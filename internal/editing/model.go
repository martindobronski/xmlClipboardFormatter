// Package editing enthält den Textpuffer des Editors: Cursor, Auswahl,
// Einfügen, Löschen und Undo. Es kennt weder Fyne noch XML und ist deshalb
// ohne Fenster testbar.
package editing

import (
	"strings"
	"time"
)

const (
	maxUndo       = 500
	undoGruppenMs = 700
)

type snapshot struct {
	text   []rune
	cursor int
	anchor int
}

type editKind int

const (
	editInsert editKind = iota + 1
	editDelete
	editOther
)

// Model ist ein Textpuffer mit Cursor (an einer Zeichenposition) und Anker.
// Cursor und Anker begrenzen die Auswahl; sind sie gleich, ist nichts
// ausgewählt.
type Model struct {
	text   []rune
	cursor int
	anchor int

	undo     []snapshot
	redo     []snapshot
	lastPush time.Time
	lastKind editKind
}

func New(text string) *Model {
	m := &Model{}
	m.SetText(text)
	return m
}

func (m *Model) SetText(text string) {
	m.text = []rune(text)
	m.cursor = 0
	m.anchor = 0
	m.undo = nil
	m.redo = nil
	m.lastPush = time.Time{}
	m.lastKind = 0
}

func (m *Model) Text() string { return string(m.text) }

// Runes gibt eine Kopie zurück, damit die Oberfläche während des Zeichnens
// nicht mit einer Änderung im selben Speicher kollidiert.
func (m *Model) Runes() []rune {
	out := make([]rune, len(m.text))
	copy(out, m.text)
	return out
}

func (m *Model) Len() int { return len(m.text) }

func (m *Model) Cursor() int { return m.cursor }

func (m *Model) Anchor() int { return m.anchor }

func (m *Model) Selection() (start, end int, ok bool) {
	if m.cursor == m.anchor {
		return m.cursor, m.cursor, false
	}
	if m.cursor < m.anchor {
		return m.cursor, m.anchor, true
	}
	return m.anchor, m.cursor, true
}

func (m *Model) SelectedText() string {
	start, end, ok := m.Selection()
	if !ok {
		return ""
	}
	return string(m.text[start:end])
}

func (m *Model) SetCursor(pos int) {
	m.cursor = clamp(pos, 0, len(m.text))
	m.anchor = m.cursor
	m.lastPush = time.Time{}
	m.lastKind = 0
}

func (m *Model) SetCursorExtend(pos int) {
	m.cursor = clamp(pos, 0, len(m.text))
	m.lastPush = time.Time{}
	m.lastKind = 0
}

func (m *Model) SelectAll() {
	m.anchor = 0
	m.cursor = len(m.text)
	m.lastPush = time.Time{}
	m.lastKind = 0
}

func (m *Model) Select(start, end int) {
	m.anchor = clamp(start, 0, len(m.text))
	m.cursor = clamp(end, 0, len(m.text))
	m.lastPush = time.Time{}
	m.lastKind = 0
}

func (m *Model) Insert(s string) {
	runes := []rune(s)
	if len(runes) == 0 {
		return
	}
	m.begin(editInsert)
	m.ersetzeAuswahl(runes)
	m.cursor += len(runes)
	m.anchor = m.cursor
}

func (m *Model) Backspace() bool {
	if start, end, ok := m.Selection(); ok {
		m.begin(editDelete)
		m.loesche(start, end)
		return true
	}
	if m.cursor == 0 {
		return false
	}
	m.begin(editDelete)
	m.loesche(m.cursor-1, m.cursor)
	return true
}

func (m *Model) Delete() bool {
	if start, end, ok := m.Selection(); ok {
		m.begin(editDelete)
		m.loesche(start, end)
		return true
	}
	if m.cursor >= len(m.text) {
		return false
	}
	m.begin(editDelete)
	m.loesche(m.cursor, m.cursor+1)
	return true
}

