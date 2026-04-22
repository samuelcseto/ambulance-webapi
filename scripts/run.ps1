param (
    $command
)

if (-not $command) {
    $command = "start"
}

$ProjectRoot = "${PSScriptRoot}/.."

$env:AMBULANCE_API_ENVIRONMENT = "Development"
$env:AMBULANCE_API_PORT = "8080"
$env:AMBULANCE_API_MONGODB_USERNAME = "root"
$env:AMBULANCE_API_MONGODB_PASSWORD = "neUhaDnes"

function mongo {
    docker compose --file ${ProjectRoot}/deployments/docker-compose/compose.yaml $args
}

switch ($command) {
    "start" {
        try {
            mongo up --detach
            go run ${ProjectRoot}/cmd/ambulance-api-service
        }
        finally {
            mongo down
        }
    }
    "mongo" {
        mongo up
    }
    "openapi" {
        docker run --rm -v ${ProjectRoot}:/local openapitools/openapi-generator-cli:v6.6.0 generate -c /local/scripts/generator-cfg.yaml
    }
    "openapi-templates" {
        docker run --rm -v ${ProjectRoot}:/local openapitools/openapi-generator-cli:v6.6.0 author template --generator-name go-gin-server --output /local/scripts/templates
    }
    default {
        throw "Unknown command: $command"
    }
}
