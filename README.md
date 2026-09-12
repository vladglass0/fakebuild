# fakebuild — фейковый emerge как в Gentoo

Пишешь `build browser` → идёт анимация сборки в стиле Gentoo → запускается реальная программа (`yandex-browser-stable`).

## Сборка

```bash
go build -o build .
```

## Использование

```bash
./build browser          # анимация + запуск yandex-browser-stable в фоне, без логов
./build browser -t 5     # то же, но анимация ~5 сек
./build browser -i       # без анимации, сразу запустить программу
./build browser --fg     # запуск в том же терминале с логами (для отладки)
./build tg               # AyuGram в фоне
./build -h               # помощь
```

Аргументы после алиаса пробрасываются в запускаемую программу:
```bash
./build code ~/myproject
```

## Запуск в фоне

По умолчанию программа запускается через `setsid` со stdin/stdout/stderr
в `/dev/null`:
- терминал сразу освобождается, `build` завершается, приложение живёт само;
- никакой вывод запущенной программы в терминал не сыпется;
- `Ctrl+C` по терминалу приложение не убивает.

Для отладки есть режим foreground — вывод идёт в тот же терминал,
`build` ждёт завершения программы:

```bash
./build browser --fg
```

Режим по умолчанию задаётся в `aliases.go` полем `Foreground`,
флаги `--fg` / `--bg` перекрывают его для одного запуска.

## Как настроить свои алиасы

Алиасы живут в файле `aliases.go` — открой его, скопируй любой блок,
поменяй значения и пересобери: `go build -o build .`.

```go
"discord": {
    EmergeName:  "net-im/discord",
    Version:     "0.0.72",
    Use:         "system-ffmpeg",
    Command:     "discord", // можно с аргументами: "code --new-window"
    Args:        []string{},
    BuildTime:   8,
    Foreground:  false, // false = фон без логов, true = тот же терминал
},
```

Аргументы можно писать прямо в `Command` (понимает кавычки):
```go
Command: "doom_ascii -iwad /usr/share/games/doom/freedoom2.wad",
```
или дописывать при запуске: `./build code ~/myproject`.
`./build list` показывает только имена алиасов.

## Как сделать командой `build` везде

```bash
sudo cp build /usr/local/bin/build
build browser
```
