package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"
)

// ============================================================
//  АЛИАСЫ ЖИВУТ В ФАЙЛЕ aliases.go (рядом, в этом же пакете).
//  Правишь там — потом пересобери:  go build -o build .
// ============================================================

type BuildTarget struct {
	EmergeName  string
	Version     string
	Description string
	Use         string
	Command     string
	Args        []string
	BuildTime   int  // секунды
	Foreground  bool // фон без логов (false) или тот же терминал (true)
}

// Цвета для лога в стиле portage
const (
	green  = "\033[32;01m"
	yellow = "\033[33;01m"
	red    = "\033[31;01m"
	blue   = "\033[34;01m"
	teal   = "\033[36;01m"
	dim    = "\033[90m"
	normal = "\033[0m"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	arg := os.Args[1]

	// служебные команды
	switch arg {
	case "-h", "--help", "help":
		usage()
		return
	case "list", "-l", "--list":
		listTargets()
		return
	}

	// флаг -t N для переопределения времени: ./build browser -t 5
	// флаг -i: пропустить имитацию сборки, сразу запустить программу
	timeOverride := -1
	instant := false
	extraArgs := []string{}
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "-t" && i+1 < len(os.Args) {
			var t int
			fmt.Sscanf(os.Args[i+1], "%d", &t)
			timeOverride = t
			i++
		} else if os.Args[i] == "-i" || os.Args[i] == "--instant" {
			instant = true
		} else {
			extraArgs = append(extraArgs, os.Args[i])
		}
	}

	t, ok := targets[arg]
	if !ok {
		fmt.Printf("%s * %sUnknown target '%s'%s\n", red, "", arg, normal)
		fmt.Println("Targets:")
		listTargets()
		os.Exit(1)
	}

	if timeOverride > 0 {
		t.BuildTime = timeOverride
	}
	// аргументы из командной строки добавляем к команде запуска
	t.Args = append(t.Args, extraArgs...)

	if !instant {
		fakeBuild(t)
	}
	launch(t)
}

func usage() {
	fmt.Printf("Fast Installer by RtxBB\n\n")
	fmt.Printf("Use:\n")
	fmt.Printf("  ./build <packet> [-t very important] [-i] [args]\n")
	fmt.Printf("  ./build <packet> -i        - run without build simulation\n")
	fmt.Printf("  ./build list              - show packages\n\n")
	listTargets()
}

func listTargets() {
	aliases := make([]string, 0, len(targets))
	for alias := range targets {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		fmt.Printf("  %s\n", alias)
	}
}

// ---------------- ФЕЙКОВАЯ СБОРКА ----------------

// timeScale масштабирует все фиксированные паузы, чтобы BuildTime / -t
// задавал ВСЁ время анимации, а не только фазу компиляции.
// Полное время ≈ BuildTime: ~60% — фиксированные фазы, ~40% — компиляция.
// Базовая норма: при timeScale=1.0 фиксированные фазы длятся ~18 сек.
var timeScale float64 = 1.0

func fakeBuild(t BuildTarget) {
	start := time.Now()
	if t.BuildTime <= 0 {
		t.BuildTime = 1
	}
	timeScale = float64(t.BuildTime) * 0.6 / 18.0
	if timeScale < 0.02 {
		timeScale = 0.02
	}
	if timeScale > 20 {
		timeScale = 20
	}
	pkg := fmt.Sprintf("%s-%s", t.EmergeName, t.Version)

	phaseHeader(pkg, t)
	phaseResolve(pkg, t)
	phaseFetch(pkg, t)
	phaseDigest(pkg)
	phaseUnpack(pkg)
	phasePrepare(pkg)
	phaseConfigure(pkg, t)
	phaseCompile(pkg, t)
	phaseTest(pkg)
	phaseInstall(pkg)
	phasePostinst(pkg, t)
	phaseMerge(pkg, start)
}

