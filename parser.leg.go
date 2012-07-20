
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
)

const (
	parserIfaceVersion_16 = iota
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
	extension		Extensions
	heap	elemHeap
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
	rules [252]func() bool
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
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
		},
		/* 44 Inlines */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(c, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
		},
		/* 45 Inlines */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = p.mkList(LIST, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
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
			 if a.next == nil { yy = a; } else { yy = p.mkList(LIST, a) } 
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
			 yy = p.mkString(yytext); yy.key = HTML 
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
		/* 59 OneStarClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 60 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 61 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 62 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(EMPH, a) 
			yyval[yyp-1] = a
		},
		/* 63 OneUlClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 64 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 65 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 66 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(EMPH, a) 
			yyval[yyp-1] = a
		},
		/* 67 TwoStarClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 68 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 69 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 70 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(STRONG, a) 
			yyval[yyp-1] = a
		},
		/* 71 TwoUlClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 72 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 73 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 74 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(STRONG, a) 
			yyval[yyp-1] = a
		},
		/* 75 Image */
		func(yytext string, _ int) {
				if yy.key == LINK {
			yy.key = IMAGE
		} else {
			result := yy
			yy.children = cons(p.mkString("!"), result.children)
		}
	
		},
		/* 76 ReferenceLinkDouble */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			
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
                       
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 77 ReferenceLinkSingle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
                           if match, found := p.findReference(a.children); found {
                               yy = p.mkLink(a.children, match.url, match.title)
                               a = nil
                           } else {
                               result := p.mkElem(LIST)
                               result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), p.mkString(yytext))));
                               yy = result
                           }
                       
			yyval[yyp-1] = a
		},
		/* 78 ExplicitLink */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			l := yyval[yyp-2]
			t := yyval[yyp-3]
			 yy = p.mkLink(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil 
			yyval[yyp-1] = s
			yyval[yyp-2] = l
			yyval[yyp-3] = t
		},
		/* 79 Source */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 80 Title */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 81 AutoLinkUrl */
		func(yytext string, _ int) {
			   yy = p.mkLink(p.mkString(yytext), yytext, "") 
		},
		/* 82 AutoLinkEmail */
		func(yytext string, _ int) {
			
                    yy = p.mkLink(p.mkString(yytext), "mailto:"+yytext, "")
                
		},
		/* 83 Reference */
		func(yytext string, _ int) {
			l := yyval[yyp-1]
			s := yyval[yyp-2]
			t := yyval[yyp-3]
			 yy = p.mkLink(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE 
			yyval[yyp-3] = t
			yyval[yyp-1] = l
			yyval[yyp-2] = s
		},
		/* 84 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 85 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 86 RefSrc */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext)
           yy.key = HTML 
		},
		/* 87 RefTitle */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 88 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 89 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 p.references = reverse(a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 90 Code */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext); yy.key = CODE 
		},
		/* 91 RawHtml */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = p.mkList(LIST, nil)
                } else {
                    yy = p.mkString(yytext)
                    yy.key = HTML
                }
            
		},
		/* 92 StartList */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 93 Line */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 94 Apostrophe */
		func(yytext string, _ int) {
			 yy = p.mkElem(APOSTROPHE) 
		},
		/* 95 Ellipsis */
		func(yytext string, _ int) {
			 yy = p.mkElem(ELLIPSIS) 
		},
		/* 96 EnDash */
		func(yytext string, _ int) {
			 yy = p.mkElem(ENDASH) 
		},
		/* 97 EmDash */
		func(yytext string, _ int) {
			 yy = p.mkElem(EMDASH) 
		},
		/* 98 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 99 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 yy = p.mkList(SINGLEQUOTED, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 100 DoubleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 101 DoubleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = p.mkList(DOUBLEQUOTED, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 102 NoteReference */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			
                    if match, ok := p.find_note(ref.contents.str); ok {
                        yy = p.mkElem(NOTE)
                        yy.children = match.children
                        yy.contents.str = ""
                    } else {
                        yy = p.mkString("[^"+ref.contents.str+"]")
                    }
                
			yyval[yyp-1] = ref
		},
		/* 103 RawNoteReference */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 104 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			ref := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = ref
		},
		/* 105 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			ref := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = ref
		},
		/* 106 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			ref := yyval[yyp-2]
			   yy = p.mkList(NOTE, a)
                    yy.contents.str = ref.contents.str
                
			yyval[yyp-2] = ref
			yyval[yyp-1] = a
		},
		/* 107 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 108 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(NOTE, a)
                  yy.contents.str = "" 
			yyval[yyp-1] = a
		},
		/* 109 Notes */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 110 Notes */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 p.notes = reverse(a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 111 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 112 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(p.mkString(yytext), a) 
			yyval[yyp-1] = a
		},
		/* 113 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   yy = p.mkStringFromList(a, true)
                    yy.key = RAW
                
			yyval[yyp-1] = a
		},
		/* 114 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 115 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(DEFINITIONLIST, a) 
			yyval[yyp-1] = a
		},
		/* 116 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 117 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			
			yyval[yyp-1] = a
		},
		/* 118 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 119 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 120 DListTitle */
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
		yyPush = 121 + iota
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
	3:	{0, 0, 0, 0, 50, 232, 255, 3, 254, 255, 255, 135, 254, 255, 255, 71, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
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
			doarg(yySet, -2)
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
				doarg(yySet, -1)
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
					doarg(yySet, -1)
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
		/* 147 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line ((&[\-] ('---' '-'*)) | (&[=] ('===' '='*))) Newline) { yy = p.mkString("\n")
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
						position++
						if !matchString("--") {
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
						position++
						if !matchString("==") {
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
			if (position < len(p.Buffer)) {
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
		/* 155 OneStarOpen <- (!StarLine '*' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleStarLine]() {
				goto l760
			}
			goto l759
		l760:
			if !matchChar('*') {
				goto l759
			}
			if !p.rules[ruleSpacechar]() {
				goto l761
			}
			goto l759
		l761:
			if !p.rules[ruleNewline]() {
				goto l762
			}
			goto l759
		l762:
			return true
		l759:
			position = position0
			return false
		},
		/* 156 OneStarClose <- (!Spacechar !Newline Inline '*' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l764
			}
			goto l763
		l764:
			if !p.rules[ruleNewline]() {
				goto l765
			}
			goto l763
		l765:
			if !p.rules[ruleInline]() {
				goto l763
			}
			doarg(yySet, -1)
			if !matchChar('*') {
				goto l763
			}
			do(59)
			doarg(yyPop, 1)
			return true
		l763:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 157 EmphStar <- (OneStarOpen StartList (!OneStarClose Inline { a = cons(yy, a) })* OneStarClose { a = cons(yy, a) } { yy = p.mkList(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneStarOpen]() {
				goto l766
			}
			if !p.rules[ruleStartList]() {
				goto l766
			}
			doarg(yySet, -1)
		l767:
			{
				position768, thunkPosition768 := position, thunkPosition
				if !p.rules[ruleOneStarClose]() {
					goto l769
				}
				goto l768
			l769:
				if !p.rules[ruleInline]() {
					goto l768
				}
				do(60)
				goto l767
			l768:
				position, thunkPosition = position768, thunkPosition768
			}
			if !p.rules[ruleOneStarClose]() {
				goto l766
			}
			do(61)
			do(62)
			doarg(yyPop, 1)
			return true
		l766:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 158 OneUlOpen <- (!UlLine '_' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l771
			}
			goto l770
		l771:
			if !matchChar('_') {
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
		/* 159 OneUlClose <- (!Spacechar !Newline Inline '_' !Alphanumeric { yy = a }) */
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
			if !matchChar('_') {
				goto l774
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l777
			}
			goto l774
		l777:
			do(63)
			doarg(yyPop, 1)
			return true
		l774:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 160 EmphUl <- (OneUlOpen StartList (!OneUlClose Inline { a = cons(yy, a) })* OneUlClose { a = cons(yy, a) } { yy = p.mkList(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneUlOpen]() {
				goto l778
			}
			if !p.rules[ruleStartList]() {
				goto l778
			}
			doarg(yySet, -1)
		l779:
			{
				position780, thunkPosition780 := position, thunkPosition
				if !p.rules[ruleOneUlClose]() {
					goto l781
				}
				goto l780
			l781:
				if !p.rules[ruleInline]() {
					goto l780
				}
				do(64)
				goto l779
			l780:
				position, thunkPosition = position780, thunkPosition780
			}
			if !p.rules[ruleOneUlClose]() {
				goto l778
			}
			do(65)
			do(66)
			doarg(yyPop, 1)
			return true
		l778:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 161 Strong <- ((&[_] StrongUl) | (&[*] StrongStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l782
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleStrongUl]() {
						goto l782
					}
					break
				case '*':
					if !p.rules[ruleStrongStar]() {
						goto l782
					}
					break
				default:
					goto l782
				}
			}
			return true
		l782:
			return false
		},
		/* 162 TwoStarOpen <- (!StarLine '**' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleStarLine]() {
				goto l785
			}
			goto l784
		l785:
			if !matchString("**") {
				goto l784
			}
			if !p.rules[ruleSpacechar]() {
				goto l786
			}
			goto l784
		l786:
			if !p.rules[ruleNewline]() {
				goto l787
			}
			goto l784
		l787:
			return true
		l784:
			position = position0
			return false
		},
		/* 163 TwoStarClose <- (!Spacechar !Newline Inline '**' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l789
			}
			goto l788
		l789:
			if !p.rules[ruleNewline]() {
				goto l790
			}
			goto l788
		l790:
			if !p.rules[ruleInline]() {
				goto l788
			}
			doarg(yySet, -1)
			if !matchString("**") {
				goto l788
			}
			do(67)
			doarg(yyPop, 1)
			return true
		l788:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 164 StrongStar <- (TwoStarOpen StartList (!TwoStarClose Inline { a = cons(yy, a) })* TwoStarClose { a = cons(yy, a) } { yy = p.mkList(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoStarOpen]() {
				goto l791
			}
			if !p.rules[ruleStartList]() {
				goto l791
			}
			doarg(yySet, -1)
		l792:
			{
				position793, thunkPosition793 := position, thunkPosition
				if !p.rules[ruleTwoStarClose]() {
					goto l794
				}
				goto l793
			l794:
				if !p.rules[ruleInline]() {
					goto l793
				}
				do(68)
				goto l792
			l793:
				position, thunkPosition = position793, thunkPosition793
			}
			if !p.rules[ruleTwoStarClose]() {
				goto l791
			}
			do(69)
			do(70)
			doarg(yyPop, 1)
			return true
		l791:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 165 TwoUlOpen <- (!UlLine '__' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l796
			}
			goto l795
		l796:
			if !matchString("__") {
				goto l795
			}
			if !p.rules[ruleSpacechar]() {
				goto l797
			}
			goto l795
		l797:
			if !p.rules[ruleNewline]() {
				goto l798
			}
			goto l795
		l798:
			return true
		l795:
			position = position0
			return false
		},
		/* 166 TwoUlClose <- (!Spacechar !Newline Inline '__' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l800
			}
			goto l799
		l800:
			if !p.rules[ruleNewline]() {
				goto l801
			}
			goto l799
		l801:
			if !p.rules[ruleInline]() {
				goto l799
			}
			doarg(yySet, -1)
			if !matchString("__") {
				goto l799
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l802
			}
			goto l799
		l802:
			do(71)
			doarg(yyPop, 1)
			return true
		l799:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 167 StrongUl <- (TwoUlOpen StartList (!TwoUlClose Inline { a = cons(yy, a) })* TwoUlClose { a = cons(yy, a) } { yy = p.mkList(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoUlOpen]() {
				goto l803
			}
			if !p.rules[ruleStartList]() {
				goto l803
			}
			doarg(yySet, -1)
		l804:
			{
				position805, thunkPosition805 := position, thunkPosition
				if !p.rules[ruleTwoUlClose]() {
					goto l806
				}
				goto l805
			l806:
				if !p.rules[ruleInline]() {
					goto l805
				}
				do(72)
				goto l804
			l805:
				position, thunkPosition = position805, thunkPosition805
			}
			if !p.rules[ruleTwoUlClose]() {
				goto l803
			}
			do(73)
			do(74)
			doarg(yyPop, 1)
			return true
		l803:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 168 Image <- ('!' (ExplicitLink / ReferenceLink) {	if yy.key == LINK {
			yy.key = IMAGE
		} else {
			result := yy
			yy.children = cons(p.mkString("!"), result.children)
		}
	}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('!') {
				goto l807
			}
			if !p.rules[ruleExplicitLink]() {
				goto l809
			}
			goto l808
		l809:
			if !p.rules[ruleReferenceLink]() {
				goto l807
			}
		l808:
			do(75)
			return true
		l807:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 169 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() bool {
			if !p.rules[ruleExplicitLink]() {
				goto l812
			}
			goto l811
		l812:
			if !p.rules[ruleReferenceLink]() {
				goto l813
			}
			goto l811
		l813:
			if !p.rules[ruleAutoLink]() {
				goto l810
			}
		l811:
			return true
		l810:
			return false
		},
		/* 170 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() bool {
			if !p.rules[ruleReferenceLinkDouble]() {
				goto l816
			}
			goto l815
		l816:
			if !p.rules[ruleReferenceLinkSingle]() {
				goto l814
			}
		l815:
			return true
		l814:
			return false
		},
		/* 171 ReferenceLinkDouble <- (Label < Spnl > !'[]' Label {
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
				goto l817
			}
			doarg(yySet, -2)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto l817
			}
			end = position
			if !matchString("[]") {
				goto l818
			}
			goto l817
		l818:
			if !p.rules[ruleLabel]() {
				goto l817
			}
			doarg(yySet, -1)
			do(76)
			doarg(yyPop, 2)
			return true
		l817:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 172 ReferenceLinkSingle <- (Label < (Spnl '[]')? > {
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
				goto l819
			}
			doarg(yySet, -1)
			begin = position
			{
				position820 := position
				if !p.rules[ruleSpnl]() {
					goto l820
				}
				if !matchString("[]") {
					goto l820
				}
				goto l821
			l820:
				position = position820
			}
		l821:
			end = position
			do(77)
			doarg(yyPop, 1)
			return true
		l819:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 173 ExplicitLink <- (Label Spnl '(' Sp Source Spnl Title Sp ')' { yy = p.mkLink(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleLabel]() {
				goto l822
			}
			doarg(yySet, -2)
			if !p.rules[ruleSpnl]() {
				goto l822
			}
			if !matchChar('(') {
				goto l822
			}
			if !p.rules[ruleSp]() {
				goto l822
			}
			if !p.rules[ruleSource]() {
				goto l822
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l822
			}
			if !p.rules[ruleTitle]() {
				goto l822
			}
			doarg(yySet, -3)
			if !p.rules[ruleSp]() {
				goto l822
			}
			if !matchChar(')') {
				goto l822
			}
			do(78)
			doarg(yyPop, 3)
			return true
		l822:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 174 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			{
				position824 := position
				if !matchChar('<') {
					goto l825
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l825
				}
				end = position
				if !matchChar('>') {
					goto l825
				}
				goto l824
			l825:
				position = position824
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l823
				}
				end = position
			}
		l824:
			do(79)
			return true
		l823:
			position = position0
			return false
		},
		/* 175 SourceContents <- (((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* / '') */
		func() bool {
		l829:
			{
				position830 := position
				if position == len(p.Buffer) {
					goto l832
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l832
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l832
					}
				}
			l833:
				if position == len(p.Buffer) {
					goto l834
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l834
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l834
					}
				}
				goto l833
			l834:
				goto l831
			l832:
				if !matchChar('(') {
					goto l830
				}
				if !p.rules[ruleSourceContents]() {
					goto l830
				}
				if !matchChar(')') {
					goto l830
				}
			l831:
				goto l829
			l830:
				position = position830
			}
			goto l827
		l827:
			return true
		},
		/* 176 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = p.mkString(yytext) }) */
		func() bool {
			if !p.rules[ruleTitleSingle]() {
				goto l837
			}
			goto l836
		l837:
			if !p.rules[ruleTitleDouble]() {
				goto l838
			}
			goto l836
		l838:
			begin = position
			end = position
		l836:
			do(80)
			return true
		},
		/* 177 TitleSingle <- ('\'' < (!('\'' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '\'') */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l839
			}
			begin = position
		l840:
			{
				position841 := position
				{
					position842 := position
					if !matchChar('\'') {
						goto l842
					}
					if !p.rules[ruleSp]() {
						goto l842
					}
					{
						if position == len(p.Buffer) {
							goto l842
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l842
							}
							break
						default:
							goto l842
						}
					}
					goto l841
				l842:
					position = position842
				}
				if !matchDot() {
					goto l841
				}
				goto l840
			l841:
				position = position841
			}
			end = position
			if !matchChar('\'') {
				goto l839
			}
			return true
		l839:
			position = position0
			return false
		},
		/* 178 TitleDouble <- ('"' < (!('"' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '"') */
		func() bool {
			position0 := position
			if !matchChar('"') {
				goto l844
			}
			begin = position
		l845:
			{
				position846 := position
				{
					position847 := position
					if !matchChar('"') {
						goto l847
					}
					if !p.rules[ruleSp]() {
						goto l847
					}
					{
						if position == len(p.Buffer) {
							goto l847
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l847
							}
							break
						default:
							goto l847
						}
					}
					goto l846
				l847:
					position = position847
				}
				if !matchDot() {
					goto l846
				}
				goto l845
			l846:
				position = position846
			}
			end = position
			if !matchChar('"') {
				goto l844
			}
			return true
		l844:
			position = position0
			return false
		},
		/* 179 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() bool {
			if !p.rules[ruleAutoLinkUrl]() {
				goto l851
			}
			goto l850
		l851:
			if !p.rules[ruleAutoLinkEmail]() {
				goto l849
			}
		l850:
			return true
		l849:
			return false
		},
		/* 180 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = p.mkLink(p.mkString(yytext), yytext, "") }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l852
			}
			begin = position
			if !matchClass(2) {
				goto l852
			}
		l853:
			if !matchClass(2) {
				goto l854
			}
			goto l853
		l854:
			if !matchString("://") {
				goto l852
			}
			if !p.rules[ruleNewline]() {
				goto l857
			}
			goto l852
		l857:
			if peekChar('>') {
				goto l852
			}
			if !matchDot() {
				goto l852
			}
		l855:
			{
				position856 := position
				if !p.rules[ruleNewline]() {
					goto l858
				}
				goto l856
			l858:
				if peekChar('>') {
					goto l856
				}
				if !matchDot() {
					goto l856
				}
				goto l855
			l856:
				position = position856
			}
			end = position
			if !matchChar('>') {
				goto l852
			}
			do(81)
			return true
		l852:
			position = position0
			return false
		},
		/* 181 AutoLinkEmail <- ('<' 'mailto:'? < [-A-Za-z0-9+_./!%~$]+ '@' (!Newline !'>' .)+ > '>' {
                    yy = p.mkLink(p.mkString(yytext), "mailto:"+yytext, "")
                }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l859
			}
			if !matchString("mailto:") {
				goto l860
			}
		l860:
			begin = position
			if !matchClass(3) {
				goto l859
			}
		l862:
			if !matchClass(3) {
				goto l863
			}
			goto l862
		l863:
			if !matchChar('@') {
				goto l859
			}
			if !p.rules[ruleNewline]() {
				goto l866
			}
			goto l859
		l866:
			if peekChar('>') {
				goto l859
			}
			if !matchDot() {
				goto l859
			}
		l864:
			{
				position865 := position
				if !p.rules[ruleNewline]() {
					goto l867
				}
				goto l865
			l867:
				if peekChar('>') {
					goto l865
				}
				if !matchDot() {
					goto l865
				}
				goto l864
			l865:
				position = position865
			}
			end = position
			if !matchChar('>') {
				goto l859
			}
			do(82)
			return true
		l859:
			position = position0
			return false
		},
		/* 182 Reference <- (NonindentSpace !'[]' Label ':' Spnl RefSrc RefTitle BlankLine+ { yy = p.mkLink(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleNonindentSpace]() {
				goto l868
			}
			if !matchString("[]") {
				goto l869
			}
			goto l868
		l869:
			if !p.rules[ruleLabel]() {
				goto l868
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto l868
			}
			if !p.rules[ruleSpnl]() {
				goto l868
			}
			if !p.rules[ruleRefSrc]() {
				goto l868
			}
			doarg(yySet, -2)
			if !p.rules[ruleRefTitle]() {
				goto l868
			}
			doarg(yySet, -3)
			if !p.rules[ruleBlankLine]() {
				goto l868
			}
		l870:
			if !p.rules[ruleBlankLine]() {
				goto l871
			}
			goto l870
		l871:
			do(83)
			doarg(yyPop, 3)
			return true
		l868:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 183 Label <- ('[' ((!'^' &{p.extension.Notes}) / (&. &{!p.extension.Notes})) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l872
			}
			if peekChar('^') {
				goto l874
			}
			if !(p.extension.Notes) {
				goto l874
			}
			goto l873
		l874:
			if !(position < len(p.Buffer)) {
				goto l872
			}
			if !(!p.extension.Notes) {
				goto l872
			}
		l873:
			if !p.rules[ruleStartList]() {
				goto l872
			}
			doarg(yySet, -1)
		l875:
			{
				position876 := position
				if peekChar(']') {
					goto l876
				}
				if !p.rules[ruleInline]() {
					goto l876
				}
				do(84)
				goto l875
			l876:
				position = position876
			}
			if !matchChar(']') {
				goto l872
			}
			do(85)
			doarg(yyPop, 1)
			return true
		l872:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 184 RefSrc <- (< Nonspacechar+ > { yy = p.mkString(yytext)
           yy.key = HTML }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto l877
			}
		l878:
			if !p.rules[ruleNonspacechar]() {
				goto l879
			}
			goto l878
		l879:
			end = position
			do(86)
			return true
		l877:
			position = position0
			return false
		},
		/* 185 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRefTitleSingle]() {
				goto l882
			}
			goto l881
		l882:
			if !p.rules[ruleRefTitleDouble]() {
				goto l883
			}
			goto l881
		l883:
			if !p.rules[ruleRefTitleParens]() {
				goto l884
			}
			goto l881
		l884:
			if !p.rules[ruleEmptyTitle]() {
				goto l880
			}
		l881:
			do(87)
			return true
		l880:
			position = position0
			return false
		},
		/* 186 EmptyTitle <- (< '' >) */
		func() bool {
			begin = position
			end = position
			return true
		},
		/* 187 RefTitleSingle <- (Spnl '\'' < (!((&[\'] ('\'' Sp Newline)) | (&[\n\r] Newline)) .)* > '\'') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l886
			}
			if !matchChar('\'') {
				goto l886
			}
			begin = position
		l887:
			{
				position888 := position
				{
					position889 := position
					{
						if position == len(p.Buffer) {
							goto l889
						}
						switch p.Buffer[position] {
						case '\'':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l889
							}
							if !p.rules[ruleNewline]() {
								goto l889
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l889
							}
							break
						default:
							goto l889
						}
					}
					goto l888
				l889:
					position = position889
				}
				if !matchDot() {
					goto l888
				}
				goto l887
			l888:
				position = position888
			}
			end = position
			if !matchChar('\'') {
				goto l886
			}
			return true
		l886:
			position = position0
			return false
		},
		/* 188 RefTitleDouble <- (Spnl '"' < (!((&[\"] ('"' Sp Newline)) | (&[\n\r] Newline)) .)* > '"') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l891
			}
			if !matchChar('"') {
				goto l891
			}
			begin = position
		l892:
			{
				position893 := position
				{
					position894 := position
					{
						if position == len(p.Buffer) {
							goto l894
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l894
							}
							if !p.rules[ruleNewline]() {
								goto l894
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l894
							}
							break
						default:
							goto l894
						}
					}
					goto l893
				l894:
					position = position894
				}
				if !matchDot() {
					goto l893
				}
				goto l892
			l893:
				position = position893
			}
			end = position
			if !matchChar('"') {
				goto l891
			}
			return true
		l891:
			position = position0
			return false
		},
		/* 189 RefTitleParens <- (Spnl '(' < (!((&[)] (')' Sp Newline)) | (&[\n\r] Newline)) .)* > ')') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l896
			}
			if !matchChar('(') {
				goto l896
			}
			begin = position
		l897:
			{
				position898 := position
				{
					position899 := position
					{
						if position == len(p.Buffer) {
							goto l899
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l899
							}
							if !p.rules[ruleNewline]() {
								goto l899
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l899
							}
							break
						default:
							goto l899
						}
					}
					goto l898
				l899:
					position = position899
				}
				if !matchDot() {
					goto l898
				}
				goto l897
			l898:
				position = position898
			}
			end = position
			if !matchChar(')') {
				goto l896
			}
			return true
		l896:
			position = position0
			return false
		},
		/* 190 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l901
			}
			doarg(yySet, -1)
		l902:
			{
				position903, thunkPosition903 := position, thunkPosition
				{
					position904, thunkPosition904 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto l905
					}
					doarg(yySet, -2)
					do(88)
					goto l904
				l905:
					position, thunkPosition = position904, thunkPosition904
					if !p.rules[ruleSkipBlock]() {
						goto l903
					}
				}
			l904:
				goto l902
			l903:
				position, thunkPosition = position903, thunkPosition903
			}
			do(89)
			if !(commit(thunkPosition0)) {
				goto l901
			}
			doarg(yyPop, 2)
			return true
		l901:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 191 Ticks1 <- ('`' !'`') */
		func() bool {
			position0 := position
			if !matchChar('`') {
				goto l906
			}
			if peekChar('`') {
				goto l906
			}
			return true
		l906:
			position = position0
			return false
		},
		/* 192 Ticks2 <- ('``' !'`') */
		func() bool {
			position0 := position
			if !matchString("``") {
				goto l907
			}
			if peekChar('`') {
				goto l907
			}
			return true
		l907:
			position = position0
			return false
		},
		/* 193 Ticks3 <- ('```' !'`') */
		func() bool {
			position0 := position
			if !matchString("```") {
				goto l908
			}
			if peekChar('`') {
				goto l908
			}
			return true
		l908:
			position = position0
			return false
		},
		/* 194 Ticks4 <- ('````' !'`') */
		func() bool {
			position0 := position
			if !matchString("````") {
				goto l909
			}
			if peekChar('`') {
				goto l909
			}
			return true
		l909:
			position = position0
			return false
		},
		/* 195 Ticks5 <- ('`````' !'`') */
		func() bool {
			position0 := position
			if !matchString("`````") {
				goto l910
			}
			if peekChar('`') {
				goto l910
			}
			return true
		l910:
			position = position0
			return false
		},
		/* 196 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks1 '`'+)) | (&[\t\n\r ] (!(Sp Ticks1) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks2 '`'+)) | (&[\t\n\r ] (!(Sp Ticks2) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks3 '`'+)) | (&[\t\n\r ] (!(Sp Ticks3) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks4 '`'+)) | (&[\t\n\r ] (!(Sp Ticks4) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks5 '`'+)) | (&[\t\n\r ] (!(Sp Ticks5) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks5)) { yy = p.mkString(yytext); yy.key = CODE }) */
		func() bool {
			position0 := position
			{
				position912 := position
				if !p.rules[ruleTicks1]() {
					goto l913
				}
				if !p.rules[ruleSp]() {
					goto l913
				}
				begin = position
				if peekChar('`') {
					goto l917
				}
				if !p.rules[ruleNonspacechar]() {
					goto l917
				}
			l918:
				if peekChar('`') {
					goto l919
				}
				if !p.rules[ruleNonspacechar]() {
					goto l919
				}
				goto l918
			l919:
				goto l916
			l917:
				{
					if position == len(p.Buffer) {
						goto l913
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks1]() {
							goto l921
						}
						goto l913
					l921:
						if !matchChar('`') {
							goto l913
						}
					l922:
						if !matchChar('`') {
							goto l923
						}
						goto l922
					l923:
						break
					default:
						{
							position924 := position
							if !p.rules[ruleSp]() {
								goto l924
							}
							if !p.rules[ruleTicks1]() {
								goto l924
							}
							goto l913
						l924:
							position = position924
						}
						{
							if position == len(p.Buffer) {
								goto l913
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l913
								}
								if !p.rules[ruleBlankLine]() {
									goto l926
								}
								goto l913
							l926:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l913
								}
								break
							default:
								goto l913
							}
						}
					}
				}
			l916:
			l914:
				{
					position915 := position
					if peekChar('`') {
						goto l928
					}
					if !p.rules[ruleNonspacechar]() {
						goto l928
					}
				l929:
					if peekChar('`') {
						goto l930
					}
					if !p.rules[ruleNonspacechar]() {
						goto l930
					}
					goto l929
				l930:
					goto l927
				l928:
					{
						if position == len(p.Buffer) {
							goto l915
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks1]() {
								goto l932
							}
							goto l915
						l932:
							if !matchChar('`') {
								goto l915
							}
						l933:
							if !matchChar('`') {
								goto l934
							}
							goto l933
						l934:
							break
						default:
							{
								position935 := position
								if !p.rules[ruleSp]() {
									goto l935
								}
								if !p.rules[ruleTicks1]() {
									goto l935
								}
								goto l915
							l935:
								position = position935
							}
							{
								if position == len(p.Buffer) {
									goto l915
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l915
									}
									if !p.rules[ruleBlankLine]() {
										goto l937
									}
									goto l915
								l937:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l915
									}
									break
								default:
									goto l915
								}
							}
						}
					}
				l927:
					goto l914
				l915:
					position = position915
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l913
				}
				if !p.rules[ruleTicks1]() {
					goto l913
				}
				goto l912
			l913:
				position = position912
				if !p.rules[ruleTicks2]() {
					goto l938
				}
				if !p.rules[ruleSp]() {
					goto l938
				}
				begin = position
				if peekChar('`') {
					goto l942
				}
				if !p.rules[ruleNonspacechar]() {
					goto l942
				}
			l943:
				if peekChar('`') {
					goto l944
				}
				if !p.rules[ruleNonspacechar]() {
					goto l944
				}
				goto l943
			l944:
				goto l941
			l942:
				{
					if position == len(p.Buffer) {
						goto l938
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks2]() {
							goto l946
						}
						goto l938
					l946:
						if !matchChar('`') {
							goto l938
						}
					l947:
						if !matchChar('`') {
							goto l948
						}
						goto l947
					l948:
						break
					default:
						{
							position949 := position
							if !p.rules[ruleSp]() {
								goto l949
							}
							if !p.rules[ruleTicks2]() {
								goto l949
							}
							goto l938
						l949:
							position = position949
						}
						{
							if position == len(p.Buffer) {
								goto l938
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l938
								}
								if !p.rules[ruleBlankLine]() {
									goto l951
								}
								goto l938
							l951:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l938
								}
								break
							default:
								goto l938
							}
						}
					}
				}
			l941:
			l939:
				{
					position940 := position
					if peekChar('`') {
						goto l953
					}
					if !p.rules[ruleNonspacechar]() {
						goto l953
					}
				l954:
					if peekChar('`') {
						goto l955
					}
					if !p.rules[ruleNonspacechar]() {
						goto l955
					}
					goto l954
				l955:
					goto l952
				l953:
					{
						if position == len(p.Buffer) {
							goto l940
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks2]() {
								goto l957
							}
							goto l940
						l957:
							if !matchChar('`') {
								goto l940
							}
						l958:
							if !matchChar('`') {
								goto l959
							}
							goto l958
						l959:
							break
						default:
							{
								position960 := position
								if !p.rules[ruleSp]() {
									goto l960
								}
								if !p.rules[ruleTicks2]() {
									goto l960
								}
								goto l940
							l960:
								position = position960
							}
							{
								if position == len(p.Buffer) {
									goto l940
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l940
									}
									if !p.rules[ruleBlankLine]() {
										goto l962
									}
									goto l940
								l962:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l940
									}
									break
								default:
									goto l940
								}
							}
						}
					}
				l952:
					goto l939
				l940:
					position = position940
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l938
				}
				if !p.rules[ruleTicks2]() {
					goto l938
				}
				goto l912
			l938:
				position = position912
				if !p.rules[ruleTicks3]() {
					goto l963
				}
				if !p.rules[ruleSp]() {
					goto l963
				}
				begin = position
				if peekChar('`') {
					goto l967
				}
				if !p.rules[ruleNonspacechar]() {
					goto l967
				}
			l968:
				if peekChar('`') {
					goto l969
				}
				if !p.rules[ruleNonspacechar]() {
					goto l969
				}
				goto l968
			l969:
				goto l966
			l967:
				{
					if position == len(p.Buffer) {
						goto l963
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks3]() {
							goto l971
						}
						goto l963
					l971:
						if !matchChar('`') {
							goto l963
						}
					l972:
						if !matchChar('`') {
							goto l973
						}
						goto l972
					l973:
						break
					default:
						{
							position974 := position
							if !p.rules[ruleSp]() {
								goto l974
							}
							if !p.rules[ruleTicks3]() {
								goto l974
							}
							goto l963
						l974:
							position = position974
						}
						{
							if position == len(p.Buffer) {
								goto l963
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l963
								}
								if !p.rules[ruleBlankLine]() {
									goto l976
								}
								goto l963
							l976:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l963
								}
								break
							default:
								goto l963
							}
						}
					}
				}
			l966:
			l964:
				{
					position965 := position
					if peekChar('`') {
						goto l978
					}
					if !p.rules[ruleNonspacechar]() {
						goto l978
					}
				l979:
					if peekChar('`') {
						goto l980
					}
					if !p.rules[ruleNonspacechar]() {
						goto l980
					}
					goto l979
				l980:
					goto l977
				l978:
					{
						if position == len(p.Buffer) {
							goto l965
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks3]() {
								goto l982
							}
							goto l965
						l982:
							if !matchChar('`') {
								goto l965
							}
						l983:
							if !matchChar('`') {
								goto l984
							}
							goto l983
						l984:
							break
						default:
							{
								position985 := position
								if !p.rules[ruleSp]() {
									goto l985
								}
								if !p.rules[ruleTicks3]() {
									goto l985
								}
								goto l965
							l985:
								position = position985
							}
							{
								if position == len(p.Buffer) {
									goto l965
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l965
									}
									if !p.rules[ruleBlankLine]() {
										goto l987
									}
									goto l965
								l987:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l965
									}
									break
								default:
									goto l965
								}
							}
						}
					}
				l977:
					goto l964
				l965:
					position = position965
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l963
				}
				if !p.rules[ruleTicks3]() {
					goto l963
				}
				goto l912
			l963:
				position = position912
				if !p.rules[ruleTicks4]() {
					goto l988
				}
				if !p.rules[ruleSp]() {
					goto l988
				}
				begin = position
				if peekChar('`') {
					goto l992
				}
				if !p.rules[ruleNonspacechar]() {
					goto l992
				}
			l993:
				if peekChar('`') {
					goto l994
				}
				if !p.rules[ruleNonspacechar]() {
					goto l994
				}
				goto l993
			l994:
				goto l991
			l992:
				{
					if position == len(p.Buffer) {
						goto l988
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks4]() {
							goto l996
						}
						goto l988
					l996:
						if !matchChar('`') {
							goto l988
						}
					l997:
						if !matchChar('`') {
							goto l998
						}
						goto l997
					l998:
						break
					default:
						{
							position999 := position
							if !p.rules[ruleSp]() {
								goto l999
							}
							if !p.rules[ruleTicks4]() {
								goto l999
							}
							goto l988
						l999:
							position = position999
						}
						{
							if position == len(p.Buffer) {
								goto l988
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l988
								}
								if !p.rules[ruleBlankLine]() {
									goto l1001
								}
								goto l988
							l1001:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l988
								}
								break
							default:
								goto l988
							}
						}
					}
				}
			l991:
			l989:
				{
					position990 := position
					if peekChar('`') {
						goto l1003
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1003
					}
				l1004:
					if peekChar('`') {
						goto l1005
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1005
					}
					goto l1004
				l1005:
					goto l1002
				l1003:
					{
						if position == len(p.Buffer) {
							goto l990
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks4]() {
								goto l1007
							}
							goto l990
						l1007:
							if !matchChar('`') {
								goto l990
							}
						l1008:
							if !matchChar('`') {
								goto l1009
							}
							goto l1008
						l1009:
							break
						default:
							{
								position1010 := position
								if !p.rules[ruleSp]() {
									goto l1010
								}
								if !p.rules[ruleTicks4]() {
									goto l1010
								}
								goto l990
							l1010:
								position = position1010
							}
							{
								if position == len(p.Buffer) {
									goto l990
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l990
									}
									if !p.rules[ruleBlankLine]() {
										goto l1012
									}
									goto l990
								l1012:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l990
									}
									break
								default:
									goto l990
								}
							}
						}
					}
				l1002:
					goto l989
				l990:
					position = position990
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l988
				}
				if !p.rules[ruleTicks4]() {
					goto l988
				}
				goto l912
			l988:
				position = position912
				if !p.rules[ruleTicks5]() {
					goto l911
				}
				if !p.rules[ruleSp]() {
					goto l911
				}
				begin = position
				if peekChar('`') {
					goto l1016
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1016
				}
			l1017:
				if peekChar('`') {
					goto l1018
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1018
				}
				goto l1017
			l1018:
				goto l1015
			l1016:
				{
					if position == len(p.Buffer) {
						goto l911
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks5]() {
							goto l1020
						}
						goto l911
					l1020:
						if !matchChar('`') {
							goto l911
						}
					l1021:
						if !matchChar('`') {
							goto l1022
						}
						goto l1021
					l1022:
						break
					default:
						{
							position1023 := position
							if !p.rules[ruleSp]() {
								goto l1023
							}
							if !p.rules[ruleTicks5]() {
								goto l1023
							}
							goto l911
						l1023:
							position = position1023
						}
						{
							if position == len(p.Buffer) {
								goto l911
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l911
								}
								if !p.rules[ruleBlankLine]() {
									goto l1025
								}
								goto l911
							l1025:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l911
								}
								break
							default:
								goto l911
							}
						}
					}
				}
			l1015:
			l1013:
				{
					position1014 := position
					if peekChar('`') {
						goto l1027
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1027
					}
				l1028:
					if peekChar('`') {
						goto l1029
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1029
					}
					goto l1028
				l1029:
					goto l1026
				l1027:
					{
						if position == len(p.Buffer) {
							goto l1014
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks5]() {
								goto l1031
							}
							goto l1014
						l1031:
							if !matchChar('`') {
								goto l1014
							}
						l1032:
							if !matchChar('`') {
								goto l1033
							}
							goto l1032
						l1033:
							break
						default:
							{
								position1034 := position
								if !p.rules[ruleSp]() {
									goto l1034
								}
								if !p.rules[ruleTicks5]() {
									goto l1034
								}
								goto l1014
							l1034:
								position = position1034
							}
							{
								if position == len(p.Buffer) {
									goto l1014
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l1014
									}
									if !p.rules[ruleBlankLine]() {
										goto l1036
									}
									goto l1014
								l1036:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l1014
									}
									break
								default:
									goto l1014
								}
							}
						}
					}
				l1026:
					goto l1013
				l1014:
					position = position1014
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l911
				}
				if !p.rules[ruleTicks5]() {
					goto l911
				}
			}
		l912:
			do(90)
			return true
		l911:
			position = position0
			return false
		},
		/* 197 RawHtml <- (< (HtmlComment / HtmlBlockScript / HtmlTag) > {   if p.extension.FilterHTML {
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
				goto l1039
			}
			goto l1038
		l1039:
			if !p.rules[ruleHtmlBlockScript]() {
				goto l1040
			}
			goto l1038
		l1040:
			if !p.rules[ruleHtmlTag]() {
				goto l1037
			}
		l1038:
			end = position
			do(91)
			return true
		l1037:
			position = position0
			return false
		},
		/* 198 BlankLine <- (Sp Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1041
			}
			if !p.rules[ruleNewline]() {
				goto l1041
			}
			return true
		l1041:
			position = position0
			return false
		},
		/* 199 Quoted <- ((&[\'] ('\'' (!'\'' .)* '\'')) | (&[\"] ('"' (!'"' .)* '"'))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1042
				}
				switch p.Buffer[position] {
				case '\'':
					position++ // matchChar
				l1044:
					if position == len(p.Buffer) {
						goto l1045
					}
					switch p.Buffer[position] {
					case '\'':
						goto l1045
					default:
						position++
					}
					goto l1044
				l1045:
					if !matchChar('\'') {
						goto l1042
					}
					break
				case '"':
					position++ // matchChar
				l1046:
					if position == len(p.Buffer) {
						goto l1047
					}
					switch p.Buffer[position] {
					case '"':
						goto l1047
					default:
						position++
					}
					goto l1046
				l1047:
					if !matchChar('"') {
						goto l1042
					}
					break
				default:
					goto l1042
				}
			}
			return true
		l1042:
			position = position0
			return false
		},
		/* 200 HtmlAttribute <- (((&[\-] '-') | (&[0-9A-Za-z] [A-Za-z0-9]))+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1048
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				default:
					if !matchClass(6) {
						goto l1048
					}
				}
			}
		l1049:
			{
				if position == len(p.Buffer) {
					goto l1050
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				default:
					if !matchClass(6) {
						goto l1050
					}
				}
			}
			goto l1049
		l1050:
			if !p.rules[ruleSpnl]() {
				goto l1048
			}
			{
				position1053 := position
				if !matchChar('=') {
					goto l1053
				}
				if !p.rules[ruleSpnl]() {
					goto l1053
				}
				if !p.rules[ruleQuoted]() {
					goto l1056
				}
				goto l1055
			l1056:
				if peekChar('>') {
					goto l1053
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1053
				}
			l1057:
				if peekChar('>') {
					goto l1058
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1058
				}
				goto l1057
			l1058:
			l1055:
				goto l1054
			l1053:
				position = position1053
			}
		l1054:
			if !p.rules[ruleSpnl]() {
				goto l1048
			}
			return true
		l1048:
			position = position0
			return false
		},
		/* 201 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
		func() bool {
			position0 := position
			if !matchString("<!--") {
				goto l1059
			}
		l1060:
			{
				position1061 := position
				if !matchString("-->") {
					goto l1062
				}
				goto l1061
			l1062:
				if !matchDot() {
					goto l1061
				}
				goto l1060
			l1061:
				position = position1061
			}
			if !matchString("-->") {
				goto l1059
			}
			return true
		l1059:
			position = position0
			return false
		},
		/* 202 HtmlTag <- ('<' Spnl '/'? [A-Za-z0-9]+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l1063
			}
			if !p.rules[ruleSpnl]() {
				goto l1063
			}
			matchChar('/')
			if !matchClass(6) {
				goto l1063
			}
		l1064:
			if !matchClass(6) {
				goto l1065
			}
			goto l1064
		l1065:
			if !p.rules[ruleSpnl]() {
				goto l1063
			}
		l1066:
			if !p.rules[ruleHtmlAttribute]() {
				goto l1067
			}
			goto l1066
		l1067:
			matchChar('/')
			if !p.rules[ruleSpnl]() {
				goto l1063
			}
			if !matchChar('>') {
				goto l1063
			}
			return true
		l1063:
			position = position0
			return false
		},
		/* 203 Eof <- !. */
		func() bool {
			if (position < len(p.Buffer)) {
				goto l1068
			}
			return true
		l1068:
			return false
		},
		/* 204 Spacechar <- ((&[\t] '\t') | (&[ ] ' ')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1069
				}
				switch p.Buffer[position] {
				case '\t':
					position++ // matchChar
					break
				case ' ':
					position++ // matchChar
					break
				default:
					goto l1069
				}
			}
			return true
		l1069:
			return false
		},
		/* 205 Nonspacechar <- (!Spacechar !Newline .) */
		func() bool {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto l1072
			}
			goto l1071
		l1072:
			if !p.rules[ruleNewline]() {
				goto l1073
			}
			goto l1071
		l1073:
			if !matchDot() {
				goto l1071
			}
			return true
		l1071:
			position = position0
			return false
		},
		/* 206 Newline <- ((&[\r] ('\r' '\n'?)) | (&[\n] '\n')) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1074
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
					goto l1074
				}
			}
			return true
		l1074:
			position = position0
			return false
		},
		/* 207 Sp <- Spacechar* */
		func() bool {
		l1077:
			if !p.rules[ruleSpacechar]() {
				goto l1078
			}
			goto l1077
		l1078:
			return true
		},
		/* 208 Spnl <- (Sp (Newline Sp)?) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1079
			}
			{
				position1080 := position
				if !p.rules[ruleNewline]() {
					goto l1080
				}
				if !p.rules[ruleSp]() {
					goto l1080
				}
				goto l1081
			l1080:
				position = position1080
			}
		l1081:
			return true
		l1079:
			position = position0
			return false
		},
		/* 209 SpecialChar <- ('\'' / '"' / ((&[\\] '\\') | (&[#] '#') | (&[!] '!') | (&[<] '<') | (&[)] ')') | (&[(] '(') | (&[\]] ']') | (&[\[] '[') | (&[&] '&') | (&[`] '`') | (&[_] '_') | (&[*] '*') | (&[\"\'\-.^] ExtendedSpecialChar))) */
		func() bool {
			if !matchChar('\'') {
				goto l1084
			}
			goto l1083
		l1084:
			if !matchChar('"') {
				goto l1085
			}
			goto l1083
		l1085:
			{
				if position == len(p.Buffer) {
					goto l1082
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
						goto l1082
					}
				}
			}
		l1083:
			return true
		l1082:
			return false
		},
		/* 210 NormalChar <- (!((&[\n\r] Newline) | (&[\t ] Spacechar) | (&[!-#&-*\-.<\[-`] SpecialChar)) .) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1088
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l1088
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l1088
					}
					break
				default:
					if !p.rules[ruleSpecialChar]() {
						goto l1088
					}
				}
			}
			goto l1087
		l1088:
			if !matchDot() {
				goto l1087
			}
			return true
		l1087:
			position = position0
			return false
		},
		/* 211 NonAlphanumeric <- [\000-\057\072-\100\133-\140\173-\177] */
		func() bool {
			if !matchClass(4) {
				goto l1090
			}
			return true
		l1090:
			return false
		},
		/* 212 Alphanumeric <- ((&[\377] '\377') | (&[\376] '\376') | (&[\375] '\375') | (&[\374] '\374') | (&[\373] '\373') | (&[\372] '\372') | (&[\371] '\371') | (&[\370] '\370') | (&[\367] '\367') | (&[\366] '\366') | (&[\365] '\365') | (&[\364] '\364') | (&[\363] '\363') | (&[\362] '\362') | (&[\361] '\361') | (&[\360] '\360') | (&[\357] '\357') | (&[\356] '\356') | (&[\355] '\355') | (&[\354] '\354') | (&[\353] '\353') | (&[\352] '\352') | (&[\351] '\351') | (&[\350] '\350') | (&[\347] '\347') | (&[\346] '\346') | (&[\345] '\345') | (&[\344] '\344') | (&[\343] '\343') | (&[\342] '\342') | (&[\341] '\341') | (&[\340] '\340') | (&[\337] '\337') | (&[\336] '\336') | (&[\335] '\335') | (&[\334] '\334') | (&[\333] '\333') | (&[\332] '\332') | (&[\331] '\331') | (&[\330] '\330') | (&[\327] '\327') | (&[\326] '\326') | (&[\325] '\325') | (&[\324] '\324') | (&[\323] '\323') | (&[\322] '\322') | (&[\321] '\321') | (&[\320] '\320') | (&[\317] '\317') | (&[\316] '\316') | (&[\315] '\315') | (&[\314] '\314') | (&[\313] '\313') | (&[\312] '\312') | (&[\311] '\311') | (&[\310] '\310') | (&[\307] '\307') | (&[\306] '\306') | (&[\305] '\305') | (&[\304] '\304') | (&[\303] '\303') | (&[\302] '\302') | (&[\301] '\301') | (&[\300] '\300') | (&[\277] '\277') | (&[\276] '\276') | (&[\275] '\275') | (&[\274] '\274') | (&[\273] '\273') | (&[\272] '\272') | (&[\271] '\271') | (&[\270] '\270') | (&[\267] '\267') | (&[\266] '\266') | (&[\265] '\265') | (&[\264] '\264') | (&[\263] '\263') | (&[\262] '\262') | (&[\261] '\261') | (&[\260] '\260') | (&[\257] '\257') | (&[\256] '\256') | (&[\255] '\255') | (&[\254] '\254') | (&[\253] '\253') | (&[\252] '\252') | (&[\251] '\251') | (&[\250] '\250') | (&[\247] '\247') | (&[\246] '\246') | (&[\245] '\245') | (&[\244] '\244') | (&[\243] '\243') | (&[\242] '\242') | (&[\241] '\241') | (&[\240] '\240') | (&[\237] '\237') | (&[\236] '\236') | (&[\235] '\235') | (&[\234] '\234') | (&[\233] '\233') | (&[\232] '\232') | (&[\231] '\231') | (&[\230] '\230') | (&[\227] '\227') | (&[\226] '\226') | (&[\225] '\225') | (&[\224] '\224') | (&[\223] '\223') | (&[\222] '\222') | (&[\221] '\221') | (&[\220] '\220') | (&[\217] '\217') | (&[\216] '\216') | (&[\215] '\215') | (&[\214] '\214') | (&[\213] '\213') | (&[\212] '\212') | (&[\211] '\211') | (&[\210] '\210') | (&[\207] '\207') | (&[\206] '\206') | (&[\205] '\205') | (&[\204] '\204') | (&[\203] '\203') | (&[\202] '\202') | (&[\201] '\201') | (&[\200] '\200') | (&[0-9A-Za-z] [0-9A-Za-z])) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1091
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
						goto l1091
					}
				}
			}
			return true
		l1091:
			return false
		},
		/* 213 AlphanumericAscii <- [A-Za-z0-9] */
		func() bool {
			if !matchClass(6) {
				goto l1093
			}
			return true
		l1093:
			return false
		},
		/* 214 Digit <- [0-9] */
		func() bool {
			if !matchClass(0) {
				goto l1094
			}
			return true
		l1094:
			return false
		},
		/* 215 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1095
			}
			if !matchChar('#') {
				goto l1095
			}
			if !matchClass(7) {
				goto l1095
			}
			if !matchClass(8) {
				goto l1095
			}
		l1096:
			if !matchClass(8) {
				goto l1097
			}
			goto l1096
		l1097:
			if !matchChar(';') {
				goto l1095
			}
			end = position
			return true
		l1095:
			position = position0
			return false
		},
		/* 216 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1098
			}
			if !matchChar('#') {
				goto l1098
			}
			if !matchClass(0) {
				goto l1098
			}
		l1099:
			if !matchClass(0) {
				goto l1100
			}
			goto l1099
		l1100:
			end = position
			if !matchChar(';') {
				goto l1098
			}
			end = position
			return true
		l1098:
			position = position0
			return false
		},
		/* 217 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1101
			}
			if !matchClass(6) {
				goto l1101
			}
		l1102:
			if !matchClass(6) {
				goto l1103
			}
			goto l1102
		l1103:
			if !matchChar(';') {
				goto l1101
			}
			end = position
			return true
		l1101:
			position = position0
			return false
		},
		/* 218 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() bool {
			if !matchString("   ") {
				goto l1106
			}
			goto l1105
		l1106:
			if !matchString("  ") {
				goto l1107
			}
			goto l1105
		l1107:
			if !matchChar(' ') {
				goto l1108
			}
			goto l1105
		l1108:
		l1105:
			return true
		},
		/* 219 Indent <- ((&[ ] '    ') | (&[\t] '\t')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1109
				}
				switch p.Buffer[position] {
				case ' ':
					position++
					if !matchString("   ") {
						goto l1109
					}
					break
				case '\t':
					position++ // matchChar
					break
				default:
					goto l1109
				}
			}
			return true
		l1109:
			return false
		},
		/* 220 IndentedLine <- (Indent Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1111
			}
			if !p.rules[ruleLine]() {
				goto l1111
			}
			return true
		l1111:
			position = position0
			return false
		},
		/* 221 OptionallyIndentedLine <- (Indent? Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1113
			}
		l1113:
			if !p.rules[ruleLine]() {
				goto l1112
			}
			return true
		l1112:
			position = position0
			return false
		},
		/* 222 StartList <- (&. { yy = nil }) */
		func() bool {
			if !(position < len(p.Buffer)) {
				goto l1115
			}
			do(92)
			return true
		l1115:
			return false
		},
		/* 223 Line <- (RawLine { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRawLine]() {
				goto l1116
			}
			do(93)
			return true
		l1116:
			position = position0
			return false
		},
		/* 224 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > !.)) */
		func() bool {
			position0 := position
			{
				position1118 := position
				begin = position
			l1120:
				if position == len(p.Buffer) {
					goto l1121
				}
				switch p.Buffer[position] {
				case '\r', '\n':
					goto l1121
				default:
					position++
				}
				goto l1120
			l1121:
				if !p.rules[ruleNewline]() {
					goto l1119
				}
				end = position
				goto l1118
			l1119:
				position = position1118
				begin = position
				if !matchDot() {
					goto l1117
				}
			l1122:
				if !matchDot() {
					goto l1123
				}
				goto l1122
			l1123:
				end = position
				if (position < len(p.Buffer)) {
					goto l1117
				}
			}
		l1118:
			return true
		l1117:
			position = position0
			return false
		},
		/* 225 SkipBlock <- (HtmlBlock / ((!'#' !SetextBottom1 !SetextBottom2 !BlankLine RawLine)+ BlankLine*) / BlankLine+ / RawLine) */
		func() bool {
			position0 := position
			{
				position1125 := position
				if !p.rules[ruleHtmlBlock]() {
					goto l1126
				}
				goto l1125
			l1126:
				if peekChar('#') {
					goto l1127
				}
				if !p.rules[ruleSetextBottom1]() {
					goto l1130
				}
				goto l1127
			l1130:
				if !p.rules[ruleSetextBottom2]() {
					goto l1131
				}
				goto l1127
			l1131:
				if !p.rules[ruleBlankLine]() {
					goto l1132
				}
				goto l1127
			l1132:
				if !p.rules[ruleRawLine]() {
					goto l1127
				}
			l1128:
				{
					position1129 := position
					if peekChar('#') {
						goto l1129
					}
					if !p.rules[ruleSetextBottom1]() {
						goto l1133
					}
					goto l1129
				l1133:
					if !p.rules[ruleSetextBottom2]() {
						goto l1134
					}
					goto l1129
				l1134:
					if !p.rules[ruleBlankLine]() {
						goto l1135
					}
					goto l1129
				l1135:
					if !p.rules[ruleRawLine]() {
						goto l1129
					}
					goto l1128
				l1129:
					position = position1129
				}
			l1136:
				if !p.rules[ruleBlankLine]() {
					goto l1137
				}
				goto l1136
			l1137:
				goto l1125
			l1127:
				position = position1125
				if !p.rules[ruleBlankLine]() {
					goto l1138
				}
			l1139:
				if !p.rules[ruleBlankLine]() {
					goto l1140
				}
				goto l1139
			l1140:
				goto l1125
			l1138:
				position = position1125
				if !p.rules[ruleRawLine]() {
					goto l1124
				}
			}
		l1125:
			return true
		l1124:
			position = position0
			return false
		},
		/* 226 ExtendedSpecialChar <- ((&[^] (&{p.extension.Notes} '^')) | (&[\"\'\-.] (&{p.extension.Smart} ((&[\"] '"') | (&[\'] '\'') | (&[\-] '-') | (&[.] '.'))))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1141
				}
				switch p.Buffer[position] {
				case '^':
					if !(p.extension.Notes) {
						goto l1141
					}
					if !matchChar('^') {
						goto l1141
					}
					break
				default:
					if !(p.extension.Smart) {
						goto l1141
					}
					{
						if position == len(p.Buffer) {
							goto l1141
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
							goto l1141
						}
					}
				}
			}
			return true
		l1141:
			position = position0
			return false
		},
		/* 227 Smart <- (&{p.extension.Smart} (SingleQuoted / ((&[\'] Apostrophe) | (&[\"] DoubleQuoted) | (&[\-] Dash) | (&[.] Ellipsis)))) */
		func() bool {
			if !(p.extension.Smart) {
				goto l1144
			}
			if !p.rules[ruleSingleQuoted]() {
				goto l1146
			}
			goto l1145
		l1146:
			{
				if position == len(p.Buffer) {
					goto l1144
				}
				switch p.Buffer[position] {
				case '\'':
					if !p.rules[ruleApostrophe]() {
						goto l1144
					}
					break
				case '"':
					if !p.rules[ruleDoubleQuoted]() {
						goto l1144
					}
					break
				case '-':
					if !p.rules[ruleDash]() {
						goto l1144
					}
					break
				case '.':
					if !p.rules[ruleEllipsis]() {
						goto l1144
					}
					break
				default:
					goto l1144
				}
			}
		l1145:
			return true
		l1144:
			return false
		},
		/* 228 Apostrophe <- ('\'' { yy = p.mkElem(APOSTROPHE) }) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1148
			}
			do(94)
			return true
		l1148:
			position = position0
			return false
		},
		/* 229 Ellipsis <- (('...' / '. . .') { yy = p.mkElem(ELLIPSIS) }) */
		func() bool {
			position0 := position
			if !matchString("...") {
				goto l1151
			}
			goto l1150
		l1151:
			if !matchString(". . .") {
				goto l1149
			}
		l1150:
			do(95)
			return true
		l1149:
			position = position0
			return false
		},
		/* 230 Dash <- (EmDash / EnDash) */
		func() bool {
			if !p.rules[ruleEmDash]() {
				goto l1154
			}
			goto l1153
		l1154:
			if !p.rules[ruleEnDash]() {
				goto l1152
			}
		l1153:
			return true
		l1152:
			return false
		},
		/* 231 EnDash <- ('-' &[0-9] { yy = p.mkElem(ENDASH) }) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l1155
			}
			if !peekClass(0) {
				goto l1155
			}
			do(96)
			return true
		l1155:
			position = position0
			return false
		},
		/* 232 EmDash <- (('---' / '--') { yy = p.mkElem(EMDASH) }) */
		func() bool {
			position0 := position
			if !matchString("---") {
				goto l1158
			}
			goto l1157
		l1158:
			if !matchString("--") {
				goto l1156
			}
		l1157:
			do(97)
			return true
		l1156:
			position = position0
			return false
		},
		/* 233 SingleQuoteStart <- ('\'' !((&[\n\r] Newline) | (&[\t ] Spacechar))) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1159
			}
			{
				if position == len(p.Buffer) {
					goto l1160
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l1160
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l1160
					}
					break
				default:
					goto l1160
				}
			}
			goto l1159
		l1160:
			return true
		l1159:
			position = position0
			return false
		},
		/* 234 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1162
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l1163
			}
			goto l1162
		l1163:
			return true
		l1162:
			position = position0
			return false
		},
		/* 235 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = p.mkList(SINGLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto l1164
			}
			if !p.rules[ruleStartList]() {
				goto l1164
			}
			doarg(yySet, -1)
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1167
			}
			goto l1164
		l1167:
			if !p.rules[ruleInline]() {
				goto l1164
			}
			doarg(yySet, -2)
			do(98)
		l1165:
			{
				position1166, thunkPosition1166 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto l1168
				}
				goto l1166
			l1168:
				if !p.rules[ruleInline]() {
					goto l1166
				}
				doarg(yySet, -2)
				do(98)
				goto l1165
			l1166:
				position, thunkPosition = position1166, thunkPosition1166
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1164
			}
			do(99)
			doarg(yyPop, 2)
			return true
		l1164:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 236 DoubleQuoteStart <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1169
			}
			return true
		l1169:
			return false
		},
		/* 237 DoubleQuoteEnd <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1170
			}
			return true
		l1170:
			return false
		},
		/* 238 DoubleQuoted <- ('"' StartList (!'"' Inline { a = cons(b, a) })+ '"' { yy = p.mkList(DOUBLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('"') {
				goto l1171
			}
			if !p.rules[ruleStartList]() {
				goto l1171
			}
			doarg(yySet, -2)
			if peekChar('"') {
				goto l1171
			}
			if !p.rules[ruleInline]() {
				goto l1171
			}
			doarg(yySet, -1)
			do(100)
		l1172:
			{
				position1173, thunkPosition1173 := position, thunkPosition
				if peekChar('"') {
					goto l1173
				}
				if !p.rules[ruleInline]() {
					goto l1173
				}
				doarg(yySet, -1)
				do(100)
				goto l1172
			l1173:
				position, thunkPosition = position1173, thunkPosition1173
			}
			if !matchChar('"') {
				goto l1171
			}
			do(101)
			doarg(yyPop, 2)
			return true
		l1171:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 239 NoteReference <- (&{p.extension.Notes} RawNoteReference {
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
				goto l1174
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1174
			}
			doarg(yySet, -1)
			do(102)
			doarg(yyPop, 1)
			return true
		l1174:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 240 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !matchString("[^") {
				goto l1175
			}
			begin = position
			if !p.rules[ruleNewline]() {
				goto l1178
			}
			goto l1175
		l1178:
			if peekChar(']') {
				goto l1175
			}
			if !matchDot() {
				goto l1175
			}
		l1176:
			{
				position1177 := position
				if !p.rules[ruleNewline]() {
					goto l1179
				}
				goto l1177
			l1179:
				if peekChar(']') {
					goto l1177
				}
				if !matchDot() {
					goto l1177
				}
				goto l1176
			l1177:
				position = position1177
			}
			end = position
			if !matchChar(']') {
				goto l1175
			}
			do(103)
			return true
		l1175:
			position = position0
			return false
		},
		/* 241 Note <- (&{p.extension.Notes} NonindentSpace RawNoteReference ':' Sp StartList (RawNoteBlock { a = cons(yy, a) }) (&Indent RawNoteBlock { a = cons(yy, a) })* {   yy = p.mkList(NOTE, a)
                    yy.contents.str = ref.contents.str
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !(p.extension.Notes) {
				goto l1180
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l1180
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1180
			}
			doarg(yySet, -2)
			if !matchChar(':') {
				goto l1180
			}
			if !p.rules[ruleSp]() {
				goto l1180
			}
			if !p.rules[ruleStartList]() {
				goto l1180
			}
			doarg(yySet, -1)
			if !p.rules[ruleRawNoteBlock]() {
				goto l1180
			}
			do(104)
		l1181:
			{
				position1182, thunkPosition1182 := position, thunkPosition
				{
					position1183 := position
					if !p.rules[ruleIndent]() {
						goto l1182
					}
					position = position1183
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto l1182
				}
				do(105)
				goto l1181
			l1182:
				position, thunkPosition = position1182, thunkPosition1182
			}
			do(106)
			doarg(yyPop, 2)
			return true
		l1180:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 242 InlineNote <- (&{p.extension.Notes} '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = p.mkList(NOTE, a)
                  yy.contents.str = "" }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto l1184
			}
			if !matchString("^[") {
				goto l1184
			}
			if !p.rules[ruleStartList]() {
				goto l1184
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto l1184
			}
			if !p.rules[ruleInline]() {
				goto l1184
			}
			do(107)
		l1185:
			{
				position1186 := position
				if peekChar(']') {
					goto l1186
				}
				if !p.rules[ruleInline]() {
					goto l1186
				}
				do(107)
				goto l1185
			l1186:
				position = position1186
			}
			if !matchChar(']') {
				goto l1184
			}
			do(108)
			doarg(yyPop, 1)
			return true
		l1184:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 243 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1187
			}
			doarg(yySet, -2)
		l1188:
			{
				position1189, thunkPosition1189 := position, thunkPosition
				{
					position1190, thunkPosition1190 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto l1191
					}
					doarg(yySet, -1)
					do(109)
					goto l1190
				l1191:
					position, thunkPosition = position1190, thunkPosition1190
					if !p.rules[ruleSkipBlock]() {
						goto l1189
					}
				}
			l1190:
				goto l1188
			l1189:
				position, thunkPosition = position1189, thunkPosition1189
			}
			do(110)
			if !(commit(thunkPosition0)) {
				goto l1187
			}
			doarg(yyPop, 2)
			return true
		l1187:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 244 RawNoteBlock <- (StartList (!BlankLine OptionallyIndentedLine { a = cons(yy, a) })+ (< BlankLine* > { a = cons(p.mkString(yytext), a) }) {   yy = p.mkStringFromList(a, true)
                    yy.key = RAW
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1192
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l1195
			}
			goto l1192
		l1195:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l1192
			}
			do(111)
		l1193:
			{
				position1194 := position
				if !p.rules[ruleBlankLine]() {
					goto l1196
				}
				goto l1194
			l1196:
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto l1194
				}
				do(111)
				goto l1193
			l1194:
				position = position1194
			}
			begin = position
		l1197:
			if !p.rules[ruleBlankLine]() {
				goto l1198
			}
			goto l1197
		l1198:
			end = position
			do(112)
			do(113)
			doarg(yyPop, 1)
			return true
		l1192:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 245 DefinitionList <- (&{p.extension.Dlists} StartList (Definition { a = cons(yy, a) })+ { yy = p.mkList(DEFINITIONLIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Dlists) {
				goto l1199
			}
			if !p.rules[ruleStartList]() {
				goto l1199
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto l1199
			}
			do(114)
		l1200:
			{
				position1201, thunkPosition1201 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto l1201
				}
				do(114)
				goto l1200
			l1201:
				position, thunkPosition = position1201, thunkPosition1201
			}
			do(115)
			doarg(yyPop, 1)
			return true
		l1199:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 246 Definition <- (&(NonindentSpace !Defmark Nonspacechar RawLine BlankLine? Defmark) StartList (DListTitle { a = cons(yy, a) })+ (DefTight / DefLoose) {
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			} { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1203 := position
				if !p.rules[ruleNonindentSpace]() {
					goto l1202
				}
				if !p.rules[ruleDefmark]() {
					goto l1204
				}
				goto l1202
			l1204:
				if !p.rules[ruleNonspacechar]() {
					goto l1202
				}
				if !p.rules[ruleRawLine]() {
					goto l1202
				}
				if !p.rules[ruleBlankLine]() {
					goto l1205
				}
			l1205:
				if !p.rules[ruleDefmark]() {
					goto l1202
				}
				position = position1203
			}
			if !p.rules[ruleStartList]() {
				goto l1202
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto l1202
			}
			do(116)
		l1207:
			{
				position1208, thunkPosition1208 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto l1208
				}
				do(116)
				goto l1207
			l1208:
				position, thunkPosition = position1208, thunkPosition1208
			}
			if !p.rules[ruleDefTight]() {
				goto l1210
			}
			goto l1209
		l1210:
			if !p.rules[ruleDefLoose]() {
				goto l1202
			}
		l1209:
			do(117)
			do(118)
			doarg(yyPop, 1)
			return true
		l1202:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 247 DListTitle <- (NonindentSpace !Defmark &Nonspacechar StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline {	yy = p.mkList(LIST, a)
				yy.key = DEFTITLE
			}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l1211
			}
			if !p.rules[ruleDefmark]() {
				goto l1212
			}
			goto l1211
		l1212:
			{
				position1213 := position
				if !p.rules[ruleNonspacechar]() {
					goto l1211
				}
				position = position1213
			}
			if !p.rules[ruleStartList]() {
				goto l1211
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l1216
			}
			goto l1211
		l1216:
			if !p.rules[ruleInline]() {
				goto l1211
			}
			do(119)
		l1214:
			{
				position1215 := position
				if !p.rules[ruleEndline]() {
					goto l1217
				}
				goto l1215
			l1217:
				if !p.rules[ruleInline]() {
					goto l1215
				}
				do(119)
				goto l1214
			l1215:
				position = position1215
			}
			if !p.rules[ruleSp]() {
				goto l1211
			}
			if !p.rules[ruleNewline]() {
				goto l1211
			}
			do(120)
			doarg(yyPop, 1)
			return true
		l1211:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 248 DefTight <- (&Defmark ListTight) */
		func() bool {
			{
				position1219 := position
				if !p.rules[ruleDefmark]() {
					goto l1218
				}
				position = position1219
			}
			if !p.rules[ruleListTight]() {
				goto l1218
			}
			return true
		l1218:
			return false
		},
		/* 249 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l1220
			}
			{
				position1221 := position
				if !p.rules[ruleDefmark]() {
					goto l1220
				}
				position = position1221
			}
			if !p.rules[ruleListLoose]() {
				goto l1220
			}
			return true
		l1220:
			position = position0
			return false
		},
		/* 250 Defmark <- (NonindentSpace ((&[~] '~') | (&[:] ':')) Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l1222
			}
			{
				if position == len(p.Buffer) {
					goto l1222
				}
				switch p.Buffer[position] {
				case '~':
					position++ // matchChar
					break
				case ':':
					position++ // matchChar
					break
				default:
					goto l1222
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l1222
			}
		l1224:
			if !p.rules[ruleSpacechar]() {
				goto l1225
			}
			goto l1224
		l1225:
			return true
		l1222:
			position = position0
			return false
		},
		/* 251 DefMarker <- (&{p.extension.Dlists} Defmark) */
		func() bool {
			if !(p.extension.Dlists) {
				goto l1226
			}
			if !p.rules[ruleDefmark]() {
				goto l1226
			}
			return true
		l1226:
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
