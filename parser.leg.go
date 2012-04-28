
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
	"strings"
	"log"
	"sync"
)

// Semantic value of a parsing action.
type element struct {
	key	int
	contents
	children	*element
	next		*element
}

// Information (label, URL and title) for a link.
type link struct {
	label	*element
	url		string
	title	string
}

// Union for contents of an Element (string, list, or link).
type contents struct {
	str	string
	*link
}

// Types of semantic values returned by parsers.
const (
	LIST	= iota	/* A generic list of values. For ordered and bullet lists, see below. */
	RAW				/* Raw markdown to be processed further */
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
	H1	/* Code assumes that H1..6 are in order. */
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
	extension		Options
	tree				*element	/* Results of parse. */
	references			*element	/* List of link references found. */
	notes				*element	/* List of footnotes found. */
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
	ruleOneStarOpen
	ruleOneStarClose
	ruleEmphStar
	ruleOneUlOpen
	ruleOneUlClose
	ruleEmphUl
	ruleStrong
	ruleTwoStarOpen
	ruleTwoStarClose
	ruleStrongStar
	ruleTwoUlOpen
	ruleTwoUlClose
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
	ruleNonAlphanumeric
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
	Buffer string
	Min, Max int
	rules [250]func() bool
	ResetBuffer	func(string) string
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

func	(e *errPos) String() string {
	return fmt.Sprintf("%d:%d", e.Line, e.Pos)
}

type unexpectedCharError struct {
	After, At	errPos
	Char	byte
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
			 yy = a; yy.key = PARA 
			yyval[yyp-1] = a
		},
		/* 4 Plain */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a; yy.key = PLAIN 
			yyval[yyp-1] = a
		},
		/* 5 AtxStart */
		func(yytext string, _ int) {
			 yy = mk_element(H1 + (len(yytext) - 1)) 
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
			a := yyval[yyp-2]
			s := yyval[yyp-1]
			 yy = mk_list(s.key, a)
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
			 yy = mk_list(H1, a) 
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
			 yy = mk_list(H2, a) 
			yyval[yyp-1] = a
		},
		/* 12 BlockQuote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			  yy = mk_element(BLOCKQUOTE)
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
			 a = cons(mk_str("\n"), a) 
			yyval[yyp-1] = a
		},
		/* 16 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   yy = mk_str_from_list(a, true)
                     yy.key = RAW
                 
			yyval[yyp-1] = a
		},
		/* 17 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(mk_str("\n"), a) 
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
			 yy = mk_str_from_list(a, false) 
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
			 yy = mk_str_from_list(a, false)
                 yy.key = VERBATIM 
			yyval[yyp-1] = a
		},
		/* 22 HorizontalRule */
		func(yytext string, _ int) {
			 yy = mk_element(HRULE) 
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
			 yy = mk_list(LIST, a) 
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
			 yy = mk_list(LIST, a) 
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
			
               raw := mk_str_from_list(a, false)
               raw.key = RAW
               yy = mk_element(LISTITEM)
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
			
               raw := mk_str_from_list(a, false)
               raw.key = RAW
               yy = mk_element(LISTITEM)
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
			 yy = mk_str_from_list(a, false) 
			yyval[yyp-1] = a
		},
		/* 37 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   if len(yytext) == 0 {
                                   a = cons(mk_str("\001"), a) // block separator
                              } else {
                                   a = cons(mk_str(yytext), a)
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
			  yy = mk_str_from_list(a, false) 
			yyval[yyp-1] = a
		},
		/* 40 OrderedList */
		func(yytext string, _ int) {
			 yy.key = ORDEREDLIST 
		},
		/* 41 HtmlBlock */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTMLBLOCK
                }
            
		},
		/* 42 StyleBlock */
		func(yytext string, _ int) {
			   if p.extension.FilterStyles {
                        yy = mk_list(LIST, nil)
                    } else {
                        yy = mk_str(yytext)
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
			c := yyval[yyp-2]
			a := yyval[yyp-1]
			 a = cons(c, a) 
			yyval[yyp-2] = c
			yyval[yyp-1] = a
		},
		/* 45 Inlines */
		func(yytext string, _ int) {
			c := yyval[yyp-2]
			a := yyval[yyp-1]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = c
		},
		/* 46 Space */
		func(yytext string, _ int) {
			 yy = mk_str(" ")
          yy.key = SPACE 
		},
		/* 47 Str */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 48 EscapedChar */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 49 Entity */
		func(yytext string, _ int) {
			 yy = mk_str(yytext); yy.key = HTML 
		},
		/* 50 NormalEndline */
		func(yytext string, _ int) {
			 yy = mk_str("\n")
                    yy.key = SPACE 
		},
		/* 51 TerminalEndline */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 52 LineBreak */
		func(yytext string, _ int) {
			 yy = mk_element(LINEBREAK) 
		},
		/* 53 Symbol */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 54 UlOrStarLine */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 55 OneStarClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 56 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 57 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 58 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(EMPH, a) 
			yyval[yyp-1] = a
		},
		/* 59 OneUlClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 60 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 61 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 62 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(EMPH, a) 
			yyval[yyp-1] = a
		},
		/* 63 TwoStarClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 64 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 65 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 66 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(STRONG, a) 
			yyval[yyp-1] = a
		},
		/* 67 TwoUlClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 68 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 69 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 70 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(STRONG, a) 
			yyval[yyp-1] = a
		},
		/* 71 Image */
		func(yytext string, _ int) {
				if yy.key == LINK {
			yy.key = IMAGE
		} else {
			result := yy
			yy.children = cons(mk_str("!"), result.children)
		}
	
		},
		/* 72 ReferenceLinkDouble */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			
                           if match, found := p.findReference(b.children); found {
                               yy = mk_link(a.children, match.url, match.title);
                               a = nil
                               b = nil
                           } else {
                               result := mk_element(LIST)
                               result.children = cons(mk_str("["), cons(a, cons(mk_str("]"), cons(mk_str(yytext),
                                                   cons(mk_str("["), cons(b, mk_str("]")))))))
                               yy = result
                           }
                       
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 73 ReferenceLinkSingle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
                           if match, found := p.findReference(a.children); found {
                               yy = mk_link(a.children, match.url, match.title)
                               a = nil
                           } else {
                               result := mk_element(LIST)
                               result.children = cons(mk_str("["), cons(a, cons(mk_str("]"), mk_str(yytext))));
                               yy = result
                           }
                       
			yyval[yyp-1] = a
		},
		/* 74 ExplicitLink */
		func(yytext string, _ int) {
			t := yyval[yyp-1]
			l := yyval[yyp-2]
			s := yyval[yyp-3]
			 yy = mk_link(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil 
			yyval[yyp-2] = l
			yyval[yyp-3] = s
			yyval[yyp-1] = t
		},
		/* 75 Source */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 76 Title */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 77 AutoLinkUrl */
		func(yytext string, _ int) {
			   yy = mk_link(mk_str(yytext), yytext, "") 
		},
		/* 78 AutoLinkEmail */
		func(yytext string, _ int) {
			
                    yy = mk_link(mk_str(yytext), "mailto:"+yytext, "")
                
		},
		/* 79 Reference */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			l := yyval[yyp-2]
			t := yyval[yyp-3]
			 yy = mk_link(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE 
			yyval[yyp-3] = t
			yyval[yyp-1] = s
			yyval[yyp-2] = l
		},
		/* 80 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 81 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 82 RefSrc */
		func(yytext string, _ int) {
			 yy = mk_str(yytext)
           yy.key = HTML 
		},
		/* 83 RefTitle */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 84 References */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 85 References */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 p.references = reverse(a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 86 Code */
		func(yytext string, _ int) {
			 yy = mk_str(yytext); yy.key = CODE 
		},
		/* 87 RawHtml */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTML
                }
            
		},
		/* 88 StartList */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 89 Line */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 90 Apostrophe */
		func(yytext string, _ int) {
			 yy = mk_element(APOSTROPHE) 
		},
		/* 91 Ellipsis */
		func(yytext string, _ int) {
			 yy = mk_element(ELLIPSIS) 
		},
		/* 92 EnDash */
		func(yytext string, _ int) {
			 yy = mk_element(ENDASH) 
		},
		/* 93 EmDash */
		func(yytext string, _ int) {
			 yy = mk_element(EMDASH) 
		},
		/* 94 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 95 SingleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-2]
			a := yyval[yyp-1]
			 yy = mk_list(SINGLEQUOTED, a) 
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 96 DoubleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 97 DoubleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 yy = mk_list(DOUBLEQUOTED, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 98 NoteReference */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			
                    if match, ok := p.find_note(ref.contents.str); ok {
                        yy = mk_element(NOTE)
                        yy.children = match.children
                        yy.contents.str = ""
                    } else {
                        yy = mk_str("[^"+ref.contents.str+"]")
                    }
                
			yyval[yyp-1] = ref
		},
		/* 99 RawNoteReference */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 100 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 101 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = ref
		},
		/* 102 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			ref := yyval[yyp-1]
			   yy = mk_list(NOTE, a)
                    yy.contents.str = ref.contents.str
                
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 103 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 104 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(NOTE, a)
                  yy.contents.str = "" 
			yyval[yyp-1] = a
		},
		/* 105 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 106 Notes */
		func(yytext string, _ int) {
			b := yyval[yyp-2]
			a := yyval[yyp-1]
			 p.notes = reverse(a) 
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 107 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 108 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(mk_str(yytext), a) 
			yyval[yyp-1] = a
		},
		/* 109 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   yy = mk_str_from_list(a, true)
                    yy.key = RAW
                
			yyval[yyp-1] = a
		},
		/* 110 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 111 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(DEFINITIONLIST, a) 
			yyval[yyp-1] = a
		},
		/* 112 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 113 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			
			yyval[yyp-1] = a
		},
		/* 114 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 115 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 116 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
				yy = mk_list(LIST, a)
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
		yyPush = 117 + iota
		yyPop
		yySet
	)

	type thunk struct {
		action uint8
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
		thunks[thunkPosition].action = action
		if arg != 0 {
			thunks[thunkPosition].begin = arg // use begin to store an argument
		} else {
			thunks[thunkPosition].begin = begin
		}
		thunks[thunkPosition].end = end
		thunkPosition++
	}
	do := func(action uint8) {
		doarg(action, 0)
	}

	p.ResetBuffer = func(s string) (old string) {
		if p.Max < len(p.Buffer) {
			old = p.Buffer[position:]
		}
		p.Buffer = s
		thunkPosition = 0
		position = 0
		p.Min = 0
		p.Max = 0
		return
	}

	commit := func(thunkPosition0 int) bool {
		if thunkPosition0 == 0 {
			for i := 0; i < thunkPosition; i++ {
				b := thunks[i].begin
				e := thunks[i].end
				s := ""
				if b >= 0 && e <= len(p.Buffer) && b <= e {
					s = p.Buffer[b:e]
				}
				magic := b
				actions[thunks[i].action](s, magic)
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
	9:	{0, 6, 0, 0, 3, 82, 0, 252, 0, 0, 0, 32, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	3:	{0, 0, 0, 0, 0, 40, 255, 3, 254, 255, 255, 135, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	1:	{0, 0, 0, 0, 10, 111, 0, 80, 0, 0, 0, 184, 1, 0, 0, 56, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	0:	{0, 0, 0, 0, 0, 0, 255, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	5:	{0, 0, 0, 0, 0, 0, 255, 3, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	8:	{0, 0, 0, 0, 0, 0, 255, 3, 126, 0, 0, 0, 126, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	2:	{0, 0, 0, 0, 0, 0, 0, 0, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	6:	{0, 0, 0, 0, 0, 0, 255, 3, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	7:	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	4:	{0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 31, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
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
		/* 6 AtxStart <- (&'#' < ('######' / '#####' / '####' / '###' / '##' / '#') > { yy = mk_element(H1 + (len(yytext) - 1)) }) */
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
		/* 7 AtxHeading <- (AtxStart Sp? StartList (AtxInline { a = cons(yy, a) })+ (Sp? '#'* Sp)? Newline { yy = mk_list(s.key, a)
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
		/* 9 SetextBottom1 <- ('===' '='* Newline) */
		func() bool {
			position0 := position
			if !matchString("===") {
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
		/* 10 SetextBottom2 <- ('---' '-'* Newline) */
		func() bool {
			position0 := position
			if !matchString("---") {
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
		/* 11 SetextHeading1 <- (&(RawLine SetextBottom1) StartList (!Endline Inline { a = cons(yy, a) })+ Newline SetextBottom1 { yy = mk_list(H1, a) }) */
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
		/* 12 SetextHeading2 <- (&(RawLine SetextBottom2) StartList (!Endline Inline { a = cons(yy, a) })+ Newline SetextBottom2 { yy = mk_list(H2, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position65 := position
				if !p.rules[ruleRawLine]() {
					goto l64
				}
				if !p.rules[ruleSetextBottom2]() {
					goto l64
				}
				position = position65
			}
			if !p.rules[ruleStartList]() {
				goto l64
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l68
			}
			goto l64
		l68:
			if !p.rules[ruleInline]() {
				goto l64
			}
			do(10)
		l66:
			{
				position67 := position
				if !p.rules[ruleEndline]() {
					goto l69
				}
				goto l67
			l69:
				if !p.rules[ruleInline]() {
					goto l67
				}
				do(10)
				goto l66
			l67:
				position = position67
			}
			if !p.rules[ruleNewline]() {
				goto l64
			}
			if !p.rules[ruleSetextBottom2]() {
				goto l64
			}
			do(11)
			doarg(yyPop, 1)
			return true
		l64:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 13 Heading <- (AtxHeading / SetextHeading) */
		func() bool {
			if !p.rules[ruleAtxHeading]() {
				goto l72
			}
			goto l71
		l72:
			if !p.rules[ruleSetextHeading]() {
				goto l70
			}
		l71:
			return true
		l70:
			return false
		},
		/* 14 BlockQuote <- (BlockQuoteRaw {  yy = mk_element(BLOCKQUOTE)
                yy.children = a
             }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleBlockQuoteRaw]() {
				goto l73
			}
			doarg(yySet, -1)
			do(12)
			doarg(yyPop, 1)
			return true
		l73:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 BlockQuoteRaw <- (StartList ('>' ' '? Line { a = cons(yy, a) } (!'>' !BlankLine Line { a = cons(yy, a) })* (BlankLine { a = cons(mk_str("\n"), a) })*)+ {   yy = mk_str_from_list(a, true)
                     yy.key = RAW
                 }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l74
			}
			doarg(yySet, -1)
			if !matchChar('>') {
				goto l74
			}
			matchChar(' ')
			if !p.rules[ruleLine]() {
				goto l74
			}
			do(13)
		l77:
			{
				position78, thunkPosition78 := position, thunkPosition
				if peekChar('>') {
					goto l78
				}
				if !p.rules[ruleBlankLine]() {
					goto l79
				}
				goto l78
			l79:
				if !p.rules[ruleLine]() {
					goto l78
				}
				do(14)
				goto l77
			l78:
				position, thunkPosition = position78, thunkPosition78
			}
		l80:
			{
				position81 := position
				if !p.rules[ruleBlankLine]() {
					goto l81
				}
				do(15)
				goto l80
			l81:
				position = position81
			}
		l75:
			{
				position76, thunkPosition76 := position, thunkPosition
				if !matchChar('>') {
					goto l76
				}
				matchChar(' ')
				if !p.rules[ruleLine]() {
					goto l76
				}
				do(13)
			l82:
				{
					position83, thunkPosition83 := position, thunkPosition
					if peekChar('>') {
						goto l83
					}
					if !p.rules[ruleBlankLine]() {
						goto l84
					}
					goto l83
				l84:
					if !p.rules[ruleLine]() {
						goto l83
					}
					do(14)
					goto l82
				l83:
					position, thunkPosition = position83, thunkPosition83
				}
			l85:
				{
					position86 := position
					if !p.rules[ruleBlankLine]() {
						goto l86
					}
					do(15)
					goto l85
				l86:
					position = position86
				}
				goto l75
			l76:
				position, thunkPosition = position76, thunkPosition76
			}
			do(16)
			doarg(yyPop, 1)
			return true
		l74:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 NonblankIndentedLine <- (!BlankLine IndentedLine) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l88
			}
			goto l87
		l88:
			if !p.rules[ruleIndentedLine]() {
				goto l87
			}
			return true
		l87:
			position = position0
			return false
		},
		/* 17 VerbatimChunk <- (StartList (BlankLine { a = cons(mk_str("\n"), a) })* (NonblankIndentedLine { a = cons(yy, a) })+ { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l89
			}
			doarg(yySet, -1)
		l90:
			{
				position91 := position
				if !p.rules[ruleBlankLine]() {
					goto l91
				}
				do(17)
				goto l90
			l91:
				position = position91
			}
			if !p.rules[ruleNonblankIndentedLine]() {
				goto l89
			}
			do(18)
		l92:
			{
				position93 := position
				if !p.rules[ruleNonblankIndentedLine]() {
					goto l93
				}
				do(18)
				goto l92
			l93:
				position = position93
			}
			do(19)
			doarg(yyPop, 1)
			return true
		l89:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 18 Verbatim <- (StartList (VerbatimChunk { a = cons(yy, a) })+ { yy = mk_str_from_list(a, false)
                 yy.key = VERBATIM }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l94
			}
			doarg(yySet, -1)
			if !p.rules[ruleVerbatimChunk]() {
				goto l94
			}
			do(20)
		l95:
			{
				position96, thunkPosition96 := position, thunkPosition
				if !p.rules[ruleVerbatimChunk]() {
					goto l96
				}
				do(20)
				goto l95
			l96:
				position, thunkPosition = position96, thunkPosition96
			}
			do(21)
			doarg(yyPop, 1)
			return true
		l94:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 19 HorizontalRule <- (NonindentSpace ((&[_] ('_' Sp '_' Sp '_' (Sp '_')*)) | (&[\-] ('-' Sp '-' Sp '-' (Sp '-')*)) | (&[*] ('*' Sp '*' Sp '*' (Sp '*')*))) Sp Newline BlankLine+ { yy = mk_element(HRULE) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l97
			}
			{
				if position == len(p.Buffer) {
					goto l97
				}
				switch p.Buffer[position] {
				case '_':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l97
					}
					if !matchChar('_') {
						goto l97
					}
					if !p.rules[ruleSp]() {
						goto l97
					}
					if !matchChar('_') {
						goto l97
					}
				l99:
					{
						position100 := position
						if !p.rules[ruleSp]() {
							goto l100
						}
						if !matchChar('_') {
							goto l100
						}
						goto l99
					l100:
						position = position100
					}
					break
				case '-':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l97
					}
					if !matchChar('-') {
						goto l97
					}
					if !p.rules[ruleSp]() {
						goto l97
					}
					if !matchChar('-') {
						goto l97
					}
				l101:
					{
						position102 := position
						if !p.rules[ruleSp]() {
							goto l102
						}
						if !matchChar('-') {
							goto l102
						}
						goto l101
					l102:
						position = position102
					}
					break
				case '*':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l97
					}
					if !matchChar('*') {
						goto l97
					}
					if !p.rules[ruleSp]() {
						goto l97
					}
					if !matchChar('*') {
						goto l97
					}
				l103:
					{
						position104 := position
						if !p.rules[ruleSp]() {
							goto l104
						}
						if !matchChar('*') {
							goto l104
						}
						goto l103
					l104:
						position = position104
					}
					break
				default:
					goto l97
				}
			}
			if !p.rules[ruleSp]() {
				goto l97
			}
			if !p.rules[ruleNewline]() {
				goto l97
			}
			if !p.rules[ruleBlankLine]() {
				goto l97
			}
		l105:
			if !p.rules[ruleBlankLine]() {
				goto l106
			}
			goto l105
		l106:
			do(22)
			return true
		l97:
			position = position0
			return false
		},
		/* 20 Bullet <- (!HorizontalRule NonindentSpace ((&[\-] '-') | (&[*] '*') | (&[+] '+')) Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHorizontalRule]() {
				goto l108
			}
			goto l107
		l108:
			if !p.rules[ruleNonindentSpace]() {
				goto l107
			}
			{
				if position == len(p.Buffer) {
					goto l107
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
					goto l107
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l107
			}
		l110:
			if !p.rules[ruleSpacechar]() {
				goto l111
			}
			goto l110
		l111:
			return true
		l107:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 21 BulletList <- (&Bullet (ListTight / ListLoose) { yy.key = BULLETLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position113 := position
				if !p.rules[ruleBullet]() {
					goto l112
				}
				position = position113
			}
			if !p.rules[ruleListTight]() {
				goto l115
			}
			goto l114
		l115:
			if !p.rules[ruleListLoose]() {
				goto l112
			}
		l114:
			do(23)
			return true
		l112:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 22 ListTight <- (StartList (ListItemTight { a = cons(yy, a) })+ BlankLine* !((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l116
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItemTight]() {
				goto l116
			}
			do(24)
		l117:
			{
				position118, thunkPosition118 := position, thunkPosition
				if !p.rules[ruleListItemTight]() {
					goto l118
				}
				do(24)
				goto l117
			l118:
				position, thunkPosition = position118, thunkPosition118
			}
		l119:
			if !p.rules[ruleBlankLine]() {
				goto l120
			}
			goto l119
		l120:
			{
				if position == len(p.Buffer) {
					goto l121
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l121
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l121
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l121
					}
				}
			}
			goto l116
		l121:
			do(25)
			doarg(yyPop, 1)
			return true
		l116:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 23 ListLoose <- (StartList (ListItem BlankLine* {
                  li := b.children
                  li.contents.str += "\n\n"
                  a = cons(b, a)
              })+ { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l123
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItem]() {
				goto l123
			}
			doarg(yySet, -2)
		l126:
			if !p.rules[ruleBlankLine]() {
				goto l127
			}
			goto l126
		l127:
			do(26)
		l124:
			{
				position125, thunkPosition125 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto l125
				}
				doarg(yySet, -2)
			l128:
				if !p.rules[ruleBlankLine]() {
					goto l129
				}
				goto l128
			l129:
				do(26)
				goto l124
			l125:
				position, thunkPosition = position125, thunkPosition125
			}
			do(27)
			doarg(yyPop, 2)
			return true
		l123:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 24 ListItem <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (ListContinuationBlock { a = cons(yy, a) })* {
               raw := mk_str_from_list(a, false)
               raw.key = RAW
               yy = mk_element(LISTITEM)
               yy.children = raw
            }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				if position == len(p.Buffer) {
					goto l130
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l130
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l130
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l130
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto l130
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l130
			}
			do(28)
		l132:
			{
				position133, thunkPosition133 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l133
				}
				do(29)
				goto l132
			l133:
				position, thunkPosition = position133, thunkPosition133
			}
			do(30)
			doarg(yyPop, 1)
			return true
		l130:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 25 ListItemTight <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (!BlankLine ListContinuationBlock { a = cons(yy, a) })* !ListContinuationBlock {
               raw := mk_str_from_list(a, false)
               raw.key = RAW
               yy = mk_element(LISTITEM)
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
			do(31)
		l136:
			{
				position137, thunkPosition137 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l138
				}
				goto l137
			l138:
				if !p.rules[ruleListContinuationBlock]() {
					goto l137
				}
				do(32)
				goto l136
			l137:
				position, thunkPosition = position137, thunkPosition137
			}
			if !p.rules[ruleListContinuationBlock]() {
				goto l139
			}
			goto l134
		l139:
			do(33)
			doarg(yyPop, 1)
			return true
		l134:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 26 ListBlock <- (StartList !BlankLine Line { a = cons(yy, a) } (ListBlockLine { a = cons(yy, a) })* { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l140
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l141
			}
			goto l140
		l141:
			if !p.rules[ruleLine]() {
				goto l140
			}
			do(34)
		l142:
			{
				position143 := position
				if !p.rules[ruleListBlockLine]() {
					goto l143
				}
				do(35)
				goto l142
			l143:
				position = position143
			}
			do(36)
			doarg(yyPop, 1)
			return true
		l140:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 27 ListContinuationBlock <- (StartList (< BlankLine* > {   if len(yytext) == 0 {
                                   a = cons(mk_str("\001"), a) // block separator
                              } else {
                                   a = cons(mk_str(yytext), a)
                              }
                          }) (Indent ListBlock { a = cons(yy, a) })+ {  yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l144
			}
			doarg(yySet, -1)
			begin = position
		l145:
			if !p.rules[ruleBlankLine]() {
				goto l146
			}
			goto l145
		l146:
			end = position
			do(37)
			if !p.rules[ruleIndent]() {
				goto l144
			}
			if !p.rules[ruleListBlock]() {
				goto l144
			}
			do(38)
		l147:
			{
				position148, thunkPosition148 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l148
				}
				if !p.rules[ruleListBlock]() {
					goto l148
				}
				do(38)
				goto l147
			l148:
				position, thunkPosition = position148, thunkPosition148
			}
			do(39)
			doarg(yyPop, 1)
			return true
		l144:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 28 Enumerator <- (NonindentSpace [0-9]+ '.' Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l149
			}
			if !matchClass(0) {
				goto l149
			}
		l150:
			if !matchClass(0) {
				goto l151
			}
			goto l150
		l151:
			if !matchChar('.') {
				goto l149
			}
			if !p.rules[ruleSpacechar]() {
				goto l149
			}
		l152:
			if !p.rules[ruleSpacechar]() {
				goto l153
			}
			goto l152
		l153:
			return true
		l149:
			position = position0
			return false
		},
		/* 29 OrderedList <- (&Enumerator (ListTight / ListLoose) { yy.key = ORDEREDLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position155 := position
				if !p.rules[ruleEnumerator]() {
					goto l154
				}
				position = position155
			}
			if !p.rules[ruleListTight]() {
				goto l157
			}
			goto l156
		l157:
			if !p.rules[ruleListLoose]() {
				goto l154
			}
		l156:
			do(40)
			return true
		l154:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 ListBlockLine <- (!BlankLine !((&[:~] DefMarker) | (&[\t *+\-0-9] (Indent? ((&[*+\-] Bullet) | (&[0-9] Enumerator))))) !HorizontalRule OptionallyIndentedLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlankLine]() {
				goto l159
			}
			goto l158
		l159:
			{
				position160 := position
				{
					if position == len(p.Buffer) {
						goto l160
					}
					switch p.Buffer[position] {
					case ':', '~':
						if !p.rules[ruleDefMarker]() {
							goto l160
						}
						break
					default:
						if !p.rules[ruleIndent]() {
							goto l162
						}
					l162:
						{
							if position == len(p.Buffer) {
								goto l160
							}
							switch p.Buffer[position] {
							case '*', '+', '-':
								if !p.rules[ruleBullet]() {
									goto l160
								}
								break
							default:
								if !p.rules[ruleEnumerator]() {
									goto l160
								}
							}
						}
					}
				}
				goto l158
			l160:
				position = position160
			}
			if !p.rules[ruleHorizontalRule]() {
				goto l165
			}
			goto l158
		l165:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l158
			}
			return true
		l158:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 31 HtmlBlockOpenAddress <- ('<' Spnl ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l166
			}
			if !p.rules[ruleSpnl]() {
				goto l166
			}
			{
				if position == len(p.Buffer) {
					goto l166
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l166
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l166
					}
					break
				default:
					goto l166
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l166
			}
		l168:
			if !p.rules[ruleHtmlAttribute]() {
				goto l169
			}
			goto l168
		l169:
			if !matchChar('>') {
				goto l166
			}
			return true
		l166:
			position = position0
			return false
		},
		/* 32 HtmlBlockCloseAddress <- ('<' Spnl '/' ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l170
			}
			if !p.rules[ruleSpnl]() {
				goto l170
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l170
			}
			return true
		l170:
			position = position0
			return false
		},
		/* 33 HtmlBlockAddress <- (HtmlBlockOpenAddress (HtmlBlockAddress / (!HtmlBlockCloseAddress .))* HtmlBlockCloseAddress) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenAddress]() {
				goto l172
			}
		l173:
			{
				position174 := position
				if !p.rules[ruleHtmlBlockAddress]() {
					goto l176
				}
				goto l175
			l176:
				if !p.rules[ruleHtmlBlockCloseAddress]() {
					goto l177
				}
				goto l174
			l177:
				if !matchDot() {
					goto l174
				}
			l175:
				goto l173
			l174:
				position = position174
			}
			if !p.rules[ruleHtmlBlockCloseAddress]() {
				goto l172
			}
			return true
		l172:
			position = position0
			return false
		},
		/* 34 HtmlBlockOpenBlockquote <- ('<' Spnl ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l178
			}
			if !p.rules[ruleSpnl]() {
				goto l178
			}
			{
				if position == len(p.Buffer) {
					goto l178
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l178
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l178
					}
					break
				default:
					goto l178
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l178
			}
		l180:
			if !p.rules[ruleHtmlAttribute]() {
				goto l181
			}
			goto l180
		l181:
			if !matchChar('>') {
				goto l178
			}
			return true
		l178:
			position = position0
			return false
		},
		/* 35 HtmlBlockCloseBlockquote <- ('<' Spnl '/' ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l182
			}
			if !p.rules[ruleSpnl]() {
				goto l182
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l182
			}
			return true
		l182:
			position = position0
			return false
		},
		/* 36 HtmlBlockBlockquote <- (HtmlBlockOpenBlockquote (HtmlBlockBlockquote / (!HtmlBlockCloseBlockquote .))* HtmlBlockCloseBlockquote) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenBlockquote]() {
				goto l184
			}
		l185:
			{
				position186 := position
				if !p.rules[ruleHtmlBlockBlockquote]() {
					goto l188
				}
				goto l187
			l188:
				if !p.rules[ruleHtmlBlockCloseBlockquote]() {
					goto l189
				}
				goto l186
			l189:
				if !matchDot() {
					goto l186
				}
			l187:
				goto l185
			l186:
				position = position186
			}
			if !p.rules[ruleHtmlBlockCloseBlockquote]() {
				goto l184
			}
			return true
		l184:
			position = position0
			return false
		},
		/* 37 HtmlBlockOpenCenter <- ('<' Spnl ((&[C] 'CENTER') | (&[c] 'center')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l190
			}
			if !p.rules[ruleSpnl]() {
				goto l190
			}
			{
				if position == len(p.Buffer) {
					goto l190
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l190
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l190
					}
					break
				default:
					goto l190
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l190
			}
		l192:
			if !p.rules[ruleHtmlAttribute]() {
				goto l193
			}
			goto l192
		l193:
			if !matchChar('>') {
				goto l190
			}
			return true
		l190:
			position = position0
			return false
		},
		/* 38 HtmlBlockCloseCenter <- ('<' Spnl '/' ((&[C] 'CENTER') | (&[c] 'center')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l194
			}
			if !p.rules[ruleSpnl]() {
				goto l194
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l194
			}
			return true
		l194:
			position = position0
			return false
		},
		/* 39 HtmlBlockCenter <- (HtmlBlockOpenCenter (HtmlBlockCenter / (!HtmlBlockCloseCenter .))* HtmlBlockCloseCenter) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenCenter]() {
				goto l196
			}
		l197:
			{
				position198 := position
				if !p.rules[ruleHtmlBlockCenter]() {
					goto l200
				}
				goto l199
			l200:
				if !p.rules[ruleHtmlBlockCloseCenter]() {
					goto l201
				}
				goto l198
			l201:
				if !matchDot() {
					goto l198
				}
			l199:
				goto l197
			l198:
				position = position198
			}
			if !p.rules[ruleHtmlBlockCloseCenter]() {
				goto l196
			}
			return true
		l196:
			position = position0
			return false
		},
		/* 40 HtmlBlockOpenDir <- ('<' Spnl ((&[D] 'DIR') | (&[d] 'dir')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l202
			}
			if !p.rules[ruleSpnl]() {
				goto l202
			}
			{
				if position == len(p.Buffer) {
					goto l202
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto l202
					}
					break
				case 'd':
					position++
					if !matchString("ir") {
						goto l202
					}
					break
				default:
					goto l202
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l202
			}
		l204:
			if !p.rules[ruleHtmlAttribute]() {
				goto l205
			}
			goto l204
		l205:
			if !matchChar('>') {
				goto l202
			}
			return true
		l202:
			position = position0
			return false
		},
		/* 41 HtmlBlockCloseDir <- ('<' Spnl '/' ((&[D] 'DIR') | (&[d] 'dir')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l206
			}
			if !p.rules[ruleSpnl]() {
				goto l206
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l206
			}
			return true
		l206:
			position = position0
			return false
		},
		/* 42 HtmlBlockDir <- (HtmlBlockOpenDir (HtmlBlockDir / (!HtmlBlockCloseDir .))* HtmlBlockCloseDir) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDir]() {
				goto l208
			}
		l209:
			{
				position210 := position
				if !p.rules[ruleHtmlBlockDir]() {
					goto l212
				}
				goto l211
			l212:
				if !p.rules[ruleHtmlBlockCloseDir]() {
					goto l213
				}
				goto l210
			l213:
				if !matchDot() {
					goto l210
				}
			l211:
				goto l209
			l210:
				position = position210
			}
			if !p.rules[ruleHtmlBlockCloseDir]() {
				goto l208
			}
			return true
		l208:
			position = position0
			return false
		},
		/* 43 HtmlBlockOpenDiv <- ('<' Spnl ((&[D] 'DIV') | (&[d] 'div')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l214
			}
			if !p.rules[ruleSpnl]() {
				goto l214
			}
			{
				if position == len(p.Buffer) {
					goto l214
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto l214
					}
					break
				case 'd':
					position++
					if !matchString("iv") {
						goto l214
					}
					break
				default:
					goto l214
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l214
			}
		l216:
			if !p.rules[ruleHtmlAttribute]() {
				goto l217
			}
			goto l216
		l217:
			if !matchChar('>') {
				goto l214
			}
			return true
		l214:
			position = position0
			return false
		},
		/* 44 HtmlBlockCloseDiv <- ('<' Spnl '/' ((&[D] 'DIV') | (&[d] 'div')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l218
			}
			if !p.rules[ruleSpnl]() {
				goto l218
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l218
			}
			return true
		l218:
			position = position0
			return false
		},
		/* 45 HtmlBlockDiv <- (HtmlBlockOpenDiv (HtmlBlockDiv / (!HtmlBlockCloseDiv .))* HtmlBlockCloseDiv) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDiv]() {
				goto l220
			}
		l221:
			{
				position222 := position
				if !p.rules[ruleHtmlBlockDiv]() {
					goto l224
				}
				goto l223
			l224:
				if !p.rules[ruleHtmlBlockCloseDiv]() {
					goto l225
				}
				goto l222
			l225:
				if !matchDot() {
					goto l222
				}
			l223:
				goto l221
			l222:
				position = position222
			}
			if !p.rules[ruleHtmlBlockCloseDiv]() {
				goto l220
			}
			return true
		l220:
			position = position0
			return false
		},
		/* 46 HtmlBlockOpenDl <- ('<' Spnl ((&[D] 'DL') | (&[d] 'dl')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l226
			}
			if !p.rules[ruleSpnl]() {
				goto l226
			}
			{
				if position == len(p.Buffer) {
					goto l226
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto l226
					}
					break
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto l226
					}
					break
				default:
					goto l226
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l226
			}
		l228:
			if !p.rules[ruleHtmlAttribute]() {
				goto l229
			}
			goto l228
		l229:
			if !matchChar('>') {
				goto l226
			}
			return true
		l226:
			position = position0
			return false
		},
		/* 47 HtmlBlockCloseDl <- ('<' Spnl '/' ((&[D] 'DL') | (&[d] 'dl')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l230
			}
			if !p.rules[ruleSpnl]() {
				goto l230
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l230
			}
			return true
		l230:
			position = position0
			return false
		},
		/* 48 HtmlBlockDl <- (HtmlBlockOpenDl (HtmlBlockDl / (!HtmlBlockCloseDl .))* HtmlBlockCloseDl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDl]() {
				goto l232
			}
		l233:
			{
				position234 := position
				if !p.rules[ruleHtmlBlockDl]() {
					goto l236
				}
				goto l235
			l236:
				if !p.rules[ruleHtmlBlockCloseDl]() {
					goto l237
				}
				goto l234
			l237:
				if !matchDot() {
					goto l234
				}
			l235:
				goto l233
			l234:
				position = position234
			}
			if !p.rules[ruleHtmlBlockCloseDl]() {
				goto l232
			}
			return true
		l232:
			position = position0
			return false
		},
		/* 49 HtmlBlockOpenFieldset <- ('<' Spnl ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l238
			}
			if !p.rules[ruleSpnl]() {
				goto l238
			}
			{
				if position == len(p.Buffer) {
					goto l238
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto l238
					}
					break
				case 'f':
					position++
					if !matchString("ieldset") {
						goto l238
					}
					break
				default:
					goto l238
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l238
			}
		l240:
			if !p.rules[ruleHtmlAttribute]() {
				goto l241
			}
			goto l240
		l241:
			if !matchChar('>') {
				goto l238
			}
			return true
		l238:
			position = position0
			return false
		},
		/* 50 HtmlBlockCloseFieldset <- ('<' Spnl '/' ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l242
			}
			if !p.rules[ruleSpnl]() {
				goto l242
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l242
			}
			return true
		l242:
			position = position0
			return false
		},
		/* 51 HtmlBlockFieldset <- (HtmlBlockOpenFieldset (HtmlBlockFieldset / (!HtmlBlockCloseFieldset .))* HtmlBlockCloseFieldset) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFieldset]() {
				goto l244
			}
		l245:
			{
				position246 := position
				if !p.rules[ruleHtmlBlockFieldset]() {
					goto l248
				}
				goto l247
			l248:
				if !p.rules[ruleHtmlBlockCloseFieldset]() {
					goto l249
				}
				goto l246
			l249:
				if !matchDot() {
					goto l246
				}
			l247:
				goto l245
			l246:
				position = position246
			}
			if !p.rules[ruleHtmlBlockCloseFieldset]() {
				goto l244
			}
			return true
		l244:
			position = position0
			return false
		},
		/* 52 HtmlBlockOpenForm <- ('<' Spnl ((&[F] 'FORM') | (&[f] 'form')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l250
			}
			if !p.rules[ruleSpnl]() {
				goto l250
			}
			{
				if position == len(p.Buffer) {
					goto l250
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto l250
					}
					break
				case 'f':
					position++
					if !matchString("orm") {
						goto l250
					}
					break
				default:
					goto l250
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l250
			}
		l252:
			if !p.rules[ruleHtmlAttribute]() {
				goto l253
			}
			goto l252
		l253:
			if !matchChar('>') {
				goto l250
			}
			return true
		l250:
			position = position0
			return false
		},
		/* 53 HtmlBlockCloseForm <- ('<' Spnl '/' ((&[F] 'FORM') | (&[f] 'form')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l254
			}
			if !p.rules[ruleSpnl]() {
				goto l254
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l254
			}
			return true
		l254:
			position = position0
			return false
		},
		/* 54 HtmlBlockForm <- (HtmlBlockOpenForm (HtmlBlockForm / (!HtmlBlockCloseForm .))* HtmlBlockCloseForm) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenForm]() {
				goto l256
			}
		l257:
			{
				position258 := position
				if !p.rules[ruleHtmlBlockForm]() {
					goto l260
				}
				goto l259
			l260:
				if !p.rules[ruleHtmlBlockCloseForm]() {
					goto l261
				}
				goto l258
			l261:
				if !matchDot() {
					goto l258
				}
			l259:
				goto l257
			l258:
				position = position258
			}
			if !p.rules[ruleHtmlBlockCloseForm]() {
				goto l256
			}
			return true
		l256:
			position = position0
			return false
		},
		/* 55 HtmlBlockOpenH1 <- ('<' Spnl ((&[H] 'H1') | (&[h] 'h1')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l262
			}
			if !p.rules[ruleSpnl]() {
				goto l262
			}
			{
				if position == len(p.Buffer) {
					goto l262
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto l262
					}
					break
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto l262
					}
					break
				default:
					goto l262
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l262
			}
		l264:
			if !p.rules[ruleHtmlAttribute]() {
				goto l265
			}
			goto l264
		l265:
			if !matchChar('>') {
				goto l262
			}
			return true
		l262:
			position = position0
			return false
		},
		/* 56 HtmlBlockCloseH1 <- ('<' Spnl '/' ((&[H] 'H1') | (&[h] 'h1')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l266
			}
			if !p.rules[ruleSpnl]() {
				goto l266
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l266
			}
			return true
		l266:
			position = position0
			return false
		},
		/* 57 HtmlBlockH1 <- (HtmlBlockOpenH1 (HtmlBlockH1 / (!HtmlBlockCloseH1 .))* HtmlBlockCloseH1) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH1]() {
				goto l268
			}
		l269:
			{
				position270 := position
				if !p.rules[ruleHtmlBlockH1]() {
					goto l272
				}
				goto l271
			l272:
				if !p.rules[ruleHtmlBlockCloseH1]() {
					goto l273
				}
				goto l270
			l273:
				if !matchDot() {
					goto l270
				}
			l271:
				goto l269
			l270:
				position = position270
			}
			if !p.rules[ruleHtmlBlockCloseH1]() {
				goto l268
			}
			return true
		l268:
			position = position0
			return false
		},
		/* 58 HtmlBlockOpenH2 <- ('<' Spnl ((&[H] 'H2') | (&[h] 'h2')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l274
			}
			if !p.rules[ruleSpnl]() {
				goto l274
			}
			{
				if position == len(p.Buffer) {
					goto l274
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto l274
					}
					break
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto l274
					}
					break
				default:
					goto l274
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l274
			}
		l276:
			if !p.rules[ruleHtmlAttribute]() {
				goto l277
			}
			goto l276
		l277:
			if !matchChar('>') {
				goto l274
			}
			return true
		l274:
			position = position0
			return false
		},
		/* 59 HtmlBlockCloseH2 <- ('<' Spnl '/' ((&[H] 'H2') | (&[h] 'h2')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l278
			}
			if !p.rules[ruleSpnl]() {
				goto l278
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l278
			}
			return true
		l278:
			position = position0
			return false
		},
		/* 60 HtmlBlockH2 <- (HtmlBlockOpenH2 (HtmlBlockH2 / (!HtmlBlockCloseH2 .))* HtmlBlockCloseH2) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH2]() {
				goto l280
			}
		l281:
			{
				position282 := position
				if !p.rules[ruleHtmlBlockH2]() {
					goto l284
				}
				goto l283
			l284:
				if !p.rules[ruleHtmlBlockCloseH2]() {
					goto l285
				}
				goto l282
			l285:
				if !matchDot() {
					goto l282
				}
			l283:
				goto l281
			l282:
				position = position282
			}
			if !p.rules[ruleHtmlBlockCloseH2]() {
				goto l280
			}
			return true
		l280:
			position = position0
			return false
		},
		/* 61 HtmlBlockOpenH3 <- ('<' Spnl ((&[H] 'H3') | (&[h] 'h3')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l286
			}
			if !p.rules[ruleSpnl]() {
				goto l286
			}
			{
				if position == len(p.Buffer) {
					goto l286
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto l286
					}
					break
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto l286
					}
					break
				default:
					goto l286
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l286
			}
		l288:
			if !p.rules[ruleHtmlAttribute]() {
				goto l289
			}
			goto l288
		l289:
			if !matchChar('>') {
				goto l286
			}
			return true
		l286:
			position = position0
			return false
		},
		/* 62 HtmlBlockCloseH3 <- ('<' Spnl '/' ((&[H] 'H3') | (&[h] 'h3')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l290
			}
			if !p.rules[ruleSpnl]() {
				goto l290
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l290
			}
			return true
		l290:
			position = position0
			return false
		},
		/* 63 HtmlBlockH3 <- (HtmlBlockOpenH3 (HtmlBlockH3 / (!HtmlBlockCloseH3 .))* HtmlBlockCloseH3) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH3]() {
				goto l292
			}
		l293:
			{
				position294 := position
				if !p.rules[ruleHtmlBlockH3]() {
					goto l296
				}
				goto l295
			l296:
				if !p.rules[ruleHtmlBlockCloseH3]() {
					goto l297
				}
				goto l294
			l297:
				if !matchDot() {
					goto l294
				}
			l295:
				goto l293
			l294:
				position = position294
			}
			if !p.rules[ruleHtmlBlockCloseH3]() {
				goto l292
			}
			return true
		l292:
			position = position0
			return false
		},
		/* 64 HtmlBlockOpenH4 <- ('<' Spnl ((&[H] 'H4') | (&[h] 'h4')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l298
			}
			if !p.rules[ruleSpnl]() {
				goto l298
			}
			{
				if position == len(p.Buffer) {
					goto l298
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto l298
					}
					break
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto l298
					}
					break
				default:
					goto l298
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l298
			}
		l300:
			if !p.rules[ruleHtmlAttribute]() {
				goto l301
			}
			goto l300
		l301:
			if !matchChar('>') {
				goto l298
			}
			return true
		l298:
			position = position0
			return false
		},
		/* 65 HtmlBlockCloseH4 <- ('<' Spnl '/' ((&[H] 'H4') | (&[h] 'h4')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l302
			}
			if !p.rules[ruleSpnl]() {
				goto l302
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l302
			}
			return true
		l302:
			position = position0
			return false
		},
		/* 66 HtmlBlockH4 <- (HtmlBlockOpenH4 (HtmlBlockH4 / (!HtmlBlockCloseH4 .))* HtmlBlockCloseH4) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH4]() {
				goto l304
			}
		l305:
			{
				position306 := position
				if !p.rules[ruleHtmlBlockH4]() {
					goto l308
				}
				goto l307
			l308:
				if !p.rules[ruleHtmlBlockCloseH4]() {
					goto l309
				}
				goto l306
			l309:
				if !matchDot() {
					goto l306
				}
			l307:
				goto l305
			l306:
				position = position306
			}
			if !p.rules[ruleHtmlBlockCloseH4]() {
				goto l304
			}
			return true
		l304:
			position = position0
			return false
		},
		/* 67 HtmlBlockOpenH5 <- ('<' Spnl ((&[H] 'H5') | (&[h] 'h5')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l310
			}
			if !p.rules[ruleSpnl]() {
				goto l310
			}
			{
				if position == len(p.Buffer) {
					goto l310
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto l310
					}
					break
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto l310
					}
					break
				default:
					goto l310
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l310
			}
		l312:
			if !p.rules[ruleHtmlAttribute]() {
				goto l313
			}
			goto l312
		l313:
			if !matchChar('>') {
				goto l310
			}
			return true
		l310:
			position = position0
			return false
		},
		/* 68 HtmlBlockCloseH5 <- ('<' Spnl '/' ((&[H] 'H5') | (&[h] 'h5')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l314
			}
			if !p.rules[ruleSpnl]() {
				goto l314
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l314
			}
			return true
		l314:
			position = position0
			return false
		},
		/* 69 HtmlBlockH5 <- (HtmlBlockOpenH5 (HtmlBlockH5 / (!HtmlBlockCloseH5 .))* HtmlBlockCloseH5) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH5]() {
				goto l316
			}
		l317:
			{
				position318 := position
				if !p.rules[ruleHtmlBlockH5]() {
					goto l320
				}
				goto l319
			l320:
				if !p.rules[ruleHtmlBlockCloseH5]() {
					goto l321
				}
				goto l318
			l321:
				if !matchDot() {
					goto l318
				}
			l319:
				goto l317
			l318:
				position = position318
			}
			if !p.rules[ruleHtmlBlockCloseH5]() {
				goto l316
			}
			return true
		l316:
			position = position0
			return false
		},
		/* 70 HtmlBlockOpenH6 <- ('<' Spnl ((&[H] 'H6') | (&[h] 'h6')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l322
			}
			if !p.rules[ruleSpnl]() {
				goto l322
			}
			{
				if position == len(p.Buffer) {
					goto l322
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto l322
					}
					break
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto l322
					}
					break
				default:
					goto l322
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l322
			}
		l324:
			if !p.rules[ruleHtmlAttribute]() {
				goto l325
			}
			goto l324
		l325:
			if !matchChar('>') {
				goto l322
			}
			return true
		l322:
			position = position0
			return false
		},
		/* 71 HtmlBlockCloseH6 <- ('<' Spnl '/' ((&[H] 'H6') | (&[h] 'h6')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l326
			}
			if !p.rules[ruleSpnl]() {
				goto l326
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l326
			}
			return true
		l326:
			position = position0
			return false
		},
		/* 72 HtmlBlockH6 <- (HtmlBlockOpenH6 (HtmlBlockH6 / (!HtmlBlockCloseH6 .))* HtmlBlockCloseH6) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH6]() {
				goto l328
			}
		l329:
			{
				position330 := position
				if !p.rules[ruleHtmlBlockH6]() {
					goto l332
				}
				goto l331
			l332:
				if !p.rules[ruleHtmlBlockCloseH6]() {
					goto l333
				}
				goto l330
			l333:
				if !matchDot() {
					goto l330
				}
			l331:
				goto l329
			l330:
				position = position330
			}
			if !p.rules[ruleHtmlBlockCloseH6]() {
				goto l328
			}
			return true
		l328:
			position = position0
			return false
		},
		/* 73 HtmlBlockOpenMenu <- ('<' Spnl ((&[M] 'MENU') | (&[m] 'menu')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l334
			}
			if !p.rules[ruleSpnl]() {
				goto l334
			}
			{
				if position == len(p.Buffer) {
					goto l334
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto l334
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l334
					}
					break
				default:
					goto l334
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l334
			}
		l336:
			if !p.rules[ruleHtmlAttribute]() {
				goto l337
			}
			goto l336
		l337:
			if !matchChar('>') {
				goto l334
			}
			return true
		l334:
			position = position0
			return false
		},
		/* 74 HtmlBlockCloseMenu <- ('<' Spnl '/' ((&[M] 'MENU') | (&[m] 'menu')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l338
			}
			if !p.rules[ruleSpnl]() {
				goto l338
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l338
			}
			return true
		l338:
			position = position0
			return false
		},
		/* 75 HtmlBlockMenu <- (HtmlBlockOpenMenu (HtmlBlockMenu / (!HtmlBlockCloseMenu .))* HtmlBlockCloseMenu) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenMenu]() {
				goto l340
			}
		l341:
			{
				position342 := position
				if !p.rules[ruleHtmlBlockMenu]() {
					goto l344
				}
				goto l343
			l344:
				if !p.rules[ruleHtmlBlockCloseMenu]() {
					goto l345
				}
				goto l342
			l345:
				if !matchDot() {
					goto l342
				}
			l343:
				goto l341
			l342:
				position = position342
			}
			if !p.rules[ruleHtmlBlockCloseMenu]() {
				goto l340
			}
			return true
		l340:
			position = position0
			return false
		},
		/* 76 HtmlBlockOpenNoframes <- ('<' Spnl ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l346
			}
			if !p.rules[ruleSpnl]() {
				goto l346
			}
			{
				if position == len(p.Buffer) {
					goto l346
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto l346
					}
					break
				case 'n':
					position++
					if !matchString("oframes") {
						goto l346
					}
					break
				default:
					goto l346
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l346
			}
		l348:
			if !p.rules[ruleHtmlAttribute]() {
				goto l349
			}
			goto l348
		l349:
			if !matchChar('>') {
				goto l346
			}
			return true
		l346:
			position = position0
			return false
		},
		/* 77 HtmlBlockCloseNoframes <- ('<' Spnl '/' ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l350
			}
			if !p.rules[ruleSpnl]() {
				goto l350
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l350
			}
			return true
		l350:
			position = position0
			return false
		},
		/* 78 HtmlBlockNoframes <- (HtmlBlockOpenNoframes (HtmlBlockNoframes / (!HtmlBlockCloseNoframes .))* HtmlBlockCloseNoframes) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoframes]() {
				goto l352
			}
		l353:
			{
				position354 := position
				if !p.rules[ruleHtmlBlockNoframes]() {
					goto l356
				}
				goto l355
			l356:
				if !p.rules[ruleHtmlBlockCloseNoframes]() {
					goto l357
				}
				goto l354
			l357:
				if !matchDot() {
					goto l354
				}
			l355:
				goto l353
			l354:
				position = position354
			}
			if !p.rules[ruleHtmlBlockCloseNoframes]() {
				goto l352
			}
			return true
		l352:
			position = position0
			return false
		},
		/* 79 HtmlBlockOpenNoscript <- ('<' Spnl ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l358
			}
			if !p.rules[ruleSpnl]() {
				goto l358
			}
			{
				if position == len(p.Buffer) {
					goto l358
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l358
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l358
					}
					break
				default:
					goto l358
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l358
			}
		l360:
			if !p.rules[ruleHtmlAttribute]() {
				goto l361
			}
			goto l360
		l361:
			if !matchChar('>') {
				goto l358
			}
			return true
		l358:
			position = position0
			return false
		},
		/* 80 HtmlBlockCloseNoscript <- ('<' Spnl '/' ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l362
			}
			if !p.rules[ruleSpnl]() {
				goto l362
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l362
			}
			return true
		l362:
			position = position0
			return false
		},
		/* 81 HtmlBlockNoscript <- (HtmlBlockOpenNoscript (HtmlBlockNoscript / (!HtmlBlockCloseNoscript .))* HtmlBlockCloseNoscript) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoscript]() {
				goto l364
			}
		l365:
			{
				position366 := position
				if !p.rules[ruleHtmlBlockNoscript]() {
					goto l368
				}
				goto l367
			l368:
				if !p.rules[ruleHtmlBlockCloseNoscript]() {
					goto l369
				}
				goto l366
			l369:
				if !matchDot() {
					goto l366
				}
			l367:
				goto l365
			l366:
				position = position366
			}
			if !p.rules[ruleHtmlBlockCloseNoscript]() {
				goto l364
			}
			return true
		l364:
			position = position0
			return false
		},
		/* 82 HtmlBlockOpenOl <- ('<' Spnl ((&[O] 'OL') | (&[o] 'ol')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l370
			}
			if !p.rules[ruleSpnl]() {
				goto l370
			}
			{
				if position == len(p.Buffer) {
					goto l370
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l370
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l370
					}
					break
				default:
					goto l370
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l370
			}
		l372:
			if !p.rules[ruleHtmlAttribute]() {
				goto l373
			}
			goto l372
		l373:
			if !matchChar('>') {
				goto l370
			}
			return true
		l370:
			position = position0
			return false
		},
		/* 83 HtmlBlockCloseOl <- ('<' Spnl '/' ((&[O] 'OL') | (&[o] 'ol')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l374
			}
			if !p.rules[ruleSpnl]() {
				goto l374
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l374
			}
			return true
		l374:
			position = position0
			return false
		},
		/* 84 HtmlBlockOl <- (HtmlBlockOpenOl (HtmlBlockOl / (!HtmlBlockCloseOl .))* HtmlBlockCloseOl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenOl]() {
				goto l376
			}
		l377:
			{
				position378 := position
				if !p.rules[ruleHtmlBlockOl]() {
					goto l380
				}
				goto l379
			l380:
				if !p.rules[ruleHtmlBlockCloseOl]() {
					goto l381
				}
				goto l378
			l381:
				if !matchDot() {
					goto l378
				}
			l379:
				goto l377
			l378:
				position = position378
			}
			if !p.rules[ruleHtmlBlockCloseOl]() {
				goto l376
			}
			return true
		l376:
			position = position0
			return false
		},
		/* 85 HtmlBlockOpenP <- ('<' Spnl ((&[P] 'P') | (&[p] 'p')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l382
			}
			if !p.rules[ruleSpnl]() {
				goto l382
			}
			{
				if position == len(p.Buffer) {
					goto l382
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
					break
				case 'p':
					position++ // matchChar
					break
				default:
					goto l382
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l382
			}
		l384:
			if !p.rules[ruleHtmlAttribute]() {
				goto l385
			}
			goto l384
		l385:
			if !matchChar('>') {
				goto l382
			}
			return true
		l382:
			position = position0
			return false
		},
		/* 86 HtmlBlockCloseP <- ('<' Spnl '/' ((&[P] 'P') | (&[p] 'p')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l386
			}
			if !p.rules[ruleSpnl]() {
				goto l386
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l386
			}
			return true
		l386:
			position = position0
			return false
		},
		/* 87 HtmlBlockP <- (HtmlBlockOpenP (HtmlBlockP / (!HtmlBlockCloseP .))* HtmlBlockCloseP) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenP]() {
				goto l388
			}
		l389:
			{
				position390 := position
				if !p.rules[ruleHtmlBlockP]() {
					goto l392
				}
				goto l391
			l392:
				if !p.rules[ruleHtmlBlockCloseP]() {
					goto l393
				}
				goto l390
			l393:
				if !matchDot() {
					goto l390
				}
			l391:
				goto l389
			l390:
				position = position390
			}
			if !p.rules[ruleHtmlBlockCloseP]() {
				goto l388
			}
			return true
		l388:
			position = position0
			return false
		},
		/* 88 HtmlBlockOpenPre <- ('<' Spnl ((&[P] 'PRE') | (&[p] 'pre')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l394
			}
			if !p.rules[ruleSpnl]() {
				goto l394
			}
			{
				if position == len(p.Buffer) {
					goto l394
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto l394
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l394
					}
					break
				default:
					goto l394
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l394
			}
		l396:
			if !p.rules[ruleHtmlAttribute]() {
				goto l397
			}
			goto l396
		l397:
			if !matchChar('>') {
				goto l394
			}
			return true
		l394:
			position = position0
			return false
		},
		/* 89 HtmlBlockClosePre <- ('<' Spnl '/' ((&[P] 'PRE') | (&[p] 'pre')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l398
			}
			if !p.rules[ruleSpnl]() {
				goto l398
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l398
			}
			return true
		l398:
			position = position0
			return false
		},
		/* 90 HtmlBlockPre <- (HtmlBlockOpenPre (HtmlBlockPre / (!HtmlBlockClosePre .))* HtmlBlockClosePre) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenPre]() {
				goto l400
			}
		l401:
			{
				position402 := position
				if !p.rules[ruleHtmlBlockPre]() {
					goto l404
				}
				goto l403
			l404:
				if !p.rules[ruleHtmlBlockClosePre]() {
					goto l405
				}
				goto l402
			l405:
				if !matchDot() {
					goto l402
				}
			l403:
				goto l401
			l402:
				position = position402
			}
			if !p.rules[ruleHtmlBlockClosePre]() {
				goto l400
			}
			return true
		l400:
			position = position0
			return false
		},
		/* 91 HtmlBlockOpenTable <- ('<' Spnl ((&[T] 'TABLE') | (&[t] 'table')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l406
			}
			if !p.rules[ruleSpnl]() {
				goto l406
			}
			{
				if position == len(p.Buffer) {
					goto l406
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto l406
					}
					break
				case 't':
					position++
					if !matchString("able") {
						goto l406
					}
					break
				default:
					goto l406
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l406
			}
		l408:
			if !p.rules[ruleHtmlAttribute]() {
				goto l409
			}
			goto l408
		l409:
			if !matchChar('>') {
				goto l406
			}
			return true
		l406:
			position = position0
			return false
		},
		/* 92 HtmlBlockCloseTable <- ('<' Spnl '/' ((&[T] 'TABLE') | (&[t] 'table')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l410
			}
			if !p.rules[ruleSpnl]() {
				goto l410
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l410
			}
			return true
		l410:
			position = position0
			return false
		},
		/* 93 HtmlBlockTable <- (HtmlBlockOpenTable (HtmlBlockTable / (!HtmlBlockCloseTable .))* HtmlBlockCloseTable) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTable]() {
				goto l412
			}
		l413:
			{
				position414 := position
				if !p.rules[ruleHtmlBlockTable]() {
					goto l416
				}
				goto l415
			l416:
				if !p.rules[ruleHtmlBlockCloseTable]() {
					goto l417
				}
				goto l414
			l417:
				if !matchDot() {
					goto l414
				}
			l415:
				goto l413
			l414:
				position = position414
			}
			if !p.rules[ruleHtmlBlockCloseTable]() {
				goto l412
			}
			return true
		l412:
			position = position0
			return false
		},
		/* 94 HtmlBlockOpenUl <- ('<' Spnl ((&[U] 'UL') | (&[u] 'ul')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l418
			}
			if !p.rules[ruleSpnl]() {
				goto l418
			}
			{
				if position == len(p.Buffer) {
					goto l418
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l418
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l418
					}
					break
				default:
					goto l418
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l418
			}
		l420:
			if !p.rules[ruleHtmlAttribute]() {
				goto l421
			}
			goto l420
		l421:
			if !matchChar('>') {
				goto l418
			}
			return true
		l418:
			position = position0
			return false
		},
		/* 95 HtmlBlockCloseUl <- ('<' Spnl '/' ((&[U] 'UL') | (&[u] 'ul')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l422
			}
			if !p.rules[ruleSpnl]() {
				goto l422
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l422
			}
			return true
		l422:
			position = position0
			return false
		},
		/* 96 HtmlBlockUl <- (HtmlBlockOpenUl (HtmlBlockUl / (!HtmlBlockCloseUl .))* HtmlBlockCloseUl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenUl]() {
				goto l424
			}
		l425:
			{
				position426 := position
				if !p.rules[ruleHtmlBlockUl]() {
					goto l428
				}
				goto l427
			l428:
				if !p.rules[ruleHtmlBlockCloseUl]() {
					goto l429
				}
				goto l426
			l429:
				if !matchDot() {
					goto l426
				}
			l427:
				goto l425
			l426:
				position = position426
			}
			if !p.rules[ruleHtmlBlockCloseUl]() {
				goto l424
			}
			return true
		l424:
			position = position0
			return false
		},
		/* 97 HtmlBlockOpenDd <- ('<' Spnl ((&[D] 'DD') | (&[d] 'dd')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l430
			}
			if !p.rules[ruleSpnl]() {
				goto l430
			}
			{
				if position == len(p.Buffer) {
					goto l430
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto l430
					}
					break
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto l430
					}
					break
				default:
					goto l430
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l430
			}
		l432:
			if !p.rules[ruleHtmlAttribute]() {
				goto l433
			}
			goto l432
		l433:
			if !matchChar('>') {
				goto l430
			}
			return true
		l430:
			position = position0
			return false
		},
		/* 98 HtmlBlockCloseDd <- ('<' Spnl '/' ((&[D] 'DD') | (&[d] 'dd')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l434
			}
			if !p.rules[ruleSpnl]() {
				goto l434
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l434
			}
			return true
		l434:
			position = position0
			return false
		},
		/* 99 HtmlBlockDd <- (HtmlBlockOpenDd (HtmlBlockDd / (!HtmlBlockCloseDd .))* HtmlBlockCloseDd) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDd]() {
				goto l436
			}
		l437:
			{
				position438 := position
				if !p.rules[ruleHtmlBlockDd]() {
					goto l440
				}
				goto l439
			l440:
				if !p.rules[ruleHtmlBlockCloseDd]() {
					goto l441
				}
				goto l438
			l441:
				if !matchDot() {
					goto l438
				}
			l439:
				goto l437
			l438:
				position = position438
			}
			if !p.rules[ruleHtmlBlockCloseDd]() {
				goto l436
			}
			return true
		l436:
			position = position0
			return false
		},
		/* 100 HtmlBlockOpenDt <- ('<' Spnl ((&[D] 'DT') | (&[d] 'dt')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l442
			}
			if !p.rules[ruleSpnl]() {
				goto l442
			}
			{
				if position == len(p.Buffer) {
					goto l442
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l442
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l442
					}
					break
				default:
					goto l442
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l442
			}
		l444:
			if !p.rules[ruleHtmlAttribute]() {
				goto l445
			}
			goto l444
		l445:
			if !matchChar('>') {
				goto l442
			}
			return true
		l442:
			position = position0
			return false
		},
		/* 101 HtmlBlockCloseDt <- ('<' Spnl '/' ((&[D] 'DT') | (&[d] 'dt')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l446
			}
			if !p.rules[ruleSpnl]() {
				goto l446
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l446
			}
			return true
		l446:
			position = position0
			return false
		},
		/* 102 HtmlBlockDt <- (HtmlBlockOpenDt (HtmlBlockDt / (!HtmlBlockCloseDt .))* HtmlBlockCloseDt) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDt]() {
				goto l448
			}
		l449:
			{
				position450 := position
				if !p.rules[ruleHtmlBlockDt]() {
					goto l452
				}
				goto l451
			l452:
				if !p.rules[ruleHtmlBlockCloseDt]() {
					goto l453
				}
				goto l450
			l453:
				if !matchDot() {
					goto l450
				}
			l451:
				goto l449
			l450:
				position = position450
			}
			if !p.rules[ruleHtmlBlockCloseDt]() {
				goto l448
			}
			return true
		l448:
			position = position0
			return false
		},
		/* 103 HtmlBlockOpenFrameset <- ('<' Spnl ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l454
			}
			if !p.rules[ruleSpnl]() {
				goto l454
			}
			{
				if position == len(p.Buffer) {
					goto l454
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l454
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l454
					}
					break
				default:
					goto l454
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l454
			}
		l456:
			if !p.rules[ruleHtmlAttribute]() {
				goto l457
			}
			goto l456
		l457:
			if !matchChar('>') {
				goto l454
			}
			return true
		l454:
			position = position0
			return false
		},
		/* 104 HtmlBlockCloseFrameset <- ('<' Spnl '/' ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l458
			}
			if !p.rules[ruleSpnl]() {
				goto l458
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l458
			}
			return true
		l458:
			position = position0
			return false
		},
		/* 105 HtmlBlockFrameset <- (HtmlBlockOpenFrameset (HtmlBlockFrameset / (!HtmlBlockCloseFrameset .))* HtmlBlockCloseFrameset) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFrameset]() {
				goto l460
			}
		l461:
			{
				position462 := position
				if !p.rules[ruleHtmlBlockFrameset]() {
					goto l464
				}
				goto l463
			l464:
				if !p.rules[ruleHtmlBlockCloseFrameset]() {
					goto l465
				}
				goto l462
			l465:
				if !matchDot() {
					goto l462
				}
			l463:
				goto l461
			l462:
				position = position462
			}
			if !p.rules[ruleHtmlBlockCloseFrameset]() {
				goto l460
			}
			return true
		l460:
			position = position0
			return false
		},
		/* 106 HtmlBlockOpenLi <- ('<' Spnl ((&[L] 'LI') | (&[l] 'li')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l466
			}
			if !p.rules[ruleSpnl]() {
				goto l466
			}
			{
				if position == len(p.Buffer) {
					goto l466
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l466
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l466
					}
					break
				default:
					goto l466
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l466
			}
		l468:
			if !p.rules[ruleHtmlAttribute]() {
				goto l469
			}
			goto l468
		l469:
			if !matchChar('>') {
				goto l466
			}
			return true
		l466:
			position = position0
			return false
		},
		/* 107 HtmlBlockCloseLi <- ('<' Spnl '/' ((&[L] 'LI') | (&[l] 'li')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l470
			}
			if !p.rules[ruleSpnl]() {
				goto l470
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l470
			}
			return true
		l470:
			position = position0
			return false
		},
		/* 108 HtmlBlockLi <- (HtmlBlockOpenLi (HtmlBlockLi / (!HtmlBlockCloseLi .))* HtmlBlockCloseLi) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenLi]() {
				goto l472
			}
		l473:
			{
				position474 := position
				if !p.rules[ruleHtmlBlockLi]() {
					goto l476
				}
				goto l475
			l476:
				if !p.rules[ruleHtmlBlockCloseLi]() {
					goto l477
				}
				goto l474
			l477:
				if !matchDot() {
					goto l474
				}
			l475:
				goto l473
			l474:
				position = position474
			}
			if !p.rules[ruleHtmlBlockCloseLi]() {
				goto l472
			}
			return true
		l472:
			position = position0
			return false
		},
		/* 109 HtmlBlockOpenTbody <- ('<' Spnl ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l478
			}
			if !p.rules[ruleSpnl]() {
				goto l478
			}
			{
				if position == len(p.Buffer) {
					goto l478
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto l478
					}
					break
				case 't':
					position++
					if !matchString("body") {
						goto l478
					}
					break
				default:
					goto l478
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l478
			}
		l480:
			if !p.rules[ruleHtmlAttribute]() {
				goto l481
			}
			goto l480
		l481:
			if !matchChar('>') {
				goto l478
			}
			return true
		l478:
			position = position0
			return false
		},
		/* 110 HtmlBlockCloseTbody <- ('<' Spnl '/' ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l482
			}
			if !p.rules[ruleSpnl]() {
				goto l482
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l482
			}
			return true
		l482:
			position = position0
			return false
		},
		/* 111 HtmlBlockTbody <- (HtmlBlockOpenTbody (HtmlBlockTbody / (!HtmlBlockCloseTbody .))* HtmlBlockCloseTbody) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTbody]() {
				goto l484
			}
		l485:
			{
				position486 := position
				if !p.rules[ruleHtmlBlockTbody]() {
					goto l488
				}
				goto l487
			l488:
				if !p.rules[ruleHtmlBlockCloseTbody]() {
					goto l489
				}
				goto l486
			l489:
				if !matchDot() {
					goto l486
				}
			l487:
				goto l485
			l486:
				position = position486
			}
			if !p.rules[ruleHtmlBlockCloseTbody]() {
				goto l484
			}
			return true
		l484:
			position = position0
			return false
		},
		/* 112 HtmlBlockOpenTd <- ('<' Spnl ((&[T] 'TD') | (&[t] 'td')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l490
			}
			if !p.rules[ruleSpnl]() {
				goto l490
			}
			{
				if position == len(p.Buffer) {
					goto l490
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto l490
					}
					break
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto l490
					}
					break
				default:
					goto l490
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l490
			}
		l492:
			if !p.rules[ruleHtmlAttribute]() {
				goto l493
			}
			goto l492
		l493:
			if !matchChar('>') {
				goto l490
			}
			return true
		l490:
			position = position0
			return false
		},
		/* 113 HtmlBlockCloseTd <- ('<' Spnl '/' ((&[T] 'TD') | (&[t] 'td')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l494
			}
			if !p.rules[ruleSpnl]() {
				goto l494
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l494
			}
			return true
		l494:
			position = position0
			return false
		},
		/* 114 HtmlBlockTd <- (HtmlBlockOpenTd (HtmlBlockTd / (!HtmlBlockCloseTd .))* HtmlBlockCloseTd) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTd]() {
				goto l496
			}
		l497:
			{
				position498 := position
				if !p.rules[ruleHtmlBlockTd]() {
					goto l500
				}
				goto l499
			l500:
				if !p.rules[ruleHtmlBlockCloseTd]() {
					goto l501
				}
				goto l498
			l501:
				if !matchDot() {
					goto l498
				}
			l499:
				goto l497
			l498:
				position = position498
			}
			if !p.rules[ruleHtmlBlockCloseTd]() {
				goto l496
			}
			return true
		l496:
			position = position0
			return false
		},
		/* 115 HtmlBlockOpenTfoot <- ('<' Spnl ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l502
			}
			if !p.rules[ruleSpnl]() {
				goto l502
			}
			{
				if position == len(p.Buffer) {
					goto l502
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto l502
					}
					break
				case 't':
					position++
					if !matchString("foot") {
						goto l502
					}
					break
				default:
					goto l502
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l502
			}
		l504:
			if !p.rules[ruleHtmlAttribute]() {
				goto l505
			}
			goto l504
		l505:
			if !matchChar('>') {
				goto l502
			}
			return true
		l502:
			position = position0
			return false
		},
		/* 116 HtmlBlockCloseTfoot <- ('<' Spnl '/' ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l506
			}
			if !p.rules[ruleSpnl]() {
				goto l506
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l506
			}
			return true
		l506:
			position = position0
			return false
		},
		/* 117 HtmlBlockTfoot <- (HtmlBlockOpenTfoot (HtmlBlockTfoot / (!HtmlBlockCloseTfoot .))* HtmlBlockCloseTfoot) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTfoot]() {
				goto l508
			}
		l509:
			{
				position510 := position
				if !p.rules[ruleHtmlBlockTfoot]() {
					goto l512
				}
				goto l511
			l512:
				if !p.rules[ruleHtmlBlockCloseTfoot]() {
					goto l513
				}
				goto l510
			l513:
				if !matchDot() {
					goto l510
				}
			l511:
				goto l509
			l510:
				position = position510
			}
			if !p.rules[ruleHtmlBlockCloseTfoot]() {
				goto l508
			}
			return true
		l508:
			position = position0
			return false
		},
		/* 118 HtmlBlockOpenTh <- ('<' Spnl ((&[T] 'TH') | (&[t] 'th')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l514
			}
			if !p.rules[ruleSpnl]() {
				goto l514
			}
			{
				if position == len(p.Buffer) {
					goto l514
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto l514
					}
					break
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto l514
					}
					break
				default:
					goto l514
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l514
			}
		l516:
			if !p.rules[ruleHtmlAttribute]() {
				goto l517
			}
			goto l516
		l517:
			if !matchChar('>') {
				goto l514
			}
			return true
		l514:
			position = position0
			return false
		},
		/* 119 HtmlBlockCloseTh <- ('<' Spnl '/' ((&[T] 'TH') | (&[t] 'th')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l518
			}
			if !p.rules[ruleSpnl]() {
				goto l518
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l518
			}
			return true
		l518:
			position = position0
			return false
		},
		/* 120 HtmlBlockTh <- (HtmlBlockOpenTh (HtmlBlockTh / (!HtmlBlockCloseTh .))* HtmlBlockCloseTh) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTh]() {
				goto l520
			}
		l521:
			{
				position522 := position
				if !p.rules[ruleHtmlBlockTh]() {
					goto l524
				}
				goto l523
			l524:
				if !p.rules[ruleHtmlBlockCloseTh]() {
					goto l525
				}
				goto l522
			l525:
				if !matchDot() {
					goto l522
				}
			l523:
				goto l521
			l522:
				position = position522
			}
			if !p.rules[ruleHtmlBlockCloseTh]() {
				goto l520
			}
			return true
		l520:
			position = position0
			return false
		},
		/* 121 HtmlBlockOpenThead <- ('<' Spnl ((&[T] 'THEAD') | (&[t] 'thead')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l526
			}
			if !p.rules[ruleSpnl]() {
				goto l526
			}
			{
				if position == len(p.Buffer) {
					goto l526
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto l526
					}
					break
				case 't':
					position++
					if !matchString("head") {
						goto l526
					}
					break
				default:
					goto l526
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l526
			}
		l528:
			if !p.rules[ruleHtmlAttribute]() {
				goto l529
			}
			goto l528
		l529:
			if !matchChar('>') {
				goto l526
			}
			return true
		l526:
			position = position0
			return false
		},
		/* 122 HtmlBlockCloseThead <- ('<' Spnl '/' ((&[T] 'THEAD') | (&[t] 'thead')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l530
			}
			if !p.rules[ruleSpnl]() {
				goto l530
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l530
			}
			return true
		l530:
			position = position0
			return false
		},
		/* 123 HtmlBlockThead <- (HtmlBlockOpenThead (HtmlBlockThead / (!HtmlBlockCloseThead .))* HtmlBlockCloseThead) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenThead]() {
				goto l532
			}
		l533:
			{
				position534 := position
				if !p.rules[ruleHtmlBlockThead]() {
					goto l536
				}
				goto l535
			l536:
				if !p.rules[ruleHtmlBlockCloseThead]() {
					goto l537
				}
				goto l534
			l537:
				if !matchDot() {
					goto l534
				}
			l535:
				goto l533
			l534:
				position = position534
			}
			if !p.rules[ruleHtmlBlockCloseThead]() {
				goto l532
			}
			return true
		l532:
			position = position0
			return false
		},
		/* 124 HtmlBlockOpenTr <- ('<' Spnl ((&[T] 'TR') | (&[t] 'tr')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l538
			}
			if !p.rules[ruleSpnl]() {
				goto l538
			}
			{
				if position == len(p.Buffer) {
					goto l538
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l538
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l538
					}
					break
				default:
					goto l538
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l538
			}
		l540:
			if !p.rules[ruleHtmlAttribute]() {
				goto l541
			}
			goto l540
		l541:
			if !matchChar('>') {
				goto l538
			}
			return true
		l538:
			position = position0
			return false
		},
		/* 125 HtmlBlockCloseTr <- ('<' Spnl '/' ((&[T] 'TR') | (&[t] 'tr')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l542
			}
			if !p.rules[ruleSpnl]() {
				goto l542
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l542
			}
			return true
		l542:
			position = position0
			return false
		},
		/* 126 HtmlBlockTr <- (HtmlBlockOpenTr (HtmlBlockTr / (!HtmlBlockCloseTr .))* HtmlBlockCloseTr) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTr]() {
				goto l544
			}
		l545:
			{
				position546 := position
				if !p.rules[ruleHtmlBlockTr]() {
					goto l548
				}
				goto l547
			l548:
				if !p.rules[ruleHtmlBlockCloseTr]() {
					goto l549
				}
				goto l546
			l549:
				if !matchDot() {
					goto l546
				}
			l547:
				goto l545
			l546:
				position = position546
			}
			if !p.rules[ruleHtmlBlockCloseTr]() {
				goto l544
			}
			return true
		l544:
			position = position0
			return false
		},
		/* 127 HtmlBlockOpenScript <- ('<' Spnl ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l550
			}
			if !p.rules[ruleSpnl]() {
				goto l550
			}
			{
				if position == len(p.Buffer) {
					goto l550
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l550
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l550
					}
					break
				default:
					goto l550
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l550
			}
		l552:
			if !p.rules[ruleHtmlAttribute]() {
				goto l553
			}
			goto l552
		l553:
			if !matchChar('>') {
				goto l550
			}
			return true
		l550:
			position = position0
			return false
		},
		/* 128 HtmlBlockCloseScript <- ('<' Spnl '/' ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l554
			}
			if !p.rules[ruleSpnl]() {
				goto l554
			}
			if !matchChar('/') {
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
			if !matchChar('>') {
				goto l554
			}
			return true
		l554:
			position = position0
			return false
		},
		/* 129 HtmlBlockScript <- (HtmlBlockOpenScript (HtmlBlockScript / (!HtmlBlockCloseScript .))* HtmlBlockCloseScript) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenScript]() {
				goto l556
			}
		l557:
			{
				position558 := position
				if !p.rules[ruleHtmlBlockScript]() {
					goto l560
				}
				goto l559
			l560:
				if !p.rules[ruleHtmlBlockCloseScript]() {
					goto l561
				}
				goto l558
			l561:
				if !matchDot() {
					goto l558
				}
			l559:
				goto l557
			l558:
				position = position558
			}
			if !p.rules[ruleHtmlBlockCloseScript]() {
				goto l556
			}
			return true
		l556:
			position = position0
			return false
		},
		/* 130 HtmlBlockInTags <- (HtmlBlockAddress / HtmlBlockBlockquote / HtmlBlockCenter / HtmlBlockDir / HtmlBlockDiv / HtmlBlockDl / HtmlBlockFieldset / HtmlBlockForm / HtmlBlockH1 / HtmlBlockH2 / HtmlBlockH3 / HtmlBlockH4 / HtmlBlockH5 / HtmlBlockH6 / HtmlBlockMenu / HtmlBlockNoframes / HtmlBlockNoscript / HtmlBlockOl / HtmlBlockP / HtmlBlockPre / HtmlBlockTable / HtmlBlockUl / HtmlBlockDd / HtmlBlockDt / HtmlBlockFrameset / HtmlBlockLi / HtmlBlockTbody / HtmlBlockTd / HtmlBlockTfoot / HtmlBlockTh / HtmlBlockThead / HtmlBlockTr / HtmlBlockScript) */
		func() bool {
			if !p.rules[ruleHtmlBlockAddress]() {
				goto l564
			}
			goto l563
		l564:
			if !p.rules[ruleHtmlBlockBlockquote]() {
				goto l565
			}
			goto l563
		l565:
			if !p.rules[ruleHtmlBlockCenter]() {
				goto l566
			}
			goto l563
		l566:
			if !p.rules[ruleHtmlBlockDir]() {
				goto l567
			}
			goto l563
		l567:
			if !p.rules[ruleHtmlBlockDiv]() {
				goto l568
			}
			goto l563
		l568:
			if !p.rules[ruleHtmlBlockDl]() {
				goto l569
			}
			goto l563
		l569:
			if !p.rules[ruleHtmlBlockFieldset]() {
				goto l570
			}
			goto l563
		l570:
			if !p.rules[ruleHtmlBlockForm]() {
				goto l571
			}
			goto l563
		l571:
			if !p.rules[ruleHtmlBlockH1]() {
				goto l572
			}
			goto l563
		l572:
			if !p.rules[ruleHtmlBlockH2]() {
				goto l573
			}
			goto l563
		l573:
			if !p.rules[ruleHtmlBlockH3]() {
				goto l574
			}
			goto l563
		l574:
			if !p.rules[ruleHtmlBlockH4]() {
				goto l575
			}
			goto l563
		l575:
			if !p.rules[ruleHtmlBlockH5]() {
				goto l576
			}
			goto l563
		l576:
			if !p.rules[ruleHtmlBlockH6]() {
				goto l577
			}
			goto l563
		l577:
			if !p.rules[ruleHtmlBlockMenu]() {
				goto l578
			}
			goto l563
		l578:
			if !p.rules[ruleHtmlBlockNoframes]() {
				goto l579
			}
			goto l563
		l579:
			if !p.rules[ruleHtmlBlockNoscript]() {
				goto l580
			}
			goto l563
		l580:
			if !p.rules[ruleHtmlBlockOl]() {
				goto l581
			}
			goto l563
		l581:
			if !p.rules[ruleHtmlBlockP]() {
				goto l582
			}
			goto l563
		l582:
			if !p.rules[ruleHtmlBlockPre]() {
				goto l583
			}
			goto l563
		l583:
			if !p.rules[ruleHtmlBlockTable]() {
				goto l584
			}
			goto l563
		l584:
			if !p.rules[ruleHtmlBlockUl]() {
				goto l585
			}
			goto l563
		l585:
			if !p.rules[ruleHtmlBlockDd]() {
				goto l586
			}
			goto l563
		l586:
			if !p.rules[ruleHtmlBlockDt]() {
				goto l587
			}
			goto l563
		l587:
			if !p.rules[ruleHtmlBlockFrameset]() {
				goto l588
			}
			goto l563
		l588:
			if !p.rules[ruleHtmlBlockLi]() {
				goto l589
			}
			goto l563
		l589:
			if !p.rules[ruleHtmlBlockTbody]() {
				goto l590
			}
			goto l563
		l590:
			if !p.rules[ruleHtmlBlockTd]() {
				goto l591
			}
			goto l563
		l591:
			if !p.rules[ruleHtmlBlockTfoot]() {
				goto l592
			}
			goto l563
		l592:
			if !p.rules[ruleHtmlBlockTh]() {
				goto l593
			}
			goto l563
		l593:
			if !p.rules[ruleHtmlBlockThead]() {
				goto l594
			}
			goto l563
		l594:
			if !p.rules[ruleHtmlBlockTr]() {
				goto l595
			}
			goto l563
		l595:
			if !p.rules[ruleHtmlBlockScript]() {
				goto l562
			}
		l563:
			return true
		l562:
			return false
		},
		/* 131 HtmlBlock <- (&'<' < (HtmlBlockInTags / HtmlComment / HtmlBlockSelfClosing) > BlankLine+ {   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTMLBLOCK
                }
            }) */
		func() bool {
			position0 := position
			if !peekChar('<') {
				goto l596
			}
			begin = position
			if !p.rules[ruleHtmlBlockInTags]() {
				goto l598
			}
			goto l597
		l598:
			if !p.rules[ruleHtmlComment]() {
				goto l599
			}
			goto l597
		l599:
			if !p.rules[ruleHtmlBlockSelfClosing]() {
				goto l596
			}
		l597:
			end = position
			if !p.rules[ruleBlankLine]() {
				goto l596
			}
		l600:
			if !p.rules[ruleBlankLine]() {
				goto l601
			}
			goto l600
		l601:
			do(41)
			return true
		l596:
			position = position0
			return false
		},
		/* 132 HtmlBlockSelfClosing <- ('<' Spnl HtmlBlockType Spnl HtmlAttribute* '/' Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l602
			}
			if !p.rules[ruleSpnl]() {
				goto l602
			}
			if !p.rules[ruleHtmlBlockType]() {
				goto l602
			}
			if !p.rules[ruleSpnl]() {
				goto l602
			}
		l603:
			if !p.rules[ruleHtmlAttribute]() {
				goto l604
			}
			goto l603
		l604:
			if !matchChar('/') {
				goto l602
			}
			if !p.rules[ruleSpnl]() {
				goto l602
			}
			if !matchChar('>') {
				goto l602
			}
			return true
		l602:
			position = position0
			return false
		},
		/* 133 HtmlBlockType <- ('dir' / 'div' / 'dl' / 'fieldset' / 'form' / 'h1' / 'h2' / 'h3' / 'h4' / 'h5' / 'h6' / 'noframes' / 'p' / 'table' / 'dd' / 'tbody' / 'td' / 'tfoot' / 'th' / 'thead' / 'DIR' / 'DIV' / 'DL' / 'FIELDSET' / 'FORM' / 'H1' / 'H2' / 'H3' / 'H4' / 'H5' / 'H6' / 'NOFRAMES' / 'P' / 'TABLE' / 'DD' / 'TBODY' / 'TD' / 'TFOOT' / 'TH' / 'THEAD' / ((&[S] 'SCRIPT') | (&[T] 'TR') | (&[L] 'LI') | (&[F] 'FRAMESET') | (&[D] 'DT') | (&[U] 'UL') | (&[P] 'PRE') | (&[O] 'OL') | (&[N] 'NOSCRIPT') | (&[M] 'MENU') | (&[I] 'ISINDEX') | (&[H] 'HR') | (&[C] 'CENTER') | (&[B] 'BLOCKQUOTE') | (&[A] 'ADDRESS') | (&[s] 'script') | (&[t] 'tr') | (&[l] 'li') | (&[f] 'frameset') | (&[d] 'dt') | (&[u] 'ul') | (&[p] 'pre') | (&[o] 'ol') | (&[n] 'noscript') | (&[m] 'menu') | (&[i] 'isindex') | (&[h] 'hr') | (&[c] 'center') | (&[b] 'blockquote') | (&[a] 'address'))) */
		func() bool {
			if !matchString("dir") {
				goto l607
			}
			goto l606
		l607:
			if !matchString("div") {
				goto l608
			}
			goto l606
		l608:
			if !matchString("dl") {
				goto l609
			}
			goto l606
		l609:
			if !matchString("fieldset") {
				goto l610
			}
			goto l606
		l610:
			if !matchString("form") {
				goto l611
			}
			goto l606
		l611:
			if !matchString("h1") {
				goto l612
			}
			goto l606
		l612:
			if !matchString("h2") {
				goto l613
			}
			goto l606
		l613:
			if !matchString("h3") {
				goto l614
			}
			goto l606
		l614:
			if !matchString("h4") {
				goto l615
			}
			goto l606
		l615:
			if !matchString("h5") {
				goto l616
			}
			goto l606
		l616:
			if !matchString("h6") {
				goto l617
			}
			goto l606
		l617:
			if !matchString("noframes") {
				goto l618
			}
			goto l606
		l618:
			if !matchChar('p') {
				goto l619
			}
			goto l606
		l619:
			if !matchString("table") {
				goto l620
			}
			goto l606
		l620:
			if !matchString("dd") {
				goto l621
			}
			goto l606
		l621:
			if !matchString("tbody") {
				goto l622
			}
			goto l606
		l622:
			if !matchString("td") {
				goto l623
			}
			goto l606
		l623:
			if !matchString("tfoot") {
				goto l624
			}
			goto l606
		l624:
			if !matchString("th") {
				goto l625
			}
			goto l606
		l625:
			if !matchString("thead") {
				goto l626
			}
			goto l606
		l626:
			if !matchString("DIR") {
				goto l627
			}
			goto l606
		l627:
			if !matchString("DIV") {
				goto l628
			}
			goto l606
		l628:
			if !matchString("DL") {
				goto l629
			}
			goto l606
		l629:
			if !matchString("FIELDSET") {
				goto l630
			}
			goto l606
		l630:
			if !matchString("FORM") {
				goto l631
			}
			goto l606
		l631:
			if !matchString("H1") {
				goto l632
			}
			goto l606
		l632:
			if !matchString("H2") {
				goto l633
			}
			goto l606
		l633:
			if !matchString("H3") {
				goto l634
			}
			goto l606
		l634:
			if !matchString("H4") {
				goto l635
			}
			goto l606
		l635:
			if !matchString("H5") {
				goto l636
			}
			goto l606
		l636:
			if !matchString("H6") {
				goto l637
			}
			goto l606
		l637:
			if !matchString("NOFRAMES") {
				goto l638
			}
			goto l606
		l638:
			if !matchChar('P') {
				goto l639
			}
			goto l606
		l639:
			if !matchString("TABLE") {
				goto l640
			}
			goto l606
		l640:
			if !matchString("DD") {
				goto l641
			}
			goto l606
		l641:
			if !matchString("TBODY") {
				goto l642
			}
			goto l606
		l642:
			if !matchString("TD") {
				goto l643
			}
			goto l606
		l643:
			if !matchString("TFOOT") {
				goto l644
			}
			goto l606
		l644:
			if !matchString("TH") {
				goto l645
			}
			goto l606
		l645:
			if !matchString("THEAD") {
				goto l646
			}
			goto l606
		l646:
			{
				if position == len(p.Buffer) {
					goto l605
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l605
					}
					break
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l605
					}
					break
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l605
					}
					break
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l605
					}
					break
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l605
					}
					break
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l605
					}
					break
				case 'P':
					position++
					if !matchString("RE") {
						goto l605
					}
					break
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l605
					}
					break
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l605
					}
					break
				case 'M':
					position++
					if !matchString("ENU") {
						goto l605
					}
					break
				case 'I':
					position++
					if !matchString("SINDEX") {
						goto l605
					}
					break
				case 'H':
					position++ // matchString(`HR`)
					if !matchChar('R') {
						goto l605
					}
					break
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l605
					}
					break
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l605
					}
					break
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l605
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l605
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l605
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l605
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l605
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l605
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l605
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l605
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l605
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l605
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l605
					}
					break
				case 'i':
					position++
					if !matchString("sindex") {
						goto l605
					}
					break
				case 'h':
					position++ // matchString(`hr`)
					if !matchChar('r') {
						goto l605
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l605
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l605
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l605
					}
					break
				default:
					goto l605
				}
			}
		l606:
			return true
		l605:
			return false
		},
		/* 134 StyleOpen <- ('<' Spnl ((&[S] 'STYLE') | (&[s] 'style')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l648
			}
			if !p.rules[ruleSpnl]() {
				goto l648
			}
			{
				if position == len(p.Buffer) {
					goto l648
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto l648
					}
					break
				case 's':
					position++
					if !matchString("tyle") {
						goto l648
					}
					break
				default:
					goto l648
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l648
			}
		l650:
			if !p.rules[ruleHtmlAttribute]() {
				goto l651
			}
			goto l650
		l651:
			if !matchChar('>') {
				goto l648
			}
			return true
		l648:
			position = position0
			return false
		},
		/* 135 StyleClose <- ('<' Spnl '/' ((&[S] 'STYLE') | (&[s] 'style')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l652
			}
			if !p.rules[ruleSpnl]() {
				goto l652
			}
			if !matchChar('/') {
				goto l652
			}
			{
				if position == len(p.Buffer) {
					goto l652
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto l652
					}
					break
				case 's':
					position++
					if !matchString("tyle") {
						goto l652
					}
					break
				default:
					goto l652
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l652
			}
			if !matchChar('>') {
				goto l652
			}
			return true
		l652:
			position = position0
			return false
		},
		/* 136 InStyleTags <- (StyleOpen (!StyleClose .)* StyleClose) */
		func() bool {
			position0 := position
			if !p.rules[ruleStyleOpen]() {
				goto l654
			}
		l655:
			{
				position656 := position
				if !p.rules[ruleStyleClose]() {
					goto l657
				}
				goto l656
			l657:
				if !matchDot() {
					goto l656
				}
				goto l655
			l656:
				position = position656
			}
			if !p.rules[ruleStyleClose]() {
				goto l654
			}
			return true
		l654:
			position = position0
			return false
		},
		/* 137 StyleBlock <- (< InStyleTags > BlankLine* {   if p.extension.FilterStyles {
                        yy = mk_list(LIST, nil)
                    } else {
                        yy = mk_str(yytext)
                        yy.key = HTMLBLOCK
                    }
                }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleInStyleTags]() {
				goto l658
			}
			end = position
		l659:
			if !p.rules[ruleBlankLine]() {
				goto l660
			}
			goto l659
		l660:
			do(42)
			return true
		l658:
			position = position0
			return false
		},
		/* 138 Inlines <- (StartList ((!Endline Inline { a = cons(yy, a) }) / (Endline &Inline { a = cons(c, a) }))+ Endline? { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l661
			}
			doarg(yySet, -1)
			{
				position664 := position
				if !p.rules[ruleEndline]() {
					goto l666
				}
				goto l665
			l666:
				if !p.rules[ruleInline]() {
					goto l665
				}
				do(43)
				goto l664
			l665:
				position = position664
				if !p.rules[ruleEndline]() {
					goto l661
				}
				doarg(yySet, -2)
				{
					position667 := position
					if !p.rules[ruleInline]() {
						goto l661
					}
					position = position667
				}
				do(44)
			}
		l664:
		l662:
			{
				position663, thunkPosition663 := position, thunkPosition
				{
					position668 := position
					if !p.rules[ruleEndline]() {
						goto l670
					}
					goto l669
				l670:
					if !p.rules[ruleInline]() {
						goto l669
					}
					do(43)
					goto l668
				l669:
					position = position668
					if !p.rules[ruleEndline]() {
						goto l663
					}
					doarg(yySet, -2)
					{
						position671 := position
						if !p.rules[ruleInline]() {
							goto l663
						}
						position = position671
					}
					do(44)
				}
			l668:
				goto l662
			l663:
				position, thunkPosition = position663, thunkPosition663
			}
			if !p.rules[ruleEndline]() {
				goto l672
			}
		l672:
			do(45)
			doarg(yyPop, 2)
			return true
		l661:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 139 Inline <- (Str / Endline / UlOrStarLine / Space / Strong / Emph / Image / Link / NoteReference / InlineNote / Code / RawHtml / Entity / EscapedChar / Smart / Symbol) */
		func() bool {
			if !p.rules[ruleStr]() {
				goto l676
			}
			goto l675
		l676:
			if !p.rules[ruleEndline]() {
				goto l677
			}
			goto l675
		l677:
			if !p.rules[ruleUlOrStarLine]() {
				goto l678
			}
			goto l675
		l678:
			if !p.rules[ruleSpace]() {
				goto l679
			}
			goto l675
		l679:
			if !p.rules[ruleStrong]() {
				goto l680
			}
			goto l675
		l680:
			if !p.rules[ruleEmph]() {
				goto l681
			}
			goto l675
		l681:
			if !p.rules[ruleImage]() {
				goto l682
			}
			goto l675
		l682:
			if !p.rules[ruleLink]() {
				goto l683
			}
			goto l675
		l683:
			if !p.rules[ruleNoteReference]() {
				goto l684
			}
			goto l675
		l684:
			if !p.rules[ruleInlineNote]() {
				goto l685
			}
			goto l675
		l685:
			if !p.rules[ruleCode]() {
				goto l686
			}
			goto l675
		l686:
			if !p.rules[ruleRawHtml]() {
				goto l687
			}
			goto l675
		l687:
			if !p.rules[ruleEntity]() {
				goto l688
			}
			goto l675
		l688:
			if !p.rules[ruleEscapedChar]() {
				goto l689
			}
			goto l675
		l689:
			if !p.rules[ruleSmart]() {
				goto l690
			}
			goto l675
		l690:
			if !p.rules[ruleSymbol]() {
				goto l674
			}
		l675:
			return true
		l674:
			return false
		},
		/* 140 Space <- (Spacechar+ { yy = mk_str(" ")
          yy.key = SPACE }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto l691
			}
		l692:
			if !p.rules[ruleSpacechar]() {
				goto l693
			}
			goto l692
		l693:
			do(46)
			return true
		l691:
			position = position0
			return false
		},
		/* 141 Str <- (< NormalChar (NormalChar / ('_'+ &Alphanumeric))* > { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleNormalChar]() {
				goto l694
			}
		l695:
			{
				position696 := position
				if !p.rules[ruleNormalChar]() {
					goto l698
				}
				goto l697
			l698:
				if !matchChar('_') {
					goto l696
				}
			l699:
				if !matchChar('_') {
					goto l700
				}
				goto l699
			l700:
				{
					position701 := position
					if !p.rules[ruleAlphanumeric]() {
						goto l696
					}
					position = position701
				}
			l697:
				goto l695
			l696:
				position = position696
			}
			end = position
			do(47)
			return true
		l694:
			position = position0
			return false
		},
		/* 142 EscapedChar <- ('\\' !Newline < [-\\`|*_{}[\]()#+.!><] > { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !matchChar('\\') {
				goto l702
			}
			if !p.rules[ruleNewline]() {
				goto l703
			}
			goto l702
		l703:
			begin = position
			if !matchClass(1) {
				goto l702
			}
			end = position
			do(48)
			return true
		l702:
			position = position0
			return false
		},
		/* 143 Entity <- ((HexEntity / DecEntity / CharEntity) { yy = mk_str(yytext); yy.key = HTML }) */
		func() bool {
			position0 := position
			if !p.rules[ruleHexEntity]() {
				goto l706
			}
			goto l705
		l706:
			if !p.rules[ruleDecEntity]() {
				goto l707
			}
			goto l705
		l707:
			if !p.rules[ruleCharEntity]() {
				goto l704
			}
		l705:
			do(49)
			return true
		l704:
			position = position0
			return false
		},
		/* 144 Endline <- (LineBreak / TerminalEndline / NormalEndline) */
		func() bool {
			if !p.rules[ruleLineBreak]() {
				goto l710
			}
			goto l709
		l710:
			if !p.rules[ruleTerminalEndline]() {
				goto l711
			}
			goto l709
		l711:
			if !p.rules[ruleNormalEndline]() {
				goto l708
			}
		l709:
			return true
		l708:
			return false
		},
		/* 145 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line ((&[\-] ('---' '-'*)) | (&[=] ('===' '='*))) Newline) { yy = mk_str("\n")
                    yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l712
			}
			if !p.rules[ruleNewline]() {
				goto l712
			}
			if !p.rules[ruleBlankLine]() {
				goto l713
			}
			goto l712
		l713:
			if peekChar('>') {
				goto l712
			}
			if !p.rules[ruleAtxStart]() {
				goto l714
			}
			goto l712
		l714:
			{
				position715, thunkPosition715 := position, thunkPosition
				if !p.rules[ruleLine]() {
					goto l715
				}
				{
					if position == len(p.Buffer) {
						goto l715
					}
					switch p.Buffer[position] {
					case '-':
						position++
						if !matchString("--") {
							goto l715
						}
					l717:
						if !matchChar('-') {
							goto l718
						}
						goto l717
					l718:
						break
					case '=':
						position++
						if !matchString("==") {
							goto l715
						}
					l719:
						if !matchChar('=') {
							goto l720
						}
						goto l719
					l720:
						break
					default:
						goto l715
					}
				}
				if !p.rules[ruleNewline]() {
					goto l715
				}
				goto l712
			l715:
				position, thunkPosition = position715, thunkPosition715
			}
			do(50)
			return true
		l712:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 146 TerminalEndline <- (Sp Newline !. { yy = nil }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l721
			}
			if !p.rules[ruleNewline]() {
				goto l721
			}
			if (position < len(p.Buffer)) {
				goto l721
			}
			do(51)
			return true
		l721:
			position = position0
			return false
		},
		/* 147 LineBreak <- ('  ' NormalEndline { yy = mk_element(LINEBREAK) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("  ") {
				goto l722
			}
			if !p.rules[ruleNormalEndline]() {
				goto l722
			}
			do(52)
			return true
		l722:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 148 Symbol <- (< SpecialChar > { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleSpecialChar]() {
				goto l723
			}
			end = position
			do(53)
			return true
		l723:
			position = position0
			return false
		},
		/* 149 UlOrStarLine <- ((UlLine / StarLine) { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l726
			}
			goto l725
		l726:
			if !p.rules[ruleStarLine]() {
				goto l724
			}
		l725:
			do(54)
			return true
		l724:
			position = position0
			return false
		},
		/* 150 StarLine <- ((&[*] (< '****' '*'* >)) | (&[\t ] (< Spacechar '*'+ &Spacechar >))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l727
				}
				switch p.Buffer[position] {
				case '*':
					begin = position
					if !matchString("****") {
						goto l727
					}
				l729:
					if !matchChar('*') {
						goto l730
					}
					goto l729
				l730:
					end = position
					break
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto l727
					}
					if !matchChar('*') {
						goto l727
					}
				l731:
					if !matchChar('*') {
						goto l732
					}
					goto l731
				l732:
					{
						position733 := position
						if !p.rules[ruleSpacechar]() {
							goto l727
						}
						position = position733
					}
					end = position
					break
				default:
					goto l727
				}
			}
			return true
		l727:
			position = position0
			return false
		},
		/* 151 UlLine <- ((&[_] (< '____' '_'* >)) | (&[\t ] (< Spacechar '_'+ &Spacechar >))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l734
				}
				switch p.Buffer[position] {
				case '_':
					begin = position
					if !matchString("____") {
						goto l734
					}
				l736:
					if !matchChar('_') {
						goto l737
					}
					goto l736
				l737:
					end = position
					break
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto l734
					}
					if !matchChar('_') {
						goto l734
					}
				l738:
					if !matchChar('_') {
						goto l739
					}
					goto l738
				l739:
					{
						position740 := position
						if !p.rules[ruleSpacechar]() {
							goto l734
						}
						position = position740
					}
					end = position
					break
				default:
					goto l734
				}
			}
			return true
		l734:
			position = position0
			return false
		},
		/* 152 Emph <- ((&[_] EmphUl) | (&[*] EmphStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l741
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleEmphUl]() {
						goto l741
					}
					break
				case '*':
					if !p.rules[ruleEmphStar]() {
						goto l741
					}
					break
				default:
					goto l741
				}
			}
			return true
		l741:
			return false
		},
		/* 153 OneStarOpen <- (!StarLine '*' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleStarLine]() {
				goto l744
			}
			goto l743
		l744:
			if !matchChar('*') {
				goto l743
			}
			if !p.rules[ruleSpacechar]() {
				goto l745
			}
			goto l743
		l745:
			if !p.rules[ruleNewline]() {
				goto l746
			}
			goto l743
		l746:
			return true
		l743:
			position = position0
			return false
		},
		/* 154 OneStarClose <- (!Spacechar !Newline Inline !StrongStar '*' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l748
			}
			goto l747
		l748:
			if !p.rules[ruleNewline]() {
				goto l749
			}
			goto l747
		l749:
			if !p.rules[ruleInline]() {
				goto l747
			}
			doarg(yySet, -1)
			if !p.rules[ruleStrongStar]() {
				goto l750
			}
			goto l747
		l750:
			if !matchChar('*') {
				goto l747
			}
			do(55)
			doarg(yyPop, 1)
			return true
		l747:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 155 EmphStar <- (OneStarOpen StartList (!OneStarClose Inline { a = cons(yy, a) })* OneStarClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneStarOpen]() {
				goto l751
			}
			if !p.rules[ruleStartList]() {
				goto l751
			}
			doarg(yySet, -1)
		l752:
			{
				position753, thunkPosition753 := position, thunkPosition
				if !p.rules[ruleOneStarClose]() {
					goto l754
				}
				goto l753
			l754:
				if !p.rules[ruleInline]() {
					goto l753
				}
				do(56)
				goto l752
			l753:
				position, thunkPosition = position753, thunkPosition753
			}
			if !p.rules[ruleOneStarClose]() {
				goto l751
			}
			do(57)
			do(58)
			doarg(yyPop, 1)
			return true
		l751:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 156 OneUlOpen <- (!UlLine '_' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l756
			}
			goto l755
		l756:
			if !matchChar('_') {
				goto l755
			}
			if !p.rules[ruleSpacechar]() {
				goto l757
			}
			goto l755
		l757:
			if !p.rules[ruleNewline]() {
				goto l758
			}
			goto l755
		l758:
			return true
		l755:
			position = position0
			return false
		},
		/* 157 OneUlClose <- (!Spacechar !Newline Inline !StrongUl '_' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l760
			}
			goto l759
		l760:
			if !p.rules[ruleNewline]() {
				goto l761
			}
			goto l759
		l761:
			if !p.rules[ruleInline]() {
				goto l759
			}
			doarg(yySet, -1)
			if !p.rules[ruleStrongUl]() {
				goto l762
			}
			goto l759
		l762:
			if !matchChar('_') {
				goto l759
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l763
			}
			goto l759
		l763:
			do(59)
			doarg(yyPop, 1)
			return true
		l759:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 158 EmphUl <- (OneUlOpen StartList (!OneUlClose Inline { a = cons(yy, a) })* OneUlClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneUlOpen]() {
				goto l764
			}
			if !p.rules[ruleStartList]() {
				goto l764
			}
			doarg(yySet, -1)
		l765:
			{
				position766, thunkPosition766 := position, thunkPosition
				if !p.rules[ruleOneUlClose]() {
					goto l767
				}
				goto l766
			l767:
				if !p.rules[ruleInline]() {
					goto l766
				}
				do(60)
				goto l765
			l766:
				position, thunkPosition = position766, thunkPosition766
			}
			if !p.rules[ruleOneUlClose]() {
				goto l764
			}
			do(61)
			do(62)
			doarg(yyPop, 1)
			return true
		l764:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 159 Strong <- ((&[_] StrongUl) | (&[*] StrongStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l768
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleStrongUl]() {
						goto l768
					}
					break
				case '*':
					if !p.rules[ruleStrongStar]() {
						goto l768
					}
					break
				default:
					goto l768
				}
			}
			return true
		l768:
			return false
		},
		/* 160 TwoStarOpen <- (!StarLine '**' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleStarLine]() {
				goto l771
			}
			goto l770
		l771:
			if !matchString("**") {
				goto l770
			}
			if !p.rules[ruleSpacechar]() {
				goto l772
			}
			goto l770
		l772:
			if !p.rules[ruleNewline]() {
				goto l773
			}
			goto l770
		l773:
			return true
		l770:
			position = position0
			return false
		},
		/* 161 TwoStarClose <- (!Spacechar !Newline Inline '**' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l775
			}
			goto l774
		l775:
			if !p.rules[ruleNewline]() {
				goto l776
			}
			goto l774
		l776:
			if !p.rules[ruleInline]() {
				goto l774
			}
			doarg(yySet, -1)
			if !matchString("**") {
				goto l774
			}
			do(63)
			doarg(yyPop, 1)
			return true
		l774:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 162 StrongStar <- (TwoStarOpen StartList (!TwoStarClose Inline { a = cons(yy, a) })* TwoStarClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoStarOpen]() {
				goto l777
			}
			if !p.rules[ruleStartList]() {
				goto l777
			}
			doarg(yySet, -1)
		l778:
			{
				position779, thunkPosition779 := position, thunkPosition
				if !p.rules[ruleTwoStarClose]() {
					goto l780
				}
				goto l779
			l780:
				if !p.rules[ruleInline]() {
					goto l779
				}
				do(64)
				goto l778
			l779:
				position, thunkPosition = position779, thunkPosition779
			}
			if !p.rules[ruleTwoStarClose]() {
				goto l777
			}
			do(65)
			do(66)
			doarg(yyPop, 1)
			return true
		l777:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 163 TwoUlOpen <- (!UlLine '__' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l782
			}
			goto l781
		l782:
			if !matchString("__") {
				goto l781
			}
			if !p.rules[ruleSpacechar]() {
				goto l783
			}
			goto l781
		l783:
			if !p.rules[ruleNewline]() {
				goto l784
			}
			goto l781
		l784:
			return true
		l781:
			position = position0
			return false
		},
		/* 164 TwoUlClose <- (!Spacechar !Newline Inline '__' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l786
			}
			goto l785
		l786:
			if !p.rules[ruleNewline]() {
				goto l787
			}
			goto l785
		l787:
			if !p.rules[ruleInline]() {
				goto l785
			}
			doarg(yySet, -1)
			if !matchString("__") {
				goto l785
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l788
			}
			goto l785
		l788:
			do(67)
			doarg(yyPop, 1)
			return true
		l785:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 165 StrongUl <- (TwoUlOpen StartList (!TwoUlClose Inline { a = cons(yy, a) })* TwoUlClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoUlOpen]() {
				goto l789
			}
			if !p.rules[ruleStartList]() {
				goto l789
			}
			doarg(yySet, -1)
		l790:
			{
				position791, thunkPosition791 := position, thunkPosition
				if !p.rules[ruleTwoUlClose]() {
					goto l792
				}
				goto l791
			l792:
				if !p.rules[ruleInline]() {
					goto l791
				}
				do(68)
				goto l790
			l791:
				position, thunkPosition = position791, thunkPosition791
			}
			if !p.rules[ruleTwoUlClose]() {
				goto l789
			}
			do(69)
			do(70)
			doarg(yyPop, 1)
			return true
		l789:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 166 Image <- ('!' (ExplicitLink / ReferenceLink) {	if yy.key == LINK {
			yy.key = IMAGE
		} else {
			result := yy
			yy.children = cons(mk_str("!"), result.children)
		}
	}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('!') {
				goto l793
			}
			if !p.rules[ruleExplicitLink]() {
				goto l795
			}
			goto l794
		l795:
			if !p.rules[ruleReferenceLink]() {
				goto l793
			}
		l794:
			do(71)
			return true
		l793:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 167 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() bool {
			if !p.rules[ruleExplicitLink]() {
				goto l798
			}
			goto l797
		l798:
			if !p.rules[ruleReferenceLink]() {
				goto l799
			}
			goto l797
		l799:
			if !p.rules[ruleAutoLink]() {
				goto l796
			}
		l797:
			return true
		l796:
			return false
		},
		/* 168 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() bool {
			if !p.rules[ruleReferenceLinkDouble]() {
				goto l802
			}
			goto l801
		l802:
			if !p.rules[ruleReferenceLinkSingle]() {
				goto l800
			}
		l801:
			return true
		l800:
			return false
		},
		/* 169 ReferenceLinkDouble <- (Label < Spnl > !'[]' Label {
                           if match, found := p.findReference(b.children); found {
                               yy = mk_link(a.children, match.url, match.title);
                               a = nil
                               b = nil
                           } else {
                               result := mk_element(LIST)
                               result.children = cons(mk_str("["), cons(a, cons(mk_str("]"), cons(mk_str(yytext),
                                                   cons(mk_str("["), cons(b, mk_str("]")))))))
                               yy = result
                           }
                       }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleLabel]() {
				goto l803
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto l803
			}
			end = position
			if !matchString("[]") {
				goto l804
			}
			goto l803
		l804:
			if !p.rules[ruleLabel]() {
				goto l803
			}
			doarg(yySet, -2)
			do(72)
			doarg(yyPop, 2)
			return true
		l803:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 170 ReferenceLinkSingle <- (Label < (Spnl '[]')? > {
                           if match, found := p.findReference(a.children); found {
                               yy = mk_link(a.children, match.url, match.title)
                               a = nil
                           } else {
                               result := mk_element(LIST)
                               result.children = cons(mk_str("["), cons(a, cons(mk_str("]"), mk_str(yytext))));
                               yy = result
                           }
                       }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleLabel]() {
				goto l805
			}
			doarg(yySet, -1)
			begin = position
			{
				position806 := position
				if !p.rules[ruleSpnl]() {
					goto l806
				}
				if !matchString("[]") {
					goto l806
				}
				goto l807
			l806:
				position = position806
			}
		l807:
			end = position
			do(73)
			doarg(yyPop, 1)
			return true
		l805:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 171 ExplicitLink <- (Label Spnl '(' Sp Source Spnl Title Sp ')' { yy = mk_link(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleLabel]() {
				goto l808
			}
			doarg(yySet, -2)
			if !p.rules[ruleSpnl]() {
				goto l808
			}
			if !matchChar('(') {
				goto l808
			}
			if !p.rules[ruleSp]() {
				goto l808
			}
			if !p.rules[ruleSource]() {
				goto l808
			}
			doarg(yySet, -3)
			if !p.rules[ruleSpnl]() {
				goto l808
			}
			if !p.rules[ruleTitle]() {
				goto l808
			}
			doarg(yySet, -1)
			if !p.rules[ruleSp]() {
				goto l808
			}
			if !matchChar(')') {
				goto l808
			}
			do(74)
			doarg(yyPop, 3)
			return true
		l808:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 172 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			{
				position810 := position
				if !matchChar('<') {
					goto l811
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l811
				}
				end = position
				if !matchChar('>') {
					goto l811
				}
				goto l810
			l811:
				position = position810
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l809
				}
				end = position
			}
		l810:
			do(75)
			return true
		l809:
			position = position0
			return false
		},
		/* 173 SourceContents <- (((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* / '') */
		func() bool {
		l815:
			{
				position816 := position
				if position == len(p.Buffer) {
					goto l818
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l818
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l818
					}
				}
			l819:
				if position == len(p.Buffer) {
					goto l820
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l820
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l820
					}
				}
				goto l819
			l820:
				goto l817
			l818:
				if !matchChar('(') {
					goto l816
				}
				if !p.rules[ruleSourceContents]() {
					goto l816
				}
				if !matchChar(')') {
					goto l816
				}
			l817:
				goto l815
			l816:
				position = position816
			}
			goto l813
		l813:
			return true
		},
		/* 174 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = mk_str(yytext) }) */
		func() bool {
			if !p.rules[ruleTitleSingle]() {
				goto l823
			}
			goto l822
		l823:
			if !p.rules[ruleTitleDouble]() {
				goto l824
			}
			goto l822
		l824:
			begin = position
			end = position
		l822:
			do(76)
			return true
		},
		/* 175 TitleSingle <- ('\'' < (!('\'' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '\'') */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l825
			}
			begin = position
		l826:
			{
				position827 := position
				{
					position828 := position
					if !matchChar('\'') {
						goto l828
					}
					if !p.rules[ruleSp]() {
						goto l828
					}
					{
						if position == len(p.Buffer) {
							goto l828
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l828
							}
							break
						default:
							goto l828
						}
					}
					goto l827
				l828:
					position = position828
				}
				if !matchDot() {
					goto l827
				}
				goto l826
			l827:
				position = position827
			}
			end = position
			if !matchChar('\'') {
				goto l825
			}
			return true
		l825:
			position = position0
			return false
		},
		/* 176 TitleDouble <- ('"' < (!('"' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '"') */
		func() bool {
			position0 := position
			if !matchChar('"') {
				goto l830
			}
			begin = position
		l831:
			{
				position832 := position
				{
					position833 := position
					if !matchChar('"') {
						goto l833
					}
					if !p.rules[ruleSp]() {
						goto l833
					}
					{
						if position == len(p.Buffer) {
							goto l833
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l833
							}
							break
						default:
							goto l833
						}
					}
					goto l832
				l833:
					position = position833
				}
				if !matchDot() {
					goto l832
				}
				goto l831
			l832:
				position = position832
			}
			end = position
			if !matchChar('"') {
				goto l830
			}
			return true
		l830:
			position = position0
			return false
		},
		/* 177 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() bool {
			if !p.rules[ruleAutoLinkUrl]() {
				goto l837
			}
			goto l836
		l837:
			if !p.rules[ruleAutoLinkEmail]() {
				goto l835
			}
		l836:
			return true
		l835:
			return false
		},
		/* 178 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = mk_link(mk_str(yytext), yytext, "") }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l838
			}
			begin = position
			if !matchClass(2) {
				goto l838
			}
		l839:
			if !matchClass(2) {
				goto l840
			}
			goto l839
		l840:
			if !matchString("://") {
				goto l838
			}
			if !p.rules[ruleNewline]() {
				goto l843
			}
			goto l838
		l843:
			if peekChar('>') {
				goto l838
			}
			if !matchDot() {
				goto l838
			}
		l841:
			{
				position842 := position
				if !p.rules[ruleNewline]() {
					goto l844
				}
				goto l842
			l844:
				if peekChar('>') {
					goto l842
				}
				if !matchDot() {
					goto l842
				}
				goto l841
			l842:
				position = position842
			}
			end = position
			if !matchChar('>') {
				goto l838
			}
			do(77)
			return true
		l838:
			position = position0
			return false
		},
		/* 179 AutoLinkEmail <- ('<' < [-A-Za-z0-9+_]+ '@' (!Newline !'>' .)+ > '>' {
                    yy = mk_link(mk_str(yytext), "mailto:"+yytext, "")
                }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l845
			}
			begin = position
			if !matchClass(3) {
				goto l845
			}
		l846:
			if !matchClass(3) {
				goto l847
			}
			goto l846
		l847:
			if !matchChar('@') {
				goto l845
			}
			if !p.rules[ruleNewline]() {
				goto l850
			}
			goto l845
		l850:
			if peekChar('>') {
				goto l845
			}
			if !matchDot() {
				goto l845
			}
		l848:
			{
				position849 := position
				if !p.rules[ruleNewline]() {
					goto l851
				}
				goto l849
			l851:
				if peekChar('>') {
					goto l849
				}
				if !matchDot() {
					goto l849
				}
				goto l848
			l849:
				position = position849
			}
			end = position
			if !matchChar('>') {
				goto l845
			}
			do(78)
			return true
		l845:
			position = position0
			return false
		},
		/* 180 Reference <- (NonindentSpace !'[]' Label ':' Spnl RefSrc Spnl RefTitle BlankLine* { yy = mk_link(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleNonindentSpace]() {
				goto l852
			}
			if !matchString("[]") {
				goto l853
			}
			goto l852
		l853:
			if !p.rules[ruleLabel]() {
				goto l852
			}
			doarg(yySet, -2)
			if !matchChar(':') {
				goto l852
			}
			if !p.rules[ruleSpnl]() {
				goto l852
			}
			if !p.rules[ruleRefSrc]() {
				goto l852
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l852
			}
			if !p.rules[ruleRefTitle]() {
				goto l852
			}
			doarg(yySet, -3)
		l854:
			if !p.rules[ruleBlankLine]() {
				goto l855
			}
			goto l854
		l855:
			do(79)
			doarg(yyPop, 3)
			return true
		l852:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 181 Label <- ('[' ((!'^' &{p.extension.Notes}) / (&. &{!p.extension.Notes})) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l856
			}
			if peekChar('^') {
				goto l858
			}
			if !(p.extension.Notes) {
				goto l858
			}
			goto l857
		l858:
			if !(position < len(p.Buffer)) {
				goto l856
			}
			if !(!p.extension.Notes) {
				goto l856
			}
		l857:
			if !p.rules[ruleStartList]() {
				goto l856
			}
			doarg(yySet, -1)
		l859:
			{
				position860 := position
				if peekChar(']') {
					goto l860
				}
				if !p.rules[ruleInline]() {
					goto l860
				}
				do(80)
				goto l859
			l860:
				position = position860
			}
			if !matchChar(']') {
				goto l856
			}
			do(81)
			doarg(yyPop, 1)
			return true
		l856:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 182 RefSrc <- (< Nonspacechar+ > { yy = mk_str(yytext)
           yy.key = HTML }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto l861
			}
		l862:
			if !p.rules[ruleNonspacechar]() {
				goto l863
			}
			goto l862
		l863:
			end = position
			do(82)
			return true
		l861:
			position = position0
			return false
		},
		/* 183 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRefTitleSingle]() {
				goto l866
			}
			goto l865
		l866:
			if !p.rules[ruleRefTitleDouble]() {
				goto l867
			}
			goto l865
		l867:
			if !p.rules[ruleRefTitleParens]() {
				goto l868
			}
			goto l865
		l868:
			if !p.rules[ruleEmptyTitle]() {
				goto l864
			}
		l865:
			do(83)
			return true
		l864:
			position = position0
			return false
		},
		/* 184 EmptyTitle <- (< '' >) */
		func() bool {
			begin = position
			end = position
			return true
		},
		/* 185 RefTitleSingle <- ('\'' < (!((&[\'] ('\'' Sp Newline)) | (&[\n\r] Newline)) .)* > '\'') */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l870
			}
			begin = position
		l871:
			{
				position872 := position
				{
					position873 := position
					{
						if position == len(p.Buffer) {
							goto l873
						}
						switch p.Buffer[position] {
						case '\'':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l873
							}
							if !p.rules[ruleNewline]() {
								goto l873
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l873
							}
							break
						default:
							goto l873
						}
					}
					goto l872
				l873:
					position = position873
				}
				if !matchDot() {
					goto l872
				}
				goto l871
			l872:
				position = position872
			}
			end = position
			if !matchChar('\'') {
				goto l870
			}
			return true
		l870:
			position = position0
			return false
		},
		/* 186 RefTitleDouble <- ('"' < (!((&[\"] ('"' Sp Newline)) | (&[\n\r] Newline)) .)* > '"') */
		func() bool {
			position0 := position
			if !matchChar('"') {
				goto l875
			}
			begin = position
		l876:
			{
				position877 := position
				{
					position878 := position
					{
						if position == len(p.Buffer) {
							goto l878
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l878
							}
							if !p.rules[ruleNewline]() {
								goto l878
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l878
							}
							break
						default:
							goto l878
						}
					}
					goto l877
				l878:
					position = position878
				}
				if !matchDot() {
					goto l877
				}
				goto l876
			l877:
				position = position877
			}
			end = position
			if !matchChar('"') {
				goto l875
			}
			return true
		l875:
			position = position0
			return false
		},
		/* 187 RefTitleParens <- ('(' < (!((&[)] (')' Sp Newline)) | (&[\n\r] Newline)) .)* > ')') */
		func() bool {
			position0 := position
			if !matchChar('(') {
				goto l880
			}
			begin = position
		l881:
			{
				position882 := position
				{
					position883 := position
					{
						if position == len(p.Buffer) {
							goto l883
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l883
							}
							if !p.rules[ruleNewline]() {
								goto l883
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l883
							}
							break
						default:
							goto l883
						}
					}
					goto l882
				l883:
					position = position883
				}
				if !matchDot() {
					goto l882
				}
				goto l881
			l882:
				position = position882
			}
			end = position
			if !matchChar(')') {
				goto l880
			}
			return true
		l880:
			position = position0
			return false
		},
		/* 188 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l885
			}
			doarg(yySet, -2)
		l886:
			{
				position887, thunkPosition887 := position, thunkPosition
				{
					position888, thunkPosition888 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto l889
					}
					doarg(yySet, -1)
					do(84)
					goto l888
				l889:
					position, thunkPosition = position888, thunkPosition888
					if !p.rules[ruleSkipBlock]() {
						goto l887
					}
				}
			l888:
				goto l886
			l887:
				position, thunkPosition = position887, thunkPosition887
			}
			do(85)
			if !(commit(thunkPosition0)) {
				goto l885
			}
			doarg(yyPop, 2)
			return true
		l885:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 189 Ticks1 <- ('`' !'`') */
		func() bool {
			position0 := position
			if !matchChar('`') {
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
		/* 190 Ticks2 <- ('``' !'`') */
		func() bool {
			position0 := position
			if !matchString("``") {
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
		/* 191 Ticks3 <- ('```' !'`') */
		func() bool {
			position0 := position
			if !matchString("```") {
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
		/* 192 Ticks4 <- ('````' !'`') */
		func() bool {
			position0 := position
			if !matchString("````") {
				goto l893
			}
			if peekChar('`') {
				goto l893
			}
			return true
		l893:
			position = position0
			return false
		},
		/* 193 Ticks5 <- ('`````' !'`') */
		func() bool {
			position0 := position
			if !matchString("`````") {
				goto l894
			}
			if peekChar('`') {
				goto l894
			}
			return true
		l894:
			position = position0
			return false
		},
		/* 194 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks1 '`'+)) | (&[\t\n\r ] (!(Sp Ticks1) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks2 '`'+)) | (&[\t\n\r ] (!(Sp Ticks2) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks3 '`'+)) | (&[\t\n\r ] (!(Sp Ticks3) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks4 '`'+)) | (&[\t\n\r ] (!(Sp Ticks4) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks5 '`'+)) | (&[\t\n\r ] (!(Sp Ticks5) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks5)) { yy = mk_str(yytext); yy.key = CODE }) */
		func() bool {
			position0 := position
			{
				position896 := position
				if !p.rules[ruleTicks1]() {
					goto l897
				}
				if !p.rules[ruleSp]() {
					goto l897
				}
				begin = position
				if peekChar('`') {
					goto l901
				}
				if !p.rules[ruleNonspacechar]() {
					goto l901
				}
			l902:
				if peekChar('`') {
					goto l903
				}
				if !p.rules[ruleNonspacechar]() {
					goto l903
				}
				goto l902
			l903:
				goto l900
			l901:
				{
					if position == len(p.Buffer) {
						goto l897
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks1]() {
							goto l905
						}
						goto l897
					l905:
						if !matchChar('`') {
							goto l897
						}
					l906:
						if !matchChar('`') {
							goto l907
						}
						goto l906
					l907:
						break
					default:
						{
							position908 := position
							if !p.rules[ruleSp]() {
								goto l908
							}
							if !p.rules[ruleTicks1]() {
								goto l908
							}
							goto l897
						l908:
							position = position908
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
									goto l910
								}
								goto l897
							l910:
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
			l900:
			l898:
				{
					position899 := position
					if peekChar('`') {
						goto l912
					}
					if !p.rules[ruleNonspacechar]() {
						goto l912
					}
				l913:
					if peekChar('`') {
						goto l914
					}
					if !p.rules[ruleNonspacechar]() {
						goto l914
					}
					goto l913
				l914:
					goto l911
				l912:
					{
						if position == len(p.Buffer) {
							goto l899
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks1]() {
								goto l916
							}
							goto l899
						l916:
							if !matchChar('`') {
								goto l899
							}
						l917:
							if !matchChar('`') {
								goto l918
							}
							goto l917
						l918:
							break
						default:
							{
								position919 := position
								if !p.rules[ruleSp]() {
									goto l919
								}
								if !p.rules[ruleTicks1]() {
									goto l919
								}
								goto l899
							l919:
								position = position919
							}
							{
								if position == len(p.Buffer) {
									goto l899
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l899
									}
									if !p.rules[ruleBlankLine]() {
										goto l921
									}
									goto l899
								l921:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l899
									}
									break
								default:
									goto l899
								}
							}
						}
					}
				l911:
					goto l898
				l899:
					position = position899
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l897
				}
				if !p.rules[ruleTicks1]() {
					goto l897
				}
				goto l896
			l897:
				position = position896
				if !p.rules[ruleTicks2]() {
					goto l922
				}
				if !p.rules[ruleSp]() {
					goto l922
				}
				begin = position
				if peekChar('`') {
					goto l926
				}
				if !p.rules[ruleNonspacechar]() {
					goto l926
				}
			l927:
				if peekChar('`') {
					goto l928
				}
				if !p.rules[ruleNonspacechar]() {
					goto l928
				}
				goto l927
			l928:
				goto l925
			l926:
				{
					if position == len(p.Buffer) {
						goto l922
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks2]() {
							goto l930
						}
						goto l922
					l930:
						if !matchChar('`') {
							goto l922
						}
					l931:
						if !matchChar('`') {
							goto l932
						}
						goto l931
					l932:
						break
					default:
						{
							position933 := position
							if !p.rules[ruleSp]() {
								goto l933
							}
							if !p.rules[ruleTicks2]() {
								goto l933
							}
							goto l922
						l933:
							position = position933
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
									goto l935
								}
								goto l922
							l935:
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
			l925:
			l923:
				{
					position924 := position
					if peekChar('`') {
						goto l937
					}
					if !p.rules[ruleNonspacechar]() {
						goto l937
					}
				l938:
					if peekChar('`') {
						goto l939
					}
					if !p.rules[ruleNonspacechar]() {
						goto l939
					}
					goto l938
				l939:
					goto l936
				l937:
					{
						if position == len(p.Buffer) {
							goto l924
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks2]() {
								goto l941
							}
							goto l924
						l941:
							if !matchChar('`') {
								goto l924
							}
						l942:
							if !matchChar('`') {
								goto l943
							}
							goto l942
						l943:
							break
						default:
							{
								position944 := position
								if !p.rules[ruleSp]() {
									goto l944
								}
								if !p.rules[ruleTicks2]() {
									goto l944
								}
								goto l924
							l944:
								position = position944
							}
							{
								if position == len(p.Buffer) {
									goto l924
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l924
									}
									if !p.rules[ruleBlankLine]() {
										goto l946
									}
									goto l924
								l946:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l924
									}
									break
								default:
									goto l924
								}
							}
						}
					}
				l936:
					goto l923
				l924:
					position = position924
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l922
				}
				if !p.rules[ruleTicks2]() {
					goto l922
				}
				goto l896
			l922:
				position = position896
				if !p.rules[ruleTicks3]() {
					goto l947
				}
				if !p.rules[ruleSp]() {
					goto l947
				}
				begin = position
				if peekChar('`') {
					goto l951
				}
				if !p.rules[ruleNonspacechar]() {
					goto l951
				}
			l952:
				if peekChar('`') {
					goto l953
				}
				if !p.rules[ruleNonspacechar]() {
					goto l953
				}
				goto l952
			l953:
				goto l950
			l951:
				{
					if position == len(p.Buffer) {
						goto l947
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks3]() {
							goto l955
						}
						goto l947
					l955:
						if !matchChar('`') {
							goto l947
						}
					l956:
						if !matchChar('`') {
							goto l957
						}
						goto l956
					l957:
						break
					default:
						{
							position958 := position
							if !p.rules[ruleSp]() {
								goto l958
							}
							if !p.rules[ruleTicks3]() {
								goto l958
							}
							goto l947
						l958:
							position = position958
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
									goto l960
								}
								goto l947
							l960:
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
			l950:
			l948:
				{
					position949 := position
					if peekChar('`') {
						goto l962
					}
					if !p.rules[ruleNonspacechar]() {
						goto l962
					}
				l963:
					if peekChar('`') {
						goto l964
					}
					if !p.rules[ruleNonspacechar]() {
						goto l964
					}
					goto l963
				l964:
					goto l961
				l962:
					{
						if position == len(p.Buffer) {
							goto l949
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks3]() {
								goto l966
							}
							goto l949
						l966:
							if !matchChar('`') {
								goto l949
							}
						l967:
							if !matchChar('`') {
								goto l968
							}
							goto l967
						l968:
							break
						default:
							{
								position969 := position
								if !p.rules[ruleSp]() {
									goto l969
								}
								if !p.rules[ruleTicks3]() {
									goto l969
								}
								goto l949
							l969:
								position = position969
							}
							{
								if position == len(p.Buffer) {
									goto l949
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l949
									}
									if !p.rules[ruleBlankLine]() {
										goto l971
									}
									goto l949
								l971:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l949
									}
									break
								default:
									goto l949
								}
							}
						}
					}
				l961:
					goto l948
				l949:
					position = position949
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l947
				}
				if !p.rules[ruleTicks3]() {
					goto l947
				}
				goto l896
			l947:
				position = position896
				if !p.rules[ruleTicks4]() {
					goto l972
				}
				if !p.rules[ruleSp]() {
					goto l972
				}
				begin = position
				if peekChar('`') {
					goto l976
				}
				if !p.rules[ruleNonspacechar]() {
					goto l976
				}
			l977:
				if peekChar('`') {
					goto l978
				}
				if !p.rules[ruleNonspacechar]() {
					goto l978
				}
				goto l977
			l978:
				goto l975
			l976:
				{
					if position == len(p.Buffer) {
						goto l972
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks4]() {
							goto l980
						}
						goto l972
					l980:
						if !matchChar('`') {
							goto l972
						}
					l981:
						if !matchChar('`') {
							goto l982
						}
						goto l981
					l982:
						break
					default:
						{
							position983 := position
							if !p.rules[ruleSp]() {
								goto l983
							}
							if !p.rules[ruleTicks4]() {
								goto l983
							}
							goto l972
						l983:
							position = position983
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
									goto l985
								}
								goto l972
							l985:
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
			l975:
			l973:
				{
					position974 := position
					if peekChar('`') {
						goto l987
					}
					if !p.rules[ruleNonspacechar]() {
						goto l987
					}
				l988:
					if peekChar('`') {
						goto l989
					}
					if !p.rules[ruleNonspacechar]() {
						goto l989
					}
					goto l988
				l989:
					goto l986
				l987:
					{
						if position == len(p.Buffer) {
							goto l974
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks4]() {
								goto l991
							}
							goto l974
						l991:
							if !matchChar('`') {
								goto l974
							}
						l992:
							if !matchChar('`') {
								goto l993
							}
							goto l992
						l993:
							break
						default:
							{
								position994 := position
								if !p.rules[ruleSp]() {
									goto l994
								}
								if !p.rules[ruleTicks4]() {
									goto l994
								}
								goto l974
							l994:
								position = position994
							}
							{
								if position == len(p.Buffer) {
									goto l974
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l974
									}
									if !p.rules[ruleBlankLine]() {
										goto l996
									}
									goto l974
								l996:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l974
									}
									break
								default:
									goto l974
								}
							}
						}
					}
				l986:
					goto l973
				l974:
					position = position974
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l972
				}
				if !p.rules[ruleTicks4]() {
					goto l972
				}
				goto l896
			l972:
				position = position896
				if !p.rules[ruleTicks5]() {
					goto l895
				}
				if !p.rules[ruleSp]() {
					goto l895
				}
				begin = position
				if peekChar('`') {
					goto l1000
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1000
				}
			l1001:
				if peekChar('`') {
					goto l1002
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1002
				}
				goto l1001
			l1002:
				goto l999
			l1000:
				{
					if position == len(p.Buffer) {
						goto l895
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks5]() {
							goto l1004
						}
						goto l895
					l1004:
						if !matchChar('`') {
							goto l895
						}
					l1005:
						if !matchChar('`') {
							goto l1006
						}
						goto l1005
					l1006:
						break
					default:
						{
							position1007 := position
							if !p.rules[ruleSp]() {
								goto l1007
							}
							if !p.rules[ruleTicks5]() {
								goto l1007
							}
							goto l895
						l1007:
							position = position1007
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
									goto l1009
								}
								goto l895
							l1009:
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
			l999:
			l997:
				{
					position998 := position
					if peekChar('`') {
						goto l1011
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1011
					}
				l1012:
					if peekChar('`') {
						goto l1013
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1013
					}
					goto l1012
				l1013:
					goto l1010
				l1011:
					{
						if position == len(p.Buffer) {
							goto l998
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks5]() {
								goto l1015
							}
							goto l998
						l1015:
							if !matchChar('`') {
								goto l998
							}
						l1016:
							if !matchChar('`') {
								goto l1017
							}
							goto l1016
						l1017:
							break
						default:
							{
								position1018 := position
								if !p.rules[ruleSp]() {
									goto l1018
								}
								if !p.rules[ruleTicks5]() {
									goto l1018
								}
								goto l998
							l1018:
								position = position1018
							}
							{
								if position == len(p.Buffer) {
									goto l998
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l998
									}
									if !p.rules[ruleBlankLine]() {
										goto l1020
									}
									goto l998
								l1020:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l998
									}
									break
								default:
									goto l998
								}
							}
						}
					}
				l1010:
					goto l997
				l998:
					position = position998
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l895
				}
				if !p.rules[ruleTicks5]() {
					goto l895
				}
			}
		l896:
			do(86)
			return true
		l895:
			position = position0
			return false
		},
		/* 195 RawHtml <- (< (HtmlComment / HtmlTag) > {   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTML
                }
            }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleHtmlComment]() {
				goto l1023
			}
			goto l1022
		l1023:
			if !p.rules[ruleHtmlTag]() {
				goto l1021
			}
		l1022:
			end = position
			do(87)
			return true
		l1021:
			position = position0
			return false
		},
		/* 196 BlankLine <- (Sp Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1024
			}
			if !p.rules[ruleNewline]() {
				goto l1024
			}
			return true
		l1024:
			position = position0
			return false
		},
		/* 197 Quoted <- ((&[\'] ('\'' (!'\'' .)* '\'')) | (&[\"] ('"' (!'"' .)* '"'))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1025
				}
				switch p.Buffer[position] {
				case '\'':
					position++ // matchChar
				l1027:
					if position == len(p.Buffer) {
						goto l1028
					}
					switch p.Buffer[position] {
					case '\'':
						goto l1028
					default:
						position++
					}
					goto l1027
				l1028:
					if !matchChar('\'') {
						goto l1025
					}
					break
				case '"':
					position++ // matchChar
				l1029:
					if position == len(p.Buffer) {
						goto l1030
					}
					switch p.Buffer[position] {
					case '"':
						goto l1030
					default:
						position++
					}
					goto l1029
				l1030:
					if !matchChar('"') {
						goto l1025
					}
					break
				default:
					goto l1025
				}
			}
			return true
		l1025:
			position = position0
			return false
		},
		/* 198 HtmlAttribute <- (((&[\-] '-') | (&[0-9A-Za-z] [A-Za-z0-9]))+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1031
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				default:
					if !matchClass(6) {
						goto l1031
					}
				}
			}
		l1032:
			{
				if position == len(p.Buffer) {
					goto l1033
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				default:
					if !matchClass(6) {
						goto l1033
					}
				}
			}
			goto l1032
		l1033:
			if !p.rules[ruleSpnl]() {
				goto l1031
			}
			{
				position1036 := position
				if !matchChar('=') {
					goto l1036
				}
				if !p.rules[ruleSpnl]() {
					goto l1036
				}
				if !p.rules[ruleQuoted]() {
					goto l1039
				}
				goto l1038
			l1039:
				if peekChar('>') {
					goto l1036
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1036
				}
			l1040:
				if peekChar('>') {
					goto l1041
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1041
				}
				goto l1040
			l1041:
			l1038:
				goto l1037
			l1036:
				position = position1036
			}
		l1037:
			if !p.rules[ruleSpnl]() {
				goto l1031
			}
			return true
		l1031:
			position = position0
			return false
		},
		/* 199 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
		func() bool {
			position0 := position
			if !matchString("<!--") {
				goto l1042
			}
		l1043:
			{
				position1044 := position
				if !matchString("-->") {
					goto l1045
				}
				goto l1044
			l1045:
				if !matchDot() {
					goto l1044
				}
				goto l1043
			l1044:
				position = position1044
			}
			if !matchString("-->") {
				goto l1042
			}
			return true
		l1042:
			position = position0
			return false
		},
		/* 200 HtmlTag <- ('<' Spnl '/'? [A-Za-z0-9]+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l1046
			}
			if !p.rules[ruleSpnl]() {
				goto l1046
			}
			matchChar('/')
			if !matchClass(6) {
				goto l1046
			}
		l1047:
			if !matchClass(6) {
				goto l1048
			}
			goto l1047
		l1048:
			if !p.rules[ruleSpnl]() {
				goto l1046
			}
		l1049:
			if !p.rules[ruleHtmlAttribute]() {
				goto l1050
			}
			goto l1049
		l1050:
			matchChar('/')
			if !p.rules[ruleSpnl]() {
				goto l1046
			}
			if !matchChar('>') {
				goto l1046
			}
			return true
		l1046:
			position = position0
			return false
		},
		/* 201 Eof <- !. */
		func() bool {
			if (position < len(p.Buffer)) {
				goto l1051
			}
			return true
		l1051:
			return false
		},
		/* 202 Spacechar <- ((&[\t] '\t') | (&[ ] ' ')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1052
				}
				switch p.Buffer[position] {
				case '\t':
					position++ // matchChar
					break
				case ' ':
					position++ // matchChar
					break
				default:
					goto l1052
				}
			}
			return true
		l1052:
			return false
		},
		/* 203 Nonspacechar <- (!Spacechar !Newline .) */
		func() bool {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto l1055
			}
			goto l1054
		l1055:
			if !p.rules[ruleNewline]() {
				goto l1056
			}
			goto l1054
		l1056:
			if !matchDot() {
				goto l1054
			}
			return true
		l1054:
			position = position0
			return false
		},
		/* 204 Newline <- ((&[\r] ('\r' '\n'?)) | (&[\n] '\n')) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1057
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
					goto l1057
				}
			}
			return true
		l1057:
			position = position0
			return false
		},
		/* 205 Sp <- Spacechar* */
		func() bool {
		l1060:
			if !p.rules[ruleSpacechar]() {
				goto l1061
			}
			goto l1060
		l1061:
			return true
		},
		/* 206 Spnl <- (Sp (Newline Sp)?) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1062
			}
			{
				position1063 := position
				if !p.rules[ruleNewline]() {
					goto l1063
				}
				if !p.rules[ruleSp]() {
					goto l1063
				}
				goto l1064
			l1063:
				position = position1063
			}
		l1064:
			return true
		l1062:
			position = position0
			return false
		},
		/* 207 SpecialChar <- ((&[\\] '\\') | (&[#] '#') | (&[!] '!') | (&[<] '<') | (&[\]] ']') | (&[\[] '[') | (&[&] '&') | (&[`] '`') | (&[_] '_') | (&[*] '*') | (&[\"\'\-.^] ExtendedSpecialChar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1065
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
						goto l1065
					}
				}
			}
			return true
		l1065:
			return false
		},
		/* 208 NormalChar <- (!((&[\n\r] Newline) | (&[\t ] Spacechar) | (&[!-#&\'*\-.<\[-`] SpecialChar)) .) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1068
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l1068
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l1068
					}
					break
				default:
					if !p.rules[ruleSpecialChar]() {
						goto l1068
					}
				}
			}
			goto l1067
		l1068:
			if !matchDot() {
				goto l1067
			}
			return true
		l1067:
			position = position0
			return false
		},
		/* 209 NonAlphanumeric <- [\000-\057\072-\100\133-\140\173-\177] */
		func() bool {
			if !matchClass(4) {
				goto l1070
			}
			return true
		l1070:
			return false
		},
		/* 210 Alphanumeric <- ((&[\377] '\377') | (&[\376] '\376') | (&[\375] '\375') | (&[\374] '\374') | (&[\373] '\373') | (&[\372] '\372') | (&[\371] '\371') | (&[\370] '\370') | (&[\367] '\367') | (&[\366] '\366') | (&[\365] '\365') | (&[\364] '\364') | (&[\363] '\363') | (&[\362] '\362') | (&[\361] '\361') | (&[\360] '\360') | (&[\357] '\357') | (&[\356] '\356') | (&[\355] '\355') | (&[\354] '\354') | (&[\353] '\353') | (&[\352] '\352') | (&[\351] '\351') | (&[\350] '\350') | (&[\347] '\347') | (&[\346] '\346') | (&[\345] '\345') | (&[\344] '\344') | (&[\343] '\343') | (&[\342] '\342') | (&[\341] '\341') | (&[\340] '\340') | (&[\337] '\337') | (&[\336] '\336') | (&[\335] '\335') | (&[\334] '\334') | (&[\333] '\333') | (&[\332] '\332') | (&[\331] '\331') | (&[\330] '\330') | (&[\327] '\327') | (&[\326] '\326') | (&[\325] '\325') | (&[\324] '\324') | (&[\323] '\323') | (&[\322] '\322') | (&[\321] '\321') | (&[\320] '\320') | (&[\317] '\317') | (&[\316] '\316') | (&[\315] '\315') | (&[\314] '\314') | (&[\313] '\313') | (&[\312] '\312') | (&[\311] '\311') | (&[\310] '\310') | (&[\307] '\307') | (&[\306] '\306') | (&[\305] '\305') | (&[\304] '\304') | (&[\303] '\303') | (&[\302] '\302') | (&[\301] '\301') | (&[\300] '\300') | (&[\277] '\277') | (&[\276] '\276') | (&[\275] '\275') | (&[\274] '\274') | (&[\273] '\273') | (&[\272] '\272') | (&[\271] '\271') | (&[\270] '\270') | (&[\267] '\267') | (&[\266] '\266') | (&[\265] '\265') | (&[\264] '\264') | (&[\263] '\263') | (&[\262] '\262') | (&[\261] '\261') | (&[\260] '\260') | (&[\257] '\257') | (&[\256] '\256') | (&[\255] '\255') | (&[\254] '\254') | (&[\253] '\253') | (&[\252] '\252') | (&[\251] '\251') | (&[\250] '\250') | (&[\247] '\247') | (&[\246] '\246') | (&[\245] '\245') | (&[\244] '\244') | (&[\243] '\243') | (&[\242] '\242') | (&[\241] '\241') | (&[\240] '\240') | (&[\237] '\237') | (&[\236] '\236') | (&[\235] '\235') | (&[\234] '\234') | (&[\233] '\233') | (&[\232] '\232') | (&[\231] '\231') | (&[\230] '\230') | (&[\227] '\227') | (&[\226] '\226') | (&[\225] '\225') | (&[\224] '\224') | (&[\223] '\223') | (&[\222] '\222') | (&[\221] '\221') | (&[\220] '\220') | (&[\217] '\217') | (&[\216] '\216') | (&[\215] '\215') | (&[\214] '\214') | (&[\213] '\213') | (&[\212] '\212') | (&[\211] '\211') | (&[\210] '\210') | (&[\207] '\207') | (&[\206] '\206') | (&[\205] '\205') | (&[\204] '\204') | (&[\203] '\203') | (&[\202] '\202') | (&[\201] '\201') | (&[\200] '\200') | (&[0-9A-Za-z] [0-9A-Za-z])) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1071
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
					if !matchClass(5) {
						goto l1071
					}
				}
			}
			return true
		l1071:
			return false
		},
		/* 211 AlphanumericAscii <- [A-Za-z0-9] */
		func() bool {
			if !matchClass(6) {
				goto l1073
			}
			return true
		l1073:
			return false
		},
		/* 212 Digit <- [0-9] */
		func() bool {
			if !matchClass(0) {
				goto l1074
			}
			return true
		l1074:
			return false
		},
		/* 213 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1075
			}
			if !matchChar('#') {
				goto l1075
			}
			if !matchClass(7) {
				goto l1075
			}
			if !matchClass(8) {
				goto l1075
			}
		l1076:
			if !matchClass(8) {
				goto l1077
			}
			goto l1076
		l1077:
			if !matchChar(';') {
				goto l1075
			}
			end = position
			return true
		l1075:
			position = position0
			return false
		},
		/* 214 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1078
			}
			if !matchChar('#') {
				goto l1078
			}
			if !matchClass(0) {
				goto l1078
			}
		l1079:
			if !matchClass(0) {
				goto l1080
			}
			goto l1079
		l1080:
			end = position
			if !matchChar(';') {
				goto l1078
			}
			end = position
			return true
		l1078:
			position = position0
			return false
		},
		/* 215 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1081
			}
			if !matchClass(6) {
				goto l1081
			}
		l1082:
			if !matchClass(6) {
				goto l1083
			}
			goto l1082
		l1083:
			if !matchChar(';') {
				goto l1081
			}
			end = position
			return true
		l1081:
			position = position0
			return false
		},
		/* 216 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() bool {
			if !matchString("   ") {
				goto l1086
			}
			goto l1085
		l1086:
			if !matchString("  ") {
				goto l1087
			}
			goto l1085
		l1087:
			if !matchChar(' ') {
				goto l1088
			}
			goto l1085
		l1088:
		l1085:
			return true
		},
		/* 217 Indent <- ((&[ ] '    ') | (&[\t] '\t')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1089
				}
				switch p.Buffer[position] {
				case ' ':
					position++
					if !matchString("   ") {
						goto l1089
					}
					break
				case '\t':
					position++ // matchChar
					break
				default:
					goto l1089
				}
			}
			return true
		l1089:
			return false
		},
		/* 218 IndentedLine <- (Indent Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1091
			}
			if !p.rules[ruleLine]() {
				goto l1091
			}
			return true
		l1091:
			position = position0
			return false
		},
		/* 219 OptionallyIndentedLine <- (Indent? Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1093
			}
		l1093:
			if !p.rules[ruleLine]() {
				goto l1092
			}
			return true
		l1092:
			position = position0
			return false
		},
		/* 220 StartList <- (&. { yy = nil }) */
		func() bool {
			if !(position < len(p.Buffer)) {
				goto l1095
			}
			do(88)
			return true
		l1095:
			return false
		},
		/* 221 Line <- (RawLine { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRawLine]() {
				goto l1096
			}
			do(89)
			return true
		l1096:
			position = position0
			return false
		},
		/* 222 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > !.)) */
		func() bool {
			position0 := position
			{
				position1098 := position
				begin = position
			l1100:
				if position == len(p.Buffer) {
					goto l1101
				}
				switch p.Buffer[position] {
				case '\r', '\n':
					goto l1101
				default:
					position++
				}
				goto l1100
			l1101:
				if !p.rules[ruleNewline]() {
					goto l1099
				}
				end = position
				goto l1098
			l1099:
				position = position1098
				begin = position
				if !matchDot() {
					goto l1097
				}
			l1102:
				if !matchDot() {
					goto l1103
				}
				goto l1102
			l1103:
				end = position
				if (position < len(p.Buffer)) {
					goto l1097
				}
			}
		l1098:
			return true
		l1097:
			position = position0
			return false
		},
		/* 223 SkipBlock <- (((!BlankLine RawLine)+ BlankLine*) / BlankLine+) */
		func() bool {
			position0 := position
			{
				position1105 := position
				if !p.rules[ruleBlankLine]() {
					goto l1109
				}
				goto l1106
			l1109:
				if !p.rules[ruleRawLine]() {
					goto l1106
				}
			l1107:
				{
					position1108 := position
					if !p.rules[ruleBlankLine]() {
						goto l1110
					}
					goto l1108
				l1110:
					if !p.rules[ruleRawLine]() {
						goto l1108
					}
					goto l1107
				l1108:
					position = position1108
				}
			l1111:
				if !p.rules[ruleBlankLine]() {
					goto l1112
				}
				goto l1111
			l1112:
				goto l1105
			l1106:
				position = position1105
				if !p.rules[ruleBlankLine]() {
					goto l1104
				}
			l1113:
				if !p.rules[ruleBlankLine]() {
					goto l1114
				}
				goto l1113
			l1114:
			}
		l1105:
			return true
		l1104:
			position = position0
			return false
		},
		/* 224 ExtendedSpecialChar <- ((&[^] (&{p.extension.Notes} '^')) | (&[\"\'\-.] (&{p.extension.Smart} ((&[\"] '"') | (&[\'] '\'') | (&[\-] '-') | (&[.] '.'))))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1115
				}
				switch p.Buffer[position] {
				case '^':
					if !(p.extension.Notes) {
						goto l1115
					}
					if !matchChar('^') {
						goto l1115
					}
					break
				default:
					if !(p.extension.Smart) {
						goto l1115
					}
					{
						if position == len(p.Buffer) {
							goto l1115
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
							goto l1115
						}
					}
				}
			}
			return true
		l1115:
			position = position0
			return false
		},
		/* 225 Smart <- (&{p.extension.Smart} (SingleQuoted / ((&[\'] Apostrophe) | (&[\"] DoubleQuoted) | (&[\-] Dash) | (&[.] Ellipsis)))) */
		func() bool {
			if !(p.extension.Smart) {
				goto l1118
			}
			if !p.rules[ruleSingleQuoted]() {
				goto l1120
			}
			goto l1119
		l1120:
			{
				if position == len(p.Buffer) {
					goto l1118
				}
				switch p.Buffer[position] {
				case '\'':
					if !p.rules[ruleApostrophe]() {
						goto l1118
					}
					break
				case '"':
					if !p.rules[ruleDoubleQuoted]() {
						goto l1118
					}
					break
				case '-':
					if !p.rules[ruleDash]() {
						goto l1118
					}
					break
				case '.':
					if !p.rules[ruleEllipsis]() {
						goto l1118
					}
					break
				default:
					goto l1118
				}
			}
		l1119:
			return true
		l1118:
			return false
		},
		/* 226 Apostrophe <- ('\'' { yy = mk_element(APOSTROPHE) }) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1122
			}
			do(90)
			return true
		l1122:
			position = position0
			return false
		},
		/* 227 Ellipsis <- (('...' / '. . .') { yy = mk_element(ELLIPSIS) }) */
		func() bool {
			position0 := position
			if !matchString("...") {
				goto l1125
			}
			goto l1124
		l1125:
			if !matchString(". . .") {
				goto l1123
			}
		l1124:
			do(91)
			return true
		l1123:
			position = position0
			return false
		},
		/* 228 Dash <- (EmDash / EnDash) */
		func() bool {
			if !p.rules[ruleEmDash]() {
				goto l1128
			}
			goto l1127
		l1128:
			if !p.rules[ruleEnDash]() {
				goto l1126
			}
		l1127:
			return true
		l1126:
			return false
		},
		/* 229 EnDash <- ('-' &[0-9] { yy = mk_element(ENDASH) }) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l1129
			}
			if !peekClass(0) {
				goto l1129
			}
			do(92)
			return true
		l1129:
			position = position0
			return false
		},
		/* 230 EmDash <- (('---' / '--') { yy = mk_element(EMDASH) }) */
		func() bool {
			position0 := position
			if !matchString("---") {
				goto l1132
			}
			goto l1131
		l1132:
			if !matchString("--") {
				goto l1130
			}
		l1131:
			do(93)
			return true
		l1130:
			position = position0
			return false
		},
		/* 231 SingleQuoteStart <- ('\'' ![)!\],.;:-? \t\n] !(((&[r] 're') | (&[l] 'll') | (&[v] 've') | (&[m] 'm') | (&[t] 't') | (&[s] 's')) !Alphanumeric)) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1133
			}
			if peekClass(9) {
				goto l1133
			}
			{
				position1134 := position
				{
					if position == len(p.Buffer) {
						goto l1134
					}
					switch p.Buffer[position] {
					case 'r':
						position++ // matchString(`re`)
						if !matchChar('e') {
							goto l1134
						}
						break
					case 'l':
						position++ // matchString(`ll`)
						if !matchChar('l') {
							goto l1134
						}
						break
					case 'v':
						position++ // matchString(`ve`)
						if !matchChar('e') {
							goto l1134
						}
						break
					case 'm':
						position++ // matchChar
						break
					case 't':
						position++ // matchChar
						break
					case 's':
						position++ // matchChar
						break
					default:
						goto l1134
					}
				}
				if !p.rules[ruleAlphanumeric]() {
					goto l1136
				}
				goto l1134
			l1136:
				goto l1133
			l1134:
				position = position1134
			}
			return true
		l1133:
			position = position0
			return false
		},
		/* 232 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1137
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l1138
			}
			goto l1137
		l1138:
			return true
		l1137:
			position = position0
			return false
		},
		/* 233 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = mk_list(SINGLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto l1139
			}
			if !p.rules[ruleStartList]() {
				goto l1139
			}
			doarg(yySet, -1)
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1142
			}
			goto l1139
		l1142:
			if !p.rules[ruleInline]() {
				goto l1139
			}
			doarg(yySet, -2)
			do(94)
		l1140:
			{
				position1141, thunkPosition1141 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto l1143
				}
				goto l1141
			l1143:
				if !p.rules[ruleInline]() {
					goto l1141
				}
				doarg(yySet, -2)
				do(94)
				goto l1140
			l1141:
				position, thunkPosition = position1141, thunkPosition1141
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1139
			}
			do(95)
			doarg(yyPop, 2)
			return true
		l1139:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 234 DoubleQuoteStart <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1144
			}
			return true
		l1144:
			return false
		},
		/* 235 DoubleQuoteEnd <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1145
			}
			return true
		l1145:
			return false
		},
		/* 236 DoubleQuoted <- ('"' StartList (!'"' Inline { a = cons(b, a) })+ '"' { yy = mk_list(DOUBLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('"') {
				goto l1146
			}
			if !p.rules[ruleStartList]() {
				goto l1146
			}
			doarg(yySet, -1)
			if peekChar('"') {
				goto l1146
			}
			if !p.rules[ruleInline]() {
				goto l1146
			}
			doarg(yySet, -2)
			do(96)
		l1147:
			{
				position1148, thunkPosition1148 := position, thunkPosition
				if peekChar('"') {
					goto l1148
				}
				if !p.rules[ruleInline]() {
					goto l1148
				}
				doarg(yySet, -2)
				do(96)
				goto l1147
			l1148:
				position, thunkPosition = position1148, thunkPosition1148
			}
			if !matchChar('"') {
				goto l1146
			}
			do(97)
			doarg(yyPop, 2)
			return true
		l1146:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 237 NoteReference <- (&{p.extension.Notes} RawNoteReference {
                    if match, ok := p.find_note(ref.contents.str); ok {
                        yy = mk_element(NOTE)
                        yy.children = match.children
                        yy.contents.str = ""
                    } else {
                        yy = mk_str("[^"+ref.contents.str+"]")
                    }
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto l1149
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1149
			}
			doarg(yySet, -1)
			do(98)
			doarg(yyPop, 1)
			return true
		l1149:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 238 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !matchString("[^") {
				goto l1150
			}
			begin = position
			if !p.rules[ruleNewline]() {
				goto l1153
			}
			goto l1150
		l1153:
			if peekChar(']') {
				goto l1150
			}
			if !matchDot() {
				goto l1150
			}
		l1151:
			{
				position1152 := position
				if !p.rules[ruleNewline]() {
					goto l1154
				}
				goto l1152
			l1154:
				if peekChar(']') {
					goto l1152
				}
				if !matchDot() {
					goto l1152
				}
				goto l1151
			l1152:
				position = position1152
			}
			end = position
			if !matchChar(']') {
				goto l1150
			}
			do(99)
			return true
		l1150:
			position = position0
			return false
		},
		/* 239 Note <- (&{p.extension.Notes} NonindentSpace RawNoteReference ':' Sp StartList (RawNoteBlock { a = cons(yy, a) }) (&Indent RawNoteBlock { a = cons(yy, a) })* {   yy = mk_list(NOTE, a)
                    yy.contents.str = ref.contents.str
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !(p.extension.Notes) {
				goto l1155
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l1155
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1155
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto l1155
			}
			if !p.rules[ruleSp]() {
				goto l1155
			}
			if !p.rules[ruleStartList]() {
				goto l1155
			}
			doarg(yySet, -2)
			if !p.rules[ruleRawNoteBlock]() {
				goto l1155
			}
			do(100)
		l1156:
			{
				position1157, thunkPosition1157 := position, thunkPosition
				{
					position1158 := position
					if !p.rules[ruleIndent]() {
						goto l1157
					}
					position = position1158
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto l1157
				}
				do(101)
				goto l1156
			l1157:
				position, thunkPosition = position1157, thunkPosition1157
			}
			do(102)
			doarg(yyPop, 2)
			return true
		l1155:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 240 InlineNote <- (&{p.extension.Notes} '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = mk_list(NOTE, a)
                  yy.contents.str = "" }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto l1159
			}
			if !matchString("^[") {
				goto l1159
			}
			if !p.rules[ruleStartList]() {
				goto l1159
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto l1159
			}
			if !p.rules[ruleInline]() {
				goto l1159
			}
			do(103)
		l1160:
			{
				position1161 := position
				if peekChar(']') {
					goto l1161
				}
				if !p.rules[ruleInline]() {
					goto l1161
				}
				do(103)
				goto l1160
			l1161:
				position = position1161
			}
			if !matchChar(']') {
				goto l1159
			}
			do(104)
			doarg(yyPop, 1)
			return true
		l1159:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 241 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1162
			}
			doarg(yySet, -1)
		l1163:
			{
				position1164, thunkPosition1164 := position, thunkPosition
				{
					position1165, thunkPosition1165 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto l1166
					}
					doarg(yySet, -2)
					do(105)
					goto l1165
				l1166:
					position, thunkPosition = position1165, thunkPosition1165
					if !p.rules[ruleSkipBlock]() {
						goto l1164
					}
				}
			l1165:
				goto l1163
			l1164:
				position, thunkPosition = position1164, thunkPosition1164
			}
			do(106)
			if !(commit(thunkPosition0)) {
				goto l1162
			}
			doarg(yyPop, 2)
			return true
		l1162:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 242 RawNoteBlock <- (StartList (!BlankLine OptionallyIndentedLine { a = cons(yy, a) })+ (< BlankLine* > { a = cons(mk_str(yytext), a) }) {   yy = mk_str_from_list(a, true)
                    yy.key = RAW
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1167
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l1170
			}
			goto l1167
		l1170:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l1167
			}
			do(107)
		l1168:
			{
				position1169 := position
				if !p.rules[ruleBlankLine]() {
					goto l1171
				}
				goto l1169
			l1171:
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto l1169
				}
				do(107)
				goto l1168
			l1169:
				position = position1169
			}
			begin = position
		l1172:
			if !p.rules[ruleBlankLine]() {
				goto l1173
			}
			goto l1172
		l1173:
			end = position
			do(108)
			do(109)
			doarg(yyPop, 1)
			return true
		l1167:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 243 DefinitionList <- (&{p.extension.Dlists} StartList (Definition { a = cons(yy, a) })+ { yy = mk_list(DEFINITIONLIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Dlists) {
				goto l1174
			}
			if !p.rules[ruleStartList]() {
				goto l1174
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto l1174
			}
			do(110)
		l1175:
			{
				position1176, thunkPosition1176 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto l1176
				}
				do(110)
				goto l1175
			l1176:
				position, thunkPosition = position1176, thunkPosition1176
			}
			do(111)
			doarg(yyPop, 1)
			return true
		l1174:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 244 Definition <- (&(NonindentSpace !Defmark Nonspacechar RawLine BlankLine? Defmark) StartList (DListTitle { a = cons(yy, a) })+ (DefTight / DefLoose) {
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			} { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1178 := position
				if !p.rules[ruleNonindentSpace]() {
					goto l1177
				}
				if !p.rules[ruleDefmark]() {
					goto l1179
				}
				goto l1177
			l1179:
				if !p.rules[ruleNonspacechar]() {
					goto l1177
				}
				if !p.rules[ruleRawLine]() {
					goto l1177
				}
				if !p.rules[ruleBlankLine]() {
					goto l1180
				}
			l1180:
				if !p.rules[ruleDefmark]() {
					goto l1177
				}
				position = position1178
			}
			if !p.rules[ruleStartList]() {
				goto l1177
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto l1177
			}
			do(112)
		l1182:
			{
				position1183, thunkPosition1183 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto l1183
				}
				do(112)
				goto l1182
			l1183:
				position, thunkPosition = position1183, thunkPosition1183
			}
			if !p.rules[ruleDefTight]() {
				goto l1185
			}
			goto l1184
		l1185:
			if !p.rules[ruleDefLoose]() {
				goto l1177
			}
		l1184:
			do(113)
			do(114)
			doarg(yyPop, 1)
			return true
		l1177:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 245 DListTitle <- (NonindentSpace !Defmark &Nonspacechar StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline {	yy = mk_list(LIST, a)
				yy.key = DEFTITLE
			}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l1186
			}
			if !p.rules[ruleDefmark]() {
				goto l1187
			}
			goto l1186
		l1187:
			{
				position1188 := position
				if !p.rules[ruleNonspacechar]() {
					goto l1186
				}
				position = position1188
			}
			if !p.rules[ruleStartList]() {
				goto l1186
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l1191
			}
			goto l1186
		l1191:
			if !p.rules[ruleInline]() {
				goto l1186
			}
			do(115)
		l1189:
			{
				position1190 := position
				if !p.rules[ruleEndline]() {
					goto l1192
				}
				goto l1190
			l1192:
				if !p.rules[ruleInline]() {
					goto l1190
				}
				do(115)
				goto l1189
			l1190:
				position = position1190
			}
			if !p.rules[ruleSp]() {
				goto l1186
			}
			if !p.rules[ruleNewline]() {
				goto l1186
			}
			do(116)
			doarg(yyPop, 1)
			return true
		l1186:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 246 DefTight <- (&Defmark ListTight) */
		func() bool {
			{
				position1194 := position
				if !p.rules[ruleDefmark]() {
					goto l1193
				}
				position = position1194
			}
			if !p.rules[ruleListTight]() {
				goto l1193
			}
			return true
		l1193:
			return false
		},
		/* 247 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l1195
			}
			{
				position1196 := position
				if !p.rules[ruleDefmark]() {
					goto l1195
				}
				position = position1196
			}
			if !p.rules[ruleListLoose]() {
				goto l1195
			}
			return true
		l1195:
			position = position0
			return false
		},
		/* 248 Defmark <- (NonindentSpace ((&[~] '~') | (&[:] ':')) Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l1197
			}
			{
				if position == len(p.Buffer) {
					goto l1197
				}
				switch p.Buffer[position] {
				case '~':
					position++ // matchChar
					break
				case ':':
					position++ // matchChar
					break
				default:
					goto l1197
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l1197
			}
		l1199:
			if !p.rules[ruleSpacechar]() {
				goto l1200
			}
			goto l1199
		l1200:
			return true
		l1197:
			position = position0
			return false
		},
		/* 249 DefMarker <- (&{p.extension.Dlists} Defmark) */
		func() bool {
			if !(p.extension.Dlists) {
				goto l1201
			}
			if !p.rules[ruleDefmark]() {
				goto l1201
			}
			return true
		l1201:
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


/* mk_element - generic constructor for element
 */
var elbuf []element
var elock sync.Mutex

func mk_element(key int) *element {
	elock.Lock()
	if len(elbuf) == 0 {
		elbuf = make([]element, 1024)
	}
	e := &elbuf[0]
	elbuf = elbuf[1:]
	elock.Unlock()
	e.key = key
	return e
}

/* mk_str - constructor for STR element
 */
func mk_str(s string) (result *element) {
	result = mk_element(STR)
	result.contents.str = s
	return
}

/* mk_str_from_list - makes STR element by concatenating a
 * reversed list of strings, adding optional extra newline
 */
func mk_str_from_list(list *element, extra_newline bool) (result *element) {
	s := ""
	for list = reverse(list); list != nil; list = list.next {
		s += list.contents.str
	}

	if extra_newline {
		s += "\n"
	}
	result = mk_element(STR)
	result.contents.str = s
	return
}

/* mk_list - makes new list with key 'key' and children the reverse of 'lst'.
 * This is designed to be used with cons to build lists in a parser action.
 * The reversing is necessary because cons adds to the head of a list.
 */
func mk_list(key int, lst *element) (el *element) {
	el = mk_element(key)
	el.children = reverse(lst)
	return
}

/* mk_link - constructor for LINK element
 */
func mk_link(label *element, url, title string) (el *element) {
	el = mk_element(LINK)
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
			return false	/* No links or images within links */
		default:
			log.Fatalf("match_inlines encountered unknown key = %d\n", l1.key)
		}
		l1 = l1.next
		l2 = l2.next
	}
	return l1 == nil && l2 == nil	/* return true if both lists exhausted */
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
	LIST:			"LIST",
	RAW:			"RAW",
	SPACE:			"SPACE",
	LINEBREAK:		"LINEBREAK",
	ELLIPSIS:		"ELLIPSIS",
	EMDASH:			"EMDASH",
	ENDASH:			"ENDASH",
	APOSTROPHE:		"APOSTROPHE",
	SINGLEQUOTED:	"SINGLEQUOTED",
	DOUBLEQUOTED:	"DOUBLEQUOTED",
	STR:			"STR",
	LINK:			"LINK",
	IMAGE:			"IMAGE",
	CODE:			"CODE",
	HTML:			"HTML",
	EMPH:			"EMPH",
	STRONG:			"STRONG",
	PLAIN:			"PLAIN",
	PARA:			"PARA",
	LISTITEM:		"LISTITEM",
	BULLETLIST:		"BULLETLIST",
	ORDEREDLIST:	"ORDEREDLIST",
	H1:				"H1",
	H2:				"H2",
	H3:				"H3",
	H4:				"H4",
	H5:				"H5",
	H6:				"H6",
	BLOCKQUOTE:		"BLOCKQUOTE",
	VERBATIM:		"VERBATIM",
	HTMLBLOCK:		"HTMLBLOCK",
	HRULE:			"HRULE",
	REFERENCE:		"REFERENCE",
	NOTE:			"NOTE",
	DEFINITIONLIST:	"DEFINITIONLIST",
	DEFTITLE:		"DEFTITLE",
	DEFDATA:		"DEFDATA",
}
