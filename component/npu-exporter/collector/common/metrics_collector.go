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

// Package common for general collector
package common

import (
	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/npu-exporter/v6/collector/container"
)

// MetricsCollector metrics collector
type MetricsCollector interface {
	// Describe report metrics to prometheus
	Describe(ch chan<- *prometheus.Desc)

	// CollectToCache collect data to cache
	CollectToCache(n *NpuCollector, chipList []HuaWeiAIChip)

	// UpdatePrometheus update prometheus
	UpdatePrometheus(ch chan<- prometheus.Metric, n *NpuCollector, containerMap map[int32]container.DevicesInfo,
		chips []HuaWeiAIChip)

	// UpdateTelegraf update telegraf
	UpdateTelegraf(fieldsMap map[int]map[string]interface{}, n *NpuCollector, containerMap map[int32]container.DevicesInfo,
		chips []HuaWeiAIChip) map[int]map[string]interface{}

	// PreCollect pre handle before collect
	PreCollect(*NpuCollector, []HuaWeiAIChip)

	// PostCollect post handle after collect
	PostCollect(*NpuCollector)

	// IsSupported Check whether the current hardware supports this metric
	IsSupported(*NpuCollector) bool
}
