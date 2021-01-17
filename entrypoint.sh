#!/bin/ash

set -eo pipefail

if [ -n "$SERVICE_ACCOUNT" ]; then
    GOOGLE_APPLICATION_CREDENTIALS=/usr/local/etc/calendar-notifier/credential.json
    echo -n "$SERVICE_ACCOUNT" | base64 -d > "$GOOGLE_APPLICATION_CREDENTIALS"
fi

confpath=/usr/local/etc/calendar-notifier/config.yml
if [ -n "$CONFIG_BASE64" ]; then
    echo -n "$CONFIG_BASE64" | base64 -d > "$confpath"
elif [ -n "$CONFIG" ]; then
    echo -n "$CONFIG" > "$confpath"
fi

export GOOGLE_APPLICATION_CREDENTIALS
calendar-notifier -config "$confpath"
