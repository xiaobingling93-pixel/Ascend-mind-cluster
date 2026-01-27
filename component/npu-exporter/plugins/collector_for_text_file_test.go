/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	testFilePath           = "/test/path/file.json"
	testAbsFilePath        = "/abs/test/path/file.json"
	testMetricName         = "test_metric"
	testMetricDesc         = "test description"
	testMetricVersion      = "1.0"
	testLabelKey           = "label_key"
	testLabelValue         = "label_value"
	testErrorMsg           = "test error"
	noSuchFileErrorMsg     = "no such file or directory"
	otherErrorMsg          = "permission denied"
	testMaxDataListSize    = 129
	testMaxLabelSize       = 11
	testMaxMetricNameSize  = 129
	testMaxDescSize        = 1025
	testInvalidVersion     = "2.0"
	testEmptyString        = ""
	testZeroTimestamp      = int64(0)
	testValidTimestamp     = int64(1234567890)
	testNewMetricName      = "new_metric"
	testNewMetricDesc      = "new description"
	testNewMetricVersion   = "2.0"
	testLogFlagName        = "name"
	testLogFlagVersion     = "version"
	testLogFlagDesc        = "desc"
	testDirPath            = "/test/dir"
	num2                   = 2
	testFilePath2          = "/test/path/file2.json"
	testFilePaths          = "/test/path/file1.json,/test/path/file2.json"
	testFilePathsExceedMax = "/test/path/file1.json,/test/path/file2.json," +
		"/test/path/file3.json,/test/path/file4.json,/test/path/file5.json," +
		"/test/path/file6.json,/test/path/file7.json,/test/path/file8.json," +
		"/test/path/file9.json,/test/path/file10.json,/test/path/file11.json"
	testInvalidJSON   = "invalid json"
	testEmptyFileData = ""
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
}
func TestIsDataOk(t *testing.T) {
	convey.Convey("should return nil when all fields are valid", t, func() {
		metricsData := createValidTextMetricData()
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestIsDataOkDataList(t *testing.T) {
	convey.Convey("should return error when dataList is empty", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.DataList = []DataItem{}
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "dataList is empty")
	})

	convey.Convey("should return error when dataList size exceeds max", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.DataList = make([]DataItem, testMaxDataListSize)
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "size of dataList")
	})

	convey.Convey("should return error when label size exceeds max", t, func() {
		metricsData := createValidTextMetricData()
		label := make(map[string]string)
		for i := 0; i < testMaxLabelSize; i++ {
			label[testLabelKey+string(rune(i))] = testLabelValue
		}
		metricsData.DataList[0].Label = label
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "size of first item's Label")
	})
}

func TestIsDataOkName(t *testing.T) {
	convey.Convey("should return error when name is empty", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Name = testEmptyString
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "name field is empty")
	})

	convey.Convey("should return error when name length exceeds max", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Name = createLongString(testMaxMetricNameSize)
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "length of metric name")
	})
}

func TestIsDataOkDesc(t *testing.T) {
	convey.Convey("should return error when desc is empty", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Desc = testEmptyString
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "desc field is empty")
	})

	convey.Convey("should return error when desc length exceeds max", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Desc = createLongString(testMaxDescSize)
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "length of metric desc")
	})
}

func TestIsDataOkVersion(t *testing.T) {
	convey.Convey("should return error when version is not 1.0", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Version = testInvalidVersion
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "version should be 1.0")
	})
}

func TestIsDataOkTimestamp(t *testing.T) {
	convey.Convey("should return error when timestamp is zero or negative", t, func() {
		testCases := []struct {
			name      string
			timestamp int64
		}{
			{"zero timestamp", testZeroTimestamp},
			{"negative timestamp", int64(-1)},
		}
		for _, tc := range testCases {
			metricsData := createValidTextMetricData()
			metricsData.Timestamp = tc.timestamp
			err := isDataOk(metricsData, testFilePath)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "timestamp field")
		}
	})
}

func createValidTextMetricData() *TextMetricData {
	return &TextMetricData{
		Version:   testMetricVersion,
		Desc:      testMetricDesc,
		Name:      testMetricName,
		Timestamp: testValidTimestamp,
		DataList: []DataItem{
			{
				Label: map[string]string{
					testLabelKey: testLabelValue,
				},
				Value: 1.0,
			},
		},
	}
}

func createLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

func resetGlobalMaps() {
	metricStructInfosMap = make(map[string]metricStructInfo)
	existMetrics = make(map[string]string)
	validPaths = make([]string, 0)
}

func setupFileCheckPatches(patches *gomonkey.Patches, isDir bool, checkPathErr error,
	permissionErr error) {
	patches.ApplyFuncReturn(utils.IsDir, isDir)
	patches.ApplyFuncReturn(utils.CheckPath, testAbsFilePath, checkPathErr)
	patches.ApplyFuncReturn(os.Getuid, 0)
	patches.ApplyFuncReturn(utils.DoCheckOwnerAndPermission, permissionErr)
}

func setupValidFilePatches(patches *gomonkey.Patches) {
	setupFileCheckPatches(patches, false, nil, nil)
	validData := createValidTextMetricData()
	fileData, err := json.Marshal(validData)
	if err != nil {
		fmt.Println("marshal err: ", err)
	}
	patches.ApplyFuncReturn(utils.ReadLimitBytes, fileData, nil)
}

