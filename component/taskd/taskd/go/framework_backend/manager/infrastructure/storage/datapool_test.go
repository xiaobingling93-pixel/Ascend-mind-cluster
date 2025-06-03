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
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"taskd/common/constant"
	"taskd/toolkit_backend/net/common"
)

// All mock agent,cluster,worker
var (
	agentName = "agent1"
	agentInfo = &Agent{
		Status:    map[string]string{"status1": "value1"},
		NodeRank:  "1",
		HeartBeat: time.Now(),
		Pos:       &common.Position{Role: "agent"},
		RWMutex:   sync.RWMutex{},
	}
	workerName = "worker1"
	workerInfo = &Worker{
		Status:     map[string]string{"status1": "value1"},
		GlobalRank: "1",
		HeartBeat:  time.Now(),
		Pos:        &common.Position{Role: "worker"},
		RWMutex:    sync.RWMutex{},
	}
	clusterName = "cluster1"
	clusterInfo = &Cluster{
		Command:   map[string]string{"cmd1": "value1"},
		HeartBeat: time.Now(),
		Business:  []int32{0, 0, 0},
		RWMutex:   sync.RWMutex{},
	}
)

func newMsgQueue(length int32) *MsgQueue {
	return &MsgQueue{Queue: make([]BaseMessage, length), Mutex: sync.Mutex{}}
}

