#!/bin/zsh
#
# Doppelklick-Starter für macOS. Öffnet ein Terminalfenster und ruft start.sh
# mit denselben Argumenten auf. Der Pfad wird relativ zu dieser Datei
# bestimmt, damit der Ordner beliebig verschoben werden kann.
cd "$(dirname "$0")" || exit 1
exec ./start.sh "$@"
