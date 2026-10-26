package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// version und versionDate werden beim Bauen über ldflags gesetzt:
//
//	go build -ldflags "-X main.version=1.2.3 -X main.versionDate=26.10.2026"
//
// Ohne Angabe zeigt die Fusszeile die Version 1.0.0 und das heutige Datum.
// Das Datum ist dann das des Starts, nicht das des Builds; die ldflags sind
// der Weg für ein echtes Build-Datum.
var (
	version     = "1.0.0"
	versionDate = ""
)

func main() {
	a := app.NewWithID("de.signaliduna.xmlclipboardformatter")
	a.Settings().SetTheme(newAppTheme(false))

	icon := fyne.NewStaticResource("app-icon.svg", []byte(appIconSVG))
	a.SetIcon(icon)

	w := a.NewWindow("XML Clipboard Formatter")
	w.SetIcon(icon)

	ui := newFormatterUI(a, w)
	w.SetPadded(false)
	w.SetContent(ui.root)
	w.Canvas().Focus(ui.editor)
	w.Resize(fyne.NewSize(980, 720))
	w.CenterOnScreen()
	w.ShowAndRun()
}
