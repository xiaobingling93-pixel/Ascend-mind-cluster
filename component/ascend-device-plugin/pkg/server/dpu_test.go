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

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device/dpucontrol"
	"ascend-common/common-utils/ethtool"
)

func TestUpdateDpuHealthy(t *testing.T) {
	tests := []npuFaultTestCase{
		buildFilterDpuFaultTestCase1(),
		buildFilterDpuFaultTestCase2(),
		buildFilterDpuFaultTestCase3(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hdm := &HwDevManager{
				dpuManager: &dpucontrol.DpuFilter{
					NpuWithDpuInfos: tt.npuWithDpuInfos,
				},
			}
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": tt.deviceList,
			}
			outputs := getMockOutputCellsForEthState(tt.npuWithDpuInfos)
			patchEth := gomonkey.ApplyFuncSeq(ethtool.GetInterfaceOperState, outputs)
			hdm.updateDpuHealthy(groupDevice)
			for i, dev := range tt.deviceList {
				if got := dev.DpuHealth; got != tt.want[i] {
					t.Errorf("DpuHealthy = %v, want %v, for dev %v", got, tt.want[i], dev.CardID)
				}
			}
			patchEth.Reset()
		})
	}
}

func getMockOutputCellsForEthState(npuWithDpuInfos []dpucontrol.NpuWithDpuInfo) []gomonkey.OutputCell {
	uniqueSet := make(map[string]struct{})
	outputs := make([]gomonkey.OutputCell, 0)
	for _, dpuInfos := range npuWithDpuInfos {
		for _, dpu := range dpuInfos.DpuInfo {
			if _, exist := uniqueSet[dpu.DeviceName]; !exist {
				uniqueSet[dpu.DeviceName] = struct{}{}
				fmt.Println(dpu)
				outputs = append(outputs, gomonkey.OutputCell{Values: gomonkey.Params{dpu.Operstate, nil}})
			}
		}
	}
	return outputs
}

type npuFaultTestCase struct {
	name            string
	npuWithDpuInfos []dpucontrol.NpuWithDpuInfo
	deviceList      []*common.NpuDevice
	want            []string
}

func buildFilterDpuFaultTestCase1() npuFaultTestCase {
	dpuInfo1 := dpucontrol.BaseDpuInfo{
		DeviceName: "enps0", Operstate: "down", DeviceId: "0x1825", Vendor: "0x10b5",
	}
	dpuInfo2 := dpucontrol.BaseDpuInfo{
		DeviceName: "enps2", Operstate: "down", DeviceId: "0x1825", Vendor: "0x10b5",
	}
	return npuFaultTestCase{
		name: "No DPU state is up, DpuHealth is unhealthy",
		npuWithDpuInfos: []dpucontrol.NpuWithDpuInfo{
			{
				NpuId:   int32(0),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1, dpuInfo2},
			},
			{
				NpuId:   int32(1),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1, dpuInfo2},
			},
		},
		deviceList: []*common.NpuDevice{
			{
				DeviceName: "Ascend910-0",
				CardID:     int32(0),
				DpuHealth:  "",
			},
			{
				DeviceName: "Ascend910-1",
				CardID:     int32(1),
				DpuHealth:  "",
			},
		},
		want: []string{v1beta1.Unhealthy, v1beta1.Unhealthy},
	}
}

