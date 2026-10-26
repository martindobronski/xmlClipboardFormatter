package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Farben, die Fyne nicht kennt. Sie werden über die eigene Theme-Implementierung
// nachgeschlagen und passen sich dadurch beim Umschalten automatisch an.
const (
	colorFrame       fyne.ThemeColorName = "xmlFrame"
	colorMuted       fyne.ThemeColorName = "xmlMuted"
	colorAccentHover fyne.ThemeColorName = "xmlAccentHover"
	colorAccentText  fyne.ThemeColorName = "xmlAccentText"
	colorTonal       fyne.ThemeColorName = "xmlTonal"
	colorTonalHover  fyne.ThemeColorName = "xmlTonalHover"
	colorOnDisabled  fyne.ThemeColorName = "xmlOnDisabled"
	colorGutter      fyne.ThemeColorName = "xmlGutter"
	colorXMLKeyword  fyne.ThemeColorName = "xmlKeyword"
	colorXMLName     fyne.ThemeColorName = "xmlName"
	colorXMLString   fyne.ThemeColorName = "xmlString"
	colorXMLNumber   fyne.ThemeColorName = "xmlNumber"
	colorXMLComment  fyne.ThemeColorName = "xmlComment"
	colorXMLOperator fyne.ThemeColorName = "xmlOperator"
	colorSelection   fyne.ThemeColorName = "xmlSelection"
)

const (
	sizeButtonText  fyne.ThemeSizeName = "xmlButtonText"
	sizeStatusText  fyne.ThemeSizeName = "xmlStatusText"
	sizeVersionText fyne.ThemeSizeName = "xmlVersionText"
	sizeEditorText  fyne.ThemeSizeName = "xmlEditorText"
)

// palette entspricht dem Farbsatz aus Theme.java des JSON-Formatters. Beide
// Paletten belegen dieselben Felder; eine vergessene Farbe fällt deshalb schon
// beim Schreiben des Codes auf und nicht erst nach dem Umschalten.
type palette struct {
	dark bool

	background color.Color
	surface    color.Color
	border     color.Color
	frame      color.Color
	text       color.Color
	muted      color.Color

	accent      color.Color
	accentHover color.Color
	onAccent    color.Color
	accentText  color.Color
	tonal       color.Color
	tonalHover  color.Color
	onDisabled  color.Color

	success color.Color
	warning color.Color
	error   color.Color

	keyword  color.Color
	name     color.Color
	str      color.Color
	number   color.Color
	comment  color.Color
	operator color.Color

	selection color.Color
}

func c(rgb uint32) color.NRGBA {
	return color.NRGBA{
		R: uint8(rgb >> 16),
		G: uint8(rgb >> 8),
		B: uint8(rgb),
		A: 0xff,
	}
}

var (
	paletteDark = palette{
		dark:        true,
		background:  c(0x1E2024),
		surface:     c(0x262A2F),
		border:      c(0x383C42),
		frame:       c(0x767C85),
		text:        c(0xE8EAED),
		muted:       c(0x9AA4B0),
		accent:      c(0x1D4ED8),
		accentHover: c(0x1A44B8),
		onAccent:    c(0xFFFFFF),
		accentText:  c(0x93C5FD),
		tonal:       c(0x1E3050),
		tonalHover:  c(0x27395C),
		onDisabled:  c(0x9AA4B0),
		success:     c(0x4ADE80),
		warning:     c(0xFBBF24),
		error:       c(0xF87171),
		keyword:     c(0xBB9AF7),
		name:        c(0x7DCFFF),
		str:         c(0x9ECE6A),
		number:      c(0xE0AF68),
		comment:     c(0x8B949E),
		operator:    c(0xC6C8D1),
		selection:   color.NRGBA{R: 0x2D, G: 0x4F, B: 0x8A, A: 0xB0},
	}
	paletteLight = palette{
		dark:        false,
		background:  c(0xF4F5F7),
		surface:     c(0xFFFFFF),
		border:      c(0xD5D9DE),
		frame:       c(0x6B7280),
		text:        c(0x1F2328),
		muted:       c(0x5B6470),
		accent:      c(0x1D4ED8),
		accentHover: c(0x1A44B8),
		onAccent:    c(0xFFFFFF),
		accentText:  c(0x1D4ED8),
		tonal:       c(0xE8EFFC),
		tonalHover:  c(0xD8E4FB),
		onDisabled:  c(0x464E58),
		success:     c(0x15803D),
		warning:     c(0xB45309),
		error:       c(0xB91C1C),
		keyword:     c(0x8250DF),
		name:        c(0x0F766E),
		str:         c(0x0F7B3C),
		number:      c(0x9A3412),
		comment:     c(0x6B7280),
		operator:    c(0x374151),
		selection:   color.NRGBA{R: 0xB9, G: 0xD2, B: 0xFA, A: 0xC8},
	}
)

