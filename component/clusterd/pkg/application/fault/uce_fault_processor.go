package fault

import (
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
)

/*
The uceFaultProcessor process uce fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type uceFaultProcessor struct {
	deviceCenter             *deviceFaultProcessCenter
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64

	reportInfo *reportInfosForAllJobs
	// uceJob->jobInfo
	uceDevicesOfUceJob map[string]uceJobInfo
	// node->DeviceName->uceDeviceInfo
	uceDeviceOfNode  map[string]uceNodeInfo
	jobServerInfoMap job.JobServerInfoMap
}

// JobId->node->device->report_info
type reportInfosForAllJobs struct {
	InfoMap map[string]map[string]map[string]reportInfo
	RwMutex sync.RWMutex
}

func newUceFaultProcessor(deviceCenter *deviceFaultProcessCenter) *uceFaultProcessor {
	return &uceFaultProcessor{
		JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
		JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
		reportInfo: &reportInfosForAllJobs{
			InfoMap: make(map[string]map[string]map[string]reportInfo),
			RwMutex: sync.RWMutex{},
		},
		deviceCenter: deviceCenter,
	}
}

func (reportInfos *reportInfosForAllJobs) getInfo(jobId, nodeName, deviceName string) reportInfo {
	if reportInfos == nil {
		return reportInfo{
			RecoverTime:  constant.JobNotRecover,
			CompleteTime: constant.JobNotRecoverComplete,
		}
	}
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	if info, ok := reportInfos.InfoMap[jobId][nodeName][deviceName]; ok {
		return info
	}
	return reportInfo{
		RecoverTime:  constant.JobNotRecover,
		CompleteTime: constant.JobNotRecoverComplete,
	}
}

type uceDeviceInfo struct {
	// DeviceName has prefix Ascend910
	DeviceName   string
	FaultTime    int64
	RecoverTime  int64
	CompleteTime int64
}

type uceNodeInfo struct {
	NodeName string
	// DeviceName->DeviceInfo
	DeviceInfo map[string]uceDeviceInfo
}

type uceJobInfo struct {
	// UceNode node->nodeInfo
	JobId   string
	UceNode map[string]uceNodeInfo
}

type reportInfo struct {
	RecoverTime  int64
	CompleteTime int64
}

func (processor *uceFaultProcessor) initUceDeviceFromNodeAndReportInfo(jobId string,
	uceNode uceNodeInfo, devicesOfJobOnNode job.ServerHccl) uceNodeInfo {
	//devicesOfJobOnNode := getDevicesNameOfJobOnNode(uceNode.NodeName, serverList, jobId)

	jobUceNodeInfo := uceNodeInfo{
		NodeName:   uceNode.NodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode.DeviceList {
		deviceName := deviceID2DeviceKey(deviceOfJob.DeviceID)
		if uceDevice, ok := uceNode.DeviceInfo[deviceName]; ok {
			reportInfo := processor.reportInfo.getInfo(jobId, uceNode.NodeName, deviceName)
			jobUceNodeInfo.DeviceInfo[uceDevice.DeviceName] = uceDeviceInfo{
				DeviceName:   deviceName,
				FaultTime:    uceDevice.FaultTime,
				RecoverTime:  reportInfo.RecoverTime,
				CompleteTime: reportInfo.CompleteTime,
			}
		}
	}

	return jobUceNodeInfo
}

func (processor *uceFaultProcessor) process() {
	processor.jobServerInfoMap = processor.deviceCenter.jobServerInfoMap
	deviceInfos := processor.deviceCenter.getInfoMap()
	processor.uceDeviceOfNode = processor.getUceDeviceOfNodes(deviceInfos)
	processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs(deviceInfos)
	currentTime := time.Now().UnixMilli()
	filterDeviceInfos := processor.processUceFaultInfo(deviceInfos, currentTime)
	hwlog.RunLog.Infof("current deviceInfos %s", util.ObjToString(deviceInfos))
	hwlog.RunLog.Infof("currentTime: %d", currentTime)
	hwlog.RunLog.Infof("result deviceInfos %s", util.ObjToString(filterDeviceInfos))
	processor.deviceCenter.setInfoMap(filterDeviceInfos)
}

func (processor *uceFaultProcessor) processUceFaultInfo(
	deviceInfos map[string]*constant.DeviceInfo, currentTime int64) map[string]*constant.DeviceInfo {
	for cmName, deviceInfo := range deviceInfos {
		nodeName, err := cmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		faultList := processor.processEachNodeUceFaultInfo(nodeName, deviceInfo, currentTime)
		deviceInfo.DeviceList[device.GetFaultListKey()] = faultList
	}
	return deviceInfos
}

func (processor *uceFaultProcessor) processEachNodeUceFaultInfo(
	nodeName string, orgDeviceInfo *constant.DeviceInfo, currentTime int64) string {
	faultMap := device.GetFaultMap(orgDeviceInfo)
	for _, uceJob := range processor.uceDevicesOfUceJob {
		for deviceName, uceDevice := range uceJob.UceNode[nodeName].DeviceInfo {
			if processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime) {
				hwlog.RunLog.Warnf("uceFaultProcessor filtered uce device: %s on node %s, currentTime: %d",
					util.ObjToString(uceDevice), nodeName, currentTime)
				faultMap = processor.filterUceDeviceFaultInfo(deviceName, faultMap)
			}
		}
	}
	return device.FaultMapToArrayToString(faultMap)
}

func (processor *uceFaultProcessor) filterUceDeviceFaultInfo(
	deviceName string, faultMap map[string][]constant.DeviceFault) map[string][]constant.DeviceFault {
	for _, fault := range faultMap[deviceName] {
		// filter device's uce fault
		if device.IsUceFault(fault) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
		}
	}
	return faultMap
}

func (processor *uceFaultProcessor) canFilterUceDeviceFaultInfo(uceDevice uceDeviceInfo, currentTime int64) bool {
	if processor.currentTimeIsNotExceedReportRecoverTimeout(uceDevice, currentTime) {
		return true
	}
	if processor.RecoverTimeIsNotExceedRecoverTimeout(uceDevice) {
		if processor.currentTimeIsNotExceedRecoverCompleteTimeout(uceDevice, currentTime) {
			return true
		} else if processor.reportCompleteTimeIsNotExceedCompleteTimeout(uceDevice) {
			return true
		}
		return false
	}
	return false
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedReportRecoverTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return uceDevice.FaultTime >= currentTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) RecoverTimeIsNotExceedRecoverTimeout(
	uceDevice uceDeviceInfo) bool {
	return uceDevice.FaultTime >= uceDevice.RecoverTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedRecoverCompleteTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= currentTime
}

func (processor *uceFaultProcessor) reportCompleteTimeIsNotExceedCompleteTimeout(
	uceDevice uceDeviceInfo) bool {
	// invalid complete time
	if uceDevice.CompleteTime < uceDevice.FaultTime || uceDevice.CompleteTime < uceDevice.RecoverTime {
		return false
	}
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= uceDevice.CompleteTime
}

func (processor *uceFaultProcessor) getUceDeviceOfNodes(
	deviceInfos map[string]*constant.DeviceInfo) map[string]uceNodeInfo {
	uceNodes := make(map[string]uceNodeInfo)
	for _, deviceInfo := range deviceInfos {
		nodeName, err := cmNameToNodeName(deviceInfo.CmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		uceFaultDevicesOnNode := processor.getUceFaultDevices(nodeName, deviceInfo)

		if len(uceFaultDevicesOnNode.DeviceInfo) == 0 {
			continue
		}
		uceNodes[nodeName] = uceFaultDevicesOnNode
	}
	return uceNodes
}

func (processor *uceFaultProcessor) getUceDevicesForUceTolerateJobs(
	deviceInfos map[string]*constant.DeviceInfo) map[string]uceJobInfo {
	nodesName := getNodesNameFromDeviceInfo(deviceInfos)
	uceJobs := make(map[string]uceJobInfo)
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	for jobUid, serverList := range processor.jobServerInfoMap.InfoMap {
		// If job cannot tolerate uce fault, don't Filter device info
		if !kube.JobMgr.JobTolerateUceFault(jobUid) {
			continue
		}
		jobInfo := uceJobInfo{
			// node->uceNodeInfo
			UceNode: make(map[string]uceNodeInfo),
			JobId:   jobUid,
		}
		for _, nodeName := range nodesName {
			devicesOfJobOnNode := serverList[nodeName]
			if len(devicesOfJobOnNode.DeviceList) == 0 {
				continue
			}
			jobInfo.UceNode[nodeName] =
				processor.initUceDeviceFromNodeAndReportInfo(jobUid,
					processor.uceDeviceOfNode[nodeName], devicesOfJobOnNode)
		}
		if len(jobInfo.UceNode) != 0 {
			uceJobs[jobUid] = jobInfo
		}
	}
	return uceJobs
}

func (processor *uceFaultProcessor) getUceFaultDevices(nodeName string, deviceInfo *constant.DeviceInfo) uceNodeInfo {
	faultMap := device.GetFaultMap(deviceInfo)
	nodeInfo := uceNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}
	for _, deviceFaults := range faultMap {
		for _, fault := range deviceFaults {
			if !device.IsUceFault(fault) {
				continue
			}
			nodeInfo.DeviceInfo[fault.NPUName] = uceDeviceInfo{
				DeviceName:   fault.NPUName,
				FaultTime:    fault.FaultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
			}
		}
	}
	return nodeInfo
}
