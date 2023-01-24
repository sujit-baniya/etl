package core

import (
	"math/rand"
)

type Helper struct {
	core *Core
}

func NewHelper(core *Core) *Helper {
	helper := new(Helper)

	if core == nil {
		panic("cannot pass a nil pointer for a core, to a core helper")
	} else {
		helper.core = core
	}

	return helper
}

func (helper Helper) SaveToCache(data any) *CacheResponsePromise {
	promise := NewCacheResponsePromise()

	requestNonce := rand.Uint32()
	helper.core.C9 <- CacheRequest{Action: SaveInCache, Data: data, Nonce: requestNonce, ExpiresIn: DefaultTimeout}
	promise.nonce = requestNonce
	promise.wg.Add(1)

	return promise
}

func (helper Helper) LoadFromCache(identifier string) *CacheResponsePromise {
	promise := NewCacheResponsePromise()

	requestNonce := rand.Uint32()
	helper.core.C9 <- CacheRequest{Action: SaveInCache, Identifier: identifier, Nonce: requestNonce, ExpiresIn: DefaultTimeout}
	promise.nonce = requestNonce
	promise.wg.Add(1)

	return promise
}