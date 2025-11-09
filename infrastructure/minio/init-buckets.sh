#!/bin/sh
set -e

echo "Waiting for MinIO to be ready..."
sleep 10

export MC_HOST_minio=http://minioadmin:minioadmin@minio:9000

echo "Creating buckets..."

mc mb minio/public --ignore-existing || echo "Bucket 'public' already exists"
mc mb minio/tracks --ignore-existing || echo "Bucket 'tracks' already exists"

echo "Setting bucket policies..."

mc anonymous set public minio/public || echo "Failed to set public policy for 'public' bucket"
mc anonymous set public minio/tracks || echo "Failed to set public policy for 'tracks' bucket"

echo "Created buckets:"
mc ls minio

echo "Bucket initialization completed!"