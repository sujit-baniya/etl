# ETLFramework
A software orchestration framework for extract-transform-load process deployment and scaling. Developers can write and link custom ETL functions for data processing, that will be provisioned and scaled according to data velocity and processing demand made by the deployed functions. In production, ETL functions can be provisioned manually (or by script) through function calls over RPC using the "fast backend" framework. ETL processes can be mounted or unmounted depending on whether the administrator wishes to allow RPC calls to provision new instances of the ETL process.

### CLI Parameters

#### help
Provides descriptions for available parameters.

#### debug
Provides verbose output for the starter script.

#### key
Generated ECDSA public and private keys are outputted as x509 encoded formats.

### project 

#### 1. create project [project-name]
Creates a new ETL project associated with the <name> parameter.

#### 2. show project
Displays all the projects created on the local system

### cluster

#### 1. create cluster [cluster-name]
Creates a new cluster source and test file associated with the <name> parameter.

#### 2. delete cluster [cluster-name]
Deletes a cluster source and test file

#### 3. show cluster
Displays all clusters associated with the project with respective date-created, developer, and contact metadata.

### deploy
Runs the default entrypoint into the etl project.

### mount

#### 1. create mount [cluster-name]
Adds a cluster to the automount list in the config. When deployed, anyone with permission will be able to invoke the cluster.

#### 2. delete mount [cluster-name]
Removes a cluster from the automount list in the config. When deployed, the cluster will need to be manually mounted before it can be invoked over RPC.

#### 3. show mount
Shows a list of clusters that are automount in the project.

---

### Common Questions

#### What is the ETLFramework Core?
The core is defined as the entrypoint to the ETLFramework that allows developers to register new Clusters and inject custom configurations
using the *config.etl.json file*.

```go
c := core.NewCore()

m := Multiply{} 	// A structure implementing the etl.Cluster.Cluster interface
c.Cluster("multiply", m, cluster.Config{Identifier: "multiply"})

c.Run()	 // Starts the etl
```

#### What is a Cluster?
A cluster is defined as any structure that implements the ETLFramework.cluster.Cluster interface. Where the interface can be thought as 
the set of functions required to implement the business-logic of the Extract-Transform-Load (ETL) process. 

```go
type Cluster interface {
    ExtractFunc(output channel.OutputChannel)
    TransformFunc(input channel.InputChannel, output channel.OutputChannel)
    LoadFunc(input channel.InputChannel)
}
```

#### What does the ETLFramework do with a Cluster?
Once a cluster has been registered with the ETLFramework Core, it can be mounted and provisioned to initiate execution. Where an ETLCluster is linked by
go channels to pass data between the successive functions. The framework is responsible for monitoring the amount of data present within the channels, and if required, provisioning
additional thread to assist with data-processing.

```go
extract -[etChannel]-> transform -[tlChannel]-> load
```

etChannel : the channel between the (extract) and (transform) goroutines
tlChannel : the channel between the (transform) and (load) goroutines

##### How is provisioning handled?

Each channel (et and tl) has an associated threshold and growth factor. The developer has the option of specifying these quanities to
best match the ETL-process they are implementing or used the default as defined by the ETLFramework.

- Data Unit: A single object or structure past to the channel that is required by the next E-T-L function.
- Threshold (int): When 
- Growth Factor (double): By what scale should the number of successive functions exist if a channel is considered "congested"

#### Is Cluster Execution Guaranteed?

Each channel's completion is guaranteed by synchronous Wait Groups, where the ETLFramework Core will not
complete until every Cluster has completed processing.

#### How is Cluster Deadlock Prevented?

In theory, each channel has an infinite runtime until an operator has made a request to shut down the ETLFramework Core. Upon receiving the
shutdown interrupt, each Cluster will have 30 minutes to finish executing before being terminated. If this value does not fit your defined
scope, it can be modified in the *config.etl.json* under the "hard-terminate-time" flag as an integer representation of minutes.

### Provisioning a Cluster

#### What is Mounting?

Mounting indicates whether a cluster can be dynamically provisioned. When a cluster is registered
with the ETLFramework Core it is placed in a "registered" state, but is not operational. In order to
become operational where the cluster can be provisioned, it must be "mounted" to be placed into an operational state.

#### Why is Mounting important?

Mounting allows the operator to control what clusters are available during the lifetime of the system. When a cluster
is using excessive resources, is encountering unexpected errors, or has a possible vulnerability the operator can dismount
the cluster to stop further provisioning.

---

### HTTP Curl Interaction

#### Endpoints

##### /clusters
1. provision
2. mount
3. unmount

##### /data

1. mounts
2. supervisor
   1. lookup
   2. state

##### /statistics
1. [cluster-name]

##### /debug
1. shutdown
2. endpoints
   1. [endpoint-identifier]

#### Examples

##### Shutdown Node
curl -X GET http://127.0.0.1:8000/debug -H 'Content-Type: application/json' -d '{"function": "shutdown"}'

##### Test Cluster

###### View Mounted and Unmounted Clusters
curl -X GET http://127.0.0.1:8000/data -H 'Content-Type: application/json' -d '{"function": "mounts"}'

###### Mount Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "mount", "param":["multiply"]}'

###### UnMount Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "unmount", "param":["multiply"]}'

###### Provision Cluster
curl -X GET http://127.0.0.1:8000/clusters -H 'Content-Type: application/json' -d '{"function": "provision", "param":["multiply"]}'

##### Test Cluster Statistics
curl -X GET http://127.0.0.1:8000/statistics -H 'Content-Type: application/json' -d '{"function": "multiply", "param":["multiply"]}'
