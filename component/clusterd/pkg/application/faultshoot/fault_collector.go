/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
 */

// Package faultshoot contain fault process
package faultshoot

import (
	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
)

// DeviceInfoCollector collects device info
func DeviceInfoCollector(oldDevInfo, newDevInfo *constant.DeviceInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		hwlog.RunLog.Info("add")
	} else if operator == constant.DeleteOperator {
		hwlog.RunLog.Info("del")
	}
}

// SwitchInfoCollector collects switchinfo info of 900A3
func SwitchInfoCollector(oldSwitchInfo, newSwitchInfo *constant.SwitchInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		hwlog.RunLog.Info("add")
	} else if operator == constant.DeleteOperator {
		hwlog.RunLog.Info("del")
	}
}

// NodeCollector collects node info
func NodeCollector(oldNodeInfo, newNodeInfo *constant.NodeInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		hwlog.RunLog.Info("add")
	} else if operator == constant.DeleteOperator {
		hwlog.RunLog.Info("del")
	}
}
