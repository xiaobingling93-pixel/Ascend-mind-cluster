// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package busconfig business configuration service for grpc client
package busconfig

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/interface/grpc/config"
)

const (
	publisherTestWaitTime = 20 * time.Millisecond
)

func fakePublisher() *ConfigPublisher {
	ctx := context.Background()
	publisher := NewConfigPublisher(job1, ctx)
	return publisher
}

// TestSelectStreamAndContext for test selectStreamAndContext
func TestSelectStreamAndContext(t *testing.T) {
	convey.Convey("test selectStreamAndContext", t, func() {
		convey.Convey("01-context canceled, should shutdown work queue", func() {
			publisher := fakePublisher()
			runSuccess := false
			go func() {
				publisher.selectStreamAndContext(&mockConfigSubscribeRankTableServer{})
				runSuccess = true
			}()
			publisher.ctxCancelFunc()
			time.Sleep(publisherTestWaitTime)
			convey.So(runSuccess, convey.ShouldBeTrue)
		})
		convey.Convey("02-stream context canceled, should shutdown work queue", func() {
			publisher := fakePublisher()
			runSuccess := false
			ctx, cancel := context.WithCancel(context.Background())
			patch := gomonkey.ApplyMethodReturn(&mockConfigSubscribeRankTableServer{}, "Context", ctx)
			defer patch.Reset()
			go func() {
				publisher.selectStreamAndContext(&mockConfigSubscribeRankTableServer{})
				runSuccess = true
			}()
			cancel()
			time.Sleep(publisherTestWaitTime)
			convey.So(runSuccess, convey.ShouldBeTrue)
		})
	})
}

// TestSendRankTable for test sendRankTable
func TestSendRankTable(t *testing.T) {
	convey.Convey("test sendRankTable", t, func() {
		convey.Convey("01-work queue shut down, should not send rank table", func() {
			publisher := fakePublisher()
			runSuccess := true
			go func() {
				runSuccess = publisher.sendRankTable(&mockConfigSubscribeRankTableServer{})
			}()
			publisher.rankTableQue.ShutDown()
			time.Sleep(publisherTestWaitTime)
			convey.So(runSuccess, convey.ShouldBeFalse)
		})
		convey.Convey("02-send rank table success", func() {
			publisher := fakePublisher()
			runSuccess := false
			table := &config.RankTableStream{
				JobId:     job1,
				RankTable: rankTable,
			}
			patch := gomonkey.ApplyMethodReturn(&mockConfigSubscribeRankTableServer{}, "Send", nil)
			defer patch.Reset()
			publisher.rankTableQue.Add(table)
			runSuccess = publisher.sendRankTable(&mockConfigSubscribeRankTableServer{})
			convey.So(runSuccess, convey.ShouldBeTrue)
		})
	})
}

// TestSaveData for test SaveData
func TestSaveData(t *testing.T) {
	convey.Convey("test SaveData", t, func() {
		publisher := fakePublisher()
		publisher.SaveData(job1, rankTable)
		element, isShutdown := publisher.rankTableQue.Get()
		convey.So(isShutdown, convey.ShouldBeFalse)
		data, ok := element.(*config.RankTableStream)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(data.RankTable, convey.ShouldEqual, rankTable)
	})
}

// TestStop for test stop
func TestStop(t *testing.T) {
	convey.Convey("test stop", t, func() {
		publisher := fakePublisher()
		isShutDown := false
		go func() {
			_, isShutDown = publisher.rankTableQue.Get()
		}()
		publisher.stop()
		time.Sleep(publisherTestWaitTime)
		convey.So(isShutDown, convey.ShouldBeTrue)
	})
}
