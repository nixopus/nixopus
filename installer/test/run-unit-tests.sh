#!/usr/bin/env bash
# Run the installer unit-test suite (bats).
# Usage: bash installer/test/run-unit-tests.sh [--verbose] [file...]
#
# With no arguments runs all .bats files under installer/test/unit/.
# Pass specific .bats file paths to run only those files.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UNIT_DIR="$SCRIPT_DIR/unit"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

BATS_FLAGS=()
FILES=()

for arg in "$@"; do
    case "$arg" in
        --verbose|-v) BATS_FLAGS+=(--verbose-run) ;;
        --tap)        BATS_FLAGS+=(--formatter tap) ;;
        *)            FILES+=("$arg") ;;
    esac
done

if [ ${#FILES[@]} -eq 0 ]; then
    mapfile -t FILES < <(find "$UNIT_DIR" -name "*.bats" | sort)
fi

if [ ${#FILES[@]} -eq 0 ]; then
    echo -e "${RED}No .bats files found under $UNIT_DIR${NC}" >&2
    exit 1
fi

if ! command -v bats &>/dev/null; then
    echo -e "${RED}bats not found. Install: https://github.com/bats-core/bats-core${NC}" >&2
    exit 1
fi

echo -e "${CYAN}${BOLD}━━━ Nixopus Installer Unit Tests ━━━${NC}"
echo ""
echo -e "  bats $(bats --version)"
echo -e "  Files: ${#FILES[@]}"
echo ""

bats "${BATS_FLAGS[@]}" "${FILES[@]}"
status=$?

echo ""
if [ $status -eq 0 ]; then
    echo -e "${GREEN}${BOLD}All tests passed.${NC}"
else
    echo -e "${RED}${BOLD}Some tests failed.${NC}"
fi

exit $status
