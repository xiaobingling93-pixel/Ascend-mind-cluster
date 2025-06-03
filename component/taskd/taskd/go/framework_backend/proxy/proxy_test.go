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

// Package proxy for taskd proxy backend
package proxy

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

var initNetworkFunc = net.InitNetwork
var originalNewProxyInstance func(config *common.TaskNetConfig) error

func init() {
	originalNewProxyInstance = newProxyInstance
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// MockNet mock net
type MockNet struct {
	InitNetworkResult *net.NetInstance
	InitNetworkErr    error
}

func (m *MockNet) InitNetwork(config *common.TaskNetConfig) (*net.NetInstance, error) {
	return m.InitNetworkResult, m.InitNetworkErr
}

func TestNewProxyInstance(t *testing.T) {
	mockNet := &MockNet{
		InitNetworkResult: &net.NetInstance{},
		InitNetworkErr:    nil,
	}
	originalInitNetwork := initNetworkFunc
	defer func() {
		initNetworkFunc = originalInitNetwork
	}()
	initNetworkFunc = func(config *common.TaskNetConfig) (*net.NetInstance, error) {
		return mockNet.InitNetwork(config)
	}
	proxyConfig := &common.TaskNetConfig{
		Pos: common.Position{
			ProcessRank: "-1",
		},
	}

	convey.Convey("test newProxyInstance", t, func() {
		convey.Convey("newProxyInstance failed", func() {
			mockNet.InitNetworkErr = errors.New("test newProxyInstance failed")
			proxyInstance = nil
			err := newProxyInstance(proxyConfig)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestValidNetConfig(t *testing.T) {
	proxy := &proxyClient{}
	validConfig := &common.TaskNetConfig{
		Pos: common.Position{
			ProcessRank: "-1",
		},
	}
	invalidConfig := &common.TaskNetConfig{
		Pos: common.Position{
			ProcessRank: "0",
		},
	}

	convey.Convey("test validNetConfig func", t, func() {
		convey.Convey("valid config return nil", func() {
			err := proxy.validNetConfig(validConfig)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("invalid config return error", func() {
			err := proxy.validNetConfig(invalidConfig)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestProxyInit(t *testing.T) {
	mockNet := &MockNet{
		InitNetworkResult: &net.NetInstance{},
		InitNetworkErr:    nil,
	}
	originalInitNetwork := initNetworkFunc
	defer func() {
		initNetworkFunc = originalInitNetwork
	}()
	initNetworkFunc = func(config *common.TaskNetConfig) (*net.NetInstance, error) {
		return mockNet.InitNetwork(config)
	}

	proxy := &proxyClient{
		proxyInfo: &proxyInfo{},
	}
	proxyConfig := &common.TaskNetConfig{
		Pos: common.Position{
			ProcessRank: "-1",
		},
	}

	convey.Convey("test init func", t, func() {

		convey.Convey("valid config failed", func() {
			invalidConfig := &common.TaskNetConfig{
				Pos: common.Position{
					ProcessRank: "0",
				},
			}
			err := proxy.initNetwork(invalidConfig)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("initnetwork failed", func() {
			mockNet.InitNetworkErr = errors.New("mock initnetwork failed")
			err := proxy.initNetwork(proxyConfig)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
