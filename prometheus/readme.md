Docker Compose
```
docker compose --profile grafana up -d
docker compose down
```

To compare partition metrics on one graph
```promql
partition_queue_counter{partition=~".*"}

partition_batch_counter_total{partition=~".*"}

partition_event_counter_total{partition=~".*"}
```

To aggregate metric totals by partition:
```promql
sum (partition_queue_counter)

sum (partition_batch_counter_total)

sum (partition_event_counter_total)
```

```
Device_MoCA_Interface_1_Stats_BytesSent_total{SerialNumber=~".*"}
Device_MoCA_Interface_1_Stats_BytesReceived_total{SerialNumber=~".*"}
```
