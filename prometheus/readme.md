Docker Compose
```
docker compose up -d
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
