// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultrank contain job fault rank process
package faultrank

import (
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess/uce"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/kube"
)

var JobFaultRankProcessor *jobRankFaultInfoProcessor

type jobRankFaultInfoProcessor struct {
	jobFaultInfoMap map[string]constant.JobFaultInfo
	mutex           sync.RWMutex
}

func init() {
	JobFaultRankProcessor = &jobRankFaultInfoProcessor{
		jobFaultInfoMap: make(map[string]constant.JobFaultInfo),
		mutex:           sync.RWMutex{},
	}
}

func (processor *jobRankFaultInfoProcessor) GetJobFaultRankInfos() map[string]constant.JobFaultInfo {
	processor.mutex.RLock()
	defer processor.mutex.RUnlock()
	result := new(map[string]constant.JobFaultInfo)
	err := util.DeepCopy(result, processor.jobFaultInfoMap)
	if err != nil {
		hwlog.RunLog.Errorf("get job fault rank failed, err: %v", err)
		return nil
	}
	hwlog.RunLog.Debugf("get job fault rank: %v", util.ObjToString(*result))
	return *result
}

// GetJobFaultRankInfosFilterLevel query jobs fault rank info, and filter fault below `faultLevel`
func (processor *jobRankFaultInfoProcessor) GetJobFaultRankInfosFilterLevel(
	faultLevel string) map[string]constant.JobFaultInfo {
	jobFaultRankInfos := processor.GetJobFaultRankInfos()
	if jobFaultRankInfos == nil {
		return nil
	}
	for jobId, jobFaultInfo := range jobFaultRankInfos {
		faultList := make([]constant.FaultRank, 0)
		for _, fault := range jobFaultInfo.FaultList {
			if fault.FaultLevel != faultLevel {
				faultList = append(faultList, fault)
			}
		}
		jobFaultInfo.FaultList = faultList
		jobFaultRankInfos[jobId] = jobFaultInfo
	}
	return jobFaultRankInfos
}

func (processor *jobRankFaultInfoProcessor) SetJobFaultRankInfos(faultInfos map[string]constant.JobFaultInfo) {
	processor.mutex.Lock()
	defer processor.mutex.Unlock()
	processor.jobFaultInfoMap = faultInfos
}

func (processor *jobRankFaultInfoProcessor) FindFaultRankForJob(nodeDeviceInfoMap map[string]constant.AdvanceDeviceFaultCm,
	nodeName string, serverList map[string]constant.ServerHccl, jobId string) []constant.FaultRank {
	advanceDeviceInfo := nodeDeviceInfoMap[nodeName]
	devicesOfJobOnNode, ok := serverList[nodeName]
	faultRankList := make([]constant.FaultRank, 0)
	if !ok || len(devicesOfJobOnNode.DeviceList) == 0 {
		return faultRankList
	}

	for _, deviceInfo := range devicesOfJobOnNode.DeviceList {
		deviceName := advanceDeviceInfo.ServerType + "-" + deviceInfo.DeviceID
		faultList, found := advanceDeviceInfo.FaultDeviceList[deviceName]
		uceInManagementPlane := false
		if found {
			// scan management plane fault info. management plane may filter uce fault in uceProcessor
			for _, fault := range faultList {
				faultRank := constant.FaultRank{
					RankId:      deviceInfo.RankID,
					FaultCode:   fault.FaultCode,
					FaultLevel:  fault.FaultLevel,
					DoStepRetry: false,
				}
				if strings.Contains(fault.FaultCode, constant.UceFaultCode) {
					// management plane find uce fault
					uceInManagementPlane = true
					faultRank.DoStepRetry = processor.canDoStepRetry(jobId, nodeName, deviceName)
				}
				faultRankList = append(faultRankList, faultRank)
			}
		}
		if uceInManagementPlane {
			continue
		}
		// business plane find uce fault
		if processor.uceInBusinessPlane(jobId, nodeName, deviceName) {
			faultRankList = append(faultRankList, constant.FaultRank{
				RankId:      deviceInfo.RankID,
				FaultCode:   constant.UceFaultCode,
				FaultLevel:  constant.RestartBusiness,
				DoStepRetry: processor.canDoStepRetry(jobId, nodeName, deviceName),
			})
		}
	}
	return faultRankList
}

