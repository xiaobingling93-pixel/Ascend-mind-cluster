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
Package eventapi provides the self-defined api of diag event
*/
package eventapi

import (
	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/model/diagmodel"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/service/servicecore"
)

const apiDiagEvent = "diag"

// GetDiagEventApi 获取添加指标的api
func GetDiagEventApi() *servicecore.Api {
	return servicecore.BuildApi(apiDiagEvent, &diagmodel.DiagModel{}, apiDiagEventFunc, nil)
}

// apiDiagEventFunc 诊断事件
func apiDiagEventFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *request.Context, model *diagmodel.DiagModel) error {
	diagCtx.StartDiag(ctxData)
	return nil
}
