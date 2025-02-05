// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package service a series of controller test function
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/common"
	"clusterd/pkg/interface/grpc/pb"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func TestHandleNotifyDump(t *testing.T) {
	convey.Convey("Test handleNotifyDump", t, func() {
		ctl := &EventController{
			uuid: "test-uuid",
			jobInfo: common.JobBaseInfo{
				JobId:     "test-job-id",
				JobName:   "test-job-name",
				Namespace: "test-namespace",
			},
		}
		convey.Convey("01-update cache fault and pod fail, should return err", func() {
			mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "updateCacheFaultAndPod",
				func() ([]*pb.FaultRank, []string, error) { return nil, nil, errors.New("mock error") })
			defer mockFunc.Reset()

			result, respCode, err := ctl.handleNotifyDump()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(respCode == common.ServerInnerError, convey.ShouldBeTrue)
			convey.So(result == "", convey.ShouldBeTrue)
		})

		mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "updateCacheFaultAndPod",
			func() ([]*pb.FaultRank, []string, error) { return nil, []string{"1", "2"}, nil })
		defer mockFunc.Reset()
		convey.Convey("02-write reset info fail, should return err", func() {
			mockFunc1 := gomonkey.ApplyFuncReturn(common.WriteResetInfoToCM, nil, errors.New("mock error1"))
			defer mockFunc1.Reset()
			result, respCode, err := ctl.handleNotifyDump()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(respCode == common.OperateConfigMapError, convey.ShouldBeTrue)
			convey.So(result == "", convey.ShouldBeTrue)
		})
		mockFunc1 := gomonkey.ApplyFuncReturn(common.WriteResetInfoToCM,
			&v1.ConfigMap{Data: map[string]string{"key": "value"}}, nil)
		defer mockFunc1.Reset()
		convey.Convey("03-signal enqueue success, should return nil", func() {
			mockFunc2 := gomonkey.ApplyPrivateMethod(&EventController{}, "signalEnqueue",
				func(signal *pb.ProcessManageSignal) (string, common.RespCode, error) { return "", common.OK, nil })
			defer mockFunc2.Reset()
			_, respCode, err := ctl.handleNotifyDump()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.OK, convey.ShouldBeTrue)
		})
	})
}

func TestHandleNotifyWaitFaultFlushing(t *testing.T) {
	convey.Convey("Test handleNotifyWaitFaultFlushing", t, func() {
		ctl := &EventController{uuid: "test-uuid", jobInfo: common.JobBaseInfo{
			JobId:         "test-job-id",
			JobName:       "test-job-name",
			Namespace:     "test-namespace",
			RecoverConfig: common.RecoverConfig{PlatFormMode: true}}}
		convey.Convey("01-wait plat form strategy fail, should return timeout event", func() {
			mockFunc := gomonkey.ApplyFuncReturn(WaitPlatFormStrategyReady, "", errors.New("mock error"))
			defer mockFunc.Reset()
			result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.WaitPlatStrategyTimeout, convey.ShouldBeTrue)
			convey.So(result == common.WaitPlatStrategyTimeoutEvent, convey.ShouldBeTrue)
		})
		strategy := "dump"
		mockFunc := gomonkey.ApplyFuncReturn(WaitPlatFormStrategyReady, strategy, nil)
		defer mockFunc.Reset()
		convey.Convey("should dump when occur fault", func() {
			mockFunc1 := gomonkey.ApplyPrivateMethod(&EventController{}, "shouldDumpWhenOccurFault",
				func(*EventController) bool { return true })
			defer mockFunc1.Reset()
			convey.Convey("02-config strategies not contain plat strategy, should return err", func() {
				ctl.jobInfo.MindXConfigStrategies = []string{"exit"}
				result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(respCode == common.ServerInnerError, convey.ShouldBeTrue)
				convey.So(result == "", convey.ShouldBeTrue)
			})
			convey.Convey("03-config strategies contain plat strategy, should return dump event", func() {
				ctl.jobInfo.MindXConfigStrategies = []string{"dump", "exit"}
				result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
				convey.So(err, convey.ShouldBeNil)
				convey.So(respCode == common.OK, convey.ShouldBeTrue)
				convey.So(result == common.DumpForFaultEvent, convey.ShouldBeTrue)
			})
		})
		mockFunc1 := gomonkey.ApplyPrivateMethod(&EventController{}, "shouldDumpWhenOccurFault",
			func(*EventController) bool { return false })
		defer mockFunc1.Reset()
		convey.Convey("04-retry write reset cm fail, should return operate cm error", func() {
			mockFunc2 := gomonkey.ApplyFuncReturn(common.RetryWriteResetCM, nil, errors.New("mock error"))
			defer mockFunc2.Reset()
			result, respCode, err := ctl.handleNotifyWaitFaultFlushing()
			convey.So(err, convey.ShouldBeNil)
			convey.So(respCode == common.OperateConfigMapError, convey.ShouldBeTrue)
			convey.So(result == common.NotifyFinishEvent, convey.ShouldBeTrue)
		})
	})
}

