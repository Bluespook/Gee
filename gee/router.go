package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		make(map[string]*node),
		make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	res := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range res {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	//log.Printf("Route %4s - %s", method, pattern)
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	params := make(map[string]string)
	searchPath := parsePattern(path) //实际路径
	res := root.search(searchPath, 0)
	if res != nil {
		parts := parsePattern(res.pattern) //模糊路径
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchPath[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchPath[index:], "/")
				break
			}
		}
		return res, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, param := r.getRoute(c.Method, c.Path)
	//log.Println(n.pattern)
	if n != nil {
		c.Params = param
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 not found: %s\n", c.Path)
		})
	}
	c.Next()
}
