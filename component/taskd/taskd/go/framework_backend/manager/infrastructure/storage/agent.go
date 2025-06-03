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

	"taskd/toolkit_backend/net/common"
)

// AgentInfos all agent infos
type AgentInfos struct {
	Agents    map[string]*Agent
	AllStatus map[string]string
	RWMutex   sync.RWMutex
}

// Agent the agent info
type Agent struct {
	Config    map[string]string
	Actions   map[string]string
	Status    map[string]string
	NodeRank  string
	HeartBeat time.Time
	FaultInfo map[string]string
	Pos       *common.Position
	RWMutex   sync.RWMutex
}

func (a *AgentInfos) registerAgent(agentName string, agentInfo *Agent) error {
	a.RWMutex.Lock()
	defer a.RWMutex.Unlock()
	a.Agents[agentName] = agentInfo
	return nil
}

func (a *AgentInfos) getAgent(agentName string) (*Agent, error) {
	if agent, exists := a.Agents[agentName]; exists {
		return agent.getAgent()
	}
	return nil, fmt.Errorf("agent name is unregistered : %v", agentName)
}

func (a *Agent) getAgent() (*Agent, error) {
	a.RWMutex.RLock()
	defer a.RWMutex.RUnlock()
	return a, nil
}

func (a *AgentInfos) updateAgent(agentName string, newAgent *Agent) error {
	a.Agents[agentName].RWMutex.Lock()
	defer a.Agents[agentName].RWMutex.Unlock()
	a.Agents[agentName] = &Agent{
		Config:    newAgent.Config,
		Actions:   newAgent.Actions,
		Status:    newAgent.Status,
		NodeRank:  newAgent.NodeRank,
		HeartBeat: newAgent.HeartBeat,
		FaultInfo: newAgent.FaultInfo,
		Pos:       newAgent.Pos,
	}
	return nil
}
