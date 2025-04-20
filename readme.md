# CWMP and USP bulk data collector

## Overview

In telecom industry, CWMP (or TR-069) is a widely adopted protocol - one of the best options for the ISPs to manage sometimes millions of devices remotely. There are two main players in this protocol - the CPE (Customer Premises Equipment) devices and the ACS (Auto-Configuration Server) server that controls them. Being SOAP-based the protocol does a very good job at controlling the devices remotely, but presents challenges when it comes to collecting large amounts of telemetry data from the devices. To address this, an extension to the protocol exists that allows you to use a separate endpoint where the devices periodically submit bulk data reports containing the values of the device parameters in CSV or JSON format. In addition to the more efficient format, this completely decouples the telemetry data plane from the data plane used by the ACS to control the CPEs. **This means you have more options where to send the telemetry.** Instead of sending it to the ACS (or to a built-in component of the ACS solution resposible for this) you can choose to send it to a dedicated analytics or telemetry platform. These platforms are often more scalable and have more advanced capabilities than the ACS solution can offer. This repo explores a few such options.

- [Azure Event Hubs](#azure-event-hubs)
- [Open Telemetry](#open-telemetry)
- [MQTT](#mqtt)
- [Dapr](#dapr)

> [!NOTE]
> There is a newer, more capable and efficient gRPC based protocol - USP (or TR-369) that aims to replace CWMP but it is not yet widely adopted. USP uses the same mechanism for bulk data collection as CWMP, so the bulk data collector can also be used with USP.

## Context

```mermaid
graph LR
    ACS[ACS]
    ACS <-.-> CPE1
    ACS <-.-> CPEn
    CPE1[CPE 1] --> BDC
    CPEn[CPE N] --> BDC
    BDC((Bulk Data Collector))
```

Work in progress...

## Azure Event Hubs

This variant of the collector sends the collected data to [Azure Events Hubs](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-about) - the main Azure real-time data ingestion service. Once the data is ingested into Event Hubs, there is a large number of real-time stream processing, data analytics and data storage services that you can use to extract insights from it.

The Azure Event Hubs collector variant is relatively more complex than the others. It is worth taking a look at its internal components so you can configure it to work effectively.
```mermaid
graph LR
    subgraph BDC[Bulk Data Collector]
        C[Controller]
        C --> P1Q[Partition 1 Queue]
        P1Q --> P1P1((Partition 1<br/>Producer 1))
        P1Q --> P1Pz((Partition 1<br/>Producer Z))
        C --> PyQ[Partition Y Queue]
        PyQ --> PyP1((Partition Y<br/>Producer 1))
        PyQ --> PyPz((Partition Y<br/>Producer Z))
    end
    subgraph CPEs[CPEs]
        CPE1[CPE 1] --> C
        CPEn[CPE N] --> C
    end
    subgraph AEH[Azure Event Hubs]
        P1P1 --> P1[Partition 1]
        P1Pz --> P1[Partition 1]
        PyP1 --> Py[Partition Y]
        PyPz --> Py[Partition Y]
    end
```

### Testing

1. Add `config.env` to `cmd/azureeventhubs`

```env
# cmd/azureeventhubs/config.env

AZURE_EVENTHUBS_CONNECTION_STRING=<Add the Event Hubs connection string here>
AZURE_EVENTHUBS_EVENTHUB=<Add the Event Hub name here>
PARTITION_QUEUE_LIMIT=100
PARTITION_PRODUCERS_COUNT=1
```

2. Run Prometheus

```shell
cd prometheus
docker compose  up -d
```

3. Run the collector

```shell
cd cmd/azureeventhubs
go run main.go
```

4. Run the test

```shell
cd grafana/k6
k6 run collector.js
```

Work in progress...

## Open Telemetry

This variant of the collector sends the collected data to any [Open Telemetry](https://opentelemetry.io/docs/what-is-opentelemetry/) compatible collector. I will use [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/) distribution with [Azure Data Explorer](https://learn.microsoft.com/en-us/azure/data-explorer/) exporter.

### Testing

1. Create the necessary tables in Azure Data Explorer

```kusto
.create-merge table OTELLogs (Timestamp:datetime, ObservedTimestamp:datetime, TraceID:string, SpanID:string, SeverityText:string, SeverityNumber:int, Body:string, ResourceAttributes:dynamic, LogsAttributes:dynamic)
.create-merge table OTELMetrics (Timestamp:datetime, MetricName:string, MetricType:string, MetricUnit:string, MetricDescription:string, MetricValue:real, Host:string, ResourceAttributes:dynamic,MetricAttributes:dynamic)
.create-merge table OTELTraces (TraceID:string, SpanID:string, ParentID:string, SpanName:string, SpanStatus:string, SpanKind:string, StartTime:datetime, EndTime:datetime, ResourceAttributes:dynamic, TraceAttributes:dynamic, Events:dynamic, Links:dynamic)

.alter-merge table OTELTraces (SpanStatusMessage:string)

.alter table OTELLogs policy streamingingestion enable
.alter table OTELMetrics policy streamingingestion enable
.alter table OTELTraces policy streamingingestion enable

.add database oteldb ingestors ('aadapp=<Add Microsoft Entra ID app ID here>') 'Azure Data Explorer App Registration'
```

2. Add `config.yaml` to `otelcol-contrib`

```yaml
# cmd/otelcol-contrib/config.yaml

receivers:
  otlp:
    protocols:
      # grpc:
      #  endpoint: "0.0.0.0:4317"
      http:
        endpoint: "0.0.0.0:4317"
processors:
  batch:
exporters:
  azuredataexplorer:
    cluster_uri: <Add the Azure Data Explorer URL here>
    tenant_id: <Add the Azure tenant ID here>
    application_id: <Add the applicatioin ID here>
    application_key: <Add the application key here>
    # db_name: "oteldb"
    # logs_table_name: "OTELLogs"
    # metrics_table_name: "OTELMetrics"
    # traces_table_name: "OTELTraces"
    # logs_table_json_mapping: "otellogs_mapping"
    # metrics_table_json_mapping: "otelmetrics_mapping"
    # traces_table_json_mapping: "oteltraces_mapping"
    ingestion_type: "managed"
service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [azuredataexplorer]
    metrics:
      receivers: [otlp]
      processors: []
      exporters: [azuredataexplorer]
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [azuredataexplorer]
```

3. Add `config.env` to `cmd/otel`

```env
# cmd/otel/config.env

```

4. Add `config.yaml` to `cmd/otel`

```yaml
# cmd/otel/config.yaml

otel:
  meter:
    name: "collector"
    instruments:
      - parameterName: "Device.DeviceInfo.ProcessStatus.CPUUsage"
        name: "Device_DeviceInfo_ProcessStatus_CPUUsage"
        kind: "Int64Gauge"
        description: "Process CPU usage"
        unit: "percent"
      - parameterName: "Device.DeviceInfo.MemoryStatus.Total"
        name: "Device_DeviceInfo_MemoryStatus_Total"
        kind: "Int64Gauge"
        description: "Total memory"
        unit: "byte"
      - parameterName: "Device.DeviceInfo.MemoryStatus.Free"
        name: "Device_DeviceInfo_MemoryStatus_Free"
        kind: "Int64Gauge"
        description: "Free memory"
        unit: "byte"
      - parameterName: "Device.Ethernet.Interface.1.Stats.BytesSent"
        name: "Device_Ethernet_Interface_1_Stats_BytesSent"
        kind: "Int64Counter"
        description: "Ethernet bytes sent"
        unit: "byte"
      - parameterName: "Device.Ethernet.Interface.1.Stats.BytesReceived"
        name: "Device_Ethernet_Interface_1_Stats_BytesReceived"
        kind: "Int64Counter"
        description: "Ethernet bytes received"
        unit: "byte"
      - parameterName: "Device.MoCA.Interface.1.Stats.BytesSent"
        name: "Device_MoCA_Interface_1_Stats_BytesSent"
        kind: "Int64Counter"
        description: "MoCA bytes sent"
        unit: "byte"
      - parameterName: "Device.MoCA.Interface.1.Stats.BytesReceived"
        name: "Device_MoCA_Interface_1_Stats_BytesReceived"
        kind: "Int64Counter"
        description: "MoCA bytes received"
        unit: "byte"
```

5. Run the OTel collector

```shell
cd otelcol-contrib
docker compose up -d
```

6. Run the collector

```shell
cd cmd/otel
go run main.go
```

7. Run the test

```shell
cd grafana/k6
k6 run collector.js
```

Work in progress...

## MQTT

This variant of the collector sends the collected data to any MQTT v5 compatible broker. I will use [Azure Event Grid](https://learn.microsoft.com/en-us/azure/event-grid/) with MQTT feature enabled.

### Testing

1. Add the cert and key files to `cmd/mqtt`

2. Add `config.env` to `cmd/mqtt`

```env
# cmd/mqtt/config.env

MQTT_SERVER_URL=<Add the MQTT server URL here>
MQTT_CERT_FILE=<Add the cert file here>
MQTT_KEY_FILE=<Add the key file here>
```

3. Run the collector

```shell
cd cmd/mqtt
go run main.go
```

4. Run the test

```shell
cd grafana/k6
k6 run collector.js
```

Work in progress...

## Dapr

Work in progress...
