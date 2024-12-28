// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package job a series of job test function
package job

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/kube"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestInitCM(t *testing.T) {
	convey.Convey("test initCM", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when CreateOrUpdateConfigMap failed, return false", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return errors.New("test error")
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			result := initCM(jobInfo)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("when CreateOrUpdateConfigMap success, return true", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return nil
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			result := initCM(jobInfo)
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestUpdateCM(t *testing.T) {
	convey.Convey("test updateCM", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when framework is empty, index is 0, UpdateOrCreateConfigMap failed."+
			" return false", func() {
			mockUpdateOrCreateConfigMap := gomonkey.ApplyFunc(kube.UpdateOrCreateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return errors.New("test error")
				})
			defer mockUpdateOrCreateConfigMap.Reset()
			result := updateCM(jobInfo, 0, "")
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("when framework is pytorch, index is 1, UpdateOrCreateConfigMap success."+
			" return true", func() {
			mockUpdateOrCreateConfigMap := gomonkey.ApplyFunc(kube.UpdateOrCreateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return nil
				})
			defer mockUpdateOrCreateConfigMap.Reset()
			jobInfo.Framework = ptFramework
			result := updateCM(jobInfo, 1, "")
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestPreDeleteCM(t *testing.T) {
	convey.Convey("test preDeleteCM", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when framework is empty, totalCmNum is 0, CreateOrUpdateConfigMap success."+
			" return true", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return nil
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			hccls := make([]string, 0)
			result := preDeleteCM(jobInfo, hccls)
			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("when framework is pytorch, totalCmNum is 1, CreateOrUpdateConfigMap failed."+
			" return false", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return errors.New("test error")
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			jobInfo.TotalCmNum = 1
			jobInfo.Framework = ptFramework
			hccls := make([]string, 1)
			hccls[0] = "{}"
			result := preDeleteCM(jobInfo, hccls)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("when framework is pytorch, totalCmNum is 1 but hccls is nil, "+
			"CreateOrUpdateConfigMap success. return true", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return nil
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			jobInfo.TotalCmNum = 1
			jobInfo.Framework = ptFramework
			hccls := make([]string, 0)
			result := preDeleteCM(jobInfo, hccls)
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestDeleteCm(t *testing.T) {
	convey.Convey("test deleteCm", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when totalCmNum is 0, DeleteConfigMap success. return true", func() {
			mockDeleteConfigMap := gomonkey.ApplyFunc(kube.DeleteConfigMap,
				func(cmName, nameSpace string) error {
					return nil
				})
			defer mockDeleteConfigMap.Reset()
			result := deleteCm(jobInfo)
			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("when totalCmNum is 1, DeleteConfigMap failed. return true", func() {
			mockDeleteConfigMap := gomonkey.ApplyFunc(kube.DeleteConfigMap,
				func(cmName, nameSpace string) error {
					return errors.New("test error")
				})
			defer mockDeleteConfigMap.Reset()
			jobInfo.TotalCmNum = 1
			result := deleteCm(jobInfo)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("when totalCmNum is 1, DeleteConfigMap is notFound failed. return true", func() {
			mockDeleteConfigMap := gomonkey.ApplyFunc(kube.DeleteConfigMap,
				func(cmName, nameSpace string) error {
					var err = &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
					return err
				})
			defer mockDeleteConfigMap.Reset()
			jobInfo.TotalCmNum = 1
			result := deleteCm(jobInfo)
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestGetDefaultLabel(t *testing.T) {
	convey.Convey("test getDefaultLabel. return map have lable", t, func() {
		labels := getDefaultLabel()
		convey.So(labels[atlasRing], convey.ShouldEqual, val910)
		convey.So(labels[configmapLabel], convey.ShouldEqual, "true")
	})
}
