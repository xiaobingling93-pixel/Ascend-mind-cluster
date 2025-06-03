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

// Package manager for taskd manager backend
package manager

import (
	"context"
	"fmt"

	"ascend-common/common-utils/hwlog"
	"taskd/common/utils"
)

// ClusterInfo define the information from the cluster
type ClusterInfo struct {
	// IP indicate cluster server ip
	Ip string `json:"ip"`
	// Port indicate cluster server port
	Port string `json:"port"`
	// Name indicate cluster server service name
	Name string `json:"name"`
	// Role
	Role string `json:"role"`
}

// Config define the configuration of manager
type Config struct {
	// JobId indicate the id of the job where the manager is located
	JobId string `json:"job_id"`
	// NodeNums indicate the number of nodes where the manager is located
	NodeNums int `json:"node_nums"`
	// ProcPerNode indicate the number of business processes where the manager's job is located
	ProcPerNode int `json:"proc_per_node"`
	// PluginDir indicate the plugin dir
	PluginDir string `json:"plugin_dir"`
	// ClusterInfos indicate the information of cluster
	ClusterInfos []ClusterInfo `json:"cluster_infos"`
}

// NewTaskDManager return taskd manager instance
func NewTaskDManager(config Config) *BaseManager {
	return &BaseManager{
		Config: config,
	}
}

// BaseManager the class taskd manager backend
type BaseManager struct {
	Config
}

// Init base manger
func (m *BaseManager) Init() error {
	if err := utils.InitHwLogger("manager.log", context.Background()); err != nil {
		fmt.Printf("manager init hwlog failed, err: %v \n", err)
		return err
	}
	hwlog.RunLog.Info("manager init success!")
	return nil
}

// Start taskd manager
func (m *BaseManager) Start() error {
	if err := m.Init(); err != nil {
		fmt.Printf("manager init failed, err: %v \n", err)
		return fmt.Errorf("manager init failed, err: %v", err)
	}
	if err := m.Process(); err != nil {
		hwlog.RunLog.Errorf("manager process failed, err: %v", err)
		return fmt.Errorf("manager process failed, err: %v", err)
	}
	return nil
}

// Process task main process
func (m *BaseManager) Process() error {
	return nil
}
