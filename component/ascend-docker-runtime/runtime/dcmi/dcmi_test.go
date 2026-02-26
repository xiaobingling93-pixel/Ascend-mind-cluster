/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package dcmi

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockDcDriver implements DcDriverV1Interface for testing
type MockDcDriver struct {
	// initializeFunc is a mock function for DcInitialize
	initializeFunc func() error
	// shutdownFunc is a mock function for DcShutDown
	shutdownFunc func()
	// getCardListFunc is a mock function for DcGetCardList
	getCardListFunc func() (int32, []int32, error)
	// getDeviceNumFunc is a mock function for DcGetDeviceNumInCard
	getDeviceNumFunc func(cardID int32) (int32, error)
	// getLogicIDFunc is a mock function for DcGetDeviceLogicID
	getLogicIDFunc func(cardID, deviceID int32) (int32, error)
	// getProductTypeFunc is a mock function for DcGetProductType
	getProductTypeFunc func(cardID, deviceID int32) (string, error)
	// createVDeviceFunc is a mock function for DcCreateVDevice
	createVDeviceFunc func(cardID, deviceID int32, coreNum string) (int32, error)
	// destroyVDeviceFunc is a mock function for DcDestroyVDevice
	destroyVDeviceFunc func(cardID, deviceID int32, vDevID int32) error
	// getChipInfoFunc is a mock function for DcGetChipInfo
	getChipInfoFunc func(cardID, deviceID int32) (*ChipInfo, error)
}

// DcInitialize mocks the initialization function
func (m *MockDcDriver) DcInitialize() error {
	if m.initializeFunc != nil {
		return m.initializeFunc()
	}
	return nil
}

// DcShutDown mocks the shutdown function
func (m *MockDcDriver) DcShutDown() {
	if m.shutdownFunc != nil {
		m.shutdownFunc()
	}
}

// DcGetCardList mocks getting the card list
func (m *MockDcDriver) DcGetCardList() (int32, []int32, error) {
	if m.getCardListFunc != nil {
		return m.getCardListFunc()
	}
	return 0, []int32{}, nil
}

// DcGetDeviceNumInCard mocks getting the number of devices in a card
func (m *MockDcDriver) DcGetDeviceNumInCard(cardID int32) (int32, error) {
	if m.getDeviceNumFunc != nil {
		return m.getDeviceNumFunc(cardID)
	}
	return 0, nil
}

// DcGetDeviceLogicID mocks getting the logic ID of a device
func (m *MockDcDriver) DcGetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	if m.getLogicIDFunc != nil {
		return m.getLogicIDFunc(cardID, deviceID)
	}
	return 0, nil
}

// DcGetProductType mocks getting the product type
func (m *MockDcDriver) DcGetProductType(cardID, deviceID int32) (string, error) {
	if m.getProductTypeFunc != nil {
		return m.getProductTypeFunc(cardID, deviceID)
	}
	return "", nil
}

// DcCreateVDevice mocks creating a virtual device
func (m *MockDcDriver) DcCreateVDevice(cardID, deviceID int32, coreNum string) (int32, error) {
	if m.createVDeviceFunc != nil {
		return m.createVDeviceFunc(cardID, deviceID, coreNum)
	}
	return 0, nil
}

// DcDestroyVDevice mocks destroying a virtual device
func (m *MockDcDriver) DcDestroyVDevice(cardID, deviceID int32, vDevID int32) error {
	if m.destroyVDeviceFunc != nil {
		return m.destroyVDeviceFunc(cardID, deviceID, vDevID)
	}
	return nil
}

// DcGetChipInfo mocks getting chip information
func (m *MockDcDriver) DcGetChipInfo(cardID, deviceID int32) (*ChipInfo, error) {
	if m.getChipInfoFunc != nil {
		return m.getChipInfoFunc(cardID, deviceID)
	}
	return nil, nil
}

// MockA950Driver implements DcDriverV2Interface for testing
type MockA950Driver struct {
	// initializeFunc is a mock function for DcInitialize
	initializeFunc func() error
	// shutdownFunc is a mock function for DcShutDown
	shutdownFunc func()
	// getDeviceListFunc is a mock function for DcGetDeviceList
	getDeviceListFunc func() (int32, []int32, error)
	// createVDeviceFunc is a mock function for DcCreateVDevice
	createVDeviceFunc func(deviceID int32, coreNum string) (int32, error)
	// destroyVDeviceFunc is a mock function for DcDestroyVDevice
	destroyVDeviceFunc func(deviceID int32, vDevID int32) error
	// getChipInfoFunc is a mock function for DcGetChipInfo
	getChipInfoFunc func(deviceID int32) (*ChipInfo, error)
}

