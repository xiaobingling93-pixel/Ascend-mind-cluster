/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package half910x4 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.
*/
package half910x4

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

// TestNew
func TestNew(t *testing.T) {
	t.Run("test New", func(t *testing.T) {
		npu := New(SchedulerName)
		if npu.GetPluginName() != SchedulerName {
			t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), SchedulerName)
		}
		if npu.GetAnnoName() != util.NPU910CardName {
			t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), util.NPU910CardName)
		}
		if npu.GetAnnoPreVal() != util.NPU910CardNamePre {
			t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
				npu.GetPluginName(), util.NPU910CardNamePre)
		}
	})
}

type fields struct {
	name string
	half910x4
	args    interface{}
	wantErr bool
}

func buildTestPreStartCases() []fields {
	var i interface{}
	return []fields{
		{
			name:      "01-it will return err when i is not *rescheduler",
			half910x4: half910x4{},
			args:      i,
			wantErr:   true,
		},
		{
			name:      "02-it will return err when i is not *rescheduler",
			half910x4: half910x4{},
			args:      &rescheduling.ReScheduler{},
			wantErr:   false,
		},
	}
}

func TestPreStartRescheduling(t *testing.T) {
	tests := buildTestPreStartCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.PreStartAction(tt.args, nil); (err != nil) != tt.wantErr {
				t.Errorf("preStartRescheduling() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
