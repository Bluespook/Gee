package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request
	//request info
	Path   string
	Method string
	Params map[string]string
	//response info
	Status int
	//middleware
	handlers []HandlerFunc
	index    int
	//engine pointer
	engine *Engine
}

func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	c := &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
	return c
}

func (c *Context) Next() {
	c.index += 1
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// PostForm 例如：curl "http://localhost:9999/login" -Method POST -Body "username=geektutu&password=1234"，从最后一个参数中获取key对应的val并返回
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// Query 例如curl http://localhost:9999/hello?name="lyl"，获取?后对应key的val并返回
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) StatusCode(code int) {
	c.Status = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.StatusCode(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.StatusCode(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.StatusCode(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.StatusCode(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

func (c *Context) Param(key string) string {
	val := c.Params[key]
	return val
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}
