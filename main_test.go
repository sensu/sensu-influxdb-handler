package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusMetrics(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	event.Metrics = nil

	var apiStub = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		expectedBody := `check1,sensu_entity_name=entity1 status=0`
		assert.Contains(string(body), expectedBody)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"ok": true}`))
		require.NoError(t, err)
	}))

	config.Addr = apiStub.URL
	config.CheckStatusMetric = true
	err := sendMetrics(event)
	assert.NoError(err)
}

func TestSendMetrics(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	event.Check = nil
	event.Metrics = corev2.FixtureMetrics()

	var apiStub = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		expectedBody := `answer,foo=bar,sensu_entity_name=entity1 value=42`
		assert.Contains(string(body), expectedBody)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"ok": true}`))
		require.NoError(t, err)
	}))

	config.Addr = apiStub.URL
	err := sendMetrics(event)
	assert.NoError(err)
}

func TestSendMetricsHostStripping(t *testing.T) {
	var hostStrippingDataSet = []struct {
		entityName string
		expectedBody string
		metricName string
		stripHost    bool
	}{
		{
			entityName: "sensu.company.com",
			metricName: "sensu.company.com",
			expectedBody: `answer,foo=bar,sensu_entity_name=sensu.company.com value=42`,
			stripHost: true,
		},
		{
			entityName: "answer.company.com",
			metricName: "answer.company.com",
			expectedBody: `answer,foo=bar,sensu_entity_name=answer.company.com value=42`,
			stripHost: true,
		},
		{
			entityName: "prod-eu-db01.company.com",
			metricName: "prod-eu-db01.company.com",
			expectedBody: `answer,foo=bar,sensu_entity_name=prod-eu-db01.company.com value=42`,
			stripHost: true,
		},
		{
			entityName: "",
			metricName: "",
			expectedBody: `answer,foo=bar value=42`,
			stripHost: true,
		},
		{
			entityName: "metric01",
			metricName: "",
			expectedBody: `foo=bar,sensu_entity_name=metric01 answer=42`,
			stripHost: true,
		},
		{
			entityName: "prod-eu-db01.company.com",
			metricName: "different.company.com",
			expectedBody: `different,foo=bar,sensu_entity_name=prod-eu-db01.company.com company.com.answer=42`,
			stripHost: true,
		},
		{
			entityName: "no-stripping.company.com",
			metricName: "no-stripping.company.com",
			expectedBody: `no-stripping,foo=bar,sensu_entity_name=no-stripping.company.com company.com.answer=42`,
			stripHost: false,
		},
		{
			entityName: "company.com",
			metricName: "company.com",
			expectedBody: `company,foo=bar,sensu_entity_name=company.com com.answer=42`,
			stripHost: false,
		},
	}

	for _, tt := range hostStrippingDataSet {
		assert := assert.New(t)
		event := corev2.FixtureEvent(tt.entityName, "check1")
		event.Check = nil
		event.Metrics = corev2.FixtureMetrics()
		event.Metrics.Points[0].Name = tt.metricName + ".answer"

		var apiStub = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			assert.Contains(string(body), tt.expectedBody)
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"ok": true}`))
			require.NoError(t, err)
		}))

		config.Addr = apiStub.URL
		config.StripHost = tt.stripHost
		err := sendMetrics(event)
		assert.NoError(err)
	}
}

func TestSendAnnotation(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Status = 1
	event.Check.Occurrences = 1
	event.Check.Output = "FAILURE"
	event.Metrics = corev2.FixtureMetrics()

	var apiStub = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		expectedBody := `sensu_event,check=check1,entity=entity1 description="\"ALERT - entity1/check1 : FAILURE\"",occurrences=1i,status=1i,title="\"Sensu Event\""`
		assert.Contains(string(body), expectedBody)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"ok": true}`))
		require.NoError(t, err)
	}))

	config.Addr = apiStub.URL
	err := sendMetrics(event)
	assert.NoError(err)
}

func TestEventNeedsAnnotation(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")

	b := eventNeedsAnnotation(event)
	assert.True(b)

	event.Check.Occurrences = 2
	b = eventNeedsAnnotation(event)
	assert.False(b)

	event.Check.Status = 1
	b = eventNeedsAnnotation(event)
	assert.True(b)

	event.Check = nil
	b = eventNeedsAnnotation(event)
	assert.False(b)
}

func TestMain(t *testing.T) {
	assert := assert.New(t)
	file, _ := ioutil.TempFile(os.TempDir(), "sensu-handler-influx-db-")
	defer func() {
		_ = os.Remove(file.Name())
	}()

	event := corev2.FixtureEvent("entity1", "check1")
	event.Check = nil
	event.Metrics = corev2.FixtureMetrics()
	eventJSON, _ := json.Marshal(event)
	_, err := file.WriteString(string(eventJSON))
	require.NoError(t, err)
	require.NoError(t, file.Sync())
	_, err = file.Seek(0, 0)
	require.NoError(t, err)
	os.Stdin = file
	requestReceived := false

	var apiStub = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"ok": true}`))
		require.NoError(t, err)
	}))

	oldArgs := os.Args
	os.Args = []string{"influx-db", "-a", apiStub.URL, "-c", "-d", "foo", "-u", "bar", "-p", "baz"}
	defer func() { os.Args = oldArgs }()

	main()
	assert.True(requestReceived)
}
