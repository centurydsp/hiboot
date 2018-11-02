// Copyright 2018 John Deng (hi.devops.io@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Line 1: main package
package controller

import (
	"hidevops.io/hiboot/pkg/app/web"
	"net/http"
	"testing"
)

func GetTestApplication(t *testing.T) web.TestApplication {
	return web.RunTestApplication(t, newFooController)
}

func TestFooGet(t *testing.T) {
	GetTestApplication(t).
		Get("/foo").
		WithQueryObject(fooRequest{Name: "Peter", Age: 18}).
		Expect().Status(http.StatusOK)
}

func TestFooPost(t *testing.T) {
	GetTestApplication(t).
		Post("/foo").
		WithJSON(fooRequest{Name: "Mike", Age: 18}).
		Expect().Status(http.StatusOK)
}
