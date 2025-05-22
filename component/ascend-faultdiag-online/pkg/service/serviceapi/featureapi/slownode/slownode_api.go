/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package slownode provides API
*/
package slownode

/*
#cgo linux LDFLAGS: -ldl

#include <stdint.h>
#include <stdlib.h>

// 声明 Go 导出的回调函数
extern void slowNodeAlgoResultCallBack(char* msg);
extern void slowNodeDataParseResultCallback(char* msg);
extern void slowNodeMergeParalleGroupInfoResultCallback(char* msg);
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/service/servicecore"
)

const (
	start    = "start"
	stop     = "stop"
	reload   = "reload"
	name     = "slownode"
	byteSize = 1024
	filemode = 0644
	// SlowCalculateRanks 慢计算卡
	SlowCalculateRanks = "SlowCalculateRanks"
	// SlowCommunicationDomains 慢通信域
	SlowCommunicationDomains = "SlowCommunicationDomains"
	// SlowCommunicationRanks  慢通信卡
	SlowCommunicationRanks = "SlowCommunicationRanks"
	// SlowHostNodes hosts侧慢卡
	SlowHostNodes = "SlowHostNodes"
	// SlowIORanks 慢IO卡
	SlowIORanks = "SlowIORanks"
	// DegradationLevel 劣化百分点
	DegradationLevel = "DegradationLevel"
	// IsSlow 是否存在慢节点
	IsSlow = "IsSlow"

	maxDegradationCount int = 5

	// slow node report
	slowNodeFaultCode         = "110001010"
	slowNodeRecoveryFaultCode = "100001011"
)

var dContext *diagcontext.DiagContext
var contextData *contextdata.CtxData

// ClusterStart 开启Cluster
func ClusterStart() *servicecore.Api {
	return servicecore.BuildApi(start, &slownode.SlowNodeInput{}, ClusterStartFunc, nil)
}

// ClusterStop 关闭Cluster
func ClusterStop() *servicecore.Api {
	return servicecore.BuildApi(stop, &slownode.SlowNodeInput{}, ClusterStopFunc, nil)
}

// ClusterReload reload Cluster
func ClusterReload() *servicecore.Api {
	return servicecore.BuildApi(reload, &slownode.SlowNodeInput{}, ClusterReloadFunc, nil)
}

// NodeStart 开启Node
func NodeStart() *servicecore.Api {
	return servicecore.BuildApi(start, &slownode.SlowNodeInput{}, NodeStartFunc, nil)
}

// NodeStop 关闭Node
func NodeStop() *servicecore.Api {
	return servicecore.BuildApi(stop, &slownode.SlowNodeInput{}, NodeStopFunc, nil)
}

// NodeReload reload Node
func NodeReload() *servicecore.Api {
	return servicecore.BuildApi(reload, &slownode.SlowNodeInput{}, NodeReloadFunc, nil)
}

func startFunc(
	ctxData *contextdata.CtxData,
	diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context,
	inputModel *slownode.SlowNodeInput,
	target enum.DeployMode,
) error {
	dContext = diagCtx
	contextData = ctxData
	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		return fmt.Errorf("start failed: no handler for %s", name)
	}

	eventType := inputModel.EventType
	funcUintPtr := uintptr(unsafe.Pointer(C.slowNodeAlgoResultCallBack))

	if eventType == enum.DataParse {
		if target == enum.Cluster {
			funcUintPtr = uintptr(unsafe.Pointer(C.slowNodeMergeParalleGroupInfoResultCallback))
		} else {
			funcUintPtr = uintptr(unsafe.Pointer(C.slowNodeDataParseResultCallback))
		}
	}

	inputBytes, err := createJsonInput(enum.Register, funcUintPtr, target, inputModel)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]register %s req data: %s", inputModel.EventType, string(inputBytes))

	// 定义输出缓冲区
	output1 := make([]byte, byteSize)
	_, err = handler.ExecuteFunc(inputBytes, output1)
	if err != nil {
		return err
	}
	inputBytes, err = createJsonInput(enum.Start, 0, target, inputModel)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]start %s req data: %s", inputModel.EventType, string(inputBytes))

	// 定义输出缓冲区
	output2 := make([]byte, byteSize)
	_, err = handler.ExecuteFunc(inputBytes, output2)
	if err != nil {
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = fmt.Sprintf("start %s successfully", inputModel.EventType)
	return nil
}

func stopFunc(
	ctxData *contextdata.CtxData,
	reqCtx *request.Context,
	inputModel *slownode.SlowNodeInput,
	target enum.DeployMode,
) error {
	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		return fmt.Errorf("start failed: no handler for %s", name)
	}

	inputBytes, err := createJsonInput(enum.Stop, 0, target, inputModel)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]stop %s req data: %s", inputModel.EventType, string(inputBytes))
	// 定义输出缓冲区
	output := make([]byte, byteSize)
	_, err = handler.ExecuteFunc(inputBytes, output)
	if err != nil {
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = fmt.Sprintf("stop %s successfully", inputModel.EventType)
	return nil
}

func reloadFunc(
	ctxData *contextdata.CtxData,
	diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context,
	inputModel *slownode.SlowNodeInput,
	target enum.DeployMode,
) error {
	dContext = diagCtx
	contextData = ctxData
	handler := ctxData.Framework.SoHandlerMap[name]
	if handler == nil {
		return fmt.Errorf("start failed: no handler for %s", name)
	}
	inputBytes, err := createJsonInput(enum.Reload, 0, target, inputModel)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]reload %s req data: %s", inputModel.EventType, string(inputBytes))
	// 定义输出缓冲区
	output := make([]byte, byteSize)
	_, err = handler.ExecuteFunc(inputBytes, output)
	if err != nil {
		return err
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = fmt.Sprintf("reload %s successfully", inputModel.EventType)
	return nil
}

// ClusterStartFunc /start api对应方法，启动算法
func ClusterStartFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, inputModel *slownode.SlowNodeInput) error {
	return startFunc(ctxData, diagCtx, reqCtx, inputModel, enum.Cluster)
}

// ClusterStopFunc /stop api对应方法，停止算法
func ClusterStopFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, inputModel *slownode.SlowNodeInput) error {
	return stopFunc(ctxData, reqCtx, inputModel, enum.Cluster)
}

// ClusterReloadFunc /reload api对应方法，reload算法
func ClusterReloadFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, inputModel *slownode.SlowNodeInput) error {
	return reloadFunc(ctxData, diagCtx, reqCtx, inputModel, enum.Cluster)
}

// NodeStartFunc /start api对应方法，启动算法
func NodeStartFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, inputModel *slownode.SlowNodeInput) error {
	return startFunc(ctxData, diagCtx, reqCtx, inputModel, enum.Node)
}

// NodeStopFunc /stop api对应方法，停止算法
func NodeStopFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, inputModel *slownode.SlowNodeInput) error {
	return stopFunc(ctxData, reqCtx, inputModel, enum.Node)
}

// NodeReloadFunc /reload api对应方法，reload算法
func NodeReloadFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, inputModel *slownode.SlowNodeInput) error {
	return reloadFunc(ctxData, diagCtx, reqCtx, inputModel, enum.Node)
}

// 回调将数据放到ctx中
//
//export slowNodeAlgoResultCallBack
func slowNodeAlgoResultCallBack(cMessage *C.char) {
	message := C.GoString(cMessage)
	defer C.free(unsafe.Pointer(cMessage))

	hwlog.RunLog.Infof("[FD-OL SLOWNODE]%v: got the slow node algo callback data: %v",
		contextData.Framework.Config.Mode, message)
	if dContext == nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process slow node algo callback failed: context is nil")
		return
	}
	if contextData.Framework.Config.Mode == enum.Cluster {
		clusterProcessSlowNodeAlgoCallback(message)
	} else {
		nodeProcessSlowNodeAlgoCallback(message)
	}
}

// slowNodeDataParseResultCallback the callback func for slow node data parse.
//
//export slowNodeDataParseResultCallback
func slowNodeDataParseResultCallback(cMessage *C.char) {
	message := C.GoString(cMessage)
	defer C.free(unsafe.Pointer(cMessage))

	hwlog.RunLog.Infof("[FD-OL SLOWNODE]%v: got the data parse callback data: %v",
		contextData.Framework.Config.Mode, message)
	if dContext == nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process data parse result callback data failed: context is nil")
		return
	}
	// only fd-on running in the node will trigger this callback
	// we don't need to determine the target is cluster or node
	nodeProcessDataParseCallback(message)
}

// slowNodeMergeParalleGroupInfoResultCallback the callback func for slow node merge parallel group info.
//
//export slowNodeMergeParalleGroupInfoResultCallback
func slowNodeMergeParalleGroupInfoResultCallback(cMessage *C.char) {
	message := C.GoString(cMessage)
	defer C.free(unsafe.Pointer(cMessage))

	hwlog.RunLog.Infof("[FD-OL SLOWNODE]%v: got the merge parallel group info callback data: %s",
		contextData.Framework.Config.Mode, message)
	if dContext == nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process merge parallel group info callback data failed: context is nil")
		return
	}
	clusterProcessParallelGroupInfoCallback(message)
}

// createJsonInput create a []byte as input data of a param of execute
func createJsonInput(
	command enum.Command,
	funcPtr uintptr,
	target enum.DeployMode,
	model *slownode.SlowNodeInput,
) ([]byte, error) {
	input := slownode.Input{
		Command:   command,
		Target:    target,
		Func:      uint(funcPtr),
		EventType: enum.SlowNodeAlgo,
	}
	if model == nil {
		return json.Marshal(input)
	}
	input.EventType = model.EventType
	input.Model = model.SlowNodeAlgoInput
	if input.EventType == enum.DataParse {
		input.Model = model.DataParseInput
	}
	return json.Marshal(input)
}

func convertMaptoStruct[T slownode.NodeSlowNodeAlgoResult | slownode.ClusterSlowNodeAlgoResult](
	data map[string]map[string]any, target *T) error {
	if len(data) == 0 {
		return fmt.Errorf("callback data is empty: %s", data)
	}
	for _, v := range data {
		if len(v) == 0 {
			return fmt.Errorf("callback data is empty: %s", data)
		}
		for _, result := range v {
			dataBytes, err := json.Marshal(result)
			if err != nil {
				return err
			}
			err = json.Unmarshal(dataBytes, target)
			return err
		}
	}
	return nil
}
