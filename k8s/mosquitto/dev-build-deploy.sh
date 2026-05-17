#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

REGISTRY="rg.fr-par.scw.cloud/funcscwjeromet1q1hfov"
TAG="${1:-latest}"
PLATFORMS="linux/amd64,linux/arm64"
BUILDER="novoceo-multiarch"

# Crée le builder multi-arch si nécessaire (utilise QEMU via docker-container driver)
docker buildx inspect "$BUILDER" > /dev/null 2>&1 \
  || docker buildx create --name "$BUILDER" --driver docker-container --use
docker buildx use "$BUILDER"

echo "==> Build mosquitto multi-arch ($PLATFORMS) : $TAG"
docker buildx build \
  --platform "$PLATFORMS" \
  --file mosquitto/Dockerfile \
  --tag "$REGISTRY/mosquitto:$TAG" \
  --push \
  .

echo "==> Rollout StatefulSet"
kubectl rollout restart statefulset/mosquitto -n novoceo
kubectl rollout status statefulset/mosquitto -n novoceo
