/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package util is using for the total variable.
*/

package util

import (
	"errors"
	"testing"
)

func TestErrorCollector(t *testing.T) {
	ec := NewErrorCollector("test", DefaultPrintLimit)
	ec.Add("node1", nil)
	ec.Add("node2", errors.New("test error"))
	ec.Print()
	t.Run("01-ErrorCollector", func(t *testing.T) {
		if ec.count != 1 {
			t.Errorf("ErrorCollector.Count() = %v, want %v", ec.count, 1)
		}
		if len(ec.errors) != 1 {
			t.Errorf("ErrorCollector.Errors() = %v, want %v", len(ec.errors), 1)
		}
		if ec.errors["test error"][0] != "node2" {
			t.Errorf("ErrorCollector.Errors() = %v, want %v", ec.errors["test error"], []string{"node2"})
		}
	})

}
