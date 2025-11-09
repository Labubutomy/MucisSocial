#!/bin/bash

echo "Waiting for Redpanda to be ready..."
sleep 10

echo "Creating topics..."

rpk topic create transcoder-tasks \
    --brokers redpanda:9092 \
    --partitions 3 \
    --replicas 1 \
    --topic-config retention.ms=604800000 \
    --topic-config compression.type=snappy

echo "Topics created successfully!"

rpk topic list --brokers redpanda:9092

