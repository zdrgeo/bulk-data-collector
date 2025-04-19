package otel

import (
	"context"
	"errors"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	ErrInvalidInstrumentKind = errors.New("invalid instrument kind")
	ErrInvalidValueType      = errors.New("invalid value type")
)

const (
	OTelInstrumentKindInt64Counter         = "Int64Counter"
	OTelInstrumentKindInt64Gauge           = "Int64Gauge"
	OTelInstrumentKindInt64Histogram       = "Int64Histogram"
	OTelInstrumentKindInt64UpDownCounter   = "Int64UpDownCounter"
	OTelInstrumentKindFloat64Counter       = "Float64Counter"
	OTelInstrumentKindFloat64Gauge         = "Float64Gauge"
	OTelInstrumentKindFloat64Histogram     = "Float64Histogram"
	OTelInstrumentKindFloat64UpDownCounter = "Float64UpDownCounter"
)

type OTelInstrumentOptions struct {
	ParameterName string `json:"ParameterName"`
	Name          string `json:"Name"`
	Kind          string `json:"Kind"`
	Description   string `json:"Description"`
	Unit          string `json:"Unit"`
}

type OTelInstrument interface {
	Measure(ctx context.Context, value any, attributes attribute.Set) error
}

func NewOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (OTelInstrument, error) {
	switch options.Kind {
	case OTelInstrumentKindInt64Counter:
		return NewInt64CounterOTelInstrument(meter, options)
	case OTelInstrumentKindInt64Gauge:
		return NewInt64GaugeOTelInstrument(meter, options)
	case OTelInstrumentKindInt64Histogram:
		return NewInt64HistogramOTelInstrument(meter, options)
	case OTelInstrumentKindInt64UpDownCounter:
		return NewInt64UpDownCounterOTelInstrument(meter, options)
	case OTelInstrumentKindFloat64Counter:
		return NewFloat64CounterOTelInstrument(meter, options)
	case OTelInstrumentKindFloat64Gauge:
		return NewFloat64GaugeOTelInstrument(meter, options)
	case OTelInstrumentKindFloat64Histogram:
		return NewFloat64HistogramOTelInstrument(meter, options)
	case OTelInstrumentKindFloat64UpDownCounter:
		return NewFloat64UpDownCounterOTelInstrument(meter, options)
	}

	return nil, ErrInvalidInstrumentKind
}

func toInt64(value any) (int64, error) {
	var (
		int64Value int64
		err        error
	)

	switch v := value.(type) {
	case string:
		int64Value, err = strconv.ParseInt(v, 10, 64)
	case bool:
		if v {
			int64Value, err = 1, nil
		} else {
			int64Value, err = 0, nil
		}
	case int:
		int64Value, err = int64(v), nil
	case int64:
		int64Value, err = v, nil
	case uint:
		int64Value, err = int64(v), nil
	case uint64:
		int64Value, err = int64(v), nil
	case float32:
		int64Value, err = int64(v), nil
	case float64:
		int64Value, err = int64(v), nil
	default:
		int64Value, err = 0, ErrInvalidValueType
	}

	return int64Value, err
}

func toFloat64(value any) (float64, error) {
	var (
		float64Value float64
		err          error
	)

	switch v := value.(type) {
	case string:
		float64Value, err = strconv.ParseFloat(v, 64)
	case bool:
		if v {
			float64Value, err = 1, nil
		} else {
			float64Value, err = 0, nil
		}
	case int:
		float64Value, err = float64(v), nil
	case int64:
		float64Value, err = float64(v), nil
	case uint:
		float64Value, err = float64(v), nil
	case uint64:
		float64Value, err = float64(v), nil
	case float32:
		float64Value, err = float64(v), nil
	case float64:
		float64Value = v
	default:
		float64Value, err = 0, ErrInvalidValueType
	}

	return float64Value, err
}

type Int64CounterOTelInstrument struct {
	counter metric.Int64Counter
}

func NewInt64CounterOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (*Int64CounterOTelInstrument, error) {
	counter, err := meter.Int64Counter(options.Name, metric.WithDescription(options.Description), metric.WithUnit(options.Unit))

	if err != nil {
		return nil, err
	}

	return &Int64CounterOTelInstrument{counter: counter}, nil
}

func (i *Int64CounterOTelInstrument) Measure(ctx context.Context, value any, attributes attribute.Set) error {
	int64Value, err := toInt64(value)

	if err != nil {
		return err
	}

	i.counter.Add(ctx, int64Value, metric.WithAttributeSet(attributes))

	return nil
}

type Int64GaugeOTelInstrument struct {
	gauge metric.Int64Gauge
}

func NewInt64GaugeOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (*Int64GaugeOTelInstrument, error) {
	gauge, err := meter.Int64Gauge(options.Name, metric.WithDescription(options.Description), metric.WithUnit(options.Unit))

	if err != nil {
		return nil, err
	}

	return &Int64GaugeOTelInstrument{gauge: gauge}, nil
}

