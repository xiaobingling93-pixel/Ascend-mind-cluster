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

// Package slownode provides a sync map for storing slow node and the context
package slownode

import (
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/global"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
)

type algoControl struct {
	// StartSlowNodeAlgoSign is channel for start slow node algo
	StartSlowNodeAlgoSign chan struct{}
	// StopSlowNodeAlgoSign is channel for stop slow node algo
	StopSlowNodeAlgoSign chan struct{}
}

// dataParseControl are some parameters for node fd-ol
type dataParseControl struct {
	// StartDataParseSign is channel for start data parse
	StartDataParseSign chan struct{}
	// StopDataParseSign is channel for stop data parse
	StopDataParseSign chan struct{}
}

type cluster struct {
	isStartedHeavyProfiling bool
	WorkerNodesIP           []string
	SlowNodeAlgoRes         []*slownode.ClusterSlowNodeAlgoResult
}

// SlowNodeContext is a mixed struct with fdctx and SlowNodeContext
type SlowNodeContext struct {
	// mutex
	mu sync.Mutex

	// Job is the object data from configmap
	Job       *slownode.SlowNodeJob
	step      SlowNodeStep
	isRunning bool
	isFailed  bool

	// Deployment is the deployment type, cluster or node
	Deployment enum.DeployMode
	// StopChan is a stop signal for all goroutine
	StopChan chan struct{}
	// AllCMNAMEs is a sync map including all the cm names have been created
	AllCMNAMEs sync.Map

	algoControl
	dataParseControl
	cluster
}

// NewSlowNodeContext returns a new SlowNodeContext object
func NewSlowNodeContext(
	job *slownode.SlowNodeJob,
	deployment enum.DeployMode,
) *SlowNodeContext {
	if job == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]create slow node context failed: job is nil.")
		return nil
	}
	if deployment == enum.Cluster && global.GrpcClient == nil {
		hwlog.RunLog.Errorf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) in cluster created context failed, grpcClient is nil",
			job.JobName, job.JobId)
		return nil
	}
	ctx := &SlowNodeContext{
		Job:        job,
		Deployment: deployment,
		isRunning:  false,
		isFailed:   false,
		step:       InitialStep,
		algoControl: algoControl{
			StartSlowNodeAlgoSign: make(chan struct{}, channelCapacity),
			StopSlowNodeAlgoSign:  make(chan struct{}, channelCapacity),
		},
		dataParseControl: dataParseControl{
			StartDataParseSign: make(chan struct{}, channelCapacity),
			StopDataParseSign:  make(chan struct{}, channelCapacity),
		},
		cluster: cluster{
			SlowNodeAlgoRes: make([]*slownode.ClusterSlowNodeAlgoResult, 0),
			WorkerNodesIP:   make([]string, 0),
		},
		StopChan: make(chan struct{}),
	}
	return ctx
}

// InsertNodesIP is a function support insert nodes IP to context
func (ctx *SlowNodeContext) InsertNodesIP(nodesIP []string) {
	ctx.WorkerNodesIP = nodesIP
}

// Start the job
func (ctx *SlowNodeContext) Start() {
	ctx.isRunning = true
	ctx.InitialStep()
}

// Failed the job
func (ctx *SlowNodeContext) Failed() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.isFailed = true
}

// ResetFailed reset the job failed status
func (ctx *SlowNodeContext) ResetFailed() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.isFailed = false
}

// Stop the job
func (ctx *SlowNodeContext) Stop() {
	close(ctx.StopChan)
	ctx.isRunning = false
}

// IsRunning check if the job is running
func (ctx *SlowNodeContext) IsRunning() bool {
	return ctx.isRunning
}

// IsFailed check if the job is failed
func (ctx *SlowNodeContext) IsFailed() bool {
	return ctx.isFailed
}

// AddStep add the step
func (ctx *SlowNodeContext) AddStep() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.step++
}