// MoveHorizontal bewegt den Cursor um delta Zeichen. extend hält den Anker
// fest, sodass eine Auswahl entsteht.
func (m *Model) MoveHorizontal(delta int, extend bool) {
	ziel := clamp(m.cursor+delta, 0, len(m.text))
	m.move(ziel, extend)
}

func (m *Model) MoveTo(target int, extend bool) {
	m.move(clamp(target, 0, len(m.text)), extend)
}

func (m *Model) MoveLineStart(extend bool) {
	m.move(m.LineStart(m.cursor), extend)
}

func (m *Model) MoveLineEnd(extend bool) {
	m.move(m.LineEnd(m.cursor), extend)
}

func (m *Model) move(target int, extend bool) {
	m.cursor = clamp(target, 0, len(m.text))
	if !extend {
		m.anchor = m.cursor
	}
	m.lastPush = time.Time{}
	m.lastKind = 0
}

func (m *Model) Undo() bool {
	if len(m.undo) == 0 {
		return false
	}
	m.redo = append(m.redo, m.snapshot())
	s := m.undo[len(m.undo)-1]
	m.undo = m.undo[:len(m.undo)-1]
	m.restore(s)
	m.lastPush = time.Time{}
	m.lastKind = 0
	return true
}

func (m *Model) Redo() bool {
	if len(m.redo) == 0 {
		return false
	}
	m.undo = append(m.undo, m.snapshot())
	s := m.redo[len(m.redo)-1]
	m.redo = m.redo[:len(m.redo)-1]
	m.restore(s)
	m.lastPush = time.Time{}
	m.lastKind = 0
	return true
}

func (m *Model) CanUndo() bool { return len(m.undo) > 0 }
func (m *Model) CanRedo() bool { return len(m.redo) > 0 }

func (m *Model) snapshot() snapshot {
	text := make([]rune, len(m.text))
	copy(text, m.text)
	return snapshot{text: text, cursor: m.cursor, anchor: m.anchor}
}

func (m *Model) restore(s snapshot) {
	m.text = s.text
	m.cursor = clamp(s.cursor, 0, len(m.text))
	m.anchor = clamp(s.anchor, 0, len(m.text))
}

// begin legt einen Undo-Punkt an. Aufeinanderfolgende Änderungen derselben Art
// innerhalb eines kurzen Fensters teilen sich einen Punkt, sonst bräuchte es
// für jeden Buchstaben ein eigenes Undo.
func (m *Model) begin(kind editKind) {
	jetzt := time.Now()
	if m.lastKind == kind && kind != editOther && !m.lastPush.IsZero() &&
		jetzt.Sub(m.lastPush) < undoGruppenMs*time.Millisecond {
		return
	}
	m.undo = append(m.undo, m.snapshot())
	if len(m.undo) > maxUndo {
		m.undo = m.undo[len(m.undo)-maxUndo:]
	}
	m.redo = nil
	m.lastPush = jetzt
	m.lastKind = kind
}

func (m *Model) ersetzeAuswahl(einfuegen []rune) {
	start, end, ok := m.Selection()
	if !ok {
		start, end = m.cursor, m.cursor
	}
	neu := make([]rune, 0, len(m.text)-(end-start)+len(einfuegen))
	neu = append(neu, m.text[:start]...)
	neu = append(neu, einfuegen...)
	neu = append(neu, m.text[end:]...)
	m.text = neu
	m.cursor = start
	m.anchor = start
}

func (m *Model) loesche(start, end int) {
	neu := make([]rune, 0, len(m.text)-(end-start))
	neu = append(neu, m.text[:start]...)
	neu = append(neu, m.text[end:]...)
	m.text = neu
	m.cursor = start
	m.anchor = start
}

// LineStart liefert den Anfang der Zeile, in der pos steht. Ein pos direkt
// hinter einem '\n' ist der Anfang der nächsten Zeile.
func (m *Model) LineStart(pos int) int {
	pos = clamp(pos, 0, len(m.text))
	for pos > 0 && m.text[pos-1] != '\n' {
		pos--
	}
	return pos
}

