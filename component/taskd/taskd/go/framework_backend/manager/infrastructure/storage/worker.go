/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain w copy of the License at

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

// WorkerInfos all worker infos
type WorkerInfos struct {
	Workers   map[string]*Worker
	AllStatus map[string]string
	RWMutex   sync.RWMutex
}

// Worker the worker info
type Worker struct {
	Config     map[string]string
	Actions    map[string]string
	Status     map[string]string
	GlobalRank string
	HeartBeat  time.Time
	FaultInfo  map[string]string
	Pos        *common.Position
	RWMutex    sync.RWMutex
}

func (w *WorkerInfos) registerWorker(workerName string, workerInfo *Worker) error {
	w.RWMutex.Lock()
	defer w.RWMutex.Unlock()
	w.Workers[workerName] = workerInfo
	return nil
}

func (w *WorkerInfos) getWorker(workerName string) (*Worker, error) {
	if worker, exists := w.Workers[workerName]; exists {
		return worker.getWorker()
	}
	return nil, fmt.Errorf("worker name is unregistered : %v", workerName)
}

func (w *Worker) getWorker() (*Worker, error) {
	w.RWMutex.RLock()
	defer w.RWMutex.RUnlock()
	return w, nil
}

func (w *WorkerInfos) updateWorker(workerName string, newWorker *Worker) error {
	w.Workers[workerName].RWMutex.Lock()
	defer w.Workers[workerName].RWMutex.Unlock()
	w.Workers[workerName] = &Worker{
		Config:     newWorker.Config,
		Actions:    newWorker.Actions,
		Status:     newWorker.Status,
		GlobalRank: newWorker.GlobalRank,
		HeartBeat:  newWorker.HeartBeat,
		FaultInfo:  newWorker.FaultInfo,
		Pos:        newWorker.Pos,
	}
	return nil
}
