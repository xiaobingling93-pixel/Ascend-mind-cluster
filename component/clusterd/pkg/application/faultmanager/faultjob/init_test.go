// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultjob contain fault job process
package faultjob

import (
	"fmt"
	"testing"

	"ascend-common/common-utils/hwlog"
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../../testdata/clusterd.log"}
	hwLogConfig.MaxBackups = 30
	hwLogConfig.MaxAge = 7
	hwLogConfig.LogLevel = 0
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	initConfig()
	m.Run()
}

func initConfig() {
	relationFilePath := "../../../../build/relationFaultCustomization.json"
	if fileBytes := LoadConfigFromFile(relationFilePath); fileBytes != nil {
		initRelationFaultStrategies(fileBytes)
		initRelationFaultCodesMap()
	}
	durationFilePath := "../../../../build/faultDuration.json"
	if fileBytes := LoadConfigFromFile(durationFilePath); fileBytes != nil {
		initFaultDuration(fileBytes)
		initFaultCodeTimeOutMap()
	}
}
