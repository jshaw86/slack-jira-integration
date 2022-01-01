# Two stage docker build to reduce image overhead
FROM golang:1.13.6-alpine3.11 as build
WORKDIR /app
COPY . /app
RUN apk add make gcc musl-dev
RUN make build

FROM alpine:3.11.3
WORKDIR /app
COPY --from=build  /app/main .
EXPOSE 8080
ENTRYPOINT ["/app/main"]
