@echo off
rem Startet den XML Clipboard Formatter.
rem
rem Beim ersten Start und nach Aenderungen an den Quellen wird gebaut, danach
rem startet das Programm. Die ausfuehrbare Datei liegt in build\.
rem
rem   start.bat            starten, vorher bei Bedarf bauen
rem   start.bat -Neu       vorher neu bauen, auch wenn alles aktuell ist
rem   start.bat -Pruefen   nur die Umgebung pruefen, weder bauen noch starten
rem
rem Weitere Argumente werden an das Programm durchgereicht.
setlocal EnableExtensions
cd /d "%~dp0"

set "APP_NAME=XMLClipboardFormatter"
set "APP=build\%APP_NAME%.exe"
set "GO_MIN=1.24"
set "NEUBAUEN="
set "PRUEFEN="
set "APP_ARGUMENTE="
set "ANZAHL=0"

rem --- Argumente ------------------------------------------------------------
rem Nicht "if ... set ... & shift": das & trennt die Befehle auch dann, wenn
rem die Bedingung nicht zutrifft. Dann gingen Argumente still verloren.
:argumente
if "%~1"=="" goto goSuchen
if /i "%~1"=="-Neu" goto neu
if /i "%~1"=="--neu" goto neu
if /i "%~1"=="-Pruefen" goto pruefen
if /i "%~1"=="--pruefen" goto pruefen
set "APP_ARGUMENTE=%APP_ARGUMENTE% %1"
set /a ANZAHL+=1
shift
goto argumente

:neu
set "NEUBAUEN=1"
shift
goto argumente

:pruefen
set "PRUEFEN=1"
shift
goto argumente

rem --- Go suchen ------------------------------------------------------------
rem Reihenfolge: XMLFORMATTER_GO, GOROOT, PATH, die ueblichen Installationsorte.
rem Die Variable ist fuer Rechner gedacht, auf denen Go nicht im PATH steht.
:goSuchen
set "GO="
set "GO_QUELLE="
if defined XMLFORMATTER_GO if exist "%XMLFORMATTER_GO%" (
    set "GO=%XMLFORMATTER_GO%"
    set "GO_QUELLE=XMLFORMATTER_GO"
)
if not defined GO if defined GOROOT if exist "%GOROOT%\bin\go.exe" (
    set "GO=%GOROOT%\bin\go.exe"
    set "GO_QUELLE=GOROOT"
)
if not defined GO for /f "delims=" %%g in ('where go 2^>nul') do if not defined GO (
    set "GO=%%g"
    set "GO_QUELLE=PATH"
)
if not defined GO if exist "%ProgramFiles%\Go\bin\go.exe" (
    set "GO=%ProgramFiles%\Go\bin\go.exe"
    set "GO_QUELLE=Program Files"
)
if not defined GO if exist "%LOCALAPPDATA%\Programs\Go\bin\go.exe" (
    set "GO=%LOCALAPPDATA%\Programs\Go\bin\go.exe"
    set "GO_QUELLE=LocalAppData"
)
if not defined GO (
    echo Fehler: Go wurde nicht gefunden. Go %GO_MIN% oder neuer installieren:
    echo         https://go.dev/dl/  - danach ein neues Terminal oeffnen.
    exit /b 1
)

rem Die Version ueber PowerShell zu holen ist umstaendlicher als "go version",
rem funktioniert aber auch, wenn der Pfad zu Go Leerzeichen enthaelt.
set "XMLFMT_GO=%GO%"
set "GOVERSION="
for /f "usebackq delims=" %%v in (`powershell -NoProfile -Command "& $env:XMLFMT_GO env GOVERSION" 2^>nul`) do set "GOVERSION=%%v"
if not defined GOVERSION (
    for /f "usebackq tokens=3 delims= " %%v in (`powershell -NoProfile -Command "& $env:XMLFMT_GO version" 2^>nul`) do set "GOVERSION=%%v"
)
if not defined GOVERSION (
    echo Fehler: Go wurde gefunden, meldet aber keine Version.
    exit /b 1
)

