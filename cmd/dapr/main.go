package main

import (
	"log"
	"log/slog"
	"net/http"

	daprclient "github.com/dapr/go-sdk/client"
	"github.com/spf13/viper"
	"github.com/zdrgeo/bulk-data-collector/pkg/handlers"
	daprservices "github.com/zdrgeo/bulk-data-collector/pkg/services/dapr"
)

const (
	storeName  = "iotoperations-statestore"
	pubSubName = "iotoperations-pubsub"
	topicName  = "collector"
	clientID   = "100000-1"
	username   = "100000"
)

var (
	logger     *slog.Logger
	daprClient daprclient.Client
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

	initDapr()
}

func initDapr() {
	var err error

	daprClient, err = daprclient.NewClient()

	if err != nil {
		log.Panic(err)
	}
}

func main() {
	mainDapr()
}

func mainDapr() {
	collectorServiceOptions := &daprservices.DaprCollectorServiceOptions{
		PubSubName: pubSubName,
		TopicName:  topicName,
	}

	collectorService := daprservices.NewDaprCollectorService(daprClient, collectorServiceOptions)

	collectorHandler := handlers.NewCollectorHandler(collectorService)

	http.Handle("/collector", http.HandlerFunc(collectorHandler.Collect))

	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Panic(err)
	}
}
