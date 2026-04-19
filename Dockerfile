FROM golang:1.25 AS builder
WORKDIR /app

COPY . .
RUN ./build.sh

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/output/bin/hertz_service /app/hertz_service
COPY .env /app/.env
CMD ["/app/hertz_service"]