func TestOnlySupportDumpStrategy(t *testing.T) {
	convey.Convey("onlySupportDumpStrategy", t, func() {
		ctl := EventController{}
		convey.Convey("01-ProcessRecoverEnable is false, should return false", func() {
			ctl.jobInfo.ProcessRecoverEnable = false
			convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
		})
		ctl.jobInfo.ProcessRecoverEnable = true
		convey.Convey("02-cluster not support dump, should return false", func() {
			ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessRecoverStrategyName,
				constant.ProcessDumpStrategyName}
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeFalse)
		})
		convey.Convey("03-cluster support exit, should return false", func() {
			ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessExitStrategyName}
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeFalse)
		})
		ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessDumpStrategyName}
		convey.Convey("04-cluster support dump, not platform mode, should return true", func() {
			ctl.jobInfo.PlatFormMode = false
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeTrue)
		})
		ctl.jobInfo.PlatFormMode = true
		convey.Convey("05-platform mode, platform strategy is not dump, should return false", func() {
			ctl.platStrategy = constant.ProcessExitStrategyName
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeFalse)
		})
		convey.Convey("06-platform mode, platform strategy is dump, should return true", func() {
			ctl.platStrategy = constant.ProcessDumpStrategyName
			convey.So(ctl.onlySupportDumpStrategy(), convey.ShouldBeTrue)
		})
	})
}

func TestShouldDumpWhenOccurFault(t *testing.T) {
	ctl := EventController{}
	convey.Convey("shouldDumpWhenOccurFault", t, func() {
		convey.Convey("01-ProcessRecoverEnable is false, should return false", func() {
			ctl.jobInfo.ProcessRecoverEnable = false
			convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
		})
		ctl.jobInfo.ProcessRecoverEnable = true
		convey.Convey("02-config not support dump, should return false", func() {
			ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessRecoverStrategyName,
				constant.ProcessDumpStrategyName}
			convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
		})
		ctl.jobInfo.MindXConfigStrategies = []string{constant.ProcessDumpStrategyName}
		ctl.jobInfo.PlatFormMode = false
		convey.Convey("03-healthState is healthy, should return false", func() {
			ctl.healthState = constant.HealthyState
			convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
		})
		convey.Convey("05-healthState is not healthy", func() {
			convey.Convey("healthState is unhealthy should return true", func() {
				ctl.healthState = constant.UnHealthyState
				convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeTrue)
			})
			ctl.healthState = constant.SubHealthyState
			convey.Convey("06-healthState is subHealthy and graceExit is false, should return false", func() {
				ctl.jobInfo.GraceExit = false
				convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeFalse)
			})
			convey.Convey("07-healthState is subHealthy and graceExit is true, should return true", func() {
				ctl.jobInfo.GraceExit = true
				convey.So(ctl.shouldDumpWhenOccurFault(), convey.ShouldBeTrue)
			})
		})
	})
}

func TestUpdateCacheFaultAndPod(t *testing.T) {
	ctl := EventController{
		faultPod: map[string]string{},
	}
	convey.Convey("updateCacheFaultAndPod", t, func() {
		faultRank := []*pb.FaultRank{{RankId: "1", FaultType: constant.NormalFaultType}}
		ranks := []string{"1"}
		mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "normalFaultAssociateSameNodeRank",
			func(*EventController) ([]*pb.FaultRank, []string) { return faultRank, ranks })
		defer mockFunc.Reset()

		convey.Convey("01-get pod map failed, should return error", func() {
			mockFunc1 := gomonkey.ApplyFuncReturn(common.GetPodMap, nil, errors.New("mock error"))
			defer mockFunc1.Reset()
			_, _, err := ctl.updateCacheFaultAndPod()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-update cache fault and pod success, should not return error", func() {
			podMap := map[string]string{"1": "uuid1"}
			mockFunc1 := gomonkey.ApplyFuncReturn(common.GetPodMap, podMap, nil)
			defer mockFunc1.Reset()
			retFaultrank, retRank, err := ctl.updateCacheFaultAndPod()
			convey.So(err, convey.ShouldBeNil)
			convey.So(retFaultrank, convey.ShouldResemble, faultRank)
			convey.So(retRank, convey.ShouldResemble, ranks)
			convey.So(ctl.faultPod, convey.ShouldResemble, podMap)
			convey.So(ctl.cacheUceFault, convey.ShouldBeNil)
			convey.So(ctl.cacheNormalFault, convey.ShouldResemble, faultRank)
		})
	})
}
