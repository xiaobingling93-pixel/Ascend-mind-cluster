package fault

import (
	"sync"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/switchinfo"
)

type switchFaultProcessCenter struct {
	baseFaultCenter
	infoMap map[string]*constant.SwitchInfo
	mutex   sync.RWMutex
}

func newSwitchFaultProcessCenter() *switchFaultProcessCenter {
	return &switchFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(),
		infoMap:         make(map[string]*constant.SwitchInfo),
		mutex:           sync.RWMutex{},
	}
}

func (switchCenter *switchFaultProcessCenter) getInfoMap() map[string]*constant.SwitchInfo {
	switchCenter.mutex.RLock()
	defer switchCenter.mutex.RUnlock()
	return switchinfo.DeepCopyInfos(switchCenter.infoMap)
}

func (switchCenter *switchFaultProcessCenter) setInfoMap(infos map[string]*constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	switchCenter.infoMap = switchinfo.DeepCopyInfos(infos)
}

func (switchCenter *switchFaultProcessCenter) updateInfoFromCm(newInfo *constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	length := len(switchCenter.infoMap)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	switchCenter.infoMap[newInfo.CmName] = newInfo
}

func (switchCenter *switchFaultProcessCenter) delInfoFromCm(newInfo *constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	delete(switchCenter.infoMap, newInfo.CmName)
}
