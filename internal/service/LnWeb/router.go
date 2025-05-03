package LnWeb

import (
	"net/http"
	"strings"
)

// router 路由管理器
type router struct {
	// 存储每种HTTP方法对应的Trie树根节点
	// 示例：roots['GET'] 表示GET方法的路由树根节点
	roots map[string]*node

	// 存储路由处理函数的映射表
	// 键格式：方法-路径，示例：handlers['GET-/api/v1/:id']
	handlers map[string]HandlerFunc
}

// 创建新路由实例
func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// 解析路由模式为分段路径，处理动态参数和通配符
// 示例：
//
//	"/api/v1/:id" -> ["api", "v1", ":id"]
//	"/static/*filepath" -> ["static", "*filepath"]
func parsePattern(pattern string) []string {
	// 分割路径为字符串数组
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		// 跳过空字符串（处理开头/结尾的斜杠）
		if item != "" {
			parts = append(parts, item)
			// 遇到通配符立即停止解析后续路径
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

// 注册路由规则和处理器
// method: HTTP方法（GET/POST等）
// pattern: 路由路径模式（可含动态参数）
// handler: 对应的处理函数
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	// 解析路径模式为分段路径
	parts := parsePattern(pattern)

	// 构造处理函数的映射键（方法+原始路径）
	key := method + "-" + pattern

	// 初始化对应HTTP方法的路由树
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &node{}
	}

	// 将路由插入Trie树（从根节点开始插入）
	r.roots[method].insert(pattern, parts, 0)

	// 存储路由处理函数
	r.handlers[key] = handler
}

// 查找匹配的路由节点并解析参数
// method: HTTP方法
// path: 实际请求路径
// 返回: 匹配的节点 和 解析出的参数映射
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	// 解析实际请求路径为分段路径
	searchParts := parsePattern(path)
	params := make(map[string]string)

	// 获取对应HTTP方法的路由树根节点
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	// 在Trie树中搜索匹配的节点
	n := root.search(searchParts, 0)

	if n != nil {
		// 解析路由注册时的模式路径（注意使用n.pattern而不是请求路径）
		parts := parsePattern(n.pattern)

		// 构建参数映射表
		for idx, part := range parts {
			switch {
			case part[0] == ':':
				// 动态参数匹配（:id等）
				params[part[1:]] = searchParts[idx]
			case part[0] == '*' && len(part) > 1:
				// 通配符匹配（*filepath等）
				params[part[1:]] = strings.Join(searchParts[idx:], "/")
				break // 通配符后续路径不再处理
			}
		}
		return n, params
	}
	return nil, nil
}

// 处理请求入口
// handle 处理路由请求的核心逻辑
func (r *router) handle(c *Context) {
	// 1. 路由匹配阶段
	// 通过请求方法和路径查找匹配的路由节点
	if n, params := r.getRoute(c.Method, c.Path); n != nil {
		// 成功匹配路由时的处理流程

		// 1.1 构造路由唯一标识键
		// 格式示例：GET-/user/:id
		key := c.Method + "-" + n.pattern

		// 1.2 注入路由参数到上下文
		// 示例：路径/user/123 会解析出 {"id": "123"}
		c.Params = params

		// 1.3 追加路由处理函数到执行链
		// 注意：此时中间件已通过路由组收集到c.handlers
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		// 2. 404处理流程

		// 2.1 追加404处理函数到执行链
		c.handlers = append(c.handlers, func(ctx *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}

	// 3. 启动中间件+处理函数执行链
	// 执行顺序：中间件1 -> 中间件2 -> 路由处理函数/404处理器
	c.Next()
}

/*

典型工作流程示例：

注册路由：router.addRoute("GET", "/user/:name", userHandler)
	请求处理：
		请求路径：/user/ln
		parsePattern解析得到：["user", "ln"]
		Trie树匹配到模式/user/:name
		解析参数：name => "ln"
		构造key：GET-/user/:name
		调用对应的userHandler

	注意事项：
		通配符路由必须作为最后一个路径段
		不同HTTP方法有独立的路由树
		参数名称在模式中必须唯一
		需要修改handle方法以支持动态路由匹配

*/
