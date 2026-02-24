/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package manualfault is main test for process manually separate faults
package manualfault

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/manualfault"
)

var testErr = errors.New("test error")

const (
	node1 = "node1"
	node2 = "node2"

	dev0 = "Ascend910-0"
	dev1 = "Ascend910-1"
	dev2 = "Ascend910-2"
	dev5 = "Ascend910-5"
	dev6 = "Ascend910-6"

	code0 = "code0"
	code1 = "code1"
	code2 = "code2"

	len0 = 0
	len1 = 1
	len2 = 2
	len3 = 3
	len5 = 5

	receiveTime0 = 1770969600000 // 2026-02-13 08:00:00
	receiveTime1 = 1771059600000 // 2026-02-14 09:00:00

	podName1               = "pod1"
	podName2               = "pod2"
	podName3               = "pod3"
	podName4               = "pod4"
	podNameSpace1          = "default"
	defaultPodRankIndexKey = "0"
	podDeviceKey0          = `{"server_id":"127.0.0.1","devices":[{"device_id":"0"}]}`
	podDeviceKey2          = `{"server_id":"127.0.0.1","devices":[{"device_id":"2"}]}`
	podDeviceKey5          = `{"server_id":"127.0.0.1","devices":[{"device_id":"5"}]}`
	podDeviceKey6          = `{"server_id":"127.0.0.1","devices":[{"device_id":"6"}]}`

	job1     = "123"
	job2     = "456"
	jobName1 = "job1"
	vcJobKey = "job"
	pgName1  = "pg1"

	podGroupKey  = "scheduling.k8s.io/group-name"
	vcJobNameKey = "volcano.sh/job-name"
)

var (
	oriDevInfo1    = make(map[string]*constant.DeviceInfo)
	expDeviceInfo1 = make(map[string]*constant.DeviceInfo)
	oriDevInfo2    = make(map[string]*constant.DeviceInfo)

	podDemo1 *v1.Pod
	podDemo2 *v1.Pod
	podDemo3 *v1.Pod
	podDemo4 *v1.Pod

	devInfoMap = map[string]string{dev0: podDeviceKey0, dev2: podDeviceKey2, dev5: podDeviceKey5, dev6: podDeviceKey6}
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		panic(err)
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	if err := initLog(); err != nil {
		return err
	}
	if err := initTestDataFromYaml(); err != nil {
		return err
	}
	return nil
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func initTestDataFromYaml() error {
	const maxFileSize = 10000
	var testDataPath = "../../../../../testdata/resource/manual_fault_processor_test.yaml"

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		return errors.New("test data path is invalid")
	}
	if fileInfo.Size() > int64(maxFileSize) {
		return errors.New("test data path size is invalid")
	}
	open, err := os.Open(testDataPath)
	if err != nil {
		return errors.New("open test data file failed")
	}
	defer open.Close()
	decoder := yaml.NewYAMLOrJSONDecoder(open, maxFileSize)
	if err = decoder.Decode(&oriDevInfo1); err != nil {
		return errors.New("decode oriDevInfo1 failed")
	}
	if err = decoder.Decode(&expDeviceInfo1); err != nil {
		return errors.New("decode expDeviceInfo1 failed")
	}
	if err = decoder.Decode(&oriDevInfo2); err != nil {
		return errors.New("decode oriDevInfo2 failed")
	}
	return nil
}

func getDemoNodeInfo() map[string]manualfault.NodeCmInfo {
	return map[string]manualfault.NodeCmInfo{
		node1: {
			Total: []string{dev0, dev1},
			Detail: map[string][]manualfault.DevCmInfo{
				dev0: {
					{
						FaultCode:        code0,
						FaultLevel:       constant.ManuallySeparateNPU,
						LastSeparateTime: receiveTime0,
					},
				},
				dev1: {
					{
						FaultCode:        code1,
						FaultLevel:       constant.ManuallySeparateNPU,
						LastSeparateTime: receiveTime1,
					},
				},
			},
		},
		node2: {
			Total: []string{dev1},
			Detail: map[string][]manualfault.DevCmInfo{
				dev1: {
					{
						FaultCode:        code1,
						FaultLevel:       constant.ManuallySeparateNPU,
						LastSeparateTime: receiveTime1,
					},
				},
			},
		},
	}
}

func getDemoPod(nodeName, podName, devName, jobUid string) *v1.Pod {
	uid, err := generateRandomString(len5)
	if err != nil {
		return podDemo3
	}
	p := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNameSpace1,
			UID:       types.UID(uid),
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
		},
		Status: v1.PodStatus{},
	}

	isControlle := true
	owner := metav1.OwnerReference{
		Name:       jobName1,
		Controller: &isControlle,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid)}
	p.SetOwnerReferences([]metav1.OwnerReference{owner})
	annotation := map[string]string{
		podGroupKey:          pgName1,
		api.PodRankIndexAnno: defaultPodRankIndexKey,
		api.Pod910DeviceAnno: devInfoMap[devName],
	}
	p.SetAnnotations(annotation)
	label := map[string]string{
		vcJobNameKey: jobName1,
	}
	p.SetLabels(label)
	return p
}

func generateRandomString(length int) (string, error) {
	chars := "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	charsLen := big.NewInt(int64(len(chars)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, charsLen)
		if err != nil {
			return "", err
		}
		result[i] = chars[n.Int64()]
	}

	return string(result), nil
}
