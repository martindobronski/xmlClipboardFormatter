package main

import (
	"fmt"
	"image/color"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	"xmlclipboardformatter/internal/xmlfmt"
)

const (
	formMarginTop    = 16
	formMarginBottom = 14
	formMarginLeft   = 18
	formMarginRight  = 18
	formGap          = 12
	actionGap        = 8
	footerGap        = 10
	toastDuration    = 4 * time.Second
)

type statusKind int

const (
	statusNeutral statusKind = iota
	statusSuccess
	statusWarning
	statusError
)

// formatterUI hält die Fensterbestandteile und den Ablauf zwischen
// Zwischenablage, Editor und Formatierer. Die Formatierung selbst liegt in
// internal/xmlfmt und weiß nichts von Fyne.
type formatterUI struct {
	app fyne.App
	win fyne.Window

	root fyne.CanvasObject

	editor     *codeEditor
	cardBG     *canvas.Rectangle
	cardBorder *canvas.Rectangle

	header  *fyne.Container
	footer  *fyne.Container
	info    *fyne.Container
	actions *fyne.Container

	themeButton  *actionButton
	readButton   *actionButton
	formatButton *actionButton
	packButton   *actionButton
	writeButton  *actionButton
	exitButton   *actionButton

	statusText  *canvas.Text
	countText   *canvas.Text
	versionText *canvas.Text

	dark     bool
	busy     bool
	closed   atomic.Bool
	statusID int
	status   statusKind
}

func newFormatterUI(a fyne.App, w fyne.Window) *formatterUI {
	u := &formatterUI{app: a, win: w}

	u.editor = newCodeEditor("")
	u.editor.OnChanged = u.onTextChanged

	u.themeButton = newActionButton("Dunkel", buttonSwitch, "moon", u.toggleTheme)
	u.readButton = newActionButton("Clipboard einlesen", buttonOutlined, "clipboard", u.readClipboard)
	u.formatButton = newActionButton("XML Formatieren", buttonFull, "wand", func() { u.format(false) })
	u.packButton = newActionButton("XML Packen", buttonOutlined, "compact", func() { u.format(true) })
	u.writeButton = newActionButton("Ins Clipboard schreiben", buttonTonal, "write", u.writeClipboard)
	u.exitButton = newActionButton("Beenden", buttonLink, "", func() { a.Quit() })

	u.statusText = canvas.NewText("", theme.Color(theme.ColorNameForeground))
	u.statusText.TextSize = a.Settings().Theme().Size(sizeStatusText)
	u.statusText.Hide()

	u.countText = canvas.NewText("0 Zeilen, 0 Zeichen", theme.Color(colorMuted))
	u.countText.TextSize = a.Settings().Theme().Size(sizeStatusText)

	u.versionText = canvas.NewText(versionLine(), theme.Color(colorMuted))
	u.versionText.TextSize = a.Settings().Theme().Size(sizeVersionText)

	u.cardBG = canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	u.cardBG.CornerRadius = 14

	u.cardBorder = canvas.NewRectangle(color.Transparent)
	u.cardBorder.CornerRadius = 14
	u.cardBorder.StrokeWidth = 1

	card := container.NewStack(u.cardBG, u.cardBorder, u.editor)

	u.header = container.NewHBox(layout.NewSpacer(), u.themeButton)
	u.actions = container.New(layout.NewCustomPaddedHBoxLayout(actionGap),
		u.readButton, u.formatButton, u.packButton, u.writeButton)
	actionRow := container.NewBorder(nil, nil, u.actions, u.exitButton)

	statusArea := container.NewStack(u.statusText, u.countText)
	u.info = container.NewBorder(nil, nil, statusArea, u.versionText)

	u.footer = container.New(layout.NewCustomPaddedVBoxLayout(footerGap),
		actionRow, u.info)

	form := container.New(&formLayout{gap: formGap}, u.header, card, u.footer)
	u.root = container.New(
		layout.NewCustomPaddedLayout(formMarginTop, formMarginBottom, formMarginLeft, formMarginRight),
		form)

	w.SetOnClosed(func() { u.closed.Store(true) })

	u.applyTheme()
	u.updateCount()
	u.refreshActions()

	return u
}

func versionLine() string {
	datum := versionDate
	if datum == "" {
		datum = time.Now().Format("02.01.2006")
	}
	return "Version " + version + " vom " + datum
}

