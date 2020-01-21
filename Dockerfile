FROM golang:1.13.6-alpine3.11 AS build

WORKDIR /go/src/github.com/ww24/calendar-worker
COPY . /go/src/github.com/ww24/calendar-worker
ENV CGO_ENABLED=0
RUN go build -o /usr/local/bin/calendar-worker ./cmd/server


FROM alpine:3.11

RUN apk add --no-cache tzdata ca-certificates

COPY --from=build /usr/local/bin/calendar-worker /usr/local/bin/calendar-worker
COPY --from=build /go/src/github.com/ww24/calendar-worker/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=build /go/src/github.com/ww24/calendar-worker/config.sample.yml /usr/local/etc/calendar-worker/config.yml

ENV GOOGLE_APPLICATION_CREDENTIALS=/usr/local/etc/calendar-worker/credential.json
ENTRYPOINT [ "entrypoint.sh" ]
