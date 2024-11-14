package fault

import (
	"clusterd/pkg/application/job"
	"fmt"
	"strings"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
)

func getNodeAndDeviceFromJobIdAndRankId(jobId, rankId string, jobServerInfoMap job.JobServerInfoMap) (string, string, error) {

	for _, server := range jobServerInfoMap.InfoMap[jobId] {
		for _, dev := range server.DeviceList {
			if dev.RankID == rankId {
				return server.ServerName, dev.DeviceID, nil
			}
		}
	}
	return "", "", fmt.Errorf("not find node and device from jobId %v and rankid %v", jobId, rankId)
}

func getNodesNameFromDeviceInfo(deviceInfos map[string]*constant.DeviceInfo) []string {
	nodesName := make([]string, 0)
	for cmName, _ := range deviceInfos {
		nodeName, err := cmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		nodesName = append(nodesName, nodeName)
	}
	return nodesName
}

func cmNameToNodeName(cmName string) (string, error) {
	if !strings.HasPrefix(cmName, constant.DeviceInfoPrefix) {
		return "", fmt.Errorf("cmName has not prefix %s", constant.DeviceInfoPrefix)
	}
	return strings.TrimPrefix(cmName, constant.DeviceInfoPrefix), nil
}

func nodeNameToCmName(nodeName string) string {
	return constant.DeviceInfoPrefix + nodeName
}

func deviceID2DeviceKey(deviceID string) string {
	return constant.AscendDevPrefix + deviceID
}
