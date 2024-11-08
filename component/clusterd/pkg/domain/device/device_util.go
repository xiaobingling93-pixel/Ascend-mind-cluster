// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package device a series of device function
package device

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/api/core/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const safeDeviceSize = 1000

// ParseDeviceInfoCM get device info from configmap obj
func ParseDeviceInfoCM(obj interface{}) (*constant.DeviceInfo, error) {
	deviceCm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return &constant.DeviceInfo{}, fmt.Errorf("not device configmap")
	}
	devInfoCM := constant.DeviceInfoCM{}
	data, ok := deviceCm.Data[constant.DevInfoCMKey]
	if !ok {
		return &constant.DeviceInfo{}, fmt.Errorf("configmap %s has no %s", deviceCm.Name, constant.DevInfoCMKey)
	}

	if unmarshalErr := json.Unmarshal([]byte(data), &devInfoCM); unmarshalErr != nil {
		return &constant.DeviceInfo{}, fmt.Errorf("unmarshal failed: %v, configmap name: %s", unmarshalErr, deviceCm.Name)
	}

	if !util.EqualDataHash(devInfoCM.CheckCode, devInfoCM.DeviceInfo) {
		return &constant.DeviceInfo{}, fmt.Errorf("device info configmap %s is not valid", deviceCm.Name)
	}
	var device constant.DeviceInfo
	device.DeviceList = devInfoCM.DeviceInfo.DeviceList
	device.UpdateTime = devInfoCM.DeviceInfo.UpdateTime
	device.ServerIndex = devInfoCM.ServerIndex
	device.SuperPodID = devInfoCM.SuperPodID
	device.CmName = deviceCm.Name
	return &device, nil
}

// DeepCopy deep copy deviceInfo
func DeepCopy(info *constant.DeviceInfo) *constant.DeviceInfo {
	if info == nil {
		return nil
	}
	data, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("marshal device failed , err is %v", err)
		return nil
	}
	newDeviceInfo := &constant.DeviceInfo{}
	if err := json.Unmarshal(data, newDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal device failed , err is %v", err)
		return nil
	}
	return newDeviceInfo
}

func DeepCopyInfos(infos map[string]*constant.DeviceInfo) map[string]*constant.DeviceInfo {
	res := make(map[string]*constant.DeviceInfo)
	for key, val := range infos {
		res[key] = DeepCopy(val)
	}
	return res
}

// GetSafeData get data every 1000 DeviceInfo
func GetSafeData(deviceInfos map[string]*constant.DeviceInfo) []string {
	if len(deviceInfos) == 0 {
		return []string{}
	}
	if len(deviceInfos) <= safeDeviceSize {
		return []string{util.ObjToString(deviceInfos)}
	}
	deviceSlice := make([]string, 0, len(deviceInfos)/safeDeviceSize+1)
	childDeviceInfos := make(map[string]*constant.DeviceInfo, safeDeviceSize)
	for cmName, deviceInfo := range deviceInfos {
		childDeviceInfos[cmName] = deviceInfo
		if len(childDeviceInfos)%safeDeviceSize == 0 {
			deviceSlice = append(deviceSlice, util.ObjToString(childDeviceInfos))
			childDeviceInfos = make(map[string]*constant.DeviceInfo, safeDeviceSize)
		}
	}
	if len(childDeviceInfos) != 0 {
		deviceSlice = append(deviceSlice, util.ObjToString(childDeviceInfos))
	}
	return deviceSlice
}

