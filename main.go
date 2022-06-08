package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// HandlerConfig for runtime values
type HandlerConfig struct {
	sensu.PluginConfig
	Addr               string
	Token              string
	Bucket             string
	Org                string
	Username           string
	Password           string
	DbName             string
	Precision          string
	InsecureSkipVerify bool
	CheckStatusMetric  bool
	StripHost          bool
	Legacy             bool
}

var (
	precisionMap = map[string]time.Duration{
		"ns": time.Nanosecond,
		"us": time.Microsecond,
		"ms": time.Millisecond,
		"s":  time.Second,
	}

	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-influxdb-handler",
			Short:    "an influxdb handler built for use with sensu",
			Keyspace: "sensu.io/plugins/sensu-influxdb-handler/config",
		},
	}

	influxdbConfigOptions = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "addr",
			Env:       "INFLUXDB_ADDR",
			Argument:  "addr",
			Shorthand: "a",
			Default:   "http://localhost:8086",
			Usage:     "the url of the influxdb server, should be of the form 'http://host:port/dbname', defaults to 'http://localhost:8086' or value of INFLUXDB_ADDR env variable",
			Value:     &config.Addr,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "token",
			Env:       "INFLUXDB_TOKEN",
			Argument:  "token",
			Shorthand: "t",
			Default:   "",
			Usage:     "the authentication token needed for influxdbv2, use '<user>:<password>' as token for influxdb 1.8 compatibility",
			Value:     &config.Token,
			Secret:    true,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "bucket",
			Env:       "INFLUXDB_BUCKET",
			Argument:  "bucket",
			Shorthand: "b",
			Default:   "",
			Usage:     "the influxdbv2 bucket, use '<database>/<retention-policy>' as bucket for influxdb v1.8 compatibility",
			Value:     &config.Bucket,
			Secret:    true,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "org",
			Env:       "INFLUXDB_ORG",
			Argument:  "org",
			Shorthand: "o",
			Default:   "",
			Usage:     "the influxdbv2 org, leave empty for influxdb v1.8 compatibility",
			Value:     &config.Org,
			Secret:    true,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "username",
			Env:       "INFLUXDB_USER",
			Argument:  "username",
			Shorthand: "u",
			Default:   "",
			Usage:     "(Deprecated) the username for the given db, Transition to influxdb v1.8 compatible authentication token",
			Value:     &config.Username,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "password",
			Env:       "INFLUXDB_PASS",
			Argument:  "password",
			Shorthand: "p",
			Default:   "",
			Secret:    true,
			Usage:     "(Deprecated) the password for the given db. Transition to influxdb v1.8  compatible authentication token",
			Value:     &config.Password,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "dbName",
			Argument:  "dbName",
			Shorthand: "d",
			Default:   "",
			Usage:     "(Deprecated) influx v1.8 database to send metrics to. Transition to influxdb v1.8 compatible bucket name",
			Value:     &config.DbName,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "precision",
			Argument:  "precision",
			Shorthand: "",
			Default:   "s",
			Usage:     "the precision value of the metric",
			Value:     &config.Precision,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "insecureSkipVerify",
			Argument:  "insecureSkipVerify",
			Shorthand: "i",
			Default:   false,
			Usage:     "if true, the influx client skips https certificate verification",
			Value:     &config.InsecureSkipVerify,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "checkStatusMetric",
			Argument:  "checkStatusMetric",
			Shorthand: "c",
			Default:   false,
			Usage:     "if true, the check status result will be captured as a metric",
			Value:     &config.CheckStatusMetric,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "stripHost",
			Argument:  "stripHost",
			Shorthand: "",
			Default:   false,
			Usage:     "if true, we strip the host from the metric",
			Value:     &config.StripHost,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "legacy",
			Argument:  "legacy",
			Shorthand: "l",
			Default:   false,
			Usage:     "(Deprecated) if true, parse the metric w/ legacy format",
			Value:     &config.Legacy,
		},
	}
)

func main() {
	useStdin, err := testStdin()
	if err != nil {
		panic(err)
	}
	if useStdin {
		goHandler := sensu.NewHandler(&config.PluginConfig, influxdbConfigOptions, checkArgs, sendMetrics)
		goHandler.Execute()
	} else {
		panic(fmt.Errorf("Must supply Sensu event json on stdin\n"))
	}
}

func testStdin() (bool, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Printf("Error accessing stdin: %v\n", err)
		return false, err
	}
	//Check the Mode bitmask for Named Pipe to indicate stdin is connected
	if fi.Mode()&os.ModeNamedPipe != 0 {
		return true, nil
	}
	return false, nil
}

func checkArgs(event *corev2.Event) error {

	if _, ok := precisionMap[config.Precision]; !ok {
		keys := []string{}
		for key, _ := range precisionMap {
			keys = append(keys, key)
		}

		return fmt.Errorf("--precision must be one of: %v\n", keys)
	}
	if len(config.Addr) == 0 {
		return errors.New("--address must be provided\n")
	}
	if len(config.Bucket) > 0 && len(config.DbName) > 0 {
		return errors.New("Cannot set both --bucket and --dbName\n")
	}
	if len(config.Bucket) == 0 && len(config.DbName) == 0 {
		return errors.New("Must specify either --bucket or --dbName\n")
	}

	if len(config.Bucket) == 0 {
		if len(config.DbName) > 0 {
			config.Bucket = config.DbName
		}
	}
	if len(config.Token) == 0 {
		token := ""
		if len(config.Username) > 0 {
			token = token + string(config.Username) + ":"
		}
		if len(config.Password) > 0 {
			token = token + string(config.Password)
		}
		config.Token = token
	}
	if !event.HasMetrics() && !config.CheckStatusMetric {
		return fmt.Errorf("event does not contain metrics")
	}
	return nil
}

func sendMetrics(event *corev2.Event) error {
	return nil
	var writeErrors []error
	c := influxdb2.NewClientWithOptions(
		config.Addr,
		config.Token,
		influxdb2.DefaultOptions().
			SetPrecision(precisionMap[config.Precision]).
			SetTLSConfig(&tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			}))
	defer c.Close()
	// Get non-blocking write client
	writeAPI := c.WriteAPI(config.Org, config.Bucket)
	defer writeAPI.Flush()
	// Get errors channel
	errorsCh := writeAPI.Errors()
	// Create go proc for reading and logging errors
	go func() {
		for err := range errorsCh {
			writeErrors = append(writeErrors, err)
		}
	}()
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

		pt := influxdb2.NewPoint(name, tags, fields, timestamp)
		writeAPI.WritePoint(pt)
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

		pt := influxdb2.NewPoint("sensu_event", tags, fields, timestamp)
		writeAPI.WritePoint(pt)
	}
	//writeAPI.Flush()
	//c.Close()
	return nil
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
