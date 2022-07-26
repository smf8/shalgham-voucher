FROM golang:1.18 AS build

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

ARG GO_PROXY
ENV GOPROXY=${GO_PROXY}

RUN mkdir -p /src

# First add modules list to better utilize caching
COPY go.sum go.mod /src/

WORKDIR /src

RUN go mod download

COPY . /src

# Build components.
# Put built binaries and runtime resources in /app dir ready to be copied over or used.
RUN go build -ldflags '-w -s' .  && \
    mkdir -p /app && \
    cp ./voucher ./wait-for-it.sh /app/ && \
    cp -r /app/

#
# 2. Runtime Container
#
FROM alpine:latest

WORKDIR /app

#RUN #apt update && apt install -y ca-certificates

COPY --from=build /app /app/

CMD ["./voucher"]
