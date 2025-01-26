// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package cmprocesscenter

import (
	"clusterd/pkg/application/faultmanager/cmmanager"
	"clusterd/pkg/common/constant"
)

var SwitchCenter *SwitchFaultProcessCenter

func init() {
	manager := cmmanager.SwitchCenterCmManager
	SwitchCenter = &SwitchFaultProcessCenter{
		BaseFaultCenter: newBaseFaultCenter(manager, constant.SwitchProcessType),
	}
}

type SwitchFaultProcessCenter struct {
	BaseFaultCenter[*constant.SwitchInfo]
}
