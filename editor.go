package main

import (
	"math"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"xmlclipboardformatter/internal/editing"
	"xmlclipboardformatter/internal/xmlhighlight"
)

const (
	editorPadTop       = 9
	editorPadBottom    = 9
	editorPadRight     = 10
	editorGutter       = 50
	editorScrollbar    = 9
	editorMaxHighlight = 100_000
)

// visualRow ist eine Bildschirmzeile. Eine logische Zeile kann durch den
// weichen Umbruch mehrere Bildschirmzeilen belegen; die Zeilennummern zählen
// deshalb Bildschirmzeilen, genau wie im Java-Editor.
type visualRow struct {
	start int
	end   int
	line  int
}

// codeEditor ist ein eigenständiger XML-Editor: Textpuffer, Cursor, Auswahl,
// Einfärbung und Zeilennummern liegen in einer Widget-Schicht. Das Modell aus
// internal/editing bleibt frei von Fyne und ist ohne Fenster testbar.
type codeEditor struct {
	widget.BaseWidget

	model            *editing.Model
	runes            []rune
	spans            []xmlhighlight.Span
	spansVersion     int
	rows             []visualRow
	rowsVersion      int
	rowsBuiltVersion int
	rowsWidth        float32
	textSize         float32
	lineHeight       float32
	visibleRows      int
	textLeft         float32
	contentHeight    float32
	scrollRow        int
	advance          map[rune]float32
	advanceSize      float32

	focused  bool
	shift    bool
	selDrag  bool
	barDrag  bool
	barStart float32
	barRow   int

	OnChanged     func()
	OnFocusChange func(bool)
}

func newCodeEditor(text string) *codeEditor {
	e := &codeEditor{model: editing.New(text)}
	e.ExtendBaseWidget(e)
	e.syncRunes()
	e.spans = e.scan()
	e.rowsVersion = 1
	e.prepare(e.MinSize())
	return e
}

func (e *codeEditor) CreateRenderer() fyne.WidgetRenderer {
	r := &codeEditorRenderer{
		editor: e,
		caret:  canvas.NewRectangle(theme.Color(theme.ColorNameForeground)),
		track:  canvas.NewRectangle(theme.Color(theme.ColorNameScrollBarBackground)),
		thumb:  canvas.NewRectangle(theme.Color(theme.ColorNameScrollBar)),
	}
	r.caret.Resize(fyne.NewSize(1.5, 10))
	r.track.CornerRadius = 3
	r.thumb.CornerRadius = 3
	r.Refresh()
	return r
}

func (e *codeEditor) MinSize() fyne.Size {
	textSize := e.Theme().Size(sizeEditorText)
	if textSize <= 0 {
		textSize = 13
	}
	charWidth := fyne.MeasureText("M", textSize, fyne.TextStyle{Monospace: true}).Width
	lineHeight := fyne.MeasureText("Ag", textSize, fyne.TextStyle{Monospace: true}).Height + 3
	if lineHeight < textSize+4 {
		lineHeight = textSize + 4
	}
	width := float32(editorGutter) + charWidth*84 + editorPadRight + editorScrollbar
	height := editorPadTop + editorPadBottom + lineHeight*16
	return fyne.NewSize(width, height)
}

func (e *codeEditor) Text() string { return e.model.Text() }

// SetText ersetzt den gesamten Inhalt und meldet die Änderung. Eingaben aus
// der Zwischenablage enthalten unter Windows CRLF; im Puffer steht überall LF.
func (e *codeEditor) SetText(text string) {
	e.setText(text, 0)
	e.notify()
}

// SetTextAt ersetzt den Inhalt und setzt den Cursor an die frühere Position.
// Die Formatierer übernehmen damit die Stelle, an der man gerade arbeitet,
// ohne dass die Änderung ein zweites Mal gemeldet wird.
func (e *codeEditor) SetTextAt(text string, cursor int) {
	e.setText(text, cursor)
}

func (e *codeEditor) setText(text string, cursor int) {
	e.model.SetText(cleanText(text))
	if cursor > 0 {
		e.model.SetCursor(maxInt(0, minInt(cursor, e.model.Len())))
	}
	e.syncRunes()
	e.spans = e.scan()
	e.rowsVersion++
	e.prepare(e.Size())
	e.ensureCursorVisible()
	e.Refresh()
}

func (e *codeEditor) CursorLineColumn() (int, int) { return e.model.CursorLineColumn() }

func (e *codeEditor) CursorOffset() int { return e.model.Cursor() }