func newDataPool() *DataPool {
	return &DataPool{
		Snapshot: &SnapShot{
			AgentInfos: &AgentInfos{
				Agents:    make(map[string]*Agent),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
			WorkerInfos: &WorkerInfos{
				Workers:   make(map[string]*Worker),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
			ClusterInfos: &ClusterInfos{
				Clusters:  make(map[string]*Cluster),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
		},
		RWMutex: sync.RWMutex{},
	}
}

// TestNewMsg test new a message
func TestNewMsg(t *testing.T) {
	mq := newMsgQueue(0)
	convey.Convey("NewMsg should create valid message", t, func() {
		msg := mq.NewMsg("test123", "TEST_BIZ", &common.Position{Role: "test"}, MsgBody{MsgType: "TEST"})
		convey.So(msg.Header.Uuid, convey.ShouldEqual, "test123")
		convey.So(msg.Header.BizType, convey.ShouldEqual, "TEST_BIZ")
		convey.So(msg.Body.MsgType, convey.ShouldEqual, "TEST")
		convey.So(msg.Header.Timestamp.IsZero(), convey.ShouldBeFalse)
	})
}

// TestEnqueue test message enter the queue
func TestEnqueue(t *testing.T) {
	convey.Convey("TestEnqueue test message enter the queue success", t, func() {
		mq := newMsgQueue(0)
		msg := BaseMessage{}
		err := mq.Enqueue(msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestEnqueue test message fail", t, func() {
		mq := newMsgQueue(constant.MaxMsgQueueLength)
		msg := BaseMessage{}
		err := mq.Enqueue(msg)
		convey.So(err.Error(), convey.ShouldEqual, "message queue is full")
	})
}

// TestDequeue test message departure
func TestDequeue(t *testing.T) {
	convey.Convey("TestDequeue test message departure success", t, func() {
		msg := BaseMessage{Header: MsgHeader{}, Body: MsgBody{}}
		mq := &MsgQueue{
			Queue: []BaseMessage{msg},
			Mutex: sync.Mutex{},
		}
		dqMsg, err := mq.Dequeue()
		convey.So(dqMsg, convey.ShouldEqual, msg)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestRegisterAgent test register agent in the data pool
func TestRegisterAgent(t *testing.T) {
	dp := newDataPool()
	convey.Convey("TestRegisterAgent should add new agent", t, func() {
		err := dp.RegisterAgent(agentName, agentInfo)
		convey.So(err, convey.ShouldBeNil)
		agent, err := dp.GetAgent(agentName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(agent.NodeRank, convey.ShouldEqual, "1")
		convey.So(agent.Status["status1"], convey.ShouldEqual, "value1")
	})
}

// TestRegisterWorker test register worker in the data pool
func TestRegisterWorker(t *testing.T) {
	dp := newDataPool()
	convey.Convey("TestRegisterWorker should add new worker", t, func() {
		err := dp.RegisterWorker(workerName, workerInfo)
		convey.So(err, convey.ShouldBeNil)
		worker, err := dp.GetWorker(workerName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(worker.GlobalRank, convey.ShouldEqual, "1")
		convey.So(worker.Status["status1"], convey.ShouldEqual, "value1")
	})
}

// TestRegisterCluster test register cluster in the data pool
func TestRegisterCluster(t *testing.T) {
	dp := newDataPool()
	convey.Convey("RegisterCluster should add new cluster", t, func() {
		dp.RegisterCluster(clusterName, clusterInfo)

		cluster, err := dp.GetCluster(clusterName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(cluster.Command), convey.ShouldEqual, 1)
		convey.So(cluster.Business, convey.ShouldResemble, []int32{0, 0, 0})
	})
}

// TestUpdateAgent test update agent info in the data pool
func TestUpdateAgent(t *testing.T) {
	dp := newDataPool()
	convey.Convey("UpdateAgent should modify existing agent", t, func() {
		_ = dp.RegisterAgent(agentName, agentInfo)
		updatedAgent := &Agent{
			Status:    map[string]string{"status1": "updated"},
			NodeRank:  "2",
			HeartBeat: time.Now(),
			Pos:       &common.Position{Role: "agent"},
			RWMutex:   sync.RWMutex{},
		}
		err := dp.UpdateAgent(agentName, updatedAgent)
		convey.So(err, convey.ShouldBeNil)
		agent, _ := dp.GetAgent(agentName)
		convey.So(agent.NodeRank, convey.ShouldEqual, "2")
		convey.So(agent.Status["status1"], convey.ShouldEqual, "updated")
	})
	convey.Convey("UpdateAgent should fail with nil data pool", t, func() {
		var nilDp *DataPool
		err := nilDp.UpdateAgent(agentName, agentInfo)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "agents is not initialized")
	})
}

// TestUpdateWorker test update worker info in the data pool
func TestUpdateWorker(t *testing.T) {
	dp := newDataPool()
	convey.Convey("UpdateWorker should modify existing worker", t, func() {
		_ = dp.RegisterWorker(workerName, workerInfo)
		updatedWorker := &Worker{
			Status:     map[string]string{"status1": "updated"},
			GlobalRank: "2",
			HeartBeat:  time.Now(),
			Pos:        &common.Position{Role: "worker"},
			RWMutex:    sync.RWMutex{},
		}
		err := dp.UpdateWorker(workerName, updatedWorker)
		convey.So(err, convey.ShouldBeNil)
		worker, _ := dp.GetWorker(workerName)
		convey.So(worker.GlobalRank, convey.ShouldEqual, "2")
		convey.So(worker.Status["status1"], convey.ShouldEqual, "updated")
	})
	convey.Convey("UpdateWorker should fail with nil data pool", t, func() {
		var nilDp *DataPool
		err := nilDp.UpdateWorker(workerName, workerInfo)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "workers is not initialized")
	})
}

// TestUpdateCluster test update cluster info in the data pool
func TestUpdateCluster(t *testing.T) {
	dp := newDataPool()
	convey.Convey("UpdateCluster should modify existing cluster", t, func() {
		dp.RegisterCluster(clusterName, clusterInfo)
		business := []int32{0, 0}
		updatedCluster := &Cluster{
			Command:   map[string]string{"cmd1": "updated"},
			HeartBeat: time.Now(),
			Business:  business,
			RWMutex:   sync.RWMutex{},
		}

		err := dp.UpdateCluster(clusterName, updatedCluster)
		convey.So(err, convey.ShouldBeNil)

		cluster, _ := dp.GetCluster(clusterName)
		convey.So(cluster.Command["cmd1"], convey.ShouldEqual, "updated")
		convey.So(cluster.Business, convey.ShouldResemble, business)
	})
}

// TestGetAgent test get agent info from data pool
func TestGetAgent(t *testing.T) {
	dp := newDataPool()
	convey.Convey("GetAgent should return agent info", t, func() {
		_ = dp.RegisterAgent(agentName, agentInfo)
		agent, err := dp.GetAgent(agentName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(agent, convey.ShouldNotBeNil)
		convey.So(agent.NodeRank, convey.ShouldEqual, "1")
	})

	convey.Convey("GetAgent should fail for non-existent agent", t, func() {
		_, err := dp.GetAgent("nonexistent")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "agent name is unregistered : nonexistent")
	})
}

// TestGetWorker test get worker info from data pool
func TestGetWorker(t *testing.T) {
	dp := newDataPool()
	convey.Convey("GetWorker should return worker info", t, func() {
		_ = dp.RegisterWorker(workerName, workerInfo)
		worker, err := dp.GetWorker(workerName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(worker, convey.ShouldNotBeNil)
		convey.So(worker.GlobalRank, convey.ShouldEqual, "1")
	})

	convey.Convey("GetWorker should fail for non-existent worker", t, func() {
		_, err := dp.GetWorker("nonexistent")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "worker name is unregistered : nonexistent")
	})
}

// TestGetCluster test get cluster info from data pool
func TestGetCluster(t *testing.T) {
	dp := newDataPool()
	convey.Convey("GetCluster should return cluster info", t, func() {
		dp.RegisterCluster(clusterName, clusterInfo)

		cluster, err := dp.GetCluster(clusterName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(cluster, convey.ShouldNotBeNil)
	})

	convey.Convey("GetCluster should fail for non-existent cluster", t, func() {
		_, err := dp.GetCluster("nonexistent")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "cluster name is unregistered : nonexistent")
	})
}

// TestGetSnapShot test get data pool snapshot
func TestGetSnapShot(t *testing.T) {
	convey.Convey("TestGetSnapShot test get data pool snapshot success", t, func() {
		snapshot := &SnapShot{
			AgentInfos: &AgentInfos{
				Agents:    map[string]*Agent{"nilAgent": nil, agentName: agentInfo},
				AllStatus: map[string]string{"status": "value1"},
			},
			WorkerInfos: &WorkerInfos{
				Workers:   map[string]*Worker{"nilWorker": nil, workerName: workerInfo},
				AllStatus: map[string]string{"status": "value1"},
			},
			ClusterInfos: &ClusterInfos{
				Clusters:  map[string]*Cluster{"nilCluster": nil, clusterName: clusterInfo},
				AllStatus: map[string]string{"status": "value1"},
			},
		}
		dp := &DataPool{Snapshot: snapshot}
		getSnapshot, err := dp.GetSnapShot()
		convey.So(err, convey.ShouldBeNil)
		convey.So(getSnapshot, convey.ShouldEqual, snapshot)
	})
	convey.Convey("TestGetSnapShot test get data pool snapshot fail", t, func() {
		dp := &DataPool{Snapshot: nil}
		getSnapshot, err := dp.GetSnapShot()
		convey.So(getSnapshot, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "snapshot is null")
	})
}
