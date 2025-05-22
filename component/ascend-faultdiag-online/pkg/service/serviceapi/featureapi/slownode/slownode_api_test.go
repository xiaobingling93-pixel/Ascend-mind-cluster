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

// Package slownode is a DT collection for func in slownode_api
package slownode

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/context/sohandle"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	"ascend-faultdiag-online/pkg/model/servicemodel"
	"ascend-faultdiag-online/pkg/service/request"
)

// MockHandler 模拟 SoHandler 的行为
type MockHandler struct {
	mock.Mock
}

// ExecuteFunc 模拟 SoHandler 的 ExecuteFunc 方法
func (m *MockHandler) ExecuteFunc(input []byte, output []byte) (int, error) {
	args := m.Called(input, output)
	return args.Int(0), args.Error(1)
}

func createMockSoHandler(mockHandler *MockHandler) *sohandle.SoHandler {
	return &sohandle.SoHandler{
		SoHandle: nil,
		SoType:   "mock",
		ExecuteFunc: func(input []byte, output []byte) (int, error) {
			return mockHandler.ExecuteFunc(input, output)
		},
	}
}

func createMockCtxData(mockHandler *MockHandler) *contextdata.CtxData {
	return &contextdata.CtxData{
		Framework: &contextdata.Framework{
			Config:       nil, // 如果需要，可以提供具体的配置
			SoHandlerMap: map[string]*sohandle.SoHandler{"slownode": createMockSoHandler(mockHandler)},
			ReqQue:       make(chan *request.Context, 10),
			IsRunning:    true,
			StopChan:     make(chan struct{}),
			Logger:       log.New(os.Stdout, "TEST: ", log.LstdFlags),
		},
	}
}

func TestClusterStartFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	diagCtx := &diagcontext.DiagContext{}
	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 第一次调用 registerCallback
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 第二次调用 start

	// 调用方法
	err := ClusterStartFunc(ctxData, diagCtx, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "ClusterStartFunc 成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "slownode Cluster start successfully", reqCtx.Response.Msg, "Response 消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestClusterStartFunc_Error(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	// 创建其他必要的参数
	diagCtx := &diagcontext.DiagContext{}
	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, errors.New("mock error")).Once() // registerCallback 出错

	// 调用方法
	err := ClusterStartFunc(ctxData, diagCtx, reqCtx, inputModel)

	// 验证结果
	assert.Error(t, err, "ClusterStartFunc 返回错误")
	assert.Equal(t, "mock error", err.Error(), "错误消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestClusterStopFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	// 创建其他必要的参数
	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	diagCtx := &diagcontext.DiagContext{}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 stop

	// 调用方法
	err := ClusterStopFunc(ctxData, diagCtx, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "ClusterStopFunc 成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "slownode Cluster stop successfully", reqCtx.Response.Msg, "Response 消息正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestClusterReloadFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 reload

	// 调用方法
	err := ClusterReloadFunc(ctxData, nil, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "ClusterReloadFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "slownode Cluster reload successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestClusterReloadFunc_Error(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, errors.New("mock error")).Once() // reload 出错

	// 调用方法
	err := ClusterReloadFunc(ctxData, nil, reqCtx, inputModel)

	// 验证结果
	assert.Error(t, err, "ClusterReloadFunc 应该返回错误")
	assert.Equal(t, "mock error", err.Error(), "错误消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestNodeStartFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 registerCallBack
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 start

	// 调用方法
	err := NodeStartFunc(ctxData, nil, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "NodeStartFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "slownode Node start successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestNodeStartFunc_Error(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, errors.New("mock error")).Once() // start 出错

	// 调用方法
	err := NodeStartFunc(ctxData, nil, reqCtx, inputModel)

	// 验证结果
	assert.Error(t, err, "NodeStartFunc 应该返回错误")
	assert.Equal(t, "mock error", err.Error(), "错误消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestNodeStopFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 stop

	// 调用方法
	err := NodeStopFunc(ctxData, nil, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "NodeStopFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "slownode Node stop successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}

func TestNodeReloadFunc(t *testing.T) {
	// 创建 MockHandler
	mockHandler := new(MockHandler)

	// 创建模拟的上下文数据
	ctxData := createMockCtxData(mockHandler)

	reqCtx := &request.Context{
		Response: &servicemodel.ResponseBody{},
	}
	inputModel := &slownode.SlowNodeInput{}

	// 模拟 handler 的行为
	mockHandler.On("ExecuteFunc", mock.Anything, mock.Anything).Return(0, nil).Once() // 调用 reload

	// 调用方法
	err := NodeReloadFunc(ctxData, nil, reqCtx, inputModel)

	// 验证结果
	assert.NoError(t, err, "NodeReloadFunc 应该成功执行")
	assert.Equal(t, enum.Success, reqCtx.Response.Status, "Response 状态应该是 Success")
	assert.Equal(t, "slownode Node reload successfully", reqCtx.Response.Msg, "Response 消息应该正确")

	// 验证 MockHandler 的调用
	mockHandler.AssertExpectations(t)
}
