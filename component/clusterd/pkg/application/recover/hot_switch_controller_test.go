// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recover a series of service function
package recover

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

type hotSwitchTestCase struct {
	name                  string
	controller            *EventController
	mockFaultRanks        []*pb.FaultRank
	mockSignalEnqueueResp string
	mockSignalEnqueueCode common.RespCode
	mockSignalEnqueueErr  error
	expectedEvent         string
	expectedCode          common.RespCode
	expectedError         error
}

type signalTestConfig struct {
	signalType string
	actions    []string
	strategy   string
}

func buildHotSwitchTestCases() []hotSwitchTestCase {
	return []hotSwitchTestCase{
		{
			name:                  "successful hot switch preparation",
			controller:            &EventController{jobInfo: common.JobBaseInfo{JobId: "test-job-1"}, uuid: "test-uuid-1"},
			mockFaultRanks:        []*pb.FaultRank{{RankId: "0"}},
			mockSignalEnqueueResp: "success",
			mockSignalEnqueueCode: common.OK,
			mockSignalEnqueueErr:  nil,
			expectedEvent:         "success",
			expectedCode:          common.OK,
			expectedError:         nil,
		},
		{
			name:                  "hot switch preparation with error",
			controller:            &EventController{jobInfo: common.JobBaseInfo{JobId: "test-job-2"}, uuid: "test-uuid-2"},
			mockFaultRanks:        []*pb.FaultRank{{RankId: "1"}},
			mockSignalEnqueueResp: "",
			mockSignalEnqueueCode: common.ServerInnerError,
			mockSignalEnqueueErr:  nil,
			expectedEvent:         "",
			expectedCode:          common.ServerInnerError,
			expectedError:         nil,
		},
	}
}

func buildSignalTestCases(signalType string, actions []string, strategy string) []hotSwitchTestCase {
	return []hotSwitchTestCase{
		{
			name:                  "successful " + signalType + " handling",
			controller:            &EventController{jobInfo: common.JobBaseInfo{JobId: "test-job-1"}, uuid: "test-uuid-1"},
			mockSignalEnqueueResp: "success",
			mockSignalEnqueueCode: common.OK,
			mockSignalEnqueueErr:  nil,
			expectedEvent:         "success",
			expectedCode:          common.OK,
			expectedError:         nil,
		},
		{
			name:                  signalType + " handling with error",
			controller:            &EventController{jobInfo: common.JobBaseInfo{JobId: "test-job-2"}, uuid: "test-uuid-2"},
			mockSignalEnqueueResp: "",
			mockSignalEnqueueCode: common.ServerInnerError,
			mockSignalEnqueueErr:  nil,
			expectedEvent:         "",
			expectedCode:          common.ServerInnerError,
			expectedError:         nil,
		},
	}
}

func runSignalTest(t *testing.T, tests []hotSwitchTestCase, testFunc func(*EventController) (string, common.RespCode, error), config signalTestConfig) {
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			patch := gomonkey.ApplyPrivateMethod((*EventController)(nil), "signalEnqueue",
				func(_ *EventController, signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
					convey.So(signal.Uuid, convey.ShouldEqual, tt.controller.uuid)
					convey.So(signal.JobId, convey.ShouldEqual, tt.controller.jobInfo.JobId)
					convey.So(signal.SignalType, convey.ShouldEqual, config.signalType)
					convey.So(signal.Actions, convey.ShouldResemble, config.actions)
					convey.So(signal.ChangeStrategy, convey.ShouldEqual, config.strategy)
					return tt.mockSignalEnqueueResp, tt.mockSignalEnqueueCode, tt.mockSignalEnqueueErr
				})
			defer patch.Reset()
			event, code, err := testFunc(tt.controller)
			convey.So(event, convey.ShouldEqual, tt.expectedEvent)
			convey.So(code, convey.ShouldEqual, tt.expectedCode)
			convey.So(err, convey.ShouldEqual, tt.expectedError)
		})
	}
}

