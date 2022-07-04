FROM golang:1.16-alpine3.13 AS build

WORKDIR /go/src/github.com/ww24/calendar-notifier
COPY . /go/src/github.com/ww24/calendar-notifier
ENV CGO_ENABLED=0
RUN go build -o /usr/local/bin/calendar-notifier ./cmd/server


FROM alpine:3.16.0

RUN apk add --no-cache tzdata ca-certificates

COPY --from=build /usr/local/bin/calendar-notifier /usr/local/bin/calendar-notifier
COPY --from=build /go/src/github.com/ww24/calendar-notifier/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=build /go/src/github.com/ww24/calendar-notifier/config.sample.yml /usr/local/etc/calendar-notifier/config.yml

ENTRYPOINT [ "entrypoint.sh" ]
