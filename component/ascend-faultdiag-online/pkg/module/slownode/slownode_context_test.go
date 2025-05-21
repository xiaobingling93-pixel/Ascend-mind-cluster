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

// Package slownode is a DT collection for func in slownode_context
package slownode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
)

func TestSlowNodeContext(t *testing.T) {
	var job = &slownode.SlowNodeJob{
		SlowNodeInputBase: slownode.SlowNodeInputBase{
			JobName: "job1",
		},
		SlowNode: 1,
	}
	ctx := NewSlowNodeContext(job, enum.Node)

	// test job start
	assert.Equal(t, false, ctx.IsRunning())
	ctx.Start()
	assert.Equal(t, true, ctx.IsRunning())

	// test job stop
	ctx.Stop()
	assert.Equal(t, false, ctx.IsRunning())

	// test update
	assert.Equal(t, 1, ctx.Job.SlowNode)
	ctx.Update(&slownode.SlowNodeJob{SlowNode: 0})
	assert.Equal(t, 0, ctx.Job.SlowNode)

}
