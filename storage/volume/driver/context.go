package driver

import (
	log "github.com/sirupsen/logrus"
)

var ctx = Context{
	Log: log.StandardLogger(),
}

// Context represents driver context struct.
type Context struct {
	Log  *log.Logger
	data map[string]interface{}
}

// Contexts returns driver context instance.
func Contexts() Context {
	return ctx
}

// New is used to make a Context instance.
func (c Context) New() Context {
	return Context{
		Log:  c.Log,
		data: make(map[string]interface{}),
	}
}

// Add is used to add a key into context data.
func (c Context) Add(key string, value interface{}) {
	c.data[key] = value
}

// Del is used to delete a key from context data.
func (c Context) Del(key string) {
	delete(c.data, key)
}

// Get returns content of key from context data.
func (c Context) Get(key string) (interface{}, bool) {
	v, ok := c.data[key]
	return v, ok
}

// GetString returns string content of key from context data.
func (c Context) GetString(key string) (string, bool) {
	v, ok := c.Get(key)
	if ok {
		return v.(string), true
	}
	return "", false
}

// GetInt returns int content of key from context data.
func (c Context) GetInt(key string) (int, bool) {
	v, ok := c.Get(key)
	if ok {
		return v.(int), true
	}
	return 0, false
}
