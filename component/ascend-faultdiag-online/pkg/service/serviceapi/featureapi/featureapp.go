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
Package featureapi provides API
*/
package featureapi

import (
	"ascend-faultdiag-online/pkg/service/serviceapi/featureapi/netfault"
	"ascend-faultdiag-online/pkg/service/serviceapi/featureapi/slownode"
	"ascend-faultdiag-online/pkg/service/servicecore"
)

const apiApp = "feature"

// GetAppApi 获取指标相关api
func GetAppApi() *servicecore.Api {
	return servicecore.NewApi(apiApp, nil, []*servicecore.Api{
		slownode.GetSlowNodeApi(),
		netfault.GetNetFaultApi(),
	})
}
