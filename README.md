# go-k8s-service

Minimal Go monorepo with two microservices for learning Kubernetes deployment patterns.

## Services

| Service | Port | Endpoints |
|---------|------|-----------|
| order-service | 8080 | GET /health, GET /ready, GET /orders, POST /orders, GET /orders/{id} |
| payment-service | 8081 | GET /health, GET /ready, POST /payments, GET /payments/{id} |

## Project Structure

```
go-k8s-service/
├── order-service/
│   ├── main.go
│   ├── main_test.go
│   └── Dockerfile
├── payment-service/
│   ├── main.go
│   ├── main_test.go
│   └── Dockerfile
├── k8s/
│   ├── order-service/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── payment-service/
│       ├── deployment.yaml
│       └── service.yaml
├── postman/
│   └── postman_collection.json
├── docker-compose.yml
└── go.mod
```

## Run Locally

```bash
# single service
go run ./order-service
go run ./payment-service
```

## Test

```bash
go test ./order-service/...
go test ./payment-service/...
```

Import `postman/postman_collection.json` into Postman for manual testing.

## Docker

```bash
# build images
docker build -f order-service/Dockerfile -t abhiraj777/order-service:v2 .
docker build -f payment-service/Dockerfile -t abhiraj777/payment-service:v2 .

# push to Docker Hub
docker push abhiraj777/order-service:v2
docker push abhiraj777/payment-service:v2
```

## Deploy to Kubernetes

```bash
# create a cluster (using Kind)
kind create cluster --name go-k8s

# deploy
kubectl apply -f k8s/order-service/
kubectl apply -f k8s/payment-service/

# verify
kubectl get pods
kubectl get services
```

### K8s Resources per Service

- **Deployment** — 3 replicas, resource limits, liveness + readiness probes
- **Service** — ClusterIP on port 80, forwarding to container port

### Health Checks

| Endpoint | Probe | Behavior |
|----------|-------|----------|
| `/health` | livenessProbe | Always returns 200 |
| `/ready` | readinessProbe | Returns 503 for ~2s after startup, then 200 |

### Access Services

```bash
# port-forward to test from your machine
kubectl port-forward svc/order-service 8080:80
kubectl port-forward svc/payment-service 8081:80

# then
curl localhost:8080/health
curl localhost:8081/ready
```
