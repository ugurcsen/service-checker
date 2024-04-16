package types

import (
	"time"
)

type ServerCheckConfig struct {
	Namespaces map[string]struct {
		Port         int    `yaml:"Port"`
		Protocol     string `yaml:"Protocol"`
		Link         string `yaml:"Link"`
		SSRCheck     bool   `yaml:"SSRCheck"`
		SSRThreshold int    `yaml:"SSRThreshold"`
	} `yaml:"Namespaces"`
	Hosts      []string `yaml:"Hosts"`
	OpenSearch *struct {
		Hosts    []string `yaml:"Hosts"`
		Index    string   `yaml:"Index"`
		Username string   `yaml:"Username"`
		Password string   `yaml:"Password"`
	} `yaml:"OpenSearch"`
}

type ResultStruct struct {
	Time          time.Time
	Host          string
	StatusCode    string
	latency       time.Duration
	Latency       float64
	ContentLength int
	SSR           *bool
}

func (r *ResultStruct) SetLatency(duration time.Duration) {
	r.latency = duration
	r.Latency = duration.Seconds()
}

func (r *ResultStruct) GetLatency() time.Duration {
	return r.latency
}