// gesperrt mischt eine Farbe in den Hintergrund. Gesperrte Schaltflächen
// behalten so ihren Farbton, statt neutral grau zu wirken.
func (p palette) gesperrt(farbe color.Color) color.Color {
	anteil := 0.70
	if p.dark {
		anteil = 0.78
	}
	r, g, b, _ := farbe.RGBA()
	hr, hg, hb, _ := p.background.RGBA()
	return color.NRGBA{
		R: uint8(float64(r>>8)*(1-anteil) + float64(hr>>8)*anteil),
		G: uint8(float64(g>>8)*(1-anteil) + float64(hg>>8)*anteil),
		B: uint8(float64(b>>8)*(1-anteil) + float64(hb>>8)*anteil),
		A: 0xff,
	}
}

// appTheme ist die Fyne-Theme-Implementierung der Anwendung. Sie liefert die
// Palette unabhängig von der Systemeinstellung; die App startet hell wie die
// Java-Vorlage und wechselt nur auf ausdrücklichen Wunsch.
type appTheme struct {
	palette palette
	base    fyne.Theme
}

func newAppTheme(dark bool) *appTheme {
	p := paletteLight
	if dark {
		p = paletteDark
	}
	return &appTheme{palette: p, base: theme.DefaultTheme()}
}

func (t *appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	p := t.palette
	switch name {
	case theme.ColorNameBackground:
		return p.background
	case theme.ColorNameButton:
		return p.surface
	case theme.ColorNameDisabled:
		return p.onDisabled
	case theme.ColorNameDisabledButton:
		return p.gesperrt(p.surface)
	case theme.ColorNameError:
		return p.error
	case theme.ColorNameFocus:
		return p.accent
	case theme.ColorNameForeground:
		return p.text
	case theme.ColorNameForegroundOnError:
		return p.background
	case theme.ColorNameForegroundOnPrimary:
		return p.onAccent
	case theme.ColorNameForegroundOnSuccess:
		return p.background
	case theme.ColorNameForegroundOnWarning:
		return p.background
	case theme.ColorNameHeaderBackground:
		return p.background
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x7F, G: 0x8A, B: 0x9A, A: 0x20}
	case theme.ColorNameHyperlink:
		return p.accentText
	case theme.ColorNameInputBackground:
		return p.surface
	case theme.ColorNameInputBorder:
		return p.border
	case theme.ColorNameMenuBackground:
		return p.surface
	case theme.ColorNameOverlayBackground:
		return p.surface
	case theme.ColorNamePlaceHolder:
		return p.muted
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0x7F, G: 0x8A, B: 0x9A, A: 0x40}
	case theme.ColorNamePrimary:
		return p.accent
	case theme.ColorNameScrollBar:
		return p.frame
	case theme.ColorNameScrollBarBackground:
		return color.Transparent
	case theme.ColorNameSelection:
		return p.selection
	case theme.ColorNameSeparator:
		return p.border
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x28}
	case theme.ColorNameSuccess:
		return p.success
	case theme.ColorNameWarning:
		return p.warning

	case colorFrame:
		return p.frame
	case colorMuted:
		return p.muted
	case colorAccentHover:
		return p.accentHover
	case colorAccentText:
		return p.accentText
	case colorTonal:
		return p.tonal
	case colorTonalHover:
		return p.tonalHover
	case colorOnDisabled:
		return p.onDisabled
	case colorGutter:
		return p.muted
	case colorXMLKeyword:
		return p.keyword
	case colorXMLName:
		return p.name
	case colorXMLString:
		return p.str
	case colorXMLNumber:
		return p.number
	case colorXMLComment:
		return p.comment
	case colorXMLOperator:
		return p.operator
	case colorSelection:
		return p.selection
	}
	return t.base.Color(name, variant)
}

func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *appTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *appTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case sizeButtonText:
		return 12
	case sizeStatusText:
		return 11
	case sizeVersionText:
		return 10
	case sizeEditorText:
		return 13
	}
	return t.base.Size(name)
}

// mixColors mischt zwei Farben; anteil ist der Anteil der zweiten Farbe.
func mixColors(a, b color.Color, anteil float64) color.Color {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return color.NRGBA{
		R: uint8(float64(ar>>8)*(1-anteil) + float64(br>>8)*anteil),
		G: uint8(float64(ag>>8)*(1-anteil) + float64(bg>>8)*anteil),
		B: uint8(float64(ab>>8)*(1-anteil) + float64(bb>>8)*anteil),
		A: uint8(float64(aa>>8)*(1-anteil) + float64(ba>>8)*anteil),
	}
}

// hexColor liefert die Farbe als #rrggbb, wie sie in einem SVG steht.
func hexColor(farbe color.Color) string {
	r, g, b, _ := farbe.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// iconResource erzeugt ein einfarbiges Symbol in der gewünschten Farbe. Der
// Name enthält die Farbe, damit Fyne die gerenderten Varianten unterscheiden
// kann.
var iconCache = map[string]fyne.Resource{}

func iconResource(name string, farbe color.Color) fyne.Resource {
	hex := hexColor(farbe)
	key := name + "_" + strings.TrimPrefix(hex, "#")
	if res, ok := iconCache[key]; ok {
		return res
	}
	svg, ok := iconSVGs[name]
	if !ok {
		svg = iconSVGs["document"]
	}
	content := strings.ReplaceAll(svg, "#000000", hex)
	res := fyne.NewStaticResource(key, []byte(content))
	iconCache[key] = res
	return res
}
