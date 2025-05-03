package LnWeb

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

// HandlerFunc :
// type HandlerFunc func(http.ResponseWriter, *http.Request)
type HandlerFunc func(*Context)

// Engine : 是所有请求的统一处理引擎
type Engine struct {
	// 路由映射表：{ url : func }
	//router map[string]HandlerFunc
	router *router

	*RouterGroup
	groups []*RouterGroup // store all groups

	htmlTemplates *template.Template // for html render [ in memory ]
	funcMap       template.FuncMap   // for html render
}

// New : 构造一个 gee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	// key := method + "-" + pattern -> eg: GET-/hello 、 POST-/hello ...
	// engine.router[key] = handler
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
	// middleware slice
	var middlewares []HandlerFunc

	// 遍历所有路由组，筛选匹配请求路径的中间件
	// 注意：这里会匹配所有包含请求路径前缀的路由组，实现中间件的叠加
	for _, group := range engine.groups {
		// 如果请求路径以路由组前缀开头，则合并该组的中间件
		// 示例：请求路径为 /api/v1/users，路由组前缀为 /api 和 /api/v1 都会匹配
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			// 使用展开运算符添加中间件切片，保持中间件注册顺序
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

func (engine *Engine) SetFunMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

/*
	Engine ( core )
	├── router: ( manage all routing groups )
	└── RouterGroup ( default routing group )
		├── prefix: ""
		└── engine: ( point to Engine itself )
*/

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	engine      *Engine       // all groups share an Engine instance
}

// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, pattern string, handler HandlerFunc) {
	pattern = group.prefix + pattern
	log.Printf("Route %4s - %s\n", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Use is defined to add middleware to the group
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// Static : serve static files [for user]
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}