func TestNotifyPrepareHotSwitch(t *testing.T) {
	tests := buildHotSwitchTestCases()
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			patch := gomonkey.NewPatches()
			defer patch.Reset()
			patch.ApplyPrivateMethod(tt.controller, "normalFaultAssociateSameNodeRank",
				func(_ *EventController) ([]*pb.FaultRank, map[string]string) {
					return tt.mockFaultRanks, nil
				}).ApplyPrivateMethod(tt.controller, "signalEnqueue",
				func(_ *EventController, signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
					convey.So(signal.Uuid, convey.ShouldEqual, tt.controller.uuid)
					convey.So(signal.JobId, convey.ShouldEqual, tt.controller.jobInfo.JobId)
					convey.So(signal.SignalType, convey.ShouldEqual, constant.HotSwitchSignalType)
					convey.So(signal.Actions, convey.ShouldResemble, hotSwitchActions)
					convey.So(signal.FaultRanks, convey.ShouldResemble, tt.mockFaultRanks)
					convey.So(signal.ChangeStrategy, convey.ShouldEqual, "")
					return tt.mockSignalEnqueueResp, tt.mockSignalEnqueueCode, tt.mockSignalEnqueueErr
				})

			event, code, err := tt.controller.notifyPrepareHotSwitch()

			convey.So(event, convey.ShouldEqual, tt.expectedEvent)
			convey.So(code, convey.ShouldEqual, tt.expectedCode)
			convey.So(err, convey.ShouldEqual, tt.expectedError)
		})
	}
}

func TestNotifyNewPodFailedHandler(t *testing.T) {
	tests := buildSignalTestCases("new pod failed", stopHotSwitchActions, "")
	config := signalTestConfig{
		signalType: constant.HotSwitchSignalType,
		actions:    stopHotSwitchActions,
		strategy:   "",
	}
	testFunc := func(controller *EventController) (string, common.RespCode, error) {
		return controller.notifyNewPodFailedHandler()
	}
	runSignalTest(t, tests, testFunc, config)
}

func TestNotifyNewPodRunningHandler(t *testing.T) {
	tests := buildSignalTestCases("new pod running", newPodRunningActions, "")
	config := signalTestConfig{
		signalType: constant.HotSwitchSignalType,
		actions:    newPodRunningActions,
		strategy:   "",
	}
	testFunc := func(controller *EventController) (string, common.RespCode, error) {
		return controller.notifyNewPodRunningHandler()
	}
	runSignalTest(t, tests, testFunc, config)
}

func TestNotifyRestartTrain(t *testing.T) {
	tests := buildSignalTestCases("restart train", changeStrategyActions, constant.ProcessMigration)
	config := signalTestConfig{
		signalType: constant.ChangeStrategySignalType,
		actions:    changeStrategyActions,
		strategy:   constant.ProcessMigration,
	}
	testFunc := func(controller *EventController) (string, common.RespCode, error) {
		return controller.notifyRestartTrain()
	}
	runSignalTest(t, tests, testFunc, config)
}

type hotSwitchTestCaseB struct {
	name            string
	podName         string
	podExist        bool
	annotationExist bool
	deleteErr       error
	newPodExist     bool
	patchErr        error
	expectedCode    common.RespCode
	expectError     bool
	expectedEvent   string
	labelErr        error
	faultPods       map[string]string
}

func TestCleanStateWhenFailed(t *testing.T) {
	ctl := buildBaseController()
	testCases := buildHotSwitchTestCase1()
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			doPatches(patches, ctl, tc)
			event, code, err := ctl.cleanStateWhenFailed()
			convey.So(code, convey.ShouldEqual, tc.expectedCode)
			convey.So(event, convey.ShouldEqual, "")
			if tc.expectError {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func doPatches(patches *gomonkey.Patches, ctl *EventController, tc hotSwitchTestCaseB) {
	patchReset(patches, ctl)
	if !tc.podExist {
		patchGetPodByPodId(patches, v1.Pod{}, false)
	} else {
		testPod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "fault-pod", Namespace: "default"}}
		if tc.annotationExist {
			testPod.Annotations = map[string]string{api.InHotSwitchFlowKey: "true", api.BackupNewPodNameKey: "new-pod"}
		} else {
			testPod.Annotations = make(map[string]string)
		}
		patchGetPodByPodId(patches, testPod, true)
		patches.ApplyFuncReturn(kube.DeletePodAnnotation, tc.deleteErr)
		if tc.newPodExist {
			patches.ApplyFuncReturn(pod.GetPodByJobIdAndPodName,
				v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "new-pod", Namespace: "default"}}, true)
		} else {
			patches.ApplyFuncReturn(pod.GetPodByJobIdAndPodName, v1.Pod{}, false)
		}
		patches.ApplyFuncReturn(kube.RetryPatchPodAnnotations, tc.patchErr)
	}
}

