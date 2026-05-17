#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

REGISTRY="rg.fr-par.scw.cloud/funcscwjeromet1q1hfov"
TAG="${1:-dev}"
PLATFORMS="linux/amd64,linux/arm64"
BUILDER="novoceo-multiarch"

docker buildx inspect "$BUILDER" > /dev/null 2>&1 \
  || docker buildx create --name "$BUILDER" --driver docker-container --use
docker buildx use "$BUILDER"

echo "==> Build novoceo-front multi-arch ($PLATFORMS) : $TAG"
docker buildx build \
  --platform "$PLATFORMS" \
  --file front/Dockerfile \
  --tag "$REGISTRY/novoceo-front:$TAG" \
  --push \
  .

echo "==> Redémarrage"
kubectl rollout restart deployment/front -n novoceo
kubectl rollout status deployment/front -n novoceo
