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

// Package superpod for utils function
package superpod

import (
	"sort"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// transferSuperPodToRackIdMap get a map with rackID key, and get all nodes by rackID in this superPod
func transferSuperPodToRackIdMap(superPod map[string]nodeBaseInfo) map[int32][]nodeBaseInfo {
	superPodWithRack := make(map[int32][]nodeBaseInfo)

	for _, nNode := range superPod {
		_, ok := superPodWithRack[nNode.rackID]
		if !ok {
			superPodWithRack[nNode.rackID] = make([]nodeBaseInfo, 0)
		}
		superPodWithRack[nNode.rackID] = append(superPodWithRack[nNode.rackID], nNode)
	}

	klog.V(util.LogInfoLev).Infof("the superpod with rackID info is %v", superPodWithRack)
	return superPodWithRack
}

// the param superPodWithRack contain all rackID and nodes slice with the same rackID
// return value is a slice contains many of rackIDs sorted by the len of rack nodes ascending order.
func sortRackIdByLengthInOneSuperPod(superPodWithRack map[int32][]nodeBaseInfo) []int32 {
	var rackIdOrder []int32

	for rackId := range superPodWithRack {
		rackIdOrder = append(rackIdOrder, rackId)
	}
	sort.Slice(rackIdOrder, func(i, j int) bool {
		iLen := len(superPodWithRack[rackIdOrder[i]])
		jLen := len(superPodWithRack[rackIdOrder[j]])
		return iLen < jLen
	})
	return rackIdOrder
}

// the param superPodWithRack contain all rackID and nodes slice with the same rackID
// return value is a map which key is 1-8(=n) means how many nodes are available
// and which value is a slice contains all rackID whose nodes len is equal to n
func getRackLengthInOneSuperPod(superPodWithRack map[int32][]nodeBaseInfo) map[int][]int32 {
	// key is 1-8,value is a slice contain all rackID
	nodesLenWithRackIDMap := make(map[int][]int32, rackNodeNum)

	for rackId, nodes := range superPodWithRack {
		if nodesLenWithRackIDMap[len(nodes)] == nil {
			nodesLenWithRackIDMap[len(nodes)] = make([]int32, 0)
		}
		nodesLenWithRackIDMap[len(nodes)] = append(nodesLenWithRackIDMap[len(nodes)], rackId)
	}
	return nodesLenWithRackIDMap
}

// the input length is in range [0,8], contains value in range [0,63]
// eg: input1 []int{0,1,2,3,4,5,6,7} means 8 cards can be used
// eg: input2 []int{8,9,10,11,15} means 5 cards can be used
func getUsableNPUIndex(input []int) [nodeNPUNum]bool {
	res := [nodeNPUNum]bool{}
	for _, num := range input {
		if num < rackNodeNum*nodeNPUNum {
			res[num%nodeNPUNum] = true
		}
	}
	return res
}

