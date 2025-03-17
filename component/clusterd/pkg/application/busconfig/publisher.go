// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package busconfig business configuration service for grpc client
package busconfig

import (
	"context"
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/config"
)

const retryTimes = 3

// ConfigPublisher save the rank table and send it to the client
type ConfigPublisher struct {
	jobId          string
	rankTableQue   workqueue.Interface
	ctxContext     context.Context
	ctxCancelFunc  context.CancelFunc
	serviceContext context.Context
	lock           sync.RWMutex
}

// NewConfigPublisher create a config publisher
func NewConfigPublisher(jobId string, serviceCtx context.Context) *ConfigPublisher {
	publisher := &ConfigPublisher{
		jobId:          jobId,
		rankTableQue:   workqueue.New(),
		serviceContext: serviceCtx,
		lock:           sync.RWMutex{},
	}
	publisher.ctxContext, publisher.ctxCancelFunc = context.WithCancel(publisher.serviceContext)
	return publisher
}

func (c *ConfigPublisher) selectStreamAndContext(stream config.Config_SubscribeRankTableServer) {
	defer c.rankTableQue.ShutDown()
	for {
		select {
		case <-c.ctxContext.Done():
			hwlog.RunLog.Infof("context canceled, jobId=%s", c.jobId)
			return
		case <-stream.Context().Done():
			hwlog.RunLog.Warnf("stream is closed, do not send ranktable jobId=%s", c.jobId)
			return
		}
	}
}

func (c *ConfigPublisher) listenRankTableChange(stream config.Config_SubscribeRankTableServer) {
	hwlog.RunLog.Infof("start listen a new work queue, jobId=%s", c.jobId)
	go c.selectStreamAndContext(stream)
	for {
		if !c.sendRankTable(stream) {
			break
		}
	}
}

func (c *ConfigPublisher) sendRankTable(stream config.Config_SubscribeRankTableServer) bool {
	element, isShutdown := c.rankTableQue.Get()
	if isShutdown {
		hwlog.RunLog.Warnf("work queue shut down, do not send ranktable jobId=%s", c.jobId)
		return false
	}
	c.rankTableQue.Done(element)
	data, ok := element.(*config.RankTableStream)
	if !ok {
		hwlog.RunLog.Errorf("failed to assert element to *config.RankTableStream, jobId=%s", c.jobId)
		return true
	}
	for i := 0; i < retryTimes; i++ {
		err := stream.Send(data)
		if err == nil {
			hwlog.RunLog.Infof("send ranktable success, jobId=%s", data.JobId)
			break
		}
		hwlog.RunLog.Errorf("send ranktable failed, jobId=%s, error= %v", data.JobId, err)
		time.Sleep(time.Second)
	}
	return true
}

// SaveData save data to work queue
func (c *ConfigPublisher) SaveData(jobId, data string) {
	rankTable := &config.RankTableStream{
		JobId:     jobId,
		RankTable: data,
	}
	c.rankTableQue.Add(rankTable)
}

func (c *ConfigPublisher) stop() {
	hwlog.RunLog.Infof("jobId=%s enter publisher stop function", c.jobId)
	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
	}
	c.rankTableQue.ShutDown()
}
