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

func (helper Helper) IsDebugEnabled() bool {
	return GetConfigInstance().Debug
}

func (helper Helper) SaveToCache(data any) *CacheResponsePromise {

	var expiry float64
	if GetConfigInstance().Cache.Expiry != 0.0 {
		expiry = GetConfigInstance().Cache.Expiry
	} else {
		expiry = DefaultTimeout
	}

	requestNonce := rand.Uint32()
	helper.core.C9 <- CacheRequest{Action: CacheSaveIn, Data: data, Nonce: requestNonce, ExpiresIn: expiry}

	responseChannel := GetProvisionerMemoryInstance().CreateCacheResponseEventListener(requestNonce)
	promise := NewCacheResponsePromise(requestNonce, responseChannel)

	return promise
}

func (helper Helper) LoadFromCache(identifier string) *CacheResponsePromise {

	requestNonce := rand.Uint32()
	helper.core.C9 <- CacheRequest{Action: CacheLoadFrom, Identifier: identifier, Nonce: requestNonce}

	responseChannel := GetProvisionerMemoryInstance().CreateCacheResponseEventListener(requestNonce)
	promise := NewCacheResponsePromise(requestNonce, responseChannel)

	return promise
}

func (helper Helper) Log(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.core.C11 <- MessengerRequest{Action: MessengerLog, Cluster: cluster, Message: message, Nonce: requestNonce}
}

func (helper Helper) Warning(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.core.C11 <- MessengerRequest{Action: MessengerWarning, Cluster: cluster, Message: message, Nonce: requestNonce}
}

func (helper Helper) Fatal(cluster, message string) {

	requestNonce := rand.Uint32()
	helper.core.C11 <- MessengerRequest{Action: MessengerFatal, Cluster: cluster, Message: message, Nonce: requestNonce}
}
