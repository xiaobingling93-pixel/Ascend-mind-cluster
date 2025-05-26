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

// Package slownodeapi provides some func to process the data in cluster
package slownodeapi

import (
	"encoding/json"
	"os"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/global"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	sm "ascend-faultdiag-online/pkg/module/slownode"
	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/grpc"
	"ascend-faultdiag-online/pkg/utils/grpc/pubfault"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

// clusterProcessSlowNodeAlgoCallback process the callback data from slow node algo deployed in cluster
func clusterProcessSlowNodeAlgoCallback(message string) {
	var data = map[string]map[string]any{}
	if err := json.Unmarshal([]byte(message), &data); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json unmarshal data: %s failed: %v", message, err)
		return
	}
	var result = slownode.ClusterSlowNodeAlgoResult{}
	if err := convertMaptoStruct(data, &result); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]convert map data: %s to struct failed: %v", data, err)
		return
	}
	slowNodeCtx, ok := sm.GetSlowNodeCtxMap().Get(result.JobId)
	if !ok || !slowNodeCtx.IsRunning() {
		hwlog.RunLog.Warnf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) is not exit or not running, exit slow node algo callback process.",
			result.JobName, result.JobId)
		return
	}
	slowNodeCtx.AddRecords(&result)
	clusterProcessProfiling(slowNodeCtx, &result)
}

func clusterProcessParallelGroupInfoCallback(message string) {
	var result = slownode.DataParseResult{}
	if err := json.Unmarshal([]byte(message), &result); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]error parsing dataParseResult JSON: %v", err)
		return
	}
	slowNodeCtx, ok := sm.GetSlowNodeCtxMap().Get(result.JobId)
	if !ok || !slowNodeCtx.IsRunning() {
		hwlog.RunLog.Warnf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) is not exit or not running, exit slow node algo callback process.",
			result.JobName, result.JobId)
		return
	}
	// ClusterStep2 means started slow node algo
	if result.IsFinished && slowNodeCtx.Step() == sm.ClusterStep1 {
		slowNodeCtx.StartSlowNodeAlgo()
	}
}

func clusterProcessProfiling(slowNodeCtx *sm.SlowNodeContext, result *slownode.ClusterSlowNodeAlgoResult) {
	// if the node is not slow, and the heavy profiling is started, stop it
	if result.IsSlow != sm.IsDegradation && slowNodeCtx.IsStartedHeavyProfiling() {
		hwlog.RunLog.Infof(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) detected node is not degradation, stop heavy profiling",
			slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId)
		reportSlowNode(slowNodeCtx, result)
		slowNodeCtx.StopHeavyProfiling()
		return
	}
	// if the lastest 5 records of node are all slow, and the heavy profiling started, stop it
	if result.IsSlow == sm.IsDegradation {
		// if the node is slow, and the heavy profiling is not started, start it
		if !slowNodeCtx.IsStartedHeavyProfiling() {
			hwlog.RunLog.Infof(
				"[FD-OL SLOWNODE]job(name=%s, jobId=%s) detected node is degradation, start heavy profiling",
				slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId)
			slowNodeCtx.StartHeavyProfiling()
			slowNodeCtx.AddRecords(result)
			return
		}
		degradation := slicetool.Filter(slowNodeCtx.SlowNodeAlgoRes,
			func(record *slownode.ClusterSlowNodeAlgoResult) bool {
				return record.IsSlow == sm.IsDegradation
			})
		if len(degradation) >= maxDegradationCount {
			hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s in cluster got degradation",
				slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId)
			reportSlowNode(slowNodeCtx, result)
			slowNodeCtx.StopHeavyProfiling()
		}
	}
}

func reportSlowNode(slowNodeCtx *sm.SlowNodeContext, result *slownode.ClusterSlowNodeAlgoResult) {
	if global.GrpcClient == nil {
		client, err := grpc.GetClient(utils.GetClusterIP())
		if err != nil {
			hwlog.RunLog.Errorf("[FD-OL SLOWNODE]get grpc client failed: %s", err)
			return
		}
		global.GrpcClient = client
	}
	var faultCode = slowNodeRecoveryFaultCode
	if result.IsSlow == sm.IsDegradation {
		faultCode = slowNodeFaultCode
	}
	hostname, err := os.Hostname()
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) got hostname failed: %s",
			slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId, err)
	}
	var deviceIds = make([]int32, len(result.SlowCalculateRanks))
	for i, v := range result.SlowCalculateRanks {
		deviceIds[i] = int32(v)
	}
	var fault = []*pubfault.Fault{
		{
			FaultId:       slowNodeCtx.Job.JobId,
			FaultType:     "Node",
			FaultCode:     faultCode,
			FaultTime:     time.Now().Unix(),
			Assertion:     "once",
			FaultLocation: map[string]string{},
			Influence: []*pubfault.PubFaultInfo{
				{
					NodeName:  hostname,
					DeviceIds: deviceIds,
				},
			},
			Description: "",
		}}
	if err := global.GrpcClient.ReportFault(fault); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) report slow node detection failed: %s",
			slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId, err)
	} else {
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) report slow node detection: %+v successfully.",
			slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId, fault)
	}
}
