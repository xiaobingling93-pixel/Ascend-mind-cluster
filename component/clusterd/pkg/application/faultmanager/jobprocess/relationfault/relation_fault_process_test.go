// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package relationfault contain relation fault process
package relationfault

import (
	"testing"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func TestFaultJobProcessNetworkFault(t *testing.T) {
	t.Run("processNetworkFault_noTrigger", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   "0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := constant.FaultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0003",
		}
		networkFaults = append(networkFaults, &fault1, &fault2)
		retryEventList := make([]constant.FaultInfo, 0)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()
		t.Log(util.ObjToString(faultJob.FaultStrategy))
	})

	t.Run("processNetworkFault_not_all_network_in_RelationFaults", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   "0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := constant.FaultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0004",
		}
		networkFaults = append(networkFaults, &fault1, &fault2)
		retryEventList := make([]constant.FaultInfo, 0)
		retryEvent := constant.FaultInfo{
			NodeName:  "node-102",
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "0x0001",
		}
		retryEventList = append(retryEventList, retryEvent)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()
		t.Log(util.ObjToString(faultJob.FaultStrategy))

	})

	t.Run("processNetworkFault_right_node_and_device_separate", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   "0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := constant.FaultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0003",
		}
		fault3 := constant.FaultInfo{
			NodeName:  "node-103",
			NPUName:   "Ascend-910-2",
			FaultType: constant.SwitchFault,
			FaultCode: "0x0003",
		}
		networkFaults = append(networkFaults, &fault1, &fault2, &fault3)
		retryEventList := make([]constant.FaultInfo, 0)
		retryEvent := constant.FaultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "0x0001",
		}
		retryEventList = append(retryEventList, retryEvent)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()
		t.Log(util.ObjToString(faultJob.FaultStrategy))
	})

	t.Run("processNetworkFault_right_node_separate_device_subHealth", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   "Trigger0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SubHealthFaultStrategy,
		})
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   "Trigger0x0002",
			RelationFaults: []string{"0x0004"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := constant.FaultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0003",
		}
		nodeFault := constant.FaultInfo{
			NodeName:  "node-103",
			NPUName:   "Ascend-910-2",
			FaultType: constant.SwitchFault,
			FaultCode: "0x0004",
		}

		networkFaults = append(networkFaults, &fault1, &fault2, &nodeFault)
		retryEventList := make([]constant.FaultInfo, 0)
		retryEvent := constant.FaultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "Trigger0x0001",
		}
		retryEvent2 := constant.FaultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "Trigger0x0002",
		}
		retryEventList = append(retryEventList, retryEvent, retryEvent2)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()

		t.Log(util.ObjToString(faultJob.FaultStrategy))

	})

	t.Run("processNetworkFault_right_node_separate_device_both has sub-health and separate", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   "Trigger0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SubHealthFaultStrategy,
		})
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   "Trigger0x0002",
			RelationFaults: []string{"0x0004"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := constant.FaultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0003",
		}
		fault4 := constant.FaultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0004",
		}
		nodeFault := constant.FaultInfo{
			NodeName:  "node-103",
			NPUName:   "Ascend-910-2",
			FaultType: constant.SwitchFault,
			FaultCode: "0x0004",
		}

		networkFaults = append(networkFaults, &fault1, &fault2, &nodeFault, &fault4)
		retryEventList := make([]constant.FaultInfo, 0)
		retryEvent := constant.FaultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "Trigger0x0001",
		}
		retryEvent2 := constant.FaultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "Trigger0x0002",
		}
		retryEventList = append(retryEventList, retryEvent, retryEvent2)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()
		t.Log(util.ObjToString(faultJob.FaultStrategy))

	})
}
