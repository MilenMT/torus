package jaeger

import (
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

// Init creates a new instance of tracer and set it as GlobalTracer.
func Init(serviceName string) error {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
			// TODO
			BufferFlushInterval: 1 * time.Second,
			QueueSize:           10,
		},
	}

	tracer, _, err := cfg.New(
		serviceName,
	)
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}
