package trace

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewMeter, NewMeterProvider, NewTracerProvider, NewTracer)
