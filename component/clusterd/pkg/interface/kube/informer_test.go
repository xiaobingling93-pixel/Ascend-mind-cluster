// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube test function
package kube

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/superpod"
)

var (
	testTwoDeviceFunc = 2
	testTwoNodeFunc   = 2
)

func TestStopInformer(t *testing.T) {
	convey.Convey("TestStopInformer", t, func() {
		convey.So(StopInformer, convey.ShouldNotPanic)
	})
}

func TestCleanFuncs(t *testing.T) {
	convey.Convey("TestCleanFuncs", t, func() {
		CleanFuncs()
		convey.So(len(cmDeviceFuncs), convey.ShouldEqual, 0)
		convey.So(len(cmNodeFuncs), convey.ShouldEqual, 0)
	})
}

func TestAddCmDeviceFunc(t *testing.T) {
	convey.Convey("TestAddCmDeviceFunc", t, func() {
		convey.Convey("add one device func", func() {
			AddCmDeviceFunc(constant.Resource, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs[constant.Resource]), convey.ShouldEqual, 1)
		})
		convey.Convey("add two device func", func() {
			AddCmDeviceFunc(constant.Resource, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs[constant.Resource]), convey.ShouldEqual, testTwoDeviceFunc)
		})
		convey.Convey("add two different business func", func() {
			AddCmDeviceFunc(constant.Statistics, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs), convey.ShouldEqual, testTwoDeviceFunc)
		})
	})
}

func TestAddCmNodeFunc(t *testing.T) {
	convey.Convey("TestAddCmNodeFunc", t, func() {
		convey.Convey("add one node func", func() {
			AddCmNodeFunc(constant.Resource, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs[constant.Resource]), convey.ShouldEqual, 1)
		})
		convey.Convey("add two node func", func() {
			AddCmNodeFunc(constant.Resource, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs[constant.Resource]), convey.ShouldEqual, testTwoNodeFunc)
		})
		convey.Convey("add two different business func", func() {
			AddCmNodeFunc(constant.Statistics, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs), convey.ShouldEqual, testTwoNodeFunc)
		})
	})
}

func TestCheckConfigMapIsDeviceInfo(t *testing.T) {
	convey.Convey("test checkConfigMapIsNodeInfo", t, func() {
		var obj interface{}
		mockMatchedFalse := gomonkey.ApplyFunc(util.IsNSAndNameMatched, func(obj interface{},
			namespace string, namePrefix string) bool {
			return false
		})
		defer mockMatchedFalse.Reset()
		cmCheck := checkConfigMapIsDeviceInfo(obj)
		convey.So(cmCheck, convey.ShouldBeFalse)
	})
}

func TestCheckConfigMapIsNodeInfo(t *testing.T) {
	convey.Convey("test checkConfigMapIsNodeInfo", t, func() {
		var obj interface{}
		mockMatchedTrue := gomonkey.ApplyFunc(util.IsNSAndNameMatched, func(obj interface{},
			namespace string, namePrefix string) bool {
			return true
		})
		defer mockMatchedTrue.Reset()
		nodeCheck := checkConfigMapIsNodeInfo(obj)
		convey.So(nodeCheck, convey.ShouldBeTrue)
	})
}

func TestInitClusterDevice(t *testing.T) {
	convey.Convey("Test initClusterDevice", t, func() {
		fakeNodes := []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-1",
				},
			},
		}
		patchGetNodes := gomonkey.ApplyFunc(getNodesFromInformer, func() []*v1.Node {
			return fakeNodes
		})
		defer patchGetNodes.Reset()

		patchGetNodeDevice := gomonkey.ApplyFunc(node.GetNodeDeviceAndSuperPodID,
			func(node *v1.Node) (*api.NodeDevice, string) {
				return &api.NodeDevice{
					NodeName: node.Name,
				}, "test-superpod-id"
			})
		defer patchGetNodeDevice.Reset()

		var calledSaveNode bool
		patchSaveNode := gomonkey.ApplyFunc(superpod.SaveNode,
			func(superPodID string, node *api.NodeDevice) {
				calledSaveNode = true
				convey.So(superPodID, convey.ShouldEqual, "test-superpod-id")
				convey.So(node.NodeName, convey.ShouldEqual, "test-node-1")
			})
		defer patchSaveNode.Reset()

		initClusterDevice()
		convey.So(calledSaveNode, convey.ShouldBeFalse)
	})
}

func TestConfigMapIsEpRankTableInfo(t *testing.T) {
	convey.Convey("test ConfigMapIsEpRankTableInfo", t, func() {
		convey.Convey("when object is not a ConfigMap", func() {
			obj := "not a configmap"
			result := checkConfigMapIsEpRankTableInfo(obj)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("when ConfigMap name has the correct prefix", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: constant.MindIeRanktablePrefix + "example",
				},
			}
			result := checkConfigMapIsEpRankTableInfo(cm)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("when ConfigMap name does not have the correct prefix", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wrong-prefix-example",
				},
			}
			result := checkConfigMapIsEpRankTableInfo(cm)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestAddPodGroupFunc(t *testing.T) {
	convey.Convey("Test AddPodGroupFunc", t, func() {
		business := "testBusiness"
		func1 := func(pg1, pg2 *v1beta1.PodGroup, s string) {}
		AddPodGroupFunc(business, func1)
		convey.So(len(podGroupFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestAddPodFunc(t *testing.T) {
	convey.Convey("Test AddPodFunc", t, func() {
		business := "testBusiness"
		func1 := func(p1, p2 *v1.Pod, s string) {}
		AddPodFunc(business, func1)
		convey.So(len(podFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestAddNodeFunc(t *testing.T) {
	convey.Convey("Test AddNodeFunc", t, func() {
		business := "testBusiness"
		func1 := func(n1, n2 *v1.Node, s string) {}
		AddNodeFunc(business, func1)
		convey.So(len(nodeFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestCheckConfigMapIsSwitchInfo(t *testing.T) {
	convey.Convey("Test checkConfigMapIsSwitchInfo", t, func() {
		patch := gomonkey.ApplyFuncReturn(util.IsNSAndNameMatched, true)
		value := checkConfigMapIsSwitchInfo(nil)
		convey.ShouldBeTrue(value)
		patch.Reset()

		patch = gomonkey.ApplyFuncReturn(util.IsNSAndNameMatched, false)
		value = checkConfigMapIsSwitchInfo(nil)
		convey.ShouldBeFalse(value)
		patch.Reset()
	})
}

func TestAddCmPubFaultFunc(t *testing.T) {
	convey.Convey("Test AddCmPubFaultFunc", t, func() {
		business := "testBusiness"
		func1 := func(pf1, pf2 *api.PubFaultInfo, s string) {}

		AddCmPubFaultFunc(business, func1)

		convey.So(len(cmPubFaultFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestAddCmRankTableFunc(t *testing.T) {
	convey.Convey("Test AddCmRankTableFunc", t, func() {
		business := "testBusiness"
		func1 := func(i1, i2 interface{}, s string) {}

		AddCmRankTableFunc(business, func1)

		convey.So(len(cmRankTableFuncs[business]), convey.ShouldEqual, 1)
	})
}
