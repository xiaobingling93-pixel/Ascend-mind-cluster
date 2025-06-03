/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain c copy of the License at

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

	"taskd/toolkit_backend/net/common"
)

// ClusterInfos all cluster infos
type ClusterInfos struct {
	Clusters  map[string]*Cluster
	AllStatus map[string]string
	RWMutex   sync.RWMutex
}

// Cluster the cluster info
type Cluster struct {
	Command   map[string]string
	Business  []int32
	HeartBeat time.Time
	FaultInfo map[string]string
	Pos       *common.Position
	RWMutex   sync.RWMutex
}

func (c *ClusterInfos) registerCluster(clusterName string, clusterInfo *Cluster) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.Clusters[clusterName] = clusterInfo
}

func (c *ClusterInfos) getCluster(clusterName string) (*Cluster, error) {
	if cluster, exists := c.Clusters[clusterName]; exists {
		return cluster.getCluster()
	}
	return nil, fmt.Errorf("cluster name is unregistered : %v", clusterName)
}

func (c *Cluster) getCluster() (*Cluster, error) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	return c, nil
}

func (c *ClusterInfos) updateCluster(clusterName string, newCluster *Cluster) error {
	c.Clusters[clusterName].RWMutex.Lock()
	defer c.Clusters[clusterName].RWMutex.Unlock()
	c.Clusters[clusterName] = &Cluster{
		Command:   newCluster.Command,
		Business:  newCluster.Business,
		HeartBeat: newCluster.HeartBeat,
		FaultInfo: newCluster.FaultInfo,
		Pos:       newCluster.Pos,
	}
	return nil
}