// DisplayLineCount zählt Bildschirmzeilen; vor dem ersten Layout sind die
// Absätze die beste verfügbare Näherung.
func (e *codeEditor) DisplayLineCount() int {
	if len(e.rows) > 0 {
		return len(e.rows)
	}
	return e.model.LineCount()
}

func (e *codeEditor) CharCount() int { return e.model.Len() }

func (e *codeEditor) HasSelection() bool {
	_, _, ok := e.model.Selection()
	return ok
}

func (e *codeEditor) FocusEditor() {
	if c := fyne.CurrentApp().Driver().CanvasForObject(e); c != nil {
		c.Focus(e)
	}
}

func (e *codeEditor) scan() []xmlhighlight.Span {
	if len(e.runes) > editorMaxHighlight {
		return nil
	}
	return xmlhighlight.Scan(string(e.runes))
}

func (e *codeEditor) syncRunes() {
	e.runes = e.model.Runes()
}

func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.ReplaceAll(text, "\x00", "")
}

func (e *codeEditor) prepare(size fyne.Size) {
	textSize := e.Theme().Size(sizeEditorText)
	if textSize <= 0 {
		textSize = 13
	}
	lineHeight := fyne.MeasureText("Ag", textSize, fyne.TextStyle{Monospace: true}).Height + 3
	if lineHeight < textSize+4 {
		lineHeight = textSize + 4
	}
	e.textSize = textSize
	e.lineHeight = lineHeight
	e.textLeft = editorGutter

	width := size.Width - e.textLeft - editorPadRight - editorScrollbar
	if width < 40 {
		width = 40
	}
	if e.rowsWidth != width || e.rowsBuiltVersion != e.rowsVersion {
		e.rows = e.buildRows(width)
		e.rowsWidth = width
		e.rowsBuiltVersion = e.rowsVersion
	}

	e.contentHeight = editorPadTop + float32(len(e.rows))*e.lineHeight + editorPadBottom
	e.visibleRows = int((size.Height - editorPadTop - editorPadBottom) / e.lineHeight)
	if e.visibleRows < 1 {
		e.visibleRows = 1
	}
	e.clampScroll()
}

func (e *codeEditor) buildRows(width float32) []visualRow {
	r := e.runes
	var rows []visualRow
	line := 0
	start := 0
	for start <= len(r) {
		end := start
		for end < len(r) && r[end] != '\n' {
			end++
		}
		if start == end {
			rows = append(rows, visualRow{start: start, end: end, line: line})
		} else {
			pos := start
			for pos < end {
				x := float32(0)
				i := pos
				lastSpace := -1
				for i < end {
					w := e.advanceFor(r[i])
					if x+w > width && i > pos {
						break
					}
					x += w
					if r[i] == ' ' || r[i] == '\t' {
						lastSpace = i
					}
					i++
				}
				chunkEnd := i
				if i < end && lastSpace >= pos {
					chunkEnd = lastSpace + 1
				}
				if chunkEnd <= pos {
					chunkEnd = pos + 1
				}
				rows = append(rows, visualRow{start: pos, end: chunkEnd, line: line})
				pos = chunkEnd
			}
		}
		if end >= len(r) {
			break
		}
		start = end + 1
		line++
	}
	if len(rows) == 0 {
		rows = []visualRow{{start: 0, end: 0, line: 0}}
	}
	return rows
}

func (e *codeEditor) advanceFor(r rune) float32 {
	if r == '\t' {
		return 4 * e.advanceFor(' ')
	}
	if e.advance == nil || e.advanceSize != e.textSize {
		e.advance = make(map[rune]float32)
		e.advanceSize = e.textSize
	}
	if w, ok := e.advance[r]; ok {
		return w
	}
	w := fyne.MeasureText(string(r), e.textSize, fyne.TextStyle{Monospace: true}).Width
	e.advance[r] = w
	return w
}

func (e *codeEditor) rowAtOffset(pos int) int {
	if len(e.rows) == 0 {
		return -1
	}
	for i, row := range e.rows {
		if pos >= row.start && pos < row.end {
			return i
		}
		if row.start == row.end && pos == row.start {
			return i
		}
	}
	for i := 0; i < len(e.rows)-1; i++ {
		if e.rows[i].end == pos && e.rows[i+1].line == e.rows[i].line &&
			e.rows[i+1].start == pos {
			return i + 1
		}
	}
	for i := len(e.rows) - 1; i >= 0; i-- {
		if pos >= e.rows[i].start && pos <= e.rows[i].end {
			return i
		}
	}
	return len(e.rows) - 1
}

