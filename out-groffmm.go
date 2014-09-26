/*  Original C version https://github.com/jgm/peg-markdown/
 *	Copyright 2008 John MacFarlane (jgm at berkeley dot edu).
 *
 *  Modifications and translation from C into Go
 *  based on markdown_output.c
 *	Copyright 2010 Michael Teichgräber (mt at wmipf dot de)
 *
 *  This program is free software; you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License or the MIT
 *  license.  See LICENSE for details.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 */

package markdown

// groff mm output functions

import (
	"log"
	"strings"
)

type troffOut struct {
	baseWriter
	strikeMacroWritten bool
	inListItem         bool
	escape             *strings.Replacer
}

// Returns a formatter that writes the document in groff mm format.
func ToGroffMM(w Writer) Formatter {
	f := new(troffOut)
	f.baseWriter = baseWriter{w, 2}
	f.escape = strings.NewReplacer(`\`, `\e`)
	return f
}
func (f *troffOut) FormatBlock(tree *Element) {
	f.elist(tree)
}
func (f *troffOut) Finish() {
	f.WriteByte('\n')
	f.padded = 2
}

func (h *troffOut) sp() *troffOut {
	h.pad(2)
	return h
}
func (h *troffOut) br() *troffOut {
	h.pad(1)
	return h
}

func (h *troffOut) skipPadding() *troffOut {
	h.padded = 2
	return h
}

// write a string
func (w *troffOut) s(s string) *troffOut {
	w.WriteString(s)
	return w
}

// write string, escape '\'
func (w *troffOut) str(s string) *troffOut {
	if strings.HasPrefix(s, ".") {
		w.WriteString(`\[char46]`)
		s = s[1:]
	}
	w.escape.WriteString(w, s)
	return w
}

func (w *troffOut) children(el *Element) *troffOut {
	return w.elist(el.Children)
}
func (w *troffOut) inline(pfx string, el *Element, sfx string) *troffOut {
	return w.s(pfx).children(el).s(sfx)
}

func (w *troffOut) req(name string) *troffOut {
	return w.br().s(".").s(name)
}

// write a list of elements
func (w *troffOut) elist(list *Element) *troffOut {
	for i := 0; list != nil; i++ {
		w.elem(list, i == 0)
		list = list.Next
	}
	return w
}

func (w *troffOut) elem(elt *Element, isFirst bool) *troffOut {
	var s string

	switch elt.Key {
	case SPACE:
		s = elt.Contents.Str
	case LINEBREAK:
		w.req("br\n")
	case STR:
		w.str(elt.Contents.Str)
	case ELLIPSIS:
		s = "..."
	case EMDASH:
		s = `\[em]`
	case ENDASH:
		s = `\[en]`
	case APOSTROPHE:
		s = "'"
	case SINGLEQUOTED:
		w.inline("`", elt, "'")
	case DOUBLEQUOTED:
		w.inline(`\[lq]`, elt, `\[rq]`)
	case CODE:
		w.s(`\fC`).str(elt.Contents.Str).s(`\fR`)
	case HTML:
		/* don't print HTML */
	case LINK:
		link := elt.Contents.Link
		w.elist(link.Label)
		w.s(" (").s(link.URL).s(")")
	case IMAGE:
		w.s("[IMAGE: ").elist(elt.Contents.Link.Label).s("]")
		/* not supported */
	case EMPH:
		w.inline(`\fI`, elt, `\fR`)
	case STRONG:
		w.inline(`\fB`, elt, `\fR`)
	case STRIKE:
		w.s("\\c\n")
		if !w.strikeMacroWritten {
			w.s(`.de ST
.nr width \w'\\$1'
\Z@\v'-.25m'\l'\\n[width]u'@\\$1\c
..
`)
			w.strikeMacroWritten = true
		}
		w.inline(".ST \"", elt, `"`).br()
	case LIST:
		w.children(elt)
	case RAW:
		/* Shouldn't occur - these are handled by process_raw_blocks() */
		log.Fatalf("RAW")
	case H1, H2, H3, H4, H5, H6:
		h := ".H " + string('1'+elt.Key-H1) + ` "` /* assumes H1 ... H6 are in order */
		w.br().inline(h, elt, `"`)
	case PLAIN:
		w.br().children(elt)
	case PARA:
		if !w.inListItem || !isFirst {
			w.req("P\n").children(elt)
		} else {
			w.br().children(elt)
		}
	case HRULE:
		w.br().s(`\l'\n(.lu*8u/10u'`)
	case HTMLBLOCK:
		/* don't print HTML block */
	case VERBATIM:
		w.req("VERBON 2\n")
		w.str(elt.Contents.Str)
		w.s(".VERBOFF")
	case BULLETLIST:
		w.req("BL").children(elt).req("LE 1")
	case ORDEREDLIST:
		w.req("AL").children(elt).req("LE 1")
	case DEFINITIONLIST:
		w.req(`BVL \\n(Pin`).children(elt).req("LE 1")
	case DEFTITLE:
		w.req(`DLI "`).children(elt).s(`"`)
	case DEFDATA:
		w.children(elt)
		w.req("br")
	case LISTITEM:
		w.req("LI\n")
		w.inListItem = true
		w.skipPadding()
		w.children(elt)
		w.inListItem = false
	case BLOCKQUOTE:
		w.req("DS I\n")
		w.skipPadding()
		w.children(elt)
		w.req("DE")
	case NOTE:
		/* if Contents.Str == 0, then print note; else ignore, since this
		 * is a note block that has been incorporated into the notes list */
		if elt.Contents.Str == "" {
			w.s("\\*F\n")
			w.s(".FS\n")
			w.skipPadding()
			w.children(elt)
			w.req("FE\n")
		}
	case REFERENCE:
		/* Nonprinting */
	default:
		log.Fatalf("troffOut.elem encountered unknown Element key = %d\n", elt.Key)
	}
	if s != "" {
		w.s(s)
	}
	return w
}
