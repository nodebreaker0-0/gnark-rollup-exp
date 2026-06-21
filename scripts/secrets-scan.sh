#!/usr/bin/env bash
# Fail if anything that looks like real secret material is committed.
# Scoped to active source and docs; matches secret VALUE shapes, not prose like
# the word "private key" (so docs that discuss keys are fine).
set -euo pipefail

SCAN_PATHS=(prove examples rollup scripts specs decisions Makefile README.md CHARTER.md delegation_matrix.md .specify)

# Patterns: PEM private-key blocks, Slack tokens, OpenAI-style keys, AWS access
# key IDs, and 64-hex literals embedded in quotes (a likely raw private key).
PATTERNS=(
	'-----BEGIN [A-Z ]*PRIVATE KEY-----'
	'xox[baprs]-[0-9A-Za-z-]{10,}'
	'sk-[A-Za-z0-9]{20,}'
	'AKIA[0-9A-Z]{16}'
	'"[0-9a-fA-F]{64}"'
)

hits=0
for pat in "${PATTERNS[@]}"; do
	if grep -REn --binary-files=without-match "$pat" "${SCAN_PATHS[@]}" 2>/dev/null; then
		echo "secrets-scan: pattern matched: $pat" >&2
		hits=1
	fi
done

if [ "$hits" -ne 0 ]; then
	echo "secrets-scan: FAIL — potential secret committed" >&2
	exit 1
fi

echo "secrets: OK"
