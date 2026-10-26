# Anleitung für macOS

Diese Anleitung richtet sich an alle, die den XML Clipboard Formatter auf einem
Mac benutzen wollen. Sie setzt keine Programmierkenntnisse voraus — nur ein
Terminal, in das Sie einen Befehl eintippen können.

> **Sie suchen die Anleitung für Windows?** Dann nehmen Sie
> [anleitung-windows.md](anleitung-windows.md).

---

## 1. Wofür ist das Programm?

Sie markieren XML in einem beliebigen Programm, kopieren es mit **Cmd+C** in
die Zwischenablage und formatieren es mit einem Klick:

```xml
<bestellung><kunde><name>Müller &amp; Söhne</name></kunde></bestellung>
```

wird zu:

```xml
<bestellung>
    <kunde>
        <name>Müller &amp; Söhne</name>
    </kunde>
</bestellung>
```

Das Ergebnis liegt danach wieder in der Zwischenablage, sodass Sie es mit
**Cmd+V** an der gewünschten Stelle einsetzen können. Sie brauchen dafür kein
Programm, das XML kennt, und keine Internetverbindung.

Genauso geht es umgekehrt: **XML Packen** entfernt den Leerraum zwischen den
Elementen. Das ist die Form, die viele Werkzeuge erwarten, wenn XML in eine
Zeile oder in ein Feld ohne Umbrüche soll.

## 2. Was Sie brauchen

| | |
| --- | --- |
| **Go 1.24 oder neuer** | zwingend, das Programm ist in Go geschrieben |
| **Xcode Command Line Tools** | zwingend, Fyne braucht einen C-Compiler |
| **Git** | nur zum Holen des Quelltexts; ein ZIP funktioniert auch |

Die Command Line Tools bringen `clang` und die macOS-Systemheader mit. Ohne sie
bricht der erste Bau mit einer Meldung über einen fehlenden C-Compiler ab.
Go selbst bringt keinen C-Compiler mit.

## 3. Schritt 1: Go prüfen

Öffnen Sie das Terminal über **Finder → Programme → Dienstprogramme →
Terminal** (oder Cmd+Leertaste, „Terminal“ eintippen, Enter) und geben Sie ein:

```sh
go version
```

Sieht die Antwort so oder ähnlich aus, ist alles in Ordnung:

```
go version go1.26.1 darwin/arm64
```

