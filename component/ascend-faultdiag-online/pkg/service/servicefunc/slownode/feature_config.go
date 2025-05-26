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

// Package slownode a series of function
package slownode

import (
	"reflect"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/global"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	sm "ascend-faultdiag-online/pkg/module/slownode"
)

var (
	jobFuncList                  = []func(*slownode.SlowNodeJob, *slownode.SlowNodeJob, string){}
	nodeSlowNodeAlgoResFuncList  = []func(*slownode.NodeSlowNodeAlgoResult, *slownode.NodeSlowNodeAlgoResult, string){}
	nodeDataProfilingResFuncList = []func(
		*slownode.NodeDataProfilingResult, *slownode.NodeDataProfilingResult, string){}
	informerCh = make(chan struct{})

	informerHandlers = []InformerHandlerType{}
)

// InformerHandlerType register event handler for informer
type InformerHandlerType struct {
	filterFunc func(any) bool
	handler    func(any, any, string)
}

func registerHandlers(filter func(any) bool, handler func(any, any, string)) {
	informerHandlers = append(informerHandlers, InformerHandlerType{filter, handler})
}

// StopInformer stop informer when loss-leader
func StopInformer() {
	if informerCh != nil {
		close(informerCh)
		return
	}
	hwlog.RunLog.Warn("[FD-OL SLOWNODE]stop CM informer: channel is nil will not close it")
}

// CleanFunc clean func when loss-leader
func CleanFunc() {
	jobFuncList = jobFuncList[:0]
	nodeSlowNodeAlgoResFuncList = nodeSlowNodeAlgoResFuncList[:0]
	nodeDataProfilingResFuncList = nodeDataProfilingResFuncList[:0]
}

// AddCMHandler Add one or one more func in local funcList
func AddCMHandler[T any](handlers *[]func(old, new *T, op string), newHandlers ...func(old, new *T, op string)) {
	*handlers = append(*handlers, newHandlers...)
}

// InitCMInformer init configmap informer
func InitCMInformer() {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(global.K8sClient.ClientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = sm.CmConsumer + "=" + sm.CmConsumerValue
		}))
	cmInformer := informerFactory.Core().V1().ConfigMaps().Informer()

	for _, item := range informerHandlers {
		informerHandler := item
		cmInformer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: informerHandler.filterFunc,
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj any) {
					informerHandler.handler(nil, obj, AddOperator)
				},
				UpdateFunc: func(oldObj, newObj any) {
					if !reflect.DeepEqual(oldObj, newObj) {
						informerHandler.handler(oldObj, newObj, UpdateOperator)
					}
				},
				DeleteFunc: func(obj any) {
					informerHandler.handler(nil, obj, DeleteOperator)
				},
			},
		})
	}
	hwlog.RunLog.Info("[FD-OL SLOWNODE]started to watch configMap.")
	informerFactory.Start(informerCh)
}

func slowNodeFeatureHandler(oldObj, newObj any, operator string) {
	genericHandler(oldObj, newObj, operator, sm.SlowNodeFatureCMKey, jobFuncList)
}

func nodeSlowNodeJobHandler(oldObj, newObj any, operator string) {
	genericHandler(oldObj, newObj, operator, sm.NodeSlowNodeJobCMKey, jobFuncList)
}

func nodeSlowNodeAlgoHandler(oldObj, newObj any, operator string) {
	genericHandler(oldObj, newObj, operator, sm.NodeSlowNodeAlgoResultCMKey, nodeSlowNodeAlgoResFuncList)
}

func nodeDataProfilingHandler(oldObj, newObj any, operator string) {
	genericHandler(oldObj, newObj, operator, sm.NodeDataProfilingResultCMKey, nodeDataProfilingResFuncList)
}

func genericHandler[T slownode.NodeSlowNodeAlgoResult | slownode.NodeDataProfilingResult | slownode.SlowNodeJob](
	oldObj, newObj any, operator string, cmKey string, handlerFuncs []func(*T, *T, string)) {
	var oldObjTyped, newObjTyped *T

	if oldObj != nil {
		oldObjTyped = new(T)
		if err := ParseCMResult(oldObj, cmKey, oldObjTyped); err != nil {
			hwlog.RunLog.Errorf("[FD-OL SLOWNODE]parsed old cm: %v failed: %v", oldObj, err)
			return
		}
	}

	newObjTyped = new(T)
	if err := ParseCMResult(newObj, cmKey, newObjTyped); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]parsed new cm: %v failed: %v", newObj, err)
		return
	}
	for _, f := range handlerFuncs {
		f(oldObjTyped, newObjTyped, operator)
	}
}

// filterSlowNodeFeature filter func for slow node feature
func filterSlowNodeFeature(obj any) bool {
	return IsNameMatched(obj, sm.SlowNodeFeaturePrefix)
}

// filterSlowNodeFeature filter func for slow node feature
func filterNodeSlowNodeJob(obj any) bool {
	return IsNameMatched(obj, sm.NodeSlowNodeJobPrefix)
}

// filterSlowNodeAlgoResult filter func for slow node algo result
func filterSlowNodeAlgoResult(obj any) bool {
	return IsNameMatched(obj, sm.NodeSlowNodeAlgoResultPrefix)
}

// filterDataProfilingResult filter func for data profiling result
func filterDataProfilingResult(obj any) bool {
	return IsNameMatched(obj, sm.NodeDataProfilingResultPrefix)
}

// IsNameMatched check whether its namespace and name match the configmap
func IsNameMatched(obj any, namePrefix string) bool {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]cannot convert obj: %v to ConfigMap", obj)
		return false
	}
	return strings.HasPrefix(cm.Name, namePrefix)
}
