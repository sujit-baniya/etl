package core

import (
	"github.com/GabeCordo/etl/components/database"
	"github.com/GabeCordo/etl/components/utils"
	"sync"
)

type DatabaseAction uint8

const (
	DatabaseStore DatabaseAction = iota
	DatabaseFetch
	DatabaseReplace
	DatabaseDelete
	DatabaseUpperPing
	DatabaseLowerPing
)

type DatabaseRequest struct {
	Action  DatabaseAction    `json:"Action"`
	Nonce   uint32            `json:"Nonce"`
	Origin  Module            `json:"origin"`
	Type    database.DataType `json:"type"`
	Cluster string            `json:"cluster"` // aka. Identifier
	Data    any               `json:"data"`    // *cluster.Response `json:"Data"`
}

type DatabaseResponse struct {
	Nonce   uint32 `json:"Nonce"`
	Success bool   `json:"Success"`
	Data    any    `json:"statistics"` // []database.Entry or cluster.Config
}

type DatabaseThread struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 <-chan DatabaseRequest  // Database is receiving core from the http_thread
	C2 chan<- DatabaseResponse // Database is sending responses to the http_thread

	C3 chan<- MessengerRequest  // Database is sending core to the Messenger
	C4 <-chan MessengerResponse // Database is receiving responses from the Messenger

	C7 <-chan DatabaseRequest  // Database is receiving core from the Supervisor
	C8 chan<- DatabaseResponse // Database is sending responses to the Supervisor

	messengerResponseTable *utils.ResponseTable

	accepting bool
	wg        sync.WaitGroup
}

func NewDatabase(channels ...interface{}) (*DatabaseThread, bool) {
	database := new(DatabaseThread)
	var ok bool

	database.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	database.C1, ok = (channels[1]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	database.C2, ok = (channels[2]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}
	database.C3, ok = (channels[3]).(chan MessengerRequest)
	if !ok {
		return nil, ok
	}
	database.C4, ok = (channels[4]).(chan MessengerResponse)
	if !ok {
		return nil, ok
	}
	database.C7, ok = (channels[5]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	database.C8, ok = (channels[6]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}

	database.messengerResponseTable = utils.NewResponseTable()

	return database, ok
}
