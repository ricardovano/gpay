# Build project
FROM golang:1.20-alpine AS builder

RUN apk update --no-cache \
    && apk --no-cache add build-base git openssh-client ca-certificates

WORKDIR /app

RUN echo "Configuring gitconfig" \
    && echo '[url "git@github.com:ricardovano"]' >> /root/.gitconfig \
    && echo '  insteadOf = https://github.com/ricardovano' >> /root/.gitconfig \
    && echo '[url "git@github.com:ricardovano"]' >> /root/.gitconfig   \
    && echo '  insteadOf = https://github.com/ricardovano' >> /root/.gitconfig \
    && mkdir -p ~/.ssh && ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

COPY go.mod ./
COPY go.sum ./

RUN --mount=type=ssh go mod download -x

COPY . /go/src/github.com/ricardovano/qpay

WORKDIR /go/src/github.com/ricardovano/qpay/cmd/main

RUN go build -tags musl -buildmode=pie -o qpay

FROM alpine:3

RUN apk --no-cache add tzdata

WORKDIR /app

COPY --from=builder /go/src/github.com/ricardovano/qpay/cmd/main/qpay .

USER nobody

ENTRYPOINT [ "/app/qpay" ]