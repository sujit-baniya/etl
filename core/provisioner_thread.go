package core

import (
	"github.com/GabeCordo/etl/components/cluster"
	"log"
	"time"
)

const (
	DefaultHardTerminateTime = 30 // minutes
)

var provisioner *cluster.Provisioner

func GetProvisionerInstance() *cluster.Provisioner {

	if provisioner == nil {
		provisioner = cluster.NewProvisioner()
	}

	return provisioner
}

func (provisionerThread *ProvisionerThread) Setup() {
	provisionerThread.accepting = true
	provisionerInstance := GetProvisionerInstance() // create the supervisor if it doesn't exist

	GetProvisionerMemoryInstance() // create the instance

	// auto-mounting is supported within the etl Config; if a cluster Identifier is added
	// to the config under 'auto-mount', it is added to the map of Operational functions
	for _, identifier := range GetConfigInstance().AutoMount {
		provisionerInstance.Mount(identifier)
	}
}

func (provisionerThread *ProvisionerThread) Start() {
	go func() {
		// request coming from http_server
		for request := range provisionerThread.C5 {
			if !provisionerThread.accepting {
				break
			}
			provisionerThread.wg.Add(1)
			provisionerThread.ProcessIncomingRequests(request)
		}

		provisionerThread.wg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C8 {
			if !provisionerThread.accepting {
				break
			}
			provisionerThread.ProcessesIncomingDatabaseResponses(response)
		}

		provisionerThread.wg.Wait()
	}()
	go func() {
		for response := range provisionerThread.C10 {
			if !provisionerThread.accepting {
				break
			}
			provisionerThread.ProcessIncomingCacheResponses(response)
		}

		provisionerThread.wg.Wait()
	}()

	provisionerThread.wg.Wait()
}

func (provisionerThread *ProvisionerThread) ProcessIncomingRequests(request ProvisionerRequest) {
	provisionerInstance := GetProvisionerInstance()

	if request.Action == Mount {
		provisionerInstance.Mount(request.Cluster)
		log.Printf("Mounted cluster {%s}", request.Cluster)
		provisionerThread.wg.Done()
	} else if request.Action == UnMount {
		provisionerInstance.UnMount(request.Cluster)
		log.Printf("UnMounted cluster {%s}", request.Cluster)
		provisionerThread.wg.Done()
	} else if request.Action == Provision {
		if !provisionerInstance.IsMounted(request.Cluster) {
			log.Printf("Could not provision cluster {%s}; cluster was not mounted", request.Cluster)
			provisionerThread.wg.Done()
			return
		} else {
			log.Printf("Provisioning cluster {%s}", request.Cluster)
		}

		clstr, cnfg, register, ok := provisionerInstance.Function(request.Cluster)
		if !ok {
			log.Println("there is a corrupted cluster in the supervisor")
			provisionerThread.wg.Done()
			return
		}

		var supervisor *cluster.Supervisor
		if cnfg == nil {
			supervisor = cluster.NewSupervisor(*clstr)
		} else {
			supervisor = cluster.NewCustomSupervisor(*clstr, cnfg)
		}
		register.Register(supervisor)

		go func() {
			var response *cluster.Response

			c := make(chan struct{})
			go func() {
				defer close(c)
				response = supervisor.Start()
			}()

			// block until the supervisor completes
			<-c

			// don't send the statistics of the cluster to the database unless an Identifier has been
			// given to the cluster for grouping purposes
			if len(supervisor.Config.Identifier) != 0 {
				// saves statistics to the database thread
				dbRequest := DatabaseRequest{Action: Store, Origin: Provisioner, Cluster: supervisor.Config.Identifier, Data: response}
				provisionerThread.C7 <- dbRequest

				// sends a completion message to the messenger thread to write to a log file or send an email regarding completion
				msgRequest := MessengerRequest{Action: Close, Cluster: supervisor.Config.Identifier}
				provisionerThread.C11 <- msgRequest

				// provide the console with output indicating that the cluster has completed
				// we already provide output when a cluster is provisioned, so it completes the state
				if GetConfigInstance().Debug {
					duration := time.Now().Sub(supervisor.StartTime)
					log.Printf("Cluster transformations complete, took %dhr %dm %ds {%s}\n",
						int(duration.Hours()),
						int(duration.Minutes()),
						int(duration.Seconds()),
						supervisor.Config.Identifier,
					)
				}
			}

			// let the provisioner thread decrement the semaphore otherwise we will be stuck in deadlock waiting for
			// the provisioned cluster to complete before allowing the etl-framework to shut down
			provisionerThread.wg.Done()
		}()
	} else if request.Action == Teardown {
		// TODO - not implemented
	}
}

func (provisionerThread *ProvisionerThread) ProcessesIncomingDatabaseResponses(response DatabaseResponse) {
	// TODO - implement
	// GetProvisionerMemoryInstance().LoadDatabaseResponse(response.Nonce, response)
}

func (provisionerThread *ProvisionerThread) ProcessIncomingCacheResponses(response CacheResponse) {
	GetProvisionerMemoryInstance().SendCacheResponseEvent(response.Nonce, response)
}

func (provisionerThread *ProvisionerThread) Teardown() {
	provisionerThread.accepting = false

	provisionerThread.wg.Wait()
}
