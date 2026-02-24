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

// Package manualfault test for processing manual separate npu info
package manualfault

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
)

const (
	node1 = "node1"
	node2 = "node2"

	dev1 = "dev1"
	dev2 = "dev2"
	dev3 = "dev3"

	code1 = "code1"

	len0 = 0
	len1 = 1
	len2 = 2

	receiveTime1 = 1771059600000 // 2026-02-14 09:00:00
	receiveTime3 = 1771059620000 // 2026-02-14 09:00:20
	receiveTime4 = 1771059630000 // 2026-02-14 09:00:30
)

func getDemoNodeInfo() map[string]manualfault.NodeCmInfo {
	return map[string]manualfault.NodeCmInfo{
		node1: {
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

func getDemoCm() *v1.ConfigMap {
	info := getDemoNodeInfo()
	data := manualfault.ConvertNodeInfoToCmData(info)
	cm := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Data:       data,
	}
	return cm
}

func TestCheckDiffAndDelete(t *testing.T) {
	convey.Convey("test func checkDiffAndDelete success", t, testCheckDiffAndDelete)
	convey.Convey("test func checkDiffAndDelete failed, manual cm is nil", t, testErrGetCm)
}

func testCheckDiffAndDelete() {
	prepareData()
	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	manualfault.LastCmInfo = faultCmInfo
	p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, getDemoCm(), nil)
	defer p1.Reset()
	checkDiffAndDelete()
	faultCmInfo2, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo2), convey.ShouldEqual, len1)
	info, ok := faultCmInfo2[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1})
	info, ok = faultCmInfo2[node2]
	convey.So(ok, convey.ShouldBeFalse)
}

func testErrGetCm() {
	manualfault.InitFaultCmInfo()
	p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, nil, testErr)
	defer p1.Reset()
	checkDiffAndDelete()
	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo), convey.ShouldEqual, len0)
}

const (
	defaultFaultWindowHours = 24
	defaultFaultThreshold   = 3
	defaultFaultFreeHours   = 48
)

var validPolicy = conf.ManuallySeparatePolicy{
	Enabled: true,
	Separate: struct {
		FaultWindowHours int `yaml:"fault_window_hours"`
		FaultThreshold   int `yaml:"fault_threshold"`
	}{
		FaultWindowHours: defaultFaultWindowHours,
		FaultThreshold:   defaultFaultThreshold,
	},
	Release: struct {
		FaultFreeHours int `yaml:"fault_free_hours"`
	}{
		FaultFreeHours: defaultFaultFreeHours,
	},
}

func TestRelease(t *testing.T) {
	convey.Convey("test func release success", t, testRelease)
	convey.Convey("test func release failed, deep copy error", t, testErrDeepCp)
	convey.Convey("test func release failed, close release switch", t, testCloseRelease)
}

func testRelease() {
	testTime := time.Date(2026, 02, 16, 9, 0, 5, 0, time.UTC)
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&time.Time{}), "UnixMilli", func(_ *time.Time) int64 {
		return testTime.UnixMilli()
	})
	defer patch.Reset()
	conf.SetManualSeparatePolicy(validPolicy)
	prepareData()

	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo), convey.ShouldEqual, len2)
	info, ok := faultCmInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})

	release()
	faultCmInfo2, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo2), convey.ShouldEqual, len2)
	info, ok = faultCmInfo2[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev2})
	info, ok = faultCmInfo2[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
}

func testErrDeepCp() {
	conf.SetManualSeparatePolicy(validPolicy)
	prepareData()

	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo), convey.ShouldEqual, len2)
	info, ok := faultCmInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
	p1 := gomonkey.ApplyFuncReturn(util.DeepCopy, testErr)

	release()
	p1.Reset()
	faultCmInfo2, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo2), convey.ShouldEqual, len2)
	info, ok = faultCmInfo2[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo2[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
}

func testCloseRelease() {
	closeRelease := validPolicy
	closeRelease.Enabled = false
	conf.SetManualSeparatePolicy(closeRelease)
	prepareData()

	faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo), convey.ShouldEqual, len2)
	info, ok := faultCmInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
	p1 := gomonkey.ApplyFuncReturn(util.DeepCopy, testErr)

	release()
	p1.Reset()
	faultCmInfo2, err := manualfault.FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faultCmInfo2), convey.ShouldEqual, len2)
	info, ok = faultCmInfo2[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	info, ok = faultCmInfo2[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev3})
}

func prepareData() {
	manualfault.InitFaultCmInfo()
	fault1 := manualfault.FaultInfo{
		NodeName:    node1,
		DevName:     dev1,
		FaultCode:   code1,
		ReceiveTime: time.Now().Add(-(defaultFaultFreeHours + 1) * time.Hour).UnixMilli(), // before 49h
	}
	fault2 := manualfault.FaultInfo{
		NodeName:    node1,
		DevName:     dev2,
		FaultCode:   code1,
		ReceiveTime: time.Now().Add(-(defaultFaultFreeHours - 1) * time.Hour).UnixMilli(), // before 47h
	}
	fault3 := manualfault.FaultInfo{
		NodeName:    node2,
		DevName:     dev3,
		FaultCode:   code1,
		ReceiveTime: time.Now().Add(-(defaultFaultFreeHours - 1) * time.Hour).UnixMilli(), // before 47h
	}
	manualfault.FaultCmInfo.AddSeparateDev(fault1)
	manualfault.FaultCmInfo.AddSeparateDev(fault2)
	manualfault.FaultCmInfo.AddSeparateDev(fault3)
}

func TestProcess(t *testing.T) {
	const processInterval = 500 * time.Millisecond
	convey.Convey("test func ProcessManuSep success", t, func() {
		var hasExecuted bool
		var p1 = gomonkey.ApplyFunc(manualfault.UpdateOrCreateManualCm, func() {
			hasExecuted = true
			return
		})
		defer p1.Reset()
		ctx, cancel := context.WithCancel(context.TODO())
		go ProcessManuSep(ctx)
		time.Sleep(processInterval)
		cancel()
		convey.So(hasExecuted, convey.ShouldBeFalse)
	})
}

func TestLoadManualCmInfo(t *testing.T) {
	convey.Convey("test func LoadManualCmInfo success", t, func() {
		manualfault.InitFaultCmInfo()
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, getDemoCm(), nil)
		defer p1.Reset()
		LoadManualCmInfo()
		faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(faultCmInfo, convey.ShouldResemble, getDemoNodeInfo())
	})
	convey.Convey("test func LoadManualCmInfo success, cm is nil", t, func() {
		manualfault.InitFaultCmInfo()
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, nil, nil)
		defer p1.Reset()
		LoadManualCmInfo()
		faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(faultCmInfo), convey.ShouldEqual, 0)
	})
	convey.Convey("test func LoadManualCmInfo failed, get manual cm failed", t, func() {
		manualfault.InitFaultCmInfo()
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, nil, testErr)
		defer p1.Reset()
		LoadManualCmInfo()
		faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(faultCmInfo), convey.ShouldEqual, len0)
	})
	convey.Convey("test func LoadManualCmInfo failed, parse cm info failed", t, func() {
		manualfault.InitFaultCmInfo()
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, getDemoCm(), nil).
			ApplyFuncReturn(manualfault.ParseManualCm, nil, testErr)
		defer p1.Reset()
		LoadManualCmInfo()
		faultCmInfo, err := manualfault.FaultCmInfo.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(faultCmInfo), convey.ShouldEqual, len0)
	})
}
