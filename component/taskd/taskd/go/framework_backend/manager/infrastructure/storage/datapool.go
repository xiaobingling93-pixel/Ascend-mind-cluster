/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package storage for taskd manager backend data type
package storage

import (
	"fmt"
	"sync"
	"time"

	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/toolkit_backend/net/common"
)

// DataPool is used to store message info
type DataPool struct {
	Snapshot *SnapShot
	RWMutex  sync.RWMutex
}

// SnapShot is used to store agent worker cluster infos
type SnapShot struct {
	AgentInfos   *AgentInfos
	WorkerInfos  *WorkerInfos
	ClusterInfos *ClusterInfos
}

// MsgQueue the queue store message
type MsgQueue struct {
	Queue []BaseMessage
	Mutex sync.Mutex
}

// NewMsg new a base message to enqueue
func (mq *MsgQueue) NewMsg(uuid string, bizType string, src *common.Position, msgBody MsgBody) BaseMessage {
	return BaseMessage{
		Header: MsgHeader{
			Uuid:      uuid,
			BizType:   bizType,
			Src:       src,
			Timestamp: time.Now(),
		},
		Body: msgBody,
	}
}

// Enqueue message enter the queue
func (mq *MsgQueue) Enqueue(msg BaseMessage) error {
	mq.Mutex.Lock()
	defer mq.Mutex.Unlock()
	if len(mq.Queue) >= constant.MaxMsgQueueLength {
		return fmt.Errorf("message queue is full")
	}

	mq.Queue = append(mq.Queue, msg)
	return nil
}

// Dequeue message departure
func (mq *MsgQueue) Dequeue() (BaseMessage, error) {
	if len(mq.Queue) == 0 {
		return BaseMessage{}, fmt.Errorf("message queue is empty")
	}
	mq.Mutex.Lock()
	defer mq.Mutex.Unlock()
	msg := mq.Queue[0]
	mq.Queue = mq.Queue[1:]
	return msg, nil
}

// RegisterAgent register agent in the data pool
func (d *DataPool) RegisterAgent(agentName string, agentInfo *Agent) error {
	err := d.Snapshot.AgentInfos.registerAgent(agentName, agentInfo)
	return err
}

// RegisterWorker register worker in the data pool
func (d *DataPool) RegisterWorker(workerName string, workerInfo *Worker) error {
	err := d.Snapshot.WorkerInfos.registerWorker(workerName, workerInfo)
	return err
}

// RegisterCluster register cluster in the data pool
func (d *DataPool) RegisterCluster(clusterName string, clusterInfo *Cluster) {
	d.Snapshot.ClusterInfos.registerCluster(clusterName, clusterInfo)
}

// UpdateAgent update agent info in the data pool
func (d *DataPool) UpdateAgent(agentName string, agentInfo *Agent) error {
	if d == nil || d.Snapshot == nil || d.Snapshot.AgentInfos == nil || d.Snapshot.AgentInfos.Agents == nil {
		return fmt.Errorf("agents is not initialized")
	}
	err := d.Snapshot.AgentInfos.updateAgent(agentName, agentInfo)
	return err
}

// UpdateWorker update worker info in the data pool
func (d *DataPool) UpdateWorker(workerName string, workerInfo *Worker) error {
	if d == nil || d.Snapshot == nil || d.Snapshot.WorkerInfos == nil || d.Snapshot.WorkerInfos.Workers == nil {
		return fmt.Errorf("workers is not initialized")
	}
	err := d.Snapshot.WorkerInfos.updateWorker(workerName, workerInfo)
	return err
}

// UpdateCluster update cluster info in the data pool
func (d *DataPool) UpdateCluster(clusterName string, clusterInfo *Cluster) error {
	if d == nil || d.Snapshot == nil || d.Snapshot.ClusterInfos == nil || d.Snapshot.ClusterInfos.Clusters == nil {
		return fmt.Errorf("clusters is not initialized")
	}
	err := d.Snapshot.ClusterInfos.updateCluster(clusterName, clusterInfo)
	return err
}

// GetAgent return agent info about agent name
func (d *DataPool) GetAgent(agentName string) (*Agent, error) {
	return d.Snapshot.AgentInfos.getAgent(agentName)
}

// GetWorker return worker info about worker name
func (d *DataPool) GetWorker(workerName string) (*Worker, error) {
	return d.Snapshot.WorkerInfos.getWorker(workerName)
}

// GetCluster return cluster info about cluster name
func (d *DataPool) GetCluster(clusterName string) (*Cluster, error) {
	return d.Snapshot.ClusterInfos.getCluster(clusterName)
}

