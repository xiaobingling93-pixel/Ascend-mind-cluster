// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package busconfig business configuration service for grpc client
package busconfig

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/config"
)

const (
	waitTime = 100 * time.Millisecond
)

// BusinessConfigServer business config server
type BusinessConfigServer struct {
	serviceCtx      context.Context
	configPublisher map[string]*ConfigPublisher
	lock            sync.RWMutex
	config.UnimplementedConfigServer
}

// NewBusinessConfigServer create a business config server
func NewBusinessConfigServer(ctx context.Context) *BusinessConfigServer {
	server := &BusinessConfigServer{
		serviceCtx:      ctx,
		configPublisher: make(map[string]*ConfigPublisher),
		lock:            sync.RWMutex{},
	}
	return server
}

// rankTableChange The callback function when the rank table changes
func (c *BusinessConfigServer) rankTableChange(jobId, data string) (bool, error) {
	publisher, ok := c.getPublisher(jobId)
	if !ok {
		return true, errors.New("job not registered")
	}
	hwlog.RunLog.Infof("ranktable changed, jobId=%s", jobId)
	publisher.SaveData(jobId, data)
	return false, nil
}

// Register is task register service
func (c *BusinessConfigServer) Register(ctx context.Context, req *config.ClientInfo) (*config.Status, error) {
	hwlog.RunLog.Infof("business config service receive Register request, jobId=%s, role=%s",
		req.JobId, req.Role)
	publisher, ok := c.getPublisher(req.JobId)
	if ok && publisher != nil {
		publisher.stop()
		for {
			if _, ok = c.getPublisher(req.JobId); !ok {
				break
			}
			time.Sleep(waitTime)
		}
	}
	c.addPublisher(req.JobId)
	return &config.Status{Code: int32(common.OK), Info: "register success"}, nil
}

// SubscribeRankTable subscribe rank table from ClusterD
func (c *BusinessConfigServer) SubscribeRankTable(request *config.ClientInfo,
	stream config.Config_SubscribeRankTableServer) error {
	hwlog.RunLog.Infof("receive Subscribe ranktable request, jobId=%s, rule=%s",
		request.JobId, request.Role)
	publisher, ok := c.getPublisher(request.JobId)
	if !ok || publisher == nil {
		hwlog.RunLog.Warnf("jobId=%s not registered, role=%s", request.JobId, request.Role)
		return fmt.Errorf("jobId=%s not registered, role=%s", request.JobId, request.Role)
	}
	publisher.listenRankTableChange(stream)
	c.deletePublisher(request.JobId)
	return nil
}

func (c *BusinessConfigServer) getPublisher(jobId string) (*ConfigPublisher, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	publisher, ok := c.configPublisher[jobId]
	return publisher, ok
}

func (c *BusinessConfigServer) deletePublisher(jobId string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.configPublisher, jobId)
}

func (c *BusinessConfigServer) addPublisher(jobId string) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	publisher := NewConfigPublisher(jobId, c.serviceCtx)
	c.configPublisher[jobId] = publisher
}
