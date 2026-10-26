package editing

import (
	"testing"
)

func TestInsertUndAuswahl(t *testing.T) {
	t.Parallel()
	m := New("hallo welt")
	m.SetCursor(5)
	m.Insert(",")
	if got := m.Text(); got != "hallo, welt" {
		t.Fatalf("Text = %q", got)
	}
	if m.Cursor() != 6 {
		t.Fatalf("Cursor = %d", m.Cursor())
	}
}

func TestAuswahlWirdErsetzt(t *testing.T) {
	t.Parallel()
	m := New("hallo welt")
	m.Select(6, 10)
	m.Insert("du")
	if got := m.Text(); got != "hallo du" {
		t.Fatalf("Text = %q", got)
	}
	if _, _, ok := m.Selection(); ok {
		t.Fatal("Auswahl steht noch")
	}
}

func TestBackspaceUndDelete(t *testing.T) {
	t.Parallel()
	m := New("abcd")
	m.SetCursor(2)
	if !m.Backspace() {
		t.Fatal("Backspace meldet nichts gelöscht")
	}
	if got := m.Text(); got != "acd" {
		t.Fatalf("Text = %q", got)
	}
	if !m.Delete() {
		t.Fatal("Delete meldet nichts gelöscht")
	}
	if got := m.Text(); got != "ad" {
		t.Fatalf("Text = %q", got)
	}
}

func TestZeilen(t *testing.T) {
	t.Parallel()
	m := New("eins\nzwei\n\nvier")
	if got := m.LineCount(); got != 4 {
		t.Fatalf("LineCount = %d", got)
	}
	start, end := m.Line(6)
	if start != 5 || end != 9 {
		t.Fatalf("Line(6) = %d,%d", start, end)
	}
	m.SetCursorLineColumn(2, 3)
	if m.Cursor() != 7 {
		t.Fatalf("Cursor = %d, want 7", m.Cursor())
	}
	line, col := m.CursorLineColumn()
	if line != 2 || col != 3 {
		t.Fatalf("LineColumn = %d,%d", line, col)
	}
}

func TestEinrueckungUebernehmen(t *testing.T) {
	t.Parallel()
	m := New("  <a>\n    <b/>\n  </a>")
	m.SetCursor(m.LineEnd(m.LineStart(6)))
	if got := m.LeadingWhitespace(m.Cursor()); got != "    " {
		t.Fatalf("LeadingWhitespace = %q", got)
	}
}

func TestUndoGruppiertTippen(t *testing.T) {
	t.Parallel()
	m := New("")
	m.Insert("a")
	m.Insert("b")
	m.Insert("c")
	if got := m.Text(); got != "abc" {
		t.Fatalf("Text = %q", got)
	}
	if !m.Undo() {
		t.Fatal("Undo meldet nichts")
	}
	if got := m.Text(); got != "" {
		t.Fatalf("nach Undo = %q", got)
	}
	if !m.Redo() {
		t.Fatal("Redo meldet nichts")
	}
	if got := m.Text(); got != "abc" {
		t.Fatalf("nach Redo = %q", got)
	}
}

func TestUndoNachBewegungIstEigenesStueck(t *testing.T) {
	t.Parallel()
	m := New("")
	m.Insert("a")
	m.MoveHorizontal(0, false)
	m.Insert("b")
	m.Undo()
	if got := m.Text(); got != "a" {
		t.Fatalf("nach Undo = %q", got)
	}
}

func TestIndentOutdent(t *testing.T) {
	t.Parallel()
	m := New("a\nb\nc")
	m.Select(0, len(m.Text()))
	m.IndentSelection()
	if got := m.Text(); got != "    a\n    b\n    c" {
		t.Fatalf("nach Indent = %q", got)
	}
	m.Select(0, len(m.Text()))
	m.OutdentSelection()
	if got := m.Text(); got != "a\nb\nc" {
		t.Fatalf("nach Outdent = %q", got)
	}
}

func TestSetTextLoeschtUndo(t *testing.T) {
	t.Parallel()
	m := New("alt")
	m.Insert("!")
	m.SetText("neu")
	if m.CanUndo() {
		t.Fatal("Undo lebt nach SetText weiter")
	}
}
