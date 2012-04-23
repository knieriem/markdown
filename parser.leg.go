
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

type Doc struct {
	parser		*yyParser
	extension	Options

	tree				*element	/* Results of parse. */
	references			*element	/* List of link references found. */
	notes				*element	/* List of footnotes found. */
}


const (
	ruleDoc = iota
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
	*Doc
	Buffer string
	Min, Max int
	rules [249]func() bool
	ResetBuffer	func(string) string
}

func (p *yyParser) Parse(ruleId int) (err error) {
	if p.rules[ruleId]() {
		return
	}
	return p.parseErr()
}

type ErrPos struct {
	Line, Pos int
}

func	(e *ErrPos) String() string {
	return fmt.Sprintf("%d:%d", e.Line, e.Pos)
}

type UnexpectedCharError struct {
	After, At	ErrPos
	Char	byte
}

func (e *UnexpectedCharError) Error() string {
	return fmt.Sprintf("%v: unexpected character '%c'", &e.At, e.Char)
}

type UnexpectedEOFError struct {
	After ErrPos
}

func (e *UnexpectedEOFError) Error() string {
	return fmt.Sprintf("%v: unexpected end of file", &e.After)
}

func (p *yyParser) parseErr() (err error) {
	var pos, after ErrPos
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
		err = &UnexpectedEOFError{after}
	} else {
		err = &UnexpectedCharError{after, pos, p.Buffer[p.Max]}
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
		/* 2 Para */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a; yy.key = PARA 
			yyval[yyp-1] = a
		},
		/* 3 Plain */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a; yy.key = PLAIN 
			yyval[yyp-1] = a
		},
		/* 4 AtxStart */
		func(yytext string, _ int) {
			 yy = mk_element(H1 + (len(yytext) - 1)) 
		},
		/* 5 AtxHeading */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = s
			yyval[yyp-2] = a
		},
		/* 6 AtxHeading */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = mk_list(s.key, a)
              s = nil 
			yyval[yyp-2] = a
			yyval[yyp-1] = s
		},
		/* 7 SetextHeading1 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 8 SetextHeading1 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(H1, a) 
			yyval[yyp-1] = a
		},
		/* 9 SetextHeading2 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 10 SetextHeading2 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(H2, a) 
			yyval[yyp-1] = a
		},
		/* 11 BlockQuote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			  yy = mk_element(BLOCKQUOTE)
                yy.children = a
             
			yyval[yyp-1] = a
		},
		/* 12 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
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
			 a = cons(mk_str("\n"), a) 
			yyval[yyp-1] = a
		},
		/* 15 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   yy = mk_str_from_list(a, true)
                     yy.key = RAW
                 
			yyval[yyp-1] = a
		},
		/* 16 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(mk_str("\n"), a) 
			yyval[yyp-1] = a
		},
		/* 17 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 18 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_str_from_list(a, false) 
			yyval[yyp-1] = a
		},
		/* 19 Verbatim */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 20 Verbatim */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_str_from_list(a, false)
                 yy.key = VERBATIM 
			yyval[yyp-1] = a
		},
		/* 21 HorizontalRule */
		func(yytext string, _ int) {
			 yy = mk_element(HRULE) 
		},
		/* 22 BulletList */
		func(yytext string, _ int) {
			 yy.key = BULLETLIST 
		},
		/* 23 ListTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 24 ListTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 25 ListLoose */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			
                  li := b.children
                  li.contents.str += "\n\n"
                  a = cons(b, a)
              
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 26 ListLoose */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 27 ListItem */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
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
			
               raw := mk_str_from_list(a, false)
               raw.key = RAW
               yy = mk_element(LISTITEM)
               yy.children = raw
            
			yyval[yyp-1] = a
		},
		/* 30 ListItemTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
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
			
               raw := mk_str_from_list(a, false)
               raw.key = RAW
               yy = mk_element(LISTITEM)
               yy.children = raw
            
			yyval[yyp-1] = a
		},
		/* 33 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
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
			 yy = mk_str_from_list(a, false) 
			yyval[yyp-1] = a
		},
		/* 36 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   if len(yytext) == 0 {
                                   a = cons(mk_str("\001"), a) // block separator
                              } else {
                                   a = cons(mk_str(yytext), a)
                              }
                          
			yyval[yyp-1] = a
		},
		/* 37 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 38 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			  yy = mk_str_from_list(a, false) 
			yyval[yyp-1] = a
		},
		/* 39 OrderedList */
		func(yytext string, _ int) {
			 yy.key = ORDEREDLIST 
		},
		/* 40 HtmlBlock */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTMLBLOCK
                }
            
		},
		/* 41 StyleBlock */
		func(yytext string, _ int) {
			   if p.extension.FilterStyles {
                        yy = mk_list(LIST, nil)
                    } else {
                        yy = mk_str(yytext)
                        yy.key = HTMLBLOCK
                    }
                
		},
		/* 42 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = c
		},
		/* 43 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			 a = cons(c, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = c
		},
		/* 44 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = c
		},
		/* 45 Space */
		func(yytext string, _ int) {
			 yy = mk_str(" ")
          yy.key = SPACE 
		},
		/* 46 Str */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 47 EscapedChar */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 48 Entity */
		func(yytext string, _ int) {
			 yy = mk_str(yytext); yy.key = HTML 
		},
		/* 49 NormalEndline */
		func(yytext string, _ int) {
			 yy = mk_str("\n")
                    yy.key = SPACE 
		},
		/* 50 TerminalEndline */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 51 LineBreak */
		func(yytext string, _ int) {
			 yy = mk_element(LINEBREAK) 
		},
		/* 52 Symbol */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 53 UlOrStarLine */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 54 OneStarClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 55 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
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
			 yy = mk_list(EMPH, a) 
			yyval[yyp-1] = a
		},
		/* 58 OneUlClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 59 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
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
			 yy = mk_list(EMPH, a) 
			yyval[yyp-1] = a
		},
		/* 62 TwoStarClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 63 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
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
			 yy = mk_list(STRONG, a) 
			yyval[yyp-1] = a
		},
		/* 66 TwoUlClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 67 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
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
			 yy = mk_list(STRONG, a) 
			yyval[yyp-1] = a
		},
		/* 70 Image */
		func(yytext string, _ int) {
				if yy.key == LINK {
			yy.key = IMAGE
		} else {
			result := yy
			yy.children = cons(mk_str("!"), result.children)
		}
	
		},
		/* 71 ReferenceLinkDouble */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			
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
                       
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 72 ReferenceLinkSingle */
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
		/* 73 ExplicitLink */
		func(yytext string, _ int) {
			l := yyval[yyp-1]
			s := yyval[yyp-2]
			t := yyval[yyp-3]
			 yy = mk_link(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil 
			yyval[yyp-1] = l
			yyval[yyp-2] = s
			yyval[yyp-3] = t
		},
		/* 74 Source */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 75 Title */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 76 AutoLinkUrl */
		func(yytext string, _ int) {
			   yy = mk_link(mk_str(yytext), yytext, "") 
		},
		/* 77 AutoLinkEmail */
		func(yytext string, _ int) {
			
                    yy = mk_link(mk_str(yytext), "mailto:"+yytext, "")
                
		},
		/* 78 Reference */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			l := yyval[yyp-2]
			t := yyval[yyp-3]
			 yy = mk_link(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE 
			yyval[yyp-1] = s
			yyval[yyp-2] = l
			yyval[yyp-3] = t
		},
		/* 79 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 80 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 81 RefSrc */
		func(yytext string, _ int) {
			 yy = mk_str(yytext)
           yy.key = HTML 
		},
		/* 82 RefTitle */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 83 References */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 84 References */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			 p.references = reverse(a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 85 Code */
		func(yytext string, _ int) {
			 yy = mk_str(yytext); yy.key = CODE 
		},
		/* 86 RawHtml */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTML
                }
            
		},
		/* 87 StartList */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 88 Line */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 89 Apostrophe */
		func(yytext string, _ int) {
			 yy = mk_element(APOSTROPHE) 
		},
		/* 90 Ellipsis */
		func(yytext string, _ int) {
			 yy = mk_element(ELLIPSIS) 
		},
		/* 91 EnDash */
		func(yytext string, _ int) {
			 yy = mk_element(ENDASH) 
		},
		/* 92 EmDash */
		func(yytext string, _ int) {
			 yy = mk_element(EMDASH) 
		},
		/* 93 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 94 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 yy = mk_list(SINGLEQUOTED, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 95 DoubleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 96 DoubleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 yy = mk_list(DOUBLEQUOTED, a) 
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 97 NoteReference */
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
		/* 98 RawNoteReference */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 99 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = ref
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
			   yy = mk_list(NOTE, a)
                    yy.contents.str = ref.contents.str
                
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 102 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 103 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(NOTE, a)
                  yy.contents.str = "" 
			yyval[yyp-1] = a
		},
		/* 104 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 105 Notes */
		func(yytext string, _ int) {
			b := yyval[yyp-2]
			a := yyval[yyp-1]
			 p.notes = reverse(a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 106 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 107 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(mk_str(yytext), a) 
			yyval[yyp-1] = a
		},
		/* 108 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   yy = mk_str_from_list(a, true)
                    yy.key = RAW
                
			yyval[yyp-1] = a
		},
		/* 109 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 110 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(DEFINITIONLIST, a) 
			yyval[yyp-1] = a
		},
		/* 111 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 112 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			
			yyval[yyp-1] = a
		},
		/* 113 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 114 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 115 DListTitle */
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
		yyPush = 116 + iota
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
		/* 1 Block <- (BlankLine* (BlockQuote / Verbatim / Note / Reference / HorizontalRule / Heading / DefinitionList / OrderedList / BulletList / HtmlBlock / StyleBlock / Para / Plain)) */
		func() bool {
			position0 := position
		l4:
			if !p.rules[ruleBlankLine]() {
				goto l5
			}
			goto l4
		l5:
			if !p.rules[ruleBlockQuote]() {
				goto l7
			}
			goto l6
		l7:
			if !p.rules[ruleVerbatim]() {
				goto l8
			}
			goto l6
		l8:
			if !p.rules[ruleNote]() {
				goto l9
			}
			goto l6
		l9:
			if !p.rules[ruleReference]() {
				goto l10
			}
			goto l6
		l10:
			if !p.rules[ruleHorizontalRule]() {
				goto l11
			}
			goto l6
		l11:
			if !p.rules[ruleHeading]() {
				goto l12
			}
			goto l6
		l12:
			if !p.rules[ruleDefinitionList]() {
				goto l13
			}
			goto l6
		l13:
			if !p.rules[ruleOrderedList]() {
				goto l14
			}
			goto l6
		l14:
			if !p.rules[ruleBulletList]() {
				goto l15
			}
			goto l6
		l15:
			if !p.rules[ruleHtmlBlock]() {
				goto l16
			}
			goto l6
		l16:
			if !p.rules[ruleStyleBlock]() {
				goto l17
			}
			goto l6
		l17:
			if !p.rules[rulePara]() {
				goto l18
			}
			goto l6
		l18:
			if !p.rules[rulePlain]() {
				goto l3
			}
		l6:
			return true
		l3:
			position = position0
			return false
		},
		/* 2 Para <- (NonindentSpace Inlines BlankLine+ { yy = a; yy.key = PARA }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l19
			}
			if !p.rules[ruleInlines]() {
				goto l19
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l19
			}
		l20:
			if !p.rules[ruleBlankLine]() {
				goto l21
			}
			goto l20
		l21:
			do(2)
			doarg(yyPop, 1)
			return true
		l19:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 3 Plain <- (Inlines { yy = a; yy.key = PLAIN }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleInlines]() {
				goto l22
			}
			doarg(yySet, -1)
			do(3)
			doarg(yyPop, 1)
			return true
		l22:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 4 AtxInline <- (!Newline !(Sp? '#'* Sp Newline) Inline) */
		func() bool {
			position0 := position
			if !p.rules[ruleNewline]() {
				goto l24
			}
			goto l23
		l24:
			{
				position25 := position
				if !p.rules[ruleSp]() {
					goto l26
				}
			l26:
			l28:
				if !matchChar('#') {
					goto l29
				}
				goto l28
			l29:
				if !p.rules[ruleSp]() {
					goto l25
				}
				if !p.rules[ruleNewline]() {
					goto l25
				}
				goto l23
			l25:
				position = position25
			}
			if !p.rules[ruleInline]() {
				goto l23
			}
			return true
		l23:
			position = position0
			return false
		},
		/* 5 AtxStart <- (&'#' < ('######' / '#####' / '####' / '###' / '##' / '#') > { yy = mk_element(H1 + (len(yytext) - 1)) }) */
		func() bool {
			position0 := position
			if !peekChar('#') {
				goto l30
			}
			begin = position
			if !matchString("######") {
				goto l32
			}
			goto l31
		l32:
			if !matchString("#####") {
				goto l33
			}
			goto l31
		l33:
			if !matchString("####") {
				goto l34
			}
			goto l31
		l34:
			if !matchString("###") {
				goto l35
			}
			goto l31
		l35:
			if !matchString("##") {
				goto l36
			}
			goto l31
		l36:
			if !matchChar('#') {
				goto l30
			}
		l31:
			end = position
			do(4)
			return true
		l30:
			position = position0
			return false
		},
		/* 6 AtxHeading <- (AtxStart Sp? StartList (AtxInline { a = cons(yy, a) })+ (Sp? '#'* Sp)? Newline { yy = mk_list(s.key, a)
              s = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleAtxStart]() {
				goto l37
			}
			doarg(yySet, -1)
			if !p.rules[ruleSp]() {
				goto l38
			}
		l38:
			if !p.rules[ruleStartList]() {
				goto l37
			}
			doarg(yySet, -2)
			if !p.rules[ruleAtxInline]() {
				goto l37
			}
			do(5)
		l40:
			{
				position41 := position
				if !p.rules[ruleAtxInline]() {
					goto l41
				}
				do(5)
				goto l40
			l41:
				position = position41
			}
			{
				position42 := position
				if !p.rules[ruleSp]() {
					goto l44
				}
			l44:
			l46:
				if !matchChar('#') {
					goto l47
				}
				goto l46
			l47:
				if !p.rules[ruleSp]() {
					goto l42
				}
				goto l43
			l42:
				position = position42
			}
		l43:
			if !p.rules[ruleNewline]() {
				goto l37
			}
			do(6)
			doarg(yyPop, 2)
			return true
		l37:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 7 SetextHeading <- (SetextHeading1 / SetextHeading2) */
		func() bool {
			if !p.rules[ruleSetextHeading1]() {
				goto l50
			}
			goto l49
		l50:
			if !p.rules[ruleSetextHeading2]() {
				goto l48
			}
		l49:
			return true
		l48:
			return false
		},
		/* 8 SetextBottom1 <- ('===' '='* Newline) */
		func() bool {
			position0 := position
			if !matchString("===") {
				goto l51
			}
		l52:
			if !matchChar('=') {
				goto l53
			}
			goto l52
		l53:
			if !p.rules[ruleNewline]() {
				goto l51
			}
			return true
		l51:
			position = position0
			return false
		},
		/* 9 SetextBottom2 <- ('---' '-'* Newline) */
		func() bool {
			position0 := position
			if !matchString("---") {
				goto l54
			}
		l55:
			if !matchChar('-') {
				goto l56
			}
			goto l55
		l56:
			if !p.rules[ruleNewline]() {
				goto l54
			}
			return true
		l54:
			position = position0
			return false
		},
		/* 10 SetextHeading1 <- (&(RawLine SetextBottom1) StartList (!Endline Inline { a = cons(yy, a) })+ Newline SetextBottom1 { yy = mk_list(H1, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position58 := position
				if !p.rules[ruleRawLine]() {
					goto l57
				}
				if !p.rules[ruleSetextBottom1]() {
					goto l57
				}
				position = position58
			}
			if !p.rules[ruleStartList]() {
				goto l57
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l61
			}
			goto l57
		l61:
			if !p.rules[ruleInline]() {
				goto l57
			}
			do(7)
		l59:
			{
				position60 := position
				if !p.rules[ruleEndline]() {
					goto l62
				}
				goto l60
			l62:
				if !p.rules[ruleInline]() {
					goto l60
				}
				do(7)
				goto l59
			l60:
				position = position60
			}
			if !p.rules[ruleNewline]() {
				goto l57
			}
			if !p.rules[ruleSetextBottom1]() {
				goto l57
			}
			do(8)
			doarg(yyPop, 1)
			return true
		l57:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 11 SetextHeading2 <- (&(RawLine SetextBottom2) StartList (!Endline Inline { a = cons(yy, a) })+ Newline SetextBottom2 { yy = mk_list(H2, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position64 := position
				if !p.rules[ruleRawLine]() {
					goto l63
				}
				if !p.rules[ruleSetextBottom2]() {
					goto l63
				}
				position = position64
			}
			if !p.rules[ruleStartList]() {
				goto l63
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l67
			}
			goto l63
		l67:
			if !p.rules[ruleInline]() {
				goto l63
			}
			do(9)
		l65:
			{
				position66 := position
				if !p.rules[ruleEndline]() {
					goto l68
				}
				goto l66
			l68:
				if !p.rules[ruleInline]() {
					goto l66
				}
				do(9)
				goto l65
			l66:
				position = position66
			}
			if !p.rules[ruleNewline]() {
				goto l63
			}
			if !p.rules[ruleSetextBottom2]() {
				goto l63
			}
			do(10)
			doarg(yyPop, 1)
			return true
		l63:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 12 Heading <- (AtxHeading / SetextHeading) */
		func() bool {
			if !p.rules[ruleAtxHeading]() {
				goto l71
			}
			goto l70
		l71:
			if !p.rules[ruleSetextHeading]() {
				goto l69
			}
		l70:
			return true
		l69:
			return false
		},
		/* 13 BlockQuote <- (BlockQuoteRaw {  yy = mk_element(BLOCKQUOTE)
                yy.children = a
             }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleBlockQuoteRaw]() {
				goto l72
			}
			doarg(yySet, -1)
			do(11)
			doarg(yyPop, 1)
			return true
		l72:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 14 BlockQuoteRaw <- (StartList ('>' ' '? Line { a = cons(yy, a) } (!'>' !BlankLine Line { a = cons(yy, a) })* (BlankLine { a = cons(mk_str("\n"), a) })*)+ {   yy = mk_str_from_list(a, true)
                     yy.key = RAW
                 }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l73
			}
			doarg(yySet, -1)
			if !matchChar('>') {
				goto l73
			}
			matchChar(' ')
			if !p.rules[ruleLine]() {
				goto l73
			}
			do(12)
		l76:
			{
				position77, thunkPosition77 := position, thunkPosition
				if peekChar('>') {
					goto l77
				}
				if !p.rules[ruleBlankLine]() {
					goto l78
				}
				goto l77
			l78:
				if !p.rules[ruleLine]() {
					goto l77
				}
				do(13)
				goto l76
			l77:
				position, thunkPosition = position77, thunkPosition77
			}
		l79:
			{
				position80 := position
				if !p.rules[ruleBlankLine]() {
					goto l80
				}
				do(14)
				goto l79
			l80:
				position = position80
			}
		l74:
			{
				position75, thunkPosition75 := position, thunkPosition
				if !matchChar('>') {
					goto l75
				}
				matchChar(' ')
				if !p.rules[ruleLine]() {
					goto l75
				}
				do(12)
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
					do(13)
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
					do(14)
					goto l84
				l85:
					position = position85
				}
				goto l74
			l75:
				position, thunkPosition = position75, thunkPosition75
			}
			do(15)
			doarg(yyPop, 1)
			return true
		l73:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 NonblankIndentedLine <- (!BlankLine IndentedLine) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l87
			}
			goto l86
		l87:
			if !p.rules[ruleIndentedLine]() {
				goto l86
			}
			return true
		l86:
			position = position0
			return false
		},
		/* 16 VerbatimChunk <- (StartList (BlankLine { a = cons(mk_str("\n"), a) })* (NonblankIndentedLine { a = cons(yy, a) })+ { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l88
			}
			doarg(yySet, -1)
		l89:
			{
				position90 := position
				if !p.rules[ruleBlankLine]() {
					goto l90
				}
				do(16)
				goto l89
			l90:
				position = position90
			}
			if !p.rules[ruleNonblankIndentedLine]() {
				goto l88
			}
			do(17)
		l91:
			{
				position92 := position
				if !p.rules[ruleNonblankIndentedLine]() {
					goto l92
				}
				do(17)
				goto l91
			l92:
				position = position92
			}
			do(18)
			doarg(yyPop, 1)
			return true
		l88:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 17 Verbatim <- (StartList (VerbatimChunk { a = cons(yy, a) })+ { yy = mk_str_from_list(a, false)
                 yy.key = VERBATIM }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l93
			}
			doarg(yySet, -1)
			if !p.rules[ruleVerbatimChunk]() {
				goto l93
			}
			do(19)
		l94:
			{
				position95, thunkPosition95 := position, thunkPosition
				if !p.rules[ruleVerbatimChunk]() {
					goto l95
				}
				do(19)
				goto l94
			l95:
				position, thunkPosition = position95, thunkPosition95
			}
			do(20)
			doarg(yyPop, 1)
			return true
		l93:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 18 HorizontalRule <- (NonindentSpace ((&[_] ('_' Sp '_' Sp '_' (Sp '_')*)) | (&[\-] ('-' Sp '-' Sp '-' (Sp '-')*)) | (&[*] ('*' Sp '*' Sp '*' (Sp '*')*))) Sp Newline BlankLine+ { yy = mk_element(HRULE) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l96
			}
			{
				if position == len(p.Buffer) {
					goto l96
				}
				switch p.Buffer[position] {
				case '_':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l96
					}
					if !matchChar('_') {
						goto l96
					}
					if !p.rules[ruleSp]() {
						goto l96
					}
					if !matchChar('_') {
						goto l96
					}
				l98:
					{
						position99 := position
						if !p.rules[ruleSp]() {
							goto l99
						}
						if !matchChar('_') {
							goto l99
						}
						goto l98
					l99:
						position = position99
					}
					break
				case '-':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l96
					}
					if !matchChar('-') {
						goto l96
					}
					if !p.rules[ruleSp]() {
						goto l96
					}
					if !matchChar('-') {
						goto l96
					}
				l100:
					{
						position101 := position
						if !p.rules[ruleSp]() {
							goto l101
						}
						if !matchChar('-') {
							goto l101
						}
						goto l100
					l101:
						position = position101
					}
					break
				case '*':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l96
					}
					if !matchChar('*') {
						goto l96
					}
					if !p.rules[ruleSp]() {
						goto l96
					}
					if !matchChar('*') {
						goto l96
					}
				l102:
					{
						position103 := position
						if !p.rules[ruleSp]() {
							goto l103
						}
						if !matchChar('*') {
							goto l103
						}
						goto l102
					l103:
						position = position103
					}
					break
				default:
					goto l96
				}
			}
			if !p.rules[ruleSp]() {
				goto l96
			}
			if !p.rules[ruleNewline]() {
				goto l96
			}
			if !p.rules[ruleBlankLine]() {
				goto l96
			}
		l104:
			if !p.rules[ruleBlankLine]() {
				goto l105
			}
			goto l104
		l105:
			do(21)
			return true
		l96:
			position = position0
			return false
		},
		/* 19 Bullet <- (!HorizontalRule NonindentSpace ((&[\-] '-') | (&[*] '*') | (&[+] '+')) Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHorizontalRule]() {
				goto l107
			}
			goto l106
		l107:
			if !p.rules[ruleNonindentSpace]() {
				goto l106
			}
			{
				if position == len(p.Buffer) {
					goto l106
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
					goto l106
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l106
			}
		l109:
			if !p.rules[ruleSpacechar]() {
				goto l110
			}
			goto l109
		l110:
			return true
		l106:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 20 BulletList <- (&Bullet (ListTight / ListLoose) { yy.key = BULLETLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position112 := position
				if !p.rules[ruleBullet]() {
					goto l111
				}
				position = position112
			}
			if !p.rules[ruleListTight]() {
				goto l114
			}
			goto l113
		l114:
			if !p.rules[ruleListLoose]() {
				goto l111
			}
		l113:
			do(22)
			return true
		l111:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 21 ListTight <- (StartList (ListItemTight { a = cons(yy, a) })+ BlankLine* !((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l115
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItemTight]() {
				goto l115
			}
			do(23)
		l116:
			{
				position117, thunkPosition117 := position, thunkPosition
				if !p.rules[ruleListItemTight]() {
					goto l117
				}
				do(23)
				goto l116
			l117:
				position, thunkPosition = position117, thunkPosition117
			}
		l118:
			if !p.rules[ruleBlankLine]() {
				goto l119
			}
			goto l118
		l119:
			{
				if position == len(p.Buffer) {
					goto l120
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l120
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l120
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l120
					}
				}
			}
			goto l115
		l120:
			do(24)
			doarg(yyPop, 1)
			return true
		l115:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 22 ListLoose <- (StartList (ListItem BlankLine* {
                  li := b.children
                  li.contents.str += "\n\n"
                  a = cons(b, a)
              })+ { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l122
			}
			doarg(yySet, -2)
			if !p.rules[ruleListItem]() {
				goto l122
			}
			doarg(yySet, -1)
		l125:
			if !p.rules[ruleBlankLine]() {
				goto l126
			}
			goto l125
		l126:
			do(25)
		l123:
			{
				position124, thunkPosition124 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto l124
				}
				doarg(yySet, -1)
			l127:
				if !p.rules[ruleBlankLine]() {
					goto l128
				}
				goto l127
			l128:
				do(25)
				goto l123
			l124:
				position, thunkPosition = position124, thunkPosition124
			}
			do(26)
			doarg(yyPop, 2)
			return true
		l122:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 23 ListItem <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (ListContinuationBlock { a = cons(yy, a) })* {
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
					goto l129
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l129
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l129
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l129
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto l129
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l129
			}
			do(27)
		l131:
			{
				position132, thunkPosition132 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l132
				}
				do(28)
				goto l131
			l132:
				position, thunkPosition = position132, thunkPosition132
			}
			do(29)
			doarg(yyPop, 1)
			return true
		l129:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 24 ListItemTight <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (!BlankLine ListContinuationBlock { a = cons(yy, a) })* !ListContinuationBlock {
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
					goto l133
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l133
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l133
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l133
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto l133
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l133
			}
			do(30)
		l135:
			{
				position136, thunkPosition136 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l137
				}
				goto l136
			l137:
				if !p.rules[ruleListContinuationBlock]() {
					goto l136
				}
				do(31)
				goto l135
			l136:
				position, thunkPosition = position136, thunkPosition136
			}
			if !p.rules[ruleListContinuationBlock]() {
				goto l138
			}
			goto l133
		l138:
			do(32)
			doarg(yyPop, 1)
			return true
		l133:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 25 ListBlock <- (StartList !BlankLine Line { a = cons(yy, a) } (ListBlockLine { a = cons(yy, a) })* { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l139
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l140
			}
			goto l139
		l140:
			if !p.rules[ruleLine]() {
				goto l139
			}
			do(33)
		l141:
			{
				position142 := position
				if !p.rules[ruleListBlockLine]() {
					goto l142
				}
				do(34)
				goto l141
			l142:
				position = position142
			}
			do(35)
			doarg(yyPop, 1)
			return true
		l139:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 26 ListContinuationBlock <- (StartList (< BlankLine* > {   if len(yytext) == 0 {
                                   a = cons(mk_str("\001"), a) // block separator
                              } else {
                                   a = cons(mk_str(yytext), a)
                              }
                          }) (Indent ListBlock { a = cons(yy, a) })+ {  yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l143
			}
			doarg(yySet, -1)
			begin = position
		l144:
			if !p.rules[ruleBlankLine]() {
				goto l145
			}
			goto l144
		l145:
			end = position
			do(36)
			if !p.rules[ruleIndent]() {
				goto l143
			}
			if !p.rules[ruleListBlock]() {
				goto l143
			}
			do(37)
		l146:
			{
				position147, thunkPosition147 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l147
				}
				if !p.rules[ruleListBlock]() {
					goto l147
				}
				do(37)
				goto l146
			l147:
				position, thunkPosition = position147, thunkPosition147
			}
			do(38)
			doarg(yyPop, 1)
			return true
		l143:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 27 Enumerator <- (NonindentSpace [0-9]+ '.' Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l148
			}
			if !matchClass(0) {
				goto l148
			}
		l149:
			if !matchClass(0) {
				goto l150
			}
			goto l149
		l150:
			if !matchChar('.') {
				goto l148
			}
			if !p.rules[ruleSpacechar]() {
				goto l148
			}
		l151:
			if !p.rules[ruleSpacechar]() {
				goto l152
			}
			goto l151
		l152:
			return true
		l148:
			position = position0
			return false
		},
		/* 28 OrderedList <- (&Enumerator (ListTight / ListLoose) { yy.key = ORDEREDLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position154 := position
				if !p.rules[ruleEnumerator]() {
					goto l153
				}
				position = position154
			}
			if !p.rules[ruleListTight]() {
				goto l156
			}
			goto l155
		l156:
			if !p.rules[ruleListLoose]() {
				goto l153
			}
		l155:
			do(39)
			return true
		l153:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 29 ListBlockLine <- (!BlankLine !((&[:~] DefMarker) | (&[\t *+\-0-9] (Indent? ((&[*+\-] Bullet) | (&[0-9] Enumerator))))) !HorizontalRule OptionallyIndentedLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlankLine]() {
				goto l158
			}
			goto l157
		l158:
			{
				position159 := position
				{
					if position == len(p.Buffer) {
						goto l159
					}
					switch p.Buffer[position] {
					case ':', '~':
						if !p.rules[ruleDefMarker]() {
							goto l159
						}
						break
					default:
						if !p.rules[ruleIndent]() {
							goto l161
						}
					l161:
						{
							if position == len(p.Buffer) {
								goto l159
							}
							switch p.Buffer[position] {
							case '*', '+', '-':
								if !p.rules[ruleBullet]() {
									goto l159
								}
								break
							default:
								if !p.rules[ruleEnumerator]() {
									goto l159
								}
							}
						}
					}
				}
				goto l157
			l159:
				position = position159
			}
			if !p.rules[ruleHorizontalRule]() {
				goto l164
			}
			goto l157
		l164:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l157
			}
			return true
		l157:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 HtmlBlockOpenAddress <- ('<' Spnl ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l165
			}
			if !p.rules[ruleSpnl]() {
				goto l165
			}
			{
				if position == len(p.Buffer) {
					goto l165
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l165
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l165
					}
					break
				default:
					goto l165
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l165
			}
		l167:
			if !p.rules[ruleHtmlAttribute]() {
				goto l168
			}
			goto l167
		l168:
			if !matchChar('>') {
				goto l165
			}
			return true
		l165:
			position = position0
			return false
		},
		/* 31 HtmlBlockCloseAddress <- ('<' Spnl '/' ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l169
			}
			if !p.rules[ruleSpnl]() {
				goto l169
			}
			if !matchChar('/') {
				goto l169
			}
			{
				if position == len(p.Buffer) {
					goto l169
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l169
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l169
					}
					break
				default:
					goto l169
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l169
			}
			if !matchChar('>') {
				goto l169
			}
			return true
		l169:
			position = position0
			return false
		},
		/* 32 HtmlBlockAddress <- (HtmlBlockOpenAddress (HtmlBlockAddress / (!HtmlBlockCloseAddress .))* HtmlBlockCloseAddress) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenAddress]() {
				goto l171
			}
		l172:
			{
				position173 := position
				if !p.rules[ruleHtmlBlockAddress]() {
					goto l175
				}
				goto l174
			l175:
				if !p.rules[ruleHtmlBlockCloseAddress]() {
					goto l176
				}
				goto l173
			l176:
				if !matchDot() {
					goto l173
				}
			l174:
				goto l172
			l173:
				position = position173
			}
			if !p.rules[ruleHtmlBlockCloseAddress]() {
				goto l171
			}
			return true
		l171:
			position = position0
			return false
		},
		/* 33 HtmlBlockOpenBlockquote <- ('<' Spnl ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l177
			}
			if !p.rules[ruleSpnl]() {
				goto l177
			}
			{
				if position == len(p.Buffer) {
					goto l177
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l177
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l177
					}
					break
				default:
					goto l177
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l177
			}
		l179:
			if !p.rules[ruleHtmlAttribute]() {
				goto l180
			}
			goto l179
		l180:
			if !matchChar('>') {
				goto l177
			}
			return true
		l177:
			position = position0
			return false
		},
		/* 34 HtmlBlockCloseBlockquote <- ('<' Spnl '/' ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l181
			}
			if !p.rules[ruleSpnl]() {
				goto l181
			}
			if !matchChar('/') {
				goto l181
			}
			{
				if position == len(p.Buffer) {
					goto l181
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l181
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l181
					}
					break
				default:
					goto l181
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l181
			}
			if !matchChar('>') {
				goto l181
			}
			return true
		l181:
			position = position0
			return false
		},
		/* 35 HtmlBlockBlockquote <- (HtmlBlockOpenBlockquote (HtmlBlockBlockquote / (!HtmlBlockCloseBlockquote .))* HtmlBlockCloseBlockquote) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenBlockquote]() {
				goto l183
			}
		l184:
			{
				position185 := position
				if !p.rules[ruleHtmlBlockBlockquote]() {
					goto l187
				}
				goto l186
			l187:
				if !p.rules[ruleHtmlBlockCloseBlockquote]() {
					goto l188
				}
				goto l185
			l188:
				if !matchDot() {
					goto l185
				}
			l186:
				goto l184
			l185:
				position = position185
			}
			if !p.rules[ruleHtmlBlockCloseBlockquote]() {
				goto l183
			}
			return true
		l183:
			position = position0
			return false
		},
		/* 36 HtmlBlockOpenCenter <- ('<' Spnl ((&[C] 'CENTER') | (&[c] 'center')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l189
			}
			if !p.rules[ruleSpnl]() {
				goto l189
			}
			{
				if position == len(p.Buffer) {
					goto l189
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l189
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l189
					}
					break
				default:
					goto l189
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l189
			}
		l191:
			if !p.rules[ruleHtmlAttribute]() {
				goto l192
			}
			goto l191
		l192:
			if !matchChar('>') {
				goto l189
			}
			return true
		l189:
			position = position0
			return false
		},
		/* 37 HtmlBlockCloseCenter <- ('<' Spnl '/' ((&[C] 'CENTER') | (&[c] 'center')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l193
			}
			if !p.rules[ruleSpnl]() {
				goto l193
			}
			if !matchChar('/') {
				goto l193
			}
			{
				if position == len(p.Buffer) {
					goto l193
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l193
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l193
					}
					break
				default:
					goto l193
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l193
			}
			if !matchChar('>') {
				goto l193
			}
			return true
		l193:
			position = position0
			return false
		},
		/* 38 HtmlBlockCenter <- (HtmlBlockOpenCenter (HtmlBlockCenter / (!HtmlBlockCloseCenter .))* HtmlBlockCloseCenter) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenCenter]() {
				goto l195
			}
		l196:
			{
				position197 := position
				if !p.rules[ruleHtmlBlockCenter]() {
					goto l199
				}
				goto l198
			l199:
				if !p.rules[ruleHtmlBlockCloseCenter]() {
					goto l200
				}
				goto l197
			l200:
				if !matchDot() {
					goto l197
				}
			l198:
				goto l196
			l197:
				position = position197
			}
			if !p.rules[ruleHtmlBlockCloseCenter]() {
				goto l195
			}
			return true
		l195:
			position = position0
			return false
		},
		/* 39 HtmlBlockOpenDir <- ('<' Spnl ((&[D] 'DIR') | (&[d] 'dir')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l201
			}
			if !p.rules[ruleSpnl]() {
				goto l201
			}
			{
				if position == len(p.Buffer) {
					goto l201
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto l201
					}
					break
				case 'd':
					position++
					if !matchString("ir") {
						goto l201
					}
					break
				default:
					goto l201
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l201
			}
		l203:
			if !p.rules[ruleHtmlAttribute]() {
				goto l204
			}
			goto l203
		l204:
			if !matchChar('>') {
				goto l201
			}
			return true
		l201:
			position = position0
			return false
		},
		/* 40 HtmlBlockCloseDir <- ('<' Spnl '/' ((&[D] 'DIR') | (&[d] 'dir')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l205
			}
			if !p.rules[ruleSpnl]() {
				goto l205
			}
			if !matchChar('/') {
				goto l205
			}
			{
				if position == len(p.Buffer) {
					goto l205
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto l205
					}
					break
				case 'd':
					position++
					if !matchString("ir") {
						goto l205
					}
					break
				default:
					goto l205
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l205
			}
			if !matchChar('>') {
				goto l205
			}
			return true
		l205:
			position = position0
			return false
		},
		/* 41 HtmlBlockDir <- (HtmlBlockOpenDir (HtmlBlockDir / (!HtmlBlockCloseDir .))* HtmlBlockCloseDir) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDir]() {
				goto l207
			}
		l208:
			{
				position209 := position
				if !p.rules[ruleHtmlBlockDir]() {
					goto l211
				}
				goto l210
			l211:
				if !p.rules[ruleHtmlBlockCloseDir]() {
					goto l212
				}
				goto l209
			l212:
				if !matchDot() {
					goto l209
				}
			l210:
				goto l208
			l209:
				position = position209
			}
			if !p.rules[ruleHtmlBlockCloseDir]() {
				goto l207
			}
			return true
		l207:
			position = position0
			return false
		},
		/* 42 HtmlBlockOpenDiv <- ('<' Spnl ((&[D] 'DIV') | (&[d] 'div')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l213
			}
			if !p.rules[ruleSpnl]() {
				goto l213
			}
			{
				if position == len(p.Buffer) {
					goto l213
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto l213
					}
					break
				case 'd':
					position++
					if !matchString("iv") {
						goto l213
					}
					break
				default:
					goto l213
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l213
			}
		l215:
			if !p.rules[ruleHtmlAttribute]() {
				goto l216
			}
			goto l215
		l216:
			if !matchChar('>') {
				goto l213
			}
			return true
		l213:
			position = position0
			return false
		},
		/* 43 HtmlBlockCloseDiv <- ('<' Spnl '/' ((&[D] 'DIV') | (&[d] 'div')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l217
			}
			if !p.rules[ruleSpnl]() {
				goto l217
			}
			if !matchChar('/') {
				goto l217
			}
			{
				if position == len(p.Buffer) {
					goto l217
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto l217
					}
					break
				case 'd':
					position++
					if !matchString("iv") {
						goto l217
					}
					break
				default:
					goto l217
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l217
			}
			if !matchChar('>') {
				goto l217
			}
			return true
		l217:
			position = position0
			return false
		},
		/* 44 HtmlBlockDiv <- (HtmlBlockOpenDiv (HtmlBlockDiv / (!HtmlBlockCloseDiv .))* HtmlBlockCloseDiv) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDiv]() {
				goto l219
			}
		l220:
			{
				position221 := position
				if !p.rules[ruleHtmlBlockDiv]() {
					goto l223
				}
				goto l222
			l223:
				if !p.rules[ruleHtmlBlockCloseDiv]() {
					goto l224
				}
				goto l221
			l224:
				if !matchDot() {
					goto l221
				}
			l222:
				goto l220
			l221:
				position = position221
			}
			if !p.rules[ruleHtmlBlockCloseDiv]() {
				goto l219
			}
			return true
		l219:
			position = position0
			return false
		},
		/* 45 HtmlBlockOpenDl <- ('<' Spnl ((&[D] 'DL') | (&[d] 'dl')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l225
			}
			if !p.rules[ruleSpnl]() {
				goto l225
			}
			{
				if position == len(p.Buffer) {
					goto l225
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto l225
					}
					break
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto l225
					}
					break
				default:
					goto l225
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l225
			}
		l227:
			if !p.rules[ruleHtmlAttribute]() {
				goto l228
			}
			goto l227
		l228:
			if !matchChar('>') {
				goto l225
			}
			return true
		l225:
			position = position0
			return false
		},
		/* 46 HtmlBlockCloseDl <- ('<' Spnl '/' ((&[D] 'DL') | (&[d] 'dl')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l229
			}
			if !p.rules[ruleSpnl]() {
				goto l229
			}
			if !matchChar('/') {
				goto l229
			}
			{
				if position == len(p.Buffer) {
					goto l229
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto l229
					}
					break
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto l229
					}
					break
				default:
					goto l229
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l229
			}
			if !matchChar('>') {
				goto l229
			}
			return true
		l229:
			position = position0
			return false
		},
		/* 47 HtmlBlockDl <- (HtmlBlockOpenDl (HtmlBlockDl / (!HtmlBlockCloseDl .))* HtmlBlockCloseDl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDl]() {
				goto l231
			}
		l232:
			{
				position233 := position
				if !p.rules[ruleHtmlBlockDl]() {
					goto l235
				}
				goto l234
			l235:
				if !p.rules[ruleHtmlBlockCloseDl]() {
					goto l236
				}
				goto l233
			l236:
				if !matchDot() {
					goto l233
				}
			l234:
				goto l232
			l233:
				position = position233
			}
			if !p.rules[ruleHtmlBlockCloseDl]() {
				goto l231
			}
			return true
		l231:
			position = position0
			return false
		},
		/* 48 HtmlBlockOpenFieldset <- ('<' Spnl ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l237
			}
			if !p.rules[ruleSpnl]() {
				goto l237
			}
			{
				if position == len(p.Buffer) {
					goto l237
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto l237
					}
					break
				case 'f':
					position++
					if !matchString("ieldset") {
						goto l237
					}
					break
				default:
					goto l237
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l237
			}
		l239:
			if !p.rules[ruleHtmlAttribute]() {
				goto l240
			}
			goto l239
		l240:
			if !matchChar('>') {
				goto l237
			}
			return true
		l237:
			position = position0
			return false
		},
		/* 49 HtmlBlockCloseFieldset <- ('<' Spnl '/' ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l241
			}
			if !p.rules[ruleSpnl]() {
				goto l241
			}
			if !matchChar('/') {
				goto l241
			}
			{
				if position == len(p.Buffer) {
					goto l241
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto l241
					}
					break
				case 'f':
					position++
					if !matchString("ieldset") {
						goto l241
					}
					break
				default:
					goto l241
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l241
			}
			if !matchChar('>') {
				goto l241
			}
			return true
		l241:
			position = position0
			return false
		},
		/* 50 HtmlBlockFieldset <- (HtmlBlockOpenFieldset (HtmlBlockFieldset / (!HtmlBlockCloseFieldset .))* HtmlBlockCloseFieldset) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFieldset]() {
				goto l243
			}
		l244:
			{
				position245 := position
				if !p.rules[ruleHtmlBlockFieldset]() {
					goto l247
				}
				goto l246
			l247:
				if !p.rules[ruleHtmlBlockCloseFieldset]() {
					goto l248
				}
				goto l245
			l248:
				if !matchDot() {
					goto l245
				}
			l246:
				goto l244
			l245:
				position = position245
			}
			if !p.rules[ruleHtmlBlockCloseFieldset]() {
				goto l243
			}
			return true
		l243:
			position = position0
			return false
		},
		/* 51 HtmlBlockOpenForm <- ('<' Spnl ((&[F] 'FORM') | (&[f] 'form')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l249
			}
			if !p.rules[ruleSpnl]() {
				goto l249
			}
			{
				if position == len(p.Buffer) {
					goto l249
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto l249
					}
					break
				case 'f':
					position++
					if !matchString("orm") {
						goto l249
					}
					break
				default:
					goto l249
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l249
			}
		l251:
			if !p.rules[ruleHtmlAttribute]() {
				goto l252
			}
			goto l251
		l252:
			if !matchChar('>') {
				goto l249
			}
			return true
		l249:
			position = position0
			return false
		},
		/* 52 HtmlBlockCloseForm <- ('<' Spnl '/' ((&[F] 'FORM') | (&[f] 'form')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l253
			}
			if !p.rules[ruleSpnl]() {
				goto l253
			}
			if !matchChar('/') {
				goto l253
			}
			{
				if position == len(p.Buffer) {
					goto l253
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto l253
					}
					break
				case 'f':
					position++
					if !matchString("orm") {
						goto l253
					}
					break
				default:
					goto l253
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l253
			}
			if !matchChar('>') {
				goto l253
			}
			return true
		l253:
			position = position0
			return false
		},
		/* 53 HtmlBlockForm <- (HtmlBlockOpenForm (HtmlBlockForm / (!HtmlBlockCloseForm .))* HtmlBlockCloseForm) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenForm]() {
				goto l255
			}
		l256:
			{
				position257 := position
				if !p.rules[ruleHtmlBlockForm]() {
					goto l259
				}
				goto l258
			l259:
				if !p.rules[ruleHtmlBlockCloseForm]() {
					goto l260
				}
				goto l257
			l260:
				if !matchDot() {
					goto l257
				}
			l258:
				goto l256
			l257:
				position = position257
			}
			if !p.rules[ruleHtmlBlockCloseForm]() {
				goto l255
			}
			return true
		l255:
			position = position0
			return false
		},
		/* 54 HtmlBlockOpenH1 <- ('<' Spnl ((&[H] 'H1') | (&[h] 'h1')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l261
			}
			if !p.rules[ruleSpnl]() {
				goto l261
			}
			{
				if position == len(p.Buffer) {
					goto l261
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto l261
					}
					break
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto l261
					}
					break
				default:
					goto l261
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l261
			}
		l263:
			if !p.rules[ruleHtmlAttribute]() {
				goto l264
			}
			goto l263
		l264:
			if !matchChar('>') {
				goto l261
			}
			return true
		l261:
			position = position0
			return false
		},
		/* 55 HtmlBlockCloseH1 <- ('<' Spnl '/' ((&[H] 'H1') | (&[h] 'h1')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l265
			}
			if !p.rules[ruleSpnl]() {
				goto l265
			}
			if !matchChar('/') {
				goto l265
			}
			{
				if position == len(p.Buffer) {
					goto l265
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto l265
					}
					break
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto l265
					}
					break
				default:
					goto l265
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l265
			}
			if !matchChar('>') {
				goto l265
			}
			return true
		l265:
			position = position0
			return false
		},
		/* 56 HtmlBlockH1 <- (HtmlBlockOpenH1 (HtmlBlockH1 / (!HtmlBlockCloseH1 .))* HtmlBlockCloseH1) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH1]() {
				goto l267
			}
		l268:
			{
				position269 := position
				if !p.rules[ruleHtmlBlockH1]() {
					goto l271
				}
				goto l270
			l271:
				if !p.rules[ruleHtmlBlockCloseH1]() {
					goto l272
				}
				goto l269
			l272:
				if !matchDot() {
					goto l269
				}
			l270:
				goto l268
			l269:
				position = position269
			}
			if !p.rules[ruleHtmlBlockCloseH1]() {
				goto l267
			}
			return true
		l267:
			position = position0
			return false
		},
		/* 57 HtmlBlockOpenH2 <- ('<' Spnl ((&[H] 'H2') | (&[h] 'h2')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l273
			}
			if !p.rules[ruleSpnl]() {
				goto l273
			}
			{
				if position == len(p.Buffer) {
					goto l273
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto l273
					}
					break
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto l273
					}
					break
				default:
					goto l273
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l273
			}
		l275:
			if !p.rules[ruleHtmlAttribute]() {
				goto l276
			}
			goto l275
		l276:
			if !matchChar('>') {
				goto l273
			}
			return true
		l273:
			position = position0
			return false
		},
		/* 58 HtmlBlockCloseH2 <- ('<' Spnl '/' ((&[H] 'H2') | (&[h] 'h2')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l277
			}
			if !p.rules[ruleSpnl]() {
				goto l277
			}
			if !matchChar('/') {
				goto l277
			}
			{
				if position == len(p.Buffer) {
					goto l277
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto l277
					}
					break
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto l277
					}
					break
				default:
					goto l277
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l277
			}
			if !matchChar('>') {
				goto l277
			}
			return true
		l277:
			position = position0
			return false
		},
		/* 59 HtmlBlockH2 <- (HtmlBlockOpenH2 (HtmlBlockH2 / (!HtmlBlockCloseH2 .))* HtmlBlockCloseH2) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH2]() {
				goto l279
			}
		l280:
			{
				position281 := position
				if !p.rules[ruleHtmlBlockH2]() {
					goto l283
				}
				goto l282
			l283:
				if !p.rules[ruleHtmlBlockCloseH2]() {
					goto l284
				}
				goto l281
			l284:
				if !matchDot() {
					goto l281
				}
			l282:
				goto l280
			l281:
				position = position281
			}
			if !p.rules[ruleHtmlBlockCloseH2]() {
				goto l279
			}
			return true
		l279:
			position = position0
			return false
		},
		/* 60 HtmlBlockOpenH3 <- ('<' Spnl ((&[H] 'H3') | (&[h] 'h3')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l285
			}
			if !p.rules[ruleSpnl]() {
				goto l285
			}
			{
				if position == len(p.Buffer) {
					goto l285
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto l285
					}
					break
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto l285
					}
					break
				default:
					goto l285
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l285
			}
		l287:
			if !p.rules[ruleHtmlAttribute]() {
				goto l288
			}
			goto l287
		l288:
			if !matchChar('>') {
				goto l285
			}
			return true
		l285:
			position = position0
			return false
		},
		/* 61 HtmlBlockCloseH3 <- ('<' Spnl '/' ((&[H] 'H3') | (&[h] 'h3')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l289
			}
			if !p.rules[ruleSpnl]() {
				goto l289
			}
			if !matchChar('/') {
				goto l289
			}
			{
				if position == len(p.Buffer) {
					goto l289
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto l289
					}
					break
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto l289
					}
					break
				default:
					goto l289
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l289
			}
			if !matchChar('>') {
				goto l289
			}
			return true
		l289:
			position = position0
			return false
		},
		/* 62 HtmlBlockH3 <- (HtmlBlockOpenH3 (HtmlBlockH3 / (!HtmlBlockCloseH3 .))* HtmlBlockCloseH3) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH3]() {
				goto l291
			}
		l292:
			{
				position293 := position
				if !p.rules[ruleHtmlBlockH3]() {
					goto l295
				}
				goto l294
			l295:
				if !p.rules[ruleHtmlBlockCloseH3]() {
					goto l296
				}
				goto l293
			l296:
				if !matchDot() {
					goto l293
				}
			l294:
				goto l292
			l293:
				position = position293
			}
			if !p.rules[ruleHtmlBlockCloseH3]() {
				goto l291
			}
			return true
		l291:
			position = position0
			return false
		},
		/* 63 HtmlBlockOpenH4 <- ('<' Spnl ((&[H] 'H4') | (&[h] 'h4')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l297
			}
			if !p.rules[ruleSpnl]() {
				goto l297
			}
			{
				if position == len(p.Buffer) {
					goto l297
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto l297
					}
					break
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto l297
					}
					break
				default:
					goto l297
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l297
			}
		l299:
			if !p.rules[ruleHtmlAttribute]() {
				goto l300
			}
			goto l299
		l300:
			if !matchChar('>') {
				goto l297
			}
			return true
		l297:
			position = position0
			return false
		},
		/* 64 HtmlBlockCloseH4 <- ('<' Spnl '/' ((&[H] 'H4') | (&[h] 'h4')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l301
			}
			if !p.rules[ruleSpnl]() {
				goto l301
			}
			if !matchChar('/') {
				goto l301
			}
			{
				if position == len(p.Buffer) {
					goto l301
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto l301
					}
					break
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto l301
					}
					break
				default:
					goto l301
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l301
			}
			if !matchChar('>') {
				goto l301
			}
			return true
		l301:
			position = position0
			return false
		},
		/* 65 HtmlBlockH4 <- (HtmlBlockOpenH4 (HtmlBlockH4 / (!HtmlBlockCloseH4 .))* HtmlBlockCloseH4) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH4]() {
				goto l303
			}
		l304:
			{
				position305 := position
				if !p.rules[ruleHtmlBlockH4]() {
					goto l307
				}
				goto l306
			l307:
				if !p.rules[ruleHtmlBlockCloseH4]() {
					goto l308
				}
				goto l305
			l308:
				if !matchDot() {
					goto l305
				}
			l306:
				goto l304
			l305:
				position = position305
			}
			if !p.rules[ruleHtmlBlockCloseH4]() {
				goto l303
			}
			return true
		l303:
			position = position0
			return false
		},
		/* 66 HtmlBlockOpenH5 <- ('<' Spnl ((&[H] 'H5') | (&[h] 'h5')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l309
			}
			if !p.rules[ruleSpnl]() {
				goto l309
			}
			{
				if position == len(p.Buffer) {
					goto l309
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto l309
					}
					break
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto l309
					}
					break
				default:
					goto l309
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l309
			}
		l311:
			if !p.rules[ruleHtmlAttribute]() {
				goto l312
			}
			goto l311
		l312:
			if !matchChar('>') {
				goto l309
			}
			return true
		l309:
			position = position0
			return false
		},
		/* 67 HtmlBlockCloseH5 <- ('<' Spnl '/' ((&[H] 'H5') | (&[h] 'h5')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l313
			}
			if !p.rules[ruleSpnl]() {
				goto l313
			}
			if !matchChar('/') {
				goto l313
			}
			{
				if position == len(p.Buffer) {
					goto l313
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto l313
					}
					break
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto l313
					}
					break
				default:
					goto l313
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l313
			}
			if !matchChar('>') {
				goto l313
			}
			return true
		l313:
			position = position0
			return false
		},
		/* 68 HtmlBlockH5 <- (HtmlBlockOpenH5 (HtmlBlockH5 / (!HtmlBlockCloseH5 .))* HtmlBlockCloseH5) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH5]() {
				goto l315
			}
		l316:
			{
				position317 := position
				if !p.rules[ruleHtmlBlockH5]() {
					goto l319
				}
				goto l318
			l319:
				if !p.rules[ruleHtmlBlockCloseH5]() {
					goto l320
				}
				goto l317
			l320:
				if !matchDot() {
					goto l317
				}
			l318:
				goto l316
			l317:
				position = position317
			}
			if !p.rules[ruleHtmlBlockCloseH5]() {
				goto l315
			}
			return true
		l315:
			position = position0
			return false
		},
		/* 69 HtmlBlockOpenH6 <- ('<' Spnl ((&[H] 'H6') | (&[h] 'h6')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l321
			}
			if !p.rules[ruleSpnl]() {
				goto l321
			}
			{
				if position == len(p.Buffer) {
					goto l321
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto l321
					}
					break
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto l321
					}
					break
				default:
					goto l321
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l321
			}
		l323:
			if !p.rules[ruleHtmlAttribute]() {
				goto l324
			}
			goto l323
		l324:
			if !matchChar('>') {
				goto l321
			}
			return true
		l321:
			position = position0
			return false
		},
		/* 70 HtmlBlockCloseH6 <- ('<' Spnl '/' ((&[H] 'H6') | (&[h] 'h6')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l325
			}
			if !p.rules[ruleSpnl]() {
				goto l325
			}
			if !matchChar('/') {
				goto l325
			}
			{
				if position == len(p.Buffer) {
					goto l325
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto l325
					}
					break
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto l325
					}
					break
				default:
					goto l325
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l325
			}
			if !matchChar('>') {
				goto l325
			}
			return true
		l325:
			position = position0
			return false
		},
		/* 71 HtmlBlockH6 <- (HtmlBlockOpenH6 (HtmlBlockH6 / (!HtmlBlockCloseH6 .))* HtmlBlockCloseH6) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH6]() {
				goto l327
			}
		l328:
			{
				position329 := position
				if !p.rules[ruleHtmlBlockH6]() {
					goto l331
				}
				goto l330
			l331:
				if !p.rules[ruleHtmlBlockCloseH6]() {
					goto l332
				}
				goto l329
			l332:
				if !matchDot() {
					goto l329
				}
			l330:
				goto l328
			l329:
				position = position329
			}
			if !p.rules[ruleHtmlBlockCloseH6]() {
				goto l327
			}
			return true
		l327:
			position = position0
			return false
		},
		/* 72 HtmlBlockOpenMenu <- ('<' Spnl ((&[M] 'MENU') | (&[m] 'menu')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l333
			}
			if !p.rules[ruleSpnl]() {
				goto l333
			}
			{
				if position == len(p.Buffer) {
					goto l333
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto l333
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l333
					}
					break
				default:
					goto l333
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l333
			}
		l335:
			if !p.rules[ruleHtmlAttribute]() {
				goto l336
			}
			goto l335
		l336:
			if !matchChar('>') {
				goto l333
			}
			return true
		l333:
			position = position0
			return false
		},
		/* 73 HtmlBlockCloseMenu <- ('<' Spnl '/' ((&[M] 'MENU') | (&[m] 'menu')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l337
			}
			if !p.rules[ruleSpnl]() {
				goto l337
			}
			if !matchChar('/') {
				goto l337
			}
			{
				if position == len(p.Buffer) {
					goto l337
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto l337
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l337
					}
					break
				default:
					goto l337
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l337
			}
			if !matchChar('>') {
				goto l337
			}
			return true
		l337:
			position = position0
			return false
		},
		/* 74 HtmlBlockMenu <- (HtmlBlockOpenMenu (HtmlBlockMenu / (!HtmlBlockCloseMenu .))* HtmlBlockCloseMenu) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenMenu]() {
				goto l339
			}
		l340:
			{
				position341 := position
				if !p.rules[ruleHtmlBlockMenu]() {
					goto l343
				}
				goto l342
			l343:
				if !p.rules[ruleHtmlBlockCloseMenu]() {
					goto l344
				}
				goto l341
			l344:
				if !matchDot() {
					goto l341
				}
			l342:
				goto l340
			l341:
				position = position341
			}
			if !p.rules[ruleHtmlBlockCloseMenu]() {
				goto l339
			}
			return true
		l339:
			position = position0
			return false
		},
		/* 75 HtmlBlockOpenNoframes <- ('<' Spnl ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l345
			}
			if !p.rules[ruleSpnl]() {
				goto l345
			}
			{
				if position == len(p.Buffer) {
					goto l345
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto l345
					}
					break
				case 'n':
					position++
					if !matchString("oframes") {
						goto l345
					}
					break
				default:
					goto l345
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l345
			}
		l347:
			if !p.rules[ruleHtmlAttribute]() {
				goto l348
			}
			goto l347
		l348:
			if !matchChar('>') {
				goto l345
			}
			return true
		l345:
			position = position0
			return false
		},
		/* 76 HtmlBlockCloseNoframes <- ('<' Spnl '/' ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l349
			}
			if !p.rules[ruleSpnl]() {
				goto l349
			}
			if !matchChar('/') {
				goto l349
			}
			{
				if position == len(p.Buffer) {
					goto l349
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto l349
					}
					break
				case 'n':
					position++
					if !matchString("oframes") {
						goto l349
					}
					break
				default:
					goto l349
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l349
			}
			if !matchChar('>') {
				goto l349
			}
			return true
		l349:
			position = position0
			return false
		},
		/* 77 HtmlBlockNoframes <- (HtmlBlockOpenNoframes (HtmlBlockNoframes / (!HtmlBlockCloseNoframes .))* HtmlBlockCloseNoframes) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoframes]() {
				goto l351
			}
		l352:
			{
				position353 := position
				if !p.rules[ruleHtmlBlockNoframes]() {
					goto l355
				}
				goto l354
			l355:
				if !p.rules[ruleHtmlBlockCloseNoframes]() {
					goto l356
				}
				goto l353
			l356:
				if !matchDot() {
					goto l353
				}
			l354:
				goto l352
			l353:
				position = position353
			}
			if !p.rules[ruleHtmlBlockCloseNoframes]() {
				goto l351
			}
			return true
		l351:
			position = position0
			return false
		},
		/* 78 HtmlBlockOpenNoscript <- ('<' Spnl ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l357
			}
			if !p.rules[ruleSpnl]() {
				goto l357
			}
			{
				if position == len(p.Buffer) {
					goto l357
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l357
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l357
					}
					break
				default:
					goto l357
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l357
			}
		l359:
			if !p.rules[ruleHtmlAttribute]() {
				goto l360
			}
			goto l359
		l360:
			if !matchChar('>') {
				goto l357
			}
			return true
		l357:
			position = position0
			return false
		},
		/* 79 HtmlBlockCloseNoscript <- ('<' Spnl '/' ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l361
			}
			if !p.rules[ruleSpnl]() {
				goto l361
			}
			if !matchChar('/') {
				goto l361
			}
			{
				if position == len(p.Buffer) {
					goto l361
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l361
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l361
					}
					break
				default:
					goto l361
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l361
			}
			if !matchChar('>') {
				goto l361
			}
			return true
		l361:
			position = position0
			return false
		},
		/* 80 HtmlBlockNoscript <- (HtmlBlockOpenNoscript (HtmlBlockNoscript / (!HtmlBlockCloseNoscript .))* HtmlBlockCloseNoscript) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoscript]() {
				goto l363
			}
		l364:
			{
				position365 := position
				if !p.rules[ruleHtmlBlockNoscript]() {
					goto l367
				}
				goto l366
			l367:
				if !p.rules[ruleHtmlBlockCloseNoscript]() {
					goto l368
				}
				goto l365
			l368:
				if !matchDot() {
					goto l365
				}
			l366:
				goto l364
			l365:
				position = position365
			}
			if !p.rules[ruleHtmlBlockCloseNoscript]() {
				goto l363
			}
			return true
		l363:
			position = position0
			return false
		},
		/* 81 HtmlBlockOpenOl <- ('<' Spnl ((&[O] 'OL') | (&[o] 'ol')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l369
			}
			if !p.rules[ruleSpnl]() {
				goto l369
			}
			{
				if position == len(p.Buffer) {
					goto l369
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l369
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l369
					}
					break
				default:
					goto l369
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l369
			}
		l371:
			if !p.rules[ruleHtmlAttribute]() {
				goto l372
			}
			goto l371
		l372:
			if !matchChar('>') {
				goto l369
			}
			return true
		l369:
			position = position0
			return false
		},
		/* 82 HtmlBlockCloseOl <- ('<' Spnl '/' ((&[O] 'OL') | (&[o] 'ol')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l373
			}
			if !p.rules[ruleSpnl]() {
				goto l373
			}
			if !matchChar('/') {
				goto l373
			}
			{
				if position == len(p.Buffer) {
					goto l373
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l373
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l373
					}
					break
				default:
					goto l373
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l373
			}
			if !matchChar('>') {
				goto l373
			}
			return true
		l373:
			position = position0
			return false
		},
		/* 83 HtmlBlockOl <- (HtmlBlockOpenOl (HtmlBlockOl / (!HtmlBlockCloseOl .))* HtmlBlockCloseOl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenOl]() {
				goto l375
			}
		l376:
			{
				position377 := position
				if !p.rules[ruleHtmlBlockOl]() {
					goto l379
				}
				goto l378
			l379:
				if !p.rules[ruleHtmlBlockCloseOl]() {
					goto l380
				}
				goto l377
			l380:
				if !matchDot() {
					goto l377
				}
			l378:
				goto l376
			l377:
				position = position377
			}
			if !p.rules[ruleHtmlBlockCloseOl]() {
				goto l375
			}
			return true
		l375:
			position = position0
			return false
		},
		/* 84 HtmlBlockOpenP <- ('<' Spnl ((&[P] 'P') | (&[p] 'p')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l381
			}
			if !p.rules[ruleSpnl]() {
				goto l381
			}
			{
				if position == len(p.Buffer) {
					goto l381
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
					break
				case 'p':
					position++ // matchChar
					break
				default:
					goto l381
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l381
			}
		l383:
			if !p.rules[ruleHtmlAttribute]() {
				goto l384
			}
			goto l383
		l384:
			if !matchChar('>') {
				goto l381
			}
			return true
		l381:
			position = position0
			return false
		},
		/* 85 HtmlBlockCloseP <- ('<' Spnl '/' ((&[P] 'P') | (&[p] 'p')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l385
			}
			if !p.rules[ruleSpnl]() {
				goto l385
			}
			if !matchChar('/') {
				goto l385
			}
			{
				if position == len(p.Buffer) {
					goto l385
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
					break
				case 'p':
					position++ // matchChar
					break
				default:
					goto l385
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l385
			}
			if !matchChar('>') {
				goto l385
			}
			return true
		l385:
			position = position0
			return false
		},
		/* 86 HtmlBlockP <- (HtmlBlockOpenP (HtmlBlockP / (!HtmlBlockCloseP .))* HtmlBlockCloseP) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenP]() {
				goto l387
			}
		l388:
			{
				position389 := position
				if !p.rules[ruleHtmlBlockP]() {
					goto l391
				}
				goto l390
			l391:
				if !p.rules[ruleHtmlBlockCloseP]() {
					goto l392
				}
				goto l389
			l392:
				if !matchDot() {
					goto l389
				}
			l390:
				goto l388
			l389:
				position = position389
			}
			if !p.rules[ruleHtmlBlockCloseP]() {
				goto l387
			}
			return true
		l387:
			position = position0
			return false
		},
		/* 87 HtmlBlockOpenPre <- ('<' Spnl ((&[P] 'PRE') | (&[p] 'pre')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l393
			}
			if !p.rules[ruleSpnl]() {
				goto l393
			}
			{
				if position == len(p.Buffer) {
					goto l393
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto l393
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l393
					}
					break
				default:
					goto l393
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l393
			}
		l395:
			if !p.rules[ruleHtmlAttribute]() {
				goto l396
			}
			goto l395
		l396:
			if !matchChar('>') {
				goto l393
			}
			return true
		l393:
			position = position0
			return false
		},
		/* 88 HtmlBlockClosePre <- ('<' Spnl '/' ((&[P] 'PRE') | (&[p] 'pre')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l397
			}
			if !p.rules[ruleSpnl]() {
				goto l397
			}
			if !matchChar('/') {
				goto l397
			}
			{
				if position == len(p.Buffer) {
					goto l397
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto l397
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l397
					}
					break
				default:
					goto l397
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l397
			}
			if !matchChar('>') {
				goto l397
			}
			return true
		l397:
			position = position0
			return false
		},
		/* 89 HtmlBlockPre <- (HtmlBlockOpenPre (HtmlBlockPre / (!HtmlBlockClosePre .))* HtmlBlockClosePre) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenPre]() {
				goto l399
			}
		l400:
			{
				position401 := position
				if !p.rules[ruleHtmlBlockPre]() {
					goto l403
				}
				goto l402
			l403:
				if !p.rules[ruleHtmlBlockClosePre]() {
					goto l404
				}
				goto l401
			l404:
				if !matchDot() {
					goto l401
				}
			l402:
				goto l400
			l401:
				position = position401
			}
			if !p.rules[ruleHtmlBlockClosePre]() {
				goto l399
			}
			return true
		l399:
			position = position0
			return false
		},
		/* 90 HtmlBlockOpenTable <- ('<' Spnl ((&[T] 'TABLE') | (&[t] 'table')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l405
			}
			if !p.rules[ruleSpnl]() {
				goto l405
			}
			{
				if position == len(p.Buffer) {
					goto l405
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto l405
					}
					break
				case 't':
					position++
					if !matchString("able") {
						goto l405
					}
					break
				default:
					goto l405
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l405
			}
		l407:
			if !p.rules[ruleHtmlAttribute]() {
				goto l408
			}
			goto l407
		l408:
			if !matchChar('>') {
				goto l405
			}
			return true
		l405:
			position = position0
			return false
		},
		/* 91 HtmlBlockCloseTable <- ('<' Spnl '/' ((&[T] 'TABLE') | (&[t] 'table')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l409
			}
			if !p.rules[ruleSpnl]() {
				goto l409
			}
			if !matchChar('/') {
				goto l409
			}
			{
				if position == len(p.Buffer) {
					goto l409
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto l409
					}
					break
				case 't':
					position++
					if !matchString("able") {
						goto l409
					}
					break
				default:
					goto l409
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l409
			}
			if !matchChar('>') {
				goto l409
			}
			return true
		l409:
			position = position0
			return false
		},
		/* 92 HtmlBlockTable <- (HtmlBlockOpenTable (HtmlBlockTable / (!HtmlBlockCloseTable .))* HtmlBlockCloseTable) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTable]() {
				goto l411
			}
		l412:
			{
				position413 := position
				if !p.rules[ruleHtmlBlockTable]() {
					goto l415
				}
				goto l414
			l415:
				if !p.rules[ruleHtmlBlockCloseTable]() {
					goto l416
				}
				goto l413
			l416:
				if !matchDot() {
					goto l413
				}
			l414:
				goto l412
			l413:
				position = position413
			}
			if !p.rules[ruleHtmlBlockCloseTable]() {
				goto l411
			}
			return true
		l411:
			position = position0
			return false
		},
		/* 93 HtmlBlockOpenUl <- ('<' Spnl ((&[U] 'UL') | (&[u] 'ul')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l417
			}
			if !p.rules[ruleSpnl]() {
				goto l417
			}
			{
				if position == len(p.Buffer) {
					goto l417
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l417
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l417
					}
					break
				default:
					goto l417
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l417
			}
		l419:
			if !p.rules[ruleHtmlAttribute]() {
				goto l420
			}
			goto l419
		l420:
			if !matchChar('>') {
				goto l417
			}
			return true
		l417:
			position = position0
			return false
		},
		/* 94 HtmlBlockCloseUl <- ('<' Spnl '/' ((&[U] 'UL') | (&[u] 'ul')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l421
			}
			if !p.rules[ruleSpnl]() {
				goto l421
			}
			if !matchChar('/') {
				goto l421
			}
			{
				if position == len(p.Buffer) {
					goto l421
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l421
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l421
					}
					break
				default:
					goto l421
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l421
			}
			if !matchChar('>') {
				goto l421
			}
			return true
		l421:
			position = position0
			return false
		},
		/* 95 HtmlBlockUl <- (HtmlBlockOpenUl (HtmlBlockUl / (!HtmlBlockCloseUl .))* HtmlBlockCloseUl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenUl]() {
				goto l423
			}
		l424:
			{
				position425 := position
				if !p.rules[ruleHtmlBlockUl]() {
					goto l427
				}
				goto l426
			l427:
				if !p.rules[ruleHtmlBlockCloseUl]() {
					goto l428
				}
				goto l425
			l428:
				if !matchDot() {
					goto l425
				}
			l426:
				goto l424
			l425:
				position = position425
			}
			if !p.rules[ruleHtmlBlockCloseUl]() {
				goto l423
			}
			return true
		l423:
			position = position0
			return false
		},
		/* 96 HtmlBlockOpenDd <- ('<' Spnl ((&[D] 'DD') | (&[d] 'dd')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l429
			}
			if !p.rules[ruleSpnl]() {
				goto l429
			}
			{
				if position == len(p.Buffer) {
					goto l429
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto l429
					}
					break
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto l429
					}
					break
				default:
					goto l429
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l429
			}
		l431:
			if !p.rules[ruleHtmlAttribute]() {
				goto l432
			}
			goto l431
		l432:
			if !matchChar('>') {
				goto l429
			}
			return true
		l429:
			position = position0
			return false
		},
		/* 97 HtmlBlockCloseDd <- ('<' Spnl '/' ((&[D] 'DD') | (&[d] 'dd')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l433
			}
			if !p.rules[ruleSpnl]() {
				goto l433
			}
			if !matchChar('/') {
				goto l433
			}
			{
				if position == len(p.Buffer) {
					goto l433
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto l433
					}
					break
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto l433
					}
					break
				default:
					goto l433
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l433
			}
			if !matchChar('>') {
				goto l433
			}
			return true
		l433:
			position = position0
			return false
		},
		/* 98 HtmlBlockDd <- (HtmlBlockOpenDd (HtmlBlockDd / (!HtmlBlockCloseDd .))* HtmlBlockCloseDd) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDd]() {
				goto l435
			}
		l436:
			{
				position437 := position
				if !p.rules[ruleHtmlBlockDd]() {
					goto l439
				}
				goto l438
			l439:
				if !p.rules[ruleHtmlBlockCloseDd]() {
					goto l440
				}
				goto l437
			l440:
				if !matchDot() {
					goto l437
				}
			l438:
				goto l436
			l437:
				position = position437
			}
			if !p.rules[ruleHtmlBlockCloseDd]() {
				goto l435
			}
			return true
		l435:
			position = position0
			return false
		},
		/* 99 HtmlBlockOpenDt <- ('<' Spnl ((&[D] 'DT') | (&[d] 'dt')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l441
			}
			if !p.rules[ruleSpnl]() {
				goto l441
			}
			{
				if position == len(p.Buffer) {
					goto l441
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l441
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l441
					}
					break
				default:
					goto l441
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l441
			}
		l443:
			if !p.rules[ruleHtmlAttribute]() {
				goto l444
			}
			goto l443
		l444:
			if !matchChar('>') {
				goto l441
			}
			return true
		l441:
			position = position0
			return false
		},
		/* 100 HtmlBlockCloseDt <- ('<' Spnl '/' ((&[D] 'DT') | (&[d] 'dt')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l445
			}
			if !p.rules[ruleSpnl]() {
				goto l445
			}
			if !matchChar('/') {
				goto l445
			}
			{
				if position == len(p.Buffer) {
					goto l445
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l445
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l445
					}
					break
				default:
					goto l445
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l445
			}
			if !matchChar('>') {
				goto l445
			}
			return true
		l445:
			position = position0
			return false
		},
		/* 101 HtmlBlockDt <- (HtmlBlockOpenDt (HtmlBlockDt / (!HtmlBlockCloseDt .))* HtmlBlockCloseDt) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDt]() {
				goto l447
			}
		l448:
			{
				position449 := position
				if !p.rules[ruleHtmlBlockDt]() {
					goto l451
				}
				goto l450
			l451:
				if !p.rules[ruleHtmlBlockCloseDt]() {
					goto l452
				}
				goto l449
			l452:
				if !matchDot() {
					goto l449
				}
			l450:
				goto l448
			l449:
				position = position449
			}
			if !p.rules[ruleHtmlBlockCloseDt]() {
				goto l447
			}
			return true
		l447:
			position = position0
			return false
		},
		/* 102 HtmlBlockOpenFrameset <- ('<' Spnl ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l453
			}
			if !p.rules[ruleSpnl]() {
				goto l453
			}
			{
				if position == len(p.Buffer) {
					goto l453
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l453
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l453
					}
					break
				default:
					goto l453
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l453
			}
		l455:
			if !p.rules[ruleHtmlAttribute]() {
				goto l456
			}
			goto l455
		l456:
			if !matchChar('>') {
				goto l453
			}
			return true
		l453:
			position = position0
			return false
		},
		/* 103 HtmlBlockCloseFrameset <- ('<' Spnl '/' ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l457
			}
			if !p.rules[ruleSpnl]() {
				goto l457
			}
			if !matchChar('/') {
				goto l457
			}
			{
				if position == len(p.Buffer) {
					goto l457
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l457
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l457
					}
					break
				default:
					goto l457
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l457
			}
			if !matchChar('>') {
				goto l457
			}
			return true
		l457:
			position = position0
			return false
		},
		/* 104 HtmlBlockFrameset <- (HtmlBlockOpenFrameset (HtmlBlockFrameset / (!HtmlBlockCloseFrameset .))* HtmlBlockCloseFrameset) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFrameset]() {
				goto l459
			}
		l460:
			{
				position461 := position
				if !p.rules[ruleHtmlBlockFrameset]() {
					goto l463
				}
				goto l462
			l463:
				if !p.rules[ruleHtmlBlockCloseFrameset]() {
					goto l464
				}
				goto l461
			l464:
				if !matchDot() {
					goto l461
				}
			l462:
				goto l460
			l461:
				position = position461
			}
			if !p.rules[ruleHtmlBlockCloseFrameset]() {
				goto l459
			}
			return true
		l459:
			position = position0
			return false
		},
		/* 105 HtmlBlockOpenLi <- ('<' Spnl ((&[L] 'LI') | (&[l] 'li')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l465
			}
			if !p.rules[ruleSpnl]() {
				goto l465
			}
			{
				if position == len(p.Buffer) {
					goto l465
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l465
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l465
					}
					break
				default:
					goto l465
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l465
			}
		l467:
			if !p.rules[ruleHtmlAttribute]() {
				goto l468
			}
			goto l467
		l468:
			if !matchChar('>') {
				goto l465
			}
			return true
		l465:
			position = position0
			return false
		},
		/* 106 HtmlBlockCloseLi <- ('<' Spnl '/' ((&[L] 'LI') | (&[l] 'li')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l469
			}
			if !p.rules[ruleSpnl]() {
				goto l469
			}
			if !matchChar('/') {
				goto l469
			}
			{
				if position == len(p.Buffer) {
					goto l469
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l469
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l469
					}
					break
				default:
					goto l469
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l469
			}
			if !matchChar('>') {
				goto l469
			}
			return true
		l469:
			position = position0
			return false
		},
		/* 107 HtmlBlockLi <- (HtmlBlockOpenLi (HtmlBlockLi / (!HtmlBlockCloseLi .))* HtmlBlockCloseLi) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenLi]() {
				goto l471
			}
		l472:
			{
				position473 := position
				if !p.rules[ruleHtmlBlockLi]() {
					goto l475
				}
				goto l474
			l475:
				if !p.rules[ruleHtmlBlockCloseLi]() {
					goto l476
				}
				goto l473
			l476:
				if !matchDot() {
					goto l473
				}
			l474:
				goto l472
			l473:
				position = position473
			}
			if !p.rules[ruleHtmlBlockCloseLi]() {
				goto l471
			}
			return true
		l471:
			position = position0
			return false
		},
		/* 108 HtmlBlockOpenTbody <- ('<' Spnl ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l477
			}
			if !p.rules[ruleSpnl]() {
				goto l477
			}
			{
				if position == len(p.Buffer) {
					goto l477
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto l477
					}
					break
				case 't':
					position++
					if !matchString("body") {
						goto l477
					}
					break
				default:
					goto l477
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l477
			}
		l479:
			if !p.rules[ruleHtmlAttribute]() {
				goto l480
			}
			goto l479
		l480:
			if !matchChar('>') {
				goto l477
			}
			return true
		l477:
			position = position0
			return false
		},
		/* 109 HtmlBlockCloseTbody <- ('<' Spnl '/' ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l481
			}
			if !p.rules[ruleSpnl]() {
				goto l481
			}
			if !matchChar('/') {
				goto l481
			}
			{
				if position == len(p.Buffer) {
					goto l481
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto l481
					}
					break
				case 't':
					position++
					if !matchString("body") {
						goto l481
					}
					break
				default:
					goto l481
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l481
			}
			if !matchChar('>') {
				goto l481
			}
			return true
		l481:
			position = position0
			return false
		},
		/* 110 HtmlBlockTbody <- (HtmlBlockOpenTbody (HtmlBlockTbody / (!HtmlBlockCloseTbody .))* HtmlBlockCloseTbody) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTbody]() {
				goto l483
			}
		l484:
			{
				position485 := position
				if !p.rules[ruleHtmlBlockTbody]() {
					goto l487
				}
				goto l486
			l487:
				if !p.rules[ruleHtmlBlockCloseTbody]() {
					goto l488
				}
				goto l485
			l488:
				if !matchDot() {
					goto l485
				}
			l486:
				goto l484
			l485:
				position = position485
			}
			if !p.rules[ruleHtmlBlockCloseTbody]() {
				goto l483
			}
			return true
		l483:
			position = position0
			return false
		},
		/* 111 HtmlBlockOpenTd <- ('<' Spnl ((&[T] 'TD') | (&[t] 'td')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l489
			}
			if !p.rules[ruleSpnl]() {
				goto l489
			}
			{
				if position == len(p.Buffer) {
					goto l489
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto l489
					}
					break
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto l489
					}
					break
				default:
					goto l489
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l489
			}
		l491:
			if !p.rules[ruleHtmlAttribute]() {
				goto l492
			}
			goto l491
		l492:
			if !matchChar('>') {
				goto l489
			}
			return true
		l489:
			position = position0
			return false
		},
		/* 112 HtmlBlockCloseTd <- ('<' Spnl '/' ((&[T] 'TD') | (&[t] 'td')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l493
			}
			if !p.rules[ruleSpnl]() {
				goto l493
			}
			if !matchChar('/') {
				goto l493
			}
			{
				if position == len(p.Buffer) {
					goto l493
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto l493
					}
					break
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto l493
					}
					break
				default:
					goto l493
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l493
			}
			if !matchChar('>') {
				goto l493
			}
			return true
		l493:
			position = position0
			return false
		},
		/* 113 HtmlBlockTd <- (HtmlBlockOpenTd (HtmlBlockTd / (!HtmlBlockCloseTd .))* HtmlBlockCloseTd) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTd]() {
				goto l495
			}
		l496:
			{
				position497 := position
				if !p.rules[ruleHtmlBlockTd]() {
					goto l499
				}
				goto l498
			l499:
				if !p.rules[ruleHtmlBlockCloseTd]() {
					goto l500
				}
				goto l497
			l500:
				if !matchDot() {
					goto l497
				}
			l498:
				goto l496
			l497:
				position = position497
			}
			if !p.rules[ruleHtmlBlockCloseTd]() {
				goto l495
			}
			return true
		l495:
			position = position0
			return false
		},
		/* 114 HtmlBlockOpenTfoot <- ('<' Spnl ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l501
			}
			if !p.rules[ruleSpnl]() {
				goto l501
			}
			{
				if position == len(p.Buffer) {
					goto l501
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto l501
					}
					break
				case 't':
					position++
					if !matchString("foot") {
						goto l501
					}
					break
				default:
					goto l501
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l501
			}
		l503:
			if !p.rules[ruleHtmlAttribute]() {
				goto l504
			}
			goto l503
		l504:
			if !matchChar('>') {
				goto l501
			}
			return true
		l501:
			position = position0
			return false
		},
		/* 115 HtmlBlockCloseTfoot <- ('<' Spnl '/' ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l505
			}
			if !p.rules[ruleSpnl]() {
				goto l505
			}
			if !matchChar('/') {
				goto l505
			}
			{
				if position == len(p.Buffer) {
					goto l505
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto l505
					}
					break
				case 't':
					position++
					if !matchString("foot") {
						goto l505
					}
					break
				default:
					goto l505
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l505
			}
			if !matchChar('>') {
				goto l505
			}
			return true
		l505:
			position = position0
			return false
		},
		/* 116 HtmlBlockTfoot <- (HtmlBlockOpenTfoot (HtmlBlockTfoot / (!HtmlBlockCloseTfoot .))* HtmlBlockCloseTfoot) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTfoot]() {
				goto l507
			}
		l508:
			{
				position509 := position
				if !p.rules[ruleHtmlBlockTfoot]() {
					goto l511
				}
				goto l510
			l511:
				if !p.rules[ruleHtmlBlockCloseTfoot]() {
					goto l512
				}
				goto l509
			l512:
				if !matchDot() {
					goto l509
				}
			l510:
				goto l508
			l509:
				position = position509
			}
			if !p.rules[ruleHtmlBlockCloseTfoot]() {
				goto l507
			}
			return true
		l507:
			position = position0
			return false
		},
		/* 117 HtmlBlockOpenTh <- ('<' Spnl ((&[T] 'TH') | (&[t] 'th')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l513
			}
			if !p.rules[ruleSpnl]() {
				goto l513
			}
			{
				if position == len(p.Buffer) {
					goto l513
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto l513
					}
					break
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto l513
					}
					break
				default:
					goto l513
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l513
			}
		l515:
			if !p.rules[ruleHtmlAttribute]() {
				goto l516
			}
			goto l515
		l516:
			if !matchChar('>') {
				goto l513
			}
			return true
		l513:
			position = position0
			return false
		},
		/* 118 HtmlBlockCloseTh <- ('<' Spnl '/' ((&[T] 'TH') | (&[t] 'th')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l517
			}
			if !p.rules[ruleSpnl]() {
				goto l517
			}
			if !matchChar('/') {
				goto l517
			}
			{
				if position == len(p.Buffer) {
					goto l517
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto l517
					}
					break
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto l517
					}
					break
				default:
					goto l517
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l517
			}
			if !matchChar('>') {
				goto l517
			}
			return true
		l517:
			position = position0
			return false
		},
		/* 119 HtmlBlockTh <- (HtmlBlockOpenTh (HtmlBlockTh / (!HtmlBlockCloseTh .))* HtmlBlockCloseTh) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTh]() {
				goto l519
			}
		l520:
			{
				position521 := position
				if !p.rules[ruleHtmlBlockTh]() {
					goto l523
				}
				goto l522
			l523:
				if !p.rules[ruleHtmlBlockCloseTh]() {
					goto l524
				}
				goto l521
			l524:
				if !matchDot() {
					goto l521
				}
			l522:
				goto l520
			l521:
				position = position521
			}
			if !p.rules[ruleHtmlBlockCloseTh]() {
				goto l519
			}
			return true
		l519:
			position = position0
			return false
		},
		/* 120 HtmlBlockOpenThead <- ('<' Spnl ((&[T] 'THEAD') | (&[t] 'thead')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l525
			}
			if !p.rules[ruleSpnl]() {
				goto l525
			}
			{
				if position == len(p.Buffer) {
					goto l525
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto l525
					}
					break
				case 't':
					position++
					if !matchString("head") {
						goto l525
					}
					break
				default:
					goto l525
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l525
			}
		l527:
			if !p.rules[ruleHtmlAttribute]() {
				goto l528
			}
			goto l527
		l528:
			if !matchChar('>') {
				goto l525
			}
			return true
		l525:
			position = position0
			return false
		},
		/* 121 HtmlBlockCloseThead <- ('<' Spnl '/' ((&[T] 'THEAD') | (&[t] 'thead')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l529
			}
			if !p.rules[ruleSpnl]() {
				goto l529
			}
			if !matchChar('/') {
				goto l529
			}
			{
				if position == len(p.Buffer) {
					goto l529
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto l529
					}
					break
				case 't':
					position++
					if !matchString("head") {
						goto l529
					}
					break
				default:
					goto l529
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l529
			}
			if !matchChar('>') {
				goto l529
			}
			return true
		l529:
			position = position0
			return false
		},
		/* 122 HtmlBlockThead <- (HtmlBlockOpenThead (HtmlBlockThead / (!HtmlBlockCloseThead .))* HtmlBlockCloseThead) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenThead]() {
				goto l531
			}
		l532:
			{
				position533 := position
				if !p.rules[ruleHtmlBlockThead]() {
					goto l535
				}
				goto l534
			l535:
				if !p.rules[ruleHtmlBlockCloseThead]() {
					goto l536
				}
				goto l533
			l536:
				if !matchDot() {
					goto l533
				}
			l534:
				goto l532
			l533:
				position = position533
			}
			if !p.rules[ruleHtmlBlockCloseThead]() {
				goto l531
			}
			return true
		l531:
			position = position0
			return false
		},
		/* 123 HtmlBlockOpenTr <- ('<' Spnl ((&[T] 'TR') | (&[t] 'tr')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l537
			}
			if !p.rules[ruleSpnl]() {
				goto l537
			}
			{
				if position == len(p.Buffer) {
					goto l537
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l537
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l537
					}
					break
				default:
					goto l537
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l537
			}
		l539:
			if !p.rules[ruleHtmlAttribute]() {
				goto l540
			}
			goto l539
		l540:
			if !matchChar('>') {
				goto l537
			}
			return true
		l537:
			position = position0
			return false
		},
		/* 124 HtmlBlockCloseTr <- ('<' Spnl '/' ((&[T] 'TR') | (&[t] 'tr')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l541
			}
			if !p.rules[ruleSpnl]() {
				goto l541
			}
			if !matchChar('/') {
				goto l541
			}
			{
				if position == len(p.Buffer) {
					goto l541
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l541
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l541
					}
					break
				default:
					goto l541
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l541
			}
			if !matchChar('>') {
				goto l541
			}
			return true
		l541:
			position = position0
			return false
		},
		/* 125 HtmlBlockTr <- (HtmlBlockOpenTr (HtmlBlockTr / (!HtmlBlockCloseTr .))* HtmlBlockCloseTr) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTr]() {
				goto l543
			}
		l544:
			{
				position545 := position
				if !p.rules[ruleHtmlBlockTr]() {
					goto l547
				}
				goto l546
			l547:
				if !p.rules[ruleHtmlBlockCloseTr]() {
					goto l548
				}
				goto l545
			l548:
				if !matchDot() {
					goto l545
				}
			l546:
				goto l544
			l545:
				position = position545
			}
			if !p.rules[ruleHtmlBlockCloseTr]() {
				goto l543
			}
			return true
		l543:
			position = position0
			return false
		},
		/* 126 HtmlBlockOpenScript <- ('<' Spnl ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l549
			}
			if !p.rules[ruleSpnl]() {
				goto l549
			}
			{
				if position == len(p.Buffer) {
					goto l549
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l549
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l549
					}
					break
				default:
					goto l549
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l549
			}
		l551:
			if !p.rules[ruleHtmlAttribute]() {
				goto l552
			}
			goto l551
		l552:
			if !matchChar('>') {
				goto l549
			}
			return true
		l549:
			position = position0
			return false
		},
		/* 127 HtmlBlockCloseScript <- ('<' Spnl '/' ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l553
			}
			if !p.rules[ruleSpnl]() {
				goto l553
			}
			if !matchChar('/') {
				goto l553
			}
			{
				if position == len(p.Buffer) {
					goto l553
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l553
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l553
					}
					break
				default:
					goto l553
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l553
			}
			if !matchChar('>') {
				goto l553
			}
			return true
		l553:
			position = position0
			return false
		},
		/* 128 HtmlBlockScript <- (HtmlBlockOpenScript (HtmlBlockScript / (!HtmlBlockCloseScript .))* HtmlBlockCloseScript) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenScript]() {
				goto l555
			}
		l556:
			{
				position557 := position
				if !p.rules[ruleHtmlBlockScript]() {
					goto l559
				}
				goto l558
			l559:
				if !p.rules[ruleHtmlBlockCloseScript]() {
					goto l560
				}
				goto l557
			l560:
				if !matchDot() {
					goto l557
				}
			l558:
				goto l556
			l557:
				position = position557
			}
			if !p.rules[ruleHtmlBlockCloseScript]() {
				goto l555
			}
			return true
		l555:
			position = position0
			return false
		},
		/* 129 HtmlBlockInTags <- (HtmlBlockAddress / HtmlBlockBlockquote / HtmlBlockCenter / HtmlBlockDir / HtmlBlockDiv / HtmlBlockDl / HtmlBlockFieldset / HtmlBlockForm / HtmlBlockH1 / HtmlBlockH2 / HtmlBlockH3 / HtmlBlockH4 / HtmlBlockH5 / HtmlBlockH6 / HtmlBlockMenu / HtmlBlockNoframes / HtmlBlockNoscript / HtmlBlockOl / HtmlBlockP / HtmlBlockPre / HtmlBlockTable / HtmlBlockUl / HtmlBlockDd / HtmlBlockDt / HtmlBlockFrameset / HtmlBlockLi / HtmlBlockTbody / HtmlBlockTd / HtmlBlockTfoot / HtmlBlockTh / HtmlBlockThead / HtmlBlockTr / HtmlBlockScript) */
		func() bool {
			if !p.rules[ruleHtmlBlockAddress]() {
				goto l563
			}
			goto l562
		l563:
			if !p.rules[ruleHtmlBlockBlockquote]() {
				goto l564
			}
			goto l562
		l564:
			if !p.rules[ruleHtmlBlockCenter]() {
				goto l565
			}
			goto l562
		l565:
			if !p.rules[ruleHtmlBlockDir]() {
				goto l566
			}
			goto l562
		l566:
			if !p.rules[ruleHtmlBlockDiv]() {
				goto l567
			}
			goto l562
		l567:
			if !p.rules[ruleHtmlBlockDl]() {
				goto l568
			}
			goto l562
		l568:
			if !p.rules[ruleHtmlBlockFieldset]() {
				goto l569
			}
			goto l562
		l569:
			if !p.rules[ruleHtmlBlockForm]() {
				goto l570
			}
			goto l562
		l570:
			if !p.rules[ruleHtmlBlockH1]() {
				goto l571
			}
			goto l562
		l571:
			if !p.rules[ruleHtmlBlockH2]() {
				goto l572
			}
			goto l562
		l572:
			if !p.rules[ruleHtmlBlockH3]() {
				goto l573
			}
			goto l562
		l573:
			if !p.rules[ruleHtmlBlockH4]() {
				goto l574
			}
			goto l562
		l574:
			if !p.rules[ruleHtmlBlockH5]() {
				goto l575
			}
			goto l562
		l575:
			if !p.rules[ruleHtmlBlockH6]() {
				goto l576
			}
			goto l562
		l576:
			if !p.rules[ruleHtmlBlockMenu]() {
				goto l577
			}
			goto l562
		l577:
			if !p.rules[ruleHtmlBlockNoframes]() {
				goto l578
			}
			goto l562
		l578:
			if !p.rules[ruleHtmlBlockNoscript]() {
				goto l579
			}
			goto l562
		l579:
			if !p.rules[ruleHtmlBlockOl]() {
				goto l580
			}
			goto l562
		l580:
			if !p.rules[ruleHtmlBlockP]() {
				goto l581
			}
			goto l562
		l581:
			if !p.rules[ruleHtmlBlockPre]() {
				goto l582
			}
			goto l562
		l582:
			if !p.rules[ruleHtmlBlockTable]() {
				goto l583
			}
			goto l562
		l583:
			if !p.rules[ruleHtmlBlockUl]() {
				goto l584
			}
			goto l562
		l584:
			if !p.rules[ruleHtmlBlockDd]() {
				goto l585
			}
			goto l562
		l585:
			if !p.rules[ruleHtmlBlockDt]() {
				goto l586
			}
			goto l562
		l586:
			if !p.rules[ruleHtmlBlockFrameset]() {
				goto l587
			}
			goto l562
		l587:
			if !p.rules[ruleHtmlBlockLi]() {
				goto l588
			}
			goto l562
		l588:
			if !p.rules[ruleHtmlBlockTbody]() {
				goto l589
			}
			goto l562
		l589:
			if !p.rules[ruleHtmlBlockTd]() {
				goto l590
			}
			goto l562
		l590:
			if !p.rules[ruleHtmlBlockTfoot]() {
				goto l591
			}
			goto l562
		l591:
			if !p.rules[ruleHtmlBlockTh]() {
				goto l592
			}
			goto l562
		l592:
			if !p.rules[ruleHtmlBlockThead]() {
				goto l593
			}
			goto l562
		l593:
			if !p.rules[ruleHtmlBlockTr]() {
				goto l594
			}
			goto l562
		l594:
			if !p.rules[ruleHtmlBlockScript]() {
				goto l561
			}
		l562:
			return true
		l561:
			return false
		},
		/* 130 HtmlBlock <- (&'<' < (HtmlBlockInTags / HtmlComment / HtmlBlockSelfClosing) > BlankLine+ {   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTMLBLOCK
                }
            }) */
		func() bool {
			position0 := position
			if !peekChar('<') {
				goto l595
			}
			begin = position
			if !p.rules[ruleHtmlBlockInTags]() {
				goto l597
			}
			goto l596
		l597:
			if !p.rules[ruleHtmlComment]() {
				goto l598
			}
			goto l596
		l598:
			if !p.rules[ruleHtmlBlockSelfClosing]() {
				goto l595
			}
		l596:
			end = position
			if !p.rules[ruleBlankLine]() {
				goto l595
			}
		l599:
			if !p.rules[ruleBlankLine]() {
				goto l600
			}
			goto l599
		l600:
			do(40)
			return true
		l595:
			position = position0
			return false
		},
		/* 131 HtmlBlockSelfClosing <- ('<' Spnl HtmlBlockType Spnl HtmlAttribute* '/' Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l601
			}
			if !p.rules[ruleSpnl]() {
				goto l601
			}
			if !p.rules[ruleHtmlBlockType]() {
				goto l601
			}
			if !p.rules[ruleSpnl]() {
				goto l601
			}
		l602:
			if !p.rules[ruleHtmlAttribute]() {
				goto l603
			}
			goto l602
		l603:
			if !matchChar('/') {
				goto l601
			}
			if !p.rules[ruleSpnl]() {
				goto l601
			}
			if !matchChar('>') {
				goto l601
			}
			return true
		l601:
			position = position0
			return false
		},
		/* 132 HtmlBlockType <- ('dir' / 'div' / 'dl' / 'fieldset' / 'form' / 'h1' / 'h2' / 'h3' / 'h4' / 'h5' / 'h6' / 'noframes' / 'p' / 'table' / 'dd' / 'tbody' / 'td' / 'tfoot' / 'th' / 'thead' / 'DIR' / 'DIV' / 'DL' / 'FIELDSET' / 'FORM' / 'H1' / 'H2' / 'H3' / 'H4' / 'H5' / 'H6' / 'NOFRAMES' / 'P' / 'TABLE' / 'DD' / 'TBODY' / 'TD' / 'TFOOT' / 'TH' / 'THEAD' / ((&[S] 'SCRIPT') | (&[T] 'TR') | (&[L] 'LI') | (&[F] 'FRAMESET') | (&[D] 'DT') | (&[U] 'UL') | (&[P] 'PRE') | (&[O] 'OL') | (&[N] 'NOSCRIPT') | (&[M] 'MENU') | (&[I] 'ISINDEX') | (&[H] 'HR') | (&[C] 'CENTER') | (&[B] 'BLOCKQUOTE') | (&[A] 'ADDRESS') | (&[s] 'script') | (&[t] 'tr') | (&[l] 'li') | (&[f] 'frameset') | (&[d] 'dt') | (&[u] 'ul') | (&[p] 'pre') | (&[o] 'ol') | (&[n] 'noscript') | (&[m] 'menu') | (&[i] 'isindex') | (&[h] 'hr') | (&[c] 'center') | (&[b] 'blockquote') | (&[a] 'address'))) */
		func() bool {
			if !matchString("dir") {
				goto l606
			}
			goto l605
		l606:
			if !matchString("div") {
				goto l607
			}
			goto l605
		l607:
			if !matchString("dl") {
				goto l608
			}
			goto l605
		l608:
			if !matchString("fieldset") {
				goto l609
			}
			goto l605
		l609:
			if !matchString("form") {
				goto l610
			}
			goto l605
		l610:
			if !matchString("h1") {
				goto l611
			}
			goto l605
		l611:
			if !matchString("h2") {
				goto l612
			}
			goto l605
		l612:
			if !matchString("h3") {
				goto l613
			}
			goto l605
		l613:
			if !matchString("h4") {
				goto l614
			}
			goto l605
		l614:
			if !matchString("h5") {
				goto l615
			}
			goto l605
		l615:
			if !matchString("h6") {
				goto l616
			}
			goto l605
		l616:
			if !matchString("noframes") {
				goto l617
			}
			goto l605
		l617:
			if !matchChar('p') {
				goto l618
			}
			goto l605
		l618:
			if !matchString("table") {
				goto l619
			}
			goto l605
		l619:
			if !matchString("dd") {
				goto l620
			}
			goto l605
		l620:
			if !matchString("tbody") {
				goto l621
			}
			goto l605
		l621:
			if !matchString("td") {
				goto l622
			}
			goto l605
		l622:
			if !matchString("tfoot") {
				goto l623
			}
			goto l605
		l623:
			if !matchString("th") {
				goto l624
			}
			goto l605
		l624:
			if !matchString("thead") {
				goto l625
			}
			goto l605
		l625:
			if !matchString("DIR") {
				goto l626
			}
			goto l605
		l626:
			if !matchString("DIV") {
				goto l627
			}
			goto l605
		l627:
			if !matchString("DL") {
				goto l628
			}
			goto l605
		l628:
			if !matchString("FIELDSET") {
				goto l629
			}
			goto l605
		l629:
			if !matchString("FORM") {
				goto l630
			}
			goto l605
		l630:
			if !matchString("H1") {
				goto l631
			}
			goto l605
		l631:
			if !matchString("H2") {
				goto l632
			}
			goto l605
		l632:
			if !matchString("H3") {
				goto l633
			}
			goto l605
		l633:
			if !matchString("H4") {
				goto l634
			}
			goto l605
		l634:
			if !matchString("H5") {
				goto l635
			}
			goto l605
		l635:
			if !matchString("H6") {
				goto l636
			}
			goto l605
		l636:
			if !matchString("NOFRAMES") {
				goto l637
			}
			goto l605
		l637:
			if !matchChar('P') {
				goto l638
			}
			goto l605
		l638:
			if !matchString("TABLE") {
				goto l639
			}
			goto l605
		l639:
			if !matchString("DD") {
				goto l640
			}
			goto l605
		l640:
			if !matchString("TBODY") {
				goto l641
			}
			goto l605
		l641:
			if !matchString("TD") {
				goto l642
			}
			goto l605
		l642:
			if !matchString("TFOOT") {
				goto l643
			}
			goto l605
		l643:
			if !matchString("TH") {
				goto l644
			}
			goto l605
		l644:
			if !matchString("THEAD") {
				goto l645
			}
			goto l605
		l645:
			{
				if position == len(p.Buffer) {
					goto l604
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l604
					}
					break
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l604
					}
					break
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l604
					}
					break
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l604
					}
					break
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l604
					}
					break
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l604
					}
					break
				case 'P':
					position++
					if !matchString("RE") {
						goto l604
					}
					break
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l604
					}
					break
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l604
					}
					break
				case 'M':
					position++
					if !matchString("ENU") {
						goto l604
					}
					break
				case 'I':
					position++
					if !matchString("SINDEX") {
						goto l604
					}
					break
				case 'H':
					position++ // matchString(`HR`)
					if !matchChar('R') {
						goto l604
					}
					break
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l604
					}
					break
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l604
					}
					break
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l604
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l604
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l604
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l604
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l604
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l604
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l604
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l604
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l604
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l604
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l604
					}
					break
				case 'i':
					position++
					if !matchString("sindex") {
						goto l604
					}
					break
				case 'h':
					position++ // matchString(`hr`)
					if !matchChar('r') {
						goto l604
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l604
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l604
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l604
					}
					break
				default:
					goto l604
				}
			}
		l605:
			return true
		l604:
			return false
		},
		/* 133 StyleOpen <- ('<' Spnl ((&[S] 'STYLE') | (&[s] 'style')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l647
			}
			if !p.rules[ruleSpnl]() {
				goto l647
			}
			{
				if position == len(p.Buffer) {
					goto l647
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto l647
					}
					break
				case 's':
					position++
					if !matchString("tyle") {
						goto l647
					}
					break
				default:
					goto l647
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l647
			}
		l649:
			if !p.rules[ruleHtmlAttribute]() {
				goto l650
			}
			goto l649
		l650:
			if !matchChar('>') {
				goto l647
			}
			return true
		l647:
			position = position0
			return false
		},
		/* 134 StyleClose <- ('<' Spnl '/' ((&[S] 'STYLE') | (&[s] 'style')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l651
			}
			if !p.rules[ruleSpnl]() {
				goto l651
			}
			if !matchChar('/') {
				goto l651
			}
			{
				if position == len(p.Buffer) {
					goto l651
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto l651
					}
					break
				case 's':
					position++
					if !matchString("tyle") {
						goto l651
					}
					break
				default:
					goto l651
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l651
			}
			if !matchChar('>') {
				goto l651
			}
			return true
		l651:
			position = position0
			return false
		},
		/* 135 InStyleTags <- (StyleOpen (!StyleClose .)* StyleClose) */
		func() bool {
			position0 := position
			if !p.rules[ruleStyleOpen]() {
				goto l653
			}
		l654:
			{
				position655 := position
				if !p.rules[ruleStyleClose]() {
					goto l656
				}
				goto l655
			l656:
				if !matchDot() {
					goto l655
				}
				goto l654
			l655:
				position = position655
			}
			if !p.rules[ruleStyleClose]() {
				goto l653
			}
			return true
		l653:
			position = position0
			return false
		},
		/* 136 StyleBlock <- (< InStyleTags > BlankLine* {   if p.extension.FilterStyles {
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
				goto l657
			}
			end = position
		l658:
			if !p.rules[ruleBlankLine]() {
				goto l659
			}
			goto l658
		l659:
			do(41)
			return true
		l657:
			position = position0
			return false
		},
		/* 137 Inlines <- (StartList ((!Endline Inline { a = cons(yy, a) }) / (Endline &Inline { a = cons(c, a) }))+ Endline? { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l660
			}
			doarg(yySet, -1)
			{
				position663 := position
				if !p.rules[ruleEndline]() {
					goto l665
				}
				goto l664
			l665:
				if !p.rules[ruleInline]() {
					goto l664
				}
				do(42)
				goto l663
			l664:
				position = position663
				if !p.rules[ruleEndline]() {
					goto l660
				}
				doarg(yySet, -2)
				{
					position666 := position
					if !p.rules[ruleInline]() {
						goto l660
					}
					position = position666
				}
				do(43)
			}
		l663:
		l661:
			{
				position662, thunkPosition662 := position, thunkPosition
				{
					position667 := position
					if !p.rules[ruleEndline]() {
						goto l669
					}
					goto l668
				l669:
					if !p.rules[ruleInline]() {
						goto l668
					}
					do(42)
					goto l667
				l668:
					position = position667
					if !p.rules[ruleEndline]() {
						goto l662
					}
					doarg(yySet, -2)
					{
						position670 := position
						if !p.rules[ruleInline]() {
							goto l662
						}
						position = position670
					}
					do(43)
				}
			l667:
				goto l661
			l662:
				position, thunkPosition = position662, thunkPosition662
			}
			if !p.rules[ruleEndline]() {
				goto l671
			}
		l671:
			do(44)
			doarg(yyPop, 2)
			return true
		l660:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 138 Inline <- (Str / Endline / UlOrStarLine / Space / Strong / Emph / Image / Link / NoteReference / InlineNote / Code / RawHtml / Entity / EscapedChar / Smart / Symbol) */
		func() bool {
			if !p.rules[ruleStr]() {
				goto l675
			}
			goto l674
		l675:
			if !p.rules[ruleEndline]() {
				goto l676
			}
			goto l674
		l676:
			if !p.rules[ruleUlOrStarLine]() {
				goto l677
			}
			goto l674
		l677:
			if !p.rules[ruleSpace]() {
				goto l678
			}
			goto l674
		l678:
			if !p.rules[ruleStrong]() {
				goto l679
			}
			goto l674
		l679:
			if !p.rules[ruleEmph]() {
				goto l680
			}
			goto l674
		l680:
			if !p.rules[ruleImage]() {
				goto l681
			}
			goto l674
		l681:
			if !p.rules[ruleLink]() {
				goto l682
			}
			goto l674
		l682:
			if !p.rules[ruleNoteReference]() {
				goto l683
			}
			goto l674
		l683:
			if !p.rules[ruleInlineNote]() {
				goto l684
			}
			goto l674
		l684:
			if !p.rules[ruleCode]() {
				goto l685
			}
			goto l674
		l685:
			if !p.rules[ruleRawHtml]() {
				goto l686
			}
			goto l674
		l686:
			if !p.rules[ruleEntity]() {
				goto l687
			}
			goto l674
		l687:
			if !p.rules[ruleEscapedChar]() {
				goto l688
			}
			goto l674
		l688:
			if !p.rules[ruleSmart]() {
				goto l689
			}
			goto l674
		l689:
			if !p.rules[ruleSymbol]() {
				goto l673
			}
		l674:
			return true
		l673:
			return false
		},
		/* 139 Space <- (Spacechar+ { yy = mk_str(" ")
          yy.key = SPACE }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto l690
			}
		l691:
			if !p.rules[ruleSpacechar]() {
				goto l692
			}
			goto l691
		l692:
			do(45)
			return true
		l690:
			position = position0
			return false
		},
		/* 140 Str <- (< NormalChar (NormalChar / ('_'+ &Alphanumeric))* > { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleNormalChar]() {
				goto l693
			}
		l694:
			{
				position695 := position
				if !p.rules[ruleNormalChar]() {
					goto l697
				}
				goto l696
			l697:
				if !matchChar('_') {
					goto l695
				}
			l698:
				if !matchChar('_') {
					goto l699
				}
				goto l698
			l699:
				{
					position700 := position
					if !p.rules[ruleAlphanumeric]() {
						goto l695
					}
					position = position700
				}
			l696:
				goto l694
			l695:
				position = position695
			}
			end = position
			do(46)
			return true
		l693:
			position = position0
			return false
		},
		/* 141 EscapedChar <- ('\\' !Newline < [-\\`|*_{}[\]()#+.!><] > { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !matchChar('\\') {
				goto l701
			}
			if !p.rules[ruleNewline]() {
				goto l702
			}
			goto l701
		l702:
			begin = position
			if !matchClass(1) {
				goto l701
			}
			end = position
			do(47)
			return true
		l701:
			position = position0
			return false
		},
		/* 142 Entity <- ((HexEntity / DecEntity / CharEntity) { yy = mk_str(yytext); yy.key = HTML }) */
		func() bool {
			position0 := position
			if !p.rules[ruleHexEntity]() {
				goto l705
			}
			goto l704
		l705:
			if !p.rules[ruleDecEntity]() {
				goto l706
			}
			goto l704
		l706:
			if !p.rules[ruleCharEntity]() {
				goto l703
			}
		l704:
			do(48)
			return true
		l703:
			position = position0
			return false
		},
		/* 143 Endline <- (LineBreak / TerminalEndline / NormalEndline) */
		func() bool {
			if !p.rules[ruleLineBreak]() {
				goto l709
			}
			goto l708
		l709:
			if !p.rules[ruleTerminalEndline]() {
				goto l710
			}
			goto l708
		l710:
			if !p.rules[ruleNormalEndline]() {
				goto l707
			}
		l708:
			return true
		l707:
			return false
		},
		/* 144 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line ((&[\-] ('---' '-'*)) | (&[=] ('===' '='*))) Newline) { yy = mk_str("\n")
                    yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l711
			}
			if !p.rules[ruleNewline]() {
				goto l711
			}
			if !p.rules[ruleBlankLine]() {
				goto l712
			}
			goto l711
		l712:
			if peekChar('>') {
				goto l711
			}
			if !p.rules[ruleAtxStart]() {
				goto l713
			}
			goto l711
		l713:
			{
				position714, thunkPosition714 := position, thunkPosition
				if !p.rules[ruleLine]() {
					goto l714
				}
				{
					if position == len(p.Buffer) {
						goto l714
					}
					switch p.Buffer[position] {
					case '-':
						position++
						if !matchString("--") {
							goto l714
						}
					l716:
						if !matchChar('-') {
							goto l717
						}
						goto l716
					l717:
						break
					case '=':
						position++
						if !matchString("==") {
							goto l714
						}
					l718:
						if !matchChar('=') {
							goto l719
						}
						goto l718
					l719:
						break
					default:
						goto l714
					}
				}
				if !p.rules[ruleNewline]() {
					goto l714
				}
				goto l711
			l714:
				position, thunkPosition = position714, thunkPosition714
			}
			do(49)
			return true
		l711:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 145 TerminalEndline <- (Sp Newline !. { yy = nil }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l720
			}
			if !p.rules[ruleNewline]() {
				goto l720
			}
			if (position < len(p.Buffer)) {
				goto l720
			}
			do(50)
			return true
		l720:
			position = position0
			return false
		},
		/* 146 LineBreak <- ('  ' NormalEndline { yy = mk_element(LINEBREAK) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("  ") {
				goto l721
			}
			if !p.rules[ruleNormalEndline]() {
				goto l721
			}
			do(51)
			return true
		l721:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 147 Symbol <- (< SpecialChar > { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleSpecialChar]() {
				goto l722
			}
			end = position
			do(52)
			return true
		l722:
			position = position0
			return false
		},
		/* 148 UlOrStarLine <- ((UlLine / StarLine) { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l725
			}
			goto l724
		l725:
			if !p.rules[ruleStarLine]() {
				goto l723
			}
		l724:
			do(53)
			return true
		l723:
			position = position0
			return false
		},
		/* 149 StarLine <- ((&[*] (< '****' '*'* >)) | (&[\t ] (< Spacechar '*'+ &Spacechar >))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l726
				}
				switch p.Buffer[position] {
				case '*':
					begin = position
					if !matchString("****") {
						goto l726
					}
				l728:
					if !matchChar('*') {
						goto l729
					}
					goto l728
				l729:
					end = position
					break
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto l726
					}
					if !matchChar('*') {
						goto l726
					}
				l730:
					if !matchChar('*') {
						goto l731
					}
					goto l730
				l731:
					{
						position732 := position
						if !p.rules[ruleSpacechar]() {
							goto l726
						}
						position = position732
					}
					end = position
					break
				default:
					goto l726
				}
			}
			return true
		l726:
			position = position0
			return false
		},
		/* 150 UlLine <- ((&[_] (< '____' '_'* >)) | (&[\t ] (< Spacechar '_'+ &Spacechar >))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l733
				}
				switch p.Buffer[position] {
				case '_':
					begin = position
					if !matchString("____") {
						goto l733
					}
				l735:
					if !matchChar('_') {
						goto l736
					}
					goto l735
				l736:
					end = position
					break
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto l733
					}
					if !matchChar('_') {
						goto l733
					}
				l737:
					if !matchChar('_') {
						goto l738
					}
					goto l737
				l738:
					{
						position739 := position
						if !p.rules[ruleSpacechar]() {
							goto l733
						}
						position = position739
					}
					end = position
					break
				default:
					goto l733
				}
			}
			return true
		l733:
			position = position0
			return false
		},
		/* 151 Emph <- ((&[_] EmphUl) | (&[*] EmphStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l740
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleEmphUl]() {
						goto l740
					}
					break
				case '*':
					if !p.rules[ruleEmphStar]() {
						goto l740
					}
					break
				default:
					goto l740
				}
			}
			return true
		l740:
			return false
		},
		/* 152 OneStarOpen <- (!StarLine '*' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleStarLine]() {
				goto l743
			}
			goto l742
		l743:
			if !matchChar('*') {
				goto l742
			}
			if !p.rules[ruleSpacechar]() {
				goto l744
			}
			goto l742
		l744:
			if !p.rules[ruleNewline]() {
				goto l745
			}
			goto l742
		l745:
			return true
		l742:
			position = position0
			return false
		},
		/* 153 OneStarClose <- (!Spacechar !Newline Inline !StrongStar '*' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l747
			}
			goto l746
		l747:
			if !p.rules[ruleNewline]() {
				goto l748
			}
			goto l746
		l748:
			if !p.rules[ruleInline]() {
				goto l746
			}
			doarg(yySet, -1)
			if !p.rules[ruleStrongStar]() {
				goto l749
			}
			goto l746
		l749:
			if !matchChar('*') {
				goto l746
			}
			do(54)
			doarg(yyPop, 1)
			return true
		l746:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 154 EmphStar <- (OneStarOpen StartList (!OneStarClose Inline { a = cons(yy, a) })* OneStarClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneStarOpen]() {
				goto l750
			}
			if !p.rules[ruleStartList]() {
				goto l750
			}
			doarg(yySet, -1)
		l751:
			{
				position752, thunkPosition752 := position, thunkPosition
				if !p.rules[ruleOneStarClose]() {
					goto l753
				}
				goto l752
			l753:
				if !p.rules[ruleInline]() {
					goto l752
				}
				do(55)
				goto l751
			l752:
				position, thunkPosition = position752, thunkPosition752
			}
			if !p.rules[ruleOneStarClose]() {
				goto l750
			}
			do(56)
			do(57)
			doarg(yyPop, 1)
			return true
		l750:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 155 OneUlOpen <- (!UlLine '_' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l755
			}
			goto l754
		l755:
			if !matchChar('_') {
				goto l754
			}
			if !p.rules[ruleSpacechar]() {
				goto l756
			}
			goto l754
		l756:
			if !p.rules[ruleNewline]() {
				goto l757
			}
			goto l754
		l757:
			return true
		l754:
			position = position0
			return false
		},
		/* 156 OneUlClose <- (!Spacechar !Newline Inline !StrongUl '_' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l759
			}
			goto l758
		l759:
			if !p.rules[ruleNewline]() {
				goto l760
			}
			goto l758
		l760:
			if !p.rules[ruleInline]() {
				goto l758
			}
			doarg(yySet, -1)
			if !p.rules[ruleStrongUl]() {
				goto l761
			}
			goto l758
		l761:
			if !matchChar('_') {
				goto l758
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l762
			}
			goto l758
		l762:
			do(58)
			doarg(yyPop, 1)
			return true
		l758:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 157 EmphUl <- (OneUlOpen StartList (!OneUlClose Inline { a = cons(yy, a) })* OneUlClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneUlOpen]() {
				goto l763
			}
			if !p.rules[ruleStartList]() {
				goto l763
			}
			doarg(yySet, -1)
		l764:
			{
				position765, thunkPosition765 := position, thunkPosition
				if !p.rules[ruleOneUlClose]() {
					goto l766
				}
				goto l765
			l766:
				if !p.rules[ruleInline]() {
					goto l765
				}
				do(59)
				goto l764
			l765:
				position, thunkPosition = position765, thunkPosition765
			}
			if !p.rules[ruleOneUlClose]() {
				goto l763
			}
			do(60)
			do(61)
			doarg(yyPop, 1)
			return true
		l763:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 158 Strong <- ((&[_] StrongUl) | (&[*] StrongStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l767
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleStrongUl]() {
						goto l767
					}
					break
				case '*':
					if !p.rules[ruleStrongStar]() {
						goto l767
					}
					break
				default:
					goto l767
				}
			}
			return true
		l767:
			return false
		},
		/* 159 TwoStarOpen <- (!StarLine '**' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleStarLine]() {
				goto l770
			}
			goto l769
		l770:
			if !matchString("**") {
				goto l769
			}
			if !p.rules[ruleSpacechar]() {
				goto l771
			}
			goto l769
		l771:
			if !p.rules[ruleNewline]() {
				goto l772
			}
			goto l769
		l772:
			return true
		l769:
			position = position0
			return false
		},
		/* 160 TwoStarClose <- (!Spacechar !Newline Inline '**' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l774
			}
			goto l773
		l774:
			if !p.rules[ruleNewline]() {
				goto l775
			}
			goto l773
		l775:
			if !p.rules[ruleInline]() {
				goto l773
			}
			doarg(yySet, -1)
			if !matchString("**") {
				goto l773
			}
			do(62)
			doarg(yyPop, 1)
			return true
		l773:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 161 StrongStar <- (TwoStarOpen StartList (!TwoStarClose Inline { a = cons(yy, a) })* TwoStarClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoStarOpen]() {
				goto l776
			}
			if !p.rules[ruleStartList]() {
				goto l776
			}
			doarg(yySet, -1)
		l777:
			{
				position778, thunkPosition778 := position, thunkPosition
				if !p.rules[ruleTwoStarClose]() {
					goto l779
				}
				goto l778
			l779:
				if !p.rules[ruleInline]() {
					goto l778
				}
				do(63)
				goto l777
			l778:
				position, thunkPosition = position778, thunkPosition778
			}
			if !p.rules[ruleTwoStarClose]() {
				goto l776
			}
			do(64)
			do(65)
			doarg(yyPop, 1)
			return true
		l776:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 162 TwoUlOpen <- (!UlLine '__' !Spacechar !Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l781
			}
			goto l780
		l781:
			if !matchString("__") {
				goto l780
			}
			if !p.rules[ruleSpacechar]() {
				goto l782
			}
			goto l780
		l782:
			if !p.rules[ruleNewline]() {
				goto l783
			}
			goto l780
		l783:
			return true
		l780:
			position = position0
			return false
		},
		/* 163 TwoUlClose <- (!Spacechar !Newline Inline '__' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSpacechar]() {
				goto l785
			}
			goto l784
		l785:
			if !p.rules[ruleNewline]() {
				goto l786
			}
			goto l784
		l786:
			if !p.rules[ruleInline]() {
				goto l784
			}
			doarg(yySet, -1)
			if !matchString("__") {
				goto l784
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l787
			}
			goto l784
		l787:
			do(66)
			doarg(yyPop, 1)
			return true
		l784:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 164 StrongUl <- (TwoUlOpen StartList (!TwoUlClose Inline { a = cons(yy, a) })* TwoUlClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoUlOpen]() {
				goto l788
			}
			if !p.rules[ruleStartList]() {
				goto l788
			}
			doarg(yySet, -1)
		l789:
			{
				position790, thunkPosition790 := position, thunkPosition
				if !p.rules[ruleTwoUlClose]() {
					goto l791
				}
				goto l790
			l791:
				if !p.rules[ruleInline]() {
					goto l790
				}
				do(67)
				goto l789
			l790:
				position, thunkPosition = position790, thunkPosition790
			}
			if !p.rules[ruleTwoUlClose]() {
				goto l788
			}
			do(68)
			do(69)
			doarg(yyPop, 1)
			return true
		l788:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 165 Image <- ('!' (ExplicitLink / ReferenceLink) {	if yy.key == LINK {
			yy.key = IMAGE
		} else {
			result := yy
			yy.children = cons(mk_str("!"), result.children)
		}
	}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('!') {
				goto l792
			}
			if !p.rules[ruleExplicitLink]() {
				goto l794
			}
			goto l793
		l794:
			if !p.rules[ruleReferenceLink]() {
				goto l792
			}
		l793:
			do(70)
			return true
		l792:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 166 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() bool {
			if !p.rules[ruleExplicitLink]() {
				goto l797
			}
			goto l796
		l797:
			if !p.rules[ruleReferenceLink]() {
				goto l798
			}
			goto l796
		l798:
			if !p.rules[ruleAutoLink]() {
				goto l795
			}
		l796:
			return true
		l795:
			return false
		},
		/* 167 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() bool {
			if !p.rules[ruleReferenceLinkDouble]() {
				goto l801
			}
			goto l800
		l801:
			if !p.rules[ruleReferenceLinkSingle]() {
				goto l799
			}
		l800:
			return true
		l799:
			return false
		},
		/* 168 ReferenceLinkDouble <- (Label < Spnl > !'[]' Label {
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
				goto l802
			}
			doarg(yySet, -2)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto l802
			}
			end = position
			if !matchString("[]") {
				goto l803
			}
			goto l802
		l803:
			if !p.rules[ruleLabel]() {
				goto l802
			}
			doarg(yySet, -1)
			do(71)
			doarg(yyPop, 2)
			return true
		l802:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 169 ReferenceLinkSingle <- (Label < (Spnl '[]')? > {
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
				goto l804
			}
			doarg(yySet, -1)
			begin = position
			{
				position805 := position
				if !p.rules[ruleSpnl]() {
					goto l805
				}
				if !matchString("[]") {
					goto l805
				}
				goto l806
			l805:
				position = position805
			}
		l806:
			end = position
			do(72)
			doarg(yyPop, 1)
			return true
		l804:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 170 ExplicitLink <- (Label Spnl '(' Sp Source Spnl Title Sp ')' { yy = mk_link(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleLabel]() {
				goto l807
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l807
			}
			if !matchChar('(') {
				goto l807
			}
			if !p.rules[ruleSp]() {
				goto l807
			}
			if !p.rules[ruleSource]() {
				goto l807
			}
			doarg(yySet, -2)
			if !p.rules[ruleSpnl]() {
				goto l807
			}
			if !p.rules[ruleTitle]() {
				goto l807
			}
			doarg(yySet, -3)
			if !p.rules[ruleSp]() {
				goto l807
			}
			if !matchChar(')') {
				goto l807
			}
			do(73)
			doarg(yyPop, 3)
			return true
		l807:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 171 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			{
				position809 := position
				if !matchChar('<') {
					goto l810
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l810
				}
				end = position
				if !matchChar('>') {
					goto l810
				}
				goto l809
			l810:
				position = position809
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l808
				}
				end = position
			}
		l809:
			do(74)
			return true
		l808:
			position = position0
			return false
		},
		/* 172 SourceContents <- (((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* / '') */
		func() bool {
		l814:
			{
				position815 := position
				if position == len(p.Buffer) {
					goto l817
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l817
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l817
					}
				}
			l818:
				if position == len(p.Buffer) {
					goto l819
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l819
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l819
					}
				}
				goto l818
			l819:
				goto l816
			l817:
				if !matchChar('(') {
					goto l815
				}
				if !p.rules[ruleSourceContents]() {
					goto l815
				}
				if !matchChar(')') {
					goto l815
				}
			l816:
				goto l814
			l815:
				position = position815
			}
			goto l812
		l812:
			return true
		},
		/* 173 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = mk_str(yytext) }) */
		func() bool {
			if !p.rules[ruleTitleSingle]() {
				goto l822
			}
			goto l821
		l822:
			if !p.rules[ruleTitleDouble]() {
				goto l823
			}
			goto l821
		l823:
			begin = position
			end = position
		l821:
			do(75)
			return true
		},
		/* 174 TitleSingle <- ('\'' < (!('\'' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '\'') */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l824
			}
			begin = position
		l825:
			{
				position826 := position
				{
					position827 := position
					if !matchChar('\'') {
						goto l827
					}
					if !p.rules[ruleSp]() {
						goto l827
					}
					{
						if position == len(p.Buffer) {
							goto l827
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l827
							}
							break
						default:
							goto l827
						}
					}
					goto l826
				l827:
					position = position827
				}
				if !matchDot() {
					goto l826
				}
				goto l825
			l826:
				position = position826
			}
			end = position
			if !matchChar('\'') {
				goto l824
			}
			return true
		l824:
			position = position0
			return false
		},
		/* 175 TitleDouble <- ('"' < (!('"' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '"') */
		func() bool {
			position0 := position
			if !matchChar('"') {
				goto l829
			}
			begin = position
		l830:
			{
				position831 := position
				{
					position832 := position
					if !matchChar('"') {
						goto l832
					}
					if !p.rules[ruleSp]() {
						goto l832
					}
					{
						if position == len(p.Buffer) {
							goto l832
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l832
							}
							break
						default:
							goto l832
						}
					}
					goto l831
				l832:
					position = position832
				}
				if !matchDot() {
					goto l831
				}
				goto l830
			l831:
				position = position831
			}
			end = position
			if !matchChar('"') {
				goto l829
			}
			return true
		l829:
			position = position0
			return false
		},
		/* 176 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() bool {
			if !p.rules[ruleAutoLinkUrl]() {
				goto l836
			}
			goto l835
		l836:
			if !p.rules[ruleAutoLinkEmail]() {
				goto l834
			}
		l835:
			return true
		l834:
			return false
		},
		/* 177 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = mk_link(mk_str(yytext), yytext, "") }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l837
			}
			begin = position
			if !matchClass(2) {
				goto l837
			}
		l838:
			if !matchClass(2) {
				goto l839
			}
			goto l838
		l839:
			if !matchString("://") {
				goto l837
			}
			if !p.rules[ruleNewline]() {
				goto l842
			}
			goto l837
		l842:
			if peekChar('>') {
				goto l837
			}
			if !matchDot() {
				goto l837
			}
		l840:
			{
				position841 := position
				if !p.rules[ruleNewline]() {
					goto l843
				}
				goto l841
			l843:
				if peekChar('>') {
					goto l841
				}
				if !matchDot() {
					goto l841
				}
				goto l840
			l841:
				position = position841
			}
			end = position
			if !matchChar('>') {
				goto l837
			}
			do(76)
			return true
		l837:
			position = position0
			return false
		},
		/* 178 AutoLinkEmail <- ('<' < [-A-Za-z0-9+_]+ '@' (!Newline !'>' .)+ > '>' {
                    yy = mk_link(mk_str(yytext), "mailto:"+yytext, "")
                }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l844
			}
			begin = position
			if !matchClass(3) {
				goto l844
			}
		l845:
			if !matchClass(3) {
				goto l846
			}
			goto l845
		l846:
			if !matchChar('@') {
				goto l844
			}
			if !p.rules[ruleNewline]() {
				goto l849
			}
			goto l844
		l849:
			if peekChar('>') {
				goto l844
			}
			if !matchDot() {
				goto l844
			}
		l847:
			{
				position848 := position
				if !p.rules[ruleNewline]() {
					goto l850
				}
				goto l848
			l850:
				if peekChar('>') {
					goto l848
				}
				if !matchDot() {
					goto l848
				}
				goto l847
			l848:
				position = position848
			}
			end = position
			if !matchChar('>') {
				goto l844
			}
			do(77)
			return true
		l844:
			position = position0
			return false
		},
		/* 179 Reference <- (NonindentSpace !'[]' Label ':' Spnl RefSrc Spnl RefTitle BlankLine* { yy = mk_link(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleNonindentSpace]() {
				goto l851
			}
			if !matchString("[]") {
				goto l852
			}
			goto l851
		l852:
			if !p.rules[ruleLabel]() {
				goto l851
			}
			doarg(yySet, -2)
			if !matchChar(':') {
				goto l851
			}
			if !p.rules[ruleSpnl]() {
				goto l851
			}
			if !p.rules[ruleRefSrc]() {
				goto l851
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l851
			}
			if !p.rules[ruleRefTitle]() {
				goto l851
			}
			doarg(yySet, -3)
		l853:
			if !p.rules[ruleBlankLine]() {
				goto l854
			}
			goto l853
		l854:
			do(78)
			doarg(yyPop, 3)
			return true
		l851:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 180 Label <- ('[' ((!'^' &{p.extension.Notes}) / (&. &{!p.extension.Notes})) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l855
			}
			if peekChar('^') {
				goto l857
			}
			if !(p.extension.Notes) {
				goto l857
			}
			goto l856
		l857:
			if !(position < len(p.Buffer)) {
				goto l855
			}
			if !(!p.extension.Notes) {
				goto l855
			}
		l856:
			if !p.rules[ruleStartList]() {
				goto l855
			}
			doarg(yySet, -1)
		l858:
			{
				position859 := position
				if peekChar(']') {
					goto l859
				}
				if !p.rules[ruleInline]() {
					goto l859
				}
				do(79)
				goto l858
			l859:
				position = position859
			}
			if !matchChar(']') {
				goto l855
			}
			do(80)
			doarg(yyPop, 1)
			return true
		l855:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 181 RefSrc <- (< Nonspacechar+ > { yy = mk_str(yytext)
           yy.key = HTML }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto l860
			}
		l861:
			if !p.rules[ruleNonspacechar]() {
				goto l862
			}
			goto l861
		l862:
			end = position
			do(81)
			return true
		l860:
			position = position0
			return false
		},
		/* 182 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRefTitleSingle]() {
				goto l865
			}
			goto l864
		l865:
			if !p.rules[ruleRefTitleDouble]() {
				goto l866
			}
			goto l864
		l866:
			if !p.rules[ruleRefTitleParens]() {
				goto l867
			}
			goto l864
		l867:
			if !p.rules[ruleEmptyTitle]() {
				goto l863
			}
		l864:
			do(82)
			return true
		l863:
			position = position0
			return false
		},
		/* 183 EmptyTitle <- (< '' >) */
		func() bool {
			begin = position
			end = position
			return true
		},
		/* 184 RefTitleSingle <- ('\'' < (!((&[\'] ('\'' Sp Newline)) | (&[\n\r] Newline)) .)* > '\'') */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l869
			}
			begin = position
		l870:
			{
				position871 := position
				{
					position872 := position
					{
						if position == len(p.Buffer) {
							goto l872
						}
						switch p.Buffer[position] {
						case '\'':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l872
							}
							if !p.rules[ruleNewline]() {
								goto l872
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l872
							}
							break
						default:
							goto l872
						}
					}
					goto l871
				l872:
					position = position872
				}
				if !matchDot() {
					goto l871
				}
				goto l870
			l871:
				position = position871
			}
			end = position
			if !matchChar('\'') {
				goto l869
			}
			return true
		l869:
			position = position0
			return false
		},
		/* 185 RefTitleDouble <- ('"' < (!((&[\"] ('"' Sp Newline)) | (&[\n\r] Newline)) .)* > '"') */
		func() bool {
			position0 := position
			if !matchChar('"') {
				goto l874
			}
			begin = position
		l875:
			{
				position876 := position
				{
					position877 := position
					{
						if position == len(p.Buffer) {
							goto l877
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l877
							}
							if !p.rules[ruleNewline]() {
								goto l877
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l877
							}
							break
						default:
							goto l877
						}
					}
					goto l876
				l877:
					position = position877
				}
				if !matchDot() {
					goto l876
				}
				goto l875
			l876:
				position = position876
			}
			end = position
			if !matchChar('"') {
				goto l874
			}
			return true
		l874:
			position = position0
			return false
		},
		/* 186 RefTitleParens <- ('(' < (!((&[)] (')' Sp Newline)) | (&[\n\r] Newline)) .)* > ')') */
		func() bool {
			position0 := position
			if !matchChar('(') {
				goto l879
			}
			begin = position
		l880:
			{
				position881 := position
				{
					position882 := position
					{
						if position == len(p.Buffer) {
							goto l882
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l882
							}
							if !p.rules[ruleNewline]() {
								goto l882
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l882
							}
							break
						default:
							goto l882
						}
					}
					goto l881
				l882:
					position = position882
				}
				if !matchDot() {
					goto l881
				}
				goto l880
			l881:
				position = position881
			}
			end = position
			if !matchChar(')') {
				goto l879
			}
			return true
		l879:
			position = position0
			return false
		},
		/* 187 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l884
			}
			doarg(yySet, -2)
		l885:
			{
				position886, thunkPosition886 := position, thunkPosition
				{
					position887, thunkPosition887 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto l888
					}
					doarg(yySet, -1)
					do(83)
					goto l887
				l888:
					position, thunkPosition = position887, thunkPosition887
					if !p.rules[ruleSkipBlock]() {
						goto l886
					}
				}
			l887:
				goto l885
			l886:
				position, thunkPosition = position886, thunkPosition886
			}
			do(84)
			if !(commit(thunkPosition0)) {
				goto l884
			}
			doarg(yyPop, 2)
			return true
		l884:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 188 Ticks1 <- ('`' !'`') */
		func() bool {
			position0 := position
			if !matchChar('`') {
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
		/* 189 Ticks2 <- ('``' !'`') */
		func() bool {
			position0 := position
			if !matchString("``") {
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
		/* 190 Ticks3 <- ('```' !'`') */
		func() bool {
			position0 := position
			if !matchString("```") {
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
		/* 191 Ticks4 <- ('````' !'`') */
		func() bool {
			position0 := position
			if !matchString("````") {
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
		/* 192 Ticks5 <- ('`````' !'`') */
		func() bool {
			position0 := position
			if !matchString("`````") {
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
		/* 193 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks1 '`'+)) | (&[\t\n\r ] (!(Sp Ticks1) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks2 '`'+)) | (&[\t\n\r ] (!(Sp Ticks2) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks3 '`'+)) | (&[\t\n\r ] (!(Sp Ticks3) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks4 '`'+)) | (&[\t\n\r ] (!(Sp Ticks4) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks5 '`'+)) | (&[\t\n\r ] (!(Sp Ticks5) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks5)) { yy = mk_str(yytext); yy.key = CODE }) */
		func() bool {
			position0 := position
			{
				position895 := position
				if !p.rules[ruleTicks1]() {
					goto l896
				}
				if !p.rules[ruleSp]() {
					goto l896
				}
				begin = position
				if peekChar('`') {
					goto l900
				}
				if !p.rules[ruleNonspacechar]() {
					goto l900
				}
			l901:
				if peekChar('`') {
					goto l902
				}
				if !p.rules[ruleNonspacechar]() {
					goto l902
				}
				goto l901
			l902:
				goto l899
			l900:
				{
					if position == len(p.Buffer) {
						goto l896
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks1]() {
							goto l904
						}
						goto l896
					l904:
						if !matchChar('`') {
							goto l896
						}
					l905:
						if !matchChar('`') {
							goto l906
						}
						goto l905
					l906:
						break
					default:
						{
							position907 := position
							if !p.rules[ruleSp]() {
								goto l907
							}
							if !p.rules[ruleTicks1]() {
								goto l907
							}
							goto l896
						l907:
							position = position907
						}
						{
							if position == len(p.Buffer) {
								goto l896
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l896
								}
								if !p.rules[ruleBlankLine]() {
									goto l909
								}
								goto l896
							l909:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l896
								}
								break
							default:
								goto l896
							}
						}
					}
				}
			l899:
			l897:
				{
					position898 := position
					if peekChar('`') {
						goto l911
					}
					if !p.rules[ruleNonspacechar]() {
						goto l911
					}
				l912:
					if peekChar('`') {
						goto l913
					}
					if !p.rules[ruleNonspacechar]() {
						goto l913
					}
					goto l912
				l913:
					goto l910
				l911:
					{
						if position == len(p.Buffer) {
							goto l898
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks1]() {
								goto l915
							}
							goto l898
						l915:
							if !matchChar('`') {
								goto l898
							}
						l916:
							if !matchChar('`') {
								goto l917
							}
							goto l916
						l917:
							break
						default:
							{
								position918 := position
								if !p.rules[ruleSp]() {
									goto l918
								}
								if !p.rules[ruleTicks1]() {
									goto l918
								}
								goto l898
							l918:
								position = position918
							}
							{
								if position == len(p.Buffer) {
									goto l898
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l898
									}
									if !p.rules[ruleBlankLine]() {
										goto l920
									}
									goto l898
								l920:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l898
									}
									break
								default:
									goto l898
								}
							}
						}
					}
				l910:
					goto l897
				l898:
					position = position898
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l896
				}
				if !p.rules[ruleTicks1]() {
					goto l896
				}
				goto l895
			l896:
				position = position895
				if !p.rules[ruleTicks2]() {
					goto l921
				}
				if !p.rules[ruleSp]() {
					goto l921
				}
				begin = position
				if peekChar('`') {
					goto l925
				}
				if !p.rules[ruleNonspacechar]() {
					goto l925
				}
			l926:
				if peekChar('`') {
					goto l927
				}
				if !p.rules[ruleNonspacechar]() {
					goto l927
				}
				goto l926
			l927:
				goto l924
			l925:
				{
					if position == len(p.Buffer) {
						goto l921
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks2]() {
							goto l929
						}
						goto l921
					l929:
						if !matchChar('`') {
							goto l921
						}
					l930:
						if !matchChar('`') {
							goto l931
						}
						goto l930
					l931:
						break
					default:
						{
							position932 := position
							if !p.rules[ruleSp]() {
								goto l932
							}
							if !p.rules[ruleTicks2]() {
								goto l932
							}
							goto l921
						l932:
							position = position932
						}
						{
							if position == len(p.Buffer) {
								goto l921
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l921
								}
								if !p.rules[ruleBlankLine]() {
									goto l934
								}
								goto l921
							l934:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l921
								}
								break
							default:
								goto l921
							}
						}
					}
				}
			l924:
			l922:
				{
					position923 := position
					if peekChar('`') {
						goto l936
					}
					if !p.rules[ruleNonspacechar]() {
						goto l936
					}
				l937:
					if peekChar('`') {
						goto l938
					}
					if !p.rules[ruleNonspacechar]() {
						goto l938
					}
					goto l937
				l938:
					goto l935
				l936:
					{
						if position == len(p.Buffer) {
							goto l923
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks2]() {
								goto l940
							}
							goto l923
						l940:
							if !matchChar('`') {
								goto l923
							}
						l941:
							if !matchChar('`') {
								goto l942
							}
							goto l941
						l942:
							break
						default:
							{
								position943 := position
								if !p.rules[ruleSp]() {
									goto l943
								}
								if !p.rules[ruleTicks2]() {
									goto l943
								}
								goto l923
							l943:
								position = position943
							}
							{
								if position == len(p.Buffer) {
									goto l923
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l923
									}
									if !p.rules[ruleBlankLine]() {
										goto l945
									}
									goto l923
								l945:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l923
									}
									break
								default:
									goto l923
								}
							}
						}
					}
				l935:
					goto l922
				l923:
					position = position923
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l921
				}
				if !p.rules[ruleTicks2]() {
					goto l921
				}
				goto l895
			l921:
				position = position895
				if !p.rules[ruleTicks3]() {
					goto l946
				}
				if !p.rules[ruleSp]() {
					goto l946
				}
				begin = position
				if peekChar('`') {
					goto l950
				}
				if !p.rules[ruleNonspacechar]() {
					goto l950
				}
			l951:
				if peekChar('`') {
					goto l952
				}
				if !p.rules[ruleNonspacechar]() {
					goto l952
				}
				goto l951
			l952:
				goto l949
			l950:
				{
					if position == len(p.Buffer) {
						goto l946
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks3]() {
							goto l954
						}
						goto l946
					l954:
						if !matchChar('`') {
							goto l946
						}
					l955:
						if !matchChar('`') {
							goto l956
						}
						goto l955
					l956:
						break
					default:
						{
							position957 := position
							if !p.rules[ruleSp]() {
								goto l957
							}
							if !p.rules[ruleTicks3]() {
								goto l957
							}
							goto l946
						l957:
							position = position957
						}
						{
							if position == len(p.Buffer) {
								goto l946
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l946
								}
								if !p.rules[ruleBlankLine]() {
									goto l959
								}
								goto l946
							l959:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l946
								}
								break
							default:
								goto l946
							}
						}
					}
				}
			l949:
			l947:
				{
					position948 := position
					if peekChar('`') {
						goto l961
					}
					if !p.rules[ruleNonspacechar]() {
						goto l961
					}
				l962:
					if peekChar('`') {
						goto l963
					}
					if !p.rules[ruleNonspacechar]() {
						goto l963
					}
					goto l962
				l963:
					goto l960
				l961:
					{
						if position == len(p.Buffer) {
							goto l948
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks3]() {
								goto l965
							}
							goto l948
						l965:
							if !matchChar('`') {
								goto l948
							}
						l966:
							if !matchChar('`') {
								goto l967
							}
							goto l966
						l967:
							break
						default:
							{
								position968 := position
								if !p.rules[ruleSp]() {
									goto l968
								}
								if !p.rules[ruleTicks3]() {
									goto l968
								}
								goto l948
							l968:
								position = position968
							}
							{
								if position == len(p.Buffer) {
									goto l948
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l948
									}
									if !p.rules[ruleBlankLine]() {
										goto l970
									}
									goto l948
								l970:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l948
									}
									break
								default:
									goto l948
								}
							}
						}
					}
				l960:
					goto l947
				l948:
					position = position948
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l946
				}
				if !p.rules[ruleTicks3]() {
					goto l946
				}
				goto l895
			l946:
				position = position895
				if !p.rules[ruleTicks4]() {
					goto l971
				}
				if !p.rules[ruleSp]() {
					goto l971
				}
				begin = position
				if peekChar('`') {
					goto l975
				}
				if !p.rules[ruleNonspacechar]() {
					goto l975
				}
			l976:
				if peekChar('`') {
					goto l977
				}
				if !p.rules[ruleNonspacechar]() {
					goto l977
				}
				goto l976
			l977:
				goto l974
			l975:
				{
					if position == len(p.Buffer) {
						goto l971
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks4]() {
							goto l979
						}
						goto l971
					l979:
						if !matchChar('`') {
							goto l971
						}
					l980:
						if !matchChar('`') {
							goto l981
						}
						goto l980
					l981:
						break
					default:
						{
							position982 := position
							if !p.rules[ruleSp]() {
								goto l982
							}
							if !p.rules[ruleTicks4]() {
								goto l982
							}
							goto l971
						l982:
							position = position982
						}
						{
							if position == len(p.Buffer) {
								goto l971
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l971
								}
								if !p.rules[ruleBlankLine]() {
									goto l984
								}
								goto l971
							l984:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l971
								}
								break
							default:
								goto l971
							}
						}
					}
				}
			l974:
			l972:
				{
					position973 := position
					if peekChar('`') {
						goto l986
					}
					if !p.rules[ruleNonspacechar]() {
						goto l986
					}
				l987:
					if peekChar('`') {
						goto l988
					}
					if !p.rules[ruleNonspacechar]() {
						goto l988
					}
					goto l987
				l988:
					goto l985
				l986:
					{
						if position == len(p.Buffer) {
							goto l973
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks4]() {
								goto l990
							}
							goto l973
						l990:
							if !matchChar('`') {
								goto l973
							}
						l991:
							if !matchChar('`') {
								goto l992
							}
							goto l991
						l992:
							break
						default:
							{
								position993 := position
								if !p.rules[ruleSp]() {
									goto l993
								}
								if !p.rules[ruleTicks4]() {
									goto l993
								}
								goto l973
							l993:
								position = position993
							}
							{
								if position == len(p.Buffer) {
									goto l973
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l973
									}
									if !p.rules[ruleBlankLine]() {
										goto l995
									}
									goto l973
								l995:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l973
									}
									break
								default:
									goto l973
								}
							}
						}
					}
				l985:
					goto l972
				l973:
					position = position973
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l971
				}
				if !p.rules[ruleTicks4]() {
					goto l971
				}
				goto l895
			l971:
				position = position895
				if !p.rules[ruleTicks5]() {
					goto l894
				}
				if !p.rules[ruleSp]() {
					goto l894
				}
				begin = position
				if peekChar('`') {
					goto l999
				}
				if !p.rules[ruleNonspacechar]() {
					goto l999
				}
			l1000:
				if peekChar('`') {
					goto l1001
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1001
				}
				goto l1000
			l1001:
				goto l998
			l999:
				{
					if position == len(p.Buffer) {
						goto l894
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks5]() {
							goto l1003
						}
						goto l894
					l1003:
						if !matchChar('`') {
							goto l894
						}
					l1004:
						if !matchChar('`') {
							goto l1005
						}
						goto l1004
					l1005:
						break
					default:
						{
							position1006 := position
							if !p.rules[ruleSp]() {
								goto l1006
							}
							if !p.rules[ruleTicks5]() {
								goto l1006
							}
							goto l894
						l1006:
							position = position1006
						}
						{
							if position == len(p.Buffer) {
								goto l894
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l894
								}
								if !p.rules[ruleBlankLine]() {
									goto l1008
								}
								goto l894
							l1008:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l894
								}
								break
							default:
								goto l894
							}
						}
					}
				}
			l998:
			l996:
				{
					position997 := position
					if peekChar('`') {
						goto l1010
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1010
					}
				l1011:
					if peekChar('`') {
						goto l1012
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1012
					}
					goto l1011
				l1012:
					goto l1009
				l1010:
					{
						if position == len(p.Buffer) {
							goto l997
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks5]() {
								goto l1014
							}
							goto l997
						l1014:
							if !matchChar('`') {
								goto l997
							}
						l1015:
							if !matchChar('`') {
								goto l1016
							}
							goto l1015
						l1016:
							break
						default:
							{
								position1017 := position
								if !p.rules[ruleSp]() {
									goto l1017
								}
								if !p.rules[ruleTicks5]() {
									goto l1017
								}
								goto l997
							l1017:
								position = position1017
							}
							{
								if position == len(p.Buffer) {
									goto l997
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l997
									}
									if !p.rules[ruleBlankLine]() {
										goto l1019
									}
									goto l997
								l1019:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l997
									}
									break
								default:
									goto l997
								}
							}
						}
					}
				l1009:
					goto l996
				l997:
					position = position997
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l894
				}
				if !p.rules[ruleTicks5]() {
					goto l894
				}
			}
		l895:
			do(85)
			return true
		l894:
			position = position0
			return false
		},
		/* 194 RawHtml <- (< (HtmlComment / HtmlTag) > {   if p.extension.FilterHTML {
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
				goto l1022
			}
			goto l1021
		l1022:
			if !p.rules[ruleHtmlTag]() {
				goto l1020
			}
		l1021:
			end = position
			do(86)
			return true
		l1020:
			position = position0
			return false
		},
		/* 195 BlankLine <- (Sp Newline) */
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
		/* 196 Quoted <- ((&[\'] ('\'' (!'\'' .)* '\'')) | (&[\"] ('"' (!'"' .)* '"'))) */
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
		/* 197 HtmlAttribute <- (((&[\-] '-') | (&[0-9A-Za-z] [A-Za-z0-9]))+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
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
					if !matchClass(6) {
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
					if !matchClass(6) {
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
		/* 198 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
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
		/* 199 HtmlTag <- ('<' Spnl '/'? [A-Za-z0-9]+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l1045
			}
			if !p.rules[ruleSpnl]() {
				goto l1045
			}
			matchChar('/')
			if !matchClass(6) {
				goto l1045
			}
		l1046:
			if !matchClass(6) {
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
		/* 200 Eof <- !. */
		func() bool {
			if (position < len(p.Buffer)) {
				goto l1050
			}
			return true
		l1050:
			return false
		},
		/* 201 Spacechar <- ((&[\t] '\t') | (&[ ] ' ')) */
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
		/* 202 Nonspacechar <- (!Spacechar !Newline .) */
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
		/* 203 Newline <- ((&[\r] ('\r' '\n'?)) | (&[\n] '\n')) */
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
		/* 204 Sp <- Spacechar* */
		func() bool {
		l1059:
			if !p.rules[ruleSpacechar]() {
				goto l1060
			}
			goto l1059
		l1060:
			return true
		},
		/* 205 Spnl <- (Sp (Newline Sp)?) */
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
		/* 206 SpecialChar <- ((&[\\] '\\') | (&[#] '#') | (&[!] '!') | (&[<] '<') | (&[\]] ']') | (&[\[] '[') | (&[&] '&') | (&[`] '`') | (&[_] '_') | (&[*] '*') | (&[\"\'\-.^] ExtendedSpecialChar)) */
		func() bool {
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
			return true
		l1064:
			return false
		},
		/* 207 NormalChar <- (!((&[\n\r] Newline) | (&[\t ] Spacechar) | (&[!-#&\'*\-.<\[-`] SpecialChar)) .) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1067
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l1067
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l1067
					}
					break
				default:
					if !p.rules[ruleSpecialChar]() {
						goto l1067
					}
				}
			}
			goto l1066
		l1067:
			if !matchDot() {
				goto l1066
			}
			return true
		l1066:
			position = position0
			return false
		},
		/* 208 NonAlphanumeric <- [\000-\057\072-\100\133-\140\173-\177] */
		func() bool {
			if !matchClass(4) {
				goto l1069
			}
			return true
		l1069:
			return false
		},
		/* 209 Alphanumeric <- ((&[\377] '\377') | (&[\376] '\376') | (&[\375] '\375') | (&[\374] '\374') | (&[\373] '\373') | (&[\372] '\372') | (&[\371] '\371') | (&[\370] '\370') | (&[\367] '\367') | (&[\366] '\366') | (&[\365] '\365') | (&[\364] '\364') | (&[\363] '\363') | (&[\362] '\362') | (&[\361] '\361') | (&[\360] '\360') | (&[\357] '\357') | (&[\356] '\356') | (&[\355] '\355') | (&[\354] '\354') | (&[\353] '\353') | (&[\352] '\352') | (&[\351] '\351') | (&[\350] '\350') | (&[\347] '\347') | (&[\346] '\346') | (&[\345] '\345') | (&[\344] '\344') | (&[\343] '\343') | (&[\342] '\342') | (&[\341] '\341') | (&[\340] '\340') | (&[\337] '\337') | (&[\336] '\336') | (&[\335] '\335') | (&[\334] '\334') | (&[\333] '\333') | (&[\332] '\332') | (&[\331] '\331') | (&[\330] '\330') | (&[\327] '\327') | (&[\326] '\326') | (&[\325] '\325') | (&[\324] '\324') | (&[\323] '\323') | (&[\322] '\322') | (&[\321] '\321') | (&[\320] '\320') | (&[\317] '\317') | (&[\316] '\316') | (&[\315] '\315') | (&[\314] '\314') | (&[\313] '\313') | (&[\312] '\312') | (&[\311] '\311') | (&[\310] '\310') | (&[\307] '\307') | (&[\306] '\306') | (&[\305] '\305') | (&[\304] '\304') | (&[\303] '\303') | (&[\302] '\302') | (&[\301] '\301') | (&[\300] '\300') | (&[\277] '\277') | (&[\276] '\276') | (&[\275] '\275') | (&[\274] '\274') | (&[\273] '\273') | (&[\272] '\272') | (&[\271] '\271') | (&[\270] '\270') | (&[\267] '\267') | (&[\266] '\266') | (&[\265] '\265') | (&[\264] '\264') | (&[\263] '\263') | (&[\262] '\262') | (&[\261] '\261') | (&[\260] '\260') | (&[\257] '\257') | (&[\256] '\256') | (&[\255] '\255') | (&[\254] '\254') | (&[\253] '\253') | (&[\252] '\252') | (&[\251] '\251') | (&[\250] '\250') | (&[\247] '\247') | (&[\246] '\246') | (&[\245] '\245') | (&[\244] '\244') | (&[\243] '\243') | (&[\242] '\242') | (&[\241] '\241') | (&[\240] '\240') | (&[\237] '\237') | (&[\236] '\236') | (&[\235] '\235') | (&[\234] '\234') | (&[\233] '\233') | (&[\232] '\232') | (&[\231] '\231') | (&[\230] '\230') | (&[\227] '\227') | (&[\226] '\226') | (&[\225] '\225') | (&[\224] '\224') | (&[\223] '\223') | (&[\222] '\222') | (&[\221] '\221') | (&[\220] '\220') | (&[\217] '\217') | (&[\216] '\216') | (&[\215] '\215') | (&[\214] '\214') | (&[\213] '\213') | (&[\212] '\212') | (&[\211] '\211') | (&[\210] '\210') | (&[\207] '\207') | (&[\206] '\206') | (&[\205] '\205') | (&[\204] '\204') | (&[\203] '\203') | (&[\202] '\202') | (&[\201] '\201') | (&[\200] '\200') | (&[0-9A-Za-z] [0-9A-Za-z])) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1070
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
						goto l1070
					}
				}
			}
			return true
		l1070:
			return false
		},
		/* 210 AlphanumericAscii <- [A-Za-z0-9] */
		func() bool {
			if !matchClass(6) {
				goto l1072
			}
			return true
		l1072:
			return false
		},
		/* 211 Digit <- [0-9] */
		func() bool {
			if !matchClass(0) {
				goto l1073
			}
			return true
		l1073:
			return false
		},
		/* 212 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1074
			}
			if !matchChar('#') {
				goto l1074
			}
			if !matchClass(7) {
				goto l1074
			}
			if !matchClass(8) {
				goto l1074
			}
		l1075:
			if !matchClass(8) {
				goto l1076
			}
			goto l1075
		l1076:
			if !matchChar(';') {
				goto l1074
			}
			end = position
			return true
		l1074:
			position = position0
			return false
		},
		/* 213 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1077
			}
			if !matchChar('#') {
				goto l1077
			}
			if !matchClass(0) {
				goto l1077
			}
		l1078:
			if !matchClass(0) {
				goto l1079
			}
			goto l1078
		l1079:
			end = position
			if !matchChar(';') {
				goto l1077
			}
			end = position
			return true
		l1077:
			position = position0
			return false
		},
		/* 214 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1080
			}
			if !matchClass(6) {
				goto l1080
			}
		l1081:
			if !matchClass(6) {
				goto l1082
			}
			goto l1081
		l1082:
			if !matchChar(';') {
				goto l1080
			}
			end = position
			return true
		l1080:
			position = position0
			return false
		},
		/* 215 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() bool {
			if !matchString("   ") {
				goto l1085
			}
			goto l1084
		l1085:
			if !matchString("  ") {
				goto l1086
			}
			goto l1084
		l1086:
			if !matchChar(' ') {
				goto l1087
			}
			goto l1084
		l1087:
		l1084:
			return true
		},
		/* 216 Indent <- ((&[ ] '    ') | (&[\t] '\t')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1088
				}
				switch p.Buffer[position] {
				case ' ':
					position++
					if !matchString("   ") {
						goto l1088
					}
					break
				case '\t':
					position++ // matchChar
					break
				default:
					goto l1088
				}
			}
			return true
		l1088:
			return false
		},
		/* 217 IndentedLine <- (Indent Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1090
			}
			if !p.rules[ruleLine]() {
				goto l1090
			}
			return true
		l1090:
			position = position0
			return false
		},
		/* 218 OptionallyIndentedLine <- (Indent? Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1092
			}
		l1092:
			if !p.rules[ruleLine]() {
				goto l1091
			}
			return true
		l1091:
			position = position0
			return false
		},
		/* 219 StartList <- (&. { yy = nil }) */
		func() bool {
			if !(position < len(p.Buffer)) {
				goto l1094
			}
			do(87)
			return true
		l1094:
			return false
		},
		/* 220 Line <- (RawLine { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRawLine]() {
				goto l1095
			}
			do(88)
			return true
		l1095:
			position = position0
			return false
		},
		/* 221 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > !.)) */
		func() bool {
			position0 := position
			{
				position1097 := position
				begin = position
			l1099:
				if position == len(p.Buffer) {
					goto l1100
				}
				switch p.Buffer[position] {
				case '\r', '\n':
					goto l1100
				default:
					position++
				}
				goto l1099
			l1100:
				if !p.rules[ruleNewline]() {
					goto l1098
				}
				end = position
				goto l1097
			l1098:
				position = position1097
				begin = position
				if !matchDot() {
					goto l1096
				}
			l1101:
				if !matchDot() {
					goto l1102
				}
				goto l1101
			l1102:
				end = position
				if (position < len(p.Buffer)) {
					goto l1096
				}
			}
		l1097:
			return true
		l1096:
			position = position0
			return false
		},
		/* 222 SkipBlock <- (((!BlankLine RawLine)+ BlankLine*) / BlankLine+) */
		func() bool {
			position0 := position
			{
				position1104 := position
				if !p.rules[ruleBlankLine]() {
					goto l1108
				}
				goto l1105
			l1108:
				if !p.rules[ruleRawLine]() {
					goto l1105
				}
			l1106:
				{
					position1107 := position
					if !p.rules[ruleBlankLine]() {
						goto l1109
					}
					goto l1107
				l1109:
					if !p.rules[ruleRawLine]() {
						goto l1107
					}
					goto l1106
				l1107:
					position = position1107
				}
			l1110:
				if !p.rules[ruleBlankLine]() {
					goto l1111
				}
				goto l1110
			l1111:
				goto l1104
			l1105:
				position = position1104
				if !p.rules[ruleBlankLine]() {
					goto l1103
				}
			l1112:
				if !p.rules[ruleBlankLine]() {
					goto l1113
				}
				goto l1112
			l1113:
			}
		l1104:
			return true
		l1103:
			position = position0
			return false
		},
		/* 223 ExtendedSpecialChar <- ((&[^] (&{p.extension.Notes} '^')) | (&[\"\'\-.] (&{p.extension.Smart} ((&[\"] '"') | (&[\'] '\'') | (&[\-] '-') | (&[.] '.'))))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1114
				}
				switch p.Buffer[position] {
				case '^':
					if !(p.extension.Notes) {
						goto l1114
					}
					if !matchChar('^') {
						goto l1114
					}
					break
				default:
					if !(p.extension.Smart) {
						goto l1114
					}
					{
						if position == len(p.Buffer) {
							goto l1114
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
							goto l1114
						}
					}
				}
			}
			return true
		l1114:
			position = position0
			return false
		},
		/* 224 Smart <- (&{p.extension.Smart} (SingleQuoted / ((&[\'] Apostrophe) | (&[\"] DoubleQuoted) | (&[\-] Dash) | (&[.] Ellipsis)))) */
		func() bool {
			if !(p.extension.Smart) {
				goto l1117
			}
			if !p.rules[ruleSingleQuoted]() {
				goto l1119
			}
			goto l1118
		l1119:
			{
				if position == len(p.Buffer) {
					goto l1117
				}
				switch p.Buffer[position] {
				case '\'':
					if !p.rules[ruleApostrophe]() {
						goto l1117
					}
					break
				case '"':
					if !p.rules[ruleDoubleQuoted]() {
						goto l1117
					}
					break
				case '-':
					if !p.rules[ruleDash]() {
						goto l1117
					}
					break
				case '.':
					if !p.rules[ruleEllipsis]() {
						goto l1117
					}
					break
				default:
					goto l1117
				}
			}
		l1118:
			return true
		l1117:
			return false
		},
		/* 225 Apostrophe <- ('\'' { yy = mk_element(APOSTROPHE) }) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1121
			}
			do(89)
			return true
		l1121:
			position = position0
			return false
		},
		/* 226 Ellipsis <- (('...' / '. . .') { yy = mk_element(ELLIPSIS) }) */
		func() bool {
			position0 := position
			if !matchString("...") {
				goto l1124
			}
			goto l1123
		l1124:
			if !matchString(". . .") {
				goto l1122
			}
		l1123:
			do(90)
			return true
		l1122:
			position = position0
			return false
		},
		/* 227 Dash <- (EmDash / EnDash) */
		func() bool {
			if !p.rules[ruleEmDash]() {
				goto l1127
			}
			goto l1126
		l1127:
			if !p.rules[ruleEnDash]() {
				goto l1125
			}
		l1126:
			return true
		l1125:
			return false
		},
		/* 228 EnDash <- ('-' &[0-9] { yy = mk_element(ENDASH) }) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l1128
			}
			if !peekClass(0) {
				goto l1128
			}
			do(91)
			return true
		l1128:
			position = position0
			return false
		},
		/* 229 EmDash <- (('---' / '--') { yy = mk_element(EMDASH) }) */
		func() bool {
			position0 := position
			if !matchString("---") {
				goto l1131
			}
			goto l1130
		l1131:
			if !matchString("--") {
				goto l1129
			}
		l1130:
			do(92)
			return true
		l1129:
			position = position0
			return false
		},
		/* 230 SingleQuoteStart <- ('\'' ![)!\],.;:-? \t\n] !(((&[r] 're') | (&[l] 'll') | (&[v] 've') | (&[m] 'm') | (&[t] 't') | (&[s] 's')) !Alphanumeric)) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1132
			}
			if peekClass(9) {
				goto l1132
			}
			{
				position1133 := position
				{
					if position == len(p.Buffer) {
						goto l1133
					}
					switch p.Buffer[position] {
					case 'r':
						position++ // matchString(`re`)
						if !matchChar('e') {
							goto l1133
						}
						break
					case 'l':
						position++ // matchString(`ll`)
						if !matchChar('l') {
							goto l1133
						}
						break
					case 'v':
						position++ // matchString(`ve`)
						if !matchChar('e') {
							goto l1133
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
						goto l1133
					}
				}
				if !p.rules[ruleAlphanumeric]() {
					goto l1135
				}
				goto l1133
			l1135:
				goto l1132
			l1133:
				position = position1133
			}
			return true
		l1132:
			position = position0
			return false
		},
		/* 231 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1136
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l1137
			}
			goto l1136
		l1137:
			return true
		l1136:
			position = position0
			return false
		},
		/* 232 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = mk_list(SINGLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto l1138
			}
			if !p.rules[ruleStartList]() {
				goto l1138
			}
			doarg(yySet, -1)
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1141
			}
			goto l1138
		l1141:
			if !p.rules[ruleInline]() {
				goto l1138
			}
			doarg(yySet, -2)
			do(93)
		l1139:
			{
				position1140, thunkPosition1140 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto l1142
				}
				goto l1140
			l1142:
				if !p.rules[ruleInline]() {
					goto l1140
				}
				doarg(yySet, -2)
				do(93)
				goto l1139
			l1140:
				position, thunkPosition = position1140, thunkPosition1140
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1138
			}
			do(94)
			doarg(yyPop, 2)
			return true
		l1138:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 233 DoubleQuoteStart <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1143
			}
			return true
		l1143:
			return false
		},
		/* 234 DoubleQuoteEnd <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1144
			}
			return true
		l1144:
			return false
		},
		/* 235 DoubleQuoted <- ('"' StartList (!'"' Inline { a = cons(b, a) })+ '"' { yy = mk_list(DOUBLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('"') {
				goto l1145
			}
			if !p.rules[ruleStartList]() {
				goto l1145
			}
			doarg(yySet, -1)
			if peekChar('"') {
				goto l1145
			}
			if !p.rules[ruleInline]() {
				goto l1145
			}
			doarg(yySet, -2)
			do(95)
		l1146:
			{
				position1147, thunkPosition1147 := position, thunkPosition
				if peekChar('"') {
					goto l1147
				}
				if !p.rules[ruleInline]() {
					goto l1147
				}
				doarg(yySet, -2)
				do(95)
				goto l1146
			l1147:
				position, thunkPosition = position1147, thunkPosition1147
			}
			if !matchChar('"') {
				goto l1145
			}
			do(96)
			doarg(yyPop, 2)
			return true
		l1145:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 236 NoteReference <- (&{p.extension.Notes} RawNoteReference {
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
				goto l1148
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1148
			}
			doarg(yySet, -1)
			do(97)
			doarg(yyPop, 1)
			return true
		l1148:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 237 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = mk_str(yytext) }) */
		func() bool {
			position0 := position
			if !matchString("[^") {
				goto l1149
			}
			begin = position
			if !p.rules[ruleNewline]() {
				goto l1152
			}
			goto l1149
		l1152:
			if peekChar(']') {
				goto l1149
			}
			if !matchDot() {
				goto l1149
			}
		l1150:
			{
				position1151 := position
				if !p.rules[ruleNewline]() {
					goto l1153
				}
				goto l1151
			l1153:
				if peekChar(']') {
					goto l1151
				}
				if !matchDot() {
					goto l1151
				}
				goto l1150
			l1151:
				position = position1151
			}
			end = position
			if !matchChar(']') {
				goto l1149
			}
			do(98)
			return true
		l1149:
			position = position0
			return false
		},
		/* 238 Note <- (&{p.extension.Notes} NonindentSpace RawNoteReference ':' Sp StartList (RawNoteBlock { a = cons(yy, a) }) (&Indent RawNoteBlock { a = cons(yy, a) })* {   yy = mk_list(NOTE, a)
                    yy.contents.str = ref.contents.str
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !(p.extension.Notes) {
				goto l1154
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l1154
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1154
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto l1154
			}
			if !p.rules[ruleSp]() {
				goto l1154
			}
			if !p.rules[ruleStartList]() {
				goto l1154
			}
			doarg(yySet, -2)
			if !p.rules[ruleRawNoteBlock]() {
				goto l1154
			}
			do(99)
		l1155:
			{
				position1156, thunkPosition1156 := position, thunkPosition
				{
					position1157 := position
					if !p.rules[ruleIndent]() {
						goto l1156
					}
					position = position1157
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto l1156
				}
				do(100)
				goto l1155
			l1156:
				position, thunkPosition = position1156, thunkPosition1156
			}
			do(101)
			doarg(yyPop, 2)
			return true
		l1154:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 239 InlineNote <- (&{p.extension.Notes} '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = mk_list(NOTE, a)
                  yy.contents.str = "" }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto l1158
			}
			if !matchString("^[") {
				goto l1158
			}
			if !p.rules[ruleStartList]() {
				goto l1158
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto l1158
			}
			if !p.rules[ruleInline]() {
				goto l1158
			}
			do(102)
		l1159:
			{
				position1160 := position
				if peekChar(']') {
					goto l1160
				}
				if !p.rules[ruleInline]() {
					goto l1160
				}
				do(102)
				goto l1159
			l1160:
				position = position1160
			}
			if !matchChar(']') {
				goto l1158
			}
			do(103)
			doarg(yyPop, 1)
			return true
		l1158:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 240 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1161
			}
			doarg(yySet, -1)
		l1162:
			{
				position1163, thunkPosition1163 := position, thunkPosition
				{
					position1164, thunkPosition1164 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto l1165
					}
					doarg(yySet, -2)
					do(104)
					goto l1164
				l1165:
					position, thunkPosition = position1164, thunkPosition1164
					if !p.rules[ruleSkipBlock]() {
						goto l1163
					}
				}
			l1164:
				goto l1162
			l1163:
				position, thunkPosition = position1163, thunkPosition1163
			}
			do(105)
			if !(commit(thunkPosition0)) {
				goto l1161
			}
			doarg(yyPop, 2)
			return true
		l1161:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 241 RawNoteBlock <- (StartList (!BlankLine OptionallyIndentedLine { a = cons(yy, a) })+ (< BlankLine* > { a = cons(mk_str(yytext), a) }) {   yy = mk_str_from_list(a, true)
                    yy.key = RAW
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1166
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l1169
			}
			goto l1166
		l1169:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l1166
			}
			do(106)
		l1167:
			{
				position1168 := position
				if !p.rules[ruleBlankLine]() {
					goto l1170
				}
				goto l1168
			l1170:
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto l1168
				}
				do(106)
				goto l1167
			l1168:
				position = position1168
			}
			begin = position
		l1171:
			if !p.rules[ruleBlankLine]() {
				goto l1172
			}
			goto l1171
		l1172:
			end = position
			do(107)
			do(108)
			doarg(yyPop, 1)
			return true
		l1166:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 242 DefinitionList <- (&{p.extension.Dlists} StartList (Definition { a = cons(yy, a) })+ { yy = mk_list(DEFINITIONLIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Dlists) {
				goto l1173
			}
			if !p.rules[ruleStartList]() {
				goto l1173
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto l1173
			}
			do(109)
		l1174:
			{
				position1175, thunkPosition1175 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto l1175
				}
				do(109)
				goto l1174
			l1175:
				position, thunkPosition = position1175, thunkPosition1175
			}
			do(110)
			doarg(yyPop, 1)
			return true
		l1173:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 243 Definition <- (&((!Defmark RawLine)+ BlankLine? Defmark) StartList (DListTitle { a = cons(yy, a) })+ (DefTight / DefLoose) {
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			} { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1177 := position
				if !p.rules[ruleDefmark]() {
					goto l1180
				}
				goto l1176
			l1180:
				if !p.rules[ruleRawLine]() {
					goto l1176
				}
			l1178:
				{
					position1179 := position
					if !p.rules[ruleDefmark]() {
						goto l1181
					}
					goto l1179
				l1181:
					if !p.rules[ruleRawLine]() {
						goto l1179
					}
					goto l1178
				l1179:
					position = position1179
				}
				if !p.rules[ruleBlankLine]() {
					goto l1182
				}
			l1182:
				if !p.rules[ruleDefmark]() {
					goto l1176
				}
				position = position1177
			}
			if !p.rules[ruleStartList]() {
				goto l1176
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto l1176
			}
			do(111)
		l1184:
			{
				position1185, thunkPosition1185 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto l1185
				}
				do(111)
				goto l1184
			l1185:
				position, thunkPosition = position1185, thunkPosition1185
			}
			if !p.rules[ruleDefTight]() {
				goto l1187
			}
			goto l1186
		l1187:
			if !p.rules[ruleDefLoose]() {
				goto l1176
			}
		l1186:
			do(112)
			do(113)
			doarg(yyPop, 1)
			return true
		l1176:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 244 DListTitle <- (NonindentSpace !Defmark &Nonspacechar StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline {	yy = mk_list(LIST, a)
				yy.key = DEFTITLE
			}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l1188
			}
			if !p.rules[ruleDefmark]() {
				goto l1189
			}
			goto l1188
		l1189:
			{
				position1190 := position
				if !p.rules[ruleNonspacechar]() {
					goto l1188
				}
				position = position1190
			}
			if !p.rules[ruleStartList]() {
				goto l1188
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l1193
			}
			goto l1188
		l1193:
			if !p.rules[ruleInline]() {
				goto l1188
			}
			do(114)
		l1191:
			{
				position1192 := position
				if !p.rules[ruleEndline]() {
					goto l1194
				}
				goto l1192
			l1194:
				if !p.rules[ruleInline]() {
					goto l1192
				}
				do(114)
				goto l1191
			l1192:
				position = position1192
			}
			if !p.rules[ruleSp]() {
				goto l1188
			}
			if !p.rules[ruleNewline]() {
				goto l1188
			}
			do(115)
			doarg(yyPop, 1)
			return true
		l1188:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 245 DefTight <- (&Defmark ListTight) */
		func() bool {
			{
				position1196 := position
				if !p.rules[ruleDefmark]() {
					goto l1195
				}
				position = position1196
			}
			if !p.rules[ruleListTight]() {
				goto l1195
			}
			return true
		l1195:
			return false
		},
		/* 246 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l1197
			}
			{
				position1198 := position
				if !p.rules[ruleDefmark]() {
					goto l1197
				}
				position = position1198
			}
			if !p.rules[ruleListLoose]() {
				goto l1197
			}
			return true
		l1197:
			position = position0
			return false
		},
		/* 247 Defmark <- (NonindentSpace ((&[~] '~') | (&[:] ':')) Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l1199
			}
			{
				if position == len(p.Buffer) {
					goto l1199
				}
				switch p.Buffer[position] {
				case '~':
					position++ // matchChar
					break
				case ':':
					position++ // matchChar
					break
				default:
					goto l1199
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l1199
			}
		l1201:
			if !p.rules[ruleSpacechar]() {
				goto l1202
			}
			goto l1201
		l1202:
			return true
		l1199:
			position = position0
			return false
		},
		/* 248 DefMarker <- (&{p.extension.Dlists} Defmark) */
		func() bool {
			if !(p.extension.Dlists) {
				goto l1203
			}
			if !p.rules[ruleDefmark]() {
				goto l1203
			}
			return true
		l1203:
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
func (d *Doc) findReference(label *element) (*link, bool) {
	for cur := d.references; cur != nil; cur = cur.next {
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
func (d *Doc) find_note(label string) (*element, bool) {
	for el := d.notes; el != nil; el = el.next {
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
