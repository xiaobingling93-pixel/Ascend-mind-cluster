// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package main a series of main function
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"clusterd/pkg/application/resource"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	sv "clusterd/pkg/interface/grpc"
	"clusterd/pkg/interface/kube"
)

const (
	defaultLogFile = "/var/log/mindx-dl/clusterd/clusterd.log"

	leaseDuration = 5 * time.Second
	renewDeadline = 3 * time.Second
	retryPeriod   = 2 * time.Second
)

var (
	hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile}
	// BuildVersion build version
	BuildVersion string
	// BuildName build name
	BuildName string
	version   bool
	server    *sv.ClusterInfoMgrServer
)

func leaderElectAndRun() error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to get hostname: %v", err)
	}
	id := hostname + "_" + string(uuid.NewUUID())
	rl, err := resourcelock.New(resourcelock.ConfigMapsLeasesResourceLock, constant.DLNamespace, constant.ComponentName,
		kube.GetClientK8s().ClientSet.CoreV1(),
		kube.GetClientK8s().ClientSet.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
		})
	if err != nil {
		return fmt.Errorf("couldn't create resource lock: %v", err)
	}

	leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDeadline,
		RetryPeriod:   retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				hwlog.RunLog.Warnf("leader election obtain, start informer")
				kube.InitCMInformer()
				kube.InitPodInformer()
				kube.InitPGInformer(ctx)
				kube.AddCmNodeFunc(constant.Resource, resource.NodeCollector)
				kube.AddCmDeviceFunc(constant.Resource, resource.DeviceInfoCollector)
				kube.AddCmSwitchFunc(constant.Resource, resource.SwitchInfoCollector)
				go resource.Report()
			},
			OnStoppedLeading: func() {
				hwlog.RunLog.Warnf("leader election lost, stop informer")
				kube.StopInformer()
				kube.CleanFuncs()
				resource.StopReport()
			},
			OnNewLeader: func(identity string) {
				hwlog.RunLog.Warnf("new leader is %s", identity)
			},
		},
	})
	return nil
}

func main() {
	flag.Parse()
	if version {
		fmt.Printf("%s version: %s \n", BuildName, BuildVersion)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	// init hwlog
	if err := hwlog.InitRunLogger(hwLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	if !checkParameters() {
		return
	}
	err := kube.InitClientK8s()
	if err != nil {
		hwlog.RunLog.Errorf("new client config err: %v", err)
		return
	}
	err = kube.InitClientVolcano()
	if err != nil {
		hwlog.RunLog.Errorf("new volcano client config err: %v", err)
	}
	server = sv.NewClusterInfoMgrServer([]grpc.ServerOption{grpc.MaxRecvMsgSize(constant.MaxGRPCRecvMsgSize),
		grpc.MaxConcurrentStreams(constant.MaxGRPCConcurrentStreams)})
	if err = server.Start(); err != nil {
		hwlog.RunLog.Errorf("cluster info server start failed, err: %#v", err)
	}
	// election and running process
	if err := leaderElectAndRun(); err != nil {
		hwlog.RunLog.Errorf("leader election failed,err is %v", err)
		return
	}
	signalCatch(cancel)
}

func init() {
	flag.BoolVar(&version, "version", false, "the version of the program")

	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup run log files, range [7, 700] days")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile,
		"Run log file path. if the file size exceeds 20MB, will be rotated")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup operator logs, range is (0, 30]")
}

func checkParameters() bool {
	return true
}

func signalCatch(cancel context.CancelFunc) {
	osSignalChan := util.NewSignalWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	if osSignalChan == nil {
		hwlog.RunLog.Error("create stop signal channel failed")
		return
	}
	select {
	case sig, sigEnd := <-osSignalChan:
		if !sigEnd {
			hwlog.RunLog.Info("catch system stop signal channel is closed")
			return
		}
		hwlog.RunLog.Infof("receive system signal: %s, ClusterD shutting down", sig.String())
		if server != nil {
			server.Stop(false)
		}
		cancel()
	}
}