// formLayout legt Kopfzeile, Editorfläche und Fusszeile untereinander. Die
// beiden Abstände gehören hierher und nicht in die Kinder, weil nur so die
// Editorfläche den Rest des Fensters wirklich ausfüllt.
type formLayout struct {
	gap float32
}

func (l *formLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}
	header, card, footer := objects[0], objects[1], objects[2]
	headerH := header.MinSize().Height
	footerH := footer.MinSize().Height

	header.Move(fyne.NewPos(0, 0))
	header.Resize(fyne.NewSize(size.Width, headerH))

	y := headerH + l.gap
	cardH := size.Height - y - footerH - l.gap
	if cardH < 0 {
		cardH = 0
	}
	card.Move(fyne.NewPos(0, y))
	card.Resize(fyne.NewSize(size.Width, cardH))

	footer.Move(fyne.NewPos(0, size.Height-footerH))
	footer.Resize(fyne.NewSize(size.Width, footerH))
}

func (l *formLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 3 {
		return fyne.NewSize(0, 0)
	}
	header, card, footer := objects[0], objects[1], objects[2]
	width := header.MinSize().Width
	width = max32(width, card.MinSize().Width)
	width = max32(width, footer.MinSize().Width)
	height := header.MinSize().Height + l.gap + card.MinSize().Height +
		l.gap + footer.MinSize().Height
	return fyne.NewSize(width, height)
}

func (u *formatterUI) onTextChanged() {
	u.updateCount()
	u.refreshActions()
}

func (u *formatterUI) updateCount() {
	text := u.editor.Text()
	if text == "" {
		u.countText.Text = "0 Zeilen, 0 Zeichen"
	} else {
		zeilen := u.editor.DisplayLineCount()
		wort := "Zeilen"
		if zeilen == 1 {
			wort = "Zeile"
		}
		u.countText.Text = fmt.Sprintf("%d %s, %d Zeichen", zeilen, wort, u.editor.CharCount())
	}
	u.countText.Refresh()
	u.info.Refresh()
}

func (u *formatterUI) refreshActions() {
	valid := !u.busy && xmlfmt.LooksLikeXML(u.editor.Text())

	u.readButton.Enable()
	if u.busy {
		u.readButton.Disable()
	}
	for _, b := range []*actionButton{u.formatButton, u.packButton, u.writeButton} {
		if valid {
			b.Enable()
		} else {
			b.Disable()
		}
	}

	if u.busy {
		return
	}
	if valid || xmlfmt.TrimSpace(u.editor.Text()) == "" {
		u.setStatus("", statusNeutral)
	} else {
		u.setStatus("⚠ Der Text sieht nicht nach XML aus. Formatieren, Packen und Schreiben sind blockiert.",
			statusWarning)
	}
}

func (u *formatterUI) setBusy(busy bool) {
	u.busy = busy
	u.refreshActions()
}

func (u *formatterUI) setStatus(text string, kind statusKind) {
	u.statusID++
	u.status = kind
	u.statusText.Text = text
	u.statusText.Color = u.statusColor()
	u.statusText.Refresh()

	if text == "" {
		u.statusText.Hide()
		u.countText.Show()
	} else {
		u.statusText.Show()
		u.countText.Hide()
	}
	u.info.Refresh()

	if kind == statusSuccess && text != "" && !u.busy {
		id := u.statusID
		time.AfterFunc(toastDuration, func() {
			if u.closed.Load() {
				return
			}
			fyne.Do(func() {
				if u.statusID != id || u.busy {
					return
				}
				u.setStatus("", statusNeutral)
			})
		})
	}
}

func (u *formatterUI) statusColor() color.Color {
	switch u.status {
	case statusSuccess:
		return u.themeColor(theme.ColorNameSuccess)
	case statusWarning:
		return u.themeColor(theme.ColorNameWarning)
	case statusError:
		return u.themeColor(theme.ColorNameError)
	}
	return u.themeColor(colorMuted)
}

func (u *formatterUI) readClipboard() {
	u.setBusy(true)
	u.setStatus("Lese Zwischenablage ...", statusNeutral)

	text := fyne.CurrentApp().Clipboard().Content()
	if text == "" {
		u.setBusy(false)
		u.setStatus("⚠ Die Zwischenablage ist leer.", statusWarning)
		return
	}
	if xmlfmt.TrimSpace(text) == "" {
		u.setBusy(false)
		u.setStatus("⚠ Die Zwischenablage enthält keinen lesbaren Text.", statusWarning)
		return
	}

	u.editor.SetText(text)
	u.setBusy(false)
	if xmlfmt.LooksLikeXML(text) {
		u.setStatus("✓ XML geladen. Formatieren, Packen und Schreiben sind freigeschaltet.", statusSuccess)
	} else {
		u.setStatus("⚠ Text geladen, sieht aber nicht nach XML aus. Aktionen blockiert.", statusWarning)
	}
}

