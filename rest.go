package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

// ControllerAction is a type for all controller actions
type ControllerAction func(HttpBundle) ResponseSender

// Handler implements http.Handler and contains the router and controllers for the REST api
type Handler struct {
	controllers  map[string]map[string]reflect.Value
	router       router
	interceptors []Interceptor
}

// NewHandler returns a new Handler with controllers and router initialized
func NewHandler(router router) *Handler {
	controllers := make(map[string]map[string]reflect.Value)
	for _, route := range router.Routes {
		controllerVal := reflect.ValueOf(route.controller)
		controllers[route.controllerName] = getControllerActions(controllerVal)
	}
	return &Handler{controllers, router, make([]Interceptor, 0)}
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

func (rh *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := rh.router.Match(r.URL.Path)
	if route == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	context := Context{
		Request:        r,
		ResponseWriter: w,
		Route:          route,
	}

	rh.invokeInterceptors(context)

	action, actionExists := route.action[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	actionValue := rh.controllers[route.controllerName][action]
	if !actionValue.IsValid() {
		log.Println(fmt.Errorf("Action '%v' does not exist", action))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	hb := HttpBundle{
		request:     r,
		response:    w,
		routeParams: route.Params,
	}

	result := actionValue.Call([]reflect.Value{reflect.ValueOf(hb)})

	resp := result[0].Interface().(ResponseSender)
	resp.Send(w)
}

// HttpBundle holds the *http.Request, http.ResponseWriter, and routeParams of the request
type HttpBundle struct {
	request     *http.Request
	response    http.ResponseWriter
	routeParams map[string]interface{}
}

// Request returns *http.Request
func (h HttpBundle) Request() *http.Request {
	return h.request
}

// Response returns http.ResponseWriter
func (h HttpBundle) Response() http.ResponseWriter {
	return h.response
}

// RouteParams returns route params as map[string]interface{}
func (h HttpBundle) RouteParams() map[string]interface{} {
	return h.routeParams
}

// FormData returns data related to the request from GET, POST, or PUT
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

// BindJsonEntity binds the JSON body from the request to an interface{}
func (h *HttpBundle) BindJsonEntity(i interface{}) error {
	body, err := ioutil.ReadAll(h.request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, i)
}

type Interceptor func(Context) bool

func (h *Handler) AddInterceptor(i Interceptor) {
	h.interceptors = append(h.interceptors, i)
}

func (h *Handler) invokeInterceptors(c Context) {
	i := 0
	for i < len(h.interceptors)-1 && h.interceptors[i](c) {
		i++
	}
}

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Route          *Route
}
