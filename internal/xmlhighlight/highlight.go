// Package xmlhighlight zerlegt XML-Text in farbige Bereiche. Der Scanner ist
// absichtlich nachsichtig: er läuft auch über unvollständiges XML, denn
// eingefärbt wird, während der Nutzer tippt. Die eigentliche Prüfung macht
// internal/xmlfmt.
package xmlhighlight

// Kind ist die Bedeutung eines Bereichs.
type Kind int

const (
	Plain Kind = iota
	Bracket
	TagName
	AttrName
	AttrValue
	Comment
	CDATA
	Doctype
	ProcessingInstruction
	Entity
)

// Span ist ein halboffener Bereich [Start, End) in Runen.
type Span struct {
	Start int
	End   int
	Kind  Kind
}

// Scan liefert lückenlose Bereiche in Runen-Positionen. Zeilenumbrüche gehören
// zum umgebenden Text; der Editor fragt ohnehin nur Zeilenbereiche ab.
func Scan(text string) []Span {
	r := []rune(text)
	var out []Span

	add := func(start, end int, kind Kind) {
		if start >= end {
			return
		}
		if start < 0 {
			start = 0
		}
		if end > len(r) {
			end = len(r)
		}
		if start >= end {
			return
		}
		if len(out) > 0 && out[len(out)-1].Kind == kind && out[len(out)-1].End == start {
			out[len(out)-1].End = end
			return
		}
		out = append(out, Span{Start: start, End: end, Kind: kind})
	}

	i := 0
	for i < len(r) {
		switch {
		case hasPrefix(r, i, "<!--"):
			end := afterToken(r, i+4, "-->")
			add(i, end, Comment)
			i = end
		case hasPrefix(r, i, "<![CDATA["):
			end := afterToken(r, i+9, "]]>")
			add(i, end, CDATA)
			i = end
		case hasPrefix(r, i, "<?"):
			end := afterToken(r, i+2, "?>")
			add(i, end, ProcessingInstruction)
			i = end
		case hasPrefix(r, i, "<!"):
			end := doctypeEnd(r, i+2)
			add(i, end, Doctype)
			i = end
		case r[i] == '<':
			end := tag(r, i, add)
			i = end
		default:
			end := i
			for end < len(r) && r[end] != '<' {
				end++
			}
			textRun(r, i, end, add)
			i = end
		}
	}
	return lueckenFuelle(out, len(r))
}

// lueckenFuelle macht aus den gesammelten Bereichen eine lückenlose Folge.
// Zwischen Attributen und um Gleichheitszeichen stehen Leerzeichen, die keine
// eigene Bedeutung haben; sie werden als Plain geführt. Der Editor kann sich
// dadurch auf eine einzige Schleife verlassen.
func lueckenFuelle(spans []Span, gesamt int) []Span {
	if gesamt == 0 {
		return nil
	}
	out := make([]Span, 0, len(spans)+4)
	pos := 0
	for _, s := range spans {
		if s.Start > pos {
			out = ergaenze(out, pos, s.Start, Plain)
		}
		out = ergaenze(out, s.Start, s.End, s.Kind)
		pos = s.End
	}
	if pos < gesamt {
		out = ergaenze(out, pos, gesamt, Plain)
	}
	return out
}

func ergaenze(spans []Span, start, end int, kind Kind) []Span {
	if start >= end {
		return spans
	}
	if len(spans) > 0 && spans[len(spans)-1].Kind == kind && spans[len(spans)-1].End == start {
		spans[len(spans)-1].End = end
		return spans
	}
	return append(spans, Span{Start: start, End: end, Kind: kind})
}

func hasPrefix(r []rune, pos int, prefix string) bool {
	p := []rune(prefix)
	if pos+len(p) > len(r) {
		return false
	}
	for i := range p {
		if r[pos+i] != p[i] {
			return false
		}
	}
	return true
}

// afterToken liefert die Position hinter dem Abschluss oder das Textende,
// wenn er fehlt. Ein offener Kommentar wird bis zum Ende eingefärbt.
func afterToken(r []rune, from int, token string) int {
	t := []rune(token)
	for i := from; i+len(t) <= len(r); i++ {
		match := true
		for j := range t {
			if r[i+j] != t[j] {
				match = false
				break
			}
		}
		if match {
			return i + len(t)
		}
	}
	return len(r)
}