func (e *codeEditor) xAtOffset(row visualRow, offset int) float32 {
	r := e.runes
	if offset < row.start {
		offset = row.start
	}
	if offset > row.end {
		offset = row.end
	}
	x := float32(0)
	for i := row.start; i < offset; i++ {
		x += e.advanceFor(r[i])
	}
	return x
}

func (e *codeEditor) offsetAtX(row visualRow, x float32) int {
	r := e.runes
	pos := row.start
	cur := float32(0)
	for pos < row.end {
		w := e.advanceFor(r[pos])
		if x < cur+w/2 {
			break
		}
		cur += w
		pos++
	}
	return pos
}

func (e *codeEditor) positionAt(pos fyne.Position) int {
	if len(e.rows) == 0 {
		return 0
	}
	y := pos.Y - editorPadTop
	idx := e.scrollRow
	if e.lineHeight > 0 {
		idx += int(y / e.lineHeight)
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(e.rows) {
		idx = len(e.rows) - 1
	}
	row := e.rows[idx]
	if pos.X <= e.textLeft {
		return row.start
	}
	return e.offsetAtX(row, pos.X-e.textLeft)
}

func (e *codeEditor) clampScroll() {
	maxRow := len(e.rows) - e.visibleRows
	if maxRow < 0 {
		maxRow = 0
	}
	if e.scrollRow > maxRow {
		e.scrollRow = maxRow
	}
	if e.scrollRow < 0 {
		e.scrollRow = 0
	}
}

func (e *codeEditor) ensureCursorVisible() {
	if len(e.rows) == 0 || e.visibleRows <= 0 {
		return
	}
	idx := e.rowAtOffset(e.model.Cursor())
	if idx < 0 {
		return
	}
	if idx < e.scrollRow {
		e.scrollRow = idx
	}
	if idx >= e.scrollRow+e.visibleRows {
		e.scrollRow = idx - e.visibleRows + 1
	}
	e.clampScroll()
}

func (e *codeEditor) maxScrollRow() int {
	maxRow := len(e.rows) - e.visibleRows
	if maxRow < 0 {
		maxRow = 0
	}
	return maxRow
}

func (e *codeEditor) afterEdit() {
	e.syncRunes()
	e.spans = e.scan()
	e.rowsVersion++
	e.prepare(e.Size())
	e.ensureCursorVisible()
	e.Refresh()
	e.notify()
}

func (e *codeEditor) afterMove() {
	e.prepare(e.Size())
	e.ensureCursorVisible()
	e.Refresh()
}

func (e *codeEditor) notify() {
	if e.OnChanged != nil {
		e.OnChanged()
	}
}

func (e *codeEditor) FocusGained() {
	e.focused = true
	e.Refresh()
	if e.OnFocusChange != nil {
		e.OnFocusChange(true)
	}
}

func (e *codeEditor) FocusLost() {
	e.focused = false
	e.shift = false
	e.Refresh()
	if e.OnFocusChange != nil {
		e.OnFocusChange(false)
	}
}

func (e *codeEditor) TypedRune(r rune) {
	if r == 0 || r < ' ' {
		return
	}
	e.model.Insert(string(r))
	e.afterEdit()
}

func (e *codeEditor) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyBackspace:
		if e.model.Backspace() {
			e.afterEdit()
		}
	case fyne.KeyDelete:
		if e.model.Delete() {
			e.afterEdit()
		}
	case fyne.KeyReturn, fyne.KeyEnter:
		indent := e.model.LeadingWhitespace(e.model.Cursor())
		e.model.Insert("\n" + indent)
		e.afterEdit()
	case fyne.KeyTab:
		if e.shift {
			e.model.OutdentSelection()
		} else {
			e.model.IndentSelection()
		}
		e.afterEdit()
	case fyne.KeyLeft:
		e.moveHorizontal(-1)
	case fyne.KeyRight:
		e.moveHorizontal(1)
	case fyne.KeyUp:
		e.moveVertical(-1)
	case fyne.KeyDown:
		e.moveVertical(1)
	case fyne.KeyPageUp:
		e.movePage(-1)
	case fyne.KeyPageDown:
		e.movePage(1)
	case fyne.KeyHome:
		e.model.MoveLineStart(e.shift)
		e.afterMove()
	case fyne.KeyEnd:
		e.model.MoveLineEnd(e.shift)
		e.afterMove()
	}
}

func (e *codeEditor) AcceptsTab() bool { return true }

