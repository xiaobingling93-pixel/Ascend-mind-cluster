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

// Package device a series of device function
package device

import (
	"encoding/json"
	"sync"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
)

// ResetTool tool for npu reset
type ResetTool struct {
	client    *kubeclient.ClientK8s
	resetInfo *ResetInfo
	mu        sync.RWMutex
}

// ResetInfo information of npu reset
type ResetInfo struct {
	// ThirdPartyResetDevs devices waits for third party to reset
	ThirdPartyResetDevs []ResetFailDevice
	// ManualResetDevs devices waits for manually reset
	ManualResetDevs []ResetFailDevice
}

// ResetFailDevice device that fail to be reset
type ResetFailDevice struct {
	// CardId npu card id
	CardId int32
	// DeviceId npu device id
	DeviceId int32
	// AssociatedCardId card id of the associated npu
	AssociatedCardId int32
	// PhyId npu physic id
	PhyID int32
}

// WriteMode the mode determines how the content is written
type WriteMode int

const (
	// WMOverwrite write mode which will overwrite content
	WMOverwrite WriteMode = iota
	// WMAppend write mode which will append to content
	WMAppend
)

var (
	instance *ResetTool
	once     sync.Once
)

// ResetToolInstance return the single instance of reset tool, load reset info from node annotation
func ResetToolInstance(client *kubeclient.ClientK8s) *ResetTool {
	once.Do(func() {
		resetTool := ResetTool{
			client:    client,
			resetInfo: &ResetInfo{},
		}
		curNode, err := client.GetNode()
		if err != nil {
			hwlog.RunLog.Errorf("fail to get node from k8s, err: %v", err)
			instance = &resetTool
			return
		}
		if curNode.Annotations == nil {
			instance = &resetTool
			return
		}
		resetTool.resetInfo = readAnnotation(curNode.Annotations, common.ResetInfoAnnotationKey)
		instance = &resetTool
	})
	return instance
}

// WriteResetInfo write reset info into cache and node annotation
func (tool *ResetTool) WriteResetInfo(resetInfo ResetInfo, writeMode WriteMode) {
	tool.mu.Lock()
	tool.resetInfo.ThirdPartyResetDevs = mergeFailDevs(tool.resetInfo.ThirdPartyResetDevs,
		resetInfo.ThirdPartyResetDevs, writeMode)
	tool.resetInfo.ManualResetDevs = mergeFailDevs(tool.resetInfo.ManualResetDevs,
		resetInfo.ManualResetDevs, writeMode)
	hwlog.RunLog.Infof("reset info change: %v", *tool.resetInfo)
	dataBytes, err := json.Marshal(*tool.resetInfo)
	if err != nil {
		hwlog.RunLog.Errorf("marshal reset info erroo, data: %v, err: %v", *tool.resetInfo, err)
		tool.mu.Unlock()
		return
	}
	tool.mu.Unlock()
	tool.writeNodeAnnotation(string(dataBytes))
}

// ReadResetInfo read reset info from cache
func (tool *ResetTool) ReadResetInfo() ResetInfo {
	tool.mu.RLock()
	defer tool.mu.RUnlock()
	return *tool.resetInfo
}

func (tool *ResetTool) writeNodeAnnotation(resetStr string) {
	if err := tool.client.AddAnnotation(common.ResetInfoAnnotationKey, resetStr); err != nil {
		hwlog.RunLog.Errorf("fail to write reset info to node annotation, err: %v", err)
	}
}

func mergeFailDevs(curDevs []ResetFailDevice, newDevs []ResetFailDevice, writeMode WriteMode) []ResetFailDevice {
	if writeMode == WMOverwrite {
		return newDevs
	}
	if writeMode == WMAppend {
		curDevs = append(curDevs, newDevs...)
		return curDevs
	}
	hwlog.RunLog.Errorf("write mode %v is invalid", writeMode)
	return curDevs
}

func readAnnotation(annotation map[string]string, key string) *ResetInfo {
	if _, exist := annotation[key]; !exist {
		return &ResetInfo{}
	}
	var ret ResetInfo
	if err := json.Unmarshal([]byte(annotation[key]), &ret); err != nil {
		hwlog.RunLog.Errorf("unmarshal node annotation failed, err: %v", err)
		return &ResetInfo{}
	}
	return &ret
}