// tag färbt einen Start- oder Endtag und liefert das Ende hinter '>'.
func tag(r []rune, start int, add func(int, int, Kind)) int {
	n := len(r)
	i := start
	add(i, i+1, Bracket)
	i++

	if i < n && r[i] == '/' {
		add(i, i+1, Bracket)
		i++
	}

	nameStart := i
	for i < n && !isSpace(r[i]) && r[i] != '>' && r[i] != '/' {
		i++
	}
	if i > nameStart {
		add(nameStart, i, TagName)
	} else {
		// "<" ohne Namen: der Rest der Zeile bleibt Text.
		for i < n && r[i] != '>' {
			i++
		}
		if i < n {
			i++
		}
		return i
	}

	for i < n {
		switch {
		case r[i] == '>':
			add(i, i+1, Bracket)
			return i + 1
		case r[i] == '/' && i+1 < n && r[i+1] == '>':
			add(i, i+2, Bracket)
			return i + 2
		case isSpace(r[i]):
			i++
			continue
		}

		attrStart := i
		for i < n && !isSpace(r[i]) && r[i] != '=' && r[i] != '>' && r[i] != '/' {
			i++
		}
		if i > attrStart {
			add(attrStart, i, AttrName)
		}
		for i < n && isSpace(r[i]) {
			i++
		}
		if i < n && r[i] == '=' {
			add(i, i+1, Bracket)
			i++
			for i < n && isSpace(r[i]) {
				i++
			}
			if i < n && (r[i] == '"' || r[i] == '\'') {
				quote := r[i]
				valueStart := i
				i++
				contentStart := i
				for i < n && r[i] != quote {
					i++
				}
				contentEnd := i
				if i < n {
					i++
				}
				add(valueStart, contentStart, AttrValue)
				wertRun(r, contentStart, contentEnd, add)
				add(contentEnd, i, AttrValue)
			} else {
				valueStart := i
				for i < n && !isSpace(r[i]) && r[i] != '>' {
					i++
				}
				if i > valueStart {
					add(valueStart, i, AttrValue)
				}
			}
		}
		if i == attrStart {
			// Kein Fortschritt möglich (etwa bei einem einzelnen '/'); das
			// Zeichen bleibt Teil des Tags, damit die Schleife endet.
			add(i, i+1, Bracket)
			i++
		}
	}
	return i
}

// doctypeEnd findet das '>' hinter einer DOCTYPE-Deklaration und beachtet
// dabei Anführungszeichen und die interne Teilmenge in eckigen Klammern.
func doctypeEnd(r []rune, from int) int {
	depth := 0
	var quote rune
	for i := from; i < len(r); i++ {
		if quote != 0 {
			if r[i] == quote {
				quote = 0
			}
			continue
		}
		switch r[i] {
		case '"', '\'':
			quote = r[i]
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case '>':
			if depth == 0 {
				return i + 1
			}
		}
	}
	return len(r)
}

// textRun färbt Text ein und hebt dabei Entitäten hervor.
func textRun(r []rune, start, end int, add func(int, int, Kind)) {
	runWithKind(r, start, end, add, Plain)
}

// wertRun tut dasselbe für Attributwerte. Die Anführungszeichen bleiben
// AttrValue; nur die Entitäten darin bekommen ihre eigene Farbe.
func wertRun(r []rune, start, end int, add func(int, int, Kind)) {
	runWithKind(r, start, end, add, AttrValue)
}

func runWithKind(r []rune, start, end int, add func(int, int, Kind), basis Kind) {
	i := start
	plain := start
	for i < end {
		if r[i] != '&' {
			i++
			continue
		}
		semi := entityEnd(r, i, end)
		if semi < 0 {
			i++
			continue
		}
		add(plain, i, basis)
		add(i, semi, Entity)
		i = semi
		plain = semi
	}
	add(plain, end, basis)
}

// entityEnd liefert das Ende einer Entität oder -1, wenn keine vorliegt.
func entityEnd(r []rune, start, end int) int {
	limit := start + 1
	if limit > end {
		return -1
	}
	for i := start + 1; i < end && i < start+34; i++ {
		switch r[i] {
		case ';':
			if i == start+1 {
				return -1
			}
			return i + 1
		case '&', '<', '>', '"', '\'', ' ', '\t', '\n', '\r', '=':
			return -1
		}
	}
	return -1
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
