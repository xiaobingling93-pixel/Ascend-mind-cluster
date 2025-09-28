// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

type fakeProcessor struct{}

func (f *fakeProcessor) Process(info any) any {
	return info
}

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

func TestBaseFaultCenterProcess(t *testing.T) {
	t.Run("TestBaseFaultCenterProcess", func(t *testing.T) {
		manager := cmmanager.DeviceCenterCmManager
		baseCenter := newBaseFaultCenter(manager, constant.DeviceProcessType)
		baseCenter.addProcessors([]constant.FaultProcessor{&fakeProcessor{}})
		notifyChan := make(chan int, 1)
		baseCenter.Register(notifyChan)
		baseCenter.Process()
		if baseCenter.GetProcessedCm() == nil {
			t.Errorf("TestBaseFaultCenterProcess failed")
		}
	})
}

func TestNotifySubscriber(t *testing.T) {
	testCases := []struct {
		name             string
		channelList      []chan int
		centerType       int
		expectLogWarning bool
	}{
		{name: "no subscribers",
			channelList: []chan int{},
			centerType:  1},
		{name: "one subscriber with buffer",
			channelList: []chan int{make(chan int, 1)},
			centerType:  constant.DeviceProcessType},
		{name: "multiple subscribers",
			channelList: []chan int{make(chan int, 1), make(chan int, 1)},
			centerType:  constant.NodeProcessType},
		{name: "nil channel in list",
			channelList: []chan int{nil, make(chan int, 1)},
			centerType:  constant.SwitchProcessType},
	}
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			baseCenter := &baseFaultCenter[constant.ConfigMapInterface]{
				centerType:           tc.centerType,
				subscribeChannelList: tc.channelList,
			}
			baseCenter.notifySubscriber()
			for _, ch := range tc.channelList {
				if ch == nil || len(ch) == 0 {
					continue
				}
				select {
				case val := <-ch:
					convey.So(val, convey.ShouldEqual, tc.centerType)
				default:
					convey.So(false, convey.ShouldBeTrue, "testNotifySubscriber failed")
				}
			}
		})
	}
}
