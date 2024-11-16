// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultshoot contain fault process
package faultshoot

import (
	"sync"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

func newNodeFaultProcessCenter() *nodeFaultProcessCenter {
	return &nodeFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(),
		processedCm:     make(map[string]*constant.NodeInfo),
		devicePluginCm:  make(map[string]*constant.NodeInfo),
		mutex:           sync.RWMutex{},
	}
}

func (nodeCenter *nodeFaultProcessCenter) getInfoMap() map[string]*constant.NodeInfo {
	nodeCenter.mutex.RLock()
	defer nodeCenter.mutex.RUnlock()
	return node.DeepCopyInfos(nodeCenter.processedCm)
}

func (nodeCenter *nodeFaultProcessCenter) setInfoMap(infos map[string]*constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	nodeCenter.processedCm = node.DeepCopyInfos(infos)
}

func (nodeCenter *nodeFaultProcessCenter) updateInfoFromCm(newInfo *constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	length := len(nodeCenter.devicePluginCm)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	nodeCenter.devicePluginCm[newInfo.CmName] = newInfo
}

func (nodeCenter *nodeFaultProcessCenter) delInfoFromCm(newInfo *constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	delete(nodeCenter.devicePluginCm, newInfo.CmName)
}

func (nodeCenter *nodeFaultProcessCenter) process() {
	nodeCenter.setInfoMap(nodeCenter.devicePluginCm)
	nodeCenter.baseFaultCenter.process()
}
