# Sensu InfluxDB Handler
TravisCI: [![TravisCI Build Status](https://travis-ci.org/sensu/sensu-influxdb-handler.svg?branch=master)](https://travis-ci.org/sensu/sensu-influxdb-handler)

The Sensu InfluxDB Handler is a [Sensu Event Handler][3] that sends metrics to
the time series database [InfluxDB][2]. [Sensu][1] can collect metrics using
check output metric extraction or the StatsD listener. Those collected metrics
pass through the event pipeline, allowing Sensu to deliver the metrics to the
configured metric event handlers. This InfluxDB handler will allow you to
store, instrument, and visualize the metric data from Sensu.

Check out [The Sensu Blog][5] or [Sensu Docs][6] for a step by step guide!

## Installation

Download the latest version of the sensu-influxdb-handler from [releases][4],
or create an executable script from this source.

From the local path of the sensu-influxdb-handler repository:
```
go build -o /usr/local/bin/sensu-influxdb-handler main.go
```

## Configuration

Example Sensu Go handler definition:

```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "influxdb"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-influxdb-handler -a 'http://influxdb.default.svc.cluster.local:8086' -d sensu -u sensu -p password",
        "timeout": 10,
        "filters": [
            "has_metrics"
        ]
    }
}
```

Example Sensu Go check definition:

```json
{
    "api_version": "core/v2",
    "type": "CheckConfig",
    "metadata": {
        "namespace": "default",
        "name": "dummy-app-prometheus"
    },
    "spec": {
        "command": "sensu-prometheus-collector -exporter-url http://localhost:8080/metrics",
        "subscriptions":[
            "dummy"
        ],
        "publish": true,
        "interval": 10,
        "output_metric_format": "influxdb_line",
        "output_metric_handlers": [
            "influxdb"
        ]
    }
}
```

That's right, you can collect different types of metrics (ex. Influx,
Graphite, OpenTSDB, Nagios, etc.), Sensu will extract and transform
them, and this handler will populate them into your InfluxDB.

## Usage Examples

Help:
```
Usage:
  sensu-influxdb-handler [flags]

Flags:
  -a, --addr string       the address of the influxdb server, should be of the form 'http://host:port'
  -d, --db-name string    the influxdb to send metrics to
  -h, --help              help for sensu-influxdb-handler
  -p, --password string   the password for the given db
  -u, --username string   the username for the given db
```

## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://github.com/influxdata/influxdb
[3]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/sensu/sensu-influxdb-handler/releases
[5]: https://blog.sensu.io/check-output-metric-extraction-with-influxdb-grafana
[6]: https://docs.sensu.io/sensu-go/5.0/guides/influx-db-metric-handler/
