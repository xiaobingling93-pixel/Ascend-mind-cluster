// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultjob contain fault job process
package faultjob

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func (fJob *FaultJob) handleJobFault(relationFault []*constant.FaultInfo,
	triggerFault []constant.FaultInfo, strategyList []constant.RelationFaultStrategy) (map[string]string, map[string][]constant.DeviceStrategy) {

	nodeLvList := make(map[string]string)
	deviceLvList := make(map[string][]constant.DeviceStrategy)

	if len(relationFault) <= 0 || len(triggerFault) <= 0 || len(strategyList) <= 0 {
		return nodeLvList, deviceLvList
	}
	hwlog.RunLog.Infof("----relationFault: %s, triggerFault: %s, strategyList: %s----",
		util.ObjToString(relationFault), util.ObjToString(triggerFault), util.ObjToString(strategyList))

	// get code device map
	relationCodeDeviceMap, nodeDeviceList, triggerCodeDeviceMap := fJob.transferFaultToMap(relationFault, triggerFault)

	// find all matching configuration tables
	curFaultTables := fJob.getCodeMatchedTables(relationCodeDeviceMap, triggerCodeDeviceMap, strategyList)
	if len(curFaultTables) <= 0 {
		return nodeLvList, deviceLvList
	}
	hwlog.RunLog.Infof("----curFaultTables: %s----", util.ObjToString(curFaultTables))

	// get matched devices
	allUnhealthyDevices, allSubHealthDevices := fJob.getFaultDevices(curFaultTables, relationCodeDeviceMap)
	hwlog.RunLog.Infof("----allUnhealthyDevices: %s, allSubHealthDevices: %s----",
		util.ObjToString(allUnhealthyDevices), util.ObjToString(allSubHealthDevices))

	// get the nodes and devices that isolated or sub-health
	nodeLvList, deviceLvList = fJob.getAllNodeDeviceList(allUnhealthyDevices, allSubHealthDevices, nodeDeviceList)
	hwlog.RunLog.Infof("----nodeLvList: %s, deviceLvList: %s----",
		util.ObjToString(nodeLvList), util.ObjToString(deviceLvList))

	return nodeLvList, deviceLvList
}

func (fJob *FaultJob) getNPUUnderSwitch(faultDevices []*constant.FaultInfo) []constant.DeviceStrategy {
	return make([]constant.DeviceStrategy, 0)
}

func (fJob *FaultJob) getAllNodeDeviceList(allUnhealthyDevices, allSubHealthDevices []*constant.FaultInfo,
	nodeDeviceList map[string]string) (map[string]string, map[string][]constant.DeviceStrategy) {
	nodeLvList := make(map[string]string)
	deviceLvList := make(map[string][]constant.DeviceStrategy)

	// both isolation and sub-health are required, choose the highest level
	// node level higher than device
	for nodeName := range nodeDeviceList {
		deviceList := make([]constant.DeviceStrategy, 0)
		nodeLevel := false
		for _, device := range allSubHealthDevices {
			if device.NodeName != nodeName {
				continue
			}

			if device.FaultType == constant.SwitchFault {
				if fJob.FindNPUUnderSwitch {
					deviceList = append(deviceList, fJob.getNPUUnderSwitch(allSubHealthDevices)...)
				} else {
					nodeLvList[nodeName] = constant.SubHealthFaultStrategy
					nodeLevel = true
				}
			} else {
				if fJob.getDeviceIdx(device, deviceList) < 0 {
					deviceList = append(deviceList, constant.DeviceStrategy{NPUName: device.NPUName, Strategy: constant.SubHealthFaultStrategy})
				}
			}
		}
		if !nodeLevel {
			deviceLvList[nodeName] = deviceList
		}
	}

	for nodeName := range nodeDeviceList {
		deviceList := deviceLvList[nodeName]
		nodeLevel := false
		for _, device := range allUnhealthyDevices {
			if device.NodeName != nodeName {
				continue
			}

			if device.FaultType == constant.SwitchFault {
				if fJob.FindNPUUnderSwitch {
					deviceList = append(deviceList, fJob.getNPUUnderSwitch(allUnhealthyDevices)...)
				} else {
					nodeLvList[nodeName] = constant.SeparateFaultStrategy
					nodeLevel = true
				}
			} else {
				idx := fJob.getDeviceIdx(device, deviceList)
				if idx < 0 {
					deviceList = append(deviceList, constant.DeviceStrategy{NPUName: device.NPUName, Strategy: constant.SeparateFaultStrategy})
				} else {
					deviceList[idx] = constant.DeviceStrategy{NPUName: device.NPUName, Strategy: constant.SeparateFaultStrategy}
				}
			}
		}
		if !nodeLevel {
			deviceLvList[nodeName] = deviceList
		}
	}
	return nodeLvList, deviceLvList
}

