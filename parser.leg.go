/*  Original C version https://github.com/jgm/peg-markdown/
 *	Copyright 2008 John MacFarlane (jgm at berkeley dot edu).
 *
 *  Modifications and translation from C into Go
 *  based on markdown_parser.leg and utility_functions.c
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

// PEG grammar and parser actions for markdown syntax.

import (
	"fmt"
	"io"
	"log"
	"strings"
)

const (
	parserIfaceVersion_16 = iota
)

// Semantic value of a parsing action.
type element struct {
	key int
	contents
	children *element
	next     *element
}

// Information (label, URL and title) for a link.
type link struct {
	label *element
	url   string
	title string
}

// Union for contents of an Element (string, list, or link).
type contents struct {
	str string
	*link
}

// Types of semantic values returned by parsers.
const (
	LIST = iota /* A generic list of values. For ordered and bullet lists, see below. */
	RAW         /* Raw markdown to be processed further */
	SPACE
	LINEBREAK
	ELLIPSIS
	EMDASH
	ENDASH
	APOSTROPHE
	SINGLEQUOTED
	DOUBLEQUOTED
	STR
	LINK
	IMAGE
	CODE
	HTML
	EMPH
	STRONG
	PLAIN
	PARA
	LISTITEM
	BULLETLIST
	ORDEREDLIST
	H1 /* Code assumes that H1..6 are in order. */
	H2
	H3
	H4
	H5
	H6
	BLOCKQUOTE
	VERBATIM
	HTMLBLOCK
	HRULE
	REFERENCE
	NOTE
	DEFINITIONLIST
	DEFTITLE
	DEFDATA
	numVAL
)

type state struct {
	extension  Extensions
	heap       elemHeap
	tree       *element /* Results of parse. */
	references *element /* List of link references found. */
	notes      *element /* List of footnotes found. */
}

const (
	ruleDoc = iota
	ruleDocblock
	ruleBlock
	rulePara
	rulePlain
	ruleAtxInline
	ruleAtxStart
	ruleAtxHeading
	ruleSetextHeading
	ruleSetextBottom1
	ruleSetextBottom2
	ruleSetextHeading1
	ruleSetextHeading2
	ruleHeading
	ruleBlockQuote
	ruleBlockQuoteRaw
	ruleNonblankIndentedLine
	ruleVerbatimChunk
	ruleVerbatim
	ruleHorizontalRule
	ruleBullet
	ruleBulletList
	ruleListTight
	ruleListLoose
	ruleListItem
	ruleListItemTight
	ruleListBlock
	ruleListContinuationBlock
	ruleEnumerator
	ruleOrderedList
	ruleListBlockLine
	ruleHtmlBlockOpenAddress
	ruleHtmlBlockCloseAddress
	ruleHtmlBlockAddress
	ruleHtmlBlockOpenBlockquote
	ruleHtmlBlockCloseBlockquote
	ruleHtmlBlockBlockquote
	ruleHtmlBlockOpenCenter
	ruleHtmlBlockCloseCenter
	ruleHtmlBlockCenter
	ruleHtmlBlockOpenDir
	ruleHtmlBlockCloseDir
	ruleHtmlBlockDir
	ruleHtmlBlockOpenDiv
	ruleHtmlBlockCloseDiv
	ruleHtmlBlockDiv
	ruleHtmlBlockOpenDl
	ruleHtmlBlockCloseDl
	ruleHtmlBlockDl
	ruleHtmlBlockOpenFieldset
	ruleHtmlBlockCloseFieldset
	ruleHtmlBlockFieldset
	ruleHtmlBlockOpenForm
	ruleHtmlBlockCloseForm
	ruleHtmlBlockForm
	ruleHtmlBlockOpenH1
	ruleHtmlBlockCloseH1
	ruleHtmlBlockH1
	ruleHtmlBlockOpenH2
	ruleHtmlBlockCloseH2
	ruleHtmlBlockH2
	ruleHtmlBlockOpenH3
	ruleHtmlBlockCloseH3
	ruleHtmlBlockH3
	ruleHtmlBlockOpenH4
	ruleHtmlBlockCloseH4
	ruleHtmlBlockH4
	ruleHtmlBlockOpenH5
	ruleHtmlBlockCloseH5
	ruleHtmlBlockH5
	ruleHtmlBlockOpenH6
	ruleHtmlBlockCloseH6
	ruleHtmlBlockH6
	ruleHtmlBlockOpenMenu
	ruleHtmlBlockCloseMenu
	ruleHtmlBlockMenu
	ruleHtmlBlockOpenNoframes
	ruleHtmlBlockCloseNoframes
	ruleHtmlBlockNoframes
	ruleHtmlBlockOpenNoscript
	ruleHtmlBlockCloseNoscript
	ruleHtmlBlockNoscript
	ruleHtmlBlockOpenOl
	ruleHtmlBlockCloseOl
	ruleHtmlBlockOl
	ruleHtmlBlockOpenP
	ruleHtmlBlockCloseP
	ruleHtmlBlockP
	ruleHtmlBlockOpenPre
	ruleHtmlBlockClosePre
	ruleHtmlBlockPre
	ruleHtmlBlockOpenTable
	ruleHtmlBlockCloseTable
	ruleHtmlBlockTable
	ruleHtmlBlockOpenUl
	ruleHtmlBlockCloseUl
	ruleHtmlBlockUl
	ruleHtmlBlockOpenDd
	ruleHtmlBlockCloseDd
	ruleHtmlBlockDd
	ruleHtmlBlockOpenDt
	ruleHtmlBlockCloseDt
	ruleHtmlBlockDt
	ruleHtmlBlockOpenFrameset
	ruleHtmlBlockCloseFrameset
	ruleHtmlBlockFrameset
	ruleHtmlBlockOpenLi
	ruleHtmlBlockCloseLi
	ruleHtmlBlockLi
	ruleHtmlBlockOpenTbody
	ruleHtmlBlockCloseTbody
	ruleHtmlBlockTbody
	ruleHtmlBlockOpenTd
	ruleHtmlBlockCloseTd
	ruleHtmlBlockTd
	ruleHtmlBlockOpenTfoot
	ruleHtmlBlockCloseTfoot
	ruleHtmlBlockTfoot
	ruleHtmlBlockOpenTh
	ruleHtmlBlockCloseTh
	ruleHtmlBlockTh
	ruleHtmlBlockOpenThead
	ruleHtmlBlockCloseThead
	ruleHtmlBlockThead
	ruleHtmlBlockOpenTr
	ruleHtmlBlockCloseTr
	ruleHtmlBlockTr
	ruleHtmlBlockOpenScript
	ruleHtmlBlockCloseScript
	ruleHtmlBlockScript
	ruleHtmlBlockInTags
	ruleHtmlBlock
	ruleHtmlBlockSelfClosing
	ruleHtmlBlockType
	ruleStyleOpen
	ruleStyleClose
	ruleInStyleTags
	ruleStyleBlock
	ruleInlines
	ruleInline
	ruleSpace
	ruleStr
	ruleStrChunk
	ruleAposChunk
	ruleEscapedChar
	ruleEntity
	ruleEndline
	ruleNormalEndline
	ruleTerminalEndline
	ruleLineBreak
	ruleSymbol
	ruleUlOrStarLine
	ruleStarLine
	ruleUlLine
	ruleEmph
	ruleWhitespace
	ruleEmphStar
	ruleEmphUl
	ruleStrong
	ruleStrongStar
	ruleStrongUl
	ruleImage
	ruleLink
	ruleReferenceLink
	ruleReferenceLinkDouble
	ruleReferenceLinkSingle
	ruleExplicitLink
	ruleSource
	ruleSourceContents
	ruleTitle
	ruleTitleSingle
	ruleTitleDouble
	ruleAutoLink
	ruleAutoLinkUrl
	ruleAutoLinkEmail
	ruleReference
	ruleLabel
	ruleRefSrc
	ruleRefTitle
	ruleEmptyTitle
	ruleRefTitleSingle
	ruleRefTitleDouble
	ruleRefTitleParens
	ruleReferences
	ruleTicks1
	ruleTicks2
	ruleTicks3
	ruleTicks4
	ruleTicks5
	ruleCode
	ruleRawHtml
	ruleBlankLine
	ruleQuoted
	ruleHtmlAttribute
	ruleHtmlComment
	ruleHtmlTag
	ruleEof
	ruleSpacechar
	ruleNonspacechar
	ruleNewline
	ruleSp
	ruleSpnl
	ruleSpecialChar
	ruleNormalChar
	ruleAlphanumeric
	ruleAlphanumericAscii
	ruleDigit
	ruleHexEntity
	ruleDecEntity
	ruleCharEntity
	ruleNonindentSpace
	ruleIndent
	ruleIndentedLine
	ruleOptionallyIndentedLine
	ruleStartList
	ruleLine
	ruleRawLine
	ruleSkipBlock
	ruleExtendedSpecialChar
	ruleSmart
	ruleApostrophe
	ruleEllipsis
	ruleDash
	ruleEnDash
	ruleEmDash
	ruleSingleQuoteStart
	ruleSingleQuoteEnd
	ruleSingleQuoted
	ruleDoubleQuoteStart
	ruleDoubleQuoteEnd
	ruleDoubleQuoted
	ruleNoteReference
	ruleRawNoteReference
	ruleNote
	ruleInlineNote
	ruleNotes
	ruleRawNoteBlock
	ruleDefinitionList
	ruleDefinition
	ruleDListTitle
	ruleDefTight
	ruleDefLoose
	ruleDefmark
	ruleDefMarker
)

type yyParser struct {
	state
	Buffer      string
	Min, Max    int
	rules       [244]func() bool
	ResetBuffer func(string) string
}

func (p *yyParser) Parse(ruleId int) (err error) {
	if p.rules[ruleId]() {
		return
	}
	return p.parseErr()
}

type errPos struct {
	Line, Pos int
}

func (e *errPos) String() string {
	return fmt.Sprintf("%d:%d", e.Line, e.Pos)
}

type unexpectedCharError struct {
	After, At errPos
	Char      byte
}

func (e *unexpectedCharError) Error() string {
	return fmt.Sprintf("%v: unexpected character '%c'", &e.At, e.Char)
}

type unexpectedEOFError struct {
	After errPos
}

func (e *unexpectedEOFError) Error() string {
	return fmt.Sprintf("%v: unexpected end of file", &e.After)
}

func (p *yyParser) parseErr() (err error) {
	var pos, after errPos
	pos.Line = 1
	for i, c := range p.Buffer[0:] {
		if c == '\n' {
			pos.Line++
			pos.Pos = 0
		} else {
			pos.Pos++
		}
		if i == p.Min {
			if p.Min != p.Max {
				after = pos
			} else {
				break
			}
		} else if i == p.Max {
			break
		}
	}
	if p.Max >= len(p.Buffer) {
		err = &unexpectedEOFError{after}
	} else {
		err = &unexpectedCharError{after, pos, p.Buffer[p.Max]}
	}
	return
}

