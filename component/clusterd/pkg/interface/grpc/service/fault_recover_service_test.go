// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package service a series of fault recover service test function
package service

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/common"
)

func TestNotifyFaultInfoForJob(t *testing.T) {
	convey.Convey("Test notifyFaultInfoForJob", t, func() {
		svr := &FaultRecoverService{
			eventCtl: map[string]*EventController{
				"mockJob1": {jobInfo: common.JobBaseInfo{
					RecoverConfig: common.RecoverConfig{GraceExit: false}},
				},
			},
		}
		convey.Convey("01-controller not exist, should not add event", func() {
			mockJob := "mockJob2"
			info := constant.JobFaultInfo{JobId: mockJob}
			svr.notifyFaultInfoForJob(info)
			convey.So(svr.eventCtl[mockJob], convey.ShouldBeNil)
		})
		convey.Convey("02-subHealthy fault and not graceExit, should not add event", func() {
			mockJob := "mockJob1"
			info := constant.JobFaultInfo{
				JobId:        mockJob,
				HealthyState: constant.SubHealthyState,
			}
			svr.notifyFaultInfoForJob(info)
			convey.So(svr.eventCtl[mockJob].healthState, convey.ShouldBeEmpty)
		})
		convey.Convey("03-unHealthy fault, should add event", func() {
			mockJob := "mockJob1"
			info := constant.JobFaultInfo{
				JobId:        mockJob,
				HealthyState: constant.UnHealthyState,
				FaultList: []constant.FaultRank{
					{RankId: "1", DoStepRetry: false},
					{RankId: "2", DoStepRetry: true},
				},
			}
			mockFunc := gomonkey.ApplyPrivateMethod(&EventController{}, "addEvent",
				func(*EventController, string) {})
			defer mockFunc.Reset()
			svr.notifyFaultInfoForJob(info)
			ctl := svr.eventCtl[mockJob]
			convey.So(ctl, convey.ShouldNotBeNil)
			convey.So(ctl.healthState == constant.UnHealthyState, convey.ShouldBeTrue)
			convey.So(len(ctl.cacheNormalFault) == 1, convey.ShouldBeTrue)
			convey.So(len(ctl.cacheUceFault) == 1, convey.ShouldBeTrue)
		})
	})
}
