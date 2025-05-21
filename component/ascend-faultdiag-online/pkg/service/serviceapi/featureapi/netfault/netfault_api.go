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

// Package netfault for fault network feature
package netfault

/*
#cgo linux LDFLAGS: -ldl

#include <stdint.h>
#include <stdlib.h>

// 声明 Go 导出的回调函数
extern void netfaultResultCallBack(char* msg);

*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"unsafe"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature"
	"ascend-faultdiag-online/pkg/model/feature/netfault"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

var dContext *diagcontext.DiagContext
var contextData *contextdata.CtxData

// ControllerStartFunc start controller of cluster
func ControllerStartFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, model *feature.Status) error {
	dContext = diagCtx
	contextData = ctxData

	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		err := errors.New("no handler for " + name)
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}

	input, err := createJSONInput(registerCallBack, uintptr(unsafe.Pointer(C.netfaultResultCallBack)), cluster)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}
	// 定义输出缓冲区
	output1 := make([]byte, byteSize)
	_, err = handler.ExecuteFunc([]byte(input), output1)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}

	input, err = createJSONInput(apiStart, 0, cluster)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}
	// 定义输出缓冲区
	output2 := make([]byte, byteSize)
	_, err = handler.ExecuteFunc([]byte(input), output2)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller start err: %v", err)
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = "netfault controller start successfully"
	return nil
}

// ControllerStopFunc stop controller of cluster
func ControllerStopFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, model *feature.Status) error {
	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		err := errors.New("no handler for " + name)
		hwlog.RunLog.Errorf("netfault controller stop err: %v", err)
		return err
	}
	input, err := createJSONInput(apiStop, 0, cluster)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller stop err: %v", err)
		return err
	}
	// 定义输出缓冲区
	output := make([]byte, byteSize)
	_, err = handler.ExecuteFunc([]byte(input), output)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller stop err: %v", err)
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = "netfault controller stop successfully"
	return nil
}

// ControllerReloadFunc reload controller of cluster
func ControllerReloadFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, model *feature.Status) error {
	dContext = diagCtx
	contextData = ctxData
	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		err := errors.New("no handler for " + name)
		hwlog.RunLog.Errorf("netfault controller reload err: %v", err)
		return err
	}
	input, err := createJSONInput(apiReload, 0, cluster)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller reload err: %v", err)
		return err
	}
	// 定义输出缓冲区
	output := make([]byte, byteSize)
	_, err = handler.ExecuteFunc([]byte(input), output)
	if err != nil {
		hwlog.RunLog.Errorf("netfault controller reload err: %v", err)
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = "netfault controller reload successfully"
	return nil
}

// 回调将数据放到ctx中
//
//export netfaultResultCallBack
func netfaultResultCallBack(cMessage *C.char) {
	message := C.GoString(cMessage)
	defer C.free(unsafe.Pointer(cMessage))

	if dContext == nil {
		hwlog.RunLog.Errorf("netfault result callback err: context is nil")
		return
	}
	if enum.Cluster != contextData.Framework.Config.Mode {
		return
	}
	var clusterResult []netfault.ClusterResult
	if err := json.Unmarshal([]byte(message), &clusterResult); err != nil {
		hwlog.RunLog.Errorf("error parsing clusterResult JSON: %v", err)
		return
	}
	if len(clusterResult) > 0 {
		hwlog.RunLog.Infof("the result of callback is %+v", clusterResult)
	}
	parseAndAddMetric(clusterResult, dContext)
	return
}

// 解析结果入库到MetricPool
func parseAndAddMetric(clusterResult []netfault.ClusterResult, context *diagcontext.DiagContext) {
	models := make([]*metricmodel.MetricReqModel, 0)
	// 通用的创建 MetricReqModel 的方法
	createMetric := func(domainType string, domainValue string, name string,
		valueType enum.MetricValueType, value string) *metricmodel.MetricReqModel {
		return &metricmodel.MetricReqModel{
			Domain: []*metricmodel.DomainItem{
				{
					DomainType: enum.MetricDomainType(domainType),
					Value:      domainValue,
				},
			},
			Name:      name,
			ValueType: valueType,
			Value:     value,
		}
	}
	// 从检测结果添加指标模型
	for i, result := range clusterResult {
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.TaskId,
			enum.StringMetric, result.TaskID))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.MinLossRate,
			enum.FloatMetric, fmt.Sprintf("%f", result.MinLossRate)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.MaxLossRate,
			enum.FloatMetric, fmt.Sprintf("%f", result.MaxLossRate)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.AvgLossRate,
			enum.FloatMetric, fmt.Sprintf("%f", result.AvgLossRate)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.MinDelay,
			enum.FloatMetric, fmt.Sprintf("%f", result.MinDelay)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.MaxDelay,
			enum.FloatMetric, fmt.Sprintf("%f", result.MaxDelay)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.AvgDelay,
			enum.FloatMetric, fmt.Sprintf("%f", result.AvgDelay)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.FaultType,
			enum.FloatMetric, strconv.Itoa(result.FaultType)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.SrcID,
			enum.StringMetric, result.SrcID))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.SrcType,
			enum.FloatMetric, strconv.Itoa(result.SrcType)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.DstID,
			enum.StringMetric, result.DstID))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.DstType,
			enum.FloatMetric, strconv.Itoa(result.DstType)))
		models = append(models, createMetric(enum.NetworkDomain, strconv.Itoa(i), constants.Level,
			enum.FloatMetric, strconv.Itoa(result.Level)))
	}
	addMetricFromClusterResult(models, context)
}

func addMetricFromClusterResult(models []*metricmodel.MetricReqModel, context *diagcontext.DiagContext) {
	// 统一处理 MetricPool 添加逻辑
	for _, metric := range models {
		if slicetool.ValueIn(metric.ValueType, []enum.MetricValueType{enum.FloatMetric, enum.StringMetric}) != nil {
			contextData.Framework.Logger.Println("Unknown Metric Type", metric.ValueType)
			continue
		}
		domain := context.DomainFactory.GetInstance(metric.Domain)
		context.MetricPool.AddMetric(&diagcontext.Metric{Domain: domain, Name: metric.Name},
			metric.Value, metric.ValueType)
	}
}

// 构造 JSON 输入
func createJSONInput(command string, funcPtr uintptr, target string) (string, error) {
	input := map[string]interface{}{
		inputCommand: command,
		inputTarget:  target,
	}

	if funcPtr != 0 {
		input[inputFunc] = funcPtr
	}
	jsonData, err := json.Marshal(input)
	return string(jsonData), err
}
