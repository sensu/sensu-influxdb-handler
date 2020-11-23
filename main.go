package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// HandlerConfig for runtime values
type HandlerConfig struct {
	sensu.PluginConfig
	Addr               string
	Username           string
	Password           string
	DbName             string
	Precision          string
	InsecureSkipVerify bool
	CheckStatusMetric  bool
	StripHost          bool
	Legacy             bool
}

const (
	addr               = "addr"
	username           = "username"
	password           = "password"
	dbName             = "db-name"
	precision          = "precision"
	insecureSkipVerify = "insecure-skip-verify"
	checkStatusMetric  = "check-status-metric"
	stripHost          = "strip-host"
	legacy             = "legacy-format"
)

var (
	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-influxdb-handler",
			Short:    "an influxdb handler built for use with sensu",
			Keyspace: "sensu.io/plugins/sensu-influxdb-handler/config",
		},
	}

	influxdbConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      addr,
			Env:       "INFLUXDB_ADDR",
			Argument:  addr,
			Shorthand: "a",
			Default:   "http://localhost:8086",
			Usage:     "the address of the influxdb server, should be of the form 'http://host:port', defaults to 'http://localhost:8086' or value of INFLUXDB_ADDR env variable",
			Value:     &config.Addr,
		},
		{
			Path:      username,
			Env:       "INFLUXDB_USER",
			Argument:  username,
			Shorthand: "u",
			Default:   "",
			Usage:     "the username for the given db, defaults to value of INFLUXDB_USER env variable",
			Value:     &config.Username,
		},
		{
			Path:      password,
			Env:       "INFLUXDB_PASS",
			Argument:  password,
			Shorthand: "p",
			Default:   "",
			Secret:    true,
			Usage:     "the password for the given db, defaults to value of INFLUXDB_PASS env variable",
			Value:     &config.Password,
		},
		{
			Path:      dbName,
			Argument:  dbName,
			Shorthand: "d",
			Default:   "",
			Usage:     "the influxdb to send metrics to",
			Value:     &config.DbName,
		},
		{
			Path:      precision,
			Argument:  precision,
			Shorthand: "",
			Default:   "s",
			Usage:     "the precision value of the metric",
			Value:     &config.Precision,
		},
		{
			Path:      insecureSkipVerify,
			Argument:  insecureSkipVerify,
			Shorthand: "i",
			Default:   false,
			Usage:     "if true, the influx client skips https certificate verification",
			Value:     &config.InsecureSkipVerify,
		},
		{
			Path:      checkStatusMetric,
			Argument:  checkStatusMetric,
			Shorthand: "c",
			Default:   false,
			Usage:     "if true, the check status result will be captured as a metric",
			Value:     &config.CheckStatusMetric,
		},
		{
			Path:      stripHost,
			Argument:  stripHost,
			Shorthand: "",
			Default:   false,
			Usage:     "if true, we strip the host from the metric",
			Value:     &config.StripHost,
		},
		{
			Path:      legacy,
			Argument:  legacy,
			Shorthand: "l",
			Default:   false,
			Usage:     "if true, parse the metric w/ legacy format",
			Value:     &config.Legacy,
		},
	}
)

func main() {
	goHandler := sensu.NewGoHandler(&config.PluginConfig, influxdbConfigOptions, checkArgs, sendMetrics)
	goHandler.Execute()
}

func checkArgs(event *corev2.Event) error {
	if len(config.DbName) == 0 {
		return errors.New("missing db name")
	}
	if !event.HasMetrics() && !config.CheckStatusMetric {
		return fmt.Errorf("event does not contain metrics")
	}
	return nil
}

