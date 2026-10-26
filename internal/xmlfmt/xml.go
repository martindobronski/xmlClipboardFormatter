// Package xmlfmt liest, prüft und formatiert XML ohne Netzwerk, ohne
// Fremdbibliothek und ohne Rekursion im Parser. Es ist die Fachlogik der
// Anwendung; die Oberfläche kennt kein XML.
package xmlfmt

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Error ist ein Fehler in der Eingabe mit einer Stelle, die man im Textfeld
// wiederfindet. Die Zeilen und Spalten zählen ab 1 und rechnen in Zeichen,
// nicht in Bytes.
type Error struct {
	Message string
	Line    int
	Column  int
}

func (e *Error) Error() string { return e.Meldung() }

// Meldung ist der Text, den die Oberfläche in die Statuszeile schreibt.
func (e *Error) Meldung() string {
	return fmt.Sprintf("Zeile %d, Spalte %d: %s", e.Line, e.Column, e.Message)
}

func fehler(line, column int, format string, args ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, args...), Line: line, Column: column}
}

// TrimSpace schneidet nur XML-Leerraum an den Rändern ab: Leerzeichen, Tab,
// Wagenrücklauf und Zeilenumbruch. Sichtbare Zeichen wie ein geschütztes
// Leerzeichen (U+00A0) bleiben stehen.
func TrimSpace(text string) string {
	return strings.Trim(text, " \t\r\n")
}

// LooksLikeXML entscheidet, ob ein Text als XML behandelt wird. Die Prüfung
// ist absichtlich eine Vermutung: sie läuft nach jedem Tastendruck und darf
// einen halb getippten Text nicht sperren. Der Formatter bleibt der Richter.
func LooksLikeXML(text string) bool {
	text = strings.TrimPrefix(text, "\uFEFF")
	text = strings.TrimLeft(text, " \t\r\n")
	if text == "" || text[0] != '<' {
		return false
	}
	rest := text[1:]
	if rest == "" {
		return false
	}
	switch rest[0] {
	case '!', '?':
		return true
	}
	r, _ := utf8.DecodeRuneInString(rest)
	return isNameStart(r)
}

// normalize vereinheitlicht Zeilenenden und entfernt ein BOM am Anfang.
// XML schreibt Zeilenenden beim Lesen ohnehin auf LF um; hier wird dieselbe
// Regel auf den Text angewendet, den die App wieder ausgibt.
func normalize(text string) string {
	text = strings.TrimPrefix(text, "\uFEFF")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func isNameStart(r rune) bool {
	return r == '_' || r == ':' || unicode.IsLetter(r)
}

func isNameChar(r rune) bool {
	return isNameStart(r) || unicode.IsDigit(r) || r == '-' || r == '.' || r == '\u00B7' ||
		unicode.IsMark(r)
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func isHex(r rune) bool {
	return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
