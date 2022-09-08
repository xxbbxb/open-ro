FROM golang:1.19-alpine as builder
WORKDIR /build
RUN apk add ca-certificates tzdata
COPY ./ ./
RUN go build ./

FROM alpine:3.15
WORKDIR /app
COPY --from=builder /build/openrobot ./openrobot
ENTRYPOINT ./openrobot