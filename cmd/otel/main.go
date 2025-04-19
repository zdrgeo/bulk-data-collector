package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	handlers "github.com/zdrgeo/bulk-data-collector/pkg/handlers"
	otelservices "github.com/zdrgeo/bulk-data-collector/pkg/services/otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	logger *slog.Logger
)

func init() {
	logger = slog.Default()
	// Use otelslog bridge to integrate with OpenTelemetry (https://pkg.go.dev/go.opentelemetry.io/otel/sdk/log)
	// logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{AddSource: true}))
	// logger := slog.New(slog.NewJSONHandler(nil, &slog.HandlerOptions{AddSource: true}))

	viper.AddConfigPath(".")
	// viper.SetConfigFile(".env")
	// viper.SetConfigName("config")
	// viper.SetConfigType("env") // "env", "json", "yaml"
	viper.SetEnvPrefix("bulk_data_collector")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Panic(err)
	}

	initOTel()
}

func initOTel() {
	encoder := json.NewEncoder(os.Stdout)

	encoder.SetIndent("", "  ")

	stdoutExporter, err := stdoutmetric.New(
		stdoutmetric.WithEncoder(encoder),
		stdoutmetric.WithoutTimestamps(),
	)

	if err != nil {
		log.Panic(err)
	}

	_ = stdoutExporter

	otlpgrpcExporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
		otlpmetricgrpc.WithInsecure(),
	)

	if err != nil {
		log.Panic(err)
	}

	_ = otlpgrpcExporter

	otlphttpExporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpoint("localhost:4317"),
		otlpmetrichttp.WithInsecure(),
	)

	if err != nil {
		log.Panic(err)
	}

	_ = otlphttpExporter

	resource := resource.NewSchemaless(
		semconv.ServiceName("bulk-data-collector"),
	)

	_ = resource

	periodicReader := metric.NewPeriodicReader(
		// stdoutExporter,
		// otlpgrpcExporter,
		otlphttpExporter,
	)

	_ = periodicReader

	prometheusExporter, err := prometheus.New()

	if err != nil {
		log.Fatal(err)
	}

	_ = prometheusExporter

	meterProvider := metric.NewMeterProvider(
		// metric.WithResource(resource),
		// metric.WithReader(periodicReader),
		metric.WithReader(prometheusExporter),
	)

	otel.SetMeterProvider(meterProvider)
}

func main() {
	mainOTel()
}

func mainOTel() {
	otelConfig := viper.Sub("otel")

	collectorServiceOptions := &otelservices.OTelCollectorServiceOptions{}

	if err := otelConfig.Unmarshal(collectorServiceOptions); err != nil {
		log.Panic(err)
	}

	collectorService, err := otelservices.NewOTelCollectorService(collectorServiceOptions)

	if err != nil {
		log.Panic(err)
	}

	collectorHandler := handlers.NewCollectorHandler(collectorService)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/collector", http.HandlerFunc(collectorHandler.Collect))

	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Panic(err)
	}
}
