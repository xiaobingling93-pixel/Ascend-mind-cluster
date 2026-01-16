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

// Package kubeclient a series of k8s function ut
package kubeclient

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestClientK8sMethodWriteDeviceInfoDataIntoCMA5(t *testing.T) {
	convey.Convey("test ClientK8s method WriteDeviceInfoDataIntoCMA5", t, func() {
		utKubeClient, err := initK8S()
		if err != nil {
			t.Fatal("test WriteDeviceInfoDataIntoCMA5 init kubernetes failed")
		}
		nodeDeivceData := common.NodeDeviceInfoCache{
			DeviceInfo: common.NodeDeviceInfo{DeviceList: map[string]string{}},
		}
		convey.Convey("01-should return err when common.MarshalData failed", func() {
			patch := gomonkey.ApplyFunc(common.MarshalData, func(data interface{}) []byte {
				return nil
			})
			defer patch.Reset()
			ret := utKubeClient.WriteDeviceInfoDataIntoCMA5(&nodeDeivceData, "", common.SwitchFaultInfo{}, common.DpuInfo{})
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return success when all is called success", func() {
			patch := gomonkey.ApplyPrivateMethod(utKubeClient, "createOrUpdateDeviceCM",
				func(cm *v1.ConfigMap) error {
					return nil
				})
			defer patch.Reset()
			ret := utKubeClient.WriteDeviceInfoDataIntoCMA5(&nodeDeivceData, "", common.SwitchFaultInfo{}, common.DpuInfo{})
			convey.So(ret, convey.ShouldBeNil)
		})
	})
}
