package egg

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Context struct {
	Writer     http.ResponseWriter
	Req        *http.Request
	Path       string
	Method     string
	Params     map[string]string
	StatusCode int
	handlers   []HandlerFunc
	index      int
	engine     *Engine
}

type H map[string]interface{}

//newContext 构造函数
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

//调用下一个中间件，该方法只能在中间件内部调用
func (c *Context) Next() {
	//第一次进入，++为0
	c.index++
	l := len(c.handlers)
	for ; c.index < l; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

//Query 获取Get请求参数
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

//PostForm 获取Post请求数据
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

//Status add status code
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

//SetHeader add header
func (c *Context) SetHeader(k, v string) {
	c.Writer.Header().Set(k, v)
}

//String return str in resp
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

//JSON build and return json in resp 使用模版方法抽象-> gin.Render
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json;charset=utf-8")
	c.Status(code)
	if err := writeJSON(c.Writer, obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
		//panic(err)
	}

}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

//writeJSON
func writeJSON(w http.ResponseWriter, obj interface{}) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplate.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
}
