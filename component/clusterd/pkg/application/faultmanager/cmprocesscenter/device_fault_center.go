// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package cmprocesscenter

import (
	"clusterd/pkg/application/faultmanager/cmmanager"
	"clusterd/pkg/application/faultmanager/uce"
	"clusterd/pkg/application/faultmanager/uce_accompany"
	"clusterd/pkg/common/constant"
)

var DeviceCenter *DeviceFaultProcessCenter

// DeviceFaultProcessCenter
type DeviceFaultProcessCenter struct {
	BaseFaultCenter[*constant.DeviceInfo]
}

func init() {
	manager := cmmanager.DeviceCenterCmManager
	DeviceCenter = &DeviceFaultProcessCenter{
		BaseFaultCenter: newBaseFaultCenter(manager, constant.DeviceProcessType),
	}

	DeviceCenter.addProcessors([]constant.FaultProcessor{
		uce_accompany.UceAccompanyProcessor, // this processor filter the uce accompany faults, before processorForUceFault
		uce.UceProcessor,                    // this processor filter the uce faults.
	})
}
