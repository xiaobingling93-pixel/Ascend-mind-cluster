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

// Package manager is to provide other service tools, i.e. clusterd
package manager

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"

	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	_ "taskd/common/testtool"
	"taskd/common/utils"
)

type fakeClient struct {
	pb.RecoverClient
}

func (f *fakeClient) Init(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) Register(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeProcessManageSignal(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (pb.Recover_SubscribeProcessManageSignalClient, error) {
	return nil, nil
}

func (f *fakeClient) ReportStopComplete(ctx context.Context, in *pb.StopCompleteRequest, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) ReportRecoverStrategy(ctx context.Context, in *pb.RecoverStrategyRequest, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) ReportRecoverStatus(ctx context.Context, in *pb.RecoverStatusRequest, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) ReportProcessFault(ctx context.Context, in *pb.ProcessFaultRequest, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) SwitchNicTrack(ctx context.Context, in *pb.SwitchNics, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeSwitchNicSignal(ctx context.Context, in *pb.SwitchNicRequest, opts ...grpc.CallOption) (pb.Recover_SubscribeSwitchNicSignalClient, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeNotifySwitch(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (pb.Recover_SubscribeNotifySwitchClient, error) {
	return nil, nil
}

func (f *fakeClient) ReplySwitchNicResult(ctx context.Context, in *pb.SwitchResult, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) HealthCheck(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) StressTest(ctx context.Context, in *pb.StressTestParam, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeStressTestResponse(ctx context.Context, in *pb.StressTestRequest, opts ...grpc.CallOption) (pb.Recover_SubscribeStressTestResponseClient, error) {
	return nil, nil
}

func (f *fakeClient) SubscribeNotifyExecStressTest(ctx context.Context, in *pb.ClientInfo, opts ...grpc.CallOption) (pb.Recover_SubscribeNotifyExecStressTestClient, error) {
	return nil, nil
}

func (f *fakeClient) ReplyStressTestResult(ctx context.Context, in *pb.StressTestResult, opts ...grpc.CallOption) (*pb.Status, error) {
	return nil, nil
}

func TestReportControllerInfoToClusterd(t *testing.T) {
	convey.Convey("get clusterd addr failed", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) {
			return "", fmt.Errorf("get clusterd address err")
		})
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("init clusterd connect err", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, fmt.Errorf("grpc.Dial err"))

		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("send message to clusterd failed", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportRecoverStatus, false)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.RecoverStatus})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("action is unknown action", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: "action"})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestReportControllerInfoToClusterd2(t *testing.T) {
	convey.Convey("message.Action is RecoverStatus", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportRecoverStatus, true)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.RecoverStatus})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("message.Action is ProcessFault", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportProcessFault, true)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.ProcessFault})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("message.Action is RecoverStrategy", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportRecoverStrategy, true)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.RecoverStrategy})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("message.Action is StopComplete", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(utils.GetClusterdAddr, func() (string, error) { return "127.0.0.1", nil }).
			ApplyFuncReturn(grpc.Dial, nil, nil).
			ApplyMethodReturn(&grpc.ClientConn{}, "Close", nil).
			ApplyFuncReturn(reportStopComplete, true)
		res := ReportControllerInfoToClusterd(&constant.ControllerMessage{Action: constant.StopComplete})
		convey.So(res, convey.ShouldEqual, true)
	})
}

func TestReportRecoverStatus(t *testing.T) {
	message := &constant.ControllerMessage{
		Code: 0,
		Msg:  "msg",
		FaultRanks: map[int]int{
			0: 0,
		},
	}
	convey.Convey("managerInstance is nil", t, func() {
		preInst := managerInstance
		managerInstance = nil
		defer func() {
			managerInstance = preInst
		}()
		res := reportRecoverStatus(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("report recover status to clusterd ok", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		res := reportRecoverStatus(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("report recover status to clusterd failed", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodReturn(&fakeClient{}, "ReportRecoverStatus", &pb.Status{}, fmt.Errorf("err"))
		res := reportRecoverStatus(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestReportProcessFault(t *testing.T) {
	message := &constant.ControllerMessage{
		Code: 0,
		Msg:  "msg",
		FaultRanks: map[int]int{
			0: 0,
		},
	}
	convey.Convey("managerInstance is nil", t, func() {
		preInst := managerInstance
		managerInstance = nil
		defer func() {
			managerInstance = preInst
		}()
		res := reportProcessFault(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("report process fault to clusterd ok", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		res := reportProcessFault(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("report process fault to clusterd failed", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodReturn(&fakeClient{}, "ReportProcessFault", &pb.Status{}, fmt.Errorf("err"))
		res := reportProcessFault(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestReportRecoverStrategy(t *testing.T) {
	message := &constant.ControllerMessage{
		Code: 0,
		Msg:  "msg",
		FaultRanks: map[int]int{
			0: 0,
		},
	}
	convey.Convey("managerInstance is nil", t, func() {
		preInst := managerInstance
		managerInstance = nil
		defer func() {
			managerInstance = preInst
		}()
		res := reportRecoverStrategy(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("report  strategy to clusterd ok", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		res := reportRecoverStrategy(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("report strategy to clusterd failed", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodReturn(&fakeClient{}, "ReportRecoverStrategy", &pb.Status{}, fmt.Errorf("err"))
		res := reportRecoverStrategy(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
}

func TestReportStopComplete(t *testing.T) {
	message := &constant.ControllerMessage{
		Code: 0,
		Msg:  "msg",
		FaultRanks: map[int]int{
			0: 0,
		},
	}
	convey.Convey("managerInstance is nil", t, func() {
		preInst := managerInstance
		managerInstance = nil
		defer func() {
			managerInstance = preInst
		}()
		res := reportStopComplete(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("report stop complete to clusterd ok", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		res := reportStopComplete(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, true)
	})
	convey.Convey("report stop complete to clusterd failed", t, func() {
		preInst := managerInstance
		managerInstance = &BaseManager{}
		defer func() {
			managerInstance = preInst
		}()
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodReturn(&fakeClient{}, "ReportStopComplete", &pb.Status{}, fmt.Errorf("err"))
		res := reportStopComplete(message, &fakeClient{})
		convey.So(res, convey.ShouldEqual, false)
	})
}
