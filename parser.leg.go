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
	parserIfaceVersion_17 = iota
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
	STRIKE
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
	ruleHtmlBlockOpenHead
	ruleHtmlBlockCloseHead
	ruleHtmlBlockHead
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
	ruleTwoTildeOpen
	ruleTwoTildeClose
	ruleStrike
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
	ruleTildeLine
)

type yyParser struct {
	state
	Buffer      string
	Min, Max    int
	rules       [251]func() bool
	commit      func(int) bool
	ResetBuffer func(string) string
}

func (p *yyParser) Parse(ruleId int) (err error) {
	if p.rules[ruleId]() {
		// Make sure thunkPosition is 0 (there may be a yyPop action on the stack).
		p.commit(0)
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
		if p.Min == p.Max {
			err = io.EOF
		} else {
			err = &unexpectedEOFError{after}
		}
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
			yyval[yyp-1] = s
			yyval[yyp-2] = a
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

			yyval[yyp-1] = a
			yyval[yyp-2] = b
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
			yyval[yyp-1] = a
			yyval[yyp-2] = c
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
			yyval[yyp-1] = a
			yyval[yyp-2] = b
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
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 63 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 64 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			yy = p.mkList(EMPH, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 65 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 66 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			yy = p.mkList(STRONG, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 67 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 68 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			yy = p.mkList(STRONG, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 69 TwoTildeClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = a
			yyval[yyp-1] = a
		},
		/* 70 Strike */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 71 Strike */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			yy = p.mkList(STRIKE, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 72 Image */
		func(yytext string, _ int) {
			if yy.key == LINK {
				yy.key = IMAGE
			} else {
				result := yy
				yy.children = cons(p.mkString("!"), result.children)
			}

		},
		/* 73 ReferenceLinkDouble */
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
		/* 74 ReferenceLinkSingle */
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
		/* 75 ExplicitLink */
		func(yytext string, _ int) {
			l := yyval[yyp-1]
			s := yyval[yyp-2]
			t := yyval[yyp-3]
			yy = p.mkLink(l.children, s.contents.str, t.contents.str)
			s = nil
			t = nil
			l = nil
			yyval[yyp-1] = l
			yyval[yyp-2] = s
			yyval[yyp-3] = t
		},
		/* 76 Source */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 77 Title */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 78 AutoLinkUrl */
		func(yytext string, _ int) {
			yy = p.mkLink(p.mkString(yytext), yytext, "")
		},
		/* 79 AutoLinkEmail */
		func(yytext string, _ int) {

			yy = p.mkLink(p.mkString(yytext), "mailto:"+yytext, "")

		},
		/* 80 Reference */
		func(yytext string, _ int) {
			l := yyval[yyp-1]
			s := yyval[yyp-2]
			t := yyval[yyp-3]
			yy = p.mkLink(l.children, s.contents.str, t.contents.str)
			s = nil
			t = nil
			l = nil
			yy.key = REFERENCE
			yyval[yyp-1] = l
			yyval[yyp-2] = s
			yyval[yyp-3] = t
		},
		/* 81 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 82 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(LIST, a)
			yyval[yyp-1] = a
		},
		/* 83 RefSrc */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
			yy.key = HTML
		},
		/* 84 RefTitle */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 85 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 86 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			p.references = reverse(a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 87 Code */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
			yy.key = CODE
		},
		/* 88 RawHtml */
		func(yytext string, _ int) {
			if p.extension.FilterHTML {
				yy = p.mkList(LIST, nil)
			} else {
				yy = p.mkString(yytext)
				yy.key = HTML
			}

		},
		/* 89 StartList */
		func(yytext string, _ int) {
			yy = nil
		},
		/* 90 Line */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 91 Apostrophe */
		func(yytext string, _ int) {
			yy = p.mkElem(APOSTROPHE)
		},
		/* 92 Ellipsis */
		func(yytext string, _ int) {
			yy = p.mkElem(ELLIPSIS)
		},
		/* 93 EnDash */
		func(yytext string, _ int) {
			yy = p.mkElem(ENDASH)
		},
		/* 94 EmDash */
		func(yytext string, _ int) {
			yy = p.mkElem(EMDASH)
		},
		/* 95 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 96 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			yy = p.mkList(SINGLEQUOTED, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 97 DoubleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 98 DoubleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			yy = p.mkList(DOUBLEQUOTED, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 99 NoteReference */
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
		/* 100 RawNoteReference */
		func(yytext string, _ int) {
			yy = p.mkString(yytext)
		},
		/* 101 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(yy, a)
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 102 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			a = cons(yy, a)
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 103 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			yy = p.mkList(NOTE, a)
			yy.contents.str = ref.contents.str

			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 104 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 105 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(NOTE, a)
			yy.contents.str = ""
			yyval[yyp-1] = a
		},
		/* 106 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			a = cons(b, a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 107 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			p.notes = reverse(a)
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 108 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 109 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(p.mkString(yytext), a)
			yyval[yyp-1] = a
		},
		/* 110 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkStringFromList(a, true)
			yy.key = RAW

			yyval[yyp-1] = a
		},
		/* 111 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 112 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(DEFINITIONLIST, a)
			yyval[yyp-1] = a
		},
		/* 113 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 114 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]

			for e := yy.children; e != nil; e = e.next {
				e.key = DEFDATA
			}
			a = cons(yy, a)

			yyval[yyp-1] = a
		},
		/* 115 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			yy = p.mkList(LIST, a)
			yyval[yyp-1] = a
		},
		/* 116 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			a = cons(yy, a)
			yyval[yyp-1] = a
		},
		/* 117 DListTitle */
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
		yyPush = 118 + iota
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

	p.commit = func(thunkPosition0 int) bool {
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
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
		loop:
			{
				position1 := position
				if !p.rules[ruleBlock]() {
					goto out
				}
				do(0)
				goto loop
			out:
				position = position1
			}
			do(1)
			if !(p.commit(thunkPosition0)) {
				goto ko
			}
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 1 Docblock <- (Block { p.tree = yy } commit) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlock]() {
				goto ko
			}
			do(2)
			if !(p.commit(thunkPosition0)) {
				goto ko
			}
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 2 Block <- (BlankLine* (BlockQuote / Verbatim / Note / Reference / HorizontalRule / Heading / DefinitionList / OrderedList / BulletList / HtmlBlock / StyleBlock / Para / Plain)) */
		func() (match bool) {
			position0 := position
		loop:
			if !p.rules[ruleBlankLine]() {
				goto out
			}
			goto loop
		out:
			if !p.rules[ruleBlockQuote]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleVerbatim]() {
				goto nextAlt5
			}
			goto ok
		nextAlt5:
			if !p.rules[ruleNote]() {
				goto nextAlt6
			}
			goto ok
		nextAlt6:
			if !p.rules[ruleReference]() {
				goto nextAlt7
			}
			goto ok
		nextAlt7:
			if !p.rules[ruleHorizontalRule]() {
				goto nextAlt8
			}
			goto ok
		nextAlt8:
			if !p.rules[ruleHeading]() {
				goto nextAlt9
			}
			goto ok
		nextAlt9:
			if !p.rules[ruleDefinitionList]() {
				goto nextAlt10
			}
			goto ok
		nextAlt10:
			if !p.rules[ruleOrderedList]() {
				goto nextAlt11
			}
			goto ok
		nextAlt11:
			if !p.rules[ruleBulletList]() {
				goto nextAlt12
			}
			goto ok
		nextAlt12:
			if !p.rules[ruleHtmlBlock]() {
				goto nextAlt13
			}
			goto ok
		nextAlt13:
			if !p.rules[ruleStyleBlock]() {
				goto nextAlt14
			}
			goto ok
		nextAlt14:
			if !p.rules[rulePara]() {
				goto nextAlt15
			}
			goto ok
		nextAlt15:
			if !p.rules[rulePlain]() {
				goto ko
			}
		ok:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 3 Para <- (NonindentSpace Inlines BlankLine+ { yy = a; yy.key = PARA }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto ko
			}
			if !p.rules[ruleInlines]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto ko
			}
		loop:
			if !p.rules[ruleBlankLine]() {
				goto out
			}
			goto loop
		out:
			do(3)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 4 Plain <- (Inlines { yy = a; yy.key = PLAIN }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleInlines]() {
				goto ko
			}
			doarg(yySet, -1)
			do(4)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 5 AtxInline <- (!Newline !(Sp '#'* Sp Newline) Inline) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleNewline]() {
				goto ok
			}
			goto ko
		ok:
			{
				position1 := position
				if !p.rules[ruleSp]() {
					goto ok2
				}
			loop:
				if !matchChar('#') {
					goto out
				}
				goto loop
			out:
				if !p.rules[ruleSp]() {
					goto ok2
				}
				if !p.rules[ruleNewline]() {
					goto ok2
				}
				goto ko
			ok2:
				position = position1
			}
			if !p.rules[ruleInline]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 6 AtxStart <- (&'#' < ('######' / '#####' / '####' / '###' / '##' / '#') > { yy = p.mkElem(H1 + (len(yytext) - 1)) }) */
		func() (match bool) {
			position0 := position
			if !peekChar('#') {
				goto ko
			}
			begin = position
			if !matchString("######") {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !matchString("#####") {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !matchString("####") {
				goto nextAlt4
			}
			goto ok
		nextAlt4:
			if !matchString("###") {
				goto nextAlt5
			}
			goto ok
		nextAlt5:
			if !matchString("##") {
				goto nextAlt6
			}
			goto ok
		nextAlt6:
			if !matchChar('#') {
				goto ko
			}
		ok:
			end = position
			do(5)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 7 AtxHeading <- (AtxStart Sp StartList (AtxInline { a = cons(yy, a) })+ (Sp '#'* Sp)? Newline { yy = p.mkList(s.key, a)
		   s = nil }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleAtxStart]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -2)
			if !p.rules[ruleAtxInline]() {
				goto ko
			}
			do(6)
		loop:
			{
				position1 := position
				if !p.rules[ruleAtxInline]() {
					goto out
				}
				do(6)
				goto loop
			out:
				position = position1
			}
			{
				position2 := position
				if !p.rules[ruleSp]() {
					goto ko3
				}
			loop5:
				if !matchChar('#') {
					goto out6
				}
				goto loop5
			out6:
				if !p.rules[ruleSp]() {
					goto ko3
				}
				goto ok
			ko3:
				position = position2
			}
		ok:
			if !p.rules[ruleNewline]() {
				goto ko
			}
			do(7)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 8 SetextHeading <- (SetextHeading1 / SetextHeading2) */
		func() (match bool) {
			if !p.rules[ruleSetextHeading1]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleSetextHeading2]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 9 SetextBottom1 <- ('='+ Newline) */
		func() (match bool) {
			position0 := position
			if !matchChar('=') {
				goto ko
			}
		loop:
			if !matchChar('=') {
				goto out
			}
			goto loop
		out:
			if !p.rules[ruleNewline]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 10 SetextBottom2 <- ('-'+ Newline) */
		func() (match bool) {
			position0 := position
			if !matchChar('-') {
				goto ko
			}
		loop:
			if !matchChar('-') {
				goto out
			}
			goto loop
		out:
			if !p.rules[ruleNewline]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 11 SetextHeading1 <- (&(RawLine SetextBottom1) StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline SetextBottom1 { yy = p.mkList(H1, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1 := position
				if !p.rules[ruleRawLine]() {
					goto ko
				}
				if !p.rules[ruleSetextBottom1]() {
					goto ko
				}
				position = position1
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleInline]() {
				goto ko
			}
			do(8)
		loop:
			{
				position2 := position
				if !p.rules[ruleEndline]() {
					goto ok5
				}
				goto out
			ok5:
				if !p.rules[ruleInline]() {
					goto out
				}
				do(8)
				goto loop
			out:
				position = position2
			}
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ko
			}
			if !p.rules[ruleSetextBottom1]() {
				goto ko
			}
			do(9)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 12 SetextHeading2 <- (&(RawLine SetextBottom2) StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline SetextBottom2 { yy = p.mkList(H2, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1 := position
				if !p.rules[ruleRawLine]() {
					goto ko
				}
				if !p.rules[ruleSetextBottom2]() {
					goto ko
				}
				position = position1
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleInline]() {
				goto ko
			}
			do(10)
		loop:
			{
				position2 := position
				if !p.rules[ruleEndline]() {
					goto ok5
				}
				goto out
			ok5:
				if !p.rules[ruleInline]() {
					goto out
				}
				do(10)
				goto loop
			out:
				position = position2
			}
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ko
			}
			if !p.rules[ruleSetextBottom2]() {
				goto ko
			}
			do(11)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 13 Heading <- (SetextHeading / AtxHeading) */
		func() (match bool) {
			if !p.rules[ruleSetextHeading]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleAtxHeading]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 14 BlockQuote <- (BlockQuoteRaw {  yy = p.mkElem(BLOCKQUOTE)
		   yy.children = a
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleBlockQuoteRaw]() {
				goto ko
			}
			doarg(yySet, -1)
			do(12)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 15 BlockQuoteRaw <- (StartList ('>' ' '? Line { a = cons(yy, a) } (!'>' !BlankLine Line { a = cons(yy, a) })* (BlankLine { a = cons(p.mkString("\n"), a) })*)+ {   yy = p.mkStringFromList(a, true)
		    yy.key = RAW
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchChar('>') {
				goto ko
			}
			matchChar(' ')
			if !p.rules[ruleLine]() {
				goto ko
			}
			do(13)
		loop3:
			{
				position1, thunkPosition1 := position, thunkPosition
				if peekChar('>') {
					goto out4
				}
				if !p.rules[ruleBlankLine]() {
					goto ok
				}
				goto out4
			ok:
				if !p.rules[ruleLine]() {
					goto out4
				}
				do(14)
				goto loop3
			out4:
				position, thunkPosition = position1, thunkPosition1
			}
		loop6:
			{
				position2 := position
				if !p.rules[ruleBlankLine]() {
					goto out7
				}
				do(15)
				goto loop6
			out7:
				position = position2
			}
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchChar('>') {
					goto out
				}
				matchChar(' ')
				if !p.rules[ruleLine]() {
					goto out
				}
				do(13)
			loop8:
				{
					position4, thunkPosition4 := position, thunkPosition
					if peekChar('>') {
						goto out9
					}
					if !p.rules[ruleBlankLine]() {
						goto ok10
					}
					goto out9
				ok10:
					if !p.rules[ruleLine]() {
						goto out9
					}
					do(14)
					goto loop8
				out9:
					position, thunkPosition = position4, thunkPosition4
				}
			loop11:
				{
					position5 := position
					if !p.rules[ruleBlankLine]() {
						goto out12
					}
					do(15)
					goto loop11
				out12:
					position = position5
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			do(16)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 16 NonblankIndentedLine <- (!BlankLine IndentedLine) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleIndentedLine]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 17 VerbatimChunk <- (StartList (BlankLine { a = cons(p.mkString("\n"), a) })* (NonblankIndentedLine { a = cons(yy, a) })+ { yy = p.mkStringFromList(a, false) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
		loop:
			{
				position1 := position
				if !p.rules[ruleBlankLine]() {
					goto out
				}
				do(17)
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleNonblankIndentedLine]() {
				goto ko
			}
			do(18)
		loop3:
			{
				position2 := position
				if !p.rules[ruleNonblankIndentedLine]() {
					goto out4
				}
				do(18)
				goto loop3
			out4:
				position = position2
			}
			do(19)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 18 Verbatim <- (StartList (VerbatimChunk { a = cons(yy, a) })+ { yy = p.mkStringFromList(a, false)
		   yy.key = VERBATIM }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleVerbatimChunk]() {
				goto ko
			}
			do(20)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleVerbatimChunk]() {
					goto out
				}
				do(20)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			do(21)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 19 HorizontalRule <- (NonindentSpace ((&[_] ('_' Sp '_' Sp '_' (Sp '_')*)) | (&[\-] ('-' Sp '-' Sp '-' (Sp '-')*)) | (&[*] ('*' Sp '*' Sp '*' (Sp '*')*))) Sp Newline BlankLine+ { yy = p.mkElem(HRULE) }) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '_':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto ko
					}
					if !matchChar('_') {
						goto ko
					}
					if !p.rules[ruleSp]() {
						goto ko
					}
					if !matchChar('_') {
						goto ko
					}
				loop:
					{
						position1 := position
						if !p.rules[ruleSp]() {
							goto out
						}
						if !matchChar('_') {
							goto out
						}
						goto loop
					out:
						position = position1
					}
				case '-':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto ko
					}
					if !matchChar('-') {
						goto ko
					}
					if !p.rules[ruleSp]() {
						goto ko
					}
					if !matchChar('-') {
						goto ko
					}
				loop4:
					{
						position2 := position
						if !p.rules[ruleSp]() {
							goto out5
						}
						if !matchChar('-') {
							goto out5
						}
						goto loop4
					out5:
						position = position2
					}
				case '*':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto ko
					}
					if !matchChar('*') {
						goto ko
					}
					if !p.rules[ruleSp]() {
						goto ko
					}
					if !matchChar('*') {
						goto ko
					}
				loop6:
					{
						position3 := position
						if !p.rules[ruleSp]() {
							goto out7
						}
						if !matchChar('*') {
							goto out7
						}
						goto loop6
					out7:
						position = position3
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ko
			}
			if !p.rules[ruleBlankLine]() {
				goto ko
			}
		loop8:
			if !p.rules[ruleBlankLine]() {
				goto out9
			}
			goto loop8
		out9:
			do(22)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 20 Bullet <- (!HorizontalRule NonindentSpace ((&[\-] '-') | (&[*] '*') | (&[+] '+')) Spacechar+) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHorizontalRule]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleNonindentSpace]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
				case '*':
					position++ // matchChar
				case '+':
					position++ // matchChar
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto ko
			}
		loop:
			if !p.rules[ruleSpacechar]() {
				goto out
			}
			goto loop
		out:
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 21 BulletList <- (&Bullet (ListTight / ListLoose) { yy.key = BULLETLIST }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1 := position
				if !p.rules[ruleBullet]() {
					goto ko
				}
				position = position1
			}
			if !p.rules[ruleListTight]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleListLoose]() {
				goto ko
			}
		ok:
			do(23)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 22 ListTight <- (StartList (ListItemTight { a = cons(yy, a) })+ BlankLine* !((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) { yy = p.mkList(LIST, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItemTight]() {
				goto ko
			}
			do(24)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleListItemTight]() {
					goto out
				}
				do(24)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
		loop3:
			if !p.rules[ruleBlankLine]() {
				goto out4
			}
			goto loop3
		out4:
			{
				if position == len(p.Buffer) {
					goto ok
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto ok
					}
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto ok
					}
				default:
					if !p.rules[ruleEnumerator]() {
						goto ok
					}
				}
			}
			goto ko
		ok:
			do(25)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 23 ListLoose <- (StartList (ListItem BlankLine* {
		    li := b.children
		    li.contents.str += "\n\n"
		    a = cons(b, a)
		})+ { yy = p.mkList(LIST, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItem]() {
				goto ko
			}
			doarg(yySet, -2)
		loop3:
			if !p.rules[ruleBlankLine]() {
				goto out4
			}
			goto loop3
		out4:
			do(26)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto out
				}
				doarg(yySet, -2)
			loop5:
				if !p.rules[ruleBlankLine]() {
					goto out6
				}
				goto loop5
			out6:
				do(26)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			do(27)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 24 ListItem <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (ListContinuationBlock { a = cons(yy, a) })* {
		   raw := p.mkStringFromList(a, false)
		   raw.key = RAW
		   yy = p.mkElem(LISTITEM)
		   yy.children = raw
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto ko
					}
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto ko
					}
				default:
					if !p.rules[ruleEnumerator]() {
						goto ko
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto ko
			}
			do(28)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto out
				}
				do(29)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			do(30)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 25 ListItemTight <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (!BlankLine ListContinuationBlock { a = cons(yy, a) })* !ListContinuationBlock {
		   raw := p.mkStringFromList(a, false)
		   raw.key = RAW
		   yy = p.mkElem(LISTITEM)
		   yy.children = raw
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto ko
					}
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto ko
					}
				default:
					if !p.rules[ruleEnumerator]() {
						goto ko
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto ko
			}
			do(31)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto ok4
				}
				goto out
			ok4:
				if !p.rules[ruleListContinuationBlock]() {
					goto out
				}
				do(32)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleListContinuationBlock]() {
				goto ok5
			}
			goto ko
		ok5:
			do(33)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 26 ListBlock <- (StartList !BlankLine Line { a = cons(yy, a) } (ListBlockLine { a = cons(yy, a) })* { yy = p.mkStringFromList(a, false) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleLine]() {
				goto ko
			}
			do(34)
		loop:
			{
				position1 := position
				if !p.rules[ruleListBlockLine]() {
					goto out
				}
				do(35)
				goto loop
			out:
				position = position1
			}
			do(36)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 27 ListContinuationBlock <- (StartList (< BlankLine* > {   if len(yytext) == 0 {
		         a = cons(p.mkString("\001"), a) // block separator
		    } else {
		         a = cons(p.mkString(yytext), a)
		    }
		}) (Indent ListBlock { a = cons(yy, a) })+ {  yy = p.mkStringFromList(a, false) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			begin = position
		loop:
			if !p.rules[ruleBlankLine]() {
				goto out
			}
			goto loop
		out:
			end = position
			do(37)
			if !p.rules[ruleIndent]() {
				goto ko
			}
			if !p.rules[ruleListBlock]() {
				goto ko
			}
			do(38)
		loop3:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto out4
				}
				if !p.rules[ruleListBlock]() {
					goto out4
				}
				do(38)
				goto loop3
			out4:
				position, thunkPosition = position1, thunkPosition1
			}
			do(39)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 28 Enumerator <- (NonindentSpace [0-9]+ '.' Spacechar+) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto ko
			}
			if !matchClass(0) {
				goto ko
			}
		loop:
			if !matchClass(0) {
				goto out
			}
			goto loop
		out:
			if !matchChar('.') {
				goto ko
			}
			if !p.rules[ruleSpacechar]() {
				goto ko
			}
		loop3:
			if !p.rules[ruleSpacechar]() {
				goto out4
			}
			goto loop3
		out4:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 29 OrderedList <- (&Enumerator (ListTight / ListLoose) { yy.key = ORDEREDLIST }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1 := position
				if !p.rules[ruleEnumerator]() {
					goto ko
				}
				position = position1
			}
			if !p.rules[ruleListTight]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleListLoose]() {
				goto ko
			}
		ok:
			do(40)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 30 ListBlockLine <- (!BlankLine !((&[:~] DefMarker) | (&[\t *+\-0-9] (Indent? ((&[*+\-] Bullet) | (&[0-9] Enumerator))))) !HorizontalRule OptionallyIndentedLine) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlankLine]() {
				goto ok
			}
			goto ko
		ok:
			{
				position1 := position
				{
					if position == len(p.Buffer) {
						goto ok2
					}
					switch p.Buffer[position] {
					case ':', '~':
						if !p.rules[ruleDefMarker]() {
							goto ok2
						}
					default:
						if !p.rules[ruleIndent]() {
							goto ko4
						}
					ko4:
						{
							if position == len(p.Buffer) {
								goto ok2
							}
							switch p.Buffer[position] {
							case '*', '+', '-':
								if !p.rules[ruleBullet]() {
									goto ok2
								}
							default:
								if !p.rules[ruleEnumerator]() {
									goto ok2
								}
							}
						}
					}
				}
				goto ko
			ok2:
				position = position1
			}
			if !p.rules[ruleHorizontalRule]() {
				goto ok7
			}
			goto ko
		ok7:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto ko
			}
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 31 HtmlBlockOpenAddress <- ('<' Spnl ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto ko
					}
				case 'a':
					position++
					if !matchString("ddress") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 32 HtmlBlockCloseAddress <- ('<' Spnl '/' ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto ko
					}
				case 'a':
					position++
					if !matchString("ddress") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 33 HtmlBlockAddress <- (HtmlBlockOpenAddress (HtmlBlockAddress / (!HtmlBlockCloseAddress .))* HtmlBlockCloseAddress) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenAddress]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockAddress]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseAddress]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseAddress]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 34 HtmlBlockOpenBlockquote <- ('<' Spnl ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto ko
					}
				case 'b':
					position++
					if !matchString("lockquote") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 35 HtmlBlockCloseBlockquote <- ('<' Spnl '/' ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto ko
					}
				case 'b':
					position++
					if !matchString("lockquote") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 36 HtmlBlockBlockquote <- (HtmlBlockOpenBlockquote (HtmlBlockBlockquote / (!HtmlBlockCloseBlockquote .))* HtmlBlockCloseBlockquote) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenBlockquote]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockBlockquote]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseBlockquote]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseBlockquote]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 37 HtmlBlockOpenCenter <- ('<' Spnl ((&[C] 'CENTER') | (&[c] 'center')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto ko
					}
				case 'c':
					position++
					if !matchString("enter") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 38 HtmlBlockCloseCenter <- ('<' Spnl '/' ((&[C] 'CENTER') | (&[c] 'center')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto ko
					}
				case 'c':
					position++
					if !matchString("enter") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 39 HtmlBlockCenter <- (HtmlBlockOpenCenter (HtmlBlockCenter / (!HtmlBlockCloseCenter .))* HtmlBlockCloseCenter) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenCenter]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockCenter]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseCenter]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseCenter]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 40 HtmlBlockOpenDir <- ('<' Spnl ((&[D] 'DIR') | (&[d] 'dir')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto ko
					}
				case 'd':
					position++
					if !matchString("ir") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 41 HtmlBlockCloseDir <- ('<' Spnl '/' ((&[D] 'DIR') | (&[d] 'dir')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto ko
					}
				case 'd':
					position++
					if !matchString("ir") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 42 HtmlBlockDir <- (HtmlBlockOpenDir (HtmlBlockDir / (!HtmlBlockCloseDir .))* HtmlBlockCloseDir) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDir]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockDir]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseDir]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseDir]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 43 HtmlBlockOpenDiv <- ('<' Spnl ((&[D] 'DIV') | (&[d] 'div')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto ko
					}
				case 'd':
					position++
					if !matchString("iv") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 44 HtmlBlockCloseDiv <- ('<' Spnl '/' ((&[D] 'DIV') | (&[d] 'div')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto ko
					}
				case 'd':
					position++
					if !matchString("iv") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 45 HtmlBlockDiv <- (HtmlBlockOpenDiv (HtmlBlockDiv / (!HtmlBlockCloseDiv .))* HtmlBlockCloseDiv) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDiv]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockDiv]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseDiv]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseDiv]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 46 HtmlBlockOpenDl <- ('<' Spnl ((&[D] 'DL') | (&[d] 'dl')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto ko
					}
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 47 HtmlBlockCloseDl <- ('<' Spnl '/' ((&[D] 'DL') | (&[d] 'dl')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto ko
					}
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 48 HtmlBlockDl <- (HtmlBlockOpenDl (HtmlBlockDl / (!HtmlBlockCloseDl .))* HtmlBlockCloseDl) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDl]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockDl]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseDl]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseDl]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 49 HtmlBlockOpenFieldset <- ('<' Spnl ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto ko
					}
				case 'f':
					position++
					if !matchString("ieldset") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 50 HtmlBlockCloseFieldset <- ('<' Spnl '/' ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto ko
					}
				case 'f':
					position++
					if !matchString("ieldset") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 51 HtmlBlockFieldset <- (HtmlBlockOpenFieldset (HtmlBlockFieldset / (!HtmlBlockCloseFieldset .))* HtmlBlockCloseFieldset) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFieldset]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockFieldset]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseFieldset]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseFieldset]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 52 HtmlBlockOpenForm <- ('<' Spnl ((&[F] 'FORM') | (&[f] 'form')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto ko
					}
				case 'f':
					position++
					if !matchString("orm") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 53 HtmlBlockCloseForm <- ('<' Spnl '/' ((&[F] 'FORM') | (&[f] 'form')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto ko
					}
				case 'f':
					position++
					if !matchString("orm") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 54 HtmlBlockForm <- (HtmlBlockOpenForm (HtmlBlockForm / (!HtmlBlockCloseForm .))* HtmlBlockCloseForm) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenForm]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockForm]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseForm]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseForm]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 55 HtmlBlockOpenH1 <- ('<' Spnl ((&[H] 'H1') | (&[h] 'h1')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 56 HtmlBlockCloseH1 <- ('<' Spnl '/' ((&[H] 'H1') | (&[h] 'h1')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 57 HtmlBlockH1 <- (HtmlBlockOpenH1 (HtmlBlockH1 / (!HtmlBlockCloseH1 .))* HtmlBlockCloseH1) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH1]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockH1]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseH1]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseH1]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 58 HtmlBlockOpenH2 <- ('<' Spnl ((&[H] 'H2') | (&[h] 'h2')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 59 HtmlBlockCloseH2 <- ('<' Spnl '/' ((&[H] 'H2') | (&[h] 'h2')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 60 HtmlBlockH2 <- (HtmlBlockOpenH2 (HtmlBlockH2 / (!HtmlBlockCloseH2 .))* HtmlBlockCloseH2) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH2]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockH2]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseH2]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseH2]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 61 HtmlBlockOpenH3 <- ('<' Spnl ((&[H] 'H3') | (&[h] 'h3')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 62 HtmlBlockCloseH3 <- ('<' Spnl '/' ((&[H] 'H3') | (&[h] 'h3')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 63 HtmlBlockH3 <- (HtmlBlockOpenH3 (HtmlBlockH3 / (!HtmlBlockCloseH3 .))* HtmlBlockCloseH3) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH3]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockH3]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseH3]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseH3]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 64 HtmlBlockOpenH4 <- ('<' Spnl ((&[H] 'H4') | (&[h] 'h4')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 65 HtmlBlockCloseH4 <- ('<' Spnl '/' ((&[H] 'H4') | (&[h] 'h4')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 66 HtmlBlockH4 <- (HtmlBlockOpenH4 (HtmlBlockH4 / (!HtmlBlockCloseH4 .))* HtmlBlockCloseH4) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH4]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockH4]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseH4]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseH4]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 67 HtmlBlockOpenH5 <- ('<' Spnl ((&[H] 'H5') | (&[h] 'h5')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 68 HtmlBlockCloseH5 <- ('<' Spnl '/' ((&[H] 'H5') | (&[h] 'h5')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 69 HtmlBlockH5 <- (HtmlBlockOpenH5 (HtmlBlockH5 / (!HtmlBlockCloseH5 .))* HtmlBlockCloseH5) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH5]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockH5]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseH5]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseH5]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 70 HtmlBlockOpenH6 <- ('<' Spnl ((&[H] 'H6') | (&[h] 'h6')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 71 HtmlBlockCloseH6 <- ('<' Spnl '/' ((&[H] 'H6') | (&[h] 'h6')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto ko
					}
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 72 HtmlBlockH6 <- (HtmlBlockOpenH6 (HtmlBlockH6 / (!HtmlBlockCloseH6 .))* HtmlBlockCloseH6) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH6]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockH6]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseH6]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseH6]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 73 HtmlBlockOpenMenu <- ('<' Spnl ((&[M] 'MENU') | (&[m] 'menu')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto ko
					}
				case 'm':
					position++
					if !matchString("enu") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 74 HtmlBlockCloseMenu <- ('<' Spnl '/' ((&[M] 'MENU') | (&[m] 'menu')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto ko
					}
				case 'm':
					position++
					if !matchString("enu") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 75 HtmlBlockMenu <- (HtmlBlockOpenMenu (HtmlBlockMenu / (!HtmlBlockCloseMenu .))* HtmlBlockCloseMenu) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenMenu]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockMenu]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseMenu]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseMenu]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 76 HtmlBlockOpenNoframes <- ('<' Spnl ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto ko
					}
				case 'n':
					position++
					if !matchString("oframes") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 77 HtmlBlockCloseNoframes <- ('<' Spnl '/' ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto ko
					}
				case 'n':
					position++
					if !matchString("oframes") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 78 HtmlBlockNoframes <- (HtmlBlockOpenNoframes (HtmlBlockNoframes / (!HtmlBlockCloseNoframes .))* HtmlBlockCloseNoframes) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoframes]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockNoframes]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseNoframes]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseNoframes]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 79 HtmlBlockOpenNoscript <- ('<' Spnl ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto ko
					}
				case 'n':
					position++
					if !matchString("oscript") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 80 HtmlBlockCloseNoscript <- ('<' Spnl '/' ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto ko
					}
				case 'n':
					position++
					if !matchString("oscript") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 81 HtmlBlockNoscript <- (HtmlBlockOpenNoscript (HtmlBlockNoscript / (!HtmlBlockCloseNoscript .))* HtmlBlockCloseNoscript) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoscript]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockNoscript]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseNoscript]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseNoscript]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 82 HtmlBlockOpenOl <- ('<' Spnl ((&[O] 'OL') | (&[o] 'ol')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto ko
					}
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 83 HtmlBlockCloseOl <- ('<' Spnl '/' ((&[O] 'OL') | (&[o] 'ol')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto ko
					}
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 84 HtmlBlockOl <- (HtmlBlockOpenOl (HtmlBlockOl / (!HtmlBlockCloseOl .))* HtmlBlockCloseOl) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenOl]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockOl]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseOl]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseOl]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 85 HtmlBlockOpenP <- ('<' Spnl ((&[P] 'P') | (&[p] 'p')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
				case 'p':
					position++ // matchChar
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 86 HtmlBlockCloseP <- ('<' Spnl '/' ((&[P] 'P') | (&[p] 'p')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
				case 'p':
					position++ // matchChar
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 87 HtmlBlockP <- (HtmlBlockOpenP (HtmlBlockP / (!HtmlBlockCloseP .))* HtmlBlockCloseP) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenP]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockP]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseP]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseP]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 88 HtmlBlockOpenPre <- ('<' Spnl ((&[P] 'PRE') | (&[p] 'pre')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto ko
					}
				case 'p':
					position++
					if !matchString("re") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 89 HtmlBlockClosePre <- ('<' Spnl '/' ((&[P] 'PRE') | (&[p] 'pre')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto ko
					}
				case 'p':
					position++
					if !matchString("re") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 90 HtmlBlockPre <- (HtmlBlockOpenPre (HtmlBlockPre / (!HtmlBlockClosePre .))* HtmlBlockClosePre) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenPre]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockPre]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockClosePre]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockClosePre]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 91 HtmlBlockOpenTable <- ('<' Spnl ((&[T] 'TABLE') | (&[t] 'table')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto ko
					}
				case 't':
					position++
					if !matchString("able") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 92 HtmlBlockCloseTable <- ('<' Spnl '/' ((&[T] 'TABLE') | (&[t] 'table')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto ko
					}
				case 't':
					position++
					if !matchString("able") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 93 HtmlBlockTable <- (HtmlBlockOpenTable (HtmlBlockTable / (!HtmlBlockCloseTable .))* HtmlBlockCloseTable) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTable]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockTable]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseTable]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseTable]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 94 HtmlBlockOpenUl <- ('<' Spnl ((&[U] 'UL') | (&[u] 'ul')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto ko
					}
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 95 HtmlBlockCloseUl <- ('<' Spnl '/' ((&[U] 'UL') | (&[u] 'ul')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto ko
					}
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 96 HtmlBlockUl <- (HtmlBlockOpenUl (HtmlBlockUl / (!HtmlBlockCloseUl .))* HtmlBlockCloseUl) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenUl]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockUl]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseUl]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseUl]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 97 HtmlBlockOpenDd <- ('<' Spnl ((&[D] 'DD') | (&[d] 'dd')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto ko
					}
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 98 HtmlBlockCloseDd <- ('<' Spnl '/' ((&[D] 'DD') | (&[d] 'dd')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto ko
					}
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 99 HtmlBlockDd <- (HtmlBlockOpenDd (HtmlBlockDd / (!HtmlBlockCloseDd .))* HtmlBlockCloseDd) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDd]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockDd]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseDd]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseDd]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 100 HtmlBlockOpenDt <- ('<' Spnl ((&[D] 'DT') | (&[d] 'dt')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto ko
					}
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 101 HtmlBlockCloseDt <- ('<' Spnl '/' ((&[D] 'DT') | (&[d] 'dt')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto ko
					}
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 102 HtmlBlockDt <- (HtmlBlockOpenDt (HtmlBlockDt / (!HtmlBlockCloseDt .))* HtmlBlockCloseDt) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDt]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockDt]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseDt]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseDt]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 103 HtmlBlockOpenFrameset <- ('<' Spnl ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto ko
					}
				case 'f':
					position++
					if !matchString("rameset") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 104 HtmlBlockCloseFrameset <- ('<' Spnl '/' ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto ko
					}
				case 'f':
					position++
					if !matchString("rameset") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 105 HtmlBlockFrameset <- (HtmlBlockOpenFrameset (HtmlBlockFrameset / (!HtmlBlockCloseFrameset .))* HtmlBlockCloseFrameset) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFrameset]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockFrameset]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseFrameset]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseFrameset]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 106 HtmlBlockOpenLi <- ('<' Spnl ((&[L] 'LI') | (&[l] 'li')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto ko
					}
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 107 HtmlBlockCloseLi <- ('<' Spnl '/' ((&[L] 'LI') | (&[l] 'li')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto ko
					}
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 108 HtmlBlockLi <- (HtmlBlockOpenLi (HtmlBlockLi / (!HtmlBlockCloseLi .))* HtmlBlockCloseLi) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenLi]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockLi]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseLi]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseLi]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 109 HtmlBlockOpenTbody <- ('<' Spnl ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto ko
					}
				case 't':
					position++
					if !matchString("body") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 110 HtmlBlockCloseTbody <- ('<' Spnl '/' ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto ko
					}
				case 't':
					position++
					if !matchString("body") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 111 HtmlBlockTbody <- (HtmlBlockOpenTbody (HtmlBlockTbody / (!HtmlBlockCloseTbody .))* HtmlBlockCloseTbody) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTbody]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockTbody]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseTbody]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseTbody]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 112 HtmlBlockOpenTd <- ('<' Spnl ((&[T] 'TD') | (&[t] 'td')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto ko
					}
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 113 HtmlBlockCloseTd <- ('<' Spnl '/' ((&[T] 'TD') | (&[t] 'td')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto ko
					}
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 114 HtmlBlockTd <- (HtmlBlockOpenTd (HtmlBlockTd / (!HtmlBlockCloseTd .))* HtmlBlockCloseTd) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTd]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockTd]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseTd]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseTd]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 115 HtmlBlockOpenTfoot <- ('<' Spnl ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto ko
					}
				case 't':
					position++
					if !matchString("foot") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 116 HtmlBlockCloseTfoot <- ('<' Spnl '/' ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto ko
					}
				case 't':
					position++
					if !matchString("foot") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 117 HtmlBlockTfoot <- (HtmlBlockOpenTfoot (HtmlBlockTfoot / (!HtmlBlockCloseTfoot .))* HtmlBlockCloseTfoot) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTfoot]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockTfoot]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseTfoot]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseTfoot]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 118 HtmlBlockOpenTh <- ('<' Spnl ((&[T] 'TH') | (&[t] 'th')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto ko
					}
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 119 HtmlBlockCloseTh <- ('<' Spnl '/' ((&[T] 'TH') | (&[t] 'th')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto ko
					}
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 120 HtmlBlockTh <- (HtmlBlockOpenTh (HtmlBlockTh / (!HtmlBlockCloseTh .))* HtmlBlockCloseTh) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTh]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockTh]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseTh]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseTh]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 121 HtmlBlockOpenThead <- ('<' Spnl ((&[T] 'THEAD') | (&[t] 'thead')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto ko
					}
				case 't':
					position++
					if !matchString("head") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 122 HtmlBlockCloseThead <- ('<' Spnl '/' ((&[T] 'THEAD') | (&[t] 'thead')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto ko
					}
				case 't':
					position++
					if !matchString("head") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 123 HtmlBlockThead <- (HtmlBlockOpenThead (HtmlBlockThead / (!HtmlBlockCloseThead .))* HtmlBlockCloseThead) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenThead]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockThead]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseThead]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseThead]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 124 HtmlBlockOpenTr <- ('<' Spnl ((&[T] 'TR') | (&[t] 'tr')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto ko
					}
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 125 HtmlBlockCloseTr <- ('<' Spnl '/' ((&[T] 'TR') | (&[t] 'tr')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto ko
					}
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 126 HtmlBlockTr <- (HtmlBlockOpenTr (HtmlBlockTr / (!HtmlBlockCloseTr .))* HtmlBlockCloseTr) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTr]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockTr]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if !p.rules[ruleHtmlBlockCloseTr]() {
					goto ok5
				}
				goto out
			ok5:
				if !matchDot() {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseTr]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 127 HtmlBlockOpenScript <- ('<' Spnl ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto ko
					}
				case 's':
					position++
					if !matchString("cript") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 128 HtmlBlockCloseScript <- ('<' Spnl '/' ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto ko
					}
				case 's':
					position++
					if !matchString("cript") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 129 HtmlBlockScript <- (HtmlBlockOpenScript (!HtmlBlockCloseScript .)* HtmlBlockCloseScript) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenScript]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockCloseScript]() {
					goto ok
				}
				goto out
			ok:
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseScript]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 130 HtmlBlockOpenHead <- ('<' Spnl ((&[H] 'HEAD') | (&[h] 'head')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++
					if !matchString("EAD") {
						goto ko
					}
				case 'h':
					position++
					if !matchString("ead") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 131 HtmlBlockCloseHead <- ('<' Spnl '/' ((&[H] 'HEAD') | (&[h] 'head')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'H':
					position++
					if !matchString("EAD") {
						goto ko
					}
				case 'h':
					position++
					if !matchString("ead") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 132 HtmlBlockHead <- (HtmlBlockOpenHead (!HtmlBlockCloseHead .)* HtmlBlockCloseHead) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenHead]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleHtmlBlockCloseHead]() {
					goto ok
				}
				goto out
			ok:
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleHtmlBlockCloseHead]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 133 HtmlBlockInTags <- (HtmlBlockAddress / HtmlBlockBlockquote / HtmlBlockCenter / HtmlBlockDir / HtmlBlockDiv / HtmlBlockDl / HtmlBlockFieldset / HtmlBlockForm / HtmlBlockH1 / HtmlBlockH2 / HtmlBlockH3 / HtmlBlockH4 / HtmlBlockH5 / HtmlBlockH6 / HtmlBlockMenu / HtmlBlockNoframes / HtmlBlockNoscript / HtmlBlockOl / HtmlBlockP / HtmlBlockPre / HtmlBlockTable / HtmlBlockUl / HtmlBlockDd / HtmlBlockDt / HtmlBlockFrameset / HtmlBlockLi / HtmlBlockTbody / HtmlBlockTd / HtmlBlockTfoot / HtmlBlockTh / HtmlBlockThead / HtmlBlockTr / HtmlBlockScript / HtmlBlockHead) */
		func() (match bool) {
			if !p.rules[ruleHtmlBlockAddress]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleHtmlBlockBlockquote]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !p.rules[ruleHtmlBlockCenter]() {
				goto nextAlt4
			}
			goto ok
		nextAlt4:
			if !p.rules[ruleHtmlBlockDir]() {
				goto nextAlt5
			}
			goto ok
		nextAlt5:
			if !p.rules[ruleHtmlBlockDiv]() {
				goto nextAlt6
			}
			goto ok
		nextAlt6:
			if !p.rules[ruleHtmlBlockDl]() {
				goto nextAlt7
			}
			goto ok
		nextAlt7:
			if !p.rules[ruleHtmlBlockFieldset]() {
				goto nextAlt8
			}
			goto ok
		nextAlt8:
			if !p.rules[ruleHtmlBlockForm]() {
				goto nextAlt9
			}
			goto ok
		nextAlt9:
			if !p.rules[ruleHtmlBlockH1]() {
				goto nextAlt10
			}
			goto ok
		nextAlt10:
			if !p.rules[ruleHtmlBlockH2]() {
				goto nextAlt11
			}
			goto ok
		nextAlt11:
			if !p.rules[ruleHtmlBlockH3]() {
				goto nextAlt12
			}
			goto ok
		nextAlt12:
			if !p.rules[ruleHtmlBlockH4]() {
				goto nextAlt13
			}
			goto ok
		nextAlt13:
			if !p.rules[ruleHtmlBlockH5]() {
				goto nextAlt14
			}
			goto ok
		nextAlt14:
			if !p.rules[ruleHtmlBlockH6]() {
				goto nextAlt15
			}
			goto ok
		nextAlt15:
			if !p.rules[ruleHtmlBlockMenu]() {
				goto nextAlt16
			}
			goto ok
		nextAlt16:
			if !p.rules[ruleHtmlBlockNoframes]() {
				goto nextAlt17
			}
			goto ok
		nextAlt17:
			if !p.rules[ruleHtmlBlockNoscript]() {
				goto nextAlt18
			}
			goto ok
		nextAlt18:
			if !p.rules[ruleHtmlBlockOl]() {
				goto nextAlt19
			}
			goto ok
		nextAlt19:
			if !p.rules[ruleHtmlBlockP]() {
				goto nextAlt20
			}
			goto ok
		nextAlt20:
			if !p.rules[ruleHtmlBlockPre]() {
				goto nextAlt21
			}
			goto ok
		nextAlt21:
			if !p.rules[ruleHtmlBlockTable]() {
				goto nextAlt22
			}
			goto ok
		nextAlt22:
			if !p.rules[ruleHtmlBlockUl]() {
				goto nextAlt23
			}
			goto ok
		nextAlt23:
			if !p.rules[ruleHtmlBlockDd]() {
				goto nextAlt24
			}
			goto ok
		nextAlt24:
			if !p.rules[ruleHtmlBlockDt]() {
				goto nextAlt25
			}
			goto ok
		nextAlt25:
			if !p.rules[ruleHtmlBlockFrameset]() {
				goto nextAlt26
			}
			goto ok
		nextAlt26:
			if !p.rules[ruleHtmlBlockLi]() {
				goto nextAlt27
			}
			goto ok
		nextAlt27:
			if !p.rules[ruleHtmlBlockTbody]() {
				goto nextAlt28
			}
			goto ok
		nextAlt28:
			if !p.rules[ruleHtmlBlockTd]() {
				goto nextAlt29
			}
			goto ok
		nextAlt29:
			if !p.rules[ruleHtmlBlockTfoot]() {
				goto nextAlt30
			}
			goto ok
		nextAlt30:
			if !p.rules[ruleHtmlBlockTh]() {
				goto nextAlt31
			}
			goto ok
		nextAlt31:
			if !p.rules[ruleHtmlBlockThead]() {
				goto nextAlt32
			}
			goto ok
		nextAlt32:
			if !p.rules[ruleHtmlBlockTr]() {
				goto nextAlt33
			}
			goto ok
		nextAlt33:
			if !p.rules[ruleHtmlBlockScript]() {
				goto nextAlt34
			}
			goto ok
		nextAlt34:
			if !p.rules[ruleHtmlBlockHead]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 134 HtmlBlock <- (&'<' < (HtmlBlockInTags / HtmlComment / HtmlBlockSelfClosing) > BlankLine+ {   if p.extension.FilterHTML {
		        yy = p.mkList(LIST, nil)
		    } else {
		        yy = p.mkString(yytext)
		        yy.key = HTMLBLOCK
		    }
		}) */
		func() (match bool) {
			position0 := position
			if !peekChar('<') {
				goto ko
			}
			begin = position
			if !p.rules[ruleHtmlBlockInTags]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleHtmlComment]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !p.rules[ruleHtmlBlockSelfClosing]() {
				goto ko
			}
		ok:
			end = position
			if !p.rules[ruleBlankLine]() {
				goto ko
			}
		loop:
			if !p.rules[ruleBlankLine]() {
				goto out
			}
			goto loop
		out:
			do(41)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 135 HtmlBlockSelfClosing <- ('<' Spnl HtmlBlockType Spnl HtmlAttribute* '/' Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !p.rules[ruleHtmlBlockType]() {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('/') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 136 HtmlBlockType <- ('dir' / 'div' / 'dl' / 'fieldset' / 'form' / 'h1' / 'h2' / 'h3' / 'h4' / 'h5' / 'h6' / 'noframes' / 'p' / 'table' / 'dd' / 'tbody' / 'td' / 'tfoot' / 'th' / 'thead' / 'DIR' / 'DIV' / 'DL' / 'FIELDSET' / 'FORM' / 'H1' / 'H2' / 'H3' / 'H4' / 'H5' / 'H6' / 'NOFRAMES' / 'P' / 'TABLE' / 'DD' / 'TBODY' / 'TD' / 'TFOOT' / 'TH' / 'THEAD' / ((&[S] 'SCRIPT') | (&[T] 'TR') | (&[L] 'LI') | (&[F] 'FRAMESET') | (&[D] 'DT') | (&[U] 'UL') | (&[P] 'PRE') | (&[O] 'OL') | (&[N] 'NOSCRIPT') | (&[M] 'MENU') | (&[I] 'ISINDEX') | (&[H] 'HR') | (&[C] 'CENTER') | (&[B] 'BLOCKQUOTE') | (&[A] 'ADDRESS') | (&[s] 'script') | (&[t] 'tr') | (&[l] 'li') | (&[f] 'frameset') | (&[d] 'dt') | (&[u] 'ul') | (&[p] 'pre') | (&[o] 'ol') | (&[n] 'noscript') | (&[m] 'menu') | (&[i] 'isindex') | (&[h] 'hr') | (&[c] 'center') | (&[b] 'blockquote') | (&[a] 'address'))) */
		func() (match bool) {
			if !matchString("dir") {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !matchString("div") {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !matchString("dl") {
				goto nextAlt4
			}
			goto ok
		nextAlt4:
			if !matchString("fieldset") {
				goto nextAlt5
			}
			goto ok
		nextAlt5:
			if !matchString("form") {
				goto nextAlt6
			}
			goto ok
		nextAlt6:
			if !matchString("h1") {
				goto nextAlt7
			}
			goto ok
		nextAlt7:
			if !matchString("h2") {
				goto nextAlt8
			}
			goto ok
		nextAlt8:
			if !matchString("h3") {
				goto nextAlt9
			}
			goto ok
		nextAlt9:
			if !matchString("h4") {
				goto nextAlt10
			}
			goto ok
		nextAlt10:
			if !matchString("h5") {
				goto nextAlt11
			}
			goto ok
		nextAlt11:
			if !matchString("h6") {
				goto nextAlt12
			}
			goto ok
		nextAlt12:
			if !matchString("noframes") {
				goto nextAlt13
			}
			goto ok
		nextAlt13:
			if !matchChar('p') {
				goto nextAlt14
			}
			goto ok
		nextAlt14:
			if !matchString("table") {
				goto nextAlt15
			}
			goto ok
		nextAlt15:
			if !matchString("dd") {
				goto nextAlt16
			}
			goto ok
		nextAlt16:
			if !matchString("tbody") {
				goto nextAlt17
			}
			goto ok
		nextAlt17:
			if !matchString("td") {
				goto nextAlt18
			}
			goto ok
		nextAlt18:
			if !matchString("tfoot") {
				goto nextAlt19
			}
			goto ok
		nextAlt19:
			if !matchString("th") {
				goto nextAlt20
			}
			goto ok
		nextAlt20:
			if !matchString("thead") {
				goto nextAlt21
			}
			goto ok
		nextAlt21:
			if !matchString("DIR") {
				goto nextAlt22
			}
			goto ok
		nextAlt22:
			if !matchString("DIV") {
				goto nextAlt23
			}
			goto ok
		nextAlt23:
			if !matchString("DL") {
				goto nextAlt24
			}
			goto ok
		nextAlt24:
			if !matchString("FIELDSET") {
				goto nextAlt25
			}
			goto ok
		nextAlt25:
			if !matchString("FORM") {
				goto nextAlt26
			}
			goto ok
		nextAlt26:
			if !matchString("H1") {
				goto nextAlt27
			}
			goto ok
		nextAlt27:
			if !matchString("H2") {
				goto nextAlt28
			}
			goto ok
		nextAlt28:
			if !matchString("H3") {
				goto nextAlt29
			}
			goto ok
		nextAlt29:
			if !matchString("H4") {
				goto nextAlt30
			}
			goto ok
		nextAlt30:
			if !matchString("H5") {
				goto nextAlt31
			}
			goto ok
		nextAlt31:
			if !matchString("H6") {
				goto nextAlt32
			}
			goto ok
		nextAlt32:
			if !matchString("NOFRAMES") {
				goto nextAlt33
			}
			goto ok
		nextAlt33:
			if !matchChar('P') {
				goto nextAlt34
			}
			goto ok
		nextAlt34:
			if !matchString("TABLE") {
				goto nextAlt35
			}
			goto ok
		nextAlt35:
			if !matchString("DD") {
				goto nextAlt36
			}
			goto ok
		nextAlt36:
			if !matchString("TBODY") {
				goto nextAlt37
			}
			goto ok
		nextAlt37:
			if !matchString("TD") {
				goto nextAlt38
			}
			goto ok
		nextAlt38:
			if !matchString("TFOOT") {
				goto nextAlt39
			}
			goto ok
		nextAlt39:
			if !matchString("TH") {
				goto nextAlt40
			}
			goto ok
		nextAlt40:
			if !matchString("THEAD") {
				goto nextAlt41
			}
			goto ok
		nextAlt41:
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						return
					}
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						return
					}
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						return
					}
				case 'F':
					position++
					if !matchString("RAMESET") {
						return
					}
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						return
					}
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						return
					}
				case 'P':
					position++
					if !matchString("RE") {
						return
					}
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						return
					}
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						return
					}
				case 'M':
					position++
					if !matchString("ENU") {
						return
					}
				case 'I':
					position++
					if !matchString("SINDEX") {
						return
					}
				case 'H':
					position++ // matchString(`HR`)
					if !matchChar('R') {
						return
					}
				case 'C':
					position++
					if !matchString("ENTER") {
						return
					}
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						return
					}
				case 'A':
					position++
					if !matchString("DDRESS") {
						return
					}
				case 's':
					position++
					if !matchString("cript") {
						return
					}
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						return
					}
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						return
					}
				case 'f':
					position++
					if !matchString("rameset") {
						return
					}
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						return
					}
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						return
					}
				case 'p':
					position++
					if !matchString("re") {
						return
					}
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						return
					}
				case 'n':
					position++
					if !matchString("oscript") {
						return
					}
				case 'm':
					position++
					if !matchString("enu") {
						return
					}
				case 'i':
					position++
					if !matchString("sindex") {
						return
					}
				case 'h':
					position++ // matchString(`hr`)
					if !matchChar('r') {
						return
					}
				case 'c':
					position++
					if !matchString("enter") {
						return
					}
				case 'b':
					position++
					if !matchString("lockquote") {
						return
					}
				case 'a':
					position++
					if !matchString("ddress") {
						return
					}
				default:
					return
				}
			}
		ok:
			match = true
			return
		},
		/* 137 StyleOpen <- ('<' Spnl ((&[S] 'STYLE') | (&[s] 'style')) Spnl HtmlAttribute* '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto ko
					}
				case 's':
					position++
					if !matchString("tyle") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop:
			if !p.rules[ruleHtmlAttribute]() {
				goto out
			}
			goto loop
		out:
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 138 StyleClose <- ('<' Spnl '/' ((&[S] 'STYLE') | (&[s] 'style')) Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('/') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto ko
					}
				case 's':
					position++
					if !matchString("tyle") {
						goto ko
					}
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 139 InStyleTags <- (StyleOpen (!StyleClose .)* StyleClose) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleStyleOpen]() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleStyleClose]() {
					goto ok
				}
				goto out
			ok:
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			if !p.rules[ruleStyleClose]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 140 StyleBlock <- (< InStyleTags > BlankLine* {   if p.extension.FilterStyles {
		        yy = p.mkList(LIST, nil)
		    } else {
		        yy = p.mkString(yytext)
		        yy.key = HTMLBLOCK
		    }
		}) */
		func() (match bool) {
			position0 := position
			begin = position
			if !p.rules[ruleInStyleTags]() {
				goto ko
			}
			end = position
		loop:
			if !p.rules[ruleBlankLine]() {
				goto out
			}
			goto loop
		out:
			do(42)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 141 Inlines <- (StartList ((!Endline Inline { a = cons(yy, a) }) / (Endline &Inline { a = cons(c, a) }))+ Endline? { yy = p.mkList(LIST, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			{
				position1 := position
				if !p.rules[ruleEndline]() {
					goto ok5
				}
				goto nextAlt
			ok5:
				if !p.rules[ruleInline]() {
					goto nextAlt
				}
				do(43)
				goto ok
			nextAlt:
				position = position1
				if !p.rules[ruleEndline]() {
					goto ko
				}
				doarg(yySet, -2)
				{
					position2 := position
					if !p.rules[ruleInline]() {
						goto ko
					}
					position = position2
				}
				do(44)
			}
		ok:
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				{
					position4 := position
					if !p.rules[ruleEndline]() {
						goto ok9
					}
					goto nextAlt8
				ok9:
					if !p.rules[ruleInline]() {
						goto nextAlt8
					}
					do(43)
					goto ok7
				nextAlt8:
					position = position4
					if !p.rules[ruleEndline]() {
						goto out
					}
					doarg(yySet, -2)
					{
						position5 := position
						if !p.rules[ruleInline]() {
							goto out
						}
						position = position5
					}
					do(44)
				}
			ok7:
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleEndline]() {
				goto ko11
			}
		ko11:
			do(45)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 142 Inline <- (Str / Endline / UlOrStarLine / Space / Strong / Emph / Strike / Image / Link / NoteReference / InlineNote / Code / RawHtml / Entity / EscapedChar / Smart / Symbol) */
		func() (match bool) {
			if !p.rules[ruleStr]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleEndline]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !p.rules[ruleUlOrStarLine]() {
				goto nextAlt4
			}
			goto ok
		nextAlt4:
			if !p.rules[ruleSpace]() {
				goto nextAlt5
			}
			goto ok
		nextAlt5:
			if !p.rules[ruleStrong]() {
				goto nextAlt6
			}
			goto ok
		nextAlt6:
			if !p.rules[ruleEmph]() {
				goto nextAlt7
			}
			goto ok
		nextAlt7:
			if !p.rules[ruleStrike]() {
				goto nextAlt8
			}
			goto ok
		nextAlt8:
			if !p.rules[ruleImage]() {
				goto nextAlt9
			}
			goto ok
		nextAlt9:
			if !p.rules[ruleLink]() {
				goto nextAlt10
			}
			goto ok
		nextAlt10:
			if !p.rules[ruleNoteReference]() {
				goto nextAlt11
			}
			goto ok
		nextAlt11:
			if !p.rules[ruleInlineNote]() {
				goto nextAlt12
			}
			goto ok
		nextAlt12:
			if !p.rules[ruleCode]() {
				goto nextAlt13
			}
			goto ok
		nextAlt13:
			if !p.rules[ruleRawHtml]() {
				goto nextAlt14
			}
			goto ok
		nextAlt14:
			if !p.rules[ruleEntity]() {
				goto nextAlt15
			}
			goto ok
		nextAlt15:
			if !p.rules[ruleEscapedChar]() {
				goto nextAlt16
			}
			goto ok
		nextAlt16:
			if !p.rules[ruleSmart]() {
				goto nextAlt17
			}
			goto ok
		nextAlt17:
			if !p.rules[ruleSymbol]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 143 Space <- (Spacechar+ { yy = p.mkString(" ")
		   yy.key = SPACE }) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto ko
			}
		loop:
			if !p.rules[ruleSpacechar]() {
				goto out
			}
			goto loop
		out:
			do(46)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 144 Str <- (StartList < NormalChar+ > { a = cons(p.mkString(yytext), a) } (StrChunk { a = cons(yy, a) })* { if a.next == nil { yy = a; } else { yy = p.mkList(LIST, a) } }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleNormalChar]() {
				goto ko
			}
		loop:
			if !p.rules[ruleNormalChar]() {
				goto out
			}
			goto loop
		out:
			end = position
			do(47)
		loop3:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleStrChunk]() {
					goto out4
				}
				do(48)
				goto loop3
			out4:
				position, thunkPosition = position1, thunkPosition1
			}
			do(49)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 145 StrChunk <- ((< (NormalChar / ('_'+ &Alphanumeric))+ > { yy = p.mkString(yytext) }) / AposChunk) */
		func() (match bool) {
			position0 := position
			{
				position1 := position
				begin = position
				if !p.rules[ruleNormalChar]() {
					goto nextAlt6
				}
				goto ok5
			nextAlt6:
				if !matchChar('_') {
					goto nextAlt
				}
			loop7:
				if !matchChar('_') {
					goto out8
				}
				goto loop7
			out8:
				{
					position2 := position
					if !p.rules[ruleAlphanumeric]() {
						goto nextAlt
					}
					position = position2
				}
			ok5:
			loop:
				{
					position2 := position
					if !p.rules[ruleNormalChar]() {
						goto nextAlt11
					}
					goto ok10
				nextAlt11:
					if !matchChar('_') {
						goto out
					}
				loop12:
					if !matchChar('_') {
						goto out13
					}
					goto loop12
				out13:
					{
						position4 := position
						if !p.rules[ruleAlphanumeric]() {
							goto out
						}
						position = position4
					}
				ok10:
					goto loop
				out:
					position = position2
				}
				end = position
				do(50)
				goto ok
			nextAlt:
				position = position1
				if !p.rules[ruleAposChunk]() {
					goto ko
				}
			}
		ok:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 146 AposChunk <- (&{p.extension.Smart} '\'' &Alphanumeric { yy = p.mkElem(APOSTROPHE) }) */
		func() (match bool) {
			position0 := position
			if !(p.extension.Smart) {
				goto ko
			}
			if !matchChar('\'') {
				goto ko
			}
			{
				position1 := position
				if !p.rules[ruleAlphanumeric]() {
					goto ko
				}
				position = position1
			}
			do(51)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 147 EscapedChar <- ('\\' !Newline < [-\\`|*_{}[\]()#+.!><] > { yy = p.mkString(yytext) }) */
		func() (match bool) {
			position0 := position
			if !matchChar('\\') {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ok
			}
			goto ko
		ok:
			begin = position
			if !matchClass(1) {
				goto ko
			}
			end = position
			do(52)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 148 Entity <- ((HexEntity / DecEntity / CharEntity) { yy = p.mkString(yytext); yy.key = HTML }) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleHexEntity]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleDecEntity]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !p.rules[ruleCharEntity]() {
				goto ko
			}
		ok:
			do(53)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 149 Endline <- (LineBreak / TerminalEndline / NormalEndline) */
		func() (match bool) {
			if !p.rules[ruleLineBreak]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleTerminalEndline]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !p.rules[ruleNormalEndline]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 150 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line ((&[\-] '-'+) | (&[=] '='+)) Newline) { yy = p.mkString("\n")
		   yy.key = SPACE }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ko
			}
			if !p.rules[ruleBlankLine]() {
				goto ok
			}
			goto ko
		ok:
			if peekChar('>') {
				goto ko
			}
			if !p.rules[ruleAtxStart]() {
				goto ok2
			}
			goto ko
		ok2:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleLine]() {
					goto ok3
				}
				{
					if position == len(p.Buffer) {
						goto ok3
					}
					switch p.Buffer[position] {
					case '-':
						if !matchChar('-') {
							goto ok3
						}
					loop:
						if !matchChar('-') {
							goto out
						}
						goto loop
					out:
						break
					case '=':
						if !matchChar('=') {
							goto ok3
						}
					loop7:
						if !matchChar('=') {
							goto out8
						}
						goto loop7
					out8:
						break
					default:
						goto ok3
					}
				}
				if !p.rules[ruleNewline]() {
					goto ok3
				}
				goto ko
			ok3:
				position, thunkPosition = position1, thunkPosition1
			}
			do(54)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 151 TerminalEndline <- (Sp Newline !. { yy = nil }) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ko
			}
			if position < len(p.Buffer) {
				goto ko
			}
			do(55)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 152 LineBreak <- ('  ' NormalEndline { yy = p.mkElem(LINEBREAK) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("  ") {
				goto ko
			}
			if !p.rules[ruleNormalEndline]() {
				goto ko
			}
			do(56)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 153 Symbol <- (< SpecialChar > { yy = p.mkString(yytext) }) */
		func() (match bool) {
			position0 := position
			begin = position
			if !p.rules[ruleSpecialChar]() {
				goto ko
			}
			end = position
			do(57)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 154 UlOrStarLine <- ((UlLine / StarLine) { yy = p.mkString(yytext) }) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleStarLine]() {
				goto ko
			}
		ok:
			do(58)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 155 StarLine <- ((&[*] (< '****' '*'* >)) | (&[\t ] (< Spacechar '*'+ &Spacechar >))) */
		func() (match bool) {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '*':
					begin = position
					if !matchString("****") {
						goto ko
					}
				loop:
					if !matchChar('*') {
						goto out
					}
					goto loop
				out:
					end = position
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto ko
					}
					if !matchChar('*') {
						goto ko
					}
				loop4:
					if !matchChar('*') {
						goto out5
					}
					goto loop4
				out5:
					{
						position1 := position
						if !p.rules[ruleSpacechar]() {
							goto ko
						}
						position = position1
					}
					end = position
				default:
					goto ko
				}
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 156 UlLine <- ((&[_] (< '____' '_'* >)) | (&[\t ] (< Spacechar '_'+ &Spacechar >))) */
		func() (match bool) {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '_':
					begin = position
					if !matchString("____") {
						goto ko
					}
				loop:
					if !matchChar('_') {
						goto out
					}
					goto loop
				out:
					end = position
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto ko
					}
					if !matchChar('_') {
						goto ko
					}
				loop4:
					if !matchChar('_') {
						goto out5
					}
					goto loop4
				out5:
					{
						position1 := position
						if !p.rules[ruleSpacechar]() {
							goto ko
						}
						position = position1
					}
					end = position
				default:
					goto ko
				}
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 157 Emph <- ((&[_] EmphUl) | (&[*] EmphStar)) */
		func() (match bool) {
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleEmphUl]() {
						return
					}
				case '*':
					if !p.rules[ruleEmphStar]() {
						return
					}
				default:
					return
				}
			}
			match = true
			return
		},
		/* 158 Whitespace <- ((&[\n\r] Newline) | (&[\t ] Spacechar)) */
		func() (match bool) {
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						return
					}
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						return
					}
				default:
					return
				}
			}
			match = true
			return
		},
		/* 159 EmphStar <- ('*' !Whitespace StartList ((!'*' Inline { a = cons(b, a) }) / (StrongStar { a = cons(b, a) }))+ '*' { yy = p.mkList(EMPH, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('*') {
				goto ko
			}
			if !p.rules[ruleWhitespace]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			{
				position1, thunkPosition1 := position, thunkPosition
				if peekChar('*') {
					goto nextAlt
				}
				if !p.rules[ruleInline]() {
					goto nextAlt
				}
				doarg(yySet, -2)
				do(59)
				goto ok4
			nextAlt:
				position, thunkPosition = position1, thunkPosition1
				if !p.rules[ruleStrongStar]() {
					goto ko
				}
				doarg(yySet, -2)
				do(60)
			}
		ok4:
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				{
					position3, thunkPosition3 := position, thunkPosition
					if peekChar('*') {
						goto nextAlt7
					}
					if !p.rules[ruleInline]() {
						goto nextAlt7
					}
					doarg(yySet, -2)
					do(59)
					goto ok6
				nextAlt7:
					position, thunkPosition = position3, thunkPosition3
					if !p.rules[ruleStrongStar]() {
						goto out
					}
					doarg(yySet, -2)
					do(60)
				}
			ok6:
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !matchChar('*') {
				goto ko
			}
			do(61)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 160 EmphUl <- ('_' !Whitespace StartList ((!'_' Inline { a = cons(b, a) }) / (StrongUl { a = cons(b, a) }))+ '_' { yy = p.mkList(EMPH, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('_') {
				goto ko
			}
			if !p.rules[ruleWhitespace]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			{
				position1, thunkPosition1 := position, thunkPosition
				if peekChar('_') {
					goto nextAlt
				}
				if !p.rules[ruleInline]() {
					goto nextAlt
				}
				doarg(yySet, -2)
				do(62)
				goto ok4
			nextAlt:
				position, thunkPosition = position1, thunkPosition1
				if !p.rules[ruleStrongUl]() {
					goto ko
				}
				doarg(yySet, -2)
				do(63)
			}
		ok4:
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				{
					position3, thunkPosition3 := position, thunkPosition
					if peekChar('_') {
						goto nextAlt7
					}
					if !p.rules[ruleInline]() {
						goto nextAlt7
					}
					doarg(yySet, -2)
					do(62)
					goto ok6
				nextAlt7:
					position, thunkPosition = position3, thunkPosition3
					if !p.rules[ruleStrongUl]() {
						goto out
					}
					doarg(yySet, -2)
					do(63)
				}
			ok6:
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !matchChar('_') {
				goto ko
			}
			do(64)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 161 Strong <- ((&[_] StrongUl) | (&[*] StrongStar)) */
		func() (match bool) {
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleStrongUl]() {
						return
					}
				case '*':
					if !p.rules[ruleStrongStar]() {
						return
					}
				default:
					return
				}
			}
			match = true
			return
		},
		/* 162 StrongStar <- ('**' !Whitespace StartList (!'**' Inline { a = cons(b, a) })+ '**' { yy = p.mkList(STRONG, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchString("**") {
				goto ko
			}
			if !p.rules[ruleWhitespace]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchString("**") {
				goto ok4
			}
			goto ko
		ok4:
			if !p.rules[ruleInline]() {
				goto ko
			}
			doarg(yySet, -2)
			do(65)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchString("**") {
					goto ok5
				}
				goto out
			ok5:
				if !p.rules[ruleInline]() {
					goto out
				}
				doarg(yySet, -2)
				do(65)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !matchString("**") {
				goto ko
			}
			do(66)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 163 StrongUl <- ('__' !Whitespace StartList (!'__' Inline { a = cons(b, a) })+ '__' { yy = p.mkList(STRONG, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchString("__") {
				goto ko
			}
			if !p.rules[ruleWhitespace]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchString("__") {
				goto ok4
			}
			goto ko
		ok4:
			if !p.rules[ruleInline]() {
				goto ko
			}
			doarg(yySet, -2)
			do(67)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchString("__") {
					goto ok5
				}
				goto out
			ok5:
				if !p.rules[ruleInline]() {
					goto out
				}
				doarg(yySet, -2)
				do(67)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !matchString("__") {
				goto ko
			}
			do(68)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 164 TwoTildeOpen <- (&{p.extension.Strike} !TildeLine '~~' !Spacechar !Newline) */
		func() (match bool) {
			position0 := position
			if !(p.extension.Strike) {
				goto ko
			}
			if !p.rules[ruleTildeLine]() {
				goto ok
			}
			goto ko
		ok:
			if !matchString("~~") {
				goto ko
			}
			if !p.rules[ruleSpacechar]() {
				goto ok2
			}
			goto ko
		ok2:
			if !p.rules[ruleNewline]() {
				goto ok3
			}
			goto ko
		ok3:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 165 TwoTildeClose <- (&{p.extension.Strike} !Spacechar !Newline Inline '~~' { yy = a; }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Strike) {
				goto ko
			}
			if !p.rules[ruleSpacechar]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleNewline]() {
				goto ok2
			}
			goto ko
		ok2:
			if !p.rules[ruleInline]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchString("~~") {
				goto ko
			}
			do(69)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 166 Strike <- (&{p.extension.Strike} '~~' !Whitespace StartList (!'~~' Inline { a = cons(b, a) })+ '~~' { yy = p.mkList(STRIKE, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !(p.extension.Strike) {
				goto ko
			}
			if !matchString("~~") {
				goto ko
			}
			if !p.rules[ruleWhitespace]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchString("~~") {
				goto ok4
			}
			goto ko
		ok4:
			if !p.rules[ruleInline]() {
				goto ko
			}
			doarg(yySet, -2)
			do(70)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchString("~~") {
					goto ok5
				}
				goto out
			ok5:
				if !p.rules[ruleInline]() {
					goto out
				}
				doarg(yySet, -2)
				do(70)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !matchString("~~") {
				goto ko
			}
			do(71)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 167 Image <- ('!' (ExplicitLink / ReferenceLink) {	if yy.key == LINK {
				yy.key = IMAGE
			} else {
				result := yy
				yy.children = cons(p.mkString("!"), result.children)
			}
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('!') {
				goto ko
			}
			if !p.rules[ruleExplicitLink]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleReferenceLink]() {
				goto ko
			}
		ok:
			do(72)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 168 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() (match bool) {
			if !p.rules[ruleExplicitLink]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleReferenceLink]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !p.rules[ruleAutoLink]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 169 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() (match bool) {
			if !p.rules[ruleReferenceLinkDouble]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleReferenceLinkSingle]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 170 ReferenceLinkDouble <- (Label < Spnl > !'[]' Label {
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
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleLabel]() {
				goto ko
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			end = position
			if !matchString("[]") {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleLabel]() {
				goto ko
			}
			doarg(yySet, -2)
			do(73)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 171 ReferenceLinkSingle <- (Label < (Spnl '[]')? > {
		    if match, found := p.findReference(a.children); found {
		        yy = p.mkLink(a.children, match.url, match.title)
		        a = nil
		    } else {
		        result := p.mkElem(LIST)
		        result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), p.mkString(yytext))));
		        yy = result
		    }
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleLabel]() {
				goto ko
			}
			doarg(yySet, -1)
			begin = position
			{
				position1 := position
				if !p.rules[ruleSpnl]() {
					goto ko1
				}
				if !matchString("[]") {
					goto ko1
				}
				goto ok
			ko1:
				position = position1
			}
		ok:
			end = position
			do(74)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 172 ExplicitLink <- (Label '(' Sp Source Spnl Title Sp ')' { yy = p.mkLink(l.children, s.contents.str, t.contents.str)
		   s = nil
		   t = nil
		   l = nil }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleLabel]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchChar('(') {
				goto ko
			}
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleSource]() {
				goto ko
			}
			doarg(yySet, -2)
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !p.rules[ruleTitle]() {
				goto ko
			}
			doarg(yySet, -3)
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !matchChar(')') {
				goto ko
			}
			do(75)
			doarg(yyPop, 3)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 173 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = p.mkString(yytext) }) */
		func() (match bool) {
			position0 := position
			{
				position1 := position
				if !matchChar('<') {
					goto nextAlt
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto nextAlt
				}
				end = position
				if !matchChar('>') {
					goto nextAlt
				}
				goto ok
			nextAlt:
				position = position1
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto ko
				}
				end = position
			}
		ok:
			do(76)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 174 SourceContents <- ((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* */
		func() (match bool) {
		loop:
			{
				position1 := position
				if position == len(p.Buffer) {
					goto nextAlt
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto nextAlt
				default:
					if !p.rules[ruleNonspacechar]() {
						goto nextAlt
					}
				}
			loop5:
				if position == len(p.Buffer) {
					goto out6
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto out6
				default:
					if !p.rules[ruleNonspacechar]() {
						goto out6
					}
				}
				goto loop5
			out6:
				goto ok
			nextAlt:
				if !matchChar('(') {
					goto out
				}
				if !p.rules[ruleSourceContents]() {
					goto out
				}
				if !matchChar(')') {
					goto out
				}
			ok:
				goto loop
			out:
				position = position1
			}
			match = true
			return
		},
		/* 175 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = p.mkString(yytext) }) */
		func() (match bool) {
			if !p.rules[ruleTitleSingle]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleTitleDouble]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			begin = position
			end = position
		ok:
			do(77)
			match = true
			return
		},
		/* 176 TitleSingle <- ('\'' < (!('\'' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '\'') */
		func() (match bool) {
			position0 := position
			if !matchChar('\'') {
				goto ko
			}
			begin = position
		loop:
			{
				position1 := position
				{
					position2 := position
					if !matchChar('\'') {
						goto ok
					}
					if !p.rules[ruleSp]() {
						goto ok
					}
					{
						if position == len(p.Buffer) {
							goto ok
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto ok
							}
						default:
							goto ok
						}
					}
					goto out
				ok:
					position = position2
				}
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			end = position
			if !matchChar('\'') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 177 TitleDouble <- ('"' < (!('"' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '"') */
		func() (match bool) {
			position0 := position
			if !matchChar('"') {
				goto ko
			}
			begin = position
		loop:
			{
				position1 := position
				{
					position2 := position
					if !matchChar('"') {
						goto ok
					}
					if !p.rules[ruleSp]() {
						goto ok
					}
					{
						if position == len(p.Buffer) {
							goto ok
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto ok
							}
						default:
							goto ok
						}
					}
					goto out
				ok:
					position = position2
				}
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			end = position
			if !matchChar('"') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 178 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() (match bool) {
			if !p.rules[ruleAutoLinkUrl]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleAutoLinkEmail]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 179 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = p.mkLink(p.mkString(yytext), yytext, "") }) */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			begin = position
			if !matchClass(2) {
				goto ko
			}
		loop:
			if !matchClass(2) {
				goto out
			}
			goto loop
		out:
			if !matchString("://") {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ok
			}
			goto ko
		ok:
			if peekChar('>') {
				goto ko
			}
			if !matchDot() {
				goto ko
			}
		loop3:
			{
				position1 := position
				if !p.rules[ruleNewline]() {
					goto ok6
				}
				goto out4
			ok6:
				if peekChar('>') {
					goto out4
				}
				if !matchDot() {
					goto out4
				}
				goto loop3
			out4:
				position = position1
			}
			end = position
			if !matchChar('>') {
				goto ko
			}
			do(78)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 180 AutoLinkEmail <- ('<' 'mailto:'? < [-A-Za-z0-9+_./!%~$]+ '@' (!Newline !'>' .)+ > '>' {
		    yy = p.mkLink(p.mkString(yytext), "mailto:"+yytext, "")
		}) */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !matchString("mailto:") {
				goto ko1
			}
		ko1:
			begin = position
			if !matchClass(3) {
				goto ko
			}
		loop:
			if !matchClass(3) {
				goto out
			}
			goto loop
		out:
			if !matchChar('@') {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ok7
			}
			goto ko
		ok7:
			if peekChar('>') {
				goto ko
			}
			if !matchDot() {
				goto ko
			}
		loop5:
			{
				position1 := position
				if !p.rules[ruleNewline]() {
					goto ok8
				}
				goto out6
			ok8:
				if peekChar('>') {
					goto out6
				}
				if !matchDot() {
					goto out6
				}
				goto loop5
			out6:
				position = position1
			}
			end = position
			if !matchChar('>') {
				goto ko
			}
			do(79)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 181 Reference <- (NonindentSpace !'[]' Label ':' Spnl RefSrc RefTitle BlankLine+ { yy = p.mkLink(l.children, s.contents.str, t.contents.str)
		   s = nil
		   t = nil
		   l = nil
		   yy.key = REFERENCE }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleNonindentSpace]() {
				goto ko
			}
			if !matchString("[]") {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleLabel]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !p.rules[ruleRefSrc]() {
				goto ko
			}
			doarg(yySet, -2)
			if !p.rules[ruleRefTitle]() {
				goto ko
			}
			doarg(yySet, -3)
			if !p.rules[ruleBlankLine]() {
				goto ko
			}
		loop:
			if !p.rules[ruleBlankLine]() {
				goto out
			}
			goto loop
		out:
			do(80)
			doarg(yyPop, 3)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 182 Label <- ('[' ((!'^' &{p.extension.Notes}) / (&. &{!p.extension.Notes})) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = p.mkList(LIST, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto ko
			}
			if peekChar('^') {
				goto nextAlt
			}
			if !(p.extension.Notes) {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !(position < len(p.Buffer)) {
				goto ko
			}
			if !(!p.extension.Notes) {
				goto ko
			}
		ok:
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
		loop:
			{
				position1 := position
				if peekChar(']') {
					goto out
				}
				if !p.rules[ruleInline]() {
					goto out
				}
				do(81)
				goto loop
			out:
				position = position1
			}
			if !matchChar(']') {
				goto ko
			}
			do(82)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 183 RefSrc <- (< Nonspacechar+ > { yy = p.mkString(yytext)
		   yy.key = HTML }) */
		func() (match bool) {
			position0 := position
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto ko
			}
		loop:
			if !p.rules[ruleNonspacechar]() {
				goto out
			}
			goto loop
		out:
			end = position
			do(83)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 184 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = p.mkString(yytext) }) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleRefTitleSingle]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleRefTitleDouble]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !p.rules[ruleRefTitleParens]() {
				goto nextAlt4
			}
			goto ok
		nextAlt4:
			if !p.rules[ruleEmptyTitle]() {
				goto ko
			}
		ok:
			do(84)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 185 EmptyTitle <- (< '' >) */
		func() (match bool) {
			begin = position
			end = position
			match = true
			return
		},
		/* 186 RefTitleSingle <- (Spnl '\'' < (!((&[\'] ('\'' Sp Newline)) | (&[\n\r] Newline)) .)* > '\'') */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('\'') {
				goto ko
			}
			begin = position
		loop:
			{
				position1 := position
				{
					position2 := position
					{
						if position == len(p.Buffer) {
							goto ok
						}
						switch p.Buffer[position] {
						case '\'':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto ok
							}
							if !p.rules[ruleNewline]() {
								goto ok
							}
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto ok
							}
						default:
							goto ok
						}
					}
					goto out
				ok:
					position = position2
				}
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			end = position
			if !matchChar('\'') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 187 RefTitleDouble <- (Spnl '"' < (!((&[\"] ('"' Sp Newline)) | (&[\n\r] Newline)) .)* > '"') */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('"') {
				goto ko
			}
			begin = position
		loop:
			{
				position1 := position
				{
					position2 := position
					{
						if position == len(p.Buffer) {
							goto ok
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto ok
							}
							if !p.rules[ruleNewline]() {
								goto ok
							}
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto ok
							}
						default:
							goto ok
						}
					}
					goto out
				ok:
					position = position2
				}
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			end = position
			if !matchChar('"') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 188 RefTitleParens <- (Spnl '(' < (!((&[)] (')' Sp Newline)) | (&[\n\r] Newline)) .)* > ')') */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('(') {
				goto ko
			}
			begin = position
		loop:
			{
				position1 := position
				{
					position2 := position
					{
						if position == len(p.Buffer) {
							goto ok
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto ok
							}
							if !p.rules[ruleNewline]() {
								goto ok
							}
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto ok
							}
						default:
							goto ok
						}
					}
					goto out
				ok:
					position = position2
				}
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			end = position
			if !matchChar(')') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 189 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				{
					position2, thunkPosition2 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto nextAlt
					}
					doarg(yySet, -2)
					do(85)
					goto ok
				nextAlt:
					position, thunkPosition = position2, thunkPosition2
					if !p.rules[ruleSkipBlock]() {
						goto out
					}
				}
			ok:
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			do(86)
			if !(p.commit(thunkPosition0)) {
				goto ko
			}
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 190 Ticks1 <- ('`' !'`') */
		func() (match bool) {
			position0 := position
			if !matchChar('`') {
				goto ko
			}
			if peekChar('`') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 191 Ticks2 <- ('``' !'`') */
		func() (match bool) {
			position0 := position
			if !matchString("``") {
				goto ko
			}
			if peekChar('`') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 192 Ticks3 <- ('```' !'`') */
		func() (match bool) {
			position0 := position
			if !matchString("```") {
				goto ko
			}
			if peekChar('`') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 193 Ticks4 <- ('````' !'`') */
		func() (match bool) {
			position0 := position
			if !matchString("````") {
				goto ko
			}
			if peekChar('`') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 194 Ticks5 <- ('`````' !'`') */
		func() (match bool) {
			position0 := position
			if !matchString("`````") {
				goto ko
			}
			if peekChar('`') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 195 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks1 '`'+)) | (&[\t\n\r ] (!(Sp Ticks1) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks2 '`'+)) | (&[\t\n\r ] (!(Sp Ticks2) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks3 '`'+)) | (&[\t\n\r ] (!(Sp Ticks3) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks4 '`'+)) | (&[\t\n\r ] (!(Sp Ticks4) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks5 '`'+)) | (&[\t\n\r ] (!(Sp Ticks5) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks5)) { yy = p.mkString(yytext); yy.key = CODE }) */
		func() (match bool) {
			position0 := position
			{
				position1 := position
				if !p.rules[ruleTicks1]() {
					goto nextAlt
				}
				if !p.rules[ruleSp]() {
					goto nextAlt
				}
				begin = position
				if peekChar('`') {
					goto nextAlt6
				}
				if !p.rules[ruleNonspacechar]() {
					goto nextAlt6
				}
			loop7:
				if peekChar('`') {
					goto out8
				}
				if !p.rules[ruleNonspacechar]() {
					goto out8
				}
				goto loop7
			out8:
				goto ok5
			nextAlt6:
				{
					if position == len(p.Buffer) {
						goto nextAlt
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks1]() {
							goto ok10
						}
						goto nextAlt
					ok10:
						if !matchChar('`') {
							goto nextAlt
						}
					loop11:
						if !matchChar('`') {
							goto out12
						}
						goto loop11
					out12:
						break
					default:
						{
							position2 := position
							if !p.rules[ruleSp]() {
								goto ok13
							}
							if !p.rules[ruleTicks1]() {
								goto ok13
							}
							goto nextAlt
						ok13:
							position = position2
						}
						{
							if position == len(p.Buffer) {
								goto nextAlt
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto nextAlt
								}
								if !p.rules[ruleBlankLine]() {
									goto ok15
								}
								goto nextAlt
							ok15:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto nextAlt
								}
							default:
								goto nextAlt
							}
						}
					}
				}
			ok5:
			loop:
				{
					position2 := position
					if peekChar('`') {
						goto nextAlt17
					}
					if !p.rules[ruleNonspacechar]() {
						goto nextAlt17
					}
				loop18:
					if peekChar('`') {
						goto out19
					}
					if !p.rules[ruleNonspacechar]() {
						goto out19
					}
					goto loop18
				out19:
					goto ok16
				nextAlt17:
					{
						if position == len(p.Buffer) {
							goto out
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks1]() {
								goto ok21
							}
							goto out
						ok21:
							if !matchChar('`') {
								goto out
							}
						loop22:
							if !matchChar('`') {
								goto out23
							}
							goto loop22
						out23:
							break
						default:
							{
								position4 := position
								if !p.rules[ruleSp]() {
									goto ok24
								}
								if !p.rules[ruleTicks1]() {
									goto ok24
								}
								goto out
							ok24:
								position = position4
							}
							{
								if position == len(p.Buffer) {
									goto out
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto out
									}
									if !p.rules[ruleBlankLine]() {
										goto ok26
									}
									goto out
								ok26:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto out
									}
								default:
									goto out
								}
							}
						}
					}
				ok16:
					goto loop
				out:
					position = position2
				}
				end = position
				if !p.rules[ruleSp]() {
					goto nextAlt
				}
				if !p.rules[ruleTicks1]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				position = position1
				if !p.rules[ruleTicks2]() {
					goto nextAlt27
				}
				if !p.rules[ruleSp]() {
					goto nextAlt27
				}
				begin = position
				if peekChar('`') {
					goto nextAlt31
				}
				if !p.rules[ruleNonspacechar]() {
					goto nextAlt31
				}
			loop32:
				if peekChar('`') {
					goto out33
				}
				if !p.rules[ruleNonspacechar]() {
					goto out33
				}
				goto loop32
			out33:
				goto ok30
			nextAlt31:
				{
					if position == len(p.Buffer) {
						goto nextAlt27
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks2]() {
							goto ok35
						}
						goto nextAlt27
					ok35:
						if !matchChar('`') {
							goto nextAlt27
						}
					loop36:
						if !matchChar('`') {
							goto out37
						}
						goto loop36
					out37:
						break
					default:
						{
							position5 := position
							if !p.rules[ruleSp]() {
								goto ok38
							}
							if !p.rules[ruleTicks2]() {
								goto ok38
							}
							goto nextAlt27
						ok38:
							position = position5
						}
						{
							if position == len(p.Buffer) {
								goto nextAlt27
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto nextAlt27
								}
								if !p.rules[ruleBlankLine]() {
									goto ok40
								}
								goto nextAlt27
							ok40:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto nextAlt27
								}
							default:
								goto nextAlt27
							}
						}
					}
				}
			ok30:
			loop28:
				{
					position5 := position
					if peekChar('`') {
						goto nextAlt42
					}
					if !p.rules[ruleNonspacechar]() {
						goto nextAlt42
					}
				loop43:
					if peekChar('`') {
						goto out44
					}
					if !p.rules[ruleNonspacechar]() {
						goto out44
					}
					goto loop43
				out44:
					goto ok41
				nextAlt42:
					{
						if position == len(p.Buffer) {
							goto out29
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks2]() {
								goto ok46
							}
							goto out29
						ok46:
							if !matchChar('`') {
								goto out29
							}
						loop47:
							if !matchChar('`') {
								goto out48
							}
							goto loop47
						out48:
							break
						default:
							{
								position7 := position
								if !p.rules[ruleSp]() {
									goto ok49
								}
								if !p.rules[ruleTicks2]() {
									goto ok49
								}
								goto out29
							ok49:
								position = position7
							}
							{
								if position == len(p.Buffer) {
									goto out29
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto out29
									}
									if !p.rules[ruleBlankLine]() {
										goto ok51
									}
									goto out29
								ok51:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto out29
									}
								default:
									goto out29
								}
							}
						}
					}
				ok41:
					goto loop28
				out29:
					position = position5
				}
				end = position
				if !p.rules[ruleSp]() {
					goto nextAlt27
				}
				if !p.rules[ruleTicks2]() {
					goto nextAlt27
				}
				goto ok
			nextAlt27:
				position = position1
				if !p.rules[ruleTicks3]() {
					goto nextAlt52
				}
				if !p.rules[ruleSp]() {
					goto nextAlt52
				}
				begin = position
				if peekChar('`') {
					goto nextAlt56
				}
				if !p.rules[ruleNonspacechar]() {
					goto nextAlt56
				}
			loop57:
				if peekChar('`') {
					goto out58
				}
				if !p.rules[ruleNonspacechar]() {
					goto out58
				}
				goto loop57
			out58:
				goto ok55
			nextAlt56:
				{
					if position == len(p.Buffer) {
						goto nextAlt52
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks3]() {
							goto ok60
						}
						goto nextAlt52
					ok60:
						if !matchChar('`') {
							goto nextAlt52
						}
					loop61:
						if !matchChar('`') {
							goto out62
						}
						goto loop61
					out62:
						break
					default:
						{
							position8 := position
							if !p.rules[ruleSp]() {
								goto ok63
							}
							if !p.rules[ruleTicks3]() {
								goto ok63
							}
							goto nextAlt52
						ok63:
							position = position8
						}
						{
							if position == len(p.Buffer) {
								goto nextAlt52
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto nextAlt52
								}
								if !p.rules[ruleBlankLine]() {
									goto ok65
								}
								goto nextAlt52
							ok65:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto nextAlt52
								}
							default:
								goto nextAlt52
							}
						}
					}
				}
			ok55:
			loop53:
				{
					position8 := position
					if peekChar('`') {
						goto nextAlt67
					}
					if !p.rules[ruleNonspacechar]() {
						goto nextAlt67
					}
				loop68:
					if peekChar('`') {
						goto out69
					}
					if !p.rules[ruleNonspacechar]() {
						goto out69
					}
					goto loop68
				out69:
					goto ok66
				nextAlt67:
					{
						if position == len(p.Buffer) {
							goto out54
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks3]() {
								goto ok71
							}
							goto out54
						ok71:
							if !matchChar('`') {
								goto out54
							}
						loop72:
							if !matchChar('`') {
								goto out73
							}
							goto loop72
						out73:
							break
						default:
							{
								position10 := position
								if !p.rules[ruleSp]() {
									goto ok74
								}
								if !p.rules[ruleTicks3]() {
									goto ok74
								}
								goto out54
							ok74:
								position = position10
							}
							{
								if position == len(p.Buffer) {
									goto out54
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto out54
									}
									if !p.rules[ruleBlankLine]() {
										goto ok76
									}
									goto out54
								ok76:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto out54
									}
								default:
									goto out54
								}
							}
						}
					}
				ok66:
					goto loop53
				out54:
					position = position8
				}
				end = position
				if !p.rules[ruleSp]() {
					goto nextAlt52
				}
				if !p.rules[ruleTicks3]() {
					goto nextAlt52
				}
				goto ok
			nextAlt52:
				position = position1
				if !p.rules[ruleTicks4]() {
					goto nextAlt77
				}
				if !p.rules[ruleSp]() {
					goto nextAlt77
				}
				begin = position
				if peekChar('`') {
					goto nextAlt81
				}
				if !p.rules[ruleNonspacechar]() {
					goto nextAlt81
				}
			loop82:
				if peekChar('`') {
					goto out83
				}
				if !p.rules[ruleNonspacechar]() {
					goto out83
				}
				goto loop82
			out83:
				goto ok80
			nextAlt81:
				{
					if position == len(p.Buffer) {
						goto nextAlt77
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks4]() {
							goto ok85
						}
						goto nextAlt77
					ok85:
						if !matchChar('`') {
							goto nextAlt77
						}
					loop86:
						if !matchChar('`') {
							goto out87
						}
						goto loop86
					out87:
						break
					default:
						{
							position11 := position
							if !p.rules[ruleSp]() {
								goto ok88
							}
							if !p.rules[ruleTicks4]() {
								goto ok88
							}
							goto nextAlt77
						ok88:
							position = position11
						}
						{
							if position == len(p.Buffer) {
								goto nextAlt77
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto nextAlt77
								}
								if !p.rules[ruleBlankLine]() {
									goto ok90
								}
								goto nextAlt77
							ok90:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto nextAlt77
								}
							default:
								goto nextAlt77
							}
						}
					}
				}
			ok80:
			loop78:
				{
					position11 := position
					if peekChar('`') {
						goto nextAlt92
					}
					if !p.rules[ruleNonspacechar]() {
						goto nextAlt92
					}
				loop93:
					if peekChar('`') {
						goto out94
					}
					if !p.rules[ruleNonspacechar]() {
						goto out94
					}
					goto loop93
				out94:
					goto ok91
				nextAlt92:
					{
						if position == len(p.Buffer) {
							goto out79
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks4]() {
								goto ok96
							}
							goto out79
						ok96:
							if !matchChar('`') {
								goto out79
							}
						loop97:
							if !matchChar('`') {
								goto out98
							}
							goto loop97
						out98:
							break
						default:
							{
								position13 := position
								if !p.rules[ruleSp]() {
									goto ok99
								}
								if !p.rules[ruleTicks4]() {
									goto ok99
								}
								goto out79
							ok99:
								position = position13
							}
							{
								if position == len(p.Buffer) {
									goto out79
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto out79
									}
									if !p.rules[ruleBlankLine]() {
										goto ok101
									}
									goto out79
								ok101:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto out79
									}
								default:
									goto out79
								}
							}
						}
					}
				ok91:
					goto loop78
				out79:
					position = position11
				}
				end = position
				if !p.rules[ruleSp]() {
					goto nextAlt77
				}
				if !p.rules[ruleTicks4]() {
					goto nextAlt77
				}
				goto ok
			nextAlt77:
				position = position1
				if !p.rules[ruleTicks5]() {
					goto ko
				}
				if !p.rules[ruleSp]() {
					goto ko
				}
				begin = position
				if peekChar('`') {
					goto nextAlt105
				}
				if !p.rules[ruleNonspacechar]() {
					goto nextAlt105
				}
			loop106:
				if peekChar('`') {
					goto out107
				}
				if !p.rules[ruleNonspacechar]() {
					goto out107
				}
				goto loop106
			out107:
				goto ok104
			nextAlt105:
				{
					if position == len(p.Buffer) {
						goto ko
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks5]() {
							goto ok109
						}
						goto ko
					ok109:
						if !matchChar('`') {
							goto ko
						}
					loop110:
						if !matchChar('`') {
							goto out111
						}
						goto loop110
					out111:
						break
					default:
						{
							position14 := position
							if !p.rules[ruleSp]() {
								goto ok112
							}
							if !p.rules[ruleTicks5]() {
								goto ok112
							}
							goto ko
						ok112:
							position = position14
						}
						{
							if position == len(p.Buffer) {
								goto ko
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto ko
								}
								if !p.rules[ruleBlankLine]() {
									goto ok114
								}
								goto ko
							ok114:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto ko
								}
							default:
								goto ko
							}
						}
					}
				}
			ok104:
			loop102:
				{
					position14 := position
					if peekChar('`') {
						goto nextAlt116
					}
					if !p.rules[ruleNonspacechar]() {
						goto nextAlt116
					}
				loop117:
					if peekChar('`') {
						goto out118
					}
					if !p.rules[ruleNonspacechar]() {
						goto out118
					}
					goto loop117
				out118:
					goto ok115
				nextAlt116:
					{
						if position == len(p.Buffer) {
							goto out103
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks5]() {
								goto ok120
							}
							goto out103
						ok120:
							if !matchChar('`') {
								goto out103
							}
						loop121:
							if !matchChar('`') {
								goto out122
							}
							goto loop121
						out122:
							break
						default:
							{
								position16 := position
								if !p.rules[ruleSp]() {
									goto ok123
								}
								if !p.rules[ruleTicks5]() {
									goto ok123
								}
								goto out103
							ok123:
								position = position16
							}
							{
								if position == len(p.Buffer) {
									goto out103
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto out103
									}
									if !p.rules[ruleBlankLine]() {
										goto ok125
									}
									goto out103
								ok125:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto out103
									}
								default:
									goto out103
								}
							}
						}
					}
				ok115:
					goto loop102
				out103:
					position = position14
				}
				end = position
				if !p.rules[ruleSp]() {
					goto ko
				}
				if !p.rules[ruleTicks5]() {
					goto ko
				}
			}
		ok:
			do(87)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 196 RawHtml <- (< (HtmlComment / HtmlBlockScript / HtmlTag) > {   if p.extension.FilterHTML {
		        yy = p.mkList(LIST, nil)
		    } else {
		        yy = p.mkString(yytext)
		        yy.key = HTML
		    }
		}) */
		func() (match bool) {
			position0 := position
			begin = position
			if !p.rules[ruleHtmlComment]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleHtmlBlockScript]() {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !p.rules[ruleHtmlTag]() {
				goto ko
			}
		ok:
			end = position
			do(88)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 197 BlankLine <- (Sp Newline) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 198 Quoted <- ((&[\'] ('\'' (!'\'' .)* '\'')) | (&[\"] ('"' (!'"' .)* '"'))) */
		func() (match bool) {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '\'':
					position++ // matchChar
				loop:
					if position == len(p.Buffer) {
						goto out
					}
					switch p.Buffer[position] {
					case '\'':
						goto out
					default:
						position++
					}
					goto loop
				out:
					if !matchChar('\'') {
						goto ko
					}
				case '"':
					position++ // matchChar
				loop4:
					if position == len(p.Buffer) {
						goto out5
					}
					switch p.Buffer[position] {
					case '"':
						goto out5
					default:
						position++
					}
					goto loop4
				out5:
					if !matchChar('"') {
						goto ko
					}
				default:
					goto ko
				}
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 199 HtmlAttribute <- (((&[\-] '-') | (&[0-9A-Za-z] [A-Za-z0-9]))+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
		func() (match bool) {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
				default:
					if !matchClass(5) {
						goto ko
					}
				}
			}
		loop:
			{
				if position == len(p.Buffer) {
					goto out
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
				default:
					if !matchClass(5) {
						goto out
					}
				}
			}
			goto loop
		out:
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			{
				position1 := position
				if !matchChar('=') {
					goto ko5
				}
				if !p.rules[ruleSpnl]() {
					goto ko5
				}
				if !p.rules[ruleQuoted]() {
					goto nextAlt
				}
				goto ok7
			nextAlt:
				if peekChar('>') {
					goto ko5
				}
				if !p.rules[ruleNonspacechar]() {
					goto ko5
				}
			loop9:
				if peekChar('>') {
					goto out10
				}
				if !p.rules[ruleNonspacechar]() {
					goto out10
				}
				goto loop9
			out10:
			ok7:
				goto ok6
			ko5:
				position = position1
			}
		ok6:
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 200 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
		func() (match bool) {
			position0 := position
			if !matchString("<!--") {
				goto ko
			}
		loop:
			{
				position1 := position
				if !matchString("-->") {
					goto ok
				}
				goto out
			ok:
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			if !matchString("-->") {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 201 HtmlTag <- ('<' Spnl '/'? [A-Za-z0-9]+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() (match bool) {
			position0 := position
			if !matchChar('<') {
				goto ko
			}
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			matchChar('/')
			if !matchClass(5) {
				goto ko
			}
		loop:
			if !matchClass(5) {
				goto out
			}
			goto loop
		out:
			if !p.rules[ruleSpnl]() {
				goto ko
			}
		loop3:
			if !p.rules[ruleHtmlAttribute]() {
				goto out4
			}
			goto loop3
		out4:
			matchChar('/')
			if !p.rules[ruleSpnl]() {
				goto ko
			}
			if !matchChar('>') {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 202 Eof <- !. */
		func() (match bool) {
			if position < len(p.Buffer) {
				return
			}
			match = true
			return
		},
		/* 203 Spacechar <- ((&[\t] '\t') | (&[ ] ' ')) */
		func() (match bool) {
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case '\t':
					position++ // matchChar
				case ' ':
					position++ // matchChar
				default:
					return
				}
			}
			match = true
			return
		},
		/* 204 Nonspacechar <- (!Spacechar !Newline .) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleNewline]() {
				goto ok2
			}
			goto ko
		ok2:
			if !matchDot() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 205 Newline <- ((&[\r] ('\r' '\n'?)) | (&[\n] '\n')) */
		func() (match bool) {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '\r':
					position++ // matchChar
					matchChar('\n')
				case '\n':
					position++ // matchChar
				default:
					goto ko
				}
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 206 Sp <- Spacechar* */
		func() (match bool) {
		loop:
			if !p.rules[ruleSpacechar]() {
				goto out
			}
			goto loop
		out:
			match = true
			return
		},
		/* 207 Spnl <- (Sp (Newline Sp)?) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleSp]() {
				goto ko
			}
			{
				position1 := position
				if !p.rules[ruleNewline]() {
					goto ko1
				}
				if !p.rules[ruleSp]() {
					goto ko1
				}
				goto ok
			ko1:
				position = position1
			}
		ok:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 208 SpecialChar <- ('\'' / '"' / ((&[\\] '\\') | (&[#] '#') | (&[!] '!') | (&[<] '<') | (&[)] ')') | (&[(] '(') | (&[\]] ']') | (&[\[] '[') | (&[&] '&') | (&[`] '`') | (&[_] '_') | (&[*] '*') | (&[~] '~') | (&[\"\'\-.^] ExtendedSpecialChar))) */
		func() (match bool) {
			if !matchChar('\'') {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !matchChar('"') {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case '\\':
					position++ // matchChar
				case '#':
					position++ // matchChar
				case '!':
					position++ // matchChar
				case '<':
					position++ // matchChar
				case ')':
					position++ // matchChar
				case '(':
					position++ // matchChar
				case ']':
					position++ // matchChar
				case '[':
					position++ // matchChar
				case '&':
					position++ // matchChar
				case '`':
					position++ // matchChar
				case '_':
					position++ // matchChar
				case '*':
					position++ // matchChar
				case '~':
					position++ // matchChar
				default:
					if !p.rules[ruleExtendedSpecialChar]() {
						return
					}
				}
			}
		ok:
			match = true
			return
		},
		/* 209 NormalChar <- (!((&[\n\r] Newline) | (&[\t ] Spacechar) | (&[!-#&-*\-.<\[-`~] SpecialChar)) .) */
		func() (match bool) {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto ok
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto ok
					}
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto ok
					}
				default:
					if !p.rules[ruleSpecialChar]() {
						goto ok
					}
				}
			}
			goto ko
		ok:
			if !matchDot() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 210 Alphanumeric <- ((&[\377] '\377') | (&[\376] '\376') | (&[\375] '\375') | (&[\374] '\374') | (&[\373] '\373') | (&[\372] '\372') | (&[\371] '\371') | (&[\370] '\370') | (&[\367] '\367') | (&[\366] '\366') | (&[\365] '\365') | (&[\364] '\364') | (&[\363] '\363') | (&[\362] '\362') | (&[\361] '\361') | (&[\360] '\360') | (&[\357] '\357') | (&[\356] '\356') | (&[\355] '\355') | (&[\354] '\354') | (&[\353] '\353') | (&[\352] '\352') | (&[\351] '\351') | (&[\350] '\350') | (&[\347] '\347') | (&[\346] '\346') | (&[\345] '\345') | (&[\344] '\344') | (&[\343] '\343') | (&[\342] '\342') | (&[\341] '\341') | (&[\340] '\340') | (&[\337] '\337') | (&[\336] '\336') | (&[\335] '\335') | (&[\334] '\334') | (&[\333] '\333') | (&[\332] '\332') | (&[\331] '\331') | (&[\330] '\330') | (&[\327] '\327') | (&[\326] '\326') | (&[\325] '\325') | (&[\324] '\324') | (&[\323] '\323') | (&[\322] '\322') | (&[\321] '\321') | (&[\320] '\320') | (&[\317] '\317') | (&[\316] '\316') | (&[\315] '\315') | (&[\314] '\314') | (&[\313] '\313') | (&[\312] '\312') | (&[\311] '\311') | (&[\310] '\310') | (&[\307] '\307') | (&[\306] '\306') | (&[\305] '\305') | (&[\304] '\304') | (&[\303] '\303') | (&[\302] '\302') | (&[\301] '\301') | (&[\300] '\300') | (&[\277] '\277') | (&[\276] '\276') | (&[\275] '\275') | (&[\274] '\274') | (&[\273] '\273') | (&[\272] '\272') | (&[\271] '\271') | (&[\270] '\270') | (&[\267] '\267') | (&[\266] '\266') | (&[\265] '\265') | (&[\264] '\264') | (&[\263] '\263') | (&[\262] '\262') | (&[\261] '\261') | (&[\260] '\260') | (&[\257] '\257') | (&[\256] '\256') | (&[\255] '\255') | (&[\254] '\254') | (&[\253] '\253') | (&[\252] '\252') | (&[\251] '\251') | (&[\250] '\250') | (&[\247] '\247') | (&[\246] '\246') | (&[\245] '\245') | (&[\244] '\244') | (&[\243] '\243') | (&[\242] '\242') | (&[\241] '\241') | (&[\240] '\240') | (&[\237] '\237') | (&[\236] '\236') | (&[\235] '\235') | (&[\234] '\234') | (&[\233] '\233') | (&[\232] '\232') | (&[\231] '\231') | (&[\230] '\230') | (&[\227] '\227') | (&[\226] '\226') | (&[\225] '\225') | (&[\224] '\224') | (&[\223] '\223') | (&[\222] '\222') | (&[\221] '\221') | (&[\220] '\220') | (&[\217] '\217') | (&[\216] '\216') | (&[\215] '\215') | (&[\214] '\214') | (&[\213] '\213') | (&[\212] '\212') | (&[\211] '\211') | (&[\210] '\210') | (&[\207] '\207') | (&[\206] '\206') | (&[\205] '\205') | (&[\204] '\204') | (&[\203] '\203') | (&[\202] '\202') | (&[\201] '\201') | (&[\200] '\200') | (&[0-9A-Za-z] [0-9A-Za-z])) */
		func() (match bool) {
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case '\377':
					position++ // matchChar
				case '\376':
					position++ // matchChar
				case '\375':
					position++ // matchChar
				case '\374':
					position++ // matchChar
				case '\373':
					position++ // matchChar
				case '\372':
					position++ // matchChar
				case '\371':
					position++ // matchChar
				case '\370':
					position++ // matchChar
				case '\367':
					position++ // matchChar
				case '\366':
					position++ // matchChar
				case '\365':
					position++ // matchChar
				case '\364':
					position++ // matchChar
				case '\363':
					position++ // matchChar
				case '\362':
					position++ // matchChar
				case '\361':
					position++ // matchChar
				case '\360':
					position++ // matchChar
				case '\357':
					position++ // matchChar
				case '\356':
					position++ // matchChar
				case '\355':
					position++ // matchChar
				case '\354':
					position++ // matchChar
				case '\353':
					position++ // matchChar
				case '\352':
					position++ // matchChar
				case '\351':
					position++ // matchChar
				case '\350':
					position++ // matchChar
				case '\347':
					position++ // matchChar
				case '\346':
					position++ // matchChar
				case '\345':
					position++ // matchChar
				case '\344':
					position++ // matchChar
				case '\343':
					position++ // matchChar
				case '\342':
					position++ // matchChar
				case '\341':
					position++ // matchChar
				case '\340':
					position++ // matchChar
				case '\337':
					position++ // matchChar
				case '\336':
					position++ // matchChar
				case '\335':
					position++ // matchChar
				case '\334':
					position++ // matchChar
				case '\333':
					position++ // matchChar
				case '\332':
					position++ // matchChar
				case '\331':
					position++ // matchChar
				case '\330':
					position++ // matchChar
				case '\327':
					position++ // matchChar
				case '\326':
					position++ // matchChar
				case '\325':
					position++ // matchChar
				case '\324':
					position++ // matchChar
				case '\323':
					position++ // matchChar
				case '\322':
					position++ // matchChar
				case '\321':
					position++ // matchChar
				case '\320':
					position++ // matchChar
				case '\317':
					position++ // matchChar
				case '\316':
					position++ // matchChar
				case '\315':
					position++ // matchChar
				case '\314':
					position++ // matchChar
				case '\313':
					position++ // matchChar
				case '\312':
					position++ // matchChar
				case '\311':
					position++ // matchChar
				case '\310':
					position++ // matchChar
				case '\307':
					position++ // matchChar
				case '\306':
					position++ // matchChar
				case '\305':
					position++ // matchChar
				case '\304':
					position++ // matchChar
				case '\303':
					position++ // matchChar
				case '\302':
					position++ // matchChar
				case '\301':
					position++ // matchChar
				case '\300':
					position++ // matchChar
				case '\277':
					position++ // matchChar
				case '\276':
					position++ // matchChar
				case '\275':
					position++ // matchChar
				case '\274':
					position++ // matchChar
				case '\273':
					position++ // matchChar
				case '\272':
					position++ // matchChar
				case '\271':
					position++ // matchChar
				case '\270':
					position++ // matchChar
				case '\267':
					position++ // matchChar
				case '\266':
					position++ // matchChar
				case '\265':
					position++ // matchChar
				case '\264':
					position++ // matchChar
				case '\263':
					position++ // matchChar
				case '\262':
					position++ // matchChar
				case '\261':
					position++ // matchChar
				case '\260':
					position++ // matchChar
				case '\257':
					position++ // matchChar
				case '\256':
					position++ // matchChar
				case '\255':
					position++ // matchChar
				case '\254':
					position++ // matchChar
				case '\253':
					position++ // matchChar
				case '\252':
					position++ // matchChar
				case '\251':
					position++ // matchChar
				case '\250':
					position++ // matchChar
				case '\247':
					position++ // matchChar
				case '\246':
					position++ // matchChar
				case '\245':
					position++ // matchChar
				case '\244':
					position++ // matchChar
				case '\243':
					position++ // matchChar
				case '\242':
					position++ // matchChar
				case '\241':
					position++ // matchChar
				case '\240':
					position++ // matchChar
				case '\237':
					position++ // matchChar
				case '\236':
					position++ // matchChar
				case '\235':
					position++ // matchChar
				case '\234':
					position++ // matchChar
				case '\233':
					position++ // matchChar
				case '\232':
					position++ // matchChar
				case '\231':
					position++ // matchChar
				case '\230':
					position++ // matchChar
				case '\227':
					position++ // matchChar
				case '\226':
					position++ // matchChar
				case '\225':
					position++ // matchChar
				case '\224':
					position++ // matchChar
				case '\223':
					position++ // matchChar
				case '\222':
					position++ // matchChar
				case '\221':
					position++ // matchChar
				case '\220':
					position++ // matchChar
				case '\217':
					position++ // matchChar
				case '\216':
					position++ // matchChar
				case '\215':
					position++ // matchChar
				case '\214':
					position++ // matchChar
				case '\213':
					position++ // matchChar
				case '\212':
					position++ // matchChar
				case '\211':
					position++ // matchChar
				case '\210':
					position++ // matchChar
				case '\207':
					position++ // matchChar
				case '\206':
					position++ // matchChar
				case '\205':
					position++ // matchChar
				case '\204':
					position++ // matchChar
				case '\203':
					position++ // matchChar
				case '\202':
					position++ // matchChar
				case '\201':
					position++ // matchChar
				case '\200':
					position++ // matchChar
				default:
					if !matchClass(4) {
						return
					}
				}
			}
			match = true
			return
		},
		/* 211 AlphanumericAscii <- [A-Za-z0-9] */
		func() (match bool) {
			if !matchClass(5) {
				return
			}
			match = true
			return
		},
		/* 212 Digit <- [0-9] */
		func() (match bool) {
			if !matchClass(0) {
				return
			}
			match = true
			return
		},
		/* 213 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() (match bool) {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto ko
			}
			if !matchChar('#') {
				goto ko
			}
			if !matchClass(6) {
				goto ko
			}
			if !matchClass(7) {
				goto ko
			}
		loop:
			if !matchClass(7) {
				goto out
			}
			goto loop
		out:
			if !matchChar(';') {
				goto ko
			}
			end = position
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 214 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() (match bool) {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto ko
			}
			if !matchChar('#') {
				goto ko
			}
			if !matchClass(0) {
				goto ko
			}
		loop:
			if !matchClass(0) {
				goto out
			}
			goto loop
		out:
			end = position
			if !matchChar(';') {
				goto ko
			}
			end = position
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 215 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() (match bool) {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto ko
			}
			if !matchClass(5) {
				goto ko
			}
		loop:
			if !matchClass(5) {
				goto out
			}
			goto loop
		out:
			if !matchChar(';') {
				goto ko
			}
			end = position
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 216 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() (match bool) {
			if !matchString("   ") {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !matchString("  ") {
				goto nextAlt3
			}
			goto ok
		nextAlt3:
			if !matchChar(' ') {
				goto nextAlt4
			}
			goto ok
		nextAlt4:
		ok:
			match = true
			return
		},
		/* 217 Indent <- ((&[ ] '    ') | (&[\t] '\t')) */
		func() (match bool) {
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case ' ':
					position++
					if !matchString("   ") {
						return
					}
				case '\t':
					position++ // matchChar
				default:
					return
				}
			}
			match = true
			return
		},
		/* 218 IndentedLine <- (Indent Line) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto ko
			}
			if !p.rules[ruleLine]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 219 OptionallyIndentedLine <- (Indent? Line) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto ko1
			}
		ko1:
			if !p.rules[ruleLine]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 220 StartList <- (&. { yy = nil }) */
		func() (match bool) {
			if !(position < len(p.Buffer)) {
				return
			}
			do(89)
			match = true
			return
		},
		/* 221 Line <- (RawLine { yy = p.mkString(yytext) }) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleRawLine]() {
				goto ko
			}
			do(90)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 222 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > !.)) */
		func() (match bool) {
			position0 := position
			{
				position1 := position
				begin = position
			loop:
				if position == len(p.Buffer) {
					goto out
				}
				switch p.Buffer[position] {
				case '\r', '\n':
					goto out
				default:
					position++
				}
				goto loop
			out:
				if !p.rules[ruleNewline]() {
					goto nextAlt
				}
				end = position
				goto ok
			nextAlt:
				position = position1
				begin = position
				if !matchDot() {
					goto ko
				}
			loop5:
				if !matchDot() {
					goto out6
				}
				goto loop5
			out6:
				end = position
				if position < len(p.Buffer) {
					goto ko
				}
			}
		ok:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 223 SkipBlock <- (HtmlBlock / ((!'#' !SetextBottom1 !SetextBottom2 !BlankLine RawLine)+ BlankLine*) / BlankLine+ / RawLine) */
		func() (match bool) {
			position0 := position
			{
				position1 := position
				if !p.rules[ruleHtmlBlock]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				if peekChar('#') {
					goto nextAlt3
				}
				if !p.rules[ruleSetextBottom1]() {
					goto ok6
				}
				goto nextAlt3
			ok6:
				if !p.rules[ruleSetextBottom2]() {
					goto ok7
				}
				goto nextAlt3
			ok7:
				if !p.rules[ruleBlankLine]() {
					goto ok8
				}
				goto nextAlt3
			ok8:
				if !p.rules[ruleRawLine]() {
					goto nextAlt3
				}
			loop:
				{
					position2 := position
					if peekChar('#') {
						goto out
					}
					if !p.rules[ruleSetextBottom1]() {
						goto ok9
					}
					goto out
				ok9:
					if !p.rules[ruleSetextBottom2]() {
						goto ok10
					}
					goto out
				ok10:
					if !p.rules[ruleBlankLine]() {
						goto ok11
					}
					goto out
				ok11:
					if !p.rules[ruleRawLine]() {
						goto out
					}
					goto loop
				out:
					position = position2
				}
			loop12:
				if !p.rules[ruleBlankLine]() {
					goto out13
				}
				goto loop12
			out13:
				goto ok
			nextAlt3:
				position = position1
				if !p.rules[ruleBlankLine]() {
					goto nextAlt14
				}
			loop15:
				if !p.rules[ruleBlankLine]() {
					goto out16
				}
				goto loop15
			out16:
				goto ok
			nextAlt14:
				position = position1
				if !p.rules[ruleRawLine]() {
					goto ko
				}
			}
		ok:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 224 ExtendedSpecialChar <- ((&[^] (&{p.extension.Notes} '^')) | (&[\"\'\-.] (&{p.extension.Smart} ((&[\"] '"') | (&[\'] '\'') | (&[\-] '-') | (&[.] '.'))))) */
		func() (match bool) {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '^':
					if !(p.extension.Notes) {
						goto ko
					}
					if !matchChar('^') {
						goto ko
					}
				default:
					if !(p.extension.Smart) {
						goto ko
					}
					{
						if position == len(p.Buffer) {
							goto ko
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
						case '\'':
							position++ // matchChar
						case '-':
							position++ // matchChar
						case '.':
							position++ // matchChar
						default:
							goto ko
						}
					}
				}
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 225 Smart <- (&{p.extension.Smart} (SingleQuoted / ((&[\'] Apostrophe) | (&[\"] DoubleQuoted) | (&[\-] Dash) | (&[.] Ellipsis)))) */
		func() (match bool) {
			if !(p.extension.Smart) {
				return
			}
			if !p.rules[ruleSingleQuoted]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			{
				if position == len(p.Buffer) {
					return
				}
				switch p.Buffer[position] {
				case '\'':
					if !p.rules[ruleApostrophe]() {
						return
					}
				case '"':
					if !p.rules[ruleDoubleQuoted]() {
						return
					}
				case '-':
					if !p.rules[ruleDash]() {
						return
					}
				case '.':
					if !p.rules[ruleEllipsis]() {
						return
					}
				default:
					return
				}
			}
		ok:
			match = true
			return
		},
		/* 226 Apostrophe <- ('\'' { yy = p.mkElem(APOSTROPHE) }) */
		func() (match bool) {
			position0 := position
			if !matchChar('\'') {
				goto ko
			}
			do(91)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 227 Ellipsis <- (('...' / '. . .') { yy = p.mkElem(ELLIPSIS) }) */
		func() (match bool) {
			position0 := position
			if !matchString("...") {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !matchString(". . .") {
				goto ko
			}
		ok:
			do(92)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 228 Dash <- (EmDash / EnDash) */
		func() (match bool) {
			if !p.rules[ruleEmDash]() {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !p.rules[ruleEnDash]() {
				return
			}
		ok:
			match = true
			return
		},
		/* 229 EnDash <- ('-' &[0-9] { yy = p.mkElem(ENDASH) }) */
		func() (match bool) {
			position0 := position
			if !matchChar('-') {
				goto ko
			}
			if !peekClass(0) {
				goto ko
			}
			do(93)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 230 EmDash <- (('---' / '--') { yy = p.mkElem(EMDASH) }) */
		func() (match bool) {
			position0 := position
			if !matchString("---") {
				goto nextAlt
			}
			goto ok
		nextAlt:
			if !matchString("--") {
				goto ko
			}
		ok:
			do(94)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 231 SingleQuoteStart <- ('\'' !((&[\n\r] Newline) | (&[\t ] Spacechar))) */
		func() (match bool) {
			position0 := position
			if !matchChar('\'') {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ok
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto ok
					}
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto ok
					}
				default:
					goto ok
				}
			}
			goto ko
		ok:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 232 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() (match bool) {
			position0 := position
			if !matchChar('\'') {
				goto ko
			}
			if !p.rules[ruleAlphanumeric]() {
				goto ok
			}
			goto ko
		ok:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 233 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = p.mkList(SINGLEQUOTED, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto ko
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleSingleQuoteEnd]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleInline]() {
				goto ko
			}
			doarg(yySet, -2)
			do(95)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto ok4
				}
				goto out
			ok4:
				if !p.rules[ruleInline]() {
					goto out
				}
				doarg(yySet, -2)
				do(95)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto ko
			}
			do(96)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 234 DoubleQuoteStart <- '"' */
		func() (match bool) {
			if !matchChar('"') {
				return
			}
			match = true
			return
		},
		/* 235 DoubleQuoteEnd <- '"' */
		func() (match bool) {
			if !matchChar('"') {
				return
			}
			match = true
			return
		},
		/* 236 DoubleQuoted <- ('"' StartList (!'"' Inline { a = cons(b, a) })+ '"' { yy = p.mkList(DOUBLEQUOTED, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('"') {
				goto ko
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if peekChar('"') {
				goto ko
			}
			if !p.rules[ruleInline]() {
				goto ko
			}
			doarg(yySet, -2)
			do(97)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if peekChar('"') {
					goto out
				}
				if !p.rules[ruleInline]() {
					goto out
				}
				doarg(yySet, -2)
				do(97)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !matchChar('"') {
				goto ko
			}
			do(98)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 237 NoteReference <- (&{p.extension.Notes} RawNoteReference {
		    if match, ok := p.find_note(ref.contents.str); ok {
		        yy = p.mkElem(NOTE)
		        yy.children = match.children
		        yy.contents.str = ""
		    } else {
		        yy = p.mkString("[^"+ref.contents.str+"]")
		    }
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto ko
			}
			if !p.rules[ruleRawNoteReference]() {
				goto ko
			}
			doarg(yySet, -1)
			do(99)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 238 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = p.mkString(yytext) }) */
		func() (match bool) {
			position0 := position
			if !matchString("[^") {
				goto ko
			}
			begin = position
			if !p.rules[ruleNewline]() {
				goto ok
			}
			goto ko
		ok:
			if peekChar(']') {
				goto ko
			}
			if !matchDot() {
				goto ko
			}
		loop:
			{
				position1 := position
				if !p.rules[ruleNewline]() {
					goto ok4
				}
				goto out
			ok4:
				if peekChar(']') {
					goto out
				}
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position = position1
			}
			end = position
			if !matchChar(']') {
				goto ko
			}
			do(100)
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 239 Note <- (&{p.extension.Notes} NonindentSpace RawNoteReference ':' Sp StartList (RawNoteBlock { a = cons(yy, a) }) (&Indent RawNoteBlock { a = cons(yy, a) })* {   yy = p.mkList(NOTE, a)
		    yy.contents.str = ref.contents.str
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !(p.extension.Notes) {
				goto ko
			}
			if !p.rules[ruleNonindentSpace]() {
				goto ko
			}
			if !p.rules[ruleRawNoteReference]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto ko
			}
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -2)
			if !p.rules[ruleRawNoteBlock]() {
				goto ko
			}
			do(101)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				{
					position2 := position
					if !p.rules[ruleIndent]() {
						goto out
					}
					position = position2
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto out
				}
				do(102)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			do(103)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 240 InlineNote <- (&{p.extension.Notes} '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = p.mkList(NOTE, a)
		   yy.contents.str = "" }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto ko
			}
			if !matchString("^[") {
				goto ko
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto ko
			}
			if !p.rules[ruleInline]() {
				goto ko
			}
			do(104)
		loop:
			{
				position1 := position
				if peekChar(']') {
					goto out
				}
				if !p.rules[ruleInline]() {
					goto out
				}
				do(104)
				goto loop
			out:
				position = position1
			}
			if !matchChar(']') {
				goto ko
			}
			do(105)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 241 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				{
					position2, thunkPosition2 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto nextAlt
					}
					doarg(yySet, -2)
					do(106)
					goto ok
				nextAlt:
					position, thunkPosition = position2, thunkPosition2
					if !p.rules[ruleSkipBlock]() {
						goto out
					}
				}
			ok:
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			do(107)
			if !(p.commit(thunkPosition0)) {
				goto ko
			}
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 242 RawNoteBlock <- (StartList (!BlankLine OptionallyIndentedLine { a = cons(yy, a) })+ (< BlankLine* > { a = cons(p.mkString(yytext), a) }) {   yy = p.mkStringFromList(a, true)
		    yy.key = RAW
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto ok
			}
			goto ko
		ok:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto ko
			}
			do(108)
		loop:
			{
				position1 := position
				if !p.rules[ruleBlankLine]() {
					goto ok4
				}
				goto out
			ok4:
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto out
				}
				do(108)
				goto loop
			out:
				position = position1
			}
			begin = position
		loop5:
			if !p.rules[ruleBlankLine]() {
				goto out6
			}
			goto loop5
		out6:
			end = position
			do(109)
			do(110)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 243 DefinitionList <- (&{p.extension.Dlists} StartList (Definition { a = cons(yy, a) })+ { yy = p.mkList(DEFINITIONLIST, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Dlists) {
				goto ko
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto ko
			}
			do(111)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto out
				}
				do(111)
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			do(112)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 244 Definition <- (&(NonindentSpace !Defmark Nonspacechar RawLine BlankLine? Defmark) StartList (DListTitle { a = cons(yy, a) })+ (DefTight / DefLoose) {
			for e := yy.children; e != nil; e = e.next {
				e.key = DEFDATA
			}
			a = cons(yy, a)
		} { yy = p.mkList(LIST, a) }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1 := position
				if !p.rules[ruleNonindentSpace]() {
					goto ko
				}
				if !p.rules[ruleDefmark]() {
					goto ok
				}
				goto ko
			ok:
				if !p.rules[ruleNonspacechar]() {
					goto ko
				}
				if !p.rules[ruleRawLine]() {
					goto ko
				}
				if !p.rules[ruleBlankLine]() {
					goto ko3
				}
			ko3:
				if !p.rules[ruleDefmark]() {
					goto ko
				}
				position = position1
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto ko
			}
			do(113)
		loop:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto out
				}
				do(113)
				goto loop
			out:
				position, thunkPosition = position2, thunkPosition2
			}
			if !p.rules[ruleDefTight]() {
				goto nextAlt
			}
			goto ok7
		nextAlt:
			if !p.rules[ruleDefLoose]() {
				goto ko
			}
		ok7:
			do(114)
			do(115)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 245 DListTitle <- (NonindentSpace !Defmark &Nonspacechar StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline {	yy = p.mkList(LIST, a)
			yy.key = DEFTITLE
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto ko
			}
			if !p.rules[ruleDefmark]() {
				goto ok
			}
			goto ko
		ok:
			{
				position1 := position
				if !p.rules[ruleNonspacechar]() {
					goto ko
				}
				position = position1
			}
			if !p.rules[ruleStartList]() {
				goto ko
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto ok5
			}
			goto ko
		ok5:
			if !p.rules[ruleInline]() {
				goto ko
			}
			do(116)
		loop:
			{
				position2 := position
				if !p.rules[ruleEndline]() {
					goto ok6
				}
				goto out
			ok6:
				if !p.rules[ruleInline]() {
					goto out
				}
				do(116)
				goto loop
			out:
				position = position2
			}
			if !p.rules[ruleSp]() {
				goto ko
			}
			if !p.rules[ruleNewline]() {
				goto ko
			}
			do(117)
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 246 DefTight <- (&Defmark ListTight) */
		func() (match bool) {
			{
				position1 := position
				if !p.rules[ruleDefmark]() {
					return
				}
				position = position1
			}
			if !p.rules[ruleListTight]() {
				return
			}
			match = true
			return
		},
		/* 247 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto ko
			}
			{
				position1 := position
				if !p.rules[ruleDefmark]() {
					goto ko
				}
				position = position1
			}
			if !p.rules[ruleListLoose]() {
				goto ko
			}
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 248 Defmark <- (NonindentSpace ((&[~] '~') | (&[:] ':')) Spacechar+) */
		func() (match bool) {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto ko
			}
			{
				if position == len(p.Buffer) {
					goto ko
				}
				switch p.Buffer[position] {
				case '~':
					position++ // matchChar
				case ':':
					position++ // matchChar
				default:
					goto ko
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto ko
			}
		loop:
			if !p.rules[ruleSpacechar]() {
				goto out
			}
			goto loop
		out:
			match = true
			return
		ko:
			position = position0
			return
		},
		/* 249 DefMarker <- (&{p.extension.Dlists} Defmark) */
		func() (match bool) {
			if !(p.extension.Dlists) {
				return
			}
			if !p.rules[ruleDefmark]() {
				return
			}
			match = true
			return
		},
		nil,
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
	STRIKE:         "STRIKE",
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
