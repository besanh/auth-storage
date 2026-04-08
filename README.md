# Auth Storage Server

## Documentation
Comprehensive documentation is available in the [docs/](file:///Users/anhle/Downloads/BesAnh/k8s-practice/auth_storage/server/docs) folder:
- [Architecture Overview](file:///Users/anhle/Downloads/BesAnh/k8s-practice/auth_storage/server/docs/architecture.md)
- [Sequence Flows](file:///Users/anhle/Downloads/BesAnh/k8s-practice/auth_storage/server/docs/flows.md)
- [Integration Guide](file:///Users/anhle/Downloads/BesAnh/k8s-practice/auth_storage/server/docs/integration_guide.md)
- [OpenAPI Specification](file:///Users/anhle/Downloads/BesAnh/k8s-practice/auth_storage/server/docs/openapi.yaml)
- [Development Guide](file:///Users/anhle/Downloads/BesAnh/k8s-practice/auth_storage/server/docs/development_guide.md)

## Infrastructure Setup

This project requires **Postgres**, **Redis**, **SpiceDB**, and **Vault**. You can start all dependencies using Docker Compose:

```bash
docker compose up -d
```

### 🔐 Multi-Step Vault Initialization
Vault is configured in **Persistent Server Mode**. Follow these steps for first-time setup:

1.  **Initialize**: `make vault-init` (Save the 5 Unseal Keys and Root Token).
2.  **Unseal**: Run `make vault-unseal` **3 times** with different keys.
3.  **Configure Environment**: Update `.vault.env` with your new root token.
4.  **Finish Init**: Run `make init-vault` to enable the KV-V2 engine.

**Note**: You must run `make vault-unseal` (3 keys) every time the container restarts.

## UI Dashboards
*   **Vault UI**: [http://localhost:8200](http://localhost:8200) (Use Token login)
*   **Swagger Docs**: Run `make swagger` and visit [http://localhost:8080](http://localhost:8080)


---

# Kratos Project Template

## Install Kratos
```
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
```
## Create a service
```
# Create a template project
kratos new server

cd server
# Add a proto template
kratos proto add api/server/server.proto
# Generate the proto code
kratos proto client api/server/server.proto
# Generate the source code of service by proto file
kratos proto server api/server/server.proto -t internal/service

go generate ./...
go build -o ./bin/ ./...
./bin/server -conf ./configs
```
## Generate other auxiliary files by Makefile
```
# Download and update dependencies
make init
# Generate API files (include: pb.go, http, grpc, validate, swagger) by proto file
make api
# Generate all files
make all
```
## Automated Initialization (wire)
```
# install wire
go get github.com/google/wire/cmd/wire

# generate wire
cd cmd/server
wire
```

## Docker
```bash
# build
docker build -t <your-docker-image-name> .

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```

