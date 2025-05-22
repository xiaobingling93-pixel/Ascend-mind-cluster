/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package global is some variables from all
package global

import (
	"ascend-faultdiag-online/pkg/utils/configmap"
	"ascend-faultdiag-online/pkg/utils/grpc"
)

var (
	// K8sClient is a global k8s client
	K8sClient *configmap.ClientK8s
	// GrpcClient is a global grpc client
	GrpcClient *grpc.Client
)
