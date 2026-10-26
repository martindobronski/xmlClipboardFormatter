package main

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
)

// neueTestUI baut dieselbe Oberfläche wie main, nur ohne echtes Fenster.
// SetPadded(false) gehört dazu: sonst rechnet der Test mit anderen Rändern
// als die Anwendung und prüft eine Oberfläche, die niemand sieht.
func neueTestUI(t *testing.T) *formatterUI {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	a.Settings().SetTheme(newAppTheme(false))

	w := test.NewWindow(nil)
	t.Cleanup(w.Close)
	w.SetPadded(false)

	ui := newFormatterUI(a, w)
	w.SetContent(ui.root)
	w.Resize(fyne.NewSize(980, 720))
	return ui
}

// warteAufStatus wartet, bis die asynchrone Aktion ihre Meldung gesetzt hat.
// Feste Wartezeiten wären auf einem ausgelasteten Rechner irgendwann zu kurz.
func warteAufStatus(t *testing.T, ui *formatterUI, want string) {
	t.Helper()
	frist := time.Now().Add(2 * time.Second)
	for time.Now().Before(frist) {
		if ui.statusText.Text == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("Status = %q, erwartet %q", ui.statusText.Text, want)
}

func TestFormatButton(t *testing.T) {
	ui := neueTestUI(t)
	ui.editor.SetText("<root><a x='1'>text</a><b/></root>")

	test.Tap(ui.formatButton)
	warteAufStatus(t, ui, "✓ XML lesbar formatiert.")

	want := "<root>\n    <a x='1'>text</a>\n    <b />\n</root>\n"
	if got := ui.editor.Text(); got != want {
		t.Fatalf("formatiert:\n%q\nerwartet:\n%q", got, want)
	}
}

func TestPackButton(t *testing.T) {
	ui := neueTestUI(t)
	ui.editor.SetText("<root>\n    <a x='1'>text</a>\n    <b />\n</root>")

	test.Tap(ui.packButton)
	warteAufStatus(t, ui, "✓ XML auf eine Zeile gepackt.")

	want := "<root><a x='1'>text</a><b /></root>\n"
	if got := ui.editor.Text(); got != want {
		t.Fatalf("gepackt:\n%q\nerwartet:\n%q", got, want)
	}
}

func TestSchreibenBehaeltGeschuetztesLeerzeichen(t *testing.T) {
	ui := neueTestUI(t)
	eingabe := "<root/>\u00A0"
	ui.editor.SetText(eingabe)

	test.Tap(ui.writeButton)
	if got := ui.app.Clipboard().Content(); got != eingabe {
		t.Fatalf("Zwischenablage:\n%q\nerwartet:\n%q", got, eingabe)
	}
}

func TestFehlerLaesstTextStehen(t *testing.T) {
	ui := neueTestUI(t)
	eingabe := "<root><a></root>"
	ui.editor.SetText(eingabe)

	test.Tap(ui.formatButton)
	frist := time.Now().Add(2 * time.Second)
	for time.Now().Before(frist) && !strings.HasPrefix(ui.statusText.Text, "❌ ") {
		time.Sleep(10 * time.Millisecond)
	}

	if !strings.Contains(ui.statusText.Text, "Erwartet </a>, gefunden </root>.") {
		t.Fatalf("Status = %q", ui.statusText.Text)
	}
	if got := ui.editor.Text(); got != eingabe {
		t.Fatalf("Text wurde verändert: %q", got)
	}
}

func TestAktionenSindOhneXMLGesperrt(t *testing.T) {
	ui := neueTestUI(t)
	ui.editor.SetText("Einkaufsliste")

	for _, b := range []*actionButton{ui.formatButton, ui.packButton, ui.writeButton} {
		if !b.Disabled() {
			t.Fatalf("%q ist ohne XML nicht gesperrt", b.Text)
		}
	}
	if ui.readButton.Disabled() {
		t.Fatal("Clipboard einlesen ist gesperrt")
	}
}

func TestRedoShortcutMitShift(t *testing.T) {
	ui := neueTestUI(t)
	e := ui.editor
	e.SetText("")
	e.model.Insert("a")
	e.afterEdit()

	e.TypedShortcut(&fyne.ShortcutUndo{})
	if got := e.Text(); got != "" {
		t.Fatalf("nach Rückgängig: %q", got)
	}

	// Fyne reicht Cmd+Shift+Z beziehungsweise Strg+Shift+Z als eigenen
	// Shortcut durch; der Editor muss ihn selbst auf Wiederholen abbilden.
	e.TypedShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyZ,
		Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift,
	})
	if got := e.Text(); got != "a" {
		t.Fatalf("nach Wiederholen: %q", got)
	}
}

func TestThemeSwitch(t *testing.T) {
	ui := neueTestUI(t)

	ui.toggleTheme()
	if !ui.dark {
		t.Fatal("Theme wurde nicht dunkel")
	}
	if got := ui.themeButton.Text; got != "Hell" {
		t.Fatalf("Schaltertext: %q", got)
	}

	ui.toggleTheme()
	if ui.dark {
		t.Fatal("Theme wurde nicht wieder hell")
	}
	if got := ui.themeButton.Text; got != "Dunkel" {
		t.Fatalf("Schaltertext: %q", got)
	}
}
