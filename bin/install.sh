#!/bin/bash
# Установка команды build в систему.
# После этого `build <alias>` работает из любой папки:
#   build browser       - анимация + запуск в фоне
#   build browser -i    - сразу запуск, без анимации
#   build list          - показать алиасы
# Скрипт ссылается на этот репозиторий, поэтому правки
# aliases.go подхватываются сразу, переустановка не нужна.
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TARGET="/usr/local/bin/build"
if [ -w /usr/local/bin ]; then
  ln -sf "$ROOT/bin/build.sh" "$TARGET"
else
  echo "Нужен sudo для записи в /usr/local/bin"
  sudo ln -sf "$ROOT/bin/build.sh" "$TARGET"
fi
echo "Готово: $TARGET -> $ROOT/bin/build.sh"
echo "Проверка:"
"$TARGET" list
