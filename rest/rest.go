package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

type ControllerAction func(HttpBundle) Response

type RestHandler struct {
	controllers map[string]map[string]reflect.Value
	router      router
}

func NewRestHandler(router router) RestHandler {
	controllers := make(map[string]map[string]reflect.Value)
	for _, route := range router.Routes {
		controllerVal := reflect.ValueOf(route.Controller)
		controllers[route.ControllerName] = getControllerActions(controllerVal)
	}
	return RestHandler{controllers, router}
}

func getControllerActions(controllerVal reflect.Value) map[string]reflect.Value {
	actions := make(map[string]reflect.Value)
	for i := 0; i < controllerVal.NumField(); i++ {
		fieldVal := controllerVal.Field(i)
		if _, ok := fieldVal.Interface().(ControllerAction); ok {
			actions[controllerVal.Type().Field(i).Name] = fieldVal
		}
	}
	return actions
}

func (rh RestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := rh.router.Match(r.URL.Path)
	if route == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	action, actionExists := route.Action[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	actionValue := rh.controllers[route.ControllerName][action]
	if !actionValue.IsValid() {
		log.Println(errors.New(fmt.Sprintf("Action '%v' does not exist", action)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	hb := HttpBundle{
		request:     r,
		response:    w,
		routeParams: route.Params,
	}

	result := actionValue.Call([]reflect.Value{reflect.ValueOf(hb)})

	resp := result[0].Interface().(Response)
	resp.Send(w)
}

type HttpBundle struct {
	request     *http.Request
	response    http.ResponseWriter
	routeParams map[string]interface{}
}

func (h HttpBundle) Request() *http.Request {
	return h.request
}

func (h HttpBundle) Response() http.ResponseWriter {
	return h.response
}

func (h HttpBundle) RouteParams() map[string]interface{} {
	return h.routeParams
}

func (h HttpBundle) FormData() url.Values {
	h.request.ParseForm()
	switch h.request.Method {
	case "POST":
		fallthrough
	case "PUT":
		return h.request.PostForm
	default:
		return h.request.Form
	}
}

func (h *HttpBundle) BindEntity(i interface{}) error {
	body, err := ioutil.ReadAll(h.request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, i)
}

// TODO: add more safetly nets and verifications.  A lot of assumptions going on
// TODO: add request logging
