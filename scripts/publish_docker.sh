#!/bin/bash
# ==============================================================================
# RupertFrameworks: GymRat Docker Multi-Arch Build & Publish Script
# Builds multi-platform Docker images (linux/amd64, linux/arm64) and pushes to Docker Hub.
# ==============================================================================

set -e

DOCKER_USER="${DOCKER_USER:-jazonknight}"
IMAGE_NAME="rufw-gymrat"
VERSION="${1:-v2.0.7}"

echo "📦 [GymRat Docker Publish] Target Registry: ${DOCKER_USER}/${IMAGE_NAME}"
echo "🏷️  [GymRat Docker Publish] Tagging Version: ${VERSION} and latest"
echo "--------------------------------------------------------"

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t "${DOCKER_USER}/${IMAGE_NAME}:latest" \
  -t "${DOCKER_USER}/${IMAGE_NAME}:${VERSION}" \
  --push .

echo "--------------------------------------------------------"
echo "🎉 [GymRat Docker Publish] Successfully published ${DOCKER_USER}/${IMAGE_NAME}:${VERSION} to Docker Hub!"
