FROM mcr.microsoft.com/devcontainers/go:1-bookworm

# Install sigrok-cli and development tools
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        sigrok-cli \
        yamllint \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install golangci-lint
RUN curl -sSfL https://golangci-lint.run/install.sh \
    | sh -s -- -b /usr/local/bin v2.8.0

# Verify installations
RUN sigrok-cli --version \
    && yamllint --version \
    && golangci-lint version