func sendMetrics(event *corev2.Event) error {
	var pt *client.Point
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:               config.Addr,
		Username:           config.Username,
		Password:           config.Password,
		InsecureSkipVerify: config.InsecureSkipVerify,
	})
	if err != nil {
		return err
	}
	defer c.Close()

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.DbName,
		Precision: config.Precision,
	})
	if err != nil {
		return err
	}

	// Add the check status field as a metric if requested. Measurement recorded as the check name.
	if config.CheckStatusMetric && event.HasCheck() {
		var statusMetric = &corev2.MetricPoint{
			Name:      event.Check.Name + ".status",
			Value:     float64(event.Check.Status),
			Timestamp: event.Timestamp,
		}
		// bootstrap the event for metrics
		if !event.HasMetrics() {
			event.Metrics = new(corev2.Metrics)
			event.Metrics.Points = make([]*corev2.MetricPoint, 0)
		}
		event.Metrics.Points = append(event.Metrics.Points, statusMetric)
	}

	for _, point := range event.Metrics.Points {

		if config.StripHost && strings.HasPrefix(point.Name, event.Entity.Name) {
			// Adding a char since we also want to strip the dot
			point.Name = point.Name[len(event.Entity.Name)+1:]
		}

		name := setName(point.Name)

		fields := setFields(point.Name, point.Value)

		timestamp, err := setTime(point.Timestamp)
		if err != nil {
			return err
		}

		tags := setTags(event.Entity.Name, point.Tags)

		pt, err = client.NewPoint(name, tags, fields, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}

	// 1.x handler parity
	annotate := eventNeedsAnnotation(event)

	if annotate {
		tags := make(map[string]string)
		tags["entity"] = event.Entity.Name
		tags["check"] = event.Check.Name

		title := fmt.Sprintf("%q", "Sensu Event")
		description := fmt.Sprintf("%q", sensu.FormattedMessage(event))
		fields := make(map[string]interface{})
		fields["title"] = title
		fields["description"] = description
		fields["status"] = event.Check.Status
		fields["occurrences"] = event.Check.Occurrences

		stringTimestamp := strconv.FormatInt(event.Timestamp, 10)
		if len(stringTimestamp) > 10 {
			stringTimestamp = stringTimestamp[:10]
		}
		t, err := strconv.ParseInt(stringTimestamp, 10, 64)
		if err != nil {
			return err
		}
		timestamp := time.Unix(t, 0)

		pt, err = client.NewPoint("sensu_event", tags, fields, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}

	if err = c.Write(bp); err != nil {
		return err
	}

	return c.Close()
}

// Determine if an event needs an annotation
func eventNeedsAnnotation(event *corev2.Event) bool {
	// No check, no need to be here
	if !event.HasCheck() {
		return false
	}

	// Alert (should this only happen on occurrence == 1?)
	if event.Check.Status != 0 {
		return true
	}

	// Status 0, steady as she goes, not an alert
	if event.Check.Occurrences > 1 {
		return false
	}

	// Status 0, but first occurrence so it's a resolution, assumed
	return true
}

// set tagkey name
func setFields(name string, value interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	//Legacy always uses value as the key
	if config.Legacy {
		fields["value"] = value
		return fields
	}

	nameField := strings.Split(name, ".")
	// names with '.', use first part as measurement name and rest as key for the value
	if len(nameField) > 1 {
		fields[strings.Join(nameField[1:], ".")] = value
		return fields
	}

	fields["value"] = value
	return fields
}

func setTags(name string, tags []*corev2.MetricTag) map[string]string {
	ntags := make(map[string]string)

	if config.Legacy {
		ntags["host"] = name
	} else {
		ntags["sensu_entity_name"] = name
	}

	for _, tag := range tags {
		ntags[tag.Name] = tag.Value
	}

	return ntags
}

func setTime(timestamp int64) (time.Time, error) {
	stringTimestamp := strconv.FormatInt(timestamp, 10)
	if len(stringTimestamp) > 10 {
		stringTimestamp = stringTimestamp[:10]
	}
	t, err := strconv.ParseInt(stringTimestamp, 10, 64)
	if err != nil {
		return time.Now(), err
	}

	return time.Unix(t, 0), nil
}

// set mesurement name
func setName(name string) string {
	//Legacy always returns full name
	if config.Legacy {
		return name
	}

	// if name includes '.' then only the first one is used
	return strings.Split(name, ".")[0]
}
