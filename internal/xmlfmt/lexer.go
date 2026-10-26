package xmlfmt

type tokenKind int

const (
	tokText tokenKind = iota
	tokStart
	tokEnd
	tokComment
	tokCDATA
	tokPI
	tokDoctype
)

type attr struct {
	name   string
	value  string
	quote  rune
	line   int
	column int
}

type token struct {
	kind        tokenKind
	name        string
	attrs       []attr
	selfClosing bool
	line        int
	column      int
	start       int
	end         int
}

type lexer struct {
	src    []rune
	pos    int
	line   int
	column int
	tokens []token
}

func lex(input string) ([]token, []rune, error) {
	normalized := normalize(input)
	src := []rune(normalized)
	l := &lexer{src: src, line: 1, column: 1}
	if err := l.run(); err != nil {
		return nil, nil, err
	}
	return l.tokens, src, nil
}

func (l *lexer) run() error {
	for l.pos < len(l.src) {
		if l.src[l.pos] == '<' {
			if err := l.markup(); err != nil {
				return err
			}
			continue
		}
		if err := l.text(); err != nil {
			return err
		}
	}
	return nil
}

func (l *lexer) markup() error {
	line, column := l.line, l.column
	switch {
	case l.prefix("<!--"):
		return l.comment(line, column)
	case l.prefix("<![CDATA["):
		return l.cdata(line, column)
	case l.prefix("<!DOCTYPE"):
		return l.doctype(line, column)
	case l.prefix("<?"):
		return l.processingInstruction(line, column)
	case l.prefix("</"):
		return l.endTag(line, column)
	case l.prefix("<!"):
		return fehler(line, column, "Unbekannte Deklaration. Erlaubt sind <!-- -->, <![CDATA[ ]]> und <!DOCTYPE ...>.")
	default:
		return l.startTag(line, column)
	}
}

func (l *lexer) text() error {
	start, line, column := l.pos, l.line, l.column
	for l.pos < len(l.src) && l.src[l.pos] != '<' {
		switch {
		case l.src[l.pos] == '&':
			if err := l.entity(); err != nil {
				return err
			}
		case l.src[l.pos] == ']' && l.pos+2 < len(l.src) &&
			l.src[l.pos+1] == ']' && l.src[l.pos+2] == '>':
			return fehler(l.line, l.column, "Die Zeichenfolge ']]>' darf nicht im Text stehen.")
		default:
			l.advance()
		}
	}
	l.tokens = append(l.tokens, token{kind: tokText, line: line, column: column, start: start, end: l.pos})
	return nil
}

func (l *lexer) comment(line, column int) error {
	start := l.pos
	l.consume(len("<!--"))
	for {
		if l.pos >= len(l.src) {
			return fehler(line, column, "Kommentar wird nicht geschlossen.")
		}
		if l.prefix("-->") {
			l.consume(3)
			break
		}
		if l.prefix("--") {
			return fehler(l.line, l.column, "Im Kommentar darf '--' nicht vorkommen.")
		}
		l.advance()
	}
	l.tokens = append(l.tokens, token{kind: tokComment, line: line, column: column, start: start, end: l.pos})
	return nil
}

func (l *lexer) cdata(line, column int) error {
	start := l.pos
	l.consume(len("<![CDATA["))
	for {
		if l.pos >= len(l.src) {
			return fehler(line, column, "CDATA-Abschnitt wird nicht geschlossen.")
		}
		if l.prefix("]]>") {
			l.consume(3)
			break
		}
		l.advance()
	}
	l.tokens = append(l.tokens, token{kind: tokCDATA, line: line, column: column, start: start, end: l.pos})
	return nil
}

func (l *lexer) processingInstruction(line, column int) error {
	start := l.pos
	l.consume(2)
	name, err := l.name()
	if err != nil {
		return err
	}
	for {
		if l.pos >= len(l.src) {
			return fehler(line, column, "Verarbeitungsanweisung wird nicht geschlossen.")
		}
		if l.prefix("?>") {
			l.consume(2)
			break
		}
		l.advance()
	}
	l.tokens = append(l.tokens, token{kind: tokPI, name: name, line: line, column: column, start: start, end: l.pos})
	return nil
}

