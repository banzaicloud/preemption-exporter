## Preemptible instance termination exporter

Prometheus [exporters](https://prometheus.io/docs/instrumenting/writing_exporters) are used to export metrics from third-party systems as Prometheus metrics - this is an exporter to scrape for GCP spot price termination notice on the instance for [Hollowtrees](https://github.com/banzaicloud/hollowtrees).

### Preemptible instance lifecycle

Preemptible VMs are Google Compute Engine VM instances that last a maximum of 24 hours and provide no availability guarantees. Preemptible VMs are priced lower than standard Compute Engine VMs and offer the same machine types and options.

>Note: reemptible instances require available CPU quotas like regular instances. To avoid preemptible instances consuming the CPU quotas for your regular instances, you can request a special `Preemptible CPU` quota. After Compute Engine grants you preemptible CPU quota in that region, all preemptible instances will count against that quota. All regular instances will continue to count against the regular CPU quota.

### Preemptible instance termination notice

It can be determined if an instance was preempted from inside the instance itself. This is useful if we want to handle a shutdown due to a Compute Engine preeemption differently from a normal shutdown in a shutdown script. To do this, simply check the metadata server for the `preempted` value in your instance's default instance metadata.

For example, use curl from within your instance to obtain the value for preempted:
```
curl "http://metadata.google.internal/computeMetadata/v1/instance/preempted" -H "Metadata-Flavor: Google"
TRUE
```

If this value is TRUE, the instance was preempted by Compute Engine, otherwise it will be FALSE.


### Quick start

The project uses the [promu](https://github.com/prometheus/promu) Prometheus utility tool. To build the exporter `promu` needs to be installed. To install promu and build the exporter:

```
go get github.com/prometheus/promu
promu build
```

The following options can be configured when starting the exporter:

```
./preemption-exporter --help
Usage of ./preemption-exporte:
  -bind-addr string
        bind address for the metrics server (default ":9189")
  -log-level string
        log level (default "info")
  -metadata-endpoint string
        metadata endpoint to query (default "http://169.254.169.254/latest/meta-data/")
  -metrics-path string
        path to metrics endpoint (default "/metrics")

```

### Test locally

The GCP instance metadata is available at `http://metadata.google.internal/computeMetadata/v1/instance`. By default this is the endpoint that is being queried by the exporter but it is quite hard to reproduce a termination notice on a GCP instance for testing, so the meta-data endpoint can be changed in the configuration.
There is a test server in the `utils` directory that can be used to mock the behavior of the metadata endpoint. It listens on port 9092 and provides dummy responses for `/id` and `/scheduling/preempted`. It can be started with:
```
go run util/test_server.go
```
The exporter can be started with this configuration to query this endpoint locally:
```
./preemption-exporter --metadata-endpoint http://localhost:9092/computeMetadata/v1/instance/ --log-level debug
```



### Default Hollowtrees node exporters associated to alerts:

* AWS spot instance termination [exporter](https://github.com/banzaicloud/spot-termination-exporter)
* AWS autoscaling group [exporter](https://github.com/banzaicloud/aws-autoscaling-exporter)
