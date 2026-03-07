/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package incrementfault is test for process increment faults
package incrementfault

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/manualfault"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
)

const (
	defaultFaultWindowHours = 24
	defaultFaultThreshold   = 3
	defaultFaultFreeHours   = 48
)

var (
	validPolicy = conf.ManuallySeparatePolicy{
		Enabled: true,
		Separate: struct {
			FaultWindowHours int `yaml:"fault_window_hours"`
			FaultThreshold   int `yaml:"fault_threshold"`
		}{
			FaultWindowHours: defaultFaultWindowHours,
			FaultThreshold:   defaultFaultThreshold,
		},
		Release: struct {
			FaultFreeHours int `yaml:"fault_free_hours"`
		}{
			FaultFreeHours: defaultFaultFreeHours,
		},
	}
)

func TestProcessor(t *testing.T) {
	preparePodStorage()
	var p1 = gomonkey.ApplyFuncReturn(podgroup.GetPodGroup, v1beta1.PodGroup{
		ObjectMeta: v1.ObjectMeta{Name: pgName1},
	})
	defer p1.Reset()
	convey.Convey("test func 'Process', manually separate npu enabled is false", t, testCloseManualSep)
	convey.Convey("test func 'Process', input type is invalid", t, testInputIsNil)
	convey.Convey("test func 'Process', when have job, update job fault mgr", t, testWhenHaveJob)
	convey.Convey("test func 'Process', when have increment fault", t, testIncrementFault)
	convey.Convey("test func 'Process', when fault is same", t, testSameFault)
}

func testCloseManualSep() {
	validPolicy1 := validPolicy
	validPolicy1.Enabled = false
	conf.SetManualSeparatePolicy(validPolicy1)
	resetRelatedCache()
	ori := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	IncrementFaultProcessor.Process(ori)
	convey.So(manualfault.FaultCmInfo.Len(), convey.ShouldEqual, 0)
}

func testInputIsNil() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	res := IncrementFaultProcessor.Process(nil)
	convey.So(res, convey.ShouldBeNil)
}

func testWhenHaveJob() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	pod.SavePod(podDemo1)
	pod.SavePod(podDemo2)
	defer pod.DeletePod(podDemo1)
	defer pod.DeletePod(podDemo2)

	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    updateFaultReceiveTime(faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1)),
		UpdateConfigmap: nil,
	}
	_ = IncrementFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	jobFaults := manualfault.JobFaultMgr.GetFaultsByJobId(job1)
	convey.So(len(jobFaults), convey.ShouldEqual, len0)
	jobFaults = manualfault.JobFaultMgr.GetFaultsByJobId(job2)
	convey.So(len(jobFaults), convey.ShouldEqual, len1)
}

func testIncrementFault() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	pod.SavePod(podDemo1)
	pod.SavePod(podDemo3)
	pod.SavePod(podDemo4)
	defer func() {
		pod.DeletePod(podDemo1)
		pod.DeletePod(podDemo3)
		pod.DeletePod(podDemo4)
	}()

	// increment fault is node1 Ascend910-6 and node2 Ascend910-2
	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    updateFaultReceiveTime(faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo2)),
		UpdateConfigmap: nil,
	}
	IncrementFaultProcessor.nodeDeviceCmMap = updateFaultReceiveTime(faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1))
	_ = IncrementFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	jobFaults := manualfault.JobFaultMgr.GetFaultsByJobId(job1)
	convey.So(len(jobFaults), convey.ShouldEqual, len2)
}

