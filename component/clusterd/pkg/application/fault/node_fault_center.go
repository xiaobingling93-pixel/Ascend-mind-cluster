package fault

import (
	"sync"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
)

// nodeFaultProcessCenter
type nodeFaultProcessCenter struct {
	baseFaultCenter
	infoMap map[string]*constant.NodeInfo
	mutex   sync.RWMutex
}

func newNodeFaultProcessCenter() *nodeFaultProcessCenter {
	return &nodeFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(),
		infoMap:         make(map[string]*constant.NodeInfo),
		mutex:           sync.RWMutex{},
	}
}

func (nodeCenter *nodeFaultProcessCenter) getInfoMap() map[string]*constant.NodeInfo {
	nodeCenter.mutex.RLock()
	defer nodeCenter.mutex.RUnlock()
	return node.DeepCopyInfos(nodeCenter.infoMap)
}

func (nodeCenter *nodeFaultProcessCenter) setInfoMap(infos map[string]*constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	nodeCenter.infoMap = node.DeepCopyInfos(infos)
}

func (nodeCenter *nodeFaultProcessCenter) updateInfoFromCm(newInfo *constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	length := len(nodeCenter.infoMap)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	nodeCenter.infoMap[newInfo.CmName] = newInfo
}

func (nodeCenter *nodeFaultProcessCenter) delInfoFromCm(newInfo *constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	delete(nodeCenter.infoMap, newInfo.CmName)
}
