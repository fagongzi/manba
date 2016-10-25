#!/usr/bin/env bash
set -e

# Test to make sure we didn't get any linting errors
#   while also writing the linting errors to stderr
# DEV: `test -z` ensures the value is empty
# DEV: `| tee >(cat >&2)` takes the stdout from `golint` and replays it on stderr
#   this way we see the linting errors while still capturing them for `test -z`
test -z "`golint | tee >(cat >&2)`"
