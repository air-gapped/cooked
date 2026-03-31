#!/bin/sh
set -e
update-ca-certificates 2>/dev/null || true
exec cooked "$@"
