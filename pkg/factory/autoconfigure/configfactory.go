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

// Package autoconfigure implement ConfigurableFactory
package autoconfigure

import (
	"errors"
	"github.com/hidevopsio/hiboot/pkg/factory/autoconfigure/depends"
	"github.com/hidevopsio/hiboot/pkg/factory/instantiate"
	"github.com/hidevopsio/hiboot/pkg/inject"
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/hidevopsio/hiboot/pkg/system"
	"github.com/hidevopsio/hiboot/pkg/utils/cmap"
	"github.com/hidevopsio/hiboot/pkg/utils/gotest"
	"github.com/hidevopsio/hiboot/pkg/utils/io"
	"github.com/hidevopsio/hiboot/pkg/utils/reflector"
	"github.com/hidevopsio/hiboot/pkg/utils/replacer"
	"github.com/hidevopsio/hiboot/pkg/utils/str"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const (
	System            = "system"
	application       = "application"
	config            = "config"
	yaml              = "yaml"
	appProfilesActive = "APP_PROFILES_ACTIVE"
)

var (
	// ErrInvalidMethod method is invalid
	ErrInvalidMethod = errors.New("[factory] method is invalid")

	// ErrFactoryCannotBeNil means that the InstantiateFactory can not be nil
	ErrFactoryCannotBeNil = errors.New("[factory] InstantiateFactory can not be nil")

	// ErrFactoryIsNotInitialized means that the InstantiateFactory is not initialized
	ErrFactoryIsNotInitialized = errors.New("[factory] InstantiateFactory is not initialized")

	// ErrInvalidObjectType means that the Configuration type is invalid, it should embeds app.PreConfiguration
	ErrInvalidObjectType = errors.New("[factory] invalid Configuration type, one of app.Configuration, app.PreConfiguration, or app.PostConfiguration need to be embedded")

	// ErrConfigurationNameIsTaken means that the configuration name is already taken
	ErrConfigurationNameIsTaken = errors.New("[factory] configuration name is already taken")

	// ErrComponentNameIsTaken means that the component name is already taken
	ErrComponentNameIsTaken = errors.New("[factory] component name is already taken")
)

type ConfigurableFactory struct {
	*instantiate.InstantiateFactory
	configurations cmap.ConcurrentMap
	systemConfig   *system.Configuration
	builder        *system.Builder

	preConfigContainer  cmap.ConcurrentMap
	configContainer     cmap.ConcurrentMap
	postConfigContainer cmap.ConcurrentMap

	preConfigureContainer  []interface{}
	configureContainer     []interface{}
	postConfigureContainer []interface{}
}

// Initialize initialize ConfigurableFactory
func (f *ConfigurableFactory) Initialize(configurations cmap.ConcurrentMap) (err error) {
	if f.InstantiateFactory == nil {
		return ErrFactoryCannotBeNil
	}
	if !f.Initialized() {
		return ErrFactoryIsNotInitialized
	}
	f.configurations = configurations
	f.SetInstance("configurations", configurations)

	f.preConfigContainer = cmap.New()
	f.configContainer = cmap.New()
	f.postConfigContainer = cmap.New()
	return
}

// SystemConfiguration getter
func (f *ConfigurableFactory) SystemConfiguration() *system.Configuration {
	return f.systemConfig
}

// Configuration getter
func (f *ConfigurableFactory) Configuration(name string) interface{} {
	cfg, ok := f.configurations.Get(name)
	if ok {
		return cfg
	}
	return nil
}

// BuildSystemConfig build system configuration
func (f *ConfigurableFactory) BuildSystemConfig() (systemConfig *system.Configuration, err error) {
	workDir := io.GetWorkDir()
	systemConfig = new(system.Configuration)
	profile := os.Getenv(appProfilesActive)
	f.builder = &system.Builder{
		Path:       filepath.Join(workDir, config),
		Name:       application,
		FileType:   yaml,
		Profile:    profile,
		ConfigType: systemConfig,
	}

	f.SetInstance("systemConfiguration", systemConfig)
	inject.DefaultValue(systemConfig)
	_, err = f.builder.Build()
	if err != nil {
		return
	}
	// TODO: should separate instance to system and app
	inject.IntoObject(systemConfig)
	replacer.Replace(systemConfig, systemConfig)

	f.configurations.Set(System, systemConfig)

	f.systemConfig = systemConfig
	return systemConfig, err
}

// Build build all auto configurations
func (f *ConfigurableFactory) Build(configs [][]interface{}) {
	// categorize configurations first, then inject object if necessary
	for _, item := range configs {
		name, inst := f.ParseInstance("Configuration", item...)
		if name == "" || name == "configuration" {
			continue
		}

		ifcField := reflector.GetEmbeddedInterfaceField(inst)

		if ifcField.Anonymous {
			switch ifcField.Name {
			case "Configuration":
				f.configureContainer = append(f.configureContainer, inst)
			case "PreConfiguration":
				f.preConfigureContainer = append(f.preConfigureContainer, inst)
			case "PostConfiguration":
				f.postConfigureContainer = append(f.postConfigureContainer, inst)
			default:
				continue
			}
		} else {
			err := ErrInvalidObjectType
			log.Error(err)
			continue
		}
	}

	f.build(f.preConfigureContainer)
	f.build(f.configureContainer)
	f.build(f.postConfigureContainer)

}

// InstantiateByName instantiate by method name
func (f *ConfigurableFactory) InstantiateByName(configuration interface{}, name string) (inst interface{}, err error) {
	objVal := reflect.ValueOf(configuration)
	method, ok := objVal.Type().MethodByName(name)
	if ok {
		return f.InstantiateMethod(configuration, method, name)
	}
	return nil, ErrInvalidMethod
}

// InstantiateMethod instantiate by iterated methods
func (f *ConfigurableFactory) InstantiateMethod(configuration interface{}, method reflect.Method, methodName string) (inst interface{}, err error) {
	//log.Debugf("method: %v", methodName)
	instanceName := str.LowerFirst(methodName)
	if inst = f.GetInstance(instanceName); inst != nil {
		//log.Debugf("instance %v exists", instanceName)
		return
	}
	numIn := method.Type.NumIn()
	// only 1 arg is supported so far
	argv := make([]reflect.Value, numIn)
	argv[0] = reflect.ValueOf(configuration)
	for a := 1; a < numIn; a++ {
		// TODO: eliminate duplications
		mth := method.Type.In(a)
		iTyp := reflector.IndirectType(mth)
		mthName := str.ToLowerCamel(iTyp.Name())
		depInst := f.GetInstance(mthName)
		if depInst == nil {
			pkgName := io.DirName(iTyp.PkgPath())
			alternativeName := str.ToLowerCamel(pkgName) + iTyp.Name()
			depInst = f.GetInstance(alternativeName)
			if depInst == nil {
				// TODO: check it it's dependency circle
				// TODO: check if it depends on the instance of another configuration
				depInst, err = f.InstantiateByName(configuration, strings.Title(mthName))
				if err != nil {
					depInst, err = f.InstantiateByName(configuration, strings.Title(alternativeName))
				}
			}
		}
		if depInst == nil {
			log.Errorf("[factory] failed to inject dependency as it can not be found")
		}
		argv[a] = reflect.ValueOf(depInst)
	}
	// inject instance into method
	retVal := method.Func.Call(argv)
	// save instance
	if retVal != nil && retVal[0].CanInterface() {
		inst = retVal[0].Interface()
		//log.Debugf("instantiated: %v", instance)
		f.SetInstance(instanceName, inst)
	}
	return
}

// Instantiate run instantiation by method
func (f *ConfigurableFactory) Instantiate(configuration interface{}) (err error) {
	cv := reflect.ValueOf(configuration)

	// inject configuration before instantiation

	configType := cv.Type()
	//log.Debug("type: ", configType)
	//name := configType.Elem().Name()
	//log.Debug("fieldName: ", name)

	// call Init
	numOfMethod := cv.NumMethod()
	//log.Debug("methods: ", numOfMethod)
	for mi := 0; mi < numOfMethod; mi++ {
		method := configType.Method(mi)
		// skip Init method
		if method.Name != "Init" {
			_, err = f.InstantiateMethod(configuration, method, method.Name)
			if err != nil {
				return
			}
		}
	}
	return
}

// appProfilesActive getter
func (f *ConfigurableFactory) appProfilesActive() string {
	if f.systemConfig == nil {
		return os.Getenv(appProfilesActive)
	}
	return f.systemConfig.App.Profiles.Active
}

// build
func (f *ConfigurableFactory) build(cfgContainer []interface{}) {
	// sort dependencies
	sort.Sort(depends.ByDependency(cfgContainer))

	isTestRunning := gotest.IsRunning()
	for _, item := range cfgContainer {
		name, configType := f.ParseInstance("Configuration", item)

		// TODO: should check if profiles is enabled str.InSlice(name, sysconf.App.Profiles.Include)
		if !isTestRunning && f.systemConfig != nil && !str.InSlice(name, f.systemConfig.App.Profiles.Include) {
			continue
		}
		log.Infof("Auto configure %v starter", name)

		// inject properties
		f.builder.ConfigType = configType

		// inject default value
		inject.DefaultValue(configType)

		cf, err := f.builder.Build(name, f.appProfilesActive())

		// TODO: check if cf.DependsOn
		if cf == nil {
			log.Warnf("failed to build %v configuration with error %v", name, err)
		} else {
			// replace references and environment variables
			if f.systemConfig != nil {
				replacer.Replace(cf, f.systemConfig)
			}
			inject.IntoObject(cf)
			replacer.Replace(cf, cf)

			// instantiation
			if err == nil {
				// create instances
				f.Instantiate(cf)
				// save configuration
				if _, ok := f.configurations.Get(name); ok {
					log.Fatalf("[factory] configuration name %v is already taken", name)
				}
				f.configurations.Set(name, cf)
			}
		}
	}
}