func (e *codeEditor) KeyDown(key *fyne.KeyEvent) {
	if key.Name == desktop.KeyShiftLeft || key.Name == desktop.KeyShiftRight {
		e.shift = true
	}
}

func (e *codeEditor) KeyUp(key *fyne.KeyEvent) {
	if key.Name == desktop.KeyShiftLeft || key.Name == desktop.KeyShiftRight {
		e.shift = false
	}
}

func (e *codeEditor) moveHorizontal(delta int) {
	if start, end, ok := e.model.Selection(); ok && !e.shift {
		if delta < 0 {
			e.model.SetCursor(start)
		} else {
			e.model.SetCursor(end)
		}
		e.afterMove()
		return
	}
	e.model.MoveHorizontal(delta, e.shift)
	e.afterMove()
}

func (e *codeEditor) moveVertical(delta int) {
	e.moveToRow(e.rowAtOffset(e.model.Cursor())+delta, e.columnX())
}

func (e *codeEditor) movePage(direction int) {
	step := e.visibleRows - 1
	if step < 1 {
		step = 1
	}
	e.moveToRow(e.rowAtOffset(e.model.Cursor())+direction*step, e.columnX())
}

func (e *codeEditor) columnX() float32 {
	idx := e.rowAtOffset(e.model.Cursor())
	if idx < 0 {
		return 0
	}
	return e.xAtOffset(e.rows[idx], e.model.Cursor())
}

func (e *codeEditor) moveToRow(target int, x float32) {
	if len(e.rows) == 0 || target < 0 || target >= len(e.rows) {
		return
	}
	pos := e.offsetAtX(e.rows[target], x)
	e.model.MoveTo(pos, e.shift)
	e.afterMove()
}

func (e *codeEditor) TypedShortcut(shortcut fyne.Shortcut) {
	switch s := shortcut.(type) {
	case *fyne.ShortcutPaste:
		text := cleanText(s.Clipboard.Content())
		if text != "" {
			e.model.Insert(text)
			e.afterEdit()
		}
	case *fyne.ShortcutCopy:
		if text := e.model.SelectedText(); text != "" {
			s.Clipboard.SetContent(text)
		}
	case *fyne.ShortcutCut:
		if text := e.model.SelectedText(); text != "" {
			s.Clipboard.SetContent(text)
			e.model.Backspace()
			e.afterEdit()
		}
	case *fyne.ShortcutSelectAll:
		e.model.SelectAll()
		e.afterMove()
	case *fyne.ShortcutUndo:
		if e.model.Undo() {
			e.afterEdit()
		}
	case *fyne.ShortcutRedo:
		if e.model.Redo() {
			e.afterEdit()
		}
	case *desktop.CustomShortcut:
		e.customShortcut(s)
	}
}

func (e *codeEditor) customShortcut(s *desktop.CustomShortcut) {
	extend := e.shift || s.Modifier&fyne.KeyModifierShift != 0
	word := s.Modifier&fyne.KeyModifierAlt != 0
	line := false
	if runtime.GOOS == "darwin" {
		line = s.Modifier&fyne.KeyModifierShortcutDefault != 0
	} else {
		word = word || s.Modifier&fyne.KeyModifierShortcutDefault != 0
	}
	switch s.KeyName {
	case fyne.KeyLeft:
		switch {
		case word:
			e.model.MoveTo(e.wordLeft(e.model.Cursor()), extend)
		case line:
			e.model.MoveLineStart(extend)
		default:
			e.model.MoveHorizontal(-1, extend)
		}
		e.afterMove()
	case fyne.KeyRight:
		switch {
		case word:
			e.model.MoveTo(e.wordRight(e.model.Cursor()), extend)
		case line:
			e.model.MoveLineEnd(extend)
		default:
			e.model.MoveHorizontal(1, extend)
		}
		e.afterMove()
	case fyne.KeyZ:
		// Fyne liefert Cmd+Shift+Z beziehungsweise Strg+Shift+Z als eigenen
		// Shortcut, nicht als ShortcutRedo. Ohne diesen Zweig wäre er stumm.
		if s.Modifier&fyne.KeyModifierShift != 0 &&
			s.Modifier&fyne.KeyModifierShortcutDefault != 0 {
			if e.model.Redo() {
				e.afterEdit()
			}
		}
	case fyne.KeyUp:
		if s.Modifier&fyne.KeyModifierShortcutDefault != 0 {
			e.model.MoveTo(0, extend)
			e.afterMove()
		}
	case fyne.KeyDown:
		if s.Modifier&fyne.KeyModifierShortcutDefault != 0 {
			e.model.MoveTo(e.model.Len(), extend)
			e.afterMove()
		}
	}
}

