# Anleitung für Windows

Diese Anleitung richtet sich an alle, die den XML Clipboard Formatter auf einem
Windows-Rechner benutzen wollen. Sie setzt keine Programmierkenntnisse voraus —
nur eine Eingabeaufforderung, in die Sie einen Befehl eintippen können.

> **Sie suchen die Anleitung für macOS?** Dann nehmen Sie
> [anleitung-macos.md](anleitung-macos.md).

---

## 1. Wofür ist das Programm?

Sie markieren XML in einem beliebigen Programm, kopieren es mit **Strg+C** in
die Zwischenablage und formatieren es mit einem Klick:

```xml
<bestellung><kunde><name>Müller & Söhne</name></kunde></bestellung>
```

wird zu:

```xml
<bestellung>
    <kunde>
        <name>Müller & Söhne</name>
    </kunde>
</bestellung>
```

Das Ergebnis liegt danach wieder in der Zwischenablage, sodass Sie es mit
**Strg+V** an der gewünschten Stelle einsetzen können. Sie brauchen dafür kein
Programm, das XML kennt, und keine Internetverbindung.

Genauso geht es umgekehrt: **XML Packen** entfernt den Leerraum zwischen den
Elementen. Das ist die Form, die viele Werkzeuge erwarten, wenn XML in eine
Zeile oder in ein Feld ohne Umbrüche soll.

## 2. Was Sie brauchen

|                        |                                                         |
| ---------------------- | ------------------------------------------------------- |
| **Windows 10 oder 11** | ältere Fassungen werden nicht geprüft                   |
| **Go 1.24 oder neuer** | zwingend, das Programm ist in Go geschrieben            |
| **mingw-w64 (gcc)**    | zwingend, Fyne braucht cgo und einen C-Compiler         |
| **Git**                | nur zum Holen des Quelltexts; ein ZIP funktioniert auch |

PowerShell ist ab Windows 10 vorhanden und wird von `start.bat` für die
Versions- und Datumsabfrage benutzt. Sie müssen nichts daran einstellen.

Go und mingw-w64 sind nur zum **Bauen** nötig. Wenn die `.exe` einmal gebaut
ist, startet sie auch ohne Go. `start.bat` sucht Go aber immer, auch wenn
nichts zu bauen ist; ohne Go starten Sie die fertige
`build\XMLClipboardFormatter.exe` deshalb direkt.

## 3. Schritt 1: Go installieren

