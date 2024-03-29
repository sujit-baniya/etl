package supervisor

import (
	"fmt"
	"github.com/GabeCordo/etl/components/channel"
	"github.com/GabeCordo/etl/components/cluster"
	"time"
)

const (
	DefaultNumberOfClusters       = 1
	DefaultMonitorRefreshDuration = 1
	DefaultChannelThreshold       = 10
	DefaultChannelGrowthFactor    = 2
)

func NewSupervisor(clusterName string, clusterImplementation cluster.Cluster) *Supervisor {
	supervisor := new(Supervisor)

	supervisor.group = clusterImplementation
	supervisor.Config = cluster.Config{
		clusterName,
		DefaultChannelThreshold,
		DefaultNumberOfClusters,
		DefaultNumberOfClusters,
		DefaultChannelGrowthFactor,
		DefaultChannelThreshold,
		DefaultChannelGrowthFactor,
		DefaultChannelGrowthFactor,
	}
	supervisor.Stats = cluster.NewStatistics()
	supervisor.etChannel = channel.NewManagedChannel(supervisor.Config.ETChannelThreshold, supervisor.Config.ETChannelGrowthFactor)
	supervisor.tlChannel = channel.NewManagedChannel(supervisor.Config.TLChannelThreshold, supervisor.Config.TLChannelGrowthFactor)

	return supervisor
}

func NewCustomSupervisor(clusterImplementation cluster.Cluster, config cluster.Config) *Supervisor {
	supervisor := new(Supervisor)

	/**
	 * Note: we may wish to dynamically modify the threshold and growth-factor rates
	 *       used by the managed channels to vary how provisioning of new transform and
	 *       load goroutines are created. This allows us to create an autonomous system
	 *       that "self improves" if the output of the monitor is looped back
	 */

	supervisor.group = clusterImplementation
	supervisor.Config = config
	supervisor.Stats = cluster.NewStatistics()
	supervisor.etChannel = channel.NewManagedChannel(config.ETChannelThreshold, config.ETChannelGrowthFactor)
	supervisor.tlChannel = channel.NewManagedChannel(config.TLChannelThreshold, config.TLChannelGrowthFactor)

	return supervisor
}

func (supervisor *Supervisor) Event(event Event) bool {
	supervisor.mutex.Lock()
	defer supervisor.mutex.Unlock()

	if supervisor.State == UnTouched {
		if event == Startup {
			supervisor.State = Running
		} else {
			return false
		}
	} else if supervisor.State == Running {
		if event == StartProvision {
			supervisor.State = Provisioning
		} else if event == Error {
			supervisor.State = Failed
		} else if event == TearedDown {
			supervisor.State = Terminated
		} else {
			return false
		}
	} else if supervisor.State == Provisioning {
		if event == EndProvision {
			supervisor.State = Running
		} else if event == Error {
			supervisor.State = Failed
		} else {
			return false
		}
	} else if (supervisor.State == Failed) || (supervisor.State == Terminated) {
		return false
	}

	return true // represents a boolean ~ hasStateChanged?
}

func (supervisor *Supervisor) Start() (response *cluster.Response) {
	supervisor.Event(Startup)
	defer supervisor.Event(TearedDown)
	defer func() {
		if r := recover(); r != nil {
			response = cluster.NewResponse(
				supervisor.Config,
				supervisor.Stats,
				time.Now().Sub(supervisor.StartTime),
				true,
			)
		}
	}()

	supervisor.StartTime = time.Now()

	// start creating the default frontend goroutines
	supervisor.Provision(cluster.Extract)
	supervisor.waitGroup.Add(1)

	for i := 0; i < supervisor.Config.StartWithNTransformClusters; i++ {
		supervisor.Provision(cluster.Transform)
		supervisor.waitGroup.Add(1)
	}
	for i := 0; i < supervisor.Config.StartWithNLoadClusters; i++ {
		supervisor.Provision(cluster.Load)
		supervisor.waitGroup.Add(1)
	}
	// end creating the default frontend goroutines

	// every N seconds we should check if the etChannel or tlChannel is congested
	// and requires us to provision additional nodes
	go supervisor.Runtime()

	supervisor.waitGroup.Wait() // wait for the Extract-Transform-Load (ETL) Cycle to Complete

	response = cluster.NewResponse(
		supervisor.Config,
		supervisor.Stats,
		time.Now().Sub(supervisor.StartTime),
		false,
	)

	return response
}

func (supervisor *Supervisor) Runtime() {
	for {
		// is etChannel congested?
		if supervisor.etChannel.State == channel.Congested {
			supervisor.Stats.NumEtThresholdBreaches++
			n := supervisor.Stats.NumProvisionedTransformRoutes
			for n > 0 {
				supervisor.Provision(cluster.Transform)
				n--
			}
			supervisor.Stats.NumProvisionedTransformRoutes *= supervisor.etChannel.Config.GrowthFactor
		}

		// is tlChannel congested?
		if supervisor.tlChannel.State == channel.Congested {
			supervisor.Stats.NumTlThresholdBreaches++
			n := supervisor.Stats.NumProvisionedLoadRoutines
			for n > 0 {
				supervisor.Provision(cluster.Load)
				n--
			}
			supervisor.Stats.NumProvisionedLoadRoutines *= supervisor.tlChannel.Config.GrowthFactor
		}

		// check if the channel is congested after DefaultMonitorRefreshDuration seconds
		time.Sleep(DefaultMonitorRefreshDuration * time.Second)
	}
}

func (supervisor *Supervisor) Provision(segment cluster.Segment) {
	supervisor.Event(StartProvision)
	defer supervisor.Event(EndProvision)

	go func() {
		switch segment {
		case cluster.Extract:
			supervisor.Stats.NumProvisionedExtractRoutines++
			supervisor.group.ExtractFunc(supervisor.etChannel.Channel)
			break
		case cluster.Transform: // transform
			supervisor.Stats.NumProvisionedTransformRoutes++
			supervisor.group.TransformFunc(supervisor.etChannel.Channel, supervisor.tlChannel.Channel)
			break
		default: // load
			supervisor.Stats.NumProvisionedLoadRoutines++
			supervisor.group.LoadFunc(supervisor.tlChannel.Channel)
			break
		}
		supervisor.waitGroup.Done() // notify the wait group a process has completed ~ if all are finished we close the monitor
	}()
}

func (supervisor *Supervisor) Print() {
	fmt.Printf("Id: %d\n", supervisor.Id)
	fmt.Printf("Cluster: %s\n", supervisor.Config.Identifier)
}

func (status Status) String() string {
	switch status {
	case UnTouched:
		return "UnTouched"
	case Running:
		return "Running"
	case Provisioning:
		return "Provisioning"
	case Failed:
		return "Failed"
	case Terminated:
		return "Terminated"
	default:
		return "None"
	}
}