func (p *yyParser) Init() {
	var position int
	var yyp int
	var yy *element
	var yyval = make([]*element, 256)

	actions := [...]func(string, int){
		/* 0 Doc */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 1 Doc */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			p.tree = reverse(a)
			yyval[yyp-1] = a
		},
		/* 2 Docblock */
		func(yytext string, _ int) {
			p.tree = yy
		},
		/* 3 Para */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = a
			yy.key = PARA
			yyval[yyp-1] = a
		},
		/* 4 Plain */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = a
			yy.key = PLAIN
			yyval[yyp-1] = a
		},
		/* 5 AtxStart */
		func(yytext string, _ int) {
			yy = p.mkElem(H1 + (len(yytext) - 1))
		},
		/* 6 AtxHeading */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(yy, a)
			yyval[yyp-2] = a
			yyval[yyp-1] = s
		},
		/* 7 AtxHeading */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			a := yyval[yyp-2]
			yy = p.mkList(s.key, a)
			s = nil
			yyval[yyp-1] = s
			yyval[yyp-2] = a
		},
		/* 8 SetextHeading1 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 9 SetextHeading1 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(H1, a)
			yyval[yyp-1] = a
		},
		/* 10 SetextHeading2 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 11 SetextHeading2 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(H2, a)
			yyval[yyp-1] = a
		},
		/* 12 BlockQuote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkElem(BLOCKQUOTE)
			yy.children = a

			yyval[yyp-1] = a
		},
		/* 13 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 14 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 15 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(p.mkString("\n"), a)
			yyval[yyp-1] = a
		},
		/* 16 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkStringFromList(a, true)
			yy.key = RAW

			yyval[yyp-1] = a
		},
		/* 17 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(p.mkString("\n"), a)
			yyval[yyp-1] = a
		},
		/* 18 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 19 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkStringFromList(a, false)
			yyval[yyp-1] = a
		},
		/* 20 Verbatim */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 21 Verbatim */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkStringFromList(a, false)
			yy.key = VERBATIM
			yyval[yyp-1] = a
		},
		/* 22 HorizontalRule */
		func(yytext string, _ int) {
			yy = p.mkElem(HRULE)
		},
		/* 23 BulletList */
		func(yytext string, _ int) {
			yy.key = BULLETLIST
		},
		/* 24 ListTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 25 ListTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(LIST, a)
			yyval[yyp-1] = a
		},
		/* 26 ListLoose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]

			li := b.children
			li.contents.str += "\n\n"
			a = cons(b, a)

			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 27 ListLoose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			yy = p.mkList(LIST, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 28 ListItem */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 29 ListItem */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 30 ListItem */
		func(yytext string, _ int) {
			a := yyval[yyp-1]

			raw := p.mkStringFromList(a, false)
			raw.key = RAW
			yy = p.mkElem(LISTITEM)
			yy.children = raw

			yyval[yyp-1] = a
		},
		/* 31 ListItemTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 32 ListItemTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 33 ListItemTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]

			raw := p.mkStringFromList(a, false)
			raw.key = RAW
			yy = p.mkElem(LISTITEM)
			yy.children = raw

			yyval[yyp-1] = a
		},
		/* 34 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 35 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 36 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkStringFromList(a, false)
			yyval[yyp-1] = a
		},
		/* 37 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			if len(yytext) == 0 {
				a = cons(p.mkString("\001"), a) // block separator
			} else {
				a = cons(p.mkString(yytext), a)
			}

			yyval[yyp-1] = a
		},
		/* 38 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 39 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkStringFromList(a, false)
			yyval[yyp-1] = a
		},
		/* 40 OrderedList */
		func(yytext string, _ int) {
			yy.key = ORDEREDLIST
		},
		/* 41 HtmlBlock */
		func(yytext string, _ int) {
			if p.extension.FilterHTML {
				yy = p.mkList(LIST, nil)
			} else {
				yy = p.mkString(yytext)
				yy.key = HTMLBLOCK
			}

		},
		/* 42 StyleBlock */
		func(yytext string, _ int) {
			if p.extension.FilterStyles {
				yy = p.mkList(LIST, nil)
			} else {
				yy = p.mkString(yytext)
				yy.key = HTMLBLOCK
			}

		},
		/* 43 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			a = cons(yy, a)
			yyval[yyp-2] = c
			yyval[yyp-1] = a
		},
		/* 44 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			a = cons(c, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = c
		},
		/* 45 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			yy = p.mkList(LIST, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = c
		},
		/* 46 Space */
		func(yytext string, _ int) {
			yy = p.mkString(" ")
			yy.key = SPACE
		},
		/* 47 Str */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(p.mkString(yytext), a)
			yyval[yyp-1] = a
		},
		/* 48 Str */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 49 Str */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			if a.next == nil {
				yy = a
			} else {
				yy = p.mkList(LIST, a)
			}
			yyval[yyp-1] = a
		},
		/* 50 StrChunk */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 51 AposChunk */
		func(yytext string, _ int) {
			yy = p.mkElem(APOSTROPHE)
		},
		/* 52 EscapedChar */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 53 Entity */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
			yy.key = HTML
		},
		/* 54 NormalEndline */
		func(yytext string, _ int) {
			yy = p.mkString("\n")
			yy.key = SPACE
		},
		/* 55 TerminalEndline */
		func(yytext string, _ int) {
			yy = nil
		},
		/* 56 LineBreak */
		func(yytext string, _ int) {
			yy = p.mkElem(LINEBREAK)
		},
		/* 57 Symbol */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 58 UlOrStarLine */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 59 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 60 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 61 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			yy = p.mkList(EMPH, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 62 EmphUl */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 63 EmphUl */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 64 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			yy = p.mkList(EMPH, a)
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 65 StrongStar */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 66 StrongStar */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			yy = p.mkList(STRONG, a)
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 67 StrongUl */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 68 StrongUl */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			yy = p.mkList(STRONG, a)
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 69 Image */
		func(yytext string, _ int) {
			if yy.key == LINK {
				yy.key = IMAGE
			} else {
				result := yy
				yy.children = cons(p.mkString("!"), result.children)
			}

		},
		/* 70 ReferenceLinkDouble */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]

			if match, found := p.findReference(b.children); found {
				yy = p.mkLink(a.children, match.url, match.title)
				a = nil
				b = nil
			} else {
				result := p.mkElem(LIST)
				result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), cons(p.mkString(yytext),
					cons(p.mkString("["), cons(b, p.mkString("]")))))))
				yy = result
			}

			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 71 ReferenceLinkSingle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]

			if match, found := p.findReference(a.children); found {
				yy = p.mkLink(a.children, match.url, match.title)
				a = nil
			} else {
				result := p.mkElem(LIST)
				result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), p.mkString(yytext))))
				yy = result
			}

			yyval[yyp-1] = a
		},
		/* 72 ExplicitLink */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			t := yyval[yyp-2]
			l := yyval[yyp-3]
			yy = p.mkLink(l.children, s.contents.str, t.contents.str)
			s = nil
			t = nil
			l = nil
			yyval[yyp-1] = s
			yyval[yyp-2] = t
			yyval[yyp-3] = l
		},
		/* 73 Source */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 74 Title */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 75 AutoLinkUrl */
		func(yytext string, _ int) {
			yy = p.mkLink(p.mkString(yytext), yytext, "")
		},
		/* 76 AutoLinkEmail */
		func(yytext string, _ int) {

			yy = p.mkLink(p.mkString(yytext), "mailto:"+yytext, "")

		},
		/* 77 Reference */
		func(yytext string, _ int) {
			l := yyval[yyp-1]
			t := yyval[yyp-2]
			s := yyval[yyp-3]
			yy = p.mkLink(l.children, s.contents.str, t.contents.str)
			s = nil
			t = nil
			l = nil
			yy.key = REFERENCE
			yyval[yyp-2] = t
			yyval[yyp-3] = s
			yyval[yyp-1] = l
		},
		/* 78 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 79 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(LIST, a)
			yyval[yyp-1] = a
		},
		/* 80 RefSrc */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
			yy.key = HTML
		},
		/* 81 RefTitle */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 82 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 83 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			p.references = reverse(a)
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 84 Code */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
			yy.key = CODE
		},
		/* 85 RawHtml */
		func(yytext string, _ int) {
			if p.extension.FilterHTML {
				yy = p.mkList(LIST, nil)
			} else {
				yy = p.mkString(yytext)
				yy.key = HTML
			}

		},
		/* 86 StartList */
		func(yytext string, _ int) {
			yy = nil
		},
		/* 87 Line */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 88 Apostrophe */
		func(yytext string, _ int) {
			yy = p.mkElem(APOSTROPHE)
		},
		/* 89 Ellipsis */
		func(yytext string, _ int) {
			yy = p.mkElem(ELLIPSIS)
		},
		/* 90 EnDash */
		func(yytext string, _ int) {
			yy = p.mkElem(ENDASH)
		},
		/* 91 EmDash */
		func(yytext string, _ int) {
			yy = p.mkElem(EMDASH)
		},
		/* 92 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 93 SingleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-2]
			a := yyval[yyp-1]
			yy = p.mkList(SINGLEQUOTED, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 94 DoubleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 95 DoubleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			yy = p.mkList(DOUBLEQUOTED, a)
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 96 NoteReference */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]

			if match, ok := p.find_note(ref.contents.str); ok {
				yy = p.mkElem(NOTE)
				yy.children = match.children
				yy.contents.str = ""
			} else {
				yy = p.mkString("[^" + ref.contents.str + "]")
			}

			yyval[yyp-1] = ref
		},
		/* 97 RawNoteReference */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 98 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(yy, a)
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 99 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			ref := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 100 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			ref := yyval[yyp-1]
			yy = p.mkList(NOTE, a)
			yy.contents.str = ref.contents.str

			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 101 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 102 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(NOTE, a)
			yy.contents.str = ""
			yyval[yyp-1] = a
		},
		/* 103 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 104 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			p.notes = reverse(a)
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 105 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 106 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(p.mkString(yytext), a)
			yyval[yyp-1] = a
		},
		/* 107 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkStringFromList(a, true)
			yy.key = RAW

			yyval[yyp-1] = a
		},
		/* 108 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 109 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(DEFINITIONLIST, a)
			yyval[yyp-1] = a
		},
		/* 110 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 111 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]

			for e := yy.children; e != nil; e = e.next {
				e.key = DEFDATA
			}
			a = cons(yy, a)

			yyval[yyp-1] = a
		},
		/* 112 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(LIST, a)
			yyval[yyp-1] = a
		},
		/* 113 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 114 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(LIST, a)
			yy.key = DEFTITLE

			yyval[yyp-1] = a
		},

		/* yyPush */
		func(_ string, count int) {
			yyp += count
			if yyp >= len(yyval) {
				s := make([]*element, cap(yyval)+256)
				copy(s, yyval)
				yyval = s
			}
		},
		/* yyPop */
		func(_ string, count int) {
			yyp -= count
		},
		/* yySet */
		func(_ string, count int) {
			yyval[yyp+count] = yy
		},
	}
	const (
		yyPush = 115 + iota
		yyPop
		yySet
	)

	type thunk struct {
		action     uint8
		begin, end int
	}
	var thunkPosition, begin, end int
	thunks := make([]thunk, 32)
	doarg := func(action uint8, arg int) {
		if thunkPosition == len(thunks) {
			newThunks := make([]thunk, 2*len(thunks))
			copy(newThunks, thunks)
			thunks = newThunks
		}
		t := &thunks[thunkPosition]
		thunkPosition++
		t.action = action
		if arg != 0 {
			t.begin = arg // use begin to store an argument
		} else {
			t.begin = begin
		}
		t.end = end
	}
	do := func(action uint8) {
		doarg(action, 0)
	}

	p.ResetBuffer = func(s string) (old string) {
		if position < len(p.Buffer) {
			old = p.Buffer[position:]
		}
		p.Buffer = s
		thunkPosition = 0
		position = 0
		p.Min = 0
		p.Max = 0
		end = 0
		return
	}

	commit := func(thunkPosition0 int) bool {
		if thunkPosition0 == 0 {
			s := ""
			for _, t := range thunks[:thunkPosition] {
				b := t.begin
				if b >= 0 && b <= t.end {
					s = p.Buffer[b:t.end]
				}
				magic := b
				actions[t.action](s, magic)
			}
			p.Min = position
			thunkPosition = 0
			return true
		}
		return false
	}
	matchDot := func() bool {
		if position < len(p.Buffer) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	matchChar := func(c byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] == c) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	peekChar := func(c byte) bool {
		return position < len(p.Buffer) && p.Buffer[position] == c
	}

	matchString := func(s string) bool {
		length := len(s)
		next := position + length
		if (next <= len(p.Buffer)) && p.Buffer[position] == s[0] && (p.Buffer[position:next] == s) {
			position = next
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	classes := [...][32]uint8{
		3: {0, 0, 0, 0, 50, 232, 255, 3, 254, 255, 255, 135, 254, 255, 255, 71, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		1: {0, 0, 0, 0, 10, 111, 0, 80, 0, 0, 0, 184, 1, 0, 0, 56, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		0: {0, 0, 0, 0, 0, 0, 255, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		4: {0, 0, 0, 0, 0, 0, 255, 3, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		7: {0, 0, 0, 0, 0, 0, 255, 3, 126, 0, 0, 0, 126, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		2: {0, 0, 0, 0, 0, 0, 0, 0, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		5: {0, 0, 0, 0, 0, 0, 255, 3, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		6: {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	matchClass := func(class uint) bool {
		if (position < len(p.Buffer)) &&
			((classes[class][p.Buffer[position]>>3] & (1 << (p.Buffer[position] & 7))) != 0) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	peekClass := func(class uint) bool {
		if (position < len(p.Buffer)) &&
			((classes[class][p.Buffer[position]>>3] & (1 << (p.Buffer[position] & 7))) != 0) {
			return true
		}
		return false
	}

	p.rules = [...]func() bool{

		/* 0 Doc <- (StartList (Block { a = cons(yy, a) })* { p.tree = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l0
			}
			doarg(yySet, -1)
		l1:
			{
				position2 := position
				if !p.rules[ruleBlock]() {
					goto l2
				}
				do(0)
				goto l1
			l2:
				position = position2
			}
			do(1)
			if !(commit(thunkPosition0)) {
				goto l0
			}
			doarg(yyPop, 1)
			return true
		l0:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 1 Docblock <- (Block { p.tree = yy } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlock]() {
				goto l3
			}
			do(2)
			if !(commit(thunkPosition0)) {
				goto l3
			}
			return true
		l3:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 2 Block <- (BlankLine* (BlockQuote / Verbatim / Note / Reference / HorizontalRule / Heading / DefinitionList / OrderedList / BulletList / HtmlBlock / StyleBlock / Para / Plain)) */
		func() bool {
			position0 := position
		l5:
			if !p.rules[ruleBlankLine]() {
				goto l6
			}
			goto l5
		l6:
			if !p.rules[ruleBlockQuote]() {
				goto l8
			}
			goto l7
		l8:
			if !p.rules[ruleVerbatim]() {
				goto l9
			}
			goto l7
		l9:
			if !p.rules[ruleNote]() {
				goto l10
			}
			goto l7
		l10:
			if !p.rules[ruleReference]() {
				goto l11
			}
			goto l7
		l11:
			if !p.rules[ruleHorizontalRule]() {
				goto l12
			}
			goto l7
		l12:
			if !p.rules[ruleHeading]() {
				goto l13
			}
			goto l7
		l13:
			if !p.rules[ruleDefinitionList]() {
				goto l14
			}
			goto l7
		l14:
			if !p.rules[ruleOrderedList]() {
				goto l15
			}
			goto l7
		l15:
			if !p.rules[ruleBulletList]() {
				goto l16
			}
			goto l7
		l16:
			if !p.rules[ruleHtmlBlock]() {
				goto l17
			}
			goto l7
		l17:
			if !p.rules[ruleStyleBlock]() {
				goto l18
			}
			goto l7
		l18:
			if !p.rules[rulePara]() {
				goto l19
			}
			goto l7
		l19:
			if !p.rules[rulePlain]() {
				goto l4
			}
		l7:
			return true
		l4:
			position = position0
			return false
		},
		/* 3 Para <- (NonindentSpace Inlines BlankLine+ { yy = a; yy.key = PARA }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l20
			}
			if !p.rules[ruleInlines]() {
				goto l20
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l20
			}
		l21:
			if !p.rules[ruleBlankLine]() {
				goto l22
			}
			goto l21
		l22:
			do(3)
			doarg(yyPop, 1)
			return true
		l20:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 4 Plain <- (Inlines { yy = a; yy.key = PLAIN }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleInlines]() {
				goto l23
			}
			doarg(yySet, -1)
			do(4)
			doarg(yyPop, 1)
			return true
		l23:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 5 AtxInline <- (!Newline !(Sp? '#'* Sp Newline) Inline) */
		func() bool {
			position0 := position
			if !p.rules[ruleNewline]() {
				goto l25
			}
			goto l24
		l25:
			{
				position26 := position
				if !p.rules[ruleSp]() {
					goto l27
				}
			l27:
			l29:
				if !matchChar('#') {
					goto l30
				}
				goto l29
			l30:
				if !p.rules[ruleSp]() {
					goto l26
				}
				if !p.rules[ruleNewline]() {
					goto l26
				}
				goto l24
			l26:
				position = position26
			}
			if !p.rules[ruleInline]() {
				goto l24
			}
			return true
		l24:
			position = position0
			return false
		},
		/* 6 AtxStart <- (&'#' < ('######' / '#####' / '####' / '###' / '##' / '#') > { yy = p.mkElem(H1 + (len(yytext) - 1)) }) */
		func() bool {
			position0 := position
			if !peekChar('#') {
				goto l31
			}
			begin = position
			if !matchString("######") {
				goto l33
			}
			goto l32
		l33:
			if !matchString("#####") {
				goto l34
			}
			goto l32
		l34:
			if !matchString("####") {
				goto l35
			}
			goto l32
		l35:
			if !matchString("###") {
				goto l36
			}
			goto l32
		l36:
			if !matchString("##") {
				goto l37
			}
			goto l32
		l37:
			if !matchChar('#') {
				goto l31
			}
		l32:
			end = position
			do(5)
			return true
		l31:
			position = position0
			return false
		},
		/* 7 AtxHeading <- (AtxStart Sp? StartList (AtxInline { a = cons(yy, a) })+ (Sp? '#'* Sp)? Newline { yy = p.mkList(s.key, a)
		   s = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleAtxStart]() {
				goto l38
			}
			doarg(yySet, -1)
			if !p.rules[ruleSp]() {
				goto l39
			}
		l39:
			if !p.rules[ruleStartList]() {
				goto l38
			}
			doarg(yySet, -2)
			if !p.rules[ruleAtxInline]() {
				goto l38
			}
			do(6)
		l41:
			{
				position42 := position
				if !p.rules[ruleAtxInline]() {
					goto l42
				}
				do(6)
				goto l41
			l42:
				position = position42
			}
			{
				position43 := position
				if !p.rules[ruleSp]() {
					goto l45
				}
			l45:
			l47:
				if !matchChar('#') {
					goto l48
				}
				goto l47
			l48:
				if !p.rules[ruleSp]() {
					goto l43
				}
				goto l44
			l43:
				position = position43
			}
		l44:
			if !p.rules[ruleNewline]() {
				goto l38
			}
			do(7)
			doarg(yyPop, 2)
			return true
		l38:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 8 SetextHeading <- (SetextHeading1 / SetextHeading2) */
		func() bool {
			if !p.rules[ruleSetextHeading1]() {
				goto l51
			}
			goto l50
		l51:
			if !p.rules[ruleSetextHeading2]() {
				goto l49
			}
		l50:
			return true
		l49:
			return false
		},
		/* 9 SetextBottom1 <- ('='+ Newline) */
		func() bool {
			position0 := position
			if !matchChar('=') {
				goto l52
			}
		l53:
			if !matchChar('=') {
				goto l54
			}
			goto l53
		l54:
			if !p.rules[ruleNewline]() {
				goto l52
			}
			return true
		l52:
			position = position0
			return false
		},
		/* 10 SetextBottom2 <- ('-'+ Newline) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l55
			}
		l56:
			if !matchChar('-') {
				goto l57
			}
			goto l56
		l57:
			if !p.rules[ruleNewline]() {
				goto l55
			}
			return true
		l55:
			position = position0
			return false
		},
		/* 11 SetextHeading1 <- (&(RawLine SetextBottom1) StartList (!Endline Inline { a = cons(yy, a) })+ Sp? Newline SetextBottom1 { yy = p.mkList(H1, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position59 := position
				if !p.rules[ruleRawLine]() {
					goto l58
				}
				if !p.rules[ruleSetextBottom1]() {
					goto l58
				}
				position = position59
			}
			if !p.rules[ruleStartList]() {
				goto l58
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l62
			}
			goto l58
		l62:
			if !p.rules[ruleInline]() {
				goto l58
			}
			do(8)
		l60:
			{
				position61 := position
				if !p.rules[ruleEndline]() {
					goto l63
				}
				goto l61
			l63:
				if !p.rules[ruleInline]() {
					goto l61
				}
				do(8)
				goto l60
			l61:
				position = position61
			}
			if !p.rules[ruleSp]() {
				goto l64
			}
		l64:
			if !p.rules[ruleNewline]() {
				goto l58
			}
			if !p.rules[ruleSetextBottom1]() {
				goto l58
			}
			do(9)
			doarg(yyPop, 1)
			return true
		l58:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 12 SetextHeading2 <- (&(RawLine SetextBottom2) StartList (!Endline Inline { a = cons(yy, a) })+ Sp? Newline SetextBottom2 { yy = p.mkList(H2, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position67 := position
				if !p.rules[ruleRawLine]() {
					goto l66
				}
				if !p.rules[ruleSetextBottom2]() {
					goto l66
				}
				position = position67
			}
			if !p.rules[ruleStartList]() {
				goto l66
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l70
			}
			goto l66
		l70:
			if !p.rules[ruleInline]() {
				goto l66
			}
			do(10)
		l68:
			{
				position69 := position
				if !p.rules[ruleEndline]() {
					goto l71
				}
				goto l69
			l71:
				if !p.rules[ruleInline]() {
					goto l69
				}
				do(10)
				goto l68
			l69:
				position = position69
			}
			if !p.rules[ruleSp]() {
				goto l72
			}
		l72:
			if !p.rules[ruleNewline]() {
				goto l66
			}
			if !p.rules[ruleSetextBottom2]() {
				goto l66
			}
			do(11)
			doarg(yyPop, 1)
			return true
		l66:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 13 Heading <- (SetextHeading / AtxHeading) */
		func() bool {
			if !p.rules[ruleSetextHeading]() {
				goto l76
			}
			goto l75
		l76:
			if !p.rules[ruleAtxHeading]() {
				goto l74
			}
		l75:
			return true
		l74:
			return false
		},
		/* 14 BlockQuote <- (BlockQuoteRaw {  yy = p.mkElem(BLOCKQUOTE)
		   yy.children = a
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleBlockQuoteRaw]() {
				goto l77
			}
			doarg(yySet, -1)
			do(12)
			doarg(yyPop, 1)
			return true
		l77:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 BlockQuoteRaw <- (StartList ('>' ' '? Line { a = cons(yy, a) } (!'>' !BlankLine Line { a = cons(yy, a) })* (BlankLine { a = cons(p.mkString("\n"), a) })*)+ {   yy = p.mkStringFromList(a, true)
		    yy.key = RAW
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l78
			}
			doarg(yySet, -1)
			if !matchChar('>') {
				goto l78
			}
			matchChar(' ')
			if !p.rules[ruleLine]() {
				goto l78
			}
			do(13)
		l81:
			{
				position82, thunkPosition82 := position, thunkPosition
				if peekChar('>') {
					goto l82
				}
				if !p.rules[ruleBlankLine]() {
					goto l83
				}
				goto l82
			l83:
				if !p.rules[ruleLine]() {
					goto l82
				}
				do(14)
				goto l81
			l82:
				position, thunkPosition = position82, thunkPosition82
			}
		l84:
			{
				position85 := position
				if !p.rules[ruleBlankLine]() {
					goto l85
				}
				do(15)
				goto l84
			l85:
				position = position85
			}
		l79:
			{
				position80, thunkPosition80 := position, thunkPosition
				if !matchChar('>') {
					goto l80
				}
				matchChar(' ')
				if !p.rules[ruleLine]() {
					goto l80
				}
				do(13)
			l86:
				{
					position87, thunkPosition87 := position, thunkPosition
					if peekChar('>') {
						goto l87
					}
					if !p.rules[ruleBlankLine]() {
						goto l88
					}
					goto l87
				l88:
					if !p.rules[ruleLine]() {
						goto l87
					}
					do(14)
					goto l86
				l87:
					position, thunkPosition = position87, thunkPosition87
				}
			l89:
				{
					position90 := position
					if !p.rules[ruleBlankLine]() {
						goto l90
					}
					do(15)
					goto l89
				l90:
					position = position90
				}
				goto l79
			l80:
				position, thunkPosition = position80, thunkPosition80
			}
			do(16)
			doarg(yyPop, 1)
			return true
		l78:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 NonblankIndentedLine <- (!BlankLine IndentedLine) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l92
			}
			goto l91
		l92:
			if !p.rules[ruleIndentedLine]() {
				goto l91
			}
			return true
		l91:
			position = position0
			return false
		},
		/* 17 VerbatimChunk <- (StartList (BlankLine { a = cons(p.mkString("\n"), a) })* (NonblankIndentedLine { a = cons(yy, a) })+ { yy = p.mkStringFromList(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l93
			}
			doarg(yySet, -1)
		l94:
			{
				position95 := position
				if !p.rules[ruleBlankLine]() {
					goto l95
				}
				do(17)
				goto l94
			l95:
				position = position95
			}
			if !p.rules[ruleNonblankIndentedLine]() {
				goto l93
			}
			do(18)
		l96:
			{
				position97 := position
				if !p.rules[ruleNonblankIndentedLine]() {
					goto l97
				}
				do(18)
				goto l96
			l97:
				position = position97
			}
			do(19)
			doarg(yyPop, 1)
			return true
		l93:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 18 Verbatim <- (StartList (VerbatimChunk { a = cons(yy, a) })+ { yy = p.mkStringFromList(a, false)
		   yy.key = VERBATIM }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l98
			}
			doarg(yySet, -1)
			if !p.rules[ruleVerbatimChunk]() {
				goto l98
			}
			do(20)
		l99:
			{
				position100, thunkPosition100 := position, thunkPosition
				if !p.rules[ruleVerbatimChunk]() {
					goto l100
				}
				do(20)
				goto l99
			l100:
				position, thunkPosition = position100, thunkPosition100
			}
			do(21)
			doarg(yyPop, 1)
			return true
		l98:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 19 HorizontalRule <- (NonindentSpace ((&[_] ('_' Sp '_' Sp '_' (Sp '_')*)) | (&[\-] ('-' Sp '-' Sp '-' (Sp '-')*)) | (&[*] ('*' Sp '*' Sp '*' (Sp '*')*))) Sp Newline BlankLine+ { yy = p.mkElem(HRULE) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l101
			}
			{
				if position == len(p.Buffer) {
					goto l101
				}
				switch p.Buffer[position] {
				case '_':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l101
					}
					if !matchChar('_') {
						goto l101
					}
					if !p.rules[ruleSp]() {
						goto l101
					}
					if !matchChar('_') {
						goto l101
					}
				l103:
					{
						position104 := position
						if !p.rules[ruleSp]() {
							goto l104
						}
						if !matchChar('_') {
							goto l104
						}
						goto l103
					l104:
						position = position104
					}
					break
				case '-':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l101
					}
					if !matchChar('-') {
						goto l101
					}
					if !p.rules[ruleSp]() {
						goto l101
					}
					if !matchChar('-') {
						goto l101
					}
				l105:
					{
						position106 := position
						if !p.rules[ruleSp]() {
							goto l106
						}
						if !matchChar('-') {
							goto l106
						}
						goto l105
					l106:
						position = position106
					}
					break
				case '*':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l101
					}
					if !matchChar('*') {
						goto l101
					}
					if !p.rules[ruleSp]() {
						goto l101
					}
					if !matchChar('*') {
						goto l101
					}
				l107:
					{
						position108 := position
						if !p.rules[ruleSp]() {
							goto l108
						}
						if !matchChar('*') {
							goto l108
						}
						goto l107
					l108:
						position = position108
					}
					break
				default:
					goto l101
				}
			}
			if !p.rules[ruleSp]() {
				goto l101
			}
			if !p.rules[ruleNewline]() {
				goto l101
			}
			if !p.rules[ruleBlankLine]() {
				goto l101
			}
		l109:
			if !p.rules[ruleBlankLine]() {
				goto l110
			}
			goto l109
		l110:
			do(22)
			return true
		l101:
			position = position0
			return false
		},
		/* 20 Bullet <- (!HorizontalRule NonindentSpace ((&[\-] '-') | (&[*] '*') | (&[+] '+')) Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHorizontalRule]() {
				goto l112
			}
			goto l111
		l112:
			if !p.rules[ruleNonindentSpace]() {
				goto l111
			}
			{
				if position == len(p.Buffer) {
					goto l111
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				case '*':
					position++ // matchChar
					break
				case '+':
					position++ // matchChar
					break
				default:
					goto l111
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l111
			}
		l114:
			if !p.rules[ruleSpacechar]() {
				goto l115
			}
			goto l114
		l115:
			return true
		l111:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 21 BulletList <- (&Bullet (ListTight / ListLoose) { yy.key = BULLETLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position117 := position
				if !p.rules[ruleBullet]() {
					goto l116
				}
				position = position117
			}
			if !p.rules[ruleListTight]() {
				goto l119
			}
			goto l118
		l119:
			if !p.rules[ruleListLoose]() {
				goto l116
			}
		l118:
			do(23)
			return true
		l116:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 22 ListTight <- (StartList (ListItemTight { a = cons(yy, a) })+ BlankLine* !((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l120
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItemTight]() {
				goto l120
			}
			do(24)
		l121:
			{
				position122, thunkPosition122 := position, thunkPosition
				if !p.rules[ruleListItemTight]() {
					goto l122
				}
				do(24)
				goto l121
			l122:
				position, thunkPosition = position122, thunkPosition122
			}
		l123:
			if !p.rules[ruleBlankLine]() {
				goto l124
			}
			goto l123
		l124:
			{
				if position == len(p.Buffer) {
					goto l125
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l125
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l125
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l125
					}
				}
			}
			goto l120
		l125:
			do(25)
			doarg(yyPop, 1)
			return true
		l120:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 23 ListLoose <- (StartList (ListItem BlankLine* {
		    li := b.children
		    li.contents.str += "\n\n"
		    a = cons(b, a)
		})+ { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l127
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItem]() {
				goto l127
			}
			doarg(yySet, -2)
		l130:
			if !p.rules[ruleBlankLine]() {
				goto l131
			}
			goto l130
		l131:
			do(26)
		l128:
			{
				position129, thunkPosition129 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto l129
				}
				doarg(yySet, -2)
			l132:
				if !p.rules[ruleBlankLine]() {
					goto l133
				}
				goto l132
			l133:
				do(26)
				goto l128
			l129:
				position, thunkPosition = position129, thunkPosition129
			}
			do(27)
			doarg(yyPop, 2)
			return true
		l127:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 24 ListItem <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (ListContinuationBlock { a = cons(yy, a) })* {
		   raw := p.mkStringFromList(a, false)
		   raw.key = RAW
		   yy = p.mkElem(LISTITEM)
		   yy.children = raw
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				if position == len(p.Buffer) {
					goto l134
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l134
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l134
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l134
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto l134
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l134
			}
			do(28)
		l136:
			{
				position137, thunkPosition137 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l137
				}
				do(29)
				goto l136
			l137:
				position, thunkPosition = position137, thunkPosition137
			}
			do(30)
			doarg(yyPop, 1)
			return true
		l134:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 25 ListItemTight <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (!BlankLine ListContinuationBlock { a = cons(yy, a) })* !ListContinuationBlock {
		   raw := p.mkStringFromList(a, false)
		   raw.key = RAW
		   yy = p.mkElem(LISTITEM)
		   yy.children = raw
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				if position == len(p.Buffer) {
					goto l138
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l138
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l138
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l138
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto l138
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l138
			}
			do(31)
		l140:
			{
				position141, thunkPosition141 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l142
				}
				goto l141
			l142:
				if !p.rules[ruleListContinuationBlock]() {
					goto l141
				}
				do(32)
				goto l140
			l141:
				position, thunkPosition = position141, thunkPosition141
			}
			if !p.rules[ruleListContinuationBlock]() {
				goto l143
			}
			goto l138
		l143:
			do(33)
			doarg(yyPop, 1)
			return true
		l138:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 26 ListBlock <- (StartList !BlankLine Line { a = cons(yy, a) } (ListBlockLine { a = cons(yy, a) })* { yy = p.mkStringFromList(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l144
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l145
			}
			goto l144
		l145:
			if !p.rules[ruleLine]() {
				goto l144
			}
			do(34)
		l146:
			{
				position147 := position
				if !p.rules[ruleListBlockLine]() {
					goto l147
				}
				do(35)
				goto l146
			l147:
				position = position147
			}
			do(36)
			doarg(yyPop, 1)
			return true
		l144:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 27 ListContinuationBlock <- (StartList (< BlankLine* > {   if len(yytext) == 0 {
		         a = cons(p.mkString("\001"), a) // block separator
		    } else {
		         a = cons(p.mkString(yytext), a)
		    }
		}) (Indent ListBlock { a = cons(yy, a) })+ {  yy = p.mkStringFromList(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l148
			}
			doarg(yySet, -1)
			begin = position
		l149:
			if !p.rules[ruleBlankLine]() {
				goto l150
			}
			goto l149
		l150:
			end = position
			do(37)
			if !p.rules[ruleIndent]() {
				goto l148
			}
			if !p.rules[ruleListBlock]() {
				goto l148
			}
			do(38)
		l151:
			{
				position152, thunkPosition152 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l152
				}
				if !p.rules[ruleListBlock]() {
					goto l152
				}
				do(38)
				goto l151
			l152:
				position, thunkPosition = position152, thunkPosition152
			}
			do(39)
			doarg(yyPop, 1)
			return true
		l148:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 28 Enumerator <- (NonindentSpace [0-9]+ '.' Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l153
			}
			if !matchClass(0) {
				goto l153
			}
		l154:
			if !matchClass(0) {
				goto l155
			}
			goto l154
		l155:
			if !matchChar('.') {
				goto l153
			}
			if !p.rules[ruleSpacechar]() {
				goto l153
			}
		l156:
			if !p.rules[ruleSpacechar]() {
				goto l157
			}
			goto l156
		l157:
			return true
		l153:
			position = position0
			return false
		},
		/* 29 OrderedList <- (&Enumerator (ListTight / ListLoose) { yy.key = ORDEREDLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position159 := position
				if !p.rules[ruleEnumerator]() {
					goto l158
				}
				position = position159
			}
			if !p.rules[ruleListTight]() {
				goto l161
			}
			goto l160
		l161:
			if !p.rules[ruleListLoose]() {
				goto l158
			}
		l160:
			do(40)
			return true
		l158:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 ListBlockLine <- (!BlankLine !((&[:~] DefMarker) | (&[\t *+\-0-9] (Indent? ((&[*+\-] Bullet) | (&[0-9] Enumerator))))) !HorizontalRule OptionallyIndentedLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlankLine]() {
				goto l163
			}
			goto l162
		l163:
			{
				position164 := position
				{
					if position == len(p.Buffer) {
						goto l164
					}
					switch p.Buffer[position] {
					case ':', '~':
						if !p.rules[ruleDefMarker]() {
							goto l164
						}
						break
					default:
						if !p.rules[ruleIndent]() {
							goto l166
						}
					l166:
						{
							if position == len(p.Buffer) {
								goto l164
							}
							switch p.Buffer[position] {
							case '*', '+', '-':
								if !p.rules[ruleBullet]() {
									goto l164
								}
								break
							default:
								if !p.rules[ruleEnumerator]() {
									goto l164
								}
							}
						}
					}
				}
				goto l162
			l164:
				position = position164
			}
			if !p.rules[ruleHorizontalRule]() {
				goto l169
			}
			goto l162
		l169:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l162
			}
			return true
		l162:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 31 HtmlBlockOpenAddress <- ('<' Spnl ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l170
			}
			if !p.rules[ruleSpnl]() {
				goto l170
			}
			{
				if position == len(p.Buffer) {
					goto l170
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l170
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l170
					}
					break
				default:
					goto l170
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l170
			}
		l172:
			if !p.rules[ruleHtmlAttribute]() {
				goto l173
			}
			goto l172
		l173:
			if !matchChar('>') {
				goto l170
			}
			return true
		l170:
			position = position0
			return false
		},
		/* 32 HtmlBlockCloseAddress <- ('<' Spnl '/' ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l174
			}
			if !p.rules[ruleSpnl]() {
				goto l174
			}
			if !matchChar('/') {
				goto l174
			}
			{
				if position == len(p.Buffer) {
					goto l174
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l174
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l174
					}
					break
				default:
					goto l174
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l174
			}
			if !matchChar('>') {
				goto l174
			}
			return true
		l174:
			position = position0
			return false
		},
		/* 33 HtmlBlockAddress <- (HtmlBlockOpenAddress (HtmlBlockAddress / (!HtmlBlockCloseAddress .))* HtmlBlockCloseAddress) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenAddress]() {
				goto l176
			}
		l177:
			{
				position178 := position
				if !p.rules[ruleHtmlBlockAddress]() {
					goto l180
				}
				goto l179
			l180:
				if !p.rules[ruleHtmlBlockCloseAddress]() {
					goto l181
				}
				goto l178
			l181:
				if !matchDot() {
					goto l178
				}
			l179:
				goto l177
			l178:
				position = position178
			}
			if !p.rules[ruleHtmlBlockCloseAddress]() {
				goto l176
			}
			return true
		l176:
			position = position0
			return false
		},
		/* 34 HtmlBlockOpenBlockquote <- ('<' Spnl ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l182
			}
			if !p.rules[ruleSpnl]() {
				goto l182
			}
			{
				if position == len(p.Buffer) {
					goto l182
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l182
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l182
					}
					break
				default:
					goto l182
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l182
			}
		l184:
			if !p.rules[ruleHtmlAttribute]() {
				goto l185
			}
			goto l184
		l185:
			if !matchChar('>') {
				goto l182
			}
			return true
		l182:
			position = position0
			return false
		},
		/* 35 HtmlBlockCloseBlockquote <- ('<' Spnl '/' ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l186
			}
			if !p.rules[ruleSpnl]() {
				goto l186
			}
			if !matchChar('/') {
				goto l186
			}
			{
				if position == len(p.Buffer) {
					goto l186
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l186
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l186
					}
					break
				default:
					goto l186
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l186
			}
			if !matchChar('>') {
				goto l186
			}
			return true
		l186:
			position = position0
			return false
		},
		/* 36 HtmlBlockBlockquote <- (HtmlBlockOpenBlockquote (HtmlBlockBlockquote / (!HtmlBlockCloseBlockquote .))* HtmlBlockCloseBlockquote) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenBlockquote]() {
				goto l188
			}
		l189:
			{
				position190 := position
				if !p.rules[ruleHtmlBlockBlockquote]() {
					goto l192
				}
				goto l191
			l192:
				if !p.rules[ruleHtmlBlockCloseBlockquote]() {
					goto l193
				}
				goto l190
			l193:
				if !matchDot() {
					goto l190
				}
			l191:
				goto l189
			l190:
				position = position190
			}
			if !p.rules[ruleHtmlBlockCloseBlockquote]() {
				goto l188
			}
			return true
		l188:
			position = position0
			return false
		},
		/* 37 HtmlBlockOpenCenter <- ('<' Spnl ((&[C] 'CENTER') | (&[c] 'center')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l194
			}
			if !p.rules[ruleSpnl]() {
				goto l194
			}
			{
				if position == len(p.Buffer) {
					goto l194
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l194
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l194
					}
					break
				default:
					goto l194
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l194
			}
		l196:
			if !p.rules[ruleHtmlAttribute]() {
				goto l197
			}
			goto l196
		l197:
			if !matchChar('>') {
				goto l194
			}
			return true
		l194:
			position = position0
			return false
		},
		/* 38 HtmlBlockCloseCenter <- ('<' Spnl '/' ((&[C] 'CENTER') | (&[c] 'center')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l198
			}
			if !p.rules[ruleSpnl]() {
				goto l198
			}
			if !matchChar('/') {
				goto l198
			}
			{
				if position == len(p.Buffer) {
					goto l198
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l198
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l198
					}
					break
				default:
					goto l198
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l198
			}
			if !matchChar('>') {
				goto l198
			}
			return true
		l198:
			position = position0
			return false
		},
		/* 39 HtmlBlockCenter <- (HtmlBlockOpenCenter (HtmlBlockCenter / (!HtmlBlockCloseCenter .))* HtmlBlockCloseCenter) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenCenter]() {
				goto l200
			}
		l201:
			{
				position202 := position
				if !p.rules[ruleHtmlBlockCenter]() {
					goto l204
				}
				goto l203
			l204:
				if !p.rules[ruleHtmlBlockCloseCenter]() {
					goto l205
				}
				goto l202
			l205:
				if !matchDot() {
					goto l202
				}
			l203:
				goto l201
			l202:
				position = position202
			}
			if !p.rules[ruleHtmlBlockCloseCenter]() {
				goto l200
			}
			return true
		l200:
			position = position0
			return false
		},
		/* 40 HtmlBlockOpenDir <- ('<' Spnl ((&[D] 'DIR') | (&[d] 'dir')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l206
			}
			if !p.rules[ruleSpnl]() {
				goto l206
			}
			{
				if position == len(p.Buffer) {
					goto l206
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto l206
					}
					break
				case 'd':
					position++
					if !matchString("ir") {
						goto l206
					}
					break
				default:
					goto l206
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l206
			}
		l208:
			if !p.rules[ruleHtmlAttribute]() {
				goto l209
			}
			goto l208
		l209:
			if !matchChar('>') {
				goto l206
			}
			return true
		l206:
			position = position0
			return false
		},
		/* 41 HtmlBlockCloseDir <- ('<' Spnl '/' ((&[D] 'DIR') | (&[d] 'dir')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l210
			}
			if !p.rules[ruleSpnl]() {
				goto l210
			}
			if !matchChar('/') {
				goto l210
			}
			{
				if position == len(p.Buffer) {
					goto l210
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto l210
					}
					break
				case 'd':
					position++
					if !matchString("ir") {
						goto l210
					}
					break
				default:
					goto l210
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l210
			}
			if !matchChar('>') {
				goto l210
			}
			return true
		l210:
			position = position0
			return false
		},
		/* 42 HtmlBlockDir <- (HtmlBlockOpenDir (HtmlBlockDir / (!HtmlBlockCloseDir .))* HtmlBlockCloseDir) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDir]() {
				goto l212
			}
		l213:
			{
				position214 := position
				if !p.rules[ruleHtmlBlockDir]() {
					goto l216
				}
				goto l215
			l216:
				if !p.rules[ruleHtmlBlockCloseDir]() {
					goto l217
				}
				goto l214
			l217:
				if !matchDot() {
					goto l214
				}
			l215:
				goto l213
			l214:
				position = position214
			}
			if !p.rules[ruleHtmlBlockCloseDir]() {
				goto l212
			}
			return true
		l212:
			position = position0
			return false
		},
		/* 43 HtmlBlockOpenDiv <- ('<' Spnl ((&[D] 'DIV') | (&[d] 'div')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l218
			}
			if !p.rules[ruleSpnl]() {
				goto l218
			}
			{
				if position == len(p.Buffer) {
					goto l218
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto l218
					}
					break
				case 'd':
					position++
					if !matchString("iv") {
						goto l218
					}
					break
				default:
					goto l218
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l218
			}
		l220:
			if !p.rules[ruleHtmlAttribute]() {
				goto l221
			}
			goto l220
		l221:
			if !matchChar('>') {
				goto l218
			}
			return true
		l218:
			position = position0
			return false
		},
		/* 44 HtmlBlockCloseDiv <- ('<' Spnl '/' ((&[D] 'DIV') | (&[d] 'div')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l222
			}
			if !p.rules[ruleSpnl]() {
				goto l222
			}
			if !matchChar('/') {
				goto l222
			}
			{
				if position == len(p.Buffer) {
					goto l222
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto l222
					}
					break
				case 'd':
					position++
					if !matchString("iv") {
						goto l222
					}
					break
				default:
					goto l222
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l222
			}
			if !matchChar('>') {
				goto l222
			}
			return true
		l222:
			position = position0
			return false
		},
		/* 45 HtmlBlockDiv <- (HtmlBlockOpenDiv (HtmlBlockDiv / (!HtmlBlockCloseDiv .))* HtmlBlockCloseDiv) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDiv]() {
				goto l224
			}
		l225:
			{
				position226 := position
				if !p.rules[ruleHtmlBlockDiv]() {
					goto l228
				}
				goto l227
			l228:
				if !p.rules[ruleHtmlBlockCloseDiv]() {
					goto l229
				}
				goto l226
			l229:
				if !matchDot() {
					goto l226
				}
			l227:
				goto l225
			l226:
				position = position226
			}
			if !p.rules[ruleHtmlBlockCloseDiv]() {
				goto l224
			}
			return true
		l224:
			position = position0
			return false
		},
		/* 46 HtmlBlockOpenDl <- ('<' Spnl ((&[D] 'DL') | (&[d] 'dl')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l230
			}
			if !p.rules[ruleSpnl]() {
				goto l230
			}
			{
				if position == len(p.Buffer) {
					goto l230
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto l230
					}
					break
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto l230
					}
					break
				default:
					goto l230
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l230
			}
		l232:
			if !p.rules[ruleHtmlAttribute]() {
				goto l233
			}
			goto l232
		l233:
			if !matchChar('>') {
				goto l230
			}
			return true
		l230:
			position = position0
			return false
		},
		/* 47 HtmlBlockCloseDl <- ('<' Spnl '/' ((&[D] 'DL') | (&[d] 'dl')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l234
			}
			if !p.rules[ruleSpnl]() {
				goto l234
			}
			if !matchChar('/') {
				goto l234
			}
			{
				if position == len(p.Buffer) {
					goto l234
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto l234
					}
					break
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto l234
					}
					break
				default:
					goto l234
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l234
			}
			if !matchChar('>') {
				goto l234
			}
			return true
		l234:
			position = position0
			return false
		},
		/* 48 HtmlBlockDl <- (HtmlBlockOpenDl (HtmlBlockDl / (!HtmlBlockCloseDl .))* HtmlBlockCloseDl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDl]() {
				goto l236
			}
		l237:
			{
				position238 := position
				if !p.rules[ruleHtmlBlockDl]() {
					goto l240
				}
				goto l239
			l240:
				if !p.rules[ruleHtmlBlockCloseDl]() {
					goto l241
				}
				goto l238
			l241:
				if !matchDot() {
					goto l238
				}
			l239:
				goto l237
			l238:
				position = position238
			}
			if !p.rules[ruleHtmlBlockCloseDl]() {
				goto l236
			}
			return true
		l236:
			position = position0
			return false
		},
		/* 49 HtmlBlockOpenFieldset <- ('<' Spnl ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l242
			}
			if !p.rules[ruleSpnl]() {
				goto l242
			}
			{
				if position == len(p.Buffer) {
					goto l242
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto l242
					}
					break
				case 'f':
					position++
					if !matchString("ieldset") {
						goto l242
					}
					break
				default:
					goto l242
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l242
			}
		l244:
			if !p.rules[ruleHtmlAttribute]() {
				goto l245
			}
			goto l244
		l245:
			if !matchChar('>') {
				goto l242
			}
			return true
		l242:
			position = position0
			return false
		},
		/* 50 HtmlBlockCloseFieldset <- ('<' Spnl '/' ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l246
			}
			if !p.rules[ruleSpnl]() {
				goto l246
			}
			if !matchChar('/') {
				goto l246
			}
			{
				if position == len(p.Buffer) {
					goto l246
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto l246
					}
					break
				case 'f':
					position++
					if !matchString("ieldset") {
						goto l246
					}
					break
				default:
					goto l246
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l246
			}
			if !matchChar('>') {
				goto l246
			}
			return true
		l246:
			position = position0
			return false
		},
		/* 51 HtmlBlockFieldset <- (HtmlBlockOpenFieldset (HtmlBlockFieldset / (!HtmlBlockCloseFieldset .))* HtmlBlockCloseFieldset) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFieldset]() {
				goto l248
			}
		l249:
			{
				position250 := position
				if !p.rules[ruleHtmlBlockFieldset]() {
					goto l252
				}
				goto l251
			l252:
				if !p.rules[ruleHtmlBlockCloseFieldset]() {
					goto l253
				}
				goto l250
			l253:
				if !matchDot() {
					goto l250
				}
			l251:
				goto l249
			l250:
				position = position250
			}
			if !p.rules[ruleHtmlBlockCloseFieldset]() {
				goto l248
			}
			return true
		l248:
			position = position0
			return false
		},
		/* 52 HtmlBlockOpenForm <- ('<' Spnl ((&[F] 'FORM') | (&[f] 'form')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l254
			}
			if !p.rules[ruleSpnl]() {
				goto l254
			}
			{
				if position == len(p.Buffer) {
					goto l254
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto l254
					}
					break
				case 'f':
					position++
					if !matchString("orm") {
						goto l254
					}
					break
				default:
					goto l254
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l254
			}
		l256:
			if !p.rules[ruleHtmlAttribute]() {
				goto l257
			}
			goto l256
		l257:
			if !matchChar('>') {
				goto l254
			}
			return true
		l254:
			position = position0
			return false
		},
		/* 53 HtmlBlockCloseForm <- ('<' Spnl '/' ((&[F] 'FORM') | (&[f] 'form')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l258
			}
			if !p.rules[ruleSpnl]() {
				goto l258
			}
			if !matchChar('/') {
				goto l258
			}
			{
				if position == len(p.Buffer) {
					goto l258
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto l258
					}
					break
				case 'f':
					position++
					if !matchString("orm") {
						goto l258
					}
					break
				default:
					goto l258
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l258
			}
			if !matchChar('>') {
				goto l258
			}
			return true
		l258:
			position = position0
			return false
		},
		/* 54 HtmlBlockForm <- (HtmlBlockOpenForm (HtmlBlockForm / (!HtmlBlockCloseForm .))* HtmlBlockCloseForm) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenForm]() {
				goto l260
			}
		l261:
			{
				position262 := position
				if !p.rules[ruleHtmlBlockForm]() {
					goto l264
				}
				goto l263
			l264:
				if !p.rules[ruleHtmlBlockCloseForm]() {
					goto l265
				}
				goto l262
			l265:
				if !matchDot() {
					goto l262
				}
			l263:
				goto l261
			l262:
				position = position262
			}
			if !p.rules[ruleHtmlBlockCloseForm]() {
				goto l260
			}
			return true
		l260:
			position = position0
			return false
		},
		/* 55 HtmlBlockOpenH1 <- ('<' Spnl ((&[H] 'H1') | (&[h] 'h1')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l266
			}
			if !p.rules[ruleSpnl]() {
				goto l266
			}
			{
				if position == len(p.Buffer) {
					goto l266
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto l266
					}
					break
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto l266
					}
					break
				default:
					goto l266
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l266
			}
		l268:
			if !p.rules[ruleHtmlAttribute]() {
				goto l269
			}
			goto l268
		l269:
			if !matchChar('>') {
				goto l266
			}
			return true
		l266:
			position = position0
			return false
		},
		/* 56 HtmlBlockCloseH1 <- ('<' Spnl '/' ((&[H] 'H1') | (&[h] 'h1')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l270
			}
			if !p.rules[ruleSpnl]() {
				goto l270
			}
			if !matchChar('/') {
				goto l270
			}
			{
				if position == len(p.Buffer) {
					goto l270
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto l270
					}
					break
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto l270
					}
					break
				default:
					goto l270
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l270
			}
			if !matchChar('>') {
				goto l270
			}
			return true
		l270:
			position = position0
			return false
		},
		/* 57 HtmlBlockH1 <- (HtmlBlockOpenH1 (HtmlBlockH1 / (!HtmlBlockCloseH1 .))* HtmlBlockCloseH1) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH1]() {
				goto l272
			}
		l273:
			{
				position274 := position
				if !p.rules[ruleHtmlBlockH1]() {
					goto l276
				}
				goto l275
			l276:
				if !p.rules[ruleHtmlBlockCloseH1]() {
					goto l277
				}
				goto l274
			l277:
				if !matchDot() {
					goto l274
				}
			l275:
				goto l273
			l274:
				position = position274
			}
			if !p.rules[ruleHtmlBlockCloseH1]() {
				goto l272
			}
			return true
		l272:
			position = position0
			return false
		},
		/* 58 HtmlBlockOpenH2 <- ('<' Spnl ((&[H] 'H2') | (&[h] 'h2')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l278
			}
			if !p.rules[ruleSpnl]() {
				goto l278
			}
			{
				if position == len(p.Buffer) {
					goto l278
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto l278
					}
					break
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto l278
					}
					break
				default:
					goto l278
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l278
			}
		l280:
			if !p.rules[ruleHtmlAttribute]() {
				goto l281
			}
			goto l280
		l281:
			if !matchChar('>') {
				goto l278
			}
			return true
		l278:
			position = position0
			return false
		},
		/* 59 HtmlBlockCloseH2 <- ('<' Spnl '/' ((&[H] 'H2') | (&[h] 'h2')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l282
			}
			if !p.rules[ruleSpnl]() {
				goto l282
			}
			if !matchChar('/') {
				goto l282
			}
			{
				if position == len(p.Buffer) {
					goto l282
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto l282
					}
					break
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto l282
					}
					break
				default:
					goto l282
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l282
			}
			if !matchChar('>') {
				goto l282
			}
			return true
		l282:
			position = position0
			return false
		},
		/* 60 HtmlBlockH2 <- (HtmlBlockOpenH2 (HtmlBlockH2 / (!HtmlBlockCloseH2 .))* HtmlBlockCloseH2) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH2]() {
				goto l284
			}
		l285:
			{
				position286 := position
				if !p.rules[ruleHtmlBlockH2]() {
					goto l288
				}
				goto l287
			l288:
				if !p.rules[ruleHtmlBlockCloseH2]() {
					goto l289
				}
				goto l286
			l289:
				if !matchDot() {
					goto l286
				}
			l287:
				goto l285
			l286:
				position = position286
			}
			if !p.rules[ruleHtmlBlockCloseH2]() {
				goto l284
			}
			return true
		l284:
			position = position0
			return false
		},
		/* 61 HtmlBlockOpenH3 <- ('<' Spnl ((&[H] 'H3') | (&[h] 'h3')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l290
			}
			if !p.rules[ruleSpnl]() {
				goto l290
			}
			{
				if position == len(p.Buffer) {
					goto l290
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto l290
					}
					break
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto l290
					}
					break
				default:
					goto l290
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l290
			}
		l292:
			if !p.rules[ruleHtmlAttribute]() {
				goto l293
			}
			goto l292
		l293:
			if !matchChar('>') {
				goto l290
			}
			return true
		l290:
			position = position0
			return false
		},
		/* 62 HtmlBlockCloseH3 <- ('<' Spnl '/' ((&[H] 'H3') | (&[h] 'h3')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l294
			}
			if !p.rules[ruleSpnl]() {
				goto l294
			}
			if !matchChar('/') {
				goto l294
			}
			{
				if position == len(p.Buffer) {
					goto l294
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto l294
					}
					break
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto l294
					}
					break
				default:
					goto l294
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l294
			}
			if !matchChar('>') {
				goto l294
			}
			return true
		l294:
			position = position0
			return false
		},
		/* 63 HtmlBlockH3 <- (HtmlBlockOpenH3 (HtmlBlockH3 / (!HtmlBlockCloseH3 .))* HtmlBlockCloseH3) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH3]() {
				goto l296
			}
		l297:
			{
				position298 := position
				if !p.rules[ruleHtmlBlockH3]() {
					goto l300
				}
				goto l299
			l300:
				if !p.rules[ruleHtmlBlockCloseH3]() {
					goto l301
				}
				goto l298
			l301:
				if !matchDot() {
					goto l298
				}
			l299:
				goto l297
			l298:
				position = position298
			}
			if !p.rules[ruleHtmlBlockCloseH3]() {
				goto l296
			}
			return true
		l296:
			position = position0
			return false
		},
		/* 64 HtmlBlockOpenH4 <- ('<' Spnl ((&[H] 'H4') | (&[h] 'h4')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l302
			}
			if !p.rules[ruleSpnl]() {
				goto l302
			}
			{
				if position == len(p.Buffer) {
					goto l302
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto l302
					}
					break
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto l302
					}
					break
				default:
					goto l302
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l302
			}
		l304:
			if !p.rules[ruleHtmlAttribute]() {
				goto l305
			}
			goto l304
		l305:
			if !matchChar('>') {
				goto l302
			}
			return true
		l302:
			position = position0
			return false
		},
		/* 65 HtmlBlockCloseH4 <- ('<' Spnl '/' ((&[H] 'H4') | (&[h] 'h4')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l306
			}
			if !p.rules[ruleSpnl]() {
				goto l306
			}
			if !matchChar('/') {
				goto l306
			}
			{
				if position == len(p.Buffer) {
					goto l306
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto l306
					}
					break
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto l306
					}
					break
				default:
					goto l306
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l306
			}
			if !matchChar('>') {
				goto l306
			}
			return true
		l306:
			position = position0
			return false
		},
		/* 66 HtmlBlockH4 <- (HtmlBlockOpenH4 (HtmlBlockH4 / (!HtmlBlockCloseH4 .))* HtmlBlockCloseH4) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH4]() {
				goto l308
			}
		l309:
			{
				position310 := position
				if !p.rules[ruleHtmlBlockH4]() {
					goto l312
				}
				goto l311
			l312:
				if !p.rules[ruleHtmlBlockCloseH4]() {
					goto l313
				}
				goto l310
			l313:
				if !matchDot() {
					goto l310
				}
			l311:
				goto l309
			l310:
				position = position310
			}
			if !p.rules[ruleHtmlBlockCloseH4]() {
				goto l308
			}
			return true
		l308:
			position = position0
			return false
		},
		/* 67 HtmlBlockOpenH5 <- ('<' Spnl ((&[H] 'H5') | (&[h] 'h5')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l314
			}
			if !p.rules[ruleSpnl]() {
				goto l314
			}
			{
				if position == len(p.Buffer) {
					goto l314
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto l314
					}
					break
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto l314
					}
					break
				default:
					goto l314
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l314
			}
		l316:
			if !p.rules[ruleHtmlAttribute]() {
				goto l317
			}
			goto l316
		l317:
			if !matchChar('>') {
				goto l314
			}
			return true
		l314:
			position = position0
			return false
		},
		/* 68 HtmlBlockCloseH5 <- ('<' Spnl '/' ((&[H] 'H5') | (&[h] 'h5')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l318
			}
			if !p.rules[ruleSpnl]() {
				goto l318
			}
			if !matchChar('/') {
				goto l318
			}
			{
				if position == len(p.Buffer) {
					goto l318
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto l318
					}
					break
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto l318
					}
					break
				default:
					goto l318
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l318
			}
			if !matchChar('>') {
				goto l318
			}
			return true
		l318:
			position = position0
			return false
		},
		/* 69 HtmlBlockH5 <- (HtmlBlockOpenH5 (HtmlBlockH5 / (!HtmlBlockCloseH5 .))* HtmlBlockCloseH5) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH5]() {
				goto l320
			}
		l321:
			{
				position322 := position
				if !p.rules[ruleHtmlBlockH5]() {
					goto l324
				}
				goto l323
			l324:
				if !p.rules[ruleHtmlBlockCloseH5]() {
					goto l325
				}
				goto l322
			l325:
				if !matchDot() {
					goto l322
				}
			l323:
				goto l321
			l322:
				position = position322
			}
			if !p.rules[ruleHtmlBlockCloseH5]() {
				goto l320
			}
			return true
		l320:
			position = position0
			return false
		},
		/* 70 HtmlBlockOpenH6 <- ('<' Spnl ((&[H] 'H6') | (&[h] 'h6')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l326
			}
			if !p.rules[ruleSpnl]() {
				goto l326
			}
			{
				if position == len(p.Buffer) {
					goto l326
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto l326
					}
					break
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto l326
					}
					break
				default:
					goto l326
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l326
			}
		l328:
			if !p.rules[ruleHtmlAttribute]() {
				goto l329
			}
			goto l328
		l329:
			if !matchChar('>') {
				goto l326
			}
			return true
		l326:
			position = position0
			return false
		},
		/* 71 HtmlBlockCloseH6 <- ('<' Spnl '/' ((&[H] 'H6') | (&[h] 'h6')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l330
			}
			if !p.rules[ruleSpnl]() {
				goto l330
			}
			if !matchChar('/') {
				goto l330
			}
			{
				if position == len(p.Buffer) {
					goto l330
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto l330
					}
					break
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto l330
					}
					break
				default:
					goto l330
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l330
			}
			if !matchChar('>') {
				goto l330
			}
			return true
		l330:
			position = position0
			return false
		},
		/* 72 HtmlBlockH6 <- (HtmlBlockOpenH6 (HtmlBlockH6 / (!HtmlBlockCloseH6 .))* HtmlBlockCloseH6) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH6]() {
				goto l332
			}
		l333:
			{
				position334 := position
				if !p.rules[ruleHtmlBlockH6]() {
					goto l336
				}
				goto l335
			l336:
				if !p.rules[ruleHtmlBlockCloseH6]() {
					goto l337
				}
				goto l334
			l337:
				if !matchDot() {
					goto l334
				}
			l335:
				goto l333
			l334:
				position = position334
			}
			if !p.rules[ruleHtmlBlockCloseH6]() {
				goto l332
			}
			return true
		l332:
			position = position0
			return false
		},
		/* 73 HtmlBlockOpenMenu <- ('<' Spnl ((&[M] 'MENU') | (&[m] 'menu')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l338
			}
			if !p.rules[ruleSpnl]() {
				goto l338
			}
			{
				if position == len(p.Buffer) {
					goto l338
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto l338
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l338
					}
					break
				default:
					goto l338
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l338
			}
		l340:
			if !p.rules[ruleHtmlAttribute]() {
				goto l341
			}
			goto l340
		l341:
			if !matchChar('>') {
				goto l338
			}
			return true
		l338:
			position = position0
			return false
		},
		/* 74 HtmlBlockCloseMenu <- ('<' Spnl '/' ((&[M] 'MENU') | (&[m] 'menu')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l342
			}
			if !p.rules[ruleSpnl]() {
				goto l342
			}
			if !matchChar('/') {
				goto l342
			}
			{
				if position == len(p.Buffer) {
					goto l342
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto l342
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l342
					}
					break
				default:
					goto l342
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l342
			}
			if !matchChar('>') {
				goto l342
			}
			return true
		l342:
			position = position0
			return false
		},
		/* 75 HtmlBlockMenu <- (HtmlBlockOpenMenu (HtmlBlockMenu / (!HtmlBlockCloseMenu .))* HtmlBlockCloseMenu) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenMenu]() {
				goto l344
			}
		l345:
			{
				position346 := position
				if !p.rules[ruleHtmlBlockMenu]() {
					goto l348
				}
				goto l347
			l348:
				if !p.rules[ruleHtmlBlockCloseMenu]() {
					goto l349
				}
				goto l346
			l349:
				if !matchDot() {
					goto l346
				}
			l347:
				goto l345
			l346:
				position = position346
			}
			if !p.rules[ruleHtmlBlockCloseMenu]() {
				goto l344
			}
			return true
		l344:
			position = position0
			return false
		},
		/* 76 HtmlBlockOpenNoframes <- ('<' Spnl ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l350
			}
			if !p.rules[ruleSpnl]() {
				goto l350
			}
			{
				if position == len(p.Buffer) {
					goto l350
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto l350
					}
					break
				case 'n':
					position++
					if !matchString("oframes") {
						goto l350
					}
					break
				default:
					goto l350
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l350
			}
		l352:
			if !p.rules[ruleHtmlAttribute]() {
				goto l353
			}
			goto l352
		l353:
			if !matchChar('>') {
				goto l350
			}
			return true
		l350:
			position = position0
			return false
		},
		/* 77 HtmlBlockCloseNoframes <- ('<' Spnl '/' ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l354
			}
			if !p.rules[ruleSpnl]() {
				goto l354
			}
			if !matchChar('/') {
				goto l354
			}
			{
				if position == len(p.Buffer) {
					goto l354
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto l354
					}
					break
				case 'n':
					position++
					if !matchString("oframes") {
						goto l354
					}
					break
				default:
					goto l354
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l354
			}
			if !matchChar('>') {
				goto l354
			}
			return true
		l354:
			position = position0
			return false
		},
		/* 78 HtmlBlockNoframes <- (HtmlBlockOpenNoframes (HtmlBlockNoframes / (!HtmlBlockCloseNoframes .))* HtmlBlockCloseNoframes) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoframes]() {
				goto l356
			}
		l357:
			{
				position358 := position
				if !p.rules[ruleHtmlBlockNoframes]() {
					goto l360
				}
				goto l359
			l360:
				if !p.rules[ruleHtmlBlockCloseNoframes]() {
					goto l361
				}
				goto l358
			l361:
				if !matchDot() {
					goto l358
				}
			l359:
				goto l357
			l358:
				position = position358
			}
			if !p.rules[ruleHtmlBlockCloseNoframes]() {
				goto l356
			}
			return true
		l356:
			position = position0
			return false
		},
		/* 79 HtmlBlockOpenNoscript <- ('<' Spnl ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l362
			}
			if !p.rules[ruleSpnl]() {
				goto l362
			}
			{
				if position == len(p.Buffer) {
					goto l362
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l362
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l362
					}
					break
				default:
					goto l362
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l362
			}
		l364:
			if !p.rules[ruleHtmlAttribute]() {
				goto l365
			}
			goto l364
		l365:
			if !matchChar('>') {
				goto l362
			}
			return true
		l362:
			position = position0
			return false
		},
		/* 80 HtmlBlockCloseNoscript <- ('<' Spnl '/' ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l366
			}
			if !p.rules[ruleSpnl]() {
				goto l366
			}
			if !matchChar('/') {
				goto l366
			}
			{
				if position == len(p.Buffer) {
					goto l366
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l366
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l366
					}
					break
				default:
					goto l366
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l366
			}
			if !matchChar('>') {
				goto l366
			}
			return true
		l366:
			position = position0
			return false
		},
		/* 81 HtmlBlockNoscript <- (HtmlBlockOpenNoscript (HtmlBlockNoscript / (!HtmlBlockCloseNoscript .))* HtmlBlockCloseNoscript) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoscript]() {
				goto l368
			}
		l369:
			{
				position370 := position
				if !p.rules[ruleHtmlBlockNoscript]() {
					goto l372
				}
				goto l371
			l372:
				if !p.rules[ruleHtmlBlockCloseNoscript]() {
					goto l373
				}
				goto l370
			l373:
				if !matchDot() {
					goto l370
				}
			l371:
				goto l369
			l370:
				position = position370
			}
			if !p.rules[ruleHtmlBlockCloseNoscript]() {
				goto l368
			}
			return true
		l368:
			position = position0
			return false
		},
		/* 82 HtmlBlockOpenOl <- ('<' Spnl ((&[O] 'OL') | (&[o] 'ol')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l374
			}
			if !p.rules[ruleSpnl]() {
				goto l374
			}
			{
				if position == len(p.Buffer) {
					goto l374
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l374
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l374
					}
					break
				default:
					goto l374
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l374
			}
		l376:
			if !p.rules[ruleHtmlAttribute]() {
				goto l377
			}
			goto l376
		l377:
			if !matchChar('>') {
				goto l374
			}
			return true
		l374:
			position = position0
			return false
		},
		/* 83 HtmlBlockCloseOl <- ('<' Spnl '/' ((&[O] 'OL') | (&[o] 'ol')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l378
			}
			if !p.rules[ruleSpnl]() {
				goto l378
			}
			if !matchChar('/') {
				goto l378
			}
			{
				if position == len(p.Buffer) {
					goto l378
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l378
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l378
					}
					break
				default:
					goto l378
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l378
			}
			if !matchChar('>') {
				goto l378
			}
			return true
		l378:
			position = position0
			return false
		},
		/* 84 HtmlBlockOl <- (HtmlBlockOpenOl (HtmlBlockOl / (!HtmlBlockCloseOl .))* HtmlBlockCloseOl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenOl]() {
				goto l380
			}
		l381:
			{
				position382 := position
				if !p.rules[ruleHtmlBlockOl]() {
					goto l384
				}
				goto l383
			l384:
				if !p.rules[ruleHtmlBlockCloseOl]() {
					goto l385
				}
				goto l382
			l385:
				if !matchDot() {
					goto l382
				}
			l383:
				goto l381
			l382:
				position = position382
			}
			if !p.rules[ruleHtmlBlockCloseOl]() {
				goto l380
			}
			return true
		l380:
			position = position0
			return false
		},
		/* 85 HtmlBlockOpenP <- ('<' Spnl ((&[P] 'P') | (&[p] 'p')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l386
			}
			if !p.rules[ruleSpnl]() {
				goto l386
			}
			{
				if position == len(p.Buffer) {
					goto l386
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
					break
				case 'p':
					position++ // matchChar
					break
				default:
					goto l386
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l386
			}
		l388:
			if !p.rules[ruleHtmlAttribute]() {
				goto l389
			}
			goto l388
		l389:
			if !matchChar('>') {
				goto l386
			}
			return true
		l386:
			position = position0
			return false
		},
		/* 86 HtmlBlockCloseP <- ('<' Spnl '/' ((&[P] 'P') | (&[p] 'p')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l390
			}
			if !p.rules[ruleSpnl]() {
				goto l390
			}
			if !matchChar('/') {
				goto l390
			}
			{
				if position == len(p.Buffer) {
					goto l390
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
					break
				case 'p':
					position++ // matchChar
					break
				default:
					goto l390
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l390
			}
			if !matchChar('>') {
				goto l390
			}
			return true
		l390:
			position = position0
			return false
		},
		/* 87 HtmlBlockP <- (HtmlBlockOpenP (HtmlBlockP / (!HtmlBlockCloseP .))* HtmlBlockCloseP) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenP]() {
				goto l392
			}
		l393:
			{
				position394 := position
				if !p.rules[ruleHtmlBlockP]() {
					goto l396
				}
				goto l395
			l396:
				if !p.rules[ruleHtmlBlockCloseP]() {
					goto l397
				}
				goto l394
			l397:
				if !matchDot() {
					goto l394
				}
			l395:
				goto l393
			l394:
				position = position394
			}
			if !p.rules[ruleHtmlBlockCloseP]() {
				goto l392
			}
			return true
		l392:
			position = position0
			return false
		},
		/* 88 HtmlBlockOpenPre <- ('<' Spnl ((&[P] 'PRE') | (&[p] 'pre')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l398
			}
			if !p.rules[ruleSpnl]() {
				goto l398
			}
			{
				if position == len(p.Buffer) {
					goto l398
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto l398
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l398
					}
					break
				default:
					goto l398
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l398
			}
		l400:
			if !p.rules[ruleHtmlAttribute]() {
				goto l401
			}
			goto l400
		l401:
			if !matchChar('>') {
				goto l398
			}
			return true
		l398:
			position = position0
			return false
		},
		/* 89 HtmlBlockClosePre <- ('<' Spnl '/' ((&[P] 'PRE') | (&[p] 'pre')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l402
			}
			if !p.rules[ruleSpnl]() {
				goto l402
			}
			if !matchChar('/') {
				goto l402
			}
			{
				if position == len(p.Buffer) {
					goto l402
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto l402
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l402
					}
					break
				default:
					goto l402
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l402
			}
			if !matchChar('>') {
				goto l402
			}
			return true
		l402:
			position = position0
			return false
		},
		/* 90 HtmlBlockPre <- (HtmlBlockOpenPre (HtmlBlockPre / (!HtmlBlockClosePre .))* HtmlBlockClosePre) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenPre]() {
				goto l404
			}
		l405:
			{
				position406 := position
				if !p.rules[ruleHtmlBlockPre]() {
					goto l408
				}
				goto l407
			l408:
				if !p.rules[ruleHtmlBlockClosePre]() {
					goto l409
				}
				goto l406
			l409:
				if !matchDot() {
					goto l406
				}
			l407:
				goto l405
			l406:
				position = position406
			}
			if !p.rules[ruleHtmlBlockClosePre]() {
				goto l404
			}
			return true
		l404:
			position = position0
			return false
		},
		/* 91 HtmlBlockOpenTable <- ('<' Spnl ((&[T] 'TABLE') | (&[t] 'table')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l410
			}
			if !p.rules[ruleSpnl]() {
				goto l410
			}
			{
				if position == len(p.Buffer) {
					goto l410
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto l410
					}
					break
				case 't':
					position++
					if !matchString("able") {
						goto l410
					}
					break
				default:
					goto l410
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l410
			}
		l412:
			if !p.rules[ruleHtmlAttribute]() {
				goto l413
			}
			goto l412
		l413:
			if !matchChar('>') {
				goto l410
			}
			return true
		l410:
			position = position0
			return false
		},
		/* 92 HtmlBlockCloseTable <- ('<' Spnl '/' ((&[T] 'TABLE') | (&[t] 'table')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l414
			}
			if !p.rules[ruleSpnl]() {
				goto l414
			}
			if !matchChar('/') {
				goto l414
			}
			{
				if position == len(p.Buffer) {
					goto l414
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto l414
					}
					break
				case 't':
					position++
					if !matchString("able") {
						goto l414
					}
					break
				default:
					goto l414
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l414
			}
			if !matchChar('>') {
				goto l414
			}
			return true
		l414:
			position = position0
			return false
		},
		/* 93 HtmlBlockTable <- (HtmlBlockOpenTable (HtmlBlockTable / (!HtmlBlockCloseTable .))* HtmlBlockCloseTable) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTable]() {
				goto l416
			}
		l417:
			{
				position418 := position
				if !p.rules[ruleHtmlBlockTable]() {
					goto l420
				}
				goto l419
			l420:
				if !p.rules[ruleHtmlBlockCloseTable]() {
					goto l421
				}
				goto l418
			l421:
				if !matchDot() {
					goto l418
				}
			l419:
				goto l417
			l418:
				position = position418
			}
			if !p.rules[ruleHtmlBlockCloseTable]() {
				goto l416
			}
			return true
		l416:
			position = position0
			return false
		},
		/* 94 HtmlBlockOpenUl <- ('<' Spnl ((&[U] 'UL') | (&[u] 'ul')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l422
			}
			if !p.rules[ruleSpnl]() {
				goto l422
			}
			{
				if position == len(p.Buffer) {
					goto l422
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l422
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l422
					}
					break
				default:
					goto l422
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l422
			}
		l424:
			if !p.rules[ruleHtmlAttribute]() {
				goto l425
			}
			goto l424
		l425:
			if !matchChar('>') {
				goto l422
			}
			return true
		l422:
			position = position0
			return false
		},
		/* 95 HtmlBlockCloseUl <- ('<' Spnl '/' ((&[U] 'UL') | (&[u] 'ul')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l426
			}
			if !p.rules[ruleSpnl]() {
				goto l426
			}
			if !matchChar('/') {
				goto l426
			}
			{
				if position == len(p.Buffer) {
					goto l426
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l426
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l426
					}
					break
				default:
					goto l426
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l426
			}
			if !matchChar('>') {
				goto l426
			}
			return true
		l426:
			position = position0
			return false
		},
		/* 96 HtmlBlockUl <- (HtmlBlockOpenUl (HtmlBlockUl / (!HtmlBlockCloseUl .))* HtmlBlockCloseUl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenUl]() {
				goto l428
			}
		l429:
			{
				position430 := position
				if !p.rules[ruleHtmlBlockUl]() {
					goto l432
				}
				goto l431
			l432:
				if !p.rules[ruleHtmlBlockCloseUl]() {
					goto l433
				}
				goto l430
			l433:
				if !matchDot() {
					goto l430
				}
			l431:
				goto l429
			l430:
				position = position430
			}
			if !p.rules[ruleHtmlBlockCloseUl]() {
				goto l428
			}
			return true
		l428:
			position = position0
			return false
		},
		/* 97 HtmlBlockOpenDd <- ('<' Spnl ((&[D] 'DD') | (&[d] 'dd')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l434
			}
			if !p.rules[ruleSpnl]() {
				goto l434
			}
			{
				if position == len(p.Buffer) {
					goto l434
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto l434
					}
					break
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto l434
					}
					break
				default:
					goto l434
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l434
			}
		l436:
			if !p.rules[ruleHtmlAttribute]() {
				goto l437
			}
			goto l436
		l437:
			if !matchChar('>') {
				goto l434
			}
			return true
		l434:
			position = position0
			return false
		},
		/* 98 HtmlBlockCloseDd <- ('<' Spnl '/' ((&[D] 'DD') | (&[d] 'dd')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l438
			}
			if !p.rules[ruleSpnl]() {
				goto l438
			}
			if !matchChar('/') {
				goto l438
			}
			{
				if position == len(p.Buffer) {
					goto l438
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto l438
					}
					break
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto l438
					}
					break
				default:
					goto l438
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l438
			}
			if !matchChar('>') {
				goto l438
			}
			return true
		l438:
			position = position0
			return false
		},
		/* 99 HtmlBlockDd <- (HtmlBlockOpenDd (HtmlBlockDd / (!HtmlBlockCloseDd .))* HtmlBlockCloseDd) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDd]() {
				goto l440
			}
		l441:
			{
				position442 := position
				if !p.rules[ruleHtmlBlockDd]() {
					goto l444
				}
				goto l443
			l444:
				if !p.rules[ruleHtmlBlockCloseDd]() {
					goto l445
				}
				goto l442
			l445:
				if !matchDot() {
					goto l442
				}
			l443:
				goto l441
			l442:
				position = position442
			}
			if !p.rules[ruleHtmlBlockCloseDd]() {
				goto l440
			}
			return true
		l440:
			position = position0
			return false
		},
		/* 100 HtmlBlockOpenDt <- ('<' Spnl ((&[D] 'DT') | (&[d] 'dt')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l446
			}
			if !p.rules[ruleSpnl]() {
				goto l446
			}
			{
				if position == len(p.Buffer) {
					goto l446
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l446
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l446
					}
					break
				default:
					goto l446
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l446
			}
		l448:
			if !p.rules[ruleHtmlAttribute]() {
				goto l449
			}
			goto l448
		l449:
			if !matchChar('>') {
				goto l446
			}
			return true
		l446:
			position = position0
			return false
		},
		/* 101 HtmlBlockCloseDt <- ('<' Spnl '/' ((&[D] 'DT') | (&[d] 'dt')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l450
			}
			if !p.rules[ruleSpnl]() {
				goto l450
			}
			if !matchChar('/') {
				goto l450
			}
			{
				if position == len(p.Buffer) {
					goto l450
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l450
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l450
					}
					break
				default:
					goto l450
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l450
			}
			if !matchChar('>') {
				goto l450
			}
			return true
		l450:
			position = position0
			return false
		},
		/* 102 HtmlBlockDt <- (HtmlBlockOpenDt (HtmlBlockDt / (!HtmlBlockCloseDt .))* HtmlBlockCloseDt) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDt]() {
				goto l452
			}
		l453:
			{
				position454 := position
				if !p.rules[ruleHtmlBlockDt]() {
					goto l456
				}
				goto l455
			l456:
				if !p.rules[ruleHtmlBlockCloseDt]() {
					goto l457
				}
				goto l454
			l457:
				if !matchDot() {
					goto l454
				}
			l455:
				goto l453
			l454:
				position = position454
			}
			if !p.rules[ruleHtmlBlockCloseDt]() {
				goto l452
			}
			return true
		l452:
			position = position0
			return false
		},
		/* 103 HtmlBlockOpenFrameset <- ('<' Spnl ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l458
			}
			if !p.rules[ruleSpnl]() {
				goto l458
			}
			{
				if position == len(p.Buffer) {
					goto l458
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l458
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l458
					}
					break
				default:
					goto l458
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l458
			}
		l460:
			if !p.rules[ruleHtmlAttribute]() {
				goto l461
			}
			goto l460
		l461:
			if !matchChar('>') {
				goto l458
			}
			return true
		l458:
			position = position0
			return false
		},
		/* 104 HtmlBlockCloseFrameset <- ('<' Spnl '/' ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l462
			}
			if !p.rules[ruleSpnl]() {
				goto l462
			}
			if !matchChar('/') {
				goto l462
			}
			{
				if position == len(p.Buffer) {
					goto l462
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l462
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l462
					}
					break
				default:
					goto l462
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l462
			}
			if !matchChar('>') {
				goto l462
			}
			return true
		l462:
			position = position0
			return false
		},
		/* 105 HtmlBlockFrameset <- (HtmlBlockOpenFrameset (HtmlBlockFrameset / (!HtmlBlockCloseFrameset .))* HtmlBlockCloseFrameset) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFrameset]() {
				goto l464
			}
		l465:
			{
				position466 := position
				if !p.rules[ruleHtmlBlockFrameset]() {
					goto l468
				}
				goto l467
			l468:
				if !p.rules[ruleHtmlBlockCloseFrameset]() {
					goto l469
				}
				goto l466
			l469:
				if !matchDot() {
					goto l466
				}
			l467:
				goto l465
			l466:
				position = position466
			}
			if !p.rules[ruleHtmlBlockCloseFrameset]() {
				goto l464
			}
			return true
		l464:
			position = position0
			return false
		},
		/* 106 HtmlBlockOpenLi <- ('<' Spnl ((&[L] 'LI') | (&[l] 'li')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l470
			}
			if !p.rules[ruleSpnl]() {
				goto l470
			}
			{
				if position == len(p.Buffer) {
					goto l470
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l470
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l470
					}
					break
				default:
					goto l470
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l470
			}
		l472:
			if !p.rules[ruleHtmlAttribute]() {
				goto l473
			}
			goto l472
		l473:
			if !matchChar('>') {
				goto l470
			}
			return true
		l470:
			position = position0
			return false
		},
		/* 107 HtmlBlockCloseLi <- ('<' Spnl '/' ((&[L] 'LI') | (&[l] 'li')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l474
			}
			if !p.rules[ruleSpnl]() {
				goto l474
			}
			if !matchChar('/') {
				goto l474
			}
			{
				if position == len(p.Buffer) {
					goto l474
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l474
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l474
					}
					break
				default:
					goto l474
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l474
			}
			if !matchChar('>') {
				goto l474
			}
			return true
		l474:
			position = position0
			return false
		},
		/* 108 HtmlBlockLi <- (HtmlBlockOpenLi (HtmlBlockLi / (!HtmlBlockCloseLi .))* HtmlBlockCloseLi) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenLi]() {
				goto l476
			}
		l477:
			{
				position478 := position
				if !p.rules[ruleHtmlBlockLi]() {
					goto l480
				}
				goto l479
			l480:
				if !p.rules[ruleHtmlBlockCloseLi]() {
					goto l481
				}
				goto l478
			l481:
				if !matchDot() {
					goto l478
				}
			l479:
				goto l477
			l478:
				position = position478
			}
			if !p.rules[ruleHtmlBlockCloseLi]() {
				goto l476
			}
			return true
		l476:
			position = position0
			return false
		},
		/* 109 HtmlBlockOpenTbody <- ('<' Spnl ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l482
			}
			if !p.rules[ruleSpnl]() {
				goto l482
			}
			{
				if position == len(p.Buffer) {
					goto l482
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto l482
					}
					break
				case 't':
					position++
					if !matchString("body") {
						goto l482
					}
					break
				default:
					goto l482
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l482
			}
		l484:
			if !p.rules[ruleHtmlAttribute]() {
				goto l485
			}
			goto l484
		l485:
			if !matchChar('>') {
				goto l482
			}
			return true
		l482:
			position = position0
			return false
		},
		/* 110 HtmlBlockCloseTbody <- ('<' Spnl '/' ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l486
			}
			if !p.rules[ruleSpnl]() {
				goto l486
			}
			if !matchChar('/') {
				goto l486
			}
			{
				if position == len(p.Buffer) {
					goto l486
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto l486
					}
					break
				case 't':
					position++
					if !matchString("body") {
						goto l486
					}
					break
				default:
					goto l486
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l486
			}
			if !matchChar('>') {
				goto l486
			}
			return true
		l486:
			position = position0
			return false
		},
		/* 111 HtmlBlockTbody <- (HtmlBlockOpenTbody (HtmlBlockTbody / (!HtmlBlockCloseTbody .))* HtmlBlockCloseTbody) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTbody]() {
				goto l488
			}
		l489:
			{
				position490 := position
				if !p.rules[ruleHtmlBlockTbody]() {
					goto l492
				}
				goto l491
			l492:
				if !p.rules[ruleHtmlBlockCloseTbody]() {
					goto l493
				}
				goto l490
			l493:
				if !matchDot() {
					goto l490
				}
			l491:
				goto l489
			l490:
				position = position490
			}
			if !p.rules[ruleHtmlBlockCloseTbody]() {
				goto l488
			}
			return true
		l488:
			position = position0
			return false
		},
		/* 112 HtmlBlockOpenTd <- ('<' Spnl ((&[T] 'TD') | (&[t] 'td')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l494
			}
			if !p.rules[ruleSpnl]() {
				goto l494
			}
			{
				if position == len(p.Buffer) {
					goto l494
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto l494
					}
					break
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto l494
					}
					break
				default:
					goto l494
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l494
			}
		l496:
			if !p.rules[ruleHtmlAttribute]() {
				goto l497
			}
			goto l496
		l497:
			if !matchChar('>') {
				goto l494
			}
			return true
		l494:
			position = position0
			return false
		},
		/* 113 HtmlBlockCloseTd <- ('<' Spnl '/' ((&[T] 'TD') | (&[t] 'td')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l498
			}
			if !p.rules[ruleSpnl]() {
				goto l498
			}
			if !matchChar('/') {
				goto l498
			}
			{
				if position == len(p.Buffer) {
					goto l498
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto l498
					}
					break
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto l498
					}
					break
				default:
					goto l498
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l498
			}
			if !matchChar('>') {
				goto l498
			}
			return true
		l498:
			position = position0
			return false
		},
		/* 114 HtmlBlockTd <- (HtmlBlockOpenTd (HtmlBlockTd / (!HtmlBlockCloseTd .))* HtmlBlockCloseTd) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTd]() {
				goto l500
			}
		l501:
			{
				position502 := position
				if !p.rules[ruleHtmlBlockTd]() {
					goto l504
				}
				goto l503
			l504:
				if !p.rules[ruleHtmlBlockCloseTd]() {
					goto l505
				}
				goto l502
			l505:
				if !matchDot() {
					goto l502
				}
			l503:
				goto l501
			l502:
				position = position502
			}
			if !p.rules[ruleHtmlBlockCloseTd]() {
				goto l500
			}
			return true
		l500:
			position = position0
			return false
		},
		/* 115 HtmlBlockOpenTfoot <- ('<' Spnl ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l506
			}
			if !p.rules[ruleSpnl]() {
				goto l506
			}
			{
				if position == len(p.Buffer) {
					goto l506
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto l506
					}
					break
				case 't':
					position++
					if !matchString("foot") {
						goto l506
					}
					break
				default:
					goto l506
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l506
			}
		l508:
			if !p.rules[ruleHtmlAttribute]() {
				goto l509
			}
			goto l508
		l509:
			if !matchChar('>') {
				goto l506
			}
			return true
		l506:
			position = position0
			return false
		},
		/* 116 HtmlBlockCloseTfoot <- ('<' Spnl '/' ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l510
			}
			if !p.rules[ruleSpnl]() {
				goto l510
			}
			if !matchChar('/') {
				goto l510
			}
			{
				if position == len(p.Buffer) {
					goto l510
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto l510
					}
					break
				case 't':
					position++
					if !matchString("foot") {
						goto l510
					}
					break
				default:
					goto l510
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l510
			}
			if !matchChar('>') {
				goto l510
			}
			return true
		l510:
			position = position0
			return false
		},
		/* 117 HtmlBlockTfoot <- (HtmlBlockOpenTfoot (HtmlBlockTfoot / (!HtmlBlockCloseTfoot .))* HtmlBlockCloseTfoot) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTfoot]() {
				goto l512
			}
		l513:
			{
				position514 := position
				if !p.rules[ruleHtmlBlockTfoot]() {
					goto l516
				}
				goto l515
			l516:
				if !p.rules[ruleHtmlBlockCloseTfoot]() {
					goto l517
				}
				goto l514
			l517:
				if !matchDot() {
					goto l514
				}
			l515:
				goto l513
			l514:
				position = position514
			}
			if !p.rules[ruleHtmlBlockCloseTfoot]() {
				goto l512
			}
			return true
		l512:
			position = position0
			return false
		},
		/* 118 HtmlBlockOpenTh <- ('<' Spnl ((&[T] 'TH') | (&[t] 'th')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l518
			}
			if !p.rules[ruleSpnl]() {
				goto l518
			}
			{
				if position == len(p.Buffer) {
					goto l518
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto l518
					}
					break
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto l518
					}
					break
				default:
					goto l518
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l518
			}
		l520:
			if !p.rules[ruleHtmlAttribute]() {
				goto l521
			}
			goto l520
		l521:
			if !matchChar('>') {
				goto l518
			}
			return true
		l518:
			position = position0
			return false
		},
		/* 119 HtmlBlockCloseTh <- ('<' Spnl '/' ((&[T] 'TH') | (&[t] 'th')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l522
			}
			if !p.rules[ruleSpnl]() {
				goto l522
			}
			if !matchChar('/') {
				goto l522
			}
			{
				if position == len(p.Buffer) {
					goto l522
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto l522
					}
					break
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto l522
					}
					break
				default:
					goto l522
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l522
			}
			if !matchChar('>') {
				goto l522
			}
			return true
		l522:
			position = position0
			return false
		},
		/* 120 HtmlBlockTh <- (HtmlBlockOpenTh (HtmlBlockTh / (!HtmlBlockCloseTh .))* HtmlBlockCloseTh) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTh]() {
				goto l524
			}
		l525:
			{
				position526 := position
				if !p.rules[ruleHtmlBlockTh]() {
					goto l528
				}
				goto l527
			l528:
				if !p.rules[ruleHtmlBlockCloseTh]() {
					goto l529
				}
				goto l526
			l529:
				if !matchDot() {
					goto l526
				}
			l527:
				goto l525
			l526:
				position = position526
			}
			if !p.rules[ruleHtmlBlockCloseTh]() {
				goto l524
			}
			return true
		l524:
			position = position0
			return false
		},
		/* 121 HtmlBlockOpenThead <- ('<' Spnl ((&[T] 'THEAD') | (&[t] 'thead')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l530
			}
			if !p.rules[ruleSpnl]() {
				goto l530
			}
			{
				if position == len(p.Buffer) {
					goto l530
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto l530
					}
					break
				case 't':
					position++
					if !matchString("head") {
						goto l530
					}
					break
				default:
					goto l530
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l530
			}
		l532:
			if !p.rules[ruleHtmlAttribute]() {
				goto l533
			}
			goto l532
		l533:
			if !matchChar('>') {
				goto l530
			}
			return true
		l530:
			position = position0
			return false
		},
		/* 122 HtmlBlockCloseThead <- ('<' Spnl '/' ((&[T] 'THEAD') | (&[t] 'thead')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l534
			}
			if !p.rules[ruleSpnl]() {
				goto l534
			}
			if !matchChar('/') {
				goto l534
			}
			{
				if position == len(p.Buffer) {
					goto l534
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto l534
					}
					break
				case 't':
					position++
					if !matchString("head") {
						goto l534
					}
					break
				default:
					goto l534
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l534
			}
			if !matchChar('>') {
				goto l534
			}
			return true
		l534:
			position = position0
			return false
		},
		/* 123 HtmlBlockThead <- (HtmlBlockOpenThead (HtmlBlockThead / (!HtmlBlockCloseThead .))* HtmlBlockCloseThead) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenThead]() {
				goto l536
			}
		l537:
			{
				position538 := position
				if !p.rules[ruleHtmlBlockThead]() {
					goto l540
				}
				goto l539
			l540:
				if !p.rules[ruleHtmlBlockCloseThead]() {
					goto l541
				}
				goto l538
			l541:
				if !matchDot() {
					goto l538
				}
			l539:
				goto l537
			l538:
				position = position538
			}
			if !p.rules[ruleHtmlBlockCloseThead]() {
				goto l536
			}
			return true
		l536:
			position = position0
			return false
		},
		/* 124 HtmlBlockOpenTr <- ('<' Spnl ((&[T] 'TR') | (&[t] 'tr')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l542
			}
			if !p.rules[ruleSpnl]() {
				goto l542
			}
			{
				if position == len(p.Buffer) {
					goto l542
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l542
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l542
					}
					break
				default:
					goto l542
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l542
			}
		l544:
			if !p.rules[ruleHtmlAttribute]() {
				goto l545
			}
			goto l544
		l545:
			if !matchChar('>') {
				goto l542
			}
			return true
		l542:
			position = position0
			return false
		},
		/* 125 HtmlBlockCloseTr <- ('<' Spnl '/' ((&[T] 'TR') | (&[t] 'tr')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l546
			}
			if !p.rules[ruleSpnl]() {
				goto l546
			}
			if !matchChar('/') {
				goto l546
			}
			{
				if position == len(p.Buffer) {
					goto l546
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l546
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l546
					}
					break
				default:
					goto l546
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l546
			}
			if !matchChar('>') {
				goto l546
			}
			return true
		l546:
			position = position0
			return false
		},
		/* 126 HtmlBlockTr <- (HtmlBlockOpenTr (HtmlBlockTr / (!HtmlBlockCloseTr .))* HtmlBlockCloseTr) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTr]() {
				goto l548
			}
		l549:
			{
				position550 := position
				if !p.rules[ruleHtmlBlockTr]() {
					goto l552
				}
				goto l551
			l552:
				if !p.rules[ruleHtmlBlockCloseTr]() {
					goto l553
				}
				goto l550
			l553:
				if !matchDot() {
					goto l550
				}
			l551:
				goto l549
			l550:
				position = position550
			}
			if !p.rules[ruleHtmlBlockCloseTr]() {
				goto l548
			}
			return true
		l548:
			position = position0
			return false
		},
		/* 127 HtmlBlockOpenScript <- ('<' Spnl ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l554
			}
			if !p.rules[ruleSpnl]() {
				goto l554
			}
			{
				if position == len(p.Buffer) {
					goto l554
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l554
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l554
					}
					break
				default:
					goto l554
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l554
			}
		l556:
			if !p.rules[ruleHtmlAttribute]() {
				goto l557
			}
			goto l556
		l557:
			if !matchChar('>') {
				goto l554
			}
			return true
		l554:
			position = position0
			return false
		},
		/* 128 HtmlBlockCloseScript <- ('<' Spnl '/' ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l558
			}
			if !p.rules[ruleSpnl]() {
				goto l558
			}
			if !matchChar('/') {
				goto l558
			}
			{
				if position == len(p.Buffer) {
					goto l558
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l558
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l558
					}
					break
				default:
					goto l558
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l558
			}
			if !matchChar('>') {
				goto l558
			}
			return true
		l558:
			position = position0
			return false
		},
		/* 129 HtmlBlockScript <- (HtmlBlockOpenScript (!HtmlBlockCloseScript .)* HtmlBlockCloseScript) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenScript]() {
				goto l560
			}
		l561:
			{
				position562 := position
				if !p.rules[ruleHtmlBlockCloseScript]() {
					goto l563
				}
				goto l562
			l563:
				if !matchDot() {
					goto l562
				}
				goto l561
			l562:
				position = position562
			}
			if !p.rules[ruleHtmlBlockCloseScript]() {
				goto l560
			}
			return true
		l560:
			position = position0
			return false
		},
		/* 130 HtmlBlockInTags <- (HtmlBlockAddress / HtmlBlockBlockquote / HtmlBlockCenter / HtmlBlockDir / HtmlBlockDiv / HtmlBlockDl / HtmlBlockFieldset / HtmlBlockForm / HtmlBlockH1 / HtmlBlockH2 / HtmlBlockH3 / HtmlBlockH4 / HtmlBlockH5 / HtmlBlockH6 / HtmlBlockMenu / HtmlBlockNoframes / HtmlBlockNoscript / HtmlBlockOl / HtmlBlockP / HtmlBlockPre / HtmlBlockTable / HtmlBlockUl / HtmlBlockDd / HtmlBlockDt / HtmlBlockFrameset / HtmlBlockLi / HtmlBlockTbody / HtmlBlockTd / HtmlBlockTfoot / HtmlBlockTh / HtmlBlockThead / HtmlBlockTr / HtmlBlockScript) */
		func() bool {
			if !p.rules[ruleHtmlBlockAddress]() {
				goto l566
			}
			goto l565
		l566:
			if !p.rules[ruleHtmlBlockBlockquote]() {
				goto l567
			}
			goto l565
		l567:
			if !p.rules[ruleHtmlBlockCenter]() {
				goto l568
			}
			goto l565
		l568:
			if !p.rules[ruleHtmlBlockDir]() {
				goto l569
			}
			goto l565
		l569:
			if !p.rules[ruleHtmlBlockDiv]() {
				goto l570
			}
			goto l565
		l570:
			if !p.rules[ruleHtmlBlockDl]() {
				goto l571
			}
			goto l565
		l571:
			if !p.rules[ruleHtmlBlockFieldset]() {
				goto l572
			}
			goto l565
		l572:
			if !p.rules[ruleHtmlBlockForm]() {
				goto l573
			}
			goto l565
		l573:
			if !p.rules[ruleHtmlBlockH1]() {
				goto l574
			}
			goto l565
		l574:
			if !p.rules[ruleHtmlBlockH2]() {
				goto l575
			}
			goto l565
		l575:
			if !p.rules[ruleHtmlBlockH3]() {
				goto l576
			}
			goto l565
		l576:
			if !p.rules[ruleHtmlBlockH4]() {
				goto l577
			}
			goto l565
		l577:
			if !p.rules[ruleHtmlBlockH5]() {
				goto l578
			}
			goto l565
		l578:
			if !p.rules[ruleHtmlBlockH6]() {
				goto l579
			}
			goto l565
		l579:
			if !p.rules[ruleHtmlBlockMenu]() {
				goto l580
			}
			goto l565
		l580:
			if !p.rules[ruleHtmlBlockNoframes]() {
				goto l581
			}
			goto l565
		l581:
			if !p.rules[ruleHtmlBlockNoscript]() {
				goto l582
			}
			goto l565
		l582:
			if !p.rules[ruleHtmlBlockOl]() {
				goto l583
			}
			goto l565
		l583:
			if !p.rules[ruleHtmlBlockP]() {
				goto l584
			}
			goto l565
		l584:
			if !p.rules[ruleHtmlBlockPre]() {
				goto l585
			}
			goto l565
		l585:
			if !p.rules[ruleHtmlBlockTable]() {
				goto l586
			}
			goto l565
		l586:
			if !p.rules[ruleHtmlBlockUl]() {
				goto l587
			}
			goto l565
		l587:
			if !p.rules[ruleHtmlBlockDd]() {
				goto l588
			}
			goto l565
		l588:
			if !p.rules[ruleHtmlBlockDt]() {
				goto l589
			}
			goto l565
		l589:
			if !p.rules[ruleHtmlBlockFrameset]() {
				goto l590
			}
			goto l565
		l590:
			if !p.rules[ruleHtmlBlockLi]() {
				goto l591
			}
			goto l565
		l591:
			if !p.rules[ruleHtmlBlockTbody]() {
				goto l592
			}
			goto l565
		l592:
			if !p.rules[ruleHtmlBlockTd]() {
				goto l593
			}
			goto l565
		l593:
			if !p.rules[ruleHtmlBlockTfoot]() {
				goto l594
			}
			goto l565
		l594:
			if !p.rules[ruleHtmlBlockTh]() {
				goto l595
			}
			goto l565
		l595:
			if !p.rules[ruleHtmlBlockThead]() {
				goto l596
			}
			goto l565
		l596:
			if !p.rules[ruleHtmlBlockTr]() {
				goto l597
			}
			goto l565
		l597:
			if !p.rules[ruleHtmlBlockScript]() {
				goto l564
			}
		l565:
			return true
		l564:
			return false
		},
		/* 131 HtmlBlock <- (&'<' < (HtmlBlockInTags / HtmlComment / HtmlBlockSelfClosing) > BlankLine+ {   if p.extension.FilterHTML {
		        yy = p.mkList(LIST, nil)
		    } else {
		        yy = p.mkString(yytext)
		        yy.key = HTMLBLOCK
		    }
		}) */
		func() bool {
			position0 := position
			if !peekChar('<') {
				goto l598
			}
			begin = position
			if !p.rules[ruleHtmlBlockInTags]() {
				goto l600
			}
			goto l599
		l600:
			if !p.rules[ruleHtmlComment]() {
				goto l601
			}
			goto l599
		l601:
			if !p.rules[ruleHtmlBlockSelfClosing]() {
				goto l598
			}
		l599:
			end = position
			if !p.rules[ruleBlankLine]() {
				goto l598
			}
		l602:
			if !p.rules[ruleBlankLine]() {
				goto l603
			}
			goto l602
		l603:
			do(41)
			return true
		l598:
			position = position0
			return false
		},
		/* 132 HtmlBlockSelfClosing <- ('<' Spnl HtmlBlockType Spnl HtmlAttribute* '/' Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l604
			}
			if !p.rules[ruleSpnl]() {
				goto l604
			}
			if !p.rules[ruleHtmlBlockType]() {
				goto l604
			}
			if !p.rules[ruleSpnl]() {
				goto l604
			}
		l605:
			if !p.rules[ruleHtmlAttribute]() {
				goto l606
			}
			goto l605
		l606:
			if !matchChar('/') {
				goto l604
			}
			if !p.rules[ruleSpnl]() {
				goto l604
			}
			if !matchChar('>') {
				goto l604
			}
			return true
		l604:
			position = position0
			return false
		},
		/* 133 HtmlBlockType <- ('dir' / 'div' / 'dl' / 'fieldset' / 'form' / 'h1' / 'h2' / 'h3' / 'h4' / 'h5' / 'h6' / 'noframes' / 'p' / 'table' / 'dd' / 'tbody' / 'td' / 'tfoot' / 'th' / 'thead' / 'DIR' / 'DIV' / 'DL' / 'FIELDSET' / 'FORM' / 'H1' / 'H2' / 'H3' / 'H4' / 'H5' / 'H6' / 'NOFRAMES' / 'P' / 'TABLE' / 'DD' / 'TBODY' / 'TD' / 'TFOOT' / 'TH' / 'THEAD' / ((&[S] 'SCRIPT') | (&[T] 'TR') | (&[L] 'LI') | (&[F] 'FRAMESET') | (&[D] 'DT') | (&[U] 'UL') | (&[P] 'PRE') | (&[O] 'OL') | (&[N] 'NOSCRIPT') | (&[M] 'MENU') | (&[I] 'ISINDEX') | (&[H] 'HR') | (&[C] 'CENTER') | (&[B] 'BLOCKQUOTE') | (&[A] 'ADDRESS') | (&[s] 'script') | (&[t] 'tr') | (&[l] 'li') | (&[f] 'frameset') | (&[d] 'dt') | (&[u] 'ul') | (&[p] 'pre') | (&[o] 'ol') | (&[n] 'noscript') | (&[m] 'menu') | (&[i] 'isindex') | (&[h] 'hr') | (&[c] 'center') | (&[b] 'blockquote') | (&[a] 'address'))) */
		func() bool {
			if !matchString("dir") {
				goto l609
			}
			goto l608
		l609:
			if !matchString("div") {
				goto l610
			}
			goto l608
		l610:
			if !matchString("dl") {
				goto l611
			}
			goto l608
		l611:
			if !matchString("fieldset") {
				goto l612
			}
			goto l608
		l612:
			if !matchString("form") {
				goto l613
			}
			goto l608
		l613:
			if !matchString("h1") {
				goto l614
			}
			goto l608
		l614:
			if !matchString("h2") {
				goto l615
			}
			goto l608
		l615:
			if !matchString("h3") {
				goto l616
			}
			goto l608
		l616:
			if !matchString("h4") {
				goto l617
			}
			goto l608
		l617:
			if !matchString("h5") {
				goto l618
			}
			goto l608
		l618:
			if !matchString("h6") {
				goto l619
			}
			goto l608
		l619:
			if !matchString("noframes") {
				goto l620
			}
			goto l608
		l620:
			if !matchChar('p') {
				goto l621
			}
			goto l608
		l621:
			if !matchString("table") {
				goto l622
			}
			goto l608
		l622:
			if !matchString("dd") {
				goto l623
			}
			goto l608
		l623:
			if !matchString("tbody") {
				goto l624
			}
			goto l608
		l624:
			if !matchString("td") {
				goto l625
			}
			goto l608
		l625:
			if !matchString("tfoot") {
				goto l626
			}
			goto l608
		l626:
			if !matchString("th") {
				goto l627
			}
			goto l608
		l627:
			if !matchString("thead") {
				goto l628
			}
			goto l608
		l628:
			if !matchString("DIR") {
				goto l629
			}
			goto l608
		l629:
			if !matchString("DIV") {
				goto l630
			}
			goto l608
		l630:
			if !matchString("DL") {
				goto l631
			}
			goto l608
		l631:
			if !matchString("FIELDSET") {
				goto l632
			}
			goto l608
		l632:
			if !matchString("FORM") {
				goto l633
			}
			goto l608
		l633:
			if !matchString("H1") {
				goto l634
			}
			goto l608
		l634:
			if !matchString("H2") {
				goto l635
			}
			goto l608
		l635:
			if !matchString("H3") {
				goto l636
			}
			goto l608
		l636:
			if !matchString("H4") {
				goto l637
			}
			goto l608
		l637:
			if !matchString("H5") {
				goto l638
			}
			goto l608
		l638:
			if !matchString("H6") {
				goto l639
			}
			goto l608
		l639:
			if !matchString("NOFRAMES") {
				goto l640
			}
			goto l608
		l640:
			if !matchChar('P') {
				goto l641
			}
			goto l608
		l641:
			if !matchString("TABLE") {
				goto l642
			}
			goto l608
		l642:
			if !matchString("DD") {
				goto l643
			}
			goto l608
		l643:
			if !matchString("TBODY") {
				goto l644
			}
			goto l608
		l644:
			if !matchString("TD") {
				goto l645
			}
			goto l608
		l645:
			if !matchString("TFOOT") {
				goto l646
			}
			goto l608
		l646:
			if !matchString("TH") {
				goto l647
			}
			goto l608
		l647:
			if !matchString("THEAD") {
				goto l648
			}
			goto l608
		l648:
			{
				if position == len(p.Buffer) {
					goto l607
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l607
					}
					break
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l607
					}
					break
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l607
					}
					break
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l607
					}
					break
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l607
					}
					break
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l607
					}
					break
				case 'P':
					position++
					if !matchString("RE") {
						goto l607
					}
					break
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l607
					}
					break
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l607
					}
					break
				case 'M':
					position++
					if !matchString("ENU") {
						goto l607
					}
					break
				case 'I':
					position++
					if !matchString("SINDEX") {
						goto l607
					}
					break
				case 'H':
					position++ // matchString(`HR`)
					if !matchChar('R') {
						goto l607
					}
					break
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l607
					}
					break
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l607
					}
					break
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l607
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l607
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l607
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l607
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l607
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l607
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l607
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l607
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l607
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l607
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l607
					}
					break
				case 'i':
					position++
					if !matchString("sindex") {
						goto l607
					}
					break
				case 'h':
					position++ // matchString(`hr`)
					if !matchChar('r') {
						goto l607
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l607
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l607
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l607
					}
					break
				default:
					goto l607
				}
			}
		l608:
			return true
		l607:
			return false
		},
		/* 134 StyleOpen <- ('<' Spnl ((&[S] 'STYLE') | (&[s] 'style')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l650
			}
			if !p.rules[ruleSpnl]() {
				goto l650
			}
			{
				if position == len(p.Buffer) {
					goto l650
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto l650
					}
					break
				case 's':
					position++
					if !matchString("tyle") {
						goto l650
					}
					break
				default:
					goto l650
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l650
			}
		l652:
			if !p.rules[ruleHtmlAttribute]() {
				goto l653
			}
			goto l652
		l653:
			if !matchChar('>') {
				goto l650
			}
			return true
		l650:
			position = position0
			return false
		},
		/* 135 StyleClose <- ('<' Spnl '/' ((&[S] 'STYLE') | (&[s] 'style')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l654
			}
			if !p.rules[ruleSpnl]() {
				goto l654
			}
			if !matchChar('/') {
				goto l654
			}
			{
				if position == len(p.Buffer) {
					goto l654
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto l654
					}
					break
				case 's':
					position++
					if !matchString("tyle") {
						goto l654
					}
					break
				default:
					goto l654
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l654
			}
			if !matchChar('>') {
				goto l654
			}
			return true
		l654:
			position = position0
			return false
		},
		/* 136 InStyleTags <- (StyleOpen (!StyleClose .)* StyleClose) */
		func() bool {
			position0 := position
			if !p.rules[ruleStyleOpen]() {
				goto l656
			}
		l657:
			{
				position658 := position
				if !p.rules[ruleStyleClose]() {
					goto l659
				}
				goto l658
			l659:
				if !matchDot() {
					goto l658
				}
				goto l657
			l658:
				position = position658
			}
			if !p.rules[ruleStyleClose]() {
				goto l656
			}
			return true
		l656:
			position = position0
			return false
		},
		/* 137 StyleBlock <- (< InStyleTags > BlankLine* {   if p.extension.FilterStyles {
		        yy = p.mkList(LIST, nil)
		    } else {
		        yy = p.mkString(yytext)
		        yy.key = HTMLBLOCK
		    }
		}) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleInStyleTags]() {
				goto l660
			}
			end = position
		l661:
			if !p.rules[ruleBlankLine]() {
				goto l662
			}
			goto l661
		l662:
			do(42)
			return true
		l660:
			position = position0
			return false
		},
		/* 138 Inlines <- (StartList ((!Endline Inline { a = cons(yy, a) }) / (Endline &Inline { a = cons(c, a) }))+ Endline? { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l663
			}
			doarg(yySet, -1)
			{
				position666 := position
				if !p.rules[ruleEndline]() {
					goto l668
				}
				goto l667
			l668:
				if !p.rules[ruleInline]() {
					goto l667
				}
				do(43)
				goto l666
			l667:
				position = position666
				if !p.rules[ruleEndline]() {
					goto l663
				}
				doarg(yySet, -2)
				{
					position669 := position
					if !p.rules[ruleInline]() {
						goto l663
					}
					position = position669
				}
				do(44)
			}
		l666:
		l664:
			{
				position665, thunkPosition665 := position, thunkPosition
				{
					position670 := position
					if !p.rules[ruleEndline]() {
						goto l672
					}
					goto l671
				l672:
					if !p.rules[ruleInline]() {
						goto l671
					}
					do(43)
					goto l670
				l671:
					position = position670
					if !p.rules[ruleEndline]() {
						goto l665
					}
					doarg(yySet, -2)
					{
						position673 := position
						if !p.rules[ruleInline]() {
							goto l665
						}
						position = position673
					}
					do(44)
				}
			l670:
				goto l664
			l665:
				position, thunkPosition = position665, thunkPosition665
			}
			if !p.rules[ruleEndline]() {
				goto l674
			}
		l674:
			do(45)
			doarg(yyPop, 2)
			return true
		l663:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 139 Inline <- (Str / Endline / UlOrStarLine / Space / Strong / Emph / Image / Link / NoteReference / InlineNote / Code / RawHtml / Entity / EscapedChar / Smart / Symbol) */
		func() bool {
			if !p.rules[ruleStr]() {
				goto l678
			}
			goto l677
		l678:
			if !p.rules[ruleEndline]() {
				goto l679
			}
			goto l677
		l679:
			if !p.rules[ruleUlOrStarLine]() {
				goto l680
			}
			goto l677
		l680:
			if !p.rules[ruleSpace]() {
				goto l681
			}
			goto l677
		l681:
			if !p.rules[ruleStrong]() {
				goto l682
			}
			goto l677
		l682:
			if !p.rules[ruleEmph]() {
				goto l683
			}
			goto l677
		l683:
			if !p.rules[ruleImage]() {
				goto l684
			}
			goto l677
		l684:
			if !p.rules[ruleLink]() {
				goto l685
			}
			goto l677
		l685:
			if !p.rules[ruleNoteReference]() {
				goto l686
			}
			goto l677
		l686:
			if !p.rules[ruleInlineNote]() {
				goto l687
			}
			goto l677
		l687:
			if !p.rules[ruleCode]() {
				goto l688
			}
			goto l677
		l688:
			if !p.rules[ruleRawHtml]() {
				goto l689
			}
			goto l677
		l689:
			if !p.rules[ruleEntity]() {
				goto l690
			}
			goto l677
		l690:
			if !p.rules[ruleEscapedChar]() {
				goto l691
			}
			goto l677
		l691:
			if !p.rules[ruleSmart]() {
				goto l692
			}
			goto l677
		l692:
			if !p.rules[ruleSymbol]() {
				goto l676
			}
		l677:
			return true
		l676:
			return false
		},
		/* 140 Space <- (Spacechar+ { yy = p.mkString(" ")
		   yy.key = SPACE }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto l693
			}
		l694:
			if !p.rules[ruleSpacechar]() {
				goto l695
			}
			goto l694
		l695:
			do(46)
			return true
		l693:
			position = position0
			return false
		},
		/* 141 Str <- (StartList < NormalChar+ > { a = cons(p.mkString(yytext), a) } (StrChunk { a = cons(yy, a) })* { if a.next == nil { yy = a; } else { yy = p.mkList(LIST, a) } }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l696
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleNormalChar]() {
				goto l696
			}
		l697:
			if !p.rules[ruleNormalChar]() {
				goto l698
			}
			goto l697
		l698:
			end = position
			do(47)
		l699:
			{
				position700, thunkPosition700 := position, thunkPosition
				if !p.rules[ruleStrChunk]() {
					goto l700
				}
				do(48)
				goto l699
			l700:
				position, thunkPosition = position700, thunkPosition700
			}
			do(49)
			doarg(yyPop, 1)
			return true
		l696:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 142 StrChunk <- ((< (NormalChar / ('_'+ &Alphanumeric))+ > { yy = p.mkString(yytext) }) / AposChunk) */
		func() bool {
			position0 := position
			{
				position702 := position
				begin = position
				if !p.rules[ruleNormalChar]() {
					goto l707
				}
				goto l706
			l707:
				if !matchChar('_') {
					goto l703
				}
			l708:
				if !matchChar('_') {
					goto l709
				}
				goto l708
			l709:
				{
					position710 := position
					if !p.rules[ruleAlphanumeric]() {
						goto l703
					}
					position = position710
				}
			l706:
			l704:
				{
					position705 := position
					if !p.rules[ruleNormalChar]() {
						goto l712
					}
					goto l711
				l712:
					if !matchChar('_') {
						goto l705
					}
				l713:
					if !matchChar('_') {
						goto l714
					}
					goto l713
				l714:
					{
						position715 := position
						if !p.rules[ruleAlphanumeric]() {
							goto l705
						}
						position = position715
					}
				l711:
					goto l704
				l705:
					position = position705
				}
				end = position
				do(50)
				goto l702
			l703:
				position = position702
				if !p.rules[ruleAposChunk]() {
					goto l701
				}
			}
		l702:
			return true
		l701:
			position = position0
			return false
		},
		/* 143 AposChunk <- (&{p.extension.Smart} '\'' &Alphanumeric { yy = p.mkElem(APOSTROPHE) }) */
		func() bool {
			position0 := position
			if !(p.extension.Smart) {
				goto l716
			}
			if !matchChar('\'') {
				goto l716
			}
			{
				position717 := position
				if !p.rules[ruleAlphanumeric]() {
					goto l716
				}
				position = position717
			}
			do(51)
			return true
		l716:
			position = position0
			return false
		},
		/* 144 EscapedChar <- ('\\' !Newline < [-\\`|*_{}[\]()#+.!><] > { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !matchChar('\\') {
				goto l718
			}
			if !p.rules[ruleNewline]() {
				goto l719
			}
			goto l718
		l719:
			begin = position
			if !matchClass(1) {
				goto l718
			}
			end = position
			do(52)
			return true
		l718:
			position = position0
			return false
		},
		/* 145 Entity <- ((HexEntity / DecEntity / CharEntity) { yy = p.mkString(yytext); yy.key = HTML }) */
		func() bool {
			position0 := position
			if !p.rules[ruleHexEntity]() {
				goto l722
			}
			goto l721
		l722:
			if !p.rules[ruleDecEntity]() {
				goto l723
			}
			goto l721
		l723:
			if !p.rules[ruleCharEntity]() {
				goto l720
			}
		l721:
			do(53)
			return true
		l720:
			position = position0
			return false
		},
		/* 146 Endline <- (LineBreak / TerminalEndline / NormalEndline) */
		func() bool {
			if !p.rules[ruleLineBreak]() {
				goto l726
			}
			goto l725
		l726:
			if !p.rules[ruleTerminalEndline]() {
				goto l727
			}
			goto l725
		l727:
			if !p.rules[ruleNormalEndline]() {
				goto l724
			}
		l725:
			return true
		l724:
			return false
		},
		/* 147 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line ((&[\-] '-'+) | (&[=] '='+)) Newline) { yy = p.mkString("\n")
		   yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l728
			}
			if !p.rules[ruleNewline]() {
				goto l728
			}
			if !p.rules[ruleBlankLine]() {
				goto l729
			}
			goto l728
		l729:
			if peekChar('>') {
				goto l728
			}
			if !p.rules[ruleAtxStart]() {
				goto l730
			}
			goto l728
		l730:
			{
				position731, thunkPosition731 := position, thunkPosition
				if !p.rules[ruleLine]() {
					goto l731
				}
				{
					if position == len(p.Buffer) {
						goto l731
					}
					switch p.Buffer[position] {
					case '-':
						if !matchChar('-') {
							goto l731
						}
					l733:
						if !matchChar('-') {
							goto l734
						}
						goto l733
					l734:
						break
					case '=':
						if !matchChar('=') {
							goto l731
						}
					l735:
						if !matchChar('=') {
							goto l736
						}
						goto l735
					l736:
						break
					default:
						goto l731
					}
				}
				if !p.rules[ruleNewline]() {
					goto l731
				}
				goto l728
			l731:
				position, thunkPosition = position731, thunkPosition731
			}
			do(54)
			return true
		l728:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 148 TerminalEndline <- (Sp Newline !. { yy = nil }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l737
			}
			if !p.rules[ruleNewline]() {
				goto l737
			}
			if position < len(p.Buffer) {
				goto l737
			}
			do(55)
			return true
		l737:
			position = position0
			return false
		},
		/* 149 LineBreak <- ('  ' NormalEndline { yy = p.mkElem(LINEBREAK) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("  ") {
				goto l738
			}
			if !p.rules[ruleNormalEndline]() {
				goto l738
			}
			do(56)
			return true
		l738:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 150 Symbol <- (< SpecialChar > { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleSpecialChar]() {
				goto l739
			}
			end = position
			do(57)
			return true
		l739:
			position = position0
			return false
		},
		/* 151 UlOrStarLine <- ((UlLine / StarLine) { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l742
			}
			goto l741
		l742:
			if !p.rules[ruleStarLine]() {
				goto l740
			}
		l741:
			do(58)
			return true
		l740:
			position = position0
			return false
		},
		/* 152 StarLine <- ((&[*] (< '****' '*'* >)) | (&[\t ] (< Spacechar '*'+ &Spacechar >))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l743
				}
				switch p.Buffer[position] {
				case '*':
					begin = position
					if !matchString("****") {
						goto l743
					}
				l745:
					if !matchChar('*') {
						goto l746
					}
					goto l745
				l746:
					end = position
					break
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto l743
					}
					if !matchChar('*') {
						goto l743
					}
				l747:
					if !matchChar('*') {
						goto l748
					}
					goto l747
				l748:
					{
						position749 := position
						if !p.rules[ruleSpacechar]() {
							goto l743
						}
						position = position749
					}
					end = position
					break
				default:
					goto l743
				}
			}
			return true
		l743:
			position = position0
			return false
		},
		/* 153 UlLine <- ((&[_] (< '____' '_'* >)) | (&[\t ] (< Spacechar '_'+ &Spacechar >))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l750
				}
				switch p.Buffer[position] {
				case '_':
					begin = position
					if !matchString("____") {
						goto l750
					}
				l752:
					if !matchChar('_') {
						goto l753
					}
					goto l752
				l753:
					end = position
					break
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto l750
					}
					if !matchChar('_') {
						goto l750
					}
				l754:
					if !matchChar('_') {
						goto l755
					}
					goto l754
				l755:
					{
						position756 := position
						if !p.rules[ruleSpacechar]() {
							goto l750
						}
						position = position756
					}
					end = position
					break
				default:
					goto l750
				}
			}
			return true
		l750:
			position = position0
			return false
		},
		/* 154 Emph <- ((&[_] EmphUl) | (&[*] EmphStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l757
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleEmphUl]() {
						goto l757
					}
					break
				case '*':
					if !p.rules[ruleEmphStar]() {
						goto l757
					}
					break
				default:
					goto l757
				}
			}
			return true
		l757:
			return false
		},
		/* 155 Whitespace <- ((&[\n\r] Newline) | (&[\t ] Spacechar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l759
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l759
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l759
					}
					break
				default:
					goto l759
				}
			}
			return true
		l759:
			return false
		},
		/* 156 EmphStar <- ('*' !Whitespace StartList ((!'*' Inline { a = cons(b, a) }) / (StrongStar { a = cons(b, a) }))+ '*' { yy = p.mkList(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('*') {
				goto l761
			}
			if !p.rules[ruleWhitespace]() {
				goto l762
			}
			goto l761
		l762:
			if !p.rules[ruleStartList]() {
				goto l761
			}
			doarg(yySet, -1)
			{
				position765, thunkPosition765 := position, thunkPosition
				if peekChar('*') {
					goto l766
				}
				if !p.rules[ruleInline]() {
					goto l766
				}
				doarg(yySet, -2)
				do(59)
				goto l765
			l766:
				position, thunkPosition = position765, thunkPosition765
				if !p.rules[ruleStrongStar]() {
					goto l761
				}
				doarg(yySet, -2)
				do(60)
			}
		l765:
		l763:
			{
				position764, thunkPosition764 := position, thunkPosition
				{
					position767, thunkPosition767 := position, thunkPosition
					if peekChar('*') {
						goto l768
					}
					if !p.rules[ruleInline]() {
						goto l768
					}
					doarg(yySet, -2)
					do(59)
					goto l767
				l768:
					position, thunkPosition = position767, thunkPosition767
					if !p.rules[ruleStrongStar]() {
						goto l764
					}
					doarg(yySet, -2)
					do(60)
				}
			l767:
				goto l763
			l764:
				position, thunkPosition = position764, thunkPosition764
			}
			if !matchChar('*') {
				goto l761
			}
			do(61)
			doarg(yyPop, 2)
			return true
		l761:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 157 EmphUl <- ('_' !Whitespace StartList ((!'_' Inline { a = cons(b, a) }) / (StrongUl { a = cons(b, a) }))+ '_' { yy = p.mkList(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('_') {
				goto l769
			}
			if !p.rules[ruleWhitespace]() {
				goto l770
			}
			goto l769
		l770:
			if !p.rules[ruleStartList]() {
				goto l769
			}
			doarg(yySet, -2)
			{
				position773, thunkPosition773 := position, thunkPosition
				if peekChar('_') {
					goto l774
				}
				if !p.rules[ruleInline]() {
					goto l774
				}
				doarg(yySet, -1)
				do(62)
				goto l773
			l774:
				position, thunkPosition = position773, thunkPosition773
				if !p.rules[ruleStrongUl]() {
					goto l769
				}
				doarg(yySet, -1)
				do(63)
			}
		l773:
		l771:
			{
				position772, thunkPosition772 := position, thunkPosition
				{
					position775, thunkPosition775 := position, thunkPosition
					if peekChar('_') {
						goto l776
					}
					if !p.rules[ruleInline]() {
						goto l776
					}
					doarg(yySet, -1)
					do(62)
					goto l775
				l776:
					position, thunkPosition = position775, thunkPosition775
					if !p.rules[ruleStrongUl]() {
						goto l772
					}
					doarg(yySet, -1)
					do(63)
				}
			l775:
				goto l771
			l772:
				position, thunkPosition = position772, thunkPosition772
			}
			if !matchChar('_') {
				goto l769
			}
			do(64)
			doarg(yyPop, 2)
			return true
		l769:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 158 Strong <- ((&[_] StrongUl) | (&[*] StrongStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l777
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleStrongUl]() {
						goto l777
					}
					break
				case '*':
					if !p.rules[ruleStrongStar]() {
						goto l777
					}
					break
				default:
					goto l777
				}
			}
			return true
		l777:
			return false
		},
		/* 159 StrongStar <- ('**' !Whitespace StartList (!'**' Inline { a = cons(b, a) })+ '**' { yy = p.mkList(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchString("**") {
				goto l779
			}
			if !p.rules[ruleWhitespace]() {
				goto l780
			}
			goto l779
		l780:
			if !p.rules[ruleStartList]() {
				goto l779
			}
			doarg(yySet, -2)
			if !matchString("**") {
				goto l783
			}
			goto l779
		l783:
			if !p.rules[ruleInline]() {
				goto l779
			}
			doarg(yySet, -1)
			do(65)
		l781:
			{
				position782, thunkPosition782 := position, thunkPosition
				if !matchString("**") {
					goto l784
				}
				goto l782
			l784:
				if !p.rules[ruleInline]() {
					goto l782
				}
				doarg(yySet, -1)
				do(65)
				goto l781
			l782:
				position, thunkPosition = position782, thunkPosition782
			}
			if !matchString("**") {
				goto l779
			}
			do(66)
			doarg(yyPop, 2)
			return true
		l779:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 160 StrongUl <- ('__' !Whitespace StartList (!'__' Inline { a = cons(b, a) })+ '__' { yy = p.mkList(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchString("__") {
				goto l785
			}
			if !p.rules[ruleWhitespace]() {
				goto l786
			}
			goto l785
		l786:
			if !p.rules[ruleStartList]() {
				goto l785
			}
			doarg(yySet, -2)
			if !matchString("__") {
				goto l789
			}
			goto l785
		l789:
			if !p.rules[ruleInline]() {
				goto l785
			}
			doarg(yySet, -1)
			do(67)
		l787:
			{
				position788, thunkPosition788 := position, thunkPosition
				if !matchString("__") {
					goto l790
				}
				goto l788
			l790:
				if !p.rules[ruleInline]() {
					goto l788
				}
				doarg(yySet, -1)
				do(67)
				goto l787
			l788:
				position, thunkPosition = position788, thunkPosition788
			}
			if !matchString("__") {
				goto l785
			}
			do(68)
			doarg(yyPop, 2)
			return true
		l785:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 161 Image <- ('!' (ExplicitLink / ReferenceLink) {	if yy.key == LINK {
				yy.key = IMAGE
			} else {
				result := yy
				yy.children = cons(p.mkString("!"), result.children)
			}
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('!') {
				goto l791
			}
			if !p.rules[ruleExplicitLink]() {
				goto l793
			}
			goto l792
		l793:
			if !p.rules[ruleReferenceLink]() {
				goto l791
			}
		l792:
			do(69)
			return true
		l791:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 162 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() bool {
			if !p.rules[ruleExplicitLink]() {
				goto l796
			}
			goto l795
		l796:
			if !p.rules[ruleReferenceLink]() {
				goto l797
			}
			goto l795
		l797:
			if !p.rules[ruleAutoLink]() {
				goto l794
			}
		l795:
			return true
		l794:
			return false
		},
		/* 163 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() bool {
			if !p.rules[ruleReferenceLinkDouble]() {
				goto l800
			}
			goto l799
		l800:
			if !p.rules[ruleReferenceLinkSingle]() {
				goto l798
			}
		l799:
			return true
		l798:
			return false
		},
		/* 164 ReferenceLinkDouble <- (Label < Spnl > !'[]' Label {
		    if match, found := p.findReference(b.children); found {
		        yy = p.mkLink(a.children, match.url, match.title);
		        a = nil
		        b = nil
		    } else {
		        result := p.mkElem(LIST)
		        result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), cons(p.mkString(yytext),
		                            cons(p.mkString("["), cons(b, p.mkString("]")))))))
		        yy = result
		    }
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleLabel]() {
				goto l801
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto l801
			}
			end = position
			if !matchString("[]") {
				goto l802
			}
			goto l801
		l802:
			if !p.rules[ruleLabel]() {
				goto l801
			}
			doarg(yySet, -2)
			do(70)
			doarg(yyPop, 2)
			return true
		l801:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 165 ReferenceLinkSingle <- (Label < (Spnl '[]')? > {
		    if match, found := p.findReference(a.children); found {
		        yy = p.mkLink(a.children, match.url, match.title)
		        a = nil
		    } else {
		        result := p.mkElem(LIST)
		        result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), p.mkString(yytext))));
		        yy = result
		    }
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleLabel]() {
				goto l803
			}
			doarg(yySet, -1)
			begin = position
			{
				position804 := position
				if !p.rules[ruleSpnl]() {
					goto l804
				}
				if !matchString("[]") {
					goto l804
				}
				goto l805
			l804:
				position = position804
			}
		l805:
			end = position
			do(71)
			doarg(yyPop, 1)
			return true
		l803:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 166 ExplicitLink <- (Label '(' Sp Source Spnl Title Sp ')' { yy = p.mkLink(l.children, s.contents.str, t.contents.str)
		   s = nil
		   t = nil
		   l = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleLabel]() {
				goto l806
			}
			doarg(yySet, -3)
			if !matchChar('(') {
				goto l806
			}
			if !p.rules[ruleSp]() {
				goto l806
			}
			if !p.rules[ruleSource]() {
				goto l806
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l806
			}
			if !p.rules[ruleTitle]() {
				goto l806
			}
			doarg(yySet, -2)
			if !p.rules[ruleSp]() {
				goto l806
			}
			if !matchChar(')') {
				goto l806
			}
			do(72)
			doarg(yyPop, 3)
			return true
		l806:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 167 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			{
				position808 := position
				if !matchChar('<') {
					goto l809
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l809
				}
				end = position
				if !matchChar('>') {
					goto l809
				}
				goto l808
			l809:
				position = position808
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l807
				}
				end = position
			}
		l808:
			do(73)
			return true
		l807:
			position = position0
			return false
		},
		/* 168 SourceContents <- ((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* */
		func() bool {
		l811:
			{
				position812 := position
				if position == len(p.Buffer) {
					goto l814
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l814
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l814
					}
				}
			l815:
				if position == len(p.Buffer) {
					goto l816
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l816
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l816
					}
				}
				goto l815
			l816:
				goto l813
			l814:
				if !matchChar('(') {
					goto l812
				}
				if !p.rules[ruleSourceContents]() {
					goto l812
				}
				if !matchChar(')') {
					goto l812
				}
			l813:
				goto l811
			l812:
				position = position812
			}
			return true
		},
		/* 169 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = p.mkString(yytext) }) */
		func() bool {
			if !p.rules[ruleTitleSingle]() {
				goto l819
			}
			goto l818
		l819:
			if !p.rules[ruleTitleDouble]() {
				goto l820
			}
			goto l818
		l820:
			begin = position
			end = position
		l818:
			do(74)
			return true
		},
		/* 170 TitleSingle <- ('\'' < (!('\'' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '\'') */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l821
			}
			begin = position
		l822:
			{
				position823 := position
				{
					position824 := position
					if !matchChar('\'') {
						goto l824
					}
					if !p.rules[ruleSp]() {
						goto l824
					}
					{
						if position == len(p.Buffer) {
							goto l824
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l824
							}
							break
						default:
							goto l824
						}
					}
					goto l823
				l824:
					position = position824
				}
				if !matchDot() {
					goto l823
				}
				goto l822
			l823:
				position = position823
			}
			end = position
			if !matchChar('\'') {
				goto l821
			}
			return true
		l821:
			position = position0
			return false
		},
		/* 171 TitleDouble <- ('"' < (!('"' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '"') */
		func() bool {
			position0 := position
			if !matchChar('"') {
				goto l826
			}
			begin = position
		l827:
			{
				position828 := position
				{
					position829 := position
					if !matchChar('"') {
						goto l829
					}
					if !p.rules[ruleSp]() {
						goto l829
					}
					{
						if position == len(p.Buffer) {
							goto l829
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l829
							}
							break
						default:
							goto l829
						}
					}
					goto l828
				l829:
					position = position829
				}
				if !matchDot() {
					goto l828
				}
				goto l827
			l828:
				position = position828
			}
			end = position
			if !matchChar('"') {
				goto l826
			}
			return true
		l826:
			position = position0
			return false
		},
		/* 172 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() bool {
			if !p.rules[ruleAutoLinkUrl]() {
				goto l833
			}
			goto l832
		l833:
			if !p.rules[ruleAutoLinkEmail]() {
				goto l831
			}
		l832:
			return true
		l831:
			return false
		},
		/* 173 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = p.mkLink(p.mkString(yytext), yytext, "") }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l834
			}
			begin = position
			if !matchClass(2) {
				goto l834
			}
		l835:
			if !matchClass(2) {
				goto l836
			}
			goto l835
		l836:
			if !matchString("://") {
				goto l834
			}
			if !p.rules[ruleNewline]() {
				goto l839
			}
			goto l834
		l839:
			if peekChar('>') {
				goto l834
			}
			if !matchDot() {
				goto l834
			}
		l837:
			{
				position838 := position
				if !p.rules[ruleNewline]() {
					goto l840
				}
				goto l838
			l840:
				if peekChar('>') {
					goto l838
				}
				if !matchDot() {
					goto l838
				}
				goto l837
			l838:
				position = position838
			}
			end = position
			if !matchChar('>') {
				goto l834
			}
			do(75)
			return true
		l834:
			position = position0
			return false
		},
		/* 174 AutoLinkEmail <- ('<' 'mailto:'? < [-A-Za-z0-9+_./!%~$]+ '@' (!Newline !'>' .)+ > '>' {
		    yy = p.mkLink(p.mkString(yytext), "mailto:"+yytext, "")
		}) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l841
			}
			if !matchString("mailto:") {
				goto l842
			}
		l842:
			begin = position
			if !matchClass(3) {
				goto l841
			}
		l844:
			if !matchClass(3) {
				goto l845
			}
			goto l844
		l845:
			if !matchChar('@') {
				goto l841
			}
			if !p.rules[ruleNewline]() {
				goto l848
			}
			goto l841
		l848:
			if peekChar('>') {
				goto l841
			}
			if !matchDot() {
				goto l841
			}
		l846:
			{
				position847 := position
				if !p.rules[ruleNewline]() {
					goto l849
				}
				goto l847
			l849:
				if peekChar('>') {
					goto l847
				}
				if !matchDot() {
					goto l847
				}
				goto l846
			l847:
				position = position847
			}
			end = position
			if !matchChar('>') {
				goto l841
			}
			do(76)
			return true
		l841:
			position = position0
			return false
		},
		/* 175 Reference <- (NonindentSpace !'[]' Label ':' Spnl RefSrc RefTitle BlankLine+ { yy = p.mkLink(l.children, s.contents.str, t.contents.str)
		   s = nil
		   t = nil
		   l = nil
		   yy.key = REFERENCE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleNonindentSpace]() {
				goto l850
			}
			if !matchString("[]") {
				goto l851
			}
			goto l850
		l851:
			if !p.rules[ruleLabel]() {
				goto l850
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto l850
			}
			if !p.rules[ruleSpnl]() {
				goto l850
			}
			if !p.rules[ruleRefSrc]() {
				goto l850
			}
			doarg(yySet, -3)
			if !p.rules[ruleRefTitle]() {
				goto l850
			}
			doarg(yySet, -2)
			if !p.rules[ruleBlankLine]() {
				goto l850
			}
		l852:
			if !p.rules[ruleBlankLine]() {
				goto l853
			}
			goto l852
		l853:
			do(77)
			doarg(yyPop, 3)
			return true
		l850:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 176 Label <- ('[' ((!'^' &{p.extension.Notes}) / (&. &{!p.extension.Notes})) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l854
			}
			if peekChar('^') {
				goto l856
			}
			if !(p.extension.Notes) {
				goto l856
			}
			goto l855
		l856:
			if !(position < len(p.Buffer)) {
				goto l854
			}
			if !(!p.extension.Notes) {
				goto l854
			}
		l855:
			if !p.rules[ruleStartList]() {
				goto l854
			}
			doarg(yySet, -1)
		l857:
			{
				position858 := position
				if peekChar(']') {
					goto l858
				}
				if !p.rules[ruleInline]() {
					goto l858
				}
				do(78)
				goto l857
			l858:
				position = position858
			}
			if !matchChar(']') {
				goto l854
			}
			do(79)
			doarg(yyPop, 1)
			return true
		l854:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 177 RefSrc <- (< Nonspacechar+ > { yy = p.mkString(yytext)
		   yy.key = HTML }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto l859
			}
		l860:
			if !p.rules[ruleNonspacechar]() {
				goto l861
			}
			goto l860
		l861:
			end = position
			do(80)
			return true
		l859:
			position = position0
			return false
		},
		/* 178 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRefTitleSingle]() {
				goto l864
			}
			goto l863
		l864:
			if !p.rules[ruleRefTitleDouble]() {
				goto l865
			}
			goto l863
		l865:
			if !p.rules[ruleRefTitleParens]() {
				goto l866
			}
			goto l863
		l866:
			if !p.rules[ruleEmptyTitle]() {
				goto l862
			}
		l863:
			do(81)
			return true
		l862:
			position = position0
			return false
		},
		/* 179 EmptyTitle <- (< '' >) */
		func() bool {
			begin = position
			end = position
			return true
		},
		/* 180 RefTitleSingle <- (Spnl '\'' < (!((&[\'] ('\'' Sp Newline)) | (&[\n\r] Newline)) .)* > '\'') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l868
			}
			if !matchChar('\'') {
				goto l868
			}
			begin = position
		l869:
			{
				position870 := position
				{
					position871 := position
					{
						if position == len(p.Buffer) {
							goto l871
						}
						switch p.Buffer[position] {
						case '\'':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l871
							}
							if !p.rules[ruleNewline]() {
								goto l871
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l871
							}
							break
						default:
							goto l871
						}
					}
					goto l870
				l871:
					position = position871
				}
				if !matchDot() {
					goto l870
				}
				goto l869
			l870:
				position = position870
			}
			end = position
			if !matchChar('\'') {
				goto l868
			}
			return true
		l868:
			position = position0
			return false
		},
		/* 181 RefTitleDouble <- (Spnl '"' < (!((&[\"] ('"' Sp Newline)) | (&[\n\r] Newline)) .)* > '"') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l873
			}
			if !matchChar('"') {
				goto l873
			}
			begin = position
		l874:
			{
				position875 := position
				{
					position876 := position
					{
						if position == len(p.Buffer) {
							goto l876
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l876
							}
							if !p.rules[ruleNewline]() {
								goto l876
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l876
							}
							break
						default:
							goto l876
						}
					}
					goto l875
				l876:
					position = position876
				}
				if !matchDot() {
					goto l875
				}
				goto l874
			l875:
				position = position875
			}
			end = position
			if !matchChar('"') {
				goto l873
			}
			return true
		l873:
			position = position0
			return false
		},
		/* 182 RefTitleParens <- (Spnl '(' < (!((&[)] (')' Sp Newline)) | (&[\n\r] Newline)) .)* > ')') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l878
			}
			if !matchChar('(') {
				goto l878
			}
			begin = position
		l879:
			{
				position880 := position
				{
					position881 := position
					{
						if position == len(p.Buffer) {
							goto l881
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l881
							}
							if !p.rules[ruleNewline]() {
								goto l881
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l881
							}
							break
						default:
							goto l881
						}
					}
					goto l880
				l881:
					position = position881
				}
				if !matchDot() {
					goto l880
				}
				goto l879
			l880:
				position = position880
			}
			end = position
			if !matchChar(')') {
				goto l878
			}
			return true
		l878:
			position = position0
			return false
		},
		/* 183 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l883
			}
			doarg(yySet, -1)
		l884:
			{
				position885, thunkPosition885 := position, thunkPosition
				{
					position886, thunkPosition886 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto l887
					}
					doarg(yySet, -2)
					do(82)
					goto l886
				l887:
					position, thunkPosition = position886, thunkPosition886
					if !p.rules[ruleSkipBlock]() {
						goto l885
					}
				}
			l886:
				goto l884
			l885:
				position, thunkPosition = position885, thunkPosition885
			}
			do(83)
			if !(commit(thunkPosition0)) {
				goto l883
			}
			doarg(yyPop, 2)
			return true
		l883:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 184 Ticks1 <- ('`' !'`') */
		func() bool {
			position0 := position
			if !matchChar('`') {
				goto l888
			}
			if peekChar('`') {
				goto l888
			}
			return true
		l888:
			position = position0
			return false
		},
		/* 185 Ticks2 <- ('``' !'`') */
		func() bool {
			position0 := position
			if !matchString("``") {
				goto l889
			}
			if peekChar('`') {
				goto l889
			}
			return true
		l889:
			position = position0
			return false
		},
		/* 186 Ticks3 <- ('```' !'`') */
		func() bool {
			position0 := position
			if !matchString("```") {
				goto l890
			}
			if peekChar('`') {
				goto l890
			}
			return true
		l890:
			position = position0
			return false
		},
		/* 187 Ticks4 <- ('````' !'`') */
		func() bool {
			position0 := position
			if !matchString("````") {
				goto l891
			}
			if peekChar('`') {
				goto l891
			}
			return true
		l891:
			position = position0
			return false
		},
		/* 188 Ticks5 <- ('`````' !'`') */
		func() bool {
			position0 := position
			if !matchString("`````") {
				goto l892
			}
			if peekChar('`') {
				goto l892
			}
			return true
		l892:
			position = position0
			return false
		},
		/* 189 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks1 '`'+)) | (&[\t\n\r ] (!(Sp Ticks1) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks2 '`'+)) | (&[\t\n\r ] (!(Sp Ticks2) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks3 '`'+)) | (&[\t\n\r ] (!(Sp Ticks3) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks4 '`'+)) | (&[\t\n\r ] (!(Sp Ticks4) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks5 '`'+)) | (&[\t\n\r ] (!(Sp Ticks5) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks5)) { yy = p.mkString(yytext); yy.key = CODE }) */
		func() bool {
			position0 := position
			{
				position894 := position
				if !p.rules[ruleTicks1]() {
					goto l895
				}
				if !p.rules[ruleSp]() {
					goto l895
				}
				begin = position
				if peekChar('`') {
					goto l899
				}
				if !p.rules[ruleNonspacechar]() {
					goto l899
				}
			l900:
				if peekChar('`') {
					goto l901
				}
				if !p.rules[ruleNonspacechar]() {
					goto l901
				}
				goto l900
			l901:
				goto l898
			l899:
				{
					if position == len(p.Buffer) {
						goto l895
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks1]() {
							goto l903
						}
						goto l895
					l903:
						if !matchChar('`') {
							goto l895
						}
					l904:
						if !matchChar('`') {
							goto l905
						}
						goto l904
					l905:
						break
					default:
						{
							position906 := position
							if !p.rules[ruleSp]() {
								goto l906
							}
							if !p.rules[ruleTicks1]() {
								goto l906
							}
							goto l895
						l906:
							position = position906
						}
						{
							if position == len(p.Buffer) {
								goto l895
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l895
								}
								if !p.rules[ruleBlankLine]() {
									goto l908
								}
								goto l895
							l908:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l895
								}
								break
							default:
								goto l895
							}
						}
					}
				}
			l898:
			l896:
				{
					position897 := position
					if peekChar('`') {
						goto l910
					}
					if !p.rules[ruleNonspacechar]() {
						goto l910
					}
				l911:
					if peekChar('`') {
						goto l912
					}
					if !p.rules[ruleNonspacechar]() {
						goto l912
					}
					goto l911
				l912:
					goto l909
				l910:
					{
						if position == len(p.Buffer) {
							goto l897
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks1]() {
								goto l914
							}
							goto l897
						l914:
							if !matchChar('`') {
								goto l897
							}
						l915:
							if !matchChar('`') {
								goto l916
							}
							goto l915
						l916:
							break
						default:
							{
								position917 := position
								if !p.rules[ruleSp]() {
									goto l917
								}
								if !p.rules[ruleTicks1]() {
									goto l917
								}
								goto l897
							l917:
								position = position917
							}
							{
								if position == len(p.Buffer) {
									goto l897
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l897
									}
									if !p.rules[ruleBlankLine]() {
										goto l919
									}
									goto l897
								l919:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l897
									}
									break
								default:
									goto l897
								}
							}
						}
					}
				l909:
					goto l896
				l897:
					position = position897
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l895
				}
				if !p.rules[ruleTicks1]() {
					goto l895
				}
				goto l894
			l895:
				position = position894
				if !p.rules[ruleTicks2]() {
					goto l920
				}
				if !p.rules[ruleSp]() {
					goto l920
				}
				begin = position
				if peekChar('`') {
					goto l924
				}
				if !p.rules[ruleNonspacechar]() {
					goto l924
				}
			l925:
				if peekChar('`') {
					goto l926
				}
				if !p.rules[ruleNonspacechar]() {
					goto l926
				}
				goto l925
			l926:
				goto l923
			l924:
				{
					if position == len(p.Buffer) {
						goto l920
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks2]() {
							goto l928
						}
						goto l920
					l928:
						if !matchChar('`') {
							goto l920
						}
					l929:
						if !matchChar('`') {
							goto l930
						}
						goto l929
					l930:
						break
					default:
						{
							position931 := position
							if !p.rules[ruleSp]() {
								goto l931
							}
							if !p.rules[ruleTicks2]() {
								goto l931
							}
							goto l920
						l931:
							position = position931
						}
						{
							if position == len(p.Buffer) {
								goto l920
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l920
								}
								if !p.rules[ruleBlankLine]() {
									goto l933
								}
								goto l920
							l933:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l920
								}
								break
							default:
								goto l920
							}
						}
					}
				}
			l923:
			l921:
				{
					position922 := position
					if peekChar('`') {
						goto l935
					}
					if !p.rules[ruleNonspacechar]() {
						goto l935
					}
				l936:
					if peekChar('`') {
						goto l937
					}
					if !p.rules[ruleNonspacechar]() {
						goto l937
					}
					goto l936
				l937:
					goto l934
				l935:
					{
						if position == len(p.Buffer) {
							goto l922
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks2]() {
								goto l939
							}
							goto l922
						l939:
							if !matchChar('`') {
								goto l922
							}
						l940:
							if !matchChar('`') {
								goto l941
							}
							goto l940
						l941:
							break
						default:
							{
								position942 := position
								if !p.rules[ruleSp]() {
									goto l942
								}
								if !p.rules[ruleTicks2]() {
									goto l942
								}
								goto l922
							l942:
								position = position942
							}
							{
								if position == len(p.Buffer) {
									goto l922
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l922
									}
									if !p.rules[ruleBlankLine]() {
										goto l944
									}
									goto l922
								l944:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l922
									}
									break
								default:
									goto l922
								}
							}
						}
					}
				l934:
					goto l921
				l922:
					position = position922
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l920
				}
				if !p.rules[ruleTicks2]() {
					goto l920
				}
				goto l894
			l920:
				position = position894
				if !p.rules[ruleTicks3]() {
					goto l945
				}
				if !p.rules[ruleSp]() {
					goto l945
				}
				begin = position
				if peekChar('`') {
					goto l949
				}
				if !p.rules[ruleNonspacechar]() {
					goto l949
				}
			l950:
				if peekChar('`') {
					goto l951
				}
				if !p.rules[ruleNonspacechar]() {
					goto l951
				}
				goto l950
			l951:
				goto l948
			l949:
				{
					if position == len(p.Buffer) {
						goto l945
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks3]() {
							goto l953
						}
						goto l945
					l953:
						if !matchChar('`') {
							goto l945
						}
					l954:
						if !matchChar('`') {
							goto l955
						}
						goto l954
					l955:
						break
					default:
						{
							position956 := position
							if !p.rules[ruleSp]() {
								goto l956
							}
							if !p.rules[ruleTicks3]() {
								goto l956
							}
							goto l945
						l956:
							position = position956
						}
						{
							if position == len(p.Buffer) {
								goto l945
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l945
								}
								if !p.rules[ruleBlankLine]() {
									goto l958
								}
								goto l945
							l958:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l945
								}
								break
							default:
								goto l945
							}
						}
					}
				}
			l948:
			l946:
				{
					position947 := position
					if peekChar('`') {
						goto l960
					}
					if !p.rules[ruleNonspacechar]() {
						goto l960
					}
				l961:
					if peekChar('`') {
						goto l962
					}
					if !p.rules[ruleNonspacechar]() {
						goto l962
					}
					goto l961
				l962:
					goto l959
				l960:
					{
						if position == len(p.Buffer) {
							goto l947
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks3]() {
								goto l964
							}
							goto l947
						l964:
							if !matchChar('`') {
								goto l947
							}
						l965:
							if !matchChar('`') {
								goto l966
							}
							goto l965
						l966:
							break
						default:
							{
								position967 := position
								if !p.rules[ruleSp]() {
									goto l967
								}
								if !p.rules[ruleTicks3]() {
									goto l967
								}
								goto l947
							l967:
								position = position967
							}
							{
								if position == len(p.Buffer) {
									goto l947
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l947
									}
									if !p.rules[ruleBlankLine]() {
										goto l969
									}
									goto l947
								l969:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l947
									}
									break
								default:
									goto l947
								}
							}
						}
					}
				l959:
					goto l946
				l947:
					position = position947
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l945
				}
				if !p.rules[ruleTicks3]() {
					goto l945
				}
				goto l894
			l945:
				position = position894
				if !p.rules[ruleTicks4]() {
					goto l970
				}
				if !p.rules[ruleSp]() {
					goto l970
				}
				begin = position
				if peekChar('`') {
					goto l974
				}
				if !p.rules[ruleNonspacechar]() {
					goto l974
				}
			l975:
				if peekChar('`') {
					goto l976
				}
				if !p.rules[ruleNonspacechar]() {
					goto l976
				}
				goto l975
			l976:
				goto l973
			l974:
				{
					if position == len(p.Buffer) {
						goto l970
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks4]() {
							goto l978
						}
						goto l970
					l978:
						if !matchChar('`') {
							goto l970
						}
					l979:
						if !matchChar('`') {
							goto l980
						}
						goto l979
					l980:
						break
					default:
						{
							position981 := position
							if !p.rules[ruleSp]() {
								goto l981
							}
							if !p.rules[ruleTicks4]() {
								goto l981
							}
							goto l970
						l981:
							position = position981
						}
						{
							if position == len(p.Buffer) {
								goto l970
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l970
								}
								if !p.rules[ruleBlankLine]() {
									goto l983
								}
								goto l970
							l983:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l970
								}
								break
							default:
								goto l970
							}
						}
					}
				}
			l973:
			l971:
				{
					position972 := position
					if peekChar('`') {
						goto l985
					}
					if !p.rules[ruleNonspacechar]() {
						goto l985
					}
				l986:
					if peekChar('`') {
						goto l987
					}
					if !p.rules[ruleNonspacechar]() {
						goto l987
					}
					goto l986
				l987:
					goto l984
				l985:
					{
						if position == len(p.Buffer) {
							goto l972
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks4]() {
								goto l989
							}
							goto l972
						l989:
							if !matchChar('`') {
								goto l972
							}
						l990:
							if !matchChar('`') {
								goto l991
							}
							goto l990
						l991:
							break
						default:
							{
								position992 := position
								if !p.rules[ruleSp]() {
									goto l992
								}
								if !p.rules[ruleTicks4]() {
									goto l992
								}
								goto l972
							l992:
								position = position992
							}
							{
								if position == len(p.Buffer) {
									goto l972
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l972
									}
									if !p.rules[ruleBlankLine]() {
										goto l994
									}
									goto l972
								l994:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l972
									}
									break
								default:
									goto l972
								}
							}
						}
					}
				l984:
					goto l971
				l972:
					position = position972
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l970
				}
				if !p.rules[ruleTicks4]() {
					goto l970
				}
				goto l894
			l970:
				position = position894
				if !p.rules[ruleTicks5]() {
					goto l893
				}
				if !p.rules[ruleSp]() {
					goto l893
				}
				begin = position
				if peekChar('`') {
					goto l998
				}
				if !p.rules[ruleNonspacechar]() {
					goto l998
				}
			l999:
				if peekChar('`') {
					goto l1000
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1000
				}
				goto l999
			l1000:
				goto l997
			l998:
				{
					if position == len(p.Buffer) {
						goto l893
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks5]() {
							goto l1002
						}
						goto l893
					l1002:
						if !matchChar('`') {
							goto l893
						}
					l1003:
						if !matchChar('`') {
							goto l1004
						}
						goto l1003
					l1004:
						break
					default:
						{
							position1005 := position
							if !p.rules[ruleSp]() {
								goto l1005
							}
							if !p.rules[ruleTicks5]() {
								goto l1005
							}
							goto l893
						l1005:
							position = position1005
						}
						{
							if position == len(p.Buffer) {
								goto l893
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l893
								}
								if !p.rules[ruleBlankLine]() {
									goto l1007
								}
								goto l893
							l1007:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l893
								}
								break
							default:
								goto l893
							}
						}
					}
				}
			l997:
			l995:
				{
					position996 := position
					if peekChar('`') {
						goto l1009
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1009
					}
				l1010:
					if peekChar('`') {
						goto l1011
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1011
					}
					goto l1010
				l1011:
					goto l1008
				l1009:
					{
						if position == len(p.Buffer) {
							goto l996
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks5]() {
								goto l1013
							}
							goto l996
						l1013:
							if !matchChar('`') {
								goto l996
							}
						l1014:
							if !matchChar('`') {
								goto l1015
							}
							goto l1014
						l1015:
							break
						default:
							{
								position1016 := position
								if !p.rules[ruleSp]() {
									goto l1016
								}
								if !p.rules[ruleTicks5]() {
									goto l1016
								}
								goto l996
							l1016:
								position = position1016
							}
							{
								if position == len(p.Buffer) {
									goto l996
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l996
									}
									if !p.rules[ruleBlankLine]() {
										goto l1018
									}
									goto l996
								l1018:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l996
									}
									break
								default:
									goto l996
								}
							}
						}
					}
				l1008:
					goto l995
				l996:
					position = position996
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l893
				}
				if !p.rules[ruleTicks5]() {
					goto l893
				}
			}
		l894:
			do(84)
			return true
		l893:
			position = position0
			return false
		},
		/* 190 RawHtml <- (< (HtmlComment / HtmlBlockScript / HtmlTag) > {   if p.extension.FilterHTML {
		        yy = p.mkList(LIST, nil)
		    } else {
		        yy = p.mkString(yytext)
		        yy.key = HTML
		    }
		}) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleHtmlComment]() {
				goto l1021
			}
			goto l1020
		l1021:
			if !p.rules[ruleHtmlBlockScript]() {
				goto l1022
			}
			goto l1020
		l1022:
			if !p.rules[ruleHtmlTag]() {
				goto l1019
			}
		l1020:
			end = position
			do(85)
			return true
		l1019:
			position = position0
			return false
		},
		/* 191 BlankLine <- (Sp Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1023
			}
			if !p.rules[ruleNewline]() {
				goto l1023
			}
			return true
		l1023:
			position = position0
			return false
		},
		/* 192 Quoted <- ((&[\'] ('\'' (!'\'' .)* '\'')) | (&[\"] ('"' (!'"' .)* '"'))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1024
				}
				switch p.Buffer[position] {
				case '\'':
					position++ // matchChar
				l1026:
					if position == len(p.Buffer) {
						goto l1027
					}
					switch p.Buffer[position] {
					case '\'':
						goto l1027
					default:
						position++
					}
					goto l1026
				l1027:
					if !matchChar('\'') {
						goto l1024
					}
					break
				case '"':
					position++ // matchChar
				l1028:
					if position == len(p.Buffer) {
						goto l1029
					}
					switch p.Buffer[position] {
					case '"':
						goto l1029
					default:
						position++
					}
					goto l1028
				l1029:
					if !matchChar('"') {
						goto l1024
					}
					break
				default:
					goto l1024
				}
			}
			return true
		l1024:
			position = position0
			return false
		},
		/* 193 HtmlAttribute <- (((&[\-] '-') | (&[0-9A-Za-z] [A-Za-z0-9]))+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1030
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				default:
					if !matchClass(5) {
						goto l1030
					}
				}
			}
		l1031:
			{
				if position == len(p.Buffer) {
					goto l1032
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				default:
					if !matchClass(5) {
						goto l1032
					}
				}
			}
			goto l1031
		l1032:
			if !p.rules[ruleSpnl]() {
				goto l1030
			}
			{
				position1035 := position
				if !matchChar('=') {
					goto l1035
				}
				if !p.rules[ruleSpnl]() {
					goto l1035
				}
				if !p.rules[ruleQuoted]() {
					goto l1038
				}
				goto l1037
			l1038:
				if peekChar('>') {
					goto l1035
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1035
				}
			l1039:
				if peekChar('>') {
					goto l1040
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1040
				}
				goto l1039
			l1040:
			l1037:
				goto l1036
			l1035:
				position = position1035
			}
		l1036:
			if !p.rules[ruleSpnl]() {
				goto l1030
			}
			return true
		l1030:
			position = position0
			return false
		},
		/* 194 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
		func() bool {
			position0 := position
			if !matchString("<!--") {
				goto l1041
			}
		l1042:
			{
				position1043 := position
				if !matchString("-->") {
					goto l1044
				}
				goto l1043
			l1044:
				if !matchDot() {
					goto l1043
				}
				goto l1042
			l1043:
				position = position1043
			}
			if !matchString("-->") {
				goto l1041
			}
			return true
		l1041:
			position = position0
			return false
		},
		/* 195 HtmlTag <- ('<' Spnl '/'? [A-Za-z0-9]+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l1045
			}
			if !p.rules[ruleSpnl]() {
				goto l1045
			}
			matchChar('/')
			if !matchClass(5) {
				goto l1045
			}
		l1046:
			if !matchClass(5) {
				goto l1047
			}
			goto l1046
		l1047:
			if !p.rules[ruleSpnl]() {
				goto l1045
			}
		l1048:
			if !p.rules[ruleHtmlAttribute]() {
				goto l1049
			}
			goto l1048
		l1049:
			matchChar('/')
			if !p.rules[ruleSpnl]() {
				goto l1045
			}
			if !matchChar('>') {
				goto l1045
			}
			return true
		l1045:
			position = position0
			return false
		},
		/* 196 Eof <- !. */
		func() bool {
			if position < len(p.Buffer) {
				goto l1050
			}
			return true
		l1050:
			return false
		},
		/* 197 Spacechar <- ((&[\t] '\t') | (&[ ] ' ')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1051
				}
				switch p.Buffer[position] {
				case '\t':
					position++ // matchChar
					break
				case ' ':
					position++ // matchChar
					break
				default:
					goto l1051
				}
			}
			return true
		l1051:
			return false
		},
		/* 198 Nonspacechar <- (!Spacechar !Newline .) */
		func() bool {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto l1054
			}
			goto l1053
		l1054:
			if !p.rules[ruleNewline]() {
				goto l1055
			}
			goto l1053
		l1055:
			if !matchDot() {
				goto l1053
			}
			return true
		l1053:
			position = position0
			return false
		},
		/* 199 Newline <- ((&[\r] ('\r' '\n'?)) | (&[\n] '\n')) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1056
				}
				switch p.Buffer[position] {
				case '\r':
					position++ // matchChar
					matchChar('\n')
					break
				case '\n':
					position++ // matchChar
					break
				default:
					goto l1056
				}
			}
			return true
		l1056:
			position = position0
			return false
		},
		/* 200 Sp <- Spacechar* */
		func() bool {
		l1059:
			if !p.rules[ruleSpacechar]() {
				goto l1060
			}
			goto l1059
		l1060:
			return true
		},
		/* 201 Spnl <- (Sp (Newline Sp)?) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1061
			}
			{
				position1062 := position
				if !p.rules[ruleNewline]() {
					goto l1062
				}
				if !p.rules[ruleSp]() {
					goto l1062
				}
				goto l1063
			l1062:
				position = position1062
			}
		l1063:
			return true
		l1061:
			position = position0
			return false
		},
		/* 202 SpecialChar <- ('\'' / '"' / ((&[\\] '\\') | (&[#] '#') | (&[!] '!') | (&[<] '<') | (&[)] ')') | (&[(] '(') | (&[\]] ']') | (&[\[] '[') | (&[&] '&') | (&[`] '`') | (&[_] '_') | (&[*] '*') | (&[\"\'\-.^] ExtendedSpecialChar))) */
		func() bool {
			if !matchChar('\'') {
				goto l1066
			}
			goto l1065
		l1066:
			if !matchChar('"') {
				goto l1067
			}
			goto l1065
		l1067:
			{
				if position == len(p.Buffer) {
					goto l1064
				}
				switch p.Buffer[position] {
				case '\\':
					position++ // matchChar
					break
				case '#':
					position++ // matchChar
					break
				case '!':
					position++ // matchChar
					break
				case '<':
					position++ // matchChar
					break
				case ')':
					position++ // matchChar
					break
				case '(':
					position++ // matchChar
					break
				case ']':
					position++ // matchChar
					break
				case '[':
					position++ // matchChar
					break
				case '&':
					position++ // matchChar
					break
				case '`':
					position++ // matchChar
					break
				case '_':
					position++ // matchChar
					break
				case '*':
					position++ // matchChar
					break
				default:
					if !p.rules[ruleExtendedSpecialChar]() {
						goto l1064
					}
				}
			}
		l1065:
			return true
		l1064:
			return false
		},
		/* 203 NormalChar <- (!((&[\n\r] Newline) | (&[\t ] Spacechar) | (&[!-#&-*\-.<\[-`] SpecialChar)) .) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1070
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l1070
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l1070
					}
					break
				default:
					if !p.rules[ruleSpecialChar]() {
						goto l1070
					}
				}
			}
			goto l1069
		l1070:
			if !matchDot() {
				goto l1069
			}
			return true
		l1069:
			position = position0
			return false
		},
		/* 204 Alphanumeric <- ((&[\377] '\377') | (&[\376] '\376') | (&[\375] '\375') | (&[\374] '\374') | (&[\373] '\373') | (&[\372] '\372') | (&[\371] '\371') | (&[\370] '\370') | (&[\367] '\367') | (&[\366] '\366') | (&[\365] '\365') | (&[\364] '\364') | (&[\363] '\363') | (&[\362] '\362') | (&[\361] '\361') | (&[\360] '\360') | (&[\357] '\357') | (&[\356] '\356') | (&[\355] '\355') | (&[\354] '\354') | (&[\353] '\353') | (&[\352] '\352') | (&[\351] '\351') | (&[\350] '\350') | (&[\347] '\347') | (&[\346] '\346') | (&[\345] '\345') | (&[\344] '\344') | (&[\343] '\343') | (&[\342] '\342') | (&[\341] '\341') | (&[\340] '\340') | (&[\337] '\337') | (&[\336] '\336') | (&[\335] '\335') | (&[\334] '\334') | (&[\333] '\333') | (&[\332] '\332') | (&[\331] '\331') | (&[\330] '\330') | (&[\327] '\327') | (&[\326] '\326') | (&[\325] '\325') | (&[\324] '\324') | (&[\323] '\323') | (&[\322] '\322') | (&[\321] '\321') | (&[\320] '\320') | (&[\317] '\317') | (&[\316] '\316') | (&[\315] '\315') | (&[\314] '\314') | (&[\313] '\313') | (&[\312] '\312') | (&[\311] '\311') | (&[\310] '\310') | (&[\307] '\307') | (&[\306] '\306') | (&[\305] '\305') | (&[\304] '\304') | (&[\303] '\303') | (&[\302] '\302') | (&[\301] '\301') | (&[\300] '\300') | (&[\277] '\277') | (&[\276] '\276') | (&[\275] '\275') | (&[\274] '\274') | (&[\273] '\273') | (&[\272] '\272') | (&[\271] '\271') | (&[\270] '\270') | (&[\267] '\267') | (&[\266] '\266') | (&[\265] '\265') | (&[\264] '\264') | (&[\263] '\263') | (&[\262] '\262') | (&[\261] '\261') | (&[\260] '\260') | (&[\257] '\257') | (&[\256] '\256') | (&[\255] '\255') | (&[\254] '\254') | (&[\253] '\253') | (&[\252] '\252') | (&[\251] '\251') | (&[\250] '\250') | (&[\247] '\247') | (&[\246] '\246') | (&[\245] '\245') | (&[\244] '\244') | (&[\243] '\243') | (&[\242] '\242') | (&[\241] '\241') | (&[\240] '\240') | (&[\237] '\237') | (&[\236] '\236') | (&[\235] '\235') | (&[\234] '\234') | (&[\233] '\233') | (&[\232] '\232') | (&[\231] '\231') | (&[\230] '\230') | (&[\227] '\227') | (&[\226] '\226') | (&[\225] '\225') | (&[\224] '\224') | (&[\223] '\223') | (&[\222] '\222') | (&[\221] '\221') | (&[\220] '\220') | (&[\217] '\217') | (&[\216] '\216') | (&[\215] '\215') | (&[\214] '\214') | (&[\213] '\213') | (&[\212] '\212') | (&[\211] '\211') | (&[\210] '\210') | (&[\207] '\207') | (&[\206] '\206') | (&[\205] '\205') | (&[\204] '\204') | (&[\203] '\203') | (&[\202] '\202') | (&[\201] '\201') | (&[\200] '\200') | (&[0-9A-Za-z] [0-9A-Za-z])) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1072
				}
				switch p.Buffer[position] {
				case '\377':
					position++ // matchChar
					break
				case '\376':
					position++ // matchChar
					break
				case '\375':
					position++ // matchChar
					break
				case '\374':
					position++ // matchChar
					break
				case '\373':
					position++ // matchChar
					break
				case '\372':
					position++ // matchChar
					break
				case '\371':
					position++ // matchChar
					break
				case '\370':
					position++ // matchChar
					break
				case '\367':
					position++ // matchChar
					break
				case '\366':
					position++ // matchChar
					break
				case '\365':
					position++ // matchChar
					break
				case '\364':
					position++ // matchChar
					break
				case '\363':
					position++ // matchChar
					break
				case '\362':
					position++ // matchChar
					break
				case '\361':
					position++ // matchChar
					break
				case '\360':
					position++ // matchChar
					break
				case '\357':
					position++ // matchChar
					break
				case '\356':
					position++ // matchChar
					break
				case '\355':
					position++ // matchChar
					break
				case '\354':
					position++ // matchChar
					break
				case '\353':
					position++ // matchChar
					break
				case '\352':
					position++ // matchChar
					break
				case '\351':
					position++ // matchChar
					break
				case '\350':
					position++ // matchChar
					break
				case '\347':
					position++ // matchChar
					break
				case '\346':
					position++ // matchChar
					break
				case '\345':
					position++ // matchChar
					break
				case '\344':
					position++ // matchChar
					break
				case '\343':
					position++ // matchChar
					break
				case '\342':
					position++ // matchChar
					break
				case '\341':
					position++ // matchChar
					break
				case '\340':
					position++ // matchChar
					break
				case '\337':
					position++ // matchChar
					break
				case '\336':
					position++ // matchChar
					break
				case '\335':
					position++ // matchChar
					break
				case '\334':
					position++ // matchChar
					break
				case '\333':
					position++ // matchChar
					break
				case '\332':
					position++ // matchChar
					break
				case '\331':
					position++ // matchChar
					break
				case '\330':
					position++ // matchChar
					break
				case '\327':
					position++ // matchChar
					break
				case '\326':
					position++ // matchChar
					break
				case '\325':
					position++ // matchChar
					break
				case '\324':
					position++ // matchChar
					break
				case '\323':
					position++ // matchChar
					break
				case '\322':
					position++ // matchChar
					break
				case '\321':
					position++ // matchChar
					break
				case '\320':
					position++ // matchChar
					break
				case '\317':
					position++ // matchChar
					break
				case '\316':
					position++ // matchChar
					break
				case '\315':
					position++ // matchChar
					break
				case '\314':
					position++ // matchChar
					break
				case '\313':
					position++ // matchChar
					break
				case '\312':
					position++ // matchChar
					break
				case '\311':
					position++ // matchChar
					break
				case '\310':
					position++ // matchChar
					break
				case '\307':
					position++ // matchChar
					break
				case '\306':
					position++ // matchChar
					break
				case '\305':
					position++ // matchChar
					break
				case '\304':
					position++ // matchChar
					break
				case '\303':
					position++ // matchChar
					break
				case '\302':
					position++ // matchChar
					break
				case '\301':
					position++ // matchChar
					break
				case '\300':
					position++ // matchChar
					break
				case '\277':
					position++ // matchChar
					break
				case '\276':
					position++ // matchChar
					break
				case '\275':
					position++ // matchChar
					break
				case '\274':
					position++ // matchChar
					break
				case '\273':
					position++ // matchChar
					break
				case '\272':
					position++ // matchChar
					break
				case '\271':
					position++ // matchChar
					break
				case '\270':
					position++ // matchChar
					break
				case '\267':
					position++ // matchChar
					break
				case '\266':
					position++ // matchChar
					break
				case '\265':
					position++ // matchChar
					break
				case '\264':
					position++ // matchChar
					break
				case '\263':
					position++ // matchChar
					break
				case '\262':
					position++ // matchChar
					break
				case '\261':
					position++ // matchChar
					break
				case '\260':
					position++ // matchChar
					break
				case '\257':
					position++ // matchChar
					break
				case '\256':
					position++ // matchChar
					break
				case '\255':
					position++ // matchChar
					break
				case '\254':
					position++ // matchChar
					break
				case '\253':
					position++ // matchChar
					break
				case '\252':
					position++ // matchChar
					break
				case '\251':
					position++ // matchChar
					break
				case '\250':
					position++ // matchChar
					break
				case '\247':
					position++ // matchChar
					break
				case '\246':
					position++ // matchChar
					break
				case '\245':
					position++ // matchChar
					break
				case '\244':
					position++ // matchChar
					break
				case '\243':
					position++ // matchChar
					break
				case '\242':
					position++ // matchChar
					break
				case '\241':
					position++ // matchChar
					break
				case '\240':
					position++ // matchChar
					break
				case '\237':
					position++ // matchChar
					break
				case '\236':
					position++ // matchChar
					break
				case '\235':
					position++ // matchChar
					break
				case '\234':
					position++ // matchChar
					break
				case '\233':
					position++ // matchChar
					break
				case '\232':
					position++ // matchChar
					break
				case '\231':
					position++ // matchChar
					break
				case '\230':
					position++ // matchChar
					break
				case '\227':
					position++ // matchChar
					break
				case '\226':
					position++ // matchChar
					break
				case '\225':
					position++ // matchChar
					break
				case '\224':
					position++ // matchChar
					break
				case '\223':
					position++ // matchChar
					break
				case '\222':
					position++ // matchChar
					break
				case '\221':
					position++ // matchChar
					break
				case '\220':
					position++ // matchChar
					break
				case '\217':
					position++ // matchChar
					break
				case '\216':
					position++ // matchChar
					break
				case '\215':
					position++ // matchChar
					break
				case '\214':
					position++ // matchChar
					break
				case '\213':
					position++ // matchChar
					break
				case '\212':
					position++ // matchChar
					break
				case '\211':
					position++ // matchChar
					break
				case '\210':
					position++ // matchChar
					break
				case '\207':
					position++ // matchChar
					break
				case '\206':
					position++ // matchChar
					break
				case '\205':
					position++ // matchChar
					break
				case '\204':
					position++ // matchChar
					break
				case '\203':
					position++ // matchChar
					break
				case '\202':
					position++ // matchChar
					break
				case '\201':
					position++ // matchChar
					break
				case '\200':
					position++ // matchChar
					break
				default:
					if !matchClass(4) {
						goto l1072
					}
				}
			}
			return true
		l1072:
			return false
		},
		/* 205 AlphanumericAscii <- [A-Za-z0-9] */
		func() bool {
			if !matchClass(5) {
				goto l1074
			}
			return true
		l1074:
			return false
		},
		/* 206 Digit <- [0-9] */
		func() bool {
			if !matchClass(0) {
				goto l1075
			}
			return true
		l1075:
			return false
		},
		/* 207 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1076
			}
			if !matchChar('#') {
				goto l1076
			}
			if !matchClass(6) {
				goto l1076
			}
			if !matchClass(7) {
				goto l1076
			}
		l1077:
			if !matchClass(7) {
				goto l1078
			}
			goto l1077
		l1078:
			if !matchChar(';') {
				goto l1076
			}
			end = position
			return true
		l1076:
			position = position0
			return false
		},
		/* 208 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1079
			}
			if !matchChar('#') {
				goto l1079
			}
			if !matchClass(0) {
				goto l1079
			}
		l1080:
			if !matchClass(0) {
				goto l1081
			}
			goto l1080
		l1081:
			end = position
			if !matchChar(';') {
				goto l1079
			}
			end = position
			return true
		l1079:
			position = position0
			return false
		},
		/* 209 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1082
			}
			if !matchClass(5) {
				goto l1082
			}
		l1083:
			if !matchClass(5) {
				goto l1084
			}
			goto l1083
		l1084:
			if !matchChar(';') {
				goto l1082
			}
			end = position
			return true
		l1082:
			position = position0
			return false
		},
		/* 210 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() bool {
			if !matchString("   ") {
				goto l1087
			}
			goto l1086
		l1087:
			if !matchString("  ") {
				goto l1088
			}
			goto l1086
		l1088:
			if !matchChar(' ') {
				goto l1089
			}
			goto l1086
		l1089:
		l1086:
			return true
		},
		/* 211 Indent <- ((&[ ] '    ') | (&[\t] '\t')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1090
				}
				switch p.Buffer[position] {
				case ' ':
					position++
					if !matchString("   ") {
						goto l1090
					}
					break
				case '\t':
					position++ // matchChar
					break
				default:
					goto l1090
				}
			}
			return true
		l1090:
			return false
		},
		/* 212 IndentedLine <- (Indent Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1092
			}
			if !p.rules[ruleLine]() {
				goto l1092
			}
			return true
		l1092:
			position = position0
			return false
		},
		/* 213 OptionallyIndentedLine <- (Indent? Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1094
			}
		l1094:
			if !p.rules[ruleLine]() {
				goto l1093
			}
			return true
		l1093:
			position = position0
			return false
		},
		/* 214 StartList <- (&. { yy = nil }) */
		func() bool {
			if !(position < len(p.Buffer)) {
				goto l1096
			}
			do(86)
			return true
		l1096:
			return false
		},
		/* 215 Line <- (RawLine { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRawLine]() {
				goto l1097
			}
			do(87)
			return true
		l1097:
			position = position0
			return false
		},
		/* 216 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > !.)) */
		func() bool {
			position0 := position
			{
				position1099 := position
				begin = position
			l1101:
				if position == len(p.Buffer) {
					goto l1102
				}
				switch p.Buffer[position] {
				case '\r', '\n':
					goto l1102
				default:
					position++
				}
				goto l1101
			l1102:
				if !p.rules[ruleNewline]() {
					goto l1100
				}
				end = position
				goto l1099
			l1100:
				position = position1099
				begin = position
				if !matchDot() {
					goto l1098
				}
			l1103:
				if !matchDot() {
					goto l1104
				}
				goto l1103
			l1104:
				end = position
				if position < len(p.Buffer) {
					goto l1098
				}
			}
		l1099:
			return true
		l1098:
			position = position0
			return false
		},
		/* 217 SkipBlock <- (HtmlBlock / ((!'#' !SetextBottom1 !SetextBottom2 !BlankLine RawLine)+ BlankLine*) / BlankLine+ / RawLine) */
		func() bool {
			position0 := position
			{
				position1106 := position
				if !p.rules[ruleHtmlBlock]() {
					goto l1107
				}
				goto l1106
			l1107:
				if peekChar('#') {
					goto l1108
				}
				if !p.rules[ruleSetextBottom1]() {
					goto l1111
				}
				goto l1108
			l1111:
				if !p.rules[ruleSetextBottom2]() {
					goto l1112
				}
				goto l1108
			l1112:
				if !p.rules[ruleBlankLine]() {
					goto l1113
				}
				goto l1108
			l1113:
				if !p.rules[ruleRawLine]() {
					goto l1108
				}
			l1109:
				{
					position1110 := position
					if peekChar('#') {
						goto l1110
					}
					if !p.rules[ruleSetextBottom1]() {
						goto l1114
					}
					goto l1110
				l1114:
					if !p.rules[ruleSetextBottom2]() {
						goto l1115
					}
					goto l1110
				l1115:
					if !p.rules[ruleBlankLine]() {
						goto l1116
					}
					goto l1110
				l1116:
					if !p.rules[ruleRawLine]() {
						goto l1110
					}
					goto l1109
				l1110:
					position = position1110
				}
			l1117:
				if !p.rules[ruleBlankLine]() {
					goto l1118
				}
				goto l1117
			l1118:
				goto l1106
			l1108:
				position = position1106
				if !p.rules[ruleBlankLine]() {
					goto l1119
				}
			l1120:
				if !p.rules[ruleBlankLine]() {
					goto l1121
				}
				goto l1120
			l1121:
				goto l1106
			l1119:
				position = position1106
				if !p.rules[ruleRawLine]() {
					goto l1105
				}
			}
		l1106:
			return true
		l1105:
			position = position0
			return false
		},
		/* 218 ExtendedSpecialChar <- ((&[^] (&{p.extension.Notes} '^')) | (&[\"\'\-.] (&{p.extension.Smart} ((&[\"] '"') | (&[\'] '\'') | (&[\-] '-') | (&[.] '.'))))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1122
				}
				switch p.Buffer[position] {
				case '^':
					if !(p.extension.Notes) {
						goto l1122
					}
					if !matchChar('^') {
						goto l1122
					}
					break
				default:
					if !(p.extension.Smart) {
						goto l1122
					}
					{
						if position == len(p.Buffer) {
							goto l1122
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
							break
						case '\'':
							position++ // matchChar
							break
						case '-':
							position++ // matchChar
							break
						case '.':
							position++ // matchChar
							break
						default:
							goto l1122
						}
					}
				}
			}
			return true
		l1122:
			position = position0
			return false
		},
		/* 219 Smart <- (&{p.extension.Smart} (SingleQuoted / ((&[\'] Apostrophe) | (&[\"] DoubleQuoted) | (&[\-] Dash) | (&[.] Ellipsis)))) */
		func() bool {
			if !(p.extension.Smart) {
				goto l1125
			}
			if !p.rules[ruleSingleQuoted]() {
				goto l1127
			}
			goto l1126
		l1127:
			{
				if position == len(p.Buffer) {
					goto l1125
				}
				switch p.Buffer[position] {
				case '\'':
					if !p.rules[ruleApostrophe]() {
						goto l1125
					}
					break
				case '"':
					if !p.rules[ruleDoubleQuoted]() {
						goto l1125
					}
					break
				case '-':
					if !p.rules[ruleDash]() {
						goto l1125
					}
					break
				case '.':
					if !p.rules[ruleEllipsis]() {
						goto l1125
					}
					break
				default:
					goto l1125
				}
			}
		l1126:
			return true
		l1125:
			return false
		},
		/* 220 Apostrophe <- ('\'' { yy = p.mkElem(APOSTROPHE) }) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1129
			}
			do(88)
			return true
		l1129:
			position = position0
			return false
		},
		/* 221 Ellipsis <- (('...' / '. . .') { yy = p.mkElem(ELLIPSIS) }) */
		func() bool {
			position0 := position
			if !matchString("...") {
				goto l1132
			}
			goto l1131
		l1132:
			if !matchString(". . .") {
				goto l1130
			}
		l1131:
			do(89)
			return true
		l1130:
			position = position0
			return false
		},
		/* 222 Dash <- (EmDash / EnDash) */
		func() bool {
			if !p.rules[ruleEmDash]() {
				goto l1135
			}
			goto l1134
		l1135:
			if !p.rules[ruleEnDash]() {
				goto l1133
			}
		l1134:
			return true
		l1133:
			return false
		},
		/* 223 EnDash <- ('-' &[0-9] { yy = p.mkElem(ENDASH) }) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l1136
			}
			if !peekClass(0) {
				goto l1136
			}
			do(90)
			return true
		l1136:
			position = position0
			return false
		},
		/* 224 EmDash <- (('---' / '--') { yy = p.mkElem(EMDASH) }) */
		func() bool {
			position0 := position
			if !matchString("---") {
				goto l1139
			}
			goto l1138
		l1139:
			if !matchString("--") {
				goto l1137
			}
		l1138:
			do(91)
			return true
		l1137:
			position = position0
			return false
		},
		/* 225 SingleQuoteStart <- ('\'' !((&[\n\r] Newline) | (&[\t ] Spacechar))) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1140
			}
			{
				if position == len(p.Buffer) {
					goto l1141
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l1141
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l1141
					}
					break
				default:
					goto l1141
				}
			}
			goto l1140
		l1141:
			return true
		l1140:
			position = position0
			return false
		},
		/* 226 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1143
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l1144
			}
			goto l1143
		l1144:
			return true
		l1143:
			position = position0
			return false
		},
		/* 227 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = p.mkList(SINGLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto l1145
			}
			if !p.rules[ruleStartList]() {
				goto l1145
			}
			doarg(yySet, -1)
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1148
			}
			goto l1145
		l1148:
			if !p.rules[ruleInline]() {
				goto l1145
			}
			doarg(yySet, -2)
			do(92)
		l1146:
			{
				position1147, thunkPosition1147 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto l1149
				}
				goto l1147
			l1149:
				if !p.rules[ruleInline]() {
					goto l1147
				}
				doarg(yySet, -2)
				do(92)
				goto l1146
			l1147:
				position, thunkPosition = position1147, thunkPosition1147
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1145
			}
			do(93)
			doarg(yyPop, 2)
			return true
		l1145:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 228 DoubleQuoteStart <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1150
			}
			return true
		l1150:
			return false
		},
		/* 229 DoubleQuoteEnd <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1151
			}
			return true
		l1151:
			return false
		},
		/* 230 DoubleQuoted <- ('"' StartList (!'"' Inline { a = cons(b, a) })+ '"' { yy = p.mkList(DOUBLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('"') {
				goto l1152
			}
			if !p.rules[ruleStartList]() {
				goto l1152
			}
			doarg(yySet, -2)
			if peekChar('"') {
				goto l1152
			}
			if !p.rules[ruleInline]() {
				goto l1152
			}
			doarg(yySet, -1)
			do(94)
		l1153:
			{
				position1154, thunkPosition1154 := position, thunkPosition
				if peekChar('"') {
					goto l1154
				}
				if !p.rules[ruleInline]() {
					goto l1154
				}
				doarg(yySet, -1)
				do(94)
				goto l1153
			l1154:
				position, thunkPosition = position1154, thunkPosition1154
			}
			if !matchChar('"') {
				goto l1152
			}
			do(95)
			doarg(yyPop, 2)
			return true
		l1152:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 231 NoteReference <- (&{p.extension.Notes} RawNoteReference {
		    if match, ok := p.find_note(ref.contents.str); ok {
		        yy = p.mkElem(NOTE)
		        yy.children = match.children
		        yy.contents.str = ""
		    } else {
		        yy = p.mkString("[^"+ref.contents.str+"]")
		    }
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto l1155
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1155
			}
			doarg(yySet, -1)
			do(96)
			doarg(yyPop, 1)
			return true
		l1155:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 232 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !matchString("[^") {
				goto l1156
			}
			begin = position
			if !p.rules[ruleNewline]() {
				goto l1159
			}
			goto l1156
		l1159:
			if peekChar(']') {
				goto l1156
			}
			if !matchDot() {
				goto l1156
			}
		l1157:
			{
				position1158 := position
				if !p.rules[ruleNewline]() {
					goto l1160
				}
				goto l1158
			l1160:
				if peekChar(']') {
					goto l1158
				}
				if !matchDot() {
					goto l1158
				}
				goto l1157
			l1158:
				position = position1158
			}
			end = position
			if !matchChar(']') {
				goto l1156
			}
			do(97)
			return true
		l1156:
			position = position0
			return false
		},
		/* 233 Note <- (&{p.extension.Notes} NonindentSpace RawNoteReference ':' Sp StartList (RawNoteBlock { a = cons(yy, a) }) (&Indent RawNoteBlock { a = cons(yy, a) })* {   yy = p.mkList(NOTE, a)
		    yy.contents.str = ref.contents.str
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !(p.extension.Notes) {
				goto l1161
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l1161
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1161
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto l1161
			}
			if !p.rules[ruleSp]() {
				goto l1161
			}
			if !p.rules[ruleStartList]() {
				goto l1161
			}
			doarg(yySet, -2)
			if !p.rules[ruleRawNoteBlock]() {
				goto l1161
			}
			do(98)
		l1162:
			{
				position1163, thunkPosition1163 := position, thunkPosition
				{
					position1164 := position
					if !p.rules[ruleIndent]() {
						goto l1163
					}
					position = position1164
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto l1163
				}
				do(99)
				goto l1162
			l1163:
				position, thunkPosition = position1163, thunkPosition1163
			}
			do(100)
			doarg(yyPop, 2)
			return true
		l1161:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 234 InlineNote <- (&{p.extension.Notes} '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = p.mkList(NOTE, a)
		   yy.contents.str = "" }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto l1165
			}
			if !matchString("^[") {
				goto l1165
			}
			if !p.rules[ruleStartList]() {
				goto l1165
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto l1165
			}
			if !p.rules[ruleInline]() {
				goto l1165
			}
			do(101)
		l1166:
			{
				position1167 := position
				if peekChar(']') {
					goto l1167
				}
				if !p.rules[ruleInline]() {
					goto l1167
				}
				do(101)
				goto l1166
			l1167:
				position = position1167
			}
			if !matchChar(']') {
				goto l1165
			}
			do(102)
			doarg(yyPop, 1)
			return true
		l1165:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 235 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1168
			}
			doarg(yySet, -1)
		l1169:
			{
				position1170, thunkPosition1170 := position, thunkPosition
				{
					position1171, thunkPosition1171 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto l1172
					}
					doarg(yySet, -2)
					do(103)
					goto l1171
				l1172:
					position, thunkPosition = position1171, thunkPosition1171
					if !p.rules[ruleSkipBlock]() {
						goto l1170
					}
				}
			l1171:
				goto l1169
			l1170:
				position, thunkPosition = position1170, thunkPosition1170
			}
			do(104)
			if !(commit(thunkPosition0)) {
				goto l1168
			}
			doarg(yyPop, 2)
			return true
		l1168:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 236 RawNoteBlock <- (StartList (!BlankLine OptionallyIndentedLine { a = cons(yy, a) })+ (< BlankLine* > { a = cons(p.mkString(yytext), a) }) {   yy = p.mkStringFromList(a, true)
		    yy.key = RAW
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1173
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l1176
			}
			goto l1173
		l1176:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l1173
			}
			do(105)
		l1174:
			{
				position1175 := position
				if !p.rules[ruleBlankLine]() {
					goto l1177
				}
				goto l1175
			l1177:
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto l1175
				}
				do(105)
				goto l1174
			l1175:
				position = position1175
			}
			begin = position
		l1178:
			if !p.rules[ruleBlankLine]() {
				goto l1179
			}
			goto l1178
		l1179:
			end = position
			do(106)
			do(107)
			doarg(yyPop, 1)
			return true
		l1173:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 237 DefinitionList <- (&{p.extension.Dlists} StartList (Definition { a = cons(yy, a) })+ { yy = p.mkList(DEFINITIONLIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Dlists) {
				goto l1180
			}
			if !p.rules[ruleStartList]() {
				goto l1180
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto l1180
			}
			do(108)
		l1181:
			{
				position1182, thunkPosition1182 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto l1182
				}
				do(108)
				goto l1181
			l1182:
				position, thunkPosition = position1182, thunkPosition1182
			}
			do(109)
			doarg(yyPop, 1)
			return true
		l1180:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 238 Definition <- (&(NonindentSpace !Defmark Nonspacechar RawLine BlankLine? Defmark) StartList (DListTitle { a = cons(yy, a) })+ (DefTight / DefLoose) {
			for e := yy.children; e != nil; e = e.next {
				e.key = DEFDATA
			}
			a = cons(yy, a)
		} { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1184 := position
				if !p.rules[ruleNonindentSpace]() {
					goto l1183
				}
				if !p.rules[ruleDefmark]() {
					goto l1185
				}
				goto l1183
			l1185:
				if !p.rules[ruleNonspacechar]() {
					goto l1183
				}
				if !p.rules[ruleRawLine]() {
					goto l1183
				}
				if !p.rules[ruleBlankLine]() {
					goto l1186
				}
			l1186:
				if !p.rules[ruleDefmark]() {
					goto l1183
				}
				position = position1184
			}
			if !p.rules[ruleStartList]() {
				goto l1183
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto l1183
			}
			do(110)
		l1188:
			{
				position1189, thunkPosition1189 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto l1189
				}
				do(110)
				goto l1188
			l1189:
				position, thunkPosition = position1189, thunkPosition1189
			}
			if !p.rules[ruleDefTight]() {
				goto l1191
			}
			goto l1190
		l1191:
			if !p.rules[ruleDefLoose]() {
				goto l1183
			}
		l1190:
			do(111)
			do(112)
			doarg(yyPop, 1)
			return true
		l1183:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 239 DListTitle <- (NonindentSpace !Defmark &Nonspacechar StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline {	yy = p.mkList(LIST, a)
			yy.key = DEFTITLE
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l1192
			}
			if !p.rules[ruleDefmark]() {
				goto l1193
			}
			goto l1192
		l1193:
			{
				position1194 := position
				if !p.rules[ruleNonspacechar]() {
					goto l1192
				}
				position = position1194
			}
			if !p.rules[ruleStartList]() {
				goto l1192
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l1197
			}
			goto l1192
		l1197:
			if !p.rules[ruleInline]() {
				goto l1192
			}
			do(113)
		l1195:
			{
				position1196 := position
				if !p.rules[ruleEndline]() {
					goto l1198
				}
				goto l1196
			l1198:
				if !p.rules[ruleInline]() {
					goto l1196
				}
				do(113)
				goto l1195
			l1196:
				position = position1196
			}
			if !p.rules[ruleSp]() {
				goto l1192
			}
			if !p.rules[ruleNewline]() {
				goto l1192
			}
			do(114)
			doarg(yyPop, 1)
			return true
		l1192:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 240 DefTight <- (&Defmark ListTight) */
		func() bool {
			{
				position1200 := position
				if !p.rules[ruleDefmark]() {
					goto l1199
				}
				position = position1200
			}
			if !p.rules[ruleListTight]() {
				goto l1199
			}
			return true
		l1199:
			return false
		},
		/* 241 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l1201
			}
			{
				position1202 := position
				if !p.rules[ruleDefmark]() {
					goto l1201
				}
				position = position1202
			}
			if !p.rules[ruleListLoose]() {
				goto l1201
			}
			return true
		l1201:
			position = position0
			return false
		},
		/* 242 Defmark <- (NonindentSpace ((&[~] '~') | (&[:] ':')) Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l1203
			}
			{
				if position == len(p.Buffer) {
					goto l1203
				}
				switch p.Buffer[position] {
				case '~':
					position++ // matchChar
					break
				case ':':
					position++ // matchChar
					break
				default:
					goto l1203
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l1203
			}
		l1205:
			if !p.rules[ruleSpacechar]() {
				goto l1206
			}
			goto l1205
		l1206:
			return true
		l1203:
			position = position0
			return false
		},
		/* 243 DefMarker <- (&{p.extension.Dlists} Defmark) */
		func() bool {
			if !(p.extension.Dlists) {
				goto l1207
			}
			if !p.rules[ruleDefmark]() {
				goto l1207
			}
			return true
		l1207:
			return false
		},
	}
}

