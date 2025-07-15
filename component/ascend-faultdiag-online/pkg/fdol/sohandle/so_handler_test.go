/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, convey.Software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package sohandle is test collection for func in convey.So_handler
package sohandle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	fileMode0755 = 0755
	fileMode0644 = 0644
)

func TestFilterSoFiles(t *testing.T) {
	convey.Convey("Given a directory with .so files", t, func() {
		// Create a temporary directory for testing
		filePath := "so_test"
		err := os.Mkdir(filePath, fileMode0755)
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(filePath)

		// Create test files
		testFiles := []struct {
			name    string
			isDir   bool
			content string
		}{
			{"lib1.so", false, "test content"},
			{"lib2.so", false, "test content"},
			{"notlib.txt", false, "test content"},
			{"subdir", true, ""},
			{"subdir/lib3.so", false, "test content"},
			{"subdir/notlib.go", false, "test content"},
		}

		for _, tf := range testFiles {
			path := filepath.Join(filePath, tf.name)
			if tf.isDir {
				err := os.Mkdir(path, fileMode0755)
				convey.So(err, convey.ShouldBeNil)
			} else {
				err := os.WriteFile(path, []byte(tf.content), fileMode0644)
				convey.So(err, convey.ShouldBeNil)
			}
		}

		convey.Convey("When filtering so files", func() {
			soFiles, err := filterSoFiles(filePath)
			convey.Convey("It should return only so files", func() {
				convey.So(err, convey.ShouldBeNil)
				convey.So(soFiles, convey.ShouldContain, filepath.Join(filePath, "lib1.so"))
				convey.So(soFiles, convey.ShouldContain, filepath.Join(filePath, "lib2.so"))
				convey.So(soFiles, convey.ShouldContain, filepath.Join(filePath, "subdir", "lib3.so"))
			})
		})
		testFilterSoFilesWithError(filePath)
	})
}

func testFilterSoFilesWithError(filePath string) {

	convey.Convey("When the directory doesn't exist", func() {
		_, err := filterSoFiles(filepath.Join(filePath, "nonexistent"))
		convey.Convey("It should return an error", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})

	convey.Convey("When reaching max file count", func() {
		// Create maxfile + 1 files
		for i := 0; i < maxFileCount; i++ {
			path := filepath.Join(filePath, fmt.Sprintf("lib%d.so", i))
			err := os.WriteFile(path, []byte("test"), fileMode0644)
			convey.So(err, convey.ShouldBeNil)
		}

		_, err := filterSoFiles(filePath)

		convey.Convey("It should return an error", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "reach the max file count")
		})
	})

	convey.Convey("When directory is empty", func() {
		emptyDir, err := os.MkdirTemp("", "empty_test")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(emptyDir)

		soFiles, err := filterSoFiles(emptyDir)

		convey.Convey("It should return empty slice", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(soFiles, convey.ShouldBeEmpty)
		})
	})
}
