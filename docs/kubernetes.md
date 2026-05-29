# Déploiement Kubernetes (k3s)

Cluster k3s sur `100.64.0.10`, namespace `novoceo`.

## Structure des manifests

```
k8s/
  namespace.yaml
  mosquitto/
    configmap.yaml      # mosquitto.conf (référence, non monté — config générée par l'entrypoint)
    pvc.yaml            # service headless pour DNS inter-pods (StatefulSet)
    deployment.yaml     # StatefulSet 2 replicas — chaque pod a son propre PVC
    service.yaml        # NodePort 32500 -> 1883
    dev-build-deploy.sh # build multi-arch (amd64 + arm64) + rollout
  snug/
    deployment.yaml     # client MQTT sortant, 1 replica, pas de Service
    dev-build-deploy.sh
  recorder/
    configmap.yaml      # MQTT_BROKER, DB_HOST/PORT/NAME
    secret.yaml         # DB_USER, DB_PASSWORD (base64) — ne pas commiter
    deployment.yaml     # 1 replica — singleton writer MQTT -> PostgreSQL
    dev-build-deploy.sh
  api/
    configmap.yaml      # MQTT_BROKER, DB_HOST/PORT/NAME, API_PORT
    secret.yaml         # DB_USER, DB_PASSWORD, API_TOKEN (base64)
    deployment.yaml     # 2 replicas actif/actif — stateless
    service.yaml
    ingress.yaml        # novoceo.api.local.happyapi.fr — TLS Let's Encrypt
    dev-build-deploy.sh
  front/
    configmap.yaml      # API_URL, FRONT_PORT
    secret.yaml         # API_TOKEN (base64)
    deployment.yaml     # 2 replicas actif/actif — stateless
    service.yaml
    ingress.yaml        # novoceo.front.local.happyapi.fr — TLS Let's Encrypt
    dev-build-deploy.sh
```

## Registry d'images

`rg.fr-par.scw.cloud/funcscwjeromet1q1hfov/`

| Image | Tag | Architectures | Replicas | Mode |
|-------|-----|---------------|----------|------|
| `mosquitto` | `latest` | amd64 + arm64 | 2 | StatefulSet + bridge |
| `novoceo-api` | `dev` | amd64 + arm64 | 2 | actif/actif |
| `novoceo-front` | `dev` | amd64 + arm64 | 2 | actif/actif |
| `recorder` | `dev` | amd64 + arm64 | 1 | singleton — restart auto |
| `snug` | `dev` | amd64 + arm64 | 1 | singleton — restart auto |

`recorder` et `snug` restent à 1 replica : ce sont des consommateurs MQTT avec effets de bord
(écriture DB, commande Zigbee). 2 replicas provoquerait des doublons. Kubernetes redémarre
automatiquement le pod en cas de crash (`restartPolicy: Always`).

## Déploiement initial complet

```bash
# 1. Namespace
kubectl apply -f k8s/namespace.yaml

# 2. Mosquitto (broker HA — StatefulSet 2 pods)
# Prérequis : builder multi-arch disponible
./k8s/mosquitto/dev-build-deploy.sh   # build + push + rollout
kubectl apply -f k8s/mosquitto/pvc.yaml        # headless service en premier
kubectl apply -f k8s/mosquitto/deployment.yaml # puis le StatefulSet
kubectl apply -f k8s/mosquitto/service.yaml
kubectl apply -f k8s/mosquitto/configmap.yaml  # référence, non monté

# 3. Recorder
kubectl apply -f k8s/recorder/configmap.yaml
kubectl apply -f k8s/recorder/secret.yaml    # contient les credentials DB
kubectl apply -f k8s/recorder/deployment.yaml

# 4. API
kubectl apply -f k8s/api/configmap.yaml
kubectl apply -f k8s/api/secret.yaml
kubectl apply -f k8s/api/

# 5. Front
kubectl apply -f k8s/front/configmap.yaml
kubectl apply -f k8s/front/secret.yaml
kubectl apply -f k8s/front/

# 6. Snug
./k8s/snug/dev-build-deploy.sh
kubectl apply -f k8s/snug/deployment.yaml
```

## Scripts de build/deploy rapide

Tous les services utilisent `docker buildx` pour produire des images multi-arch `linux/amd64,linux/arm64`.
Le builder `novoceo-multiarch` (driver `docker-container`) est créé automatiquement à la première exécution.

```bash
# Build + push + rollout restart
./k8s/api/dev-build-deploy.sh           # tag: dev (défaut)
./k8s/api/dev-build-deploy.sh v1.2.3   # tag custom

./k8s/front/dev-build-deploy.sh
./k8s/recorder/dev-build-deploy.sh
./k8s/snug/dev-build-deploy.sh

# Mosquitto multi-arch (amd64 + arm64)
./k8s/mosquitto/dev-build-deploy.sh          # tag: latest (défaut)
./k8s/mosquitto/dev-build-deploy.sh v1.2.3  # tag custom
```

## Opérations courantes

```bash
# Statut global
kubectl get all -n novoceo

# Logs en temps réel
kubectl logs -n novoceo -l app=mosquitto -f
kubectl logs -n novoceo -l app=recorder -f
kubectl logs -n novoceo -l app=api -f
kubectl logs -n novoceo -l app=front -f
kubectl logs -n novoceo -l app=snug -f

# Tester le broker depuis le réseau local
mosquitto_pub -h 100.64.0.10 -p 32500 -t test -m "hello"
mosquitto_sub -h 100.64.0.10 -p 32500 -t "#"

# Supprimer tout
kubectl delete namespace novoceo
```

## Ingress et TLS

L'ingress controller est nginx avec cert-manager (issuer `letsencrypt`) :

| Service | URL |
|---------|-----|
| front | `https://novoceo.front.local.happyapi.fr` |
| api | `https://novoceo.api.local.happyapi.fr` |

## Ressources allouées

| Pod | Replicas | CPU req/limit | Mémoire req/limit |
|-----|----------|--------------|------------------|
| mosquitto | 2 | 50m / 200m | 32Mi / 128Mi |
| recorder | 1 | 50m / 200m | 32Mi / 128Mi |
| api | 2 | 50m / 200m | 32Mi / 128Mi |
| front | 2 | 20m / 100m | 16Mi / 64Mi |
| snug | 1 | 10m / 100m | 16Mi / 64Mi |

## Sécurité

Les fichiers `secret.yaml` contiennent les credentials encodés en base64.
Ils ne doivent pas être commités dans un repo public.
En prod, les injecter via Sealed Secrets ou un vault externe.
