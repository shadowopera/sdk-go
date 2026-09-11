#!/usr/bin/env bash

[[ "$TRACE" ]] && set -x
pushd "$(dirname "$0")" > /dev/null
trap __EXIT EXIT

colorful=false
if [[ -t 1 ]] && [[ -n "${TERM:-}" ]]; then
    colorful=true
fi

function __EXIT() {
    popd > /dev/null
}

function printMessage() {
    local timestamp=$([ -n "${WTS:-}" ] && [ "${WTS}" != "0" ] && date +'[%Y-%m-%d %H:%M:%S] ')
    >&2 echo "${timestamp}$*"
}

function printError() {
    $colorful && tput setaf 1 || true
    local timestamp=$([ -n "${WTS:-}" ] && [ "${WTS}" != "0" ] && date +'[%Y-%m-%d %H:%M:%S] ')
    >&2 echo "${timestamp}ERROR: $*"
    $colorful && tput sgr0 || true
}

# Move to project root
cd ..

outDir=../docs/archmage/src/content/docs/overview-go
if [[ ! -d "$outDir" ]]; then
    printError "$outDir does not exist"
    exit 1
fi

# Process README.md for Starlight
printMessage "Generating $outDir/sdk-go.mdx ..."
{
    echo "---"
    echo "title: 'Go SDK Overview'"
    echo "sidebar:"
    echo "  label: Overview"
    echo "  order: 1"
    echo "---"
    echo ""
    perl -0777 -pe 's/\n---\s+## Development.*//s' README.md | \
        perl -pe 's|\./images/archmage\.jpg|../../../assets/archmage/archmage.jpg|g'
} > "$outDir/sdk-go.mdx"

# Stage all changes
printMessage "Staging changes in docs site ..."
cd ../docs
git add -A

echo
printMessage "Done."