powershell -NoProfile -Command "$m=[regex]::Match($env:GOVERSION,'^go(\d+)\.(\d+)'); if (-not $m.Success) { exit 1 }; $a=[int]$m.Groups[1].Value; $b=[int]$m.Groups[2].Value; if (($a -lt 1) -or (($a -eq 1) -and ($b -lt 24))) { exit 1 }"
if errorlevel 1 (
    echo Fehler: Go %GOVERSION% ist zu alt. Mindestens Go %GO_MIN% wird gebraucht.
    echo         https://go.dev/dl/
    exit /b 1
)

rem --- C-Compiler suchen ----------------------------------------------------
rem Fyne zeichnet mit OpenGL und braucht deshalb cgo und einen C-Compiler.
rem Unter Windows ist das mingw-w64 (gcc), zum Beispiel aus MSYS2.
set "CCBIN="
for /f "delims=" %%c in ('where gcc 2^>nul') do if not defined CCBIN set "CCBIN=%%c"
if not defined CCBIN for /f "delims=" %%c in ('where clang 2^>nul') do if not defined CCBIN set "CCBIN=%%c"

rem --- Muss gebaut werden? --------------------------------------------------
rem Der Vergleich laeuft ueber PowerShell, weil %~t von der Systemsprache
rem abhaengt und ein Textvergleich je nach Rechner falsch entscheiden wuerde.
set "BAUEN="
if defined NEUBAUEN set "BAUEN=1"
if not exist "%APP%" set "BAUEN=1"
if not defined BAUEN (
    set "XMLFMT_APP=%APP%"
    for /f "usebackq delims=" %%b in (`powershell -NoProfile -Command "$exe=(Get-Item -LiteralPath $env:XMLFMT_APP).LastWriteTime; $neu=$false; foreach ($f in (Get-ChildItem -LiteralPath . -Recurse -File)) { if (($f.Name -like '*.go') -or ($f.Name -eq 'go.mod') -or ($f.Name -eq 'go.sum') -or ($f.Name -eq 'start.bat')) { if ($f.LastWriteTime -gt $exe) { $neu=$true } } }; if ($neu) { 'BAUEN' }" 2^>nul`) do set "BAUEN=1"
)

if defined PRUEFEN goto pruefAusgabe

if defined BAUEN if not defined CCBIN (
    echo Fehler: Kein C-Compiler gefunden. Fuer Fyne wird mingw-w64 gebraucht,
    echo         zum Beispiel gcc aus MSYS2: https://www.msys2.org/
    exit /b 1
)
if defined BAUEN goto bauen
goto starten

rem --- Pruefmodus -----------------------------------------------------------
:pruefAusgabe
echo Go      : %GO%  (%GO_QUELLE%)
echo Version : %GOVERSION% (mindestens Go %GO_MIN%)
if defined CCBIN (echo Compiler: %CCBIN%) else (echo Compiler: nicht gefunden)
if exist "%APP%" (echo App     : %APP% - vorhanden) else (echo App     : %APP% - fehlt)
if defined BAUEN (echo Bauen   : ja) else (echo Bauen   : nein, Quellen sind nicht neuer)
echo Argumente: %ANZAHL% an das Programm
if defined APP_ARGUMENTE echo            %APP_ARGUMENTE%
exit /b 0

rem --- Bauen ----------------------------------------------------------------
:bauen
echo Bauen ...
if not exist build mkdir build
set "DATUM="
for /f "usebackq delims=" %%d in (`powershell -NoProfile -Command "Get-Date -Format 'dd.MM.yyyy'"`) do set "DATUM=%%d"
if not defined DATUM set "DATUM=%DATE%"
set "LDFLAGS=-H=windowsgui -X main.versionDate=%DATUM%"
set "CGO_ENABLED=1"
call "%GO%" build -trimpath -ldflags "%LDFLAGS%" -o "%APP%" .
if errorlevel 1 (
    echo Fehler: der Build ist fehlgeschlagen.
    exit /b 1
)

rem --- Starten --------------------------------------------------------------
:starten
if not exist "%APP%" (
    echo Fehler: "%APP%" fehlt trotz Bauvorgang.
    exit /b 1
)
"%APP%"%APP_ARGUMENTE%
exit /b %ERRORLEVEL%