func (e *codeEditor) wordLeft(pos int) int {
	r := e.runes
	for pos > 0 && unicode.IsSpace(r[pos-1]) {
		pos--
	}
	for pos > 0 && isWordRune(r[pos-1]) {
		pos--
	}
	return pos
}

func (e *codeEditor) wordRight(pos int) int {
	r := e.runes
	for pos < len(r) && unicode.IsSpace(r[pos]) {
		pos++
	}
	for pos < len(r) && isWordRune(r[pos]) {
		pos++
	}
	return pos
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == ':' || r == '.'
}

func (e *codeEditor) Tapped(*fyne.PointEvent) {
	e.FocusEditor()
}

func (e *codeEditor) TappedSecondary(pe *fyne.PointEvent) {
	e.FocusEditor()
	e.prepare(e.Size())
	pos := e.positionAt(pe.Position)
	// Ein Rechtsklick innerhalb der Auswahl lässt sie stehen; nur ein Klick
	// daneben setzt den Cursor neu. Sonst wäre "Kopieren" im Menü tot, sobald
	// man es mit der Maus öffnet.
	if start, end, ok := e.model.Selection(); ok {
		if pos < start || pos > end {
			e.model.SetCursor(pos)
		}
	} else {
		e.model.SetCursor(pos)
	}
	e.afterMove()

	driver := fyne.CurrentApp().Driver()
	c := driver.CanvasForObject(e)
	if c == nil {
		return
	}
	clipboard := fyne.CurrentApp().Clipboard()
	hasSel := e.HasSelection()
	cut := fyne.NewMenuItem("Ausschneiden", func() {
		text := e.model.SelectedText()
		if text != "" {
			clipboard.SetContent(text)
			e.model.Backspace()
			e.afterEdit()
		}
		e.FocusEditor()
	})
	cut.Disabled = !hasSel
	copyItem := fyne.NewMenuItem("Kopieren", func() {
		if text := e.model.SelectedText(); text != "" {
			clipboard.SetContent(text)
		}
		e.FocusEditor()
	})
	copyItem.Disabled = !hasSel
	paste := fyne.NewMenuItem("Einfügen", func() {
		text := cleanText(clipboard.Content())
		if text != "" {
			e.model.Insert(text)
			e.afterEdit()
		}
		e.FocusEditor()
	})
	paste.Disabled = clipboard.Content() == ""
	all := fyne.NewMenuItem("Alles auswählen", func() {
		e.model.SelectAll()
		e.afterMove()
		e.FocusEditor()
	})
	undo := fyne.NewMenuItem("Rückgängig", func() {
		if e.model.Undo() {
			e.afterEdit()
		}
		e.FocusEditor()
	})
	undo.Disabled = !e.model.CanUndo()
	redo := fyne.NewMenuItem("Wiederholen", func() {
		if e.model.Redo() {
			e.afterEdit()
		}
		e.FocusEditor()
	})
	redo.Disabled = !e.model.CanRedo()

	menu := widget.NewPopUpMenu(fyne.NewMenu("", cut, copyItem, paste, all,
		fyne.NewMenuItemSeparator(), undo, redo), c)
	menu.ShowAtPosition(driver.AbsolutePositionForObject(e).Add(pe.Position))
}

func (e *codeEditor) DoubleTapped(pe *fyne.PointEvent) {
	e.FocusEditor()
	e.prepare(e.Size())
	pos := e.positionAt(pe.Position)
	start, end := e.wordAt(pos)
	e.model.Select(start, end)
	e.selDrag = false
	e.afterMove()
}

func (e *codeEditor) wordAt(pos int) (int, int) {
	r := e.runes
	if len(r) == 0 {
		return 0, 0
	}
	if pos >= len(r) {
		pos = len(r) - 1
	}
	if isWordRune(r[pos]) {
		start, end := pos, pos+1
		for start > 0 && isWordRune(r[start-1]) {
			start--
		}
		for end < len(r) && isWordRune(r[end]) {
			end++
		}
		return start, end
	}
	if unicode.IsSpace(r[pos]) {
		start, end := pos, pos+1
		for start > 0 && unicode.IsSpace(r[start-1]) && r[start-1] != '\n' {
			start--
		}
		for end < len(r) && unicode.IsSpace(r[end]) && r[end] != '\n' {
			end++
		}
		return start, end
	}
	return pos, pos
}

