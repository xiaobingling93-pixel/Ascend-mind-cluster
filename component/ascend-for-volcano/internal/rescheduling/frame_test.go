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
Package rescheduling is using for HuaWei Ascend pin fault
*/
package rescheduling

import (
	"encoding/json"
	"sync"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	sliceIndexZero      = 0
	sliceIndexOne       = 1
	sliceIndexTwo       = 2
	sliceIndexThree     = 3
	sliceIndexFour      = 4
	fakeNodeName        = "node0"
	annoCards           = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"
	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
	unhealthyNPU        = "huawei.com/Ascend910-Unhealthy"
)

func fakeReSchedulerFaultTask(isFault bool, paras []string, podCreateTime int64) FaultTask {
	if len(paras) < test.NPUIndex5 {
		return FaultTask{}
	}
	name := paras[sliceIndexZero]
	ns := paras[sliceIndexOne]
	nodeName := paras[sliceIndexTwo]
	rankIndex := paras[sliceIndexFour]
	faultTask := FaultTask{
		IsFaultTask:   isFault,
		TaskUID:       api.TaskID(`"` + ns + `"-"` + name + `"`),
		TaskName:      name,
		TaskNamespace: ns,
		NodeName:      nodeName,
		NodeRankIndex: rankIndex,
		UseCardName:   []string{"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"},
		PodCreateTime: podCreateTime,
	}
	return faultTask
}

func fakeSchedulerJobEmptyTask(jobName, namespace string) plugin.SchedulerJob {
	job0 := plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{
				Name:      api.JobID(jobName),
				NameSpace: namespace,
				Selector:  map[string]string{util.AcceleratorType: util.ModuleAcceleratorType},
				Label: map[string]string{
					JobRescheduleLabelKey: JobGraceRescheduleLabelValue,
				},
			},
			NPUJob: &util.NPUJob{
				ReqNPUName: util.NPU910CardName,
				ReqNPUNum:  0,
				Tasks:      make(map[api.TaskID]util.NPUTask, util.NPUIndex2),
			},
		},
	}
	return job0
}

type PreStartActionTestCase struct {
	name    string
	ssn     *framework.Session
	env     *plugin.ScheduleEnv
	wantErr bool
}

func newDefaultHandler() *plugin.ScheduleHandler {
	isFirstSession := true
	scheduleHandler := &plugin.ScheduleHandler{
		NPUPlugins: map[string]plugin.NPUBuilder{},
		ScheduleEnv: plugin.ScheduleEnv{
			IsFirstSession:   &isFirstSession,
			Jobs:             map[api.JobID]plugin.SchedulerJob{},
			JobSeverInfos:    map[api.JobID]struct{}{},
			JobDeleteFlag:    map[api.JobID]struct{}{},
			JobSinglePodFlag: map[api.JobID]bool{},
			Nodes:            map[string]plugin.NPUNode{},
			DeviceInfos: &plugin.DeviceInfosWithMutex{
				Mutex:   sync.Mutex{},
				Devices: map[string]plugin.NodeDeviceInfoWithID{},
			},
			NodeInfosFromCm: &plugin.NodeInfosFromCmWithMutex{
				Mutex: sync.Mutex{},
				Nodes: map[string]plugin.NodeDNodeInfo{},
			},
			SwitchInfosFromCm: &plugin.SwitchInfosFromCmWithMutex{
				Mutex:    sync.Mutex{},
				Switches: map[string]plugin.SwitchFaultInfo{},
			},
			FrameAttr: plugin.VolcanoFrame{},
			NslbAttr:  &plugin.NslbParameters{},
			SuperPodInfo: &plugin.SuperPodInfo{
				SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
				SuperPodFaultTaskNodes:    map[api.JobID][]string{},
				SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{},
			},
			JobPendingMessage: map[api.JobID]map[string]map[string]struct{}{},
		},
	}
	scheduleHandler.NPUPlugins[util.NPU910CardName] = func(string2 string) plugin.ISchedulerPlugin {
		return nil
	}
	return scheduleHandler
}

