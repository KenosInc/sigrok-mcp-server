# Stage 1: Build Go binary
FROM golang:1.25-bookworm AS go-builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /sigrok-mcp-server ./cmd/sigrok-mcp-server

# Stage 2: Build sigrok from git (kingst-la2016 driver requires unreleased libsigrok > 0.5.2)
FROM debian:bookworm AS sigrok-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        git gcc g++ make autoconf autoconf-archive automake libtool pkg-config \
        libglib2.0-dev libzip-dev libusb-1.0-0-dev libftdi1-dev libhidapi-dev \
        libieee1284-3-dev libserialport-dev \
        python3-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone --depth 1 git://sigrok.org/libsigrok /tmp/libsigrok \
    && cd /tmp/libsigrok \
    && ./autogen.sh \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && ldconfig \
    && rm -rf /tmp/libsigrok

RUN git clone --depth 1 git://sigrok.org/libsigrokdecode /tmp/libsigrokdecode \
    && cd /tmp/libsigrokdecode \
    && ./autogen.sh \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && ldconfig \
    && rm -rf /tmp/libsigrokdecode

RUN git clone --depth 1 git://sigrok.org/sigrok-cli /tmp/sigrok-cli \
    && cd /tmp/sigrok-cli \
    && ./autogen.sh \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && rm -rf /tmp/sigrok-cli

# Stage 3: Runtime
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libglib2.0-0 libzip4 libusb-1.0-0 libftdi1-2 \
        libhidapi-libusb0 libhidapi-hidraw0 \
        libieee1284-3 libserialport0 \
        python3 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy sigrok binaries, libraries, and decoders from builder.
# Use tar via sh -c to preserve symlinks (COPY dereferences them, breaking ldconfig).
COPY --from=sigrok-builder /usr/local/bin/sigrok-cli /usr/local/bin/sigrok-cli
RUN --mount=from=sigrok-builder,source=/usr/local/lib,target=/sigrok-lib \
    sh -c 'cd /sigrok-lib && tar cf - libsigrok.so* libsigrokdecode.so* | tar xf - -C /usr/local/lib'
COPY --from=sigrok-builder /usr/local/share/libsigrokdecode/ /usr/local/share/libsigrokdecode/
RUN ldconfig

# Copy Go binary from builder
COPY --from=go-builder /sigrok-mcp-server /usr/local/bin/sigrok-mcp-server

# Smoke-test: verify sigrok-cli works and includes kingst driver
RUN sigrok-cli --version \
    && sigrok-cli --list-supported | grep -q kingst

ENTRYPOINT ["sigrok-mcp-server"]
