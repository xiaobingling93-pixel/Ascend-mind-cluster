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
	"errors"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
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

// ValidNPUJob check jobs' required NPU number and mode.
// ssn.AddJobValidFn -> JobValid -> Job.ValidJobFn -> ValidNPUJob
func (tp *module910a5SuperPod) ValidNPUJob() *api.ValidateResult {
	errStr := "check npu job failed"
	if tp == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("%s, err is %v", errStr, err)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: errStr,
		}
	}
	// register all check func in order
	checkers := []jobCheckerFunc{
		tp.checkSpBlock,
		tp.checkTpBlockNum,
		tp.calculateTpBlockAndCheck,
		tp.checkJobReqNpuNum,
	}
	for _, checker := range checkers {
		if err := checker(); err != nil {
			klog.V(util.LogErrorLev).Infof("%s %s", errStr, err.Message)
			return err
		}
	}

	return nil
}

// CheckNodeNPUByTask to check node NPU for each task
// ssn.AddPredicateFn -> NodePredicate -> CheckNodeNPUByTask -> filter node for score
func (tp *module910a5SuperPod) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	errStr := "check npu node by task failed"
	// valid argument
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("%s, err is %v", errStr, err)
		return err
	}
	checkers := []nodeCheckerFunc{
		tp.checkNodeStaticParams,
		tp.checkNodeNPUNums,
	}

	for _, checker := range checkers {
		if err := checker(task, node); err != nil {
			klog.V(util.LogErrorLev).Infof("%s %s", errStr, err.Error())
			return err
		}
	}
	return nil
}