/*
 * List manipulation functions
 */

/* cons - cons an element onto a list, returning pointer to new head
 */
func cons(new, list *element) *element {
	new.next = list
	return new
}

/* reverse - reverse a list, returning pointer to new list
 */
func reverse(list *element) (new *element) {
	for list != nil {
		next := list.next
		new = cons(list, new)
		list = next
	}
	return
}

/*
 *  Auxiliary functions for parsing actions.
 *  These make it easier to build up data structures (including lists)
 *  in the parsing actions.
 */

/* p.mkElem - generic constructor for element
 */
func (p *yyParser) mkElem(key int) *element {
	r := p.state.heap.row
	if len(r) == 0 {
		r = p.state.heap.nextRow()
	}
	e := &r[0]
	*e = element{}
	p.state.heap.row = r[1:]
	e.key = key
	return e
}

/* p.mkString - constructor for STR element
 */
func (p *yyParser) mkString(s string) (result *element) {
	result = p.mkElem(STR)
	result.contents.str = s
	return
}

/* p.mkStringFromList - makes STR element by concatenating a
 * reversed list of strings, adding optional extra newline
 */
func (p *yyParser) mkStringFromList(list *element, extra_newline bool) (result *element) {
	s := ""
	for list = reverse(list); list != nil; list = list.next {
		s += list.contents.str
	}

	if extra_newline {
		s += "\n"
	}
	result = p.mkElem(STR)
	result.contents.str = s
	return
}

