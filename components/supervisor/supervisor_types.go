package supervisor

import (
	"github.com/GabeCordo/etl/components/channel"
	"github.com/GabeCordo/etl/components/cluster"
	"sync"
	"time"
)

const (
	MaxConcurrentSupervisors = 24
)

type Status uint8

const (
	UnTouched Status = iota
	Running
	Provisioning
	Failed
	Terminated
	Unknown
)

type Event uint8

const (
	Startup        Event = 0
	StartProvision       = 1
	EndProvision         = 2
	Error                = 3
	TearedDown           = 4
	StartReport          = 5
	EndReport            = 6
)

type SupervisorData struct {
	Id        uint64          `json:"id"`
	State     Status          `json:"state"`
	Mode      cluster.OnCrash `json:"on-crash"`
	StartTime time.Time       `json:"start-time"`
}

type Supervisor struct {
	Id        uint64              `json:"id"`
	group     cluster.Cluster     `json:"group"`
	Config    cluster.Config      `json:"config"`
	Stats     *cluster.Statistics `json:"stats"`
	State     Status              `json:"status"`
	mode      cluster.OnCrash     `json:"on-crash"`
	StartTime time.Time           `json:"start-time"`

	etChannel *channel.ManagedChannel
	tlChannel *channel.ManagedChannel

	waitGroup sync.WaitGroup
	mutex     sync.RWMutex
}

type SupervisorV2 struct {
	Data   SupervisorData      `json:"data"`
	Config *cluster.Config     `json:"config"`
	Stats  *cluster.Statistics `json:"stats"`

	etChannel *channel.ManagedChannel
	tlChannel *channel.ManagedChannel

	waitGroup sync.WaitGroup
	mutex     sync.RWMutex
}
