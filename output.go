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

// HTML output functions

import (
	"os"
	"fmt"
	"log"
	"rand"
	"strings"
)

type Writer interface {
	WriteString(string) (int, os.Error)
	WriteRune(int) (int, os.Error)
	WriteByte(byte) os.Error
}

type htmlOut struct {
	Writer
	padded	int

	notenum		int
	endNotes	[]*element	/* List of endnotes to print after main content. */
}

// WriteHtml prints a document tree in HTML format using the specified Writer.
//
func (d *Doc) WriteHtml(w Writer) int {
	out := new(htmlOut)
	out.Writer = w
	out.padded = 2
	out.elist(d.tree, false)
	if len(out.endNotes) != 0 {
		out.pad(2)
		out.printEndnotes()
	}
	out.WriteByte('\n')
	return 0
}

// pad - add newlines if needed
func (h *htmlOut) pad(n int) *htmlOut {
	for ; n > h.padded; n-- {
		h.WriteByte('\n')
	}
	h.padded = n
	return h
}

func (h *htmlOut) pset(n int) *htmlOut {
	h.padded = n
	return h
}

// print a string
func (w *htmlOut) s(s string) *htmlOut {
	w.WriteString(s)
	return w
}


/* print string, escaping for HTML  
 * If obfuscate selected, convert characters to hex or decimal entities at random
 */
func (w *htmlOut) str(s string, obfuscate bool) *htmlOut {
	var ws string
	var i0 = 0

	for i, r := range s {
		switch r {
		case '&':
			ws = "&amp;"
		case '<':
			ws = "&lt;"
		case '>':
			ws = "&gt;"
		case '"':
			ws = "&quot;"
		default:
			if obfuscate {
				if rand.Intn(1) == 0 {
					ws = fmt.Sprintf("&#%d;", r)
				} else {
					ws = fmt.Sprintf("&#%x;", r)
				}
			} else {
				if i0 == -1 {
					i0 = i
				}
				continue
			}
		}
		if i0 != -1 {
			w.WriteString(s[i0:i])
			i0 = -1
		}
		w.WriteString(ws)
	}
	if i0 != -1 {
		w.WriteString(s[i0:])
	}
	return w
}

/* print a list of elements
 */
func (w *htmlOut) elist(list *element, obfuscate bool) *htmlOut {
	for list != nil {
		w.elem(list, obfuscate)
		list = list.next
	}
	return w
}

// print an element
func (w *htmlOut) elem(elt *element, obfuscate bool) *htmlOut {
	var s string

	switch elt.key {
	case SPACE:
		s = elt.contents.str
	case LINEBREAK:
		s = "<br/>\n"
	case STR:
		w.str(elt.contents.str, obfuscate)
	case ELLIPSIS:
		s = "&hellip;"
	case EMDASH:
		s = "&mdash;"
	case ENDASH:
		s = "&ndash;"
	case APOSTROPHE:
		s = "&rsquo;"
	case SINGLEQUOTED:
		w.s("&lsquo;").elist(elt.children, obfuscate).s("&rsquo;")
	case DOUBLEQUOTED:
		w.s("&ldquo;").elist(elt.children, obfuscate).s("&rdquo;")
	case CODE:
		w.s("<code>").str(elt.contents.str, obfuscate).s("</code>")
	case HTML:
		s = elt.contents.str
	case LINK:
		if strings.Index(elt.contents.link.url, "mailto:") == 0 {
			obfuscate = true	/* obfuscate mailto: links */
		}
		w.s(`<a href="`).str(elt.contents.link.url, obfuscate).s(`"`)
		if len(elt.contents.link.title) > 0 {
			w.s(` title="`).str(elt.contents.link.title, obfuscate).s(`"`)
		}
		w.s(">").elist(elt.contents.link.label, obfuscate).s("</a>")
	case IMAGE:
		w.s(`<img src="`).str(elt.contents.link.url, obfuscate).s(`" alt="`)
		w.elist(elt.contents.link.label, obfuscate).s(`"`)
		if len(elt.contents.link.title) > 0 {
			w.s(` title="`).str(elt.contents.link.title, obfuscate).s(`"`)
		}
		w.s(" />")
	case EMPH:
		w.s("<em>").elist(elt.children, obfuscate).s("</em>")
	case STRONG:
		w.s("<strong>").elist(elt.children, obfuscate).s("</strong>")
	case LIST:
		w.elist(elt.children, obfuscate)
	case RAW:
		/* Shouldn't occur - these are handled by process_raw_blocks() */
		log.Exitf("RAW")
	case H1, H2, H3, H4, H5, H6:
		h := "h" + string('1'+elt.key-H1) + ">"	/* assumes H1 ... H6 are in order */
		w.pad(2).s("<").s(h).elist(elt.children, obfuscate).s("</").s(h).pset(0)
	case PLAIN:
		w.pad(1).elist(elt.children, obfuscate).pset(0)
	case PARA:
		w.pad(2).s("<p>").elist(elt.children, obfuscate).s("</p>").pset(0)
	case HRULE:
		w.pad(2).s("<hr />").pset(0)
	case HTMLBLOCK:
		w.pad(2).s(elt.contents.str).pset(0)
	case VERBATIM:
		w.pad(2).s("<pre><code>").str(elt.contents.str, obfuscate).s("</code></pre>").pset(0)
	case BULLETLIST:
		w.pad(2).s("<ul>").pset(0).elist(elt.children, obfuscate).pad(1).s("</ul>").pset(0)
	case ORDEREDLIST:
		w.pad(2).s("<ol>").pset(0).elist(elt.children, obfuscate).pad(1).s("</ol>").pset(0)
	case DEFINITIONLIST:
		w.pad(2).s("<dl>").pset(0).elist(elt.children, obfuscate).pad(1).s("</dl>").pset(0)
	case DEFTITLE:
		w.pad(1).s("<dt>").pset(2).elist(elt.children, obfuscate).s("</dt>").pset(0)
	case DEFDATA:
		w.pad(1).s("<dd>").pset(2).elist(elt.children, obfuscate).s("</dd>").pset(0)
	case LISTITEM:
		w.pad(1).s("<li>").pset(2).elist(elt.children, obfuscate).s("</li>").pset(0)
	case BLOCKQUOTE:
		w.pad(2).s("<blockquote>\n").pset(2).elist(elt.children, obfuscate).pad(1).s("</blockquote>").pset(0)
	case REFERENCE:
		/* Nonprinting */
	case NOTE:
		/* if contents.str == 0, then print note; else ignore, since this
		 * is a note block that has been incorporated into the notes list
		 */
		if elt.contents.str == "" {
			w.endNotes = append(w.endNotes, elt)	/* add an endnote to global endnotes list */
			w.notenum++
			nn := w.notenum
			s = fmt.Sprintf(`<a class="noteref" id="fnref%d" href="#fn%d" title="Jump to note %d">[%d]</a>`,
				nn, nn, nn, nn)
		}
	default:
		log.Exitf("htmlOut.elem encountered unknown element key = %d\n", elt.key)
	}
	if s != "" {
		w.s(s)
	}
	return w
}


func (w *htmlOut) printEndnotes() {
	counter := 0

	w.s("<hr/>\n<ol id=\"notes\">")
	for _, elt := range w.endNotes {
		counter++
		w.pad(1).s(fmt.Sprintf("<li id=\"fn%d\">\n", counter)).pset(2)
		w.elist(elt.children, false)
		w.s(fmt.Sprintf(" <a href=\"#fnref%d\" title=\"Jump back to reference\">[back]</a>", counter))
		w.pad(1).s("</li>")
	}
	w.pad(1).s("</ol>")
}