// 0. Шапка + emerge --info
func phaseHeader(pkg string, t BuildTarget) {
	fmt.Printf("%s _-----_%s\n", teal, normal)
	fmt.Printf("%s(       \\%s  Gentoo Linux Emerge v2.0\n", teal, normal)
	fmt.Printf("%s\\    ___\\%s  Live Build Simulation\n", teal, normal)
	fmt.Printf("%s \\/ /     %s CPU: %s | Jobs: %s\n", teal, normal, fakeCPU(), fakeJobs())
	fmt.Printf("%s * %sPortage %s2.3.99%s  gcc %s  glibc %s  kernel %s%s\n",
		green, normal, blue, normal, fakeGCC(), fakeGlibc(), fakeKernel(), normal)
	fmt.Printf("%s * %sFEATURES=\"%s\"  MAKEOPTS=\"%s\"%s\n", green, normal, fakeFeatures(), fakeJobs(), normal)
	fmt.Printf("%s * %sCFLAGS=\"-O2 -march=native -pipe -fno-plt\"  CXXFLAGS=\"-O2 -march=native -pipe -fno-plt\"%s\n", green, normal, normal)
	fmt.Printf("%s * %sCPU_FLAGS_X86=\"%s\"%s\n", green, normal, fakeCPUFlags(), normal)
	fmt.Println()
}

// 1. Расчёт зависимостей с деревом
func phaseResolve(pkg string, t BuildTarget) {
	fmt.Printf("%sThese are the packages that would be merged, in order:%s\n\n", green, normal)
	dots("Calculating dependencies", 500)
	dots("Resolving dependencies", 400)

	deps := fakeDeps(t.EmergeName)
	for _, d := range deps {
		fmt.Printf("[ebuild  %sN%s     ] %s %sUSE=\"%s\"%s\n", dim, normal, d, dim, randomUse(), normal)
	}
	fmt.Printf("[ebuild  %sN%s %s%s%s ] USE=\"%s\" LINGUAS=\"%s\" %s%d KiB%s\n\n",
		green, normal, blue, pkg, normal, t.Use, fakeLinguas(), green, 80000+rand.Intn(400000), normal)

	fmt.Printf("Total: %d packages (%d new), Size of downloads: %d KiB\n\n",
		len(deps)+1, len(deps)+1, 80000+rand.Intn(400000))
	sleep(400)
	fmt.Printf("%s>>>%s Emerging (1 of %d) %s%s%s\n", green, normal, len(deps)+1, blue, pkg, normal)
	fmt.Printf("%s>>>%s Dependency resolution took %.2f s\n", green, normal, 0.5+rand.Float64()*2)
	sleep(500)
}

// 2. Скачивание дистфайлов с прогресс-баром
func phaseFetch(pkg string, t BuildTarget) {
	tarball := fmt.Sprintf("%s.tar.xz", pkg[strings.LastIndex(pkg, "/")+1:])
	mirror := pick([]string{
		"https://distfiles.gentoo.org/distfiles",
		"https://mirror.yandex.ru/gentoo-distfiles/distfiles",
		"https://gentoo-mirror.net/distfiles",
	})
	fmt.Printf("%s>>>%s Downloading '%s/%s'\n", green, normal, mirror, tarball)
	downloadBar(1200+time.Duration(rand.Intn(800))*time.Millisecond,
		85000+rand.Intn(300000))
	// иногда докачка / второе зеркало
	if rand.Float32() < 0.35 {
		fmt.Printf(" --2026-09-12 23:59:59--  %s/%s\n", mirror, tarball)
		fmt.Printf("Resolving mirror... connected.\n")
		fmt.Printf("HTTP request sent, awaiting response... 206 Partial Content\n")
		downloadBar(500*time.Millisecond, 12000+rand.Intn(20000))
	}
	fmt.Printf("%s>>>%s Fetching... done (%d KiB)\n", green, normal, 80000+rand.Intn(200000))
	sleep(200)
}

// 3. Проверка манифеста
func phaseDigest(pkg string) {
	fmt.Printf("%s>>>%s Verifying ebuild manifests ... %s[ ok ]%s\n", green, normal, green, normal)
	sleep(250)
	checks := []string{"SHA256", "SHA512", "BLAKE2B", "size"}
	for _, c := range checks {
		fmt.Printf("%s>>>%s Checking %s digest ... %s[ ok ]%s\n", green, normal, c, green, normal)
		sleep(120)
	}
	fmt.Printf("%s>>>%s Verifying signature (openpgp-keys-gentoo-release) ... %s[ ok ]%s\n",
		green, normal, green, normal)
	sleep(250)
}