func (l *lexer) doctype(line, column int) error {
	start := l.pos
	l.consume(len("<!DOCTYPE"))
	if l.pos >= len(l.src) || !isSpace(l.src[l.pos]) {
		return fehler(l.line, l.column, "Nach <!DOCTYPE muss ein Leerzeichen und ein Name kommen.")
	}
	l.skipSpace()
	name, err := l.name()
	if err != nil {
		return err
	}

	depth := 0
	var quote rune
	for {
		if l.pos >= len(l.src) {
			return fehler(line, column, "Dokumenttyp-Deklaration wird nicht geschlossen.")
		}
		r := l.src[l.pos]
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			l.advance()
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
			l.advance()
		case '[':
			depth++
			l.advance()
		case ']':
			if depth > 0 {
				depth--
			}
			l.advance()
		case '>':
			if depth == 0 {
				l.advance()
				l.tokens = append(l.tokens, token{kind: tokDoctype, name: name, line: line, column: column, start: start, end: l.pos})
				return nil
			}
			l.advance()
		default:
			l.advance()
		}
	}
}

func (l *lexer) startTag(line, column int) error {
	start := l.pos
	l.advance() // '<'
	name, err := l.name()
	if err != nil {
		return err
	}

	var attrs []attr
	gesehen := map[string]bool{}
	selbst := false
	for {
		vorLuft := l.pos
		l.skipSpace()
		luft := l.pos > vorLuft
		if l.pos >= len(l.src) {
			return fehler(line, column, "Tag <%s> wird nicht geschlossen.", name)
		}
		if l.src[l.pos] == '>' {
			l.advance()
			break
		}
		if l.src[l.pos] == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '>' {
			l.advance()
			l.advance()
			selbst = true
			break
		}
		if l.src[l.pos] == '/' {
			return fehler(l.line, l.column, "Auf '/' im Tag muss '>' folgen.")
		}
		if !luft {
			return fehler(l.line, l.column, "Zwischen Attributen fehlt ein Leerzeichen.")
		}

		attrLine, attrColumn := l.line, l.column
		attrName, err := l.name()
		if err != nil {
			return err
		}
		if gesehen[attrName] {
			return fehler(attrLine, attrColumn, "Das Attribut '%s' steht mehrfach im selben Tag.", attrName)
		}
		gesehen[attrName] = true

		l.skipSpace()
		if l.pos >= len(l.src) || l.src[l.pos] != '=' {
			return fehler(l.line, l.column, "Nach dem Attributnamen '%s' wird '=' erwartet.", attrName)
		}
		l.advance()
		l.skipSpace()
		if l.pos >= len(l.src) || (l.src[l.pos] != '"' && l.src[l.pos] != '\'') {
			return fehler(l.line, l.column, "Der Wert von '%s' muss in Anführungszeichen stehen.", attrName)
		}
		quote := l.src[l.pos]
		l.advance()
		valueStart := l.pos
		for {
			if l.pos >= len(l.src) {
				return fehler(attrLine, attrColumn, "Der Wert von '%s' wird nicht geschlossen.", attrName)
			}
			r := l.src[l.pos]
			if r == quote {
				break
			}
			if r == '<' {
				return fehler(l.line, l.column, "Im Wert von '%s' darf kein '<' stehen. Es gehört als &lt; hinein.", attrName)
			}
			if r == '&' {
				if err := l.entity(); err != nil {
					return err
				}
				continue
			}
			l.advance()
		}
		value := string(l.src[valueStart:l.pos])
		l.advance() // schliessendes Anfuehrungszeichen
		attrs = append(attrs, attr{name: attrName, value: value, quote: quote, line: attrLine, column: attrColumn})
	}

	l.tokens = append(l.tokens, token{
		kind: tokStart, name: name, attrs: attrs, selfClosing: selbst,
		line: line, column: column, start: start, end: l.pos,
	})
	return nil
}

