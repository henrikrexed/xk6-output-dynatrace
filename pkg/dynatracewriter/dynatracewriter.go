package dynatracewriter

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

type Output struct {
	config *Config
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

	newconfig, err := config.ConstructConfig()
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
	dynatraceMetrics := o.convertToTimeDynatraceData(samplesContainers)
	nts = len(dynatraceMetrics)

	if nts == 0 {
		o.logger.Debug("no data to send")
		return
	}

	o.logger.WithField("nts", nts).Debug("Converted samples to time series in preparation for sending.")

	results := batchSend(
		dynatraceMetrics,
		o.config.Url,
		o.config.Headers,
		o.config.BatchSize,
		o.config.MaxConcurrentExports,
		o.logger,
	)

	var failed int
	for _, r := range results {
		if r.err != nil {
			failed++
			o.logger.WithError(r.err).WithField("batch", r.batchIndex).
				Error("Dynatrace: batch send failed")
		}
	}
	if failed > 0 {
		o.logger.WithField("failed_batches", failed).WithField("total_batches", len(results)).
			Warn("Dynatrace: some batches failed to send")
	}
}

func generatePayload(dynatraceMetrics []dynatraceMetric) string {

    var result=""
    for i:= 0; i < len(dynatraceMetrics); i++ {
        result+=dynatraceMetrics[i].toText()+"\n"
    }

    return result
}

func (o *Output) convertToTimeDynatraceData(samplesContainers []metrics.SampleContainer) []dynatraceMetric {
	var dynTimeSeries []dynatraceMetric

	for _, samplesContainer := range samplesContainers {
		samples := samplesContainer.GetSamples()

		for _, sample := range samples {
			// Prometheus remote write treats each label array in TimeSeries as the same
			// for all Samples in those TimeSeries (https://github.com/prometheus/prometheus/blob/03d084f8629477907cab39fc3d314b375eeac010/storage/remote/write_handler.go#L75).
			// But K6 metrics can have different tags per each Sample so in order not to
			// lose info in tags or assign tags wrongly, let's store each Sample in a different TimeSeries, for now.
			// This approach also allows to avoid hard to replicate issues with duplicate timestamps.

            dynametric := samleToDynametric( sample)
            if &dynametric.metricValue != nil {
                o.logger.Debug("metric name : " + dynametric.metricKeyName)
                dynTimeSeries = append  (dynTimeSeries, dynametric)
            } else {
                o.logger.Debug("The value is missing")
            }
		}

		// Do not blow up if remote endpoint is overloaded and responds too slowly.
		// TODO: consider other approaches
		if flushTooLong && len(dynTimeSeries) > 150000 {
			break
		}
	}

	return dynTimeSeries
}