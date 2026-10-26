# Was sich geändert hat

Kurz gehalten, mit dem Grund dahinter. Wer den Grund nicht braucht, liest die
fett gedruckte Zeile und lässt den Rest stehen.

## 1.0.0 — erste Fassung

**Das Programm liest XML aus der Zwischenablage, macht es lesbar oder packt es
zusammen und schreibt es zurück.**

Es gibt keine Vorgängerfassung, deshalb steht unten nicht „geändert“, sondern
was es gibt und warum es so gebaut ist.

### Es bleibt XML, auch in der Schreibweise

Der naheliegende Weg wäre `encoding/xml` aus der Go-Standardbibliothek gewesen.
Der Leser dort liefert aber keine Schreibweise, sondern Werte: Anführungszeichen
eines Attributs sind weg, CDATA ist nicht mehr von gewöhnlichem Text zu
unterscheiden, und Entity-Referenzen werden aufgelöst. Wer daraus wieder XML
baut, bekommt ein neues Dokument, nicht das alte mit anderem Leerraum.

Deshalb gibt es einen eigenen Lexer und einen eigenen Parser. Sie kennen
Start-Tags, End-Tags, Attribute, Text, Kommentare, CDATA,
Verarbeitungsanweisungen und DOCTYPE als eigene Stücke. Die Ausgabe setzt nur
Leerraum neu. Attribute behalten Reihenfolge, Schreibweise, Anführungszeichen
und Inhalt; `a='x &amp; y'` bleibt genau so stehen. Ein ausdrücklich
geschriebenes `<leer></leer>` wird nicht zu `<leer />`, und ein vorhandenes
`<leer />` bleibt ein leeres Element.

### Zwei Wege zum selben Inhalt

**XML Formatieren** stellt ein Element je Zeile dar, mit vier Leerzeichen
Einrückung. **XML Packen** entfernt den Leerraum zwischen den Elementen, für
Befehlszeilen und Felder ohne Umbrüche. Beide lesen denselben Text und ändern
kein sinntragendes Zeichen.

Drei Fälle brauchen Sonderregeln, weil XML dort nicht beliebig umbrochen werden
darf:

- **Nur Text** in einem Element bleibt einzeilig. `<name>Müller</name>` wird
  nicht zu drei Zeilen.
- **Gemischter Inhalt** aus Text und Elementen bleibt Zeichen für Zeichen
  stehen. `<p>Hallo <b>Welt</b>!</p>` würde durch zusätzliche Umbrüche einen
  anderen Text ergeben.
- **Nur Markup** in einem Element wird blockweise eingerückt. Dort ist
  Leerraum ohne Bedeutung und darf neu gesetzt werden.

Leerraum in einem Element, das nur Text enthält, bleibt auch beim Packen
stehen. Er kann Bedeutung haben, und das Programm rät nicht, ob er gemeint ist.

### Jede Ausgabe wird geprüft

Nach dem Formatieren und Packen wird die eigene Ausgabe erneut gelesen und
Stück für Stück mit der Eingabe verglichen. Stimmt die Zahl der Stücke nicht
oder weicht eines ab, bricht der Vorgang ab, statt ein Ergebnis zu liefern.
Leerraum zwischen Markup zählt beim Vergleich nicht — ihn hinzuzufügen oder
wegzulassen ist der Zweck des Formatierers. Sichtbarer Text, Attribute,
Kommentare, CDATA, Verarbeitungsanweisungen und DOCTYPE müssen gleich bleiben.

Die Prüfung kostet wenig und läuft bei jedem Schreiben mit, nicht nur im Test.
Sie findet nicht nur einen Fehler in der Ausgabe, sondern auch einen im Lexer;
beide wären sonst erst sichtbar, wenn jemand die Datei öffnet, die dieses
Programm geschrieben hat.

### Die Schranke in der Statuszeile

Formatieren, Packen und Schreiben bleiben gesperrt, solange der Text nicht als
XML erkannt wird. Erkannt wird am ersten Zeichen nach BOM und Leerraum: `<`,
danach `!`, `?` oder ein Namensanfang. Das ist eine Vermutung, keine Prüfung,
und zwar mit Absicht: Eine volle Prüfung müsste nach jedem Tastendruck laufen
und würde genau den Text sperren, an dem gerade getippt wird. Der Formatter
bleibt der Richter und nennt die Stelle.

### Ein Fehler ändert nichts

Kommt kein gültiges XML heraus, bleibt der Text stehen, in dem er war, und die
Meldung nennt Zeile und Spalte. Es wird nicht versucht, „irgendwie“ zu
reparieren, und es wird nichts geraten: Ein Text, der als Datei weitergeht,
muss von hier aus nachvollziehbar bleiben. Die Meldung beginnt mit einem Kreuz
und nicht mit einem Haken, damit sie auf den ersten Blick von den
Erfolgsmeldungen zu unterscheiden ist.