Laden Sie den Installer von [go.dev/dl](https://go.dev/dl/) herunter
(„Microsoft Windows“, `.msi`) und führen Sie ihn aus. Die Standardeinstellungen
passen.

Öffnen Sie danach ein **neues** Eingabeaufforderungsfenster — ein bereits
offenes kennt den neuen `PATH` nicht — und prüfen Sie:

```bat
go version
```

Sieht die Antwort so oder ähnlich aus, ist alles in Ordnung:

```
go version go1.26.1 windows/amd64
```

Steht dort `'go' is not recognized` oder eine Versionszahl unter 1.24, fehlt Go
oder ist zu alt. Installieren Sie es erneut und öffnen Sie ein neues Terminal.

## 4. Schritt 2: mingw-w64 installieren

Fyne zeichnet mit OpenGL und braucht deshalb cgo. Auf Windows kommt der
C-Compiler aus mingw-w64. Der bequemste Weg ist [MSYS2](https://www.msys2.org/):

1. MSYS2 herunterladen, installieren und im Startmenü **MSYS2 MINGW64**
   öffnen. (Nicht die „MSYS2 MSYS“-Shell.)

2. In dieser Shell eingeben:
   
   ```sh
   pacman -S --needed base-devel mingw-w64-x86_64-gcc
   ```
   
   Die Frage nach der Installation mit Enter bestätigen. Das Paket ist einige
   hundert Megabyte gross; der Download dauert einen Moment.

3. Das Fenster schliessen.

Jetzt muss der Compiler auch ausserhalb der MSYS2-Shell gefunden werden. Der
übliche Pfad ist `C:\msys64\mingw64\bin`.

**PATH ergänzen:**

1. Windows-Taste drücken, „Umgebungsvariablen“ eintippen und
   **Systemumgebungsvariablen bearbeiten** öffnen.
2. Unten bei **Benutzervariablen** den Eintrag `Path` auswählen und
   **Bearbeiten** anklicken.
3. **Neu** wählen und `C:\msys64\mingw64\bin` eintragen.
4. Mit **OK** schliessen und ein **neues** Eingabeaufforderungsfenster öffnen.

Prüfen:

```bat
gcc --version
```

Die erste Zeile nennt mingw-w64 und eine Versionsnummer. Kommt stattdessen
`'gcc' is not recognized`, steht der Pfad noch nicht im `PATH`, oder es wurde
ein anderer Ordner installiert. Prüfen Sie mit `where gcc`, welcher Compiler
gefunden wird. Steht dort ein anderer als der aus `C:\msys64\mingw64\bin`,
schieben Sie den MSYS2-Eintrag in der PATH-Liste nach oben.

> Es gibt auch fertige mingw-w64-Builds ausserhalb von MSYS2. Wichtig ist nur:
> `gcc` muss in der Eingabeaufforderung funktionieren, und es muss ein
> 64-Bit-Compiler sein.

## 5. Schritt 3: Das Programm holen

Wechseln Sie in einen Ordner, in dem das Programm liegen darf, und holen Sie es
mit Git:

```bat
git clone https://github.com/BENUTZERNAME/xmlClipboardFormatter.git
cd xmlClipboardFormatter
```

`BENUTZERNAME` steht für den GitHub-Namen des Repository-Inhabers.

Wer kein Git hat, lädt auf der Projektseite **Code → Download ZIP**, entpackt
die Datei (Rechtsklick → **Alle extrahieren**) und wechselt im Terminal in den
entpackten Ordner.

## 6. Schritt 4: Einmal bauen und starten

Ein **Doppelklick auf `start.bat`** startet das Programm. Beim ersten Mal
schreibt das Skript `Bauen ...` und lädt die Fyne-Module nach. Das dauert je
nach Verbindung ein bis drei Minuten. Danach öffnet sich das Fenster
**XML Clipboard Formatter**.

Alternativ in einem Eingabeaufforderungsfenster:

```bat
cd C:\Pfad\zum\xmlClipboardFormatter
start.bat
```

Die ausführbare Datei liegt danach unter:

```
build\XMLClipboardFormatter.exe
```

Sie können sie auch direkt starten. Für den normalen Gebrauch ist `start.bat`
der bequemere Weg: Es baut automatisch nach, wenn Sie oder jemand anderes am
Quelltext etwas geändert hat.

### Wenn sich das Fenster sofort schliesst

Dann ist beim Bauen etwas schiefgegangen, und die Meldung war zu schnell weg.
Öffnen Sie die Eingabeaufforderung, wechseln Sie in den Ordner und starten Sie:

```bat
start.bat
```

Jetzt bleibt die Ausgabe stehen. Die häufigsten Ursachen stehen unten unter
„Wenn etwas nicht klappt“.

### Was die Schalter können

| Befehl               | Wirkung                                                        |
| -------------------- | -------------------------------------------------------------- |
| `start.bat`          | starten, vorher bei Bedarf bauen                               |
| `start.bat -Neu`     | vorher neu bauen, auch wenn alles aktuell ist, dann starten    |
| `start.bat -Pruefen` | nur nachsehen, ob die Umgebung passt — baut und startet nichts |

`-Pruefen` gibt ungefähr das aus:

```
Go      : C:\Program Files\Go\bin\go.exe  (PATH)
Version : go1.26.1 (mindestens Go 1.24)
Compiler: C:\msys64\mingw64\bin\gcc.exe
App     : build\XMLClipboardFormatter.exe - vorhanden
Bauen   : nein, Quellen sind nicht neuer
Argumente: 0 an das Programm
```

Steht bei `App` „fehlt“ oder bei `Bauen` „ja“, ist das kein Fehler: Der nächste
Start ohne `-Pruefen` baut dann. Steht bei `Compiler` „nicht gefunden“, fehlt
mingw-w64 oder der `PATH`-Eintrag.

### Wenn Go an einer ungewöhnlichen Stelle liegt

`start.bat` sucht Go in dieser Reihenfolge: die Umgebungsvariable
`XMLFORMATTER_GO`, `%GOROOT%\bin\go.exe`, den `PATH`, `C:\Program Files\Go` und
`%LOCALAPPDATA%\Programs\Go`. Für einen einzelnen Start in PowerShell:

```powershell
$env:XMLFORMATTER_GO = 'C:\Program Files\Go\bin\go.exe'
.\start.bat
```

Oder in der Eingabeaufforderung:

```bat
set "XMLFORMATTER_GO=C:\Program Files\Go\bin\go.exe"
start.bat
```

Für dauerhafte Nutzung setzen Sie die Variable über
**Systemumgebungsvariablen bearbeiten** bei den Benutzervariablen. Danach ein
neues Terminal öffnen.

## 7. Bedienung

![Oberfläche im hellen Design](bild-hell.png)

Das Fenster trägt den Titel **XML Clipboard Formatter**. Rechts oben steht der
Knopf zum Umschalten zwischen dunkel und hell. Darunter liegt das grosse
Textfeld mit den Zeilennummern links. Ganz unten stehen die Aktionsknöpfe,
darunter die Statuszeile und rechts die Versionszeile, zum Beispiel
`Version 1.0.0 vom 26.10.2026`.

### Die fünf Knöpfe

| Knopf                       | Wirkung                                          |
| --------------------------- | ------------------------------------------------ |
| **Clipboard einlesen**      | holt den Text aus der Zwischenablage in das Feld |
| **XML Formatieren**         | macht den Text lesbar, ein Element je Zeile      |
| **XML Packen**              | entfernt den Leerraum zwischen den Elementen     |
| **Ins Clipboard schreiben** | legt den Text aus dem Feld in die Zwischenablage |
| **Beenden**                 | beendet das Programm; steht rechts als Textlink  |

**Das sind zwei Schritte, nicht einer:** *XML Formatieren* und *XML Packen*
legen nichts in die Zwischenablage. Erst *Ins Clipboard schreiben* tut das. So
können Sie sich das Ergebnis ansehen, bevor Sie es irgendwo einsetzen.
*Ins Clipboard schreiben* schneidet dabei XML-Leerraum – Leerzeichen, Tabs und
Zeilenumbrüche – am Anfang und am Ende ab; ein geschütztes Leerzeichen bleibt
stehen. Das Textfeld bleibt unverändert.

**Formatieren oder Packen?** Beides liest denselben Text und schreibt ihn in
anderer Gestalt zurück. *Formatieren* ist für Menschen, *Packen* für Maschinen:
etwa für eine Befehlszeile, eine `.env`-Datei oder ein Feld ohne Umbrüche.

### Normaler Arbeitsablauf

1. XML in Ihrem Programm markieren und mit **Strg+C** kopieren
2. **Clipboard einlesen** → `✓ XML geladen. Formatieren, Packen und Schreiben sind freigeschaltet.`
3. **XML Formatieren** → `✓ XML lesbar formatiert.`
4. **Ins Clipboard schreiben** → `✓ XML erfolgreich in die Zwischenablage kopiert.`
5. In Ihr Programm zurück, **Strg+V** einfügen

War beim Formatieren nichts zu tun, meldet das Programm das ausdrücklich:
`✓ War schon lesbar formatiert.` Beim Packen entsprechend
`✓ War schon gepackt.` Erfolgsmeldungen verschwinden nach vier Sekunden.
Warnungen und Fehler bleiben stehen, bis Sie wieder etwas tun.

Steht in der Zwischenablage Text, der nicht nach XML aussieht, meldet
**Clipboard einlesen**:

```
⚠ Text geladen, sieht aber nicht nach XML aus. Aktionen blockiert.
```

Der Text steht trotzdem im Feld — Sie können ihn ansehen und von Hand ändern.
Ist die Zwischenablage leer, bleibt das Feld unverändert und die Statuszeile
sagt `⚠ Die Zwischenablage ist leer.`

### Die Statuszeile

Die Statuszeile ist das Einzige, was Ihnen sagt, ob die Knöpfe etwas tun.
Formatieren, Packen und Schreiben sind genau dann grau, wenn der Text nicht
nach XML aussieht. **Clipboard einlesen** bleibt immer anklickbar — Sie können
also jederzeit nachsehen, was in der Zwischenablage steht.

Erkannt wird ein Text am ersten Zeichen nach einem BOM und Leerraum: Es muss
`<` sein, danach muss `!`, `?` oder ein Namensanfang stehen. `<r/>`,
`<?xml ...?>` und `<!-- ... -->` gelten damit als XML. `Einkaufsliste`, `{`
und `</r>` gelten nicht als XML. Das ist eine Vermutung, keine Prüfung: Der
Formatter liest beim Formatieren den ganzen Text und nennt die genaue Stelle,
wenn etwas nicht stimmt.

Steht gerade keine Meldung an, zeigt die Statuszeile den Zähler, zum Beispiel
`12 Zeilen, 340 Zeichen`.

### Die Zahlen rechts und links

Links im Textfeld stehen Zeilennummern. Sie zählen die **Zeilen, die Sie gerade
sehen**, nicht die Absätze im Text: Das Feld ist so breit wie das Fenster, und
eine längere Zeile bricht weich um. Bekommt eine Zeile davon zwei
Bildschirmzeilen, stehen dort auch zwei Zahlen. Beim Blättern und Tippen bleibt
die Nummer bei der Zeile, zu der sie gehört.

Der Zähler in der Statuszeile nennt dieselbe Zahl. Ist das Feld leer, steht dort
`0 Zeilen, 0 Zeichen`; die Nummerierung beginnt trotzdem bei 1. Ein
Zeilenumbruch zählt als ein Zeichen. Zeichen werden als Unicode-Zeichen gezählt,
nicht als Bytes.

### Hell oder dunkel

Der Knopf oben rechts zeigt an, **wohin** Sie wechseln, nicht wo Sie sind: Im
hellen Design steht dort ein Mond mit **Dunkel**, im dunklen eine Sonne mit
**Hell**. Ein Klick schaltet um. Die Wahl gilt für diese Sitzung; beim nächsten
Start ist es wieder hell. Das Programm merkt sich die Einstellung absichtlich
nicht.

![Oberfläche im dunklen Design](bild-dunkel.png)

### Bearbeiten im Textfeld

Das Textfeld ist ein eigener Editor, kein schreibgeschütztes Vorschaufeld. Die
üblichen Kürzel funktionieren:

| Kürzel                                  | Wirkung                                                                         |
| --------------------------------------- | ------------------------------------------------------------------------------- |
| **Strg+C / Strg+X / Strg+V**            | kopieren, ausschneiden, einfügen                                                |
| **Strg+A**                              | alles auswählen                                                                 |
| **Strg+Z**                              | rückgängig                                                                      |
| **Strg+Shift+Z oder Strg+Y**            | wiederholen                                                                     |
| **Tab**                                 | Auswahl einrücken; ohne Auswahl vier Leerzeichen an der Cursorposition einfügen |
| **Shift+Tab**                           | bis zu vier führende Leerzeichen am Zeilenanfang zurücknehmen                   |
| **Enter**                               | neue Zeile mit der Einrückung der aktuellen Zeile                               |
| **Pfeiltasten, Bild auf/ab, Pos1/Ende** | Cursor bewegen; mit Shift wird ausgewählt                                       |
| **Strg+← / Strg+→**                     | wortweise nach links beziehungsweise rechts                                     |
| **Strg+↑ / Strg+↓**                     | an den Anfang beziehungsweise das Ende des Textes                               |

Ein Rechtsklick öffnet ein Menü mit Ausschneiden, Kopieren, Einfügen, Alles
auswählen, Rückgängig und Wiederholen. Gemerkt werden bis zu 500
Undo-Schritte; schnelles Tippen wird zu einem Schritt zusammengefasst. Ein neu
eingelesener Text sowie jede Ausgabe von Formatieren oder Packen beginnen einen
neuen Verlauf.

Mehr als 100 000 Zeichen werden nicht mehr eingefärbt. Der Text bleibt auch
darüber hinaus vollständig bearbeitbar, und Formatieren und Packen lesen
weiterhin alles.

### Was beim Formatieren aus Ihrem Text wird

| Regel                                                |                                                                                                           |
| ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Einrückung                                           | vier Leerzeichen je Ebene                                                                                 |
| Elemente                                             | ein Element je Zeile, wenn das Elternelement nur Markup enthält                                           |
| Text                                                 | bleibt unverändert in seinem Element; zusätzliche Umbrüche werden nicht eingefügt                         |
| Gemischter Inhalt                                    | Text und Elemente nebeneinander bleiben unverändert stehen                                                |
| Attribute                                            | Reihenfolge, Schreibweise, Anführungszeichen und Inhalt bleiben                                           |
| Kommentare, CDATA, DOCTYPE, Verarbeitungsanweisungen | bleiben erhalten                                                                                          |
| Leere Elemente                                       | im umbrochenen Baum als `<leer />`; `<leer></leer>` bleibt ausdrücklich; in gemischtem Inhalt unverändert |
| BOM und Zeilenenden                                  | BOM fällt weg, `CRLF` wird zu `LF`                                                                        |
| Schluss                                              | genau ein Zeilenumbruch am Ende                                                                           |

Die vollständige Beschreibung, einschliesslich der Fehlermeldungen, steht im
[README](../README.md#was-es-einliest-und-was-zurückgibt).

### Wenn im XML etwas nicht stimmt

Statt zu raten, lässt das Programm die Eingabe stehen und sagt Ihnen, wo es
aufgehört hat:

```
❌ Zeile 3, Spalte 5: Erwartet </a>, gefunden </r>. Der Text ist unverändert geblieben.
```

Der Grund steht wörtlich so im Programm. Es wird nichts korrigiert und nichts
geraten. Prüfen Sie die genannte Zeile und Spalte, ändern Sie den Text, und
formatieren Sie erneut.

## 8. Wenn etwas nicht klappt

### „Der Knopf XML Formatieren ist grau“

Die Statuszeile erklärt es: Im Textfeld steht nichts, womit XML anfangen kann.
Drücken Sie **Clipboard einlesen** und schauen Sie, ob die Statuszeile
umspringt. Oder tippen Sie ein `<` in das Feld — damit ist der Text erkannt und
die Knöpfe werden frei, sofern ein Namensanfang folgt.

### „Fehler: Go wurde nicht gefunden“

Go fehlt oder ist nicht im `PATH`. Installieren Sie es, öffnen Sie ein neues
Terminal und prüfen Sie mit `go version`. Alternativ:

```bat
set "XMLFORMATTER_GO=C:\Program Files\Go\bin\go.exe"
start.bat
```

### „Fehler: Go … ist zu alt“

Installieren Sie Go 1.24 oder neuer. Die Version steht in der Meldung.

### „Fehler: Kein C-Compiler gefunden“

mingw-w64 fehlt, oder `C:\msys64\mingw64\bin` steht nicht im `PATH`. Prüfen Sie
mit `gcc --version` in einem **neuen** Terminal. Die Installation steht in
Schritt 2.

### `gcc: error: unrecognized command-line option` oder ähnliche Compilerfehler

Es wird ein falscher `gcc` gefunden, zum Beispiel von einem anderen
Entwicklungspaket. Prüfen Sie:

```bat
where gcc
```

Der erste Eintrag muss auf `C:\msys64\mingw64\bin\gcc.exe` zeigen. Verschieben
Sie den Eintrag in den Umgebungsvariablen nach oben und öffnen Sie ein neues
Terminal.

### „Fehler: der Build ist fehlgeschlagen“

Die Zeilen davor nennen den Grund. Häufig:

| Meldung                                              | Ursache                                                      |
| ---------------------------------------------------- | ------------------------------------------------------------ |
| `go: downloading ...` und dann ein Netzwerkfehler    | Der erste Bau braucht eine Verbindung für die Fyne-Module    |
| `cgo: C compiler "gcc" not found`                    | mingw-w64 fehlt oder ist nicht im `PATH`                     |
| `collect2.exe: error: ld returned 1 exit status`     | Der Linker ist unvollständig; mingw-w64 neu installieren     |
| `go.mod requires go >= 1.24.0`                       | Go ist zu alt                                                |
| `Access is denied` beim Schreiben von `build\...exe` | Das Programm läuft noch. Fenster schliessen und erneut bauen |

### „Windows hat den Start dieser App verhindert“ oder SmartScreen

Die `.exe` ist selbst gebaut und nicht signiert. Windows zeigt für unbekannte
Programme diesen Hinweis. Wenn Sie die Datei selbst aus diesem Quelltext gebaut
haben, können Sie über **Weitere Informationen → Trotzdem ausführen** starten.
Laden Sie keine fremde `.exe` herunter, nur weil sie so heisst.

### Das Fenster ist nicht zu sehen

Klicken Sie auf den Eintrag in der Taskleiste. Wenn dort nichts ist, starten
Sie `start.bat` in der Eingabeaufforderung und lesen Sie die Ausgabe. Ein
zweites Fenster verhindert das Programm nicht; es liest und schreibt dieselbe
Zwischenablage.

### Das Programm reagiert nicht auf die Zwischenablage

Prüfen Sie mit **Strg+C** in einem anderen Programm, ob dort Text kopiert wurde.
Bilder und Dateien werden nicht übernommen. Die Statuszeile sagt, wenn die
Zwischenablage leer ist.

### `-Neu` hilft nicht

Löschen Sie den Build-Ordner und starten Sie erneut:

```bat
rmdir /s /q build
start.bat
```

Damit wird alles neu übersetzt. Der Go-Modulcache bleibt erhalten.

## 9. Aktualisieren und entfernen

Aktualisieren:

```bat
git pull
start.bat -Neu
```

Ohne Git: Neues ZIP herunterladen und den Ordner ersetzen; `build` darf
gelöscht werden.

Entfernen: Programm beenden und den Projektordner löschen. Das Programm
speichert Theme und Fensterzustand nicht. Der Go-Modulcache unter
`%USERPROFILE%\go\pkg\mod` bleibt davon unberührt und kann bei Bedarf mit
`go clean -modcache` geleert werden. MSYS2 und Go können über die
Windows-Einstellungen deinstalliert werden, wenn Sie sie nicht mehr brauchen.

## 10. Hinweise für alle, die den Quelltext bearbeiten

`start.bat` hat Windows-Zeilenenden und enthält nur ASCII. Beides ist Absicht:
Eine Batch-Datei mit Unix-Zeilenenden läuft nicht zuverlässig, und Sonderzeichen
in Kommentaren können in einer anderen Codepage als Müll ankommen. Die Datei
`.gitattributes` hält `*.bat` beim Auschecken auf CRLF und die Shell-Skripte
auf LF.

`start.bat` ruft PowerShell für die Go-Version und den Zeitstempelvergleich auf.
Das ist Absicht: `%~t` liefert ein Datum in der Sprache des Rechners, und ein
Textvergleich würde je nach System falsch entscheiden. PowerShell gehört ab
Windows 10 zum System und wird nicht zusätzlich installiert.

Wer den Build von Hand machen will:

```bat
go build -trimpath -ldflags "-H=windowsgui" -o build\XMLClipboardFormatter.exe .
```

Ohne `-ldflags` zeigt die Fusszeile die eingebaute Version und das Datum des
Starts. `start.bat` setzt das Build-Datum selbst; die Version steht in
`main.go`.

## 11. Stand

Version 1.0.0, erste Fassung. Was sich geändert hat, steht in
[aenderungen.md](aenderungen.md).