// BusinessDataIsNotEqual determine the business data is not equal
func BusinessDataIsNotEqual(oldDevInfo *constant.DeviceInfo, devInfo *constant.DeviceInfo) bool {
	if oldDevInfo == nil && devInfo == nil {
		hwlog.RunLog.Debug("both oldDevInfo and devInfo are nil")
		return false
	}
	if oldDevInfo == nil || devInfo == nil {
		hwlog.RunLog.Debug("one of oldDevInfo and devInfo is not empty, and the other is empty")
		return true
	}
	if len(oldDevInfo.DeviceList) != len(devInfo.DeviceList) {
		hwlog.RunLog.Debug("the length of the deviceList of oldDevInfo is not equal to that of the deviceList of devInfo")
		return true
	}
	for nKey, nValue := range oldDevInfo.DeviceList {
		oValue, exists := devInfo.DeviceList[nKey]
		if !exists || nValue != oValue {
			hwlog.RunLog.Debug("neither oldDevInfo nor devInfo is empty, but oldDevInfo is not equal to devInfo")
			return true
		}
	}
	hwlog.RunLog.Debug("oldDevInfo is equal to devInfo")
	return false
}

func GetFaultMap(devInfo *constant.DeviceInfo) map[string][]constant.DeviceFault {
	if devInfo == nil {
		hwlog.RunLog.Error(fmt.Errorf("get fault list for node failed. devInfo is nil"))
		return make(map[string][]constant.DeviceFault)
	}
	if devInfo == nil || devInfo.DeviceList == nil {
		hwlog.RunLog.Error(fmt.Errorf("get fault list for node %v failed. device list does not exist", devInfo.CmName))
		return make(map[string][]constant.DeviceFault)
	}
	if faultList, ok := devInfo.DeviceList[GetFaultListKey()]; ok {
		var devicesFault []constant.DeviceFault
		err := json.Unmarshal([]byte(faultList), &devicesFault)
		if err != nil {
			hwlog.RunLog.Error(fmt.Errorf("get fault list for node %v failed. "+
				"Json unmarshall exception: %v", devInfo.CmName, err))
			return make(map[string][]constant.DeviceFault)
		}
		deviceFaultMap := make(map[string][]constant.DeviceFault)
		for _, deviceFault := range devicesFault {
			if deviceFault.FaultTime == 0 {
				deviceFault.FaultTime = constant.DeviceNotFault
			}
			if _, ok := deviceFaultMap[deviceFault.NPUName]; !ok {
				deviceFaultMap[deviceFault.NPUName] = make([]constant.DeviceFault, 0)
			}
			deviceFaultMap[deviceFault.NPUName] = append(deviceFaultMap[deviceFault.NPUName], deviceFault)
		}
		return deviceFaultMap
	}
	hwlog.RunLog.Error(fmt.Errorf("get fault list for node %v failed. fault list does not exist", devInfo.CmName))
	return make(map[string][]constant.DeviceFault)
}

func DeleteFaultFromFaultMap(faultMap map[string][]constant.DeviceFault,
	delFault constant.DeviceFault) map[string][]constant.DeviceFault {
	deviceFaults, ok := faultMap[delFault.NPUName]
	if !ok {
		return faultMap
	}
	newDeviceFaults := make([]constant.DeviceFault, 0)
	for _, fault := range deviceFaults {
		if reflect.DeepEqual(delFault, fault) {
			continue
		}
		newDeviceFaults = append(newDeviceFaults, fault)
	}
	faultMap[delFault.NPUName] = newDeviceFaults
	return faultMap
}

func FaultMapToArrayToString(faultMap map[string][]constant.DeviceFault) string {
	array := make([]constant.DeviceFault, 0)
	for _, faults := range faultMap {
		array = append(array, faults...)
	}
	return util.ObjToString(array)
}

// TODO key应该是什么
func GetFaultListKey() string {
	return "huawei.com/Ascend910-Fault"
}

// TODO 如何判断device fault是uce故障
func IsUceFault(faultDevice constant.DeviceFault) bool {
	if strings.Contains(faultDevice.FaultCode, constant.UCE_FAULT_CODE) {
		return true
	}
	return false
}

// TODO 如何判断device fault是uce伴随故障
func IsUceAccompanyFault(faultDevice constant.DeviceFault) bool {
	return strings.Contains(faultDevice.FaultCode, constant.AIC_FAULT_CODE) ||
		strings.Contains(faultDevice.FaultCode, constant.AIV_FAULT_CODE)
}
