package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

var (
	addr     string
	dbName   string
	username string
	password string
	stdin    *os.File
)

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensu-influxdb-handler",
		Short: "an influxdb handler built for use with sensu",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&addr,
		"addr",
		"a",
		"",
		"the address of the influxdb server, should be of the form 'http://host:port'")

	cmd.Flags().StringVarP(&dbName,
		"db-name",
		"d",
		"",
		"the influxdb to send metrics to")

	cmd.Flags().StringVarP(&username,
		"username",
		"u",
		"",
		"the username for the given db")

	cmd.Flags().StringVarP(&password,
		"password",
		"p",
		"",
		"the password for the given db")

	_ = cmd.MarkFlagRequired("addr")
	_ = cmd.MarkFlagRequired("db-name")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return fmt.Errorf("invalid argument(s) received")
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err)
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", err)
	}

	if err = event.Validate(); err != nil {
		return fmt.Errorf("failed to validate event: %s", err)
	}

	if !event.HasMetrics() {
		return fmt.Errorf("event does not contain metrics")
	}

	return sendMetrics(event)
}

func sendMetrics(event *types.Event) error {
	var pt *client.Point
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}
	defer c.Close()

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbName,
		Precision: "s",
	})
	if err != nil {
		return err
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

	if err = c.Write(bp); err != nil {
		return err
	}

	return c.Close()
}
