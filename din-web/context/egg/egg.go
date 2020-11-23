package egg

import (
	"log"
	"net/http"
)

//HandlerFunc
type HandlerFunc func(c *Context)

type Engine struct {
	router *router
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := NewContext(w, req)
	engine.router.handle(c)
}

func New() *Engine {
	return &Engine{
		newRouter(),
	}
}

func (engine *Engine) addRouter(method, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)
	engine.router.addRouter(method, pattern, handler)
}

func (engine *Engine) Get(pattern string, handler HandlerFunc) {
	engine.addRouter(http.MethodGet, pattern, handler)
}

func (engine *Engine) Post(pattern string, handler HandlerFunc) {
	engine.addRouter(http.MethodPost, pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}
