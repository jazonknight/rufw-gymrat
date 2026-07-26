#!/bin/bash
# ==============================================================================
# RupertFrameworks: GymRat Local Build & Run Script
# Compiles the Go binary and starts the local server.
# ==============================================================================

set -e

PORT="${PORT:-8080}"
BINARY_NAME="gymrat_local"

echo "🏋️‍♂️  [GymRat] Building Go executable binary..."
go build -o "${BINARY_NAME}" main.go

echo "✅ [GymRat] Build successful!"
echo "🚀 [GymRat] Starting GymRat REST server on http://localhost:${PORT}"
echo "--------------------------------------------------------"

./"${BINARY_NAME}" -server -port "${PORT}"