func (u *formatterUI) writeClipboard() {
	text := xmlfmt.TrimSpace(u.editor.Text())
	if text == "" {
		u.setStatus("⚠ Das Textfeld ist leer. Nichts zu kopieren.", statusWarning)
		return
	}

	u.setBusy(true)
	u.setStatus("Schreibe in die Zwischenablage ...", statusNeutral)
	fyne.CurrentApp().Clipboard().SetContent(text)
	u.setBusy(false)
	u.setStatus("✓ XML erfolgreich in die Zwischenablage kopiert.", statusSuccess)
}

func (u *formatterUI) format(compact bool) {
	original := u.editor.Text()
	if xmlfmt.TrimSpace(original) == "" {
		return
	}
	cursor := u.editor.CursorOffset()

	u.setBusy(true)
	if compact {
		u.setStatus("Packe ...", statusNeutral)
	} else {
		u.setStatus("Formatiere ...", statusNeutral)
	}

	go func() {
		var ausgabe string
		var err error
		if compact {
			ausgabe, err = xmlfmt.Compact(original)
		} else {
			ausgabe, err = xmlfmt.Pretty(original)
		}

		fyne.Do(func() {
			if u.closed.Load() {
				return
			}
			u.setBusy(false)
			if err != nil {
				// Die Eingabe bleibt stehen. Ein geratener Ersatztext wäre
				// schlimmer als eine Fehlermeldung, weil niemand hinterher
				// sagen könnte, was sich geändert hat.
				u.setStatus("❌ "+err.Error()+" Der Text ist unverändert geblieben.", statusError)
				return
			}

			veraendert := ausgabe != original
			u.editor.SetTextAt(ausgabe, cursor)
			u.updateCount()
			u.refreshActions()
			u.setStatus(erfolgsMeldung(compact, veraendert), statusSuccess)
		})
	}()
}

func erfolgsMeldung(compact, veraendert bool) string {
	if compact {
		if veraendert {
			return "✓ XML auf eine Zeile gepackt."
		}
		return "✓ War schon gepackt."
	}
	if veraendert {
		return "✓ XML lesbar formatiert."
	}
	return "✓ War schon lesbar formatiert."
}

func (u *formatterUI) toggleTheme() {
	u.dark = !u.dark
	u.app.Settings().SetTheme(newAppTheme(u.dark))
	u.applyTheme()
}

func (u *formatterUI) themeColor(name fyne.ThemeColorName) color.Color {
	variant := theme.VariantLight
	if u.dark {
		variant = theme.VariantDark
	}
	return u.app.Settings().Theme().Color(name, variant)
}

func (u *formatterUI) themeSize(name fyne.ThemeSizeName) float32 {
	return u.app.Settings().Theme().Size(name)
}

func (u *formatterUI) applyTheme() {
	surface := u.themeColor(theme.ColorNameInputBackground)
	frame := u.themeColor(colorFrame)

	u.cardBG.FillColor = surface
	u.cardBorder.StrokeColor = mixColors(surface, frame, 0.35)
	u.cardBG.Refresh()
	u.cardBorder.Refresh()

	u.countText.TextSize = u.themeSize(sizeStatusText)
	u.countText.Color = u.themeColor(colorMuted)
	u.countText.Refresh()

	u.statusText.TextSize = u.themeSize(sizeStatusText)
	u.statusText.Color = u.statusColor()
	u.statusText.Refresh()

	u.versionText.TextSize = u.themeSize(sizeVersionText)
	u.versionText.Color = u.themeColor(colorMuted)
	u.versionText.Refresh()

	if u.dark {
		u.themeButton.SetText("Hell")
		u.themeButton.SetIconName("sun")
	} else {
		u.themeButton.SetText("Dunkel")
		u.themeButton.SetIconName("moon")
	}

	for _, b := range []*actionButton{
		u.themeButton, u.readButton, u.formatButton, u.packButton,
		u.writeButton, u.exitButton,
	} {
		b.Refresh()
	}
	u.editor.Refresh()
	u.header.Refresh()
	u.footer.Refresh()
}

var (
	_ fyne.Layout = (*formLayout)(nil)
	_ fyne.Widget = (*actionButton)(nil)
	_ fyne.Widget = (*codeEditor)(nil)
)
