package core

import (
	"github.com/GabeCordo/etl/components/cache"
	"log"
	"time"
)

var CacheInstance *cache.Cache

func GetCacheInstance() *cache.Cache {
	if CacheInstance == nil {
		CacheInstance = cache.NewCache()
	}
	return CacheInstance
}

func (cache *CacheThread) Setup() {
	cache.accepting = true
}

func (cache *CacheThread) Start() {
	go func() {
		// request from http_server
		for request := range cache.C9 {
			if !cache.accepting {
				break
			}
			cache.wg.Add(1)
			cache.ProcessIncomingRequest(&request)
		}
	}()

	go func() {
		// cleaning the cache of expired records
		for cache.accepting {
			time.Sleep(1 * time.Minute)
			// every minute, attempt to clean the cache by removing any records that
			// may have expired since we last checked
			GetCacheInstance().Clean()
		}
	}()

	cache.wg.Wait()
}

func (cache *CacheThread) ProcessIncomingRequest(request *CacheRequest) {
	if request.Action == CacheSaveIn {
		cache.ProcessSaveRequest(request)
	} else if request.Action == CacheLoadFrom {
		cache.ProcessLoadRequest(request)
	} else if request.Action == CacheLowerPing {
		cache.ProcessPingCache(request)
	}

	cache.wg.Done()
}

// ProcessSaveRequest will insert or override an existing cache record
func (cache CacheThread) ProcessSaveRequest(request *CacheRequest) {
	var response CacheResponse
	if _, found := GetCacheInstance().Get(request.Identifier); found {
		GetCacheInstance().Swap(request.Identifier, request.Data, request.ExpiresIn)
		response = CacheResponse{Identifier: request.Identifier, Data: nil, Nonce: request.Nonce, Success: true}
	} else {
		newIdentifier := GetCacheInstance().Save(request.Data, request.ExpiresIn)
		response = CacheResponse{Identifier: newIdentifier, Data: nil, Nonce: request.Nonce, Success: true}
	}
	cache.C10 <- response
}

func (cache CacheThread) ProcessLoadRequest(request *CacheRequest) {
	cacheData, isFoundAndNotExpired := GetCacheInstance().Get(request.Identifier)
	cache.C10 <- CacheResponse{
		Identifier: request.Identifier,
		Data:       cacheData,
		Nonce:      request.Nonce,
		Success:    isFoundAndNotExpired && (cacheData != nil),
	}
}

func (cache *CacheThread) ProcessPingCache(request *CacheRequest) {

	if GetConfigInstance().Debug {
		log.Println("[etl_cache] received ping over C9")
	}

	cache.C10 <- CacheResponse{Nonce: request.Nonce, Success: true}
}

func (cache *CacheThread) Teardown() {
	cache.accepting = false
}
