package LnWeb

import "strings"

// node 结构体表示路由前缀树（Trie树）的节点，用于实现路由匹配
type node struct {
	// pattern 存储当前节点代表的完整路由路径，只有叶子节点才会设置该值
	// 例如：/p/:lang
	pattern string

	// part 表示当前节点对应的路由部分
	// 例如：在路径 /p/:lang 中，可能为 ":lang"
	part string

	// children 存储子节点，用于构建树结构
	children []*node

	// isWild 标识是否是通配节点
	// 当 part 以 ':'（参数）或 '*'（通配符）开头时为 true，表示需要特殊匹配
	isWild bool
}

// matchChild 查找第一个匹配的子节点，用于路由插入
// part: 当前要匹配的路由部分
// 返回值: 第一个匹配的节点（优先精确匹配，其次通配节点）
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		// 当子节点的 part 完全匹配，或子节点是通配节点时匹配成功
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// matchChildren 查找所有匹配的子节点，用于路由搜索
// part: 当前要匹配的路由部分
// 返回值: 所有可能匹配的节点集合（包含精确匹配和通配节点）
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// insert 插入路由到前缀树中
// pattern: 完整路由路径（如 /p/:lang）
// parts: 按 '/' 分割后的路由部分数组（如 ["p", ":lang"]）
// height: 当前处理的路由部分层级（从 0 开始）
func (n *node) insert(pattern string, parts []string, height int) {
	// 递归终止条件：当处理到最后一个路由部分时，设置当前节点的 pattern
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	// 获取当前层级的路由部分
	part := parts[height]

	// 查找是否存在匹配的子节点
	child := n.matchChild(part)
	if child == nil {
		// 创建新节点，并判断是否为通配节点
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*', // 以':'或'*'开头则为通配
		}
		// 将新节点加入子节点列表
		n.children = append(n.children, child)
	}

	// 递归插入下一个路由部分
	child.insert(pattern, parts, height+1)
}

// search 在前缀树中搜索匹配的路由
// parts: 按 '/' 分割后的请求路径部分
// height: 当前处理的路由部分层级
// 返回值: 匹配的叶子节点（nil 表示未找到）
func (n *node) search(parts []string, height int) *node {
	// 递归终止条件：
	// 1. 已处理完所有路由部分
	// 2. 当前节点是通配符节点（* 开头，匹配后续所有路径）
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		// 只有叶子节点才有非空 pattern
		if n.pattern == "" {
			return nil
		}
		return n
	}

	// 获取当前层级的路由部分
	part := parts[height]

	// 获取所有可能匹配的子节点（包含通配节点）
	children := n.matchChildren(part)

	// 递归搜索所有可能的子路径
	for _, child := range children {
		// 深度优先搜索
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

/*

示例工作流程：
	插入路由 /p/:lang/doc：
		拆分为 parts ["p", ":lang", "doc"]
		逐层创建节点：根 -> p -> :lang -> doc（在doc节点设置pattern）

	搜索路径 /p/go/doc：
		拆分为 parts ["p", "go", "doc"]
		匹配路径：根 -> p（精确匹配） -> :lang（参数匹配） -> doc（精确匹配）

*/
