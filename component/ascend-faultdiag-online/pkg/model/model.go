/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package model is common struct
package model

// HcclJson is the chip info the training job uses
type HcclJson struct {
	// ServerList is a node info list which traning job uses
	ServerList []struct {
		// ServerId is the ip of node
		ServerId string `json:"server_id"`
		// ServerSn is the series number of node
		ServerSn string `json:"server_sn"`
		// Device is a list including rank info
		Device []struct {
			// RankId is the rank id which traning job uses
			RankId string `json:"rank_id"`
		} `json:"device"`
	} `json:"server_list"`
}

// JobSummary is a struct for cm data in job-summary
type JobSummary struct {
	// JobName name of the slow node detection job
	JobName string
	// JobId uniqe id of each job, got from job-summary
	JobId string
	// Namespace is the namespace of the job
	Namespace string
	// JobStatus is the status of the job, pending/running/complete/failed
	JobStatus string
	// HcclJson is the chip info the training job using
	HcclJson HcclJson
	// Operator add/delete
	Operator string
}
