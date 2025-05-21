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

// Package slownode is a DT collection for func in slownode_conf_map
package slownode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
)

func TestCtxMap(t *testing.T) {
	// insert a context
	var jobName = "job1"
	var job = &slownode.SlowNodeJob{}
	job.JobName = jobName
	job.SlowNode = 1
	ctx := NewSlowNodeContext(job, enum.Node)
	slowNodeMap := GetSlowNodeCtxMap()
	slowNodeMap.Insert(jobName, ctx)

	// get the context
	value, ok := slowNodeMap.Get(jobName)
	assert.Equal(t, true, ok)
	assert.Equal(t, jobName, value.Job.JobName)
	assert.Equal(t, 1, value.Job.SlowNode)

	// delete the context
	slowNodeMap.Delete(jobName)
	_, ok = slowNodeMap.Get(jobName)
	assert.Equal(t, false, ok)

	// insert job
	slowNodeMap.Insert(jobName, ctx)

	// clear all contexts
	slowNodeMap.Clear()
	_, ok = slowNodeMap.Get(jobName)
	assert.Equal(t, false, ok)
}
