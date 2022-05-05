package dynatracewriter

import (
	"fmt"
	"time"
    "net/http"
	//nolint:staticcheck

	"github.com/sirupsen/logrus"
	"go.k6.io/k6/output"
	"go.k6.io/k6/stats"
)

type Output struct {
	config Config
	periodicFlusher *output.PeriodicFlusher
	output.SampleBuffer
    params  output.Params
	logger logrus.FieldLogger
}

var _ output.Output = new(Output)

// toggle to indicate whether we should stop dropping samples
var flushTooLong bool

func New(params output.Params) (*Output, error) {
	config, err := GetConsolidatedConfig(params.JSONConfig, params.Environment, params.ConfigArgument)
	if err != nil {
		return nil, err
	}

	newconfig, err := config.ConstructRemoteConfig()
	if err != nil {
		return nil, err
	}

	return &Output{
		config:  newconfig,
		logger:  params.Logger,
	}, nil
}

func (*Output) Description() string {
	return "Output k6 metrics to Dynatrace metrics ingest api"
}

func (o *Output) Start() error {
	if periodicFlusher, err := output.NewPeriodicFlusher(time.Duration(o.config.FlushPeriod.Duration), o.flush); err != nil {
		return err
	} else {
		o.periodicFlusher = periodicFlusher
	}
	o.logger.Debug("Dynatrace: starting dynatrace-write")

	return nil
}

func (o *Output) Stop() error {
	o.logger.Debug("Dynatrace: stopping dynatrace-write")
	o.periodicFlusher.Stop()
	return nil
}

func (o *Output) flush() {
	var (
		start = time.Now()
		nts   int
	)

	defer func() {
		d := time.Since(start)
		if d > time.Duration(o.config.FlushPeriod.Duration) {
			// There is no intermediary storage so warn if writing to remote write endpoint becomes too slow
			o.logger.WithField("nts", nts).
				Warn(fmt.Sprintf("Remote write took %s while flush period is %s. Some samples may be dropped.",
					d.String(), o.config.FlushPeriod.String()))
			flushTooLong = true
		} else {
			o.logger.WithField("nts", nts).Debug(fmt.Sprintf("Remote write took %s.", d.String()))
			flushTooLong = false
		}
	}()

	samplesContainers := o.GetBufferedSamples()

	// Remote write endpoint accepts TimeSeries structure defined in gRPC. It must:
	// a) contain Labels array
	// b) have a __name__ label: without it, metric might be unquerable or even rejected
	// as a metric without a name. This behaviour depends on underlying storage used.
	// c) not have duplicate timestamps within 1 timeseries, see https://github.com/prometheus/prometheus/issues/9210
	// Prometheus write handler processes only some fields as of now, so here we'll add only them.
	dynatraceMetric := o.convertToTimeDynatraceData(samplesContainers)
	nts = len(dynatraceMetric)

	o.logger.WithField("nts", nts).Debug("Converted samples to time series in preparation for sending.")

    payload =generatePayload(dynatraceMetric)
	request, error := http.NewRequest("POST", o.config.Url, bytes.NewBuffer(payload))
	for key,value := range o.config.Headers {
	    request.Header.Set(key, value)
	}

    client := &http.Client{}
    response, error := client.Do(request)
    if error != nil {
        po.logger.WithError(error).Fatal("Failed to send timeseries.")
    }
    defer response.Body.Close()

    o.logger.Debug("response Status:" + response.Status)
    o.logger.Debug("response Headers:" +  response.Header)
    body, _ := ioutil.ReadAll(response.Body)
    o.logger.Debug("response Body:"+ string(body))

}

func generatePayload(dynatraceMetric *[]dynatraceMetric) string {

    var result=""
    for i:= 0; e < len(dynatraceMetric); i++{
        result+=dynatraceMetric[i].toText()+"\n"
    }

    return result
}

func (o *Output) convertToTimeDynatraceData(samplesContainers []stats.SampleContainer) []dynatraceMetric {
	dynTimeSeries := make([]dynatraceMetric, 0)

	for _, samplesContainer := range samplesContainers {
		samples := samplesContainer.GetSamples()

		for _, sample := range samples {
			// Prometheus remote write treats each label array in TimeSeries as the same
			// for all Samples in those TimeSeries (https://github.com/prometheus/prometheus/blob/03d084f8629477907cab39fc3d314b375eeac010/storage/remote/write_handler.go#L75).
			// But K6 metrics can have different tags per each Sample so in order not to
			// lose info in tags or assign tags wrongly, let's store each Sample in a different TimeSeries, for now.
			// This approach also allows to avoid hard to replicate issues with duplicate timestamps.

            dynametric := samleToDynametric( sample)


            dynTimeSeries = append(dynTimeSeries, dynametric...)

		}

		// Do not blow up if remote endpoint is overloaded and responds too slowly.
		// TODO: consider other approaches
		if flushTooLong && len(promTimeSeries) > 150000 {
			break
		}
	}

	return dynTimeSeries
}