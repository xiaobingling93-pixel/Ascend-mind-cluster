// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"testing"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/faultrank"
	"clusterd/pkg/common/constant"
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../testdata/clusterd.log"}
	hwLogConfig.MaxBackups = 30
	hwLogConfig.MaxAge = 7
	hwLogConfig.LogLevel = 0
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	m.Run()
}

func TestCallbackForReportUceInfo(t *testing.T) {
	t.Run("TestCallbackForReportUceInfo", func(t *testing.T) {
		infos := make([]constant.ReportRecoverInfo, 0)
		infos = append(infos, constant.ReportRecoverInfo{})
		err := GlobalFaultProcessCenter.CallbackForReportUceInfo(infos)
		if err != nil {
			t.Error("TestCallbackForReportUceInfo fail")
		}
	})
}

func TestRegister(t *testing.T) {
	t.Run("TestRegister", func(t *testing.T) {
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.AllProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.DeviceProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.NodeProcessType)
		GlobalFaultProcessCenter.Register(make(chan int, 1), constant.SwitchProcessType)
	})
}

func TestQueryJobsFaultInfo(t *testing.T) {
	t.Run("TestQueryJobsFaultInfo", func(t *testing.T) {
		faultrank.JobFaultRankProcessor.SetJobFaultRankInfos(map[string]constant.JobFaultInfo{"test": {}})
		jobsFaultInfo := GlobalFaultProcessCenter.QueryJobsFaultInfo(constant.NotHandleFault)
		if len(jobsFaultInfo) != 1 {
			t.Error("TestQueryJobsFaultInfo fail")
		}
	})
}