/* p.mkList - makes new list with key 'key' and children the reverse of 'lst'.
 * This is designed to be used with cons to build lists in a parser action.
 * The reversing is necessary because cons adds to the head of a list.
 */
func (p *yyParser) mkList(key int, lst *element) (el *element) {
	el = p.mkElem(key)
	el.children = reverse(lst)
	return
}

/* p.mkLink - constructor for LINK element
 */
func (p *yyParser) mkLink(label *element, url, title string) (el *element) {
	el = p.mkElem(LINK)
	el.contents.link = &link{label: label, url: url, title: title}
	return
}

/* match_inlines - returns true if inline lists match (case-insensitive...)
 */
func match_inlines(l1, l2 *element) bool {
	for l1 != nil && l2 != nil {
		if l1.key != l2.key {
			return false
		}
		switch l1.key {
		case SPACE, LINEBREAK, ELLIPSIS, EMDASH, ENDASH, APOSTROPHE:
			break
		case CODE, STR, HTML:
			if strings.ToUpper(l1.contents.str) != strings.ToUpper(l2.contents.str) {
				return false
			}
		case EMPH, STRONG, LIST, SINGLEQUOTED, DOUBLEQUOTED:
			if !match_inlines(l1.children, l2.children) {
				return false
			}
		case LINK, IMAGE:
			return false /* No links or images within links */
		default:
			log.Fatalf("match_inlines encountered unknown key = %d\n", l1.key)
		}
		l1 = l1.next
		l2 = l2.next
	}
	return l1 == nil && l2 == nil /* return true if both lists exhausted */
}

