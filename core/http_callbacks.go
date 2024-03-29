package core

import (
	"encoding/json"
	"fmt"
	"github.com/GabeCordo/etl/components/cluster"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type JSONResponse struct {
	Status      int    `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	Data        any    `json:"data,omitempty"`
}

type ClusterConfigJSONBody struct {
	Cluster     string  `json:"cluster"`
	Version     float64 `json:"version,omitempty"`
	Mounted     bool    `json:"mounted"`
	DynamicPath string  `json:"dynamic-path,omitempty"`
}

func (httpThread *HttpThread) clusterCallback(w http.ResponseWriter, r *http.Request) {

	request := &ClusterConfigJSONBody{}
	err := json.NewDecoder(r.Body).Decode(request)
	if (r.Method != "GET") && (err != nil) {
		fmt.Println("bad")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		if clusterList, success := ClusterList(); success {
			bytes, _ := json.Marshal(clusterList)
			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if r.Method == "PUT" {
		if request.Mounted {
			success := ClusterMount(httpThread.C5, httpThread.provisionerResponseTable, request.Cluster)
			if !success {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			success := ClusterUnMount(httpThread.C5, httpThread.provisionerResponseTable, request.Cluster)
			if !success {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	} else if r.Method == "POST" {
		success, description := DynamicallyRegisterCluster(httpThread.C5, httpThread.provisionerResponseTable, request.Cluster, request.DynamicPath, request.Mounted)
		if !success {
			response := &JSONResponse{Description: description}
			bytes, _ := json.Marshal(response)
			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
	} else if r.Method == "DELETE" {
		success := DynamicallyDeleteCluster(httpThread.C5, httpThread.provisionerResponseTable, request.Cluster)
		if !success {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type SupervisorConfigJSONBody struct {
	Cluster    string `json:"cluster"`
	Config     string `json:"config"`
	Supervisor uint64 `json:"id,omitempty"`
}

type SupervisorProvisionJSONResponse struct {
	Cluster    string `json:"cluster,omitempty"`
	Supervisor uint64 `json:"id,omitempty"`
}

func (httpThread *HttpThread) supervisorCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	var request SupervisorConfigJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if (r.Method != "GET") && (err != nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {

		clusterName, foundClusterName := urlMapping["cluster"]
		supervisorIdStr, foundSupervisorId := urlMapping["id"]

		if !foundSupervisorId || !foundClusterName {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			if supervisorId, err := strconv.ParseUint(supervisorIdStr[0], 10, 64); err != nil {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				if supervisor, found := SupervisorLookup(clusterName[0], supervisorId); found {
					bytes, _ := json.Marshal(supervisor)
					if _, err := w.Write(bytes); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}
		}
	} else if r.Method == "POST" {
		if supervisorId, success, description := SupervisorProvision(httpThread.C5, httpThread.provisionerResponseTable, request.Cluster, request.Config); success {
			response := &SupervisorProvisionJSONResponse{Cluster: request.Cluster, Supervisor: supervisorId}
			bytes, _ := json.Marshal(response)
			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(description))
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (httpThread *HttpThread) configCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	request := &cluster.Config{}
	err := json.NewDecoder(r.Body).Decode(request)
	if (r.Method != "GET") && (err != nil) {
		fmt.Println("bad")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {

		log.Println("Config Get")

		clusterName, foundClusterName := urlMapping["cluster"]
		if foundClusterName {
			if config, found := GetConfigFromDatabase(httpThread.C1, httpThread.databaseResponseTable, clusterName[0]); found {
				bytes, _ := json.Marshal(config)
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	} else if r.Method == "POST" {

		isOk := StoreConfigInDatabase(httpThread.C1, httpThread.databaseResponseTable, *request)
		if !isOk {
			w.WriteHeader(http.StatusConflict)
		}

	} else if r.Method == "PUT" {
		isOk := ReplaceConfigInDatabase(httpThread.C1, httpThread.databaseResponseTable, *request)
		if !isOk {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (httpThread *HttpThread) statisticCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	if r.Method == "GET" {

		clusterName, clusterNameFound := urlMapping["cluster"]
		if clusterNameFound {
			statistics, found := FindStatistics(httpThread.C1, httpThread.databaseResponseTable, clusterName[0])
			if found {
				bytes, err := json.Marshal(statistics)
				if err == nil {
					if _, err = w.Write(bytes); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type DebugJSONBody struct {
	Action string `json:"action"`
}

type DebugJSONResponse struct {
	Duration time.Duration `json:"time-elapsed"`
	Success  bool          `json:"success"`
}

func (httpThread *HttpThread) debugCallback(w http.ResponseWriter, r *http.Request) {

	var request DebugJSONBody
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		if request.Action == "shutdown" {
			ShutdownNode(httpThread.Interrupt)
		} else if request.Action == "ping" {
			startTime := time.Now()
			success := PingNodeChannels(httpThread.C1, httpThread.databaseResponseTable, httpThread.C5, httpThread.provisionerResponseTable)
			response := DebugJSONResponse{Success: success, Duration: time.Now().Sub(startTime)}
			bytes, err := json.Marshal(response)
			if err == nil {
				if _, err := w.Write(bytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
