# go-k8s-service

A minimal Go monorepo demonstrating Kubernetes deployment patterns.

## Services

- **order-service** — manages orders (CRUD)
- **payment-service** — handles payment processing

## Structure

\```
go-k8s-service/
├── order-service/       Go service + Dockerfile
├── payment-service/     Go service + Dockerfile
├── k8s/                 Kubernetes manifests
└── docker-compose.yml   Local development
\```

## Running locally

\```bash
docker-compose up
\```

## Deploying to Kubernetes

\```bash
kubectl apply -f k8s/
\```
