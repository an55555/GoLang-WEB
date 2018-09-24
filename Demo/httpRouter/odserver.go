package odserver

import (
	"fmt"
	"net/http"
	"reflect"
)

const (
	HF1 = "func(http.ResponseWriter, *http.Request)"
	HF2 = "func(*odserver.Context)"
)

type IOdServer interface {
	ExecuteFunc(fo FuncObject, c *Context)
}

type OdServer struct {
	HConfig *httpConfig
	*Router
	*StaticPathConfig
}

//接口函数单位，即我们编写代码逻辑的函数，用户自定义实现
type HandlerFunc1 func(w http.ResponseWriter, req *http.Request)

////接口处理函数2，用户自定义实现
//type HandlerFunc2 func(c *Context)

func Default() *OdServer {
	return &OdServer{
		HConfig:          DefaultConfig(),
		Router:           NewRouter(),
		StaticPathConfig: NewStaticPathConfig(),
	}
}

//实现Handler接口，匹配方法以及路径
func (o *OdServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if bol, url, path := o.MapStaticPath(req.URL.Path); bol {
		o.doStaticPath(w, req, url, path)
	} else {
		//设置http头部信息
		for key, value := range o.HConfig.Header() {
			w.Header().Add(key, value)
		}

		//转发给doHandler进行执行
		o.doHandler(w, req)
	}
}

//判断需要执行的Http Method，从而查找对应的接口并且执行
func (o *OdServer) doHandler(w http.ResponseWriter, req *http.Request) {
	c := NewContext(req, w)
	if fo, exist := o.doMapping(req); exist {
		o.ExecuteFunc(fo, c)
	} else {
		fmt.Fprintln(c.GoResW(), "404")
	}
}

func (o *OdServer) doMapping(req *http.Request) (FuncObject, bool) {
	isFind := false
	var ho FuncObject
	switch req.Method {
	case http.MethodGet:
		{
			ho, isFind = o.doUrlMapping(req.URL.RequestURI(), GET)
		}
	case http.MethodPost:
		{
			ho, isFind = o.doUrlMapping(req.URL.RequestURI(), POST)
		}
	case http.MethodDelete:
		{
			ho, isFind = o.doUrlMapping(req.URL.RequestURI(), DELETE)
		}
	case http.MethodPut:
		{
			ho, isFind = o.doUrlMapping(req.URL.RequestURI(), PUT)
		}
	default:
		{

		}
	}
	return ho, isFind
}

//执行编写的接口
func (o *OdServer) ExecuteFunc(fo FuncObject, c *Context) {
	c.Params = fo.params
	c.ParamsName = fo.paramsName
	hf := fo.f
	ft := reflect.TypeOf(hf)
	fv := reflect.ValueOf(hf)
	var params []reflect.Value
	//单独设置该请求的http头部信息
	for key, value := range fo.Header() {
		c.Rw.Header().Add(key, value)
	}

	if ft == nil {
		return
	}

	switch ft.String() {
	case HF1:
		{
			fmt.Println("ft1", ft.String())
			params = make([]reflect.Value, 2)
			params[0] = reflect.ValueOf(c.GoResW())
			params[1] = reflect.ValueOf(c.GoReq())
		}
	case HF2:
		{
			fmt.Println("ft2", ft.String())
			params = make([]reflect.Value, 1)
			params[0] = reflect.ValueOf(c)
		}
	default:
		{
			fmt.Println("ft3", ft.String())
			fmt.Fprintln(c.GoResW(), "404")
			return
		}
	}
	fmt.Println("fv", fv)
	fv.Call(params)
}
