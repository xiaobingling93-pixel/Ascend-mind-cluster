// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package cmprocesscenter

import (
	"clusterd/pkg/application/faultmanager/cmmanager"
	"testing"

	"clusterd/pkg/common/constant"
)

type fakeProcessor struct{}

func (f *fakeProcessor) Process(info any) any {
	return info
}

func TestBaseFaultCenterProcess(t *testing.T) {
	t.Run("TestBaseFaultCenterProcess", func(t *testing.T) {
		manager := cmmanager.DeviceCenterCmManager
		baseCenter := newBaseFaultCenter(manager, constant.DeviceProcessType)
		baseCenter.addProcessors([]constant.FaultProcessor{&fakeProcessor{}})
		notifyChan := make(chan int, 1)
		baseCenter.Register(notifyChan)
		baseCenter.Process()
		if baseCenter.GetProcessedCm() == nil {
			t.Errorf("TestBaseFaultCenterProcess failed")
		}
	})
}
