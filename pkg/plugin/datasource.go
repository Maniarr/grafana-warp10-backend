package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/mitchellh/mapstructure"

	warp10 "github.com/miton18/go-warp10/base"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	Client warp10.Client
}

type DatasourceConfig struct {
	BaseUrl string
}

type DatasourceSecuredConfig struct {
	Token string
}

func NewDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var config DatasourceConfig

	_ = json.Unmarshal(settings.JSONData, &config)

	client := warp10.NewClient(config.BaseUrl)
	client.ReadToken = settings.DecryptedSecureJSONData["token"]

	return &Datasource{
		Client: *client,
	}, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Debug("Query", req)
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	Warpscript string
	Alias      string
	ShowLabels bool
}

type GTS struct {
	// Name of the time serie
	ClassName string `json:"c" mapstructure:"c"`
	// Key/value of the GTS labels (changing one key or his value create a new GTS)
	Labels map[string]string `json:"l" mapstructure:"l"`
	// Key/value of the GTS attributes (can be setted/updated/removed without creating a new GTS)
	Attributes warp10.Attributes `json:"a" mapstructure:"a"`
	// Timestamp of the last datapoint received on this GTS (% last activity window)
	LastActivity int64 `json:"la" mapstructure:"la"`
	// Array of datapoints of this GTS
	Values [][2]float64 `json:"v" mapstructure:"v"`
}

type GTSList = []*GTS

func UnmarshalWarp10Response(raw []interface{}) GTSList {
	gtsList := make(GTSList, 0)

	log.DefaultLogger.Debug("Raw", raw)

	for _, value := range raw {
		switch value.(type) {
		case []interface{}:
			log.DefaultLogger.Debug("[]interface{}")
			gtsList = append(gtsList, UnmarshalWarp10Response(value.([]interface{}))...)
		case interface{}:
			log.DefaultLogger.Debug("interface{}")
			var gts GTS

			err := mapstructure.Decode(value.(map[string]interface{}), &gts)

			if err != nil {
				log.DefaultLogger.Error("UnmarshalWarp10Response", "cast error", err)
			}

			gtsList = append(gtsList, &gts)
		}
	}

	return gtsList
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	var qm queryModel

	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	if strings.Contains(qm.Warpscript, "$read_token") {
		qm.Warpscript = strings.ReplaceAll(qm.Warpscript, "$read_token", fmt.Sprintf("'%v'", d.Client.ReadToken))
	}

	if strings.Contains(qm.Warpscript, "$fromISO") {
		qm.Warpscript = strings.ReplaceAll(qm.Warpscript, "$fromISO", fmt.Sprintf("'%v'", query.TimeRange.From.Format(time.RFC3339)))
	}

	if strings.Contains(qm.Warpscript, "$toISO") {
		qm.Warpscript = strings.ReplaceAll(qm.Warpscript, "$toISO", fmt.Sprintf("'%v'", query.TimeRange.To.Format(time.RFC3339)))
	}

	if strings.Contains(qm.Warpscript, "$interval") {
		qm.Warpscript = strings.ReplaceAll(qm.Warpscript, "$interval", strconv.FormatInt(query.Interval.Microseconds(), 10))
	}

	log.DefaultLogger.Debug("Execute query", query, qm)

	resp, err := d.Client.Exec(qm.Warpscript)

	if err != nil {
		log.DefaultLogger.Error("Error to exec query", err)

		response.Error = err

		return response
	}

	log.DefaultLogger.Debug("Query response", string(resp))

	var warp10Response []interface{}

	err = json.Unmarshal(resp, &warp10Response)

	if err != nil {
		log.DefaultLogger.Error("Error to parse warp10 response", err)

		response.Error = err

		return response
	}

	gtsList := UnmarshalWarp10Response(warp10Response)

	log.DefaultLogger.Debug("Warp10 response", gtsList)

	for _, gts := range gtsList {
		gtsName := gts.ClassName
		labels := make(map[string]string)

		if qm.Alias != "" {
			gtsName = qm.Alias

			for key, value := range gts.Labels {
				labelKey := fmt.Sprintf("$label_%s", key)

				if strings.Contains(gtsName, labelKey) {
					gtsName = strings.ReplaceAll(gtsName, labelKey, value)
				}
			}
		}

		if gts.Labels != nil && qm.ShowLabels {
			labels = gts.Labels
		}

		frame := data.NewFrame(query.RefID)

		times := []time.Time{}
		values := []float64{}

		for _, value := range gts.Values {
			times = append(times, time.Unix(int64(value[0])/1000000, int64(value[0])%1000000))
			values = append(values, float64(value[1]))
		}

		frame.Fields = append(frame.Fields,
			data.NewField("time", labels, times),
			data.NewField(gtsName, labels, values),
		)

		response.Frames = append(response.Frames, frame)
	}

	return response
}

type TokenInfo struct {
	ReadTokenDecodeError *string `json:"ReadTokenDecodeError,omitempty"`
	Type                 *string `json:"Type,omitempty"`
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "Data source is working"

	if d.Client.Host == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "No base url provided",
		}, nil
	}

	if d.Client.ReadToken == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "No read token provided",
		}, nil
	}

	resp, err := d.Client.Exec(fmt.Sprintf("'%s' TOKENINFO", d.Client.ReadToken))

	if err != nil {
		status = backend.HealthStatusError
		message = err.Error()

		if err.Error() == "" {
			message = "Base url not correct"
		}
	}

	var infos []TokenInfo

	err = json.Unmarshal(resp, &infos)

	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}

	if len(infos) < 1 {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Error to check token",
		}, nil
	}

	info := infos[0]

	if info.ReadTokenDecodeError != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: *info.ReadTokenDecodeError,
		}, nil
	}

	if *info.Type != "READ" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Provided token is not a read token",
		}, nil
	}

	// TODO: check if token is expired

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
