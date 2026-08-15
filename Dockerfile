# syntax=docker/dockerfile:1

FROM golang:1.26.2-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SOURCE_COMMIT=unavailable
RUN BUILD_REVISION="${SOURCE_COMMIT}"; \
    if [ -z "${BUILD_REVISION}" ] || [ "${BUILD_REVISION}" = "unavailable" ]; then \
        BUILD_REVISION="$(find assets -type f -exec sha256sum {} + | sort | sha256sum | cut -c1-12)"; \
    fi; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w -X bogmater/weekscale-web/internal/version.buildRevision=${BUILD_REVISION}" \
        -o /out/weekscale-web \
        ./cmd/web

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S weekscale \
    && adduser -S -G weekscale weekscale

COPY --from=build --chown=weekscale:weekscale /out/weekscale-web /usr/local/bin/weekscale-web

USER weekscale

ENV HTTP_PORT=3333

EXPOSE 3333

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:3333/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/weekscale-web"]
