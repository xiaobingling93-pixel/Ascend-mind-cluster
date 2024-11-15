// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import (
	"clusterd/pkg/interface/grpc/pb"
)

// FaultRecoverService is a service for fault recover
type FaultRecoverService struct {
	pb.UnimplementedRecoverServer
}

// NewFaultRecoverService return a new instance of FaultRecoverService
func NewFaultRecoverService() *FaultRecoverService {
	return &FaultRecoverService{}
}

// DeleteJob clear registered resources
func (s *FaultRecoverService) DeleteJob(jobId string) {

}
