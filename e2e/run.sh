#!/bin/bash
set -e

ETCDCTL_API=3 etcdctl del --prefix / --endpoints 172.100.0.40:2379

FUSIS_HOST='172.100.0.2'
PORT=8000

/app/wait-for -t 60 $FUSIS_HOST:$PORT -- echo "Fusis API is up"

echo "==> Adding service-1"
curl -X POST \
  http://$FUSIS_HOST:$PORT/services \
  -H 'content-type: application/json' \
  -d '{
    "name": "service-1",
    "address": "10.100.0.10",
    "mode": "nat",
    "port": 80,
    "protocol": "tcp",
    "scheduler": "rr"
  }'
echo ""

echo "==> Adding service-2"
curl -X POST \
  http://$FUSIS_HOST:$PORT/services \
  -H 'content-type: application/json' \
  -d '{
    "name": "service-2",
    "mode": "tunnel",
    "port": 8080,
    "protocol": "tcp",
    "scheduler": "lc"
  }'
echo ""

printf "==> Assert services"
services_resp=$(curl -s http://$FUSIS_HOST:$PORT/services)
expected_resp='[{"Name":"service-1","Address":"10.100.0.10","Port":80,"Protocol":"tcp","Scheduler":"rr","Mode":"nat","Persistent":0},{"Name":"service-2","Address":"10.100.0.1","Port":8080,"Protocol":"tcp","Scheduler":"lc","Mode":"tunnel","Persistent":0}]'
if [ $services_resp != $expected_resp ]; then
  echo ""
  echo "[fail] Services response is wrong"
  echo "[fail] Services response => $services_resp"
  exit 2
fi
echo "...OK"

echo "==> Adding destination-1"
curl -X POST \
  http://$FUSIS_HOST:$PORT/services/service-1/destinations \
  -H 'content-type: application/json' \
  -d '{
    "name": "dest-1",
    "address": "172.100.0.50",
    "port": 80
  }'
echo ""

echo "==> Adding destination-2"
curl -X POST \
  http://$FUSIS_HOST:$PORT/services/service-1/destinations \
  -H 'content-type: application/json' \
  -d '{
    "name": "dest-2",
    "address": "172.100.0.60",
    "port": 80
  }'
echo ""

printf "==> Assert destinations"
service_resp=$(curl -s http://$FUSIS_HOST:$PORT/services/service-1)
expected_resp='{"Address":"10.100.0.10","Destinations":[{"Name":"dest-1","Address":"172.100.0.50","Port":80,"Weight":1,"Mode":"nat","ServiceId":"service-1"},{"Name":"dest-2","Address":"172.100.0.60","Port":80,"Weight":1,"Mode":"nat","ServiceId":"service-1"}],"Mode":"nat","Name":"service-1","Persistent":0,"Port":80,"Protocol":"tcp","Scheduler":"rr"}'
if [ $service_resp != $expected_resp ]; then
  echo ""
  echo "[fail] Service response is wrong"
  echo "[fail] Service response => $service_resp"
  exit 2
fi
echo "...OK"


printf "==> Assert vip requests"
vip_resp1=$(curl -s http://10.100.0.10)
vip_resp2=$(curl -s http://10.100.0.10)

expected_resp="Welcome to nginx"
if [[ $vip_resp1 != *$expected_resp* ]] || [[ $vip_resp2 != *$expected_resp* ]]; then
  echo ""
  echo "[fail] VIP response is wrong"
  echo "[fail] VIP response => $vip_resp"
  exit 2
fi
echo "...OK"
