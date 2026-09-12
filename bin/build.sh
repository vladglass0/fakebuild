#!/bin/bash
# Запуск без компиляции:  bin/build.sh <alias> [флаги] [аргументы]
#   bin/build.sh browser          - анимация + запуск в фоне
#   bin/build.sh browser -i       - сразу запуск, без анимации
#   bin/build.sh browser -t 5     - анимация ~5 секунд
#   bin/build.sh list             - показать алиасы
# Все аргументы передаются программе как есть (go run . "$@").
set -e
# readlink -f раскрывает симлинк (/usr/local/bin/build -> .../bin/build.sh),
# поэтому работает и напрямую, и через ссылку из любого места.
SELF="$(readlink -f "${BASH_SOURCE[0]}")"
cd "$(dirname "$SELF")/.."
exec go run . "$@"
