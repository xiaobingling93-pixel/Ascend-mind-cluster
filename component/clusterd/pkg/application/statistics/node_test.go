// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics test for statistic funcs about node
package statistics

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/statistics"
)

const (
	nodeSN   = "nodeSN"
	nodeName = "nodeName"
)

func TestUpdateNodeSNAndNameCache(t *testing.T) {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeName,
			Annotations: map[string]string{nodeAnnotation: nodeSN},
		},
	}

	convey.Convey("test func UpdateNodeSNAndNameCache, node is nil", t, func() {
		UpdateNodeSNAndNameCache(nil, nil, constant.AddOperator)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, add node when node does not exist", t, func() {
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
		UpdateNodeSNAndNameCache(nil, node, constant.AddOperator)
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 1)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, add node when node exist", t, func() {
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 1)
		UpdateNodeSNAndNameCache(nil, node, constant.AddOperator)
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 1)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, delete node when node exist", t, func() {
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 1)
		UpdateNodeSNAndNameCache(nil, node, constant.DeleteOperator)
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, delete node when node does not exist", t, func() {
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
		UpdateNodeSNAndNameCache(nil, node, constant.DeleteOperator)
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, invalid operator", t, func() {
		UpdateNodeSNAndNameCache(nil, node, "invalid operator")
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, label does not exist", t, testLabelNotExist)
}

func testLabelNotExist() {
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeName,
			Annotations: map[string]string{"label not exist": nodeSN},
		},
	}
	UpdateNodeSNAndNameCache(nil, node1, constant.AddOperator)
	convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
}
