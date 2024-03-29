package core

import (
	"github.com/GabeCordo/etl/components/utils"
	"sync"
)

type SupervisorAction int8

const (
	ProvisionerProvision SupervisorAction = iota
	ProvisionerDynamicLoad
	ProvisionerDynamicDelete
	ProvisionerMount
	ProvisionerUnMount
	ProvisionerTeardown
	ProvisionerLowerPing
)

type ProvisionerRequest struct {
	Action     SupervisorAction `json:"Action"`
	Nonce      uint32           `json:"Nonce"`
	Cluster    string           `json:"cluster"`
	Mount      bool             `json:"mount,omitempty"`
	Config     string           `json:"config,omitempty"`
	Path       string           `json:"path,omitempty"`
	Parameters []string         `json:"parameters,omitempty"`
}

type ProvisionerResponse struct {
	Nonce        uint32 `json:"nonce"`
	Success      bool   `json:"success"`
	Cluster      string `json:"cluster"`
	Description  string `json:"description"`
	SupervisorId uint64 `json:"supervisor-id"`
}

type ProvisionerThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C5 <-chan ProvisionerRequest  // Supervisor is receiving core from the http_thread
	C6 chan<- ProvisionerResponse // Supervisor is sending responses to the http_thread

	C7 chan<- DatabaseRequest  // Supervisor is sending core to the database
	C8 <-chan DatabaseResponse // Supervisor is receiving responses from the database

	C9  chan<- CacheRequest  // Provisioner is sending requests to the cache
	C10 <-chan CacheResponse // Provisioner is receiving responses from the CacheThread

	C11 chan<- MessengerRequest // Provisioner is sending request to the messenger

	databaseResponseTable *utils.ResponseTable
	cacheResponseTable    *utils.ResponseTable

	accepting bool
	wg        sync.WaitGroup
}

func NewProvisioner(channels ...interface{}) (*ProvisionerThread, bool) {
	provisioner := new(ProvisionerThread)
	var ok bool

	provisioner.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	provisioner.C5, ok = (channels[1]).(chan ProvisionerRequest)
	if !ok {
		return nil, ok
	}
	provisioner.C6, ok = (channels[2]).(chan ProvisionerResponse)
	if !ok {
		return nil, ok
	}
	provisioner.C7, ok = (channels[3]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	provisioner.C8, ok = (channels[4]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}
	provisioner.C9, ok = (channels[5]).(chan CacheRequest)
	if !ok {
		return nil, ok
	}
	provisioner.C10, ok = (channels[6]).(chan CacheResponse)
	if !ok {
		return nil, ok
	}
	provisioner.C11, ok = (channels[7]).(chan MessengerRequest)
	if !ok {
		return nil, ok
	}

	provisioner.databaseResponseTable = utils.NewResponseTable()
	provisioner.cacheResponseTable = utils.NewResponseTable()

	return provisioner, ok
}