func updateFaultReceiveTime(allConfigmap map[string]*constant.AdvanceDeviceFaultCm) map[string]*constant.AdvanceDeviceFaultCm {
	var newAllCm = make(map[string]*constant.AdvanceDeviceFaultCm)
	for name, cm := range allConfigmap {
		var faultDeviceLists = make(map[string][]constant.DeviceFault)
		for node, faults := range cm.FaultDeviceList {
			var newFaults []constant.DeviceFault
			for _, fault := range faults {
				var newFaultTimeAndLevel = make(map[string]constant.FaultTimeAndLevel)
				for code, level := range fault.FaultTimeAndLevelMap {
					level.FaultReceivedTime = time.Now().UnixMilli()
					newFaultTimeAndLevel[code] = level
				}
				fault.FaultTimeAndLevelMap = newFaultTimeAndLevel
				newFaults = append(newFaults, fault)
			}
			faultDeviceLists[node] = newFaults
			cm.FaultDeviceList = faultDeviceLists
		}
		newAllCm[name] = cm
	}
	return newAllCm
}

func testSameFault() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	pod.SavePod(podDemo1)
	pod.SavePod(podDemo3)
	defer func() {
		pod.DeletePod(podDemo1)
		pod.DeletePod(podDemo3)
	}()
	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    updateFaultReceiveTime(faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1)),
		UpdateConfigmap: nil,
	}
	IncrementFaultProcessor.nodeDeviceCmMap = updateFaultReceiveTime(faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1))
	_ = IncrementFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	jobFaults := manualfault.JobFaultMgr.GetFaultsByJobId(job1)
	convey.So(len(jobFaults), convey.ShouldEqual, len0)
}

func resetRelatedCache() {
	manualfault.InitJobFaultManager(constant.DefaultSlidingWindow)
	manualfault.InitCounter()
	manualfault.InitFaultCmInfo()
	IncrementFaultProcessor.nodeDeviceCmMap = make(map[string]*constant.AdvanceDeviceFaultCm)
}

func preparePodStorage() {
	podDemo1 = getDemoPod(node1, podName1, dev0, job1)
	podDemo2 = getDemoPod(node1, podName2, dev5, job2)
	podDemo3 = getDemoPod(node1, podName3, dev6, job1)
	podDemo4 = getDemoPod(node2, podName4, dev2, job1)
}

func TestIsContainFault(t *testing.T) {
	oldFaults := []constant.DeviceFault{
		{
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				code1: {
					FaultLevel:        constant.NotHandleFault,
					FaultReceivedTime: receiveTime1,
				},
				code0: {
					FaultLevel:        constant.NotHandleFault,
					FaultReceivedTime: receiveTime1,
				},
			},
		},
	}
	convey.Convey("test func 'isContainFault', contain new fault", t, func() {
		res := isContainFault(getNewFault1(), oldFaults)
		convey.So(res, convey.ShouldBeTrue)
	})
	convey.Convey("test func 'isContainFault', not contain new fault", t, func() {
		// fault received time is different
		res := isContainFault(getNewFault2(), oldFaults)
		convey.So(res, convey.ShouldBeFalse)

		// fault level is different
		res = isContainFault(getNewFault3(), oldFaults)
		convey.So(res, convey.ShouldBeFalse)

		// not contain fault code
		res = isContainFault(getNewFault4(), oldFaults)
		convey.So(res, convey.ShouldBeFalse)
	})
}

func getNewFault1() constant.DeviceFault {
	return constant.DeviceFault{
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			code1: {
				FaultLevel:        constant.NotHandleFault,
				FaultReceivedTime: receiveTime1,
			},
		},
	}
}

func getNewFault2() constant.DeviceFault {
	return constant.DeviceFault{
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			code1: {
				FaultLevel:        constant.NotHandleFault,
				FaultReceivedTime: receiveTime1 + 1,
			},
		},
	}
}

func getNewFault3() constant.DeviceFault {
	return constant.DeviceFault{
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			code1: {
				FaultLevel:        constant.RestartRequest,
				FaultReceivedTime: receiveTime1,
			},
		},
	}
}

func getNewFault4() constant.DeviceFault {
	return constant.DeviceFault{
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			code2: {
				FaultLevel:        constant.NotHandleFault,
				FaultReceivedTime: receiveTime1,
			},
		},
	}
}
