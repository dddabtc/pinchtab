#!/bin/bash
# 27-screenshot-flags.sh — CLI screenshot flags

source "$(dirname "$0")/common.sh"

# SKIP: screenshot -o/-q flags are consumed by cobra even with FParseErrWhitelist.
# They need to be registered as proper cobra flags to reach the args slice.
# The docker runner also has a read-only filesystem outside /tmp.
# TODO: register -o, -q, --tab as cobra flags on screenshotCmd