type checkAndProcessFileTestCase struct {
	name                  string
	path                  string
	setupPatches          func(*gomonkey.Patches)
	expectedResult        bool
	expectedValidPathsLen int
	needResetGlobalMaps   bool
}

func getCheckAndProcessFileTestCases() []checkAndProcessFileTestCase {
	return []checkAndProcessFileTestCase{
		{name: "should return false when path is empty", path: testEmptyString,
			setupPatches:   nil,
			expectedResult: false},
		{name: "should return false when path is directory",
			path:           testDirPath,
			setupPatches:   func(patches *gomonkey.Patches) { patches.ApplyFuncReturn(utils.IsDir, true) },
			expectedResult: false},
		{name: "should return true when file does not exist", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				notExistErr := errors.New(noSuchFileErrorMsg)
				setupFileCheckPatches(patches, false, nil, notExistErr)
			},
			expectedResult: true},
		{name: "should return false when checkFilePermission fails", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				testErr := errors.New(testErrorMsg)
				setupFileCheckPatches(patches, false, nil, testErr)
			},
			expectedResult: false},
		{name: "should return true when file is empty", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				setupFileCheckPatches(patches, false, nil, nil)
				patches.ApplyFuncReturn(utils.ReadLimitBytes, []byte(testEmptyFileData), nil)
			},
			expectedResult: true},
		{name: "should return true when read file returns EOF", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				setupFileCheckPatches(patches, false, nil, nil)
				eofErr := errors.New("EOF")
				patches.ApplyFuncReturn(utils.ReadLimitBytes, nil, eofErr)
			},
			expectedResult: true},
		{name: "should return false when read file fails", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				setupFileCheckPatches(patches, false, nil, nil)
				testErr := errors.New(testErrorMsg)
				patches.ApplyFuncReturn(utils.ReadLimitBytes, nil, testErr)
			},
			expectedResult: false},
		{name: "should return false when processFileData succeeds", path: testFilePath,
			setupPatches:          func(patches *gomonkey.Patches) { setupValidFilePatches(patches) },
			expectedResult:        false,
			expectedValidPathsLen: 1,
			needResetGlobalMaps:   true},
	}
}

func TestCheckAndProcessFile(t *testing.T) {
	testCases := getCheckAndProcessFileTestCases()

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			if tc.needResetGlobalMaps {
				resetGlobalMaps()
			}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			if tc.setupPatches != nil {
				tc.setupPatches(patches)
			}

			result := checkAndProcessFile(tc.path)
			convey.So(result, convey.ShouldEqual, tc.expectedResult)
			if tc.expectedValidPathsLen > 0 {
				convey.So(len(validPaths), convey.ShouldEqual, tc.expectedValidPathsLen)
			}
			if tc.needResetGlobalMaps {
				resetGlobalMaps()
			}
		})
	}
}

func TestPreCheckPaths(t *testing.T) {
	convey.Convey("should exit immediately when all files exist", t, func() {
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		setupValidFilePatches(patches)
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		patches.ApplyFuncReturn(time.NewTicker, ticker)

		preCheckPaths([]string{testFilePath})
		convey.So(len(validPaths), convey.ShouldEqual, 1)
		resetGlobalMaps()
	})

	convey.Convey("should wait for missing files and timeout", t, func() {
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		notExistErr := errors.New(noSuchFileErrorMsg)
		setupFileCheckPatches(patches, false, nil, notExistErr)
		patches.ApplyFuncReturn(time.Time.Add, time.Now())

		preCheckPaths([]string{testFilePath})
		convey.So(len(validPaths), convey.ShouldEqual, 0)
		resetGlobalMaps()
	})
}

func TestIsSupported(t *testing.T) {
	convey.Convey("should return false when filePath is empty", t, func() {
		filePath = testEmptyString
		collector := &TextMetricsInfoCollector{}
		result := collector.IsSupported(nil)
		convey.So(result, convey.ShouldBeFalse)
	})

	convey.Convey("should return true when files are valid", t, func() {
		filePath = testFilePaths
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		setupValidFilePatches(patches)
		patches.ApplyFuncReturn(time.Time.Add, time.Now())

		collector := &TextMetricsInfoCollector{}
		result := collector.IsSupported(nil)
		convey.So(result, convey.ShouldBeTrue)
		resetGlobalMaps()
	})

	convey.Convey("should truncate paths when exceeds maxFileNumber", t, func() {
		filePath = testFilePathsExceedMax
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		setupValidFilePatches(patches)
		patches.ApplyFuncReturn(time.Time.Add, time.Now())

		collector := &TextMetricsInfoCollector{}
		result := collector.IsSupported(nil)
		convey.So(result, convey.ShouldBeTrue)
		resetGlobalMaps()
	})

	convey.Convey("should return false when no valid paths found", t, func() {
		filePath = testFilePaths
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		notExistErr := errors.New(noSuchFileErrorMsg)
		setupFileCheckPatches(patches, false, nil, notExistErr)
		patches.ApplyFuncReturn(time.Time.Add, time.Now())

		collector := &TextMetricsInfoCollector{}
		result := collector.IsSupported(nil)
		convey.So(result, convey.ShouldBeFalse)
		resetGlobalMaps()
	})
}
