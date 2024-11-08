package resource

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
	"clusterd/pkg/interface/kube"
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"sync"
	"time"
)

var processorForUceAccompanyFault = &uceAccompanyFaultProcessor{
	DiagnosisAccompanyTimeout: constant.DiagnosisAccompanyTimeout,
	uceAccompanyFaultQue:      make(map[string]map[string][]constant.DeviceFault),
	uceFaultTime:              make(map[string]map[string]int64),
}
var processorForUceFault = &uceFaultProcessor{
	JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
	JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
	mindIoReportInfo: &mindIoReportInfosForAllJobs{
		Infos:   make(map[string]map[string]map[string]mindIoReportInfo),
		RwMutex: sync.RWMutex{},
	},
}
var faultProcessCenter = &FaultProcessCenter{
	deviceFaultProcessor: []faultProcessor{processorForUceAccompanyFault, processorForUceFault},
}

func getUceFaultProcessor() (*uceFaultProcessor, error) {
	for _, processor := range faultProcessCenter.deviceFaultProcessor {
		if processor, ok := processor.(*uceFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceFaultProcessor in FaultProcessCenter")
}

func getUceAccompanyFaultProcessor() (*uceAccompanyFaultProcessor, error) {
	for _, processor := range faultProcessCenter.deviceFaultProcessor {
		if processor, ok := processor.(*uceAccompanyFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceAccompanyFaultProcessor in FaultProcessCenter")
}

// The faultProcessor process the fault information.
type faultProcessor interface {
	process()
}

// The FaultProcessCenter maintain the fault information.
type FaultProcessCenter struct {
	deviceInfos          map[string]*constant.DeviceInfo
	nodeInfos            map[string]*constant.NodeInfo
	switchInfos          map[string]*constant.SwitchInfo
	deviceFaultProcessor []faultProcessor
	nodeFaultProcessor   []faultProcessor
	switchFaultProcessor []faultProcessor
}

func (faultProcessCenter *FaultProcessCenter) QueryDeviceInfoToReport() map[string]*constant.DeviceInfo {
	cmManager.Lock()
	deviceInfos := device.DeepCopyInfos(cmManager.deviceInfoMap)
	cmManager.Unlock()
	faultProcessCenter.deviceInfos = deviceInfos
	for _, processor := range faultProcessCenter.deviceFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.deviceInfos
}

func (faultProcessCenter *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	cmManager.Lock()
	switchInfos := switchinfo.DeepCopyInfos(cmManager.switchInfoMap)
	cmManager.Unlock()
	faultProcessCenter.switchInfos = switchInfos
	for _, processor := range faultProcessCenter.switchFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.switchInfos
}

func (faultProcessCenter *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	cmManager.Lock()
	nodeInfos := node.DeepCopyInfos(cmManager.nodeInfoMap)
	cmManager.Unlock()
	faultProcessCenter.nodeInfos = nodeInfos
	for _, processor := range faultProcessCenter.nodeFaultProcessor {
		processor.process()
	}
	return faultProcessCenter.nodeInfos
}

/*
The uceFaultProcessor process uce fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type uceFaultProcessor struct {
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64

	mindIoReportInfo *mindIoReportInfosForAllJobs
	// uceJob->jobInfo
	uceDevicesOfUceJob map[string]uceJobInfo
	// node->DeviceName->uceDeviceInfo
	uceDeviceOfNode map[string]uceNodeInfo
}

// JobId->node->device->report_info
type mindIoReportInfosForAllJobs struct {
	Infos   map[string]map[string]map[string]mindIoReportInfo
	RwMutex sync.RWMutex
}

func (reportInfos *mindIoReportInfosForAllJobs) getInfo(jobId, nodeName, deviceName string) mindIoReportInfo {
	if reportInfos == nil {
		return mindIoReportInfo{
			RecoverTime:  constant.JobNotRecover,
			CompleteTime: constant.JobNotRecoverComplete,
		}
	}
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	if info, ok := reportInfos.Infos[jobId][nodeName][deviceName]; ok {
		return info
	}
	return mindIoReportInfo{
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

type mindIoReportInfo struct {
	RecoverTime  int64
	CompleteTime int64
}

func (jobInfo *uceJobInfo) initUceDeviceFromNodeAndMindIo(
	uceNode uceNodeInfo, reportInfos *mindIoReportInfosForAllJobs, serverList []*job.ServerHccl) uceNodeInfo {
	devicesOfJobOnNode := jobInfo.getDevicesNameOfJobOnNode(uceNode.NodeName, serverList)

	jobUceNodeInfo := uceNodeInfo{
		NodeName:   uceNode.NodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode {
		if uceDevice, ok := uceNode.DeviceInfo[deviceOfJob]; ok {
			reportInfo := reportInfos.getInfo(jobInfo.JobId, uceNode.NodeName, deviceOfJob)
			jobUceNodeInfo.DeviceInfo[uceDevice.DeviceName] = uceDeviceInfo{
				DeviceName:   deviceOfJob,
				FaultTime:    uceDevice.FaultTime,
				RecoverTime:  reportInfo.RecoverTime,
				CompleteTime: reportInfo.CompleteTime,
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
	processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
	currentTime := time.Now().UnixMilli()
	faultProcessCenter.deviceInfos = processor.processUceFaultInfo(currentTime)
}

func (processor *uceFaultProcessor) processUceFaultInfo(currentTime int64) map[string]*constant.DeviceInfo {
	deviceInfos := device.DeepCopyInfos(faultProcessCenter.deviceInfos)
	for cmName, deviceInfo := range deviceInfos {
		nodeName, err := util.CmNameToNodeName(cmName)
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
	if processor.currentTimeIsNotExceedMindIoReportRecoverTimeout(uceDevice, currentTime) {
		return true
	}
	if processor.mindIoRecoverTimeIsNotExceedRecoverTimeout(uceDevice) {
		if processor.currentTimeIsNotExceedMindIoReportCompleteTimeout(uceDevice, currentTime) {
			return true
		} else if processor.mindIoReportCompleteTimeIsNotExceedCompleteTimeout(uceDevice) {
			return true
		}
		return false
	}
	return false
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedMindIoReportRecoverTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return uceDevice.FaultTime >= currentTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) mindIoRecoverTimeIsNotExceedRecoverTimeout(
	uceDevice uceDeviceInfo) bool {
	return uceDevice.FaultTime >= uceDevice.RecoverTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedMindIoReportCompleteTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= currentTime
}

func (processor *uceFaultProcessor) mindIoReportCompleteTimeIsNotExceedCompleteTimeout(
	uceDevice uceDeviceInfo) bool {
	// invalid complete time
	if uceDevice.CompleteTime < uceDevice.FaultTime || uceDevice.CompleteTime < uceDevice.RecoverTime {
		return false
	}
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= uceDevice.CompleteTime
}

func (processor *uceFaultProcessor) getUceDeviceOfNodes() map[string]uceNodeInfo {
	uceNodes := make(map[string]uceNodeInfo)
	for _, deviceInfo := range faultProcessCenter.deviceInfos {
		nodeName, err := util.CmNameToNodeName(deviceInfo.CmName)
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

func (processor *uceFaultProcessor) getUceDevicesForUceTolerateJobs() map[string]uceJobInfo {
	nodesName := processor.getNodesNameFromDeviceInfo()
	uceJobs := make(map[string]uceJobInfo)
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	for jobUid, worker := range kube.JobMgr.BsWorker {
		// If job cannot tolerate uce fault, don't Filter device info
		if !processor.jobTolerateUceFault(jobUid) {
			continue
		}
		workerInfo := worker.GetWorkerInfo()
		serverList := workerInfo.CMData.GetServerList()
		jobInfo := uceJobInfo{
			// node->uceNodeInfo
			UceNode: make(map[string]uceNodeInfo),
			JobId:   jobUid,
		}
		for _, nodeName := range nodesName {
			devicesOfJobOnNode := jobInfo.getDevicesNameOfJobOnNode(nodeName, serverList)
			if len(devicesOfJobOnNode) == 0 {
				continue
			}
			jobInfo.UceNode[nodeName] =
				jobInfo.initUceDeviceFromNodeAndMindIo(
					processor.uceDeviceOfNode[nodeName], processor.mindIoReportInfo, serverList)
		}
		if len(jobInfo.UceNode) != 0 {
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

func (jobInfo *uceJobInfo) getDevicesNameOfJobOnNode(nodeName string, serverList []*job.ServerHccl) []string {
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

// CallbackForReportUceInfo cluster grpc should call back for report uce fault situation
func CallbackForReportUceInfo(jobUid, rankId string, recoverTime int64) error {
	processor, err := getUceFaultProcessor()
	if err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	nodeName, deviceId, err := kube.JobMgr.GetNodeAndDeviceFromJobIdAndRankId(jobUid, rankId)
	if err != nil {
		err = fmt.Errorf("mindIO report info failed, exception: %v", err)
		hwlog.RunLog.Error(err)
		return err
	}
	deviceName := util.DeviceID2DeviceKey(deviceId)
	processor.mindIoReportInfo.RwMutex.Lock()
	defer processor.mindIoReportInfo.RwMutex.Unlock()
	reportInfo := processor.mindIoReportInfo.Infos
	info := mindIoReportInfo{
		RecoverTime:  recoverTime,
		CompleteTime: constant.JobNotRecoverComplete,
	}
	if reportInfo == nil {
		reportInfo = make(map[string]map[string]map[string]mindIoReportInfo)
	}
	if _, ok := reportInfo[jobUid]; !ok {
		reportInfo[jobUid] = make(map[string]map[string]mindIoReportInfo)
		if _, ok := reportInfo[jobUid][nodeName]; !ok {
			reportInfo[jobUid][nodeName] = make(map[string]mindIoReportInfo)
		}
		reportInfo[jobUid][nodeName][deviceName] = info
	} else {
		if _, ok := reportInfo[jobUid][nodeName]; !ok {
			reportInfo[jobUid][nodeName] = make(map[string]mindIoReportInfo)
		}
		reportInfo[jobUid][nodeName][deviceName] = info
	}
	processor.mindIoReportInfo.Infos = reportInfo
	return nil
}

// uceAccompanyFaultProcessor:
// aic aiv fault can be 1) accompanied by uce fault, also can 2) curr alone.
// if 1) aic aiv fault should be filtered. Once find aic fault, check if there is an uce fault 5s ago
// if 2) aic aiv fault should not be retained.
type uceAccompanyFaultProcessor struct {
	// maintain 5s ago device info
	DiagnosisAccompanyTimeout int64
	// nodeName -> deviceName -> faultQue
	uceAccompanyFaultQue map[string]map[string][]constant.DeviceFault
	// uceFaultTime
	uceFaultTime map[string]map[string]int64
}

func (processor *uceAccompanyFaultProcessor) uceAccompanyFaultInQue(deviceInfos map[string]*constant.DeviceInfo) {
	for _, deviceInfo := range deviceInfos {
		nodeName, err := util.CmNameToNodeName(deviceInfo.CmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		processor.uceAccompanyFaultForNode(nodeName, deviceInfo)
	}
}

func (processor *uceAccompanyFaultProcessor) uceAccompanyFaultForNode(nodeName string, deviceInfo *constant.DeviceInfo) {
	if _, ok := processor.uceAccompanyFaultQue[nodeName]; !ok {
		processor.uceAccompanyFaultQue[nodeName] = make(map[string][]constant.DeviceFault)
		processor.uceFaultTime[nodeName] = make(map[string]int64)
	}
	faultMap := device.GetFaultMap(deviceInfo)
	for deviceName, deviceFaults := range faultMap {
		for _, fault := range deviceFaults {
			if device.IsUceFault(fault) {
				processor.uceFaultTime[nodeName][deviceName] = fault.FaultTime
				continue
			}
			if !device.IsUceAccompanyFault(fault) {
				continue
			}
			if _, ok := processor.uceAccompanyFaultQue[nodeName][deviceName]; !ok {
				processor.uceAccompanyFaultQue[nodeName][deviceName] = make([]constant.DeviceFault, 0)
			}

			// in que
			faultsInfo := processor.uceAccompanyFaultQue[nodeName][deviceName]
			processor.uceAccompanyFaultQue[nodeName][deviceName] = append(faultsInfo, fault)
		}
	}
}

func (processor *uceAccompanyFaultProcessor) filterFaultInfos(currentTime int64,
	deviceInfos map[string]*constant.DeviceInfo) map[string]*constant.DeviceInfo {
	for nodeName, nodeFaults := range processor.uceAccompanyFaultQue {
		faultMap := device.GetFaultMap(deviceInfos[util.NodeNameToCmName(nodeName)])
		for deviceName, deviceFaultQue := range nodeFaults {
			newQue, newFaultMap := processor.filterFaultDevice(faultMap, currentTime, nodeName, deviceName, deviceFaultQue)
			nodeFaults[deviceName] = newQue
			faultMap = newFaultMap
		}
		deviceInfos[util.NodeNameToCmName(nodeName)].DeviceList[device.GetFaultListKey()] = device.FaultMapToArrayToString(faultMap)
	}
	return deviceInfos
}

func (processor *uceAccompanyFaultProcessor) filterFaultDevice(
	faultMap map[string][]constant.DeviceFault, currentTime int64, nodeName, deviceName string,
	deviceFaultQue []constant.DeviceFault) ([]constant.DeviceFault, map[string][]constant.DeviceFault) {
	newDeviceFaultQue := make([]constant.DeviceFault, 0)
	for _, fault := range deviceFaultQue {
		uceFaultTime := processor.getDeviceUceFaultTime(nodeName, deviceName)
		accompanyFaultTime := fault.FaultTime
		// if is accompanied fault, filter
		if processor.isAccompaniedFaultByUce(uceFaultTime, accompanyFaultTime) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
			continue
		}
		// if current is not exceed diagnosis time,
		// then cannot decide fault is accompany or not, filter, and in que to decide in next turn.
		if !processor.isCurrentExceedDiagnosisTimeout(currentTime, accompanyFaultTime) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
			newDeviceFaultQue = append(newDeviceFaultQue, fault)
		}
	}
	return newDeviceFaultQue, faultMap
}

func (processor *uceAccompanyFaultProcessor) getDeviceUceFaultTime(nodeName, deviceName string) int64 {
	if faultTime, ok := processor.uceFaultTime[nodeName][deviceName]; ok {
		return faultTime
	}
	return constant.DeviceNotFault
}

func (processor *uceAccompanyFaultProcessor) isAccompaniedFaultByUce(
	uceFaultTime, uceAccompanyFaultTime int64) bool {
	return util.Abs(uceFaultTime-uceAccompanyFaultTime) <= processor.DiagnosisAccompanyTimeout
}

func (processor *uceAccompanyFaultProcessor) isCurrentExceedDiagnosisTimeout(
	currentTime, uceAccompanyFaultTime int64) bool {
	return uceAccompanyFaultTime < currentTime-processor.DiagnosisAccompanyTimeout
}

func (processor *uceAccompanyFaultProcessor) process() {
	deviceInfos := faultProcessCenter.deviceInfos
	processor.uceAccompanyFaultInQue(deviceInfos)
	currentTime := time.Now().UnixMilli()
	filteredFaultInfos := processor.filterFaultInfos(currentTime, deviceInfos)
	faultProcessCenter.deviceInfos = filteredFaultInfos
}
