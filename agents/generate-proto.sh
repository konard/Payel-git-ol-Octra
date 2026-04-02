#!/bin/bash
# Generate Go code from proto files

PROTO_DIR="proto"
OUTPUT_DIR="internal/fetcher/grpc"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR/agentspb"

# Generate Go code
protoc \
  -I"$PROTO_DIR" \
  --go_out="$OUTPUT_DIR" \
  --go-grpc_out="$OUTPUT_DIR" \
  "$PROTO_DIR/agents.proto"

echo "✅ Proto files generated successfully"
