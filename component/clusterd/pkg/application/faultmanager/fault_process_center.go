// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"context"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocesscenter"
	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/application/faultmanager/faultrank"
	"clusterd/pkg/application/faultmanager/jobprocesscenter"
	"clusterd/pkg/common/constant"
)

// GlobalFaultProcessCenter is a global instance of FaultProcessCenter used for processing faults.
var GlobalFaultProcessCenter *FaultProcessCenter

// FaultProcessCenter processes the faults and coordinates the fault handling among different components.
type FaultProcessCenter struct {
	NotifyProcessChan chan int
}

func (center *FaultProcessCenter) Process() {
	cmprocesscenter.DeviceCenter.Process()
	cmprocesscenter.NodeCenter.Process()
	cmprocesscenter.SwitchCenter.Process()
	allConfigmapContent := constant.AllConfigmapContent{
		DeviceCm: cmprocesscenter.DeviceCenter.GetProcessedCm(),
		SwitchCm: cmprocesscenter.SwitchCenter.GetProcessedCm(),
		NodeCm:   cmprocesscenter.NodeCenter.GetProcessedCm(),
	}
	faultrank.FaultProcessor.Process(allConfigmapContent)
	jobprocesscenter.FaultJobCenter.Process()
}

func init() {
	GlobalFaultProcessCenter = &FaultProcessCenter{
		NotifyProcessChan: make(chan int, 1000),
	}
}

func (center *FaultProcessCenter) NotifyFaultCenterProcess(whichToProcess int) {
	center.NotifyProcessChan <- whichToProcess
}

func (center *FaultProcessCenter) Work(ctx context.Context) {
	go func() {
		hwlog.RunLog.Info("FaultProcessCenter start work")
		centerTicker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				hwlog.RunLog.Info("FaultProcessCenter stop work")
				return
			case whichToProcess := <-center.NotifyProcessChan:
				switch whichToProcess {
				case constant.AllProcessType:
					center.Process()
				case constant.DeviceProcessType:
					cmprocesscenter.DeviceCenter.Process()
				case constant.NodeProcessType:
					cmprocesscenter.NodeCenter.Process()
				case constant.SwitchProcessType:
					cmprocesscenter.SwitchCenter.Process()
				default:
					hwlog.RunLog.Errorf("wrong number %d to process", whichToProcess)
				}
			case <-centerTicker.C:
				center.Process()
			}
		}
	}()
}

// CallbackForReportUceInfo callback function to report uce info
func (center *FaultProcessCenter) CallbackForReportUceInfo(infos []constant.ReportRecoverInfo) error {
	for _, info := range infos {
		collector.ReportInfoCollector.ReportUceInfo(info.JobId, info.Rank, info.RecoverTime)
	}
	center.NotifyFaultCenterProcess(constant.DeviceProcessType)
	return nil
}

// Register to notify fault occurrence
func (center *FaultProcessCenter) Register(ch chan int, whichToRegister int) {
	switch whichToRegister {
	case constant.SwitchProcessType:
		cmprocesscenter.SwitchCenter.Register(ch)
	case constant.NodeProcessType:
		cmprocesscenter.NodeCenter.Register(ch)
	case constant.DeviceProcessType:
		cmprocesscenter.DeviceCenter.Register(ch)
	case constant.AllProcessType:
		cmprocesscenter.SwitchCenter.Register(ch)
		cmprocesscenter.NodeCenter.Register(ch)
		cmprocesscenter.DeviceCenter.Register(ch)
	default:
		hwlog.RunLog.Errorf("Wrong number %d, cannot decide which to register", whichToRegister)
	}
}

// QueryJobsFaultInfo query jobs fault rank info, and filter fault below `faultLevel`
func (center *FaultProcessCenter) QueryJobsFaultInfo(faultLevel string) map[string]constant.JobFaultInfo {
	return faultrank.JobFaultRankProcessor.GetJobFaultRankInfosFilterLevel(faultLevel)
}

// QueryDeviceInfoToReport query device info to report
func (center *FaultProcessCenter) QueryDeviceInfoToReport() map[string]*constant.DeviceInfo {
	return cmprocesscenter.DeviceCenter.GetProcessedCm()
}

// QuerySwitchInfoToReport query switch info to report
func (center *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	return cmprocesscenter.SwitchCenter.GetProcessedCm()
}

// QueryNodeInfoToReport query node info to report
func (center *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	return cmprocesscenter.NodeCenter.GetProcessedCm()
}
