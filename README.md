# ProPush
Collect host CPU, Memory, Hard disk and Network card traffic data in Prometheus data format and realize active statistical reporting.

The core code is located in the collector directory. You can learn how to use ProPush with cmd/main.go.



## How to work?

### 1. GetInstance

Get collector instance just like this:

```go
ins := collector.GetInstance()
```

### 2. GetMetrics

By GetMetrics():

```go
ins.GetMetrics()
```

You can get system metrics  like this:

```
# HELP propush_cpu_usage Usage of the cpus in 100ms.
# TYPE propush_cpu_usage counter
propush_cpu_usage{intervel="100ms"} 12.5
# HELP propush_disk_usage Filesystem usage.
# TYPE propush_disk_usage gauge
propush_disk_usage{device="/dev/vda1",fstype="ext3",mountpoint="/"} 30.810608473120354
# HELP propush_memory_usage Memory usage.
# TYPE propush_memory_usage gauge
propush_memory_usage 23.708612084527985
# HELP propush_network_netio Network I/O (MB).
# TYPE propush_network_netio gauge
propush_network_netio 104479.32768344879
# HELP propush_scrape_collector_duration_seconds node_exporter: Duration of a collector scrape.
# TYPE propush_scrape_collector_duration_seconds gauge
propush_scrape_collector_duration_seconds{collector="cpu"} 0.100620312
propush_scrape_collector_duration_seconds{collector="disk"} 0.000630578
propush_scrape_collector_duration_seconds{collector="mem"} 0.000238015
propush_scrape_collector_duration_seconds{collector="netio"} 0.000351507
# HELP propush_scrape_collector_success node_exporter: Whether a collector succeeded.
# TYPE propush_scrape_collector_success gauge
propush_scrape_collector_success{collector="cpu"} 1
propush_scrape_collector_success{collector="disk"} 1
propush_scrape_collector_success{collector="mem"} 1
propush_scrape_collector_success{collector="netio"} 1
```

### 3. PushMetrics

Push the metrics to [prometheus pushgateway](https://github.com/prometheus/pushgateway).

For high availability, you can call this function multiple times by traversing the pushgateway endpoints.

```go
ins.PushMetrics("http://127.0.0.1:9091", ins.GetMetrics())
```



## Advanced

You can combine the [ordered map](https://github.com/binacsgo/treemap) to implement the regular deletion strategy of expired metrics.

A more elegant and generic solution for this will be updated in the near future.

