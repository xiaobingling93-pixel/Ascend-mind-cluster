/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type FaultJobTestField struct {
	ReScheduleKey       string
	IsFaultJob          bool
	IsInSession         bool
	JobName             string
	JobUID              api.JobID
	JobNamespace        string
	JobRankIds          []string
	NodeNames           []string
	FaultTasks          []FaultTask
	UpdateTime          int64
	JobRankIdCreateTime int64
	FaultTypes          []string
	DeleteExecutedFlag  bool
	ElasticScheduling   string
}

type FaultJobForceDeleteJobArgs struct {
	ssn             *framework.Session
	schedulerJob    *plugin.SchedulerJob
	cacheFuncBefore func()
	cacheFuncAfter  func()
}

type FaultJobForceDeleteJobTests struct {
	name    string
	fields  FaultJobTestField
	args    FaultJobForceDeleteJobArgs
	wantErr bool
}

func buildFaultJobForceDeleteJobTests() []FaultJobForceDeleteJobTests {
	var tmpPatch *gomonkey.Patches = nil
	faultTask1 := fakeReSchedulerFaultTask(true, []string{"pod0", "vcjob", "node0", "job0", "0"}, 0,
		"ppppppppppppp")
	faultTask2 := fakeReSchedulerFaultTask(false, []string{"pod1", "vcjob", "node1", "job0", "1"}, 0,
		"ppppppppppppp")
	schedulerJob := fakeSchedulerJobEmptyTask("job0", "vcjob")
	test1 := FaultJobForceDeleteJobTests{
		name: "01-FaultJobForceDeleteJob()-delete success",
		fields: FaultJobTestField{
			JobName:             "job0",
			JobUID:              "vcjob/job0",
			JobNamespace:        "vcjob",
			JobRankIds:          nil,
			NodeNames:           nil,
			FaultTasks:          []FaultTask{faultTask1, faultTask2},
			UpdateTime:          0,
			JobRankIdCreateTime: 0,
			FaultTypes:          nil,
			DeleteExecutedFlag:  false,
		},
		args: FaultJobForceDeleteJobArgs{
			ssn:          test.FakeSSNReSchedule(),
			schedulerJob: &schedulerJob,
			cacheFuncBefore: func() {
				tmpPatch = gomonkey.ApplyMethod(reflect.TypeOf(&FaultTask{}), "DeleteRealPodByTask",
					func(_ *FaultTask, _ *framework.Session, _ int64) error { return nil })
			},
			cacheFuncAfter: func() {
				if tmpPatch != nil {
					tmpPatch.Reset()
				}
			},
		},
		wantErr: false,
	}
	tests := []FaultJobForceDeleteJobTests{
		test1,
	}
	return tests
}

// TestFaultJobForceDeleteJob test for force delete function
func TestFaultJobForceDeleteJob(t *testing.T) {
	env := plugin.ScheduleEnv{
		SuperPodInfo: &plugin.SuperPodInfo{
			SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
			SuperPodFaultTaskNodes:    map[api.JobID][]string{},
			SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}},
	}
	tests := buildFaultJobForceDeleteJobTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFuncBefore()
			fJob := &FaultJob{
				ReScheduleKey:       tt.fields.ReScheduleKey,
				IsFaultJob:          tt.fields.IsFaultJob,
				IsInSession:         tt.fields.IsInSession,
				JobName:             tt.fields.JobName,
				JobUID:              tt.fields.JobUID,
				JobNamespace:        tt.fields.JobNamespace,
				JobRankIds:          tt.fields.JobRankIds,
				NodeNames:           tt.fields.NodeNames,
				FaultTasks:          tt.fields.FaultTasks,
				UpdateTime:          tt.fields.UpdateTime,
				JobRankIdCreateTime: tt.fields.JobRankIdCreateTime,
				FaultTypes:          tt.fields.FaultTypes,
				DeleteExecutedFlag:  tt.fields.DeleteExecutedFlag,
				ElasticScheduling:   tt.fields.ElasticScheduling,
			}
			if err := fJob.ForceDeleteJob(tt.args.ssn, tt.args.schedulerJob, env); (err != nil) != tt.wantErr {
				t.Errorf("ForceDeleteJob() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFuncAfter()
		})
	}
}

