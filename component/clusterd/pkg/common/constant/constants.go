// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

import (
	"math"
	"time"
)

const (
	// ComponentName is the name of component
	ComponentName = "clusterd"

	// RetryTime is the retry time loading configmap
	RetryTime = 3
	// RetrySleepTime is the sleep time retry loading configmap
	RetrySleepTime = 50 * time.Millisecond

	// MaxGRPCRecvMsgSize 4MB
	MaxGRPCRecvMsgSize = 4 * 1024 * 1024
	// MaxGRPCConcurrentStreams limit on the number of concurrent streams to each ServerTransport.
	MaxGRPCConcurrentStreams = 64
	// MaxConcurrentLimit limit over listener
	MaxConcurrentLimit = 1024
	// MaxIPConnectionLimit limit over ip
	MaxIPConnectionLimit = 512
	// CacheSize cache for ip
	CacheSize = 1024

	// GrpcPort is the grpc port
	GrpcPort = ":8899"
	// JobErrCount err count return
	JobErrCount = 3

	// DefaultNamespace represents the default value of namespace
	DefaultNamespace = "default"
	// TestName represents the default value of name
	TestName = "test"
	// NoResourceOnServer represents no resources on the server
	NoResourceOnServer = "the server could not find the requested resource"

	// CheckFaultGapSecond check fault gap seconds
	CheckFaultGapSecond = 10
)

// fault code const
const (
	UCE_FAULT_CODE = "80E01801"
	AIC_FAULT_CODE = "80C98009"
	AIV_FAULT_CODE = "80CB8009"
)

// fault processor const
const (
	JobNotRecover             = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	JobNotRecoverComplete     = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	DeviceNotFault            = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	DiagnosisAccompanyTimeout = int64(5 * time.Second)
	JobReportRecoverTimeout   = int64(10 * time.Second)
	JobReportCompleteTimeout  = int64(30 * time.Second)
)
