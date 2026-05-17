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

echo "==> Build snug multi-arch ($PLATFORMS) : $TAG"
docker buildx build \
  --platform "$PLATFORMS" \
  --file snug/Dockerfile \
  --tag "$REGISTRY/snug:$TAG" \
  --push \
  .

echo "==> Redémarrage"
kubectl rollout restart deployment/snug -n novoceo
kubectl rollout status deployment/snug -n novoceo
