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

# Copy prebuild script
COPY prebuild-themes.sh /tmp/prebuild-themes.sh
RUN chmod +x /tmp/prebuild-themes.sh

# Prebuild all themes (this happens at build time, not runtime)
RUN /tmp/prebuild-themes.sh && rm /tmp/prebuild-themes.sh

# Expose webhook port
EXPOSE 8080

# Set working directory
WORKDIR /workspace

# Run gitdocai in preview mode (Cloud)
ENTRYPOINT ["gitdocai-cli", "preview"]