FROM golang:1.20 as build
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM debian:bookworm-slim
RUN apt-get update && apt-get install --yes ca-certificates
RUN groupadd -r app && useradd --no-log-init -r -g app app
USER app
COPY --from=build /go/bin/app /
COPY swagger /swagger
ENV APP_ADDR ":8080"
EXPOSE 8080
ENTRYPOINT ["/app"]