Steht dort `zsh: command not found: go` oder eine Versionszahl unter 1.24,
fehlt Go oder ist zu alt. Installieren Sie es über [go.dev/dl](https://go.dev/dl/)
(„Apple macOS“, `.pkg`) und öffnen Sie danach ein **neues** Terminalfenster.
Das ist wichtig: Ein bereits offenes Terminal kennt den neuen `PATH` nicht.

## 4. Schritt 2: Compiler prüfen

```sh
xcode-select -p
```

Kommt ein Pfad wie `/Library/Developer/CommandLineTools`, ist alles bereit.
Kommt eine Fehlermeldung, installieren Sie die Tools einmalig:

```sh
xcode-select --install
```

Es öffnet sich ein Dialog. Nehmen Sie ihn an, warten Sie, bis die Installation
fertig ist, und prüfen Sie danach noch einmal mit `xcode-select -p`.

## 5. Schritt 3: Das Programm holen

Wechseln Sie in einen Ordner, in dem das Programm liegen darf, und holen Sie es
mit Git:

```sh
git clone https://github.com/BENUTZERNAME/xmlClipboardFormatter.git
cd xmlClipboardFormatter
```

`BENUTZERNAME` steht für den GitHub-Namen des Repository-Inhabers.

Wer kein Git hat, lädt auf der Projektseite **Code → Download ZIP**, entpackt
die Datei und wechselt im Terminal in den entpackten Ordner.

## 6. Schritt 4: Einmal bauen und starten

```sh
./start.sh
```

Beim ersten Mal schreibt das Skript `Bauen ...` und lädt die Fyne-Module nach.
Das dauert je nach Verbindung ein bis zwei Minuten. Danach öffnet sich das
Fenster **XML Clipboard Formatter**. Beim nächsten Start geht es schnell, weil
nur noch geprüft wird, ob eine Quelldatei neuer ist als die fertige Datei.

Die ausführbare Datei liegt danach unter `build/XMLClipboardFormatter`. Sie
können sie auch direkt starten:

```sh
./build/XMLClipboardFormatter
```

Für den normalen Gebrauch ist `./start.sh` aber der bequemere Weg: Es baut
automatisch nach, wenn Sie oder jemand anderes am Quelltext etwas geändert hat.

### Ohne Terminal: Start per Doppelklick

Im Ordner liegt `start.command`. Ein **Doppelklick im Finder** öffnet ein
Terminalfenster und ruft `start.sh` auf. Das Terminalfenster bleibt offen,
solange das Programm läuft; das ist bei einem Skriptstarter so.

Beim ersten Doppelklick kann macOS fragen, ob die Datei geöffnet werden darf.
Bestätigen Sie mit **Öffnen**. Falls der Finder die Datei trotzdem nicht
startet, prüfen Sie einmalig die Ausführungsrechte:

```sh
chmod +x start.sh start.command
```

Danach funktioniert auch der Doppelklick. Wenn nicht, bleibt der Weg über das
Terminal: Dort sehen Sie die Meldung, an der es liegt.

### Was die Schalter können

| Befehl | Wirkung |
| --- | --- |
| `./start.sh` | starten, vorher bei Bedarf bauen |
| `./start.sh -Neu` | vorher neu bauen, auch wenn alles aktuell ist, dann starten |
| `./start.sh -Pruefen` | nur nachsehen, ob die Umgebung passt — baut und startet nichts |

`-Pruefen` gibt ungefähr das aus:

```
Go      : /usr/local/go/bin/go  (PATH)
Version : go1.26.1 (mindestens Go 1.24)
Compiler: /usr/bin/clang
App     : build/XMLClipboardFormatter - vorhanden
Bauen   : nein, Quellen sind nicht neuer
Argumente: 0 an das Programm
```

Steht bei `App` „fehlt“ oder bei `Bauen` „ja“, ist das kein Fehler: Der nächste
Start ohne `-Pruefen` baut dann. Steht bei `Compiler` „nicht gefunden“, fehlen
die Command Line Tools.

### Wenn Go an einer ungewöhnlichen Stelle liegt

`start.sh` sucht Go in dieser Reihenfolge: die Umgebungsvariable
`XMLFORMATTER_GO`, dann `$GOROOT/bin/go`, dann den `PATH`. Für einen einzelnen
Start können Sie den Pfad vorgeben:

```sh
XMLFORMATTER_GO=/opt/homebrew/bin/go ./start.sh
```

Für dauerhafte Nutzung legen Sie die Variable in Ihrer Shell-Konfiguration ab
(`~/.zshrc`):

```sh
export XMLFORMATTER_GO=/opt/homebrew/bin/go
```

Danach ein neues Terminal öffnen oder `source ~/.zshrc` ausführen.

## 7. Bedienung

![Oberfläche im hellen Design](bild-hell.png)

Das Fenster trägt den Titel **XML Clipboard Formatter**. Rechts oben steht der
Knopf zum Umschalten zwischen dunkel und hell. Darunter liegt das grosse
Textfeld mit den Zeilennummern links. Ganz unten stehen die Aktionsknöpfe,
darunter die Statuszeile und rechts die Versionszeile, zum Beispiel
`Version 1.0.0 vom 26.10.2026`.

### Die fünf Knöpfe

| Knopf | Wirkung |
| --- | --- |
| **Clipboard einlesen** | holt den Text aus der Zwischenablage in das Feld |
| **XML Formatieren** | macht den Text lesbar, ein Element je Zeile |
| **XML Packen** | entfernt den Leerraum zwischen den Elementen |
| **Ins Clipboard schreiben** | legt den Text aus dem Feld in die Zwischenablage |
| **Beenden** | beendet das Programm; steht rechts als Textlink |

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

1. XML in Ihrem Programm markieren und mit **Cmd+C** kopieren
2. **Clipboard einlesen** → `✓ XML geladen. Formatieren, Packen und Schreiben sind freigeschaltet.`
3. **XML Formatieren** → `✓ XML lesbar formatiert.`
4. **Ins Clipboard schreiben** → `✓ XML erfolgreich in die Zwischenablage kopiert.`
5. In Ihr Programm zurück, **Cmd+V** einfügen

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

| Kürzel | Wirkung |
| --- | --- |
| **Cmd+C / Cmd+X / Cmd+V** | kopieren, ausschneiden, einfügen |
| **Cmd+A** | alles auswählen |
| **Cmd+Z** | rückgängig |
| **Shift+Cmd+Z oder Cmd+Y** | wiederholen |
| **Tab** | Auswahl einrücken; ohne Auswahl vier Leerzeichen an der Cursorposition einfügen |
| **Shift+Tab** | bis zu vier führende Leerzeichen am Zeilenanfang zurücknehmen |
| **Enter** | neue Zeile mit der Einrückung der aktuellen Zeile |
| **Pfeiltasten, Bild auf/ab, Home/End** | Cursor bewegen; mit Shift wird ausgewählt |
| **Cmd+← / Cmd+→** | an den Zeilenanfang beziehungsweise das Zeilenende |
| **Alt+← / Alt+→** | wortweise nach links beziehungsweise rechts |
| **Cmd+↑ / Cmd+↓** | an den Anfang beziehungsweise das Ende des Textes |

Ein Rechtsklick öffnet ein Menü mit Ausschneiden, Kopieren, Einfügen, Alles
auswählen, Rückgängig und Wiederholen. Gemerkt werden bis zu 500
Undo-Schritte; schnelles Tippen wird zu einem Schritt zusammengefasst. Ein neu
eingelesener Text sowie jede Ausgabe von Formatieren oder Packen beginnen einen
neuen Verlauf.

Mehr als 100 000 Zeichen werden nicht mehr eingefärbt. Der Text bleibt auch
darüber hinaus vollständig bearbeitbar, und Formatieren und Packen lesen
weiterhin alles.

### Was beim Formatieren aus Ihrem Text wird

| Regel | |
| --- | --- |
| Einrückung | vier Leerzeichen je Ebene |
| Elemente | ein Element je Zeile, wenn das Elternelement nur Markup enthält |
| Text | bleibt unverändert in seinem Element; zusätzliche Umbrüche werden nicht eingefügt |
| Gemischter Inhalt | Text und Elemente nebeneinander bleiben unverändert stehen |
| Attribute | Reihenfolge, Schreibweise, Anführungszeichen und Inhalt bleiben |
| Kommentare, CDATA, DOCTYPE, Verarbeitungsanweisungen | bleiben erhalten |
| Leere Elemente | im umbrochenen Baum als `<leer />`; `<leer></leer>` bleibt ausdrücklich; in gemischtem Inhalt unverändert |
| BOM und Zeilenenden | BOM fällt weg, `CRLF` wird zu `LF` |
| Schluss | genau ein Zeilenumbruch am Ende |

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

```sh
XMLFORMATTER_GO=/vollständiger/pfad/zu/go ./start.sh
```

### „Fehler: Go … ist zu alt“

Installieren Sie Go 1.24 oder neuer. Die Version steht in der Meldung.

### „Fehler: Kein C-Compiler gefunden“

Die Xcode Command Line Tools fehlen:

```sh
xcode-select --install
```

Danach `./start.sh -Pruefen` erneut ausführen; bei `Compiler` muss ein Pfad
stehen.

### „xcrun: error: invalid active developer path“

Die Command Line Tools sind nach einem macOS-Update nicht mehr richtig
verknüpft. Einmalig:

```sh
xcode-select --reset
xcode-select --install
```

### „./start.sh: Permission denied“

```sh
chmod +x start.sh start.command
```

### „go: downloading …“ bleibt stehen

Der erste Bau lädt die Fyne-Module. Ohne Netzwerk geht das nicht. Prüfen Sie
die Verbindung oder einen Proxy. Nach dem ersten erfolgreichen Bau startet das
Programm auch ohne Netzwerk, solange der Go-Modulcache vorhanden ist.

### „ld: warning: ignoring duplicate libraries: '-lobjc'“

Diese Warnung des macOS-Linkers ist harmlos. Der Bau läuft weiter, und das
Programm startet. Sie können sie ignorieren.

### Das Fenster ist nicht zu sehen

Klicken Sie auf das Programm im Dock. Wenn dort nichts ist, starten Sie
`./start.sh` im Terminal und lesen Sie die Ausgabe. Ein zweites Fenster
verhindert das Programm nicht; es liest und schreibt dieselbe Zwischenablage.

### Das Programm reagiert nicht auf die Zwischenablage

Prüfen Sie mit **Cmd+C** in einem anderen Programm, ob dort Text kopiert wurde.
Bilder und Dateien werden nicht übernommen. Die Statuszeile sagt, wenn die
Zwischenablage leer ist.

### `-Neu` hilft nicht

Löschen Sie den Build-Ordner und starten Sie erneut:

```sh
rm -rf build
./start.sh
```

Damit wird alles neu übersetzt. Der Go-Modulcache bleibt erhalten.

## 9. Aktualisieren und entfernen

Aktualisieren:

```sh
git pull
./start.sh -Neu
```

Ohne Git: Neues ZIP herunterladen und den Ordner ersetzen; `build/` darf
gelöscht werden.

Entfernen: Programm beenden und den Projektordner löschen. Das Programm
speichert Theme und Fensterzustand nicht. Der Go-Modulcache unter
`~/go/pkg/mod` bleibt davon unberührt und kann bei Bedarf mit
`go clean -modcache` geleert werden.

## 10. Stand

Version 1.0.0, erste Fassung. Was sich geändert hat, steht in
[aenderungen.md](aenderungen.md).
