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
	"bytes"
	"io"
	"log"
	"strings"
)

// Markdown Options:
type Options struct {
	Smart        bool
	Notes        bool
	FilterHTML   bool
	FilterStyles bool
	Dlists       bool
}

// Parse converts a Markdown document into a tree for later output processing.
func Parse(r io.Reader, opt Options) *Doc {
	d := new(Doc)
	d.extension = opt

	d.parser = new(yyParser)
	d.parser.Doc = d
	d.parser.Init()

	s := preformat(r)

	d.parseRule(ruleReferences, s)
	if opt.Notes {
		d.parseRule(ruleNotes, s)
	}
	raw := d.parseMarkdown(s)
	d.tree = d.processRawBlocks(raw)
	return d
}

func (d *Doc) parseRule(rule int, s string) {
	m := d.parser
	if m.ResetBuffer(s) != "" {
		log.Fatalf("Buffer not empty")
	}
	if err := m.Parse(rule); err != nil {
		log.Fatalln("markdown:", err)
	}
}

func (d *Doc) parseMarkdown(text string) *element {
	d.parseRule(ruleDoc, text)
	return d.tree
}

/* process_raw_blocks - traverses an element list, replacing any RAW elements with
 * the result of parsing them as markdown text, and recursing into the children
 * of parent elements.  The result should be a tree of elements without any RAWs.
 */
func (d *Doc) processRawBlocks(input *element) *element {

	for current := input; current != nil; current = current.next {
		if current.key == RAW {
			/* \001 is used to indicate boundaries between nested lists when there
			 * is no blank line.  We split the string by \001 and parse
			 * each chunk separately.
			 */
			current.key = LIST
			current.children = nil
			listEnd := &current.children
			for _, contents := range strings.Split(current.contents.str, "\001") {
				if list := d.parseMarkdown(contents); list != nil {
					*listEnd = list
					for list.next != nil {
						list = list.next
					}
					listEnd = &list.next
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
func preformat(r io.Reader) (s string) {
	charstotab := TABSTOP
	buf := make([]byte, 32768)

	b := bytes.NewBuffer(make([]byte, 0, 32768))
	for {
		n, err := r.Read(buf)
		if err != nil {
			break
		}
		i0 := 0
		for i := range buf[:n] {
			switch buf[i] {
			case '\t':
				b.Write(buf[i0:i])
				for ; charstotab > 0; charstotab-- {
					b.WriteByte(' ')
				}
				i0 = i + 1
			case '\n':
				b.Write(buf[i0 : i+1])
				i0 = i + 1
				charstotab = TABSTOP
			default:
				charstotab--
			}
			if charstotab == 0 {
				charstotab = TABSTOP
			}
		}
		b.Write(buf[i0:n])
	}

	b.WriteString("\n")
	return b.String()
}