// 4. Распаковка
func phaseUnpack(pkg string) {
	fmt.Printf("%s>>>%s Unpacking source...\n", green, normal)
	fmt.Printf("%s>>>%s Unpacking %s.tar.xz to /var/tmp/portage/%s/work\n",
		green, normal, pkg[strings.LastIndex(pkg, "/")+1:], pkg)
	// фейковый tar-вывод
	files := 1500 + rand.Intn(4000)
	for i := 0; i < 3; i++ {
		fmt.Printf("%star: %s/%s/file%d.cpp%s\n", dim, pkg, randomDir(), rand.Intn(9000), normal)
		sleep(120)
	}
	fmt.Printf("%s... %d files, %d MiB unpacked%s\n", dim, files, 200+rand.Intn(900), normal)
	spin(700)
	fmt.Printf("%s>>>%s Source unpacked in /var/tmp/portage/%s/work\n", green, normal, pkg)
	sleep(200)
}

// 5. Патчи + eautoreconf
func phasePrepare(pkg string) {
	fmt.Printf("%s>>>%s Preparing source in /var/tmp/portage/%s/work ...\n", green, normal, pkg)
	patches := []string{
		fmt.Sprintf("%s-gentoo-defaults.patch", shortName(pkg)),
		"gcc-14-compat-musl-fixes.patch",
		"chromium-system-ffmpeg-r3.patch",
		"fix-widevine-wayland-r1.patch",
		"respect-CFLAGS-LDFLAGS.patch",
	}
	n := 2 + rand.Intn(3)
	for i := 0; i < n; i++ {
		fmt.Printf(" * Applying %s ... %s[ ok ]%s\n", patches[rand.Intn(len(patches))], green, normal)
		sleep(180)
	}
	fmt.Printf(" * Applying user patches from /etc/portage/patches/%s ... %s[ ok ]%s\n", pkg, green, normal)
	sleep(150)
	fmt.Printf(" * Removing bundled libraries (ffmpeg, libvpx, icu, zlib) ... done\n")
	sleep(150)
	fmt.Printf(" * Running eautoreconf in '%s' ...\n", randomDir())
	fmt.Printf("%sautoreconf -fi%s\n", dim, normal)
	spin(600)
	fmt.Printf(" * Running elibtoolize in: %s/work/%s/\n", pkg, shortName(pkg))
	sleep(200)
	fmt.Printf(" * eapply_user: no user patches found, skipping\n")
	sleep(150)
}

// 6. Configure / cmake
func phaseConfigure(pkg string, t BuildTarget) {
	fmt.Printf("%s>>>%s Configuring source in /var/tmp/portage/%s/work ...\n", green, normal, pkg)
	sleep(200)
	checks := []string{
		"checking for x86_64-pc-linux-gnu-gcc... x86_64-pc-linux-gnu-gcc",
		"checking whether the C compiler works... yes",
		"checking for x86_64-pc-linux-gnu-g++... x86_64-pc-linux-gnu-g++",
		"checking for cmake >= 3.25... yes",
		"checking for ninja... yes",
		"checking for pkg-config... /usr/bin/pkg-config",
		"checking for nss >= 3.90... yes",
		"checking for alsa-lib... yes",
		"checking for pulseaudio... yes",
		"checking for system-ffmpeg... yes",
		"checking for wayland-client... yes",
	}
	for _, c := range checks {
		fmt.Printf("%s%s%s\n", dim, c, normal)
		sleep(90)
	}
	fmt.Printf(" * econf: updating %s/build/config.sub ... %s[ ok ]%s\n", shortName(pkg), green, normal)
	sleep(150)
	fmt.Printf(" * CFLAGS=\"-O2 -march=native -pipe -fno-plt\" CXXFLAGS=\"-O2 -march=native -pipe -fno-plt\" LDFLAGS=\"-Wl,-O1 -Wl,--as-needed\"\n")
	fmt.Printf(" * USE=\"%s\"\n", t.Use)
	fmt.Printf(" * CMAKE_BUILD_TYPE=Release  CMAKE_TOOLCHAIN_FILE=/usr/share/cmake/gentoo.cmake\n")
	for _, l := range []string{
		"-- Found Python3: /usr/bin/python3.12",
		"-- Found Ninja: /usr/bin/ninja",
		"-- Found ALSA: /usr/lib64/libasound.so",
		"-- Found VAAPI: /usr/lib64/libva.so",
		"-- Configuring done",
		"-- Generating done",
	} {
		fmt.Printf("%s%s%s\n", teal, l, normal)
		sleep(120)
	}
	spin(700)
	fmt.Printf(" * configure: creating ./config.status ... done\n")
	sleep(200)
}

