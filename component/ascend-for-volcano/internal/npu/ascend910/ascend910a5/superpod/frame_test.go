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

// Package superpod for base function ut
package superpod

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	testSchedulerName = "test"
)

// NewPluginTestCase use for New func test
type NewPluginTestCase struct {
	Name    string
	WantErr error

	ScheduleName  string
	PluginName    string
	AnnoName      string
	AnnoPreValue  string
	MaxNodeNPUNum int
}

func buildNewTestCase() []NewPluginTestCase {
	return []NewPluginTestCase{
		{
			Name:          "01-NewTest should return nil when Schedule Name is huawei.com/Ascend910900SuperPod-A5-8",
			ScheduleName:  util.SuperPodx8SchedulerName,
			PluginName:    util.SuperPodx8SchedulerName,
			AnnoName:      util.NPU910CardName,
			AnnoPreValue:  util.NPU910CardNamePre,
			MaxNodeNPUNum: npuNumber8,
			WantErr:       nil,
		},
		{
			Name:          "02-NewTest should return nil when Schedule Name is test",
			ScheduleName:  testSchedulerName,
			PluginName:    testSchedulerName,
			AnnoName:      util.NPU910CardName,
			AnnoPreValue:  util.NPU910CardNamePre,
			MaxNodeNPUNum: npuNumber8,
			WantErr:       nil,
		},
	}
}

func TestNew(t *testing.T) {
	testCases := buildNewTestCase()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			npu := New(tt.ScheduleName)
			if npu.GetPluginName() != tt.PluginName {
				t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), tt.PluginName)
			}
			if npu.GetAnnoName() != tt.AnnoName {
				t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), tt.AnnoName)
			}
			if npu.GetAnnoPreVal() != tt.AnnoPreValue {
				t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
					npu.GetPluginName(), tt.AnnoPreValue)
			}
			if npu.MaxNodeNPUNum != tt.MaxNodeNPUNum {
				t.Errorf("New() npu MaxNodeNPUNum: %d, wantMaxNodeNPUNum: %d.",
					npu.MaxNodeNPUNum, tt.MaxNodeNPUNum)
			}
		})
	}
}