// TestGetVirSupPodId test getVirSupPodId
func TestGetVirSupPodId(t *testing.T) {
	env := plugin.ScheduleEnv{
		SuperPodInfo: &plugin.SuperPodInfo{
			SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
			SuperPodFaultTaskNodes:    map[api.JobID][]string{},
			SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}},
	}
	fJob := &FaultJob{
		JobUID: "test",
	}
	t.Run("getVirSupPodId", func(t *testing.T) {
		if result := fJob.getVirSupPodId("node", env); result != "" {
			t.Errorf("return empty when SuperPodReschdInfo "+
				"doesn't have job test result = %v, want %v", result, "")
		}
		env.SuperPodInfo.SuperPodReschdInfo["test"] = map[string][]plugin.SuperNode{
			"0": {
				plugin.SuperNode{
					SuperPodID: 1,
					Name:       "node1",
				},
			},
		}
		if result := fJob.getVirSupPodId("node", env); result != "" {
			t.Errorf("return empty when name != node1 result = %v, want %v", result, "")
		}
		if result := fJob.getVirSupPodId("node1", env); result != "0" {
			t.Errorf("return 0 when name == node1 result = %v, want %v", result, "0")
		}
	})
}

// TestIsContainTask test isContainTask
func TestIsContainTask(t *testing.T) {
	env := plugin.ScheduleEnv{
		SuperPodInfo: &plugin.SuperPodInfo{
			SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
			SuperPodFaultTaskNodes:    map[api.JobID][]string{},
			SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}},
	}
	fJob := &FaultJob{
		JobUID: "test",
	}
	t.Run("isContainTask", func(t *testing.T) {
		if result := fJob.isContainTask([]string{"0"}, "node", env); result != false {
			t.Errorf("return false when SuperPodReschdInfo doesn't have job test "+
				"result = %v, want %v", result, false)
		}
		env.SuperPodInfo.SuperPodReschdInfo["test"] = map[string][]plugin.SuperNode{
			"0": {
				plugin.SuperNode{
					SuperPodID: 1,
					Name:       "node1",
				},
			},
		}
		if result := fJob.isContainTask([]string{"0"}, "node", env); result != false {
			t.Errorf("return false when name != node1 result = %v, want %v", result, false)
		}
		if result := fJob.isContainTask([]string{"0"}, "node1", env); result != true {
			t.Errorf("return true when name == node1 result = %v, want %v", result, true)
		}
	})
}

// TestGetIds test getIds
func TestGetIds(t *testing.T) {
	env := plugin.ScheduleEnv{
		SuperPodInfo: &plugin.SuperPodInfo{
			SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
			SuperPodFaultTaskNodes:    map[api.JobID][]string{},
			SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}},
	}
	fJob := &FaultJob{
		JobUID: "test",
	}
	t.Run("getIds", func(t *testing.T) {
		if result := fJob.getIds(env); result != nil {
			t.Errorf("return [] when SuperPodFaultTaskNodes doesn't have job test "+
				"result = %v, want %v", result, nil)
		}
		env.SuperPodInfo.SuperPodFaultTaskNodes["test"] = []string{"node"}
		env.SuperPodInfo.SuperPodReschdInfo["test"] = map[string][]plugin.SuperNode{
			"0": {
				plugin.SuperNode{
					SuperPodID: 1,
					Name:       "node1",
				},
			},
		}
		if result := fJob.getIds(env); result != nil {
			t.Errorf("return [] when name != node1 result = %v, want %v", result, nil)
		}
		env.SuperPodInfo.SuperPodFaultTaskNodes["test"] = []string{"node1"}
		if result := fJob.getIds(env); !reflect.DeepEqual(result, []string{"0"}) {
			t.Errorf("return [0] when name == node1 result = %v, want %v", result, true)
		}
	})
}

func TestGetJobFaultRescheduleLabel(t *testing.T) {
	t.Run("01-GetJobFaultRescheduleLabel return error when job is nil", func(t *testing.T) {
		fJob := &FaultJob{JobUID: "test"}
		res := fJob.GetJobFaultRescheduleLabel(nil)
		if res != JobOffRescheduleLabelValue {
			t.Errorf("GetJobFaultRescheduleLabel() res = %v, wantRes is %v", res, JobOffRescheduleLabelValue)
		}
	})
}

type getJobElasticSchedulingLabelTestCase struct {
	name    string
	fJob    *FaultJob
	job     *plugin.SchedulerJob
	wantRes string
}