func (processor *jobRankFaultInfoProcessor) canDoStepRetry(jobId, nodeName, deviceName string) bool {
	uceDevice, found := uce.UceProcessor.GetUceDeviceFromJob(jobId, nodeName, deviceName)
	if !found {
		hwlog.RunLog.Debugf("job %s's uce fault is not on node %s device %s", jobId, nodeName, deviceName)
		return false
	}
	doStepRetry := faultdomain.CanDoStepRetry(&uceDevice)
	hwlog.RunLog.Debugf("uceDevice %s stepretry %v", util.ObjToString(uceDevice), doStepRetry)
	return doStepRetry
}

func (processor *jobRankFaultInfoProcessor) uceInBusinessPlane(jobId, nodeName, deviceName string) bool {
	uceDevice, found := uce.UceProcessor.GetUceDeviceFromJob(jobId, nodeName, deviceName)
	// business plane didn't find uce fault
	if !found {
		hwlog.RunLog.Debugf("business plane didn't find uce fault")
		return false
	}
	// business plane found uce fault
	return faultdomain.ValidBusinessRecoverTime(uceDevice.RecoverTime)
}

func (fpi *jobRankFaultInfoProcessor) Process(info any) any {
	allConfigmap, ok := info.(constant.AllConfigmapContent)
	if !ok {
		hwlog.RunLog.Error("convert info to AllConfigmapContent failed")
		return info
	}
	deviceInfos := allConfigmap.DeviceCm
	deviceCmForNodeMap := faultdomain.GetAdvanceDeviceCmForNodeMap(deviceInfos)
	hwlog.RunLog.Debugf("deviceInfos: %#v", deviceInfos)
	nodeInfos := allConfigmap.NodeCm
	hwlog.RunLog.Debugf("nodeInfos: %#v", nodeInfos)
	switchInfos := allConfigmap.SwitchCm
	hwlog.RunLog.Debugf("switchInfos: %#v", switchInfos)

	jobFaultInfos := make(map[string]constant.JobFaultInfo)
	jobServerInfoMap := job.GetJobServerInfoMap()
	for jobId, serverList := range jobServerInfoMap.InfoMap {
		jobFaultInfo := constant.JobFaultInfo{
			JobId:     jobId,
			FaultList: make([]constant.FaultRank, 0),
		}
		hwlog.RunLog.Debugf("serverList: %d", len(serverList))
		fpi.findNodeDeviceAndSwitchFault(serverList, nodeInfos, &jobFaultInfo, switchInfos, deviceCmForNodeMap, jobId)
		if len(jobFaultInfo.FaultList) > 0 {
			hwlog.RunLog.Debugf("jobFaultInfo: %#v", jobFaultInfo)
		}
		jobFaultInfos[jobId] = jobFaultInfo
	}
	JobFaultRankProcessor.SetJobFaultRankInfos(jobFaultInfos)
	return nil
}

func (fpi *jobRankFaultInfoProcessor) findNodeDeviceAndSwitchFault(
	serverList map[string]constant.ServerHccl, nodeInfos map[string]*constant.NodeInfo,
	jobFaultInfo *constant.JobFaultInfo, switchInfos map[string]*constant.SwitchInfo,
	deviceCmForNodeMap map[string]constant.AdvanceDeviceFaultCm, jobId string) {
	for nodeName, server := range serverList {
		hwlog.RunLog.Debugf("nodeName: %s, server: %#v", nodeName, server)
		ni, ok := nodeInfos[constant.NodeInfoPrefix+nodeName]
		if ok && ni.NodeStatus == constant.UnHealthy {
			hwlog.RunLog.Debugf("node %s is unhealthy", nodeName)
			jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
			continue
		}
		node := kube.GetNode(nodeName)
		if node == nil || !faultdomain.IsNodeReady(node) {
			hwlog.RunLog.Debugf("node %s is not ready", nodeName)
			jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
			continue
		}
		si, ok := switchInfos[constant.SwitchInfoPrefix+nodeName]
		if ok && si.NodeStatus == constant.UnHealthy {
			hwlog.RunLog.Debugf("node %s switch is unhealthy", nodeName)
			jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
			continue
		}
		faultRankList := JobFaultRankProcessor.FindFaultRankForJob(deviceCmForNodeMap, nodeName, serverList, jobId)
		jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, faultRankList...)
	}
}

func serverHcclToFaultRank(server constant.ServerHccl) []constant.FaultRank {
	faultRanks := make([]constant.FaultRank, 0, len(server.DeviceList))
	for _, device := range server.DeviceList {
		faultRanks = append(faultRanks, constant.FaultRank{
			RankId:      device.RankID,
			FaultCode:   "",
			FaultLevel:  constant.SeparateNPU,
			DoStepRetry: false,
		})
	}
	return faultRanks
}