func fakeScheduleEnv() *plugin.ScheduleEnv {
	sHandle := newDefaultHandler()
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	for _, jobInfo := range ssn.Jobs {
		test.SetJobStatusRunning(jobInfo)
	}
	for _, node := range ssn.Nodes {
		sHandle.UpdateConfigMap(fakeDeviceInfoCMDataByNode(node.Name, fakeDeviceList()), util.AddOperator)
	}
	sHandle.InitVolcanoFrameFromSsn(ssn)
	sHandle.InitNodesFromSsn(ssn)
	sHandle.InitJobsFromSsn(ssn)
	sHandle.ScheduleEnv.FrameAttr.KubeClient = fake.NewSimpleClientset()
	node0 := sHandle.ScheduleEnv.Nodes[fakeNodeName]
	node0.Annotation[unhealthyNPU] = annoCards
	sHandle.ScheduleEnv.Nodes[fakeNodeName] = node0
	return &sHandle.ScheduleEnv
}

func fakeDeviceInfoCMDataByNode(nodeName string, deviceList map[string]string) *v1.ConfigMap {
	cmName := util.DevInfoPreName + nodeName
	const testTime = 1657527526
	cmData := plugin.NodeDeviceInfoWithDevPlugin{
		DeviceInfo: plugin.NodeDeviceInfo{
			DeviceList: deviceList,
			UpdateTime: testTime,
		},
		CheckCode: "6b8de396fd9945be231d24720ca66ed950baf0a5972717f335aad7571cb6457a",
	}
	var data = make(map[string]string, 1)
	cmDataStr, err := json.Marshal(cmData)
	if err != nil {
		return nil
	}
	data["DeviceInfoCfg"] = string(cmDataStr)
	data[util.SwitchInfoCmKey] = fakeSwitchInfos()
	var faultNPUConfigMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: "kube-system",
		},
		Data: data,
	}
	return faultNPUConfigMap
}

func fakeDeviceList() map[string]string {
	return map[string]string{
		util.NPU910CardName: annoCards,
		networkUnhealthyNPU: "",
		unhealthyNPU:        "",
	}
}

func fakeSwitchInfos() string {
	tmpSwitchInfo := plugin.SwitchFaultInfo{
		NodeStatus: util.NodeHealthyByNodeD,
	}
	if bytes, err := json.Marshal(tmpSwitchInfo); err == nil {
		return string(bytes)
	}
	return ""
}

func buildPreStartActionTestCase01() PreStartActionTestCase {
	return PreStartActionTestCase{
		name:    "01 PreStartAction will return err when ssn is nil",
		ssn:     nil,
		env:     &plugin.ScheduleEnv{},
		wantErr: true,
	}
}

func buildPreStartActionTestCase02() PreStartActionTestCase {
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	return PreStartActionTestCase{
		name:    "02 PreStartAction will return nil when pre start is ok",
		ssn:     ssn,
		env:     fakeScheduleEnv(),
		wantErr: false,
	}
}

func buildPreStartActionTestCases() []PreStartActionTestCase {
	return []PreStartActionTestCase{
		buildPreStartActionTestCase01(),
		buildPreStartActionTestCase02(),
	}
}

func TestReSchedulerPreStartAction(t *testing.T) {
	tests := buildPreStartActionTestCases()
	reScheduler, ok := NewHandler().(*ReScheduler)
	if !ok {
		return
	}
	reSchedulerCache = newReSchedulerCache()
	reSchedulerCache.FaultNodes = map[string]*FaultNode{fakeNodeName: {}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := reScheduler.PreStartAction(tt.env, tt.ssn); (err != nil) != tt.wantErr {
				t.Errorf("PreStartAction() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.ssn == nil {
				return
			}
			reScheduler.SynCacheFaultJobWithSession(tt.ssn)
		})
	}
}
