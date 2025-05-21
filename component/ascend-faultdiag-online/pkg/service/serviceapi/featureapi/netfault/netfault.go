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

// Package netfault for fault network feature
package netfault

import (
	"ascend-faultdiag-online/pkg/model/feature"
	"ascend-faultdiag-online/pkg/service/servicecore"
)

const (
	apiNetwork       = "netfault"
	apiController    = "controller"
	apiStart         = "start"
	apiStop          = "stop"
	apiReload        = "reload"
	registerCallBack = "registerCallBack"
	name             = "netfault"
	byteSize         = 1024
	cluster          = "cluster"
	inputCommand     = "command"
	inputTarget      = "target"
	inputFunc        = "func"
)

// GetNetFaultApi register /netfault api of feature
func GetNetFaultApi() *servicecore.Api {
	return servicecore.NewApi(apiNetwork, nil, []*servicecore.Api{
		GetControllerApi(),
	})
}

// GetControllerApi register /start、/stop、 /reload api of controller
func GetControllerApi() *servicecore.Api {
	return servicecore.NewApi(apiController, nil, []*servicecore.Api{
		// start
		servicecore.BuildApi(apiStart, &feature.Status{}, ControllerStartFunc, nil),
		// stop
		servicecore.BuildApi(apiStop, &feature.Status{}, ControllerStopFunc, nil),
		// reload
		servicecore.BuildApi(apiReload, &feature.Status{}, ControllerReloadFunc, nil),
	})
}
