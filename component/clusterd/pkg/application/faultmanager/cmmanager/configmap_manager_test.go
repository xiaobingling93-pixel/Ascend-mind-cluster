// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package cmmanager

import (
	"ascend-common/common-utils/hwlog"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"clusterd/pkg/common/constant"
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../../testdata/clusterd.log"}
	hwLogConfig.MaxBackups = 30
	hwLogConfig.MaxAge = 7
	hwLogConfig.LogLevel = 0
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	m.Run()
}

func TestFaultCenterCmManagerSetAndGetConfigmap(t *testing.T) {
	t.Run("", func(t *testing.T) {
		deviceCM := ConfigMap[*constant.DeviceInfo]{Data: make(map[string]*constant.DeviceInfo)}
		cm1 := &constant.DeviceInfo{
			DeviceInfoNoName: constant.DeviceInfoNoName{},
			CmName:           "cm1",
		}
		deviceCM.updateCmInfo(cm1, true)
		cm2 := &constant.DeviceInfo{
			DeviceInfoNoName: constant.DeviceInfoNoName{},
			CmName:           "cm2",
		}
		deviceCM.updateCmInfo(cm2, true)
		if !reflect.DeepEqual(deviceCM.Data["cm1"], cm1) ||
			!reflect.DeepEqual(deviceCM.Data["cm2"], cm2) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager := FaultCenterCmManager[*constant.DeviceInfo]{
			mutex:        sync.RWMutex{},
			originalCm:   ConfigMap[*constant.DeviceInfo]{},
			processingCm: ConfigMap[*constant.DeviceInfo]{},
			processedCm:  ConfigMap[*constant.DeviceInfo]{},
		}
		faultManager.updateOriginalCm(cm1, true)
		faultManager.updateOriginalCm(cm2, true)
		if !reflect.DeepEqual(deviceCM, faultManager.GetOriginalCm()) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager.SetProcessingCm(faultManager.GetOriginalCm())
		if !reflect.DeepEqual(deviceCM, faultManager.GetProcessingCm()) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager.SetProcessedCm(faultManager.GetProcessingCm())
		if !reflect.DeepEqual(deviceCM, faultManager.GetProcessedCm()) {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
		faultManager.updateOriginalCm(cm1, false)
		faultManager.updateOriginalCm(cm2, false)
		if len(faultManager.GetOriginalCm().Data) != 0 {
			t.Errorf("TestFaultCenterCmManagerSetAndGetDeviceInfoCm failed")
		}
	})
}
