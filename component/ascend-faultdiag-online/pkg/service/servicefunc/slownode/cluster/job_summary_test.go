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

// Pakcage cluster is a DT collection for func in job_summary
package cluster

import (
	"encoding/json"
	"testing"
	"reflect"
	"fmt"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model"
	"ascend-faultdiag-online/pkg/model/slownode"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
	}
	err := hwlog.InitRunLogger(&config, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func TestServersGenerator(t *testing.T) {
	convey.Convey("test serversGenerator", t, func() {
		var hcclJson = model.HcclJson{}
		servers := serversGenerator(hcclJson)
		convey.So(servers, convey.ShouldBeEmpty)
		var data = `{
		"server_list": [
			{
				"server_id": "127.0.0.1",
				"server_sn": "321123",
				"device": [
					{
						"rank_id": "1"
					},
										{
						"rank_id": "2"
					}
				]
			}
		]}`
		err := json.Unmarshal([]byte(data), &hcclJson)
		convey.So(err, convey.ShouldBeNil)
		var expect = []slownode.Server{
			{
				Sn:      "321123",
				Ip:      "127.0.0.1",
				RankIds: []string{"1", "2"},
			},
		}
		servers = serversGenerator(hcclJson)
		convey.So(reflect.DeepEqual(servers, expect), convey.ShouldBeTrue)
	})
}
