// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package dpu a series of dpu function
package dpu

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const safeDpuCMSize = 1000

// ParseDpuInfoCM get dpu info from configmap obj
func ParseDpuInfoCM(dpuCm *v1.ConfigMap) (*constant.DpuInfoCM, error) {
	dpuInfoCM := constant.DpuInfoCM{}
	data, ok := dpuCm.Data[api.DpuInfoCMDataKey]
	if !ok {
		return nil, fmt.Errorf("configmap %s has no key <%s>", dpuCm.Name, api.DpuInfoCMDataKey)
	}

	if unmarshalErr := json.Unmarshal([]byte(data), &dpuInfoCM); unmarshalErr != nil {
		return nil, fmt.Errorf("configmap %s unmarshal error: %v", dpuCm.Name, unmarshalErr)
	}

	if dpuInfoCM.BusType == "" {
		return nil, fmt.Errorf("%s has no key <%s>", api.DpuInfoCMDataKey, api.DpuInfoCMBusTypeKey)
	}
	if dpuInfoCM.DPUList == nil {
		return nil, fmt.Errorf("%s has no key <%s>", api.DpuInfoCMDataKey,
			api.DpuInfoCMDpuListKey)
	}
	if dpuInfoCM.NpuToDpusMap == nil {
		return nil, fmt.Errorf("%s has no key <%s>", api.DpuInfoCMDataKey,
			api.DpuInfoCMNpuToDpusMapKey)
	}

	result := constant.DpuInfoCM{
		CmName:       dpuCm.Name,
		DPUList:      dpuInfoCM.DPUList,
		BusType:      dpuInfoCM.BusType,
		NpuToDpusMap: dpuInfoCM.NpuToDpusMap,
		UpdateTime:   dpuInfoCM.UpdateTime,
	}
	return &result, nil
}

// DeepCopy deep copy dpu info
func DeepCopy(info *constant.DpuInfoCM) *constant.DpuInfoCM {
	if info == nil {
		return nil
	}
	data, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("%s json marshal failed: %v", api.DpuLogPrefix, err)
		return nil
	}
	newInfo := &constant.DpuInfoCM{}
	if unErr := json.Unmarshal(data, newInfo); unErr != nil {
		hwlog.RunLog.Errorf("%s json unmarshal failed: %v", api.DpuLogPrefix, unErr)
		return nil
	}
	return newInfo
}

// GetSafeData put every 1000 dpu cm info together
func GetSafeData(dpuCMInfos map[string]*constant.DpuInfoCM) []string {
	dpuCMCount := len(dpuCMInfos)
	if dpuCMCount == 0 {
		return []string{}
	}
	if dpuCMCount <= safeDpuCMSize {
		return []string{util.ObjToString(dpuCMInfos)}
	}

	dpuSlice := make([]string, 0, (dpuCMCount+safeDpuCMSize-1)/safeDpuCMSize)
	dpuSliceTemp := make(map[string]*constant.DpuInfoCM, safeDpuCMSize)
	for cmName, dpuInfo := range dpuCMInfos {
		dpuSliceTemp[cmName] = dpuInfo
		if len(dpuSliceTemp)%safeDpuCMSize == 0 {
			dpuSlice = append(dpuSlice, util.ObjToString(dpuSliceTemp))
			dpuSliceTemp = make(map[string]*constant.DpuInfoCM, safeDpuCMSize)
		}
	}
	if len(dpuSliceTemp) != 0 {
		dpuSlice = append(dpuSlice, util.ObjToString(dpuSliceTemp))
	}
	return dpuSlice
}
