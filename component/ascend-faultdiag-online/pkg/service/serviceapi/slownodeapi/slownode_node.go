/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package slownodeapi provides some func to process the data in node
package slownodeapi

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/global"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	sm "ascend-faultdiag-online/pkg/module/slownode"
	"ascend-faultdiag-online/pkg/utils"
)

func nodeProcessSlowNodeAlgoCallback(message string) {
	var data = map[string]map[string]any{}
	if err := json.Unmarshal([]byte(message), &data); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json unmarshal data: %s failed: %v", message, err)
		return
	}
	var nodeResult = slownode.NodeSlowNodeAlgoResult{}
	if err := convertMaptoStruct(data, &nodeResult); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]convert map data: %s to struct failed: %v", data, err)
		return
	}

	slowNodeCtx, ok := sm.GetSlowNodeCtxMap().Get(nodeResult.JobId)
	if !ok || !slowNodeCtx.IsRunning() {
		hwlog.RunLog.Warnf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) is not exit or not running, exit slow node algo callback process.",
			nodeResult.JobName, nodeResult.JobId)
		return
	}
	if nodeResult.IsSlow == sm.IsDegradation {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) detected node: %v degradation",
			slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId, nodeResult.NodeRank)
	}
	if err := writeAlgoResult(&nodeResult, slowNodeCtx); err != nil {
		hwlog.RunLog.Errorf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) created node slow node algo result configmap failed: %v",
			slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId, err)
	}
}

func nodeProcessDataParseCallback(message string) {
	var dataParseResult slownode.DataParseResult
	if err := json.Unmarshal([]byte(message), &dataParseResult); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json marshal data parse callback data: %v failed: %s", message, err)
		return
	}
	slowNodeCtx, ok := sm.GetSlowNodeCtxMap().Get(dataParseResult.JobId)
	if !ok || !slowNodeCtx.IsRunning() {
		hwlog.RunLog.Errorf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) is not exist or not running, exit data parse callback process",
			dataParseResult.JobName, dataParseResult.JobId)
		return
	}

	if dataParseResult.IsFinished && dataParseResult.StepCount >= sm.MinStepCount &&
		slowNodeCtx.Step() == sm.NodeStep1 {
		if err := nodeReportProfiling(slowNodeCtx); err != nil {
			return
		}
		// step from 1 to 2
		slowNodeCtx.AddStep()
		slowNodeCtx.StartSlowNodeAlgo()
	}
}

// nodeReportProfiling report the data profiling result to cm
func nodeReportProfiling(slowNodeCtx *sm.SlowNodeContext) error {
	parallelGroupInfo, err := readParallelGroupInfo(slowNodeCtx.Job.JobId)
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]read parallel group info failed: %v", err)
		return err
	}
	// get node ip
	nodeIP, err := utils.GetNodeIP()
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]get node ip failed: %s", err)
		return err
	}
	// finish profiling
	var profilingResult = slownode.NodeDataProfilingResult{
		JobName:                  slowNodeCtx.Job.JobName,
		JobId:                    slowNodeCtx.Job.JobId,
		FinishedInitialProfiling: true,
		FinishedTime:             time.Now().Unix(),
		NodeIP:                   nodeIP,
		ParallelGroupInfo:        parallelGroupInfo,
	}
	dataBytes, err := json.MarshalIndent(profilingResult, "", "  ")
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process node data parse: error serializing to JSON: %v", err)
		return err
	}

	cmName := fmt.Sprintf("%s-%s-%s", sm.NodeDataProfilingResultPrefix, slowNodeCtx.Job.JobId, nodeIP)
	slowNodeCtx.AllCMNAMEs.Store(cmName, struct{}{})
	if err = createOrUpdateCM(
		cmName, slowNodeCtx.Job.Namespace, sm.NodeDataProfilingResultCMKey, dataBytes); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]create node data profiling result configmap failed, err is %v", err)
		return err
	}
	return nil
}

func writeAlgoResult(nodeResult *slownode.NodeSlowNodeAlgoResult, slowNodeCtx *sm.SlowNodeContext) error {
	cmName := fmt.Sprintf("%s-%s-%s", sm.NodeSlowNodeAlgoResultPrefix, nodeResult.JobId, nodeResult.NodeRank)
	slowNodeCtx.AllCMNAMEs.Store(cmName, struct{}{})
	dataBytes, err := json.MarshalIndent(nodeResult, "", "  ")
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json marshal the slow node algo result failed: %v", err)
		return err
	}
	return createOrUpdateCM(cmName, slowNodeCtx.Job.Namespace, sm.NodeSlowNodeAlgoResultCMKey, dataBytes)
}

func createOrUpdateCM(name, namespace, cmKey string, data []byte) error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{sm.CmConsumer: sm.CmConsumerValue},
		},
		Data: map[string]string{
			cmKey: string(data),
		},
	}
	if err := global.K8sClient.CreateOrUpdateConfigMap(cm); err != nil {
		return err
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]create or update configmap: [key: %s, value: %s] successfully",
		cm.Name, cm.Data)
	return nil
}

func readParallelGroupInfo(jobName string) (map[string]any, error) {
	file, err := os.ReadFile(fmt.Sprintf("%s/%s/%s", sm.NodeFilePath, jobName, sm.ParallelGroupInfo))
	if err != nil {
		return nil, err
	}
	var data = map[string]any{}
	err = json.Unmarshal(file, &data)
	return data, err
}
