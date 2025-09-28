// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api/slownet"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/common/constant"
)

const (
	num1 = 1
	path = "/tmp/test.conf"
)

type fileTestCase struct {
	name                 string
	filePath             string
	checkPathErr         error
	openFileErr          error
	fileCheckerErr       error
	writeToFileErr       error
	getConfigPathErr     error
	writeConfigToFileErr error
	expectError          bool
	expectErrorStr       string
	config               *constant.CathelperConf
	file                 *os.File
	writeStringErr       error
}

func buildCommonConfig() *constant.CathelperConf {
	config := &constant.CathelperConf{
		SuppressedPeriod: 0,
		NetworkType:      num1,
		PingType:         0,
		PingTimes:        num1,
		PingInterval:     num1,
		Period:           num1,
		NetFault:         "off",
	}
	return config
}

func TestSaveConfigToFile(t *testing.T) {
	superpodID := "test-superpod-id"
	config := buildCommonConfig()
	testCases := buildFileTestCases1(config)
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			testPath := "/test/path/cathelper.conf"
			patches.ApplyFuncReturn(slownet.GetConfigPathForDetect, testPath, tc.getConfigPathErr)
			patches.ApplyFuncReturn(writeConfigToFile, tc.writeConfigToFileErr)
			err := saveConfigToFile(superpodID, tc.config)
			if tc.expectError {
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, tc.expectErrorStr)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestWriteConfigToFile(t *testing.T) {
	config := buildCommonConfig()
	testCases := buildFileTestCases2()
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFuncReturn(utils.CheckPath, "", tc.checkPathErr)
			file := &os.File{}
			if tc.openFileErr != nil {
				patches.ApplyFuncReturn(os.OpenFile, nil, tc.openFileErr)
			} else {
				patches.ApplyFuncReturn(os.OpenFile, file, nil)
			}
			patches.ApplyFuncReturn(utils.RealFileChecker, "", tc.fileCheckerErr)
			patches.ApplyFuncReturn(writeToFile, tc.writeToFileErr)
			patches.ApplyMethodReturn(file, "Close", nil)
			err := writeConfigToFile(config, tc.filePath)
			if tc.expectError {
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, tc.expectErrorStr)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestWriteToFile(t *testing.T) {
	config := buildCommonConfig()
	mockFile := &os.File{}
	testCases := buildFileTestCases3(config, mockFile)
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			if tc.writeStringErr != nil {
				patches.ApplyMethodReturn(mockFile, "WriteString", 0, tc.writeStringErr)
			} else {
				patches.ApplyMethodReturn(mockFile, "WriteString", 0, nil)
			}
			err := writeToFile(tc.config, tc.file)
			if tc.expectError {
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, tc.expectErrorStr)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func buildFileTestCases1(config *constant.CathelperConf) []fileTestCase {
	testCases := []fileTestCase{
		{
			name:           "nil config",
			config:         nil,
			expectError:    true,
			expectErrorStr: "config is nil",
		},
		{
			name:             "get config path failed",
			config:           config,
			getConfigPathErr: errors.New("get path failed"),
			expectError:      true,
			expectErrorStr:   "get path failed",
		},
		{
			name:                 "write config to file failed",
			config:               config,
			writeConfigToFileErr: errors.New("write file failed"),
			expectError:          true,
			expectErrorStr:       "write file failed",
		},
		{
			name:        "success",
			config:      config,
			expectError: false,
		},
	}
	return testCases
}

func buildFileTestCases2() []fileTestCase {
	testCases := []fileTestCase{
		{
			name:        "valid config and file path",
			filePath:    path,
			expectError: false,
		},
		{
			name:           "invalid file path",
			filePath:       "/invalid/path/test.conf",
			checkPathErr:   errors.New("invalid path"),
			expectError:    true,
			expectErrorStr: "invalid path",
		},
		{
			name:           "failed to create file - all retries failed",
			filePath:       path,
			openFileErr:    errors.New("permission denied"),
			expectError:    true,
			expectErrorStr: "permission denied",
		},
		{
			name:           "file checker failed",
			filePath:       path,
			fileCheckerErr: errors.New("file check failed"),
			expectError:    true,
			expectErrorStr: "file check failed",
		},
		{
			name:           "write to file failed",
			filePath:       path,
			writeToFileErr: errors.New("write failed"),
			expectError:    true,
			expectErrorStr: "write failed",
		},
	}
	return testCases
}

func buildFileTestCases3(config *constant.CathelperConf, mockFile *os.File) []fileTestCase {
	testCases := []fileTestCase{
		{
			name:           "nil file",
			config:         config,
			file:           nil,
			expectError:    true,
			expectErrorStr: "file pointer or config paramters are nil",
		},
		{
			name:           "nil config",
			config:         nil,
			file:           mockFile,
			expectError:    true,
			expectErrorStr: "file pointer or config paramters are nil",
		},
		{
			name:           "write string failed",
			config:         config,
			file:           mockFile,
			writeStringErr: errors.New("write failed"),
			expectError:    true,
			expectErrorStr: "write failed",
		},
		{
			name:        "success",
			config:      config,
			file:        mockFile,
			expectError: false,
		},
	}
	return testCases
}
