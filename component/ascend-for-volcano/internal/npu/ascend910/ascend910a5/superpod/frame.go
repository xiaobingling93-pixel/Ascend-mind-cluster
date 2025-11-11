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

// Package superpod for a5 schedule handler
package superpod

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// New return npu plugin
func New(name string) *module910a5SuperPod {
	m := &module910a5SuperPod{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(nodeNPUNum)
	m.scheduleStrategy = SuperPodSchedule
	m.netUnhealthyKey = networkUnhealthyNPU
	m.faultNPUKey = faultNPU
	m.isNeedAlgoAlign = false
	return m
}