func (l *lexer) endTag(line, column int) error {
	start := l.pos
	l.consume(2) // "</"
	name, err := l.name()
	if err != nil {
		return err
	}
	l.skipSpace()
	if l.pos >= len(l.src) || l.src[l.pos] != '>' {
		return fehler(l.line, l.column, "Nach </%s wird '>' erwartet.", name)
	}
	l.advance()
	l.tokens = append(l.tokens, token{kind: tokEnd, name: name, line: line, column: column, start: start, end: l.pos})
	return nil
}

// name liest einen XML-Namen und lässt Position und Spalte stehen, wo der
// Name endet. Ein leerer Name ist ein Fehler mit der Stelle, an der er
// anfangen müsste.
func (l *lexer) name() (string, error) {
	if l.pos >= len(l.src) || !isNameStart(l.src[l.pos]) {
		return "", fehler(l.line, l.column, "Hier wird ein Name erwartet.")
	}
	start := l.pos
	l.advance()
	for l.pos < len(l.src) && isNameChar(l.src[l.pos]) {
		l.advance()
	}
	return string(l.src[start:l.pos]), nil
}

// entity prüft eine Entity-Referenz und schluckt sie vollständig. Der Aufrufer
// steht dabei auf dem '&'. Benannte Entities werden nicht aufgelöst: eine
// DOCTYPE-Deklaration darf eigene definieren, und raten wäre hier falsch.
func (l *lexer) entity() error {
	line, column := l.line, l.column
	l.advance() // '&'
	if l.pos >= len(l.src) {
		return fehler(line, column, "'&' muss als Entity geschrieben werden, zum Beispiel &amp;.")
	}
	if l.src[l.pos] == '#' {
		l.advance()
		hex := false
		if l.pos < len(l.src) && (l.src[l.pos] == 'x' || l.src[l.pos] == 'X') {
			hex = true
			l.advance()
		}
		ziffern := 0
		for l.pos < len(l.src) {
			r := l.src[l.pos]
			if hex && isHex(r) || !hex && isDigit(r) {
				ziffern++
				l.advance()
				continue
			}
			break
		}
		if ziffern == 0 {
			return fehler(line, column, "Nach '&#' muss mindestens eine Ziffer kommen.")
		}
		if l.pos >= len(l.src) || l.src[l.pos] != ';' {
			return fehler(line, column, "Die Zeichenreferenz wird nicht mit ';' geschlossen.")
		}
		l.advance()
		return nil
	}
	if l.pos >= len(l.src) || !isNameStart(l.src[l.pos]) {
		return fehler(line, column, "'&' muss als Entity geschrieben werden, zum Beispiel &amp;.")
	}
	for l.pos < len(l.src) && isNameChar(l.src[l.pos]) {
		l.advance()
	}
	if l.pos >= len(l.src) || l.src[l.pos] != ';' {
		return fehler(line, column, "Die Entity wird nicht mit ';' geschlossen.")
	}
	l.advance()
	return nil
}

func (l *lexer) advance() {
	if l.pos >= len(l.src) {
		return
	}
	if l.src[l.pos] == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.pos++
}

func (l *lexer) consume(n int) {
	for i := 0; i < n; i++ {
		l.advance()
	}
}

func (l *lexer) skipSpace() {
	for l.pos < len(l.src) && isSpace(l.src[l.pos]) {
		l.advance()
	}
}

func (l *lexer) prefix(s string) bool {
	runes := []rune(s)
	if l.pos+len(runes) > len(l.src) {
		return false
	}
	for i, r := range runes {
		if l.src[l.pos+i] != r {
			return false
		}
	}
	return true
}

func (l *lexer) raw(source []rune, t *token) string {
	return string(source[t.start:t.end])
}

func hasContent(text string) bool {
	// Nur XML-Leerraum ist bedeutungslos. Zeichen wie ein geschütztes
	// Leerzeichen (U+00A0) sind sichtbarer Inhalt und dürfen beim Formen
	// nicht verschwinden.
	return TrimSpace(text) != ""
}
