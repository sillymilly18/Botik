FROM golang:1.24.5-alpine as builder

COPY . /payproc/source
COPY .env /payproc/source/config

WORKDIR /payproc/source
RUN go mod tidy
RUN go build -o ./dist/migrator ./cmd/migrator

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /payproc/source/migrations ./migrations
COPY --from=builder /payproc/source/dist/migrator .
COPY --from=builder /payproc/source/config .

CMD ./migrator
