
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

type yyParser struct {*Doc
	Buffer string
	Min, Max int
	rules [249]func() bool
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
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
		},
		/* 43 Inlines */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(c, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
		},
		/* 44 Inlines */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = mk_list(LIST, a) 
			yyval[yyp-1] = c
			yyval[yyp-2] = a
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
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 84 References */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 p.references = reverse(a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
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
			yyval[yyp-1] = a
			yyval[yyp-2] = b
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
			yyval[yyp-1] = ref
			yyval[yyp-2] = a
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
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 105 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
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
		/* 116 yyPush */
		func(_ string, count int) {
			yyp += count
			if yyp >= len(yyval) {
				s := make([]*element, cap(yyval)+200)
				copy(s, yyval)
				yyval = s
			}
		},
		/* 117 yyPop */
		func(_ string, count int) {
			yyp -= count
		},
		/* 118 yySet */
		func(_ string, count int) {
			yyval[yyp+count] = yy
		},
	}
	const (
		yyPush = 116+iota
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
	_ = peekDot

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
	_ = peekChar

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
		{0, 0, 0, 0, 0, 0, 255, 3, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 10, 111, 0, 80, 0, 0, 0, 184, 1, 0, 0, 56, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 31, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
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
		/* 4 AtxInline <- (!Newline !(Sp? '#'* Sp Newline) Inline) */
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
				{
					position26, thunkPosition26 := position, thunkPosition
					if !p.rules[ruleSp]() {
						goto l26
					}
					goto l27
				l26:
					position, thunkPosition = position26, thunkPosition26
				}
			l27:
			l28:
				{
					position29, thunkPosition29 := position, thunkPosition
					if !matchChar('#') {
						goto l29
					}
					goto l28
				l29:
					position, thunkPosition = position29, thunkPosition29
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
				goto l30
			}
			begin = position
			{
				position31, thunkPosition31 := position, thunkPosition
				if !matchString("######") {
					goto l32
				}
				goto l31
			l32:
				position, thunkPosition = position31, thunkPosition31
				if !matchString("#####") {
					goto l33
				}
				goto l31
			l33:
				position, thunkPosition = position31, thunkPosition31
				if !matchString("####") {
					goto l34
				}
				goto l31
			l34:
				position, thunkPosition = position31, thunkPosition31
				if !matchString("###") {
					goto l35
				}
				goto l31
			l35:
				position, thunkPosition = position31, thunkPosition31
				if !matchString("##") {
					goto l36
				}
				goto l31
			l36:
				position, thunkPosition = position31, thunkPosition31
				if !matchChar('#') {
					goto l30
				}
			}
		l31:
			end = position
			do(4)
			return true
		l30:
			position, thunkPosition = position0, thunkPosition0
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
			{
				position38, thunkPosition38 := position, thunkPosition
				if !p.rules[ruleSp]() {
					goto l38
				}
				goto l39
			l38:
				position, thunkPosition = position38, thunkPosition38
			}
		l39:
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
				position41, thunkPosition41 := position, thunkPosition
				if !p.rules[ruleAtxInline]() {
					goto l41
				}
				do(5)
				goto l40
			l41:
				position, thunkPosition = position41, thunkPosition41
			}
			{
				position42, thunkPosition42 := position, thunkPosition
				{
					position44, thunkPosition44 := position, thunkPosition
					if !p.rules[ruleSp]() {
						goto l44
					}
					goto l45
				l44:
					position, thunkPosition = position44, thunkPosition44
				}
			l45:
			l46:
				{
					position47, thunkPosition47 := position, thunkPosition
					if !matchChar('#') {
						goto l47
					}
					goto l46
				l47:
					position, thunkPosition = position47, thunkPosition47
				}
				if !p.rules[ruleSp]() {
					goto l42
				}
				goto l43
			l42:
				position, thunkPosition = position42, thunkPosition42
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
			position0, thunkPosition0 := position, thunkPosition
			{
				position49, thunkPosition49 := position, thunkPosition
				if !p.rules[ruleSetextHeading1]() {
					goto l50
				}
				goto l49
			l50:
				position, thunkPosition = position49, thunkPosition49
				if !p.rules[ruleSetextHeading2]() {
					goto l48
				}
			}
		l49:
			return true
		l48:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 8 SetextBottom1 <- ('===' '='* Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("===") {
				goto l51
			}
		l52:
			{
				position53, thunkPosition53 := position, thunkPosition
				if !matchChar('=') {
					goto l53
				}
				goto l52
			l53:
				position, thunkPosition = position53, thunkPosition53
			}
			if !p.rules[ruleNewline]() {
				goto l51
			}
			return true
		l51:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 9 SetextBottom2 <- ('---' '-'* Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("---") {
				goto l54
			}
		l55:
			{
				position56, thunkPosition56 := position, thunkPosition
				if !matchChar('-') {
					goto l56
				}
				goto l55
			l56:
				position, thunkPosition = position56, thunkPosition56
			}
			if !p.rules[ruleNewline]() {
				goto l54
			}
			return true
		l54:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 10 SetextHeading1 <- (&(RawLine SetextBottom1) StartList (!Endline Inline { a = cons(yy, a) })+ Newline SetextBottom1 { yy = mk_list(H1, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position58, thunkPosition58 := position, thunkPosition
				if !p.rules[ruleRawLine]() {
					goto l57
				}
				if !p.rules[ruleSetextBottom1]() {
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
			do(7)
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
				do(7)
				goto l59
			l60:
				position, thunkPosition = position60, thunkPosition60
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
				position64, thunkPosition64 := position, thunkPosition
				if !p.rules[ruleRawLine]() {
					goto l63
				}
				if !p.rules[ruleSetextBottom2]() {
					goto l63
				}
				position, thunkPosition = position64, thunkPosition64
			}
			if !p.rules[ruleStartList]() {
				goto l63
			}
			doarg(yySet, -1)
			{
				position67, thunkPosition67 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l67
				}
				goto l63
			l67:
				position, thunkPosition = position67, thunkPosition67
			}
			if !p.rules[ruleInline]() {
				goto l63
			}
			do(9)
		l65:
			{
				position66, thunkPosition66 := position, thunkPosition
				{
					position68, thunkPosition68 := position, thunkPosition
					if !p.rules[ruleEndline]() {
						goto l68
					}
					goto l66
				l68:
					position, thunkPosition = position68, thunkPosition68
				}
				if !p.rules[ruleInline]() {
					goto l66
				}
				do(9)
				goto l65
			l66:
				position, thunkPosition = position66, thunkPosition66
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
			position0, thunkPosition0 := position, thunkPosition
			{
				position70, thunkPosition70 := position, thunkPosition
				if !p.rules[ruleAtxHeading]() {
					goto l71
				}
				goto l70
			l71:
				position, thunkPosition = position70, thunkPosition70
				if !p.rules[ruleSetextHeading]() {
					goto l69
				}
			}
		l70:
			return true
		l69:
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
			{
				position76, thunkPosition76 := position, thunkPosition
				if !matchChar(' ') {
					goto l76
				}
				goto l77
			l76:
				position, thunkPosition = position76, thunkPosition76
			}
		l77:
			if !p.rules[ruleLine]() {
				goto l73
			}
			do(12)
		l78:
			{
				position79, thunkPosition79 := position, thunkPosition
				if peekChar('>') {
					goto l79
				}
				{
					position80, thunkPosition80 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l80
					}
					goto l79
				l80:
					position, thunkPosition = position80, thunkPosition80
				}
				if !p.rules[ruleLine]() {
					goto l79
				}
				do(13)
				goto l78
			l79:
				position, thunkPosition = position79, thunkPosition79
			}
		l81:
			{
				position82, thunkPosition82 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l82
				}
				do(14)
				goto l81
			l82:
				position, thunkPosition = position82, thunkPosition82
			}
		l74:
			{
				position75, thunkPosition75 := position, thunkPosition
				if !matchChar('>') {
					goto l75
				}
				{
					position83, thunkPosition83 := position, thunkPosition
					if !matchChar(' ') {
						goto l83
					}
					goto l84
				l83:
					position, thunkPosition = position83, thunkPosition83
				}
			l84:
				if !p.rules[ruleLine]() {
					goto l75
				}
				do(12)
			l85:
				{
					position86, thunkPosition86 := position, thunkPosition
					if peekChar('>') {
						goto l86
					}
					{
						position87, thunkPosition87 := position, thunkPosition
						if !p.rules[ruleBlankLine]() {
							goto l87
						}
						goto l86
					l87:
						position, thunkPosition = position87, thunkPosition87
					}
					if !p.rules[ruleLine]() {
						goto l86
					}
					do(13)
					goto l85
				l86:
					position, thunkPosition = position86, thunkPosition86
				}
			l88:
				{
					position89, thunkPosition89 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l89
					}
					do(14)
					goto l88
				l89:
					position, thunkPosition = position89, thunkPosition89
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
			position0, thunkPosition0 := position, thunkPosition
			{
				position91, thunkPosition91 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l91
				}
				goto l90
			l91:
				position, thunkPosition = position91, thunkPosition91
			}
			if !p.rules[ruleIndentedLine]() {
				goto l90
			}
			return true
		l90:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 VerbatimChunk <- (StartList (BlankLine { a = cons(mk_str("\n"), a) })* (NonblankIndentedLine { a = cons(yy, a) })+ { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l92
			}
			doarg(yySet, -1)
		l93:
			{
				position94, thunkPosition94 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l94
				}
				do(16)
				goto l93
			l94:
				position, thunkPosition = position94, thunkPosition94
			}
			if !p.rules[ruleNonblankIndentedLine]() {
				goto l92
			}
			do(17)
		l95:
			{
				position96, thunkPosition96 := position, thunkPosition
				if !p.rules[ruleNonblankIndentedLine]() {
					goto l96
				}
				do(17)
				goto l95
			l96:
				position, thunkPosition = position96, thunkPosition96
			}
			do(18)
			doarg(yyPop, 1)
			return true
		l92:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 17 Verbatim <- (StartList (VerbatimChunk { a = cons(yy, a) })+ { yy = mk_str_from_list(a, false)
                 yy.key = VERBATIM }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l97
			}
			doarg(yySet, -1)
			if !p.rules[ruleVerbatimChunk]() {
				goto l97
			}
			do(19)
		l98:
			{
				position99, thunkPosition99 := position, thunkPosition
				if !p.rules[ruleVerbatimChunk]() {
					goto l99
				}
				do(19)
				goto l98
			l99:
				position, thunkPosition = position99, thunkPosition99
			}
			do(20)
			doarg(yyPop, 1)
			return true
		l97:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 18 HorizontalRule <- (NonindentSpace ((&[_] ('_' Sp '_' Sp '_' (Sp '_')*)) | (&[\-] ('-' Sp '-' Sp '-' (Sp '-')*)) | (&[*] ('*' Sp '*' Sp '*' (Sp '*')*))) Sp Newline BlankLine+ { yy = mk_element(HRULE) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l100
			}
			{
				if position == len(p.Buffer) {
					goto l100
				}
				switch p.Buffer[position] {
				case '_':
					if !matchChar('_') {
						goto l100
					}
					if !p.rules[ruleSp]() {
						goto l100
					}
					if !matchChar('_') {
						goto l100
					}
					if !p.rules[ruleSp]() {
						goto l100
					}
					if !matchChar('_') {
						goto l100
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
				case '-':
					if !matchChar('-') {
						goto l100
					}
					if !p.rules[ruleSp]() {
						goto l100
					}
					if !matchChar('-') {
						goto l100
					}
					if !p.rules[ruleSp]() {
						goto l100
					}
					if !matchChar('-') {
						goto l100
					}
				l104:
					{
						position105, thunkPosition105 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l105
						}
						if !matchChar('-') {
							goto l105
						}
						goto l104
					l105:
						position, thunkPosition = position105, thunkPosition105
					}
				default:
					if !matchChar('*') {
						goto l100
					}
					if !p.rules[ruleSp]() {
						goto l100
					}
					if !matchChar('*') {
						goto l100
					}
					if !p.rules[ruleSp]() {
						goto l100
					}
					if !matchChar('*') {
						goto l100
					}
				l106:
					{
						position107, thunkPosition107 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l107
						}
						if !matchChar('*') {
							goto l107
						}
						goto l106
					l107:
						position, thunkPosition = position107, thunkPosition107
					}
				}
			}
			if !p.rules[ruleSp]() {
				goto l100
			}
			if !p.rules[ruleNewline]() {
				goto l100
			}
			if !p.rules[ruleBlankLine]() {
				goto l100
			}
		l108:
			{
				position109, thunkPosition109 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l109
				}
				goto l108
			l109:
				position, thunkPosition = position109, thunkPosition109
			}
			do(21)
			return true
		l100:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 19 Bullet <- (!HorizontalRule NonindentSpace ((&[\-] '-') | (&[*] '*') | (&[+] '+')) Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position111, thunkPosition111 := position, thunkPosition
				if !p.rules[ruleHorizontalRule]() {
					goto l111
				}
				goto l110
			l111:
				position, thunkPosition = position111, thunkPosition111
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l110
			}
			{
				if position == len(p.Buffer) {
					goto l110
				}
				switch p.Buffer[position] {
				case '-':
					if !matchChar('-') {
						goto l110
					}
				case '*':
					if !matchChar('*') {
						goto l110
					}
				default:
					if !matchChar('+') {
						goto l110
					}
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l110
			}
		l113:
			{
				position114, thunkPosition114 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l114
				}
				goto l113
			l114:
				position, thunkPosition = position114, thunkPosition114
			}
			return true
		l110:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 20 BulletList <- (&Bullet (ListTight / ListLoose) { yy.key = BULLETLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position116, thunkPosition116 := position, thunkPosition
				if !p.rules[ruleBullet]() {
					goto l115
				}
				position, thunkPosition = position116, thunkPosition116
			}
			{
				position117, thunkPosition117 := position, thunkPosition
				if !p.rules[ruleListTight]() {
					goto l118
				}
				goto l117
			l118:
				position, thunkPosition = position117, thunkPosition117
				if !p.rules[ruleListLoose]() {
					goto l115
				}
			}
		l117:
			do(22)
			return true
		l115:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 21 ListTight <- (StartList (ListItemTight { a = cons(yy, a) })+ BlankLine* !(Bullet / Enumerator / DefMarker) { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l119
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItemTight]() {
				goto l119
			}
			do(23)
		l120:
			{
				position121, thunkPosition121 := position, thunkPosition
				if !p.rules[ruleListItemTight]() {
					goto l121
				}
				do(23)
				goto l120
			l121:
				position, thunkPosition = position121, thunkPosition121
			}
		l122:
			{
				position123, thunkPosition123 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l123
				}
				goto l122
			l123:
				position, thunkPosition = position123, thunkPosition123
			}
			{
				position124, thunkPosition124 := position, thunkPosition
				{
					position125, thunkPosition125 := position, thunkPosition
					if !p.rules[ruleBullet]() {
						goto l126
					}
					goto l125
				l126:
					position, thunkPosition = position125, thunkPosition125
					if !p.rules[ruleEnumerator]() {
						goto l127
					}
					goto l125
				l127:
					position, thunkPosition = position125, thunkPosition125
					if !p.rules[ruleDefMarker]() {
						goto l124
					}
				}
			l125:
				goto l119
			l124:
				position, thunkPosition = position124, thunkPosition124
			}
			do(24)
			doarg(yyPop, 1)
			return true
		l119:
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
				goto l128
			}
			doarg(yySet, -1)
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
		l129:
			{
				position130, thunkPosition130 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto l130
				}
				doarg(yySet, -2)
			l133:
				{
					position134, thunkPosition134 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l134
					}
					goto l133
				l134:
					position, thunkPosition = position134, thunkPosition134
				}
				do(25)
				goto l129
			l130:
				position, thunkPosition = position130, thunkPosition130
			}
			do(26)
			doarg(yyPop, 2)
			return true
		l128:
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
				position136, thunkPosition136 := position, thunkPosition
				if !p.rules[ruleBullet]() {
					goto l137
				}
				goto l136
			l137:
				position, thunkPosition = position136, thunkPosition136
				if !p.rules[ruleEnumerator]() {
					goto l138
				}
				goto l136
			l138:
				position, thunkPosition = position136, thunkPosition136
				if !p.rules[ruleDefMarker]() {
					goto l135
				}
			}
		l136:
			if !p.rules[ruleStartList]() {
				goto l135
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l135
			}
			do(27)
		l139:
			{
				position140, thunkPosition140 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l140
				}
				do(28)
				goto l139
			l140:
				position, thunkPosition = position140, thunkPosition140
			}
			do(29)
			doarg(yyPop, 1)
			return true
		l135:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 24 ListItemTight <- ((Bullet / Enumerator / DefMarker) StartList ListBlock { a = cons(yy, a) } (!BlankLine ListContinuationBlock { a = cons(yy, a) })* !ListContinuationBlock {
               raw := mk_str_from_list(a, false)
               raw.key = RAW
               yy = mk_element(LISTITEM)
               yy.children = raw
            }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position142, thunkPosition142 := position, thunkPosition
				if !p.rules[ruleBullet]() {
					goto l143
				}
				goto l142
			l143:
				position, thunkPosition = position142, thunkPosition142
				if !p.rules[ruleEnumerator]() {
					goto l144
				}
				goto l142
			l144:
				position, thunkPosition = position142, thunkPosition142
				if !p.rules[ruleDefMarker]() {
					goto l141
				}
			}
		l142:
			if !p.rules[ruleStartList]() {
				goto l141
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l141
			}
			do(30)
		l145:
			{
				position146, thunkPosition146 := position, thunkPosition
				{
					position147, thunkPosition147 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l147
					}
					goto l146
				l147:
					position, thunkPosition = position147, thunkPosition147
				}
				if !p.rules[ruleListContinuationBlock]() {
					goto l146
				}
				do(31)
				goto l145
			l146:
				position, thunkPosition = position146, thunkPosition146
			}
			{
				position148, thunkPosition148 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l148
				}
				goto l141
			l148:
				position, thunkPosition = position148, thunkPosition148
			}
			do(32)
			doarg(yyPop, 1)
			return true
		l141:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 25 ListBlock <- (StartList !BlankLine Line { a = cons(yy, a) } (ListBlockLine { a = cons(yy, a) })* { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l149
			}
			doarg(yySet, -1)
			{
				position150, thunkPosition150 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l150
				}
				goto l149
			l150:
				position, thunkPosition = position150, thunkPosition150
			}
			if !p.rules[ruleLine]() {
				goto l149
			}
			do(33)
		l151:
			{
				position152, thunkPosition152 := position, thunkPosition
				if !p.rules[ruleListBlockLine]() {
					goto l152
				}
				do(34)
				goto l151
			l152:
				position, thunkPosition = position152, thunkPosition152
			}
			do(35)
			doarg(yyPop, 1)
			return true
		l149:
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
				goto l153
			}
			doarg(yySet, -1)
			begin = position
		l154:
			{
				position155, thunkPosition155 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l155
				}
				goto l154
			l155:
				position, thunkPosition = position155, thunkPosition155
			}
			end = position
			do(36)
			if !p.rules[ruleIndent]() {
				goto l153
			}
			if !p.rules[ruleListBlock]() {
				goto l153
			}
			do(37)
		l156:
			{
				position157, thunkPosition157 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l157
				}
				if !p.rules[ruleListBlock]() {
					goto l157
				}
				do(37)
				goto l156
			l157:
				position, thunkPosition = position157, thunkPosition157
			}
			do(38)
			doarg(yyPop, 1)
			return true
		l153:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 27 Enumerator <- (NonindentSpace [0-9]+ '.' Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l158
			}
			if !matchClass(7) {
				goto l158
			}
		l159:
			{
				position160, thunkPosition160 := position, thunkPosition
				if !matchClass(7) {
					goto l160
				}
				goto l159
			l160:
				position, thunkPosition = position160, thunkPosition160
			}
			if !matchChar('.') {
				goto l158
			}
			if !p.rules[ruleSpacechar]() {
				goto l158
			}
		l161:
			{
				position162, thunkPosition162 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l162
				}
				goto l161
			l162:
				position, thunkPosition = position162, thunkPosition162
			}
			return true
		l158:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 28 OrderedList <- (&Enumerator (ListTight / ListLoose) { yy.key = ORDEREDLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position164, thunkPosition164 := position, thunkPosition
				if !p.rules[ruleEnumerator]() {
					goto l163
				}
				position, thunkPosition = position164, thunkPosition164
			}
			{
				position165, thunkPosition165 := position, thunkPosition
				if !p.rules[ruleListTight]() {
					goto l166
				}
				goto l165
			l166:
				position, thunkPosition = position165, thunkPosition165
				if !p.rules[ruleListLoose]() {
					goto l163
				}
			}
		l165:
			do(39)
			return true
		l163:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 29 ListBlockLine <- (!BlankLine !((Indent? (Bullet / Enumerator)) / DefMarker) !HorizontalRule OptionallyIndentedLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position168, thunkPosition168 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l168
				}
				goto l167
			l168:
				position, thunkPosition = position168, thunkPosition168
			}
			{
				position169, thunkPosition169 := position, thunkPosition
				{
					position170, thunkPosition170 := position, thunkPosition
					{
						position172, thunkPosition172 := position, thunkPosition
						if !p.rules[ruleIndent]() {
							goto l172
						}
						goto l173
					l172:
						position, thunkPosition = position172, thunkPosition172
					}
				l173:
					{
						position174, thunkPosition174 := position, thunkPosition
						if !p.rules[ruleBullet]() {
							goto l175
						}
						goto l174
					l175:
						position, thunkPosition = position174, thunkPosition174
						if !p.rules[ruleEnumerator]() {
							goto l171
						}
					}
				l174:
					goto l170
				l171:
					position, thunkPosition = position170, thunkPosition170
					if !p.rules[ruleDefMarker]() {
						goto l169
					}
				}
			l170:
				goto l167
			l169:
				position, thunkPosition = position169, thunkPosition169
			}
			{
				position176, thunkPosition176 := position, thunkPosition
				if !p.rules[ruleHorizontalRule]() {
					goto l176
				}
				goto l167
			l176:
				position, thunkPosition = position176, thunkPosition176
			}
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l167
			}
			return true
		l167:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 HtmlBlockOpenAddress <- ('<' Spnl ('address' / 'ADDRESS') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l177
			}
			if !p.rules[ruleSpnl]() {
				goto l177
			}
			{
				position178, thunkPosition178 := position, thunkPosition
				if !matchString("address") {
					goto l179
				}
				goto l178
			l179:
				position, thunkPosition = position178, thunkPosition178
				if !matchString("ADDRESS") {
					goto l177
				}
			}
		l178:
			if !p.rules[ruleSpnl]() {
				goto l177
			}
		l180:
			{
				position181, thunkPosition181 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l181
				}
				goto l180
			l181:
				position, thunkPosition = position181, thunkPosition181
			}
			if !matchChar('>') {
				goto l177
			}
			return true
		l177:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 31 HtmlBlockCloseAddress <- ('<' Spnl '/' ('address' / 'ADDRESS') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position183, thunkPosition183 := position, thunkPosition
				if !matchString("address") {
					goto l184
				}
				goto l183
			l184:
				position, thunkPosition = position183, thunkPosition183
				if !matchString("ADDRESS") {
					goto l182
				}
			}
		l183:
			if !p.rules[ruleSpnl]() {
				goto l182
			}
			if !matchChar('>') {
				goto l182
			}
			return true
		l182:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 32 HtmlBlockAddress <- (HtmlBlockOpenAddress (HtmlBlockAddress / (!HtmlBlockCloseAddress .))* HtmlBlockCloseAddress) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenAddress]() {
				goto l185
			}
		l186:
			{
				position187, thunkPosition187 := position, thunkPosition
				{
					position188, thunkPosition188 := position, thunkPosition
					if !p.rules[ruleHtmlBlockAddress]() {
						goto l189
					}
					goto l188
				l189:
					position, thunkPosition = position188, thunkPosition188
					{
						position190, thunkPosition190 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseAddress]() {
							goto l190
						}
						goto l187
					l190:
						position, thunkPosition = position190, thunkPosition190
					}
					if !matchDot() {
						goto l187
					}
				}
			l188:
				goto l186
			l187:
				position, thunkPosition = position187, thunkPosition187
			}
			if !p.rules[ruleHtmlBlockCloseAddress]() {
				goto l185
			}
			return true
		l185:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 33 HtmlBlockOpenBlockquote <- ('<' Spnl ('blockquote' / 'BLOCKQUOTE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l191
			}
			if !p.rules[ruleSpnl]() {
				goto l191
			}
			{
				position192, thunkPosition192 := position, thunkPosition
				if !matchString("blockquote") {
					goto l193
				}
				goto l192
			l193:
				position, thunkPosition = position192, thunkPosition192
				if !matchString("BLOCKQUOTE") {
					goto l191
				}
			}
		l192:
			if !p.rules[ruleSpnl]() {
				goto l191
			}
		l194:
			{
				position195, thunkPosition195 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l195
				}
				goto l194
			l195:
				position, thunkPosition = position195, thunkPosition195
			}
			if !matchChar('>') {
				goto l191
			}
			return true
		l191:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 34 HtmlBlockCloseBlockquote <- ('<' Spnl '/' ('blockquote' / 'BLOCKQUOTE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l196
			}
			if !p.rules[ruleSpnl]() {
				goto l196
			}
			if !matchChar('/') {
				goto l196
			}
			{
				position197, thunkPosition197 := position, thunkPosition
				if !matchString("blockquote") {
					goto l198
				}
				goto l197
			l198:
				position, thunkPosition = position197, thunkPosition197
				if !matchString("BLOCKQUOTE") {
					goto l196
				}
			}
		l197:
			if !p.rules[ruleSpnl]() {
				goto l196
			}
			if !matchChar('>') {
				goto l196
			}
			return true
		l196:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 35 HtmlBlockBlockquote <- (HtmlBlockOpenBlockquote (HtmlBlockBlockquote / (!HtmlBlockCloseBlockquote .))* HtmlBlockCloseBlockquote) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenBlockquote]() {
				goto l199
			}
		l200:
			{
				position201, thunkPosition201 := position, thunkPosition
				{
					position202, thunkPosition202 := position, thunkPosition
					if !p.rules[ruleHtmlBlockBlockquote]() {
						goto l203
					}
					goto l202
				l203:
					position, thunkPosition = position202, thunkPosition202
					{
						position204, thunkPosition204 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseBlockquote]() {
							goto l204
						}
						goto l201
					l204:
						position, thunkPosition = position204, thunkPosition204
					}
					if !matchDot() {
						goto l201
					}
				}
			l202:
				goto l200
			l201:
				position, thunkPosition = position201, thunkPosition201
			}
			if !p.rules[ruleHtmlBlockCloseBlockquote]() {
				goto l199
			}
			return true
		l199:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 36 HtmlBlockOpenCenter <- ('<' Spnl ('center' / 'CENTER') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l205
			}
			if !p.rules[ruleSpnl]() {
				goto l205
			}
			{
				position206, thunkPosition206 := position, thunkPosition
				if !matchString("center") {
					goto l207
				}
				goto l206
			l207:
				position, thunkPosition = position206, thunkPosition206
				if !matchString("CENTER") {
					goto l205
				}
			}
		l206:
			if !p.rules[ruleSpnl]() {
				goto l205
			}
		l208:
			{
				position209, thunkPosition209 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l209
				}
				goto l208
			l209:
				position, thunkPosition = position209, thunkPosition209
			}
			if !matchChar('>') {
				goto l205
			}
			return true
		l205:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 37 HtmlBlockCloseCenter <- ('<' Spnl '/' ('center' / 'CENTER') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position211, thunkPosition211 := position, thunkPosition
				if !matchString("center") {
					goto l212
				}
				goto l211
			l212:
				position, thunkPosition = position211, thunkPosition211
				if !matchString("CENTER") {
					goto l210
				}
			}
		l211:
			if !p.rules[ruleSpnl]() {
				goto l210
			}
			if !matchChar('>') {
				goto l210
			}
			return true
		l210:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 38 HtmlBlockCenter <- (HtmlBlockOpenCenter (HtmlBlockCenter / (!HtmlBlockCloseCenter .))* HtmlBlockCloseCenter) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenCenter]() {
				goto l213
			}
		l214:
			{
				position215, thunkPosition215 := position, thunkPosition
				{
					position216, thunkPosition216 := position, thunkPosition
					if !p.rules[ruleHtmlBlockCenter]() {
						goto l217
					}
					goto l216
				l217:
					position, thunkPosition = position216, thunkPosition216
					{
						position218, thunkPosition218 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseCenter]() {
							goto l218
						}
						goto l215
					l218:
						position, thunkPosition = position218, thunkPosition218
					}
					if !matchDot() {
						goto l215
					}
				}
			l216:
				goto l214
			l215:
				position, thunkPosition = position215, thunkPosition215
			}
			if !p.rules[ruleHtmlBlockCloseCenter]() {
				goto l213
			}
			return true
		l213:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 39 HtmlBlockOpenDir <- ('<' Spnl ('dir' / 'DIR') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l219
			}
			if !p.rules[ruleSpnl]() {
				goto l219
			}
			{
				position220, thunkPosition220 := position, thunkPosition
				if !matchString("dir") {
					goto l221
				}
				goto l220
			l221:
				position, thunkPosition = position220, thunkPosition220
				if !matchString("DIR") {
					goto l219
				}
			}
		l220:
			if !p.rules[ruleSpnl]() {
				goto l219
			}
		l222:
			{
				position223, thunkPosition223 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l223
				}
				goto l222
			l223:
				position, thunkPosition = position223, thunkPosition223
			}
			if !matchChar('>') {
				goto l219
			}
			return true
		l219:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 40 HtmlBlockCloseDir <- ('<' Spnl '/' ('dir' / 'DIR') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l224
			}
			if !p.rules[ruleSpnl]() {
				goto l224
			}
			if !matchChar('/') {
				goto l224
			}
			{
				position225, thunkPosition225 := position, thunkPosition
				if !matchString("dir") {
					goto l226
				}
				goto l225
			l226:
				position, thunkPosition = position225, thunkPosition225
				if !matchString("DIR") {
					goto l224
				}
			}
		l225:
			if !p.rules[ruleSpnl]() {
				goto l224
			}
			if !matchChar('>') {
				goto l224
			}
			return true
		l224:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 41 HtmlBlockDir <- (HtmlBlockOpenDir (HtmlBlockDir / (!HtmlBlockCloseDir .))* HtmlBlockCloseDir) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDir]() {
				goto l227
			}
		l228:
			{
				position229, thunkPosition229 := position, thunkPosition
				{
					position230, thunkPosition230 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDir]() {
						goto l231
					}
					goto l230
				l231:
					position, thunkPosition = position230, thunkPosition230
					{
						position232, thunkPosition232 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDir]() {
							goto l232
						}
						goto l229
					l232:
						position, thunkPosition = position232, thunkPosition232
					}
					if !matchDot() {
						goto l229
					}
				}
			l230:
				goto l228
			l229:
				position, thunkPosition = position229, thunkPosition229
			}
			if !p.rules[ruleHtmlBlockCloseDir]() {
				goto l227
			}
			return true
		l227:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 42 HtmlBlockOpenDiv <- ('<' Spnl ('div' / 'DIV') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l233
			}
			if !p.rules[ruleSpnl]() {
				goto l233
			}
			{
				position234, thunkPosition234 := position, thunkPosition
				if !matchString("div") {
					goto l235
				}
				goto l234
			l235:
				position, thunkPosition = position234, thunkPosition234
				if !matchString("DIV") {
					goto l233
				}
			}
		l234:
			if !p.rules[ruleSpnl]() {
				goto l233
			}
		l236:
			{
				position237, thunkPosition237 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l237
				}
				goto l236
			l237:
				position, thunkPosition = position237, thunkPosition237
			}
			if !matchChar('>') {
				goto l233
			}
			return true
		l233:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 43 HtmlBlockCloseDiv <- ('<' Spnl '/' ('div' / 'DIV') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l238
			}
			if !p.rules[ruleSpnl]() {
				goto l238
			}
			if !matchChar('/') {
				goto l238
			}
			{
				position239, thunkPosition239 := position, thunkPosition
				if !matchString("div") {
					goto l240
				}
				goto l239
			l240:
				position, thunkPosition = position239, thunkPosition239
				if !matchString("DIV") {
					goto l238
				}
			}
		l239:
			if !p.rules[ruleSpnl]() {
				goto l238
			}
			if !matchChar('>') {
				goto l238
			}
			return true
		l238:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 44 HtmlBlockDiv <- (HtmlBlockOpenDiv (HtmlBlockDiv / (!HtmlBlockCloseDiv .))* HtmlBlockCloseDiv) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDiv]() {
				goto l241
			}
		l242:
			{
				position243, thunkPosition243 := position, thunkPosition
				{
					position244, thunkPosition244 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDiv]() {
						goto l245
					}
					goto l244
				l245:
					position, thunkPosition = position244, thunkPosition244
					{
						position246, thunkPosition246 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDiv]() {
							goto l246
						}
						goto l243
					l246:
						position, thunkPosition = position246, thunkPosition246
					}
					if !matchDot() {
						goto l243
					}
				}
			l244:
				goto l242
			l243:
				position, thunkPosition = position243, thunkPosition243
			}
			if !p.rules[ruleHtmlBlockCloseDiv]() {
				goto l241
			}
			return true
		l241:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 45 HtmlBlockOpenDl <- ('<' Spnl ('dl' / 'DL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l247
			}
			if !p.rules[ruleSpnl]() {
				goto l247
			}
			{
				position248, thunkPosition248 := position, thunkPosition
				if !matchString("dl") {
					goto l249
				}
				goto l248
			l249:
				position, thunkPosition = position248, thunkPosition248
				if !matchString("DL") {
					goto l247
				}
			}
		l248:
			if !p.rules[ruleSpnl]() {
				goto l247
			}
		l250:
			{
				position251, thunkPosition251 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l251
				}
				goto l250
			l251:
				position, thunkPosition = position251, thunkPosition251
			}
			if !matchChar('>') {
				goto l247
			}
			return true
		l247:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 46 HtmlBlockCloseDl <- ('<' Spnl '/' ('dl' / 'DL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l252
			}
			if !p.rules[ruleSpnl]() {
				goto l252
			}
			if !matchChar('/') {
				goto l252
			}
			{
				position253, thunkPosition253 := position, thunkPosition
				if !matchString("dl") {
					goto l254
				}
				goto l253
			l254:
				position, thunkPosition = position253, thunkPosition253
				if !matchString("DL") {
					goto l252
				}
			}
		l253:
			if !p.rules[ruleSpnl]() {
				goto l252
			}
			if !matchChar('>') {
				goto l252
			}
			return true
		l252:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 47 HtmlBlockDl <- (HtmlBlockOpenDl (HtmlBlockDl / (!HtmlBlockCloseDl .))* HtmlBlockCloseDl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDl]() {
				goto l255
			}
		l256:
			{
				position257, thunkPosition257 := position, thunkPosition
				{
					position258, thunkPosition258 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDl]() {
						goto l259
					}
					goto l258
				l259:
					position, thunkPosition = position258, thunkPosition258
					{
						position260, thunkPosition260 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDl]() {
							goto l260
						}
						goto l257
					l260:
						position, thunkPosition = position260, thunkPosition260
					}
					if !matchDot() {
						goto l257
					}
				}
			l258:
				goto l256
			l257:
				position, thunkPosition = position257, thunkPosition257
			}
			if !p.rules[ruleHtmlBlockCloseDl]() {
				goto l255
			}
			return true
		l255:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 48 HtmlBlockOpenFieldset <- ('<' Spnl ('fieldset' / 'FIELDSET') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l261
			}
			if !p.rules[ruleSpnl]() {
				goto l261
			}
			{
				position262, thunkPosition262 := position, thunkPosition
				if !matchString("fieldset") {
					goto l263
				}
				goto l262
			l263:
				position, thunkPosition = position262, thunkPosition262
				if !matchString("FIELDSET") {
					goto l261
				}
			}
		l262:
			if !p.rules[ruleSpnl]() {
				goto l261
			}
		l264:
			{
				position265, thunkPosition265 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l265
				}
				goto l264
			l265:
				position, thunkPosition = position265, thunkPosition265
			}
			if !matchChar('>') {
				goto l261
			}
			return true
		l261:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 49 HtmlBlockCloseFieldset <- ('<' Spnl '/' ('fieldset' / 'FIELDSET') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position267, thunkPosition267 := position, thunkPosition
				if !matchString("fieldset") {
					goto l268
				}
				goto l267
			l268:
				position, thunkPosition = position267, thunkPosition267
				if !matchString("FIELDSET") {
					goto l266
				}
			}
		l267:
			if !p.rules[ruleSpnl]() {
				goto l266
			}
			if !matchChar('>') {
				goto l266
			}
			return true
		l266:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 50 HtmlBlockFieldset <- (HtmlBlockOpenFieldset (HtmlBlockFieldset / (!HtmlBlockCloseFieldset .))* HtmlBlockCloseFieldset) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenFieldset]() {
				goto l269
			}
		l270:
			{
				position271, thunkPosition271 := position, thunkPosition
				{
					position272, thunkPosition272 := position, thunkPosition
					if !p.rules[ruleHtmlBlockFieldset]() {
						goto l273
					}
					goto l272
				l273:
					position, thunkPosition = position272, thunkPosition272
					{
						position274, thunkPosition274 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseFieldset]() {
							goto l274
						}
						goto l271
					l274:
						position, thunkPosition = position274, thunkPosition274
					}
					if !matchDot() {
						goto l271
					}
				}
			l272:
				goto l270
			l271:
				position, thunkPosition = position271, thunkPosition271
			}
			if !p.rules[ruleHtmlBlockCloseFieldset]() {
				goto l269
			}
			return true
		l269:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 51 HtmlBlockOpenForm <- ('<' Spnl ('form' / 'FORM') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l275
			}
			if !p.rules[ruleSpnl]() {
				goto l275
			}
			{
				position276, thunkPosition276 := position, thunkPosition
				if !matchString("form") {
					goto l277
				}
				goto l276
			l277:
				position, thunkPosition = position276, thunkPosition276
				if !matchString("FORM") {
					goto l275
				}
			}
		l276:
			if !p.rules[ruleSpnl]() {
				goto l275
			}
		l278:
			{
				position279, thunkPosition279 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l279
				}
				goto l278
			l279:
				position, thunkPosition = position279, thunkPosition279
			}
			if !matchChar('>') {
				goto l275
			}
			return true
		l275:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 52 HtmlBlockCloseForm <- ('<' Spnl '/' ('form' / 'FORM') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l280
			}
			if !p.rules[ruleSpnl]() {
				goto l280
			}
			if !matchChar('/') {
				goto l280
			}
			{
				position281, thunkPosition281 := position, thunkPosition
				if !matchString("form") {
					goto l282
				}
				goto l281
			l282:
				position, thunkPosition = position281, thunkPosition281
				if !matchString("FORM") {
					goto l280
				}
			}
		l281:
			if !p.rules[ruleSpnl]() {
				goto l280
			}
			if !matchChar('>') {
				goto l280
			}
			return true
		l280:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 53 HtmlBlockForm <- (HtmlBlockOpenForm (HtmlBlockForm / (!HtmlBlockCloseForm .))* HtmlBlockCloseForm) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenForm]() {
				goto l283
			}
		l284:
			{
				position285, thunkPosition285 := position, thunkPosition
				{
					position286, thunkPosition286 := position, thunkPosition
					if !p.rules[ruleHtmlBlockForm]() {
						goto l287
					}
					goto l286
				l287:
					position, thunkPosition = position286, thunkPosition286
					{
						position288, thunkPosition288 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseForm]() {
							goto l288
						}
						goto l285
					l288:
						position, thunkPosition = position288, thunkPosition288
					}
					if !matchDot() {
						goto l285
					}
				}
			l286:
				goto l284
			l285:
				position, thunkPosition = position285, thunkPosition285
			}
			if !p.rules[ruleHtmlBlockCloseForm]() {
				goto l283
			}
			return true
		l283:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 54 HtmlBlockOpenH1 <- ('<' Spnl ('h1' / 'H1') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l289
			}
			if !p.rules[ruleSpnl]() {
				goto l289
			}
			{
				position290, thunkPosition290 := position, thunkPosition
				if !matchString("h1") {
					goto l291
				}
				goto l290
			l291:
				position, thunkPosition = position290, thunkPosition290
				if !matchString("H1") {
					goto l289
				}
			}
		l290:
			if !p.rules[ruleSpnl]() {
				goto l289
			}
		l292:
			{
				position293, thunkPosition293 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l293
				}
				goto l292
			l293:
				position, thunkPosition = position293, thunkPosition293
			}
			if !matchChar('>') {
				goto l289
			}
			return true
		l289:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 55 HtmlBlockCloseH1 <- ('<' Spnl '/' ('h1' / 'H1') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position295, thunkPosition295 := position, thunkPosition
				if !matchString("h1") {
					goto l296
				}
				goto l295
			l296:
				position, thunkPosition = position295, thunkPosition295
				if !matchString("H1") {
					goto l294
				}
			}
		l295:
			if !p.rules[ruleSpnl]() {
				goto l294
			}
			if !matchChar('>') {
				goto l294
			}
			return true
		l294:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 56 HtmlBlockH1 <- (HtmlBlockOpenH1 (HtmlBlockH1 / (!HtmlBlockCloseH1 .))* HtmlBlockCloseH1) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH1]() {
				goto l297
			}
		l298:
			{
				position299, thunkPosition299 := position, thunkPosition
				{
					position300, thunkPosition300 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH1]() {
						goto l301
					}
					goto l300
				l301:
					position, thunkPosition = position300, thunkPosition300
					{
						position302, thunkPosition302 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH1]() {
							goto l302
						}
						goto l299
					l302:
						position, thunkPosition = position302, thunkPosition302
					}
					if !matchDot() {
						goto l299
					}
				}
			l300:
				goto l298
			l299:
				position, thunkPosition = position299, thunkPosition299
			}
			if !p.rules[ruleHtmlBlockCloseH1]() {
				goto l297
			}
			return true
		l297:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 57 HtmlBlockOpenH2 <- ('<' Spnl ('h2' / 'H2') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l303
			}
			if !p.rules[ruleSpnl]() {
				goto l303
			}
			{
				position304, thunkPosition304 := position, thunkPosition
				if !matchString("h2") {
					goto l305
				}
				goto l304
			l305:
				position, thunkPosition = position304, thunkPosition304
				if !matchString("H2") {
					goto l303
				}
			}
		l304:
			if !p.rules[ruleSpnl]() {
				goto l303
			}
		l306:
			{
				position307, thunkPosition307 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l307
				}
				goto l306
			l307:
				position, thunkPosition = position307, thunkPosition307
			}
			if !matchChar('>') {
				goto l303
			}
			return true
		l303:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 58 HtmlBlockCloseH2 <- ('<' Spnl '/' ('h2' / 'H2') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l308
			}
			if !p.rules[ruleSpnl]() {
				goto l308
			}
			if !matchChar('/') {
				goto l308
			}
			{
				position309, thunkPosition309 := position, thunkPosition
				if !matchString("h2") {
					goto l310
				}
				goto l309
			l310:
				position, thunkPosition = position309, thunkPosition309
				if !matchString("H2") {
					goto l308
				}
			}
		l309:
			if !p.rules[ruleSpnl]() {
				goto l308
			}
			if !matchChar('>') {
				goto l308
			}
			return true
		l308:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 59 HtmlBlockH2 <- (HtmlBlockOpenH2 (HtmlBlockH2 / (!HtmlBlockCloseH2 .))* HtmlBlockCloseH2) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH2]() {
				goto l311
			}
		l312:
			{
				position313, thunkPosition313 := position, thunkPosition
				{
					position314, thunkPosition314 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH2]() {
						goto l315
					}
					goto l314
				l315:
					position, thunkPosition = position314, thunkPosition314
					{
						position316, thunkPosition316 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH2]() {
							goto l316
						}
						goto l313
					l316:
						position, thunkPosition = position316, thunkPosition316
					}
					if !matchDot() {
						goto l313
					}
				}
			l314:
				goto l312
			l313:
				position, thunkPosition = position313, thunkPosition313
			}
			if !p.rules[ruleHtmlBlockCloseH2]() {
				goto l311
			}
			return true
		l311:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 60 HtmlBlockOpenH3 <- ('<' Spnl ('h3' / 'H3') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l317
			}
			if !p.rules[ruleSpnl]() {
				goto l317
			}
			{
				position318, thunkPosition318 := position, thunkPosition
				if !matchString("h3") {
					goto l319
				}
				goto l318
			l319:
				position, thunkPosition = position318, thunkPosition318
				if !matchString("H3") {
					goto l317
				}
			}
		l318:
			if !p.rules[ruleSpnl]() {
				goto l317
			}
		l320:
			{
				position321, thunkPosition321 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l321
				}
				goto l320
			l321:
				position, thunkPosition = position321, thunkPosition321
			}
			if !matchChar('>') {
				goto l317
			}
			return true
		l317:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 61 HtmlBlockCloseH3 <- ('<' Spnl '/' ('h3' / 'H3') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l322
			}
			if !p.rules[ruleSpnl]() {
				goto l322
			}
			if !matchChar('/') {
				goto l322
			}
			{
				position323, thunkPosition323 := position, thunkPosition
				if !matchString("h3") {
					goto l324
				}
				goto l323
			l324:
				position, thunkPosition = position323, thunkPosition323
				if !matchString("H3") {
					goto l322
				}
			}
		l323:
			if !p.rules[ruleSpnl]() {
				goto l322
			}
			if !matchChar('>') {
				goto l322
			}
			return true
		l322:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 62 HtmlBlockH3 <- (HtmlBlockOpenH3 (HtmlBlockH3 / (!HtmlBlockCloseH3 .))* HtmlBlockCloseH3) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH3]() {
				goto l325
			}
		l326:
			{
				position327, thunkPosition327 := position, thunkPosition
				{
					position328, thunkPosition328 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH3]() {
						goto l329
					}
					goto l328
				l329:
					position, thunkPosition = position328, thunkPosition328
					{
						position330, thunkPosition330 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH3]() {
							goto l330
						}
						goto l327
					l330:
						position, thunkPosition = position330, thunkPosition330
					}
					if !matchDot() {
						goto l327
					}
				}
			l328:
				goto l326
			l327:
				position, thunkPosition = position327, thunkPosition327
			}
			if !p.rules[ruleHtmlBlockCloseH3]() {
				goto l325
			}
			return true
		l325:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 63 HtmlBlockOpenH4 <- ('<' Spnl ('h4' / 'H4') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l331
			}
			if !p.rules[ruleSpnl]() {
				goto l331
			}
			{
				position332, thunkPosition332 := position, thunkPosition
				if !matchString("h4") {
					goto l333
				}
				goto l332
			l333:
				position, thunkPosition = position332, thunkPosition332
				if !matchString("H4") {
					goto l331
				}
			}
		l332:
			if !p.rules[ruleSpnl]() {
				goto l331
			}
		l334:
			{
				position335, thunkPosition335 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l335
				}
				goto l334
			l335:
				position, thunkPosition = position335, thunkPosition335
			}
			if !matchChar('>') {
				goto l331
			}
			return true
		l331:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 64 HtmlBlockCloseH4 <- ('<' Spnl '/' ('h4' / 'H4') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l336
			}
			if !p.rules[ruleSpnl]() {
				goto l336
			}
			if !matchChar('/') {
				goto l336
			}
			{
				position337, thunkPosition337 := position, thunkPosition
				if !matchString("h4") {
					goto l338
				}
				goto l337
			l338:
				position, thunkPosition = position337, thunkPosition337
				if !matchString("H4") {
					goto l336
				}
			}
		l337:
			if !p.rules[ruleSpnl]() {
				goto l336
			}
			if !matchChar('>') {
				goto l336
			}
			return true
		l336:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 65 HtmlBlockH4 <- (HtmlBlockOpenH4 (HtmlBlockH4 / (!HtmlBlockCloseH4 .))* HtmlBlockCloseH4) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH4]() {
				goto l339
			}
		l340:
			{
				position341, thunkPosition341 := position, thunkPosition
				{
					position342, thunkPosition342 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH4]() {
						goto l343
					}
					goto l342
				l343:
					position, thunkPosition = position342, thunkPosition342
					{
						position344, thunkPosition344 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH4]() {
							goto l344
						}
						goto l341
					l344:
						position, thunkPosition = position344, thunkPosition344
					}
					if !matchDot() {
						goto l341
					}
				}
			l342:
				goto l340
			l341:
				position, thunkPosition = position341, thunkPosition341
			}
			if !p.rules[ruleHtmlBlockCloseH4]() {
				goto l339
			}
			return true
		l339:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 66 HtmlBlockOpenH5 <- ('<' Spnl ('h5' / 'H5') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l345
			}
			if !p.rules[ruleSpnl]() {
				goto l345
			}
			{
				position346, thunkPosition346 := position, thunkPosition
				if !matchString("h5") {
					goto l347
				}
				goto l346
			l347:
				position, thunkPosition = position346, thunkPosition346
				if !matchString("H5") {
					goto l345
				}
			}
		l346:
			if !p.rules[ruleSpnl]() {
				goto l345
			}
		l348:
			{
				position349, thunkPosition349 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l349
				}
				goto l348
			l349:
				position, thunkPosition = position349, thunkPosition349
			}
			if !matchChar('>') {
				goto l345
			}
			return true
		l345:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 67 HtmlBlockCloseH5 <- ('<' Spnl '/' ('h5' / 'H5') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position351, thunkPosition351 := position, thunkPosition
				if !matchString("h5") {
					goto l352
				}
				goto l351
			l352:
				position, thunkPosition = position351, thunkPosition351
				if !matchString("H5") {
					goto l350
				}
			}
		l351:
			if !p.rules[ruleSpnl]() {
				goto l350
			}
			if !matchChar('>') {
				goto l350
			}
			return true
		l350:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 68 HtmlBlockH5 <- (HtmlBlockOpenH5 (HtmlBlockH5 / (!HtmlBlockCloseH5 .))* HtmlBlockCloseH5) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH5]() {
				goto l353
			}
		l354:
			{
				position355, thunkPosition355 := position, thunkPosition
				{
					position356, thunkPosition356 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH5]() {
						goto l357
					}
					goto l356
				l357:
					position, thunkPosition = position356, thunkPosition356
					{
						position358, thunkPosition358 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH5]() {
							goto l358
						}
						goto l355
					l358:
						position, thunkPosition = position358, thunkPosition358
					}
					if !matchDot() {
						goto l355
					}
				}
			l356:
				goto l354
			l355:
				position, thunkPosition = position355, thunkPosition355
			}
			if !p.rules[ruleHtmlBlockCloseH5]() {
				goto l353
			}
			return true
		l353:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 69 HtmlBlockOpenH6 <- ('<' Spnl ('h6' / 'H6') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l359
			}
			if !p.rules[ruleSpnl]() {
				goto l359
			}
			{
				position360, thunkPosition360 := position, thunkPosition
				if !matchString("h6") {
					goto l361
				}
				goto l360
			l361:
				position, thunkPosition = position360, thunkPosition360
				if !matchString("H6") {
					goto l359
				}
			}
		l360:
			if !p.rules[ruleSpnl]() {
				goto l359
			}
		l362:
			{
				position363, thunkPosition363 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l363
				}
				goto l362
			l363:
				position, thunkPosition = position363, thunkPosition363
			}
			if !matchChar('>') {
				goto l359
			}
			return true
		l359:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 70 HtmlBlockCloseH6 <- ('<' Spnl '/' ('h6' / 'H6') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l364
			}
			if !p.rules[ruleSpnl]() {
				goto l364
			}
			if !matchChar('/') {
				goto l364
			}
			{
				position365, thunkPosition365 := position, thunkPosition
				if !matchString("h6") {
					goto l366
				}
				goto l365
			l366:
				position, thunkPosition = position365, thunkPosition365
				if !matchString("H6") {
					goto l364
				}
			}
		l365:
			if !p.rules[ruleSpnl]() {
				goto l364
			}
			if !matchChar('>') {
				goto l364
			}
			return true
		l364:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 71 HtmlBlockH6 <- (HtmlBlockOpenH6 (HtmlBlockH6 / (!HtmlBlockCloseH6 .))* HtmlBlockCloseH6) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH6]() {
				goto l367
			}
		l368:
			{
				position369, thunkPosition369 := position, thunkPosition
				{
					position370, thunkPosition370 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH6]() {
						goto l371
					}
					goto l370
				l371:
					position, thunkPosition = position370, thunkPosition370
					{
						position372, thunkPosition372 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH6]() {
							goto l372
						}
						goto l369
					l372:
						position, thunkPosition = position372, thunkPosition372
					}
					if !matchDot() {
						goto l369
					}
				}
			l370:
				goto l368
			l369:
				position, thunkPosition = position369, thunkPosition369
			}
			if !p.rules[ruleHtmlBlockCloseH6]() {
				goto l367
			}
			return true
		l367:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 72 HtmlBlockOpenMenu <- ('<' Spnl ('menu' / 'MENU') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l373
			}
			if !p.rules[ruleSpnl]() {
				goto l373
			}
			{
				position374, thunkPosition374 := position, thunkPosition
				if !matchString("menu") {
					goto l375
				}
				goto l374
			l375:
				position, thunkPosition = position374, thunkPosition374
				if !matchString("MENU") {
					goto l373
				}
			}
		l374:
			if !p.rules[ruleSpnl]() {
				goto l373
			}
		l376:
			{
				position377, thunkPosition377 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l377
				}
				goto l376
			l377:
				position, thunkPosition = position377, thunkPosition377
			}
			if !matchChar('>') {
				goto l373
			}
			return true
		l373:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 73 HtmlBlockCloseMenu <- ('<' Spnl '/' ('menu' / 'MENU') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position379, thunkPosition379 := position, thunkPosition
				if !matchString("menu") {
					goto l380
				}
				goto l379
			l380:
				position, thunkPosition = position379, thunkPosition379
				if !matchString("MENU") {
					goto l378
				}
			}
		l379:
			if !p.rules[ruleSpnl]() {
				goto l378
			}
			if !matchChar('>') {
				goto l378
			}
			return true
		l378:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 74 HtmlBlockMenu <- (HtmlBlockOpenMenu (HtmlBlockMenu / (!HtmlBlockCloseMenu .))* HtmlBlockCloseMenu) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenMenu]() {
				goto l381
			}
		l382:
			{
				position383, thunkPosition383 := position, thunkPosition
				{
					position384, thunkPosition384 := position, thunkPosition
					if !p.rules[ruleHtmlBlockMenu]() {
						goto l385
					}
					goto l384
				l385:
					position, thunkPosition = position384, thunkPosition384
					{
						position386, thunkPosition386 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseMenu]() {
							goto l386
						}
						goto l383
					l386:
						position, thunkPosition = position386, thunkPosition386
					}
					if !matchDot() {
						goto l383
					}
				}
			l384:
				goto l382
			l383:
				position, thunkPosition = position383, thunkPosition383
			}
			if !p.rules[ruleHtmlBlockCloseMenu]() {
				goto l381
			}
			return true
		l381:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 75 HtmlBlockOpenNoframes <- ('<' Spnl ('noframes' / 'NOFRAMES') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l387
			}
			if !p.rules[ruleSpnl]() {
				goto l387
			}
			{
				position388, thunkPosition388 := position, thunkPosition
				if !matchString("noframes") {
					goto l389
				}
				goto l388
			l389:
				position, thunkPosition = position388, thunkPosition388
				if !matchString("NOFRAMES") {
					goto l387
				}
			}
		l388:
			if !p.rules[ruleSpnl]() {
				goto l387
			}
		l390:
			{
				position391, thunkPosition391 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l391
				}
				goto l390
			l391:
				position, thunkPosition = position391, thunkPosition391
			}
			if !matchChar('>') {
				goto l387
			}
			return true
		l387:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 76 HtmlBlockCloseNoframes <- ('<' Spnl '/' ('noframes' / 'NOFRAMES') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l392
			}
			if !p.rules[ruleSpnl]() {
				goto l392
			}
			if !matchChar('/') {
				goto l392
			}
			{
				position393, thunkPosition393 := position, thunkPosition
				if !matchString("noframes") {
					goto l394
				}
				goto l393
			l394:
				position, thunkPosition = position393, thunkPosition393
				if !matchString("NOFRAMES") {
					goto l392
				}
			}
		l393:
			if !p.rules[ruleSpnl]() {
				goto l392
			}
			if !matchChar('>') {
				goto l392
			}
			return true
		l392:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 77 HtmlBlockNoframes <- (HtmlBlockOpenNoframes (HtmlBlockNoframes / (!HtmlBlockCloseNoframes .))* HtmlBlockCloseNoframes) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenNoframes]() {
				goto l395
			}
		l396:
			{
				position397, thunkPosition397 := position, thunkPosition
				{
					position398, thunkPosition398 := position, thunkPosition
					if !p.rules[ruleHtmlBlockNoframes]() {
						goto l399
					}
					goto l398
				l399:
					position, thunkPosition = position398, thunkPosition398
					{
						position400, thunkPosition400 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseNoframes]() {
							goto l400
						}
						goto l397
					l400:
						position, thunkPosition = position400, thunkPosition400
					}
					if !matchDot() {
						goto l397
					}
				}
			l398:
				goto l396
			l397:
				position, thunkPosition = position397, thunkPosition397
			}
			if !p.rules[ruleHtmlBlockCloseNoframes]() {
				goto l395
			}
			return true
		l395:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 78 HtmlBlockOpenNoscript <- ('<' Spnl ('noscript' / 'NOSCRIPT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l401
			}
			if !p.rules[ruleSpnl]() {
				goto l401
			}
			{
				position402, thunkPosition402 := position, thunkPosition
				if !matchString("noscript") {
					goto l403
				}
				goto l402
			l403:
				position, thunkPosition = position402, thunkPosition402
				if !matchString("NOSCRIPT") {
					goto l401
				}
			}
		l402:
			if !p.rules[ruleSpnl]() {
				goto l401
			}
		l404:
			{
				position405, thunkPosition405 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l405
				}
				goto l404
			l405:
				position, thunkPosition = position405, thunkPosition405
			}
			if !matchChar('>') {
				goto l401
			}
			return true
		l401:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 79 HtmlBlockCloseNoscript <- ('<' Spnl '/' ('noscript' / 'NOSCRIPT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l406
			}
			if !p.rules[ruleSpnl]() {
				goto l406
			}
			if !matchChar('/') {
				goto l406
			}
			{
				position407, thunkPosition407 := position, thunkPosition
				if !matchString("noscript") {
					goto l408
				}
				goto l407
			l408:
				position, thunkPosition = position407, thunkPosition407
				if !matchString("NOSCRIPT") {
					goto l406
				}
			}
		l407:
			if !p.rules[ruleSpnl]() {
				goto l406
			}
			if !matchChar('>') {
				goto l406
			}
			return true
		l406:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 80 HtmlBlockNoscript <- (HtmlBlockOpenNoscript (HtmlBlockNoscript / (!HtmlBlockCloseNoscript .))* HtmlBlockCloseNoscript) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenNoscript]() {
				goto l409
			}
		l410:
			{
				position411, thunkPosition411 := position, thunkPosition
				{
					position412, thunkPosition412 := position, thunkPosition
					if !p.rules[ruleHtmlBlockNoscript]() {
						goto l413
					}
					goto l412
				l413:
					position, thunkPosition = position412, thunkPosition412
					{
						position414, thunkPosition414 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseNoscript]() {
							goto l414
						}
						goto l411
					l414:
						position, thunkPosition = position414, thunkPosition414
					}
					if !matchDot() {
						goto l411
					}
				}
			l412:
				goto l410
			l411:
				position, thunkPosition = position411, thunkPosition411
			}
			if !p.rules[ruleHtmlBlockCloseNoscript]() {
				goto l409
			}
			return true
		l409:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 81 HtmlBlockOpenOl <- ('<' Spnl ('ol' / 'OL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l415
			}
			if !p.rules[ruleSpnl]() {
				goto l415
			}
			{
				position416, thunkPosition416 := position, thunkPosition
				if !matchString("ol") {
					goto l417
				}
				goto l416
			l417:
				position, thunkPosition = position416, thunkPosition416
				if !matchString("OL") {
					goto l415
				}
			}
		l416:
			if !p.rules[ruleSpnl]() {
				goto l415
			}
		l418:
			{
				position419, thunkPosition419 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l419
				}
				goto l418
			l419:
				position, thunkPosition = position419, thunkPosition419
			}
			if !matchChar('>') {
				goto l415
			}
			return true
		l415:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 82 HtmlBlockCloseOl <- ('<' Spnl '/' ('ol' / 'OL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l420
			}
			if !p.rules[ruleSpnl]() {
				goto l420
			}
			if !matchChar('/') {
				goto l420
			}
			{
				position421, thunkPosition421 := position, thunkPosition
				if !matchString("ol") {
					goto l422
				}
				goto l421
			l422:
				position, thunkPosition = position421, thunkPosition421
				if !matchString("OL") {
					goto l420
				}
			}
		l421:
			if !p.rules[ruleSpnl]() {
				goto l420
			}
			if !matchChar('>') {
				goto l420
			}
			return true
		l420:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 83 HtmlBlockOl <- (HtmlBlockOpenOl (HtmlBlockOl / (!HtmlBlockCloseOl .))* HtmlBlockCloseOl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenOl]() {
				goto l423
			}
		l424:
			{
				position425, thunkPosition425 := position, thunkPosition
				{
					position426, thunkPosition426 := position, thunkPosition
					if !p.rules[ruleHtmlBlockOl]() {
						goto l427
					}
					goto l426
				l427:
					position, thunkPosition = position426, thunkPosition426
					{
						position428, thunkPosition428 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseOl]() {
							goto l428
						}
						goto l425
					l428:
						position, thunkPosition = position428, thunkPosition428
					}
					if !matchDot() {
						goto l425
					}
				}
			l426:
				goto l424
			l425:
				position, thunkPosition = position425, thunkPosition425
			}
			if !p.rules[ruleHtmlBlockCloseOl]() {
				goto l423
			}
			return true
		l423:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 84 HtmlBlockOpenP <- ('<' Spnl ('p' / 'P') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l429
			}
			if !p.rules[ruleSpnl]() {
				goto l429
			}
			{
				position430, thunkPosition430 := position, thunkPosition
				if !matchChar('p') {
					goto l431
				}
				goto l430
			l431:
				position, thunkPosition = position430, thunkPosition430
				if !matchChar('P') {
					goto l429
				}
			}
		l430:
			if !p.rules[ruleSpnl]() {
				goto l429
			}
		l432:
			{
				position433, thunkPosition433 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l433
				}
				goto l432
			l433:
				position, thunkPosition = position433, thunkPosition433
			}
			if !matchChar('>') {
				goto l429
			}
			return true
		l429:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 85 HtmlBlockCloseP <- ('<' Spnl '/' ('p' / 'P') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position435, thunkPosition435 := position, thunkPosition
				if !matchChar('p') {
					goto l436
				}
				goto l435
			l436:
				position, thunkPosition = position435, thunkPosition435
				if !matchChar('P') {
					goto l434
				}
			}
		l435:
			if !p.rules[ruleSpnl]() {
				goto l434
			}
			if !matchChar('>') {
				goto l434
			}
			return true
		l434:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 86 HtmlBlockP <- (HtmlBlockOpenP (HtmlBlockP / (!HtmlBlockCloseP .))* HtmlBlockCloseP) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenP]() {
				goto l437
			}
		l438:
			{
				position439, thunkPosition439 := position, thunkPosition
				{
					position440, thunkPosition440 := position, thunkPosition
					if !p.rules[ruleHtmlBlockP]() {
						goto l441
					}
					goto l440
				l441:
					position, thunkPosition = position440, thunkPosition440
					{
						position442, thunkPosition442 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseP]() {
							goto l442
						}
						goto l439
					l442:
						position, thunkPosition = position442, thunkPosition442
					}
					if !matchDot() {
						goto l439
					}
				}
			l440:
				goto l438
			l439:
				position, thunkPosition = position439, thunkPosition439
			}
			if !p.rules[ruleHtmlBlockCloseP]() {
				goto l437
			}
			return true
		l437:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 87 HtmlBlockOpenPre <- ('<' Spnl ('pre' / 'PRE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l443
			}
			if !p.rules[ruleSpnl]() {
				goto l443
			}
			{
				position444, thunkPosition444 := position, thunkPosition
				if !matchString("pre") {
					goto l445
				}
				goto l444
			l445:
				position, thunkPosition = position444, thunkPosition444
				if !matchString("PRE") {
					goto l443
				}
			}
		l444:
			if !p.rules[ruleSpnl]() {
				goto l443
			}
		l446:
			{
				position447, thunkPosition447 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l447
				}
				goto l446
			l447:
				position, thunkPosition = position447, thunkPosition447
			}
			if !matchChar('>') {
				goto l443
			}
			return true
		l443:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 88 HtmlBlockClosePre <- ('<' Spnl '/' ('pre' / 'PRE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l448
			}
			if !p.rules[ruleSpnl]() {
				goto l448
			}
			if !matchChar('/') {
				goto l448
			}
			{
				position449, thunkPosition449 := position, thunkPosition
				if !matchString("pre") {
					goto l450
				}
				goto l449
			l450:
				position, thunkPosition = position449, thunkPosition449
				if !matchString("PRE") {
					goto l448
				}
			}
		l449:
			if !p.rules[ruleSpnl]() {
				goto l448
			}
			if !matchChar('>') {
				goto l448
			}
			return true
		l448:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 89 HtmlBlockPre <- (HtmlBlockOpenPre (HtmlBlockPre / (!HtmlBlockClosePre .))* HtmlBlockClosePre) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenPre]() {
				goto l451
			}
		l452:
			{
				position453, thunkPosition453 := position, thunkPosition
				{
					position454, thunkPosition454 := position, thunkPosition
					if !p.rules[ruleHtmlBlockPre]() {
						goto l455
					}
					goto l454
				l455:
					position, thunkPosition = position454, thunkPosition454
					{
						position456, thunkPosition456 := position, thunkPosition
						if !p.rules[ruleHtmlBlockClosePre]() {
							goto l456
						}
						goto l453
					l456:
						position, thunkPosition = position456, thunkPosition456
					}
					if !matchDot() {
						goto l453
					}
				}
			l454:
				goto l452
			l453:
				position, thunkPosition = position453, thunkPosition453
			}
			if !p.rules[ruleHtmlBlockClosePre]() {
				goto l451
			}
			return true
		l451:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 90 HtmlBlockOpenTable <- ('<' Spnl ('table' / 'TABLE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l457
			}
			if !p.rules[ruleSpnl]() {
				goto l457
			}
			{
				position458, thunkPosition458 := position, thunkPosition
				if !matchString("table") {
					goto l459
				}
				goto l458
			l459:
				position, thunkPosition = position458, thunkPosition458
				if !matchString("TABLE") {
					goto l457
				}
			}
		l458:
			if !p.rules[ruleSpnl]() {
				goto l457
			}
		l460:
			{
				position461, thunkPosition461 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l461
				}
				goto l460
			l461:
				position, thunkPosition = position461, thunkPosition461
			}
			if !matchChar('>') {
				goto l457
			}
			return true
		l457:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 91 HtmlBlockCloseTable <- ('<' Spnl '/' ('table' / 'TABLE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position463, thunkPosition463 := position, thunkPosition
				if !matchString("table") {
					goto l464
				}
				goto l463
			l464:
				position, thunkPosition = position463, thunkPosition463
				if !matchString("TABLE") {
					goto l462
				}
			}
		l463:
			if !p.rules[ruleSpnl]() {
				goto l462
			}
			if !matchChar('>') {
				goto l462
			}
			return true
		l462:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 92 HtmlBlockTable <- (HtmlBlockOpenTable (HtmlBlockTable / (!HtmlBlockCloseTable .))* HtmlBlockCloseTable) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTable]() {
				goto l465
			}
		l466:
			{
				position467, thunkPosition467 := position, thunkPosition
				{
					position468, thunkPosition468 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTable]() {
						goto l469
					}
					goto l468
				l469:
					position, thunkPosition = position468, thunkPosition468
					{
						position470, thunkPosition470 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTable]() {
							goto l470
						}
						goto l467
					l470:
						position, thunkPosition = position470, thunkPosition470
					}
					if !matchDot() {
						goto l467
					}
				}
			l468:
				goto l466
			l467:
				position, thunkPosition = position467, thunkPosition467
			}
			if !p.rules[ruleHtmlBlockCloseTable]() {
				goto l465
			}
			return true
		l465:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 93 HtmlBlockOpenUl <- ('<' Spnl ('ul' / 'UL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l471
			}
			if !p.rules[ruleSpnl]() {
				goto l471
			}
			{
				position472, thunkPosition472 := position, thunkPosition
				if !matchString("ul") {
					goto l473
				}
				goto l472
			l473:
				position, thunkPosition = position472, thunkPosition472
				if !matchString("UL") {
					goto l471
				}
			}
		l472:
			if !p.rules[ruleSpnl]() {
				goto l471
			}
		l474:
			{
				position475, thunkPosition475 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l475
				}
				goto l474
			l475:
				position, thunkPosition = position475, thunkPosition475
			}
			if !matchChar('>') {
				goto l471
			}
			return true
		l471:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 94 HtmlBlockCloseUl <- ('<' Spnl '/' ('ul' / 'UL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l476
			}
			if !p.rules[ruleSpnl]() {
				goto l476
			}
			if !matchChar('/') {
				goto l476
			}
			{
				position477, thunkPosition477 := position, thunkPosition
				if !matchString("ul") {
					goto l478
				}
				goto l477
			l478:
				position, thunkPosition = position477, thunkPosition477
				if !matchString("UL") {
					goto l476
				}
			}
		l477:
			if !p.rules[ruleSpnl]() {
				goto l476
			}
			if !matchChar('>') {
				goto l476
			}
			return true
		l476:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 95 HtmlBlockUl <- (HtmlBlockOpenUl (HtmlBlockUl / (!HtmlBlockCloseUl .))* HtmlBlockCloseUl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenUl]() {
				goto l479
			}
		l480:
			{
				position481, thunkPosition481 := position, thunkPosition
				{
					position482, thunkPosition482 := position, thunkPosition
					if !p.rules[ruleHtmlBlockUl]() {
						goto l483
					}
					goto l482
				l483:
					position, thunkPosition = position482, thunkPosition482
					{
						position484, thunkPosition484 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseUl]() {
							goto l484
						}
						goto l481
					l484:
						position, thunkPosition = position484, thunkPosition484
					}
					if !matchDot() {
						goto l481
					}
				}
			l482:
				goto l480
			l481:
				position, thunkPosition = position481, thunkPosition481
			}
			if !p.rules[ruleHtmlBlockCloseUl]() {
				goto l479
			}
			return true
		l479:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 96 HtmlBlockOpenDd <- ('<' Spnl ('dd' / 'DD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l485
			}
			if !p.rules[ruleSpnl]() {
				goto l485
			}
			{
				position486, thunkPosition486 := position, thunkPosition
				if !matchString("dd") {
					goto l487
				}
				goto l486
			l487:
				position, thunkPosition = position486, thunkPosition486
				if !matchString("DD") {
					goto l485
				}
			}
		l486:
			if !p.rules[ruleSpnl]() {
				goto l485
			}
		l488:
			{
				position489, thunkPosition489 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l489
				}
				goto l488
			l489:
				position, thunkPosition = position489, thunkPosition489
			}
			if !matchChar('>') {
				goto l485
			}
			return true
		l485:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 97 HtmlBlockCloseDd <- ('<' Spnl '/' ('dd' / 'DD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l490
			}
			if !p.rules[ruleSpnl]() {
				goto l490
			}
			if !matchChar('/') {
				goto l490
			}
			{
				position491, thunkPosition491 := position, thunkPosition
				if !matchString("dd") {
					goto l492
				}
				goto l491
			l492:
				position, thunkPosition = position491, thunkPosition491
				if !matchString("DD") {
					goto l490
				}
			}
		l491:
			if !p.rules[ruleSpnl]() {
				goto l490
			}
			if !matchChar('>') {
				goto l490
			}
			return true
		l490:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 98 HtmlBlockDd <- (HtmlBlockOpenDd (HtmlBlockDd / (!HtmlBlockCloseDd .))* HtmlBlockCloseDd) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDd]() {
				goto l493
			}
		l494:
			{
				position495, thunkPosition495 := position, thunkPosition
				{
					position496, thunkPosition496 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDd]() {
						goto l497
					}
					goto l496
				l497:
					position, thunkPosition = position496, thunkPosition496
					{
						position498, thunkPosition498 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDd]() {
							goto l498
						}
						goto l495
					l498:
						position, thunkPosition = position498, thunkPosition498
					}
					if !matchDot() {
						goto l495
					}
				}
			l496:
				goto l494
			l495:
				position, thunkPosition = position495, thunkPosition495
			}
			if !p.rules[ruleHtmlBlockCloseDd]() {
				goto l493
			}
			return true
		l493:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 99 HtmlBlockOpenDt <- ('<' Spnl ('dt' / 'DT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l499
			}
			if !p.rules[ruleSpnl]() {
				goto l499
			}
			{
				position500, thunkPosition500 := position, thunkPosition
				if !matchString("dt") {
					goto l501
				}
				goto l500
			l501:
				position, thunkPosition = position500, thunkPosition500
				if !matchString("DT") {
					goto l499
				}
			}
		l500:
			if !p.rules[ruleSpnl]() {
				goto l499
			}
		l502:
			{
				position503, thunkPosition503 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l503
				}
				goto l502
			l503:
				position, thunkPosition = position503, thunkPosition503
			}
			if !matchChar('>') {
				goto l499
			}
			return true
		l499:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 100 HtmlBlockCloseDt <- ('<' Spnl '/' ('dt' / 'DT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l504
			}
			if !p.rules[ruleSpnl]() {
				goto l504
			}
			if !matchChar('/') {
				goto l504
			}
			{
				position505, thunkPosition505 := position, thunkPosition
				if !matchString("dt") {
					goto l506
				}
				goto l505
			l506:
				position, thunkPosition = position505, thunkPosition505
				if !matchString("DT") {
					goto l504
				}
			}
		l505:
			if !p.rules[ruleSpnl]() {
				goto l504
			}
			if !matchChar('>') {
				goto l504
			}
			return true
		l504:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 101 HtmlBlockDt <- (HtmlBlockOpenDt (HtmlBlockDt / (!HtmlBlockCloseDt .))* HtmlBlockCloseDt) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDt]() {
				goto l507
			}
		l508:
			{
				position509, thunkPosition509 := position, thunkPosition
				{
					position510, thunkPosition510 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDt]() {
						goto l511
					}
					goto l510
				l511:
					position, thunkPosition = position510, thunkPosition510
					{
						position512, thunkPosition512 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDt]() {
							goto l512
						}
						goto l509
					l512:
						position, thunkPosition = position512, thunkPosition512
					}
					if !matchDot() {
						goto l509
					}
				}
			l510:
				goto l508
			l509:
				position, thunkPosition = position509, thunkPosition509
			}
			if !p.rules[ruleHtmlBlockCloseDt]() {
				goto l507
			}
			return true
		l507:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 102 HtmlBlockOpenFrameset <- ('<' Spnl ('frameset' / 'FRAMESET') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l513
			}
			if !p.rules[ruleSpnl]() {
				goto l513
			}
			{
				position514, thunkPosition514 := position, thunkPosition
				if !matchString("frameset") {
					goto l515
				}
				goto l514
			l515:
				position, thunkPosition = position514, thunkPosition514
				if !matchString("FRAMESET") {
					goto l513
				}
			}
		l514:
			if !p.rules[ruleSpnl]() {
				goto l513
			}
		l516:
			{
				position517, thunkPosition517 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l517
				}
				goto l516
			l517:
				position, thunkPosition = position517, thunkPosition517
			}
			if !matchChar('>') {
				goto l513
			}
			return true
		l513:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 103 HtmlBlockCloseFrameset <- ('<' Spnl '/' ('frameset' / 'FRAMESET') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position519, thunkPosition519 := position, thunkPosition
				if !matchString("frameset") {
					goto l520
				}
				goto l519
			l520:
				position, thunkPosition = position519, thunkPosition519
				if !matchString("FRAMESET") {
					goto l518
				}
			}
		l519:
			if !p.rules[ruleSpnl]() {
				goto l518
			}
			if !matchChar('>') {
				goto l518
			}
			return true
		l518:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 104 HtmlBlockFrameset <- (HtmlBlockOpenFrameset (HtmlBlockFrameset / (!HtmlBlockCloseFrameset .))* HtmlBlockCloseFrameset) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenFrameset]() {
				goto l521
			}
		l522:
			{
				position523, thunkPosition523 := position, thunkPosition
				{
					position524, thunkPosition524 := position, thunkPosition
					if !p.rules[ruleHtmlBlockFrameset]() {
						goto l525
					}
					goto l524
				l525:
					position, thunkPosition = position524, thunkPosition524
					{
						position526, thunkPosition526 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseFrameset]() {
							goto l526
						}
						goto l523
					l526:
						position, thunkPosition = position526, thunkPosition526
					}
					if !matchDot() {
						goto l523
					}
				}
			l524:
				goto l522
			l523:
				position, thunkPosition = position523, thunkPosition523
			}
			if !p.rules[ruleHtmlBlockCloseFrameset]() {
				goto l521
			}
			return true
		l521:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 105 HtmlBlockOpenLi <- ('<' Spnl ('li' / 'LI') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l527
			}
			if !p.rules[ruleSpnl]() {
				goto l527
			}
			{
				position528, thunkPosition528 := position, thunkPosition
				if !matchString("li") {
					goto l529
				}
				goto l528
			l529:
				position, thunkPosition = position528, thunkPosition528
				if !matchString("LI") {
					goto l527
				}
			}
		l528:
			if !p.rules[ruleSpnl]() {
				goto l527
			}
		l530:
			{
				position531, thunkPosition531 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l531
				}
				goto l530
			l531:
				position, thunkPosition = position531, thunkPosition531
			}
			if !matchChar('>') {
				goto l527
			}
			return true
		l527:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 106 HtmlBlockCloseLi <- ('<' Spnl '/' ('li' / 'LI') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l532
			}
			if !p.rules[ruleSpnl]() {
				goto l532
			}
			if !matchChar('/') {
				goto l532
			}
			{
				position533, thunkPosition533 := position, thunkPosition
				if !matchString("li") {
					goto l534
				}
				goto l533
			l534:
				position, thunkPosition = position533, thunkPosition533
				if !matchString("LI") {
					goto l532
				}
			}
		l533:
			if !p.rules[ruleSpnl]() {
				goto l532
			}
			if !matchChar('>') {
				goto l532
			}
			return true
		l532:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 107 HtmlBlockLi <- (HtmlBlockOpenLi (HtmlBlockLi / (!HtmlBlockCloseLi .))* HtmlBlockCloseLi) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenLi]() {
				goto l535
			}
		l536:
			{
				position537, thunkPosition537 := position, thunkPosition
				{
					position538, thunkPosition538 := position, thunkPosition
					if !p.rules[ruleHtmlBlockLi]() {
						goto l539
					}
					goto l538
				l539:
					position, thunkPosition = position538, thunkPosition538
					{
						position540, thunkPosition540 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseLi]() {
							goto l540
						}
						goto l537
					l540:
						position, thunkPosition = position540, thunkPosition540
					}
					if !matchDot() {
						goto l537
					}
				}
			l538:
				goto l536
			l537:
				position, thunkPosition = position537, thunkPosition537
			}
			if !p.rules[ruleHtmlBlockCloseLi]() {
				goto l535
			}
			return true
		l535:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 108 HtmlBlockOpenTbody <- ('<' Spnl ('tbody' / 'TBODY') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l541
			}
			if !p.rules[ruleSpnl]() {
				goto l541
			}
			{
				position542, thunkPosition542 := position, thunkPosition
				if !matchString("tbody") {
					goto l543
				}
				goto l542
			l543:
				position, thunkPosition = position542, thunkPosition542
				if !matchString("TBODY") {
					goto l541
				}
			}
		l542:
			if !p.rules[ruleSpnl]() {
				goto l541
			}
		l544:
			{
				position545, thunkPosition545 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l545
				}
				goto l544
			l545:
				position, thunkPosition = position545, thunkPosition545
			}
			if !matchChar('>') {
				goto l541
			}
			return true
		l541:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 109 HtmlBlockCloseTbody <- ('<' Spnl '/' ('tbody' / 'TBODY') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position547, thunkPosition547 := position, thunkPosition
				if !matchString("tbody") {
					goto l548
				}
				goto l547
			l548:
				position, thunkPosition = position547, thunkPosition547
				if !matchString("TBODY") {
					goto l546
				}
			}
		l547:
			if !p.rules[ruleSpnl]() {
				goto l546
			}
			if !matchChar('>') {
				goto l546
			}
			return true
		l546:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 110 HtmlBlockTbody <- (HtmlBlockOpenTbody (HtmlBlockTbody / (!HtmlBlockCloseTbody .))* HtmlBlockCloseTbody) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTbody]() {
				goto l549
			}
		l550:
			{
				position551, thunkPosition551 := position, thunkPosition
				{
					position552, thunkPosition552 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTbody]() {
						goto l553
					}
					goto l552
				l553:
					position, thunkPosition = position552, thunkPosition552
					{
						position554, thunkPosition554 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTbody]() {
							goto l554
						}
						goto l551
					l554:
						position, thunkPosition = position554, thunkPosition554
					}
					if !matchDot() {
						goto l551
					}
				}
			l552:
				goto l550
			l551:
				position, thunkPosition = position551, thunkPosition551
			}
			if !p.rules[ruleHtmlBlockCloseTbody]() {
				goto l549
			}
			return true
		l549:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 111 HtmlBlockOpenTd <- ('<' Spnl ('td' / 'TD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l555
			}
			if !p.rules[ruleSpnl]() {
				goto l555
			}
			{
				position556, thunkPosition556 := position, thunkPosition
				if !matchString("td") {
					goto l557
				}
				goto l556
			l557:
				position, thunkPosition = position556, thunkPosition556
				if !matchString("TD") {
					goto l555
				}
			}
		l556:
			if !p.rules[ruleSpnl]() {
				goto l555
			}
		l558:
			{
				position559, thunkPosition559 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l559
				}
				goto l558
			l559:
				position, thunkPosition = position559, thunkPosition559
			}
			if !matchChar('>') {
				goto l555
			}
			return true
		l555:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 112 HtmlBlockCloseTd <- ('<' Spnl '/' ('td' / 'TD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l560
			}
			if !p.rules[ruleSpnl]() {
				goto l560
			}
			if !matchChar('/') {
				goto l560
			}
			{
				position561, thunkPosition561 := position, thunkPosition
				if !matchString("td") {
					goto l562
				}
				goto l561
			l562:
				position, thunkPosition = position561, thunkPosition561
				if !matchString("TD") {
					goto l560
				}
			}
		l561:
			if !p.rules[ruleSpnl]() {
				goto l560
			}
			if !matchChar('>') {
				goto l560
			}
			return true
		l560:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 113 HtmlBlockTd <- (HtmlBlockOpenTd (HtmlBlockTd / (!HtmlBlockCloseTd .))* HtmlBlockCloseTd) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTd]() {
				goto l563
			}
		l564:
			{
				position565, thunkPosition565 := position, thunkPosition
				{
					position566, thunkPosition566 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTd]() {
						goto l567
					}
					goto l566
				l567:
					position, thunkPosition = position566, thunkPosition566
					{
						position568, thunkPosition568 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTd]() {
							goto l568
						}
						goto l565
					l568:
						position, thunkPosition = position568, thunkPosition568
					}
					if !matchDot() {
						goto l565
					}
				}
			l566:
				goto l564
			l565:
				position, thunkPosition = position565, thunkPosition565
			}
			if !p.rules[ruleHtmlBlockCloseTd]() {
				goto l563
			}
			return true
		l563:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 114 HtmlBlockOpenTfoot <- ('<' Spnl ('tfoot' / 'TFOOT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l569
			}
			if !p.rules[ruleSpnl]() {
				goto l569
			}
			{
				position570, thunkPosition570 := position, thunkPosition
				if !matchString("tfoot") {
					goto l571
				}
				goto l570
			l571:
				position, thunkPosition = position570, thunkPosition570
				if !matchString("TFOOT") {
					goto l569
				}
			}
		l570:
			if !p.rules[ruleSpnl]() {
				goto l569
			}
		l572:
			{
				position573, thunkPosition573 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l573
				}
				goto l572
			l573:
				position, thunkPosition = position573, thunkPosition573
			}
			if !matchChar('>') {
				goto l569
			}
			return true
		l569:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 115 HtmlBlockCloseTfoot <- ('<' Spnl '/' ('tfoot' / 'TFOOT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l574
			}
			if !p.rules[ruleSpnl]() {
				goto l574
			}
			if !matchChar('/') {
				goto l574
			}
			{
				position575, thunkPosition575 := position, thunkPosition
				if !matchString("tfoot") {
					goto l576
				}
				goto l575
			l576:
				position, thunkPosition = position575, thunkPosition575
				if !matchString("TFOOT") {
					goto l574
				}
			}
		l575:
			if !p.rules[ruleSpnl]() {
				goto l574
			}
			if !matchChar('>') {
				goto l574
			}
			return true
		l574:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 116 HtmlBlockTfoot <- (HtmlBlockOpenTfoot (HtmlBlockTfoot / (!HtmlBlockCloseTfoot .))* HtmlBlockCloseTfoot) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTfoot]() {
				goto l577
			}
		l578:
			{
				position579, thunkPosition579 := position, thunkPosition
				{
					position580, thunkPosition580 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTfoot]() {
						goto l581
					}
					goto l580
				l581:
					position, thunkPosition = position580, thunkPosition580
					{
						position582, thunkPosition582 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTfoot]() {
							goto l582
						}
						goto l579
					l582:
						position, thunkPosition = position582, thunkPosition582
					}
					if !matchDot() {
						goto l579
					}
				}
			l580:
				goto l578
			l579:
				position, thunkPosition = position579, thunkPosition579
			}
			if !p.rules[ruleHtmlBlockCloseTfoot]() {
				goto l577
			}
			return true
		l577:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 117 HtmlBlockOpenTh <- ('<' Spnl ('th' / 'TH') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l583
			}
			if !p.rules[ruleSpnl]() {
				goto l583
			}
			{
				position584, thunkPosition584 := position, thunkPosition
				if !matchString("th") {
					goto l585
				}
				goto l584
			l585:
				position, thunkPosition = position584, thunkPosition584
				if !matchString("TH") {
					goto l583
				}
			}
		l584:
			if !p.rules[ruleSpnl]() {
				goto l583
			}
		l586:
			{
				position587, thunkPosition587 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l587
				}
				goto l586
			l587:
				position, thunkPosition = position587, thunkPosition587
			}
			if !matchChar('>') {
				goto l583
			}
			return true
		l583:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 118 HtmlBlockCloseTh <- ('<' Spnl '/' ('th' / 'TH') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l588
			}
			if !p.rules[ruleSpnl]() {
				goto l588
			}
			if !matchChar('/') {
				goto l588
			}
			{
				position589, thunkPosition589 := position, thunkPosition
				if !matchString("th") {
					goto l590
				}
				goto l589
			l590:
				position, thunkPosition = position589, thunkPosition589
				if !matchString("TH") {
					goto l588
				}
			}
		l589:
			if !p.rules[ruleSpnl]() {
				goto l588
			}
			if !matchChar('>') {
				goto l588
			}
			return true
		l588:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 119 HtmlBlockTh <- (HtmlBlockOpenTh (HtmlBlockTh / (!HtmlBlockCloseTh .))* HtmlBlockCloseTh) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTh]() {
				goto l591
			}
		l592:
			{
				position593, thunkPosition593 := position, thunkPosition
				{
					position594, thunkPosition594 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTh]() {
						goto l595
					}
					goto l594
				l595:
					position, thunkPosition = position594, thunkPosition594
					{
						position596, thunkPosition596 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTh]() {
							goto l596
						}
						goto l593
					l596:
						position, thunkPosition = position596, thunkPosition596
					}
					if !matchDot() {
						goto l593
					}
				}
			l594:
				goto l592
			l593:
				position, thunkPosition = position593, thunkPosition593
			}
			if !p.rules[ruleHtmlBlockCloseTh]() {
				goto l591
			}
			return true
		l591:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 120 HtmlBlockOpenThead <- ('<' Spnl ('thead' / 'THEAD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l597
			}
			if !p.rules[ruleSpnl]() {
				goto l597
			}
			{
				position598, thunkPosition598 := position, thunkPosition
				if !matchString("thead") {
					goto l599
				}
				goto l598
			l599:
				position, thunkPosition = position598, thunkPosition598
				if !matchString("THEAD") {
					goto l597
				}
			}
		l598:
			if !p.rules[ruleSpnl]() {
				goto l597
			}
		l600:
			{
				position601, thunkPosition601 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l601
				}
				goto l600
			l601:
				position, thunkPosition = position601, thunkPosition601
			}
			if !matchChar('>') {
				goto l597
			}
			return true
		l597:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 121 HtmlBlockCloseThead <- ('<' Spnl '/' ('thead' / 'THEAD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l602
			}
			if !p.rules[ruleSpnl]() {
				goto l602
			}
			if !matchChar('/') {
				goto l602
			}
			{
				position603, thunkPosition603 := position, thunkPosition
				if !matchString("thead") {
					goto l604
				}
				goto l603
			l604:
				position, thunkPosition = position603, thunkPosition603
				if !matchString("THEAD") {
					goto l602
				}
			}
		l603:
			if !p.rules[ruleSpnl]() {
				goto l602
			}
			if !matchChar('>') {
				goto l602
			}
			return true
		l602:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 122 HtmlBlockThead <- (HtmlBlockOpenThead (HtmlBlockThead / (!HtmlBlockCloseThead .))* HtmlBlockCloseThead) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenThead]() {
				goto l605
			}
		l606:
			{
				position607, thunkPosition607 := position, thunkPosition
				{
					position608, thunkPosition608 := position, thunkPosition
					if !p.rules[ruleHtmlBlockThead]() {
						goto l609
					}
					goto l608
				l609:
					position, thunkPosition = position608, thunkPosition608
					{
						position610, thunkPosition610 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseThead]() {
							goto l610
						}
						goto l607
					l610:
						position, thunkPosition = position610, thunkPosition610
					}
					if !matchDot() {
						goto l607
					}
				}
			l608:
				goto l606
			l607:
				position, thunkPosition = position607, thunkPosition607
			}
			if !p.rules[ruleHtmlBlockCloseThead]() {
				goto l605
			}
			return true
		l605:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 123 HtmlBlockOpenTr <- ('<' Spnl ('tr' / 'TR') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l611
			}
			if !p.rules[ruleSpnl]() {
				goto l611
			}
			{
				position612, thunkPosition612 := position, thunkPosition
				if !matchString("tr") {
					goto l613
				}
				goto l612
			l613:
				position, thunkPosition = position612, thunkPosition612
				if !matchString("TR") {
					goto l611
				}
			}
		l612:
			if !p.rules[ruleSpnl]() {
				goto l611
			}
		l614:
			{
				position615, thunkPosition615 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l615
				}
				goto l614
			l615:
				position, thunkPosition = position615, thunkPosition615
			}
			if !matchChar('>') {
				goto l611
			}
			return true
		l611:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 124 HtmlBlockCloseTr <- ('<' Spnl '/' ('tr' / 'TR') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l616
			}
			if !p.rules[ruleSpnl]() {
				goto l616
			}
			if !matchChar('/') {
				goto l616
			}
			{
				position617, thunkPosition617 := position, thunkPosition
				if !matchString("tr") {
					goto l618
				}
				goto l617
			l618:
				position, thunkPosition = position617, thunkPosition617
				if !matchString("TR") {
					goto l616
				}
			}
		l617:
			if !p.rules[ruleSpnl]() {
				goto l616
			}
			if !matchChar('>') {
				goto l616
			}
			return true
		l616:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 125 HtmlBlockTr <- (HtmlBlockOpenTr (HtmlBlockTr / (!HtmlBlockCloseTr .))* HtmlBlockCloseTr) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTr]() {
				goto l619
			}
		l620:
			{
				position621, thunkPosition621 := position, thunkPosition
				{
					position622, thunkPosition622 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTr]() {
						goto l623
					}
					goto l622
				l623:
					position, thunkPosition = position622, thunkPosition622
					{
						position624, thunkPosition624 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTr]() {
							goto l624
						}
						goto l621
					l624:
						position, thunkPosition = position624, thunkPosition624
					}
					if !matchDot() {
						goto l621
					}
				}
			l622:
				goto l620
			l621:
				position, thunkPosition = position621, thunkPosition621
			}
			if !p.rules[ruleHtmlBlockCloseTr]() {
				goto l619
			}
			return true
		l619:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 126 HtmlBlockOpenScript <- ('<' Spnl ('script' / 'SCRIPT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l625
			}
			if !p.rules[ruleSpnl]() {
				goto l625
			}
			{
				position626, thunkPosition626 := position, thunkPosition
				if !matchString("script") {
					goto l627
				}
				goto l626
			l627:
				position, thunkPosition = position626, thunkPosition626
				if !matchString("SCRIPT") {
					goto l625
				}
			}
		l626:
			if !p.rules[ruleSpnl]() {
				goto l625
			}
		l628:
			{
				position629, thunkPosition629 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l629
				}
				goto l628
			l629:
				position, thunkPosition = position629, thunkPosition629
			}
			if !matchChar('>') {
				goto l625
			}
			return true
		l625:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 127 HtmlBlockCloseScript <- ('<' Spnl '/' ('script' / 'SCRIPT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l630
			}
			if !p.rules[ruleSpnl]() {
				goto l630
			}
			if !matchChar('/') {
				goto l630
			}
			{
				position631, thunkPosition631 := position, thunkPosition
				if !matchString("script") {
					goto l632
				}
				goto l631
			l632:
				position, thunkPosition = position631, thunkPosition631
				if !matchString("SCRIPT") {
					goto l630
				}
			}
		l631:
			if !p.rules[ruleSpnl]() {
				goto l630
			}
			if !matchChar('>') {
				goto l630
			}
			return true
		l630:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 128 HtmlBlockScript <- (HtmlBlockOpenScript (HtmlBlockScript / (!HtmlBlockCloseScript .))* HtmlBlockCloseScript) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenScript]() {
				goto l633
			}
		l634:
			{
				position635, thunkPosition635 := position, thunkPosition
				{
					position636, thunkPosition636 := position, thunkPosition
					if !p.rules[ruleHtmlBlockScript]() {
						goto l637
					}
					goto l636
				l637:
					position, thunkPosition = position636, thunkPosition636
					{
						position638, thunkPosition638 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseScript]() {
							goto l638
						}
						goto l635
					l638:
						position, thunkPosition = position638, thunkPosition638
					}
					if !matchDot() {
						goto l635
					}
				}
			l636:
				goto l634
			l635:
				position, thunkPosition = position635, thunkPosition635
			}
			if !p.rules[ruleHtmlBlockCloseScript]() {
				goto l633
			}
			return true
		l633:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 129 HtmlBlockInTags <- (HtmlBlockAddress / HtmlBlockBlockquote / HtmlBlockCenter / HtmlBlockDir / HtmlBlockDiv / HtmlBlockDl / HtmlBlockFieldset / HtmlBlockForm / HtmlBlockH1 / HtmlBlockH2 / HtmlBlockH3 / HtmlBlockH4 / HtmlBlockH5 / HtmlBlockH6 / HtmlBlockMenu / HtmlBlockNoframes / HtmlBlockNoscript / HtmlBlockOl / HtmlBlockP / HtmlBlockPre / HtmlBlockTable / HtmlBlockUl / HtmlBlockDd / HtmlBlockDt / HtmlBlockFrameset / HtmlBlockLi / HtmlBlockTbody / HtmlBlockTd / HtmlBlockTfoot / HtmlBlockTh / HtmlBlockThead / HtmlBlockTr / HtmlBlockScript) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position640, thunkPosition640 := position, thunkPosition
				if !p.rules[ruleHtmlBlockAddress]() {
					goto l641
				}
				goto l640
			l641:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockBlockquote]() {
					goto l642
				}
				goto l640
			l642:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockCenter]() {
					goto l643
				}
				goto l640
			l643:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockDir]() {
					goto l644
				}
				goto l640
			l644:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockDiv]() {
					goto l645
				}
				goto l640
			l645:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockDl]() {
					goto l646
				}
				goto l640
			l646:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockFieldset]() {
					goto l647
				}
				goto l640
			l647:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockForm]() {
					goto l648
				}
				goto l640
			l648:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockH1]() {
					goto l649
				}
				goto l640
			l649:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockH2]() {
					goto l650
				}
				goto l640
			l650:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockH3]() {
					goto l651
				}
				goto l640
			l651:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockH4]() {
					goto l652
				}
				goto l640
			l652:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockH5]() {
					goto l653
				}
				goto l640
			l653:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockH6]() {
					goto l654
				}
				goto l640
			l654:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockMenu]() {
					goto l655
				}
				goto l640
			l655:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockNoframes]() {
					goto l656
				}
				goto l640
			l656:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockNoscript]() {
					goto l657
				}
				goto l640
			l657:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockOl]() {
					goto l658
				}
				goto l640
			l658:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockP]() {
					goto l659
				}
				goto l640
			l659:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockPre]() {
					goto l660
				}
				goto l640
			l660:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockTable]() {
					goto l661
				}
				goto l640
			l661:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockUl]() {
					goto l662
				}
				goto l640
			l662:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockDd]() {
					goto l663
				}
				goto l640
			l663:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockDt]() {
					goto l664
				}
				goto l640
			l664:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockFrameset]() {
					goto l665
				}
				goto l640
			l665:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockLi]() {
					goto l666
				}
				goto l640
			l666:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockTbody]() {
					goto l667
				}
				goto l640
			l667:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockTd]() {
					goto l668
				}
				goto l640
			l668:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockTfoot]() {
					goto l669
				}
				goto l640
			l669:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockTh]() {
					goto l670
				}
				goto l640
			l670:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockThead]() {
					goto l671
				}
				goto l640
			l671:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockTr]() {
					goto l672
				}
				goto l640
			l672:
				position, thunkPosition = position640, thunkPosition640
				if !p.rules[ruleHtmlBlockScript]() {
					goto l639
				}
			}
		l640:
			return true
		l639:
			position, thunkPosition = position0, thunkPosition0
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
			position0, thunkPosition0 := position, thunkPosition
			if !peekChar('<') {
				goto l673
			}
			begin = position
			{
				position674, thunkPosition674 := position, thunkPosition
				if !p.rules[ruleHtmlBlockInTags]() {
					goto l675
				}
				goto l674
			l675:
				position, thunkPosition = position674, thunkPosition674
				if !p.rules[ruleHtmlComment]() {
					goto l676
				}
				goto l674
			l676:
				position, thunkPosition = position674, thunkPosition674
				if !p.rules[ruleHtmlBlockSelfClosing]() {
					goto l673
				}
			}
		l674:
			end = position
			if !p.rules[ruleBlankLine]() {
				goto l673
			}
		l677:
			{
				position678, thunkPosition678 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l678
				}
				goto l677
			l678:
				position, thunkPosition = position678, thunkPosition678
			}
			do(40)
			return true
		l673:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 131 HtmlBlockSelfClosing <- ('<' Spnl HtmlBlockType Spnl HtmlAttribute* '/' Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l679
			}
			if !p.rules[ruleSpnl]() {
				goto l679
			}
			if !p.rules[ruleHtmlBlockType]() {
				goto l679
			}
			if !p.rules[ruleSpnl]() {
				goto l679
			}
		l680:
			{
				position681, thunkPosition681 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l681
				}
				goto l680
			l681:
				position, thunkPosition = position681, thunkPosition681
			}
			if !matchChar('/') {
				goto l679
			}
			if !p.rules[ruleSpnl]() {
				goto l679
			}
			if !matchChar('>') {
				goto l679
			}
			return true
		l679:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 132 HtmlBlockType <- ('address' / 'blockquote' / 'center' / 'dir' / 'div' / 'dl' / 'fieldset' / 'form' / 'h1' / 'h2' / 'h3' / 'h4' / 'h5' / 'h6' / 'hr' / 'isindex' / 'menu' / 'noframes' / 'noscript' / 'ol' / 'p' / 'pre' / 'table' / 'ul' / 'dd' / 'dt' / 'frameset' / 'li' / 'tbody' / 'td' / 'tfoot' / 'th' / 'thead' / 'tr' / 'script' / 'ADDRESS' / 'BLOCKQUOTE' / 'CENTER' / 'DIR' / 'DIV' / 'DL' / 'FIELDSET' / 'FORM' / 'H1' / 'H2' / 'H3' / 'H4' / 'H5' / 'H6' / 'HR' / 'ISINDEX' / 'MENU' / 'NOFRAMES' / 'NOSCRIPT' / 'OL' / 'P' / 'PRE' / 'TABLE' / 'UL' / 'DD' / 'DT' / 'FRAMESET' / 'LI' / 'TBODY' / 'TD' / 'TFOOT' / 'TH' / 'THEAD' / 'TR' / 'SCRIPT') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position683, thunkPosition683 := position, thunkPosition
				if !matchString("address") {
					goto l684
				}
				goto l683
			l684:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("blockquote") {
					goto l685
				}
				goto l683
			l685:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("center") {
					goto l686
				}
				goto l683
			l686:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("dir") {
					goto l687
				}
				goto l683
			l687:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("div") {
					goto l688
				}
				goto l683
			l688:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("dl") {
					goto l689
				}
				goto l683
			l689:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("fieldset") {
					goto l690
				}
				goto l683
			l690:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("form") {
					goto l691
				}
				goto l683
			l691:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("h1") {
					goto l692
				}
				goto l683
			l692:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("h2") {
					goto l693
				}
				goto l683
			l693:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("h3") {
					goto l694
				}
				goto l683
			l694:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("h4") {
					goto l695
				}
				goto l683
			l695:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("h5") {
					goto l696
				}
				goto l683
			l696:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("h6") {
					goto l697
				}
				goto l683
			l697:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("hr") {
					goto l698
				}
				goto l683
			l698:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("isindex") {
					goto l699
				}
				goto l683
			l699:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("menu") {
					goto l700
				}
				goto l683
			l700:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("noframes") {
					goto l701
				}
				goto l683
			l701:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("noscript") {
					goto l702
				}
				goto l683
			l702:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("ol") {
					goto l703
				}
				goto l683
			l703:
				position, thunkPosition = position683, thunkPosition683
				if !matchChar('p') {
					goto l704
				}
				goto l683
			l704:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("pre") {
					goto l705
				}
				goto l683
			l705:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("table") {
					goto l706
				}
				goto l683
			l706:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("ul") {
					goto l707
				}
				goto l683
			l707:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("dd") {
					goto l708
				}
				goto l683
			l708:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("dt") {
					goto l709
				}
				goto l683
			l709:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("frameset") {
					goto l710
				}
				goto l683
			l710:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("li") {
					goto l711
				}
				goto l683
			l711:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("tbody") {
					goto l712
				}
				goto l683
			l712:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("td") {
					goto l713
				}
				goto l683
			l713:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("tfoot") {
					goto l714
				}
				goto l683
			l714:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("th") {
					goto l715
				}
				goto l683
			l715:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("thead") {
					goto l716
				}
				goto l683
			l716:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("tr") {
					goto l717
				}
				goto l683
			l717:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("script") {
					goto l718
				}
				goto l683
			l718:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("ADDRESS") {
					goto l719
				}
				goto l683
			l719:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("BLOCKQUOTE") {
					goto l720
				}
				goto l683
			l720:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("CENTER") {
					goto l721
				}
				goto l683
			l721:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("DIR") {
					goto l722
				}
				goto l683
			l722:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("DIV") {
					goto l723
				}
				goto l683
			l723:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("DL") {
					goto l724
				}
				goto l683
			l724:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("FIELDSET") {
					goto l725
				}
				goto l683
			l725:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("FORM") {
					goto l726
				}
				goto l683
			l726:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("H1") {
					goto l727
				}
				goto l683
			l727:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("H2") {
					goto l728
				}
				goto l683
			l728:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("H3") {
					goto l729
				}
				goto l683
			l729:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("H4") {
					goto l730
				}
				goto l683
			l730:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("H5") {
					goto l731
				}
				goto l683
			l731:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("H6") {
					goto l732
				}
				goto l683
			l732:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("HR") {
					goto l733
				}
				goto l683
			l733:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("ISINDEX") {
					goto l734
				}
				goto l683
			l734:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("MENU") {
					goto l735
				}
				goto l683
			l735:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("NOFRAMES") {
					goto l736
				}
				goto l683
			l736:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("NOSCRIPT") {
					goto l737
				}
				goto l683
			l737:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("OL") {
					goto l738
				}
				goto l683
			l738:
				position, thunkPosition = position683, thunkPosition683
				if !matchChar('P') {
					goto l739
				}
				goto l683
			l739:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("PRE") {
					goto l740
				}
				goto l683
			l740:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("TABLE") {
					goto l741
				}
				goto l683
			l741:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("UL") {
					goto l742
				}
				goto l683
			l742:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("DD") {
					goto l743
				}
				goto l683
			l743:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("DT") {
					goto l744
				}
				goto l683
			l744:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("FRAMESET") {
					goto l745
				}
				goto l683
			l745:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("LI") {
					goto l746
				}
				goto l683
			l746:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("TBODY") {
					goto l747
				}
				goto l683
			l747:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("TD") {
					goto l748
				}
				goto l683
			l748:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("TFOOT") {
					goto l749
				}
				goto l683
			l749:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("TH") {
					goto l750
				}
				goto l683
			l750:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("THEAD") {
					goto l751
				}
				goto l683
			l751:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("TR") {
					goto l752
				}
				goto l683
			l752:
				position, thunkPosition = position683, thunkPosition683
				if !matchString("SCRIPT") {
					goto l682
				}
			}
		l683:
			return true
		l682:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 133 StyleOpen <- ('<' Spnl ('style' / 'STYLE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l753
			}
			if !p.rules[ruleSpnl]() {
				goto l753
			}
			{
				position754, thunkPosition754 := position, thunkPosition
				if !matchString("style") {
					goto l755
				}
				goto l754
			l755:
				position, thunkPosition = position754, thunkPosition754
				if !matchString("STYLE") {
					goto l753
				}
			}
		l754:
			if !p.rules[ruleSpnl]() {
				goto l753
			}
		l756:
			{
				position757, thunkPosition757 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l757
				}
				goto l756
			l757:
				position, thunkPosition = position757, thunkPosition757
			}
			if !matchChar('>') {
				goto l753
			}
			return true
		l753:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 134 StyleClose <- ('<' Spnl '/' ('style' / 'STYLE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l758
			}
			if !p.rules[ruleSpnl]() {
				goto l758
			}
			if !matchChar('/') {
				goto l758
			}
			{
				position759, thunkPosition759 := position, thunkPosition
				if !matchString("style") {
					goto l760
				}
				goto l759
			l760:
				position, thunkPosition = position759, thunkPosition759
				if !matchString("STYLE") {
					goto l758
				}
			}
		l759:
			if !p.rules[ruleSpnl]() {
				goto l758
			}
			if !matchChar('>') {
				goto l758
			}
			return true
		l758:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 135 InStyleTags <- (StyleOpen (!StyleClose .)* StyleClose) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleStyleOpen]() {
				goto l761
			}
		l762:
			{
				position763, thunkPosition763 := position, thunkPosition
				{
					position764, thunkPosition764 := position, thunkPosition
					if !p.rules[ruleStyleClose]() {
						goto l764
					}
					goto l763
				l764:
					position, thunkPosition = position764, thunkPosition764
				}
				if !matchDot() {
					goto l763
				}
				goto l762
			l763:
				position, thunkPosition = position763, thunkPosition763
			}
			if !p.rules[ruleStyleClose]() {
				goto l761
			}
			return true
		l761:
			position, thunkPosition = position0, thunkPosition0
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
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleInStyleTags]() {
				goto l765
			}
			end = position
		l766:
			{
				position767, thunkPosition767 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l767
				}
				goto l766
			l767:
				position, thunkPosition = position767, thunkPosition767
			}
			do(41)
			return true
		l765:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 137 Inlines <- (StartList ((!Endline Inline { a = cons(yy, a) }) / (Endline &Inline { a = cons(c, a) }))+ Endline? { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l768
			}
			doarg(yySet, -2)
			{
				position771, thunkPosition771 := position, thunkPosition
				{
					position773, thunkPosition773 := position, thunkPosition
					if !p.rules[ruleEndline]() {
						goto l773
					}
					goto l772
				l773:
					position, thunkPosition = position773, thunkPosition773
				}
				if !p.rules[ruleInline]() {
					goto l772
				}
				do(42)
				goto l771
			l772:
				position, thunkPosition = position771, thunkPosition771
				if !p.rules[ruleEndline]() {
					goto l768
				}
				doarg(yySet, -1)
				{
					position774, thunkPosition774 := position, thunkPosition
					if !p.rules[ruleInline]() {
						goto l768
					}
					position, thunkPosition = position774, thunkPosition774
				}
				do(43)
			}
		l771:
		l769:
			{
				position770, thunkPosition770 := position, thunkPosition
				{
					position775, thunkPosition775 := position, thunkPosition
					{
						position777, thunkPosition777 := position, thunkPosition
						if !p.rules[ruleEndline]() {
							goto l777
						}
						goto l776
					l777:
						position, thunkPosition = position777, thunkPosition777
					}
					if !p.rules[ruleInline]() {
						goto l776
					}
					do(42)
					goto l775
				l776:
					position, thunkPosition = position775, thunkPosition775
					if !p.rules[ruleEndline]() {
						goto l770
					}
					doarg(yySet, -1)
					{
						position778, thunkPosition778 := position, thunkPosition
						if !p.rules[ruleInline]() {
							goto l770
						}
						position, thunkPosition = position778, thunkPosition778
					}
					do(43)
				}
			l775:
				goto l769
			l770:
				position, thunkPosition = position770, thunkPosition770
			}
			{
				position779, thunkPosition779 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l779
				}
				goto l780
			l779:
				position, thunkPosition = position779, thunkPosition779
			}
		l780:
			do(44)
			doarg(yyPop, 2)
			return true
		l768:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 138 Inline <- (Str / Endline / UlOrStarLine / Space / Strong / Emph / Image / Link / NoteReference / InlineNote / Code / RawHtml / Entity / EscapedChar / Smart / Symbol) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position782, thunkPosition782 := position, thunkPosition
				if !p.rules[ruleStr]() {
					goto l783
				}
				goto l782
			l783:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleEndline]() {
					goto l784
				}
				goto l782
			l784:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleUlOrStarLine]() {
					goto l785
				}
				goto l782
			l785:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleSpace]() {
					goto l786
				}
				goto l782
			l786:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleStrong]() {
					goto l787
				}
				goto l782
			l787:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleEmph]() {
					goto l788
				}
				goto l782
			l788:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleImage]() {
					goto l789
				}
				goto l782
			l789:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleLink]() {
					goto l790
				}
				goto l782
			l790:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleNoteReference]() {
					goto l791
				}
				goto l782
			l791:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleInlineNote]() {
					goto l792
				}
				goto l782
			l792:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleCode]() {
					goto l793
				}
				goto l782
			l793:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleRawHtml]() {
					goto l794
				}
				goto l782
			l794:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleEntity]() {
					goto l795
				}
				goto l782
			l795:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleEscapedChar]() {
					goto l796
				}
				goto l782
			l796:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleSmart]() {
					goto l797
				}
				goto l782
			l797:
				position, thunkPosition = position782, thunkPosition782
				if !p.rules[ruleSymbol]() {
					goto l781
				}
			}
		l782:
			return true
		l781:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 139 Space <- (Spacechar+ { yy = mk_str(" ")
          yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSpacechar]() {
				goto l798
			}
		l799:
			{
				position800, thunkPosition800 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l800
				}
				goto l799
			l800:
				position, thunkPosition = position800, thunkPosition800
			}
			do(45)
			return true
		l798:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 140 Str <- (< NormalChar (NormalChar / ('_'+ &Alphanumeric))* > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleNormalChar]() {
				goto l801
			}
		l802:
			{
				position803, thunkPosition803 := position, thunkPosition
				{
					position804, thunkPosition804 := position, thunkPosition
					if !p.rules[ruleNormalChar]() {
						goto l805
					}
					goto l804
				l805:
					position, thunkPosition = position804, thunkPosition804
					if !matchChar('_') {
						goto l803
					}
				l806:
					{
						position807, thunkPosition807 := position, thunkPosition
						if !matchChar('_') {
							goto l807
						}
						goto l806
					l807:
						position, thunkPosition = position807, thunkPosition807
					}
					{
						position808, thunkPosition808 := position, thunkPosition
						if !p.rules[ruleAlphanumeric]() {
							goto l803
						}
						position, thunkPosition = position808, thunkPosition808
					}
				}
			l804:
				goto l802
			l803:
				position, thunkPosition = position803, thunkPosition803
			}
			end = position
			do(46)
			return true
		l801:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 141 EscapedChar <- ('\\' !Newline < [-\\`|*_{}[\]()#+.!><] > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\\') {
				goto l809
			}
			{
				position810, thunkPosition810 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l810
				}
				goto l809
			l810:
				position, thunkPosition = position810, thunkPosition810
			}
			begin = position
			if !matchClass(2) {
				goto l809
			}
			end = position
			do(47)
			return true
		l809:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 142 Entity <- ((HexEntity / DecEntity / CharEntity) { yy = mk_str(yytext); yy.key = HTML }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position812, thunkPosition812 := position, thunkPosition
				if !p.rules[ruleHexEntity]() {
					goto l813
				}
				goto l812
			l813:
				position, thunkPosition = position812, thunkPosition812
				if !p.rules[ruleDecEntity]() {
					goto l814
				}
				goto l812
			l814:
				position, thunkPosition = position812, thunkPosition812
				if !p.rules[ruleCharEntity]() {
					goto l811
				}
			}
		l812:
			do(48)
			return true
		l811:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 143 Endline <- (LineBreak / TerminalEndline / NormalEndline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position816, thunkPosition816 := position, thunkPosition
				if !p.rules[ruleLineBreak]() {
					goto l817
				}
				goto l816
			l817:
				position, thunkPosition = position816, thunkPosition816
				if !p.rules[ruleTerminalEndline]() {
					goto l818
				}
				goto l816
			l818:
				position, thunkPosition = position816, thunkPosition816
				if !p.rules[ruleNormalEndline]() {
					goto l815
				}
			}
		l816:
			return true
		l815:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 144 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line (('===' '='*) / ('---' '-'*)) Newline) { yy = mk_str("\n")
                    yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l819
			}
			if !p.rules[ruleNewline]() {
				goto l819
			}
			{
				position820, thunkPosition820 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l820
				}
				goto l819
			l820:
				position, thunkPosition = position820, thunkPosition820
			}
			if peekChar('>') {
				goto l819
			}
			{
				position821, thunkPosition821 := position, thunkPosition
				if !p.rules[ruleAtxStart]() {
					goto l821
				}
				goto l819
			l821:
				position, thunkPosition = position821, thunkPosition821
			}
			{
				position822, thunkPosition822 := position, thunkPosition
				if !p.rules[ruleLine]() {
					goto l822
				}
				{
					position823, thunkPosition823 := position, thunkPosition
					if !matchString("===") {
						goto l824
					}
				l825:
					{
						position826, thunkPosition826 := position, thunkPosition
						if !matchChar('=') {
							goto l826
						}
						goto l825
					l826:
						position, thunkPosition = position826, thunkPosition826
					}
					goto l823
				l824:
					position, thunkPosition = position823, thunkPosition823
					if !matchString("---") {
						goto l822
					}
				l827:
					{
						position828, thunkPosition828 := position, thunkPosition
						if !matchChar('-') {
							goto l828
						}
						goto l827
					l828:
						position, thunkPosition = position828, thunkPosition828
					}
				}
			l823:
				if !p.rules[ruleNewline]() {
					goto l822
				}
				goto l819
			l822:
				position, thunkPosition = position822, thunkPosition822
			}
			do(49)
			return true
		l819:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 145 TerminalEndline <- (Sp Newline Eof { yy = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l829
			}
			if !p.rules[ruleNewline]() {
				goto l829
			}
			if !p.rules[ruleEof]() {
				goto l829
			}
			do(50)
			return true
		l829:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 146 LineBreak <- ('  ' NormalEndline { yy = mk_element(LINEBREAK) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("  ") {
				goto l830
			}
			if !p.rules[ruleNormalEndline]() {
				goto l830
			}
			do(51)
			return true
		l830:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 147 Symbol <- (< SpecialChar > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleSpecialChar]() {
				goto l831
			}
			end = position
			do(52)
			return true
		l831:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 148 UlOrStarLine <- ((UlLine / StarLine) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position833, thunkPosition833 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l834
				}
				goto l833
			l834:
				position, thunkPosition = position833, thunkPosition833
				if !p.rules[ruleStarLine]() {
					goto l832
				}
			}
		l833:
			do(53)
			return true
		l832:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 149 StarLine <- ((< '****' '*'* >) / (< Spacechar '*'+ &Spacechar >)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position836, thunkPosition836 := position, thunkPosition
				begin = position
				if !matchString("****") {
					goto l837
				}
			l838:
				{
					position839, thunkPosition839 := position, thunkPosition
					if !matchChar('*') {
						goto l839
					}
					goto l838
				l839:
					position, thunkPosition = position839, thunkPosition839
				}
				end = position
				goto l836
			l837:
				position, thunkPosition = position836, thunkPosition836
				begin = position
				if !p.rules[ruleSpacechar]() {
					goto l835
				}
				if !matchChar('*') {
					goto l835
				}
			l840:
				{
					position841, thunkPosition841 := position, thunkPosition
					if !matchChar('*') {
						goto l841
					}
					goto l840
				l841:
					position, thunkPosition = position841, thunkPosition841
				}
				{
					position842, thunkPosition842 := position, thunkPosition
					if !p.rules[ruleSpacechar]() {
						goto l835
					}
					position, thunkPosition = position842, thunkPosition842
				}
				end = position
			}
		l836:
			return true
		l835:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 150 UlLine <- ((< '____' '_'* >) / (< Spacechar '_'+ &Spacechar >)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position844, thunkPosition844 := position, thunkPosition
				begin = position
				if !matchString("____") {
					goto l845
				}
			l846:
				{
					position847, thunkPosition847 := position, thunkPosition
					if !matchChar('_') {
						goto l847
					}
					goto l846
				l847:
					position, thunkPosition = position847, thunkPosition847
				}
				end = position
				goto l844
			l845:
				position, thunkPosition = position844, thunkPosition844
				begin = position
				if !p.rules[ruleSpacechar]() {
					goto l843
				}
				if !matchChar('_') {
					goto l843
				}
			l848:
				{
					position849, thunkPosition849 := position, thunkPosition
					if !matchChar('_') {
						goto l849
					}
					goto l848
				l849:
					position, thunkPosition = position849, thunkPosition849
				}
				{
					position850, thunkPosition850 := position, thunkPosition
					if !p.rules[ruleSpacechar]() {
						goto l843
					}
					position, thunkPosition = position850, thunkPosition850
				}
				end = position
			}
		l844:
			return true
		l843:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 151 Emph <- (EmphStar / EmphUl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position852, thunkPosition852 := position, thunkPosition
				if !p.rules[ruleEmphStar]() {
					goto l853
				}
				goto l852
			l853:
				position, thunkPosition = position852, thunkPosition852
				if !p.rules[ruleEmphUl]() {
					goto l851
				}
			}
		l852:
			return true
		l851:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 152 OneStarOpen <- (!StarLine '*' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position855, thunkPosition855 := position, thunkPosition
				if !p.rules[ruleStarLine]() {
					goto l855
				}
				goto l854
			l855:
				position, thunkPosition = position855, thunkPosition855
			}
			if !matchChar('*') {
				goto l854
			}
			{
				position856, thunkPosition856 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l856
				}
				goto l854
			l856:
				position, thunkPosition = position856, thunkPosition856
			}
			{
				position857, thunkPosition857 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l857
				}
				goto l854
			l857:
				position, thunkPosition = position857, thunkPosition857
			}
			return true
		l854:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 153 OneStarClose <- (!Spacechar !Newline Inline !StrongStar '*' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position859, thunkPosition859 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l859
				}
				goto l858
			l859:
				position, thunkPosition = position859, thunkPosition859
			}
			{
				position860, thunkPosition860 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l860
				}
				goto l858
			l860:
				position, thunkPosition = position860, thunkPosition860
			}
			if !p.rules[ruleInline]() {
				goto l858
			}
			doarg(yySet, -1)
			{
				position861, thunkPosition861 := position, thunkPosition
				if !p.rules[ruleStrongStar]() {
					goto l861
				}
				goto l858
			l861:
				position, thunkPosition = position861, thunkPosition861
			}
			if !matchChar('*') {
				goto l858
			}
			do(54)
			doarg(yyPop, 1)
			return true
		l858:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 154 EmphStar <- (OneStarOpen StartList (!OneStarClose Inline { a = cons(yy, a) })* OneStarClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneStarOpen]() {
				goto l862
			}
			if !p.rules[ruleStartList]() {
				goto l862
			}
			doarg(yySet, -1)
		l863:
			{
				position864, thunkPosition864 := position, thunkPosition
				{
					position865, thunkPosition865 := position, thunkPosition
					if !p.rules[ruleOneStarClose]() {
						goto l865
					}
					goto l864
				l865:
					position, thunkPosition = position865, thunkPosition865
				}
				if !p.rules[ruleInline]() {
					goto l864
				}
				do(55)
				goto l863
			l864:
				position, thunkPosition = position864, thunkPosition864
			}
			if !p.rules[ruleOneStarClose]() {
				goto l862
			}
			do(56)
			do(57)
			doarg(yyPop, 1)
			return true
		l862:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 155 OneUlOpen <- (!UlLine '_' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position867, thunkPosition867 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l867
				}
				goto l866
			l867:
				position, thunkPosition = position867, thunkPosition867
			}
			if !matchChar('_') {
				goto l866
			}
			{
				position868, thunkPosition868 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l868
				}
				goto l866
			l868:
				position, thunkPosition = position868, thunkPosition868
			}
			{
				position869, thunkPosition869 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l869
				}
				goto l866
			l869:
				position, thunkPosition = position869, thunkPosition869
			}
			return true
		l866:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 156 OneUlClose <- (!Spacechar !Newline Inline !StrongUl '_' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position871, thunkPosition871 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l871
				}
				goto l870
			l871:
				position, thunkPosition = position871, thunkPosition871
			}
			{
				position872, thunkPosition872 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l872
				}
				goto l870
			l872:
				position, thunkPosition = position872, thunkPosition872
			}
			if !p.rules[ruleInline]() {
				goto l870
			}
			doarg(yySet, -1)
			{
				position873, thunkPosition873 := position, thunkPosition
				if !p.rules[ruleStrongUl]() {
					goto l873
				}
				goto l870
			l873:
				position, thunkPosition = position873, thunkPosition873
			}
			if !matchChar('_') {
				goto l870
			}
			{
				position874, thunkPosition874 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l874
				}
				goto l870
			l874:
				position, thunkPosition = position874, thunkPosition874
			}
			do(58)
			doarg(yyPop, 1)
			return true
		l870:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 157 EmphUl <- (OneUlOpen StartList (!OneUlClose Inline { a = cons(yy, a) })* OneUlClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneUlOpen]() {
				goto l875
			}
			if !p.rules[ruleStartList]() {
				goto l875
			}
			doarg(yySet, -1)
		l876:
			{
				position877, thunkPosition877 := position, thunkPosition
				{
					position878, thunkPosition878 := position, thunkPosition
					if !p.rules[ruleOneUlClose]() {
						goto l878
					}
					goto l877
				l878:
					position, thunkPosition = position878, thunkPosition878
				}
				if !p.rules[ruleInline]() {
					goto l877
				}
				do(59)
				goto l876
			l877:
				position, thunkPosition = position877, thunkPosition877
			}
			if !p.rules[ruleOneUlClose]() {
				goto l875
			}
			do(60)
			do(61)
			doarg(yyPop, 1)
			return true
		l875:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 158 Strong <- (StrongStar / StrongUl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position880, thunkPosition880 := position, thunkPosition
				if !p.rules[ruleStrongStar]() {
					goto l881
				}
				goto l880
			l881:
				position, thunkPosition = position880, thunkPosition880
				if !p.rules[ruleStrongUl]() {
					goto l879
				}
			}
		l880:
			return true
		l879:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 159 TwoStarOpen <- (!StarLine '**' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position883, thunkPosition883 := position, thunkPosition
				if !p.rules[ruleStarLine]() {
					goto l883
				}
				goto l882
			l883:
				position, thunkPosition = position883, thunkPosition883
			}
			if !matchString("**") {
				goto l882
			}
			{
				position884, thunkPosition884 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l884
				}
				goto l882
			l884:
				position, thunkPosition = position884, thunkPosition884
			}
			{
				position885, thunkPosition885 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l885
				}
				goto l882
			l885:
				position, thunkPosition = position885, thunkPosition885
			}
			return true
		l882:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 160 TwoStarClose <- (!Spacechar !Newline Inline '**' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position887, thunkPosition887 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l887
				}
				goto l886
			l887:
				position, thunkPosition = position887, thunkPosition887
			}
			{
				position888, thunkPosition888 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l888
				}
				goto l886
			l888:
				position, thunkPosition = position888, thunkPosition888
			}
			if !p.rules[ruleInline]() {
				goto l886
			}
			doarg(yySet, -1)
			if !matchString("**") {
				goto l886
			}
			do(62)
			doarg(yyPop, 1)
			return true
		l886:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 161 StrongStar <- (TwoStarOpen StartList (!TwoStarClose Inline { a = cons(yy, a) })* TwoStarClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoStarOpen]() {
				goto l889
			}
			if !p.rules[ruleStartList]() {
				goto l889
			}
			doarg(yySet, -1)
		l890:
			{
				position891, thunkPosition891 := position, thunkPosition
				{
					position892, thunkPosition892 := position, thunkPosition
					if !p.rules[ruleTwoStarClose]() {
						goto l892
					}
					goto l891
				l892:
					position, thunkPosition = position892, thunkPosition892
				}
				if !p.rules[ruleInline]() {
					goto l891
				}
				do(63)
				goto l890
			l891:
				position, thunkPosition = position891, thunkPosition891
			}
			if !p.rules[ruleTwoStarClose]() {
				goto l889
			}
			do(64)
			do(65)
			doarg(yyPop, 1)
			return true
		l889:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 162 TwoUlOpen <- (!UlLine '__' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position894, thunkPosition894 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l894
				}
				goto l893
			l894:
				position, thunkPosition = position894, thunkPosition894
			}
			if !matchString("__") {
				goto l893
			}
			{
				position895, thunkPosition895 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l895
				}
				goto l893
			l895:
				position, thunkPosition = position895, thunkPosition895
			}
			{
				position896, thunkPosition896 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l896
				}
				goto l893
			l896:
				position, thunkPosition = position896, thunkPosition896
			}
			return true
		l893:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 163 TwoUlClose <- (!Spacechar !Newline Inline '__' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position898, thunkPosition898 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l898
				}
				goto l897
			l898:
				position, thunkPosition = position898, thunkPosition898
			}
			{
				position899, thunkPosition899 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l899
				}
				goto l897
			l899:
				position, thunkPosition = position899, thunkPosition899
			}
			if !p.rules[ruleInline]() {
				goto l897
			}
			doarg(yySet, -1)
			if !matchString("__") {
				goto l897
			}
			{
				position900, thunkPosition900 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l900
				}
				goto l897
			l900:
				position, thunkPosition = position900, thunkPosition900
			}
			do(66)
			doarg(yyPop, 1)
			return true
		l897:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 164 StrongUl <- (TwoUlOpen StartList (!TwoUlClose Inline { a = cons(yy, a) })* TwoUlClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoUlOpen]() {
				goto l901
			}
			if !p.rules[ruleStartList]() {
				goto l901
			}
			doarg(yySet, -1)
		l902:
			{
				position903, thunkPosition903 := position, thunkPosition
				{
					position904, thunkPosition904 := position, thunkPosition
					if !p.rules[ruleTwoUlClose]() {
						goto l904
					}
					goto l903
				l904:
					position, thunkPosition = position904, thunkPosition904
				}
				if !p.rules[ruleInline]() {
					goto l903
				}
				do(67)
				goto l902
			l903:
				position, thunkPosition = position903, thunkPosition903
			}
			if !p.rules[ruleTwoUlClose]() {
				goto l901
			}
			do(68)
			do(69)
			doarg(yyPop, 1)
			return true
		l901:
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
				goto l905
			}
			{
				position906, thunkPosition906 := position, thunkPosition
				if !p.rules[ruleExplicitLink]() {
					goto l907
				}
				goto l906
			l907:
				position, thunkPosition = position906, thunkPosition906
				if !p.rules[ruleReferenceLink]() {
					goto l905
				}
			}
		l906:
			do(70)
			return true
		l905:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 166 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position909, thunkPosition909 := position, thunkPosition
				if !p.rules[ruleExplicitLink]() {
					goto l910
				}
				goto l909
			l910:
				position, thunkPosition = position909, thunkPosition909
				if !p.rules[ruleReferenceLink]() {
					goto l911
				}
				goto l909
			l911:
				position, thunkPosition = position909, thunkPosition909
				if !p.rules[ruleAutoLink]() {
					goto l908
				}
			}
		l909:
			return true
		l908:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 167 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position913, thunkPosition913 := position, thunkPosition
				if !p.rules[ruleReferenceLinkDouble]() {
					goto l914
				}
				goto l913
			l914:
				position, thunkPosition = position913, thunkPosition913
				if !p.rules[ruleReferenceLinkSingle]() {
					goto l912
				}
			}
		l913:
			return true
		l912:
			position, thunkPosition = position0, thunkPosition0
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
				goto l915
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto l915
			}
			end = position
			{
				position916, thunkPosition916 := position, thunkPosition
				if !matchString("[]") {
					goto l916
				}
				goto l915
			l916:
				position, thunkPosition = position916, thunkPosition916
			}
			if !p.rules[ruleLabel]() {
				goto l915
			}
			doarg(yySet, -2)
			do(71)
			doarg(yyPop, 2)
			return true
		l915:
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
				goto l917
			}
			doarg(yySet, -1)
			begin = position
			{
				position918, thunkPosition918 := position, thunkPosition
				if !p.rules[ruleSpnl]() {
					goto l918
				}
				if !matchString("[]") {
					goto l918
				}
				goto l919
			l918:
				position, thunkPosition = position918, thunkPosition918
			}
		l919:
			end = position
			do(72)
			doarg(yyPop, 1)
			return true
		l917:
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
				goto l920
			}
			doarg(yySet, -2)
			if !p.rules[ruleSpnl]() {
				goto l920
			}
			if !matchChar('(') {
				goto l920
			}
			if !p.rules[ruleSp]() {
				goto l920
			}
			if !p.rules[ruleSource]() {
				goto l920
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l920
			}
			if !p.rules[ruleTitle]() {
				goto l920
			}
			doarg(yySet, -3)
			if !p.rules[ruleSp]() {
				goto l920
			}
			if !matchChar(')') {
				goto l920
			}
			do(73)
			doarg(yyPop, 3)
			return true
		l920:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 171 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position922, thunkPosition922 := position, thunkPosition
				if !matchChar('<') {
					goto l923
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l923
				}
				end = position
				if !matchChar('>') {
					goto l923
				}
				goto l922
			l923:
				position, thunkPosition = position922, thunkPosition922
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l921
				}
				end = position
			}
		l922:
			do(74)
			return true
		l921:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 172 SourceContents <- (((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* / '') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position925, thunkPosition925 := position, thunkPosition
			l927:
				{
					position928, thunkPosition928 := position, thunkPosition
					{
						position929, thunkPosition929 := position, thunkPosition
						if peekChar('(') {
							goto l930
						}
						if peekChar(')') {
							goto l930
						}
						if peekChar('>') {
							goto l930
						}
						if !p.rules[ruleNonspacechar]() {
							goto l930
						}
					l931:
						{
							position932, thunkPosition932 := position, thunkPosition
							if peekChar('(') {
								goto l932
							}
							if peekChar(')') {
								goto l932
							}
							if peekChar('>') {
								goto l932
							}
							if !p.rules[ruleNonspacechar]() {
								goto l932
							}
							goto l931
						l932:
							position, thunkPosition = position932, thunkPosition932
						}
						goto l929
					l930:
						position, thunkPosition = position929, thunkPosition929
						if !matchChar('(') {
							goto l928
						}
						if !p.rules[ruleSourceContents]() {
							goto l928
						}
						if !matchChar(')') {
							goto l928
						}
					}
				l929:
					goto l927
				l928:
					position, thunkPosition = position928, thunkPosition928
				}
				goto l925
				_, _ = position925, thunkPosition925
			}
		l925:
			return true
			_, _ = position0, thunkPosition0
			return false
		},
		/* 173 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position934, thunkPosition934 := position, thunkPosition
				if !p.rules[ruleTitleSingle]() {
					goto l935
				}
				goto l934
			l935:
				position, thunkPosition = position934, thunkPosition934
				if !p.rules[ruleTitleDouble]() {
					goto l936
				}
				goto l934
			l936:
				position, thunkPosition = position934, thunkPosition934
				begin = position
				if !peekDot() {
					goto l933
				}
				end = position
			}
		l934:
			do(75)
			return true
		l933:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 174 TitleSingle <- ('\'' < (!('\'' Sp (')' / Newline)) .)* > '\'') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l937
			}
			begin = position
		l938:
			{
				position939, thunkPosition939 := position, thunkPosition
				{
					position940, thunkPosition940 := position, thunkPosition
					if !matchChar('\'') {
						goto l940
					}
					if !p.rules[ruleSp]() {
						goto l940
					}
					{
						position941, thunkPosition941 := position, thunkPosition
						if !matchChar(')') {
							goto l942
						}
						goto l941
					l942:
						position, thunkPosition = position941, thunkPosition941
						if !p.rules[ruleNewline]() {
							goto l940
						}
					}
				l941:
					goto l939
				l940:
					position, thunkPosition = position940, thunkPosition940
				}
				if !matchDot() {
					goto l939
				}
				goto l938
			l939:
				position, thunkPosition = position939, thunkPosition939
			}
			end = position
			if !matchChar('\'') {
				goto l937
			}
			return true
		l937:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 175 TitleDouble <- ('"' < (!('"' Sp (')' / Newline)) .)* > '"') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l943
			}
			begin = position
		l944:
			{
				position945, thunkPosition945 := position, thunkPosition
				{
					position946, thunkPosition946 := position, thunkPosition
					if !matchChar('"') {
						goto l946
					}
					if !p.rules[ruleSp]() {
						goto l946
					}
					{
						position947, thunkPosition947 := position, thunkPosition
						if !matchChar(')') {
							goto l948
						}
						goto l947
					l948:
						position, thunkPosition = position947, thunkPosition947
						if !p.rules[ruleNewline]() {
							goto l946
						}
					}
				l947:
					goto l945
				l946:
					position, thunkPosition = position946, thunkPosition946
				}
				if !matchDot() {
					goto l945
				}
				goto l944
			l945:
				position, thunkPosition = position945, thunkPosition945
			}
			end = position
			if !matchChar('"') {
				goto l943
			}
			return true
		l943:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 176 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position950, thunkPosition950 := position, thunkPosition
				if !p.rules[ruleAutoLinkUrl]() {
					goto l951
				}
				goto l950
			l951:
				position, thunkPosition = position950, thunkPosition950
				if !p.rules[ruleAutoLinkEmail]() {
					goto l949
				}
			}
		l950:
			return true
		l949:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 177 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = mk_link(mk_str(yytext), yytext, "") }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l952
			}
			begin = position
			if !matchClass(4) {
				goto l952
			}
		l953:
			{
				position954, thunkPosition954 := position, thunkPosition
				if !matchClass(4) {
					goto l954
				}
				goto l953
			l954:
				position, thunkPosition = position954, thunkPosition954
			}
			if !matchString("://") {
				goto l952
			}
			{
				position957, thunkPosition957 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l957
				}
				goto l952
			l957:
				position, thunkPosition = position957, thunkPosition957
			}
			if peekChar('>') {
				goto l952
			}
			if !matchDot() {
				goto l952
			}
		l955:
			{
				position956, thunkPosition956 := position, thunkPosition
				{
					position958, thunkPosition958 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l958
					}
					goto l956
				l958:
					position, thunkPosition = position958, thunkPosition958
				}
				if peekChar('>') {
					goto l956
				}
				if !matchDot() {
					goto l956
				}
				goto l955
			l956:
				position, thunkPosition = position956, thunkPosition956
			}
			end = position
			if !matchChar('>') {
				goto l952
			}
			do(76)
			return true
		l952:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 178 AutoLinkEmail <- ('<' < [-A-Za-z0-9+_]+ '@' (!Newline !'>' .)+ > '>' {
                    yy = mk_link(mk_str(yytext), "mailto:"+yytext, "")
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l959
			}
			begin = position
			if !matchClass(9) {
				goto l959
			}
		l960:
			{
				position961, thunkPosition961 := position, thunkPosition
				if !matchClass(9) {
					goto l961
				}
				goto l960
			l961:
				position, thunkPosition = position961, thunkPosition961
			}
			if !matchChar('@') {
				goto l959
			}
			{
				position964, thunkPosition964 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l964
				}
				goto l959
			l964:
				position, thunkPosition = position964, thunkPosition964
			}
			if peekChar('>') {
				goto l959
			}
			if !matchDot() {
				goto l959
			}
		l962:
			{
				position963, thunkPosition963 := position, thunkPosition
				{
					position965, thunkPosition965 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l965
					}
					goto l963
				l965:
					position, thunkPosition = position965, thunkPosition965
				}
				if peekChar('>') {
					goto l963
				}
				if !matchDot() {
					goto l963
				}
				goto l962
			l963:
				position, thunkPosition = position963, thunkPosition963
			}
			end = position
			if !matchChar('>') {
				goto l959
			}
			do(77)
			return true
		l959:
			position, thunkPosition = position0, thunkPosition0
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
				goto l966
			}
			{
				position967, thunkPosition967 := position, thunkPosition
				if !matchString("[]") {
					goto l967
				}
				goto l966
			l967:
				position, thunkPosition = position967, thunkPosition967
			}
			if !p.rules[ruleLabel]() {
				goto l966
			}
			doarg(yySet, -2)
			if !matchChar(':') {
				goto l966
			}
			if !p.rules[ruleSpnl]() {
				goto l966
			}
			if !p.rules[ruleRefSrc]() {
				goto l966
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l966
			}
			if !p.rules[ruleRefTitle]() {
				goto l966
			}
			doarg(yySet, -3)
		l968:
			{
				position969, thunkPosition969 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l969
				}
				goto l968
			l969:
				position, thunkPosition = position969, thunkPosition969
			}
			do(78)
			doarg(yyPop, 3)
			return true
		l966:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 180 Label <- ('[' ((!'^' &{ p.extension.Notes }) / (&. &{ !p.extension.Notes })) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l970
			}
			{
				position971, thunkPosition971 := position, thunkPosition
				if peekChar('^') {
					goto l972
				}
				if !( p.extension.Notes ) {
					goto l972
				}
				goto l971
			l972:
				position, thunkPosition = position971, thunkPosition971
				if !peekDot() {
					goto l970
				}
				if !( !p.extension.Notes ) {
					goto l970
				}
			}
		l971:
			if !p.rules[ruleStartList]() {
				goto l970
			}
			doarg(yySet, -1)
		l973:
			{
				position974, thunkPosition974 := position, thunkPosition
				if peekChar(']') {
					goto l974
				}
				if !p.rules[ruleInline]() {
					goto l974
				}
				do(79)
				goto l973
			l974:
				position, thunkPosition = position974, thunkPosition974
			}
			if !matchChar(']') {
				goto l970
			}
			do(80)
			doarg(yyPop, 1)
			return true
		l970:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 181 RefSrc <- (< Nonspacechar+ > { yy = mk_str(yytext)
           yy.key = HTML }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto l975
			}
		l976:
			{
				position977, thunkPosition977 := position, thunkPosition
				if !p.rules[ruleNonspacechar]() {
					goto l977
				}
				goto l976
			l977:
				position, thunkPosition = position977, thunkPosition977
			}
			end = position
			do(81)
			return true
		l975:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 182 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position979, thunkPosition979 := position, thunkPosition
				if !p.rules[ruleRefTitleSingle]() {
					goto l980
				}
				goto l979
			l980:
				position, thunkPosition = position979, thunkPosition979
				if !p.rules[ruleRefTitleDouble]() {
					goto l981
				}
				goto l979
			l981:
				position, thunkPosition = position979, thunkPosition979
				if !p.rules[ruleRefTitleParens]() {
					goto l982
				}
				goto l979
			l982:
				position, thunkPosition = position979, thunkPosition979
				if !p.rules[ruleEmptyTitle]() {
					goto l978
				}
			}
		l979:
			do(82)
			return true
		l978:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 183 EmptyTitle <- (< '' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !peekDot() {
				goto l983
			}
			end = position
			return true
		l983:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 184 RefTitleSingle <- ('\'' < (!(('\'' Sp Newline) / Newline) .)* > '\'') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l984
			}
			begin = position
		l985:
			{
				position986, thunkPosition986 := position, thunkPosition
				{
					position987, thunkPosition987 := position, thunkPosition
					{
						position988, thunkPosition988 := position, thunkPosition
						if !matchChar('\'') {
							goto l989
						}
						if !p.rules[ruleSp]() {
							goto l989
						}
						if !p.rules[ruleNewline]() {
							goto l989
						}
						goto l988
					l989:
						position, thunkPosition = position988, thunkPosition988
						if !p.rules[ruleNewline]() {
							goto l987
						}
					}
				l988:
					goto l986
				l987:
					position, thunkPosition = position987, thunkPosition987
				}
				if !matchDot() {
					goto l986
				}
				goto l985
			l986:
				position, thunkPosition = position986, thunkPosition986
			}
			end = position
			if !matchChar('\'') {
				goto l984
			}
			return true
		l984:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 185 RefTitleDouble <- ('"' < (!(('"' Sp Newline) / Newline) .)* > '"') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l990
			}
			begin = position
		l991:
			{
				position992, thunkPosition992 := position, thunkPosition
				{
					position993, thunkPosition993 := position, thunkPosition
					{
						position994, thunkPosition994 := position, thunkPosition
						if !matchChar('"') {
							goto l995
						}
						if !p.rules[ruleSp]() {
							goto l995
						}
						if !p.rules[ruleNewline]() {
							goto l995
						}
						goto l994
					l995:
						position, thunkPosition = position994, thunkPosition994
						if !p.rules[ruleNewline]() {
							goto l993
						}
					}
				l994:
					goto l992
				l993:
					position, thunkPosition = position993, thunkPosition993
				}
				if !matchDot() {
					goto l992
				}
				goto l991
			l992:
				position, thunkPosition = position992, thunkPosition992
			}
			end = position
			if !matchChar('"') {
				goto l990
			}
			return true
		l990:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 186 RefTitleParens <- ('(' < (!((')' Sp Newline) / Newline) .)* > ')') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('(') {
				goto l996
			}
			begin = position
		l997:
			{
				position998, thunkPosition998 := position, thunkPosition
				{
					position999, thunkPosition999 := position, thunkPosition
					{
						position1000, thunkPosition1000 := position, thunkPosition
						if !matchChar(')') {
							goto l1001
						}
						if !p.rules[ruleSp]() {
							goto l1001
						}
						if !p.rules[ruleNewline]() {
							goto l1001
						}
						goto l1000
					l1001:
						position, thunkPosition = position1000, thunkPosition1000
						if !p.rules[ruleNewline]() {
							goto l999
						}
					}
				l1000:
					goto l998
				l999:
					position, thunkPosition = position999, thunkPosition999
				}
				if !matchDot() {
					goto l998
				}
				goto l997
			l998:
				position, thunkPosition = position998, thunkPosition998
			}
			end = position
			if !matchChar(')') {
				goto l996
			}
			return true
		l996:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 187 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1002
			}
			doarg(yySet, -1)
		l1003:
			{
				position1004, thunkPosition1004 := position, thunkPosition
				{
					position1005, thunkPosition1005 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto l1006
					}
					doarg(yySet, -2)
					do(83)
					goto l1005
				l1006:
					position, thunkPosition = position1005, thunkPosition1005
					if !p.rules[ruleSkipBlock]() {
						goto l1004
					}
				}
			l1005:
				goto l1003
			l1004:
				position, thunkPosition = position1004, thunkPosition1004
			}
			do(84)
			if !(commit(thunkPosition0)) {
				goto l1002
			}
			doarg(yyPop, 2)
			return true
		l1002:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 188 Ticks1 <- ('`' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('`') {
				goto l1007
			}
			if peekChar('`') {
				goto l1007
			}
			return true
		l1007:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 189 Ticks2 <- ('``' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("``") {
				goto l1008
			}
			if peekChar('`') {
				goto l1008
			}
			return true
		l1008:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 190 Ticks3 <- ('```' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("```") {
				goto l1009
			}
			if peekChar('`') {
				goto l1009
			}
			return true
		l1009:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 191 Ticks4 <- ('````' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("````") {
				goto l1010
			}
			if peekChar('`') {
				goto l1010
			}
			return true
		l1010:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 192 Ticks5 <- ('`````' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("`````") {
				goto l1011
			}
			if peekChar('`') {
				goto l1011
			}
			return true
		l1011:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 193 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / (!Ticks1 '`'+) / (!(Sp Ticks1) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / (!Ticks2 '`'+) / (!(Sp Ticks2) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / (!Ticks3 '`'+) / (!(Sp Ticks3) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / (!Ticks4 '`'+) / (!(Sp Ticks4) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / (!Ticks5 '`'+) / (!(Sp Ticks5) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks5)) { yy = mk_str(yytext); yy.key = CODE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1013, thunkPosition1013 := position, thunkPosition
				if !p.rules[ruleTicks1]() {
					goto l1014
				}
				if !p.rules[ruleSp]() {
					goto l1014
				}
				begin = position
				{
					position1017, thunkPosition1017 := position, thunkPosition
					if peekChar('`') {
						goto l1018
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1018
					}
				l1019:
					{
						position1020, thunkPosition1020 := position, thunkPosition
						if peekChar('`') {
							goto l1020
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1020
						}
						goto l1019
					l1020:
						position, thunkPosition = position1020, thunkPosition1020
					}
					goto l1017
				l1018:
					position, thunkPosition = position1017, thunkPosition1017
					{
						position1022, thunkPosition1022 := position, thunkPosition
						if !p.rules[ruleTicks1]() {
							goto l1022
						}
						goto l1021
					l1022:
						position, thunkPosition = position1022, thunkPosition1022
					}
					if !matchChar('`') {
						goto l1021
					}
				l1023:
					{
						position1024, thunkPosition1024 := position, thunkPosition
						if !matchChar('`') {
							goto l1024
						}
						goto l1023
					l1024:
						position, thunkPosition = position1024, thunkPosition1024
					}
					goto l1017
				l1021:
					position, thunkPosition = position1017, thunkPosition1017
					{
						position1025, thunkPosition1025 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1025
						}
						if !p.rules[ruleTicks1]() {
							goto l1025
						}
						goto l1014
					l1025:
						position, thunkPosition = position1025, thunkPosition1025
					}
					{
						position1026, thunkPosition1026 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1027
						}
						goto l1026
					l1027:
						position, thunkPosition = position1026, thunkPosition1026
						if !p.rules[ruleNewline]() {
							goto l1014
						}
						{
							position1028, thunkPosition1028 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1028
							}
							goto l1014
						l1028:
							position, thunkPosition = position1028, thunkPosition1028
						}
					}
				l1026:
				}
			l1017:
			l1015:
				{
					position1016, thunkPosition1016 := position, thunkPosition
					{
						position1029, thunkPosition1029 := position, thunkPosition
						if peekChar('`') {
							goto l1030
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1030
						}
					l1031:
						{
							position1032, thunkPosition1032 := position, thunkPosition
							if peekChar('`') {
								goto l1032
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1032
							}
							goto l1031
						l1032:
							position, thunkPosition = position1032, thunkPosition1032
						}
						goto l1029
					l1030:
						position, thunkPosition = position1029, thunkPosition1029
						{
							position1034, thunkPosition1034 := position, thunkPosition
							if !p.rules[ruleTicks1]() {
								goto l1034
							}
							goto l1033
						l1034:
							position, thunkPosition = position1034, thunkPosition1034
						}
						if !matchChar('`') {
							goto l1033
						}
					l1035:
						{
							position1036, thunkPosition1036 := position, thunkPosition
							if !matchChar('`') {
								goto l1036
							}
							goto l1035
						l1036:
							position, thunkPosition = position1036, thunkPosition1036
						}
						goto l1029
					l1033:
						position, thunkPosition = position1029, thunkPosition1029
						{
							position1037, thunkPosition1037 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1037
							}
							if !p.rules[ruleTicks1]() {
								goto l1037
							}
							goto l1016
						l1037:
							position, thunkPosition = position1037, thunkPosition1037
						}
						{
							position1038, thunkPosition1038 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1039
							}
							goto l1038
						l1039:
							position, thunkPosition = position1038, thunkPosition1038
							if !p.rules[ruleNewline]() {
								goto l1016
							}
							{
								position1040, thunkPosition1040 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1040
								}
								goto l1016
							l1040:
								position, thunkPosition = position1040, thunkPosition1040
							}
						}
					l1038:
					}
				l1029:
					goto l1015
				l1016:
					position, thunkPosition = position1016, thunkPosition1016
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1014
				}
				if !p.rules[ruleTicks1]() {
					goto l1014
				}
				goto l1013
			l1014:
				position, thunkPosition = position1013, thunkPosition1013
				if !p.rules[ruleTicks2]() {
					goto l1041
				}
				if !p.rules[ruleSp]() {
					goto l1041
				}
				begin = position
				{
					position1044, thunkPosition1044 := position, thunkPosition
					if peekChar('`') {
						goto l1045
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1045
					}
				l1046:
					{
						position1047, thunkPosition1047 := position, thunkPosition
						if peekChar('`') {
							goto l1047
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1047
						}
						goto l1046
					l1047:
						position, thunkPosition = position1047, thunkPosition1047
					}
					goto l1044
				l1045:
					position, thunkPosition = position1044, thunkPosition1044
					{
						position1049, thunkPosition1049 := position, thunkPosition
						if !p.rules[ruleTicks2]() {
							goto l1049
						}
						goto l1048
					l1049:
						position, thunkPosition = position1049, thunkPosition1049
					}
					if !matchChar('`') {
						goto l1048
					}
				l1050:
					{
						position1051, thunkPosition1051 := position, thunkPosition
						if !matchChar('`') {
							goto l1051
						}
						goto l1050
					l1051:
						position, thunkPosition = position1051, thunkPosition1051
					}
					goto l1044
				l1048:
					position, thunkPosition = position1044, thunkPosition1044
					{
						position1052, thunkPosition1052 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1052
						}
						if !p.rules[ruleTicks2]() {
							goto l1052
						}
						goto l1041
					l1052:
						position, thunkPosition = position1052, thunkPosition1052
					}
					{
						position1053, thunkPosition1053 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1054
						}
						goto l1053
					l1054:
						position, thunkPosition = position1053, thunkPosition1053
						if !p.rules[ruleNewline]() {
							goto l1041
						}
						{
							position1055, thunkPosition1055 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1055
							}
							goto l1041
						l1055:
							position, thunkPosition = position1055, thunkPosition1055
						}
					}
				l1053:
				}
			l1044:
			l1042:
				{
					position1043, thunkPosition1043 := position, thunkPosition
					{
						position1056, thunkPosition1056 := position, thunkPosition
						if peekChar('`') {
							goto l1057
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1057
						}
					l1058:
						{
							position1059, thunkPosition1059 := position, thunkPosition
							if peekChar('`') {
								goto l1059
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1059
							}
							goto l1058
						l1059:
							position, thunkPosition = position1059, thunkPosition1059
						}
						goto l1056
					l1057:
						position, thunkPosition = position1056, thunkPosition1056
						{
							position1061, thunkPosition1061 := position, thunkPosition
							if !p.rules[ruleTicks2]() {
								goto l1061
							}
							goto l1060
						l1061:
							position, thunkPosition = position1061, thunkPosition1061
						}
						if !matchChar('`') {
							goto l1060
						}
					l1062:
						{
							position1063, thunkPosition1063 := position, thunkPosition
							if !matchChar('`') {
								goto l1063
							}
							goto l1062
						l1063:
							position, thunkPosition = position1063, thunkPosition1063
						}
						goto l1056
					l1060:
						position, thunkPosition = position1056, thunkPosition1056
						{
							position1064, thunkPosition1064 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1064
							}
							if !p.rules[ruleTicks2]() {
								goto l1064
							}
							goto l1043
						l1064:
							position, thunkPosition = position1064, thunkPosition1064
						}
						{
							position1065, thunkPosition1065 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1066
							}
							goto l1065
						l1066:
							position, thunkPosition = position1065, thunkPosition1065
							if !p.rules[ruleNewline]() {
								goto l1043
							}
							{
								position1067, thunkPosition1067 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1067
								}
								goto l1043
							l1067:
								position, thunkPosition = position1067, thunkPosition1067
							}
						}
					l1065:
					}
				l1056:
					goto l1042
				l1043:
					position, thunkPosition = position1043, thunkPosition1043
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1041
				}
				if !p.rules[ruleTicks2]() {
					goto l1041
				}
				goto l1013
			l1041:
				position, thunkPosition = position1013, thunkPosition1013
				if !p.rules[ruleTicks3]() {
					goto l1068
				}
				if !p.rules[ruleSp]() {
					goto l1068
				}
				begin = position
				{
					position1071, thunkPosition1071 := position, thunkPosition
					if peekChar('`') {
						goto l1072
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1072
					}
				l1073:
					{
						position1074, thunkPosition1074 := position, thunkPosition
						if peekChar('`') {
							goto l1074
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1074
						}
						goto l1073
					l1074:
						position, thunkPosition = position1074, thunkPosition1074
					}
					goto l1071
				l1072:
					position, thunkPosition = position1071, thunkPosition1071
					{
						position1076, thunkPosition1076 := position, thunkPosition
						if !p.rules[ruleTicks3]() {
							goto l1076
						}
						goto l1075
					l1076:
						position, thunkPosition = position1076, thunkPosition1076
					}
					if !matchChar('`') {
						goto l1075
					}
				l1077:
					{
						position1078, thunkPosition1078 := position, thunkPosition
						if !matchChar('`') {
							goto l1078
						}
						goto l1077
					l1078:
						position, thunkPosition = position1078, thunkPosition1078
					}
					goto l1071
				l1075:
					position, thunkPosition = position1071, thunkPosition1071
					{
						position1079, thunkPosition1079 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1079
						}
						if !p.rules[ruleTicks3]() {
							goto l1079
						}
						goto l1068
					l1079:
						position, thunkPosition = position1079, thunkPosition1079
					}
					{
						position1080, thunkPosition1080 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1081
						}
						goto l1080
					l1081:
						position, thunkPosition = position1080, thunkPosition1080
						if !p.rules[ruleNewline]() {
							goto l1068
						}
						{
							position1082, thunkPosition1082 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1082
							}
							goto l1068
						l1082:
							position, thunkPosition = position1082, thunkPosition1082
						}
					}
				l1080:
				}
			l1071:
			l1069:
				{
					position1070, thunkPosition1070 := position, thunkPosition
					{
						position1083, thunkPosition1083 := position, thunkPosition
						if peekChar('`') {
							goto l1084
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1084
						}
					l1085:
						{
							position1086, thunkPosition1086 := position, thunkPosition
							if peekChar('`') {
								goto l1086
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1086
							}
							goto l1085
						l1086:
							position, thunkPosition = position1086, thunkPosition1086
						}
						goto l1083
					l1084:
						position, thunkPosition = position1083, thunkPosition1083
						{
							position1088, thunkPosition1088 := position, thunkPosition
							if !p.rules[ruleTicks3]() {
								goto l1088
							}
							goto l1087
						l1088:
							position, thunkPosition = position1088, thunkPosition1088
						}
						if !matchChar('`') {
							goto l1087
						}
					l1089:
						{
							position1090, thunkPosition1090 := position, thunkPosition
							if !matchChar('`') {
								goto l1090
							}
							goto l1089
						l1090:
							position, thunkPosition = position1090, thunkPosition1090
						}
						goto l1083
					l1087:
						position, thunkPosition = position1083, thunkPosition1083
						{
							position1091, thunkPosition1091 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1091
							}
							if !p.rules[ruleTicks3]() {
								goto l1091
							}
							goto l1070
						l1091:
							position, thunkPosition = position1091, thunkPosition1091
						}
						{
							position1092, thunkPosition1092 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1093
							}
							goto l1092
						l1093:
							position, thunkPosition = position1092, thunkPosition1092
							if !p.rules[ruleNewline]() {
								goto l1070
							}
							{
								position1094, thunkPosition1094 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1094
								}
								goto l1070
							l1094:
								position, thunkPosition = position1094, thunkPosition1094
							}
						}
					l1092:
					}
				l1083:
					goto l1069
				l1070:
					position, thunkPosition = position1070, thunkPosition1070
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1068
				}
				if !p.rules[ruleTicks3]() {
					goto l1068
				}
				goto l1013
			l1068:
				position, thunkPosition = position1013, thunkPosition1013
				if !p.rules[ruleTicks4]() {
					goto l1095
				}
				if !p.rules[ruleSp]() {
					goto l1095
				}
				begin = position
				{
					position1098, thunkPosition1098 := position, thunkPosition
					if peekChar('`') {
						goto l1099
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1099
					}
				l1100:
					{
						position1101, thunkPosition1101 := position, thunkPosition
						if peekChar('`') {
							goto l1101
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1101
						}
						goto l1100
					l1101:
						position, thunkPosition = position1101, thunkPosition1101
					}
					goto l1098
				l1099:
					position, thunkPosition = position1098, thunkPosition1098
					{
						position1103, thunkPosition1103 := position, thunkPosition
						if !p.rules[ruleTicks4]() {
							goto l1103
						}
						goto l1102
					l1103:
						position, thunkPosition = position1103, thunkPosition1103
					}
					if !matchChar('`') {
						goto l1102
					}
				l1104:
					{
						position1105, thunkPosition1105 := position, thunkPosition
						if !matchChar('`') {
							goto l1105
						}
						goto l1104
					l1105:
						position, thunkPosition = position1105, thunkPosition1105
					}
					goto l1098
				l1102:
					position, thunkPosition = position1098, thunkPosition1098
					{
						position1106, thunkPosition1106 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1106
						}
						if !p.rules[ruleTicks4]() {
							goto l1106
						}
						goto l1095
					l1106:
						position, thunkPosition = position1106, thunkPosition1106
					}
					{
						position1107, thunkPosition1107 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1108
						}
						goto l1107
					l1108:
						position, thunkPosition = position1107, thunkPosition1107
						if !p.rules[ruleNewline]() {
							goto l1095
						}
						{
							position1109, thunkPosition1109 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1109
							}
							goto l1095
						l1109:
							position, thunkPosition = position1109, thunkPosition1109
						}
					}
				l1107:
				}
			l1098:
			l1096:
				{
					position1097, thunkPosition1097 := position, thunkPosition
					{
						position1110, thunkPosition1110 := position, thunkPosition
						if peekChar('`') {
							goto l1111
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1111
						}
					l1112:
						{
							position1113, thunkPosition1113 := position, thunkPosition
							if peekChar('`') {
								goto l1113
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1113
							}
							goto l1112
						l1113:
							position, thunkPosition = position1113, thunkPosition1113
						}
						goto l1110
					l1111:
						position, thunkPosition = position1110, thunkPosition1110
						{
							position1115, thunkPosition1115 := position, thunkPosition
							if !p.rules[ruleTicks4]() {
								goto l1115
							}
							goto l1114
						l1115:
							position, thunkPosition = position1115, thunkPosition1115
						}
						if !matchChar('`') {
							goto l1114
						}
					l1116:
						{
							position1117, thunkPosition1117 := position, thunkPosition
							if !matchChar('`') {
								goto l1117
							}
							goto l1116
						l1117:
							position, thunkPosition = position1117, thunkPosition1117
						}
						goto l1110
					l1114:
						position, thunkPosition = position1110, thunkPosition1110
						{
							position1118, thunkPosition1118 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1118
							}
							if !p.rules[ruleTicks4]() {
								goto l1118
							}
							goto l1097
						l1118:
							position, thunkPosition = position1118, thunkPosition1118
						}
						{
							position1119, thunkPosition1119 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1120
							}
							goto l1119
						l1120:
							position, thunkPosition = position1119, thunkPosition1119
							if !p.rules[ruleNewline]() {
								goto l1097
							}
							{
								position1121, thunkPosition1121 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1121
								}
								goto l1097
							l1121:
								position, thunkPosition = position1121, thunkPosition1121
							}
						}
					l1119:
					}
				l1110:
					goto l1096
				l1097:
					position, thunkPosition = position1097, thunkPosition1097
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1095
				}
				if !p.rules[ruleTicks4]() {
					goto l1095
				}
				goto l1013
			l1095:
				position, thunkPosition = position1013, thunkPosition1013
				if !p.rules[ruleTicks5]() {
					goto l1012
				}
				if !p.rules[ruleSp]() {
					goto l1012
				}
				begin = position
				{
					position1124, thunkPosition1124 := position, thunkPosition
					if peekChar('`') {
						goto l1125
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1125
					}
				l1126:
					{
						position1127, thunkPosition1127 := position, thunkPosition
						if peekChar('`') {
							goto l1127
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1127
						}
						goto l1126
					l1127:
						position, thunkPosition = position1127, thunkPosition1127
					}
					goto l1124
				l1125:
					position, thunkPosition = position1124, thunkPosition1124
					{
						position1129, thunkPosition1129 := position, thunkPosition
						if !p.rules[ruleTicks5]() {
							goto l1129
						}
						goto l1128
					l1129:
						position, thunkPosition = position1129, thunkPosition1129
					}
					if !matchChar('`') {
						goto l1128
					}
				l1130:
					{
						position1131, thunkPosition1131 := position, thunkPosition
						if !matchChar('`') {
							goto l1131
						}
						goto l1130
					l1131:
						position, thunkPosition = position1131, thunkPosition1131
					}
					goto l1124
				l1128:
					position, thunkPosition = position1124, thunkPosition1124
					{
						position1132, thunkPosition1132 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1132
						}
						if !p.rules[ruleTicks5]() {
							goto l1132
						}
						goto l1012
					l1132:
						position, thunkPosition = position1132, thunkPosition1132
					}
					{
						position1133, thunkPosition1133 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1134
						}
						goto l1133
					l1134:
						position, thunkPosition = position1133, thunkPosition1133
						if !p.rules[ruleNewline]() {
							goto l1012
						}
						{
							position1135, thunkPosition1135 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1135
							}
							goto l1012
						l1135:
							position, thunkPosition = position1135, thunkPosition1135
						}
					}
				l1133:
				}
			l1124:
			l1122:
				{
					position1123, thunkPosition1123 := position, thunkPosition
					{
						position1136, thunkPosition1136 := position, thunkPosition
						if peekChar('`') {
							goto l1137
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1137
						}
					l1138:
						{
							position1139, thunkPosition1139 := position, thunkPosition
							if peekChar('`') {
								goto l1139
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1139
							}
							goto l1138
						l1139:
							position, thunkPosition = position1139, thunkPosition1139
						}
						goto l1136
					l1137:
						position, thunkPosition = position1136, thunkPosition1136
						{
							position1141, thunkPosition1141 := position, thunkPosition
							if !p.rules[ruleTicks5]() {
								goto l1141
							}
							goto l1140
						l1141:
							position, thunkPosition = position1141, thunkPosition1141
						}
						if !matchChar('`') {
							goto l1140
						}
					l1142:
						{
							position1143, thunkPosition1143 := position, thunkPosition
							if !matchChar('`') {
								goto l1143
							}
							goto l1142
						l1143:
							position, thunkPosition = position1143, thunkPosition1143
						}
						goto l1136
					l1140:
						position, thunkPosition = position1136, thunkPosition1136
						{
							position1144, thunkPosition1144 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1144
							}
							if !p.rules[ruleTicks5]() {
								goto l1144
							}
							goto l1123
						l1144:
							position, thunkPosition = position1144, thunkPosition1144
						}
						{
							position1145, thunkPosition1145 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1146
							}
							goto l1145
						l1146:
							position, thunkPosition = position1145, thunkPosition1145
							if !p.rules[ruleNewline]() {
								goto l1123
							}
							{
								position1147, thunkPosition1147 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1147
								}
								goto l1123
							l1147:
								position, thunkPosition = position1147, thunkPosition1147
							}
						}
					l1145:
					}
				l1136:
					goto l1122
				l1123:
					position, thunkPosition = position1123, thunkPosition1123
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1012
				}
				if !p.rules[ruleTicks5]() {
					goto l1012
				}
			}
		l1013:
			do(85)
			return true
		l1012:
			position, thunkPosition = position0, thunkPosition0
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
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			{
				position1149, thunkPosition1149 := position, thunkPosition
				if !p.rules[ruleHtmlComment]() {
					goto l1150
				}
				goto l1149
			l1150:
				position, thunkPosition = position1149, thunkPosition1149
				if !p.rules[ruleHtmlTag]() {
					goto l1148
				}
			}
		l1149:
			end = position
			do(86)
			return true
		l1148:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 195 BlankLine <- (Sp Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l1151
			}
			if !p.rules[ruleNewline]() {
				goto l1151
			}
			return true
		l1151:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 196 Quoted <- (('"' (!'"' .)* '"') / ('\'' (!'\'' .)* '\'')) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1153, thunkPosition1153 := position, thunkPosition
				if !matchChar('"') {
					goto l1154
				}
			l1155:
				{
					position1156, thunkPosition1156 := position, thunkPosition
					if peekChar('"') {
						goto l1156
					}
					if !matchDot() {
						goto l1156
					}
					goto l1155
				l1156:
					position, thunkPosition = position1156, thunkPosition1156
				}
				if !matchChar('"') {
					goto l1154
				}
				goto l1153
			l1154:
				position, thunkPosition = position1153, thunkPosition1153
				if !matchChar('\'') {
					goto l1152
				}
			l1157:
				{
					position1158, thunkPosition1158 := position, thunkPosition
					if peekChar('\'') {
						goto l1158
					}
					if !matchDot() {
						goto l1158
					}
					goto l1157
				l1158:
					position, thunkPosition = position1158, thunkPosition1158
				}
				if !matchChar('\'') {
					goto l1152
				}
			}
		l1153:
			return true
		l1152:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 197 HtmlAttribute <- ((AlphanumericAscii / '-')+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1162, thunkPosition1162 := position, thunkPosition
				if !p.rules[ruleAlphanumericAscii]() {
					goto l1163
				}
				goto l1162
			l1163:
				position, thunkPosition = position1162, thunkPosition1162
				if !matchChar('-') {
					goto l1159
				}
			}
		l1162:
		l1160:
			{
				position1161, thunkPosition1161 := position, thunkPosition
				{
					position1164, thunkPosition1164 := position, thunkPosition
					if !p.rules[ruleAlphanumericAscii]() {
						goto l1165
					}
					goto l1164
				l1165:
					position, thunkPosition = position1164, thunkPosition1164
					if !matchChar('-') {
						goto l1161
					}
				}
			l1164:
				goto l1160
			l1161:
				position, thunkPosition = position1161, thunkPosition1161
			}
			if !p.rules[ruleSpnl]() {
				goto l1159
			}
			{
				position1166, thunkPosition1166 := position, thunkPosition
				if !matchChar('=') {
					goto l1166
				}
				if !p.rules[ruleSpnl]() {
					goto l1166
				}
				{
					position1168, thunkPosition1168 := position, thunkPosition
					if !p.rules[ruleQuoted]() {
						goto l1169
					}
					goto l1168
				l1169:
					position, thunkPosition = position1168, thunkPosition1168
					if peekChar('>') {
						goto l1166
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1166
					}
				l1170:
					{
						position1171, thunkPosition1171 := position, thunkPosition
						if peekChar('>') {
							goto l1171
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1171
						}
						goto l1170
					l1171:
						position, thunkPosition = position1171, thunkPosition1171
					}
				}
			l1168:
				goto l1167
			l1166:
				position, thunkPosition = position1166, thunkPosition1166
			}
		l1167:
			if !p.rules[ruleSpnl]() {
				goto l1159
			}
			return true
		l1159:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 198 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("<!--") {
				goto l1172
			}
		l1173:
			{
				position1174, thunkPosition1174 := position, thunkPosition
				{
					position1175, thunkPosition1175 := position, thunkPosition
					if !matchString("-->") {
						goto l1175
					}
					goto l1174
				l1175:
					position, thunkPosition = position1175, thunkPosition1175
				}
				if !matchDot() {
					goto l1174
				}
				goto l1173
			l1174:
				position, thunkPosition = position1174, thunkPosition1174
			}
			if !matchString("-->") {
				goto l1172
			}
			return true
		l1172:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 199 HtmlTag <- ('<' Spnl '/'? AlphanumericAscii+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l1176
			}
			if !p.rules[ruleSpnl]() {
				goto l1176
			}
			{
				position1177, thunkPosition1177 := position, thunkPosition
				if !matchChar('/') {
					goto l1177
				}
				goto l1178
			l1177:
				position, thunkPosition = position1177, thunkPosition1177
			}
		l1178:
			if !p.rules[ruleAlphanumericAscii]() {
				goto l1176
			}
		l1179:
			{
				position1180, thunkPosition1180 := position, thunkPosition
				if !p.rules[ruleAlphanumericAscii]() {
					goto l1180
				}
				goto l1179
			l1180:
				position, thunkPosition = position1180, thunkPosition1180
			}
			if !p.rules[ruleSpnl]() {
				goto l1176
			}
		l1181:
			{
				position1182, thunkPosition1182 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l1182
				}
				goto l1181
			l1182:
				position, thunkPosition = position1182, thunkPosition1182
			}
			{
				position1183, thunkPosition1183 := position, thunkPosition
				if !matchChar('/') {
					goto l1183
				}
				goto l1184
			l1183:
				position, thunkPosition = position1183, thunkPosition1183
			}
		l1184:
			if !p.rules[ruleSpnl]() {
				goto l1176
			}
			if !matchChar('>') {
				goto l1176
			}
			return true
		l1176:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 200 Eof <- !. */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if peekDot() {
				goto l1185
			}
			return true
		l1185:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 201 Spacechar <- (' ' / '\t') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1187, thunkPosition1187 := position, thunkPosition
				if !matchChar(' ') {
					goto l1188
				}
				goto l1187
			l1188:
				position, thunkPosition = position1187, thunkPosition1187
				if !matchChar('\t') {
					goto l1186
				}
			}
		l1187:
			return true
		l1186:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 202 Nonspacechar <- (!Spacechar !Newline .) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1190, thunkPosition1190 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1190
				}
				goto l1189
			l1190:
				position, thunkPosition = position1190, thunkPosition1190
			}
			{
				position1191, thunkPosition1191 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1191
				}
				goto l1189
			l1191:
				position, thunkPosition = position1191, thunkPosition1191
			}
			if !matchDot() {
				goto l1189
			}
			return true
		l1189:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 203 Newline <- ('\n' / ('\r' '\n'?)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1193, thunkPosition1193 := position, thunkPosition
				if !matchChar('\n') {
					goto l1194
				}
				goto l1193
			l1194:
				position, thunkPosition = position1193, thunkPosition1193
				if !matchChar('\r') {
					goto l1192
				}
				{
					position1195, thunkPosition1195 := position, thunkPosition
					if !matchChar('\n') {
						goto l1195
					}
					goto l1196
				l1195:
					position, thunkPosition = position1195, thunkPosition1195
				}
			l1196:
			}
		l1193:
			return true
		l1192:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 204 Sp <- Spacechar* */
		func() bool {
		l1198:
			{
				position1199, thunkPosition1199 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1199
				}
				goto l1198
			l1199:
				position, thunkPosition = position1199, thunkPosition1199
			}
			return true
		},
		/* 205 Spnl <- (Sp (Newline Sp)?) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l1200
			}
			{
				position1201, thunkPosition1201 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1201
				}
				if !p.rules[ruleSp]() {
					goto l1201
				}
				goto l1202
			l1201:
				position, thunkPosition = position1201, thunkPosition1201
			}
		l1202:
			return true
		l1200:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 206 SpecialChar <- ((&[\\] '\\') | (&[#] '#') | (&[!] '!') | (&[<] '<') | (&[\]] ']') | (&[\[] '[') | (&[&] '&') | (&[`] '`') | (&[_] '_') | (&[*] '*') | (&[\"\'\-.^] ExtendedSpecialChar)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				if position == len(p.Buffer) {
					goto l1203
				}
				switch p.Buffer[position] {
				case '\\':
					if !matchChar('\\') {
						goto l1203
					}
				case '#':
					if !matchChar('#') {
						goto l1203
					}
				case '!':
					if !matchChar('!') {
						goto l1203
					}
				case '<':
					if !matchChar('<') {
						goto l1203
					}
				case ']':
					if !matchChar(']') {
						goto l1203
					}
				case '[':
					if !matchChar('[') {
						goto l1203
					}
				case '&':
					if !matchChar('&') {
						goto l1203
					}
				case '`':
					if !matchChar('`') {
						goto l1203
					}
				case '_':
					if !matchChar('_') {
						goto l1203
					}
				case '*':
					if !matchChar('*') {
						goto l1203
					}
				default:
					if !p.rules[ruleExtendedSpecialChar]() {
						goto l1203
					}
				}
			}
			return true
		l1203:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 207 NormalChar <- (!((&[\n\r] Newline) | (&[\t ] Spacechar) | (&[!-#&\'*\-.<\[-`] SpecialChar)) .) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1206, thunkPosition1206 := position, thunkPosition
				{
					if position == len(p.Buffer) {
						goto l1206
					}
					switch p.Buffer[position] {
					case '\n', '\r':
						if !p.rules[ruleNewline]() {
							goto l1206
						}
					case '\t', ' ':
						if !p.rules[ruleSpacechar]() {
							goto l1206
						}
					default:
						if !p.rules[ruleSpecialChar]() {
							goto l1206
						}
					}
				}
				goto l1205
			l1206:
				position, thunkPosition = position1206, thunkPosition1206
			}
			if !matchDot() {
				goto l1205
			}
			return true
		l1205:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 208 NonAlphanumeric <- [\000-\057\072-\100\133-\140\173-\177] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(3) {
				goto l1208
			}
			return true
		l1208:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 209 Alphanumeric <- ([0-9A-Za-z] / '\200' / '\201' / '\202' / '\203' / '\204' / '\205' / '\206' / '\207' / '\210' / '\211' / '\212' / '\213' / '\214' / '\215' / '\216' / '\217' / '\220' / '\221' / '\222' / '\223' / '\224' / '\225' / '\226' / '\227' / '\230' / '\231' / '\232' / '\233' / '\234' / '\235' / '\236' / '\237' / '\240' / '\241' / '\242' / '\243' / '\244' / '\245' / '\246' / '\247' / '\250' / '\251' / '\252' / '\253' / '\254' / '\255' / '\256' / '\257' / '\260' / '\261' / '\262' / '\263' / '\264' / '\265' / '\266' / '\267' / '\270' / '\271' / '\272' / '\273' / '\274' / '\275' / '\276' / '\277' / '\300' / '\301' / '\302' / '\303' / '\304' / '\305' / '\306' / '\307' / '\310' / '\311' / '\312' / '\313' / '\314' / '\315' / '\316' / '\317' / '\320' / '\321' / '\322' / '\323' / '\324' / '\325' / '\326' / '\327' / '\330' / '\331' / '\332' / '\333' / '\334' / '\335' / '\336' / '\337' / '\340' / '\341' / '\342' / '\343' / '\344' / '\345' / '\346' / '\347' / '\350' / '\351' / '\352' / '\353' / '\354' / '\355' / '\356' / '\357' / '\360' / '\361' / '\362' / '\363' / '\364' / '\365' / '\366' / '\367' / '\370' / '\371' / '\372' / '\373' / '\374' / '\375' / '\376' / '\377') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1210, thunkPosition1210 := position, thunkPosition
				if !matchClass(1) {
					goto l1211
				}
				goto l1210
			l1211:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\200") {
					goto l1212
				}
				goto l1210
			l1212:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\201") {
					goto l1213
				}
				goto l1210
			l1213:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\202") {
					goto l1214
				}
				goto l1210
			l1214:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\203") {
					goto l1215
				}
				goto l1210
			l1215:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\204") {
					goto l1216
				}
				goto l1210
			l1216:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\205") {
					goto l1217
				}
				goto l1210
			l1217:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\206") {
					goto l1218
				}
				goto l1210
			l1218:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\207") {
					goto l1219
				}
				goto l1210
			l1219:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\210") {
					goto l1220
				}
				goto l1210
			l1220:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\211") {
					goto l1221
				}
				goto l1210
			l1221:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\212") {
					goto l1222
				}
				goto l1210
			l1222:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\213") {
					goto l1223
				}
				goto l1210
			l1223:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\214") {
					goto l1224
				}
				goto l1210
			l1224:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\215") {
					goto l1225
				}
				goto l1210
			l1225:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\216") {
					goto l1226
				}
				goto l1210
			l1226:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\217") {
					goto l1227
				}
				goto l1210
			l1227:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\220") {
					goto l1228
				}
				goto l1210
			l1228:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\221") {
					goto l1229
				}
				goto l1210
			l1229:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\222") {
					goto l1230
				}
				goto l1210
			l1230:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\223") {
					goto l1231
				}
				goto l1210
			l1231:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\224") {
					goto l1232
				}
				goto l1210
			l1232:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\225") {
					goto l1233
				}
				goto l1210
			l1233:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\226") {
					goto l1234
				}
				goto l1210
			l1234:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\227") {
					goto l1235
				}
				goto l1210
			l1235:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\230") {
					goto l1236
				}
				goto l1210
			l1236:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\231") {
					goto l1237
				}
				goto l1210
			l1237:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\232") {
					goto l1238
				}
				goto l1210
			l1238:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\233") {
					goto l1239
				}
				goto l1210
			l1239:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\234") {
					goto l1240
				}
				goto l1210
			l1240:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\235") {
					goto l1241
				}
				goto l1210
			l1241:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\236") {
					goto l1242
				}
				goto l1210
			l1242:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\237") {
					goto l1243
				}
				goto l1210
			l1243:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\240") {
					goto l1244
				}
				goto l1210
			l1244:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\241") {
					goto l1245
				}
				goto l1210
			l1245:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\242") {
					goto l1246
				}
				goto l1210
			l1246:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\243") {
					goto l1247
				}
				goto l1210
			l1247:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\244") {
					goto l1248
				}
				goto l1210
			l1248:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\245") {
					goto l1249
				}
				goto l1210
			l1249:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\246") {
					goto l1250
				}
				goto l1210
			l1250:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\247") {
					goto l1251
				}
				goto l1210
			l1251:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\250") {
					goto l1252
				}
				goto l1210
			l1252:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\251") {
					goto l1253
				}
				goto l1210
			l1253:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\252") {
					goto l1254
				}
				goto l1210
			l1254:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\253") {
					goto l1255
				}
				goto l1210
			l1255:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\254") {
					goto l1256
				}
				goto l1210
			l1256:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\255") {
					goto l1257
				}
				goto l1210
			l1257:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\256") {
					goto l1258
				}
				goto l1210
			l1258:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\257") {
					goto l1259
				}
				goto l1210
			l1259:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\260") {
					goto l1260
				}
				goto l1210
			l1260:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\261") {
					goto l1261
				}
				goto l1210
			l1261:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\262") {
					goto l1262
				}
				goto l1210
			l1262:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\263") {
					goto l1263
				}
				goto l1210
			l1263:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\264") {
					goto l1264
				}
				goto l1210
			l1264:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\265") {
					goto l1265
				}
				goto l1210
			l1265:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\266") {
					goto l1266
				}
				goto l1210
			l1266:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\267") {
					goto l1267
				}
				goto l1210
			l1267:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\270") {
					goto l1268
				}
				goto l1210
			l1268:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\271") {
					goto l1269
				}
				goto l1210
			l1269:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\272") {
					goto l1270
				}
				goto l1210
			l1270:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\273") {
					goto l1271
				}
				goto l1210
			l1271:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\274") {
					goto l1272
				}
				goto l1210
			l1272:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\275") {
					goto l1273
				}
				goto l1210
			l1273:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\276") {
					goto l1274
				}
				goto l1210
			l1274:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\277") {
					goto l1275
				}
				goto l1210
			l1275:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\300") {
					goto l1276
				}
				goto l1210
			l1276:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\301") {
					goto l1277
				}
				goto l1210
			l1277:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\302") {
					goto l1278
				}
				goto l1210
			l1278:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\303") {
					goto l1279
				}
				goto l1210
			l1279:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\304") {
					goto l1280
				}
				goto l1210
			l1280:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\305") {
					goto l1281
				}
				goto l1210
			l1281:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\306") {
					goto l1282
				}
				goto l1210
			l1282:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\307") {
					goto l1283
				}
				goto l1210
			l1283:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\310") {
					goto l1284
				}
				goto l1210
			l1284:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\311") {
					goto l1285
				}
				goto l1210
			l1285:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\312") {
					goto l1286
				}
				goto l1210
			l1286:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\313") {
					goto l1287
				}
				goto l1210
			l1287:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\314") {
					goto l1288
				}
				goto l1210
			l1288:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\315") {
					goto l1289
				}
				goto l1210
			l1289:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\316") {
					goto l1290
				}
				goto l1210
			l1290:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\317") {
					goto l1291
				}
				goto l1210
			l1291:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\320") {
					goto l1292
				}
				goto l1210
			l1292:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\321") {
					goto l1293
				}
				goto l1210
			l1293:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\322") {
					goto l1294
				}
				goto l1210
			l1294:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\323") {
					goto l1295
				}
				goto l1210
			l1295:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\324") {
					goto l1296
				}
				goto l1210
			l1296:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\325") {
					goto l1297
				}
				goto l1210
			l1297:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\326") {
					goto l1298
				}
				goto l1210
			l1298:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\327") {
					goto l1299
				}
				goto l1210
			l1299:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\330") {
					goto l1300
				}
				goto l1210
			l1300:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\331") {
					goto l1301
				}
				goto l1210
			l1301:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\332") {
					goto l1302
				}
				goto l1210
			l1302:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\333") {
					goto l1303
				}
				goto l1210
			l1303:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\334") {
					goto l1304
				}
				goto l1210
			l1304:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\335") {
					goto l1305
				}
				goto l1210
			l1305:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\336") {
					goto l1306
				}
				goto l1210
			l1306:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\337") {
					goto l1307
				}
				goto l1210
			l1307:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\340") {
					goto l1308
				}
				goto l1210
			l1308:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\341") {
					goto l1309
				}
				goto l1210
			l1309:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\342") {
					goto l1310
				}
				goto l1210
			l1310:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\343") {
					goto l1311
				}
				goto l1210
			l1311:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\344") {
					goto l1312
				}
				goto l1210
			l1312:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\345") {
					goto l1313
				}
				goto l1210
			l1313:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\346") {
					goto l1314
				}
				goto l1210
			l1314:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\347") {
					goto l1315
				}
				goto l1210
			l1315:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\350") {
					goto l1316
				}
				goto l1210
			l1316:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\351") {
					goto l1317
				}
				goto l1210
			l1317:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\352") {
					goto l1318
				}
				goto l1210
			l1318:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\353") {
					goto l1319
				}
				goto l1210
			l1319:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\354") {
					goto l1320
				}
				goto l1210
			l1320:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\355") {
					goto l1321
				}
				goto l1210
			l1321:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\356") {
					goto l1322
				}
				goto l1210
			l1322:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\357") {
					goto l1323
				}
				goto l1210
			l1323:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\360") {
					goto l1324
				}
				goto l1210
			l1324:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\361") {
					goto l1325
				}
				goto l1210
			l1325:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\362") {
					goto l1326
				}
				goto l1210
			l1326:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\363") {
					goto l1327
				}
				goto l1210
			l1327:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\364") {
					goto l1328
				}
				goto l1210
			l1328:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\365") {
					goto l1329
				}
				goto l1210
			l1329:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\366") {
					goto l1330
				}
				goto l1210
			l1330:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\367") {
					goto l1331
				}
				goto l1210
			l1331:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\370") {
					goto l1332
				}
				goto l1210
			l1332:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\371") {
					goto l1333
				}
				goto l1210
			l1333:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\372") {
					goto l1334
				}
				goto l1210
			l1334:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\373") {
					goto l1335
				}
				goto l1210
			l1335:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\374") {
					goto l1336
				}
				goto l1210
			l1336:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\375") {
					goto l1337
				}
				goto l1210
			l1337:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\376") {
					goto l1338
				}
				goto l1210
			l1338:
				position, thunkPosition = position1210, thunkPosition1210
				if !matchString("\377") {
					goto l1209
				}
			}
		l1210:
			return true
		l1209:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 210 AlphanumericAscii <- [A-Za-z0-9] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(8) {
				goto l1339
			}
			return true
		l1339:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 211 Digit <- [0-9] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(7) {
				goto l1340
			}
			return true
		l1340:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 212 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1341
			}
			if !matchChar('#') {
				goto l1341
			}
			if !matchClass(5) {
				goto l1341
			}
			if !matchClass(0) {
				goto l1341
			}
		l1342:
			{
				position1343, thunkPosition1343 := position, thunkPosition
				if !matchClass(0) {
					goto l1343
				}
				goto l1342
			l1343:
				position, thunkPosition = position1343, thunkPosition1343
			}
			if !matchChar(';') {
				goto l1341
			}
			end = position
			return true
		l1341:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 213 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1344
			}
			if !matchChar('#') {
				goto l1344
			}
			if !matchClass(7) {
				goto l1344
			}
		l1345:
			{
				position1346, thunkPosition1346 := position, thunkPosition
				if !matchClass(7) {
					goto l1346
				}
				goto l1345
			l1346:
				position, thunkPosition = position1346, thunkPosition1346
			}
			end = position
			if !matchChar(';') {
				goto l1344
			}
			end = position
			return true
		l1344:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 214 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1347
			}
			if !matchClass(8) {
				goto l1347
			}
		l1348:
			{
				position1349, thunkPosition1349 := position, thunkPosition
				if !matchClass(8) {
					goto l1349
				}
				goto l1348
			l1349:
				position, thunkPosition = position1349, thunkPosition1349
			}
			if !matchChar(';') {
				goto l1347
			}
			end = position
			return true
		l1347:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 215 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1351, thunkPosition1351 := position, thunkPosition
				if !matchString("   ") {
					goto l1352
				}
				goto l1351
			l1352:
				position, thunkPosition = position1351, thunkPosition1351
				if !matchString("  ") {
					goto l1353
				}
				goto l1351
			l1353:
				position, thunkPosition = position1351, thunkPosition1351
				if !matchChar(' ') {
					goto l1354
				}
				goto l1351
			l1354:
				position, thunkPosition = position1351, thunkPosition1351
				if !peekDot() {
					goto l1350
				}
			}
		l1351:
			return true
		l1350:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 216 Indent <- ('\t' / '    ') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1356, thunkPosition1356 := position, thunkPosition
				if !matchChar('\t') {
					goto l1357
				}
				goto l1356
			l1357:
				position, thunkPosition = position1356, thunkPosition1356
				if !matchString("    ") {
					goto l1355
				}
			}
		l1356:
			return true
		l1355:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 217 IndentedLine <- (Indent Line) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleIndent]() {
				goto l1358
			}
			if !p.rules[ruleLine]() {
				goto l1358
			}
			return true
		l1358:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 218 OptionallyIndentedLine <- (Indent? Line) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1360, thunkPosition1360 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l1360
				}
				goto l1361
			l1360:
				position, thunkPosition = position1360, thunkPosition1360
			}
		l1361:
			if !p.rules[ruleLine]() {
				goto l1359
			}
			return true
		l1359:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 219 StartList <- (&. { yy = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !peekDot() {
				goto l1362
			}
			do(87)
			return true
		l1362:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 220 Line <- (RawLine { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleRawLine]() {
				goto l1363
			}
			do(88)
			return true
		l1363:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 221 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > Eof)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1365, thunkPosition1365 := position, thunkPosition
				begin = position
			l1367:
				{
					position1368, thunkPosition1368 := position, thunkPosition
					if peekChar('\r') {
						goto l1368
					}
					if peekChar('\n') {
						goto l1368
					}
					if !matchDot() {
						goto l1368
					}
					goto l1367
				l1368:
					position, thunkPosition = position1368, thunkPosition1368
				}
				if !p.rules[ruleNewline]() {
					goto l1366
				}
				end = position
				goto l1365
			l1366:
				position, thunkPosition = position1365, thunkPosition1365
				begin = position
				if !matchDot() {
					goto l1364
				}
			l1369:
				{
					position1370, thunkPosition1370 := position, thunkPosition
					if !matchDot() {
						goto l1370
					}
					goto l1369
				l1370:
					position, thunkPosition = position1370, thunkPosition1370
				}
				end = position
				if !p.rules[ruleEof]() {
					goto l1364
				}
			}
		l1365:
			return true
		l1364:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 222 SkipBlock <- (((!BlankLine RawLine)+ BlankLine*) / BlankLine+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1372, thunkPosition1372 := position, thunkPosition
				{
					position1376, thunkPosition1376 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1376
					}
					goto l1373
				l1376:
					position, thunkPosition = position1376, thunkPosition1376
				}
				if !p.rules[ruleRawLine]() {
					goto l1373
				}
			l1374:
				{
					position1375, thunkPosition1375 := position, thunkPosition
					{
						position1377, thunkPosition1377 := position, thunkPosition
						if !p.rules[ruleBlankLine]() {
							goto l1377
						}
						goto l1375
					l1377:
						position, thunkPosition = position1377, thunkPosition1377
					}
					if !p.rules[ruleRawLine]() {
						goto l1375
					}
					goto l1374
				l1375:
					position, thunkPosition = position1375, thunkPosition1375
				}
			l1378:
				{
					position1379, thunkPosition1379 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1379
					}
					goto l1378
				l1379:
					position, thunkPosition = position1379, thunkPosition1379
				}
				goto l1372
			l1373:
				position, thunkPosition = position1372, thunkPosition1372
				if !p.rules[ruleBlankLine]() {
					goto l1371
				}
			l1380:
				{
					position1381, thunkPosition1381 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1381
					}
					goto l1380
				l1381:
					position, thunkPosition = position1381, thunkPosition1381
				}
			}
		l1372:
			return true
		l1371:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 223 ExtendedSpecialChar <- ((&{ p.extension.Smart } ((&[\"] '"') | (&[\'] '\'') | (&[\-] '-') | (&[.] '.'))) / (&{ p.extension.Notes } '^')) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1383, thunkPosition1383 := position, thunkPosition
				if !( p.extension.Smart ) {
					goto l1384
				}
				{
					if position == len(p.Buffer) {
						goto l1384
					}
					switch p.Buffer[position] {
					case '"':
						if !matchChar('"') {
							goto l1384
						}
					case '\'':
						if !matchChar('\'') {
							goto l1384
						}
					case '-':
						if !matchChar('-') {
							goto l1384
						}
					default:
						if !matchChar('.') {
							goto l1384
						}
					}
				}
				goto l1383
			l1384:
				position, thunkPosition = position1383, thunkPosition1383
				if !( p.extension.Notes ) {
					goto l1382
				}
				if !matchChar('^') {
					goto l1382
				}
			}
		l1383:
			return true
		l1382:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 224 Smart <- (&{ p.extension.Smart } (SingleQuoted / ((&[\'] Apostrophe) | (&[\"] DoubleQuoted) | (&[\-] Dash) | (&[.] Ellipsis)))) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !( p.extension.Smart ) {
				goto l1386
			}
			{
				position1387, thunkPosition1387 := position, thunkPosition
				if !p.rules[ruleSingleQuoted]() {
					goto l1388
				}
				goto l1387
			l1388:
				position, thunkPosition = position1387, thunkPosition1387
				{
					if position == len(p.Buffer) {
						goto l1386
					}
					switch p.Buffer[position] {
					case '\'':
						if !p.rules[ruleApostrophe]() {
							goto l1386
						}
					case '"':
						if !p.rules[ruleDoubleQuoted]() {
							goto l1386
						}
					case '-':
						if !p.rules[ruleDash]() {
							goto l1386
						}
					default:
						if !p.rules[ruleEllipsis]() {
							goto l1386
						}
					}
				}
			}
		l1387:
			return true
		l1386:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 225 Apostrophe <- ('\'' { yy = mk_element(APOSTROPHE) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1390
			}
			do(89)
			return true
		l1390:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 226 Ellipsis <- (('...' / '. . .') { yy = mk_element(ELLIPSIS) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1392, thunkPosition1392 := position, thunkPosition
				if !matchString("...") {
					goto l1393
				}
				goto l1392
			l1393:
				position, thunkPosition = position1392, thunkPosition1392
				if !matchString(". . .") {
					goto l1391
				}
			}
		l1392:
			do(90)
			return true
		l1391:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 227 Dash <- (EmDash / EnDash) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1395, thunkPosition1395 := position, thunkPosition
				if !p.rules[ruleEmDash]() {
					goto l1396
				}
				goto l1395
			l1396:
				position, thunkPosition = position1395, thunkPosition1395
				if !p.rules[ruleEnDash]() {
					goto l1394
				}
			}
		l1395:
			return true
		l1394:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 228 EnDash <- ('-' &Digit { yy = mk_element(ENDASH) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('-') {
				goto l1397
			}
			{
				position1398, thunkPosition1398 := position, thunkPosition
				if !p.rules[ruleDigit]() {
					goto l1397
				}
				position, thunkPosition = position1398, thunkPosition1398
			}
			do(91)
			return true
		l1397:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 229 EmDash <- (('---' / '--') { yy = mk_element(EMDASH) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1400, thunkPosition1400 := position, thunkPosition
				if !matchString("---") {
					goto l1401
				}
				goto l1400
			l1401:
				position, thunkPosition = position1400, thunkPosition1400
				if !matchString("--") {
					goto l1399
				}
			}
		l1400:
			do(92)
			return true
		l1399:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 230 SingleQuoteStart <- ('\'' ![)!\],.;:-? \t\n] !(((&[r] 're') | (&[l] 'll') | (&[v] 've') | (&[m] 'm') | (&[t] 't') | (&[s] 's')) !Alphanumeric)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1402
			}
			{
				position1403, thunkPosition1403 := position, thunkPosition
				if !matchClass(6) {
					goto l1403
				}
				goto l1402
			l1403:
				position, thunkPosition = position1403, thunkPosition1403
			}
			{
				position1404, thunkPosition1404 := position, thunkPosition
				{
					if position == len(p.Buffer) {
						goto l1404
					}
					switch p.Buffer[position] {
					case 'r':
						if !matchString("re") {
							goto l1404
						}
					case 'l':
						if !matchString("ll") {
							goto l1404
						}
					case 'v':
						if !matchString("ve") {
							goto l1404
						}
					case 'm':
						if !matchChar('m') {
							goto l1404
						}
					case 't':
						if !matchChar('t') {
							goto l1404
						}
					default:
						if !matchChar('s') {
							goto l1404
						}
					}
				}
				{
					position1406, thunkPosition1406 := position, thunkPosition
					if !p.rules[ruleAlphanumeric]() {
						goto l1406
					}
					goto l1404
				l1406:
					position, thunkPosition = position1406, thunkPosition1406
				}
				goto l1402
			l1404:
				position, thunkPosition = position1404, thunkPosition1404
			}
			return true
		l1402:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 231 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1407
			}
			{
				position1408, thunkPosition1408 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l1408
				}
				goto l1407
			l1408:
				position, thunkPosition = position1408, thunkPosition1408
			}
			return true
		l1407:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 232 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = mk_list(SINGLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto l1409
			}
			if !p.rules[ruleStartList]() {
				goto l1409
			}
			doarg(yySet, -1)
			{
				position1412, thunkPosition1412 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto l1412
				}
				goto l1409
			l1412:
				position, thunkPosition = position1412, thunkPosition1412
			}
			if !p.rules[ruleInline]() {
				goto l1409
			}
			doarg(yySet, -2)
			do(93)
		l1410:
			{
				position1411, thunkPosition1411 := position, thunkPosition
				{
					position1413, thunkPosition1413 := position, thunkPosition
					if !p.rules[ruleSingleQuoteEnd]() {
						goto l1413
					}
					goto l1411
				l1413:
					position, thunkPosition = position1413, thunkPosition1413
				}
				if !p.rules[ruleInline]() {
					goto l1411
				}
				doarg(yySet, -2)
				do(93)
				goto l1410
			l1411:
				position, thunkPosition = position1411, thunkPosition1411
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1409
			}
			do(94)
			doarg(yyPop, 2)
			return true
		l1409:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 233 DoubleQuoteStart <- '"' */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l1414
			}
			return true
		l1414:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 234 DoubleQuoteEnd <- '"' */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l1415
			}
			return true
		l1415:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 235 DoubleQuoted <- (DoubleQuoteStart StartList (!DoubleQuoteEnd Inline { a = cons(b, a) })+ DoubleQuoteEnd { yy = mk_list(DOUBLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleDoubleQuoteStart]() {
				goto l1416
			}
			if !p.rules[ruleStartList]() {
				goto l1416
			}
			doarg(yySet, -1)
			{
				position1419, thunkPosition1419 := position, thunkPosition
				if !p.rules[ruleDoubleQuoteEnd]() {
					goto l1419
				}
				goto l1416
			l1419:
				position, thunkPosition = position1419, thunkPosition1419
			}
			if !p.rules[ruleInline]() {
				goto l1416
			}
			doarg(yySet, -2)
			do(95)
		l1417:
			{
				position1418, thunkPosition1418 := position, thunkPosition
				{
					position1420, thunkPosition1420 := position, thunkPosition
					if !p.rules[ruleDoubleQuoteEnd]() {
						goto l1420
					}
					goto l1418
				l1420:
					position, thunkPosition = position1420, thunkPosition1420
				}
				if !p.rules[ruleInline]() {
					goto l1418
				}
				doarg(yySet, -2)
				do(95)
				goto l1417
			l1418:
				position, thunkPosition = position1418, thunkPosition1418
			}
			if !p.rules[ruleDoubleQuoteEnd]() {
				goto l1416
			}
			do(96)
			doarg(yyPop, 2)
			return true
		l1416:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 236 NoteReference <- (&{ p.extension.Notes } RawNoteReference {
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
				goto l1421
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1421
			}
			doarg(yySet, -1)
			do(97)
			doarg(yyPop, 1)
			return true
		l1421:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 237 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("[^") {
				goto l1422
			}
			begin = position
			{
				position1425, thunkPosition1425 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1425
				}
				goto l1422
			l1425:
				position, thunkPosition = position1425, thunkPosition1425
			}
			if peekChar(']') {
				goto l1422
			}
			if !matchDot() {
				goto l1422
			}
		l1423:
			{
				position1424, thunkPosition1424 := position, thunkPosition
				{
					position1426, thunkPosition1426 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l1426
					}
					goto l1424
				l1426:
					position, thunkPosition = position1426, thunkPosition1426
				}
				if peekChar(']') {
					goto l1424
				}
				if !matchDot() {
					goto l1424
				}
				goto l1423
			l1424:
				position, thunkPosition = position1424, thunkPosition1424
			}
			end = position
			if !matchChar(']') {
				goto l1422
			}
			do(98)
			return true
		l1422:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 238 Note <- (&{ p.extension.Notes } NonindentSpace RawNoteReference ':' Sp StartList (RawNoteBlock { a = cons(yy, a) }) (&Indent RawNoteBlock { a = cons(yy, a) })* {   yy = mk_list(NOTE, a)
                    yy.contents.str = ref.contents.str
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !( p.extension.Notes ) {
				goto l1427
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l1427
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1427
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto l1427
			}
			if !p.rules[ruleSp]() {
				goto l1427
			}
			if !p.rules[ruleStartList]() {
				goto l1427
			}
			doarg(yySet, -2)
			if !p.rules[ruleRawNoteBlock]() {
				goto l1427
			}
			do(99)
		l1428:
			{
				position1429, thunkPosition1429 := position, thunkPosition
				{
					position1430, thunkPosition1430 := position, thunkPosition
					if !p.rules[ruleIndent]() {
						goto l1429
					}
					position, thunkPosition = position1430, thunkPosition1430
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto l1429
				}
				do(100)
				goto l1428
			l1429:
				position, thunkPosition = position1429, thunkPosition1429
			}
			do(101)
			doarg(yyPop, 2)
			return true
		l1427:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 239 InlineNote <- (&{ p.extension.Notes } '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = mk_list(NOTE, a)
                  yy.contents.str = "" }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !( p.extension.Notes ) {
				goto l1431
			}
			if !matchString("^[") {
				goto l1431
			}
			if !p.rules[ruleStartList]() {
				goto l1431
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto l1431
			}
			if !p.rules[ruleInline]() {
				goto l1431
			}
			do(102)
		l1432:
			{
				position1433, thunkPosition1433 := position, thunkPosition
				if peekChar(']') {
					goto l1433
				}
				if !p.rules[ruleInline]() {
					goto l1433
				}
				do(102)
				goto l1432
			l1433:
				position, thunkPosition = position1433, thunkPosition1433
			}
			if !matchChar(']') {
				goto l1431
			}
			do(103)
			doarg(yyPop, 1)
			return true
		l1431:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 240 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1434
			}
			doarg(yySet, -1)
		l1435:
			{
				position1436, thunkPosition1436 := position, thunkPosition
				{
					position1437, thunkPosition1437 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto l1438
					}
					doarg(yySet, -2)
					do(104)
					goto l1437
				l1438:
					position, thunkPosition = position1437, thunkPosition1437
					if !p.rules[ruleSkipBlock]() {
						goto l1436
					}
				}
			l1437:
				goto l1435
			l1436:
				position, thunkPosition = position1436, thunkPosition1436
			}
			do(105)
			if !(commit(thunkPosition0)) {
				goto l1434
			}
			doarg(yyPop, 2)
			return true
		l1434:
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
				goto l1439
			}
			doarg(yySet, -1)
			{
				position1442, thunkPosition1442 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l1442
				}
				goto l1439
			l1442:
				position, thunkPosition = position1442, thunkPosition1442
			}
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l1439
			}
			do(106)
		l1440:
			{
				position1441, thunkPosition1441 := position, thunkPosition
				{
					position1443, thunkPosition1443 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1443
					}
					goto l1441
				l1443:
					position, thunkPosition = position1443, thunkPosition1443
				}
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto l1441
				}
				do(106)
				goto l1440
			l1441:
				position, thunkPosition = position1441, thunkPosition1441
			}
			begin = position
		l1444:
			{
				position1445, thunkPosition1445 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l1445
				}
				goto l1444
			l1445:
				position, thunkPosition = position1445, thunkPosition1445
			}
			end = position
			do(107)
			do(108)
			doarg(yyPop, 1)
			return true
		l1439:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 242 DefinitionList <- (&{ p.extension.Dlists } StartList (Definition { a = cons(yy, a) })+ { yy = mk_list(DEFINITIONLIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !( p.extension.Dlists ) {
				goto l1446
			}
			if !p.rules[ruleStartList]() {
				goto l1446
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto l1446
			}
			do(109)
		l1447:
			{
				position1448, thunkPosition1448 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto l1448
				}
				do(109)
				goto l1447
			l1448:
				position, thunkPosition = position1448, thunkPosition1448
			}
			do(110)
			doarg(yyPop, 1)
			return true
		l1446:
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
				position1450, thunkPosition1450 := position, thunkPosition
				{
					position1453, thunkPosition1453 := position, thunkPosition
					if !p.rules[ruleDefmark]() {
						goto l1453
					}
					goto l1449
				l1453:
					position, thunkPosition = position1453, thunkPosition1453
				}
				if !p.rules[ruleRawLine]() {
					goto l1449
				}
			l1451:
				{
					position1452, thunkPosition1452 := position, thunkPosition
					{
						position1454, thunkPosition1454 := position, thunkPosition
						if !p.rules[ruleDefmark]() {
							goto l1454
						}
						goto l1452
					l1454:
						position, thunkPosition = position1454, thunkPosition1454
					}
					if !p.rules[ruleRawLine]() {
						goto l1452
					}
					goto l1451
				l1452:
					position, thunkPosition = position1452, thunkPosition1452
				}
				{
					position1455, thunkPosition1455 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1455
					}
					goto l1456
				l1455:
					position, thunkPosition = position1455, thunkPosition1455
				}
			l1456:
				if !p.rules[ruleDefmark]() {
					goto l1449
				}
				position, thunkPosition = position1450, thunkPosition1450
			}
			if !p.rules[ruleStartList]() {
				goto l1449
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto l1449
			}
			do(111)
		l1457:
			{
				position1458, thunkPosition1458 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto l1458
				}
				do(111)
				goto l1457
			l1458:
				position, thunkPosition = position1458, thunkPosition1458
			}
			{
				position1459, thunkPosition1459 := position, thunkPosition
				if !p.rules[ruleDefTight]() {
					goto l1460
				}
				goto l1459
			l1460:
				position, thunkPosition = position1459, thunkPosition1459
				if !p.rules[ruleDefLoose]() {
					goto l1449
				}
			}
		l1459:
			do(112)
			do(113)
			doarg(yyPop, 1)
			return true
		l1449:
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
				goto l1461
			}
			{
				position1462, thunkPosition1462 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1462
				}
				goto l1461
			l1462:
				position, thunkPosition = position1462, thunkPosition1462
			}
			{
				position1463, thunkPosition1463 := position, thunkPosition
				if !p.rules[ruleNonspacechar]() {
					goto l1461
				}
				position, thunkPosition = position1463, thunkPosition1463
			}
			if !p.rules[ruleStartList]() {
				goto l1461
			}
			doarg(yySet, -1)
			{
				position1466, thunkPosition1466 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l1466
				}
				goto l1461
			l1466:
				position, thunkPosition = position1466, thunkPosition1466
			}
			if !p.rules[ruleInline]() {
				goto l1461
			}
			do(114)
		l1464:
			{
				position1465, thunkPosition1465 := position, thunkPosition
				{
					position1467, thunkPosition1467 := position, thunkPosition
					if !p.rules[ruleEndline]() {
						goto l1467
					}
					goto l1465
				l1467:
					position, thunkPosition = position1467, thunkPosition1467
				}
				if !p.rules[ruleInline]() {
					goto l1465
				}
				do(114)
				goto l1464
			l1465:
				position, thunkPosition = position1465, thunkPosition1465
			}
			if !p.rules[ruleSp]() {
				goto l1461
			}
			if !p.rules[ruleNewline]() {
				goto l1461
			}
			do(115)
			doarg(yyPop, 1)
			return true
		l1461:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 245 DefTight <- (&Defmark ListTight) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1469, thunkPosition1469 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1468
				}
				position, thunkPosition = position1469, thunkPosition1469
			}
			if !p.rules[ruleListTight]() {
				goto l1468
			}
			return true
		l1468:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 246 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlankLine]() {
				goto l1470
			}
			{
				position1471, thunkPosition1471 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1470
				}
				position, thunkPosition = position1471, thunkPosition1471
			}
			if !p.rules[ruleListLoose]() {
				goto l1470
			}
			return true
		l1470:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 247 Defmark <- (NonindentSpace (':' / '~') Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l1472
			}
			{
				position1473, thunkPosition1473 := position, thunkPosition
				if !matchChar(':') {
					goto l1474
				}
				goto l1473
			l1474:
				position, thunkPosition = position1473, thunkPosition1473
				if !matchChar('~') {
					goto l1472
				}
			}
		l1473:
			if !p.rules[ruleSpacechar]() {
				goto l1472
			}
		l1475:
			{
				position1476, thunkPosition1476 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1476
				}
				goto l1475
			l1476:
				position, thunkPosition = position1476, thunkPosition1476
			}
			return true
		l1472:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 248 DefMarker <- (&{ p.extension.Dlists } Defmark) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !( p.extension.Dlists ) {
				goto l1477
			}
			if !p.rules[ruleDefmark]() {
				goto l1477
			}
			return true
		l1477:
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
