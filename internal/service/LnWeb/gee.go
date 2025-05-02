package LnWeb

import (
	"net/http"
)

// HandlerFunc :
// type HandlerFunc func(http.ResponseWriter, *http.Request)
type HandlerFunc func(*Context)

// Engine : 是所有请求的统一处理引擎
type Engine struct {
	// 路由映射表：{ url : func }
	//router map[string]HandlerFunc
	router *router
}

// New : 构造一个 gee.Engine
func New() *Engine {
	return &Engine{
		router: newRouter(),
	}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	// 用以支持：GET-/hello 、 POST-/hello ...
	//key := method + "-" + pattern
	//engine.router[key] = handler
	engine.router.addRoute(method, pattern, handler)
}

func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// ServeHTTP 即对 Handler 接口的实现
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//key := req.Method + "-" + req.URL.Path
	//if handler, ok := engine.router[key]; ok {
	//	handler(w, req)
	//} else {
	//	http.NotFound(w, req)
	//}

	c := newContext(w, req)
	engine.router.handle(c)
}
