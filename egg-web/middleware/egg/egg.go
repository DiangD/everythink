package egg

import (
	"log"
	"net/http"
	"strings"
)

//HandlerFunc
type HandlerFunc func(c *Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
}

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

func (group *RouterGroup) Use(handlers ...HandlerFunc) {
	group.middlewares = append(group.middlewares, handlers...)
}

func (group *RouterGroup) addRouter(method, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRouter(method, pattern, handler)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	engine.router.handle(c)
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (engine *Engine) addRouter(method, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)
	engine.router.addRouter(method, pattern, handler)
}

func (group *RouterGroup) Get(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodGet, pattern, handler)
}

func (group *RouterGroup) Post(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodPost, pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}
