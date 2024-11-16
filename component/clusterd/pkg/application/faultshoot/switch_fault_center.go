// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultshoot contain fault process
package faultshoot

import (
	"sync"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/switchinfo"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

func newSwitchFaultProcessCenter() *switchFaultProcessCenter {
	return &switchFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(),
		processedCm:     make(map[string]*constant.SwitchInfo),
		devicePluginCm:  make(map[string]*constant.SwitchInfo),
		mutex:           sync.RWMutex{},
	}
}

func (switchCenter *switchFaultProcessCenter) getInfoMap() map[string]*constant.SwitchInfo {
	switchCenter.mutex.RLock()
	defer switchCenter.mutex.RUnlock()
	return switchinfo.DeepCopyInfos(switchCenter.processedCm)
}

func (switchCenter *switchFaultProcessCenter) setInfoMap(infos map[string]*constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	switchCenter.processedCm = switchinfo.DeepCopyInfos(infos)
}

func (switchCenter *switchFaultProcessCenter) updateInfoFromCm(newInfo *constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	length := len(switchCenter.devicePluginCm)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	switchCenter.devicePluginCm[newInfo.CmName] = newInfo
}

func (switchCenter *switchFaultProcessCenter) delInfoFromCm(newInfo *constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	delete(switchCenter.devicePluginCm, newInfo.CmName)
}

func (switchCenter *switchFaultProcessCenter) process() {
	switchCenter.setInfoMap(switchCenter.devicePluginCm)
	switchCenter.baseFaultCenter.process()
}
