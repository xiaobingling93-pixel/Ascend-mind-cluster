package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/kube"
	"testing"
)

func TestJobRankFaultInfoProcessor_GetJobFaultRankInfos(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getJobFaultRankProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}

	t.Run("TestJobRankFaultInfoProcessor_GetJobFaultRankInfos", func(t *testing.T) {
		cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err := readObjectFromJobFaultRankTestYaml()
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.setInfoMap(cmDeviceInfos)
		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor.deviceCenter.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
		processor.process()
		if !isFaultRankMapEqual(processor.jobFaultInfoMap, expectFaultRanks) {
			t.Errorf("processor.jobFaultInfos = %s, expectFaultRanks = %s",
				util.ObjToString(processor.jobFaultInfoMap), util.ObjToString(expectFaultRanks))
		}
	})
}
