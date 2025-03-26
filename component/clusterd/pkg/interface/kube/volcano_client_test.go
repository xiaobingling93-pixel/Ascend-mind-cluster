// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

func TestNewVCClientK8s(t *testing.T) {
	convey.Convey("Test newVCClientK8s", t, func() {
		convey.Convey("Should return error when BuildConfigFromFlags fails", func() {
			patches := gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(_, _ string) (*rest.Config, error) {
				return nil, errors.New("config error")
			})
			defer patches.Reset()

			client, err := newVCClientK8s()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(client, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when NewForConfig fails", func() {
			patches := gomonkey.ApplyFunc(versioned.NewForConfig, func(_ *rest.Config) (*versioned.Clientset, error) {
				return nil, errors.New("client error")
			})
			defer patches.Reset()

			client, err := newVCClientK8s()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(client, convey.ShouldBeNil)
		})
	})
}
