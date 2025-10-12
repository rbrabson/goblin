FROM golang AS builder

# Set destination for COPY
WORKDIR /workspace

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY ./ ./
# COPY cmd/ cmd/
# COPY internal/ internal/
# COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux OARCH=amd64 go build -o goblin cmd/goblin/main.go

# Create a new image for the application code to run in
FROM alpine
LABEL org.label-schema.vendor="rbrabson" \
  org.label-schema.name="goblin bot" \
  org.label-schema.description="Deploy the goblin bot" \
  org.label-schema.vcs-ref=$VCS_REF \
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

COPY --from=builder /workspace/goblin /

RUN apk add iputils \
  bash \
  openssh \
  which \
  vim

USER 65532:65532

# Run
CMD ["/goblin"]