// 7. Компиляция — самая длинная фаза
func phaseCompile(pkg string, t BuildTarget) {
	fmt.Printf("%s>>>%s Compiling source in /var/tmp/portage/%s/work ...\n", green, normal, pkg)
	fmt.Printf(" * MAKEOPTS=\"%s\"  Ninja %s jobs, load limit %.1f\n", fakeJobs(), fakeJobs(), 8+rand.Float64()*8)
	sleep(200)

	sources := []string{
		"browser/ui/tab_strip", "browser/net/dns", "content/renderer/render_thread",
		"third_party/blink/renderer/core/dom/document", "third_party/webrtc/call/call",
		"third_party/ffmpeg/chromium/config/chromium/linux/x64/config",
		"components/sync/engine/sync_manager", "net/socket/ssl_client_socket",
		"ui/gfx/geometry/rect", "base/task/thread_pool", "media/gpu/vaapi_wrapper",
		"chrome/common/channel_info", "extensions/browser/extension_registry",
		"components/policy/core/common/policy_loader", "gpu/command_buffer/service/gles2",
		"cc/trees/layer_tree_host", "v8/src/objects/objects", "skia/src/core/SkCanvas",
		"base/files/file_util", "net/http/http_cache", "media/audio/pulse",
		"ui/views/controls/button", "components/payments/content/payment",
	}
	tools := []string{"[MOC]", "[RCC]", "[PROTOC]", "[GEN]"}

	totalFiles := 120 + rand.Intn(180) // 120-300 файлов
	// Компиляция забирает ~40% от BuildTime, остальное — фиксированные фазы
	// (они уже отмасштабированы через timeScale). Поэтому полное время ≈ BuildTime.
	totalTime := time.Duration(float64(t.BuildTime) * 0.4 * float64(time.Second))
	perFile := totalTime / time.Duration(totalFiles)
	if perFile < 1*time.Millisecond {
		perFile = 1 * time.Millisecond
	}

	compilers := []string{"x86_64-pc-linux-gnu-g++", "x86_64-pc-linux-gnu-gcc"}
	ccacheHits := 0
	start := time.Now()

	for i := 1; i <= totalFiles; i++ {
		pct := float64(i) * 100 / float64(totalFiles)
		src := sources[rand.Intn(len(sources))]
		if rand.Float32() < 0.4 {
			src = fmt.Sprintf("%s_%d", src, rand.Intn(5))
		}
		ext := ".cpp"
		if rand.Float32() < 0.2 {
			ext = ".c"
		}
		cc := compilers[0]
		if ext == ".c" {
			cc = compilers[1]
		}
		r := rand.Float32()
		switch {
		case r < 0.10:
			// moc/rcc/protoc строки
			fmt.Printf("%s %s/%s%s -> moc_%s.cpp\n", tools[rand.Intn(len(tools))], src, shortName(src), ext, shortName(src))
		case r < 0.25:
			fmt.Printf("[%3d%%] Building C object CMakeFiles/%s.dir/%s.o\n", int(pct), src, src[strings.LastIndex(src, "/")+1:])
		default:
			// иногда ccache
			ccache := ""
			if rand.Float32() < 0.12 {
				ccache = "ccache "
				ccacheHits++
			}
			flags := "-O2 -march=native -pipe -fno-plt -fPIC -Wall -Wno-deprecated"
			fmt.Printf("[%3d%%] %s%s %s -D_GNU_SOURCE %s/%s%s -c -o CMakeFiles/obj/%d.o\n",
				int(pct), ccache, cc, flags, src, shortName(src), ext, i)
		}

		// предупреждения / заметки
		if i%27 == 0 {
			fmt.Printf("%s%s: In member function ‘void Foo::Bar()’:\n", dim, shortName(src))
			fmt.Printf("%s%s:%d: warning: unused variable ‘x’ [-Wunused-variable]%s\n", dim, shortName(src), 100+rand.Intn(900), normal)
		}
		if i%53 == 0 {
			fmt.Printf("%s%s: note: use -Wno-unused to suppress%s\n", dim, shortName(src), normal)
		}
		if i%71 == 0 {
			fmt.Printf("%scc1plus: note: load average %.1f, %d jobs running%s\n", yellow, 4+rand.Float64()*12, 8+rand.Intn(16), normal)
		}

		time.Sleep(perFile)

		// прогресс-бар каждые ~10%
		if i%(totalFiles/10) == 0 || i == totalFiles {
			elapsed := time.Since(start).Truncate(time.Second)
			eta := time.Duration(float64(totalTime) * (1 - float64(i)/float64(totalFiles))).Truncate(time.Second)
			fmt.Printf("%s>>>%s %s [%d/%d] ETA %s elapsed %s | ccache hits: %d\n",
				green, normal, progressBar(int(pct), 20), i, totalFiles, eta, elapsed, ccacheHits)
		}
	}

	fmt.Printf("%s>>>%s Linking %s ...\n", green, normal, pkg)
	fmt.Printf("%s[100%%] Linking CXX executable %s%s\n", teal, shortName(t.EmergeName), normal)
	spin(900)
	// пара линковочных строк для реализма
	fmt.Printf("%sold: %s/usr/lib/gcc/x86_64-pc-linux-gnu/%s/../../../../x86_64-pc-linux-gnu/bin/ld: warning: discarding dynamic section .note.property%s\n",
		dim, dim, fakeGCC(), normal)
	fmt.Printf("[100%%] Built target %s\n", shortName(t.EmergeName))
	sleep(300)
}

