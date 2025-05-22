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

// Package slownode is a DT collection for func in slownode_node
package slownode

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/model/feature/slownode"
)

func TestConvertMaptoStruct(t *testing.T) {
	resorce := `{"slownode_default-test-pytorch-2pod-16npu":{"7.218.79.246":{"isSlow":0,"degradationLevel":"0.0%",` +
		`"jobName":"default-test-pytorch-2pod-16npu","nodeRank":"7.218.79.246","slowCalculateRanks":null,` +
		`"slowCommunicationDomains":null,"slowSendRanks":null,"slowHostNodes":null,"slowIORanks":null}}}`
	var data = map[string]map[string]any{}
	err := json.Unmarshal([]byte(resorce), &data)
	assert.Nil(t, err)
	var result = &slownode.NodeSlowNodeAlgoResult{}
	err = convertMaptoStruct(data, result)
	assert.Nil(t, err)
	assert.Equal(t, "default-test-pytorch-2pod-16npu", result.JobName)
}
