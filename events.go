package elasticspot

import (
	"time"
)

type EventDetails struct {
	Agentconnected bool `json:"agentConnected"`
	Attributes     []struct {
		Name  string `json:"name"`
		Value string `json:"value,omitempty"`
	} `json:"attributes"`
	Clusterarn           string `json:"clusterArn"`
	Containerinstancearn string `json:"containerInstanceArn"`
	Ec2Instanceid        string `json:"ec2InstanceId"`
	Registeredresources  []struct {
		Name           string   `json:"name"`
		Type           string   `json:"type"`
		Integervalue   int      `json:"integerValue,omitempty"`
		Stringsetvalue []string `json:"stringSetValue,omitempty"`
	} `json:"registeredResources"`
	Remainingresources []struct {
		Name           string   `json:"name"`
		Type           string   `json:"type"`
		Integervalue   int      `json:"integerValue,omitempty"`
		Stringsetvalue []string `json:"stringSetValue,omitempty"`
	} `json:"remainingResources"`
	Status      string `json:"status"`
	Version     int    `json:"version"`
	Versioninfo struct {
		Agenthash     string `json:"agentHash"`
		Agentversion  string `json:"agentVersion"`
		Dockerversion string `json:"dockerVersion"`
	} `json:"versionInfo"`
	Updatedat         time.Time     `json:"updatedAt"`
	Registeredat      time.Time     `json:"registeredAt"`
	Pendingtaskscount int           `json:"pendingTasksCount"`
	Runningtaskscount int           `json:"runningTasksCount"`
	Attachments       []interface{} `json:"attachments"`
}
