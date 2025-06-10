#!/bin/bash

# Kill any existing process on port 8081
lsof -ti:8081 | xargs kill -9 2>/dev/null

# Set environment variables
export APP_DB_PATH=test_accounting.db
export API_BASE_URL=http://localhost:8081
export APP_PORT=8081
export APP_LOG_LEVEL=debug
export APP_USERS="test:test"

# Clean up test database
rm -f test_accounting.db

# Start the server in the background
go run main.go &
SERVER_PID=$!

# Wait for server to start and apply migrations
sleep 2

# Run setup script to seed data
go run test/scripts/setup_db.go


sleep 2
# Run integration tests
go test -v ./test/integration

# Clean up
kill $SERVER_PID
rm -f test_accounting.db 