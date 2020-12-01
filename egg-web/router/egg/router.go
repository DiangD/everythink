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

func (r *router) handle(c *Context) {
	node, params := r.getRouter(c.Method, c.Path)
	if node != nil {
		c.Params = params
		key := c.Method + "-" + node.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}

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
