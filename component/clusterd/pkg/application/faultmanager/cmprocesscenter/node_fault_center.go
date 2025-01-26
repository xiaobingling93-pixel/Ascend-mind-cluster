// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package cmprocesscenter

import (
	"clusterd/pkg/application/faultmanager/cmmanager"
	"clusterd/pkg/common/constant"
)

var NodeCenter *NodeFaultProcessCenter

func init() {
	manager := cmmanager.NodeCenterCmManager
	NodeCenter = &NodeFaultProcessCenter{
		BaseFaultCenter: newBaseFaultCenter(manager, constant.NodeProcessType),
	}
}

// NodeFaultProcessCenter
type NodeFaultProcessCenter struct {
	BaseFaultCenter[*constant.NodeInfo]
}