/* find_reference - return true if link found in references matching label.
 * 'link' is modified with the matching url and title.
 */
func (p *yyParser) findReference(label *element) (*link, bool) {
	for cur := p.references; cur != nil; cur = cur.next {
		l := cur.contents.link
		if match_inlines(label, l.label) {
			return l, true
		}
	}
	return nil, false
}

/* find_note - return true if note found in notes matching label.
 * if found, 'result' is set to point to matched note.
 */
func (p *yyParser) find_note(label string) (*element, bool) {
	for el := p.notes; el != nil; el = el.next {
		if label == el.contents.str {
			return el, true
		}
	}
	return nil, false
}

/* print tree of elements, for debugging only.
 */
func print_tree(w io.Writer, elt *element, indent int) {
	var key string

	for elt != nil {
		for i := 0; i < indent; i++ {
			fmt.Fprint(w, "\t")
		}
		key = keynames[elt.key]
		if key == "" {
			key = "?"
		}
		if elt.key == STR {
			fmt.Fprintf(w, "%p:\t%s\t'%s'\n", elt, key, elt.contents.str)
		} else {
			fmt.Fprintf(w, "%p:\t%s %p\n", elt, key, elt.next)
		}
		if elt.children != nil {
			print_tree(w, elt.children, indent+1)
		}
		elt = elt.next
	}
}

