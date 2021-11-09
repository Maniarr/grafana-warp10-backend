package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
	"time"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/mitchellh/mapstructure"
)

var (
	_ backend.QueryDataHandler   = (*Warp10Datasource)(nil)
	_ backend.CheckHealthHandler = (*Warp10Datasource)(nil)
)

func NewWarp10Datasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &Warp10Datasource{}, nil
}

type Warp10Datasource struct {
}

func (d *Warp10Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Debug("QueryData called", "request", req)

	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		response.Responses[q.RefID] = res
	}

	return response, nil
}

type QueryModel struct {
	QueryText string `json:"queryText"`
}

type ConfigDatasource struct {
	Path string `json:"path"`
}

type Warp10Response struct {
	Name   string            `json:"c" mapstructure:"c"`
	Labels map[string]string `json:"l" mapstructure:"l"`
	Values [][2]float64      `json:"v" mapstructure:"v"`
}

func UnmarshalWarp10Response(raw_response []interface{}) []Warp10Response {
	warp10_responses := make([]Warp10Response, 0)

	for _, raw_value := range raw_response {

		switch raw_value.(type) {
		case []interface{}:
			warp10_responses = append(warp10_responses, UnmarshalWarp10Response(raw_value.([]interface{}))...)
		case interface{}:
			var response Warp10Response

			err := mapstructure.Decode(raw_value.(map[string]interface{}), &response)

			if err != nil {
				log.DefaultLogger.Error("UnmarshalWarp10Response", "cast error", err)
			}

			warp10_responses = append(warp10_responses, response)
		}
	}

	return warp10_responses
}

func (d *Warp10Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	var configDatasource ConfigDatasource

	response.Error = json.Unmarshal(pCtx.DataSourceInstanceSettings.JSONData, &configDatasource)

	if response.Error != nil {
		return response
	}

	var qm QueryModel

	response.Error = json.Unmarshal(query.JSON, &qm)

	if response.Error != nil {
		return response
	}

	if strings.Contains(qm.QueryText, "$fromISO") {
		qm.QueryText = strings.ReplaceAll(qm.QueryText, "$fromISO", "'"+query.TimeRange.From.Format(time.RFC3339)+"'")
	}

	if strings.Contains(qm.QueryText, "$toISO") {
		qm.QueryText = strings.ReplaceAll(qm.QueryText, "$toISO", "'"+query.TimeRange.To.Format(time.RFC3339)+"'")
	}

	log.DefaultLogger.Debug("QueryData called", "query", qm.QueryText)

	res, err := client.Post(configDatasource.Path, strings.NewReader(qm.QueryText), nil)

	if err != nil {
		log.DefaultLogger.Error("query", "http error", err)

		response.Error = err
	}

	buf, err := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		response.Error = errors.New(string(buf))

		return response
	}

	var warp10_response []interface{}

	response.Error = json.Unmarshal(buf, &warp10_response)

	for _, rm := range UnmarshalWarp10Response(warp10_response) {
		if rm.Labels == nil {
			continue
		}

		frame := data.NewFrame(rm.Name)

		times := []time.Time{}
		values := []float64{}

		for _, value := range rm.Values {
			times = append(times, time.Unix(int64(value[0])/1000000, int64(value[0])%1000000))
			values = append(values, value[1])
		}

		frame.Fields = append(frame.Fields,
			data.NewField("time", rm.Labels, times),
			data.NewField(rm.Name, rm.Labels, values),
		)

		response.Frames = append(response.Frames, frame)
	}

	return response
}

func (d *Warp10Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Debug("CheckHealth called", "request", req)

	var config ConfigDatasource

	err := json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &config)

	if err != nil {
		log.DefaultLogger.Info("CheckHealth", "response error", err)
	}

	res, err := client.Post(config.Path, strings.NewReader(""), http.Header{})

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	if err != nil || res.StatusCode != 200 {
		status = backend.HealthStatusError
		message = "Error to communicate with warp10"
		log.DefaultLogger.Info("CheckHealth", "response error", err)
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
		JSONDetails: []byte{},
	}, nil
}

var client = httpclient.NewClient(httpclient.WithHTTPTimeout(10000 * time.Millisecond))
