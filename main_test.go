package main

import (
	"encoding/json"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"unsafe"
)

func GetUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func SetUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}

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

func TestExecute(t *testing.T) {
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

	//Need to pass a non-default exitStatus function to the handler to test handler Execute logic used in main().
	var exitStatus int
	exitStatus = 1
	mockExit := func(i int) {
		exitStatus = i
	}

	// exitFunction is unexported field, use helper function to set it to a new value
	handler := sensu.NewGoHandler(&config.PluginConfig, influxdbConfigOptions, checkArgs, sendMetrics)
	field := reflect.ValueOf(handler).Elem().FieldByName("exitFunction")
	SetUnexportedField(field, mockExit)
	assert.NotZero(exitStatus)
	assert.False(requestReceived)
	handler.Execute()
	//statements after Execute() would not be called unless non-default exitFunction is used due to how default os.Exit() operates.
	assert.Zero(exitStatus)
	assert.True(requestReceived)
}
