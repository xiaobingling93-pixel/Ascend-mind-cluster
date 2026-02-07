/* Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package metrics for general collector
package metrics

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
)

func TestIsSupportNetworkHealthDevices(t *testing.T) {
	convey.Convey("TestIsSupportNetworkHealthDevices", t, func() {
		result := isSupportNetworkHealthDevices(api.Ascend910A3, 0)
		convey.So(result, convey.ShouldEqual, true)
		result = isSupportNetworkHealthDevices(api.Ascend910A5, api.Atlas9501DMainBoardID)
		convey.So(result, convey.ShouldEqual, true)
		result = isSupportNetworkHealthDevices(api.Ascend910A5, api.Atlas3504PMainBoardID)
		convey.So(result, convey.ShouldEqual, false)
	})
}
