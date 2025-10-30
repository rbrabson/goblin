FROM alpine AS builder

# Install make and go
RUN apk add make go

# Set destination for COPY
WORKDIR /workspace

# Copy the source code
COPY ./ ./

# Download Go modules
RUN go mod download

# Build the goblin binary
RUN make build-linux

# Create a new image for the application code to run in
FROM alpine
LABEL org.label-schema.vendor="rbrabson" \
  org.label-schema.name="goblin bot" \
  org.label-schema.description="Deploy the goblin bot" \
  org.label-schema.vcs-url=https://github.com/rbrabson/goblin.git \
  org.label-schema.license="BSD-3-Clause license" \
  org.label-schema.schema-version="1.0" \
  name="goblin-bot" \
  vendor="rbrabson" \
  description="Deploy the goblin bot" \
  summary="Deploy the goblin bot"

RUN mkdir -p /licenses
ADD LICENSE /licenses

RUN mkdir -p /config
ADD config /config/

WORKDIR /

COPY --from=builder /workspace/bin/linux/amd64/goblin /

# Uncomment this out if you are using a .env file for configuration instead of environment variables in the docker compose file
ADD .env .

RUN apk add iputils \
  bash \
  openssh \
  which \
  vim

USER 65532:65532

# Run
CMD ["/goblin"]