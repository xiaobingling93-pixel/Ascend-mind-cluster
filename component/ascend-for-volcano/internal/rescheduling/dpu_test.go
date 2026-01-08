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

// Package rescheduling is using for Huawei Ascend pin fault rescheduling
package rescheduling

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func TestGetTaskHealthStateByNodeDpu(t *testing.T) {
	tests := []getTaskHealthStateByNodeDpuTestCase{
		buildGetTaskHealthStateByNodeDpuTestCase1(), buildGetTaskHealthStateByNodeDpuTestCase2(),
		buildGetTaskHealthStateByNodeDpuTestCase3(), buildGetTaskHealthStateByNodeDpuTestCase4(),
		buildGetTaskHealthStateByNodeDpuTestCase6(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			faultNode := &FaultNode{
				NodeName:        tt.args.nodeName,
				DpuUnhealthyNPU: tt.args.DpuUnhealthyNPU,
			}
			reScheduler := &ReScheduler{
				DealReSchedulerCache: &DealReSchedulerCache{
					FaultNodes: map[string]*FaultNode{
						tt.args.nodeName: faultNode,
					},
				},
			}
			gotFault := reScheduler.getTaskHealthStateByNodeDpu(tt.args.fTask)
			if gotFault != tt.wantFault {
				t.Errorf("getTaskHealthStateByNodeDpu() = (%v), want (%v)", gotFault, tt.wantFault)
			}
		})
	}
}

type getTaskHealthStateByNodeDpuTestCase struct {
	name string
	args struct {
		fTask           *FaultTask
		nodeName        string
		busType         string
		npuToDpusMap    map[string][]string
		DpuUnhealthyNPU []string
	}
	wantFault bool
}

func buildGetTaskHealthStateByNodeDpuTestCase1() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "UBType - as long as one DPU status is up, the result should be healthy",
		args: struct {
			fTask           *FaultTask
			nodeName        string
			busType         string
			npuToDpusMap    map[string][]string
			DpuUnhealthyNPU []string
		}{
			fTask: &FaultTask{
				TaskName: "task1",
				NodeName: "node1",
				UseCardName: []string{
					"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4", "Ascend910-5",
					"Ascend910-6", "Ascend910-7",
				},
			},
			nodeName: "node1",
			busType:  util.UbType,
			npuToDpusMap: map[string][]string{
				"0": {"enps0", "enps2"},
				"1": {"enps0", "enps2"},
				"2": {"enps0", "enps2"},
				"3": {"enps0", "enps2"},
				"4": {"enps1", "enps3"},
				"5": {"enps1", "enps3"},
				"6": {"enps1", "enps3"},
				"7": {"enps1", "enps3"},
			},
			DpuUnhealthyNPU: []string{},
		},
		wantFault: false,
	}
}

func buildGetTaskHealthStateByNodeDpuTestCase2() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "UBType - both DPUs faulty",
		args: struct {
			fTask           *FaultTask
			nodeName        string
			busType         string
			npuToDpusMap    map[string][]string
			DpuUnhealthyNPU []string
		}{
			fTask: &FaultTask{
				TaskName:    "task1",
				NodeName:    "node1",
				UseCardName: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
			},
			nodeName: "node1",
			busType:  util.UbType,
			npuToDpusMap: map[string][]string{
				"0": {"enps0", "enps2"},
				"1": {"enps0", "enps2"},
				"2": {"enps0", "enps2"},
				"3": {"enps0", "enps2"},
			},
			DpuUnhealthyNPU: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
		},
		wantFault: true,
	}
}

func buildGetTaskHealthStateByNodeDpuTestCase3() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "PcieType - DPU healthy",
		args: struct {
			fTask           *FaultTask
			nodeName        string
			busType         string
			npuToDpusMap    map[string][]string
			DpuUnhealthyNPU []string
		}{
			fTask: &FaultTask{
				TaskName:    "task1",
				NodeName:    "node1",
				UseCardName: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
			},
			nodeName: "node1",
			busType:  util.PcieType,
			npuToDpusMap: map[string][]string{
				"0": {"eth1"},
				"1": {"eth2"},
				"2": {"eth3"},
				"3": {"eth4"},
			},
			DpuUnhealthyNPU: []string{},
		},
		wantFault: false,
	}
}

func buildGetTaskHealthStateByNodeDpuTestCase4() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "PcieType - DPU faulty",
		args: struct {
			fTask           *FaultTask
			nodeName        string
			busType         string
			npuToDpusMap    map[string][]string
			DpuUnhealthyNPU []string
		}{
			fTask: &FaultTask{
				TaskName:    "task1",
				NodeName:    "node1",
				UseCardName: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
			},
			nodeName: "node1",
			busType:  util.PcieType,
			npuToDpusMap: map[string][]string{
				"0": {"eth1"},
				"1": {"eth2"},
				"2": {"eth3"},
				"3": {"eth4"},
			},
			DpuUnhealthyNPU: []string{"Ascend910-0"},
		},
		wantFault: true,
	}
}

func buildGetTaskHealthStateByNodeDpuTestCase6() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "No DPU info, should be healthy",
		args: struct {
			fTask           *FaultTask
			nodeName        string
			busType         string
			npuToDpusMap    map[string][]string
			DpuUnhealthyNPU []string
		}{
			fTask: &FaultTask{
				TaskName:    "task1",
				NodeName:    "node1",
				UseCardName: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
			},
			nodeName:        "node1",
			busType:         util.PcieType,
			npuToDpusMap:    map[string][]string{},
			DpuUnhealthyNPU: []string{},
		},
		wantFault: false,
	}
}