### Grenzen, die bewusst so sind

- **10 000 Ebenen.** Der Parser arbeitet mit einem Stapel statt mit Rekursion
  und könnte tiefer lesen. Die Grenze schützt die Ausgabe: Lesbar formatieren
  rückt jede Ebene ein, und der Text wächst quadratisch mit der Tiefe. Bei
  10 000 Ebenen ist die Ausgabe bereits sehr gross. Wer nur weiterverarbeiten
  will, nimmt Packen; das bleibt linear.
- **100 000 Zeichen für die Syntaxfärbung.** Darüber bleibt der Text
  vollständig bearbeitbar, nur die Farbe endet. Ein Editor, der bei jedem
  Tastendruck den ganzen Text neu zerlegt, würde sonst irgendwann zäh.
- **Kein Schema, keine DTD, kein Netzwerk.** Das Programm prüft, ob der Text
  wohlgeformt ist. Ob ein Element an dieser Stelle erlaubt ist, weiss es nicht.
  Eine externe DTD wird nicht geladen, und eigene Entities werden nicht
  aufgelöst.
- **Kein Sitzungszustand.** Theme, Fensterposition und Inhalt werden nicht
  gespeichert. Beim Start ist das Fenster leer und hell. Das Programm schreibt
  keine Einstellungen, die jemand später wiederfinden müsste.
- **Nur Text.** Ein Bild in der Zwischenablage wird nicht in XML verwandelt.
  Das Textfeld bleibt, wie es war, und die Statuszeile sagt es.

### Der Weg vom Quelltext zum Fenster

Die Oberfläche ist Fyne 2.8.2. Fyne zeichnet mit OpenGL und braucht deshalb
cgo und einen C-Compiler: unter macOS die Xcode Command Line Tools, unter
Windows mingw-w64. Das ist der Preis dafür, dass dasselbe Programm auf beiden
Systemen als einzelne ausführbare Datei läuft, ohne Java-Laufzeit und ohne
Browser.

`start.sh` und `start.bat` bauen beim ersten Start und danach nur, wenn eine
Quelldatei neuer ist als die ausführbare Datei. `-Neu` erzwingt den Bau,
`-Pruefen` prüft Go, Compiler und Zielpfad, ohne zu bauen oder zu starten. Die
Skripte setzen das Build-Datum über `-ldflags`; die Version steht in `main.go`.
Ohne sie zeigt die Fusszeile die eingebaute Version und das Datum des Starts.

`start.bat` hat Windows-Zeilenenden und nur ASCII. Eine Batch-Datei mit
Unix-Zeilenenden läuft nicht zuverlässig, und Sonderzeichen können in einer
anderen Codepage als Müll ankommen. `.gitattributes` hält `*.bat` deshalb auf
CRLF und die Shell-Skripte auf LF.

### Zwei Stellen, an denen nicht die kürzeste Lösung gewählt wurde

- **Der Editor ist selbst gebaut.** Fyne bringt ein mehrzeiliges Textfeld mit,
  aber keine Zeilennummern, keine eigene XML-Färbung und kein Scrollen in
  ganzen Bildschirmzeilen. Ein eigenes Widget ist mehr Code, dafür passen
  Cursor, Auswahl, Undo und Darstellung zusammen. Der Textpuffer liegt getrennt
  in `internal/editing` und ist ohne Fenster testbar.
- **Die Statuszeile ist wichtiger als ein Dialog.** Erfolg, Warnung und Fehler
  stehen an derselben Stelle und in derselben Zeile. Erfolgsmeldungen
  verschwinden nach vier Sekunden, Warnungen und Fehler bleiben. Ein Dialog
  würde bei jeder Formatierung einen Klick verlangen, den niemand machen will.

### Was nicht abgesichert ist

- **Die Syntaxfärbung ist eine Anzeige, keine Prüfung.** Bei einer halb
  getippten Datei kann sie anders färben, als der Parser später entscheidet.
  Absicht: Wer tippt, sieht lieber zu viel Farbe als eine tote Fläche.
- **Die XML-Deklaration wird nicht ausgewertet.** Der Text kommt als Unicode
  aus der Zwischenablage. `encoding="..."` bleibt stehen, ändert aber nichts
  daran, wie das Programm die Zeichen liest.
- **Die macOS-Linker-Warnung `ignoring duplicate libraries: '-lobjc'`** ist
  harmlos und kommt von Fyne. Der Bau läuft weiter.

## Stand

Version 1.0.0, erste Fassung. Der Text hier beschreibt den Stand dieses
Commits. Die Bedienung steht im [README](../README.md), die Einrichtung in
[anleitung-macos.md](anleitung-macos.md) und
[anleitung-windows.md](anleitung-windows.md).
