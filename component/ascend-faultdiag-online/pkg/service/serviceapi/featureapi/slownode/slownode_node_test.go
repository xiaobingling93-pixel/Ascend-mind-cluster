package slownode

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/model/feature/slownode"
)

func TestConvertMaptoStruct(t *testing.T) {
	resorce := `{"slownode_default-test-pytorch-2pod-16npu":{"7.218.79.246":{"isSlow":0,"degradationLevel":"0.0%","jobName":"default-test-pytorch-2pod-16npu","nodeRank":"7.218.79.246","slowCalculateRanks":null,"slowCommunicationDomains":null,"slowSendRanks":null,"slowHostNodes":null,"slowIORanks":null}}}`
	var data = map[string]map[string]any{}
	err := json.Unmarshal([]byte(resorce), &data)
	assert.Nil(t, err)
	var result = &slownode.NodeSlowNodeAlgoResult{}
	err = convertMaptoStruct(data, result)
	assert.Nil(t, err)
	assert.Equal(t, "default-test-pytorch-2pod-16npu", result.JobName)
}
