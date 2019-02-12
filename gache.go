package gache

import "time"

type Cache interface {
	Group
	Group(key string) (Group, bool)
	NewGroup(key string, expiration time.Duration, getValueFunc GetValueFunc) error
	DelGroup(key string)
	GetGroupVal(gkey, vkey string) (interface{}, bool)
	SetGroupVal(gkey, vkey string, val interface{})
}

type Group interface {
	Get(key string) (interface{}, bool)
	Set(key string, val interface{})
	Del(key string)
	SetExpiration(expiration time.Duration)
	SetGetValueFunc(getValueFunc GetValueFunc)
}
type GetValueFunc func(key string) (interface{}, error)
