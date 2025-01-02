package card310x4

import (
	"errors"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	mockIllegalTaskNPUNumber = -1
	mockTaskNPUNumberZero    = 0
	mockTaskNPUNumberOne     = 1
	mockTaskNPUNumberTwo     = 2
	mockTaskNPUNumberThree   = 3
	mockTaskNPUNumberFour    = 4
)

type getNPUAllocPriorityArrayTestCase struct {
	Name          string
	TaskNPUNumber int
	WantArr       []int
	WantErr       error
}

func buildGetNPUAllocPriorityArrayTestCases() []getNPUAllocPriorityArrayTestCase {
	return []getNPUAllocPriorityArrayTestCase{
		{
			Name:          "01-getNPUAllocPriorityArray return nil when taskNPUNumber is 0",
			TaskNPUNumber: mockTaskNPUNumberZero,
		},
		{
			Name:          "02-getNPUAllocPriorityArray return array when taskNPUNumber is 1",
			TaskNPUNumber: mockTaskNPUNumberOne,
			WantArr:       []int{1, util.NPUIndex3, util.NPUIndex2, maxCardNPUNum},
		},
		{
			Name:          "03-getNPUAllocPriorityArray return array when taskNPUNumber is 2",
			TaskNPUNumber: mockTaskNPUNumberTwo,
			WantArr:       []int{util.NPUIndex2, util.NPUIndex3, maxCardNPUNum},
		},
		{
			Name:          "04-getNPUAllocPriorityArray return array when taskNPUNumber is 3",
			TaskNPUNumber: mockTaskNPUNumberThree,
			WantArr:       []int{util.NPUIndex3, maxCardNPUNum},
		},
		{
			Name:          "05-getNPUAllocPriorityArray return array when taskNPUNumber is 4",
			TaskNPUNumber: mockTaskNPUNumberFour,
			WantArr:       []int{maxCardNPUNum},
		},
		{
			Name:          "06-getNPUAllocPriorityArray return error when taskNPUNumber is -1",
			TaskNPUNumber: mockIllegalTaskNPUNumber,
			WantErr:       errors.New("illegal request npu number: -1"),
		},
	}
}

func TestGetNPUAllocPriorityArray(t *testing.T) {
	testCases := buildGetNPUAllocPriorityArrayTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if res, err := getNPUAllocPriorityArray(tt.TaskNPUNumber); !reflect.DeepEqual(res,
				tt.WantArr) && !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("getNPUAllocPriorityArray() res: %v, err: %v , wantArr: %v, wantErr: %v",
					res, err, tt.WantArr, tt.WantErr)
			}
		})
	}
}
