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

// Package topology for generate topology of Rack
package topology

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestTopoFileToStr(t *testing.T) {
	convey.Convey("test topoFileToStr", t, func() {
		convey.Convey("read file failed", func() {
			mock1 := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, nil, fmt.Errorf("mock error"))
			defer mock1.Reset()
			_, err := topoFileToStr("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		mock2 := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, nil, nil)
		defer mock2.Reset()
		convey.Convey("json valid failed", func() {
			mock3 := gomonkey.ApplyFuncReturn(json.Valid, false)
			defer mock3.Reset()
			_, err := topoFileToStr("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("json valid success", func() {
			mock3 := gomonkey.ApplyFuncReturn(json.Valid, true)
			defer mock3.Reset()
			_, err := topoFileToStr("")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestToFile test case for ToFile
func TestToFile(t *testing.T) {
	convey.Convey("Test ToFile function scenarios", t, func() {
		buildToFileTestCase1("")
		buildToFileTestCase2("")
	})
}

func buildToFileTestCase1(topoFilePath string) {
	convey.Convey("When get json string of topo info failed", func() {
		mock := gomonkey.ApplyFuncReturn(topoFileToStr, "", fmt.Errorf("mock error"))
		defer mock.Reset()
		err := ToFile(topoFilePath, "")
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("When get file stat failed (not not exist)", func() {
		mock := gomonkey.ApplyFuncReturn(topoFileToStr, "", nil).
			ApplyFuncReturn(os.Stat, nil, fmt.Errorf("mock error")).
			ApplyFuncReturn(os.IsNotExist, false)
		defer mock.Reset()
		err := ToFile(topoFilePath, "")
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("When get file stat success but get file hash failed", func() {
		mock := gomonkey.ApplyFuncReturn(topoFileToStr, "", nil).
			ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(getFileHash, [Sha256HashLength]byte{}, fmt.Errorf("mock error"))
		defer mock.Reset()
		err := ToFile(topoFilePath, "")
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("When new hash equals original hash (no write)", func() {
		mock := gomonkey.ApplyFuncReturn(topoFileToStr, "", nil).
			ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(getFileHash, [Sha256HashLength]byte{}, nil).
			ApplyFuncReturn(sha256.Sum256, [Sha256HashLength]byte{})
		defer mock.Reset()
		err := ToFile(topoFilePath, "")
		convey.So(err, convey.ShouldBeNil)
	})
}

func buildToFileTestCase2(topoFilePath string) {
	convey.Convey("When open file failed", func() {
		mock := gomonkey.ApplyFuncReturn(topoFileToStr, "", nil).
			ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(getFileHash, [Sha256HashLength]byte{}, nil).
			ApplyFuncReturn(os.OpenFile, nil, fmt.Errorf("mock error"))
		defer mock.Reset()
		err := ToFile(topoFilePath, "")
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("When write string to file failed", func() {
		mockFile := &os.File{}
		mock := gomonkey.ApplyFuncReturn(topoFileToStr, "", nil).
			ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(getFileHash, [Sha256HashLength]byte{}, nil).
			ApplyFuncReturn(os.OpenFile, mockFile, nil).
			ApplyMethodReturn(mockFile, "WriteString", 0, fmt.Errorf("mock error"))
		defer mock.Reset()
		err := ToFile(topoFilePath, "")
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("When chmod file failed", func() {
		mockFile := &os.File{}
		mock := gomonkey.ApplyFuncReturn(topoFileToStr, "", nil).
			ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(getFileHash, [Sha256HashLength]byte{}, nil).
			ApplyFuncReturn(os.OpenFile, mockFile, nil).
			ApplyMethodReturn(mockFile, "WriteString", 0, nil).
			ApplyFuncReturn(os.Chmod, fmt.Errorf("mock error"))
		defer mock.Reset()
		err := ToFile(topoFilePath, "")
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("When ToFile execute successfully", func() {
		mockFile := &os.File{}
		mock := gomonkey.ApplyFuncReturn(topoFileToStr, "", nil).
			ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(getFileHash, [Sha256HashLength]byte{}, nil).
			ApplyFuncReturn(os.OpenFile, mockFile, nil).
			ApplyMethodReturn(mockFile, "WriteString", 0, nil).
			ApplyFuncReturn(os.Chmod, nil)
		defer mock.Reset()
		err := ToFile(topoFilePath, "")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetFileHash(t *testing.T) {
	convey.Convey("test getFileHash err", t, func() {
		mockReadFile := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
			return []byte{}, errors.New("fake error")
		})
		defer mockReadFile.Reset()
		_, err := getFileHash("test1")
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("test getFileHash success", t, func() {
		mockReadFile := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
			return []byte{'a', 'b'}, nil
		})
		defer mockReadFile.Reset()
		_, err := getFileHash("test2")
		convey.So(err, convey.ShouldBeNil)
	})
}
