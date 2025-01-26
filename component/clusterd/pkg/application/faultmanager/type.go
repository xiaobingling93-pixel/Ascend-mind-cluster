// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"

	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/application/faultmanager/faultjob"
	"clusterd/pkg/application/faultmanager/faultrank"
	"clusterd/pkg/common/constant"
)

type baseFaultCenter[T constant.ConfigMapInterface] struct {
	processorList        []constant.FaultProcessor
	lastProcessTime      int64
	subscribeChannelList []chan int
	mutex                sync.Mutex
	processPeriod        int64
	cmManager            *faultCenterCmManager[T]
	centerType           int
}

// DeviceFaultProcessCenter
type DeviceFaultProcessCenter struct {
	baseFaultCenter[*constant.DeviceInfo]
}

// GlobalFaultProcessCenter is a global instance of FaultProcessCenter used for processing faults.
var GlobalFaultProcessCenter *FaultProcessCenter

// FaultProcessCenter processes the faults and coordinates the fault handling among different components.
type FaultProcessCenter struct {
	DeviceCenter      *DeviceFaultProcessCenter
	NodeCenter        *NodeFaultProcessCenter
	SwitchCenter      *SwitchFaultProcessCenter
	FaultJobCenter    *FaultJobProcessCenter
	faultJobProcessor *faultrank.FaultProcessorImpl
	jobServerInfoMap  constant.JobServerInfoMap
	NotifyProcessChan chan int
}

type FaultJobProcessCenter struct {
	jobServerInfoMap constant.JobServerInfoMap
	lastProcessTime  int64
	deviceInfoCm     map[string]*constant.DeviceInfo
	switchInfoCm     map[string]*constant.SwitchInfo
	nodeInfoCm       map[string]*constant.NodeInfo
	FaultJobs        map[string]*faultjob.FaultJob
}

type linkDownCqeFaultProcessCenter struct {
	deviceCenter        *DeviceFaultProcessCenter
	linkDownCqeFaults   map[string]map[string]map[string]cqeLinkDownFaultRank // job, node, device FaultInfo
	nodeDeviceFaultInfo map[string]constant.AdvanceDeviceFaultCm
	cqeFaultTimeList    map[string][]int64
}

type cqeLinkDownFaultRank struct {
	LinkDownFaultTime int64
	DeviceName        string
	IsLinkDown        bool
	IsCqe             bool
}

// NodeFaultProcessCenter
type NodeFaultProcessCenter struct {
	baseFaultCenter[*constant.NodeInfo]
}

type SwitchFaultProcessCenter struct {
	baseFaultCenter[*constant.SwitchInfo]
}

type configMap[T constant.ConfigMapInterface] struct {
	configmap map[string]T
}

type faultCenterCmManager[T constant.ConfigMapInterface] struct {
	mutex        sync.RWMutex
	cmBuffer     *collector.ConfigmapCollectBuffer[T]
	originalCm   configMap[T]
	processingCm configMap[T]
	processedCm  configMap[T]
}
