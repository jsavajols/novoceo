# --- Build stage ---
# Le contexte de build doit être la racine du projet (go.mod s'y trouve).
# docker build -f snug/Dockerfile -t snug .
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Copier les manifestes en premier pour bénéficier du cache Docker sur les dépendances.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produit un binaire entièrement statique, compatible avec distroless.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /snug ./snug/

# --- Final stage ---
# distroless/static : aucun shell, aucun package inutile, surface d'attaque minimale.
# :nonroot évite de faire tourner le process en root dans le pod K8s.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /snug /snug

# En K8s, les flags arrivent via les args: du manifest (broker, button, plug...).
# Ne jamais embarquer de credentials dans l'image.

ENTRYPOINT ["/snug"]
