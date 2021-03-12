package base

import (
	"sync"
)

type Context struct{ Ctx *sync.Map }

func (c *Context) Value(key string) (interface{}, bool) {
	if c.Ctx == nil {
		return nil, false
	}
	return c.Ctx.Load(key)
}

func (c *Context) SetKV(key string, v interface{}) {
	if c.Ctx == nil {
		c.Ctx = &sync.Map{}
	}
	c.Ctx.Store(key, v)
}

func (c *Context) Range(f func(key, value interface{}) bool) { c.Ctx.Range(f) }

func NewContext() *Context {
	return &Context{Ctx: &sync.Map{}}
}