func patchGetPodByPodId(patches *gomonkey.Patches, p v1.Pod, exist bool) *gomonkey.Patches {
	return patches.ApplyFuncReturn(pod.GetPodByPodId, p, exist)
}

func buildBaseController() *EventController {
	ctl := &EventController{
		currentHotSwitchFaultPodId: "test-pod-id",
		jobInfo:                    common.JobBaseInfo{JobId: "test-job-id"},
	}
	return ctl
}

func buildHotSwitchTestCase1() []hotSwitchTestCaseB {
	testCases := []hotSwitchTestCaseB{
		{name: "pod not exist",
			podExist:     false,
			expectedCode: common.OK,
			expectError:  false},
		{name: "annotation not exist",
			podExist:        true,
			annotationExist: false,
			expectedCode:    common.OK,
			expectError:     false},
		{name: "delete annotation failed",
			podExist:        true,
			annotationExist: true,
			deleteErr:       errors.New("delete failed"),
			expectedCode:    common.ServerInnerError,
			expectError:     true},
		{name: "new pod not exist",
			podExist:        true,
			annotationExist: true,
			deleteErr:       nil,
			newPodExist:     false,
			expectedCode:    common.ServerInnerError,
			expectError:     true},
		{name: "patch new pod failed",
			podExist:        true,
			annotationExist: true,
			deleteErr:       nil,
			newPodExist:     true,
			patchErr:        errors.New("patch failed"),
			expectedCode:    common.ServerInnerError,
			expectError:     true},
		{name: "success",
			podExist:        true,
			annotationExist: true,
			deleteErr:       nil,
			newPodExist:     true,
			patchErr:        nil,
			expectedCode:    common.OK,
			expectError:     false},
	}
	return testCases
}

