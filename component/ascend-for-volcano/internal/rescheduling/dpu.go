/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package rescheduling is using for Huawei Ascend pin fault rescheduling
package rescheduling

import (
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// The return value indicates whether the task is faulted
func (reScheduler *ReScheduler) getTaskHealthStateByNodeDpu(fTask *FaultTask) bool {
	// fTask.UseCardName - The card used by the pod.
	// reScheduler.FaultNodes[fTask.NodeName].DpuUnhealthyNPU - unhealthy NPU because of the corresponding DPU is
	// unhealthy
	faultNode, ok := reScheduler.FaultNodes[fTask.NodeName]
	if !ok {
		klog.V(util.LogErrorLev).Infof("%s task<%s> get fault node<%s> failed, reScheduler.FaultNodes=%+v",
			util.DpuLogPrefix, fTask.TaskName, fTask.NodeName, reScheduler.FaultNodes)
		return false
	}
	if len(faultNode.DpuUnhealthyNPU) == 0 {
		return false
	}

	for _, useCard := range fTask.UseCardName {
		for _, unhealthyCard := range faultNode.DpuUnhealthyNPU {
			if useCard == unhealthyCard {
				return true
			}
		}
	}

	return false
}
