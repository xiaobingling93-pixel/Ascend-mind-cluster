package faultshoot

import (
	"clusterd/pkg/common/constant"
	"reflect"
	"testing"
)

func Test_splitDeviceFault(t *testing.T) {
	t.Run("Test_splitDeviceFault", func(t *testing.T) {
		var faultInfo = constant.DeviceFault{
			FaultType:            "xx",
			NPUName:              "Ascend910-0",
			LargeModelFaultLevel: "xx",
			FaultLevel:           "xx",
			FaultHandling:        "xx",
			FaultCode:            "0x1,0x2",
			FaultTimeMap: map[string]int64{
				"0x1": 1,
				"0x2": 2,
			},
		}

		got := splitDeviceFault(faultInfo)
		want := []constant.DeviceFault{
			{
				FaultType:            "xx",
				NPUName:              "Ascend910-0",
				LargeModelFaultLevel: "xx",
				FaultLevel:           "xx",
				FaultHandling:        "xx",
				FaultCode:            "0x1",
				FaultTime:            1,
				FaultTimeMap: map[string]int64{
					"0x1": 1,
					"0x2": 2,
				},
			},
			{
				FaultType:            "xx",
				NPUName:              "Ascend910-0",
				LargeModelFaultLevel: "xx",
				FaultLevel:           "xx",
				FaultHandling:        "xx",
				FaultCode:            "0x2",
				FaultTime:            2,
				FaultTimeMap: map[string]int64{
					"0x1": 1,
					"0x2": 2,
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("splitDeviceFault() = %v, want %v", got, want)
		}
	})
}

func Test_mergeDeviceFault(t *testing.T) {
	t.Run("Test_mergeDeviceFault", func(t *testing.T) {
		split := []constant.DeviceFault{
			{
				FaultType:            "xx",
				NPUName:              "Ascend910-0",
				LargeModelFaultLevel: "xx",
				FaultLevel:           "xx",
				FaultHandling:        "xx",
				FaultCode:            "0x1",
				FaultTime:            1,
			},
			{
				FaultType:            "xx",
				NPUName:              "Ascend910-0",
				LargeModelFaultLevel: "xx",
				FaultLevel:           "xx",
				FaultHandling:        "xx",
				FaultCode:            "0x2",
				FaultTime:            2,
			},
		}
		want := constant.DeviceFault{
			FaultType:            "xx",
			NPUName:              "Ascend910-0",
			LargeModelFaultLevel: "xx",
			FaultLevel:           "xx",
			FaultHandling:        "xx",
			FaultCode:            "0x1,0x2",
			FaultTimeMap: map[string]int64{
				"0x1": 1,
				"0x2": 2,
			},
			FaultTime: 0,
		}
		got, err := mergeDeviceFault(split)
		if err != nil {
			t.Errorf("mergeDeviceFault() error = %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mergeDeviceFault() got = %v, want %v", got, want)
		}
	})
}
