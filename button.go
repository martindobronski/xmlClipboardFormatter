package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type buttonStyle int

const (
	buttonFull buttonStyle = iota
	buttonOutlined
	buttonTonal
	buttonLink
	buttonSwitch
)

// actionButton ist eine eigene Schaltfläche, weil Fyne keine umrandete oder
// getönte Stufe kennt und die Java-Vorlage ihre vier Aktionen genau darüber
// unterscheidet. Der Aufbau ist bewusst klein: Fläche, Symbol, Beschriftung
// und ein Fokusring.
type actionButton struct {
	widget.DisableableWidget

	Text     string
	IconName string
	Style    buttonStyle
	OnTapped func()
	Bold     bool

	hovered bool
	pressed bool
	focused bool
}

func newActionButton(text string, style buttonStyle, icon string, onTapped func()) *actionButton {
	b := &actionButton{Text: text, Style: style, IconName: icon, OnTapped: onTapped}
	b.ExtendBaseWidget(b)
	return b
}

func (b *actionButton) CreateRenderer() fyne.WidgetRenderer {
	r := &actionButtonRenderer{
		button: b,
		bg:     canvas.NewRectangle(color.Transparent),
		ring:   canvas.NewRectangle(color.Transparent),
		label:  canvas.NewText("", theme.Color(theme.ColorNameForeground)),
		icon:   canvas.NewImageFromResource(nil),
	}
	r.bg.CornerRadius = 9
	r.bg.StrokeWidth = 1
	r.ring.CornerRadius = 9
	r.ring.StrokeWidth = 2
	r.label.TextStyle = fyne.TextStyle{Bold: b.Bold}
	r.icon.FillMode = canvas.ImageFillContain
	r.objects = []fyne.CanvasObject{r.bg, r.icon, r.label, r.ring}
	r.Refresh()
	return r
}

func (b *actionButton) MinSize() fyne.Size { return b.BaseWidget.MinSize() }

func (b *actionButton) Tapped(*fyne.PointEvent) {
	b.requestFocus()
	if b.Disabled() || b.OnTapped == nil {
		return
	}
	b.OnTapped()
}

func (b *actionButton) TappedSecondary(*fyne.PointEvent) {}

func (b *actionButton) MouseIn(*desktop.MouseEvent) {
	if b.Disabled() {
		return
	}
	b.hovered = true
	canvas.Refresh(b)
}

func (b *actionButton) MouseMoved(*desktop.MouseEvent) {}

func (b *actionButton) MouseOut() {
	b.hovered = false
	canvas.Refresh(b)
}

func (b *actionButton) MouseDown(*desktop.MouseEvent) {
	if b.Disabled() {
		return
	}
	b.requestFocus()
	b.pressed = true
	canvas.Refresh(b)
}

func (b *actionButton) MouseUp(*desktop.MouseEvent) {
	if !b.pressed {
		return
	}
	b.pressed = false
	canvas.Refresh(b)
}

func (b *actionButton) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (b *actionButton) FocusGained() {
	b.focused = true
	canvas.Refresh(b)
}

func (b *actionButton) FocusLost() {
	b.focused = false
	canvas.Refresh(b)
}

func (b *actionButton) TypedRune(rune) {}

func (b *actionButton) TypedKey(key *fyne.KeyEvent) {
	if b.Disabled() {
		return
	}
	switch key.Name {
	case fyne.KeyReturn, fyne.KeyEnter, fyne.KeySpace:
		if b.OnTapped != nil {
			b.OnTapped()
		}
	}
}

func (b *actionButton) requestFocus() {
	if c := fyne.CurrentApp().Driver().CanvasForObject(b); c != nil {
		c.Focus(b)
	}
}

func (b *actionButton) SetText(text string) {
	b.Text = text
	b.Refresh()
}

func (b *actionButton) SetIconName(name string) {
	b.IconName = name
	b.Refresh()
}

type actionButtonRenderer struct {
	button *actionButton
	bg     *canvas.Rectangle
	ring   *canvas.Rectangle
	label  *canvas.Text
	icon   *canvas.Image

	objects []fyne.CanvasObject
}

func (r *actionButtonRenderer) Destroy() {}

func (r *actionButtonRenderer) MinSize() fyne.Size {
	b := r.button
	th := b.Theme()
	textSize := th.Size(sizeButtonText)
	labelText := b.Text
	if b.Style == buttonSwitch {
		// Der Schalter wechselt zwischen "Dunkel" und "Hell". Die Breite
		// richtet sich nach dem längeren Wort, damit er beim Umschalten nicht
		// springt.
		labelText = "Dunkel"
	}
	labelSize := fyne.MeasureText(labelText, textSize, fyne.TextStyle{Bold: b.Bold})
	iconSize := float32(0)
	if b.IconName != "" {
		iconSize = 15 + 7
	}
	padX, padY := float32(14), float32(8)
	if b.Style == buttonLink {
		padX, padY = 6, 6
	}
	height := labelSize.Height
	if iconSize > 0 {
		height = max32(height, 15)
	}
	return fyne.NewSize(labelSize.Width+iconSize+padX*2, height+padY*2)
}

