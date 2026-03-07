/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package manualfault is used to process manually separate faults

package manualfault

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
)

// ManualFaultProcessor is used to process manually separate faults
var ManualFaultProcessor *manualFaultProcessor

type manualFaultProcessor struct{}

func init() {
	ManualFaultProcessor = &manualFaultProcessor{}
}

// Process is used to process manually separate faults
func (p *manualFaultProcessor) Process(info any) any {
	if !conf.GetManualEnabled() {
		return info
	}

	processContent, ok := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	if !ok {
		hwlog.RunLog.Error("input is not deviceinfo type")
		return info
	}
	devInfos := processContent.AllConfigmap
	p.loadManualSepToClusterInfoCm(devInfos)
	return info
}

func (p *manualFaultProcessor) loadManualSepToClusterInfoCm(devInfos map[string]*constant.AdvanceDeviceFaultCm) {
	nodeInfo, err := manualfault.FaultCmInfo.DeepCopy()
	if err != nil {
		hwlog.RunLog.Errorf("deep copy fault cm info failed, error: %v", err)
		return
	}

	if len(nodeInfo) == 0 {
		return
	}
	for nodeName, devFaults := range nodeInfo {
		for _, devName := range devFaults.Total {
			hwlog.RunLog.Debugf("node: %s, dev: %s has already been manually separated", nodeName, devName)
			devCmInfo, ok := devInfos[nodeName]
			if !ok {
				continue
			}
			fault := constant.DeviceFault{
				FaultType:            constant.CardUnhealthy,
				NPUName:              devName,
				LargeModelFaultLevel: constant.ManuallySeparateNPU,
				FaultLevel:           constant.ManuallySeparateNPU,
				FaultHandling:        constant.ManuallySeparateNPU,
			}
			devCmInfo.AddFaultAndFix(fault)
		}
	}
	hwlog.RunLog.Debug("load manual fault cache to cluster info cm success")
}