func (i *Int64GaugeOTelInstrument) Measure(ctx context.Context, value any, attributes attribute.Set) error {
	int64Value, err := toInt64(value)

	if err != nil {
		return err
	}

	i.gauge.Record(ctx, int64Value, metric.WithAttributeSet(attributes))

	return nil
}

type Int64HistogramOTelInstrument struct {
	histogram metric.Int64Histogram
}

func NewInt64HistogramOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (*Int64HistogramOTelInstrument, error) {
	histogram, err := meter.Int64Histogram(options.Name, metric.WithDescription(options.Description), metric.WithUnit(options.Unit))

	if err != nil {
		return nil, err
	}

	return &Int64HistogramOTelInstrument{histogram: histogram}, nil
}

func (i *Int64HistogramOTelInstrument) Measure(ctx context.Context, value any, attributes attribute.Set) error {
	int64Value, err := toInt64(value)

	if err != nil {
		return err
	}

	i.histogram.Record(ctx, int64Value, metric.WithAttributeSet(attributes))

	return nil
}

type Int64UpDownCounterOTelInstrument struct {
	upDownCounter metric.Int64UpDownCounter
}

func NewInt64UpDownCounterOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (*Int64UpDownCounterOTelInstrument, error) {
	upDownCounter, err := meter.Int64UpDownCounter(options.Name, metric.WithDescription(options.Description), metric.WithUnit(options.Unit))

	if err != nil {
		return nil, err
	}

	return &Int64UpDownCounterOTelInstrument{upDownCounter: upDownCounter}, nil
}

func (i *Int64UpDownCounterOTelInstrument) Measure(ctx context.Context, value any, attributes attribute.Set) error {
	int64Value, err := toInt64(value)

	if err != nil {
		return err
	}

	i.upDownCounter.Add(ctx, int64Value, metric.WithAttributeSet(attributes))

	return nil
}

type Float64CounterOTelInstrument struct {
	counter metric.Float64Counter
}

func NewFloat64CounterOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (*Float64CounterOTelInstrument, error) {
	counter, err := meter.Float64Counter(options.Name, metric.WithDescription(options.Description), metric.WithUnit(options.Unit))

	if err != nil {
		return nil, err
	}

	return &Float64CounterOTelInstrument{counter: counter}, nil
}

func (i *Float64CounterOTelInstrument) Measure(ctx context.Context, value any, attributes attribute.Set) error {
	float64Value, err := toFloat64(value)

	if err != nil {
		return err
	}

	i.counter.Add(ctx, float64Value, metric.WithAttributeSet(attributes))

	return nil
}

type Float64GaugeOTelInstrument struct {
	gauge metric.Float64Gauge
}

func NewFloat64GaugeOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (*Float64GaugeOTelInstrument, error) {
	gauge, err := meter.Float64Gauge(options.Name, metric.WithDescription(options.Description), metric.WithUnit(options.Unit))

	if err != nil {
		return nil, err
	}

	return &Float64GaugeOTelInstrument{gauge: gauge}, nil
}

func (i *Float64GaugeOTelInstrument) Measure(ctx context.Context, value any, attributes attribute.Set) error {
	float64Value, err := toFloat64(value)

	if err != nil {
		return err
	}

	i.gauge.Record(ctx, float64Value, metric.WithAttributeSet(attributes))

	return nil
}

type Float64HistogramOTelInstrument struct {
	histogram metric.Float64Histogram
}

func NewFloat64HistogramOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (*Float64HistogramOTelInstrument, error) {
	histogram, err := meter.Float64Histogram(options.Name, metric.WithDescription(options.Description), metric.WithUnit(options.Unit))

	if err != nil {
		return nil, err
	}

	return &Float64HistogramOTelInstrument{histogram: histogram}, nil
}

func (i *Float64HistogramOTelInstrument) Measure(ctx context.Context, value any, attributes attribute.Set) error {
	float64Value, err := toFloat64(value)

	if err != nil {
		return err
	}

	i.histogram.Record(ctx, float64Value, metric.WithAttributeSet(attributes))

	return nil
}

type Float64UpDownCounterOTelInstrument struct {
	upDownCounter metric.Float64UpDownCounter
}

func NewFloat64UpDownCounterOTelInstrument(meter metric.Meter, options *OTelInstrumentOptions) (*Float64UpDownCounterOTelInstrument, error) {
	upDownCounter, err := meter.Float64UpDownCounter(options.Name, metric.WithDescription(options.Description), metric.WithUnit(options.Unit))

	if err != nil {
		return nil, err
	}

	return &Float64UpDownCounterOTelInstrument{upDownCounter: upDownCounter}, nil
}

func (i *Float64UpDownCounterOTelInstrument) Measure(ctx context.Context, value any, attributes attribute.Set) error {
	float64Value, err := toFloat64(value)

	if err != nil {
		return err
	}

	i.upDownCounter.Add(ctx, float64Value, metric.WithAttributeSet(attributes))

	return nil
}
