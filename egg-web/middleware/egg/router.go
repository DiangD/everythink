package egg

import (
	"net/http"
	"strings"
)

type router struct {
	root     map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		root:     make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	ps := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range ps {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRouter(method, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	parts := parsePattern(pattern)
	if _, ok := r.root[method]; !ok {
		r.root[method] = &node{}
	}
	r.root[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

//使用中间件后的流程 c.handlers = {logger...} logger part1 -> middleware2 -> ... ->handler -> ... part2 -> logger
func (r *router) handle(c *Context) {
	node, params := r.getRouter(c.Method, c.Path)
	if node != nil {
		c.Params = params
		key := c.Method + "-" + node.pattern
		//将path对应的handler添加到handlers尾部，中间件按添加顺序执行
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	//执行第一个中间件
	c.Next()
}

//important!!! /hello/:name <=> /hello/world
func (r *router) getRouter(method, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.root[method]
	if !ok {
		return nil, nil
	}
	node := root.search(searchParts, 0)

	if node != nil {
		parts := parsePattern(node.pattern)
		//获取path里面param的值设置到context的params里
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len([]rune(part)) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return node, params
	}
	return nil, nil
}