// GetSnapShot get data pool snapshot
func (d *DataPool) GetSnapShot() (*SnapShot, error) {
	d.RWMutex.RLock()
	defer d.RWMutex.RUnlock()
	return d.Snapshot.deepCopy()
}

func (s *SnapShot) deepCopy() (*SnapShot, error) {
	if s == nil {
		return nil, fmt.Errorf("snapshot is null")
	}
	clone := &SnapShot{}
	if s.AgentInfos != nil {
		clone.AgentInfos = deepCopyAgent(s.AgentInfos)
	}
	if s.WorkerInfos != nil {
		clone.WorkerInfos = deepCopyWorker(s.WorkerInfos)
	}
	if s.ClusterInfos != nil {
		clone.ClusterInfos = deepCopyCluster(s.ClusterInfos)
	}
	return clone, nil
}

func deepCopyAgent(agentInfos *AgentInfos) *AgentInfos {
	clone := &AgentInfos{
		Agents:    make(map[string]*Agent, len(agentInfos.Agents)),
		AllStatus: make(map[string]string, len(agentInfos.AllStatus)),
	}
	for k, v := range agentInfos.AllStatus {
		clone.AllStatus[k] = v
	}
	for k, v := range agentInfos.Agents {
		if v == nil {
			clone.Agents[k] = nil
			continue
		}
		cloneAgent := &Agent{
			Status:    v.Status,
			NodeRank:  v.NodeRank,
			HeartBeat: v.HeartBeat,
		}
		cloneAgent.Config = utils.CopyStringMap(v.Config)
		cloneAgent.Actions = utils.CopyStringMap(v.Actions)
		cloneAgent.FaultInfo = utils.CopyStringMap(v.FaultInfo)
		if v.Pos != nil {
			cloneAgent.Pos = &common.Position{
				Role:        v.Pos.Role,
				ServerRank:  v.Pos.ServerRank,
				ProcessRank: v.Pos.ProcessRank,
			}
		}
		clone.Agents[k] = cloneAgent
	}
	return clone
}

func deepCopyWorker(workerInfos *WorkerInfos) *WorkerInfos {
	clone := &WorkerInfos{
		Workers:   make(map[string]*Worker, len(workerInfos.Workers)),
		AllStatus: make(map[string]string, len(workerInfos.AllStatus)),
	}
	for k, v := range workerInfos.AllStatus {
		clone.AllStatus[k] = v
	}
	for k, v := range workerInfos.Workers {
		if v == nil {
			clone.Workers[k] = nil
			continue
		}
		cloneWorker := &Worker{
			Status:     v.Status,
			GlobalRank: v.GlobalRank,
			HeartBeat:  v.HeartBeat,
		}
		cloneWorker.Config = utils.CopyStringMap(v.Config)
		cloneWorker.Actions = utils.CopyStringMap(v.Actions)
		cloneWorker.FaultInfo = utils.CopyStringMap(v.FaultInfo)
		if v.Pos != nil {
			cloneWorker.Pos = &common.Position{
				Role:        v.Pos.Role,
				ServerRank:  v.Pos.ServerRank,
				ProcessRank: v.Pos.ProcessRank,
			}
		}

		clone.Workers[k] = cloneWorker
	}
	return clone
}

func deepCopyCluster(clusterInfos *ClusterInfos) *ClusterInfos {
	clone := &ClusterInfos{
		Clusters:  make(map[string]*Cluster, len(clusterInfos.Clusters)),
		AllStatus: make(map[string]string, len(clusterInfos.AllStatus)),
	}
	for k, v := range clusterInfos.AllStatus {
		clone.AllStatus[k] = v
	}
	for k, v := range clusterInfos.Clusters {
		if v == nil {
			clone.Clusters[k] = nil
			continue
		}
		cloneCluster := &Cluster{
			HeartBeat: v.HeartBeat,
		}
		cloneCluster.Command = utils.CopyStringMap(v.Command)
		cloneCluster.FaultInfo = utils.CopyStringMap(v.FaultInfo)
		cloneCluster.Business = make([]int32, 0)
		cloneCluster.Business = append(cloneCluster.Business, v.Business...)
		if v.Pos != nil {
			cloneCluster.Pos = &common.Position{
				Role:        v.Pos.Role,
				ServerRank:  v.Pos.ServerRank,
				ProcessRank: v.Pos.ProcessRank,
			}
		}
		clone.Clusters[k] = cloneCluster
	}
	return clone
}
