for i in $(seq 1 100); do curl -s -o /dev/null -d @Report.csv http://localhost:8088/collector; done
for i in $(seq 1 100); do curl -s -o /dev/null -d @Report.json http://localhost:8088/collector; done
