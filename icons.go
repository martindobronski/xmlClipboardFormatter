package main

// Die Symbole sind bewusst gefüllte Pfade und keine Strichzeichnungen: Fyne
// färbt SVGs über ihre Füllung ein. Jedes Symbol wird zur Laufzeit in der
// Farbe der jeweiligen Schaltfläche erzeugt (siehe iconResource).
var iconSVGs = map[string]string{
	"clipboard": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
<path fill="#000000" fill-rule="evenodd" d="M5 3h14a2 2 0 0 1 2 2v16a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Zm1 3v14h12V6H6Z"/>
<path fill="#000000" d="M8 1h8v4H8zM8 9h8v2H8zm0 4h6v2H8z"/>
</svg>`,
	"wand": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
<path fill="#000000" d="M18.4 2.1l3.5 3.5-13 13-3.5-3.5z"/>
<path fill="#000000" d="M6 0l1 2.4 2.4 1L7 4.4 6 6.8 5 4.4 2.6 3.4 5 2.4z"/>
<path fill="#000000" d="M20.4 12l.7 1.7 1.7.7-1.7.7-.7 1.7-.7-1.7-1.7-.7 1.7-.7z"/>
</svg>`,
	"compact": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
<path fill="#000000" d="M2 5h7V3l4 4-4 4V9H2z"/>
<path fill="#000000" d="M22 15h-7v-2l-4 4 4 4v-2h7z"/>
</svg>`,
	"write": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
<path fill="#000000" d="M10 1h4v10h4l-6 7-6-7h4z"/>
<path fill="#000000" d="M3 20h18v2H3z"/>
</svg>`,
	"moon": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
<path fill="#000000" d="M22 15.1A10 10 0 1 1 8.9 2a8 8 0 0 0 13.1 13.1z"/>
</svg>`,
	"sun": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
<circle fill="#000000" cx="12" cy="12" r="4.2"/>
<path fill="#000000" d="M11 1h2v4h-2zM11 19h2v4h-2zM1 11h4v2H1zM19 11h4v2h-4zM4.2 5.6l1.4-1.4 2.8 2.8-1.4 1.4zM15.6 17l1.4-1.4 2.8 2.8-1.4 1.4zM5.6 19.8l-1.4-1.4 2.8-2.8 1.4 1.4zM17 6.9l-1.4-1.4 2.8-2.8 1.4 1.4z"/>
</svg>`,
	"document": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
<path fill="#000000" fill-rule="evenodd" d="M6 2h8l5 5v15H6V2Zm2 2v16h9V8h-4V4H8Z"/>
<path fill="#000000" d="M9 11h6v1.6H9zm0 3h6v1.6H9zm0 3h4v1.6H9z"/>
</svg>`,
}

// appIcon ist das Fenster-/Dock-Symbol. Es wird nicht eingefärbt und behält
// deshalb seine beiden Farben.
const appIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
<rect width="64" height="64" rx="14" fill="#1E2024"/>
<path fill="#93C5FD" d="M22 18l-9 14 9 14h6L19 32l9-14zM42 18l9 14-9 14h-6l9-14-9-14z"/>
<path fill="#4ADE80" d="M35 16l-7 32h5l7-32z"/>
</svg>`