// DcInitialize mocks the initialization function for A950 driver
func (m *MockA950Driver) DcInitialize() error {
	if m.initializeFunc != nil {
		return m.initializeFunc()
	}
	return nil
}

// DcShutDown mocks the shutdown function for A950 driver
func (m *MockA950Driver) DcShutDown() {
	if m.shutdownFunc != nil {
		m.shutdownFunc()
	}
}

// DcGetDeviceList mocks getting the device list for A950
func (m *MockA950Driver) DcGetDeviceList() (int32, []int32, error) {
	if m.getDeviceListFunc != nil {
		return m.getDeviceListFunc()
	}
	return 0, []int32{}, nil
}

// DcCreateVDevice mocks creating a virtual device for A950
func (m *MockA950Driver) DcCreateVDevice(deviceID int32, coreNum string) (int32, error) {
	if m.createVDeviceFunc != nil {
		return m.createVDeviceFunc(deviceID, coreNum)
	}
	return 0, nil
}

// DcDestroyVDevice mocks destroying a virtual device for A950
func (m *MockA950Driver) DcDestroyVDevice(deviceID int32, vDevID int32) error {
	if m.destroyVDeviceFunc != nil {
		return m.destroyVDeviceFunc(deviceID, vDevID)
	}
	return nil
}

// DcGetChipInfo mocks getting chip information for A950
func (m *MockA950Driver) DcGetChipInfo(deviceID int32) (*ChipInfo, error) {
	if m.getChipInfoFunc != nil {
		return m.getChipInfoFunc(deviceID)
	}
	return nil, nil
}

// MockNpuWorker extends NpuV1Worker for testing with mock driver
type MockNpuWorker struct {
	*NpuV1Worker
}

// MockA950NpuWorker extends NpuV2Worker for testing with mock driver
type MockA950NpuWorker struct {
	*NpuV2Worker
}

// Test validation functions
// TestIsValidCardID tests the isValidCardID function
func TestIsValidCardID(t *testing.T) {
	tests := []struct {
		name   string
		cardID int32
		want   bool
	}{
		{"valid card id", 0, true},
		{"valid card id max", math.MaxInt32 - 1, true},
		{"valid card id boundary", math.MaxInt32, false},
		{"invalid negative", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidCardID(tt.cardID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidDeviceID tests the isValidDeviceID function
func TestIsValidDeviceID(t *testing.T) {
	tests := []struct {
		name     string
		deviceID int32
		want     bool
	}{
		{"valid device id min", 0, true},
		{"valid device id max", 3, true},
		{"invalid negative", -1, false},
		{"invalid too large", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidDeviceID(tt.deviceID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidCardIDAndDeviceID tests the isValidCardIDAndDeviceID function
func TestIsValidCardIDAndDeviceID(t *testing.T) {
	tests := []struct {
		name     string
		cardID   int32
		deviceID int32
		want     bool
	}{
		{"both valid", 0, 0, true},
		{"card invalid", -1, 0, false},
		{"device invalid", 0, 4, false},
		{"both invalid", -1, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidCardIDAndDeviceID(tt.cardID, tt.deviceID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidA950DeviceID tests the isValidA950DeviceID function
func TestIsValidA950DeviceID(t *testing.T) {
	tests := []struct {
		name     string
		deviceID int32
		want     bool
	}{
		{"valid device id min", 0, true},
		{"valid device id max", math.MaxInt32 - 1, true},
		{"valid boundary", math.MaxInt32, false},
		{"invalid negative", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidA950DeviceID(tt.deviceID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidChipInfo tests the isValidChipInfo function
func TestIsValidChipInfo(t *testing.T) {
	tests := []struct {
		name string
		chip *ChipInfo
		want bool
	}{
		{
			name: "all fields empty",
			chip: &ChipInfo{},
			want: false,
		},
		{
			name: "only name",
			chip: &ChipInfo{Name: "Ascend310"},
			want: true,
		},
		{
			name: "only type",
			chip: &ChipInfo{Type: "AI Processor"},
			want: true,
		},
		{
			name: "only version",
			chip: &ChipInfo{Version: "1.0"},
			want: true,
		},
		{
			name: "all fields",
			chip: &ChipInfo{Name: "Ascend310", Type: "AI Processor", Version: "1.0"},
			want: true,
		},
		{
			name: "nil chip",
			chip: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidChipInfo(tt.chip)
			assert.Equal(t, tt.want, got)
		})
	}
}
