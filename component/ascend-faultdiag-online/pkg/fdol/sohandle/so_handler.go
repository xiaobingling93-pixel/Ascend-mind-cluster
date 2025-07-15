/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package sohandle 包含了加载和使用动态链接库（.so）的功能。
*/
package sohandle

/*
#include <dlfcn.h>
#include <stdlib.h>
#include <string.h>

// 定义函数指针类型
typedef int (*execute_func_t)(const char* input, int input_length, char* output, int output_length);
typedef const char* (*get_type_func_t)();
typedef const char* (*get_version_func_t)();

const char* callGetType(get_type_func_t func) {
    if (func == NULL) {
        return NULL;
    }
	return func();
}

const char* callVersionType(get_version_func_t func) {
    if (func == NULL) {
        return NULL;
    }
	return func();
}

const int callExecute(execute_func_t func, char* input,int input_length, char* output, int output_length) {
    if (func == NULL) {
		//非零表示调用失败
        return -1;
    }
	return func(input, input_length, output, output_length);
}


*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unsafe"
)

// 定义SO文件常量
const (
	getType      = "getType"
	getVersion   = "getVersion"
	execute      = "execute"
	soFileSuffix = ".so"
	maxFileCount = 100
)

var dlcloseMutex sync.Mutex

// SoHandler 结构体，用于管理动态链接库的句柄、类型以及主执行函数。
type SoHandler struct {
	SoHandle    unsafe.Pointer                                 // .so 文件句柄
	SoType      string                                         // .so 文件类型
	SoVersion   string                                         // .so 的版本
	ExecuteFunc func(input []byte, output []byte) (int, error) // .so 文件中的主执行函数
}

// NewSoHandler 创建一个新的 SoHandler
func NewSoHandler(soPath string) (*SoHandler, error) {
	// 加载 .so 文件
	cs := C.CString(soPath)
	handle := C.dlopen(cs, C.RTLD_LAZY)
	defer C.free(unsafe.Pointer(cs))

	if handle == nil {
		return nil, fmt.Errorf("failed to load .so file: %s", soPath)
	}

	// 获取 .so 文件类型
	soType, err := getSoType(handle, soPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get .so file type: %v", err)
	}

	// 获取 .so 文件类型
	soVersion, err := getSoVersion(handle, soPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get .so file version: %v", err)
	}

	// 获取主执行函数
	executeFunc, err := getExecuteFunc(handle, soPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get execute function: %v", err)
	}

	return &SoHandler{
		SoHandle:    handle,
		SoType:      soType,
		SoVersion:   soVersion,
		ExecuteFunc: executeFunc,
	}, nil
}

// getSoType 获取 .so 文件类型
func getSoType(handle unsafe.Pointer, soPath string) (string, error) {
	// 获取 getType 函数指针
	cs := C.CString(getType)
	getTypeFunc := C.dlsym(handle, cs)
	defer C.free(unsafe.Pointer(cs))

	if getTypeFunc == nil {
		return "", fmt.Errorf("failed to find getType function in %s", soPath)
	}
	// 将函数地址转换为具体的函数类型
	typeFunc := C.get_type_func_t(getTypeFunc)
	// 调用函数获取类型
	typeName := C.GoString(C.callGetType(typeFunc))
	if typeName == "" {
		return "", fmt.Errorf("call [%s] func [%s] failed", soPath, "getType")
	}
	return typeName, nil
}

// getSoVersion 获取 .so 版本
func getSoVersion(handle unsafe.Pointer, soPath string) (string, error) {
	// 获取 getType 函数指针
	cs := C.CString(getVersion)
	getVersionFunc := C.dlsym(handle, cs)
	defer C.free(unsafe.Pointer(cs))

	if getVersionFunc == nil {
		return "", fmt.Errorf("failed to find getVersion function in %s", soPath)
	}
	// 将函数地址转换为具体的函数类型
	versionFunc := C.get_version_func_t(getVersionFunc)
	// 调用函数获取类型
	version := C.GoString(C.callVersionType(versionFunc))
	if version == "" {
		return "", fmt.Errorf("call [%s] func [%s] failed", soPath, "getType")
	}
	return version, nil
}

// getExecuteFunc 获取主执行函数
func getExecuteFunc(handle unsafe.Pointer, soPath string) (func(input []byte, output []byte) (int, error), error) {
	cs := C.CString(execute)
	// 获取 execute 函数指针
	executeFunc := C.dlsym(handle, cs)
	defer C.free(unsafe.Pointer(cs))

	if executeFunc == nil {
		return nil, fmt.Errorf("failed to find execute function in %s", soPath)
	}
	return func(input []byte, output []byte) (int, error) {
		// 检查输入和输出是否为空
		if len(input) == 0 || len(output) == 0 {
			return -1, errors.New("input or output buffer is empty")
		}
		// 获取输入和输出的首地址
		cInput := (*C.char)(unsafe.Pointer(&input[0]))
		cOutput := (*C.char)(unsafe.Pointer(&output[0]))

		// 获取输入和输出的长度
		inputLength := C.int(len(input))
		outputLength := C.int(len(output))

		f := C.execute_func_t(executeFunc)
		ret := C.callExecute(f, cInput, inputLength, cOutput, outputLength)
		if ret != 0 {
			return -1, fmt.Errorf("call [%s] func [%s] failed, return code [%d]", soPath, execute, ret)
		}
		return 0, nil
	}, nil
}

// Close 释放 .so 文件句柄
func (h *SoHandler) Close() error {
	dlcloseMutex.Lock()
	defer dlcloseMutex.Unlock()

	if h.SoHandle != nil {
		C.dlclose(h.SoHandle)
	}
	return nil
}

// 筛选 .so 文件的函数
func filterSoFiles(soDir string) ([]string, error) {
	var soFiles []string
	var fileCount int
	// 使用 filepath.Walk 递归遍历目录
	err := filepath.Walk(soDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileCount > maxFileCount {
			return fmt.Errorf("reach the max file count(%d)", maxFileCount)
		}
		// 检查是否为普通文件且扩展名是 .so
		if !info.IsDir() && filepath.Ext(info.Name()) == soFileSuffix {
			soFiles = append(soFiles, path)
		}
		fileCount++
		return nil
	})
	return soFiles, err
}

// GenerateSoHandlerMap 生成 .so 文件句柄映射表
func GenerateSoHandlerMap(soDir string) (map[string]*SoHandler, error) {
	soFiles, err := filterSoFiles(soDir)
	if err != nil {
		return nil, err
	}
	soHandlerMap := make(map[string]*SoHandler)
	for _, soFile := range soFiles {
		soHandler, err := NewSoHandler(soFile)
		if err != nil {
			return nil, err
		}
		soHandlerMap[soHandler.SoType] = soHandler
	}
	return soHandlerMap, nil
}
