#!/usr/bin/env bash

[[ "$TRACE" ]] && set -x
pushd `dirname "$0"` > /dev/null
trap __EXIT EXIT

colorful=false
if [[ -t 1 ]]; then
    colorful=true
fi

function __EXIT() {
    popd > /dev/null
}

function printMessage() {
    >&2 echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*"
}

function printError() {
    $colorful && tput setaf 1
    local timestamp=$([ -n "$WTS" ] && [ "$WTS" != "0" ] && date +'[%Y-%m-%d %H:%M:%S] ')
    >&2 echo "${timestamp}ERROR: $*"
    $colorful && tput setaf 7
}

function printImportantMessage() {
    $colorful && tput setaf 3
    local timestamp=$([ -n "$WTS" ] && [ "$WTS" != "0" ] && date +'[%Y-%m-%d %H:%M:%S] ')
    >&2 echo "${timestamp}$*"
    $colorful && tput setaf 7
}

go test -v -cover -coverprofile=c.out -coverpkg=github.com/shadowopera/sdk-go/archmage "$@" ./... && \
go tool cover -html=c.out && \
rm c.out
