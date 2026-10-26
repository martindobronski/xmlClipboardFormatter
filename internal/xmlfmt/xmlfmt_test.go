package xmlfmt

import (
	"strings"
	"testing"
)

func TestLooksLikeXML(t *testing.T) {
	t.Parallel()
	faelle := []struct {
		text string
		want bool
	}{
		{`<root/>`, true},
		{"\uFEFF  \n<root/>", true},
		{`<?xml version="1.0"?><root/>`, true},
		{`<!-- Kommentar --><root/>`, true},
		{`<`, false},
		{``, false},
		{`   `, false},
		{`Einkaufsliste`, false},
		{`{"a":1}`, false},
		{`<3`, false},
		{`</root>`, false},
	}
	for _, f := range faelle {
		if got := LooksLikeXML(f.text); got != f.want {
			t.Errorf("LooksLikeXML(%q) = %v, want %v", f.text, got, f.want)
		}
	}
}

func TestPrettyEinfach(t *testing.T) {
	t.Parallel()
	eingabe := `<?xml version="1.0" encoding="UTF-8"?><root><a x="1">text</a><leer/><explizit></explizit><!-- hinweis --></root>`
	want := `<?xml version="1.0" encoding="UTF-8"?>
<root>
    <a x="1">text</a>
    <leer />
    <explizit></explizit>
    <!-- hinweis -->
</root>
`
	got, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	if got != want {
		t.Errorf("Pretty:\n%s\nwant:\n%s", got, want)
	}
}

func TestPrettyAttributeBleibenWortgetreu(t *testing.T) {
	t.Parallel()
	eingabe := `<r a="1.50" b='x &amp; y' c="&lt;tag&gt;" d="doppelt"><e/></r>`
	want := "<r a=\"1.50\" b='x &amp; y' c=\"&lt;tag&gt;\" d=\"doppelt\">\n    <e />\n</r>\n"
	got, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	if got != want {
		t.Errorf("Pretty:\n%s\nwant:\n%s", got, want)
	}
}

func TestPrettyGemischtBleibtStehen(t *testing.T) {
	t.Parallel()
	eingabe := "<p>Hallo <b>Welt</b>, wie geht es dir?</p>"
	got, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	if got != eingabe+"\n" {
		t.Errorf("gemischter Inhalt wurde verändert:\n%s", got)
	}
}

func TestPrettyTextBleibtStehen(t *testing.T) {
	t.Parallel()
	eingabe := "<a>\n  zeile 1\n  zeile 2\n</a>"
	got, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	if got != eingabe+"\n" {
		t.Errorf("Textinhalt wurde verändert:\n%s", got)
	}
}

func TestPrettyGeschuetztesLeerzeichenBleibtStehen(t *testing.T) {
	t.Parallel()
	// U+00A0 ist XML-Leerraum? Nein. Es ist sichtbarer Text und darf auch
	// zwischen Elementen nicht als Einrückung verschwinden.
	eingabe := "<r><a />\u00A0<b /></r>"
	got, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	if got != eingabe+"\n" {
		t.Errorf("geschütztes Leerzeichen wurde verändert:\n%q", got)
	}
}

func TestCompactGeschuetztesLeerzeichenBleibtStehen(t *testing.T) {
	t.Parallel()
	eingabe := "<r><a />\u00A0<b /></r>"
	got, err := Compact(eingabe)
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}
	if got != eingabe+"\n" {
		t.Errorf("geschütztes Leerzeichen wurde verändert:\n%q", got)
	}
}

func TestTrimSpace(t *testing.T) {
	t.Parallel()
	faelle := []struct {
		text string
		want string
	}{
		{"<r/>", "<r/>"},
		{"  \t<r/>\r\n", "<r/>"},
		{"\u00A0<r/>\u00A0", "\u00A0<r/>\u00A0"},
		{" \u00A0<r/>\u00A0 ", "\u00A0<r/>\u00A0"},
		{"\u2003<r/>\u2003", "\u2003<r/>\u2003"},
		{"   ", ""},
	}
	for _, f := range faelle {
		if got := TrimSpace(f.text); got != f.want {
			t.Errorf("TrimSpace(%q) = %q, want %q", f.text, got, f.want)
		}
	}
}

func TestPrettyCDATA(t *testing.T) {
	t.Parallel()
	eingabe := `<r><script><![CDATA[if (a < b) { x(); }]]></script></r>`
	want := "<r>\n    <script><![CDATA[if (a < b) { x(); }]]></script>\n</r>\n"
	got, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	if got != want {
		t.Errorf("Pretty:\n%s\nwant:\n%s", got, want)
	}
}

func TestPrettyDoctype(t *testing.T) {
	t.Parallel()
	eingabe := `<!DOCTYPE r SYSTEM "r.dtd"><r><a>1</a></r>`
	want := "<!DOCTYPE r SYSTEM \"r.dtd\">\n<r>\n    <a>1</a>\n</r>\n"
	got, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	if got != want {
		t.Errorf("Pretty:\n%s\nwant:\n%s", got, want)
	}
}

