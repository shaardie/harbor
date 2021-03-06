package exporter

import (
	"encoding/json"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/prometheus/client_golang/prometheus"
)

// SystemInfoCollectorName ...
const (
	systemInfoCollectorName = "SystemInfoCollector"
	sysInfoURL              = "/api/v2.0/systeminfo"
)

var (
	harborSysInfo = typedDesc{
		desc: newDescWithLables("", "system_info", "Information of Harbor system",
			"auth_mode",
			"registry_url",
			"external_url",
			"harbor_version",
			"registry_storage_provider"),
		valueType: prometheus.GaugeValue,
	}
)

// NewSystemInfoCollector ...
func NewSystemInfoCollector(hbrCli *HarborClient) *SystemInfoCollector {
	return &SystemInfoCollector{
		HarborClient: hbrCli,
	}
}

// SystemInfoCollector ...
type SystemInfoCollector struct {
	*HarborClient
}

// Describe implements prometheus.Collector
func (hc *SystemInfoCollector) Describe(c chan<- *prometheus.Desc) {
	c <- harborSysInfo.Desc()
}

// Collect implements prometheus.Collector
func (hc *SystemInfoCollector) Collect(c chan<- prometheus.Metric) {
	for _, m := range hc.getSysInfo() {
		c <- m
	}
}

func (hc *SystemInfoCollector) getSysInfo() []prometheus.Metric {
	if CacheEnabled() {
		value, ok := CacheGet(systemInfoCollectorName)
		if ok {
			return value.([]prometheus.Metric)
		}
	}
	result := []prometheus.Metric{}
	res, err := hbrCli.Get(sysInfoURL)
	if err != nil {
		log.Errorf("request health info failed with err: %v", err)
		return result
	}
	defer res.Body.Close()
	var sysInfoResponse responseSysInfo
	json.NewDecoder(res.Body).Decode(&sysInfoResponse)
	result = append(result, harborSysInfo.MustNewConstMetric(1,
		sysInfoResponse.AuthMode,
		sysInfoResponse.RegistryURL,
		sysInfoResponse.ExternalURL,
		sysInfoResponse.HarborVersion,
		sysInfoResponse.StorageProvider))
	if CacheEnabled() {
		CachePut(systemInfoCollectorName, result)
	}
	return result
}

type responseSysInfo struct {
	AuthMode        string `json:"auth_mode"`
	RegistryURL     string `json:"registry_url"`
	ExternalURL     string `json:"external_url"`
	HarborVersion   string `json:"harbor_version"`
	StorageProvider string `json:"registry_storage_provider_name"`
}
