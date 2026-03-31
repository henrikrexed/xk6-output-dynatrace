package dynatracewriter

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// OTLPOutput sends k6 metrics to Dynatrace via OTLP/HTTP.
type OTLPOutput struct {
	config          *Config
	periodicFlusher *output.PeriodicFlusher
	output.SampleBuffer
	logger   logrus.FieldLogger
	exporter sdkmetric.Exporter
	resource *resource.Resource
}

var _ output.Output = new(OTLPOutput)

// NewOTLP creates an OTLP-based output for Dynatrace.
func NewOTLP(params output.Params) (*OTLPOutput, error) {
	config, err := GetConsolidatedConfig(params.JSONConfig, params.Environment, params.ConfigArgument)
	if err != nil {
		return nil, err
	}

	newconfig, err := config.ConstructConfig()
	if err != nil {
		return nil, err
	}

	// Build OTLP endpoint from Dynatrace URL
	otlpEndpoint := newconfig.Url
	if otlpEndpoint == "" {
		return nil, fmt.Errorf("K6_DYNATRACE_URL is required for OTLP output mode")
	}

	// Build headers with auth, excluding Content-Type which is managed by the OTLP exporter
	headers := make(map[string]string)
	for k, v := range newconfig.Headers {
		if !strings.EqualFold(k, "Content-Type") {
			headers[k] = v
		}
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(otlpEndpoint),
		otlpmetrichttp.WithHeaders(headers),
		otlpmetrichttp.WithURLPath("/api/v2/otlp/v1/metrics"),
		otlpmetrichttp.WithTemporalitySelector(func(sdkmetric.InstrumentKind) metricdata.Temporality {
			return metricdata.DeltaTemporality
		}),
	}

	if newconfig.InsecureSkipTLSVerify.Bool {
		tlsCfg := &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- user-configured skip
		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(tlsCfg))
	}

	exporter, err := otlpmetrichttp.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("k6"),
		attribute.String("k6.output", "dynatrace-otlp"),
	)

	return &OTLPOutput{
		config:   newconfig,
		logger:   params.Logger,
		exporter: exporter,
		resource: res,
	}, nil
}

func (*OTLPOutput) Description() string {
	return "Output k6 metrics to Dynatrace via OTLP/HTTP"
}

func (o *OTLPOutput) Start() error {
	if periodicFlusher, err := output.NewPeriodicFlusher(time.Duration(o.config.FlushPeriod.Duration), o.flush); err != nil {
		return err
	} else {
		o.periodicFlusher = periodicFlusher
	}
	o.logger.Info("Dynatrace OTLP: started")
	return nil
}

func (o *OTLPOutput) Stop() error {
	o.logger.Info("Dynatrace OTLP: stopping")
	o.periodicFlusher.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return o.exporter.Shutdown(ctx)
}

func (o *OTLPOutput) flush() {
	start := time.Now()
	samplesContainers := o.GetBufferedSamples()

	if len(samplesContainers) == 0 {
		return
	}

	rm := o.convertToOTLPMetrics(samplesContainers)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := o.exporter.Export(ctx, &rm); err != nil {
		o.logger.WithError(err).Error("Dynatrace OTLP: export failed")
	} else {
		o.logger.WithField("duration", time.Since(start)).Debug("Dynatrace OTLP: export complete")
	}
}

func (o *OTLPOutput) convertToOTLPMetrics(samplesContainers []metrics.SampleContainer) metricdata.ResourceMetrics {
	metricsMap := make(map[string]*metricdata.Sum[float64])

	for _, sc := range samplesContainers {
		for _, sample := range sc.GetSamples() {
			name := "k6." + sample.Metric.Name

			attrs := []attribute.KeyValue{}
			if sample.Tags != nil {
				for k, v := range sample.Tags.Map() {
					if len(k) > 0 && len(v) > 0 {
						// Normalize url and name attributes to reduce cardinality
						if k == "url" || k == "name" {
							v = normalizeURL(v)
						}
						attrs = append(attrs, attribute.String(k, v))
					}
				}
			}

			dp := metricdata.DataPoint[float64]{
				Attributes: attribute.NewSet(attrs...),
				StartTime:  sample.Time.Add(-time.Second),
				Time:       sample.Time,
				Value:      sample.Value,
			}

			if existing, ok := metricsMap[name]; ok {
				existing.DataPoints = append(existing.DataPoints, dp)
			} else {
				metricsMap[name] = &metricdata.Sum[float64]{
					DataPoints:  []metricdata.DataPoint[float64]{dp},
					Temporality: metricdata.DeltaTemporality,
					IsMonotonic: false,
				}
			}
		}
	}

	scopeMetrics := make([]metricdata.Metrics, 0, len(metricsMap))
	for name, sum := range metricsMap {
		scopeMetrics = append(scopeMetrics, metricdata.Metrics{
			Name: name,
			Data: *sum,
		})
	}

	return metricdata.ResourceMetrics{
		Resource: o.resource,
		ScopeMetrics: []metricdata.ScopeMetrics{
			{
				Scope: o.instrumentationScope(),
				Metrics: scopeMetrics,
			},
		},
	}
}

func (o *OTLPOutput) instrumentationScope() instrumentation.Scope {
	return instrumentation.Scope{
		Name:    "xk6-output-dynatrace",
		Version: "0.2.0",
	}
}
