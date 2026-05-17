# docker build -f api/Dockerfile -t novoceo-api .
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /api ./api/

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /api /api

EXPOSE 5000

ENTRYPOINT ["/api"]
