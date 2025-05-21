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

// Package faultclusterprocess is used to return cluster faults

package faultclusterprocess

import (
	"strconv"
	"testing"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/fault"
	"github.com/smartystreets/goconvey/convey"
)

func TestGetMaxLevel(t *testing.T) {
	convey.Convey("Given a NodeFaultInfo with various fault devices", t, func() {
		healthy := constant.HealthyState
		subHealthy := constant.SubHealthyState
		unHealthy := constant.UnHealthyState

		convey.Convey("When there are no fault devices", func() {
			faultInfo := fault.NodeFaultInfo{
				FaultDevice: []*fault.DeviceFaultInfo{},
			}
			maxLevel := getMaxLevel(faultInfo)
			convey.So(maxLevel, convey.ShouldEqual, healthy)
		})

		convey.Convey("When all devices are healthy", func() {
			faultInfo := fault.NodeFaultInfo{
				FaultDevice: []*fault.DeviceFaultInfo{
					{FaultLevel: healthy},
					{FaultLevel: healthy},
				},
			}
			maxLevel := getMaxLevel(faultInfo)
			convey.So(maxLevel, convey.ShouldEqual, healthy)
		})

		convey.Convey("When there is an unhealthy device", func() {
			faultInfo := fault.NodeFaultInfo{
				FaultDevice: []*fault.DeviceFaultInfo{
					{FaultLevel: healthy},
					{FaultLevel: unHealthy},
					{FaultLevel: subHealthy},
				},
			}
			maxLevel := getMaxLevel(faultInfo)
			convey.So(maxLevel, convey.ShouldEqual, unHealthy)
		})
	})
}

func TestGetNodeFaultInfo(t *testing.T) {
	convey.Convey("test getNodeFaultInfo", t, func() {
		const (
			testDeviceId   = 1
			testDeviceType = "NPU"
			testFaultCode  = "E1001"
		)

		convey.Convey("while nodeInfo is nil", func() {
			result := getNodeFaultInfo(nil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("while nodeInfo FaultDevList empty", func() {
			nodeInfo := &constant.NodeInfo{
				NodeInfoNoName: constant.NodeInfoNoName{
					FaultDevList: []*constant.FaultDev{},
					NodeStatus:   "",
				},
				CmName: "",
			}
			result := getNodeFaultInfo(nodeInfo)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("NodeStatus is NotHandle", func() {
			nodeInfo := &constant.NodeInfo{
				NodeInfoNoName: constant.NodeInfoNoName{
					FaultDevList: []*constant.FaultDev{
						&constant.FaultDev{
							DeviceId:   testDeviceId,
							DeviceType: testDeviceType,
							FaultCode:  []string{testFaultCode},
						},
					},
					NodeStatus: "NotHandle",
				},
				CmName: "",
			}
			result := getNodeFaultInfo(nodeInfo)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].DeviceId, convey.ShouldEqual, strconv.Itoa(int(testDeviceId)))
			convey.So(result[0].DeviceType, convey.ShouldEqual, testDeviceType)
			convey.So(result[0].FaultCodes, convey.ShouldResemble, []string{testFaultCode})
			convey.So(result[0].FaultLevel, convey.ShouldEqual, constant.HealthyState)
		})
	})
}
