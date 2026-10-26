#!/usr/bin/env bash
#
# Startet den XML Clipboard Formatter.
#
# Beim ersten Start und nach Änderungen an den Quellen wird gebaut, danach
# startet das Programm. Die ausführbare Datei liegt in build/.
#
#   ./start.sh            starten, vorher bei Bedarf bauen
#   ./start.sh -Neu       vorher neu bauen, auch wenn alles aktuell ist
#   ./start.sh -Pruefen   nur die Umgebung prüfen, weder bauen noch starten
#
# Weitere Argumente werden an das Programm durchgereicht.
#
# Auf macOS lässt sich das Programm auch per Doppelklick über start.command
# starten.

set -euo pipefail

cd "$(dirname "$0")"

APP_NAME="XMLClipboardFormatter"
APP="build/$APP_NAME"
GO_MIN="1.24"

fehler() {
    echo "Fehler: $*" >&2
    exit 1
}

# --- Argumente --------------------------------------------------------------
NEUBAUEN=0
PRUEFEN=0
APP_ARGUMENTE=()
for arg in "$@"; do
    case "$arg" in
        -Neu|--neu)
            NEUBAUEN=1
            ;;
        -Pruefen|--pruefen)
            PRUEFEN=1
            ;;
        *)
            APP_ARGUMENTE+=("$arg")
            ;;
    esac
done
set -- ${APP_ARGUMENTE[@]+"${APP_ARGUMENTE[@]}"}

# --- Go suchen --------------------------------------------------------------
# Reihenfolge: XMLFORMATTER_GO, GOROOT, PATH. Die Variable ist für Rechner
# gedacht, auf denen Go an einer Stelle liegt, die nicht im PATH steht.
GO=""
GO_QUELLE=""
if [ -n "${XMLFORMATTER_GO:-}" ] && [ -x "$XMLFORMATTER_GO" ]; then
    GO="$XMLFORMATTER_GO"
    GO_QUELLE="XMLFORMATTER_GO"
elif [ -n "${GOROOT:-}" ] && [ -x "$GOROOT/bin/go" ]; then
    GO="$GOROOT/bin/go"
    GO_QUELLE="GOROOT"
elif kandidat="$(command -v go 2>/dev/null)" && [ -n "$kandidat" ]; then
    GO="$kandidat"
    GO_QUELLE="PATH"
fi
[ -n "$GO" ] || fehler "Go wurde nicht gefunden. Go $GO_MIN oder neuer installieren: https://go.dev/dl/"

GO_VERSION="$("$GO" env GOVERSION 2>/dev/null || true)"
[ -n "$GO_VERSION" ] || GO_VERSION="$("$GO" version 2>/dev/null | awk '{print $3}' || true)"

# "go1.26.1" in zwei Zahlen zerlegen. Ein Entwicklungsstand wie
# "go1.27-devel" oder "devel go1.27-..." endet nach der Minorzahl; das reicht
# für den Vergleich.
GO_ZAHL="${GO_VERSION#devel }"
GO_ZAHL="${GO_ZAHL#go}"
if [[ "$GO_ZAHL" =~ ^([0-9]+)\.([0-9]+) ]]; then
    GO_MAJOR="${BASH_REMATCH[1]}"
    GO_MINOR="${BASH_REMATCH[2]}"
else
    fehler "Die Go-Version '$GO_VERSION' lässt sich nicht lesen."
fi
if [ "$GO_MAJOR" -lt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 24 ]; }; then
    fehler "Go $GO_VERSION ist zu alt. Mindestens Go $GO_MIN wird gebraucht."
fi

# --- C-Compiler suchen ------------------------------------------------------
# Fyne zeichnet mit OpenGL und braucht deshalb cgo und einen C-Compiler.
# Unter macOS liefert ihn Xcode Command Line Tools; "cc" und "clang" zeigen
# beide darauf.
CC_BIN=""
if [ -n "${CC:-}" ] && command -v "$CC" >/dev/null 2>&1; then
    CC_BIN="$(command -v "$CC")"
elif kandidat="$(command -v clang 2>/dev/null)" && [ -n "$kandidat" ]; then
    CC_BIN="$kandidat"
elif kandidat="$(command -v cc 2>/dev/null)" && [ -n "$kandidat" ]; then
    CC_BIN="$kandidat"
fi

# --- Muss gebaut werden? ----------------------------------------------------
BAUEN=0
if [ "$NEUBAUEN" = 1 ] || [ ! -f "$APP" ]; then
    BAUEN=1
elif [ -n "$(find . -type f \
        \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' -o -name 'start.sh' \) \
        -newer "$APP" -print -quit 2>/dev/null)" ]; then
    BAUEN=1
fi
if [ "$BAUEN" = 1 ] && [ -z "$CC_BIN" ]; then
    fehler "Kein C-Compiler gefunden. Unter macOS hilft: xcode-select --install"
fi

# --- Prüfmodus --------------------------------------------------------------
if [ "$PRUEFEN" = 1 ]; then
    echo "Go      : $GO  ($GO_QUELLE)"
    echo "Version : $GO_VERSION (mindestens Go $GO_MIN)"
    echo "Compiler: ${CC_BIN:-nicht gefunden}"
    if [ -f "$APP" ]; then
        echo "App     : $APP - vorhanden"
    else
        echo "App     : $APP - fehlt"
    fi
    if [ "$BAUEN" = 1 ]; then
        echo "Bauen   : ja"
    else
        echo "Bauen   : nein, Quellen sind nicht neuer"
    fi
    echo "Argumente: $# an das Programm"
    for a in "$@"; do
        echo "  [$a]"
    done
    exit 0
fi

# --- Bauen ------------------------------------------------------------------
if [ "$BAUEN" = 1 ]; then
    echo "Bauen ..."
    mkdir -p build
    DATUM="$(date '+%d.%m.%Y')"
    CGO_ENABLED=1 "$GO" build -trimpath \
        -ldflags "-X main.versionDate=$DATUM" \
        -o "$APP" .
fi

[ -f "$APP" ] || fehler "$APP fehlt trotz Bauvorgang."

# --- Starten ----------------------------------------------------------------
exec "$APP" "$@"
