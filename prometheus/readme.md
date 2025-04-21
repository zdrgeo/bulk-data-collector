Run Prometheus and Grafana

Open [Prometheus dashboard](http://localhost:9090) or [Grafana dashboard](http://localhost:3000).

```
docker compose --profile grafana up -d
docker compose down
```

Create a graph with the following metrics to monitor the number of events processed per partition

```promql
partition_queue_counter{partition=~".*"}

partition_batch_counter_total{partition=~".*"}

partition_event_counter_total{partition=~".*"}
```

or alternatively, use the following aggregate metrics to monitor the total event volume flowing through the entire pipeline

```promql
sum (partition_queue_counter)

sum (partition_batch_counter_total)

sum (partition_event_counter_total)
```

```
Device_MoCA_Interface_1_Stats_BytesSent_total{SerialNumber=~".*"}
Device_MoCA_Interface_1_Stats_BytesReceived_total{SerialNumber=~".*"}
```
