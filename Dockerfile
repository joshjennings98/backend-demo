FROM golang:1.22.0 as build

WORKDIR /build
ADD server server
ADD cli cli

WORKDIR /build/cli
RUN go mod tidy
RUN go build -o /output/backend-demo

FROM alpine:3.19.1

# Install go binary
COPY --from=build /output/backend-demo /usr/local/bin/backend-demo

ENTRYPOINT ["backend-demo"]
