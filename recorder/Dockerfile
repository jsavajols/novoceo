# --- Build stage ---
# Le contexte de build doit être la racine du projet (go.mod s'y trouve).
# docker build -f recorder/Dockerfile -t recorder .
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Copier les manifestes en premier pour bénéficier du cache Docker sur les dépendances.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produit un binaire entièrement statique, compatible avec distroless.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /recorder ./recorder/

# --- Final stage ---
# distroless/static : aucun shell, aucun package inutile, surface d'attaque minimale.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /recorder /recorder

# En K8s, les credentials arrivent via les variables d'environnement du Secret.
# Ne jamais monter le fichier .env dans l'image.

ENTRYPOINT ["/recorder"]
