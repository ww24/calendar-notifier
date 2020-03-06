#!/bin/ash

set -eo pipefail

if [ -n "$SERVICE_ACCOUNT" ]; then
    GOOGLE_APPLICATION_CREDENTIALS=/usr/local/etc/calendar-notifier/credential.json
    echo -n "$SERVICE_ACCOUNT" | base64 -d > "$GOOGLE_APPLICATION_CREDENTIALS"
fi

if [ -n "$CONFIG" ]; then
    echo -n "$CONFIG" | base64 -d > /usr/local/etc/calendar-notifier/config.yml
fi

export GOOGLE_APPLICATION_CREDENTIALS
calendar-notifier -config /usr/local/etc/calendar-notifier/config.yml
