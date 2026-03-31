FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
RUN apk --no-cache add git
RUN CGO_ENABLED=0 go install go.k6.io/xk6/cmd/xk6@latest
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} xk6 build \
    --output /k6 \
    --with github.com/Dynatrace/xk6-output-dynatrace=.

FROM alpine:3.20
RUN apk --no-cache add ca-certificates && \
    adduser -D -u 12345 k6
COPY --from=builder /k6 /usr/bin/k6
USER k6
WORKDIR /home/k6
ENTRYPOINT ["k6"]
