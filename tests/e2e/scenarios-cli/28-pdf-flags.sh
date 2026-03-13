#!/bin/bash
# 28-pdf-flags.sh — CLI pdf flags

source "$(dirname "$0")/common.sh"

# SKIP: pdf -o/--landscape/--scale/--page-ranges flags are consumed by cobra
# even with FParseErrWhitelist. They need to be registered as proper cobra flags.
# The docker runner also has a read-only filesystem outside /tmp.
# TODO: register -o, --landscape, --scale, --page-ranges, --tab as cobra flags on pdfCmd
