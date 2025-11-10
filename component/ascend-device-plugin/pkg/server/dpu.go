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

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"time"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device/dpucontrol"
	"ascend-common/api"
	"ascend-common/common-utils/ethtool"
	"ascend-common/common-utils/hwlog"
)

const (
	listenDpuInterval = 5 * time.Second
	maxUpdateInterval = 6 * time.Hour
)

var (
	lastData       []common.DpuCMData
	lastUpdateTime time.Time
)

// ListenDpu periodically query the DPU operstate and write to the configmap
func (hdm *HwDevManager) ListenDpu(ctx context.Context) {
	if common.ParamOption.RealCardType != api.Ascend910A5 || len(hdm.dpuManager.NpuWithDpuInfos) == 0 {
		return
	}

	ticker := time.NewTicker(listenDpuInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hdm.handleDpu()
		case <-ctx.Done():
			hwlog.RunLog.Warnf("%s ListenDpu stop work due to ctx.Done(): %v", api.DpuLogPrefix, ctx.Err())
			return
		}
	}
}

func (hdm *HwDevManager) handleDpu() {
	var dpuList []common.DpuCMData
	uniqueMap := make(map[string]dpucontrol.BaseDpuInfo, api.NpuCountPerNode)
	npuToDpusMap := make(map[string][]string, api.NpuCountPerNode)
	for _, npuWithDpu := range hdm.dpuManager.NpuWithDpuInfos {
		strKey := strconv.Itoa(int(npuWithDpu.NpuId))
		npuToDpusMap[strKey] = make([]string, 0)
		for _, dpu := range npuWithDpu.DpuInfo {
			if _, exists := uniqueMap[dpu.DeviceName]; !exists {
				uniqueMap[dpu.DeviceName] = dpu
			}
			npuToDpusMap[strKey] = append(npuToDpusMap[strKey], dpu.DeviceName)
		}
	}
	for _, dpu := range uniqueMap {
		state, err := ethtool.GetInterfaceOperState(dpu.DeviceName)
		if err != nil {
			hwlog.RunLog.Errorf("%s GetInterfaceOperState err: %v", api.DpuLogPrefix, err)
			state = api.DpuStatusDown
		}
		dpuList = append(dpuList, common.DpuCMData{
			Name:      dpu.DeviceName,
			Operstate: state,
			DeviceID:  dpu.DeviceId,
			VendorID:  dpu.Vendor,
		})
	}
	sort.Slice(dpuList, func(i, j int) bool {
		return dpuList[i].Name < dpuList[j].Name
	})
	needUpdate := !reflect.DeepEqual(dpuList, lastData) || time.Since(lastUpdateTime) > maxUpdateInterval
	if !needUpdate {
		return
	}
	if err := hdm.manager.GetKubeClient().WriteDpuDataIntoCM(hdm.dpuManager.UserConfig.BusType, dpuList,
		npuToDpusMap); err != nil {
		hwlog.RunLog.Errorf("%s write DPU info failed: %v", api.DpuLogPrefix, err)
		return
	}
	lastData = dpuList
	lastUpdateTime = time.Now()
}
