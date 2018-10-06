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

package inject

import (
	"github.com/hidevopsio/hiboot/pkg/utils/mapstruct"
	"reflect"
)

type injectTag struct {
	BaseTag
}

func init() {
	AddTag(new(injectTag))
}

//func (t *injectTag) IsSingleton() bool {
//	return true
//}

func (t *injectTag) Decode(object reflect.Value, field reflect.StructField, tag string) (retVal interface{}) {
	properties := t.ParseProperties(tag)

	// first, find if object is already instantiated
	if field.Type.Kind() == reflect.Ptr {
		// if object is not exist, then instantiate new object
		// parse tag and instantiate filed
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		// TODO: get instance from factory first
		o := reflect.New(ft)
		retVal = o.Interface()
		// inject field value
		if properties.Count() != 0 {
			mapstruct.Decode(retVal, properties.Items())
		}
	} else {
		return
	}
	return retVal
}
