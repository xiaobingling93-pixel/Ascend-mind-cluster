// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package api common brand moniker
package api

// common
const (
	// Pod910DeviceAnno annotation value is for generating 910 hccl rank table
	Pod910DeviceAnno = "ascend.kubectl.kubernetes.io/ascend-910-configuration"

	// ResourceNamePrefix pre resource name
	ResourceNamePrefix = "huawei.com/"

	// Ascend910 for 910 chip
	Ascend910 = "ascend910"

	// ASCEND310 ascend 310 chip
	ASCEND310 = "Ascend310"
	// ASCEND310P ascend 310P chip
	ASCEND310P = "Ascend310P"
	// ASCEND310B ascend 310B chip
	ASCEND310B = "Ascend310B"
	// ASCEND910 ascend 910 chip
	ASCEND910 = "Ascend910"
	// ASCEND910B ascend 910B chip
	ASCEND910B = "Ascend910B"
	// ASCEND910A3 ascend 910A3 chip
	ASCEND910A3 = "Ascend910A3"
)

// device plugin
const (
	// Use310PMixedInsert use 310P Mixed insert
	Use310PMixedInsert = "use310PMixedInsert"
	// Ascend310PMix dp use310PMixedInsert parameter usage
	Ascend310PMix = "ascend310P-V, ascend310P-VPro, ascend310P-IPro"
)

// npu exporter
const (
	// DevicePathPattern device path pattern
	DevicePathPattern = `^/dev/davinci\d+$`
	// HccsBWProfilingTimeStr  preset parameter name
	HccsBWProfilingTimeStr = "hccsBWProfilingTime"
	// Hccs log options domain value
	Hccs = "hccs"
	// Prefix pre statistic info
	Prefix = "npu_chip_info_hccs_statistic_info_"
	// BwPrefix pre bandwidth info
	BwPrefix = "npu_chip_info_hccs_bandwidth_info_"
)
