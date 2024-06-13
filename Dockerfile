FROM docker.io/golang:alpine3.20 as builder

COPY . /src
WORKDIR /src

RUN apk add --no-cache --update go gcc g++
RUN go build -o /src/ruvmail ./cmd/ruvmail

FROM docker.io/alpine:3.18

LABEL in.ruvcha.image.source=https://github.com/ruvcoindev/ruvmail
LABEL in.ruvcha.image.description=Ruvmail
LABEL in.ruvcha.image.licenses=MPL-2.0

COPY --from=builder /src/ruvmail /usr/bin/ruvmail

EXPOSE 2024/tcp
EXPOSE 2025/tcp
VOLUME /etc/ruvmail

ENTRYPOINT ["/usr/bin/ruvmail", "-smtp=:2025", "-imap=:2024", "-database=/etc/ruvmail/ruvmail.db"]