func (fJob *FaultJob) getDeviceIdx(device *constant.FaultInfo, deviceList []constant.DeviceStrategy) int {
	for i, d := range deviceList {
		if d.NPUName == device.NPUName {
			return i
		}
	}
	return -1
}

func (fJob *FaultJob) transferFaultToMap(relationFault []*constant.FaultInfo,
	triggerFault []constant.FaultInfo) (map[string][]*constant.FaultInfo, map[string]string, map[string]constant.FaultInfo) {
	relationCodeDeviceMap := make(map[string][]*constant.FaultInfo)
	nodeList := make(map[string]string)
	for _, device := range relationFault {
		if _, ok := nodeList[device.NodeName]; !ok {
			nodeList[device.NodeName] = device.NodeName
		}
		if _, ok := relationCodeDeviceMap[device.FaultCode]; !ok {
			relationCodeDeviceMap[device.FaultCode] = make([]*constant.FaultInfo, 0)
		}

		relationCodeDeviceMap[device.FaultCode] = append(relationCodeDeviceMap[device.FaultCode], device)
	}

	triggerCodeDeviceMap := make(map[string]constant.FaultInfo)
	for _, trigger := range triggerFault {
		triggerCodeDeviceMap[trigger.FaultCode] = trigger
	}

	return relationCodeDeviceMap, nodeList, triggerCodeDeviceMap
}

func (fJob *FaultJob) getCodeMatchedTables(relationCodeDeviceMap map[string][]*constant.FaultInfo,
	triggerCodeDeviceMap map[string]constant.FaultInfo, strategyList []constant.RelationFaultStrategy) []constant.RelationFaultStrategy {
	curFaultTables := make([]constant.RelationFaultStrategy, 0)

	for _, trigger := range strategyList {
		if _, ok := triggerCodeDeviceMap[trigger.TriggerFault]; !ok {
			continue
		}
		if !fJob.relationFaultsMatched(relationCodeDeviceMap, trigger) {
			continue
		}

		curFaultTables = append(curFaultTables, trigger)
	}
	return curFaultTables
}

func (fJob *FaultJob) relationFaultsMatched(relationCodeDeviceMap map[string][]*constant.FaultInfo, trigger constant.RelationFaultStrategy) bool {
	for _, fault := range trigger.RelationFaults {
		if _, ok := relationCodeDeviceMap[fault]; !ok {
			return false
		}
	}
	return true
}

func (fJob *FaultJob) getFaultDevices(curFaultTables []constant.RelationFaultStrategy,
	relationCodeDeviceMap map[string][]*constant.FaultInfo) ([]*constant.FaultInfo, []*constant.FaultInfo) {
	allUnhealthyDevices := make([]*constant.FaultInfo, 0)
	allSubHealthDevices := make([]*constant.FaultInfo, 0)

	for _, configTable := range curFaultTables {
		unhealthyDevices, subHealthDevices := fJob.getMatchedDevices(configTable, relationCodeDeviceMap)
		allUnhealthyDevices = append(allUnhealthyDevices, unhealthyDevices...)
		allSubHealthDevices = append(allSubHealthDevices, subHealthDevices...)
	}

	return allUnhealthyDevices, allSubHealthDevices
}

func (fJob *FaultJob) getMatchedDevices(configTable constant.RelationFaultStrategy,
	relationCodeDeviceMap map[string][]*constant.FaultInfo) ([]*constant.FaultInfo, []*constant.FaultInfo) {
	unhealthyDevices := make([]*constant.FaultInfo, 0)
	subHealthDevices := make([]*constant.FaultInfo, 0)
	for _, relationFault := range configTable.RelationFaults {
		for _, device := range relationCodeDeviceMap[relationFault] {
			if configTable.FaultStrategy == constant.SubHealthFaultStrategy {
				device.ExecutedStrategy = constant.SubHealthFaultStrategy
				subHealthDevices = append(subHealthDevices, device)
			} else if configTable.FaultStrategy == constant.SeparateFaultStrategy {
				device.ExecutedStrategy = constant.SeparateFaultStrategy
				unhealthyDevices = append(unhealthyDevices, device)
			}
		}
	}

	return unhealthyDevices, subHealthDevices
}

func (fJob *FaultJob) processNetworkFault() {
	nodeLvList, deviceLvList := fJob.handleJobFault(fJob.RelationFaults, fJob.TriggerFault, relationFaultStrategies)

	fJob.FaultStrategy = constant.FaultStrategy{
		NodeLvList:   nodeLvList,
		DeviceLvList: deviceLvList,
	}
}
