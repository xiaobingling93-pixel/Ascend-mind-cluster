// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package retry contain uce process method
package retry

import (
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/faultdomain/collector"
	"clusterd/pkg/domain/job"
)

var RetryProcessor *retryFaultProcessor

/*
The retryFaultProcessor process retry fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type retryFaultProcessor struct {
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64

	// retryJob->jobInfo
	retryDevicesOfJob map[string]constant.RetryJobInfo
	// node->DeviceName->retryDeviceInfo
	retryDeviceOfNode map[string]constant.RetryNodeInfo
	jobServerInfoMap  constant.JobServerInfoMap
	nodeDeviceCmMap   map[string]*constant.AdvanceDeviceFaultCm
}

func init() {
	RetryProcessor = &retryFaultProcessor{
		JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
		JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
	}
}

func (processor *retryFaultProcessor) initRetryDeviceFromNodeAndReportInfo(jobId string, nodeName string) constant.RetryNodeInfo {
	managerPlaneRetryNode := processor.retryDeviceOfNode[nodeName]
	devicesOfJobOnNode := processor.jobServerInfoMap.InfoMap[jobId][nodeName]
	jobRetryNodeInfo := constant.RetryNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]constant.RetryDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode.DeviceList {
		deviceName := processor.nodeDeviceCmMap[nodeName].DeviceType + "-" + deviceOfJob.DeviceID
		retryReportInfo := collector.ReportInfoCollector.GetInfo(jobId, nodeName, deviceName)
		jobRetryDevice := constant.RetryDeviceInfo{
			DeviceName:   deviceName,
			FaultTime:    constant.DeviceNotFault,
			RecoverTime:  retryReportInfo.RecoverTime,
			CompleteTime: retryReportInfo.CompleteTime,
		}
		// management plane found retry fault
		if retryDevice, ok := managerPlaneRetryNode.DeviceInfo[deviceName]; ok {
			jobRetryDevice.FaultTime = retryDevice.FaultTime
			jobRetryDevice.FaultType = retryDevice.FaultType
			jobRetryNodeInfo.DeviceInfo[deviceName] = jobRetryDevice
		} else if faultdomain.ValidBusinessRetryReportInfo(&retryReportInfo) { // business plane found retry fault
			jobRetryDevice.FaultType = retryReportInfo.FaultType
			jobRetryNodeInfo.DeviceInfo[deviceName] = jobRetryDevice
		}
	}

	return jobRetryNodeInfo
}

// Process retry fault
func (processor *retryFaultProcessor) Process(info any) any {
	processContent, ok := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	if !ok {
		hwlog.RunLog.Errorf("%v cannot convert to DeviceInfo", info)
		return info
	}

	processor.jobServerInfoMap = job.GetJobServerInfoMap()
	processor.nodeDeviceCmMap = processContent.AllConfigmap
	hwlog.RunLog.Debugf("current nodeDeviceCmMap %v", processor.nodeDeviceCmMap)

	processor.retryDeviceOfNode = processor.getRetryDeviceOfNodes()
	hwlog.RunLog.Debugf("current retryDeviceOfNode %v", processor.retryDeviceOfNode)

	processor.retryDevicesOfJob = processor.getRetryDevicesForTolerateJobs()
	hwlog.RunLog.Debugf("current retryDevicesOfJob %v", processor.retryDevicesOfJob)

	currentTime := time.Now().UnixMilli()
	hwlog.RunLog.Debugf("currentTime %d", currentTime)

	processor.processRetryFaultInfo(currentTime)

	hwlog.RunLog.Debugf("result deviceInfos %v", processContent.AllConfigmap)
	return processContent
}

func (processor *retryFaultProcessor) processRetryFaultInfo(currentTime int64) {
	for nodeName, advanceDeviceInfo := range processor.nodeDeviceCmMap {
		advanceDeviceInfo = processor.processEachNodeRetryFaultInfo(nodeName, advanceDeviceInfo, currentTime)
		processor.nodeDeviceCmMap[nodeName] = advanceDeviceInfo
	}
}

func (processor *retryFaultProcessor) processEachNodeRetryFaultInfo(
	nodeName string, deviceInfo *constant.AdvanceDeviceFaultCm, currentTime int64) *constant.AdvanceDeviceFaultCm {
	modified := false
	for _, retryJob := range processor.retryDevicesOfJob {
		for deviceName, retryDevice := range retryJob.RetryNode[nodeName].DeviceInfo {
			log := fmt.Sprintf("filter retry device: %s on node %s, "+
				"currentTime: %s, faultTime: %s, recoverTime: %s, faultType: %s ",
				retryDevice.DeviceName, nodeName,
				util.ReadableMsTime(currentTime),
				util.ReadableMsTime(retryDevice.FaultTime),
				util.ReadableMsTime(retryDevice.RecoverTime),
				retryDevice.FaultType,
			)
			if processor.canFilterRetryDeviceFaultInfo(retryDevice, currentTime) {
				hwlog.RunLog.Warn("retryFaultProcessor " + log)
				processor.filterRetryDeviceFaultInfo(deviceName, deviceInfo)
				modified = true
			} else {
				hwlog.RunLog.Warn("retryFaultProcessor cannot " + log)
			}
		}
	}
	if modified {
		faultdomain.SortDataForAdvanceDeviceInfo(deviceInfo)
	}
	return deviceInfo
}

func (processor *retryFaultProcessor) filterRetryDeviceFaultInfo(
	deviceName string, advanceDevInfo *constant.AdvanceDeviceFaultCm) {
	for _, fault := range advanceDevInfo.FaultDeviceList[deviceName] {
		// filter device's retry fault
		if faultdomain.IsUceFault(fault.FaultCode) || faultdomain.IsHcclRetryFault(fault.FaultCode) {
			advanceDevInfo.DelFaultAndFix(fault)
		}
	}
}

func (processor *retryFaultProcessor) canFilterRetryDeviceFaultInfo(retryDevice constant.RetryDeviceInfo, currentTime int64) bool {
	if processor.currentTimeIsNotExceedReportRecoverTimeout(retryDevice, currentTime) {
		return true
	}
	if processor.RecoverTimeIsNotExceedRecoverTimeout(retryDevice) {
		if processor.currentTimeIsNotExceedRecoverCompleteTimeout(retryDevice, currentTime) {
			return true
		} else if processor.reportCompleteTimeIsNotExceedCompleteTimeout(retryDevice) {
			return true
		}
		return false
	}
	return false
}

func (processor *retryFaultProcessor) currentTimeIsNotExceedReportRecoverTimeout(
	retryDevice constant.RetryDeviceInfo, currentTime int64) bool {
	return retryDevice.FaultTime >= currentTime-processor.JobReportRecoverTimeout
}

func (processor *retryFaultProcessor) RecoverTimeIsNotExceedRecoverTimeout(
	retryDevice constant.RetryDeviceInfo) bool {
	return retryDevice.FaultTime >= retryDevice.RecoverTime-processor.JobReportRecoverTimeout
}

func (processor *retryFaultProcessor) currentTimeIsNotExceedRecoverCompleteTimeout(
	retryDevice constant.RetryDeviceInfo, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+retryDevice.RecoverTime >= currentTime
}

func (processor *retryFaultProcessor) reportCompleteTimeIsNotExceedCompleteTimeout(
	retryDevice constant.RetryDeviceInfo) bool {
	// invalid complete time
	if retryDevice.CompleteTime < retryDevice.FaultTime || retryDevice.CompleteTime < retryDevice.RecoverTime {
		return false
	}
	return processor.JobReportCompleteTimeout+retryDevice.RecoverTime >= retryDevice.CompleteTime
}

func (processor *retryFaultProcessor) getRetryDeviceOfNodes() map[string]constant.RetryNodeInfo {
	retryNodes := make(map[string]constant.RetryNodeInfo)
	for nodeName, deviceInfo := range processor.nodeDeviceCmMap {
		retryFaultDevicesOnNode := processor.getRetryFaultDevices(nodeName, deviceInfo)

		if len(retryFaultDevicesOnNode.DeviceInfo) == 0 {
			continue
		}
		retryNodes[nodeName] = retryFaultDevicesOnNode
	}
	return retryNodes
}

func (processor *retryFaultProcessor) getRetryDevicesForTolerateJobs() map[string]constant.RetryJobInfo {
	nodeNameList := make([]string, 0)
	for key, _ := range processor.nodeDeviceCmMap {
		nodeNameList = append(nodeNameList, key)
	}
	retryJobs := make(map[string]constant.RetryJobInfo)
	for jobUid, serverList := range processor.jobServerInfoMap.InfoMap {
		if !processor.jobServerInfoMap.RetryTolerate[jobUid] {
			continue
		}
		jobInfo := constant.RetryJobInfo{
			RetryNode: make(map[string]constant.RetryNodeInfo),
			JobId:     jobUid,
		}
		for _, nodeName := range nodeNameList {
			devicesOfJobOnNode := serverList[nodeName]
			if len(devicesOfJobOnNode.DeviceList) == 0 {
				continue
			}
			jobInfo.RetryNode[nodeName] = processor.initRetryDeviceFromNodeAndReportInfo(jobUid, nodeName)
		}
		if len(jobInfo.RetryNode) != 0 {
			retryJobs[jobUid] = jobInfo
		}
	}
	return retryJobs
}

func (processor *retryFaultProcessor) getRetryFaultDevices(
	nodeName string, deviceInfo *constant.AdvanceDeviceFaultCm) constant.RetryNodeInfo {
	nodeInfo := constant.RetryNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]constant.RetryDeviceInfo),
	}
	for _, deviceFaults := range deviceInfo.FaultDeviceList {
		for _, fault := range deviceFaults {
			if !faultdomain.IsUceFault(fault.FaultCode) && !faultdomain.IsHcclRetryFault(fault.FaultCode) {
				continue
			}
			errorMsg := fmt.Sprintf("getRetryFaultDevices cannot find retry fault time for device %s of node %s",
				deviceInfo.CmName, nodeName)
			faultTime := faultdomain.GetFaultTime(fault, errorMsg)
			var faultType string
			if faultdomain.IsUceFault(fault.FaultCode) {
				faultType = constant.UceFaultType
			} else if faultdomain.IsHcclRetryFault(fault.FaultCode) {
				faultType = constant.HcclFaultType
			}
			nodeInfo.DeviceInfo[fault.NPUName] = constant.RetryDeviceInfo{
				DeviceName:   fault.NPUName,
				FaultTime:    faultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
				FaultType:    faultType,
			}
		}
	}
	return nodeInfo
}

func (processor *retryFaultProcessor) GetRetryDeviceFromJob(jobId, nodeName, deviceName string) (constant.RetryDeviceInfo, bool) {
	jobInfo, found := processor.retryDevicesOfJob[jobId]
	if !found {
		hwlog.RunLog.Debugf("job %s has no retry fault", jobId)
		return constant.RetryDeviceInfo{}, false
	}
	retryDevice, found := jobInfo.RetryNode[nodeName].DeviceInfo[deviceName]
	if !found {
		hwlog.RunLog.Debugf("job %s's retry fault is not on node %s device %s", jobId, nodeName, deviceName)
		return constant.RetryDeviceInfo{}, false
	}
	return retryDevice, true
}
