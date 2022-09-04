FROM golang:1.18-alpine as builder
WORKDIR /build
COPY ./ ./
RUN apk add ca-certificates tzdata && \
  ls -alt ./* && \
  go build ./cmd/robot

FROM alpine:3.14
WORKDIR /app
COPY --from=builder /build/robot ./robot
COPY cmd/robot/dev.crt cmd/robot/dev.key ./
ENTRYPOINT ./robot