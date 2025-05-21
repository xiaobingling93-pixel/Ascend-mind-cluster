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
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/enum"
)

// ctxMap is a sync map that stores slow node feature function config.
type ctxMap struct {
	// {<jobId>: {*context.FaultDiagContext, *SlowNodeContext}}
	data sync.Map
}

// Insert inserts a key-value pair into the slowNodeConfMap.
func (s *ctxMap) Insert(jobId string, value *SlowNodeContext) {
	s.data.Store(jobId, value)
}

// Get retrieves the value associated with the key from the slowNodeConfMap.
func (s *ctxMap) Get(jobId string) (*SlowNodeContext, bool) {
	value, ok := s.data.Load(jobId)
	if !ok {
		return nil, false
	}
	ctx, ok := value.(*SlowNodeContext)
	if !ok {
		return nil, false
	}
	return ctx, true
}

// Delete removes the key-value pair from the slowNodeConfMap.
func (s *ctxMap) Delete(jobId string) {
	s.data.Delete(jobId)
}

// Clear clears all key-value pairs from the slowNodeConfMap.
func (s *ctxMap) Clear() {
	s.data.Range(func(jobId, ctx any) bool {
		s.data.Delete(jobId)
		return true
	})
}

// redo is a method that loop all the data in the map and redo the failed job per 10 minutes.
func (c *ctxMap) redo() {
	for {
		time.Sleep(redoInterval)
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]start to redo the failed job")
		c.data.Range(func(jobId, ctx any) bool {
			slowNodeCtx, ok := ctx.(*SlowNodeContext)
			if !ok {
				hwlog.RunLog.Errorf("[FD-OL SLOWNODE]redo: job(jobId=%s) is not a slow node context",
					jobId)
				return true
			}
			if slowNodeCtx.IsFailed() {
				go map[enum.DeployMode]func(*SlowNodeContext){
					enum.Cluster: redoClusterJob,
					enum.Node:    redoNodeJob,
				}[slowNodeCtx.Deployment](slowNodeCtx)
			}
			return true
		})
	}
}

func redoClusterJob(ctx *SlowNodeContext) {
	ctx.ResetFailed()
	switch ctx.Step() {
	// step 0 -> need redo all profiling
	case InitialStep:
		ctx.StartAllProfiling()
	// step 1-> start all profiling, but start slow node algo failed
	case ClusterStep1:
		ctx.StartSlowNodeAlgo()
	default:
		// Do not care start/stop heavy profiling, the data parse callback will trigger it forever.
		hwlog.RunLog.Warnf("[FD-OL SLOWNODE]job(name=%s, namespace=%s) current step: %d, passing",
			ctx.Job.JobName, ctx.Job.Namespace, ctx.Step())
	}
}

func redoNodeJob(ctx *SlowNodeContext) {
	/*process the failed step
	step 0 -> initial step, means the job started but not start the data parse
			so we need to restart the data parse
	step 1 -> job started data parse, job will not have this step when failed, ignore it
	step 2 -> job report data profiling result success, but job failed, means started slow node algo faield.
			so we need to start the slow node algo
	step 3 -> job started slow node algo success, job will not have this step when failed, ignore it
	*/
	ctx.ResetFailed()
	switch ctx.Step() {
	// step 0 -> start the data parse
	case InitialStep:
		ctx.StartDataParse()
	// step2 -> start slow node algo failed
	case NodeStep2:
		ctx.StartSlowNodeAlgo()
	default:
		hwlog.RunLog.Warnf("[FD-OL SLOWNODE]job(name=%s, namespace=%s) current step: %d, passing",
			ctx.Job.JobName, ctx.Job.Namespace, ctx.Step())
	}
}

var (
	slowNodeCtxMap     *ctxMap
	slowNodeCtxMapOnce sync.Once
)

// GetSlowNodeCtxMap returns a instance of ctxMap.
func GetSlowNodeCtxMap() *ctxMap {
	slowNodeCtxMapOnce.Do(func() {
		slowNodeCtxMap = &ctxMap{}
		go slowNodeCtxMap.redo()
	})
	return slowNodeCtxMap
}
