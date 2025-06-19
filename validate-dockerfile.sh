#!/bin/bash

# Script to validate Dockerfile locally
# This script can be run locally to test the Dockerfile before pushing

set -e

echo "🔍 Validating Dockerfile..."

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    exit 1
fi

# Check if Dockerfile exists
if [ ! -f "Dockerfile" ]; then
    echo "❌ Dockerfile not found in current directory"
    exit 1
fi

echo "✅ Docker is available"
echo "✅ Dockerfile found"

# Build the Docker image
echo "🏗️  Building Docker image..."
if docker build -t dockerfile-test:latest .; then
    echo "✅ Docker image built successfully"
else
    echo "❌ Docker build failed"
    exit 1
fi

# Test basic functionality
echo "🧪 Testing Docker image..."

# Test that the image runs
if docker run --rm dockerfile-test:latest /bin/sh -c "echo 'Container started successfully'"; then
    echo "✅ Container runs successfully"
else
    echo "❌ Container failed to run"
    exit 1
fi

# Test that required binaries are present
echo "🔍 Checking required binaries..."
if docker run --rm dockerfile-test:latest /bin/sh -c "which okteto && which jq && which ruby"; then
    echo "✅ All required binaries are present"
else
    echo "❌ Some required binaries are missing"
    exit 1
fi

# Test that entrypoint is executable
if docker run --rm dockerfile-test:latest /bin/sh -c "test -x /entrypoint.sh"; then
    echo "✅ Entrypoint is executable"
else
    echo "❌ Entrypoint is not executable"
    exit 1
fi

# Clean up
echo "🧹 Cleaning up..."
docker rmi dockerfile-test:latest

echo "🎉 All tests passed! Dockerfile is valid."