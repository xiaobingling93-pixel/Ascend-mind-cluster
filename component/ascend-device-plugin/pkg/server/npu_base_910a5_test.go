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

// Package server contains unit tests for HwDevManager methods.
package server

import (
	"context"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
	apiCommon "ascend-common/devmanager/common"
)

const (
	expectedFeId   = 1
	testPortLength = 2
)

var (
	eid = apiCommon.Eid{
		Raw: [apiCommon.EidByteSize]byte{
			0, 0, 0, 0, 0, 0, 0, expectedFeId << common.FeIdIndexBit,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
	}
	urmaDevInfo = apiCommon.UrmaDeviceInfo{
		EidCount: 1,
		EidInfos: []apiCommon.UrmaEidInfo{
			{
				EidIndex: 1,
				Eid:      eid,
			},
		},
	}
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}
