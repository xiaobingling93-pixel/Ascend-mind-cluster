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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"errors"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

// WriteDeviceInfoDataIntoCMA5 write deviceinfo into config map for A5
func (ki *ClientK8s) WriteDeviceInfoDataIntoCMA5(nodeDeviceData *common.NodeDeviceInfoCache,
	manuallySeparateNPU string, switchInfo common.SwitchFaultInfo, dpuInfo common.DpuInfo) error {

	nodeDeviceData.DeviceInfo.UpdateTime = time.Now().Unix()
	nodeDeviceData.CheckCode = common.MakeDataHash(nodeDeviceData.DeviceInfo)

	var data, switchData, dpuData []byte
	if data = common.MarshalData(nodeDeviceData); len(data) == 0 {
		return errors.New("marshal nodeDeviceData failed")
	}
	if switchData = common.MarshalData(switchInfo); len(switchData) == 0 {
		return errors.New("marshal switchDeviceData failed")
	}
	if dpuData = common.MarshalData(dpuInfo); len(dpuData) == 0 {
		return errors.New("marshal DpuDeviceData failed")
	}
	deviceInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ki.DeviceInfoName,
			Namespace: api.KubeNS,
			Labels:    map[string]string{api.CIMCMLabelKey: common.CmConsumerValue},
		},
	}
	deviceInfoCM.Data = map[string]string{
		api.DeviceInfoCMDataKey:                   string(data),
		api.SwitchInfoCMDataKey:                   string(switchData),
		common.DeviceInfoCMManuallySeparateNPUKey: manuallySeparateNPU,
		api.DpuInfoCMDataKey:                      string(dpuData),
	}
	hwlog.RunLog.Debugf("write device info cache into cm: %s/%s.", deviceInfoCM.Namespace, deviceInfoCM.Name)
	return ki.createOrUpdateDeviceCM(deviceInfoCM)
}
