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
}

const (
	addr               = "addr"
	username           = "username"
	password           = "password"
	dbName             = "db-name"
	precision          = "precision"
	insecureSkipVerify = "insecure-skip-verify"
	checkStatusMetric  = "check-status-metric"
)

var (
	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:  "sensu-influxdb-handler",
			Short: "an influxdb handler built for use with sensu",
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
		var tagKey string
		nameField := strings.Split(point.Name, ".")
		name := nameField[0]
		if len(nameField) > 1 {
			tagKey = strings.Join(nameField[1:], ".")
		} else {
			tagKey = "value"
		}
		fields := map[string]interface{}{tagKey: point.Value}
		stringTimestamp := strconv.FormatInt(point.Timestamp, 10)
		if len(stringTimestamp) > 10 {
			stringTimestamp = stringTimestamp[:10]
		}
		t, err := strconv.ParseInt(stringTimestamp, 10, 64)
		if err != nil {
			return err
		}
		timestamp := time.Unix(t, 0)
		tags := make(map[string]string)
		tags["sensu_entity_name"] = event.Entity.Name
		for _, tag := range point.Tags {
			tags[tag.Name] = tag.Value
		}

		pt, err = client.NewPoint(name, tags, fields, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}

	annotate, action := event_needs_annotation(event)

	if annotate {
		tags := make(map[string]string)
		tags["entity"] = event.Entity.Name
		tags["check"] = event.Check.Name
		tags["action"] = action
		
		title := fmt.Sprintf("%q", "Sensu Event")
		description := fmt.Sprintf("%q", formatted_message(event, action))
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

func event_needs_annotation(event *corev2.Event) (bool, string) {
	// No check, no need to be here
	if !event.HasCheck() {
		return false, ""
	}

	// Alert
	if event.Check.Status != 0 {
		return true, "create"
	}

	// Status 0, steady as she goes, not an alert
	if event.Check.Occurrences > 1 {
		return false, ""
	}

	// Status 0, but first occurrence so it's a resolution, assumed
	return true, "resolve"
}

func formatted_message(event *corev2.Event, action string) string {
	var formatted_action = "ALERT"
	if action == "resolve" {
		formatted_action = "RESOLVED"
	}
	summary := strings.TrimSuffix(event.Check.Output, "\n")
	if len(summary) > 100 {
		summary = summary[:100]
	}
	return fmt.Sprintf("%s - %s/%s:%s", formatted_action, event.Entity.Name, event.Check.Name, summary)
}
