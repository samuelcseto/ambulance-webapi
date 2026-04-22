FROM openapitools/openapi-generator-cli:v6.6.0 AS api

WORKDIR /local

COPY api api
COPY scripts scripts
COPY .openapi-generator-ignore .openapi-generator-ignore

RUN docker-entrypoint.sh generate -c /local/scripts/generator-cfg.yaml

ENTRYPOINT ["bash"]

############################################
FROM golang:latest AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal/ internal/
COPY cmd/ cmd/
COPY --from=api /local/ ./

RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
      -o ./ambulance-webapi-srv ./cmd/ambulance-api-service

############################################
FROM scratch

LABEL org.opencontainers.image.authors="Samuel Cseto"
LABEL org.opencontainers.image.title="Ambulance Waiting List WebAPI Service"
LABEL org.opencontainers.image.description="WEBAPI for managing entries in ambulances' waiting list"

ENV AMBULANCE_API_ENVIRONMENT=production
ENV AMBULANCE_API_PORT=8080
ENV AMBULANCE_API_MONGODB_HOST=mongo
ENV AMBULANCE_API_MONGODB_PORT=27017
ENV AMBULANCE_API_MONGODB_DATABASE=scs-ambulance-wl
ENV AMBULANCE_API_MONGODB_COLLECTION=ambulance
ENV AMBULANCE_API_MONGODB_USERNAME=root
ENV AMBULANCE_API_MONGODB_PASSWORD=
ENV AMBULANCE_API_MONGODB_TIMEOUT_SECONDS=5

COPY --from=build /app/ambulance-webapi-srv ./

EXPOSE 8080
ENTRYPOINT ["./ambulance-webapi-srv"]
