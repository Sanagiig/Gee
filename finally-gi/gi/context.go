package gi

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Context struct {
	Method     string
	Path       string
	Req        *http.Request
	RespWriter http.ResponseWriter
	Params     map[string]string
	index      int
	handlers   []HTTPHandler
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Req:        r,
		RespWriter: w,
		Method:     r.Method,
		Path:       r.URL.Path,
		index:      0,
	}
}

func (c *Context) Next() {
	if c.index < len(c.handlers) {
		handler := c.handlers[c.index]

		c.index++
		handler(c)
		c.Next()
	}
}

func (c *Context) Query(key string) string {
	res := c.Req.URL.Query()[key]
	l := len(res)
	switch {
	case l == 1:
		return res[0]
	case l > 1:
		return strings.Join(res, ",")
	default:
		return ""
	}
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Status(code int) {
	c.RespWriter.WriteHeader(code)
}

func (c *Context) SetHeader(k string, v ...string) {
	for _, val := range v {
		m := c.RespWriter.Header()
		m.Add(k, val)
	}
}

func (c *Context) normalWrite(str *string) {
	c.Status(http.StatusOK)
	c.RespWriter.Write([]byte(*str))
}

func (c *Context) Data(data []byte) {
	c.SetHeader("Content-Type", "application/octet-stream")
	c.Status(http.StatusOK)
	c.RespWriter.Write(data)
}

func (c *Context) String(str string) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(http.StatusOK)
	c.RespWriter.Write([]byte(str))
}

func (c *Context) Json(data any) {
	c.SetHeader("Content-Type", "application/json")

	jencode := json.NewEncoder(c.RespWriter)
	err := jencode.Encode(data)
	if err != nil {
		c.Err(err.Error())
	}
}

func (c *Context) Fail(code int, str string) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.String(str)
}

func (c *Context) Err(str string) {
	c.SetHeader("Content-Type", "text/plain")
	c.normalWrite(&str)
}
