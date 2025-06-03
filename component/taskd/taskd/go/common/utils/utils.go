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

// Package utils is to provide go runtime utils
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/framework_backend/worker/monitor/profiling"
)

// InitHwLog init hwlog
func InitHwLog(ctx context.Context) error {
	var logFile string
	logFilePath := os.Getenv(constant.LogFilePathEnv)
	logFileName := "taskd-worker-" + strconv.Itoa(profiling.GlobalRankId) + ".log"
	if logFilePath == "" {
		logFile = constant.DefaultLogFilePath + logFileName
	} else {
		logFile = filepath.Join(logFilePath, logFileName)
	}
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   logFile,
		LogLevel:      constant.DefaultLogLevel,
		MaxBackups:    constant.DefaultMaxBackups,
		MaxAge:        constant.DefaultMaxAge,
		MaxLineLength: constant.DefaultMaxLineLength,
		// do not print to screen to avoid influence training log
		OnlyToFile: true,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}
	return nil
}

func marshalData(data interface{}) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data err: %v", err)
		return nil
	}
	return dataBuffer
}

// ObjToString obj to string
func ObjToString(data interface{}) string {
	var dataBuffer []byte
	if dataBuffer = marshalData(data); len(dataBuffer) == 0 {
		return ""
	}
	return string(dataBuffer)
}

// GetOnesDigit get code ones digit num
func GetOnesDigit(code int32) int32 {
	return code % constant.Ten
}

// GetTensDigit get code tens digit num
func GetTensDigit(code int32) int32 {
	return code / constant.Ten % constant.Ten
}

// GetThousandsAndHundreds get thousands and hundreds num
func GetThousandsAndHundreds(code int32) int32 {
	return code / constant.Hundred * constant.Hundred
}

// CopyStringMap copy string map
func CopyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