func (e *codeEditor) MouseDown(m *desktop.MouseEvent) {
	e.FocusEditor()
	if m.Button != desktop.MouseButtonPrimary {
		return
	}
	if e.inScrollbar(m.Position) {
		e.barDrag = true
		e.barStart = m.Position.Y
		e.barRow = e.scrollRow
		return
	}
	e.prepare(e.Size())
	pos := e.positionAt(m.Position)
	if e.shift {
		e.model.SetCursorExtend(pos)
	} else {
		e.model.SetCursor(pos)
	}
	e.selDrag = true
	e.afterMove()
}

func (e *codeEditor) MouseUp(*desktop.MouseEvent) {
	e.selDrag = false
	e.barDrag = false
}

func (e *codeEditor) Dragged(d *fyne.DragEvent) {
	if e.barDrag {
		maxRow := e.maxScrollRow()
		_, trackH, _, thumbH, ok := e.scrollbarInfo()
		if ok && trackH-thumbH > 0 {
			delta := (d.Position.Y - e.barStart) / (trackH - thumbH) * float32(maxRow)
			e.scrollRow = e.barRow + int(math.Round(float64(delta)))
			e.clampScroll()
			e.Refresh()
		}
		return
	}
	if !e.selDrag {
		return
	}
	e.prepare(e.Size())
	e.model.SetCursorExtend(e.positionAt(d.Position))
	e.ensureCursorVisible()
	e.Refresh()
}

func (e *codeEditor) DragEnd() {
	e.selDrag = false
	e.barDrag = false
}

func (e *codeEditor) Scrolled(ev *fyne.ScrollEvent) {
	dy := ev.Scrolled.DY
	if dy == 0 {
		dy = ev.Scrolled.DX
	}
	step := int(math.Round(float64(dy)))
	if step == 0 && dy != 0 {
		if dy > 0 {
			step = 1
		} else {
			step = -1
		}
	}
	if step == 0 {
		return
	}
	e.prepare(e.Size())
	e.scrollRow -= step * 3
	e.clampScroll()
	e.Refresh()
}

func (e *codeEditor) inScrollbar(pos fyne.Position) bool {
	_, _, _, _, ok := e.scrollbarInfo()
	if !ok {
		return false
	}
	return pos.X >= e.Size().Width-editorScrollbar-5 &&
		pos.Y >= editorPadTop && pos.Y <= e.Size().Height-editorPadBottom
}

// scrollbarInfo liefert Track und Daumen in Widget-Koordinaten.
func (e *codeEditor) scrollbarInfo() (trackY, trackH, thumbY, thumbH float32, ok bool) {
	if e.visibleRows <= 0 || len(e.rows) <= e.visibleRows || e.lineHeight <= 0 {
		return 0, 0, 0, 0, false
	}
	trackY = editorPadTop
	trackH = e.Size().Height - editorPadTop - editorPadBottom
	if trackH <= 20 {
		return 0, 0, 0, 0, false
	}
	thumbH = trackH * float32(e.visibleRows) / float32(len(e.rows))
	if thumbH < 28 {
		thumbH = 28
	}
	if thumbH > trackH {
		thumbH = trackH
	}
	maxRow := e.maxScrollRow()
	frac := float32(0)
	if maxRow > 0 {
		frac = float32(e.scrollRow) / float32(maxRow)
	}
	thumbY = trackY + frac*(trackH-thumbH)
	return trackY, trackH, thumbY, thumbH, true
}

func (e *codeEditor) Cursor() desktop.Cursor { return desktop.TextCursor }

type codeEditorRenderer struct {
	editor *codeEditor

	selPool  []*canvas.Rectangle
	textPool []*canvas.Text
	numPool  []*canvas.Text

	caret *canvas.Rectangle
	track *canvas.Rectangle
	thumb *canvas.Rectangle

	objects []fyne.CanvasObject

	cPlain    fyne.ThemeColorName
	cNumber   fyne.ThemeColorName
	cString   fyne.ThemeColorName
	cComment  fyne.ThemeColorName
	cKeyword  fyne.ThemeColorName
	cName     fyne.ThemeColorName
	cOperator fyne.ThemeColorName
	cMuted    fyne.ThemeColorName
}

func (r *codeEditorRenderer) Destroy() {}

func (r *codeEditorRenderer) MinSize() fyne.Size { return r.editor.MinSize() }

func (r *codeEditorRenderer) Layout(size fyne.Size) {
	r.layoutContent(size)
}

