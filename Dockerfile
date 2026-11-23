# Multi-stage venue server image. The final stage is also used by the one-shot
# seed and simulation containers in the local QA stack.
# Pin the compiler to the version the module was built and tested with.
# 1.27.4 currently hits a modernc/sqlite compiler ICE on this tree.
FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/j3d ./cmd/j3d

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && mkdir -p /data
COPY --from=build /out/j3d /usr/local/bin/j3d
ENV J3_ADDR=0.0.0.0:8787 \
    J3_DB=/data/j3.db \
    J3_VERSION=demo \
    J3_DEMO=1
EXPOSE 8787
VOLUME ["/data"]
ENTRYPOINT ["j3d"]
