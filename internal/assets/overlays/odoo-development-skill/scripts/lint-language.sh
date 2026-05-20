#!/usr/bin/env bash
# lint-language.sh — Checks for Spanish keywords in Odoo modules to enforce English policy.
set -euo pipefail

TARGET_DIR="${1:-.}"
echo "Running language policy linter on target: ${TARGET_DIR}..."

SPANISH_WORDS=(
	"vista"
	"modelo"
	"campo"
	"descripcion"
	"dirección"
	"direccion"
	"nombre"
	"telefono"
	"teléfono"
	"correo"
	"usuario"
	"codigo"
	"código"
)

FOUND=0

# Use grep to find Spanish keywords, ignoring case, excluding common directories
for word in "${SPANISH_WORDS[@]}"; do
	# Search in target directory excluding git, node_modules, and cache/venv directories
	matches=$(grep -rni \
		--exclude-dir=".git" \
		--exclude-dir="node_modules" \
		--exclude-dir="__pycache__" \
		--exclude-dir="venv" \
		--exclude-dir=".venv" \
		--exclude="lint-language.sh" \
		--exclude="*translation*" \
		"${word}" "${TARGET_DIR}" || true)

	if [ -n "${matches}" ]; then
		echo -e "\n\033[31m[FAIL] Found forbidden Spanish keyword '${word}':\033[0m"
		echo "${matches}"
		FOUND=1
	fi
done

if [ ${FOUND} -eq 1 ]; then
	echo -e "\n\033[31m[ERROR] Language Policy Violated!\033[0m"
	echo "All Odoo code, comments, views, and schemas must be in ENGLISH."
	echo "Please replace the Spanish keywords with their English equivalents."
	exit 1
else
	echo -e "\n\033[32m[PASS] Language policy check passed successfully!\033[0m"
	exit 0
fi
