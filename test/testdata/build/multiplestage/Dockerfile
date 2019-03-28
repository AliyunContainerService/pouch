FROM golang:1.12.1-alpine3.9 as build
COPY . /tmp
WORKDIR /tmp
RUN go build -o hello-world

FROM busybox:latest
WORKDIR /
COPY --from=build /tmp/hello-world .
CMD ["./hello-world"]
