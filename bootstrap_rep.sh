#!/bin/bash

APP_COMMAND="./replive"
APP_NAME="replive"
LOG_FILE="/var/log/${APP_NAME}/monitor.log"

mkdir -p $(dirname $LOG_FILE)

find . -name "replive_*.log" -exec mv {} logs/ \;

while true; do
    if ! pgrep -f "$APP_COMMAND" > /dev/null; then
        echo "$(date): $APP_NAME is not running. Starting it..." >> $LOG_FILE
        $APP_COMMAND &
        echo "$(date): $APP_NAME started with PID $!" >> $LOG_FILE
    fi
    sleep 5
done