func (r *codeEditorRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *codeEditorRenderer) Refresh() {
	r.cPlain = theme.ColorNameForeground
	r.cNumber = colorXMLNumber
	r.cString = colorXMLString
	r.cComment = colorXMLComment
	r.cKeyword = colorXMLKeyword
	r.cName = colorXMLName
	r.cOperator = colorXMLOperator
	r.cMuted = colorMuted
	r.layoutContent(r.editor.Size())
}

func (r *codeEditorRenderer) color(kind xmlhighlight.Kind) fyne.ThemeColorName {
	switch kind {
	case xmlhighlight.TagName:
		return r.cName
	case xmlhighlight.AttrName:
		return r.cNumber
	case xmlhighlight.AttrValue, xmlhighlight.CDATA:
		return r.cString
	case xmlhighlight.Comment:
		return r.cComment
	case xmlhighlight.Doctype, xmlhighlight.ProcessingInstruction:
		return r.cKeyword
	case xmlhighlight.Bracket:
		return r.cOperator
	case xmlhighlight.Entity:
		return r.cNumber
	}
	return r.cPlain
}

func (r *codeEditorRenderer) layoutContent(size fyne.Size) {
	e := r.editor
	e.prepare(size)
	r.objects = r.objects[:0]

	muted := theme.ColorForWidget(r.cMuted, e)

	// Auswahl zuerst: sie liegt unter dem Text.
	selIdx := 0
	if start, end, ok := e.model.Selection(); ok {
		for i := e.scrollRow; i < len(e.rows) && i < e.scrollRow+e.visibleRows; i++ {
			row := e.rows[i]
			if row.end <= start || row.start >= end {
				continue
			}
			a := maxInt(start, row.start)
			b := minInt(end, row.end)
			if b <= a {
				continue
			}
			x1 := e.textLeft + e.xAtOffset(row, a)
			x2 := e.textLeft + e.xAtOffset(row, b)
			if end > row.end {
				x2 += 4
			}
			if x2-x1 < 2 {
				x2 = x1 + 2
			}
			rect := r.rectAt(selIdx)
			rect.FillColor = theme.ColorForWidget(colorSelection, e)
			rect.Resize(fyne.NewSize(x2-x1, e.lineHeight-2))
			rect.Move(fyne.NewPos(x1, editorPadTop+float32(i-e.scrollRow)*e.lineHeight+1))
			rect.Show()
			rect.Refresh()
			r.objects = append(r.objects, rect)
			selIdx++
		}
	}

	runeText := e.runes
	textIdx := 0
	for i := e.scrollRow; i < len(e.rows) && i < e.scrollRow+e.visibleRows; i++ {
		row := e.rows[i]
		y := editorPadTop + float32(i-e.scrollRow)*e.lineHeight
		segments := r.segments(row)
		x := e.textLeft
		for _, seg := range segments {
			teil := string(runeText[seg.start:seg.end])
			teil = strings.ReplaceAll(teil, "\t", "    ")
			if teil == "" {
				continue
			}
			t := r.textAt(textIdx)
			t.Text = teil
			t.TextSize = e.textSize
			t.TextStyle = fyne.TextStyle{Monospace: true}
			t.Color = theme.ColorForWidget(r.color(seg.kind), e)
			t.Move(fyne.NewPos(x, y))
			t.Refresh()
			r.objects = append(r.objects, t)
			textIdx++
			for j := seg.start; j < seg.end; j++ {
				x += e.advanceFor(runeText[j])
			}
		}
	}

	// Zeilennummern in den linken Innenabstand, rechtsbündig.
	numIdx := 0
	for i := e.scrollRow; i < len(e.rows) && i < e.scrollRow+e.visibleRows; i++ {
		zahl := strconv.Itoa(i + 1)
		t := r.numAt(numIdx)
		t.Text = zahl
		t.TextSize = e.textSize - 1
		t.TextStyle = fyne.TextStyle{Monospace: true}
		t.Color = muted
		breite := fyne.MeasureText(zahl, e.textSize-1, fyne.TextStyle{Monospace: true}).Width
		t.Move(fyne.NewPos(editorGutter-8-breite,
			editorPadTop+float32(i-e.scrollRow)*e.lineHeight+1))
		t.Refresh()
		r.objects = append(r.objects, t)
		numIdx++
	}

	// Cursor.
	if e.focused {
		idx := e.rowAtOffset(e.model.Cursor())
		if idx >= e.scrollRow && idx < e.scrollRow+e.visibleRows {
			row := e.rows[idx]
			x := e.textLeft + e.xAtOffset(row, e.model.Cursor())
			y := editorPadTop + float32(idx-e.scrollRow)*e.lineHeight + 2
			r.caret.FillColor = theme.ColorForWidget(theme.ColorNameForeground, e)
			r.caret.Resize(fyne.NewSize(1.5, e.lineHeight-4))
			r.caret.Move(fyne.NewPos(x, y))
			r.caret.Refresh()
			r.objects = append(r.objects, r.caret)
		}
	}

	// Rollbalken, nur wenn der Text nicht vollständig ins Bild passt.
	if trackY, trackH, thumbY, thumbH, ok := e.scrollbarInfo(); ok {
		surface := theme.ColorForWidget(theme.ColorNameInputBackground, e)
		frame := theme.ColorForWidget(colorFrame, e)
		x := size.Width - editorScrollbar - 4
		r.track.FillColor = mixColors(surface, frame, 0.15)
		r.track.Resize(fyne.NewSize(editorScrollbar, trackH))
		r.track.Move(fyne.NewPos(x, trackY))
		r.track.Refresh()
		r.thumb.FillColor = mixColors(surface, frame, 0.75)
		r.thumb.Resize(fyne.NewSize(editorScrollbar, thumbH))
		r.thumb.Move(fyne.NewPos(x, thumbY))
		r.thumb.Refresh()
		r.objects = append(r.objects, r.track, r.thumb)
	}

	// Nicht mehr benötigte Objekte aus früheren Bildern bleiben im Pool,
	// werden aber nicht mehr gezeichnet, weil objects neu aufgebaut wird.
}

