package fault

import (
	"sync"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
)

// FaultRank fault rank info
type FaultRank struct {
	RankId    string
	FaultCode string
}

// JobFaultInfo job fault rank info
type JobFaultInfo struct {
	JobId     string
	FaultList []FaultRank
}

type jobRankFaultInfoProcessor struct {
	deviceCenter    *deviceFaultProcessCenter
	jobFaultInfoMap map[string]JobFaultInfo
	mutex           sync.RWMutex
}

func newJobRankFaultInfoProcessor(deviceCenter *deviceFaultProcessCenter) *jobRankFaultInfoProcessor {
	return &jobRankFaultInfoProcessor{
		jobFaultInfoMap: make(map[string]JobFaultInfo),
		deviceCenter:    deviceCenter,
		mutex:           sync.RWMutex{},
	}
}

func (processor *jobRankFaultInfoProcessor) getJobFaultRankInfos() map[string]JobFaultInfo {
	return processor.jobFaultInfoMap
}

func (processor *jobRankFaultInfoProcessor) setJobFaultRankInfos(faultInfos map[string]JobFaultInfo) {
	processor.jobFaultInfoMap = faultInfos
}

func (processor *jobRankFaultInfoProcessor) process() {
	deviceInfos := processor.deviceCenter.getInfoMap()
	nodesName := getNodesNameFromDeviceInfo(deviceInfos)
	jobFaultInfos := make(map[string]JobFaultInfo)
	jobServerInfoMap := processor.deviceCenter.jobServerInfoMap
	for jobId, serverList := range jobServerInfoMap.InfoMap {
		jobFaultInfo := JobFaultInfo{
			JobId:     jobId,
			FaultList: make([]FaultRank, 0),
		}

		for _, nodeName := range nodesName {
			faultRankList := processor.findFaultRankForJob(deviceInfos, nodeName, serverList)
			jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, faultRankList...)
		}
		jobFaultInfos[jobId] = jobFaultInfo
	}
	processor.setJobFaultRankInfos(jobFaultInfos)
}

func (processor *jobRankFaultInfoProcessor) findFaultRankForJob(deviceInfos map[string]*constant.DeviceInfo, nodeName string,
	serverList map[string]job.ServerHccl) []FaultRank {
	faultMap := device.GetFaultMap(deviceInfos[nodeNameToCmName(nodeName)])
	devicesOfJobOnNode, ok := serverList[nodeName]
	faultRankList := make([]FaultRank, 0)
	if !ok || len(devicesOfJobOnNode.DeviceList) == 0 {
		return faultRankList
	}
	for _, deviceInfo := range devicesOfJobOnNode.DeviceList {
		deviceName := deviceID2DeviceKey(deviceInfo.DeviceID)
		faultList, ok := faultMap[deviceName]
		if !ok {
			continue
		}
		for _, fault := range faultList {
			faultRankList = append(faultRankList, FaultRank{
				RankId:    deviceInfo.RankID,
				FaultCode: fault.FaultCode,
			})
		}
	}
	return faultRankList
}
