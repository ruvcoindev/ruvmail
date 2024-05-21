FROM docker.io/golang:alpine3.18 as builder

COPY . /src
WORKDIR /src

RUN apk add --no-cache --update go gcc g++
RUN go build -o /src/ruvmail ./cmd/ruvmail

FROM docker.io/alpine:3.18

LABEL org.opencontainers.image.source=https://github.com/ruvcoindev/ruvmail
LABEL org.opencontainers.image.description=Ruvmail
LABEL org.opencontainers.image.licenses=MPL-2.0

COPY --from=builder /src/ruvmail /usr/bin/ruvmail

EXPOSE 1143/tcp
EXPOSE 1025/tcp
VOLUME /etc/ruvmail

ENTRYPOINT ["/usr/bin/ruvmail", "-smtp=:1025", "-imap=:1143", "-database=/etc/ruvmail/ruvmail.db"]
