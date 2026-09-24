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

function printImportantMessage() {
    $colorful && tput setaf 3 || true
    local timestamp=$([ -n "${WTS:-}" ] && [ "${WTS}" != "0" ] && date +'[%Y-%m-%d %H:%M:%S] ')
    >&2 echo "${timestamp}$*"
    $colorful && tput sgr0 || true
}

# Move to project root
cd ..

function ensureCleanWorktree() {
    if [[ -n "$(git status --porcelain)" ]]; then
        printError "Working tree is not clean. Please commit or stash your changes."
        exit 1
    fi
}

if ! command -v relstep > /dev/null; then
    printError "relstep not found. Install it with: go install github.com/shadowopera/archmage/tools/relstep@latest"
    exit 1
fi

STEP_LIST=(
    checkVersion
    runTests
    updateChangelog
    tagRelease
)
STEPS=$(IFS=,; echo "${STEP_LIST[*]}")

function markStepDone() {
    relstep mark -steps "$STEPS" "$1" || exit 1
}

VERSION_ARG="${1:-}"

while true; do
    # Run relstep (pass version arg only on first iteration)
    RESULT=$(relstep next -steps "$STEPS" $VERSION_ARG 2>&1) || {
        printError "$RESULT"
        exit 1
    }
    # Clear version arg after first iteration
    VERSION_ARG=""

    if [[ "$RESULT" == "DONE" ]]; then
        printImportantMessage "Release completed!"
        printMessage "Delete release.json when ready for the next release."
        printImportantMessage "Run 'git push --follow-tags' to publish."
        exit 0
    fi

    NEXT_STEP="$RESULT"
    VERSION=$(relstep version 2>&1) || {
        printError "Failed to get version: $VERSION"
        exit 1
    }

    printImportantMessage "Step: $NEXT_STEP (v$VERSION)"

    case "$NEXT_STEP" in
        checkVersion)
            if git tag -l "v$VERSION" | grep -q "v$VERSION"; then
                printError "Git tag v$VERSION already exists."
                exit 1
            fi
            printMessage "Version v$VERSION is available."
            ensureCleanWorktree
            markStepDone "checkVersion"
            ;;

        runTests)
            printMessage "Running tests..."
            if ! go test ./...; then
                printError "Tests failed."
                exit 1
            fi
            ensureCleanWorktree
            markStepDone "runTests"
            ;;

        updateChangelog)
            printMessage "Summarize the subject and body of all commit messages since the last git tag, extract key changes, and update the root directory @CHANGELOG.md according to the 'Keep a Changelog 1.1.0' format. New version number: $VERSION"
            if ! gum confirm --affirmative=OK --negative=Cancel \
                "After updating CHANGELOG.md, click OK to commit changes."; then
                printMessage "Aborted."
                exit 1
            fi

            if ! awk -v h="## [$VERSION]" 'index($0, h) == 1 { found = 1; exit } END { exit !found }' CHANGELOG.md; then
                printError "Heading \"## [$VERSION]\" is not found in CHANGELOG.md."
                exit 1
            fi

            printMessage "Staging CHANGELOG.md..."
            if ! git add CHANGELOG.md; then
                printError "Failed to stage CHANGELOG.md."
                exit 1
            fi

            printMessage "Committing changes..."
            if ! git commit -m "docs: update changelog"; then
                printError "Failed to commit CHANGELOG.md changes."
                exit 1
            fi

            ensureCleanWorktree
            markStepDone "updateChangelog"
            ;;

        tagRelease)
            printMessage "Running defensive tests..."
            if ! go test ./...; then
                printError "Tests failed."
                exit 1
            fi

            ensureCleanWorktree

            # Create git tag (skip if already exists)
            if git tag -l "v$VERSION" | grep -q "v$VERSION"; then
                printMessage "Tag v$VERSION already exists, skipping."
            else
                if ! gum confirm "Do you want to create git tag v$VERSION?"; then
                    printMessage "Skipping git tag."
                else
                    git tag -a "v$VERSION" -m "v$VERSION"
                    printImportantMessage "Tag v$VERSION created."
                fi
            fi
            if ! git tag -l "v$VERSION" | grep -q "v$VERSION"; then
                printError "Git tag v$VERSION was not created. Please run: git tag v$VERSION"
                exit 1
            fi

            ensureCleanWorktree
            markStepDone "tagRelease"
            ;;

        *)
            printError "Unknown step: $NEXT_STEP"
            exit 1
            ;;
    esac
done
