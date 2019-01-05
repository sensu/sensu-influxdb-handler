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
        "command": "sensu-influxdb-handler -d sensu",
        "timeout": 10,
        "env_vars": [
            "INFLUXDB_ADDR=http://influxdb.default.svc.cluster.local:8086",
            "INFLUXDB_USER=sensu",
            "INFLUXDB_PASS=password"
        ],
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

**Security Note:** The InfluxDB addr, username and password are treated as a security sensitive configuration options in this example and are loaded into the handler config as an env_vars instead of as a command arguments. Command arguments are commonaly readable from the process table by other unprivaledged users on a system (ex: `ps` and `top` commands), so it's a better practise to read in sensitive information via environment variables or configuration files as part of command execution. The command flags for these configuration options are are provided as an override for testing purposes.



## Usage Examples

Help:
```
Usage:
  sensu-influxdb-handler [flags]

Flags:
  -a, --addr string            the address of the influxdb server, should be of the form 'http://host:port', defaults to value of INFLUXDB_ADDR env variable
  -d, --db-name string         the influxdb to send metrics to
  -h, --help                   help for sensu-influxdb-handler
  -i, --insecure-skip-verify   if true, the influx client skips https certificate verification
  -p, --password string        the password for the given db, defaults to value of INFLUXDB_PASS env variable
      --precision string       the precision value of the metric (default "s")
  -u, --username string        the username for the given db, defaults to value of INFLUXDB_USER env variable

```

## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://github.com/influxdata/influxdb
[3]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/sensu/sensu-influxdb-handler/releases
[5]: https://blog.sensu.io/check-output-metric-extraction-with-influxdb-grafana
[6]: https://docs.sensu.io/sensu-go/5.0/guides/influx-db-metric-handler/
