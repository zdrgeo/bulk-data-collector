### OpenTelemetry Collector

Docker

```shell
docker run --rm -p 4317:4317 -p 4318:4318 -v $(pwd)/config.yaml:/etc/otelcol-contrib/config.yaml otel/opentelemetry-collector-contrib --config /etc/otelcol-contrib/config.yaml
```

Docker Compose

```shell
docker compose up -d
docker compose down
```

Kubernetes

```shell
kubectl create configmap otelcol-contrib-config --from-file=config.yaml

kubectl apply -f deployment.yaml
kubectl delete -f deployment.yaml
```

### Azure Monitor

```kusto
customMetrics
| where name == "Device_DeviceInfo_ProcessStatus_CPUUsage"
| extend OUI = tostring(customDimensions.OUI), ProductClass = tostring(customDimensions.ProductClass), SerialNumber = tostring(customDimensions.SerialNumber)
| where SerialNumber in ("AB00", "AB01", "AB10", "AB11") 
| summarize AvvgCPUUsage = avg(value) by SerialNumber, Time = bin(timestamp, 30s)
```

### Azure Data Explorer

[OpenTelemetry Collector - Azure Data Explorer Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/azuredataexplorerexporter/README.md)

[Azure Data Explorer - OpenTelemetry Connector](https://learn.microsoft.com/en-us/azure/data-explorer/open-telemetry-connector?tabs=command-line)

```kusto
.create-merge table OTELLogs (Timestamp:datetime, ObservedTimestamp:datetime, TraceID:string, SpanID:string, SeverityText:string, SeverityNumber:int, Body:string, ResourceAttributes:dynamic, LogsAttributes:dynamic)
.create-merge table OTELMetrics (Timestamp:datetime, MetricName:string, MetricType:string, MetricUnit:string, MetricDescription:string, MetricValue:real, Host:string, ResourceAttributes:dynamic,MetricAttributes:dynamic)
.create-merge table OTELTraces (TraceID:string, SpanID:string, ParentID:string, SpanName:string, SpanStatus:string, SpanKind:string, StartTime:datetime, EndTime:datetime, ResourceAttributes:dynamic, TraceAttributes:dynamic, Events:dynamic, Links:dynamic)

.alter-merge table OTELTraces (SpanStatusMessage:string)

.alter table OTELLogs policy streamingingestion enable
.alter table OTELMetrics policy streamingingestion enable
.alter table OTELTraces policy streamingingestion enable

.add database oteldb ingestors ('aadapp=325195ae-1ad3-4170-879a-0e33f0aeb00f') 'Azure Data Explorer App Registration'
```

```kusto
OTELMetrics
| where MetricName == "Device_DeviceInfo_ProcessStatus_CPUUsage"
| extend OUI = tostring(MetricAttributes.OUI), ProductClass = tostring(MetricAttributes.ProductClass), SerialNumber = tostring(MetricAttributes.SerialNumber)
| where SerialNumber in ("AB00", "AB01", "AB10", "AB11") 
| summarize AvgCPUUsage = avg(MetricValue) by SerialNumber, Time = bin(Timestamp, 30s)
```

```kusto
.clear table OTELLogs data
.clear table OTELMetrics data
.clear table OTELTraces data
```
