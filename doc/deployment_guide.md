# Deployment Guide for Accounting Application

This guide provides step-by-step instructions to build the Docker image and deploy the Accounting application using Helm.

## Prerequisites

- Docker installed and running
- Kubernetes cluster (e.g., Minikube, GKE, EKS)
- Helm installed

## Step 1: Build the Docker Image

From the root of the repository, run the following command to build the Docker image:

```bash
docker build -t accounting:latest -f build/Dockerfile .
```

This command builds the image and tags it as `accounting:latest`.

## Step 2: Deploy Using Helm

Once the Docker image is built, deploy the application using Helm:

```bash
helm install accounting ./deploy/helm/accounting
```

This command installs the Helm chart located in `./deploy/helm/accounting` and names the release `accounting`.

## Step 3: Verify the Deployment

Check the status of the deployed resources:

```bash
kubectl get pods
kubectl get services
```

Ensure that the pod is running and the service is accessible.

## Step 4: Access the Application

If you're running this locally (e.g., using Minikube), you can access the application by port-forwarding the service:

```bash
kubectl port-forward svc/accounting 8080:8080
```

Then, open your browser and go to `http://localhost:8080`.

## Additional Configuration

You can customize the deployment by modifying the `values.yaml` file in the `deploy/helm/accounting` directory or by using the `--set` flag with the `helm install` command.

For example, to change the port:

```bash
helm install accounting ./deploy/helm/accounting --set env.APP_PORT=9090
```

## Troubleshooting

If you encounter any issues during deployment, check the logs of the running pods:

```bash
kubectl logs <pod-name>
```

Replace `<pod-name>` with the actual name of the pod. 