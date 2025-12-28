# Go API Starter - Dev Container
FROM mcr.microsoft.com/devcontainers/go:1.22-bookworm

# Install additional tools
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    postgresql-client \
    jq \
    curl \
    make \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install github.com/go-task/task/v3/cmd/task@latest \
    && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
    && go install github.com/air-verse/air@latest \
    && go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install github.com/vektra/mockery/v2@latest

# Pre-commit
RUN pip3 install pre-commit

# Set working directory
WORKDIR /workspace

# Keep container running
CMD ["sleep", "infinity"]
