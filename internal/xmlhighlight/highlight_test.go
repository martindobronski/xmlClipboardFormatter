package xmlhighlight

import (
	"strings"
	"testing"
)

// pruefeDeckung stellt sicher, dass jeder Rune genau in einem Bereich liegt.
func pruefeDeckung(t *testing.T, text string) []Span {
	t.Helper()
	spans := Scan(text)
	l := len([]rune(text))
	if l == 0 {
		if len(spans) != 0 {
			t.Fatalf("leerer Text liefert %d Bereiche", len(spans))
		}
		return spans
	}
	if len(spans) == 0 {
		t.Fatal("keine Bereiche für nichtleeren Text")
	}
	if spans[0].Start != 0 {
		t.Fatalf("erster Bereich beginnt bei %d, nicht bei 0", spans[0].Start)
	}
	pos := 0
	for i, s := range spans {
		if s.Start != pos {
			t.Fatalf("Bereich %d beginnt bei %d, erwartet %d", i, s.Start, pos)
		}
		if s.End <= s.Start {
			t.Fatalf("Bereich %d ist leer: %+v", i, s)
		}
		pos = s.End
	}
	if pos != l {
		t.Fatalf("Bereiche enden bei %d, Text hat %d Runen", pos, l)
	}
	return spans
}

func artAn(spans []Span, pos int) Kind {
	for _, s := range spans {
		if pos >= s.Start && pos < s.End {
			return s.Kind
		}
	}
	return Plain
}

func artVon(t *testing.T, spans []Span, text, teil string) Kind {
	t.Helper()
	pos := strings.Index(text, teil)
	if pos < 0 {
		t.Fatalf("%q kommt im Text nicht vor", teil)
	}
	return artAn(spans, len([]rune(text[:pos])))
}

func TestScanDecktDenGanzenTextAb(t *testing.T) {
	texte := []string{
		"",
		"   ",
		"<a/>",
		"<a  x = \"1\" >text</a>",
		"  <root>\n  <kind>Müller &amp; Söhne</kind>\n</root>\n",
		"<![CDATA[roh < & >]]>",
		"<?xml version=\"1.0\"?><root/>",
	}
	for _, text := range texte {
		pruefeDeckung(t, text)
	}
}

func TestScanKenntAlleBereiche(t *testing.T) {
	text := `<?xml version="1.0"?>
<!DOCTYPE root [ <!ELEMENT root (#PCDATA)> ]>
<root a="x" b='y'>
  <!-- Kommentar -->
  <kind>Müller &amp; Söhne</kind>
  <![CDATA[roh & <bleibt>]]>
</root>`
	spans := pruefeDeckung(t, text)

	faelle := []struct {
		teil string
		want Kind
	}{
		{"<?xml", ProcessingInstruction},
		{"<!DOCTYPE", Doctype},
		{"root a=", TagName},
		{"a=", AttrName},
		{`x"`, AttrValue},
		{"<!--", Comment},
		{"kind", TagName},
		{"&amp;", Entity},
		{"<![CDATA[", CDATA},
		{"</root>", Bracket},
	}
	for _, f := range faelle {
		if got := artVon(t, spans, text, f.teil); got != f.want {
			t.Errorf("%q: Art %v, erwartet %v", f.teil, got, f.want)
		}
	}
}

func TestScanAttributeUndEntitaeten(t *testing.T) {
	text := `<a x="1" y='&amp;' z = "ä &lt; b">&#x41;</a>`
	spans := pruefeDeckung(t, text)

	for _, teil := range []string{`x="`, `y='`, `z = "`} {
		if got := artVon(t, spans, text, teil); got != AttrName {
			t.Errorf("Attribut vor %q: Art %v, erwartet %v", teil, got, AttrName)
		}
	}
	if got := artVon(t, spans, text, "&amp;"); got != Entity {
		t.Errorf("Attribut-Entität: Art %v, erwartet %v", got, Entity)
	}
	if got := artVon(t, spans, text, "&#x41;"); got != Entity {
		t.Errorf("Text-Entität: Art %v, erwartet %v", got, Entity)
	}
}

func TestScanUnvollstaendigesXML(t *testing.T) {
	texte := []string{
		"<",
		"</",
		"<a",
		"<a x=",
		"<a x=\"offen",
		"<!-- offen",
		"<![CDATA[ offen",
		"<?xml offen",
		"<!DOCTYPE root [",
		"<a>&amp",
		"text < text",
		"<a b=/>",
	}
	for _, text := range texte {
		pruefeDeckung(t, text)
	}
}

func TestScanDoctypeMitInternerTeilmenge(t *testing.T) {
	text := `<!DOCTYPE root [
  <!ELEMENT root (#PCDATA)>
  <!ENTITY gt ">">
]>
<root/>`
	spans := pruefeDeckung(t, text)
	for i := 0; i < strings.Index(text, "]>\n"); i++ {
		if kind := artAn(spans, i); kind != Doctype {
			t.Fatalf("Zeichen %d im DOCTYPE hat Art %v", i, kind)
		}
	}
	if got := artVon(t, spans, text, "root/"); got != TagName {
		t.Errorf("Wurzel: Art %v, erwartet %v", got, TagName)
	}
}

func TestScanLeererNameBleibtText(t *testing.T) {
	text := "< notatag >"
	spans := pruefeDeckung(t, text)
	if got := artAn(spans, 0); got != Bracket {
		t.Fatalf("erwartet Bracket an 0, bekam %v", got)
	}
	if got := artAn(spans, 1); got != Plain {
		t.Fatalf("erwartet Plain an 1, bekam %v", got)
	}
}
