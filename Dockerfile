FROM ubuntu:24.04

# Build argument for gitdocai-cli version
ARG CLI_VERSION=latest

# Install system dependencies
RUN apt-get update && \
    apt-get install -y curl ca-certificates git && \
    update-ca-certificates && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install gitdocai-cli with specified version
RUN if [ "$CLI_VERSION" = "latest" ]; then \
        npm install -g gitdocai-cli; \
    else \
        npm install -g gitdocai-cli@${CLI_VERSION}; \
    fi

# Expose webhook port
EXPOSE 8080

# Set working directory
WORKDIR /workspace

# Run gitdocai in production mode
ENTRYPOINT ["gitdocai-cli", "prod"]