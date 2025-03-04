// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package kube a series of kube test function
package kube

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testPatchCMName = "test-cm-name"
	testPatchCMNS   = "default"
	testKey1        = "key1"
	testKey2        = "key2"
	testKey3        = "key3"
	testValue1      = "value1"
	testValue2      = "value2"
	testValue3      = "value3"
)

func TestPatchCMData(t *testing.T) {
	createCM(t)
	convey.Convey("test func 'PatchCMData' success", t, func() {

		// data from null to 'key1:value1, key2:value2'
		patchData := map[string]string{testKey1: testValue1, testKey2: testValue2}
		newCM, err := PatchCMData(testPatchCMName, testPatchCMNS, patchData)
		expData := map[string]string{testKey1: testValue1, testKey2: testValue2}
		convey.So(newCM.Data, convey.ShouldResemble, expData)
		convey.So(err, convey.ShouldBeNil)

		// data from 'key1:value1, key2:value2' to 'key1:value1, key2:value1, key3:value3'
		patchData = map[string]string{testKey2: testValue1, testKey3: testValue3}
		newCM, err = PatchCMData(testPatchCMName, testPatchCMNS, patchData)
		expData = map[string]string{testKey1: testValue1, testKey2: testValue1, testKey3: testValue3}
		convey.So(newCM.Data, convey.ShouldResemble, expData)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func 'PatchCMData' failed, marshal data failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(json.Marshal, nil, testErr)
		defer p1.Reset()
		testData := map[string]string{testKey1: testValue1}
		_, err := PatchCMData(testPatchCMName, testPatchCMNS, testData)
		expErr := fmt.Errorf("marshal cm data failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func createCM(t *testing.T) {
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPatchCMName,
			Namespace: testPatchCMNS,
		},
		Data: nil,
	}
	_, err := CreateConfigMap(cm)
	if err != nil {
		t.Errorf("create test cm failed, error: %v", err)
	}
}
