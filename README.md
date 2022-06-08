[![Bonsai Asset Badge](https://img.shields.io/badge/Sensu%20InfluxDB%20Handler-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-influxdb-handler)

# Sensu InfluxDBv2 Handler

- [Overview](#overview)
- [Usage examples](#usage-examples)
  - [Help output](#help-output)
  - [Environment variables](#environment-variables)
  - [Annotations](#annotations)
  - [Formatting options](#formatting-options)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Asset definition](#asset-definition)
  - [Handler definition](#handler-definition)
  - [Check definition](#check-definition)
- [InfluxDBv2 compatibility](#influxdbv2-compatibility)
- [Installation from source and contributing](#installation-from-source-and-contributing)

## Overview

The Sensu InfluxDB Handler is a [Sensu Event Handler][3] that sends metrics to
the time series database [InfluxDBv2][2]. [Sensu][1] can collect metrics using
check output metric extraction or the StatsD listener. Those collected metrics
pass through the event pipeline, allowing Sensu to deliver the metrics to the
configured metric event handlers. This InfluxDB handler will allow you to
store, instrument, and visualize the metric data from Sensu.

This handler also supports creating metrics out of check status results. This enables
operators to leverage InfluxDB as a long-term storage archive for Sensu check result
history. This feature will only work with the "-c" flag, and any check can add it as
handler.

Check out [The Sensu Blog][5] or [Sensu Docs][6] for a step by step guide!

## Usage Examples

### Help output
```
Usage:
  sensu-influxdb-handler [flags]

Flags:
  -a, --addr string          the url of the influxdb server, should be of the form 'http://host:port/dbname', defaults to 'http://localhost:8086' or value of INFLUXDB_ADDR env variable (default "http://localhost:8086")
  -b, --bucket string        the influxdbv2 bucket, use '<database>/<retention-policy>' as bucket for influxdb v1.8 compatibility
  -c, --checkStatusMetric    if true, the check status result will be captured as a metric
  -d, --dbName string        (Deprecated) influx v1.8 database to send metrics to. Transition to influxdb v1.8 compatible bucket name
  -h, --help                 help for sensu-influxdb-handler
  -i, --insecureSkipVerify   if true, the influx client skips https certificate verification
  -l, --legacy               (Deprecated) if true, parse the metric w/ legacy format
  -o, --org string           the influxdbv2 org, leave empty for influxdb v1.8 compatibility
  -p, --password string      (Deprecated) the password for the given db. Transition to influxdb v1.8  compatible authentication token
      --precision string     the precision value of the metric (default "s")
      --stripHost            if true, we strip the host from the metric
  -t, --token string         the authentication token needed for influxdbv2, use '<user>:<password>' as token for influxdb 1.8 compatibility
  -u, --username string      (Deprecated) the username for the given db, Transition to influxdb v1.8 compatible authentication token
```

### Environment variables

|Argument   |Environment Variable |
|-----------|---------------------|
|--addr     |INFLUXDB_ADDR        |
|--username |INFLUXDB_USER        |
|--password |INFLUXDB_PASS        |
|--bucket   |INFLUXDB_BUCKET      |
|--token    |INFLUXDB_TOKEN       |
|--org      |INFLUXDB_ORG         |

**Security Note:** Care should be taken to not expose the password for this handler by specifying it
on the command line or by directly setting the environment variable in the handler definition.  It is
suggested to make use of [secrets management][7] to surface it as an environment variable.  The
handler definition above references it as a secret.  Below is an example secrets definition that make
use of the built-in [env secrets provider][8].

### Annotations

All of the above command line arguments can be overridden by check or entity annotations.
The annotation consists of the key formed by appending the "long" argument specification
to the string sensu.io/plugins/sensu-influxdb-handler/config
(e.g. sensu.io/plugins/sensu-influxdb-handler/config/db-name).

For example, having the following in an agent.yml file will create an entity annotation
such that metrics generated by events on this entity will go to the `server-metrics` database instead
of the database defined in the handler.

```
namespace: "default"
subscriptions:
  - linux
backend-url:
  - "ws://127.0.0.1:8081"
annotations:
  sensu.io/plugins/sensu-influxdb-handler/config/db-name: "server-metrics"
```

### Formatting options

Default formatting: If the measurement is separated by `.` (period), takes the fist word as the measurement and subsequent word(s) as the field_set key. If there are no `.` then the measurement is taken as is and the field_set key is "value".

To change this default behavior there are 2 flags that can be used.

* `-l, --legacy-format` Keeps all elements of the measurement with no splitting and the key will always be "value". Also replaces the default tag key "sensu_entity_name" with "host".
* `--strip-host` Some metric checks put a hostname prefix to the measurement. This will strip it off for you without having to edit the check output. Used alone the default behavior will split the 2nd word element (if using `.` seperators) as the measurement.

These 2 flags can be used in concert, which would strip off the hostname but then keep the rest of the measurement as is.

## Configuration

### Asset registration

Assets are the best way to make use of this handler. If you're not using an asset, please consider doing so! If you're using sensuctl 5.13 or later, you can use the following command to add the asset:

`sensuctl asset add sensu/sensu-influxdb-handler`

If you're using an earlier version of sensuctl, you can download the asset definition from [this project's Bonsai Asset Index page](https://bonsai.sensu.io/assets/sensu/sensu-influxdb-handler).


### Asset definition

```yml
---
type: Asset
api_version: core/v2
metadata:
  name: sensu-influxdb-handler_linux_amd64
  labels:
  annotations:
    io.sensu.bonsai.url: https://bonsai.sensu.io/assets/sensu/sensu-influxdb-handler
    io.sensu.bonsai.api_url: https://bonsai.sensu.io/api/v1/assets/sensu/sensu-influxdb-handler
    io.sensu.bonsai.tier: Supported
    io.sensu.bonsai.version: 3.1.2
    io.sensu.bonsai.namespace: sensu
    io.sensu.bonsai.name: sensu-influxdb-handler
    io.sensu.bonsai.tags: ''
spec:
  url: https://assets.bonsai.sensu.io/b28f8719a48aa8ea80c603f97e402975a98cea47/sensu-influxdb-handler_3.1.2_linux_amd64.tar.gz
  sha512: 612c6ff9928841090c4d23bf20aaf7558e4eed8977a848cf9e2899bb13a13e7540bac2b63e324f39d9b1257bb479676bc155b24e21bf93c722b812b0f15cb3bd
  filters:
  - entity.system.os == 'linux'
  - entity.system.arch == 'amd64'
```

### Handler definition

```yml
---
api_version: core/v2
type: Handler
metadata:
  namespace: default
  name: influxdb
spec:
  type: pipe
  command: sensu-influxdb-handler -d sensu
  timeout: 10
  env_vars:
  - INFLUXDB_ADDR=http://influxdb.default.svc.cluster.local:8086
  - INFLUXDB_USER=sensu
  - INFLUXDB_PASS=password
  filters:
  - has_metrics
  runtime_assets:
  - sensu/sensu-influxdb-handler
```

### Check definition
```yml
---
api_version: core/v2
type: CheckConfig
metadata:
  namespace: default
  name: dummy-app-prometheus
spec:
  command: sensu-prometheus-collector -exporter-url http://localhost:8080/metrics
  subscriptions:
  - dummy
  publish: true
  interval: 10
  output_metric_format: influxdb_line
  output_metric_handlers:
  - influxdb
```

That's right, you can collect different types of metrics (ex. Influx,
Graphite, OpenTSDB, Nagios, etc.), Sensu will extract and transform
them, and this handler will populate them into your InfluxDB.

**Security Note:** The InfluxDB addr, username and password are treated as a security sensitive configuration options in this example and are loaded into the handler config as env_vars instead of as a command arguments. Command arguments are commonly readable from the process table by other unprivileged users on a system (ex: `ps` and `top` commands), so it's a better practice to read in sensitive information via environment variables or configuration files as part of command execution. The command flags for these configuration options are provided as an override for testing purposes.

## InfluxDBv1.8 Compatibility
It's possible to use this plugin with InfluxDB v1.8 by specifying the approporate compatible bucket and token values.
ref: https://github.com/influxdata/influxdb-client-go#influxdb-18-api-compatibility

* Use the form 'username:password' as the token value
* Use the form 'database/retention-policy' as the bucket value

## Installation from source and contributing

Download the latest version of the sensu-influxdb-handler from [releases][4],
or create an executable from this source.

### Compiling

From the local path of the sensu-influxdb-handler repository:
```
go build -o /usr/local/bin/sensu-influxdb-handler main.go
```

To contribute to this plugin, see [CONTRIBUTING](https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md)

[1]: https://github.com/sensu/sensu-go
[2]: https://github.com/influxdata/influxdb
[3]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/sensu/sensu-influxdb-handler/releases
[5]: https://blog.sensu.io/check-output-metric-extraction-with-influxdb-grafana
[6]: https://docs.sensu.io/sensu-go/5.0/guides/influx-db-metric-handler/
[7]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/
[8]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/#use-env-for-secrets-management
