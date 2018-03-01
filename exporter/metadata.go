package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
	"github.com/shirou/gopsutil/host"
)

type terminationExporter struct {
	metadataEndpoint     string
	scrapeSuccessful     *prometheus.Desc
	terminationIndicator *prometheus.Desc
	terminationTime      *prometheus.Desc
}

type PreemptedData struct {
	Preempted string    `json:"action"`
	Time   	  time.Time `json:"time"`
}

func NewPreemptionExporter(me string) *terminationExporter {
	return &terminationExporter{
		metadataEndpoint:     me,
		scrapeSuccessful:     prometheus.NewDesc("gcp_instance_metadata_service_available", "Metadata service available", []string{"instance_id"}, nil),
		terminationIndicator: prometheus.NewDesc("gcp_instance_termination_imminent", "Instance is about to be terminated", []string{"instance_id", "preemption_status"}, nil),
		terminationTime:      prometheus.NewDesc("gcp_instance_termination_in", "Instance will be terminated in", []string{"instance_id"}, nil),
	}
}

func (c *terminationExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.scrapeSuccessful
	ch <- c.terminationIndicator
	ch <- c.terminationTime
}

func (c *terminationExporter) Collect(ch chan<- prometheus.Metric) {
	log.Info("Fetching termination data from metadata-service")
	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	var instanceId string
	idResp, err := client.Get(c.metadataEndpoint + "id")

	if err != nil {
		log.Errorf("couldn't parse instance id from metadata: %s", err.Error())
		return
	}
	if idResp.StatusCode == 404 {
		log.Errorf("couldn't parse instance id from metadata: endpoint not found",)
		return
	}
	defer idResp.Body.Close()
	body, _ := ioutil.ReadAll(idResp.Body)
	instanceId = string(body)

	var preempted string
	preemptedResp, err := client.Get(c.metadataEndpoint + "preempted")

	preemptedResp, err = client.Get(c.metadataEndpoint + "/preempted")
	if err != nil {
		log.Errorf("Failed to fetch data from metadata service: %s", err)
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessful, prometheus.GaugeValue, 0, instanceId)
		return
	} else {
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessful, prometheus.GaugeValue, 1, instanceId)

		if preemptedResp.StatusCode == 404 {
			log.Debug("/preempted action endpoint not found")
			ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 0, "", instanceId)
			return
		} else {
			defer preemptedResp.Body.Close()
			body, _ := ioutil.ReadAll(preemptedResp.Body)
			preempted = string(body)

			if err != nil {
				log.Errorf("Couldn't parse /preempted metadata: %s", err)
				ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 0, instanceId, preempted)
			} else {
				log.Infof("instance endpoint available, will be preempted: %v", preempted)
				ch <- prometheus.MustNewConstMetric(c.terminationIndicator, prometheus.GaugeValue, 1, instanceId, preempted)
				uptime, _ := host.Uptime()
				log.Infof("instance was started at : %v" , uptime)
				ch <- prometheus.MustNewConstMetric(c.terminationTime, prometheus.GaugeValue, 1, instanceId, string(uptime))
			}
		}
	}
}
