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

// Package manualfault is test for process manually separate faults
package manualfault

import (
	"sort"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/manualfault"
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
	convey.Convey("test func 'Process', manually separate npu enabled is false", t, testCloseManualSep)
	convey.Convey("test func 'Process', input type is invalid", t, testInputIsNil)
	convey.Convey("test func 'Process', load from manual cm", t, testLoadFromManualCm)
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
	ManualFaultProcessor.Process(ori)
	convey.So(manualfault.FaultCmInfo.Len(), convey.ShouldEqual, 0)
}

func testInputIsNil() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	res := ManualFaultProcessor.Process(nil)
	convey.So(res, convey.ShouldBeNil)
}

func addManualInfo(advanceFaultCm map[string]*constant.AdvanceDeviceFaultCm) map[string]*constant.AdvanceDeviceFaultCm {
	var newAdvancedFaultCm = make(map[string]*constant.AdvanceDeviceFaultCm)
	for node, advancedCm := range advanceFaultCm {
		deviceFaultMap := make(map[string][]constant.DeviceFault)
		for devName, faults := range advancedCm.FaultDeviceList {
			var newFaults []constant.DeviceFault
			for _, fault := range faults {
				fault.FaultCode = strings.Replace(fault.FaultCode, " ", "", -1)
				codes := strings.Split(fault.FaultCode, ",")
				for _, code := range codes {
					if code == "" && fault.FaultLevel == constant.ManuallySeparateNPU {
						fault.FaultCode = constant.ManuallySeparateNPU
						faultTimeAndLevel := constant.FaultTimeAndLevel{
							FaultTime:  constant.UnknownFaultTime,
							FaultLevel: constant.ManuallySeparateNPU,
						}
						fault.FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{
							constant.ManuallySeparateNPU: faultTimeAndLevel,
						}
					}
				}
				newFaults = append(newFaults, fault)
			}
			deviceFaultMap[devName] = append(deviceFaultMap[devName], newFaults...)
		}
		advancedCm.FaultDeviceList = deviceFaultMap
		newAdvancedFaultCm[node] = advancedCm
	}
	return newAdvancedFaultCm
}

func testLoadFromManualCm() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	nodeInfo := getDemoNodeInfo()
	manualfault.FaultCmInfo.SetNodeInfo(nodeInfo)

	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	resContent := ManualFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	sortDeviceFaultList(resContent.AllConfigmap)
	res := addManualInfo(resContent.AllConfigmap)
	want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expDeviceInfo1)
	sortDeviceFaultList(want)

	convey.So(res, convey.ShouldResemble, want)
}

func sortDeviceFaultList(advanceFaultCm map[string]*constant.AdvanceDeviceFaultCm) {
	for _, advanceDeviceCm := range advanceFaultCm {
		for _, fault := range advanceDeviceCm.FaultDeviceList {
			sort.Slice(fault, func(i, j int) bool {
				return util.MakeDataHash(fault[i]) < util.MakeDataHash(fault[j])
			})
		}
		sort.Strings(advanceDeviceCm.CardUnHealthy)
		sort.Strings(advanceDeviceCm.NetworkUnhealthy)
		sort.Strings(advanceDeviceCm.Recovering)
		sort.Strings(advanceDeviceCm.AvailableDeviceList)
	}
}

func resetRelatedCache() {
	manualfault.InitJobFaultManager(constant.DefaultSlidingWindow)
	manualfault.InitCounter()
	manualfault.InitFaultCmInfo()
}
