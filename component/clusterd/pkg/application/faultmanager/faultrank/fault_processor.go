// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultrank

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/kube"
)

var FaultProcessor *FaultProcessorImpl

type FaultProcessorImpl struct {
}

func init() {
	FaultProcessor = &FaultProcessorImpl{}
}

func (fpi *FaultProcessorImpl) Process(info any) any {
	allConfigmap, ok := info.(constant.AllConfigmapContent)
	if !ok {
		hwlog.RunLog.Errorf("%v cannot convert to AllConfigmapContent", info)
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
		for nodeName, server := range serverList {
			hwlog.RunLog.Debugf("nodeName: %s, server: %#v", nodeName, server)
			ni, ok := nodeInfos[constant.NodeInfoPrefix+nodeName]
			if ok && ni.NodeStatus == "UnHealthy" {
				hwlog.RunLog.Infof("node %s is unhealthy", nodeName)
				jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
				continue
			}
			node := kube.GetNode(nodeName)
			if node == nil || !faultdomain.IsNodeReady(node) {
				hwlog.RunLog.Infof("node %s is not ready", nodeName)
				jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
				continue
			}
			si, ok := switchInfos[constant.SwitchInfoPrefix+nodeName]
			if ok && si.NodeStatus == "UnHealthy" {
				hwlog.RunLog.Infof("node %s switch is unhealthy", nodeName)
				jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
				continue
			}
			faultRankList := JobFaultRankProcessor.FindFaultRankForJob(deviceCmForNodeMap, nodeName, serverList, jobId)
			jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, faultRankList...)

		}
		if len(jobFaultInfo.FaultList) > 0 {
			hwlog.RunLog.Infof("jobFaultInfo: %#v", jobFaultInfo)
		}
		jobFaultInfos[jobId] = jobFaultInfo
	}
	JobFaultRankProcessor.SetJobFaultRankInfos(jobFaultInfos)
	return nil
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
