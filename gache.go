package gache

import (
	"fmt"
	"sync"
	"time"
)

type Cache interface {
	Group
	Group(key string) (Group, bool)
	NewGroup(key string, expiration time.Duration, fillFunc FillFunc) error
	DelGroup(key string)
	GetGroupVal(gkey, vkey string) (interface{}, bool)
	SetGroupVal(gkey, vkey string, val interface{}) error
}

type Group interface {
	Get(key string) (interface{}, bool)
	Set(key string, val interface{})
	Del(key string)
	SetExpiration(expiration time.Duration)
	SetFillFunc(fillFunc FillFunc)
}

type FillFunc func(key string) (interface{}, bool)

type cache struct {
	*group
	groups map[string]*group
}

func NewCache(expiration time.Duration, fillFunc FillFunc) Cache {
	if expiration < 0 {
		expiration = 0
	}

	return &cache{
		group: &group{
			values:     make(map[string]value),
			fillFunc:   fillFunc,
			expiration: expiration,
		},
		groups: make(map[string]*group),
	}
}

func (c *cache) Group(key string) (Group, bool) {
	c.mx.Lock()
	g, ok := c.groups[key]
	c.mx.Unlock()

	return g, ok
}

func (c *cache) NewGroup(key string, expiration time.Duration, fillFunc FillFunc) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	if _, exists := c.groups[key]; exists {
		return fmt.Errorf("group with key %q already exists", key)
	}

	if expiration < 0 {
		expiration = 0
	}

	c.groups[key] = &group{
		values:     make(map[string]value),
		fillFunc:   fillFunc,
		expiration: expiration,
	}

	return nil
}

func (c *cache) DelGroup(key string) {
	c.mx.Lock()
	delete(c.groups, key)
	c.mx.Unlock()
}

func (c *cache) GetGroupVal(gkey, vkey string) (interface{}, bool) {
	c.mx.Lock()
	g, ok := c.groups[gkey]
	c.mx.Unlock()

	if !ok {
		return nil, false
	}

	return g.Get(vkey)
}

func (c *cache) SetGroupVal(gkey, vkey string, val interface{}) error {
	c.mx.Lock()
	g, ok := c.groups[gkey]
	c.mx.Unlock()

	if !ok {
		return fmt.Errorf("group with key %q doesn't exist", gkey)
	}

	g.Set(vkey, val)

	return nil
}

type value struct {
	data       interface{}
	expiration int64
}

type group struct {
	mx         sync.Mutex
	values     map[string]value
	fillFunc   FillFunc
	expiration time.Duration
}

func (g *group) Get(key string) (interface{}, bool) {
	g.mx.Lock()
	v, ok := g.values[key]
	g.mx.Unlock()

	now := time.Now()
	if ok && (v.expiration == 0 || v.expiration > now.UnixNano()) {
		return v.data, true
	}

	if g.fillFunc == nil {
		g.mx.Lock()
		delete(g.values, key)
		g.mx.Unlock()
		return nil, false
	}

	v.data, ok = g.fillFunc(key)

	if !ok {
		g.mx.Lock()
		delete(g.values, key)
		g.mx.Unlock()
		return nil, false
	}

	if g.expiration != 0 {
		v.expiration = now.Add(g.expiration).UnixNano()
	}

	g.mx.Lock()
	g.values[key] = v
	g.mx.Unlock()

	return v.data, true
}

func (g *group) Set(key string, val interface{}) {
	g.mx.Lock()

	var expiration int64
	if g.expiration != 0 {
		expiration = time.Now().Add(g.expiration).UnixNano()
	}

	g.values[key] = value{
		data:       val,
		expiration: expiration,
	}

	g.mx.Unlock()
}

func (g *group) Del(key string) {
	g.mx.Lock()
	delete(g.values, key)
	g.mx.Unlock()
}

func (g *group) SetExpiration(expiration time.Duration) {
	if expiration <= 0 {
		expiration = 0
	}

	g.mx.Lock()
	g.expiration = expiration
	g.mx.Unlock()
}

func (g *group) SetFillFunc(fillFunc FillFunc) {
	g.mx.Lock()
	g.fillFunc = fillFunc
	g.mx.Unlock()
}
