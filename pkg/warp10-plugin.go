package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/gojek/heimdall/v7/httpclient"
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
	log.DefaultLogger.Info("QueryData called", "request", req)

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

type ResponseModel struct {
	Name   string            `json:"c"`
	Labels map[string]string `json:"l"`
	Values [][2]float64      `json:"v"`
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

	res, err := client.Post(configDatasource.Path, strings.NewReader(qm.QueryText), nil)

	if err != nil {
		log.DefaultLogger.Error("query", "error http", err)

		response.Error = err
	}

	buf, err := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		response.Error = errors.New(string(buf))

		return response
	}

	var rms []ResponseModel

	response.Error = json.Unmarshal(buf, &rms)

	if response.Error != nil {
		var rms_array [][]ResponseModel

		response.Error = json.Unmarshal(buf, &rms_array)

		if response.Error != nil {
			return response
		}

		for _, rm_items := range rms_array {
			rms = append(rms, rm_items...)
		}
	}

	for _, rm := range rms {
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
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var config ConfigDatasource

	err := json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &config)

	if err != nil {
		log.DefaultLogger.Info("CheckHealth", "error", err)
	}

	res, err := client.Post(config.Path, nil, nil)
	log.DefaultLogger.Info("CheckHealth", "response", res.StatusCode)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	if err != nil || res.StatusCode != 200 {
		status = backend.HealthStatusError
		message = "randomized error"
		log.DefaultLogger.Info("CheckHealth", "error", err)
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

var client = httpclient.NewClient(httpclient.WithHTTPTimeout(10000 * time.Millisecond))