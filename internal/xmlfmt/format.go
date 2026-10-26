package xmlfmt

import (
	"fmt"
	"strings"
)

const eingerueckt = 4

// Pretty formatiert XML lesbar: ein Element je Zeile, vier Leerzeichen je
// Ebene. Textinhalt bleibt unverändert; gemischter Inhalt (Text und Elemente
// nebeneinander) wird nicht umbrochen, weil zusätzlicher Leerraum dort den
// Inhalt verändern würde.
func Pretty(input string) (string, error) {
	tokens, src, err := lex(input)
	if err != nil {
		return "", err
	}
	doc, err := parse(tokens, src)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, n := range doc.nodes {
		switch n.kind {
		case nodeElement:
			schreibeElement(&b, src, n, 0)
		case nodeComment, nodePI, nodeDoctype:
			b.WriteString(textOf(src, n.tok))
			b.WriteByte('\n')
		}
	}
	ausgabe := b.String()
	if err := pruefeAusgabe(src, tokens, ausgabe); err != nil {
		return "", err
	}
	return ausgabe, nil
}

// Compact packt XML auf möglichst wenig Zeilen. Leerraum zwischen Elementen
// fällt weg; Text, Kommentare, CDATA und Verarbeitungsanweisungen werden
// unverändert übernommen. Leerraum in einem Element, das nur Text enthält,
// bleibt stehen: er kann Bedeutung haben.
func Compact(input string) (string, error) {
	tokens, src, err := lex(input)
	if err != nil {
		return "", err
	}
	doc, err := parse(tokens, src)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, n := range doc.nodes {
		switch n.kind {
		case nodeElement:
			schreibeElementDicht(&b, src, n)
		case nodeComment, nodePI, nodeDoctype:
			b.WriteString(textOf(src, n.tok))
		}
	}
	b.WriteByte('\n')
	ausgabe := b.String()
	if err := pruefeAusgabe(src, tokens, ausgabe); err != nil {
		return "", err
	}
	return ausgabe, nil
}

type inhaltsArt int

const (
	inhaltLeer inhaltsArt = iota
	inhaltText
	inhaltBlock
	inhaltGemischt
)

func klassifiziere(n *node, src []rune) inhaltsArt {
	hatText, hatSichtbarenText, hatMarkup := false, false, false
	for _, kind := range n.children {
		switch kind.kind {
		case nodeText:
			hatText = true
			if hasContent(textOf(src, kind.tok)) {
				hatSichtbarenText = true
			}
		case nodeCDATA:
			hatText = true
			hatSichtbarenText = true
		case nodeElement, nodeComment, nodePI:
			hatMarkup = true
		}
	}
	switch {
	case hatMarkup && hatSichtbarenText:
		return inhaltGemischt
	case hatMarkup:
		return inhaltBlock
	case hatText:
		return inhaltText
	default:
		return inhaltLeer
	}
}

func schreibeElement(b *strings.Builder, src []rune, n *node, tiefe int) {
	schreibeEinrueckung(b, tiefe)
	schreibeStartTag(b, n.tok)
	if n.tok.selfClosing {
		b.WriteByte('\n')
		return
	}
	switch klassifiziere(n, src) {
	case inhaltLeer:
		schreibeEndTag(b, n.tok.name)
		b.WriteByte('\n')
	case inhaltText:
		for _, kind := range n.children {
			if kind.kind == nodeText || kind.kind == nodeCDATA {
				b.WriteString(textOf(src, kind.tok))
			}
		}
		schreibeEndTag(b, n.tok.name)
		b.WriteByte('\n')
	case inhaltGemischt:
		// Der Inhalt bleibt Zeichen für Zeichen stehen. Umbrüche und
		// Einrückungen von Hand bleiben erhalten; nur das äußere Tag wird
		// eingerückt.
		b.WriteString(string(src[n.tok.end:n.end.start]))
		schreibeEndTag(b, n.tok.name)
		b.WriteByte('\n')
	case inhaltBlock:
		b.WriteByte('\n')
		for _, kind := range n.children {
			switch kind.kind {
			case nodeElement:
				schreibeElement(b, src, kind, tiefe+1)
			case nodeComment, nodePI:
				schreibeEinrueckung(b, tiefe+1)
				b.WriteString(textOf(src, kind.tok))
				b.WriteByte('\n')
			}
		}
		schreibeEinrueckung(b, tiefe)
		schreibeEndTag(b, n.tok.name)
		b.WriteByte('\n')
	}
}