// InitialStep set the step to initial
func (ctx *SlowNodeContext) InitialStep() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.step = InitialStep
}

// Step get the step
func (ctx *SlowNodeContext) Step() SlowNodeStep {
	return ctx.step
}

// IsStartedHeavyProfiling returns a bool whether the heavy profiling starts or not
func (ctx *SlowNodeContext) IsStartedHeavyProfiling() bool {
	return ctx.isStartedHeavyProfiling
}

// StartAllProfiling start all the profiling
func (ctx *SlowNodeContext) StartAllProfiling() {
	if err := global.GrpcClient.StartAllProfiling(ctx.Job.JobName, ctx.Job.Namespace); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) started all profiling failed: %s",
			ctx.Job.JobName, ctx.Job.JobId, err.Error())
		ctx.Failed()
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) started all profiling successfully",
		ctx.Job.JobName, ctx.Job.JobId)
	// step from 0 to 1
	ctx.AddStep()
}

// StopAllProfiling stop all the profiling
func (ctx *SlowNodeContext) StopAllProfiling() {
	if err := global.GrpcClient.StopAllProfiling(ctx.Job.JobName, ctx.Job.Namespace); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) stopped all profiling failed: %s",
			ctx.Job.JobName, ctx.Job.JobId, err.Error())
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) stopped all profiling successfully",
		ctx.Job.JobName, ctx.Job.JobId)
}

// StartHeavyProfiling start the heavy profiling
func (ctx *SlowNodeContext) StartHeavyProfiling() {
	if err := global.GrpcClient.StartHeavyProfiling(ctx.Job.JobName, ctx.Job.Namespace); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) started heavy profiling failed: %s",
			ctx.Job.JobName, ctx.Job.JobId, err.Error())
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) started heavy profiling successfully",
		ctx.Job.JobName, ctx.Job.JobId)
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.isStartedHeavyProfiling = true
	ctx.SlowNodeAlgoRes = make([]*slownode.ClusterSlowNodeAlgoResult, 0)
}

// StopHeavyProfiling stop the heavy profiling
func (ctx *SlowNodeContext) StopHeavyProfiling() {
	if err := global.GrpcClient.StopHeavyProfiling(ctx.Job.JobName, ctx.Job.Namespace); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) stopped heavy profiling failed: %s",
			ctx.Job.JobName, ctx.Job.JobId, err.Error())
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) stopped heavy profiling successfully",
		ctx.Job.JobName, ctx.Job.JobId)
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.isStartedHeavyProfiling = false
}

// Update the slow node to the current job
func (ctx *SlowNodeContext) Update(job *slownode.SlowNodeJob) {
	ctx.Job.SlowNode = job.SlowNode
}

// StartSlowNodeAlgo start the slow node algo
func (ctx *SlowNodeContext) StartSlowNodeAlgo() {
	ctx.StartSlowNodeAlgoSign <- struct{}{}
}

// StopSlowNodeAlgo stop the slow node algo
func (ctx *SlowNodeContext) StopSlowNodeAlgo() {
	ctx.StopSlowNodeAlgoSign <- struct{}{}
}

// StartDataParse start the data parse
func (ctx *SlowNodeContext) StartDataParse() {
	ctx.StartDataParseSign <- struct{}{}
}

// StopDataParse stop the data parse
func (ctx *SlowNodeContext) StopDataParse() {
	ctx.StopDataParseSign <- struct{}{}
}

// AddRecords add the slow node algo result in context
func (ctx *SlowNodeContext) AddRecords(result *slownode.ClusterSlowNodeAlgoResult) {
	if len(ctx.SlowNodeAlgoRes) > recordsCapacity {
		start := len(ctx.SlowNodeAlgoRes) - recordsCapacity + 1
		ctx.SlowNodeAlgoRes = ctx.SlowNodeAlgoRes[start:]
	}
	ctx.SlowNodeAlgoRes = append(ctx.SlowNodeAlgoRes, result)
}
