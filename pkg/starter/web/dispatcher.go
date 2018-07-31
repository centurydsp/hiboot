package web

import (
	"fmt"
	"strings"
	"reflect"
	"net/http"
	"github.com/fatih/camelcase"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/hidevopsio/hiboot/pkg/utils"
	"github.com/hidevopsio/hiboot/pkg/system"
	"github.com/hidevopsio/hiboot/pkg/inject"
)


var httpMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

type dispatcher struct {
}

func (d *dispatcher) register(app *iris.Application, controllers []interface{}) error {
	for _, c := range controllers {
		field := reflect.ValueOf(c)

		fieldType := field.Type()
		//log.Debug("fieldType: ", fieldType)
		fieldName := fieldType.Elem().Name()
		//log.Debug("fieldName: ", fieldName)

		controller := field.Interface()
		//log.Debug("controller: ", controller)

		// inject component
		err := inject.IntoObject(field)
		if err != nil {
			return err
		}

		fieldValue := field.Elem()

		// get context mapping
		cp := fieldValue.FieldByName("ContextMapping")
		if !cp.IsValid() {
			return &system.InvalidControllerError{Name: fieldName}
		}
		contextMapping := fmt.Sprintf("%v", cp.Interface())

		// parse method
		fieldNames := camelcase.Split(fieldName)
		controllerName := ""
		if len(fieldNames) >= 2 {
			controllerName = strings.Replace(fieldName, fieldNames[len(fieldNames)-1], "", 1)
			controllerName = utils.LowerFirst(controllerName)
		}
		//log.Debug("controllerName: ", controllerName)
		// use controller's prefix as context mapping
		if contextMapping == "" {
			contextMapping = pathSep + controllerName
		}

		numOfMethod := field.NumMethod()
		//log.Debug("methods: ", numOfMethod)

		beforeMethod, ok := fieldType.MethodByName(BeforeMethod)
		var party iris.Party
		if ok {
			//log.Debug("contextPath: ", contextMapping)
			//log.Debug("beforeMethod.Name: ", beforeMethod.Name)
			hdl := new(handler)
			hdl.parse(beforeMethod, controller, "")
			party = app.Party(contextMapping, func(ctx context.Context) {
				hdl.call(ctx.(*Context))
			})
		} else {
			party = app.Party(contextMapping)
		}

		afterMethod, ok := fieldType.MethodByName(AfterMethod)
		if ok {
			hdl := new(handler)
			hdl.parse(afterMethod, controller, "")
			party.Done(func(ctx context.Context) {
				hdl.call(ctx.(*Context))
			})
		}

		for mi := 0; mi < numOfMethod; mi++ {
			method := fieldType.Method(mi)
			methodName := method.Name
			//log.Debug("method: ", methodName)

			ctxMap := camelcase.Split(methodName)
			httpMethod := strings.ToUpper(ctxMap[0])

			// apiContextMapping should add arguments
			//log.Debug("contextMapping: ", apiContextMapping)
			// check if it's valid http method
			if utils.StringInSlice(httpMethod, httpMethods) {
				var apiContextMapping string
				if len(ctxMap) > 2 && ctxMap[1] == "By" {
					for _, pathParam := range ctxMap[2:] {
						apiContextMapping = apiContextMapping + pathSep + "{" + strings.ToLower(pathParam) + "}"
					}
				} else {
					apiContextMapping = strings.Replace(methodName, ctxMap[0], "", 1)
					apiContextMapping = pathSep + utils.LowerFirst(apiContextMapping)
				}

				// parse all necessary requests and responses
				// create new method parser here
				hdl := new(handler)
				hdl.parse(method, controller, contextMapping + apiContextMapping)

				party.Handle(httpMethod, apiContextMapping, func(ctx context.Context) {
					hdl.call(ctx.(*Context))
					ctx.Next()
				})

			}
		}
	}
	return nil
}