# PouchContainer with Prometheus

PouchContainer supports various monitoring metrics via [Prometheus](https://prometheus.io/). Now we already have the basic golang runtime and some api latency metrics. We plan to add more in the future in below two main areas:

* Important pouchd metrics
* Full list of important api duration metrics

## How to add new metrics

We tend to use prometheus's [METRIC AND LABEL NAMING](https://prometheus.io/docs/practices/naming) best-practices in PouchContainer. So when you are going to add a new metric, do follow the metric and label naming convention.

We use prometheus [go-sdk](https://github.com/prometheus/client_golang) to monitor pouchd. It supports counter, gauge and summary metric types. For more info, please refer to [METRIC TYPES](https://prometheus.io/docs/concepts/metric_types/).

## How to use

Users can start pouchd listening on `0.0.0.0:4243` via `pouchd -l tcp://0.0.0.0:4243`, then issue `GET http://127.0.0.1:4243/metrics`  request to get a full list of prometheus-formatted metrics as below:

```
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.000111176
go_gc_duration_seconds{quantile="0.25"} 0.000198062
go_gc_duration_seconds{quantile="0.5"} 0.000269599
go_gc_duration_seconds{quantile="0.75"} 0.000474291
go_gc_duration_seconds{quantile="1"} 0.002013351
go_gc_duration_seconds_sum 0.021835193
go_gc_duration_seconds_count 52
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 22
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.9"} 1
...
# HELP http_request_size_bytes The HTTP request sizes in bytes.
# TYPE http_request_size_bytes summary
http_request_size_bytes{handler="prometheus",quantile="0.5"} NaN
http_request_size_bytes{handler="prometheus",quantile="0.9"} NaN
http_request_size_bytes{handler="prometheus",quantile="0.99"} NaN
http_request_size_bytes_sum{handler="prometheus"} 0
http_request_size_bytes_count{handler="prometheus"} 0
# HELP http_response_size_bytes The HTTP response sizes in bytes.
# TYPE http_response_size_bytes summary
http_response_size_bytes{handler="prometheus",quantile="0.5"} NaN
http_response_size_bytes{handler="prometheus",quantile="0.9"} NaN
http_response_size_bytes{handler="prometheus",quantile="0.99"} NaN
http_response_size_bytes_sum{handler="prometheus"} 0
http_response_size_bytes_count{handler="prometheus"} 0
# HELP pouch_image_pull_latency_microseconds Latency in microseconds to pull a image.
# TYPE pouch_image_pull_latency_microseconds summary
pouch_image_pull_latency_microseconds{image="docker.io/library/ubuntu:latest",quantile="0.5"} 3.7803132e+07
pouch_image_pull_latency_microseconds{image="docker.io/library/ubuntu:latest",quantile="0.9"} 3.7803132e+07
pouch_image_pull_latency_microseconds{image="docker.io/library/ubuntu:latest",quantile="0.99"} 3.7803132e+07
pouch_image_pull_latency_microseconds_sum{image="docker.io/library/ubuntu:latest"} 3.7803132e+07
pouch_image_pull_latency_microseconds_count{image="docker.io/library/ubuntu:latest"} 1
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 4.78
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 1024
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 9
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 3.4521088e+07
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.51064406778e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 4.91610112e+08
```

Then we can set up a new target to scrape this metric endpoint in prometheus. So that's it.