// 8. Тесты
func phaseTest(pkg string) {
	fmt.Printf("%s>>>%s Testing %s ...\n", green, normal, pkg)
	fmt.Printf(" * emake check TEST_VERBOSE=1\n")
	sleep(300)
	tests := []string{"unit_tests", "browser_tests", "net_unittests", "components_unittests", "sandbox_linux_unittests"}
	passed := 0
	total := 0
	for _, tt := range tests {
		n := 5 + rand.Intn(20)
		for i := 0; i < 2; i++ {
			fmt.Printf("PASS: %s.%s_%d (%d ms)\n", tt, randomWord(), rand.Intn(500), rand.Intn(900))
			sleep(80)
		}
		passed += n
		total += n
	}
	if rand.Float32() < 0.25 {
		fmt.Printf("%sSKIP: %s.FlakyTest_1 (flaky, retry)%s\n", yellow, tests[0], normal)
		sleep(100)
	}
	fmt.Printf("%sAll %d tests passed, 0 failed%s\n", teal, passed, normal)
	sleep(200)
}

// 9. Установка в image
func phaseInstall(pkg string) {
	fmt.Printf("%s>>>%s Installing %s to /var/tmp/portage/%s/image\n", green, normal, pkg, pkg)
	sleep(200)
	steps := []string{
		" * DESTDIR=/var/tmp/portage/.../image ninja install",
		" * Stripping binaries (x86_64-pc-linux-gnu-strip --strip-unneeded)",
		" * ecompress: compressing docs in /usr/share/doc",
		" * einstalldocs: README.chromium, LICENSE",
		" * dosym /opt/" + shortName(pkg) + "/" + shortName(pkg) + " /usr/bin/" + shortName(pkg),
		" * fperms 4755 /opt/" + shortName(pkg) + "/chrome-sandbox (setuid)",
		" * Installing icons (16x16 ... 512x512) + desktop file",
	}
	for _, s := range steps {
		fmt.Println(s)
		sleep(150)
	}
	spin(600)
	fmt.Printf("%s>>>%s Completed installing into the image directory\n", green, normal)
	sleep(200)
}

