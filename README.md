# Sensu InfluxDB Handler

The Sensu InfluxDB Handler is a [Sensu Event Handler][3] that sends metrics to
the time series database [InfluxDB][2]. [Sensu][1] can collect metrics using
check output metric extraction or the StatsD listener. Those collected metrics
pass through the event pipeline, allowing Sensu to deliver the metrics to the
configured metric event handlers. This InfluxDB handler will allow you to
store, instrument, and visualize the metric data from Sensu.

## Installation

Download the latest version of the sensu-influxdb-handler from [releases][4],
or create an executable script from this source.

From the local path of the sensu-influxdb-handler repository:
```
go build -o /usr/local/bin/sensu-influxdb-handler main.go
```

## Configuration

Example Sensu 2.x handler definition:
```
{
  "name": "influx-db",
  "type": "pipe",
  "command": "sensu-influxdb-handler --addr 'http://123.4.5.6:8086' --username 'foo' --password 'bar' --db-name 'myDB'"
}
```

Example Sensu 2.x check definition:
```
{
  "name": "collect-metrics",
  "command": "collect.sh",
  "interval": 10,
  "subscriptions": [
    "system"
  ],
  "output_metric_format": "graphite_plaintext",
  "output_metric_handlers": [influx-db]
}
```
That's right, you can collect different types of metrics (ex. Graphite), Sensu
will extract and transform them, and this handler will populate them into your
InfluxDB.

## Usage Examples

Help:
```
Usage:
  sensu-influxdb-handler [flags]

Flags:
  -a, --addr string       the address of the influx-db server
  -d, --db-name string    the influx-db to send metrics to
  -h, --help              help for handler-influx-db
  -p, --password string   the password for the given db
  -u, --username string   the username for the given db
```

[1]: https://github.com/sensu/sensu-go
[2]: https://github.com/influxdata/influxdb
[3]: https://docs.sensu.io/sensu-core/2.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/nikkiki/sensu-influxdb-handler/releases