func (r *actionButtonRenderer) Layout(size fyne.Size) {
	b := r.button
	th := b.Theme()
	textSize := th.Size(sizeButtonText)
	labelSize := fyne.MeasureText(b.Text, textSize, fyne.TextStyle{Bold: b.Bold})

	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))
	r.ring.Resize(fyne.NewSize(size.Width-3, size.Height-3))
	r.ring.Move(fyne.NewPos(1.5, 1.5))

	padX := float32(14)
	if b.Style == buttonLink {
		padX = 6
	}
	x := padX
	if b.IconName != "" {
		iconSize := float32(15)
		r.icon.Resize(fyne.NewSize(iconSize, iconSize))
		r.icon.Move(fyne.NewPos(x, (size.Height-iconSize)/2))
		x += iconSize + 7
	}
	r.label.TextSize = textSize
	r.label.Resize(fyne.NewSize(max32(0, size.Width-x-padX), labelSize.Height))
	r.label.Move(fyne.NewPos(x, (size.Height-labelSize.Height)/2))
}

func (r *actionButtonRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *actionButtonRenderer) Refresh() {
	b := r.button
	disabled := b.Disabled()
	hover := b.hovered || b.pressed
	focused := b.focused

	background := theme.ColorForWidget(theme.ColorNameBackground, b)
	surface := theme.ColorForWidget(theme.ColorNameInputBackground, b)
	accent := theme.ColorForWidget(theme.ColorNamePrimary, b)
	accentHover := theme.ColorForWidget(colorAccentHover, b)
	accentText := theme.ColorForWidget(colorAccentText, b)
	onAccent := theme.ColorForWidget(theme.ColorNameForegroundOnPrimary, b)
	tonal := theme.ColorForWidget(colorTonal, b)
	tonalHover := theme.ColorForWidget(colorTonalHover, b)
	frame := theme.ColorForWidget(colorFrame, b)
	muted := theme.ColorForWidget(colorMuted, b)
	onDisabled := theme.ColorForWidget(colorOnDisabled, b)
	foreground := theme.ColorForWidget(theme.ColorNameForeground, b)

	var fill color.Color = color.Transparent
	var border color.Color = color.Transparent
	var labelColor color.Color = foreground
	stroke := float32(0)

	switch b.Style {
	case buttonFull:
		fill, border, labelColor, stroke = accent, frame, onAccent, 1
		if hover {
			fill = accentHover
		}
	case buttonOutlined:
		fill, border, labelColor, stroke = background, accentText, accentText, 1
	case buttonTonal:
		fill, border, labelColor, stroke = tonal, frame, accentText, 1
		if hover {
			fill = tonalHover
		}
	case buttonLink:
		fill, labelColor = color.Transparent, muted
		if hover {
			labelColor = accentText
		}
	case buttonSwitch:
		fill, border, labelColor, stroke = surface, frame, foreground, 1
		if hover {
			fill = mixColors(surface, foreground, 0.08)
		}
	}

	if disabled {
		if b.Style == buttonLink {
			labelColor = muted
		} else {
			fill = mixColors(fill, background, 0.75)
			border = mixColors(border, background, 0.75)
			labelColor = onDisabled
		}
	}

	r.bg.FillColor = fill
	r.bg.StrokeColor = border
	r.bg.StrokeWidth = stroke
	r.bg.Refresh()

	r.label.Text = b.Text
	r.label.Color = labelColor
	r.label.Refresh()

	if b.IconName == "" {
		r.icon.Hide()
	} else {
		res := iconResource(b.IconName, labelColor)
		if r.icon.Resource == nil || r.icon.Resource.Name() != res.Name() {
			r.icon.Resource = res
			r.icon.Refresh()
		}
		r.icon.Show()
	}

	ringColor := accent
	if b.Style == buttonFull && !disabled {
		ringColor = onAccent
	}
	if focused && !disabled {
		r.ring.StrokeColor = ringColor
		r.ring.Show()
	} else {
		r.ring.Hide()
	}
	r.ring.Refresh()

	r.Layout(b.Size())
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

var (
	_ fyne.Widget        = (*actionButton)(nil)
	_ fyne.Tappable      = (*actionButton)(nil)
	_ fyne.Focusable     = (*actionButton)(nil)
	_ fyne.Disableable   = (*actionButton)(nil)
	_ desktop.Hoverable  = (*actionButton)(nil)
	_ desktop.Mouseable  = (*actionButton)(nil)
	_ desktop.Cursorable = (*actionButton)(nil)
)
