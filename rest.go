package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

// ControllerAction is a type for all controller actions
type ControllerAction func(HttpBundle) ResponseSender

// Handler implements http.Handler and contains the router and controllers for the REST api
type Handler struct {
	router       router
	interceptors []Interceptor
}

// NewHandler returns a new Handler with router initialized
func NewHandler(router router) *Handler {
	return &Handler{router, make([]Interceptor, 0)}
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

	if ok := rh.invokeInterceptors(context); !ok {
		return
	}

	action, actionExists := route.action[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	hb := HttpBundle{
		request:     r,
		response:    w,
		routeParams: route.Params,
	}

	result := action(hb)

	result.Send(w)
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

func (h *Handler) invokeInterceptors(c Context) bool {
	result := true
	for i := 0; i < len(h.interceptors) && result; i++ {
		result = h.interceptors[i](c)
	}

	return result
}

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Route          *Route
}
