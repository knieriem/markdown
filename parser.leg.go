
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
	extension	Extensions

	tree				*element	/* Results of parse. */
	references			*element	/* List of link references found. */
	notes				*element	/* List of footnotes found. */
}


const (
	ruleDoc	= iota
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
	ruleListBlock
	ruleListContinuationBlock
	ruleEnumerator
	ruleOrderedList
	ruleListBlockLine
	ruleHtmlBlockOpenAddress
	ruleHtmlBlockCloseAddress
	ruleHtmlBlockOpenBlockquote
	ruleHtmlBlockCloseBlockquote
	ruleHtmlBlockOpenCenter
	ruleHtmlBlockCloseCenter
	ruleHtmlBlockOpenDir
	ruleHtmlBlockCloseDir
	ruleHtmlBlockOpenDiv
	ruleHtmlBlockCloseDiv
	ruleHtmlBlockOpenDl
	ruleHtmlBlockCloseDl
	ruleHtmlBlockOpenFieldset
	ruleHtmlBlockCloseFieldset
	ruleHtmlBlockOpenForm
	ruleHtmlBlockCloseForm
	ruleHtmlBlockOpenH1
	ruleHtmlBlockCloseH1
	ruleHtmlBlockOpenH2
	ruleHtmlBlockCloseH2
	ruleHtmlBlockOpenH3
	ruleHtmlBlockCloseH3
	ruleHtmlBlockOpenH4
	ruleHtmlBlockCloseH4
	ruleHtmlBlockOpenH5
	ruleHtmlBlockCloseH5
	ruleHtmlBlockOpenH6
	ruleHtmlBlockCloseH6
	ruleHtmlBlockOpenMenu
	ruleHtmlBlockCloseMenu
	ruleHtmlBlockOpenNoframes
	ruleHtmlBlockCloseNoframes
	ruleHtmlBlockOpenNoscript
	ruleHtmlBlockCloseNoscript
	ruleHtmlBlockOpenOl
	ruleHtmlBlockCloseOl
	ruleHtmlBlockOpenP
	ruleHtmlBlockCloseP
	ruleHtmlBlockOpenPre
	ruleHtmlBlockClosePre
	ruleHtmlBlockOpenTable
	ruleHtmlBlockCloseTable
	ruleHtmlBlockOpenUl
	ruleHtmlBlockCloseUl
	ruleHtmlBlockOpenDd
	ruleHtmlBlockCloseDd
	ruleHtmlBlockOpenDt
	ruleHtmlBlockCloseDt
	ruleHtmlBlockOpenFrameset
	ruleHtmlBlockCloseFrameset
	ruleHtmlBlockOpenLi
	ruleHtmlBlockCloseLi
	ruleHtmlBlockOpenTbody
	ruleHtmlBlockCloseTbody
	ruleHtmlBlockOpenTd
	ruleHtmlBlockCloseTd
	ruleHtmlBlockOpenTfoot
	ruleHtmlBlockCloseTfoot
	ruleHtmlBlockOpenTh
	ruleHtmlBlockCloseTh
	ruleHtmlBlockOpenThead
	ruleHtmlBlockCloseThead
	ruleHtmlBlockOpenTr
	ruleHtmlBlockCloseTr
	ruleHtmlBlockOpenScript
	ruleHtmlBlockCloseScript
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
	ruleAlphanumeric
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

type yyParser struct {*Doc
	Buffer string
	Min, Max int
	rules [213]func() bool
	ResetBuffer	func(string) string
}

func (p *yyParser) Parse(ruleId int) bool {
	if p.rules[ruleId]() {
		return true
	}
	return false
}
func (p *yyParser) PrintError() {
	line := 1
	character := 0
	for i, c := range p.Buffer[0:] {
		if c == '\n' {
			line++
			character = 0
		} else {
			character++
		}
		if i == p.Min {
			if p.Min != p.Max {
				fmt.Printf("parse error after line %v character %v\n", line, character)
			} else {
				break
			}
		} else if i == p.Max {
			break
		}
	}
	fmt.Printf("parse error: unexpected ")
	if p.Max >= len(p.Buffer) {
		fmt.Printf("end of file found\n")
	} else {
		fmt.Printf("'%c' at line %v character %v\n", p.Buffer[p.Max], line, character)
	}
}
func (p *yyParser) Init() {
	var position int
	var yyp int
	var yy *element
	var yyval = make([]*element, 200)

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
			yyval[yyp-1] = s
			yyval[yyp-2] = a
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
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			
                  li := b.children
                  li.contents.str += "\n\n"
                  a = cons(b, a)
              
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 26 ListLoose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
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
		/* 30 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 31 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 32 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_str_from_list(a, false) 
			yyval[yyp-1] = a
		},
		/* 33 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   if len(yytext) == 0 {
                                   a = cons(mk_str("\001"), a) // block separator
                              } else {
                                   a = cons(mk_str(yytext), a)
                              }
                          
			yyval[yyp-1] = a
		},
		/* 34 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 35 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			  yy = mk_str_from_list(a, false) 
			yyval[yyp-1] = a
		},
		/* 36 OrderedList */
		func(yytext string, _ int) {
			 yy.key = ORDEREDLIST 
		},
		/* 37 HtmlBlock */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTMLBLOCK
                }
            
		},
		/* 38 StyleBlock */
		func(yytext string, _ int) {
			   if p.extension.FilterStyles {
                        yy = mk_list(LIST, nil)
                    } else {
                        yy = mk_str(yytext)
                        yy.key = HTMLBLOCK
                    }
                
		},
		/* 39 Inlines */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
		},
		/* 40 Inlines */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(c, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
		},
		/* 41 Inlines */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
		},
		/* 42 Space */
		func(yytext string, _ int) {
			 yy = mk_str(" ")
          yy.key = SPACE 
		},
		/* 43 Str */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 44 EscapedChar */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 45 Entity */
		func(yytext string, _ int) {
			 yy = mk_str(yytext); yy.key = HTML 
		},
		/* 46 NormalEndline */
		func(yytext string, _ int) {
			 yy = mk_str("\n")
                    yy.key = SPACE 
		},
		/* 47 TerminalEndline */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 48 LineBreak */
		func(yytext string, _ int) {
			 yy = mk_element(LINEBREAK) 
		},
		/* 49 Symbol */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 50 UlOrStarLine */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 51 OneStarClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 52 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 53 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 54 EmphStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(EMPH, a) 
			yyval[yyp-1] = a
		},
		/* 55 OneUlClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 56 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 57 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 58 EmphUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(EMPH, a) 
			yyval[yyp-1] = a
		},
		/* 59 TwoStarClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 60 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 61 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 62 StrongStar */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(STRONG, a) 
			yyval[yyp-1] = a
		},
		/* 63 TwoUlClose */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a 
			yyval[yyp-1] = a
		},
		/* 64 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 65 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 66 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(STRONG, a) 
			yyval[yyp-1] = a
		},
		/* 67 Image */
		func(yytext string, _ int) {
			 yy.key = IMAGE 
		},
		/* 68 ReferenceLinkDouble */
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
		/* 69 ReferenceLinkSingle */
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
		/* 70 ExplicitLink */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			l := yyval[yyp-2]
			t := yyval[yyp-3]
			 yy = mk_link(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil 
			yyval[yyp-1] = s
			yyval[yyp-2] = l
			yyval[yyp-3] = t
		},
		/* 71 Source */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 72 Title */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 73 AutoLinkUrl */
		func(yytext string, _ int) {
			   yy = mk_link(mk_str(yytext), yytext, "") 
		},
		/* 74 AutoLinkEmail */
		func(yytext string, _ int) {
			
                    yy = mk_link(mk_str(yytext), "mailto:"+yytext, "")
                
		},
		/* 75 Reference */
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
		/* 76 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 77 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 78 RefSrc */
		func(yytext string, _ int) {
			 yy = mk_str(yytext)
           yy.key = HTML 
		},
		/* 79 RefTitle */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 80 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 81 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 p.references = reverse(a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 82 Code */
		func(yytext string, _ int) {
			 yy = mk_str(yytext); yy.key = CODE 
		},
		/* 83 RawHtml */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTML
                }
            
		},
		/* 84 StartList */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 85 Line */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 86 Apostrophe */
		func(yytext string, _ int) {
			 yy = mk_element(APOSTROPHE) 
		},
		/* 87 Ellipsis */
		func(yytext string, _ int) {
			 yy = mk_element(ELLIPSIS) 
		},
		/* 88 EnDash */
		func(yytext string, _ int) {
			 yy = mk_element(ENDASH) 
		},
		/* 89 EmDash */
		func(yytext string, _ int) {
			 yy = mk_element(EMDASH) 
		},
		/* 90 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 91 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 yy = mk_list(SINGLEQUOTED, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 92 DoubleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 93 DoubleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 yy = mk_list(DOUBLEQUOTED, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 94 NoteReference */
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
		/* 95 RawNoteReference */
		func(yytext string, _ int) {
			 yy = mk_str(yytext) 
		},
		/* 96 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 97 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 98 Note */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			a := yyval[yyp-2]
			   yy = mk_list(NOTE, a)
                    yy.contents.str = ref.contents.str
                
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
		},
		/* 99 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 100 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(NOTE, a)
                  yy.contents.str = "" 
			yyval[yyp-1] = a
		},
		/* 101 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 102 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 p.notes = reverse(a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 103 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 104 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(mk_str(yytext), a) 
			yyval[yyp-1] = a
		},
		/* 105 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   yy = mk_str_from_list(a, true)
                    yy.key = RAW
                
			yyval[yyp-1] = a
		},
		/* 106 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 107 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(DEFINITIONLIST, a) 
			yyval[yyp-1] = a
		},
		/* 108 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 109 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			
			yyval[yyp-1] = a
		},
		/* 110 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 111 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 112 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
				yy = mk_list(LIST, a)
				yy.key = DEFTITLE
			
			yyval[yyp-1] = a
		},
		/* 113 yyPush */
		func(_ string, count int) {
			yyp += count
			if yyp >= len(yyval) {
				s := make([]*element, cap(yyval)+200)
				copy(s, yyval)
				yyval = s
			}
		},
		/* 114 yyPop */
		func(_ string, count int) {
			yyp -= count
		},
		/* 115 yySet */
		func(_ string, count int) {
			yyval[yyp+count] = yy
		},
	}
	const (
		yyPush = 113+iota
		yyPop
		yySet
	)

	var thunkPosition, begin, end int
	thunks := make([]struct {action uint8; begin, end int}, 32)
	doarg := func(action uint8, arg int) {
		if thunkPosition == len(thunks) {
			newThunks := make([]struct {action uint8; begin, end int}, 2 * len(thunks))
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
			old = p.Buffer[p.Max:]
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
				if b>=0 && e<=len(p.Buffer) && b<=e {
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
	peekDot := func() bool {
		return position < len(p.Buffer)
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
		if (next <= len(p.Buffer)) && (p.Buffer[position:next] == s) {
			position = next
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	classes := [...][32]uint8{
		{0, 0, 0, 0, 0, 0, 255, 3, 126, 0, 0, 0, 126, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 10, 111, 0, 80, 0, 0, 0, 184, 1, 0, 0, 56, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 6, 0, 0, 3, 82, 0, 252, 0, 0, 0, 32, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 255, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 255, 3, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 40, 255, 3, 254, 255, 255, 135, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
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
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleBlock]() {
					goto l2
				}
				do(0)
				goto l1
			l2:
				position, thunkPosition = position2, thunkPosition2
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
			position0, thunkPosition0 := position, thunkPosition
		l4:
			{
				position5, thunkPosition5 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l5
				}
				goto l4
			l5:
				position, thunkPosition = position5, thunkPosition5
			}
			{
				position6, thunkPosition6 := position, thunkPosition
				if !p.rules[ruleBlockQuote]() {
					goto l7
				}
				goto l6
			l7:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleVerbatim]() {
					goto l8
				}
				goto l6
			l8:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleNote]() {
					goto l9
				}
				goto l6
			l9:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleReference]() {
					goto l10
				}
				goto l6
			l10:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleHorizontalRule]() {
					goto l11
				}
				goto l6
			l11:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleHeading]() {
					goto l12
				}
				goto l6
			l12:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleDefinitionList]() {
					goto l13
				}
				goto l6
			l13:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleOrderedList]() {
					goto l14
				}
				goto l6
			l14:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleBulletList]() {
					goto l15
				}
				goto l6
			l15:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleHtmlBlock]() {
					goto l16
				}
				goto l6
			l16:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[ruleStyleBlock]() {
					goto l17
				}
				goto l6
			l17:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[rulePara]() {
					goto l18
				}
				goto l6
			l18:
				position, thunkPosition = position6, thunkPosition6
				if !p.rules[rulePlain]() {
					goto l3
				}
			}
		l6:
			return true
		l3:
			position, thunkPosition = position0, thunkPosition0
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
			{
				position21, thunkPosition21 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l21
				}
				goto l20
			l21:
				position, thunkPosition = position21, thunkPosition21
			}
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
		/* 4 AtxInline <- (!Newline !(Sp '#'* Sp Newline) Inline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position24, thunkPosition24 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l24
				}
				goto l23
			l24:
				position, thunkPosition = position24, thunkPosition24
			}
			{
				position25, thunkPosition25 := position, thunkPosition
				if !p.rules[ruleSp]() {
					goto l25
				}
			l26:
				{
					position27, thunkPosition27 := position, thunkPosition
					if !matchChar('#') {
						goto l27
					}
					goto l26
				l27:
					position, thunkPosition = position27, thunkPosition27
				}
				if !p.rules[ruleSp]() {
					goto l25
				}
				if !p.rules[ruleNewline]() {
					goto l25
				}
				goto l23
			l25:
				position, thunkPosition = position25, thunkPosition25
			}
			if !p.rules[ruleInline]() {
				goto l23
			}
			return true
		l23:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 5 AtxStart <- (&'#' < ('######' / '#####' / '####' / '###' / '##' / '#') > { yy = mk_element(H1 + (len(yytext) - 1)) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !peekChar('#') {
				goto l28
			}
			begin = position
			{
				position29, thunkPosition29 := position, thunkPosition
				if !matchString("######") {
					goto l30
				}
				goto l29
			l30:
				position, thunkPosition = position29, thunkPosition29
				if !matchString("#####") {
					goto l31
				}
				goto l29
			l31:
				position, thunkPosition = position29, thunkPosition29
				if !matchString("####") {
					goto l32
				}
				goto l29
			l32:
				position, thunkPosition = position29, thunkPosition29
				if !matchString("###") {
					goto l33
				}
				goto l29
			l33:
				position, thunkPosition = position29, thunkPosition29
				if !matchString("##") {
					goto l34
				}
				goto l29
			l34:
				position, thunkPosition = position29, thunkPosition29
				if !matchChar('#') {
					goto l28
				}
			}
		l29:
			end = position
			do(4)
			return true
		l28:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 6 AtxHeading <- (AtxStart Sp StartList (AtxInline { a = cons(yy, a) })+ (Sp '#'* Sp)? Newline { yy = mk_list(s.key, a)
              s = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleAtxStart]() {
				goto l35
			}
			doarg(yySet, -1)
			if !p.rules[ruleSp]() {
				goto l35
			}
			if !p.rules[ruleStartList]() {
				goto l35
			}
			doarg(yySet, -2)
			if !p.rules[ruleAtxInline]() {
				goto l35
			}
			do(5)
		l36:
			{
				position37, thunkPosition37 := position, thunkPosition
				if !p.rules[ruleAtxInline]() {
					goto l37
				}
				do(5)
				goto l36
			l37:
				position, thunkPosition = position37, thunkPosition37
			}
			{
				position38, thunkPosition38 := position, thunkPosition
				if !p.rules[ruleSp]() {
					goto l38
				}
			l40:
				{
					position41, thunkPosition41 := position, thunkPosition
					if !matchChar('#') {
						goto l41
					}
					goto l40
				l41:
					position, thunkPosition = position41, thunkPosition41
				}
				if !p.rules[ruleSp]() {
					goto l38
				}
				goto l39
			l38:
				position, thunkPosition = position38, thunkPosition38
			}
		l39:
			if !p.rules[ruleNewline]() {
				goto l35
			}
			do(6)
			doarg(yyPop, 2)
			return true
		l35:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 7 SetextHeading <- (SetextHeading1 / SetextHeading2) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position43, thunkPosition43 := position, thunkPosition
				if !p.rules[ruleSetextHeading1]() {
					goto l44
				}
				goto l43
			l44:
				position, thunkPosition = position43, thunkPosition43
				if !p.rules[ruleSetextHeading2]() {
					goto l42
				}
			}
		l43:
			return true
		l42:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 8 SetextBottom1 <- ('===' '='* Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("===") {
				goto l45
			}
		l46:
			{
				position47, thunkPosition47 := position, thunkPosition
				if !matchChar('=') {
					goto l47
				}
				goto l46
			l47:
				position, thunkPosition = position47, thunkPosition47
			}
			if !p.rules[ruleNewline]() {
				goto l45
			}
			return true
		l45:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 9 SetextBottom2 <- ('---' '-'* Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("---") {
				goto l48
			}
		l49:
			{
				position50, thunkPosition50 := position, thunkPosition
				if !matchChar('-') {
					goto l50
				}
				goto l49
			l50:
				position, thunkPosition = position50, thunkPosition50
			}
			if !p.rules[ruleNewline]() {
				goto l48
			}
			return true
		l48:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 10 SetextHeading1 <- (&(RawLine SetextBottom1) StartList (!Endline Inline { a = cons(yy, a) })+ Newline SetextBottom1 { yy = mk_list(H1, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position52, thunkPosition52 := position, thunkPosition
				if !p.rules[ruleRawLine]() {
					goto l51
				}
				if !p.rules[ruleSetextBottom1]() {
					goto l51
				}
				position, thunkPosition = position52, thunkPosition52
			}
			if !p.rules[ruleStartList]() {
				goto l51
			}
			doarg(yySet, -1)
			{
				position55, thunkPosition55 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l55
				}
				goto l51
			l55:
				position, thunkPosition = position55, thunkPosition55
			}
			if !p.rules[ruleInline]() {
				goto l51
			}
			do(7)
		l53:
			{
				position54, thunkPosition54 := position, thunkPosition
				{
					position56, thunkPosition56 := position, thunkPosition
					if !p.rules[ruleEndline]() {
						goto l56
					}
					goto l54
				l56:
					position, thunkPosition = position56, thunkPosition56
				}
				if !p.rules[ruleInline]() {
					goto l54
				}
				do(7)
				goto l53
			l54:
				position, thunkPosition = position54, thunkPosition54
			}
			if !p.rules[ruleNewline]() {
				goto l51
			}
			if !p.rules[ruleSetextBottom1]() {
				goto l51
			}
			do(8)
			doarg(yyPop, 1)
			return true
		l51:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 11 SetextHeading2 <- (&(RawLine SetextBottom2) StartList (!Endline Inline { a = cons(yy, a) })+ Newline SetextBottom2 { yy = mk_list(H2, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position58, thunkPosition58 := position, thunkPosition
				if !p.rules[ruleRawLine]() {
					goto l57
				}
				if !p.rules[ruleSetextBottom2]() {
					goto l57
				}
				position, thunkPosition = position58, thunkPosition58
			}
			if !p.rules[ruleStartList]() {
				goto l57
			}
			doarg(yySet, -1)
			{
				position61, thunkPosition61 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l61
				}
				goto l57
			l61:
				position, thunkPosition = position61, thunkPosition61
			}
			if !p.rules[ruleInline]() {
				goto l57
			}
			do(9)
		l59:
			{
				position60, thunkPosition60 := position, thunkPosition
				{
					position62, thunkPosition62 := position, thunkPosition
					if !p.rules[ruleEndline]() {
						goto l62
					}
					goto l60
				l62:
					position, thunkPosition = position62, thunkPosition62
				}
				if !p.rules[ruleInline]() {
					goto l60
				}
				do(9)
				goto l59
			l60:
				position, thunkPosition = position60, thunkPosition60
			}
			if !p.rules[ruleNewline]() {
				goto l57
			}
			if !p.rules[ruleSetextBottom2]() {
				goto l57
			}
			do(10)
			doarg(yyPop, 1)
			return true
		l57:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 12 Heading <- (AtxHeading / SetextHeading) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position64, thunkPosition64 := position, thunkPosition
				if !p.rules[ruleAtxHeading]() {
					goto l65
				}
				goto l64
			l65:
				position, thunkPosition = position64, thunkPosition64
				if !p.rules[ruleSetextHeading]() {
					goto l63
				}
			}
		l64:
			return true
		l63:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 13 BlockQuote <- (BlockQuoteRaw {  yy = mk_element(BLOCKQUOTE)
                yy.children = a
             }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleBlockQuoteRaw]() {
				goto l66
			}
			doarg(yySet, -1)
			do(11)
			doarg(yyPop, 1)
			return true
		l66:
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
				goto l67
			}
			doarg(yySet, -1)
			if !matchChar('>') {
				goto l67
			}
			{
				position70, thunkPosition70 := position, thunkPosition
				if !matchChar(' ') {
					goto l70
				}
				goto l71
			l70:
				position, thunkPosition = position70, thunkPosition70
			}
		l71:
			if !p.rules[ruleLine]() {
				goto l67
			}
			do(12)
		l72:
			{
				position73, thunkPosition73 := position, thunkPosition
				if peekChar('>') {
					goto l73
				}
				{
					position74, thunkPosition74 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l74
					}
					goto l73
				l74:
					position, thunkPosition = position74, thunkPosition74
				}
				if !p.rules[ruleLine]() {
					goto l73
				}
				do(13)
				goto l72
			l73:
				position, thunkPosition = position73, thunkPosition73
			}
		l75:
			{
				position76, thunkPosition76 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l76
				}
				do(14)
				goto l75
			l76:
				position, thunkPosition = position76, thunkPosition76
			}
		l68:
			{
				position69, thunkPosition69 := position, thunkPosition
				if !matchChar('>') {
					goto l69
				}
				{
					position77, thunkPosition77 := position, thunkPosition
					if !matchChar(' ') {
						goto l77
					}
					goto l78
				l77:
					position, thunkPosition = position77, thunkPosition77
				}
			l78:
				if !p.rules[ruleLine]() {
					goto l69
				}
				do(12)
			l79:
				{
					position80, thunkPosition80 := position, thunkPosition
					if peekChar('>') {
						goto l80
					}
					{
						position81, thunkPosition81 := position, thunkPosition
						if !p.rules[ruleBlankLine]() {
							goto l81
						}
						goto l80
					l81:
						position, thunkPosition = position81, thunkPosition81
					}
					if !p.rules[ruleLine]() {
						goto l80
					}
					do(13)
					goto l79
				l80:
					position, thunkPosition = position80, thunkPosition80
				}
			l82:
				{
					position83, thunkPosition83 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l83
					}
					do(14)
					goto l82
				l83:
					position, thunkPosition = position83, thunkPosition83
				}
				goto l68
			l69:
				position, thunkPosition = position69, thunkPosition69
			}
			do(15)
			doarg(yyPop, 1)
			return true
		l67:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 NonblankIndentedLine <- (!BlankLine IndentedLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position85, thunkPosition85 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l85
				}
				goto l84
			l85:
				position, thunkPosition = position85, thunkPosition85
			}
			if !p.rules[ruleIndentedLine]() {
				goto l84
			}
			return true
		l84:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 VerbatimChunk <- (StartList (BlankLine { a = cons(mk_str("\n"), a) })* (NonblankIndentedLine { a = cons(yy, a) })+ { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l86
			}
			doarg(yySet, -1)
		l87:
			{
				position88, thunkPosition88 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l88
				}
				do(16)
				goto l87
			l88:
				position, thunkPosition = position88, thunkPosition88
			}
			if !p.rules[ruleNonblankIndentedLine]() {
				goto l86
			}
			do(17)
		l89:
			{
				position90, thunkPosition90 := position, thunkPosition
				if !p.rules[ruleNonblankIndentedLine]() {
					goto l90
				}
				do(17)
				goto l89
			l90:
				position, thunkPosition = position90, thunkPosition90
			}
			do(18)
			doarg(yyPop, 1)
			return true
		l86:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 17 Verbatim <- (StartList (VerbatimChunk { a = cons(yy, a) })+ { yy = mk_str_from_list(a, false)
                 yy.key = VERBATIM }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l91
			}
			doarg(yySet, -1)
			if !p.rules[ruleVerbatimChunk]() {
				goto l91
			}
			do(19)
		l92:
			{
				position93, thunkPosition93 := position, thunkPosition
				if !p.rules[ruleVerbatimChunk]() {
					goto l93
				}
				do(19)
				goto l92
			l93:
				position, thunkPosition = position93, thunkPosition93
			}
			do(20)
			doarg(yyPop, 1)
			return true
		l91:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 18 HorizontalRule <- (NonindentSpace (('*' Sp '*' Sp '*' (Sp '*')*) / ('-' Sp '-' Sp '-' (Sp '-')*) / ('_' Sp '_' Sp '_' (Sp '_')*)) Sp Newline BlankLine+ { yy = mk_element(HRULE) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l94
			}
			{
				position95, thunkPosition95 := position, thunkPosition
				if !matchChar('*') {
					goto l96
				}
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
			l97:
				{
					position98, thunkPosition98 := position, thunkPosition
					if !p.rules[ruleSp]() {
						goto l98
					}
					if !matchChar('*') {
						goto l98
					}
					goto l97
				l98:
					position, thunkPosition = position98, thunkPosition98
				}
				goto l95
			l96:
				position, thunkPosition = position95, thunkPosition95
				if !matchChar('-') {
					goto l99
				}
				if !p.rules[ruleSp]() {
					goto l99
				}
				if !matchChar('-') {
					goto l99
				}
				if !p.rules[ruleSp]() {
					goto l99
				}
				if !matchChar('-') {
					goto l99
				}
			l100:
				{
					position101, thunkPosition101 := position, thunkPosition
					if !p.rules[ruleSp]() {
						goto l101
					}
					if !matchChar('-') {
						goto l101
					}
					goto l100
				l101:
					position, thunkPosition = position101, thunkPosition101
				}
				goto l95
			l99:
				position, thunkPosition = position95, thunkPosition95
				if !matchChar('_') {
					goto l94
				}
				if !p.rules[ruleSp]() {
					goto l94
				}
				if !matchChar('_') {
					goto l94
				}
				if !p.rules[ruleSp]() {
					goto l94
				}
				if !matchChar('_') {
					goto l94
				}
			l102:
				{
					position103, thunkPosition103 := position, thunkPosition
					if !p.rules[ruleSp]() {
						goto l103
					}
					if !matchChar('_') {
						goto l103
					}
					goto l102
				l103:
					position, thunkPosition = position103, thunkPosition103
				}
			}
		l95:
			if !p.rules[ruleSp]() {
				goto l94
			}
			if !p.rules[ruleNewline]() {
				goto l94
			}
			if !p.rules[ruleBlankLine]() {
				goto l94
			}
		l104:
			{
				position105, thunkPosition105 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l105
				}
				goto l104
			l105:
				position, thunkPosition = position105, thunkPosition105
			}
			do(21)
			return true
		l94:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 19 Bullet <- (!HorizontalRule NonindentSpace ('+' / '*' / '-') Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position107, thunkPosition107 := position, thunkPosition
				if !p.rules[ruleHorizontalRule]() {
					goto l107
				}
				goto l106
			l107:
				position, thunkPosition = position107, thunkPosition107
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l106
			}
			{
				position108, thunkPosition108 := position, thunkPosition
				if !matchChar('+') {
					goto l109
				}
				goto l108
			l109:
				position, thunkPosition = position108, thunkPosition108
				if !matchChar('*') {
					goto l110
				}
				goto l108
			l110:
				position, thunkPosition = position108, thunkPosition108
				if !matchChar('-') {
					goto l106
				}
			}
		l108:
			if !p.rules[ruleSpacechar]() {
				goto l106
			}
		l111:
			{
				position112, thunkPosition112 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l112
				}
				goto l111
			l112:
				position, thunkPosition = position112, thunkPosition112
			}
			return true
		l106:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 20 BulletList <- (&Bullet (ListTight / ListLoose) { yy.key = BULLETLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position114, thunkPosition114 := position, thunkPosition
				if !p.rules[ruleBullet]() {
					goto l113
				}
				position, thunkPosition = position114, thunkPosition114
			}
			{
				position115, thunkPosition115 := position, thunkPosition
				if !p.rules[ruleListTight]() {
					goto l116
				}
				goto l115
			l116:
				position, thunkPosition = position115, thunkPosition115
				if !p.rules[ruleListLoose]() {
					goto l113
				}
			}
		l115:
			do(22)
			return true
		l113:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 21 ListTight <- (StartList (ListItem { a = cons(yy, a) })+ BlankLine* !(Bullet / Enumerator / DefMarker) { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l117
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItem]() {
				goto l117
			}
			do(23)
		l118:
			{
				position119, thunkPosition119 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto l119
				}
				do(23)
				goto l118
			l119:
				position, thunkPosition = position119, thunkPosition119
			}
		l120:
			{
				position121, thunkPosition121 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l121
				}
				goto l120
			l121:
				position, thunkPosition = position121, thunkPosition121
			}
			{
				position122, thunkPosition122 := position, thunkPosition
				{
					position123, thunkPosition123 := position, thunkPosition
					if !p.rules[ruleBullet]() {
						goto l124
					}
					goto l123
				l124:
					position, thunkPosition = position123, thunkPosition123
					if !p.rules[ruleEnumerator]() {
						goto l125
					}
					goto l123
				l125:
					position, thunkPosition = position123, thunkPosition123
					if !p.rules[ruleDefMarker]() {
						goto l122
					}
				}
			l123:
				goto l117
			l122:
				position, thunkPosition = position122, thunkPosition122
			}
			do(24)
			doarg(yyPop, 1)
			return true
		l117:
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
				goto l126
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItem]() {
				goto l126
			}
			doarg(yySet, -2)
		l129:
			{
				position130, thunkPosition130 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l130
				}
				goto l129
			l130:
				position, thunkPosition = position130, thunkPosition130
			}
			do(25)
		l127:
			{
				position128, thunkPosition128 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto l128
				}
				doarg(yySet, -2)
			l131:
				{
					position132, thunkPosition132 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l132
					}
					goto l131
				l132:
					position, thunkPosition = position132, thunkPosition132
				}
				do(25)
				goto l127
			l128:
				position, thunkPosition = position128, thunkPosition128
			}
			do(26)
			doarg(yyPop, 2)
			return true
		l126:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 23 ListItem <- ((Bullet / Enumerator / DefMarker) StartList ListBlock { a = cons(yy, a) } (ListContinuationBlock { a = cons(yy, a) })* {
               raw := mk_str_from_list(a, false)
               raw.key = RAW
               yy = mk_element(LISTITEM)
               yy.children = raw
            }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position134, thunkPosition134 := position, thunkPosition
				if !p.rules[ruleBullet]() {
					goto l135
				}
				goto l134
			l135:
				position, thunkPosition = position134, thunkPosition134
				if !p.rules[ruleEnumerator]() {
					goto l136
				}
				goto l134
			l136:
				position, thunkPosition = position134, thunkPosition134
				if !p.rules[ruleDefMarker]() {
					goto l133
				}
			}
		l134:
			if !p.rules[ruleStartList]() {
				goto l133
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l133
			}
			do(27)
		l137:
			{
				position138, thunkPosition138 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l138
				}
				do(28)
				goto l137
			l138:
				position, thunkPosition = position138, thunkPosition138
			}
			do(29)
			doarg(yyPop, 1)
			return true
		l133:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 24 ListBlock <- (StartList Line { a = cons(yy, a) } (ListBlockLine { a = cons(yy, a) })* { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l139
			}
			doarg(yySet, -1)
			if !p.rules[ruleLine]() {
				goto l139
			}
			do(30)
		l140:
			{
				position141, thunkPosition141 := position, thunkPosition
				if !p.rules[ruleListBlockLine]() {
					goto l141
				}
				do(31)
				goto l140
			l141:
				position, thunkPosition = position141, thunkPosition141
			}
			do(32)
			doarg(yyPop, 1)
			return true
		l139:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 25 ListContinuationBlock <- (StartList (< BlankLine* > {   if len(yytext) == 0 {
                                   a = cons(mk_str("\001"), a) // block separator
                              } else {
                                   a = cons(mk_str(yytext), a)
                              }
                          }) (Indent ListBlock { a = cons(yy, a) })+ {  yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l142
			}
			doarg(yySet, -1)
			begin = position
		l143:
			{
				position144, thunkPosition144 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l144
				}
				goto l143
			l144:
				position, thunkPosition = position144, thunkPosition144
			}
			end = position
			do(33)
			if !p.rules[ruleIndent]() {
				goto l142
			}
			if !p.rules[ruleListBlock]() {
				goto l142
			}
			do(34)
		l145:
			{
				position146, thunkPosition146 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l146
				}
				if !p.rules[ruleListBlock]() {
					goto l146
				}
				do(34)
				goto l145
			l146:
				position, thunkPosition = position146, thunkPosition146
			}
			do(35)
			doarg(yyPop, 1)
			return true
		l142:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 26 Enumerator <- (NonindentSpace [0-9]+ '.' Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l147
			}
			if !matchClass(5) {
				goto l147
			}
		l148:
			{
				position149, thunkPosition149 := position, thunkPosition
				if !matchClass(5) {
					goto l149
				}
				goto l148
			l149:
				position, thunkPosition = position149, thunkPosition149
			}
			if !matchChar('.') {
				goto l147
			}
			if !p.rules[ruleSpacechar]() {
				goto l147
			}
		l150:
			{
				position151, thunkPosition151 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l151
				}
				goto l150
			l151:
				position, thunkPosition = position151, thunkPosition151
			}
			return true
		l147:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 27 OrderedList <- (&Enumerator (ListTight / ListLoose) { yy.key = ORDEREDLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position153, thunkPosition153 := position, thunkPosition
				if !p.rules[ruleEnumerator]() {
					goto l152
				}
				position, thunkPosition = position153, thunkPosition153
			}
			{
				position154, thunkPosition154 := position, thunkPosition
				if !p.rules[ruleListTight]() {
					goto l155
				}
				goto l154
			l155:
				position, thunkPosition = position154, thunkPosition154
				if !p.rules[ruleListLoose]() {
					goto l152
				}
			}
		l154:
			do(36)
			return true
		l152:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 28 ListBlockLine <- (!((Indent? (Bullet / Enumerator)) / DefMarker) !BlankLine !HorizontalRule OptionallyIndentedLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position157, thunkPosition157 := position, thunkPosition
				{
					position158, thunkPosition158 := position, thunkPosition
					{
						position160, thunkPosition160 := position, thunkPosition
						if !p.rules[ruleIndent]() {
							goto l160
						}
						goto l161
					l160:
						position, thunkPosition = position160, thunkPosition160
					}
				l161:
					{
						position162, thunkPosition162 := position, thunkPosition
						if !p.rules[ruleBullet]() {
							goto l163
						}
						goto l162
					l163:
						position, thunkPosition = position162, thunkPosition162
						if !p.rules[ruleEnumerator]() {
							goto l159
						}
					}
				l162:
					goto l158
				l159:
					position, thunkPosition = position158, thunkPosition158
					if !p.rules[ruleDefMarker]() {
						goto l157
					}
				}
			l158:
				goto l156
			l157:
				position, thunkPosition = position157, thunkPosition157
			}
			{
				position164, thunkPosition164 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l164
				}
				goto l156
			l164:
				position, thunkPosition = position164, thunkPosition164
			}
			{
				position165, thunkPosition165 := position, thunkPosition
				if !p.rules[ruleHorizontalRule]() {
					goto l165
				}
				goto l156
			l165:
				position, thunkPosition = position165, thunkPosition165
			}
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l156
			}
			return true
		l156:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 29 HtmlBlockOpenAddress <- ('<' Spnl ('address' / 'ADDRESS') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l166
			}
			if !p.rules[ruleSpnl]() {
				goto l166
			}
			{
				position167, thunkPosition167 := position, thunkPosition
				if !matchString("address") {
					goto l168
				}
				goto l167
			l168:
				position, thunkPosition = position167, thunkPosition167
				if !matchString("ADDRESS") {
					goto l166
				}
			}
		l167:
			if !p.rules[ruleSpnl]() {
				goto l166
			}
		l169:
			{
				position170, thunkPosition170 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l170
				}
				goto l169
			l170:
				position, thunkPosition = position170, thunkPosition170
			}
			if !matchChar('>') {
				goto l166
			}
			return true
		l166:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 HtmlBlockCloseAddress <- ('<' Spnl '/' ('address' / 'ADDRESS') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l171
			}
			if !p.rules[ruleSpnl]() {
				goto l171
			}
			if !matchChar('/') {
				goto l171
			}
			{
				position172, thunkPosition172 := position, thunkPosition
				if !matchString("address") {
					goto l173
				}
				goto l172
			l173:
				position, thunkPosition = position172, thunkPosition172
				if !matchString("ADDRESS") {
					goto l171
				}
			}
		l172:
			if !p.rules[ruleSpnl]() {
				goto l171
			}
			if !matchChar('>') {
				goto l171
			}
			return true
		l171:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 31 HtmlBlockOpenBlockquote <- ('<' Spnl ('blockquote' / 'BLOCKQUOTE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l174
			}
			if !p.rules[ruleSpnl]() {
				goto l174
			}
			{
				position175, thunkPosition175 := position, thunkPosition
				if !matchString("blockquote") {
					goto l176
				}
				goto l175
			l176:
				position, thunkPosition = position175, thunkPosition175
				if !matchString("BLOCKQUOTE") {
					goto l174
				}
			}
		l175:
			if !p.rules[ruleSpnl]() {
				goto l174
			}
		l177:
			{
				position178, thunkPosition178 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l178
				}
				goto l177
			l178:
				position, thunkPosition = position178, thunkPosition178
			}
			if !matchChar('>') {
				goto l174
			}
			return true
		l174:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 32 HtmlBlockCloseBlockquote <- ('<' Spnl '/' ('blockquote' / 'BLOCKQUOTE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l179
			}
			if !p.rules[ruleSpnl]() {
				goto l179
			}
			if !matchChar('/') {
				goto l179
			}
			{
				position180, thunkPosition180 := position, thunkPosition
				if !matchString("blockquote") {
					goto l181
				}
				goto l180
			l181:
				position, thunkPosition = position180, thunkPosition180
				if !matchString("BLOCKQUOTE") {
					goto l179
				}
			}
		l180:
			if !p.rules[ruleSpnl]() {
				goto l179
			}
			if !matchChar('>') {
				goto l179
			}
			return true
		l179:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 33 HtmlBlockOpenCenter <- ('<' Spnl ('center' / 'CENTER') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l182
			}
			if !p.rules[ruleSpnl]() {
				goto l182
			}
			{
				position183, thunkPosition183 := position, thunkPosition
				if !matchString("center") {
					goto l184
				}
				goto l183
			l184:
				position, thunkPosition = position183, thunkPosition183
				if !matchString("CENTER") {
					goto l182
				}
			}
		l183:
			if !p.rules[ruleSpnl]() {
				goto l182
			}
		l185:
			{
				position186, thunkPosition186 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l186
				}
				goto l185
			l186:
				position, thunkPosition = position186, thunkPosition186
			}
			if !matchChar('>') {
				goto l182
			}
			return true
		l182:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 34 HtmlBlockCloseCenter <- ('<' Spnl '/' ('center' / 'CENTER') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l187
			}
			if !p.rules[ruleSpnl]() {
				goto l187
			}
			if !matchChar('/') {
				goto l187
			}
			{
				position188, thunkPosition188 := position, thunkPosition
				if !matchString("center") {
					goto l189
				}
				goto l188
			l189:
				position, thunkPosition = position188, thunkPosition188
				if !matchString("CENTER") {
					goto l187
				}
			}
		l188:
			if !p.rules[ruleSpnl]() {
				goto l187
			}
			if !matchChar('>') {
				goto l187
			}
			return true
		l187:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 35 HtmlBlockOpenDir <- ('<' Spnl ('dir' / 'DIR') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l190
			}
			if !p.rules[ruleSpnl]() {
				goto l190
			}
			{
				position191, thunkPosition191 := position, thunkPosition
				if !matchString("dir") {
					goto l192
				}
				goto l191
			l192:
				position, thunkPosition = position191, thunkPosition191
				if !matchString("DIR") {
					goto l190
				}
			}
		l191:
			if !p.rules[ruleSpnl]() {
				goto l190
			}
		l193:
			{
				position194, thunkPosition194 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l194
				}
				goto l193
			l194:
				position, thunkPosition = position194, thunkPosition194
			}
			if !matchChar('>') {
				goto l190
			}
			return true
		l190:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 36 HtmlBlockCloseDir <- ('<' Spnl '/' ('dir' / 'DIR') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l195
			}
			if !p.rules[ruleSpnl]() {
				goto l195
			}
			if !matchChar('/') {
				goto l195
			}
			{
				position196, thunkPosition196 := position, thunkPosition
				if !matchString("dir") {
					goto l197
				}
				goto l196
			l197:
				position, thunkPosition = position196, thunkPosition196
				if !matchString("DIR") {
					goto l195
				}
			}
		l196:
			if !p.rules[ruleSpnl]() {
				goto l195
			}
			if !matchChar('>') {
				goto l195
			}
			return true
		l195:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 37 HtmlBlockOpenDiv <- ('<' Spnl ('div' / 'DIV') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l198
			}
			if !p.rules[ruleSpnl]() {
				goto l198
			}
			{
				position199, thunkPosition199 := position, thunkPosition
				if !matchString("div") {
					goto l200
				}
				goto l199
			l200:
				position, thunkPosition = position199, thunkPosition199
				if !matchString("DIV") {
					goto l198
				}
			}
		l199:
			if !p.rules[ruleSpnl]() {
				goto l198
			}
		l201:
			{
				position202, thunkPosition202 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l202
				}
				goto l201
			l202:
				position, thunkPosition = position202, thunkPosition202
			}
			if !matchChar('>') {
				goto l198
			}
			return true
		l198:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 38 HtmlBlockCloseDiv <- ('<' Spnl '/' ('div' / 'DIV') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l203
			}
			if !p.rules[ruleSpnl]() {
				goto l203
			}
			if !matchChar('/') {
				goto l203
			}
			{
				position204, thunkPosition204 := position, thunkPosition
				if !matchString("div") {
					goto l205
				}
				goto l204
			l205:
				position, thunkPosition = position204, thunkPosition204
				if !matchString("DIV") {
					goto l203
				}
			}
		l204:
			if !p.rules[ruleSpnl]() {
				goto l203
			}
			if !matchChar('>') {
				goto l203
			}
			return true
		l203:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 39 HtmlBlockOpenDl <- ('<' Spnl ('dl' / 'DL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l206
			}
			if !p.rules[ruleSpnl]() {
				goto l206
			}
			{
				position207, thunkPosition207 := position, thunkPosition
				if !matchString("dl") {
					goto l208
				}
				goto l207
			l208:
				position, thunkPosition = position207, thunkPosition207
				if !matchString("DL") {
					goto l206
				}
			}
		l207:
			if !p.rules[ruleSpnl]() {
				goto l206
			}
		l209:
			{
				position210, thunkPosition210 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l210
				}
				goto l209
			l210:
				position, thunkPosition = position210, thunkPosition210
			}
			if !matchChar('>') {
				goto l206
			}
			return true
		l206:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 40 HtmlBlockCloseDl <- ('<' Spnl '/' ('dl' / 'DL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l211
			}
			if !p.rules[ruleSpnl]() {
				goto l211
			}
			if !matchChar('/') {
				goto l211
			}
			{
				position212, thunkPosition212 := position, thunkPosition
				if !matchString("dl") {
					goto l213
				}
				goto l212
			l213:
				position, thunkPosition = position212, thunkPosition212
				if !matchString("DL") {
					goto l211
				}
			}
		l212:
			if !p.rules[ruleSpnl]() {
				goto l211
			}
			if !matchChar('>') {
				goto l211
			}
			return true
		l211:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 41 HtmlBlockOpenFieldset <- ('<' Spnl ('fieldset' / 'FIELDSET') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l214
			}
			if !p.rules[ruleSpnl]() {
				goto l214
			}
			{
				position215, thunkPosition215 := position, thunkPosition
				if !matchString("fieldset") {
					goto l216
				}
				goto l215
			l216:
				position, thunkPosition = position215, thunkPosition215
				if !matchString("FIELDSET") {
					goto l214
				}
			}
		l215:
			if !p.rules[ruleSpnl]() {
				goto l214
			}
		l217:
			{
				position218, thunkPosition218 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l218
				}
				goto l217
			l218:
				position, thunkPosition = position218, thunkPosition218
			}
			if !matchChar('>') {
				goto l214
			}
			return true
		l214:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 42 HtmlBlockCloseFieldset <- ('<' Spnl '/' ('fieldset' / 'FIELDSET') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l219
			}
			if !p.rules[ruleSpnl]() {
				goto l219
			}
			if !matchChar('/') {
				goto l219
			}
			{
				position220, thunkPosition220 := position, thunkPosition
				if !matchString("fieldset") {
					goto l221
				}
				goto l220
			l221:
				position, thunkPosition = position220, thunkPosition220
				if !matchString("FIELDSET") {
					goto l219
				}
			}
		l220:
			if !p.rules[ruleSpnl]() {
				goto l219
			}
			if !matchChar('>') {
				goto l219
			}
			return true
		l219:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 43 HtmlBlockOpenForm <- ('<' Spnl ('form' / 'FORM') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l222
			}
			if !p.rules[ruleSpnl]() {
				goto l222
			}
			{
				position223, thunkPosition223 := position, thunkPosition
				if !matchString("form") {
					goto l224
				}
				goto l223
			l224:
				position, thunkPosition = position223, thunkPosition223
				if !matchString("FORM") {
					goto l222
				}
			}
		l223:
			if !p.rules[ruleSpnl]() {
				goto l222
			}
		l225:
			{
				position226, thunkPosition226 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l226
				}
				goto l225
			l226:
				position, thunkPosition = position226, thunkPosition226
			}
			if !matchChar('>') {
				goto l222
			}
			return true
		l222:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 44 HtmlBlockCloseForm <- ('<' Spnl '/' ('form' / 'FORM') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l227
			}
			if !p.rules[ruleSpnl]() {
				goto l227
			}
			if !matchChar('/') {
				goto l227
			}
			{
				position228, thunkPosition228 := position, thunkPosition
				if !matchString("form") {
					goto l229
				}
				goto l228
			l229:
				position, thunkPosition = position228, thunkPosition228
				if !matchString("FORM") {
					goto l227
				}
			}
		l228:
			if !p.rules[ruleSpnl]() {
				goto l227
			}
			if !matchChar('>') {
				goto l227
			}
			return true
		l227:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 45 HtmlBlockOpenH1 <- ('<' Spnl ('h1' / 'H1') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l230
			}
			if !p.rules[ruleSpnl]() {
				goto l230
			}
			{
				position231, thunkPosition231 := position, thunkPosition
				if !matchString("h1") {
					goto l232
				}
				goto l231
			l232:
				position, thunkPosition = position231, thunkPosition231
				if !matchString("H1") {
					goto l230
				}
			}
		l231:
			if !p.rules[ruleSpnl]() {
				goto l230
			}
		l233:
			{
				position234, thunkPosition234 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l234
				}
				goto l233
			l234:
				position, thunkPosition = position234, thunkPosition234
			}
			if !matchChar('>') {
				goto l230
			}
			return true
		l230:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 46 HtmlBlockCloseH1 <- ('<' Spnl '/' ('h1' / 'H1') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l235
			}
			if !p.rules[ruleSpnl]() {
				goto l235
			}
			if !matchChar('/') {
				goto l235
			}
			{
				position236, thunkPosition236 := position, thunkPosition
				if !matchString("h1") {
					goto l237
				}
				goto l236
			l237:
				position, thunkPosition = position236, thunkPosition236
				if !matchString("H1") {
					goto l235
				}
			}
		l236:
			if !p.rules[ruleSpnl]() {
				goto l235
			}
			if !matchChar('>') {
				goto l235
			}
			return true
		l235:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 47 HtmlBlockOpenH2 <- ('<' Spnl ('h2' / 'H2') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l238
			}
			if !p.rules[ruleSpnl]() {
				goto l238
			}
			{
				position239, thunkPosition239 := position, thunkPosition
				if !matchString("h2") {
					goto l240
				}
				goto l239
			l240:
				position, thunkPosition = position239, thunkPosition239
				if !matchString("H2") {
					goto l238
				}
			}
		l239:
			if !p.rules[ruleSpnl]() {
				goto l238
			}
		l241:
			{
				position242, thunkPosition242 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l242
				}
				goto l241
			l242:
				position, thunkPosition = position242, thunkPosition242
			}
			if !matchChar('>') {
				goto l238
			}
			return true
		l238:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 48 HtmlBlockCloseH2 <- ('<' Spnl '/' ('h2' / 'H2') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l243
			}
			if !p.rules[ruleSpnl]() {
				goto l243
			}
			if !matchChar('/') {
				goto l243
			}
			{
				position244, thunkPosition244 := position, thunkPosition
				if !matchString("h2") {
					goto l245
				}
				goto l244
			l245:
				position, thunkPosition = position244, thunkPosition244
				if !matchString("H2") {
					goto l243
				}
			}
		l244:
			if !p.rules[ruleSpnl]() {
				goto l243
			}
			if !matchChar('>') {
				goto l243
			}
			return true
		l243:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 49 HtmlBlockOpenH3 <- ('<' Spnl ('h3' / 'H3') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l246
			}
			if !p.rules[ruleSpnl]() {
				goto l246
			}
			{
				position247, thunkPosition247 := position, thunkPosition
				if !matchString("h3") {
					goto l248
				}
				goto l247
			l248:
				position, thunkPosition = position247, thunkPosition247
				if !matchString("H3") {
					goto l246
				}
			}
		l247:
			if !p.rules[ruleSpnl]() {
				goto l246
			}
		l249:
			{
				position250, thunkPosition250 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l250
				}
				goto l249
			l250:
				position, thunkPosition = position250, thunkPosition250
			}
			if !matchChar('>') {
				goto l246
			}
			return true
		l246:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 50 HtmlBlockCloseH3 <- ('<' Spnl '/' ('h3' / 'H3') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l251
			}
			if !p.rules[ruleSpnl]() {
				goto l251
			}
			if !matchChar('/') {
				goto l251
			}
			{
				position252, thunkPosition252 := position, thunkPosition
				if !matchString("h3") {
					goto l253
				}
				goto l252
			l253:
				position, thunkPosition = position252, thunkPosition252
				if !matchString("H3") {
					goto l251
				}
			}
		l252:
			if !p.rules[ruleSpnl]() {
				goto l251
			}
			if !matchChar('>') {
				goto l251
			}
			return true
		l251:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 51 HtmlBlockOpenH4 <- ('<' Spnl ('h4' / 'H4') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l254
			}
			if !p.rules[ruleSpnl]() {
				goto l254
			}
			{
				position255, thunkPosition255 := position, thunkPosition
				if !matchString("h4") {
					goto l256
				}
				goto l255
			l256:
				position, thunkPosition = position255, thunkPosition255
				if !matchString("H4") {
					goto l254
				}
			}
		l255:
			if !p.rules[ruleSpnl]() {
				goto l254
			}
		l257:
			{
				position258, thunkPosition258 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l258
				}
				goto l257
			l258:
				position, thunkPosition = position258, thunkPosition258
			}
			if !matchChar('>') {
				goto l254
			}
			return true
		l254:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 52 HtmlBlockCloseH4 <- ('<' Spnl '/' ('h4' / 'H4') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l259
			}
			if !p.rules[ruleSpnl]() {
				goto l259
			}
			if !matchChar('/') {
				goto l259
			}
			{
				position260, thunkPosition260 := position, thunkPosition
				if !matchString("h4") {
					goto l261
				}
				goto l260
			l261:
				position, thunkPosition = position260, thunkPosition260
				if !matchString("H4") {
					goto l259
				}
			}
		l260:
			if !p.rules[ruleSpnl]() {
				goto l259
			}
			if !matchChar('>') {
				goto l259
			}
			return true
		l259:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 53 HtmlBlockOpenH5 <- ('<' Spnl ('h5' / 'H5') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l262
			}
			if !p.rules[ruleSpnl]() {
				goto l262
			}
			{
				position263, thunkPosition263 := position, thunkPosition
				if !matchString("h5") {
					goto l264
				}
				goto l263
			l264:
				position, thunkPosition = position263, thunkPosition263
				if !matchString("H5") {
					goto l262
				}
			}
		l263:
			if !p.rules[ruleSpnl]() {
				goto l262
			}
		l265:
			{
				position266, thunkPosition266 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l266
				}
				goto l265
			l266:
				position, thunkPosition = position266, thunkPosition266
			}
			if !matchChar('>') {
				goto l262
			}
			return true
		l262:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 54 HtmlBlockCloseH5 <- ('<' Spnl '/' ('h5' / 'H5') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l267
			}
			if !p.rules[ruleSpnl]() {
				goto l267
			}
			if !matchChar('/') {
				goto l267
			}
			{
				position268, thunkPosition268 := position, thunkPosition
				if !matchString("h5") {
					goto l269
				}
				goto l268
			l269:
				position, thunkPosition = position268, thunkPosition268
				if !matchString("H5") {
					goto l267
				}
			}
		l268:
			if !p.rules[ruleSpnl]() {
				goto l267
			}
			if !matchChar('>') {
				goto l267
			}
			return true
		l267:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 55 HtmlBlockOpenH6 <- ('<' Spnl ('h6' / 'H6') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l270
			}
			if !p.rules[ruleSpnl]() {
				goto l270
			}
			{
				position271, thunkPosition271 := position, thunkPosition
				if !matchString("h6") {
					goto l272
				}
				goto l271
			l272:
				position, thunkPosition = position271, thunkPosition271
				if !matchString("H6") {
					goto l270
				}
			}
		l271:
			if !p.rules[ruleSpnl]() {
				goto l270
			}
		l273:
			{
				position274, thunkPosition274 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l274
				}
				goto l273
			l274:
				position, thunkPosition = position274, thunkPosition274
			}
			if !matchChar('>') {
				goto l270
			}
			return true
		l270:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 56 HtmlBlockCloseH6 <- ('<' Spnl '/' ('h6' / 'H6') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l275
			}
			if !p.rules[ruleSpnl]() {
				goto l275
			}
			if !matchChar('/') {
				goto l275
			}
			{
				position276, thunkPosition276 := position, thunkPosition
				if !matchString("h6") {
					goto l277
				}
				goto l276
			l277:
				position, thunkPosition = position276, thunkPosition276
				if !matchString("H6") {
					goto l275
				}
			}
		l276:
			if !p.rules[ruleSpnl]() {
				goto l275
			}
			if !matchChar('>') {
				goto l275
			}
			return true
		l275:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 57 HtmlBlockOpenMenu <- ('<' Spnl ('menu' / 'MENU') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l278
			}
			if !p.rules[ruleSpnl]() {
				goto l278
			}
			{
				position279, thunkPosition279 := position, thunkPosition
				if !matchString("menu") {
					goto l280
				}
				goto l279
			l280:
				position, thunkPosition = position279, thunkPosition279
				if !matchString("MENU") {
					goto l278
				}
			}
		l279:
			if !p.rules[ruleSpnl]() {
				goto l278
			}
		l281:
			{
				position282, thunkPosition282 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l282
				}
				goto l281
			l282:
				position, thunkPosition = position282, thunkPosition282
			}
			if !matchChar('>') {
				goto l278
			}
			return true
		l278:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 58 HtmlBlockCloseMenu <- ('<' Spnl '/' ('menu' / 'MENU') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l283
			}
			if !p.rules[ruleSpnl]() {
				goto l283
			}
			if !matchChar('/') {
				goto l283
			}
			{
				position284, thunkPosition284 := position, thunkPosition
				if !matchString("menu") {
					goto l285
				}
				goto l284
			l285:
				position, thunkPosition = position284, thunkPosition284
				if !matchString("MENU") {
					goto l283
				}
			}
		l284:
			if !p.rules[ruleSpnl]() {
				goto l283
			}
			if !matchChar('>') {
				goto l283
			}
			return true
		l283:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 59 HtmlBlockOpenNoframes <- ('<' Spnl ('noframes' / 'NOFRAMES') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l286
			}
			if !p.rules[ruleSpnl]() {
				goto l286
			}
			{
				position287, thunkPosition287 := position, thunkPosition
				if !matchString("noframes") {
					goto l288
				}
				goto l287
			l288:
				position, thunkPosition = position287, thunkPosition287
				if !matchString("NOFRAMES") {
					goto l286
				}
			}
		l287:
			if !p.rules[ruleSpnl]() {
				goto l286
			}
		l289:
			{
				position290, thunkPosition290 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l290
				}
				goto l289
			l290:
				position, thunkPosition = position290, thunkPosition290
			}
			if !matchChar('>') {
				goto l286
			}
			return true
		l286:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 60 HtmlBlockCloseNoframes <- ('<' Spnl '/' ('noframes' / 'NOFRAMES') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l291
			}
			if !p.rules[ruleSpnl]() {
				goto l291
			}
			if !matchChar('/') {
				goto l291
			}
			{
				position292, thunkPosition292 := position, thunkPosition
				if !matchString("noframes") {
					goto l293
				}
				goto l292
			l293:
				position, thunkPosition = position292, thunkPosition292
				if !matchString("NOFRAMES") {
					goto l291
				}
			}
		l292:
			if !p.rules[ruleSpnl]() {
				goto l291
			}
			if !matchChar('>') {
				goto l291
			}
			return true
		l291:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 61 HtmlBlockOpenNoscript <- ('<' Spnl ('noscript' / 'NOSCRIPT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l294
			}
			if !p.rules[ruleSpnl]() {
				goto l294
			}
			{
				position295, thunkPosition295 := position, thunkPosition
				if !matchString("noscript") {
					goto l296
				}
				goto l295
			l296:
				position, thunkPosition = position295, thunkPosition295
				if !matchString("NOSCRIPT") {
					goto l294
				}
			}
		l295:
			if !p.rules[ruleSpnl]() {
				goto l294
			}
		l297:
			{
				position298, thunkPosition298 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l298
				}
				goto l297
			l298:
				position, thunkPosition = position298, thunkPosition298
			}
			if !matchChar('>') {
				goto l294
			}
			return true
		l294:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 62 HtmlBlockCloseNoscript <- ('<' Spnl '/' ('noscript' / 'NOSCRIPT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l299
			}
			if !p.rules[ruleSpnl]() {
				goto l299
			}
			if !matchChar('/') {
				goto l299
			}
			{
				position300, thunkPosition300 := position, thunkPosition
				if !matchString("noscript") {
					goto l301
				}
				goto l300
			l301:
				position, thunkPosition = position300, thunkPosition300
				if !matchString("NOSCRIPT") {
					goto l299
				}
			}
		l300:
			if !p.rules[ruleSpnl]() {
				goto l299
			}
			if !matchChar('>') {
				goto l299
			}
			return true
		l299:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 63 HtmlBlockOpenOl <- ('<' Spnl ('ol' / 'OL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l302
			}
			if !p.rules[ruleSpnl]() {
				goto l302
			}
			{
				position303, thunkPosition303 := position, thunkPosition
				if !matchString("ol") {
					goto l304
				}
				goto l303
			l304:
				position, thunkPosition = position303, thunkPosition303
				if !matchString("OL") {
					goto l302
				}
			}
		l303:
			if !p.rules[ruleSpnl]() {
				goto l302
			}
		l305:
			{
				position306, thunkPosition306 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l306
				}
				goto l305
			l306:
				position, thunkPosition = position306, thunkPosition306
			}
			if !matchChar('>') {
				goto l302
			}
			return true
		l302:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 64 HtmlBlockCloseOl <- ('<' Spnl '/' ('ol' / 'OL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l307
			}
			if !p.rules[ruleSpnl]() {
				goto l307
			}
			if !matchChar('/') {
				goto l307
			}
			{
				position308, thunkPosition308 := position, thunkPosition
				if !matchString("ol") {
					goto l309
				}
				goto l308
			l309:
				position, thunkPosition = position308, thunkPosition308
				if !matchString("OL") {
					goto l307
				}
			}
		l308:
			if !p.rules[ruleSpnl]() {
				goto l307
			}
			if !matchChar('>') {
				goto l307
			}
			return true
		l307:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 65 HtmlBlockOpenP <- ('<' Spnl ('p' / 'P') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l310
			}
			if !p.rules[ruleSpnl]() {
				goto l310
			}
			{
				position311, thunkPosition311 := position, thunkPosition
				if !matchChar('p') {
					goto l312
				}
				goto l311
			l312:
				position, thunkPosition = position311, thunkPosition311
				if !matchChar('P') {
					goto l310
				}
			}
		l311:
			if !p.rules[ruleSpnl]() {
				goto l310
			}
		l313:
			{
				position314, thunkPosition314 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l314
				}
				goto l313
			l314:
				position, thunkPosition = position314, thunkPosition314
			}
			if !matchChar('>') {
				goto l310
			}
			return true
		l310:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 66 HtmlBlockCloseP <- ('<' Spnl '/' ('p' / 'P') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l315
			}
			if !p.rules[ruleSpnl]() {
				goto l315
			}
			if !matchChar('/') {
				goto l315
			}
			{
				position316, thunkPosition316 := position, thunkPosition
				if !matchChar('p') {
					goto l317
				}
				goto l316
			l317:
				position, thunkPosition = position316, thunkPosition316
				if !matchChar('P') {
					goto l315
				}
			}
		l316:
			if !p.rules[ruleSpnl]() {
				goto l315
			}
			if !matchChar('>') {
				goto l315
			}
			return true
		l315:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 67 HtmlBlockOpenPre <- ('<' Spnl ('pre' / 'PRE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l318
			}
			if !p.rules[ruleSpnl]() {
				goto l318
			}
			{
				position319, thunkPosition319 := position, thunkPosition
				if !matchString("pre") {
					goto l320
				}
				goto l319
			l320:
				position, thunkPosition = position319, thunkPosition319
				if !matchString("PRE") {
					goto l318
				}
			}
		l319:
			if !p.rules[ruleSpnl]() {
				goto l318
			}
		l321:
			{
				position322, thunkPosition322 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l322
				}
				goto l321
			l322:
				position, thunkPosition = position322, thunkPosition322
			}
			if !matchChar('>') {
				goto l318
			}
			return true
		l318:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 68 HtmlBlockClosePre <- ('<' Spnl '/' ('pre' / 'PRE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l323
			}
			if !p.rules[ruleSpnl]() {
				goto l323
			}
			if !matchChar('/') {
				goto l323
			}
			{
				position324, thunkPosition324 := position, thunkPosition
				if !matchString("pre") {
					goto l325
				}
				goto l324
			l325:
				position, thunkPosition = position324, thunkPosition324
				if !matchString("PRE") {
					goto l323
				}
			}
		l324:
			if !p.rules[ruleSpnl]() {
				goto l323
			}
			if !matchChar('>') {
				goto l323
			}
			return true
		l323:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 69 HtmlBlockOpenTable <- ('<' Spnl ('table' / 'TABLE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l326
			}
			if !p.rules[ruleSpnl]() {
				goto l326
			}
			{
				position327, thunkPosition327 := position, thunkPosition
				if !matchString("table") {
					goto l328
				}
				goto l327
			l328:
				position, thunkPosition = position327, thunkPosition327
				if !matchString("TABLE") {
					goto l326
				}
			}
		l327:
			if !p.rules[ruleSpnl]() {
				goto l326
			}
		l329:
			{
				position330, thunkPosition330 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l330
				}
				goto l329
			l330:
				position, thunkPosition = position330, thunkPosition330
			}
			if !matchChar('>') {
				goto l326
			}
			return true
		l326:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 70 HtmlBlockCloseTable <- ('<' Spnl '/' ('table' / 'TABLE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l331
			}
			if !p.rules[ruleSpnl]() {
				goto l331
			}
			if !matchChar('/') {
				goto l331
			}
			{
				position332, thunkPosition332 := position, thunkPosition
				if !matchString("table") {
					goto l333
				}
				goto l332
			l333:
				position, thunkPosition = position332, thunkPosition332
				if !matchString("TABLE") {
					goto l331
				}
			}
		l332:
			if !p.rules[ruleSpnl]() {
				goto l331
			}
			if !matchChar('>') {
				goto l331
			}
			return true
		l331:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 71 HtmlBlockOpenUl <- ('<' Spnl ('ul' / 'UL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l334
			}
			if !p.rules[ruleSpnl]() {
				goto l334
			}
			{
				position335, thunkPosition335 := position, thunkPosition
				if !matchString("ul") {
					goto l336
				}
				goto l335
			l336:
				position, thunkPosition = position335, thunkPosition335
				if !matchString("UL") {
					goto l334
				}
			}
		l335:
			if !p.rules[ruleSpnl]() {
				goto l334
			}
		l337:
			{
				position338, thunkPosition338 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l338
				}
				goto l337
			l338:
				position, thunkPosition = position338, thunkPosition338
			}
			if !matchChar('>') {
				goto l334
			}
			return true
		l334:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 72 HtmlBlockCloseUl <- ('<' Spnl '/' ('ul' / 'UL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l339
			}
			if !p.rules[ruleSpnl]() {
				goto l339
			}
			if !matchChar('/') {
				goto l339
			}
			{
				position340, thunkPosition340 := position, thunkPosition
				if !matchString("ul") {
					goto l341
				}
				goto l340
			l341:
				position, thunkPosition = position340, thunkPosition340
				if !matchString("UL") {
					goto l339
				}
			}
		l340:
			if !p.rules[ruleSpnl]() {
				goto l339
			}
			if !matchChar('>') {
				goto l339
			}
			return true
		l339:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 73 HtmlBlockOpenDd <- ('<' Spnl ('dd' / 'DD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l342
			}
			if !p.rules[ruleSpnl]() {
				goto l342
			}
			{
				position343, thunkPosition343 := position, thunkPosition
				if !matchString("dd") {
					goto l344
				}
				goto l343
			l344:
				position, thunkPosition = position343, thunkPosition343
				if !matchString("DD") {
					goto l342
				}
			}
		l343:
			if !p.rules[ruleSpnl]() {
				goto l342
			}
		l345:
			{
				position346, thunkPosition346 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l346
				}
				goto l345
			l346:
				position, thunkPosition = position346, thunkPosition346
			}
			if !matchChar('>') {
				goto l342
			}
			return true
		l342:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 74 HtmlBlockCloseDd <- ('<' Spnl '/' ('dd' / 'DD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l347
			}
			if !p.rules[ruleSpnl]() {
				goto l347
			}
			if !matchChar('/') {
				goto l347
			}
			{
				position348, thunkPosition348 := position, thunkPosition
				if !matchString("dd") {
					goto l349
				}
				goto l348
			l349:
				position, thunkPosition = position348, thunkPosition348
				if !matchString("DD") {
					goto l347
				}
			}
		l348:
			if !p.rules[ruleSpnl]() {
				goto l347
			}
			if !matchChar('>') {
				goto l347
			}
			return true
		l347:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 75 HtmlBlockOpenDt <- ('<' Spnl ('dt' / 'DT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l350
			}
			if !p.rules[ruleSpnl]() {
				goto l350
			}
			{
				position351, thunkPosition351 := position, thunkPosition
				if !matchString("dt") {
					goto l352
				}
				goto l351
			l352:
				position, thunkPosition = position351, thunkPosition351
				if !matchString("DT") {
					goto l350
				}
			}
		l351:
			if !p.rules[ruleSpnl]() {
				goto l350
			}
		l353:
			{
				position354, thunkPosition354 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l354
				}
				goto l353
			l354:
				position, thunkPosition = position354, thunkPosition354
			}
			if !matchChar('>') {
				goto l350
			}
			return true
		l350:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 76 HtmlBlockCloseDt <- ('<' Spnl '/' ('dt' / 'DT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l355
			}
			if !p.rules[ruleSpnl]() {
				goto l355
			}
			if !matchChar('/') {
				goto l355
			}
			{
				position356, thunkPosition356 := position, thunkPosition
				if !matchString("dt") {
					goto l357
				}
				goto l356
			l357:
				position, thunkPosition = position356, thunkPosition356
				if !matchString("DT") {
					goto l355
				}
			}
		l356:
			if !p.rules[ruleSpnl]() {
				goto l355
			}
			if !matchChar('>') {
				goto l355
			}
			return true
		l355:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 77 HtmlBlockOpenFrameset <- ('<' Spnl ('frameset' / 'FRAMESET') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l358
			}
			if !p.rules[ruleSpnl]() {
				goto l358
			}
			{
				position359, thunkPosition359 := position, thunkPosition
				if !matchString("frameset") {
					goto l360
				}
				goto l359
			l360:
				position, thunkPosition = position359, thunkPosition359
				if !matchString("FRAMESET") {
					goto l358
				}
			}
		l359:
			if !p.rules[ruleSpnl]() {
				goto l358
			}
		l361:
			{
				position362, thunkPosition362 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l362
				}
				goto l361
			l362:
				position, thunkPosition = position362, thunkPosition362
			}
			if !matchChar('>') {
				goto l358
			}
			return true
		l358:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 78 HtmlBlockCloseFrameset <- ('<' Spnl '/' ('frameset' / 'FRAMESET') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l363
			}
			if !p.rules[ruleSpnl]() {
				goto l363
			}
			if !matchChar('/') {
				goto l363
			}
			{
				position364, thunkPosition364 := position, thunkPosition
				if !matchString("frameset") {
					goto l365
				}
				goto l364
			l365:
				position, thunkPosition = position364, thunkPosition364
				if !matchString("FRAMESET") {
					goto l363
				}
			}
		l364:
			if !p.rules[ruleSpnl]() {
				goto l363
			}
			if !matchChar('>') {
				goto l363
			}
			return true
		l363:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 79 HtmlBlockOpenLi <- ('<' Spnl ('li' / 'LI') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l366
			}
			if !p.rules[ruleSpnl]() {
				goto l366
			}
			{
				position367, thunkPosition367 := position, thunkPosition
				if !matchString("li") {
					goto l368
				}
				goto l367
			l368:
				position, thunkPosition = position367, thunkPosition367
				if !matchString("LI") {
					goto l366
				}
			}
		l367:
			if !p.rules[ruleSpnl]() {
				goto l366
			}
		l369:
			{
				position370, thunkPosition370 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l370
				}
				goto l369
			l370:
				position, thunkPosition = position370, thunkPosition370
			}
			if !matchChar('>') {
				goto l366
			}
			return true
		l366:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 80 HtmlBlockCloseLi <- ('<' Spnl '/' ('li' / 'LI') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l371
			}
			if !p.rules[ruleSpnl]() {
				goto l371
			}
			if !matchChar('/') {
				goto l371
			}
			{
				position372, thunkPosition372 := position, thunkPosition
				if !matchString("li") {
					goto l373
				}
				goto l372
			l373:
				position, thunkPosition = position372, thunkPosition372
				if !matchString("LI") {
					goto l371
				}
			}
		l372:
			if !p.rules[ruleSpnl]() {
				goto l371
			}
			if !matchChar('>') {
				goto l371
			}
			return true
		l371:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 81 HtmlBlockOpenTbody <- ('<' Spnl ('tbody' / 'TBODY') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l374
			}
			if !p.rules[ruleSpnl]() {
				goto l374
			}
			{
				position375, thunkPosition375 := position, thunkPosition
				if !matchString("tbody") {
					goto l376
				}
				goto l375
			l376:
				position, thunkPosition = position375, thunkPosition375
				if !matchString("TBODY") {
					goto l374
				}
			}
		l375:
			if !p.rules[ruleSpnl]() {
				goto l374
			}
		l377:
			{
				position378, thunkPosition378 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l378
				}
				goto l377
			l378:
				position, thunkPosition = position378, thunkPosition378
			}
			if !matchChar('>') {
				goto l374
			}
			return true
		l374:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 82 HtmlBlockCloseTbody <- ('<' Spnl '/' ('tbody' / 'TBODY') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l379
			}
			if !p.rules[ruleSpnl]() {
				goto l379
			}
			if !matchChar('/') {
				goto l379
			}
			{
				position380, thunkPosition380 := position, thunkPosition
				if !matchString("tbody") {
					goto l381
				}
				goto l380
			l381:
				position, thunkPosition = position380, thunkPosition380
				if !matchString("TBODY") {
					goto l379
				}
			}
		l380:
			if !p.rules[ruleSpnl]() {
				goto l379
			}
			if !matchChar('>') {
				goto l379
			}
			return true
		l379:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 83 HtmlBlockOpenTd <- ('<' Spnl ('td' / 'TD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l382
			}
			if !p.rules[ruleSpnl]() {
				goto l382
			}
			{
				position383, thunkPosition383 := position, thunkPosition
				if !matchString("td") {
					goto l384
				}
				goto l383
			l384:
				position, thunkPosition = position383, thunkPosition383
				if !matchString("TD") {
					goto l382
				}
			}
		l383:
			if !p.rules[ruleSpnl]() {
				goto l382
			}
		l385:
			{
				position386, thunkPosition386 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l386
				}
				goto l385
			l386:
				position, thunkPosition = position386, thunkPosition386
			}
			if !matchChar('>') {
				goto l382
			}
			return true
		l382:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 84 HtmlBlockCloseTd <- ('<' Spnl '/' ('td' / 'TD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l387
			}
			if !p.rules[ruleSpnl]() {
				goto l387
			}
			if !matchChar('/') {
				goto l387
			}
			{
				position388, thunkPosition388 := position, thunkPosition
				if !matchString("td") {
					goto l389
				}
				goto l388
			l389:
				position, thunkPosition = position388, thunkPosition388
				if !matchString("TD") {
					goto l387
				}
			}
		l388:
			if !p.rules[ruleSpnl]() {
				goto l387
			}
			if !matchChar('>') {
				goto l387
			}
			return true
		l387:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 85 HtmlBlockOpenTfoot <- ('<' Spnl ('tfoot' / 'TFOOT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l390
			}
			if !p.rules[ruleSpnl]() {
				goto l390
			}
			{
				position391, thunkPosition391 := position, thunkPosition
				if !matchString("tfoot") {
					goto l392
				}
				goto l391
			l392:
				position, thunkPosition = position391, thunkPosition391
				if !matchString("TFOOT") {
					goto l390
				}
			}
		l391:
			if !p.rules[ruleSpnl]() {
				goto l390
			}
		l393:
			{
				position394, thunkPosition394 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l394
				}
				goto l393
			l394:
				position, thunkPosition = position394, thunkPosition394
			}
			if !matchChar('>') {
				goto l390
			}
			return true
		l390:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 86 HtmlBlockCloseTfoot <- ('<' Spnl '/' ('tfoot' / 'TFOOT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l395
			}
			if !p.rules[ruleSpnl]() {
				goto l395
			}
			if !matchChar('/') {
				goto l395
			}
			{
				position396, thunkPosition396 := position, thunkPosition
				if !matchString("tfoot") {
					goto l397
				}
				goto l396
			l397:
				position, thunkPosition = position396, thunkPosition396
				if !matchString("TFOOT") {
					goto l395
				}
			}
		l396:
			if !p.rules[ruleSpnl]() {
				goto l395
			}
			if !matchChar('>') {
				goto l395
			}
			return true
		l395:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 87 HtmlBlockOpenTh <- ('<' Spnl ('th' / 'TH') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l398
			}
			if !p.rules[ruleSpnl]() {
				goto l398
			}
			{
				position399, thunkPosition399 := position, thunkPosition
				if !matchString("th") {
					goto l400
				}
				goto l399
			l400:
				position, thunkPosition = position399, thunkPosition399
				if !matchString("TH") {
					goto l398
				}
			}
		l399:
			if !p.rules[ruleSpnl]() {
				goto l398
			}
		l401:
			{
				position402, thunkPosition402 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l402
				}
				goto l401
			l402:
				position, thunkPosition = position402, thunkPosition402
			}
			if !matchChar('>') {
				goto l398
			}
			return true
		l398:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 88 HtmlBlockCloseTh <- ('<' Spnl '/' ('th' / 'TH') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l403
			}
			if !p.rules[ruleSpnl]() {
				goto l403
			}
			if !matchChar('/') {
				goto l403
			}
			{
				position404, thunkPosition404 := position, thunkPosition
				if !matchString("th") {
					goto l405
				}
				goto l404
			l405:
				position, thunkPosition = position404, thunkPosition404
				if !matchString("TH") {
					goto l403
				}
			}
		l404:
			if !p.rules[ruleSpnl]() {
				goto l403
			}
			if !matchChar('>') {
				goto l403
			}
			return true
		l403:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 89 HtmlBlockOpenThead <- ('<' Spnl ('thead' / 'THEAD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l406
			}
			if !p.rules[ruleSpnl]() {
				goto l406
			}
			{
				position407, thunkPosition407 := position, thunkPosition
				if !matchString("thead") {
					goto l408
				}
				goto l407
			l408:
				position, thunkPosition = position407, thunkPosition407
				if !matchString("THEAD") {
					goto l406
				}
			}
		l407:
			if !p.rules[ruleSpnl]() {
				goto l406
			}
		l409:
			{
				position410, thunkPosition410 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l410
				}
				goto l409
			l410:
				position, thunkPosition = position410, thunkPosition410
			}
			if !matchChar('>') {
				goto l406
			}
			return true
		l406:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 90 HtmlBlockCloseThead <- ('<' Spnl '/' ('thead' / 'THEAD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l411
			}
			if !p.rules[ruleSpnl]() {
				goto l411
			}
			if !matchChar('/') {
				goto l411
			}
			{
				position412, thunkPosition412 := position, thunkPosition
				if !matchString("thead") {
					goto l413
				}
				goto l412
			l413:
				position, thunkPosition = position412, thunkPosition412
				if !matchString("THEAD") {
					goto l411
				}
			}
		l412:
			if !p.rules[ruleSpnl]() {
				goto l411
			}
			if !matchChar('>') {
				goto l411
			}
			return true
		l411:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 91 HtmlBlockOpenTr <- ('<' Spnl ('tr' / 'TR') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l414
			}
			if !p.rules[ruleSpnl]() {
				goto l414
			}
			{
				position415, thunkPosition415 := position, thunkPosition
				if !matchString("tr") {
					goto l416
				}
				goto l415
			l416:
				position, thunkPosition = position415, thunkPosition415
				if !matchString("TR") {
					goto l414
				}
			}
		l415:
			if !p.rules[ruleSpnl]() {
				goto l414
			}
		l417:
			{
				position418, thunkPosition418 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l418
				}
				goto l417
			l418:
				position, thunkPosition = position418, thunkPosition418
			}
			if !matchChar('>') {
				goto l414
			}
			return true
		l414:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 92 HtmlBlockCloseTr <- ('<' Spnl '/' ('tr' / 'TR') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l419
			}
			if !p.rules[ruleSpnl]() {
				goto l419
			}
			if !matchChar('/') {
				goto l419
			}
			{
				position420, thunkPosition420 := position, thunkPosition
				if !matchString("tr") {
					goto l421
				}
				goto l420
			l421:
				position, thunkPosition = position420, thunkPosition420
				if !matchString("TR") {
					goto l419
				}
			}
		l420:
			if !p.rules[ruleSpnl]() {
				goto l419
			}
			if !matchChar('>') {
				goto l419
			}
			return true
		l419:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 93 HtmlBlockOpenScript <- ('<' Spnl ('script' / 'SCRIPT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l422
			}
			if !p.rules[ruleSpnl]() {
				goto l422
			}
			{
				position423, thunkPosition423 := position, thunkPosition
				if !matchString("script") {
					goto l424
				}
				goto l423
			l424:
				position, thunkPosition = position423, thunkPosition423
				if !matchString("SCRIPT") {
					goto l422
				}
			}
		l423:
			if !p.rules[ruleSpnl]() {
				goto l422
			}
		l425:
			{
				position426, thunkPosition426 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l426
				}
				goto l425
			l426:
				position, thunkPosition = position426, thunkPosition426
			}
			if !matchChar('>') {
				goto l422
			}
			return true
		l422:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 94 HtmlBlockCloseScript <- ('<' Spnl '/' ('script' / 'SCRIPT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l427
			}
			if !p.rules[ruleSpnl]() {
				goto l427
			}
			if !matchChar('/') {
				goto l427
			}
			{
				position428, thunkPosition428 := position, thunkPosition
				if !matchString("script") {
					goto l429
				}
				goto l428
			l429:
				position, thunkPosition = position428, thunkPosition428
				if !matchString("SCRIPT") {
					goto l427
				}
			}
		l428:
			if !p.rules[ruleSpnl]() {
				goto l427
			}
			if !matchChar('>') {
				goto l427
			}
			return true
		l427:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 95 HtmlBlockInTags <- ((HtmlBlockOpenAddress (HtmlBlockInTags / (!HtmlBlockCloseAddress .))* HtmlBlockCloseAddress) / (HtmlBlockOpenBlockquote (HtmlBlockInTags / (!HtmlBlockCloseBlockquote .))* HtmlBlockCloseBlockquote) / (HtmlBlockOpenCenter (HtmlBlockInTags / (!HtmlBlockCloseCenter .))* HtmlBlockCloseCenter) / (HtmlBlockOpenDir (HtmlBlockInTags / (!HtmlBlockCloseDir .))* HtmlBlockCloseDir) / (HtmlBlockOpenDiv (HtmlBlockInTags / (!HtmlBlockCloseDiv .))* HtmlBlockCloseDiv) / (HtmlBlockOpenDl (HtmlBlockInTags / (!HtmlBlockCloseDl .))* HtmlBlockCloseDl) / (HtmlBlockOpenFieldset (HtmlBlockInTags / (!HtmlBlockCloseFieldset .))* HtmlBlockCloseFieldset) / (HtmlBlockOpenForm (HtmlBlockInTags / (!HtmlBlockCloseForm .))* HtmlBlockCloseForm) / (HtmlBlockOpenH1 (HtmlBlockInTags / (!HtmlBlockCloseH1 .))* HtmlBlockCloseH1) / (HtmlBlockOpenH2 (HtmlBlockInTags / (!HtmlBlockCloseH2 .))* HtmlBlockCloseH2) / (HtmlBlockOpenH3 (HtmlBlockInTags / (!HtmlBlockCloseH3 .))* HtmlBlockCloseH3) / (HtmlBlockOpenH4 (HtmlBlockInTags / (!HtmlBlockCloseH4 .))* HtmlBlockCloseH4) / (HtmlBlockOpenH5 (HtmlBlockInTags / (!HtmlBlockCloseH5 .))* HtmlBlockCloseH5) / (HtmlBlockOpenH6 (HtmlBlockInTags / (!HtmlBlockCloseH6 .))* HtmlBlockCloseH6) / (HtmlBlockOpenMenu (HtmlBlockInTags / (!HtmlBlockCloseMenu .))* HtmlBlockCloseMenu) / (HtmlBlockOpenNoframes (HtmlBlockInTags / (!HtmlBlockCloseNoframes .))* HtmlBlockCloseNoframes) / (HtmlBlockOpenNoscript (HtmlBlockInTags / (!HtmlBlockCloseNoscript .))* HtmlBlockCloseNoscript) / (HtmlBlockOpenOl (HtmlBlockInTags / (!HtmlBlockCloseOl .))* HtmlBlockCloseOl) / (HtmlBlockOpenP (HtmlBlockInTags / (!HtmlBlockCloseP .))* HtmlBlockCloseP) / (HtmlBlockOpenPre (HtmlBlockInTags / (!HtmlBlockClosePre .))* HtmlBlockClosePre) / (HtmlBlockOpenTable (HtmlBlockInTags / (!HtmlBlockCloseTable .))* HtmlBlockCloseTable) / (HtmlBlockOpenUl (HtmlBlockInTags / (!HtmlBlockCloseUl .))* HtmlBlockCloseUl) / (HtmlBlockOpenDd (HtmlBlockInTags / (!HtmlBlockCloseDd .))* HtmlBlockCloseDd) / (HtmlBlockOpenDt (HtmlBlockInTags / (!HtmlBlockCloseDt .))* HtmlBlockCloseDt) / (HtmlBlockOpenFrameset (HtmlBlockInTags / (!HtmlBlockCloseFrameset .))* HtmlBlockCloseFrameset) / (HtmlBlockOpenLi (HtmlBlockInTags / (!HtmlBlockCloseLi .))* HtmlBlockCloseLi) / (HtmlBlockOpenTbody (HtmlBlockInTags / (!HtmlBlockCloseTbody .))* HtmlBlockCloseTbody) / (HtmlBlockOpenTd (HtmlBlockInTags / (!HtmlBlockCloseTd .))* HtmlBlockCloseTd) / (HtmlBlockOpenTfoot (HtmlBlockInTags / (!HtmlBlockCloseTfoot .))* HtmlBlockCloseTfoot) / (HtmlBlockOpenTh (HtmlBlockInTags / (!HtmlBlockCloseTh .))* HtmlBlockCloseTh) / (HtmlBlockOpenThead (HtmlBlockInTags / (!HtmlBlockCloseThead .))* HtmlBlockCloseThead) / (HtmlBlockOpenTr (HtmlBlockInTags / (!HtmlBlockCloseTr .))* HtmlBlockCloseTr) / (HtmlBlockOpenScript (HtmlBlockInTags / (!HtmlBlockCloseScript .))* HtmlBlockCloseScript)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position431, thunkPosition431 := position, thunkPosition
				if !p.rules[ruleHtmlBlockOpenAddress]() {
					goto l432
				}
			l433:
				{
					position434, thunkPosition434 := position, thunkPosition
					{
						position435, thunkPosition435 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l436
						}
						goto l435
					l436:
						position, thunkPosition = position435, thunkPosition435
						{
							position437, thunkPosition437 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseAddress]() {
								goto l437
							}
							goto l434
						l437:
							position, thunkPosition = position437, thunkPosition437
						}
						if !matchDot() {
							goto l434
						}
					}
				l435:
					goto l433
				l434:
					position, thunkPosition = position434, thunkPosition434
				}
				if !p.rules[ruleHtmlBlockCloseAddress]() {
					goto l432
				}
				goto l431
			l432:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenBlockquote]() {
					goto l438
				}
			l439:
				{
					position440, thunkPosition440 := position, thunkPosition
					{
						position441, thunkPosition441 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l442
						}
						goto l441
					l442:
						position, thunkPosition = position441, thunkPosition441
						{
							position443, thunkPosition443 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseBlockquote]() {
								goto l443
							}
							goto l440
						l443:
							position, thunkPosition = position443, thunkPosition443
						}
						if !matchDot() {
							goto l440
						}
					}
				l441:
					goto l439
				l440:
					position, thunkPosition = position440, thunkPosition440
				}
				if !p.rules[ruleHtmlBlockCloseBlockquote]() {
					goto l438
				}
				goto l431
			l438:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenCenter]() {
					goto l444
				}
			l445:
				{
					position446, thunkPosition446 := position, thunkPosition
					{
						position447, thunkPosition447 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l448
						}
						goto l447
					l448:
						position, thunkPosition = position447, thunkPosition447
						{
							position449, thunkPosition449 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseCenter]() {
								goto l449
							}
							goto l446
						l449:
							position, thunkPosition = position449, thunkPosition449
						}
						if !matchDot() {
							goto l446
						}
					}
				l447:
					goto l445
				l446:
					position, thunkPosition = position446, thunkPosition446
				}
				if !p.rules[ruleHtmlBlockCloseCenter]() {
					goto l444
				}
				goto l431
			l444:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenDir]() {
					goto l450
				}
			l451:
				{
					position452, thunkPosition452 := position, thunkPosition
					{
						position453, thunkPosition453 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l454
						}
						goto l453
					l454:
						position, thunkPosition = position453, thunkPosition453
						{
							position455, thunkPosition455 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseDir]() {
								goto l455
							}
							goto l452
						l455:
							position, thunkPosition = position455, thunkPosition455
						}
						if !matchDot() {
							goto l452
						}
					}
				l453:
					goto l451
				l452:
					position, thunkPosition = position452, thunkPosition452
				}
				if !p.rules[ruleHtmlBlockCloseDir]() {
					goto l450
				}
				goto l431
			l450:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenDiv]() {
					goto l456
				}
			l457:
				{
					position458, thunkPosition458 := position, thunkPosition
					{
						position459, thunkPosition459 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l460
						}
						goto l459
					l460:
						position, thunkPosition = position459, thunkPosition459
						{
							position461, thunkPosition461 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseDiv]() {
								goto l461
							}
							goto l458
						l461:
							position, thunkPosition = position461, thunkPosition461
						}
						if !matchDot() {
							goto l458
						}
					}
				l459:
					goto l457
				l458:
					position, thunkPosition = position458, thunkPosition458
				}
				if !p.rules[ruleHtmlBlockCloseDiv]() {
					goto l456
				}
				goto l431
			l456:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenDl]() {
					goto l462
				}
			l463:
				{
					position464, thunkPosition464 := position, thunkPosition
					{
						position465, thunkPosition465 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l466
						}
						goto l465
					l466:
						position, thunkPosition = position465, thunkPosition465
						{
							position467, thunkPosition467 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseDl]() {
								goto l467
							}
							goto l464
						l467:
							position, thunkPosition = position467, thunkPosition467
						}
						if !matchDot() {
							goto l464
						}
					}
				l465:
					goto l463
				l464:
					position, thunkPosition = position464, thunkPosition464
				}
				if !p.rules[ruleHtmlBlockCloseDl]() {
					goto l462
				}
				goto l431
			l462:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenFieldset]() {
					goto l468
				}
			l469:
				{
					position470, thunkPosition470 := position, thunkPosition
					{
						position471, thunkPosition471 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l472
						}
						goto l471
					l472:
						position, thunkPosition = position471, thunkPosition471
						{
							position473, thunkPosition473 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseFieldset]() {
								goto l473
							}
							goto l470
						l473:
							position, thunkPosition = position473, thunkPosition473
						}
						if !matchDot() {
							goto l470
						}
					}
				l471:
					goto l469
				l470:
					position, thunkPosition = position470, thunkPosition470
				}
				if !p.rules[ruleHtmlBlockCloseFieldset]() {
					goto l468
				}
				goto l431
			l468:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenForm]() {
					goto l474
				}
			l475:
				{
					position476, thunkPosition476 := position, thunkPosition
					{
						position477, thunkPosition477 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l478
						}
						goto l477
					l478:
						position, thunkPosition = position477, thunkPosition477
						{
							position479, thunkPosition479 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseForm]() {
								goto l479
							}
							goto l476
						l479:
							position, thunkPosition = position479, thunkPosition479
						}
						if !matchDot() {
							goto l476
						}
					}
				l477:
					goto l475
				l476:
					position, thunkPosition = position476, thunkPosition476
				}
				if !p.rules[ruleHtmlBlockCloseForm]() {
					goto l474
				}
				goto l431
			l474:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenH1]() {
					goto l480
				}
			l481:
				{
					position482, thunkPosition482 := position, thunkPosition
					{
						position483, thunkPosition483 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l484
						}
						goto l483
					l484:
						position, thunkPosition = position483, thunkPosition483
						{
							position485, thunkPosition485 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseH1]() {
								goto l485
							}
							goto l482
						l485:
							position, thunkPosition = position485, thunkPosition485
						}
						if !matchDot() {
							goto l482
						}
					}
				l483:
					goto l481
				l482:
					position, thunkPosition = position482, thunkPosition482
				}
				if !p.rules[ruleHtmlBlockCloseH1]() {
					goto l480
				}
				goto l431
			l480:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenH2]() {
					goto l486
				}
			l487:
				{
					position488, thunkPosition488 := position, thunkPosition
					{
						position489, thunkPosition489 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l490
						}
						goto l489
					l490:
						position, thunkPosition = position489, thunkPosition489
						{
							position491, thunkPosition491 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseH2]() {
								goto l491
							}
							goto l488
						l491:
							position, thunkPosition = position491, thunkPosition491
						}
						if !matchDot() {
							goto l488
						}
					}
				l489:
					goto l487
				l488:
					position, thunkPosition = position488, thunkPosition488
				}
				if !p.rules[ruleHtmlBlockCloseH2]() {
					goto l486
				}
				goto l431
			l486:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenH3]() {
					goto l492
				}
			l493:
				{
					position494, thunkPosition494 := position, thunkPosition
					{
						position495, thunkPosition495 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l496
						}
						goto l495
					l496:
						position, thunkPosition = position495, thunkPosition495
						{
							position497, thunkPosition497 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseH3]() {
								goto l497
							}
							goto l494
						l497:
							position, thunkPosition = position497, thunkPosition497
						}
						if !matchDot() {
							goto l494
						}
					}
				l495:
					goto l493
				l494:
					position, thunkPosition = position494, thunkPosition494
				}
				if !p.rules[ruleHtmlBlockCloseH3]() {
					goto l492
				}
				goto l431
			l492:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenH4]() {
					goto l498
				}
			l499:
				{
					position500, thunkPosition500 := position, thunkPosition
					{
						position501, thunkPosition501 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l502
						}
						goto l501
					l502:
						position, thunkPosition = position501, thunkPosition501
						{
							position503, thunkPosition503 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseH4]() {
								goto l503
							}
							goto l500
						l503:
							position, thunkPosition = position503, thunkPosition503
						}
						if !matchDot() {
							goto l500
						}
					}
				l501:
					goto l499
				l500:
					position, thunkPosition = position500, thunkPosition500
				}
				if !p.rules[ruleHtmlBlockCloseH4]() {
					goto l498
				}
				goto l431
			l498:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenH5]() {
					goto l504
				}
			l505:
				{
					position506, thunkPosition506 := position, thunkPosition
					{
						position507, thunkPosition507 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l508
						}
						goto l507
					l508:
						position, thunkPosition = position507, thunkPosition507
						{
							position509, thunkPosition509 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseH5]() {
								goto l509
							}
							goto l506
						l509:
							position, thunkPosition = position509, thunkPosition509
						}
						if !matchDot() {
							goto l506
						}
					}
				l507:
					goto l505
				l506:
					position, thunkPosition = position506, thunkPosition506
				}
				if !p.rules[ruleHtmlBlockCloseH5]() {
					goto l504
				}
				goto l431
			l504:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenH6]() {
					goto l510
				}
			l511:
				{
					position512, thunkPosition512 := position, thunkPosition
					{
						position513, thunkPosition513 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l514
						}
						goto l513
					l514:
						position, thunkPosition = position513, thunkPosition513
						{
							position515, thunkPosition515 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseH6]() {
								goto l515
							}
							goto l512
						l515:
							position, thunkPosition = position515, thunkPosition515
						}
						if !matchDot() {
							goto l512
						}
					}
				l513:
					goto l511
				l512:
					position, thunkPosition = position512, thunkPosition512
				}
				if !p.rules[ruleHtmlBlockCloseH6]() {
					goto l510
				}
				goto l431
			l510:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenMenu]() {
					goto l516
				}
			l517:
				{
					position518, thunkPosition518 := position, thunkPosition
					{
						position519, thunkPosition519 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l520
						}
						goto l519
					l520:
						position, thunkPosition = position519, thunkPosition519
						{
							position521, thunkPosition521 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseMenu]() {
								goto l521
							}
							goto l518
						l521:
							position, thunkPosition = position521, thunkPosition521
						}
						if !matchDot() {
							goto l518
						}
					}
				l519:
					goto l517
				l518:
					position, thunkPosition = position518, thunkPosition518
				}
				if !p.rules[ruleHtmlBlockCloseMenu]() {
					goto l516
				}
				goto l431
			l516:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenNoframes]() {
					goto l522
				}
			l523:
				{
					position524, thunkPosition524 := position, thunkPosition
					{
						position525, thunkPosition525 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l526
						}
						goto l525
					l526:
						position, thunkPosition = position525, thunkPosition525
						{
							position527, thunkPosition527 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseNoframes]() {
								goto l527
							}
							goto l524
						l527:
							position, thunkPosition = position527, thunkPosition527
						}
						if !matchDot() {
							goto l524
						}
					}
				l525:
					goto l523
				l524:
					position, thunkPosition = position524, thunkPosition524
				}
				if !p.rules[ruleHtmlBlockCloseNoframes]() {
					goto l522
				}
				goto l431
			l522:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenNoscript]() {
					goto l528
				}
			l529:
				{
					position530, thunkPosition530 := position, thunkPosition
					{
						position531, thunkPosition531 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l532
						}
						goto l531
					l532:
						position, thunkPosition = position531, thunkPosition531
						{
							position533, thunkPosition533 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseNoscript]() {
								goto l533
							}
							goto l530
						l533:
							position, thunkPosition = position533, thunkPosition533
						}
						if !matchDot() {
							goto l530
						}
					}
				l531:
					goto l529
				l530:
					position, thunkPosition = position530, thunkPosition530
				}
				if !p.rules[ruleHtmlBlockCloseNoscript]() {
					goto l528
				}
				goto l431
			l528:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenOl]() {
					goto l534
				}
			l535:
				{
					position536, thunkPosition536 := position, thunkPosition
					{
						position537, thunkPosition537 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l538
						}
						goto l537
					l538:
						position, thunkPosition = position537, thunkPosition537
						{
							position539, thunkPosition539 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseOl]() {
								goto l539
							}
							goto l536
						l539:
							position, thunkPosition = position539, thunkPosition539
						}
						if !matchDot() {
							goto l536
						}
					}
				l537:
					goto l535
				l536:
					position, thunkPosition = position536, thunkPosition536
				}
				if !p.rules[ruleHtmlBlockCloseOl]() {
					goto l534
				}
				goto l431
			l534:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenP]() {
					goto l540
				}
			l541:
				{
					position542, thunkPosition542 := position, thunkPosition
					{
						position543, thunkPosition543 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l544
						}
						goto l543
					l544:
						position, thunkPosition = position543, thunkPosition543
						{
							position545, thunkPosition545 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseP]() {
								goto l545
							}
							goto l542
						l545:
							position, thunkPosition = position545, thunkPosition545
						}
						if !matchDot() {
							goto l542
						}
					}
				l543:
					goto l541
				l542:
					position, thunkPosition = position542, thunkPosition542
				}
				if !p.rules[ruleHtmlBlockCloseP]() {
					goto l540
				}
				goto l431
			l540:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenPre]() {
					goto l546
				}
			l547:
				{
					position548, thunkPosition548 := position, thunkPosition
					{
						position549, thunkPosition549 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l550
						}
						goto l549
					l550:
						position, thunkPosition = position549, thunkPosition549
						{
							position551, thunkPosition551 := position, thunkPosition
							if !p.rules[ruleHtmlBlockClosePre]() {
								goto l551
							}
							goto l548
						l551:
							position, thunkPosition = position551, thunkPosition551
						}
						if !matchDot() {
							goto l548
						}
					}
				l549:
					goto l547
				l548:
					position, thunkPosition = position548, thunkPosition548
				}
				if !p.rules[ruleHtmlBlockClosePre]() {
					goto l546
				}
				goto l431
			l546:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenTable]() {
					goto l552
				}
			l553:
				{
					position554, thunkPosition554 := position, thunkPosition
					{
						position555, thunkPosition555 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l556
						}
						goto l555
					l556:
						position, thunkPosition = position555, thunkPosition555
						{
							position557, thunkPosition557 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseTable]() {
								goto l557
							}
							goto l554
						l557:
							position, thunkPosition = position557, thunkPosition557
						}
						if !matchDot() {
							goto l554
						}
					}
				l555:
					goto l553
				l554:
					position, thunkPosition = position554, thunkPosition554
				}
				if !p.rules[ruleHtmlBlockCloseTable]() {
					goto l552
				}
				goto l431
			l552:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenUl]() {
					goto l558
				}
			l559:
				{
					position560, thunkPosition560 := position, thunkPosition
					{
						position561, thunkPosition561 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l562
						}
						goto l561
					l562:
						position, thunkPosition = position561, thunkPosition561
						{
							position563, thunkPosition563 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseUl]() {
								goto l563
							}
							goto l560
						l563:
							position, thunkPosition = position563, thunkPosition563
						}
						if !matchDot() {
							goto l560
						}
					}
				l561:
					goto l559
				l560:
					position, thunkPosition = position560, thunkPosition560
				}
				if !p.rules[ruleHtmlBlockCloseUl]() {
					goto l558
				}
				goto l431
			l558:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenDd]() {
					goto l564
				}
			l565:
				{
					position566, thunkPosition566 := position, thunkPosition
					{
						position567, thunkPosition567 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l568
						}
						goto l567
					l568:
						position, thunkPosition = position567, thunkPosition567
						{
							position569, thunkPosition569 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseDd]() {
								goto l569
							}
							goto l566
						l569:
							position, thunkPosition = position569, thunkPosition569
						}
						if !matchDot() {
							goto l566
						}
					}
				l567:
					goto l565
				l566:
					position, thunkPosition = position566, thunkPosition566
				}
				if !p.rules[ruleHtmlBlockCloseDd]() {
					goto l564
				}
				goto l431
			l564:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenDt]() {
					goto l570
				}
			l571:
				{
					position572, thunkPosition572 := position, thunkPosition
					{
						position573, thunkPosition573 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l574
						}
						goto l573
					l574:
						position, thunkPosition = position573, thunkPosition573
						{
							position575, thunkPosition575 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseDt]() {
								goto l575
							}
							goto l572
						l575:
							position, thunkPosition = position575, thunkPosition575
						}
						if !matchDot() {
							goto l572
						}
					}
				l573:
					goto l571
				l572:
					position, thunkPosition = position572, thunkPosition572
				}
				if !p.rules[ruleHtmlBlockCloseDt]() {
					goto l570
				}
				goto l431
			l570:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenFrameset]() {
					goto l576
				}
			l577:
				{
					position578, thunkPosition578 := position, thunkPosition
					{
						position579, thunkPosition579 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l580
						}
						goto l579
					l580:
						position, thunkPosition = position579, thunkPosition579
						{
							position581, thunkPosition581 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseFrameset]() {
								goto l581
							}
							goto l578
						l581:
							position, thunkPosition = position581, thunkPosition581
						}
						if !matchDot() {
							goto l578
						}
					}
				l579:
					goto l577
				l578:
					position, thunkPosition = position578, thunkPosition578
				}
				if !p.rules[ruleHtmlBlockCloseFrameset]() {
					goto l576
				}
				goto l431
			l576:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenLi]() {
					goto l582
				}
			l583:
				{
					position584, thunkPosition584 := position, thunkPosition
					{
						position585, thunkPosition585 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l586
						}
						goto l585
					l586:
						position, thunkPosition = position585, thunkPosition585
						{
							position587, thunkPosition587 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseLi]() {
								goto l587
							}
							goto l584
						l587:
							position, thunkPosition = position587, thunkPosition587
						}
						if !matchDot() {
							goto l584
						}
					}
				l585:
					goto l583
				l584:
					position, thunkPosition = position584, thunkPosition584
				}
				if !p.rules[ruleHtmlBlockCloseLi]() {
					goto l582
				}
				goto l431
			l582:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenTbody]() {
					goto l588
				}
			l589:
				{
					position590, thunkPosition590 := position, thunkPosition
					{
						position591, thunkPosition591 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l592
						}
						goto l591
					l592:
						position, thunkPosition = position591, thunkPosition591
						{
							position593, thunkPosition593 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseTbody]() {
								goto l593
							}
							goto l590
						l593:
							position, thunkPosition = position593, thunkPosition593
						}
						if !matchDot() {
							goto l590
						}
					}
				l591:
					goto l589
				l590:
					position, thunkPosition = position590, thunkPosition590
				}
				if !p.rules[ruleHtmlBlockCloseTbody]() {
					goto l588
				}
				goto l431
			l588:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenTd]() {
					goto l594
				}
			l595:
				{
					position596, thunkPosition596 := position, thunkPosition
					{
						position597, thunkPosition597 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l598
						}
						goto l597
					l598:
						position, thunkPosition = position597, thunkPosition597
						{
							position599, thunkPosition599 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseTd]() {
								goto l599
							}
							goto l596
						l599:
							position, thunkPosition = position599, thunkPosition599
						}
						if !matchDot() {
							goto l596
						}
					}
				l597:
					goto l595
				l596:
					position, thunkPosition = position596, thunkPosition596
				}
				if !p.rules[ruleHtmlBlockCloseTd]() {
					goto l594
				}
				goto l431
			l594:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenTfoot]() {
					goto l600
				}
			l601:
				{
					position602, thunkPosition602 := position, thunkPosition
					{
						position603, thunkPosition603 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l604
						}
						goto l603
					l604:
						position, thunkPosition = position603, thunkPosition603
						{
							position605, thunkPosition605 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseTfoot]() {
								goto l605
							}
							goto l602
						l605:
							position, thunkPosition = position605, thunkPosition605
						}
						if !matchDot() {
							goto l602
						}
					}
				l603:
					goto l601
				l602:
					position, thunkPosition = position602, thunkPosition602
				}
				if !p.rules[ruleHtmlBlockCloseTfoot]() {
					goto l600
				}
				goto l431
			l600:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenTh]() {
					goto l606
				}
			l607:
				{
					position608, thunkPosition608 := position, thunkPosition
					{
						position609, thunkPosition609 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l610
						}
						goto l609
					l610:
						position, thunkPosition = position609, thunkPosition609
						{
							position611, thunkPosition611 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseTh]() {
								goto l611
							}
							goto l608
						l611:
							position, thunkPosition = position611, thunkPosition611
						}
						if !matchDot() {
							goto l608
						}
					}
				l609:
					goto l607
				l608:
					position, thunkPosition = position608, thunkPosition608
				}
				if !p.rules[ruleHtmlBlockCloseTh]() {
					goto l606
				}
				goto l431
			l606:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenThead]() {
					goto l612
				}
			l613:
				{
					position614, thunkPosition614 := position, thunkPosition
					{
						position615, thunkPosition615 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l616
						}
						goto l615
					l616:
						position, thunkPosition = position615, thunkPosition615
						{
							position617, thunkPosition617 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseThead]() {
								goto l617
							}
							goto l614
						l617:
							position, thunkPosition = position617, thunkPosition617
						}
						if !matchDot() {
							goto l614
						}
					}
				l615:
					goto l613
				l614:
					position, thunkPosition = position614, thunkPosition614
				}
				if !p.rules[ruleHtmlBlockCloseThead]() {
					goto l612
				}
				goto l431
			l612:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenTr]() {
					goto l618
				}
			l619:
				{
					position620, thunkPosition620 := position, thunkPosition
					{
						position621, thunkPosition621 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l622
						}
						goto l621
					l622:
						position, thunkPosition = position621, thunkPosition621
						{
							position623, thunkPosition623 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseTr]() {
								goto l623
							}
							goto l620
						l623:
							position, thunkPosition = position623, thunkPosition623
						}
						if !matchDot() {
							goto l620
						}
					}
				l621:
					goto l619
				l620:
					position, thunkPosition = position620, thunkPosition620
				}
				if !p.rules[ruleHtmlBlockCloseTr]() {
					goto l618
				}
				goto l431
			l618:
				position, thunkPosition = position431, thunkPosition431
				if !p.rules[ruleHtmlBlockOpenScript]() {
					goto l430
				}
			l624:
				{
					position625, thunkPosition625 := position, thunkPosition
					{
						position626, thunkPosition626 := position, thunkPosition
						if !p.rules[ruleHtmlBlockInTags]() {
							goto l627
						}
						goto l626
					l627:
						position, thunkPosition = position626, thunkPosition626
						{
							position628, thunkPosition628 := position, thunkPosition
							if !p.rules[ruleHtmlBlockCloseScript]() {
								goto l628
							}
							goto l625
						l628:
							position, thunkPosition = position628, thunkPosition628
						}
						if !matchDot() {
							goto l625
						}
					}
				l626:
					goto l624
				l625:
					position, thunkPosition = position625, thunkPosition625
				}
				if !p.rules[ruleHtmlBlockCloseScript]() {
					goto l430
				}
			}
		l431:
			return true
		l430:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 96 HtmlBlock <- (&'<' < (HtmlBlockInTags / HtmlComment / HtmlBlockSelfClosing) > BlankLine+ {   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTMLBLOCK
                }
            }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !peekChar('<') {
				goto l629
			}
			begin = position
			{
				position630, thunkPosition630 := position, thunkPosition
				if !p.rules[ruleHtmlBlockInTags]() {
					goto l631
				}
				goto l630
			l631:
				position, thunkPosition = position630, thunkPosition630
				if !p.rules[ruleHtmlComment]() {
					goto l632
				}
				goto l630
			l632:
				position, thunkPosition = position630, thunkPosition630
				if !p.rules[ruleHtmlBlockSelfClosing]() {
					goto l629
				}
			}
		l630:
			end = position
			if !p.rules[ruleBlankLine]() {
				goto l629
			}
		l633:
			{
				position634, thunkPosition634 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l634
				}
				goto l633
			l634:
				position, thunkPosition = position634, thunkPosition634
			}
			do(37)
			return true
		l629:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 97 HtmlBlockSelfClosing <- ('<' Spnl HtmlBlockType Spnl HtmlAttribute* '/' Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l635
			}
			if !p.rules[ruleSpnl]() {
				goto l635
			}
			if !p.rules[ruleHtmlBlockType]() {
				goto l635
			}
			if !p.rules[ruleSpnl]() {
				goto l635
			}
		l636:
			{
				position637, thunkPosition637 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l637
				}
				goto l636
			l637:
				position, thunkPosition = position637, thunkPosition637
			}
			if !matchChar('/') {
				goto l635
			}
			if !p.rules[ruleSpnl]() {
				goto l635
			}
			if !matchChar('>') {
				goto l635
			}
			return true
		l635:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 98 HtmlBlockType <- ('address' / 'blockquote' / 'center' / 'dir' / 'div' / 'dl' / 'fieldset' / 'form' / 'h1' / 'h2' / 'h3' / 'h4' / 'h5' / 'h6' / 'hr' / 'isindex' / 'menu' / 'noframes' / 'noscript' / 'ol' / 'p' / 'pre' / 'table' / 'ul' / 'dd' / 'dt' / 'frameset' / 'li' / 'tbody' / 'td' / 'tfoot' / 'th' / 'thead' / 'tr' / 'script' / 'ADDRESS' / 'BLOCKQUOTE' / 'CENTER' / 'DIR' / 'DIV' / 'DL' / 'FIELDSET' / 'FORM' / 'H1' / 'H2' / 'H3' / 'H4' / 'H5' / 'H6' / 'HR' / 'ISINDEX' / 'MENU' / 'NOFRAMES' / 'NOSCRIPT' / 'OL' / 'P' / 'PRE' / 'TABLE' / 'UL' / 'DD' / 'DT' / 'FRAMESET' / 'LI' / 'TBODY' / 'TD' / 'TFOOT' / 'TH' / 'THEAD' / 'TR' / 'SCRIPT') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position639, thunkPosition639 := position, thunkPosition
				if !matchString("address") {
					goto l640
				}
				goto l639
			l640:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("blockquote") {
					goto l641
				}
				goto l639
			l641:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("center") {
					goto l642
				}
				goto l639
			l642:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("dir") {
					goto l643
				}
				goto l639
			l643:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("div") {
					goto l644
				}
				goto l639
			l644:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("dl") {
					goto l645
				}
				goto l639
			l645:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("fieldset") {
					goto l646
				}
				goto l639
			l646:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("form") {
					goto l647
				}
				goto l639
			l647:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("h1") {
					goto l648
				}
				goto l639
			l648:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("h2") {
					goto l649
				}
				goto l639
			l649:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("h3") {
					goto l650
				}
				goto l639
			l650:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("h4") {
					goto l651
				}
				goto l639
			l651:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("h5") {
					goto l652
				}
				goto l639
			l652:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("h6") {
					goto l653
				}
				goto l639
			l653:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("hr") {
					goto l654
				}
				goto l639
			l654:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("isindex") {
					goto l655
				}
				goto l639
			l655:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("menu") {
					goto l656
				}
				goto l639
			l656:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("noframes") {
					goto l657
				}
				goto l639
			l657:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("noscript") {
					goto l658
				}
				goto l639
			l658:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("ol") {
					goto l659
				}
				goto l639
			l659:
				position, thunkPosition = position639, thunkPosition639
				if !matchChar('p') {
					goto l660
				}
				goto l639
			l660:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("pre") {
					goto l661
				}
				goto l639
			l661:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("table") {
					goto l662
				}
				goto l639
			l662:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("ul") {
					goto l663
				}
				goto l639
			l663:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("dd") {
					goto l664
				}
				goto l639
			l664:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("dt") {
					goto l665
				}
				goto l639
			l665:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("frameset") {
					goto l666
				}
				goto l639
			l666:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("li") {
					goto l667
				}
				goto l639
			l667:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("tbody") {
					goto l668
				}
				goto l639
			l668:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("td") {
					goto l669
				}
				goto l639
			l669:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("tfoot") {
					goto l670
				}
				goto l639
			l670:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("th") {
					goto l671
				}
				goto l639
			l671:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("thead") {
					goto l672
				}
				goto l639
			l672:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("tr") {
					goto l673
				}
				goto l639
			l673:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("script") {
					goto l674
				}
				goto l639
			l674:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("ADDRESS") {
					goto l675
				}
				goto l639
			l675:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("BLOCKQUOTE") {
					goto l676
				}
				goto l639
			l676:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("CENTER") {
					goto l677
				}
				goto l639
			l677:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("DIR") {
					goto l678
				}
				goto l639
			l678:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("DIV") {
					goto l679
				}
				goto l639
			l679:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("DL") {
					goto l680
				}
				goto l639
			l680:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("FIELDSET") {
					goto l681
				}
				goto l639
			l681:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("FORM") {
					goto l682
				}
				goto l639
			l682:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("H1") {
					goto l683
				}
				goto l639
			l683:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("H2") {
					goto l684
				}
				goto l639
			l684:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("H3") {
					goto l685
				}
				goto l639
			l685:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("H4") {
					goto l686
				}
				goto l639
			l686:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("H5") {
					goto l687
				}
				goto l639
			l687:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("H6") {
					goto l688
				}
				goto l639
			l688:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("HR") {
					goto l689
				}
				goto l639
			l689:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("ISINDEX") {
					goto l690
				}
				goto l639
			l690:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("MENU") {
					goto l691
				}
				goto l639
			l691:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("NOFRAMES") {
					goto l692
				}
				goto l639
			l692:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("NOSCRIPT") {
					goto l693
				}
				goto l639
			l693:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("OL") {
					goto l694
				}
				goto l639
			l694:
				position, thunkPosition = position639, thunkPosition639
				if !matchChar('P') {
					goto l695
				}
				goto l639
			l695:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("PRE") {
					goto l696
				}
				goto l639
			l696:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("TABLE") {
					goto l697
				}
				goto l639
			l697:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("UL") {
					goto l698
				}
				goto l639
			l698:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("DD") {
					goto l699
				}
				goto l639
			l699:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("DT") {
					goto l700
				}
				goto l639
			l700:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("FRAMESET") {
					goto l701
				}
				goto l639
			l701:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("LI") {
					goto l702
				}
				goto l639
			l702:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("TBODY") {
					goto l703
				}
				goto l639
			l703:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("TD") {
					goto l704
				}
				goto l639
			l704:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("TFOOT") {
					goto l705
				}
				goto l639
			l705:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("TH") {
					goto l706
				}
				goto l639
			l706:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("THEAD") {
					goto l707
				}
				goto l639
			l707:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("TR") {
					goto l708
				}
				goto l639
			l708:
				position, thunkPosition = position639, thunkPosition639
				if !matchString("SCRIPT") {
					goto l638
				}
			}
		l639:
			return true
		l638:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 99 StyleOpen <- ('<' Spnl ('style' / 'STYLE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l709
			}
			if !p.rules[ruleSpnl]() {
				goto l709
			}
			{
				position710, thunkPosition710 := position, thunkPosition
				if !matchString("style") {
					goto l711
				}
				goto l710
			l711:
				position, thunkPosition = position710, thunkPosition710
				if !matchString("STYLE") {
					goto l709
				}
			}
		l710:
			if !p.rules[ruleSpnl]() {
				goto l709
			}
		l712:
			{
				position713, thunkPosition713 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l713
				}
				goto l712
			l713:
				position, thunkPosition = position713, thunkPosition713
			}
			if !matchChar('>') {
				goto l709
			}
			return true
		l709:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 100 StyleClose <- ('<' Spnl '/' ('style' / 'STYLE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l714
			}
			if !p.rules[ruleSpnl]() {
				goto l714
			}
			if !matchChar('/') {
				goto l714
			}
			{
				position715, thunkPosition715 := position, thunkPosition
				if !matchString("style") {
					goto l716
				}
				goto l715
			l716:
				position, thunkPosition = position715, thunkPosition715
				if !matchString("STYLE") {
					goto l714
				}
			}
		l715:
			if !p.rules[ruleSpnl]() {
				goto l714
			}
			if !matchChar('>') {
				goto l714
			}
			return true
		l714:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 101 InStyleTags <- (StyleOpen (!StyleClose .)* StyleClose) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleStyleOpen]() {
				goto l717
			}
		l718:
			{
				position719, thunkPosition719 := position, thunkPosition
				{
					position720, thunkPosition720 := position, thunkPosition
					if !p.rules[ruleStyleClose]() {
						goto l720
					}
					goto l719
				l720:
					position, thunkPosition = position720, thunkPosition720
				}
				if !matchDot() {
					goto l719
				}
				goto l718
			l719:
				position, thunkPosition = position719, thunkPosition719
			}
			if !p.rules[ruleStyleClose]() {
				goto l717
			}
			return true
		l717:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 102 StyleBlock <- (< InStyleTags > BlankLine* {   if p.extension.FilterStyles {
                        yy = mk_list(LIST, nil)
                    } else {
                        yy = mk_str(yytext)
                        yy.key = HTMLBLOCK
                    }
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleInStyleTags]() {
				goto l721
			}
			end = position
		l722:
			{
				position723, thunkPosition723 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l723
				}
				goto l722
			l723:
				position, thunkPosition = position723, thunkPosition723
			}
			do(38)
			return true
		l721:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 103 Inlines <- (StartList ((!Endline Inline { a = cons(yy, a) }) / (Endline &Inline { a = cons(c, a) }))+ Endline? { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l724
			}
			doarg(yySet, -2)
			{
				position727, thunkPosition727 := position, thunkPosition
				{
					position729, thunkPosition729 := position, thunkPosition
					if !p.rules[ruleEndline]() {
						goto l729
					}
					goto l728
				l729:
					position, thunkPosition = position729, thunkPosition729
				}
				if !p.rules[ruleInline]() {
					goto l728
				}
				do(39)
				goto l727
			l728:
				position, thunkPosition = position727, thunkPosition727
				if !p.rules[ruleEndline]() {
					goto l724
				}
				doarg(yySet, -1)
				{
					position730, thunkPosition730 := position, thunkPosition
					if !p.rules[ruleInline]() {
						goto l724
					}
					position, thunkPosition = position730, thunkPosition730
				}
				do(40)
			}
		l727:
		l725:
			{
				position726, thunkPosition726 := position, thunkPosition
				{
					position731, thunkPosition731 := position, thunkPosition
					{
						position733, thunkPosition733 := position, thunkPosition
						if !p.rules[ruleEndline]() {
							goto l733
						}
						goto l732
					l733:
						position, thunkPosition = position733, thunkPosition733
					}
					if !p.rules[ruleInline]() {
						goto l732
					}
					do(39)
					goto l731
				l732:
					position, thunkPosition = position731, thunkPosition731
					if !p.rules[ruleEndline]() {
						goto l726
					}
					doarg(yySet, -1)
					{
						position734, thunkPosition734 := position, thunkPosition
						if !p.rules[ruleInline]() {
							goto l726
						}
						position, thunkPosition = position734, thunkPosition734
					}
					do(40)
				}
			l731:
				goto l725
			l726:
				position, thunkPosition = position726, thunkPosition726
			}
			{
				position735, thunkPosition735 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l735
				}
				goto l736
			l735:
				position, thunkPosition = position735, thunkPosition735
			}
		l736:
			do(41)
			doarg(yyPop, 2)
			return true
		l724:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 104 Inline <- (Str / Endline / UlOrStarLine / Space / Strong / Emph / Image / Link / NoteReference / InlineNote / Code / RawHtml / Entity / EscapedChar / Smart / Symbol) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position738, thunkPosition738 := position, thunkPosition
				if !p.rules[ruleStr]() {
					goto l739
				}
				goto l738
			l739:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleEndline]() {
					goto l740
				}
				goto l738
			l740:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleUlOrStarLine]() {
					goto l741
				}
				goto l738
			l741:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleSpace]() {
					goto l742
				}
				goto l738
			l742:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleStrong]() {
					goto l743
				}
				goto l738
			l743:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleEmph]() {
					goto l744
				}
				goto l738
			l744:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleImage]() {
					goto l745
				}
				goto l738
			l745:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleLink]() {
					goto l746
				}
				goto l738
			l746:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleNoteReference]() {
					goto l747
				}
				goto l738
			l747:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleInlineNote]() {
					goto l748
				}
				goto l738
			l748:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleCode]() {
					goto l749
				}
				goto l738
			l749:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleRawHtml]() {
					goto l750
				}
				goto l738
			l750:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleEntity]() {
					goto l751
				}
				goto l738
			l751:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleEscapedChar]() {
					goto l752
				}
				goto l738
			l752:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleSmart]() {
					goto l753
				}
				goto l738
			l753:
				position, thunkPosition = position738, thunkPosition738
				if !p.rules[ruleSymbol]() {
					goto l737
				}
			}
		l738:
			return true
		l737:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 105 Space <- (Spacechar+ { yy = mk_str(" ")
          yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSpacechar]() {
				goto l754
			}
		l755:
			{
				position756, thunkPosition756 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l756
				}
				goto l755
			l756:
				position, thunkPosition = position756, thunkPosition756
			}
			do(42)
			return true
		l754:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 106 Str <- (< NormalChar (NormalChar / ('_'+ &Alphanumeric))* > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleNormalChar]() {
				goto l757
			}
		l758:
			{
				position759, thunkPosition759 := position, thunkPosition
				{
					position760, thunkPosition760 := position, thunkPosition
					if !p.rules[ruleNormalChar]() {
						goto l761
					}
					goto l760
				l761:
					position, thunkPosition = position760, thunkPosition760
					if !matchChar('_') {
						goto l759
					}
				l762:
					{
						position763, thunkPosition763 := position, thunkPosition
						if !matchChar('_') {
							goto l763
						}
						goto l762
					l763:
						position, thunkPosition = position763, thunkPosition763
					}
					{
						position764, thunkPosition764 := position, thunkPosition
						if !p.rules[ruleAlphanumeric]() {
							goto l759
						}
						position, thunkPosition = position764, thunkPosition764
					}
				}
			l760:
				goto l758
			l759:
				position, thunkPosition = position759, thunkPosition759
			}
			end = position
			do(43)
			return true
		l757:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 107 EscapedChar <- ('\\' !Newline < [-\\`|*_{}[\]()#+.!><] > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\\') {
				goto l765
			}
			{
				position766, thunkPosition766 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l766
				}
				goto l765
			l766:
				position, thunkPosition = position766, thunkPosition766
			}
			begin = position
			if !matchClass(1) {
				goto l765
			}
			end = position
			do(44)
			return true
		l765:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 108 Entity <- ((HexEntity / DecEntity / CharEntity) { yy = mk_str(yytext); yy.key = HTML }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position768, thunkPosition768 := position, thunkPosition
				if !p.rules[ruleHexEntity]() {
					goto l769
				}
				goto l768
			l769:
				position, thunkPosition = position768, thunkPosition768
				if !p.rules[ruleDecEntity]() {
					goto l770
				}
				goto l768
			l770:
				position, thunkPosition = position768, thunkPosition768
				if !p.rules[ruleCharEntity]() {
					goto l767
				}
			}
		l768:
			do(45)
			return true
		l767:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 109 Endline <- (LineBreak / TerminalEndline / NormalEndline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position772, thunkPosition772 := position, thunkPosition
				if !p.rules[ruleLineBreak]() {
					goto l773
				}
				goto l772
			l773:
				position, thunkPosition = position772, thunkPosition772
				if !p.rules[ruleTerminalEndline]() {
					goto l774
				}
				goto l772
			l774:
				position, thunkPosition = position772, thunkPosition772
				if !p.rules[ruleNormalEndline]() {
					goto l771
				}
			}
		l772:
			return true
		l771:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 110 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line (('===' '='*) / ('---' '-'*)) Newline) { yy = mk_str("\n")
                    yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l775
			}
			if !p.rules[ruleNewline]() {
				goto l775
			}
			{
				position776, thunkPosition776 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l776
				}
				goto l775
			l776:
				position, thunkPosition = position776, thunkPosition776
			}
			if peekChar('>') {
				goto l775
			}
			{
				position777, thunkPosition777 := position, thunkPosition
				if !p.rules[ruleAtxStart]() {
					goto l777
				}
				goto l775
			l777:
				position, thunkPosition = position777, thunkPosition777
			}
			{
				position778, thunkPosition778 := position, thunkPosition
				if !p.rules[ruleLine]() {
					goto l778
				}
				{
					position779, thunkPosition779 := position, thunkPosition
					if !matchString("===") {
						goto l780
					}
				l781:
					{
						position782, thunkPosition782 := position, thunkPosition
						if !matchChar('=') {
							goto l782
						}
						goto l781
					l782:
						position, thunkPosition = position782, thunkPosition782
					}
					goto l779
				l780:
					position, thunkPosition = position779, thunkPosition779
					if !matchString("---") {
						goto l778
					}
				l783:
					{
						position784, thunkPosition784 := position, thunkPosition
						if !matchChar('-') {
							goto l784
						}
						goto l783
					l784:
						position, thunkPosition = position784, thunkPosition784
					}
				}
			l779:
				if !p.rules[ruleNewline]() {
					goto l778
				}
				goto l775
			l778:
				position, thunkPosition = position778, thunkPosition778
			}
			do(46)
			return true
		l775:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 111 TerminalEndline <- (Sp Newline Eof { yy = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l785
			}
			if !p.rules[ruleNewline]() {
				goto l785
			}
			if !p.rules[ruleEof]() {
				goto l785
			}
			do(47)
			return true
		l785:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 112 LineBreak <- ('  ' NormalEndline { yy = mk_element(LINEBREAK) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("  ") {
				goto l786
			}
			if !p.rules[ruleNormalEndline]() {
				goto l786
			}
			do(48)
			return true
		l786:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 113 Symbol <- (< SpecialChar > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleSpecialChar]() {
				goto l787
			}
			end = position
			do(49)
			return true
		l787:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 114 UlOrStarLine <- ((UlLine / StarLine) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position789, thunkPosition789 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l790
				}
				goto l789
			l790:
				position, thunkPosition = position789, thunkPosition789
				if !p.rules[ruleStarLine]() {
					goto l788
				}
			}
		l789:
			do(50)
			return true
		l788:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 115 StarLine <- ((< '****' '*'* >) / (< Spacechar '*'+ &Spacechar >)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position792, thunkPosition792 := position, thunkPosition
				begin = position
				if !matchString("****") {
					goto l793
				}
			l794:
				{
					position795, thunkPosition795 := position, thunkPosition
					if !matchChar('*') {
						goto l795
					}
					goto l794
				l795:
					position, thunkPosition = position795, thunkPosition795
				}
				end = position
				goto l792
			l793:
				position, thunkPosition = position792, thunkPosition792
				begin = position
				if !p.rules[ruleSpacechar]() {
					goto l791
				}
				if !matchChar('*') {
					goto l791
				}
			l796:
				{
					position797, thunkPosition797 := position, thunkPosition
					if !matchChar('*') {
						goto l797
					}
					goto l796
				l797:
					position, thunkPosition = position797, thunkPosition797
				}
				{
					position798, thunkPosition798 := position, thunkPosition
					if !p.rules[ruleSpacechar]() {
						goto l791
					}
					position, thunkPosition = position798, thunkPosition798
				}
				end = position
			}
		l792:
			return true
		l791:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 116 UlLine <- ((< '____' '_'* >) / (< Spacechar '_'+ &Spacechar >)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position800, thunkPosition800 := position, thunkPosition
				begin = position
				if !matchString("____") {
					goto l801
				}
			l802:
				{
					position803, thunkPosition803 := position, thunkPosition
					if !matchChar('_') {
						goto l803
					}
					goto l802
				l803:
					position, thunkPosition = position803, thunkPosition803
				}
				end = position
				goto l800
			l801:
				position, thunkPosition = position800, thunkPosition800
				begin = position
				if !p.rules[ruleSpacechar]() {
					goto l799
				}
				if !matchChar('_') {
					goto l799
				}
			l804:
				{
					position805, thunkPosition805 := position, thunkPosition
					if !matchChar('_') {
						goto l805
					}
					goto l804
				l805:
					position, thunkPosition = position805, thunkPosition805
				}
				{
					position806, thunkPosition806 := position, thunkPosition
					if !p.rules[ruleSpacechar]() {
						goto l799
					}
					position, thunkPosition = position806, thunkPosition806
				}
				end = position
			}
		l800:
			return true
		l799:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 117 Emph <- (EmphStar / EmphUl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position808, thunkPosition808 := position, thunkPosition
				if !p.rules[ruleEmphStar]() {
					goto l809
				}
				goto l808
			l809:
				position, thunkPosition = position808, thunkPosition808
				if !p.rules[ruleEmphUl]() {
					goto l807
				}
			}
		l808:
			return true
		l807:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 118 OneStarOpen <- (!StarLine '*' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position811, thunkPosition811 := position, thunkPosition
				if !p.rules[ruleStarLine]() {
					goto l811
				}
				goto l810
			l811:
				position, thunkPosition = position811, thunkPosition811
			}
			if !matchChar('*') {
				goto l810
			}
			{
				position812, thunkPosition812 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l812
				}
				goto l810
			l812:
				position, thunkPosition = position812, thunkPosition812
			}
			{
				position813, thunkPosition813 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l813
				}
				goto l810
			l813:
				position, thunkPosition = position813, thunkPosition813
			}
			return true
		l810:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 119 OneStarClose <- (!Spacechar !Newline Inline !StrongStar '*' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position815, thunkPosition815 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l815
				}
				goto l814
			l815:
				position, thunkPosition = position815, thunkPosition815
			}
			{
				position816, thunkPosition816 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l816
				}
				goto l814
			l816:
				position, thunkPosition = position816, thunkPosition816
			}
			if !p.rules[ruleInline]() {
				goto l814
			}
			doarg(yySet, -1)
			{
				position817, thunkPosition817 := position, thunkPosition
				if !p.rules[ruleStrongStar]() {
					goto l817
				}
				goto l814
			l817:
				position, thunkPosition = position817, thunkPosition817
			}
			if !matchChar('*') {
				goto l814
			}
			do(51)
			doarg(yyPop, 1)
			return true
		l814:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 120 EmphStar <- (OneStarOpen StartList (!OneStarClose Inline { a = cons(yy, a) })* OneStarClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneStarOpen]() {
				goto l818
			}
			if !p.rules[ruleStartList]() {
				goto l818
			}
			doarg(yySet, -1)
		l819:
			{
				position820, thunkPosition820 := position, thunkPosition
				{
					position821, thunkPosition821 := position, thunkPosition
					if !p.rules[ruleOneStarClose]() {
						goto l821
					}
					goto l820
				l821:
					position, thunkPosition = position821, thunkPosition821
				}
				if !p.rules[ruleInline]() {
					goto l820
				}
				do(52)
				goto l819
			l820:
				position, thunkPosition = position820, thunkPosition820
			}
			if !p.rules[ruleOneStarClose]() {
				goto l818
			}
			do(53)
			do(54)
			doarg(yyPop, 1)
			return true
		l818:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 121 OneUlOpen <- (!UlLine '_' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position823, thunkPosition823 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l823
				}
				goto l822
			l823:
				position, thunkPosition = position823, thunkPosition823
			}
			if !matchChar('_') {
				goto l822
			}
			{
				position824, thunkPosition824 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l824
				}
				goto l822
			l824:
				position, thunkPosition = position824, thunkPosition824
			}
			{
				position825, thunkPosition825 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l825
				}
				goto l822
			l825:
				position, thunkPosition = position825, thunkPosition825
			}
			return true
		l822:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 122 OneUlClose <- (!Spacechar !Newline Inline !StrongUl '_' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position827, thunkPosition827 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l827
				}
				goto l826
			l827:
				position, thunkPosition = position827, thunkPosition827
			}
			{
				position828, thunkPosition828 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l828
				}
				goto l826
			l828:
				position, thunkPosition = position828, thunkPosition828
			}
			if !p.rules[ruleInline]() {
				goto l826
			}
			doarg(yySet, -1)
			{
				position829, thunkPosition829 := position, thunkPosition
				if !p.rules[ruleStrongUl]() {
					goto l829
				}
				goto l826
			l829:
				position, thunkPosition = position829, thunkPosition829
			}
			if !matchChar('_') {
				goto l826
			}
			{
				position830, thunkPosition830 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l830
				}
				goto l826
			l830:
				position, thunkPosition = position830, thunkPosition830
			}
			do(55)
			doarg(yyPop, 1)
			return true
		l826:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 123 EmphUl <- (OneUlOpen StartList (!OneUlClose Inline { a = cons(yy, a) })* OneUlClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneUlOpen]() {
				goto l831
			}
			if !p.rules[ruleStartList]() {
				goto l831
			}
			doarg(yySet, -1)
		l832:
			{
				position833, thunkPosition833 := position, thunkPosition
				{
					position834, thunkPosition834 := position, thunkPosition
					if !p.rules[ruleOneUlClose]() {
						goto l834
					}
					goto l833
				l834:
					position, thunkPosition = position834, thunkPosition834
				}
				if !p.rules[ruleInline]() {
					goto l833
				}
				do(56)
				goto l832
			l833:
				position, thunkPosition = position833, thunkPosition833
			}
			if !p.rules[ruleOneUlClose]() {
				goto l831
			}
			do(57)
			do(58)
			doarg(yyPop, 1)
			return true
		l831:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 124 Strong <- (StrongStar / StrongUl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position836, thunkPosition836 := position, thunkPosition
				if !p.rules[ruleStrongStar]() {
					goto l837
				}
				goto l836
			l837:
				position, thunkPosition = position836, thunkPosition836
				if !p.rules[ruleStrongUl]() {
					goto l835
				}
			}
		l836:
			return true
		l835:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 125 TwoStarOpen <- (!StarLine '**' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position839, thunkPosition839 := position, thunkPosition
				if !p.rules[ruleStarLine]() {
					goto l839
				}
				goto l838
			l839:
				position, thunkPosition = position839, thunkPosition839
			}
			if !matchString("**") {
				goto l838
			}
			{
				position840, thunkPosition840 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l840
				}
				goto l838
			l840:
				position, thunkPosition = position840, thunkPosition840
			}
			{
				position841, thunkPosition841 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l841
				}
				goto l838
			l841:
				position, thunkPosition = position841, thunkPosition841
			}
			return true
		l838:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 126 TwoStarClose <- (!Spacechar !Newline Inline '**' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position843, thunkPosition843 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l843
				}
				goto l842
			l843:
				position, thunkPosition = position843, thunkPosition843
			}
			{
				position844, thunkPosition844 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l844
				}
				goto l842
			l844:
				position, thunkPosition = position844, thunkPosition844
			}
			if !p.rules[ruleInline]() {
				goto l842
			}
			doarg(yySet, -1)
			if !matchString("**") {
				goto l842
			}
			do(59)
			doarg(yyPop, 1)
			return true
		l842:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 127 StrongStar <- (TwoStarOpen StartList (!TwoStarClose Inline { a = cons(yy, a) })* TwoStarClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoStarOpen]() {
				goto l845
			}
			if !p.rules[ruleStartList]() {
				goto l845
			}
			doarg(yySet, -1)
		l846:
			{
				position847, thunkPosition847 := position, thunkPosition
				{
					position848, thunkPosition848 := position, thunkPosition
					if !p.rules[ruleTwoStarClose]() {
						goto l848
					}
					goto l847
				l848:
					position, thunkPosition = position848, thunkPosition848
				}
				if !p.rules[ruleInline]() {
					goto l847
				}
				do(60)
				goto l846
			l847:
				position, thunkPosition = position847, thunkPosition847
			}
			if !p.rules[ruleTwoStarClose]() {
				goto l845
			}
			do(61)
			do(62)
			doarg(yyPop, 1)
			return true
		l845:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 128 TwoUlOpen <- (!UlLine '__' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position850, thunkPosition850 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l850
				}
				goto l849
			l850:
				position, thunkPosition = position850, thunkPosition850
			}
			if !matchString("__") {
				goto l849
			}
			{
				position851, thunkPosition851 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l851
				}
				goto l849
			l851:
				position, thunkPosition = position851, thunkPosition851
			}
			{
				position852, thunkPosition852 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l852
				}
				goto l849
			l852:
				position, thunkPosition = position852, thunkPosition852
			}
			return true
		l849:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 129 TwoUlClose <- (!Spacechar !Newline Inline '__' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position854, thunkPosition854 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l854
				}
				goto l853
			l854:
				position, thunkPosition = position854, thunkPosition854
			}
			{
				position855, thunkPosition855 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l855
				}
				goto l853
			l855:
				position, thunkPosition = position855, thunkPosition855
			}
			if !p.rules[ruleInline]() {
				goto l853
			}
			doarg(yySet, -1)
			if !matchString("__") {
				goto l853
			}
			{
				position856, thunkPosition856 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l856
				}
				goto l853
			l856:
				position, thunkPosition = position856, thunkPosition856
			}
			do(63)
			doarg(yyPop, 1)
			return true
		l853:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 130 StrongUl <- (TwoUlOpen StartList (!TwoUlClose Inline { a = cons(yy, a) })* TwoUlClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoUlOpen]() {
				goto l857
			}
			if !p.rules[ruleStartList]() {
				goto l857
			}
			doarg(yySet, -1)
		l858:
			{
				position859, thunkPosition859 := position, thunkPosition
				{
					position860, thunkPosition860 := position, thunkPosition
					if !p.rules[ruleTwoUlClose]() {
						goto l860
					}
					goto l859
				l860:
					position, thunkPosition = position860, thunkPosition860
				}
				if !p.rules[ruleInline]() {
					goto l859
				}
				do(64)
				goto l858
			l859:
				position, thunkPosition = position859, thunkPosition859
			}
			if !p.rules[ruleTwoUlClose]() {
				goto l857
			}
			do(65)
			do(66)
			doarg(yyPop, 1)
			return true
		l857:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 131 Image <- ('!' (ExplicitLink / ReferenceLink) { yy.key = IMAGE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('!') {
				goto l861
			}
			{
				position862, thunkPosition862 := position, thunkPosition
				if !p.rules[ruleExplicitLink]() {
					goto l863
				}
				goto l862
			l863:
				position, thunkPosition = position862, thunkPosition862
				if !p.rules[ruleReferenceLink]() {
					goto l861
				}
			}
		l862:
			do(67)
			return true
		l861:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 132 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position865, thunkPosition865 := position, thunkPosition
				if !p.rules[ruleExplicitLink]() {
					goto l866
				}
				goto l865
			l866:
				position, thunkPosition = position865, thunkPosition865
				if !p.rules[ruleReferenceLink]() {
					goto l867
				}
				goto l865
			l867:
				position, thunkPosition = position865, thunkPosition865
				if !p.rules[ruleAutoLink]() {
					goto l864
				}
			}
		l865:
			return true
		l864:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 133 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position869, thunkPosition869 := position, thunkPosition
				if !p.rules[ruleReferenceLinkDouble]() {
					goto l870
				}
				goto l869
			l870:
				position, thunkPosition = position869, thunkPosition869
				if !p.rules[ruleReferenceLinkSingle]() {
					goto l868
				}
			}
		l869:
			return true
		l868:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 134 ReferenceLinkDouble <- (Label < Spnl > !'[]' Label {
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
				goto l871
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto l871
			}
			end = position
			{
				position872, thunkPosition872 := position, thunkPosition
				if !matchString("[]") {
					goto l872
				}
				goto l871
			l872:
				position, thunkPosition = position872, thunkPosition872
			}
			if !p.rules[ruleLabel]() {
				goto l871
			}
			doarg(yySet, -2)
			do(68)
			doarg(yyPop, 2)
			return true
		l871:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 135 ReferenceLinkSingle <- (Label < (Spnl '[]')? > {
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
				goto l873
			}
			doarg(yySet, -1)
			begin = position
			{
				position874, thunkPosition874 := position, thunkPosition
				if !p.rules[ruleSpnl]() {
					goto l874
				}
				if !matchString("[]") {
					goto l874
				}
				goto l875
			l874:
				position, thunkPosition = position874, thunkPosition874
			}
		l875:
			end = position
			do(69)
			doarg(yyPop, 1)
			return true
		l873:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 136 ExplicitLink <- (Label Spnl '(' Sp Source Spnl Title Sp ')' { yy = mk_link(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleLabel]() {
				goto l876
			}
			doarg(yySet, -2)
			if !p.rules[ruleSpnl]() {
				goto l876
			}
			if !matchChar('(') {
				goto l876
			}
			if !p.rules[ruleSp]() {
				goto l876
			}
			if !p.rules[ruleSource]() {
				goto l876
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l876
			}
			if !p.rules[ruleTitle]() {
				goto l876
			}
			doarg(yySet, -3)
			if !p.rules[ruleSp]() {
				goto l876
			}
			if !matchChar(')') {
				goto l876
			}
			do(70)
			doarg(yyPop, 3)
			return true
		l876:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 137 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position878, thunkPosition878 := position, thunkPosition
				if !matchChar('<') {
					goto l879
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l879
				}
				end = position
				if !matchChar('>') {
					goto l879
				}
				goto l878
			l879:
				position, thunkPosition = position878, thunkPosition878
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l877
				}
				end = position
			}
		l878:
			do(71)
			return true
		l877:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 138 SourceContents <- (((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* / '') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position881, thunkPosition881 := position, thunkPosition
			l883:
				{
					position884, thunkPosition884 := position, thunkPosition
					{
						position885, thunkPosition885 := position, thunkPosition
						if peekChar('(') {
							goto l886
						}
						if peekChar(')') {
							goto l886
						}
						if peekChar('>') {
							goto l886
						}
						if !p.rules[ruleNonspacechar]() {
							goto l886
						}
					l887:
						{
							position888, thunkPosition888 := position, thunkPosition
							if peekChar('(') {
								goto l888
							}
							if peekChar(')') {
								goto l888
							}
							if peekChar('>') {
								goto l888
							}
							if !p.rules[ruleNonspacechar]() {
								goto l888
							}
							goto l887
						l888:
							position, thunkPosition = position888, thunkPosition888
						}
						goto l885
					l886:
						position, thunkPosition = position885, thunkPosition885
						if !matchChar('(') {
							goto l884
						}
						if !p.rules[ruleSourceContents]() {
							goto l884
						}
						if !matchChar(')') {
							goto l884
						}
					}
				l885:
					goto l883
				l884:
					position, thunkPosition = position884, thunkPosition884
				}
				goto l881
			l882:
				position, thunkPosition = position881, thunkPosition881
				if !matchString("") {
					goto l880
				}
			}
		l881:
			return true
		l880:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 139 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position890, thunkPosition890 := position, thunkPosition
				if !p.rules[ruleTitleSingle]() {
					goto l891
				}
				goto l890
			l891:
				position, thunkPosition = position890, thunkPosition890
				if !p.rules[ruleTitleDouble]() {
					goto l892
				}
				goto l890
			l892:
				position, thunkPosition = position890, thunkPosition890
				begin = position
				if !matchString("") {
					goto l889
				}
				end = position
			}
		l890:
			do(72)
			return true
		l889:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 140 TitleSingle <- ('\'' < (!('\'' Sp (')' / Newline)) .)* > '\'') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l893
			}
			begin = position
		l894:
			{
				position895, thunkPosition895 := position, thunkPosition
				{
					position896, thunkPosition896 := position, thunkPosition
					if !matchChar('\'') {
						goto l896
					}
					if !p.rules[ruleSp]() {
						goto l896
					}
					{
						position897, thunkPosition897 := position, thunkPosition
						if !matchChar(')') {
							goto l898
						}
						goto l897
					l898:
						position, thunkPosition = position897, thunkPosition897
						if !p.rules[ruleNewline]() {
							goto l896
						}
					}
				l897:
					goto l895
				l896:
					position, thunkPosition = position896, thunkPosition896
				}
				if !matchDot() {
					goto l895
				}
				goto l894
			l895:
				position, thunkPosition = position895, thunkPosition895
			}
			end = position
			if !matchChar('\'') {
				goto l893
			}
			return true
		l893:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 141 TitleDouble <- ('"' < (!('"' Sp (')' / Newline)) .)* > '"') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l899
			}
			begin = position
		l900:
			{
				position901, thunkPosition901 := position, thunkPosition
				{
					position902, thunkPosition902 := position, thunkPosition
					if !matchChar('"') {
						goto l902
					}
					if !p.rules[ruleSp]() {
						goto l902
					}
					{
						position903, thunkPosition903 := position, thunkPosition
						if !matchChar(')') {
							goto l904
						}
						goto l903
					l904:
						position, thunkPosition = position903, thunkPosition903
						if !p.rules[ruleNewline]() {
							goto l902
						}
					}
				l903:
					goto l901
				l902:
					position, thunkPosition = position902, thunkPosition902
				}
				if !matchDot() {
					goto l901
				}
				goto l900
			l901:
				position, thunkPosition = position901, thunkPosition901
			}
			end = position
			if !matchChar('"') {
				goto l899
			}
			return true
		l899:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 142 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position906, thunkPosition906 := position, thunkPosition
				if !p.rules[ruleAutoLinkUrl]() {
					goto l907
				}
				goto l906
			l907:
				position, thunkPosition = position906, thunkPosition906
				if !p.rules[ruleAutoLinkEmail]() {
					goto l905
				}
			}
		l906:
			return true
		l905:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 143 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = mk_link(mk_str(yytext), yytext, "") }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l908
			}
			begin = position
			if !matchClass(2) {
				goto l908
			}
		l909:
			{
				position910, thunkPosition910 := position, thunkPosition
				if !matchClass(2) {
					goto l910
				}
				goto l909
			l910:
				position, thunkPosition = position910, thunkPosition910
			}
			if !matchString("://") {
				goto l908
			}
			{
				position913, thunkPosition913 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l913
				}
				goto l908
			l913:
				position, thunkPosition = position913, thunkPosition913
			}
			if peekChar('>') {
				goto l908
			}
			if !matchDot() {
				goto l908
			}
		l911:
			{
				position912, thunkPosition912 := position, thunkPosition
				{
					position914, thunkPosition914 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l914
					}
					goto l912
				l914:
					position, thunkPosition = position914, thunkPosition914
				}
				if peekChar('>') {
					goto l912
				}
				if !matchDot() {
					goto l912
				}
				goto l911
			l912:
				position, thunkPosition = position912, thunkPosition912
			}
			end = position
			if !matchChar('>') {
				goto l908
			}
			do(73)
			return true
		l908:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 144 AutoLinkEmail <- ('<' < [-A-Za-z0-9+_]+ '@' (!Newline !'>' .)+ > '>' {
                    yy = mk_link(mk_str(yytext), "mailto:"+yytext, "")
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l915
			}
			begin = position
			if !matchClass(7) {
				goto l915
			}
		l916:
			{
				position917, thunkPosition917 := position, thunkPosition
				if !matchClass(7) {
					goto l917
				}
				goto l916
			l917:
				position, thunkPosition = position917, thunkPosition917
			}
			if !matchChar('@') {
				goto l915
			}
			{
				position920, thunkPosition920 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l920
				}
				goto l915
			l920:
				position, thunkPosition = position920, thunkPosition920
			}
			if peekChar('>') {
				goto l915
			}
			if !matchDot() {
				goto l915
			}
		l918:
			{
				position919, thunkPosition919 := position, thunkPosition
				{
					position921, thunkPosition921 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l921
					}
					goto l919
				l921:
					position, thunkPosition = position921, thunkPosition921
				}
				if peekChar('>') {
					goto l919
				}
				if !matchDot() {
					goto l919
				}
				goto l918
			l919:
				position, thunkPosition = position919, thunkPosition919
			}
			end = position
			if !matchChar('>') {
				goto l915
			}
			do(74)
			return true
		l915:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 145 Reference <- (NonindentSpace !'[]' Label ':' Spnl RefSrc Spnl RefTitle BlankLine* { yy = mk_link(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleNonindentSpace]() {
				goto l922
			}
			{
				position923, thunkPosition923 := position, thunkPosition
				if !matchString("[]") {
					goto l923
				}
				goto l922
			l923:
				position, thunkPosition = position923, thunkPosition923
			}
			if !p.rules[ruleLabel]() {
				goto l922
			}
			doarg(yySet, -2)
			if !matchChar(':') {
				goto l922
			}
			if !p.rules[ruleSpnl]() {
				goto l922
			}
			if !p.rules[ruleRefSrc]() {
				goto l922
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l922
			}
			if !p.rules[ruleRefTitle]() {
				goto l922
			}
			doarg(yySet, -3)
		l924:
			{
				position925, thunkPosition925 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l925
				}
				goto l924
			l925:
				position, thunkPosition = position925, thunkPosition925
			}
			do(75)
			doarg(yyPop, 3)
			return true
		l922:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 146 Label <- ('[' ((!'^' &{ p.extension.Notes }) / (&. &{ !p.extension.Notes })) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l926
			}
			{
				position927, thunkPosition927 := position, thunkPosition
				if peekChar('^') {
					goto l928
				}
				if !( p.extension.Notes ) {
					goto l928
				}
				goto l927
			l928:
				position, thunkPosition = position927, thunkPosition927
				if !peekDot() {
					goto l926
				}
				if !( !p.extension.Notes ) {
					goto l926
				}
			}
		l927:
			if !p.rules[ruleStartList]() {
				goto l926
			}
			doarg(yySet, -1)
		l929:
			{
				position930, thunkPosition930 := position, thunkPosition
				if peekChar(']') {
					goto l930
				}
				if !p.rules[ruleInline]() {
					goto l930
				}
				do(76)
				goto l929
			l930:
				position, thunkPosition = position930, thunkPosition930
			}
			if !matchChar(']') {
				goto l926
			}
			do(77)
			doarg(yyPop, 1)
			return true
		l926:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 147 RefSrc <- (< Nonspacechar+ > { yy = mk_str(yytext)
           yy.key = HTML }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto l931
			}
		l932:
			{
				position933, thunkPosition933 := position, thunkPosition
				if !p.rules[ruleNonspacechar]() {
					goto l933
				}
				goto l932
			l933:
				position, thunkPosition = position933, thunkPosition933
			}
			end = position
			do(78)
			return true
		l931:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 148 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position935, thunkPosition935 := position, thunkPosition
				if !p.rules[ruleRefTitleSingle]() {
					goto l936
				}
				goto l935
			l936:
				position, thunkPosition = position935, thunkPosition935
				if !p.rules[ruleRefTitleDouble]() {
					goto l937
				}
				goto l935
			l937:
				position, thunkPosition = position935, thunkPosition935
				if !p.rules[ruleRefTitleParens]() {
					goto l938
				}
				goto l935
			l938:
				position, thunkPosition = position935, thunkPosition935
				if !p.rules[ruleEmptyTitle]() {
					goto l934
				}
			}
		l935:
			do(79)
			return true
		l934:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 149 EmptyTitle <- (< '' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchString("") {
				goto l939
			}
			end = position
			return true
		l939:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 150 RefTitleSingle <- ('\'' < (!(('\'' Sp Newline) / Newline) .)* > '\'') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l940
			}
			begin = position
		l941:
			{
				position942, thunkPosition942 := position, thunkPosition
				{
					position943, thunkPosition943 := position, thunkPosition
					{
						position944, thunkPosition944 := position, thunkPosition
						if !matchChar('\'') {
							goto l945
						}
						if !p.rules[ruleSp]() {
							goto l945
						}
						if !p.rules[ruleNewline]() {
							goto l945
						}
						goto l944
					l945:
						position, thunkPosition = position944, thunkPosition944
						if !p.rules[ruleNewline]() {
							goto l943
						}
					}
				l944:
					goto l942
				l943:
					position, thunkPosition = position943, thunkPosition943
				}
				if !matchDot() {
					goto l942
				}
				goto l941
			l942:
				position, thunkPosition = position942, thunkPosition942
			}
			end = position
			if !matchChar('\'') {
				goto l940
			}
			return true
		l940:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 151 RefTitleDouble <- ('"' < (!(('"' Sp Newline) / Newline) .)* > '"') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l946
			}
			begin = position
		l947:
			{
				position948, thunkPosition948 := position, thunkPosition
				{
					position949, thunkPosition949 := position, thunkPosition
					{
						position950, thunkPosition950 := position, thunkPosition
						if !matchChar('"') {
							goto l951
						}
						if !p.rules[ruleSp]() {
							goto l951
						}
						if !p.rules[ruleNewline]() {
							goto l951
						}
						goto l950
					l951:
						position, thunkPosition = position950, thunkPosition950
						if !p.rules[ruleNewline]() {
							goto l949
						}
					}
				l950:
					goto l948
				l949:
					position, thunkPosition = position949, thunkPosition949
				}
				if !matchDot() {
					goto l948
				}
				goto l947
			l948:
				position, thunkPosition = position948, thunkPosition948
			}
			end = position
			if !matchChar('"') {
				goto l946
			}
			return true
		l946:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 152 RefTitleParens <- ('(' < (!((')' Sp Newline) / Newline) .)* > ')') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('(') {
				goto l952
			}
			begin = position
		l953:
			{
				position954, thunkPosition954 := position, thunkPosition
				{
					position955, thunkPosition955 := position, thunkPosition
					{
						position956, thunkPosition956 := position, thunkPosition
						if !matchChar(')') {
							goto l957
						}
						if !p.rules[ruleSp]() {
							goto l957
						}
						if !p.rules[ruleNewline]() {
							goto l957
						}
						goto l956
					l957:
						position, thunkPosition = position956, thunkPosition956
						if !p.rules[ruleNewline]() {
							goto l955
						}
					}
				l956:
					goto l954
				l955:
					position, thunkPosition = position955, thunkPosition955
				}
				if !matchDot() {
					goto l954
				}
				goto l953
			l954:
				position, thunkPosition = position954, thunkPosition954
			}
			end = position
			if !matchChar(')') {
				goto l952
			}
			return true
		l952:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 153 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l958
			}
			doarg(yySet, -1)
		l959:
			{
				position960, thunkPosition960 := position, thunkPosition
				{
					position961, thunkPosition961 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto l962
					}
					doarg(yySet, -2)
					do(80)
					goto l961
				l962:
					position, thunkPosition = position961, thunkPosition961
					if !p.rules[ruleSkipBlock]() {
						goto l960
					}
				}
			l961:
				goto l959
			l960:
				position, thunkPosition = position960, thunkPosition960
			}
			do(81)
			if !(commit(thunkPosition0)) {
				goto l958
			}
			doarg(yyPop, 2)
			return true
		l958:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 154 Ticks1 <- ('`' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('`') {
				goto l963
			}
			if peekChar('`') {
				goto l963
			}
			return true
		l963:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 155 Ticks2 <- ('``' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("``") {
				goto l964
			}
			if peekChar('`') {
				goto l964
			}
			return true
		l964:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 156 Ticks3 <- ('```' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("```") {
				goto l965
			}
			if peekChar('`') {
				goto l965
			}
			return true
		l965:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 157 Ticks4 <- ('````' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("````") {
				goto l966
			}
			if peekChar('`') {
				goto l966
			}
			return true
		l966:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 158 Ticks5 <- ('`````' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("`````") {
				goto l967
			}
			if peekChar('`') {
				goto l967
			}
			return true
		l967:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 159 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / (!Ticks1 '`'+) / (!(Sp Ticks1) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / (!Ticks2 '`'+) / (!(Sp Ticks2) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / (!Ticks3 '`'+) / (!(Sp Ticks3) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / (!Ticks4 '`'+) / (!(Sp Ticks4) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / (!Ticks5 '`'+) / (!(Sp Ticks5) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks5)) { yy = mk_str(yytext); yy.key = CODE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position969, thunkPosition969 := position, thunkPosition
				if !p.rules[ruleTicks1]() {
					goto l970
				}
				if !p.rules[ruleSp]() {
					goto l970
				}
				begin = position
				{
					position973, thunkPosition973 := position, thunkPosition
					if peekChar('`') {
						goto l974
					}
					if !p.rules[ruleNonspacechar]() {
						goto l974
					}
				l975:
					{
						position976, thunkPosition976 := position, thunkPosition
						if peekChar('`') {
							goto l976
						}
						if !p.rules[ruleNonspacechar]() {
							goto l976
						}
						goto l975
					l976:
						position, thunkPosition = position976, thunkPosition976
					}
					goto l973
				l974:
					position, thunkPosition = position973, thunkPosition973
					{
						position978, thunkPosition978 := position, thunkPosition
						if !p.rules[ruleTicks1]() {
							goto l978
						}
						goto l977
					l978:
						position, thunkPosition = position978, thunkPosition978
					}
					if !matchChar('`') {
						goto l977
					}
				l979:
					{
						position980, thunkPosition980 := position, thunkPosition
						if !matchChar('`') {
							goto l980
						}
						goto l979
					l980:
						position, thunkPosition = position980, thunkPosition980
					}
					goto l973
				l977:
					position, thunkPosition = position973, thunkPosition973
					{
						position981, thunkPosition981 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l981
						}
						if !p.rules[ruleTicks1]() {
							goto l981
						}
						goto l970
					l981:
						position, thunkPosition = position981, thunkPosition981
					}
					{
						position982, thunkPosition982 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l983
						}
						goto l982
					l983:
						position, thunkPosition = position982, thunkPosition982
						if !p.rules[ruleNewline]() {
							goto l970
						}
						{
							position984, thunkPosition984 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l984
							}
							goto l970
						l984:
							position, thunkPosition = position984, thunkPosition984
						}
					}
				l982:
				}
			l973:
			l971:
				{
					position972, thunkPosition972 := position, thunkPosition
					{
						position985, thunkPosition985 := position, thunkPosition
						if peekChar('`') {
							goto l986
						}
						if !p.rules[ruleNonspacechar]() {
							goto l986
						}
					l987:
						{
							position988, thunkPosition988 := position, thunkPosition
							if peekChar('`') {
								goto l988
							}
							if !p.rules[ruleNonspacechar]() {
								goto l988
							}
							goto l987
						l988:
							position, thunkPosition = position988, thunkPosition988
						}
						goto l985
					l986:
						position, thunkPosition = position985, thunkPosition985
						{
							position990, thunkPosition990 := position, thunkPosition
							if !p.rules[ruleTicks1]() {
								goto l990
							}
							goto l989
						l990:
							position, thunkPosition = position990, thunkPosition990
						}
						if !matchChar('`') {
							goto l989
						}
					l991:
						{
							position992, thunkPosition992 := position, thunkPosition
							if !matchChar('`') {
								goto l992
							}
							goto l991
						l992:
							position, thunkPosition = position992, thunkPosition992
						}
						goto l985
					l989:
						position, thunkPosition = position985, thunkPosition985
						{
							position993, thunkPosition993 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l993
							}
							if !p.rules[ruleTicks1]() {
								goto l993
							}
							goto l972
						l993:
							position, thunkPosition = position993, thunkPosition993
						}
						{
							position994, thunkPosition994 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l995
							}
							goto l994
						l995:
							position, thunkPosition = position994, thunkPosition994
							if !p.rules[ruleNewline]() {
								goto l972
							}
							{
								position996, thunkPosition996 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l996
								}
								goto l972
							l996:
								position, thunkPosition = position996, thunkPosition996
							}
						}
					l994:
					}
				l985:
					goto l971
				l972:
					position, thunkPosition = position972, thunkPosition972
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l970
				}
				if !p.rules[ruleTicks1]() {
					goto l970
				}
				goto l969
			l970:
				position, thunkPosition = position969, thunkPosition969
				if !p.rules[ruleTicks2]() {
					goto l997
				}
				if !p.rules[ruleSp]() {
					goto l997
				}
				begin = position
				{
					position1000, thunkPosition1000 := position, thunkPosition
					if peekChar('`') {
						goto l1001
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1001
					}
				l1002:
					{
						position1003, thunkPosition1003 := position, thunkPosition
						if peekChar('`') {
							goto l1003
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1003
						}
						goto l1002
					l1003:
						position, thunkPosition = position1003, thunkPosition1003
					}
					goto l1000
				l1001:
					position, thunkPosition = position1000, thunkPosition1000
					{
						position1005, thunkPosition1005 := position, thunkPosition
						if !p.rules[ruleTicks2]() {
							goto l1005
						}
						goto l1004
					l1005:
						position, thunkPosition = position1005, thunkPosition1005
					}
					if !matchChar('`') {
						goto l1004
					}
				l1006:
					{
						position1007, thunkPosition1007 := position, thunkPosition
						if !matchChar('`') {
							goto l1007
						}
						goto l1006
					l1007:
						position, thunkPosition = position1007, thunkPosition1007
					}
					goto l1000
				l1004:
					position, thunkPosition = position1000, thunkPosition1000
					{
						position1008, thunkPosition1008 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1008
						}
						if !p.rules[ruleTicks2]() {
							goto l1008
						}
						goto l997
					l1008:
						position, thunkPosition = position1008, thunkPosition1008
					}
					{
						position1009, thunkPosition1009 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1010
						}
						goto l1009
					l1010:
						position, thunkPosition = position1009, thunkPosition1009
						if !p.rules[ruleNewline]() {
							goto l997
						}
						{
							position1011, thunkPosition1011 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1011
							}
							goto l997
						l1011:
							position, thunkPosition = position1011, thunkPosition1011
						}
					}
				l1009:
				}
			l1000:
			l998:
				{
					position999, thunkPosition999 := position, thunkPosition
					{
						position1012, thunkPosition1012 := position, thunkPosition
						if peekChar('`') {
							goto l1013
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1013
						}
					l1014:
						{
							position1015, thunkPosition1015 := position, thunkPosition
							if peekChar('`') {
								goto l1015
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1015
							}
							goto l1014
						l1015:
							position, thunkPosition = position1015, thunkPosition1015
						}
						goto l1012
					l1013:
						position, thunkPosition = position1012, thunkPosition1012
						{
							position1017, thunkPosition1017 := position, thunkPosition
							if !p.rules[ruleTicks2]() {
								goto l1017
							}
							goto l1016
						l1017:
							position, thunkPosition = position1017, thunkPosition1017
						}
						if !matchChar('`') {
							goto l1016
						}
					l1018:
						{
							position1019, thunkPosition1019 := position, thunkPosition
							if !matchChar('`') {
								goto l1019
							}
							goto l1018
						l1019:
							position, thunkPosition = position1019, thunkPosition1019
						}
						goto l1012
					l1016:
						position, thunkPosition = position1012, thunkPosition1012
						{
							position1020, thunkPosition1020 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1020
							}
							if !p.rules[ruleTicks2]() {
								goto l1020
							}
							goto l999
						l1020:
							position, thunkPosition = position1020, thunkPosition1020
						}
						{
							position1021, thunkPosition1021 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1022
							}
							goto l1021
						l1022:
							position, thunkPosition = position1021, thunkPosition1021
							if !p.rules[ruleNewline]() {
								goto l999
							}
							{
								position1023, thunkPosition1023 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1023
								}
								goto l999
							l1023:
								position, thunkPosition = position1023, thunkPosition1023
							}
						}
					l1021:
					}
				l1012:
					goto l998
				l999:
					position, thunkPosition = position999, thunkPosition999
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l997
				}
				if !p.rules[ruleTicks2]() {
					goto l997
				}
				goto l969
			l997:
				position, thunkPosition = position969, thunkPosition969
				if !p.rules[ruleTicks3]() {
					goto l1024
				}
				if !p.rules[ruleSp]() {
					goto l1024
				}
				begin = position
				{
					position1027, thunkPosition1027 := position, thunkPosition
					if peekChar('`') {
						goto l1028
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1028
					}
				l1029:
					{
						position1030, thunkPosition1030 := position, thunkPosition
						if peekChar('`') {
							goto l1030
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1030
						}
						goto l1029
					l1030:
						position, thunkPosition = position1030, thunkPosition1030
					}
					goto l1027
				l1028:
					position, thunkPosition = position1027, thunkPosition1027
					{
						position1032, thunkPosition1032 := position, thunkPosition
						if !p.rules[ruleTicks3]() {
							goto l1032
						}
						goto l1031
					l1032:
						position, thunkPosition = position1032, thunkPosition1032
					}
					if !matchChar('`') {
						goto l1031
					}
				l1033:
					{
						position1034, thunkPosition1034 := position, thunkPosition
						if !matchChar('`') {
							goto l1034
						}
						goto l1033
					l1034:
						position, thunkPosition = position1034, thunkPosition1034
					}
					goto l1027
				l1031:
					position, thunkPosition = position1027, thunkPosition1027
					{
						position1035, thunkPosition1035 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1035
						}
						if !p.rules[ruleTicks3]() {
							goto l1035
						}
						goto l1024
					l1035:
						position, thunkPosition = position1035, thunkPosition1035
					}
					{
						position1036, thunkPosition1036 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1037
						}
						goto l1036
					l1037:
						position, thunkPosition = position1036, thunkPosition1036
						if !p.rules[ruleNewline]() {
							goto l1024
						}
						{
							position1038, thunkPosition1038 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1038
							}
							goto l1024
						l1038:
							position, thunkPosition = position1038, thunkPosition1038
						}
					}
				l1036:
				}
			l1027:
			l1025:
				{
					position1026, thunkPosition1026 := position, thunkPosition
					{
						position1039, thunkPosition1039 := position, thunkPosition
						if peekChar('`') {
							goto l1040
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1040
						}
					l1041:
						{
							position1042, thunkPosition1042 := position, thunkPosition
							if peekChar('`') {
								goto l1042
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1042
							}
							goto l1041
						l1042:
							position, thunkPosition = position1042, thunkPosition1042
						}
						goto l1039
					l1040:
						position, thunkPosition = position1039, thunkPosition1039
						{
							position1044, thunkPosition1044 := position, thunkPosition
							if !p.rules[ruleTicks3]() {
								goto l1044
							}
							goto l1043
						l1044:
							position, thunkPosition = position1044, thunkPosition1044
						}
						if !matchChar('`') {
							goto l1043
						}
					l1045:
						{
							position1046, thunkPosition1046 := position, thunkPosition
							if !matchChar('`') {
								goto l1046
							}
							goto l1045
						l1046:
							position, thunkPosition = position1046, thunkPosition1046
						}
						goto l1039
					l1043:
						position, thunkPosition = position1039, thunkPosition1039
						{
							position1047, thunkPosition1047 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1047
							}
							if !p.rules[ruleTicks3]() {
								goto l1047
							}
							goto l1026
						l1047:
							position, thunkPosition = position1047, thunkPosition1047
						}
						{
							position1048, thunkPosition1048 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1049
							}
							goto l1048
						l1049:
							position, thunkPosition = position1048, thunkPosition1048
							if !p.rules[ruleNewline]() {
								goto l1026
							}
							{
								position1050, thunkPosition1050 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1050
								}
								goto l1026
							l1050:
								position, thunkPosition = position1050, thunkPosition1050
							}
						}
					l1048:
					}
				l1039:
					goto l1025
				l1026:
					position, thunkPosition = position1026, thunkPosition1026
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1024
				}
				if !p.rules[ruleTicks3]() {
					goto l1024
				}
				goto l969
			l1024:
				position, thunkPosition = position969, thunkPosition969
				if !p.rules[ruleTicks4]() {
					goto l1051
				}
				if !p.rules[ruleSp]() {
					goto l1051
				}
				begin = position
				{
					position1054, thunkPosition1054 := position, thunkPosition
					if peekChar('`') {
						goto l1055
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1055
					}
				l1056:
					{
						position1057, thunkPosition1057 := position, thunkPosition
						if peekChar('`') {
							goto l1057
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1057
						}
						goto l1056
					l1057:
						position, thunkPosition = position1057, thunkPosition1057
					}
					goto l1054
				l1055:
					position, thunkPosition = position1054, thunkPosition1054
					{
						position1059, thunkPosition1059 := position, thunkPosition
						if !p.rules[ruleTicks4]() {
							goto l1059
						}
						goto l1058
					l1059:
						position, thunkPosition = position1059, thunkPosition1059
					}
					if !matchChar('`') {
						goto l1058
					}
				l1060:
					{
						position1061, thunkPosition1061 := position, thunkPosition
						if !matchChar('`') {
							goto l1061
						}
						goto l1060
					l1061:
						position, thunkPosition = position1061, thunkPosition1061
					}
					goto l1054
				l1058:
					position, thunkPosition = position1054, thunkPosition1054
					{
						position1062, thunkPosition1062 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1062
						}
						if !p.rules[ruleTicks4]() {
							goto l1062
						}
						goto l1051
					l1062:
						position, thunkPosition = position1062, thunkPosition1062
					}
					{
						position1063, thunkPosition1063 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1064
						}
						goto l1063
					l1064:
						position, thunkPosition = position1063, thunkPosition1063
						if !p.rules[ruleNewline]() {
							goto l1051
						}
						{
							position1065, thunkPosition1065 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1065
							}
							goto l1051
						l1065:
							position, thunkPosition = position1065, thunkPosition1065
						}
					}
				l1063:
				}
			l1054:
			l1052:
				{
					position1053, thunkPosition1053 := position, thunkPosition
					{
						position1066, thunkPosition1066 := position, thunkPosition
						if peekChar('`') {
							goto l1067
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1067
						}
					l1068:
						{
							position1069, thunkPosition1069 := position, thunkPosition
							if peekChar('`') {
								goto l1069
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1069
							}
							goto l1068
						l1069:
							position, thunkPosition = position1069, thunkPosition1069
						}
						goto l1066
					l1067:
						position, thunkPosition = position1066, thunkPosition1066
						{
							position1071, thunkPosition1071 := position, thunkPosition
							if !p.rules[ruleTicks4]() {
								goto l1071
							}
							goto l1070
						l1071:
							position, thunkPosition = position1071, thunkPosition1071
						}
						if !matchChar('`') {
							goto l1070
						}
					l1072:
						{
							position1073, thunkPosition1073 := position, thunkPosition
							if !matchChar('`') {
								goto l1073
							}
							goto l1072
						l1073:
							position, thunkPosition = position1073, thunkPosition1073
						}
						goto l1066
					l1070:
						position, thunkPosition = position1066, thunkPosition1066
						{
							position1074, thunkPosition1074 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1074
							}
							if !p.rules[ruleTicks4]() {
								goto l1074
							}
							goto l1053
						l1074:
							position, thunkPosition = position1074, thunkPosition1074
						}
						{
							position1075, thunkPosition1075 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1076
							}
							goto l1075
						l1076:
							position, thunkPosition = position1075, thunkPosition1075
							if !p.rules[ruleNewline]() {
								goto l1053
							}
							{
								position1077, thunkPosition1077 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1077
								}
								goto l1053
							l1077:
								position, thunkPosition = position1077, thunkPosition1077
							}
						}
					l1075:
					}
				l1066:
					goto l1052
				l1053:
					position, thunkPosition = position1053, thunkPosition1053
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1051
				}
				if !p.rules[ruleTicks4]() {
					goto l1051
				}
				goto l969
			l1051:
				position, thunkPosition = position969, thunkPosition969
				if !p.rules[ruleTicks5]() {
					goto l968
				}
				if !p.rules[ruleSp]() {
					goto l968
				}
				begin = position
				{
					position1080, thunkPosition1080 := position, thunkPosition
					if peekChar('`') {
						goto l1081
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1081
					}
				l1082:
					{
						position1083, thunkPosition1083 := position, thunkPosition
						if peekChar('`') {
							goto l1083
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1083
						}
						goto l1082
					l1083:
						position, thunkPosition = position1083, thunkPosition1083
					}
					goto l1080
				l1081:
					position, thunkPosition = position1080, thunkPosition1080
					{
						position1085, thunkPosition1085 := position, thunkPosition
						if !p.rules[ruleTicks5]() {
							goto l1085
						}
						goto l1084
					l1085:
						position, thunkPosition = position1085, thunkPosition1085
					}
					if !matchChar('`') {
						goto l1084
					}
				l1086:
					{
						position1087, thunkPosition1087 := position, thunkPosition
						if !matchChar('`') {
							goto l1087
						}
						goto l1086
					l1087:
						position, thunkPosition = position1087, thunkPosition1087
					}
					goto l1080
				l1084:
					position, thunkPosition = position1080, thunkPosition1080
					{
						position1088, thunkPosition1088 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1088
						}
						if !p.rules[ruleTicks5]() {
							goto l1088
						}
						goto l968
					l1088:
						position, thunkPosition = position1088, thunkPosition1088
					}
					{
						position1089, thunkPosition1089 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1090
						}
						goto l1089
					l1090:
						position, thunkPosition = position1089, thunkPosition1089
						if !p.rules[ruleNewline]() {
							goto l968
						}
						{
							position1091, thunkPosition1091 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1091
							}
							goto l968
						l1091:
							position, thunkPosition = position1091, thunkPosition1091
						}
					}
				l1089:
				}
			l1080:
			l1078:
				{
					position1079, thunkPosition1079 := position, thunkPosition
					{
						position1092, thunkPosition1092 := position, thunkPosition
						if peekChar('`') {
							goto l1093
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1093
						}
					l1094:
						{
							position1095, thunkPosition1095 := position, thunkPosition
							if peekChar('`') {
								goto l1095
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1095
							}
							goto l1094
						l1095:
							position, thunkPosition = position1095, thunkPosition1095
						}
						goto l1092
					l1093:
						position, thunkPosition = position1092, thunkPosition1092
						{
							position1097, thunkPosition1097 := position, thunkPosition
							if !p.rules[ruleTicks5]() {
								goto l1097
							}
							goto l1096
						l1097:
							position, thunkPosition = position1097, thunkPosition1097
						}
						if !matchChar('`') {
							goto l1096
						}
					l1098:
						{
							position1099, thunkPosition1099 := position, thunkPosition
							if !matchChar('`') {
								goto l1099
							}
							goto l1098
						l1099:
							position, thunkPosition = position1099, thunkPosition1099
						}
						goto l1092
					l1096:
						position, thunkPosition = position1092, thunkPosition1092
						{
							position1100, thunkPosition1100 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1100
							}
							if !p.rules[ruleTicks5]() {
								goto l1100
							}
							goto l1079
						l1100:
							position, thunkPosition = position1100, thunkPosition1100
						}
						{
							position1101, thunkPosition1101 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1102
							}
							goto l1101
						l1102:
							position, thunkPosition = position1101, thunkPosition1101
							if !p.rules[ruleNewline]() {
								goto l1079
							}
							{
								position1103, thunkPosition1103 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1103
								}
								goto l1079
							l1103:
								position, thunkPosition = position1103, thunkPosition1103
							}
						}
					l1101:
					}
				l1092:
					goto l1078
				l1079:
					position, thunkPosition = position1079, thunkPosition1079
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l968
				}
				if !p.rules[ruleTicks5]() {
					goto l968
				}
			}
		l969:
			do(82)
			return true
		l968:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 160 RawHtml <- (< (HtmlComment / HtmlTag) > {   if p.extension.FilterHTML {
                    yy = mk_list(LIST, nil)
                } else {
                    yy = mk_str(yytext)
                    yy.key = HTML
                }
            }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			{
				position1105, thunkPosition1105 := position, thunkPosition
				if !p.rules[ruleHtmlComment]() {
					goto l1106
				}
				goto l1105
			l1106:
				position, thunkPosition = position1105, thunkPosition1105
				if !p.rules[ruleHtmlTag]() {
					goto l1104
				}
			}
		l1105:
			end = position
			do(83)
			return true
		l1104:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 161 BlankLine <- (Sp Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l1107
			}
			if !p.rules[ruleNewline]() {
				goto l1107
			}
			return true
		l1107:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 162 Quoted <- (('"' (!'"' .)* '"') / ('\'' (!'\'' .)* '\'')) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1109, thunkPosition1109 := position, thunkPosition
				if !matchChar('"') {
					goto l1110
				}
			l1111:
				{
					position1112, thunkPosition1112 := position, thunkPosition
					if peekChar('"') {
						goto l1112
					}
					if !matchDot() {
						goto l1112
					}
					goto l1111
				l1112:
					position, thunkPosition = position1112, thunkPosition1112
				}
				if !matchChar('"') {
					goto l1110
				}
				goto l1109
			l1110:
				position, thunkPosition = position1109, thunkPosition1109
				if !matchChar('\'') {
					goto l1108
				}
			l1113:
				{
					position1114, thunkPosition1114 := position, thunkPosition
					if peekChar('\'') {
						goto l1114
					}
					if !matchDot() {
						goto l1114
					}
					goto l1113
				l1114:
					position, thunkPosition = position1114, thunkPosition1114
				}
				if !matchChar('\'') {
					goto l1108
				}
			}
		l1109:
			return true
		l1108:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 163 HtmlAttribute <- ((Alphanumeric / '-')+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1118, thunkPosition1118 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l1119
				}
				goto l1118
			l1119:
				position, thunkPosition = position1118, thunkPosition1118
				if !matchChar('-') {
					goto l1115
				}
			}
		l1118:
		l1116:
			{
				position1117, thunkPosition1117 := position, thunkPosition
				{
					position1120, thunkPosition1120 := position, thunkPosition
					if !p.rules[ruleAlphanumeric]() {
						goto l1121
					}
					goto l1120
				l1121:
					position, thunkPosition = position1120, thunkPosition1120
					if !matchChar('-') {
						goto l1117
					}
				}
			l1120:
				goto l1116
			l1117:
				position, thunkPosition = position1117, thunkPosition1117
			}
			if !p.rules[ruleSpnl]() {
				goto l1115
			}
			{
				position1122, thunkPosition1122 := position, thunkPosition
				if !matchChar('=') {
					goto l1122
				}
				if !p.rules[ruleSpnl]() {
					goto l1122
				}
				{
					position1124, thunkPosition1124 := position, thunkPosition
					if !p.rules[ruleQuoted]() {
						goto l1125
					}
					goto l1124
				l1125:
					position, thunkPosition = position1124, thunkPosition1124
					if peekChar('>') {
						goto l1122
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1122
					}
				l1126:
					{
						position1127, thunkPosition1127 := position, thunkPosition
						if peekChar('>') {
							goto l1127
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1127
						}
						goto l1126
					l1127:
						position, thunkPosition = position1127, thunkPosition1127
					}
				}
			l1124:
				goto l1123
			l1122:
				position, thunkPosition = position1122, thunkPosition1122
			}
		l1123:
			if !p.rules[ruleSpnl]() {
				goto l1115
			}
			return true
		l1115:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 164 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("<!--") {
				goto l1128
			}
		l1129:
			{
				position1130, thunkPosition1130 := position, thunkPosition
				{
					position1131, thunkPosition1131 := position, thunkPosition
					if !matchString("-->") {
						goto l1131
					}
					goto l1130
				l1131:
					position, thunkPosition = position1131, thunkPosition1131
				}
				if !matchDot() {
					goto l1130
				}
				goto l1129
			l1130:
				position, thunkPosition = position1130, thunkPosition1130
			}
			if !matchString("-->") {
				goto l1128
			}
			return true
		l1128:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 165 HtmlTag <- ('<' Spnl '/'? Alphanumeric+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l1132
			}
			if !p.rules[ruleSpnl]() {
				goto l1132
			}
			{
				position1133, thunkPosition1133 := position, thunkPosition
				if !matchChar('/') {
					goto l1133
				}
				goto l1134
			l1133:
				position, thunkPosition = position1133, thunkPosition1133
			}
		l1134:
			if !p.rules[ruleAlphanumeric]() {
				goto l1132
			}
		l1135:
			{
				position1136, thunkPosition1136 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l1136
				}
				goto l1135
			l1136:
				position, thunkPosition = position1136, thunkPosition1136
			}
			if !p.rules[ruleSpnl]() {
				goto l1132
			}
		l1137:
			{
				position1138, thunkPosition1138 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l1138
				}
				goto l1137
			l1138:
				position, thunkPosition = position1138, thunkPosition1138
			}
			{
				position1139, thunkPosition1139 := position, thunkPosition
				if !matchChar('/') {
					goto l1139
				}
				goto l1140
			l1139:
				position, thunkPosition = position1139, thunkPosition1139
			}
		l1140:
			if !p.rules[ruleSpnl]() {
				goto l1132
			}
			if !matchChar('>') {
				goto l1132
			}
			return true
		l1132:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 166 Eof <- !. */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if peekDot() {
				goto l1141
			}
			return true
		l1141:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 167 Spacechar <- (' ' / '\t') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1143, thunkPosition1143 := position, thunkPosition
				if !matchChar(' ') {
					goto l1144
				}
				goto l1143
			l1144:
				position, thunkPosition = position1143, thunkPosition1143
				if !matchChar('\t') {
					goto l1142
				}
			}
		l1143:
			return true
		l1142:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 168 Nonspacechar <- (!Spacechar !Newline .) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1146, thunkPosition1146 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1146
				}
				goto l1145
			l1146:
				position, thunkPosition = position1146, thunkPosition1146
			}
			{
				position1147, thunkPosition1147 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1147
				}
				goto l1145
			l1147:
				position, thunkPosition = position1147, thunkPosition1147
			}
			if !matchDot() {
				goto l1145
			}
			return true
		l1145:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 169 Newline <- ('\n' / ('\r' '\n'?)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1149, thunkPosition1149 := position, thunkPosition
				if !matchChar('\n') {
					goto l1150
				}
				goto l1149
			l1150:
				position, thunkPosition = position1149, thunkPosition1149
				if !matchChar('\r') {
					goto l1148
				}
				{
					position1151, thunkPosition1151 := position, thunkPosition
					if !matchChar('\n') {
						goto l1151
					}
					goto l1152
				l1151:
					position, thunkPosition = position1151, thunkPosition1151
				}
			l1152:
			}
		l1149:
			return true
		l1148:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 170 Sp <- Spacechar* */
		func() bool {
		l1154:
			{
				position1155, thunkPosition1155 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1155
				}
				goto l1154
			l1155:
				position, thunkPosition = position1155, thunkPosition1155
			}
			return true
		},
		/* 171 Spnl <- (Sp (Newline Sp)?) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l1156
			}
			{
				position1157, thunkPosition1157 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1157
				}
				if !p.rules[ruleSp]() {
					goto l1157
				}
				goto l1158
			l1157:
				position, thunkPosition = position1157, thunkPosition1157
			}
		l1158:
			return true
		l1156:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 172 SpecialChar <- ('*' / '_' / '`' / '&' / '[' / ']' / '<' / '!' / '\\' / ExtendedSpecialChar) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1160, thunkPosition1160 := position, thunkPosition
				if !matchChar('*') {
					goto l1161
				}
				goto l1160
			l1161:
				position, thunkPosition = position1160, thunkPosition1160
				if !matchChar('_') {
					goto l1162
				}
				goto l1160
			l1162:
				position, thunkPosition = position1160, thunkPosition1160
				if !matchChar('`') {
					goto l1163
				}
				goto l1160
			l1163:
				position, thunkPosition = position1160, thunkPosition1160
				if !matchChar('&') {
					goto l1164
				}
				goto l1160
			l1164:
				position, thunkPosition = position1160, thunkPosition1160
				if !matchChar('[') {
					goto l1165
				}
				goto l1160
			l1165:
				position, thunkPosition = position1160, thunkPosition1160
				if !matchChar(']') {
					goto l1166
				}
				goto l1160
			l1166:
				position, thunkPosition = position1160, thunkPosition1160
				if !matchChar('<') {
					goto l1167
				}
				goto l1160
			l1167:
				position, thunkPosition = position1160, thunkPosition1160
				if !matchChar('!') {
					goto l1168
				}
				goto l1160
			l1168:
				position, thunkPosition = position1160, thunkPosition1160
				if !matchChar('\\') {
					goto l1169
				}
				goto l1160
			l1169:
				position, thunkPosition = position1160, thunkPosition1160
				if !p.rules[ruleExtendedSpecialChar]() {
					goto l1159
				}
			}
		l1160:
			return true
		l1159:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 173 NormalChar <- (!(SpecialChar / Spacechar / Newline) .) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1171, thunkPosition1171 := position, thunkPosition
				{
					position1172, thunkPosition1172 := position, thunkPosition
					if !p.rules[ruleSpecialChar]() {
						goto l1173
					}
					goto l1172
				l1173:
					position, thunkPosition = position1172, thunkPosition1172
					if !p.rules[ruleSpacechar]() {
						goto l1174
					}
					goto l1172
				l1174:
					position, thunkPosition = position1172, thunkPosition1172
					if !p.rules[ruleNewline]() {
						goto l1171
					}
				}
			l1172:
				goto l1170
			l1171:
				position, thunkPosition = position1171, thunkPosition1171
			}
			if !matchDot() {
				goto l1170
			}
			return true
		l1170:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 174 Alphanumeric <- [A-Za-z0-9] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(6) {
				goto l1175
			}
			return true
		l1175:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 175 Digit <- [0-9] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(5) {
				goto l1176
			}
			return true
		l1176:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 176 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1177
			}
			if !matchChar('#') {
				goto l1177
			}
			if !matchClass(3) {
				goto l1177
			}
			if !matchClass(0) {
				goto l1177
			}
		l1178:
			{
				position1179, thunkPosition1179 := position, thunkPosition
				if !matchClass(0) {
					goto l1179
				}
				goto l1178
			l1179:
				position, thunkPosition = position1179, thunkPosition1179
			}
			if !matchChar(';') {
				goto l1177
			}
			end = position
			return true
		l1177:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 177 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1180
			}
			if !matchChar('#') {
				goto l1180
			}
			if !matchClass(5) {
				goto l1180
			}
		l1181:
			{
				position1182, thunkPosition1182 := position, thunkPosition
				if !matchClass(5) {
					goto l1182
				}
				goto l1181
			l1182:
				position, thunkPosition = position1182, thunkPosition1182
			}
			end = position
			if !matchChar(';') {
				goto l1180
			}
			end = position
			return true
		l1180:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 178 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1183
			}
			if !matchClass(6) {
				goto l1183
			}
		l1184:
			{
				position1185, thunkPosition1185 := position, thunkPosition
				if !matchClass(6) {
					goto l1185
				}
				goto l1184
			l1185:
				position, thunkPosition = position1185, thunkPosition1185
			}
			if !matchChar(';') {
				goto l1183
			}
			end = position
			return true
		l1183:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 179 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1187, thunkPosition1187 := position, thunkPosition
				if !matchString("   ") {
					goto l1188
				}
				goto l1187
			l1188:
				position, thunkPosition = position1187, thunkPosition1187
				if !matchString("  ") {
					goto l1189
				}
				goto l1187
			l1189:
				position, thunkPosition = position1187, thunkPosition1187
				if !matchChar(' ') {
					goto l1190
				}
				goto l1187
			l1190:
				position, thunkPosition = position1187, thunkPosition1187
				if !matchString("") {
					goto l1186
				}
			}
		l1187:
			return true
		l1186:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 180 Indent <- ('\t' / '    ') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1192, thunkPosition1192 := position, thunkPosition
				if !matchChar('\t') {
					goto l1193
				}
				goto l1192
			l1193:
				position, thunkPosition = position1192, thunkPosition1192
				if !matchString("    ") {
					goto l1191
				}
			}
		l1192:
			return true
		l1191:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 181 IndentedLine <- (Indent Line) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleIndent]() {
				goto l1194
			}
			if !p.rules[ruleLine]() {
				goto l1194
			}
			return true
		l1194:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 182 OptionallyIndentedLine <- (Indent? Line) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1196, thunkPosition1196 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l1196
				}
				goto l1197
			l1196:
				position, thunkPosition = position1196, thunkPosition1196
			}
		l1197:
			if !p.rules[ruleLine]() {
				goto l1195
			}
			return true
		l1195:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 183 StartList <- (&. { yy = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !peekDot() {
				goto l1198
			}
			do(84)
			return true
		l1198:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 184 Line <- (RawLine { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleRawLine]() {
				goto l1199
			}
			do(85)
			return true
		l1199:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 185 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > Eof)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1201, thunkPosition1201 := position, thunkPosition
				begin = position
			l1203:
				{
					position1204, thunkPosition1204 := position, thunkPosition
					if peekChar('\r') {
						goto l1204
					}
					if peekChar('\n') {
						goto l1204
					}
					if !matchDot() {
						goto l1204
					}
					goto l1203
				l1204:
					position, thunkPosition = position1204, thunkPosition1204
				}
				if !p.rules[ruleNewline]() {
					goto l1202
				}
				end = position
				goto l1201
			l1202:
				position, thunkPosition = position1201, thunkPosition1201
				begin = position
				if !matchDot() {
					goto l1200
				}
			l1205:
				{
					position1206, thunkPosition1206 := position, thunkPosition
					if !matchDot() {
						goto l1206
					}
					goto l1205
				l1206:
					position, thunkPosition = position1206, thunkPosition1206
				}
				end = position
				if !p.rules[ruleEof]() {
					goto l1200
				}
			}
		l1201:
			return true
		l1200:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 186 SkipBlock <- (((!BlankLine RawLine)+ BlankLine*) / BlankLine+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1208, thunkPosition1208 := position, thunkPosition
				{
					position1212, thunkPosition1212 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1212
					}
					goto l1209
				l1212:
					position, thunkPosition = position1212, thunkPosition1212
				}
				if !p.rules[ruleRawLine]() {
					goto l1209
				}
			l1210:
				{
					position1211, thunkPosition1211 := position, thunkPosition
					{
						position1213, thunkPosition1213 := position, thunkPosition
						if !p.rules[ruleBlankLine]() {
							goto l1213
						}
						goto l1211
					l1213:
						position, thunkPosition = position1213, thunkPosition1213
					}
					if !p.rules[ruleRawLine]() {
						goto l1211
					}
					goto l1210
				l1211:
					position, thunkPosition = position1211, thunkPosition1211
				}
			l1214:
				{
					position1215, thunkPosition1215 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1215
					}
					goto l1214
				l1215:
					position, thunkPosition = position1215, thunkPosition1215
				}
				goto l1208
			l1209:
				position, thunkPosition = position1208, thunkPosition1208
				if !p.rules[ruleBlankLine]() {
					goto l1207
				}
			l1216:
				{
					position1217, thunkPosition1217 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1217
					}
					goto l1216
				l1217:
					position, thunkPosition = position1217, thunkPosition1217
				}
			}
		l1208:
			return true
		l1207:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 187 ExtendedSpecialChar <- ((&{ p.extension.Smart } ('.' / '-' / '\'' / '"')) / (&{ p.extension.Notes } '^')) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1219, thunkPosition1219 := position, thunkPosition
				if !( p.extension.Smart ) {
					goto l1220
				}
				{
					position1221, thunkPosition1221 := position, thunkPosition
					if !matchChar('.') {
						goto l1222
					}
					goto l1221
				l1222:
					position, thunkPosition = position1221, thunkPosition1221
					if !matchChar('-') {
						goto l1223
					}
					goto l1221
				l1223:
					position, thunkPosition = position1221, thunkPosition1221
					if !matchChar('\'') {
						goto l1224
					}
					goto l1221
				l1224:
					position, thunkPosition = position1221, thunkPosition1221
					if !matchChar('"') {
						goto l1220
					}
				}
			l1221:
				goto l1219
			l1220:
				position, thunkPosition = position1219, thunkPosition1219
				if !( p.extension.Notes ) {
					goto l1218
				}
				if !matchChar('^') {
					goto l1218
				}
			}
		l1219:
			return true
		l1218:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 188 Smart <- (&{ p.extension.Smart } (Ellipsis / Dash / SingleQuoted / DoubleQuoted / Apostrophe)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !( p.extension.Smart ) {
				goto l1225
			}
			{
				position1226, thunkPosition1226 := position, thunkPosition
				if !p.rules[ruleEllipsis]() {
					goto l1227
				}
				goto l1226
			l1227:
				position, thunkPosition = position1226, thunkPosition1226
				if !p.rules[ruleDash]() {
					goto l1228
				}
				goto l1226
			l1228:
				position, thunkPosition = position1226, thunkPosition1226
				if !p.rules[ruleSingleQuoted]() {
					goto l1229
				}
				goto l1226
			l1229:
				position, thunkPosition = position1226, thunkPosition1226
				if !p.rules[ruleDoubleQuoted]() {
					goto l1230
				}
				goto l1226
			l1230:
				position, thunkPosition = position1226, thunkPosition1226
				if !p.rules[ruleApostrophe]() {
					goto l1225
				}
			}
		l1226:
			return true
		l1225:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 189 Apostrophe <- ('\'' { yy = mk_element(APOSTROPHE) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1231
			}
			do(86)
			return true
		l1231:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 190 Ellipsis <- (('...' / '. . .') { yy = mk_element(ELLIPSIS) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1233, thunkPosition1233 := position, thunkPosition
				if !matchString("...") {
					goto l1234
				}
				goto l1233
			l1234:
				position, thunkPosition = position1233, thunkPosition1233
				if !matchString(". . .") {
					goto l1232
				}
			}
		l1233:
			do(87)
			return true
		l1232:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 191 Dash <- (EmDash / EnDash) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1236, thunkPosition1236 := position, thunkPosition
				if !p.rules[ruleEmDash]() {
					goto l1237
				}
				goto l1236
			l1237:
				position, thunkPosition = position1236, thunkPosition1236
				if !p.rules[ruleEnDash]() {
					goto l1235
				}
			}
		l1236:
			return true
		l1235:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 192 EnDash <- ('-' &Digit { yy = mk_element(ENDASH) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('-') {
				goto l1238
			}
			{
				position1239, thunkPosition1239 := position, thunkPosition
				if !p.rules[ruleDigit]() {
					goto l1238
				}
				position, thunkPosition = position1239, thunkPosition1239
			}
			do(88)
			return true
		l1238:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 193 EmDash <- (('---' / '--') { yy = mk_element(EMDASH) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1241, thunkPosition1241 := position, thunkPosition
				if !matchString("---") {
					goto l1242
				}
				goto l1241
			l1242:
				position, thunkPosition = position1241, thunkPosition1241
				if !matchString("--") {
					goto l1240
				}
			}
		l1241:
			do(89)
			return true
		l1240:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 194 SingleQuoteStart <- ('\'' ![)!\],.;:-? \t\n] !(('s' / 't' / 'm' / 've' / 'll' / 're') !Alphanumeric)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1243
			}
			{
				position1244, thunkPosition1244 := position, thunkPosition
				if !matchClass(4) {
					goto l1244
				}
				goto l1243
			l1244:
				position, thunkPosition = position1244, thunkPosition1244
			}
			{
				position1245, thunkPosition1245 := position, thunkPosition
				{
					position1246, thunkPosition1246 := position, thunkPosition
					if !matchChar('s') {
						goto l1247
					}
					goto l1246
				l1247:
					position, thunkPosition = position1246, thunkPosition1246
					if !matchChar('t') {
						goto l1248
					}
					goto l1246
				l1248:
					position, thunkPosition = position1246, thunkPosition1246
					if !matchChar('m') {
						goto l1249
					}
					goto l1246
				l1249:
					position, thunkPosition = position1246, thunkPosition1246
					if !matchString("ve") {
						goto l1250
					}
					goto l1246
				l1250:
					position, thunkPosition = position1246, thunkPosition1246
					if !matchString("ll") {
						goto l1251
					}
					goto l1246
				l1251:
					position, thunkPosition = position1246, thunkPosition1246
					if !matchString("re") {
						goto l1245
					}
				}
			l1246:
				{
					position1252, thunkPosition1252 := position, thunkPosition
					if !p.rules[ruleAlphanumeric]() {
						goto l1252
					}
					goto l1245
				l1252:
					position, thunkPosition = position1252, thunkPosition1252
				}
				goto l1243
			l1245:
				position, thunkPosition = position1245, thunkPosition1245
			}
			return true
		l1243:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 195 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1253
			}
			{
				position1254, thunkPosition1254 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l1254
				}
				goto l1253
			l1254:
				position, thunkPosition = position1254, thunkPosition1254
			}
			return true
		l1253:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 196 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = mk_list(SINGLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto l1255
			}
			if !p.rules[ruleStartList]() {
				goto l1255
			}
			doarg(yySet, -1)
			{
				position1258, thunkPosition1258 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto l1258
				}
				goto l1255
			l1258:
				position, thunkPosition = position1258, thunkPosition1258
			}
			if !p.rules[ruleInline]() {
				goto l1255
			}
			doarg(yySet, -2)
			do(90)
		l1256:
			{
				position1257, thunkPosition1257 := position, thunkPosition
				{
					position1259, thunkPosition1259 := position, thunkPosition
					if !p.rules[ruleSingleQuoteEnd]() {
						goto l1259
					}
					goto l1257
				l1259:
					position, thunkPosition = position1259, thunkPosition1259
				}
				if !p.rules[ruleInline]() {
					goto l1257
				}
				doarg(yySet, -2)
				do(90)
				goto l1256
			l1257:
				position, thunkPosition = position1257, thunkPosition1257
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1255
			}
			do(91)
			doarg(yyPop, 2)
			return true
		l1255:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 197 DoubleQuoteStart <- '"' */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l1260
			}
			return true
		l1260:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 198 DoubleQuoteEnd <- '"' */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l1261
			}
			return true
		l1261:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 199 DoubleQuoted <- (DoubleQuoteStart StartList (!DoubleQuoteEnd Inline { a = cons(b, a) })+ DoubleQuoteEnd { yy = mk_list(DOUBLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleDoubleQuoteStart]() {
				goto l1262
			}
			if !p.rules[ruleStartList]() {
				goto l1262
			}
			doarg(yySet, -1)
			{
				position1265, thunkPosition1265 := position, thunkPosition
				if !p.rules[ruleDoubleQuoteEnd]() {
					goto l1265
				}
				goto l1262
			l1265:
				position, thunkPosition = position1265, thunkPosition1265
			}
			if !p.rules[ruleInline]() {
				goto l1262
			}
			doarg(yySet, -2)
			do(92)
		l1263:
			{
				position1264, thunkPosition1264 := position, thunkPosition
				{
					position1266, thunkPosition1266 := position, thunkPosition
					if !p.rules[ruleDoubleQuoteEnd]() {
						goto l1266
					}
					goto l1264
				l1266:
					position, thunkPosition = position1266, thunkPosition1266
				}
				if !p.rules[ruleInline]() {
					goto l1264
				}
				doarg(yySet, -2)
				do(92)
				goto l1263
			l1264:
				position, thunkPosition = position1264, thunkPosition1264
			}
			if !p.rules[ruleDoubleQuoteEnd]() {
				goto l1262
			}
			do(93)
			doarg(yyPop, 2)
			return true
		l1262:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 200 NoteReference <- (&{ p.extension.Notes } RawNoteReference {
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
			if !( p.extension.Notes ) {
				goto l1267
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1267
			}
			doarg(yySet, -1)
			do(94)
			doarg(yyPop, 1)
			return true
		l1267:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 201 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("[^") {
				goto l1268
			}
			begin = position
			{
				position1271, thunkPosition1271 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1271
				}
				goto l1268
			l1271:
				position, thunkPosition = position1271, thunkPosition1271
			}
			if peekChar(']') {
				goto l1268
			}
			if !matchDot() {
				goto l1268
			}
		l1269:
			{
				position1270, thunkPosition1270 := position, thunkPosition
				{
					position1272, thunkPosition1272 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l1272
					}
					goto l1270
				l1272:
					position, thunkPosition = position1272, thunkPosition1272
				}
				if peekChar(']') {
					goto l1270
				}
				if !matchDot() {
					goto l1270
				}
				goto l1269
			l1270:
				position, thunkPosition = position1270, thunkPosition1270
			}
			end = position
			if !matchChar(']') {
				goto l1268
			}
			do(95)
			return true
		l1268:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 202 Note <- (&{ p.extension.Notes } NonindentSpace RawNoteReference ':' Sp StartList (RawNoteBlock { a = cons(yy, a) }) (&Indent RawNoteBlock { a = cons(yy, a) })* {   yy = mk_list(NOTE, a)
                    yy.contents.str = ref.contents.str
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !( p.extension.Notes ) {
				goto l1273
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l1273
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1273
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto l1273
			}
			if !p.rules[ruleSp]() {
				goto l1273
			}
			if !p.rules[ruleStartList]() {
				goto l1273
			}
			doarg(yySet, -2)
			if !p.rules[ruleRawNoteBlock]() {
				goto l1273
			}
			do(96)
		l1274:
			{
				position1275, thunkPosition1275 := position, thunkPosition
				{
					position1276, thunkPosition1276 := position, thunkPosition
					if !p.rules[ruleIndent]() {
						goto l1275
					}
					position, thunkPosition = position1276, thunkPosition1276
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto l1275
				}
				do(97)
				goto l1274
			l1275:
				position, thunkPosition = position1275, thunkPosition1275
			}
			do(98)
			doarg(yyPop, 2)
			return true
		l1273:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 203 InlineNote <- (&{ p.extension.Notes } '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = mk_list(NOTE, a)
                  yy.contents.str = "" }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !( p.extension.Notes ) {
				goto l1277
			}
			if !matchString("^[") {
				goto l1277
			}
			if !p.rules[ruleStartList]() {
				goto l1277
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto l1277
			}
			if !p.rules[ruleInline]() {
				goto l1277
			}
			do(99)
		l1278:
			{
				position1279, thunkPosition1279 := position, thunkPosition
				if peekChar(']') {
					goto l1279
				}
				if !p.rules[ruleInline]() {
					goto l1279
				}
				do(99)
				goto l1278
			l1279:
				position, thunkPosition = position1279, thunkPosition1279
			}
			if !matchChar(']') {
				goto l1277
			}
			do(100)
			doarg(yyPop, 1)
			return true
		l1277:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 204 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1280
			}
			doarg(yySet, -1)
		l1281:
			{
				position1282, thunkPosition1282 := position, thunkPosition
				{
					position1283, thunkPosition1283 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto l1284
					}
					doarg(yySet, -2)
					do(101)
					goto l1283
				l1284:
					position, thunkPosition = position1283, thunkPosition1283
					if !p.rules[ruleSkipBlock]() {
						goto l1282
					}
				}
			l1283:
				goto l1281
			l1282:
				position, thunkPosition = position1282, thunkPosition1282
			}
			do(102)
			if !(commit(thunkPosition0)) {
				goto l1280
			}
			doarg(yyPop, 2)
			return true
		l1280:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 205 RawNoteBlock <- (StartList (!BlankLine OptionallyIndentedLine { a = cons(yy, a) })+ (< BlankLine* > { a = cons(mk_str(yytext), a) }) {   yy = mk_str_from_list(a, true)
                    yy.key = RAW
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1285
			}
			doarg(yySet, -1)
			{
				position1288, thunkPosition1288 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l1288
				}
				goto l1285
			l1288:
				position, thunkPosition = position1288, thunkPosition1288
			}
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l1285
			}
			do(103)
		l1286:
			{
				position1287, thunkPosition1287 := position, thunkPosition
				{
					position1289, thunkPosition1289 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1289
					}
					goto l1287
				l1289:
					position, thunkPosition = position1289, thunkPosition1289
				}
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto l1287
				}
				do(103)
				goto l1286
			l1287:
				position, thunkPosition = position1287, thunkPosition1287
			}
			begin = position
		l1290:
			{
				position1291, thunkPosition1291 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l1291
				}
				goto l1290
			l1291:
				position, thunkPosition = position1291, thunkPosition1291
			}
			end = position
			do(104)
			do(105)
			doarg(yyPop, 1)
			return true
		l1285:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 206 DefinitionList <- (&{ p.extension.Dlists } StartList (Definition { a = cons(yy, a) })+ { yy = mk_list(DEFINITIONLIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !( p.extension.Dlists ) {
				goto l1292
			}
			if !p.rules[ruleStartList]() {
				goto l1292
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto l1292
			}
			do(106)
		l1293:
			{
				position1294, thunkPosition1294 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto l1294
				}
				do(106)
				goto l1293
			l1294:
				position, thunkPosition = position1294, thunkPosition1294
			}
			do(107)
			doarg(yyPop, 1)
			return true
		l1292:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 207 Definition <- (&((!Defmark RawLine)+ BlankLine? Defmark) StartList (DListTitle { a = cons(yy, a) })+ (DefTight / DefLoose) {
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			} { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1296, thunkPosition1296 := position, thunkPosition
				{
					position1299, thunkPosition1299 := position, thunkPosition
					if !p.rules[ruleDefmark]() {
						goto l1299
					}
					goto l1295
				l1299:
					position, thunkPosition = position1299, thunkPosition1299
				}
				if !p.rules[ruleRawLine]() {
					goto l1295
				}
			l1297:
				{
					position1298, thunkPosition1298 := position, thunkPosition
					{
						position1300, thunkPosition1300 := position, thunkPosition
						if !p.rules[ruleDefmark]() {
							goto l1300
						}
						goto l1298
					l1300:
						position, thunkPosition = position1300, thunkPosition1300
					}
					if !p.rules[ruleRawLine]() {
						goto l1298
					}
					goto l1297
				l1298:
					position, thunkPosition = position1298, thunkPosition1298
				}
				{
					position1301, thunkPosition1301 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1301
					}
					goto l1302
				l1301:
					position, thunkPosition = position1301, thunkPosition1301
				}
			l1302:
				if !p.rules[ruleDefmark]() {
					goto l1295
				}
				position, thunkPosition = position1296, thunkPosition1296
			}
			if !p.rules[ruleStartList]() {
				goto l1295
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto l1295
			}
			do(108)
		l1303:
			{
				position1304, thunkPosition1304 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto l1304
				}
				do(108)
				goto l1303
			l1304:
				position, thunkPosition = position1304, thunkPosition1304
			}
			{
				position1305, thunkPosition1305 := position, thunkPosition
				if !p.rules[ruleDefTight]() {
					goto l1306
				}
				goto l1305
			l1306:
				position, thunkPosition = position1305, thunkPosition1305
				if !p.rules[ruleDefLoose]() {
					goto l1295
				}
			}
		l1305:
			do(109)
			do(110)
			doarg(yyPop, 1)
			return true
		l1295:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 208 DListTitle <- (NonindentSpace !Defmark &Nonspacechar StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline {	yy = mk_list(LIST, a)
				yy.key = DEFTITLE
			}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l1307
			}
			{
				position1308, thunkPosition1308 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1308
				}
				goto l1307
			l1308:
				position, thunkPosition = position1308, thunkPosition1308
			}
			{
				position1309, thunkPosition1309 := position, thunkPosition
				if !p.rules[ruleNonspacechar]() {
					goto l1307
				}
				position, thunkPosition = position1309, thunkPosition1309
			}
			if !p.rules[ruleStartList]() {
				goto l1307
			}
			doarg(yySet, -1)
			{
				position1312, thunkPosition1312 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l1312
				}
				goto l1307
			l1312:
				position, thunkPosition = position1312, thunkPosition1312
			}
			if !p.rules[ruleInline]() {
				goto l1307
			}
			do(111)
		l1310:
			{
				position1311, thunkPosition1311 := position, thunkPosition
				{
					position1313, thunkPosition1313 := position, thunkPosition
					if !p.rules[ruleEndline]() {
						goto l1313
					}
					goto l1311
				l1313:
					position, thunkPosition = position1313, thunkPosition1313
				}
				if !p.rules[ruleInline]() {
					goto l1311
				}
				do(111)
				goto l1310
			l1311:
				position, thunkPosition = position1311, thunkPosition1311
			}
			if !p.rules[ruleSp]() {
				goto l1307
			}
			if !p.rules[ruleNewline]() {
				goto l1307
			}
			do(112)
			doarg(yyPop, 1)
			return true
		l1307:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 209 DefTight <- (&Defmark ListTight) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1315, thunkPosition1315 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1314
				}
				position, thunkPosition = position1315, thunkPosition1315
			}
			if !p.rules[ruleListTight]() {
				goto l1314
			}
			return true
		l1314:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 210 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlankLine]() {
				goto l1316
			}
			{
				position1317, thunkPosition1317 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1316
				}
				position, thunkPosition = position1317, thunkPosition1317
			}
			if !p.rules[ruleListLoose]() {
				goto l1316
			}
			return true
		l1316:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 211 Defmark <- (NonindentSpace (':' / '~') Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l1318
			}
			{
				position1319, thunkPosition1319 := position, thunkPosition
				if !matchChar(':') {
					goto l1320
				}
				goto l1319
			l1320:
				position, thunkPosition = position1319, thunkPosition1319
				if !matchChar('~') {
					goto l1318
				}
			}
		l1319:
			if !p.rules[ruleSpacechar]() {
				goto l1318
			}
		l1321:
			{
				position1322, thunkPosition1322 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1322
				}
				goto l1321
			l1322:
				position, thunkPosition = position1322, thunkPosition1322
			}
			return true
		l1318:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 212 DefMarker <- (&{ p.extension.Dlists } Defmark) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !( p.extension.Dlists ) {
				goto l1323
			}
			if !p.rules[ruleDefmark]() {
				goto l1323
			}
			return true
		l1323:
			position, thunkPosition = position0, thunkPosition0
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

/* concat_string_list - concatenates string contents of list of STR elements.
 */
func concat_string_list(list *element) string {
	s := ""
	for list != nil {
		s += list.contents.str
		list = list.next
	}
	return s
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
	s := concat_string_list(reverse(list))
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
func mk_list(key int, lst *element) *element {
	result := mk_element(key)
	result.children = reverse(lst)
	return result
}

/* mk_link - constructor for LINK element
 */
func mk_link(label *element, url, title string) *element {
	result := mk_element(LINK)
	result.contents.link = &link{label: label, url: url, title: title}
	return result
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
func print_tree(elt *element, indent int) {
	var key string

	for elt != nil {
		for i := 0; i < indent; i++ {
			fmt.Print("\t")
		}
		key = keynames[elt.key]
		if key == "" {
			key = "?"
		}
		if elt.key == STR {
			fmt.Printf("%p:\t%s\t'%s'\n", elt, key, elt.contents.str)
		} else {
			fmt.Printf("%p:\t%s %p\n", elt, key, elt.next)
		}
		if elt.children != nil {
			print_tree(elt.children, indent+1)
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