// LineEnd liefert die Position des '\n' am Zeilenende oder das Textende.
func (m *Model) LineEnd(pos int) int {
	pos = clamp(pos, 0, len(m.text))
	for pos < len(m.text) && m.text[pos] != '\n' {
		pos++
	}
	return pos
}

// Line liefert Anfang und Ende (ohne '\n') der Zeile, in der pos steht.
func (m *Model) Line(pos int) (start, end int) {
	return m.LineStart(pos), m.LineEnd(pos)
}

func (m *Model) LineCount() int {
	if len(m.text) == 0 {
		return 1
	}
	return 1 + strings.Count(string(m.text), "\n")
}

// CursorLineColumn liefert die Stelle ab 1, wie sie in Fehlermeldungen steht.
func (m *Model) CursorLineColumn() (line, column int) {
	line, column = 1, 1
	for i := 0; i < m.cursor && i < len(m.text); i++ {
		if m.text[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

// SetCursorLineColumn setzt den Cursor auf eine Stelle ab 1. Eine Spalte
// hinter dem Zeilenende landet am Zeilenende.
func (m *Model) SetCursorLineColumn(line, column int) {
	if line < 1 {
		line = 1
	}
	if column < 1 {
		column = 1
	}
	pos := 0
	for i := 1; i < line && pos < len(m.text); i++ {
		nl := -1
		for j := pos; j < len(m.text); j++ {
			if m.text[j] == '\n' {
				nl = j
				break
			}
		}
		if nl < 0 {
			pos = len(m.text)
			break
		}
		pos = nl + 1
	}
	ende := m.LineEnd(pos)
	pos = clamp(pos+column-1, pos, ende)
	m.SetCursor(pos)
}

// LeadingWhitespace liefert die Einrückung der Zeile, in der pos steht.
// Beim Einfügen einer neuen Zeile übernimmt der Editor sie.
func (m *Model) LeadingWhitespace(pos int) string {
	start := m.LineStart(pos)
	end := m.LineEnd(pos)
	i := start
	for i < end && (m.text[i] == ' ' || m.text[i] == '\t') {
		i++
	}
	return string(m.text[start:i])
}

// IndentSelection rückt alle Zeilen der Auswahl um vier Leerzeichen ein. Ohne
// Auswahl wird am Cursor eingefügt.
func (m *Model) IndentSelection() {
	start, end, ok := m.Selection()
	if !ok {
		m.Insert("    ")
		return
	}
	m.begin(editOther)
	start = m.LineStart(start)
	end = m.LineEnd(end)
	teil := string(m.text[start:end])
	zeilen := strings.Split(teil, "\n")
	for i := range zeilen {
		zeilen[i] = "    " + zeilen[i]
	}
	neu := strings.Join(zeilen, "\n")
	m.text = append(append(append([]rune{}, m.text[:start]...), []rune(neu)...), m.text[end:]...)
	m.anchor = start
	m.cursor = start + len([]rune(neu))
}

// OutdentSelection nimmt bis zu vier führende Leerzeichen von jeder Zeile der
// Auswahl zurück.
func (m *Model) OutdentSelection() {
	start, end, ok := m.Selection()
	if !ok {
		start = m.LineStart(m.cursor)
		end = m.LineEnd(m.cursor)
	}
	m.begin(editOther)
	start = m.LineStart(start)
	end = m.LineEnd(end)
	teil := string(m.text[start:end])
	zeilen := strings.Split(teil, "\n")
	for i := range zeilen {
		n := 0
		for n < len(zeilen[i]) && n < 4 && zeilen[i][n] == ' ' {
			n++
		}
		zeilen[i] = zeilen[i][n:]
	}
	neu := strings.Join(zeilen, "\n")
	m.text = append(append(append([]rune{}, m.text[:start]...), []rune(neu)...), m.text[end:]...)
	m.anchor = start
	m.cursor = start + len([]rune(neu))
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
