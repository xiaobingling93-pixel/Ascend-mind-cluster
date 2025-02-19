/* Copyright(C) 2021-2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for general constants
package common

const (
	// npuListCacheKey Cache key
	npuListCacheKey = "npu-exporter-npu-list"
	// Cache key for parsing-device result
	tickerFailedPattern = "%s ticker failed, task shutdown"
	// UpdateCachePattern Update cache pattern
	UpdateCachePattern = "update Cache,key is %s"
)

const (
	cacheSize = 128
	// GeneralDevTagKey is the default value of devTagKey in telegraf, it means the metric is not related to any device
	GeneralDevTagKey = 0xFFFF
)
