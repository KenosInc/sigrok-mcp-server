FROM mcr.microsoft.com/devcontainers/go:1-bookworm

# sigrok has no release containing the kingst-la2016 driver, so it is
# built from source. Clone over HTTPS (git:// is unauthenticated and
# unencrypted) and pin each repository to a commit so the build is
# reproducible and cannot pick up an unreviewed upstream change.
ARG LIBSIGROK_REF=0bc2487778e660f4d3116729b6f4aee2b1996bb0
ARG LIBSIGROKDECODE_REF=71f451443029322d57376214c330b518efd84f88
ARG SIGROK_CLI_REF=f44dd91347e7ac797cefc23162b9fcf0b7329f1f
# Build dependencies for sigrok (libsigrok, libsigrokdecode, sigrok-cli).
# The Debian bookworm package ships libsigrok 0.5.2, which lacks the
# kingst-la2016 driver (added March 2020, unreleased). We build from git master.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        ca-certificates git gcc g++ make autoconf autoconf-archive automake libtool pkg-config \
        libglib2.0-dev libzip-dev libusb-1.0-0-dev libftdi1-dev libhidapi-dev \
        libieee1284-3-dev libserialport-dev \
        python3-dev \
        usbutils \
        yamllint \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Build libsigrok from git (includes kingst-la2016 driver)
RUN git init -q /tmp/libsigrok \
    && cd /tmp/libsigrok \
    && git fetch -q --depth 1 https://github.com/sigrokproject/libsigrok.git "${LIBSIGROK_REF}" \
    && git checkout -q FETCH_HEAD \
    && ./autogen.sh \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && ldconfig \
    && rm -rf /tmp/libsigrok

# Build libsigrokdecode from git
RUN git init -q /tmp/libsigrokdecode \
    && cd /tmp/libsigrokdecode \
    && git fetch -q --depth 1 https://github.com/sigrokproject/libsigrokdecode.git "${LIBSIGROKDECODE_REF}" \
    && git checkout -q FETCH_HEAD \
    && ./autogen.sh \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && ldconfig \
    && rm -rf /tmp/libsigrokdecode

# Build sigrok-cli from git
RUN git init -q /tmp/sigrok-cli \
    && cd /tmp/sigrok-cli \
    && git fetch -q --depth 1 https://github.com/sigrokproject/sigrok-cli.git "${SIGROK_CLI_REF}" \
    && git checkout -q FETCH_HEAD \
    && ./autogen.sh \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && rm -rf /tmp/sigrok-cli

# golangci-lint: pinned version — update manually or via Dependabot custom config.
# Uses pipefail so a curl failure is not masked by the pipe to sh.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sSfL https://golangci-lint.run/install.sh \
    | sh -s -- -b /usr/local/bin v2.8.0
SHELL ["/bin/sh", "-c"]

# Smoke-test: fail the build early if any tool is missing or broken
RUN sigrok-cli --version \
    && sigrok-cli --list-supported | grep -q kingst \
    && lsusb --version \
    && yamllint --version \
    && golangci-lint version
