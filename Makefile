.PHONY: test build docker deploy run

# Test the application
test:
	go test ./...

# Build the application
build:
	go build -o accounting ./main.go

# Run the application
run: build
	./accounting

# Build the Docker image
docker:
	docker build -t accounting:latest -f build/Dockerfile .

# Deploy the application using Helm
deploy:
	helm install accounting ./deploy/helm/accounting 