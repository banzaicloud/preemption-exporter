package exporter

import (
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/host"
	log "github.com/sirupsen/logrus"
)

type terminationExporter struct {
	httpCli              *http.Client
	metadataEndpoint     string
	scrapeSuccessful     *prometheus.Desc
	terminationIndicator *prometheus.Desc
	terminationTime      *prometheus.Desc
}

type PreemptedData struct {
	Preempted string    `json:"action"`
	Time      time.Time `json:"time"`
}

func NewPreemptionExporter(me string) *terminationExporter {
	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: netTransport,
	}
	return &terminationExporter{
		httpCli:              client,
		metadataEndpoint:     me,
		scrapeSuccessful:     prometheus.NewDesc("gcp_instance_metadata_service_available", "Metadata service available", []string{"instance_id"}, nil),
		terminationIndicator: prometheus.NewDesc("gcp_instance_termination_imminent", "Instance is about to be terminated", []string{"instance_id"}, nil),
		terminationTime:      prometheus.NewDesc("gcp_instance_termination_in", "Instance will be terminated in", []string{"instance_id"}, nil),
	}
}

func (c *terminationExporter) get(path string) (string, int, error) {
	req, err := http.NewRequest("GET", c.metadataEndpoint+path, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}
	return string(body), resp.StatusCode, nil
}

func (c *terminationExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.scrapeSuccessful
	ch <- c.terminationIndicator
	ch <- c.terminationTime
}

func (c *terminationExporter) Collect(ch chan<- prometheus.Metric) {
	var preemptedValue float64
	log.Info("Fetching termination data from metadata-service")
	instanceID, statusCode, err := c.get("id")
	if err != nil {
		log.Errorf("couldn't parse instance id from metadata: %s", err.Error())
		return
	}
	if statusCode == 404 {
		log.Errorf("couldn't parse instance id from metadata: endpoint not found")
		return
	}
	preempted, statusCode, err := c.get("preempted")
	if err != nil {
		log.Errorf("Failed to fetch data from metadata service: %s", err)
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessful, prometheus.GaugeValue, 0, instanceID)
		return
	} else {
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessful, prometheus.GaugeValue, 1, instanceID)
		if statusCode == 404 {
			log.Debug("/preempted action endpoint not found")
			ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 0, "", instanceID)
			return
		}
		log.Infof("instance endpoint available, will be preempted: %v", preempted)
		if isPreempted, _ := strconv.ParseBool(preempted); isPreempted {
			preemptedValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, preemptedValue, instanceID)
		uptime, _ := host.Uptime()
		log.Infof("instance was started at : %v", uptime)
		ch <- prometheus.MustNewConstMetric(c.terminationTime, prometheus.GaugeValue, float64(uptime), instanceID)
	}
}