func schreibeElementDicht(b *strings.Builder, src []rune, n *node) {
	schreibeStartTag(b, n.tok)
	if n.tok.selfClosing {
		return
	}
	switch klassifiziere(n, src) {
	case inhaltText, inhaltGemischt:
		for _, kind := range n.children {
			if kind.kind == nodeText || kind.kind == nodeCDATA || kind.kind == nodeComment || kind.kind == nodePI {
				b.WriteString(textOf(src, kind.tok))
			} else if kind.kind == nodeElement {
				schreibeElementDicht(b, src, kind)
			}
		}
	case inhaltBlock:
		for _, kind := range n.children {
			switch kind.kind {
			case nodeElement:
				schreibeElementDicht(b, src, kind)
			case nodeComment, nodePI:
				b.WriteString(textOf(src, kind.tok))
			}
		}
	}
	schreibeEndTag(b, n.tok.name)
}

func schreibeEinrueckung(b *strings.Builder, tiefe int) {
	for i := 0; i < tiefe; i++ {
		for j := 0; j < eingerueckt; j++ {
			b.WriteByte(' ')
		}
	}
}

func schreibeStartTag(b *strings.Builder, t *token) {
	b.WriteByte('<')
	b.WriteString(t.name)
	for _, a := range t.attrs {
		b.WriteByte(' ')
		b.WriteString(a.name)
		b.WriteByte('=')
		b.WriteRune(a.quote)
		b.WriteString(a.value)
		b.WriteRune(a.quote)
	}
	if t.selfClosing {
		b.WriteString(" />")
	} else {
		b.WriteByte('>')
	}
}

func schreibeEndTag(b *strings.Builder, name string) {
	b.WriteString("</")
	b.WriteString(name)
	b.WriteByte('>')
}

// pruefeAusgabe liest die eigene Ausgabe erneut und vergleicht sie Stück für
// Stück mit der Eingabe. Was diese Prüfung nicht findet, geht nicht nach
// außen. Ein Fehler hier ist ein Fehler in diesem Programm, keine Eigenschaft
// der Eingabe.
func pruefeAusgabe(source []rune, soll []token, ausgabe string) error {
	ist, istQuelle, err := lex(ausgabe)
	if err != nil {
		return fmt.Errorf("Die eigene Ausgabe lässt sich nicht lesen: %w", err)
	}
	if _, err := parse(ist, istQuelle); err != nil {
		return fmt.Errorf("Die eigene Ausgabe ist nicht wohlgeformt: %w", err)
	}
	sollStrom := strom(source, soll)
	istStrom := strom(istQuelle, ist)
	if len(sollStrom) != len(istStrom) {
		return fmt.Errorf("Die eigene Ausgabe hat %d statt %d Stücke.", len(istStrom), len(sollStrom))
	}
	for i := range sollStrom {
		if sollStrom[i] != istStrom[i] {
			return fmt.Errorf("Die eigene Ausgabe weicht an Stelle %d ab: %s wurde zu %s.",
				i+1, sollStrom[i], istStrom[i])
		}
	}
	return nil
}

// strom ist die Vergleichsform eines Tokens. Leerraum, der nur zwischen
// Markup steht, zählt nicht: ihn hinzuzufügen oder wegzulassen ist der Zweck
// des Formatierers. Sichtbarer Text, Kommentare, CDATA und Deklarationen
// müssen dagegen Zeichen für Zeichen gleich bleiben.
func strom(source []rune, tokens []token) []string {
	out := make([]string, 0, len(tokens))
	for i := range tokens {
		t := &tokens[i]
		switch t.kind {
		case tokText:
			text := textOf(source, t)
			if !hasContent(text) {
				continue
			}
			out = append(out, "text:"+text)
		case tokStart:
			var b strings.Builder
			b.WriteString("start:")
			b.WriteString(t.name)
			if t.selfClosing {
				b.WriteString("/")
			}
			for _, a := range t.attrs {
				b.WriteString("|")
				b.WriteString(a.name)
				b.WriteString("=")
				b.WriteRune(a.quote)
				b.WriteString(a.value)
				b.WriteRune(a.quote)
			}
			out = append(out, b.String())
		case tokEnd:
			out = append(out, "end:"+t.name)
		case tokComment:
			out = append(out, "comment:"+textOf(source, t))
		case tokCDATA:
			out = append(out, "cdata:"+textOf(source, t))
		case tokPI:
			out = append(out, "pi:"+textOf(source, t))
		case tokDoctype:
			out = append(out, "doctype:"+textOf(source, t))
		}
	}
	return out
}
