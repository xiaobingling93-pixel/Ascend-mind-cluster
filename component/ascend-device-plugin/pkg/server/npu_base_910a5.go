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
	"sync"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	apiCommon "ascend-common/devmanager/common"
)

var hcclTopoFilePathMap = map[int8]string{
	common.ProductTypeServer:    common.Server8PTopoPath,
	common.ProductType1D:        common.Pod1DTopoPath,
	common.ProductType2D:        common.Pod2DTopoPath,
	common.ProductType16PServer: common.Server16PTopoPath,
	common.ProductType32PServer: common.Server32PTopoPath,
	common.ProductType1PCard:    common.Card1PTopoPath,
	common.ProductType4PCard:    common.Card4PTopoPath,
}

var rankLevelInfoKeyArrMap = map[string][]string{
	common.A5300ICardName: {
		api.LevelInfoTypeIgnore, api.LevelInfoTypeIgnore, api.LevelInfoTypeIgnore, api.LevelInfoTypeRoCE,
	},
	common.A54P300ICardName: {
		api.LevelInfoTypeUB, api.LevelInfoTypeIgnore, api.LevelInfoTypeIgnore, api.LevelInfoTypeRoCE,
	},
}

const (
	size50M        = 50 * 1024 * 1024
	addrTypeEID    = "EID"
	addrTypeIPV4   = "IPV4"
	decimal        = 10
	hexadecimal    = 16
	addrNumsLength = 2
)

var npuBase *NpuBase

func init() {
	npuBase = NewNpuBase()
}

type netTypeAndFeIdList struct {
	netType  string
	feIdList []uint
}

// ProductBase for product info in os domain
type ProductBase struct {
	superPodSize   uint32
	superPodID     uint32
	serverIndex    uint32
	chassisID      uint32
	superPodType   uint8
	nodeInternalIP string
	topoFilePath   string
	cardType       string
	topoInfo       *TopoInfo
}

// NpuBase save npu base info
type NpuBase struct {
	productInfo    *ProductBase
	eidPortMap     map[string][]string
	portMapMutex   sync.RWMutex
	urmaDevInfoMap map[int32][]apiCommon.UrmaDeviceInfo
}

// NewNpuBase for new NpuBase instance
func NewNpuBase() *NpuBase {
	return &NpuBase{
		eidPortMap:     make(map[string][]string),
		portMapMutex:   sync.RWMutex{},
		urmaDevInfoMap: make(map[int32][]apiCommon.UrmaDeviceInfo),
	}
}
