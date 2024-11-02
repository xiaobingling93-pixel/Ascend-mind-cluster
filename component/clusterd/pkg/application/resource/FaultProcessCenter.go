package resource

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
	"encoding/json"
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"math"
	"time"
)

var faultProcessCenter FaultProcessCenter

const MindIoNotRecover = int64(math.MaxInt64)
const MindIoNotComplete = int64(math.MaxInt64)

// The faultProcessor process the fault information.
type faultProcessor interface {
	process()
}

type FaultProcessCenter struct {
	deviceInfos          map[string]*constant.DeviceInfo
	nodeInfos            map[string]*constant.NodeInfo
	switchInfos          map[string]*constant.SwitchInfo
	deviceFaultProcessor []faultProcessor
	nodeFaultProcessor   []faultProcessor
	switchFaultProcessor []faultProcessor
}

/*
The uceFaultProcessor process uce fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type uceFaultProcessor struct {
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64
	// jobUid->node->device->recover_moment
	mindIoReportRecoverInfo map[string]map[string]map[string]mindIoReportInfo
	// uceJob->jobInfo
	uceDevicesOfUceJob map[string]uceJobInfo
	// node->deviceName->uceDeviceInfo
	uceDeviceOfNode map[string]uceNodeInfo
}

type uceDeviceInfo struct {
	// deviceName has prefix Ascend910
	deviceName     string
	faultMoment    int64
	recoverMoment  int64
	completeMoment int64
}

type uceNodeInfo struct {
	nodeName string
	// deviceName->deviceInfo
	deviceInfo map[string]uceDeviceInfo
}

type uceJobInfo struct {
	// uceNode node->nodeInfo
	jobUid  string
	uceNode map[string]uceNodeInfo
}

type mindIoReportInfo struct {
	recoverMoment  int64
	completeMoment int64
}

func (jobInfo *uceJobInfo) getDeviceInfoFromNodeAndDevice(nodeName, deviceName string) (uceDeviceInfo, error) {
	nodeInfo, ok := jobInfo.uceNode[nodeName]
	if !ok {
		return uceDeviceInfo{}, fmt.Errorf("getDeviceInfoFromNodeAndDevice failed. Job %v doesn't on node %v",
			jobInfo.jobUid, nodeName)
	}
	deviceInfo, ok := nodeInfo.deviceInfo[deviceName]
	if !ok {
		return uceDeviceInfo{}, fmt.Errorf("getDeviceInfoFromNodeAndDevice failed. "+
			"Job %v doesn't on device %v of node %v", jobInfo.jobUid, deviceName, nodeName)
	}
	return deviceInfo, nil
}

func (jobInfo *uceJobInfo) initUceDeviceFromNodeAndMindIo(uceNode uceNodeInfo, processor *uceFaultProcessor) uceNodeInfo {
	devicesOfJobOnNode := jobInfo.getDevicesNameOfJobOnNode(uceNode.nodeName)

	jobUceNodeInfo := uceNodeInfo{
		nodeName:   uceNode.nodeName,
		deviceInfo: make(map[string]uceDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode {
		if uceDevice, ok := uceNode.deviceInfo[deviceOfJob]; ok {
			recoverMoment := MindIoNotRecover
			completeMoment := MindIoNotComplete
			if reportInfo, ok := processor.mindIoReportRecoverInfo[jobInfo.jobUid][uceNode.nodeName][deviceOfJob]; ok {
				recoverMoment = reportInfo.recoverMoment
				completeMoment = reportInfo.completeMoment
			}
			jobUceNodeInfo.deviceInfo[uceDevice.deviceName] = uceDeviceInfo{
				deviceName:     deviceOfJob,
				faultMoment:    uceDevice.faultMoment,
				recoverMoment:  recoverMoment,
				completeMoment: completeMoment,
			}
		}
	}

	return jobUceNodeInfo
}

func (processor *uceFaultProcessor) process() {
	if kube.JobMgr == nil {
		hwlog.RunLog.Infof("jobMgr is nil, cannot Filter uce fault report")
		return
	}
	processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
	processor.uceDevicesOfUceJob = processor.getUceDevicesOfUceTolerateJobs()
	currentTime := time.Now().UnixMilli()
	faultProcessCenter.deviceInfos = processor.processUceFaultInfo(faultProcessCenter.deviceInfos, currentTime)
}

func (processor *uceFaultProcessor) processUceFaultInfo(
	deviceInfos map[string]*constant.DeviceInfo, currentTime int64) map[string]*constant.DeviceInfo {
	for cmName, deviceInfo := range deviceInfos {
		nodeName, err := util.CmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		faultList := processor.processEachNodeUceFaultInfo(nodeName, deviceInfo, currentTime)
		faultListBytes, err := json.Marshal(faultList)
		if err != nil {
			hwlog.RunLog.Error("Marshal fault list for node %v failed. Exception: %v", nodeName, err)
		}

		deviceInfo.DeviceList[device.GetFaultListKey()] = string(faultListBytes)
	}
	return deviceInfos
}

func (processor *uceFaultProcessor) processEachNodeUceFaultInfo(
	nodeName string, orgDeviceInfo *constant.DeviceInfo, currentTime int64) []constant.DeviceFault {
	faultMap := device.GetFaultMap(orgDeviceInfo)
	for _, uceJob := range processor.uceDevicesOfUceJob {
		for deviceName, uceDevice := range uceJob.uceNode[nodeName].deviceInfo {
			if processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime) {
				faultMap = processor.filterUceDeviceFaultInfo(deviceName, faultMap)
			}
		}
	}
	deviceFaultList := make([]constant.DeviceFault, 0)
	for _, faultDevice := range faultMap {
		deviceFaultList = append(deviceFaultList, faultDevice)
	}
	return deviceFaultList
}

func (processor *uceFaultProcessor) filterUceDeviceFaultInfo(
	deviceName string, faultMap map[string]constant.DeviceFault) map[string]constant.DeviceFault {
	delete(faultMap, deviceName)
	return faultMap
}

func (processor *uceFaultProcessor) canFilterUceDeviceFaultInfo(uceDevice uceDeviceInfo, currentTime int64) bool {
	if processor.currentTimeIsNotExceedMindIoReportRecoverTimeout(uceDevice, currentTime) {
		return true
	}
	if processor.mindIoRecoverMomentIsNotExceedRecoverTimeout(uceDevice) {
		if processor.currentTimeIsNotExceedMindIoReportCompleteTimeout(uceDevice, currentTime) {
			return true
		} else if processor.mindIoReportCompleteMomentIsNotExceedCompleteTimeout(uceDevice) {
			return true
		}
		return false
	}
	return false
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedMindIoReportRecoverTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return uceDevice.faultMoment > currentTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) mindIoRecoverMomentIsNotExceedRecoverTimeout(
	uceDevice uceDeviceInfo) bool {
	return uceDevice.faultMoment > uceDevice.recoverMoment-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedMindIoReportCompleteTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+uceDevice.recoverMoment > currentTime
}

func (processor *uceFaultProcessor) mindIoReportCompleteMomentIsNotExceedCompleteTimeout(
	uceDevice uceDeviceInfo) bool {
	return processor.JobReportCompleteTimeout+uceDevice.recoverMoment > uceDevice.completeMoment
}

func (processor *uceFaultProcessor) getUceDeviceOfNodes() map[string]uceNodeInfo {
	uceNodes := make(map[string]uceNodeInfo)
	for cmName, deviceInfo := range faultProcessCenter.deviceInfos {
		nodeName, err := util.CmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		uceFaultDevicesOnNode := processor.getUceFaultDevices(nodeName, deviceInfo)

		if len(uceFaultDevicesOnNode.deviceInfo) == 0 {
			continue
		}
		uceNodes[nodeName] = uceFaultDevicesOnNode
	}
	return uceNodes
}

func (processor *uceFaultProcessor) getNodesNameFromDeviceInfo() []string {
	nodesName := make([]string, 0)
	for cmName, _ := range faultProcessCenter.deviceInfos {
		nodeName, err := util.CmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		nodesName = append(nodesName, nodeName)
	}
	return nodesName
}

func (processor *uceFaultProcessor) getUceDevicesOfUceTolerateJobs() map[string]uceJobInfo {
	nodesName := processor.getNodesNameFromDeviceInfo()
	uceJobs := make(map[string]uceJobInfo)
	for jobUid := range kube.JobMgr.BsWorker {
		// If job cannot tolerate uce fault, don't Filter device info
		if !processor.jobTolerateUceFault(jobUid) {
			continue
		}

		jobInfo := uceJobInfo{
			// node->uceNodeInfo
			uceNode: make(map[string]uceNodeInfo),
			jobUid:  jobUid,
		}
		for _, nodeName := range nodesName {
			devicesOfJobOnNode := jobInfo.getDevicesNameOfJobOnNode(nodeName)
			if len(devicesOfJobOnNode) == 0 {
				continue
			}
			jobInfo.uceNode[nodeName] = jobInfo.initUceDeviceFromNodeAndMindIo(processor.uceDeviceOfNode[nodeName], processor)
		}
		if len(jobInfo.uceNode) != 0 {
			uceJobs[jobUid] = jobInfo
		}
	}
	return uceJobs
}

// TODO 如何判断job是uce容忍的
func (processor *uceFaultProcessor) jobTolerateUceFault(jobUid string) bool {
	return true
}

func (processor *uceFaultProcessor) getUceFaultDevices(nodeName string, deviceInfo *constant.DeviceInfo) uceNodeInfo {
	faultList := device.GetFaultMap(deviceInfo)
	nodeInfo := uceNodeInfo{
		nodeName:   nodeName,
		deviceInfo: make(map[string]uceDeviceInfo),
	}
	for _, faultDevice := range faultList {
		nodeInfo.deviceInfo[faultDevice.NPUName] = uceDeviceInfo{
			deviceName:  faultDevice.NPUName,
			faultMoment: faultDevice.FaultTime,
		}
	}
	return nodeInfo
}

// CallbackForReportUceInfo cluster grpc should call back for report uce fault situation
func (processor *uceFaultProcessor) CallbackForReportUceInfo(jobUid, nodeName, deviceName string, info mindIoReportInfo) {
	reportInfo := processor.mindIoReportRecoverInfo
	if _, ok := reportInfo[jobUid]; !ok {
		reportInfo[jobUid] = make(map[string]map[string]mindIoReportInfo)
		if _, ok := reportInfo[jobUid][nodeName]; !ok {
			reportInfo[jobUid][nodeName] = make(map[string]mindIoReportInfo)
		} else {
			reportInfo[jobUid][nodeName][deviceName] = info
		}
	} else {
		if _, ok := reportInfo[jobUid][nodeName]; !ok {
			reportInfo[jobUid][nodeName] = make(map[string]mindIoReportInfo)
		} else {
			reportInfo[jobUid][nodeName][deviceName] = info
		}
	}
}

func (jobInfo *uceJobInfo) getDevicesNameOfJobOnNode(nodeName string) []string {
	worker := kube.JobMgr.BsWorker[jobInfo.jobUid]
	workerInfo := worker.GetWorkerInfo()
	serverList := workerInfo.CMData.GetServerList()
	var devices []*job.Device
	for _, server := range serverList {
		if server.ServerName != nodeName {
			continue
		}
		devices = server.DeviceList
	}
	devicesName := make([]string, len(devices))
	for idx, dev := range devices {
		devicesName[idx] = util.DeviceID2DeviceKey(dev.DeviceID)
	}
	return devicesName
}

func (faultProcessCenter *FaultProcessCenter) QueryDeviceInfoToReport() map[string]*constant.DeviceInfo {
	for _, processor := range faultProcessCenter.deviceFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.deviceInfos
}

func (faultProcessCenter *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	for _, processor := range faultProcessCenter.switchFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.switchInfos
}

func (faultProcessCenter *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	for _, processor := range faultProcessCenter.nodeFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.nodeInfos
}