func TestCleanStateWhenSuccess(t *testing.T) {
	ctl := &EventController{
		currentHotSwitchBackupPodId: "test-backup-pod-id",
		jobInfo:                     common.JobBaseInfo{JobId: "test-job-id"},
	}
	testCases := buildHotSwitchTestCases2()
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patchReset(patches, ctl)
			if !tc.podExist {
				patchGetPodByPodId(patches, v1.Pod{}, false)
			} else {
				testPod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "backup-pod", Namespace: "default"}}
				patchGetPodByPodId(patches, testPod, true)
				patches.ApplyFuncReturn(kube.DeletePodAnnotation, tc.deleteErr)
			}
			event, code, err := ctl.cleanStateWhenSuccess()
			convey.So(code, convey.ShouldEqual, tc.expectedCode)
			convey.So(event, convey.ShouldEqual, "")
			if tc.expectError {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func buildHotSwitchTestCases2() []hotSwitchTestCaseB {
	testCases := []hotSwitchTestCaseB{
		{
			name:         "pod not exist",
			podExist:     false,
			expectedCode: common.OK,
			expectError:  false,
		}, {
			name:         "delete annotation failed",
			podExist:     true,
			deleteErr:    errors.New("delete failed"),
			expectedCode: common.ServerInnerError,
			expectError:  true,
		}, {
			name:         "success",
			podExist:     true,
			deleteErr:    nil,
			expectedCode: common.OK,
			expectError:  false,
		},
	}
	return testCases
}

func TestNotifyStopJob(t *testing.T) {
	ctl := &EventController{
		jobInfo: common.JobBaseInfo{
			JobId:   "test-job-id",
			JobName: "test-job-name",
		},
	}

	testCases := []hotSwitchTestCaseB{
		{
			name:          "pod not found",
			podName:       "",
			expectError:   false,
			expectedCode:  common.OK,
			expectedEvent: common.NotifySuccessEvent,
		},
		{
			name:          "pod found and stop job",
			podName:       "test-pod",
			expectError:   false,
			expectedCode:  common.OK,
			expectedEvent: common.NotifySuccessEvent,
		},
	}

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patchReset(patches, ctl)
			patches.ApplyFunc(pod.GetPodByRankIndex, func(jobId string, rankIndex string) v1.Pod {
				return v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tc.podName,
						Namespace: "default",
					},
				}
			})
			patches.ApplyFuncReturn(kube.RetryPatchPodAnnotations, nil)
			patches.ApplyFuncReturn(kube.RetryPatchPodLabels, nil)
			event, code, err := ctl.notifyStopJob()
			convey.So(code, convey.ShouldEqual, tc.expectedCode)
			convey.So(event, convey.ShouldEqual, tc.expectedEvent)
			if tc.expectError {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func patchReset(patches *gomonkey.Patches, ctl *EventController) *gomonkey.Patches {
	return patches.ApplyPrivateMethod(ctl, "reset", func(bool) {})
}

func TestNotifyDeleteOldPod(t *testing.T) {
	controller := &EventController{
		currentHotSwitchFaultPodId: "test-pod-id",
		jobInfo:                    common.JobBaseInfo{JobId: "test-job-id"},
	}
	testCases := buildHotSwitchTestCase3()

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patchReset(patches, controller)
			if !tc.podExist {
				patchGetPodByPodId(patches, v1.Pod{}, false)
			} else {
				testPod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"}}
				patchGetPodByPodId(patches, testPod, true)
				patches.ApplyFuncReturn(kube.RetryPatchPodAnnotations, tc.patchErr)
				patches.ApplyFuncReturn(kube.RetryPatchPodLabels, tc.labelErr)
			}
			event, code, err := controller.notifyDeleteOldPod()
			convey.So(code, convey.ShouldEqual, tc.expectedCode)
			convey.So(event, convey.ShouldEqual, "")
			if tc.expectError {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func buildHotSwitchTestCase3() []hotSwitchTestCaseB {
	testCases := []hotSwitchTestCaseB{
		{
			name:         "pod not exist",
			podExist:     false,
			expectedCode: common.OK,
			expectError:  false,
		}, {
			name:         "patch annotation failed",
			podExist:     true,
			patchErr:     errors.New("patch annotation failed"),
			expectedCode: common.OK,
			expectError:  false,
		}, {
			name:         "patch label failed",
			podExist:     true,
			labelErr:     errors.New("patch label failed"),
			expectedCode: common.OK,
			expectError:  false,
		}, {
			name:         "success",
			podExist:     true,
			patchErr:     nil,
			labelErr:     nil,
			expectedCode: common.OK,
			expectError:  false},
	}
	return testCases
}

func TestNotifyCreateNewPod(t *testing.T) {
	controller := &EventController{
		jobInfo: common.JobBaseInfo{JobId: "test-job-id"},
	}
	testCases := buildHotSwitchTestCases3()
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethodReturn(controller, "GetFaultPod", tc.faultPods)
			if len(tc.faultPods) == 0 {
				// do nothing
			} else if !tc.podExist {
				patchGetPodByPodId(patches, v1.Pod{}, false)
			} else {
				testPod := v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
						UID:       "test-pod-uid",
					},
				}
				patchGetPodByPodId(patches, testPod, true)
				patches.ApplyFuncReturn(kube.RetryPatchPodAnnotations, tc.patchErr)
			}
			event, code, err := controller.notifyCreateNewPod()
			convey.So(code, convey.ShouldEqual, tc.expectedCode)
			convey.So(event, convey.ShouldEqual, "")
			if tc.expectError {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
			if tc.patchErr == nil && len(tc.faultPods) > 0 && tc.podExist {
				convey.So(controller.currentHotSwitchFaultPodId, convey.ShouldEqual, "test-pod-uid")
			}
		})
	}
}

func buildHotSwitchTestCases3() []hotSwitchTestCaseB {
	testCases := []hotSwitchTestCaseB{
		{
			name:         "no fault pods",
			faultPods:    map[string]string{},
			expectedCode: common.OK,
			expectError:  false,
		},
		{
			name:         "pod not exist",
			faultPods:    map[string]string{"0": "test-pod-id"},
			podExist:     false,
			expectedCode: common.OK,
			expectError:  false,
		},
		{
			name:         "patch annotation failed",
			faultPods:    map[string]string{"0": "test-pod-id"},
			podExist:     true,
			patchErr:     errors.New("patch failed"),
			expectedCode: common.OK,
			expectError:  false,
		},
		{
			name:         "success",
			faultPods:    map[string]string{"0": "test-pod-id"},
			podExist:     true,
			patchErr:     nil,
			expectedCode: common.OK,
			expectError:  false,
		},
	}
	return testCases
}
