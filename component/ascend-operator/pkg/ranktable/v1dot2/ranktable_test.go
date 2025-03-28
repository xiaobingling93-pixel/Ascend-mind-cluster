/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package v1dot2 is using for v1.2 Ranktable.
*/
package v1dot2

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
	_ "ascend-operator/pkg/testtool"
	"ascend-operator/pkg/utils"
)

func TestGatherServerList(t *testing.T) {
	convey.Convey("TestGatherServerList", t, func() {
		convey.Convey("01-servers will be set to serverList and SuperPodList", func() {
			job := &v1.AscendJob{}
			job.Annotations = map[string]string{utils.AnnoKeyOfSuperPod: "2"}
			gen := New(job)
			gen.ServerList = []*common.Server{
				{
					DeviceList: []*common.Device{
						{RankID: "0"},
						{RankID: "1"},
					},
					ServerID: "127.0.0.1",
				},
				{
					DeviceList: []*common.Device{
						{RankID: "2"},
						{RankID: "3"},
					},
					ServerID: "127.0.0.2",
				},
			}
			patch := gomonkey.ApplyMethod(new(common.BaseGenerator), "GatherServerList",
				func(*common.BaseGenerator) {})
			defer patch.Reset()
			gen.GatherServerList()
			expected := 2
			convey.So(len(gen.SuperPodList), convey.ShouldEqual, expected)
			convey.So(gen.SuperPodList[0].ServerList[0].ServerID, convey.ShouldEqual, "127.0.0.1")
		})
	})
}
