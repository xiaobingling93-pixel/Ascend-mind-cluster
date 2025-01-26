package faultrank

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	v1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/kube"
)

const (
	JobId    = "Job"
	NodeName = "Node"
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../../testdata/clusterd.log"}
	hwLogConfig.MaxBackups = 30
	hwLogConfig.MaxAge = 7
	hwLogConfig.LogLevel = 0
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	m.Run()
}

func getDemoJobServerMap() constant.JobServerInfoMap {
	return constant.JobServerInfoMap{
		InfoMap: map[string]map[string]constant.ServerHccl{
			JobId: {
				NodeName: constant.ServerHccl{
					DeviceList: []constant.Device{{
						DeviceID: "0",
						RankID:   "0",
					}, {
						DeviceID: "1",
						RankID:   "1",
					}},
				},
			},
		},
	}
}

func TestFaultProcessorImplProcess(t *testing.T) {
	t.Run("test node fail, job fault rank list should correct", func(t *testing.T) {
		processor := NewFaultProcessor()
		jobServerMap := getDemoJobServerMap()
		mockKube := gomonkey.ApplyFunc(kube.GetNode, func(name string) *v1.Node {
			return nil
		})
		mockJob := gomonkey.ApplyFunc(job.GetJobServerInfoMap, func() constant.JobServerInfoMap {
			return jobServerMap
		})
		defer func() {
			mockKube.Reset()
			mockJob.Reset()
		}()
		processor.Process(constant.AllConfigmapContent{})
		rankProcessor := processor.JobRankFaultInfoProcessor
		faultRankInfos := rankProcessor.GetJobFaultRankInfos()
		if len(faultRankInfos[JobId].FaultList) != len(jobServerMap.InfoMap[JobId][NodeName].DeviceList) {
			t.Error("TestFaultProcessorImplProcess fail")
		}
	})
}
