FROM mcr.microsoft.com/devcontainers/go:1-bookworm

# sigrok-cli is the core runtime dependency (sole interface to sigrok hardware).
# yamllint is needed for YAML linting in CI and editor hooks.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        sigrok-cli \
        usbutils \
        yamllint \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# golangci-lint: pinned version — update manually or via Dependabot custom config.
# Uses pipefail so a curl failure is not masked by the pipe to sh.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSfL https://golangci-lint.run/install.sh \
    | sh -s -- -b /usr/local/bin v2.8.0
SHELL ["/bin/sh", "-c"]

# Smoke-test: fail the build early if any tool is missing or broken
RUN sigrok-cli --version \
    && lsusb --version \
    && yamllint --version \
    && golangci-lint version
