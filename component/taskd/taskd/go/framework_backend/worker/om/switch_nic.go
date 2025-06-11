// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

/*
#include "switch.h"
*/
import "C"

import (
	"unsafe"

	"taskd/toolkit_backend/net"
)

// NetTool worker net tool
var NetTool *net.NetInstance

// SwitchNicCallback switch callback func
var switchNicCallback C.callbackfunc

func RegisterCallback(ptr uintptr) {
	switchNicCallback = (C.callbackfunc)(unsafe.Pointer(ptr))
}
