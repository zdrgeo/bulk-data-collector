package main

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/autopaho/queue/memory"
	"github.com/eclipse/paho.golang/paho"
	"github.com/spf13/viper"
	handlers "github.com/zdrgeo/bulk-data-collector/pkg/handlers"
	mqttservices "github.com/zdrgeo/bulk-data-collector/pkg/services/mqtt"
)

var (
	logger            *slog.Logger
	connectionManager *autopaho.ConnectionManager
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

	initMQTT()
}

func initMQTT() {
	var err error

	serverUrl, err := url.Parse(viper.GetString("MQTT_SERVER_URL"))

	if err != nil {
		log.Panic(err)
	}

	certificate, err := tls.LoadX509KeyPair(viper.GetString("MQTT_CERT_FILE"), viper.GetString("MQTT_KEY_FILE"))

	if err != nil {
		log.Panic(err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	clientConfig := autopaho.ClientConfig{
		Queue:                         memory.New(),
		ServerUrls:                    []*url.URL{serverUrl},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         3600,
		ConnectUsername:               viper.GetString("MQTT_CONNECT_USERNAME"),
		ConnectPassword:               []byte(viper.GetString("MQTT_CONNECT_PASSWORD")),
		TlsCfg:                        tlsCfg,
		ClientConfig: paho.ClientConfig{
			ClientID: viper.GetString("MQTT_CLIENT_ID"),
		},
	}

	if connectionManager, err = autopaho.NewConnection(context.Background(), clientConfig); err != nil {
		log.Panic(err)
	}
}

func main() {
	mainMQTT()
}

func mainMQTT() {
	mqttCollectorServiceOptions := &mqttservices.MQTTCollectorServiceOptions{
		CollectorName: viper.GetString("COLLECTOR_NAME"),
	}

	collectorService := mqttservices.NewMQTTCollectorService(connectionManager, mqttCollectorServiceOptions)

	collectorHandler := handlers.NewCollectorHandler(collectorService)

	http.Handle("/collector", http.HandlerFunc(collectorHandler.Collect))

	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Panic(err)
	}
}
