package main

// ============================================================
//  АЛИАСЫ — все твои программы живут ТУТ.
//
//  Как добавить свою:
//    1. Скопируй любой блок ниже целиком.
//    2. Поменяй значения под себя.
//    3. Пересобери:   go build -o build .
//
//  Что значит каждое поле:
//    EmergeName  — фейковое имя пакета (категория/имя), только для лога
//    Version     — версия, только для лога
//    Use         — USE-флаги, только для лога
//    Command     — реальная программа, запустится ПОСЛЕ "сборки".
//                  Можно писать вместе с аргументами:
//                    Command: "code --new-window",
//                    Command: "firefox --private-window",
//                  (понимает двойные кавычки для путей с пробелами)
//    Args        — дополнительные аргументы (обычно оставляй пустым;
//                  аргументы из командной строки дописываются сами:
//                  ./build code ~/myproject)
//    BuildTime   — сколько секунд идёт анимация сборки
//    Foreground  — false = запуск в фоне, без логов, терминал свободен
//                          (для оконных: браузер, телега, vscode)
//                  true  = запуск в том же терминале, с логами
//                          (для консольных: badapple, htop, vim)
//
//  Флаги --fg / --bg в командной строке перекрывают Foreground
//  для одного запуска:   ./build browser --fg
// ============================================================

var targets = map[string]BuildTarget{
	// ./build browser  ->  "соберёт" браузер, запустит yandex в фоне
	"browser": {
		EmergeName:  "www-client/yandex-browser",
		Version:     "24.7.1.1234-r1",
		Description: "Yandex Browser",
		Use:         "ffmpeg h264 proprietary-codecs systray",
		Command:     "yandex-browser-stable",
		Args:        []string{},
		BuildTime:   20,
		Foreground:  false,
	},

	// ./build tg  ->  "соберёт" телегу, запустит AyuGram в фоне
	"tg": {
		EmergeName:  "net-im/telegram-desktop",
		Version:     "5.7.1-r2",
		Description: "Telegram",
		Use:         "enchant gtk3 wayland",
		Command:     "AyuGram",
		Args:        []string{},
		BuildTime:   15,
		Foreground:  false,
	},

	// ./build code  ->  "соберёт" редактор, запустит VS Code в фоне
	"code": {
		EmergeName:  "app-editors/vscode",
		Version:     "1.92.2",
		Description: "VS Code",
		Use:         "-minimal telemetry",
		Command:     "code",
		Args:        []string{},
		BuildTime:   20,
		Foreground:  false,
	},

	// ./build fox  ->  "соберёт" лису, запустит Firefox в фоне
	"fox": {
		EmergeName:  "www-client/firefox",
		Version:     "128.0.3",
		Description: "Firefox",
		Use:         "gmp-autoupdate openh264 system-libvpx wayland",
		Command:     "firefox",
		Args:        []string{},
		BuildTime:   500,
		Foreground:  false,
	},

	// ./build badapple  ->  консольная штука, поэтому в том же терминале
	"badapple": {
		EmergeName:  "fun/badapple",
		Version:     "67.69.52",
		Description: "BadApple",
		Use:         "gmp-autoupdate openh264 system-libvpx wayland",
		Command:     "badapple",
		Args:        []string{},
		BuildTime:   5,
		Foreground:  true,
	},
	"doom-cli": {
		EmergeName:  "fun/doom-cli",
		Version:     "67.69.52",
		Description: "Doom CLI",
		Use:         "gazan-lib openpizdezh system-libvpx wayland",
		Command:     "doom_ascii -iwad /usr/share/games/doom/freedoom2.wad",
		Args:        []string{},
		BuildTime:   20,
		Foreground:  true,
	},
}
