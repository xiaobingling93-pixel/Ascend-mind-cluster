// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultrank

import (
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/uce"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
)

var JobFaultRankProcessor *JobRankFaultInfoProcessor

type JobRankFaultInfoProcessor struct {
	jobFaultInfoMap map[string]constant.JobFaultInfo
	mutex           sync.RWMutex
}

func init() {
	JobFaultRankProcessor = &JobRankFaultInfoProcessor{
		jobFaultInfoMap: make(map[string]constant.JobFaultInfo),
		mutex:           sync.RWMutex{},
	}
}

func (processor *JobRankFaultInfoProcessor) GetJobFaultRankInfos() map[string]constant.JobFaultInfo {
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

func (processor *JobRankFaultInfoProcessor) GetJobFaultRankInfosFilterLevel(
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

func (processor *JobRankFaultInfoProcessor) SetJobFaultRankInfos(faultInfos map[string]constant.JobFaultInfo) {
	processor.mutex.Lock()
	defer processor.mutex.Unlock()
	processor.jobFaultInfoMap = faultInfos
}

func (processor *JobRankFaultInfoProcessor) Process(info any) any { return info }

func (processor *JobRankFaultInfoProcessor) FindFaultRankForJob(nodeDeviceInfoMap map[string]constant.AdvanceDeviceFaultCm,
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

func (processor *JobRankFaultInfoProcessor) canDoStepRetry(jobId, nodeName, deviceName string) bool {
	uceDevice, found := uce.UceProcessor.GetUceDeviceFromJob(jobId, nodeName, deviceName)
	if !found {
		hwlog.RunLog.Debugf("job %s's uce fault is not on node %s device %s", jobId, nodeName, deviceName)
		return false
	}
	doStepRetry := faultdomain.CanDoStepRetry(&uceDevice)
	hwlog.RunLog.Debugf("uceDevice %s stepretry %v", util.ObjToString(uceDevice), doStepRetry)
	return doStepRetry
}

func (processor *JobRankFaultInfoProcessor) uceInBusinessPlane(jobId, nodeName, deviceName string) bool {
	uceDevice, found := uce.UceProcessor.GetUceDeviceFromJob(jobId, nodeName, deviceName)
	// business plane didn't find uce fault
	if !found {
		hwlog.RunLog.Debugf("business plane didn't find uce fault")
		return false
	}
	// business plane found uce fault
	return faultdomain.ValidBusinessRecoverTime(uceDevice.RecoverTime)
}
