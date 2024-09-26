FROM golang as builder

# Set destination for COPY
WORKDIR /workspace

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux OARCH=amd64 go build -o dgame cmd/dgame/main.go

# Create a new image for the application code to run in
FROM alpine
LABEL org.label-schema.vendor="rbrabson" \
  org.label-schema.name="dgame bot" \
  org.label-schema.description="Deploy the dgame bot" \
  org.label-schema.vcs-ref=$VCS_REF \
  org.label-schema.vcs-url=$VCS_URL \
  org.label-schema.license="BSD-3-Clause license" \
  org.label-schema.schema-version="1.0" \
  name="dgame-bot" \
  vendor="rbrabson" \
  description="Deploy the dgame bot" \
  summary="Deploy the dgame bot"

RUN mkdir -p /licenses
ADD LICENSE /licenses

WORKDIR /

COPY /store/ /store/
RUN chmod -R 777 /store

COPY /configs/ /configs/
RUN chmod -R 777 /configs

COPY --from=builder /workspace/dgame /

RUN apk add iputils \
    openssh \
    which \
    vim

USER 65532:65532

# Run
CMD ["/dgame"]