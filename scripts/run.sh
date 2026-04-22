#!/usr/bin/env bash
set -euo pipefail

command=${1:-start}
ProjectRoot="$(cd "$(dirname "$0")/.." && pwd)"

export AMBULANCE_API_ENVIRONMENT="Development"
export AMBULANCE_API_PORT="8080"
export AMBULANCE_API_MONGODB_USERNAME="root"
export AMBULANCE_API_MONGODB_PASSWORD="neUhaDnes"

mongo() {
    docker compose --file "${ProjectRoot}/deployments/docker-compose/compose.yaml" "$@"
}

case "$command" in
    start)
        trap 'mongo down' EXIT
        mongo up --detach
        go run "${ProjectRoot}/cmd/ambulance-api-service"
        ;;
    mongo)
        mongo up
        ;;
    openapi)
        docker run --rm -v "${ProjectRoot}:/local" openapitools/openapi-generator-cli:v6.6.0 \
            generate -c /local/scripts/generator-cfg.yaml
        ;;
    openapi-templates)
        docker run --rm -v "${ProjectRoot}:/local" openapitools/openapi-generator-cli:v6.6.0 \
            author template --generator-name go-gin-server --output /local/scripts/templates
        ;;
    *)
        echo "Unknown command: $command" >&2
        exit 1
        ;;
esac