func buildGetJobElasticSchedulingLabelTestCases() []getJobElasticSchedulingLabelTestCase {
	return []getJobElasticSchedulingLabelTestCase{
		{
			name:    "01-GetJobElasticSchedulingLabel return off when job is nil",
			fJob:    &FaultJob{JobUID: "test"},
			job:     nil,
			wantRes: JobOffRescheduleLabelValue,
		},
		{
			name: "02-GetJobElasticSchedulingLabel return on when ElasticSchedulingKey exist",
			fJob: &FaultJob{JobUID: "test"},
			job: &plugin.SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{
					ElasticSchedulingKey: JobOnElasticScheduling,
				}}}},
			wantRes: JobOnElasticScheduling,
		},
	}
}

func TestGetJobElasticSchedulingLabel(t *testing.T) {
	testCases := buildGetJobElasticSchedulingLabelTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.fJob.GetJobElasticSchedulingLabel(tt.job)
			if res != tt.wantRes {
				t.Errorf("GetJobElasticSchedulingLabel() res = %v, wantRes is %v", res, tt.wantRes)
			}
		})
	}
}

type isNormalJobNeedRestartTestCase struct {
	name    string
	fJob    *FaultJob
	wantRes bool
}

func buildIsNormalJobNeedRestartTestCases() []isNormalJobNeedRestartTestCase {
	return []isNormalJobNeedRestartTestCase{
		{
			name:    "01-IsNormalJobNeedRestart return false when fJob is nil",
			fJob:    nil,
			wantRes: false,
		},
		{
			name:    "02-IsNormalJobNeedRestart return true when IsSoftwareFault is true",
			fJob:    &FaultJob{FaultTasks: []FaultTask{{IsSoftwareFault: true}}},
			wantRes: true,
		},
		{
			name: "03-IsNormalJobNeedRestart return true when FaultHandling is PreSeparateNPU",
			fJob: &FaultJob{FaultTasks: []FaultTask{{
				Reason: []FaultReasonList{
					{FaultDeviceList: FaultDeviceList{FaultHandling: PreSeparateNPU}},
				},
			}}},
			wantRes: true,
		},
	}
}

func TestIsNormalJobNeedRestart(t *testing.T) {
	testCases := buildIsNormalJobNeedRestartTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if res := tt.fJob.IsNormalJobNeedRestart(); res != tt.wantRes {
				t.Errorf("IsNormalJobNeedRestart() res = %v, wantRes is %v", res, tt.wantRes)
			}
		})
	}
}

func TestGetJobFaultNPUTaskNum(t *testing.T) {
	t.Run("01-GetJobFaultNPUTaskNum return 0 when fJob.FaultTasks is empty slice",
		func(t *testing.T) {
			fJob := &FaultJob{FaultTasks: []FaultTask{}}
			if res := fJob.GetJobFaultNPUTaskNum(); res != 0 {
				t.Errorf("GetJobFaultNPUTaskNum() res = %v, wantRes = 0", res)
			}
		})
	t.Run("02-GetJobFaultNPUTaskNum return greater than 0 when fJob.FaultTasks is not empty slice",
		func(t *testing.T) {
			fJob := &FaultJob{FaultTasks: []FaultTask{{
				UseCardName: []string{"npu1", "npu2"},
			}}}
			if res := fJob.GetJobFaultNPUTaskNum(); res <= 0 {
				t.Errorf("GetJobFaultNPUTaskNum() res = %v, wantRes > 0", res)
			}
		})
}

func TestIsJobGraceDeleteSuccess(t *testing.T) {
	fJob := &FaultJob{FaultTasks: []FaultTask{{
		IsFaultTask: true,
		TaskUID:     "task01",
		UseCardName: []string{"npu1", "npu2"},
	}}}
	jobInfo := test.FakeNormalTestJob("job0", util.NPUIndex2)
	jobInfo.Tasks = map[api.TaskID]*api.TaskInfo{"task01": {Pod: &v1.Pod{}}}
	t.Run("01-isJobGraceDeleteSuccess return true when jobInfo.Tasks is not nil", func(t *testing.T) {
		if res := fJob.isJobGraceDeleteSuccess(jobInfo); res != true {
			t.Errorf("isJobGraceDeleteSuccess() res = %v, wantRes is false", res)
		}
	})
	t.Run("02-isJobGraceDeleteSuccess return true when jobInfo.PodGroup.Labels is not nil",
		func(t *testing.T) {
			fJob.PendingSessionNum = spPendingTimes
			jobInfo.PodGroup.Labels = map[string]string{
				util.SinglePodTag: util.EnableFunc,
			}
			if res := fJob.isJobGraceDeleteSuccess(jobInfo); res != true {
				t.Errorf("isJobGraceDeleteSuccess() res = %v, wantRes is false", res)
			}
		})
}