type textSegment struct {
	start int
	end   int
	kind  xmlhighlight.Kind
}

func (r *codeEditorRenderer) segments(row visualRow) []textSegment {
	e := r.editor
	if len(e.spans) == 0 {
		if row.start == row.end {
			return nil
		}
		return []textSegment{{start: row.start, end: row.end, kind: xmlhighlight.Plain}}
	}
	var out []textSegment
	pos := row.start
	for _, span := range e.spans {
		if span.End <= row.start {
			continue
		}
		if span.Start >= row.end {
			break
		}
		if span.Start > pos {
			out = append(out, textSegment{start: pos, end: span.Start, kind: xmlhighlight.Plain})
		}
		start := maxInt(span.Start, row.start)
		end := minInt(span.End, row.end)
		if end > start {
			if len(out) > 0 && out[len(out)-1].kind == span.Kind && out[len(out)-1].end == start {
				out[len(out)-1].end = end
			} else {
				out = append(out, textSegment{start: start, end: end, kind: span.Kind})
			}
			pos = end
		}
	}
	if pos < row.end {
		out = append(out, textSegment{start: pos, end: row.end, kind: xmlhighlight.Plain})
	}
	return out
}

func (r *codeEditorRenderer) rectAt(i int) *canvas.Rectangle {
	for len(r.selPool) <= i {
		rect := canvas.NewRectangle(theme.Color(theme.ColorNameSelection))
		rect.CornerRadius = 2
		r.selPool = append(r.selPool, rect)
	}
	return r.selPool[i]
}

func (r *codeEditorRenderer) textAt(i int) *canvas.Text {
	for len(r.textPool) <= i {
		t := canvas.NewText("", theme.Color(theme.ColorNameForeground))
		r.textPool = append(r.textPool, t)
	}
	return r.textPool[i]
}

func (r *codeEditorRenderer) numAt(i int) *canvas.Text {
	for len(r.numPool) <= i {
		t := canvas.NewText("", theme.Color(colorMuted))
		r.numPool = append(r.numPool, t)
	}
	return r.numPool[i]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	_ fyne.Widget            = (*codeEditor)(nil)
	_ fyne.Focusable         = (*codeEditor)(nil)
	_ fyne.Shortcutable      = (*codeEditor)(nil)
	_ fyne.Scrollable        = (*codeEditor)(nil)
	_ fyne.Draggable         = (*codeEditor)(nil)
	_ fyne.DoubleTappable    = (*codeEditor)(nil)
	_ fyne.SecondaryTappable = (*codeEditor)(nil)
	_ fyne.Tabbable          = (*codeEditor)(nil)
	_ desktop.Keyable        = (*codeEditor)(nil)
	_ desktop.Mouseable      = (*codeEditor)(nil)
	_ desktop.Cursorable     = (*codeEditor)(nil)
)
