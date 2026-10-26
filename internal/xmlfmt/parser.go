package xmlfmt

import (
	"strings"
)

type nodeKind int

const (
	nodeElement nodeKind = iota
	nodeText
	nodeComment
	nodeCDATA
	nodePI
	nodeDoctype
)

// maxDepth verhindert, dass ein Dokument den Aufrufstapel sprengt. Die
// Einrückung wächst ohnehin quadratisch mit der Tiefe; oberhalb dieser Grenze
// ist die Ausgabe für einen Menschen nicht mehr lesbar und für den Rechner
// teuer.
const maxDepth = 10_000

type node struct {
	kind     nodeKind
	tok      *token
	end      *token
	children []*node
}

func (n *node) start() int { return n.tok.start }

func (n *node) stop() int {
	if n.end != nil {
		return n.end.end
	}
	return n.tok.end
}

type document struct {
	nodes   []*node
	root    *node
	doctype *node
}

// parse baut aus dem Tokenstrom einen Baum und prüft dabei die Struktur:
// ein Wurzelelement, passende Anfangs- und End-Tags, kein Text außerhalb.
func parse(tokens []token, source []rune) (*document, error) {
	doc := &document{}
	var stack []*node

	add := func(n *node) {
		if len(stack) == 0 {
			doc.nodes = append(doc.nodes, n)
		} else {
			stack[len(stack)-1].children = append(stack[len(stack)-1].children, n)
		}
	}

	for i := range tokens {
		t := &tokens[i]
		switch t.kind {
		case tokText:
			if len(stack) == 0 {
				if hasContent(textOf(source, t)) {
					return nil, fehler(t.line, t.column, "Text außerhalb des Wurzelelements ist nicht erlaubt.")
				}
				continue
			}
			add(&node{kind: nodeText, tok: t})

		case tokComment:
			add(&node{kind: nodeComment, tok: t})

		case tokPI:
			if strings.EqualFold(t.name, "xml") {
				if len(stack) > 0 || len(doc.nodes) > 0 {
					return nil, fehler(t.line, t.column, "Die XML-Deklaration darf nur am Anfang des Dokuments stehen.")
				}
			}
			add(&node{kind: nodePI, tok: t})

		case tokCDATA:
			if len(stack) == 0 {
				return nil, fehler(t.line, t.column, "CDATA darf nur innerhalb des Wurzelelements stehen.")
			}
			add(&node{kind: nodeCDATA, tok: t})

		case tokDoctype:
			if len(stack) > 0 || doc.root != nil {
				return nil, fehler(t.line, t.column, "Die DOCTYPE-Deklaration muss vor dem Wurzelelement stehen.")
			}
			if doc.doctype != nil {
				return nil, fehler(t.line, t.column, "Es darf nur eine DOCTYPE-Deklaration geben.")
			}
			n := &node{kind: nodeDoctype, tok: t}
			doc.doctype = n
			doc.nodes = append(doc.nodes, n)

		case tokStart:
			n := &node{kind: nodeElement, tok: t}
			if len(stack) == 0 {
				if doc.root != nil {
					return nil, fehler(t.line, t.column, "Es darf nur ein Wurzelelement geben.")
				}
				doc.root = n
				doc.nodes = append(doc.nodes, n)
			} else {
				add(n)
			}
			if !t.selfClosing {
				if len(stack) >= maxDepth {
					return nil, fehler(t.line, t.column, "Zu tief verschachtelt (mehr als %d Ebenen).", maxDepth)
				}
				stack = append(stack, n)
			}

		case tokEnd:
			if len(stack) == 0 {
				return nil, fehler(t.line, t.column, "Schließendes Tag </%s> ohne öffnendes Tag.", t.name)
			}
			offen := stack[len(stack)-1]
			if offen.tok.name != t.name {
				return nil, fehler(t.line, t.column, "Erwartet </%s>, gefunden </%s>.", offen.tok.name, t.name)
			}
			offen.end = t
			stack = stack[:len(stack)-1]
		}
	}

	if len(stack) > 0 {
		offen := stack[len(stack)-1]
		return nil, fehler(offen.tok.line, offen.tok.column, "Tag <%s> wird nicht geschlossen.", offen.tok.name)
	}
	if doc.root == nil {
		return nil, fehler(1, 1, "Es fehlt ein Wurzelelement.")
	}
	if doc.doctype != nil && doc.doctype.tok.name != doc.root.tok.name {
		return nil, fehler(doc.doctype.tok.line, doc.doctype.tok.column,
			"Die DOCTYPE-Deklaration nennt '%s', das Wurzelelement heißt '%s'.",
			doc.doctype.tok.name, doc.root.tok.name)
	}
	return doc, nil
}

func textOf(source []rune, t *token) string {
	return string(source[t.start:t.end])
}
