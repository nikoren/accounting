# Accounting Project

## Overview
This project is a Go-based microservice that handles document and page management for accounting purposes. It provides APIs for creating, deleting, and managing documents and pages within a split.

## Features
- **Authentication**: Secure login and access control.
- **Document Management**: Create, delete, and manage documents.
- **Page Management**: Move pages between documents and manage page assignments.
- **Split Operations**: Finalize splits and manage split-related operations.
- **Metrics**: Monitor application performance and health.

## Getting Started

### Prerequisites
- Go 1.24 or higher
- SQLite (or your preferred database)
- Docker (optional, for containerization)

### Installation
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd accounting
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up the database:
   - Ensure your database is configured in the application settings.
   - Run the initial migration script to set up the database schema.

### Building the Application
1. Build the binary:
   ```bash
   make build
   ```

2. The binary will be created in the `bin` directory.

### Running the Application
1. Start the server:
   ```bash
   make run
   ```

2. The server will start on `http://localhost:8081`.

## API Usage
- **Authentication**: Use the `/auth/login` endpoint to authenticate.
- **Document Operations**: Use the `/documents` endpoint to create and delete documents.
- **Page Operations**: Use the `/pages/move` endpoint to move pages between documents.
- **Split Operations**: Use the `/splits/{splitID}/finalize` endpoint to finalize a split.

## Testing
### Running Tests
1. Run unit tests:
   ```bash
   make test
   ```

2. Run integration tests:
   ```bash
   make test-integration
   ```

### End-to-End Testing
- The integration tests cover end-to-end scenarios, including authentication, document management, and error handling.
- Ensure the server is running before executing the integration tests.

## Deployment
### Using Docker
1. Build the Docker image:
   ```bash
   make docker
   ```

2. Run the Docker container:
   ```bash
   make deploy 
   ```


## Contributing
- Fork the repository.
- Create a feature branch: `git checkout -b feature-name`.
- Commit your changes: `git commit -m 'Add some feature'`.
- Push to the branch: `git push origin feature-name`.
- Submit a pull request.