func buildFilterDpuFaultTestCase2() npuFaultTestCase {
	dpuInfo1 := dpucontrol.BaseDpuInfo{
		DeviceName: "enps0", Operstate: "up", DeviceId: "0x1825", Vendor: "0x10b5",
	}
	dpuInfo2 := dpucontrol.BaseDpuInfo{
		DeviceName: "enps2", Operstate: "down", DeviceId: "0x1825", Vendor: "0x10b5",
	}
	return npuFaultTestCase{
		name: "[UB type] If one of the two DPUs is in the up state, it is healthy.",
		npuWithDpuInfos: []dpucontrol.NpuWithDpuInfo{
			{
				NpuId:   int32(0),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1, dpuInfo2},
			},
			{
				NpuId:   int32(1),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1, dpuInfo2},
			},
		},
		deviceList: []*common.NpuDevice{
			{
				DeviceName: "Ascend910-0",
				CardID:     int32(0),
				DpuHealth:  "",
			},
			{
				DeviceName: "Ascend910-1",
				CardID:     int32(1),
				DpuHealth:  "",
			},
		},
		want: []string{v1beta1.Healthy, v1beta1.Healthy},
	}
}

func buildFilterDpuFaultTestCase3() npuFaultTestCase {
	dpuInfo1 := dpucontrol.BaseDpuInfo{
		DeviceName: "eth0", Operstate: "up", DeviceId: "0x1019", Vendor: "0x15b3",
	}
	dpuInfo2 := dpucontrol.BaseDpuInfo{
		DeviceName: "eth1", Operstate: "down", DeviceId: "0x1019", Vendor: "0x15b3",
	}
	return npuFaultTestCase{
		name: "[PCIe type] dpu is in the up state, it is healthy.",
		npuWithDpuInfos: []dpucontrol.NpuWithDpuInfo{
			{
				NpuId:   int32(0),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1},
			},
			{
				NpuId:   int32(1),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo2},
			},
		},
		deviceList: []*common.NpuDevice{
			{
				DeviceName: "Ascend910-0",
				CardID:     int32(0),
				DpuHealth:  "",
			},
			{
				DeviceName: "Ascend910-1",
				CardID:     int32(1),
				DpuHealth:  "",
			},
		},
		want: []string{v1beta1.Healthy, v1beta1.Unhealthy},
	}
}

func TestIsNpuMatched(t *testing.T) {
	tests := append(
		buildIsNpuMatchedTestCasesPart1(),
		buildIsNpuMatchedTestCasesPart2()...,
	)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNpuMatched(tt.args.str, tt.args.target); got != tt.want {
				t.Errorf("isNpuMatched() = %v, want %v", got, tt.want)
			}
		})
	}
}

type isNpuMatchedArgs struct {
	str    string
	target string
}

type isNpuMatchedTestCase struct {
	name string
	args isNpuMatchedArgs
	want bool
}

func buildIsNpuMatchedTestCasesPart1() []isNpuMatchedTestCase {
	return []isNpuMatchedTestCase{
		{
			name: "01-should return true when target matches cardIdStr",
			args: isNpuMatchedArgs{
				str:    "Ascend910-1",
				target: "1",
			},
			want: true,
		},
		{
			name: "02-should return false when target does not match any cardIdStr",
			args: isNpuMatchedArgs{
				str:    "Ascend910-0",
				target: "3",
			},
			want: false,
		},
		{
			name: "03-should return false when input string is empty",
			args: isNpuMatchedArgs{
				str:    "",
				target: "0",
			},
			want: false,
		},
		{
			name: "04-should return false when string does not contain '-'",
			args: isNpuMatchedArgs{
				str:    "Ascend910",
				target: "0",
			},
			want: false,
		},
	}
}

func buildIsNpuMatchedTestCasesPart2() []isNpuMatchedTestCase {
	return []isNpuMatchedTestCase{
		{
			name: "05-should return false when cardId is not an integer",
			args: isNpuMatchedArgs{
				str:    "Ascend910-abc",
				target: "0",
			},
			want: false,
		},
		{
			name: "06-should handle cardIdStr with modulo NodeNum8",
			args: isNpuMatchedArgs{
				str:    "Ascend910-9", // 9 % 8 = 1
				target: "1",
			},
			want: true,
		},
		{
			name: "07-should return false if modulo does not match target",
			args: isNpuMatchedArgs{
				str:    "Ascend910-10", // 10 % 8 = 2
				target: "3",
			},
			want: false,
		},
	}
}