var keynames = [numVAL]string{
	LIST:           "LIST",
	RAW:            "RAW",
	SPACE:          "SPACE",
	LINEBREAK:      "LINEBREAK",
	ELLIPSIS:       "ELLIPSIS",
	EMDASH:         "EMDASH",
	ENDASH:         "ENDASH",
	APOSTROPHE:     "APOSTROPHE",
	SINGLEQUOTED:   "SINGLEQUOTED",
	DOUBLEQUOTED:   "DOUBLEQUOTED",
	STR:            "STR",
	LINK:           "LINK",
	IMAGE:          "IMAGE",
	CODE:           "CODE",
	HTML:           "HTML",
	EMPH:           "EMPH",
	STRONG:         "STRONG",
	PLAIN:          "PLAIN",
	PARA:           "PARA",
	LISTITEM:       "LISTITEM",
	BULLETLIST:     "BULLETLIST",
	ORDEREDLIST:    "ORDEREDLIST",
	H1:             "H1",
	H2:             "H2",
	H3:             "H3",
	H4:             "H4",
	H5:             "H5",
	H6:             "H6",
	BLOCKQUOTE:     "BLOCKQUOTE",
	VERBATIM:       "VERBATIM",
	HTMLBLOCK:      "HTMLBLOCK",
	HRULE:          "HRULE",
	REFERENCE:      "REFERENCE",
	NOTE:           "NOTE",
	DEFINITIONLIST: "DEFINITIONLIST",
	DEFTITLE:       "DEFTITLE",
	DEFDATA:        "DEFDATA",
}