func TestCompact(t *testing.T) {
	t.Parallel()
	eingabe := `<?xml version="1.0"?>
<root>
    <a x="1">text</a>
    <leer />
    <!-- hinweis -->
    <b><c>  </c></b>
</root>`
	want := `<?xml version="1.0"?><root><a x="1">text</a><leer /><!-- hinweis --><b><c>  </c></b></root>` + "\n"
	got, err := Compact(eingabe)
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}
	if got != want {
		t.Errorf("Compact:\n%s\nwant:\n%s", got, want)
	}
}

func TestBOMUndCRLF(t *testing.T) {
	t.Parallel()
	eingabe := "\uFEFF<r>\r\n  <a>1</a>\r\n</r>\r\n"
	want := "<r>\n    <a>1</a>\n</r>\n"
	got, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	if got != want {
		t.Errorf("Pretty:\n%q\nwant:\n%q", got, want)
	}
}

func TestEmployment(t *testing.T) {
	t.Parallel()
	// Ein etwas längeres Beispiel mit Namensraum, PI und Kommentar.
	eingabe := `<?xml version="1.0"?>
<!-- Wurzel -->
<katalog xmlns="urn:beispiel" xmlns:x="urn:x">
  <?leer?>
  <artikel nummer="1">
    <name>Kaffee</name>
    <preis währung="EUR">4,50</preis>
    <x:meta x:key="a"> &amp; </x:meta>
  </artikel>
</katalog>`
	schön, err := Pretty(eingabe)
	if err != nil {
		t.Fatalf("Pretty: %v", err)
	}
	dicht, err := Compact(schön)
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}
	nochmal, err := Pretty(dicht)
	if err != nil {
		t.Fatalf("Pretty nach Compact: %v", err)
	}
	if !strings.Contains(nochmal, `<artikel nummer="1">`) {
		t.Fatalf("Attribut fehlt: %s", nochmal)
	}
	if !strings.Contains(nochmal, `<x:meta x:key="a"> &amp; </x:meta>`) {
		t.Fatalf("Text oder Attribut verändert: %s", nochmal)
	}
}

func TestFehler(t *testing.T) {
	t.Parallel()
	faelle := []struct {
		name string
		xml  string
		want string
	}{
		{"unclosed", `<r><a></r>`, "Erwartet </a>, gefunden </r>."},
		{"fehlendes Ende", `<r><a></a>`, "Tag <r> wird nicht geschlossen."},
		{"zwei Wurzeln", `<a/><b/>`, "Es darf nur ein Wurzelelement geben."},
		{"Text draußen", `Hallo <r/>`, "Text außerhalb des Wurzelelements"},
		{"geschütztes Leerzeichen draußen", "\u00A0<r/>", "Text außerhalb des Wurzelelements"},
		{"leer", ``, "Es fehlt ein Wurzelelement."},
		{"Kommentar offen", `<r><!-- x</r>`, "Kommentar wird nicht geschlossen."},
		{"CDATA offen", `<r><![CDATA[x</r>`, "CDATA-Abschnitt wird nicht geschlossen."},
		{"PI offen", `<r><?x</r>`, "Verarbeitungsanweisung wird nicht geschlossen."},
		{"Attribut ohne Wert", `<r a></r>`, "wird '=' erwartet"},
		{"Attribut ohne Anführungszeichen", `<r a=1></r>`, "muss in Anführungszeichen stehen"},
		{"doppeltes Attribut", `<r a="1" a="2"/>`, "steht mehrfach"},
		{"Entity offen", `<r>a & b</r>`, "'&' muss als Entity"},
		{"Entity ohne Semikolon", `<r>a &amp b</r>`, "nicht mit ';' geschlossen"},
		{"kleiner als im Attribut", `<r a="<"/>`, "darf kein '<'"},
		{"unbekannte Deklaration", `<!FOO><r/>`, "Unbekannte Deklaration"},
		{"DOCTYPE nach Wurzel", `<r/><!DOCTYPE r>`, "muss vor dem Wurzelelement"},
		{"falscher DOCTYPE", `<!DOCTYPE a><b/>`, "nennt 'a'"},
		{"CDATA draußen", `<![CDATA[x]]><r/>`, "CDATA darf nur innerhalb"},
		{"XML-Deklaration nicht am Anfang", `<r/><?xml version="1.0"?>`, "nur am Anfang"},
		{"Kommentar mit --", `<r><!-- a -- b --></r>`, "darf '--' nicht"},
		{"Ende ohne Anfang", `</r>`, "ohne öffnendes Tag"},
	}
	for _, f := range faelle {
		f := f
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			_, err := Pretty(f.xml)
			if err == nil {
				t.Fatalf("kein Fehler für %q", f.xml)
			}
			if !strings.Contains(err.Error(), f.want) {
				t.Fatalf("Fehler %q enthält nicht %q", err.Error(), f.want)
			}
		})
	}
}

func TestFehlerNenntZeileUndSpalte(t *testing.T) {
	t.Parallel()
	_, err := Pretty("<r>\n  <a>\n</r>")
	if err == nil {
		t.Fatal("kein Fehler")
	}
	if !strings.Contains(err.Error(), "Zeile 3, Spalte 1") {
		t.Fatalf("Stelle fehlt in %q", err.Error())
	}
}