// 10. elog / postinst хуки
func phasePostinst(pkg string, t BuildTarget) {
	fmt.Printf("%s>>>%s Emerging postinst for %s ...\n", green, normal, pkg)
	sleep(200)
	fmt.Printf(" * Messages for package %s:\n\n", pkg)
	msgs := []string{
		" * Chromium-based browsers need kernel USER_NS and PID_NS for sandbox.",
		" * Check: sysctl kernel.unprivileged_userns_clone=1",
		" * VAAPI: emerge media-libs/libva with USE=\"X wayland\" for HW decode.",
		" * If video is green, disable HW decode or update ffmpeg.",
		" * LINGUAS=\"" + fakeLinguas() + "\" enabled.",
		" * USE=\"" + t.Use + "\" — пересобери с -proprietary-codecs если нужен чистый chromium.",
	}
	for _, m := range msgs[:3+rand.Intn(3)] {
		fmt.Printf("%s%s%s\n", yellow, m, normal)
		sleep(120)
	}
	fmt.Println()
	hooks := []string{
		"Updating desktop mime database ... [ ok ]",
		"Updating shared mime info database ... [ ok ]",
		"Updating icon cache (gtk-update-icon-cache-3.0) ... [ ok ]",
		"Updating .desktop database (update-desktop-database) ... [ ok ]",
		"Updating man database (mandb -q) ... [ ok ]",
		"Running ldconfig ... [ ok ]",
	}
	for _, h := range hooks {
		fmt.Printf(" * %s\n", h)
		sleep(120)
	}
	// QA-заметки
	if rand.Float32() < 0.6 {
		fmt.Printf("%s * QA Notice: package installs setuid file: /opt/%s/chrome-sandbox%s\n", yellow, shortName(pkg), normal)
		sleep(150)
	}
	fmt.Printf("%s * GNU info directory index is up-to-date.%s\n", green, normal)
	sleep(200)
}

// 11. Мердж + очистка
func phaseMerge(pkg string, start time.Time) {
	fmt.Printf("%s>>>%s Merging %s to /\n", green, normal, pkg)
	for i := 0; i < 6; i++ {
		fmt.Printf(">>> /opt/%s/%s-%d\n>>> /usr/share/applications/%s.desktop\n>>> /usr/share/icons/hicolor/256x256/apps/%s.png\n",
			shortName(pkg), shortName(pkg), rand.Intn(100), shortName(pkg), shortName(pkg))
		sleep(100)
	}
	fmt.Printf("%s>>>%s Recording %s in \"world\" favorites file...\n", green, normal, pkg)
	sleep(250)
	elapsed := time.Since(start).Truncate(time.Second)
	fmt.Printf("%s>>>%s %s merged in %s (load avg: %.2f)\n", green, normal, pkg, elapsed, 2+rand.Float64()*8)
	fmt.Printf("%s>>>%s Auto-cleaning packages...\n", green, normal)
	sleep(300)
	fmt.Printf("%s>>>%s No outdated packages were found on your system.\n", green, normal)
	fmt.Printf("%s>>>%s Checking for preserved libraries ...%s none found\n", green, normal, normal)
	sleep(200)
	fmt.Printf("%s>>>%s Checking for revdep-rebuild ...%s no broken revdeps\n", green, normal, normal)
	sleep(200)
	fmt.Println()
	fmt.Printf(" * %sIMPORTANT:%s 1 news item needs reading for repository 'gentoo'.\n", yellow, normal)
	fmt.Printf(" * Use `eselect news read` to view new items.\n")
	fmt.Printf(" * %sAfter world updates, it is important to remove obsolete packages with `emerge --depclean`.%s\n", dim, normal)
	fmt.Printf(" * %sIt is also important to run `revdep-rebuild`.%s\n\n", dim, normal)
}

