#!/bin/bash
set -e

ps aux | egrep 'go run|go-build' | grep -v grep | awk '{print $2}' | xargs kill
$GOPATH/bin/dlv debug
