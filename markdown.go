/*  Original C version https://github.com/jgm/peg-markdown/
 *	Copyright 2008 John MacFarlane (jgm at berkeley dot edu).
 *
 *  Modifications and translation from C into Go
 *  based on markdown_lib.c and parsing_functions.c
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

// implements Parse()

import (
	"strings"
	"bytes"
)

// Markdown Extensions:
const (
	EXT_SMART	= 1 << iota
	EXT_NOTES
	EXT_FILTER_HTML
	EXT_FILTER_STYLES
)

// Parse converts a Markdown document into a tree for later output processing.
func Parse(text string, extFlags int) *Doc {
	d := new(Doc)
	d.syntaxExtensions = extFlags

	s := preformat(text)

	d.parseRule(ruleReferences, s)
	if extFlags&EXT_NOTES != 0 {
		d.parseRule(ruleNotes, s)
	}
	raw := d.parseMarkdown(s)
	d.tree = d.processRawBlocks(raw)
	return d
}

func (d *Doc) parseRule(rule int, s string) {
	m := new(yyParser)
	m.Doc = d
	m.Init()
	m.Buffer = s
	if !m.Parse(rule) {
		m.PrintError()
	}
}

func (d *Doc) parseMarkdown(text string) *element {
	m := new(yyParser)
	m.Doc = d
	m.Init()
	m.Buffer = text
	m.Parse(ruleDoc)
	return d.tree
}


/* process_raw_blocks - traverses an element list, replacing any RAW elements with
 * the result of parsing them as markdown text, and recursing into the children
 * of parent elements.  The result should be a tree of elements without any RAWs.
 */
func (d *Doc) processRawBlocks(input *element) *element {
	var last_child *element

	for current := input; current != nil; current = current.next {
		if current.key == RAW {
			/* \001 is used to indicate boundaries between nested lists when there
			 * is no blank line.  We split the string by \001 and parse
			 * each chunk separately.
			 */
			current.key = LIST
			current.children = nil
			for _, contents := range strings.Split(current.contents.str, "\001", -1) {
				list := d.parseMarkdown(contents)
				if current.children == nil {
					current.children = list
					last_child = list
				} else {
					last_child.next = list
				}
				for last_child.next != nil {
					last_child = last_child.next
				}
			}
			current.contents.str = ""
		}
		if current.children != nil {
			current.children = d.processRawBlocks(current.children)
		}
	}
	return input
}


const (
	TABSTOP = 4
)

/* preformat - allocate and copy text buffer while
 * performing tab expansion.
 */
func preformat(text string) (s string) {
	charstotab := TABSTOP
	i0 := 0

	b := bytes.NewBuffer(make([]byte, 0, len(text)+256))
	for i, _ := range text {
		switch text[i] {
		case '\t':
			b.WriteString(text[i0:i])
			for ; charstotab > 0; charstotab-- {
				b.WriteByte(' ')
			}
			i0 = i + 1
		case '\n':
			b.WriteString(text[i0 : i+1])
			i0 = i + 1
			charstotab = TABSTOP
		default:
			charstotab--
		}
		if charstotab == 0 {
			charstotab = TABSTOP
		}
	}
	b.WriteString(text[i0:])
	b.WriteString("\n\n")
	return b.String()
}