// splitCommand режет строку команды на программу + аргументы,
// понимает двойные кавычки:  code --new-window "/my dir/proj"
func splitCommand(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	esc := false
	for _, r := range s {
		switch {
		case esc:
			cur.WriteRune(r)
			esc = false
		case r == '\\' && inQuote:
			esc = true
		case r == '"':
			inQuote = !inQuote
		case (r == ' ' || r == '\t') && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// ---------------- ЗАПУСК ----------------

func launch(t BuildTarget) {
	// Command может содержать аргументы прямо в строке:
	//   Command: "code --new-window"
	//   Command: "firefox --private-window"
	// Режем на программу + аргументы (понимаем двойные кавычки).
	parts := splitCommand(t.Command)
	if len(parts) == 0 {
		fmt.Printf("%s * %sПустой Command у этого алиаса%s\n", red, normal, normal)
		os.Exit(1)
	}
	t.Command = parts[0]
	t.Args = append(parts[1:], t.Args...)

	// Режим по умолчанию берётся из алиаса (поле Foreground).
	// Флаги перекрывают его для одного запуска:
	//   --fg / -f / --foreground = в том же терминале, с логами
	//   --bg / --background       = в фоне, без логов
	foreground := t.Foreground
	filtered := make([]string, 0, len(t.Args))
	for _, a := range t.Args {
		switch a {
		case "--fg", "-f", "--foreground":
			foreground = true
		case "--bg", "--background":
			foreground = false
		default:
			filtered = append(filtered, a)
		}
	}
	t.Args = filtered

	if foreground {
		launchForeground(t)
		return
	}
	launchBackground(t)
}

func launchBackground(t BuildTarget) {
	//fmt.Printf("%s>>>%s Launching `%s %s` in background (no logs) ...\n",
	//	green, normal, t.Command, strings.Join(t.Args, " "))

	// /dev/null для stdin/stdout/stderr — ничего не сыпется в терминал
	nullR, errR := os.Open(os.DevNull)
	nullW, errW := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if errR != nil || errW != nil {
		fmt.Printf("%s * %sCannot open /dev/null, fallback to foreground%s\n", red, normal, normal)
		launchForeground(t)
		return
	}
	defer nullR.Close()
	defer nullW.Close()

	cmd := exec.Command(t.Command, t.Args...)
	cmd.Stdin = nullR
	cmd.Stdout = nullW
	cmd.Stderr = nullW
	// Отцепляемся от терминала: своя сессия, SIGHUP/Ctrl+C нас не убивают
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no such file") {
			fmt.Printf("%s * %sCommand '%s' not found. Поставь программу или поменяй Command в main.go%s\n", red, normal, t.Command, normal)
		} else {
			fmt.Printf("%s * %sFailed to start: %v%s\n", red, normal, err, normal)
		}
		os.Exit(1)
	}

	//fmt.Printf("%s>>>%s Started PID %d, терминал свободен. Логи скрыты (/dev/null).\n\n", green, normal, cmd.Process.Pid)
	// Не ждём процесс — сразу выходим, он живёт сам (setsid).
	// Reap не нужен: init подберёт зомби, терминал не блокируется.
}

func launchForeground(t BuildTarget) {
	fmt.Printf("%s>>>%s Launching `%s %s` in foreground ...\n\n", green, normal, t.Command, strings.Join(t.Args, " "))
	cmd := exec.Command(t.Command, t.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no such file") {
			fmt.Printf("%s * %sCommand '%s' not found. Поставь программу или поменяй Command в main.go%s\n", red, normal, t.Command, normal)
		} else {
			fmt.Printf("%s * %sProgram exited with error: %v%s\n", red, normal, err, normal)
		}
		os.Exit(1)
	}
}

// ---------------- helpers ----------------

func sleep(ms int) { time.Sleep(time.Duration(float64(ms)*timeScale) * time.Millisecond) }

func spin(d time.Duration) {
	d = time.Duration(float64(d) * timeScale)
	chars := []string{"|", "/", "-", "\\"}
	end := time.Now().Add(d)
	i := 0
	for time.Now().Before(end) {
		fmt.Printf("\r%s ", chars[i%len(chars)])
		time.Sleep(100 * time.Millisecond)
		i++
	}
	fmt.Printf("\r   \r")
}

func dots(msg string, d time.Duration) {
	fmt.Printf("%s... ", msg)
	d = time.Duration(float64(d) * timeScale)
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		fmt.Printf(".")
		time.Sleep(120 * time.Millisecond)
	}
	fmt.Printf(" done!\n")
}

func progressBar(pct, width int) string {
	filled := pct * width / 100
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "#"
		} else if i == filled {
			bar += ">"
		} else {
			bar += "-"
		}
	}
	bar += fmt.Sprintf("] %3d%%", pct)
	return bar
}

func downloadBar(d time.Duration, totalKB int) {
	d = time.Duration(float64(d) * timeScale)
	steps := 20
	start := time.Now()
	for i := 1; i <= steps; i++ {
		pct := i * 100 / steps
		done := totalKB * i / steps
		speed := 4000 + rand.Intn(18000)
		elapsed := time.Since(start).Truncate(time.Second)
		fmt.Printf("\r%3d%%[%s] %dK / %dK  %dK/s  ETA %ds  %s",
			pct, progressBarInner(i, steps), done, totalKB, speed, (steps-i)/4, elapsed)
		time.Sleep(d / time.Duration(steps))
	}
	fmt.Printf("\n")
}

func progressBarInner(i, total int) string {
	s := ""
	for j := 0; j < total; j++ {
		if j < i {
			s += "="
		} else if j == i {
			s += ">"
		} else {
			s += " "
		}
	}
	return s
}

func shortName(s string) string {
	if idx := strings.LastIndex(s, "/"); idx >= 0 {
		return s[idx+1:]
	}
	if idx := strings.LastIndex(s, "-"); idx >= 0 && len(s)-idx < 12 {
		return s[:idx]
	}
	return s
}

func pick(a []string) string { return a[rand.Intn(len(a))] }

func randomDir() string {
	return pick([]string{"src", "build", "out/Release", "chrome", "third_party"})
}

func randomWord() string {
	return pick([]string{"Render", "Sync", "Policy", "Tab", "DNS", "GPU", "Audio", "Net", "UI", "Sandbox"})
}

func randomUse() string {
	return pick([]string{
		"X wayland", "gtk3 -qt5", "system-ffmpeg", "pulseaudio alsa",
		"openh264 proprietary-codecs", "lto pgo",
	})
}

func fakeDeps(cat string) []string {
	pool := []string{
		"dev-libs/nss-3.98", "dev-libs/nspr-4.35", "media-libs/alsa-lib-1.2.10",
		"media-libs/mesa-24.1.3", "media-video/ffmpeg-6.1.2", "sys-libs/zlib-1.3.1",
		"dev-libs/expat-2.6.2", "net-print/cups-2.4.8", "x11-libs/gtk+-3.24.42",
		"dev-libs/glib-2.80.4", "media-libs/harfbuzz-9.0.0", "dev-libs/icu-75.1",
	}
	n := 3 + rand.Intn(4)
	var out []string
	for i := 0; i < n; i++ {
		out = append(out, pool[rand.Intn(len(pool))])
	}
	return out
}

func fakeCPU() string {
	return pick([]string{"AMD Ryzen 9 9950X3D", "Intel i9-14900KF", "AMD Ryzen 5 9600X", "Intel i5-12400F"})
}

func fakeJobs() string { return fmt.Sprintf("-j%d -l%d", 8+rand.Intn(16), 8+rand.Intn(8)) }
func fakeGCC() string  { return pick([]string{"13.3.1", "14.2.1", "12.4.0"}) }
func fakeGlibc() string {
	return pick([]string{"2.39-r6", "2.40-r2"})
}
func fakeKernel() string { return pick([]string{"6.10.9-gentoo", "6.6.47-gentoo"}) }
func fakeFeatures() string {
	return pick([]string{"ccache parallel-fetch parallel-install sandbox userpriv", "compressdebug network-sandbox parallel-fetch preserve-libs"})
}
func fakeCPUFlags() string { return "aes avx avx2 bmi1 f16c mmx popcnt sse4_2 ssse3" }
func fakeLinguas() string  { return pick([]string{"en ru de fr", "en ru", "en ru uk"}) }
