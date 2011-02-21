
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
		/* 18 HorizontalRule <- (NonindentSpace (('*' Sp '*' Sp '*' (Sp '*')*) / ('-' Sp '-' Sp '-' (Sp '-')*) / ('_' Sp '_' Sp '_' (Sp '_')*)) Sp Newline BlankLine+ { yy = mk_element(HRULE) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l100
			}
			{
				position101, thunkPosition101 := position, thunkPosition
				if !matchChar('*') {
					goto l102
				}
				if !p.rules[ruleSp]() {
					goto l102
				}
				if !matchChar('*') {
					goto l102
				}
				if !p.rules[ruleSp]() {
					goto l102
				}
				if !matchChar('*') {
					goto l102
				}
			l103:
				{
					position104, thunkPosition104 := position, thunkPosition
					if !p.rules[ruleSp]() {
						goto l104
					}
					if !matchChar('*') {
						goto l104
					}
					goto l103
				l104:
					position, thunkPosition = position104, thunkPosition104
				}
				goto l101
			l102:
				position, thunkPosition = position101, thunkPosition101
				if !matchChar('-') {
					goto l105
				}
				if !p.rules[ruleSp]() {
					goto l105
				}
				if !matchChar('-') {
					goto l105
				}
				if !p.rules[ruleSp]() {
					goto l105
				}
				if !matchChar('-') {
					goto l105
				}
			l106:
				{
					position107, thunkPosition107 := position, thunkPosition
					if !p.rules[ruleSp]() {
						goto l107
					}
					if !matchChar('-') {
						goto l107
					}
					goto l106
				l107:
					position, thunkPosition = position107, thunkPosition107
				}
				goto l101
			l105:
				position, thunkPosition = position101, thunkPosition101
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
			l108:
				{
					position109, thunkPosition109 := position, thunkPosition
					if !p.rules[ruleSp]() {
						goto l109
					}
					if !matchChar('_') {
						goto l109
					}
					goto l108
				l109:
					position, thunkPosition = position109, thunkPosition109
				}
			}
		l101:
			if !p.rules[ruleSp]() {
				goto l100
			}
			if !p.rules[ruleNewline]() {
				goto l100
			}
			if !p.rules[ruleBlankLine]() {
				goto l100
			}
		l110:
			{
				position111, thunkPosition111 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l111
				}
				goto l110
			l111:
				position, thunkPosition = position111, thunkPosition111
			}
			do(21)
			return true
		l100:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 19 Bullet <- (!HorizontalRule NonindentSpace ('+' / '*' / '-') Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position113, thunkPosition113 := position, thunkPosition
				if !p.rules[ruleHorizontalRule]() {
					goto l113
				}
				goto l112
			l113:
				position, thunkPosition = position113, thunkPosition113
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l112
			}
			{
				position114, thunkPosition114 := position, thunkPosition
				if !matchChar('+') {
					goto l115
				}
				goto l114
			l115:
				position, thunkPosition = position114, thunkPosition114
				if !matchChar('*') {
					goto l116
				}
				goto l114
			l116:
				position, thunkPosition = position114, thunkPosition114
				if !matchChar('-') {
					goto l112
				}
			}
		l114:
			if !p.rules[ruleSpacechar]() {
				goto l112
			}
		l117:
			{
				position118, thunkPosition118 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l118
				}
				goto l117
			l118:
				position, thunkPosition = position118, thunkPosition118
			}
			return true
		l112:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 20 BulletList <- (&Bullet (ListTight / ListLoose) { yy.key = BULLETLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position120, thunkPosition120 := position, thunkPosition
				if !p.rules[ruleBullet]() {
					goto l119
				}
				position, thunkPosition = position120, thunkPosition120
			}
			{
				position121, thunkPosition121 := position, thunkPosition
				if !p.rules[ruleListTight]() {
					goto l122
				}
				goto l121
			l122:
				position, thunkPosition = position121, thunkPosition121
				if !p.rules[ruleListLoose]() {
					goto l119
				}
			}
		l121:
			do(22)
			return true
		l119:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 21 ListTight <- (StartList (ListItemTight { a = cons(yy, a) })+ BlankLine* !(Bullet / Enumerator / DefMarker) { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l123
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItemTight]() {
				goto l123
			}
			do(23)
		l124:
			{
				position125, thunkPosition125 := position, thunkPosition
				if !p.rules[ruleListItemTight]() {
					goto l125
				}
				do(23)
				goto l124
			l125:
				position, thunkPosition = position125, thunkPosition125
			}
		l126:
			{
				position127, thunkPosition127 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l127
				}
				goto l126
			l127:
				position, thunkPosition = position127, thunkPosition127
			}
			{
				position128, thunkPosition128 := position, thunkPosition
				{
					position129, thunkPosition129 := position, thunkPosition
					if !p.rules[ruleBullet]() {
						goto l130
					}
					goto l129
				l130:
					position, thunkPosition = position129, thunkPosition129
					if !p.rules[ruleEnumerator]() {
						goto l131
					}
					goto l129
				l131:
					position, thunkPosition = position129, thunkPosition129
					if !p.rules[ruleDefMarker]() {
						goto l128
					}
				}
			l129:
				goto l123
			l128:
				position, thunkPosition = position128, thunkPosition128
			}
			do(24)
			doarg(yyPop, 1)
			return true
		l123:
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
				goto l132
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItem]() {
				goto l132
			}
			doarg(yySet, -2)
		l135:
			{
				position136, thunkPosition136 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l136
				}
				goto l135
			l136:
				position, thunkPosition = position136, thunkPosition136
			}
			do(25)
		l133:
			{
				position134, thunkPosition134 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto l134
				}
				doarg(yySet, -2)
			l137:
				{
					position138, thunkPosition138 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l138
					}
					goto l137
				l138:
					position, thunkPosition = position138, thunkPosition138
				}
				do(25)
				goto l133
			l134:
				position, thunkPosition = position134, thunkPosition134
			}
			do(26)
			doarg(yyPop, 2)
			return true
		l132:
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
				position140, thunkPosition140 := position, thunkPosition
				if !p.rules[ruleBullet]() {
					goto l141
				}
				goto l140
			l141:
				position, thunkPosition = position140, thunkPosition140
				if !p.rules[ruleEnumerator]() {
					goto l142
				}
				goto l140
			l142:
				position, thunkPosition = position140, thunkPosition140
				if !p.rules[ruleDefMarker]() {
					goto l139
				}
			}
		l140:
			if !p.rules[ruleStartList]() {
				goto l139
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l139
			}
			do(27)
		l143:
			{
				position144, thunkPosition144 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l144
				}
				do(28)
				goto l143
			l144:
				position, thunkPosition = position144, thunkPosition144
			}
			do(29)
			doarg(yyPop, 1)
			return true
		l139:
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
				position146, thunkPosition146 := position, thunkPosition
				if !p.rules[ruleBullet]() {
					goto l147
				}
				goto l146
			l147:
				position, thunkPosition = position146, thunkPosition146
				if !p.rules[ruleEnumerator]() {
					goto l148
				}
				goto l146
			l148:
				position, thunkPosition = position146, thunkPosition146
				if !p.rules[ruleDefMarker]() {
					goto l145
				}
			}
		l146:
			if !p.rules[ruleStartList]() {
				goto l145
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l145
			}
			do(30)
		l149:
			{
				position150, thunkPosition150 := position, thunkPosition
				{
					position151, thunkPosition151 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l151
					}
					goto l150
				l151:
					position, thunkPosition = position151, thunkPosition151
				}
				if !p.rules[ruleListContinuationBlock]() {
					goto l150
				}
				do(31)
				goto l149
			l150:
				position, thunkPosition = position150, thunkPosition150
			}
			{
				position152, thunkPosition152 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l152
				}
				goto l145
			l152:
				position, thunkPosition = position152, thunkPosition152
			}
			do(32)
			doarg(yyPop, 1)
			return true
		l145:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 25 ListBlock <- (StartList !BlankLine Line { a = cons(yy, a) } (ListBlockLine { a = cons(yy, a) })* { yy = mk_str_from_list(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l153
			}
			doarg(yySet, -1)
			{
				position154, thunkPosition154 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l154
				}
				goto l153
			l154:
				position, thunkPosition = position154, thunkPosition154
			}
			if !p.rules[ruleLine]() {
				goto l153
			}
			do(33)
		l155:
			{
				position156, thunkPosition156 := position, thunkPosition
				if !p.rules[ruleListBlockLine]() {
					goto l156
				}
				do(34)
				goto l155
			l156:
				position, thunkPosition = position156, thunkPosition156
			}
			do(35)
			doarg(yyPop, 1)
			return true
		l153:
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
				goto l157
			}
			doarg(yySet, -1)
			begin = position
		l158:
			{
				position159, thunkPosition159 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l159
				}
				goto l158
			l159:
				position, thunkPosition = position159, thunkPosition159
			}
			end = position
			do(36)
			if !p.rules[ruleIndent]() {
				goto l157
			}
			if !p.rules[ruleListBlock]() {
				goto l157
			}
			do(37)
		l160:
			{
				position161, thunkPosition161 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l161
				}
				if !p.rules[ruleListBlock]() {
					goto l161
				}
				do(37)
				goto l160
			l161:
				position, thunkPosition = position161, thunkPosition161
			}
			do(38)
			doarg(yyPop, 1)
			return true
		l157:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 27 Enumerator <- (NonindentSpace [0-9]+ '.' Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l162
			}
			if !matchClass(7) {
				goto l162
			}
		l163:
			{
				position164, thunkPosition164 := position, thunkPosition
				if !matchClass(7) {
					goto l164
				}
				goto l163
			l164:
				position, thunkPosition = position164, thunkPosition164
			}
			if !matchChar('.') {
				goto l162
			}
			if !p.rules[ruleSpacechar]() {
				goto l162
			}
		l165:
			{
				position166, thunkPosition166 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l166
				}
				goto l165
			l166:
				position, thunkPosition = position166, thunkPosition166
			}
			return true
		l162:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 28 OrderedList <- (&Enumerator (ListTight / ListLoose) { yy.key = ORDEREDLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position168, thunkPosition168 := position, thunkPosition
				if !p.rules[ruleEnumerator]() {
					goto l167
				}
				position, thunkPosition = position168, thunkPosition168
			}
			{
				position169, thunkPosition169 := position, thunkPosition
				if !p.rules[ruleListTight]() {
					goto l170
				}
				goto l169
			l170:
				position, thunkPosition = position169, thunkPosition169
				if !p.rules[ruleListLoose]() {
					goto l167
				}
			}
		l169:
			do(39)
			return true
		l167:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 29 ListBlockLine <- (!BlankLine !((Indent? (Bullet / Enumerator)) / DefMarker) !HorizontalRule OptionallyIndentedLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position172, thunkPosition172 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l172
				}
				goto l171
			l172:
				position, thunkPosition = position172, thunkPosition172
			}
			{
				position173, thunkPosition173 := position, thunkPosition
				{
					position174, thunkPosition174 := position, thunkPosition
					{
						position176, thunkPosition176 := position, thunkPosition
						if !p.rules[ruleIndent]() {
							goto l176
						}
						goto l177
					l176:
						position, thunkPosition = position176, thunkPosition176
					}
				l177:
					{
						position178, thunkPosition178 := position, thunkPosition
						if !p.rules[ruleBullet]() {
							goto l179
						}
						goto l178
					l179:
						position, thunkPosition = position178, thunkPosition178
						if !p.rules[ruleEnumerator]() {
							goto l175
						}
					}
				l178:
					goto l174
				l175:
					position, thunkPosition = position174, thunkPosition174
					if !p.rules[ruleDefMarker]() {
						goto l173
					}
				}
			l174:
				goto l171
			l173:
				position, thunkPosition = position173, thunkPosition173
			}
			{
				position180, thunkPosition180 := position, thunkPosition
				if !p.rules[ruleHorizontalRule]() {
					goto l180
				}
				goto l171
			l180:
				position, thunkPosition = position180, thunkPosition180
			}
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l171
			}
			return true
		l171:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 HtmlBlockOpenAddress <- ('<' Spnl ('address' / 'ADDRESS') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l181
			}
			if !p.rules[ruleSpnl]() {
				goto l181
			}
			{
				position182, thunkPosition182 := position, thunkPosition
				if !matchString("address") {
					goto l183
				}
				goto l182
			l183:
				position, thunkPosition = position182, thunkPosition182
				if !matchString("ADDRESS") {
					goto l181
				}
			}
		l182:
			if !p.rules[ruleSpnl]() {
				goto l181
			}
		l184:
			{
				position185, thunkPosition185 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l185
				}
				goto l184
			l185:
				position, thunkPosition = position185, thunkPosition185
			}
			if !matchChar('>') {
				goto l181
			}
			return true
		l181:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 31 HtmlBlockCloseAddress <- ('<' Spnl '/' ('address' / 'ADDRESS') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position187, thunkPosition187 := position, thunkPosition
				if !matchString("address") {
					goto l188
				}
				goto l187
			l188:
				position, thunkPosition = position187, thunkPosition187
				if !matchString("ADDRESS") {
					goto l186
				}
			}
		l187:
			if !p.rules[ruleSpnl]() {
				goto l186
			}
			if !matchChar('>') {
				goto l186
			}
			return true
		l186:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 32 HtmlBlockAddress <- (HtmlBlockOpenAddress (HtmlBlockAddress / (!HtmlBlockCloseAddress .))* HtmlBlockCloseAddress) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenAddress]() {
				goto l189
			}
		l190:
			{
				position191, thunkPosition191 := position, thunkPosition
				{
					position192, thunkPosition192 := position, thunkPosition
					if !p.rules[ruleHtmlBlockAddress]() {
						goto l193
					}
					goto l192
				l193:
					position, thunkPosition = position192, thunkPosition192
					{
						position194, thunkPosition194 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseAddress]() {
							goto l194
						}
						goto l191
					l194:
						position, thunkPosition = position194, thunkPosition194
					}
					if !matchDot() {
						goto l191
					}
				}
			l192:
				goto l190
			l191:
				position, thunkPosition = position191, thunkPosition191
			}
			if !p.rules[ruleHtmlBlockCloseAddress]() {
				goto l189
			}
			return true
		l189:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 33 HtmlBlockOpenBlockquote <- ('<' Spnl ('blockquote' / 'BLOCKQUOTE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l195
			}
			if !p.rules[ruleSpnl]() {
				goto l195
			}
			{
				position196, thunkPosition196 := position, thunkPosition
				if !matchString("blockquote") {
					goto l197
				}
				goto l196
			l197:
				position, thunkPosition = position196, thunkPosition196
				if !matchString("BLOCKQUOTE") {
					goto l195
				}
			}
		l196:
			if !p.rules[ruleSpnl]() {
				goto l195
			}
		l198:
			{
				position199, thunkPosition199 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l199
				}
				goto l198
			l199:
				position, thunkPosition = position199, thunkPosition199
			}
			if !matchChar('>') {
				goto l195
			}
			return true
		l195:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 34 HtmlBlockCloseBlockquote <- ('<' Spnl '/' ('blockquote' / 'BLOCKQUOTE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l200
			}
			if !p.rules[ruleSpnl]() {
				goto l200
			}
			if !matchChar('/') {
				goto l200
			}
			{
				position201, thunkPosition201 := position, thunkPosition
				if !matchString("blockquote") {
					goto l202
				}
				goto l201
			l202:
				position, thunkPosition = position201, thunkPosition201
				if !matchString("BLOCKQUOTE") {
					goto l200
				}
			}
		l201:
			if !p.rules[ruleSpnl]() {
				goto l200
			}
			if !matchChar('>') {
				goto l200
			}
			return true
		l200:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 35 HtmlBlockBlockquote <- (HtmlBlockOpenBlockquote (HtmlBlockBlockquote / (!HtmlBlockCloseBlockquote .))* HtmlBlockCloseBlockquote) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenBlockquote]() {
				goto l203
			}
		l204:
			{
				position205, thunkPosition205 := position, thunkPosition
				{
					position206, thunkPosition206 := position, thunkPosition
					if !p.rules[ruleHtmlBlockBlockquote]() {
						goto l207
					}
					goto l206
				l207:
					position, thunkPosition = position206, thunkPosition206
					{
						position208, thunkPosition208 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseBlockquote]() {
							goto l208
						}
						goto l205
					l208:
						position, thunkPosition = position208, thunkPosition208
					}
					if !matchDot() {
						goto l205
					}
				}
			l206:
				goto l204
			l205:
				position, thunkPosition = position205, thunkPosition205
			}
			if !p.rules[ruleHtmlBlockCloseBlockquote]() {
				goto l203
			}
			return true
		l203:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 36 HtmlBlockOpenCenter <- ('<' Spnl ('center' / 'CENTER') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l209
			}
			if !p.rules[ruleSpnl]() {
				goto l209
			}
			{
				position210, thunkPosition210 := position, thunkPosition
				if !matchString("center") {
					goto l211
				}
				goto l210
			l211:
				position, thunkPosition = position210, thunkPosition210
				if !matchString("CENTER") {
					goto l209
				}
			}
		l210:
			if !p.rules[ruleSpnl]() {
				goto l209
			}
		l212:
			{
				position213, thunkPosition213 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l213
				}
				goto l212
			l213:
				position, thunkPosition = position213, thunkPosition213
			}
			if !matchChar('>') {
				goto l209
			}
			return true
		l209:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 37 HtmlBlockCloseCenter <- ('<' Spnl '/' ('center' / 'CENTER') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l214
			}
			if !p.rules[ruleSpnl]() {
				goto l214
			}
			if !matchChar('/') {
				goto l214
			}
			{
				position215, thunkPosition215 := position, thunkPosition
				if !matchString("center") {
					goto l216
				}
				goto l215
			l216:
				position, thunkPosition = position215, thunkPosition215
				if !matchString("CENTER") {
					goto l214
				}
			}
		l215:
			if !p.rules[ruleSpnl]() {
				goto l214
			}
			if !matchChar('>') {
				goto l214
			}
			return true
		l214:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 38 HtmlBlockCenter <- (HtmlBlockOpenCenter (HtmlBlockCenter / (!HtmlBlockCloseCenter .))* HtmlBlockCloseCenter) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenCenter]() {
				goto l217
			}
		l218:
			{
				position219, thunkPosition219 := position, thunkPosition
				{
					position220, thunkPosition220 := position, thunkPosition
					if !p.rules[ruleHtmlBlockCenter]() {
						goto l221
					}
					goto l220
				l221:
					position, thunkPosition = position220, thunkPosition220
					{
						position222, thunkPosition222 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseCenter]() {
							goto l222
						}
						goto l219
					l222:
						position, thunkPosition = position222, thunkPosition222
					}
					if !matchDot() {
						goto l219
					}
				}
			l220:
				goto l218
			l219:
				position, thunkPosition = position219, thunkPosition219
			}
			if !p.rules[ruleHtmlBlockCloseCenter]() {
				goto l217
			}
			return true
		l217:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 39 HtmlBlockOpenDir <- ('<' Spnl ('dir' / 'DIR') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l223
			}
			if !p.rules[ruleSpnl]() {
				goto l223
			}
			{
				position224, thunkPosition224 := position, thunkPosition
				if !matchString("dir") {
					goto l225
				}
				goto l224
			l225:
				position, thunkPosition = position224, thunkPosition224
				if !matchString("DIR") {
					goto l223
				}
			}
		l224:
			if !p.rules[ruleSpnl]() {
				goto l223
			}
		l226:
			{
				position227, thunkPosition227 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l227
				}
				goto l226
			l227:
				position, thunkPosition = position227, thunkPosition227
			}
			if !matchChar('>') {
				goto l223
			}
			return true
		l223:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 40 HtmlBlockCloseDir <- ('<' Spnl '/' ('dir' / 'DIR') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l228
			}
			if !p.rules[ruleSpnl]() {
				goto l228
			}
			if !matchChar('/') {
				goto l228
			}
			{
				position229, thunkPosition229 := position, thunkPosition
				if !matchString("dir") {
					goto l230
				}
				goto l229
			l230:
				position, thunkPosition = position229, thunkPosition229
				if !matchString("DIR") {
					goto l228
				}
			}
		l229:
			if !p.rules[ruleSpnl]() {
				goto l228
			}
			if !matchChar('>') {
				goto l228
			}
			return true
		l228:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 41 HtmlBlockDir <- (HtmlBlockOpenDir (HtmlBlockDir / (!HtmlBlockCloseDir .))* HtmlBlockCloseDir) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDir]() {
				goto l231
			}
		l232:
			{
				position233, thunkPosition233 := position, thunkPosition
				{
					position234, thunkPosition234 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDir]() {
						goto l235
					}
					goto l234
				l235:
					position, thunkPosition = position234, thunkPosition234
					{
						position236, thunkPosition236 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDir]() {
							goto l236
						}
						goto l233
					l236:
						position, thunkPosition = position236, thunkPosition236
					}
					if !matchDot() {
						goto l233
					}
				}
			l234:
				goto l232
			l233:
				position, thunkPosition = position233, thunkPosition233
			}
			if !p.rules[ruleHtmlBlockCloseDir]() {
				goto l231
			}
			return true
		l231:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 42 HtmlBlockOpenDiv <- ('<' Spnl ('div' / 'DIV') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l237
			}
			if !p.rules[ruleSpnl]() {
				goto l237
			}
			{
				position238, thunkPosition238 := position, thunkPosition
				if !matchString("div") {
					goto l239
				}
				goto l238
			l239:
				position, thunkPosition = position238, thunkPosition238
				if !matchString("DIV") {
					goto l237
				}
			}
		l238:
			if !p.rules[ruleSpnl]() {
				goto l237
			}
		l240:
			{
				position241, thunkPosition241 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l241
				}
				goto l240
			l241:
				position, thunkPosition = position241, thunkPosition241
			}
			if !matchChar('>') {
				goto l237
			}
			return true
		l237:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 43 HtmlBlockCloseDiv <- ('<' Spnl '/' ('div' / 'DIV') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position243, thunkPosition243 := position, thunkPosition
				if !matchString("div") {
					goto l244
				}
				goto l243
			l244:
				position, thunkPosition = position243, thunkPosition243
				if !matchString("DIV") {
					goto l242
				}
			}
		l243:
			if !p.rules[ruleSpnl]() {
				goto l242
			}
			if !matchChar('>') {
				goto l242
			}
			return true
		l242:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 44 HtmlBlockDiv <- (HtmlBlockOpenDiv (HtmlBlockDiv / (!HtmlBlockCloseDiv .))* HtmlBlockCloseDiv) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDiv]() {
				goto l245
			}
		l246:
			{
				position247, thunkPosition247 := position, thunkPosition
				{
					position248, thunkPosition248 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDiv]() {
						goto l249
					}
					goto l248
				l249:
					position, thunkPosition = position248, thunkPosition248
					{
						position250, thunkPosition250 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDiv]() {
							goto l250
						}
						goto l247
					l250:
						position, thunkPosition = position250, thunkPosition250
					}
					if !matchDot() {
						goto l247
					}
				}
			l248:
				goto l246
			l247:
				position, thunkPosition = position247, thunkPosition247
			}
			if !p.rules[ruleHtmlBlockCloseDiv]() {
				goto l245
			}
			return true
		l245:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 45 HtmlBlockOpenDl <- ('<' Spnl ('dl' / 'DL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l251
			}
			if !p.rules[ruleSpnl]() {
				goto l251
			}
			{
				position252, thunkPosition252 := position, thunkPosition
				if !matchString("dl") {
					goto l253
				}
				goto l252
			l253:
				position, thunkPosition = position252, thunkPosition252
				if !matchString("DL") {
					goto l251
				}
			}
		l252:
			if !p.rules[ruleSpnl]() {
				goto l251
			}
		l254:
			{
				position255, thunkPosition255 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l255
				}
				goto l254
			l255:
				position, thunkPosition = position255, thunkPosition255
			}
			if !matchChar('>') {
				goto l251
			}
			return true
		l251:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 46 HtmlBlockCloseDl <- ('<' Spnl '/' ('dl' / 'DL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l256
			}
			if !p.rules[ruleSpnl]() {
				goto l256
			}
			if !matchChar('/') {
				goto l256
			}
			{
				position257, thunkPosition257 := position, thunkPosition
				if !matchString("dl") {
					goto l258
				}
				goto l257
			l258:
				position, thunkPosition = position257, thunkPosition257
				if !matchString("DL") {
					goto l256
				}
			}
		l257:
			if !p.rules[ruleSpnl]() {
				goto l256
			}
			if !matchChar('>') {
				goto l256
			}
			return true
		l256:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 47 HtmlBlockDl <- (HtmlBlockOpenDl (HtmlBlockDl / (!HtmlBlockCloseDl .))* HtmlBlockCloseDl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDl]() {
				goto l259
			}
		l260:
			{
				position261, thunkPosition261 := position, thunkPosition
				{
					position262, thunkPosition262 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDl]() {
						goto l263
					}
					goto l262
				l263:
					position, thunkPosition = position262, thunkPosition262
					{
						position264, thunkPosition264 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDl]() {
							goto l264
						}
						goto l261
					l264:
						position, thunkPosition = position264, thunkPosition264
					}
					if !matchDot() {
						goto l261
					}
				}
			l262:
				goto l260
			l261:
				position, thunkPosition = position261, thunkPosition261
			}
			if !p.rules[ruleHtmlBlockCloseDl]() {
				goto l259
			}
			return true
		l259:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 48 HtmlBlockOpenFieldset <- ('<' Spnl ('fieldset' / 'FIELDSET') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l265
			}
			if !p.rules[ruleSpnl]() {
				goto l265
			}
			{
				position266, thunkPosition266 := position, thunkPosition
				if !matchString("fieldset") {
					goto l267
				}
				goto l266
			l267:
				position, thunkPosition = position266, thunkPosition266
				if !matchString("FIELDSET") {
					goto l265
				}
			}
		l266:
			if !p.rules[ruleSpnl]() {
				goto l265
			}
		l268:
			{
				position269, thunkPosition269 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l269
				}
				goto l268
			l269:
				position, thunkPosition = position269, thunkPosition269
			}
			if !matchChar('>') {
				goto l265
			}
			return true
		l265:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 49 HtmlBlockCloseFieldset <- ('<' Spnl '/' ('fieldset' / 'FIELDSET') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position271, thunkPosition271 := position, thunkPosition
				if !matchString("fieldset") {
					goto l272
				}
				goto l271
			l272:
				position, thunkPosition = position271, thunkPosition271
				if !matchString("FIELDSET") {
					goto l270
				}
			}
		l271:
			if !p.rules[ruleSpnl]() {
				goto l270
			}
			if !matchChar('>') {
				goto l270
			}
			return true
		l270:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 50 HtmlBlockFieldset <- (HtmlBlockOpenFieldset (HtmlBlockFieldset / (!HtmlBlockCloseFieldset .))* HtmlBlockCloseFieldset) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenFieldset]() {
				goto l273
			}
		l274:
			{
				position275, thunkPosition275 := position, thunkPosition
				{
					position276, thunkPosition276 := position, thunkPosition
					if !p.rules[ruleHtmlBlockFieldset]() {
						goto l277
					}
					goto l276
				l277:
					position, thunkPosition = position276, thunkPosition276
					{
						position278, thunkPosition278 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseFieldset]() {
							goto l278
						}
						goto l275
					l278:
						position, thunkPosition = position278, thunkPosition278
					}
					if !matchDot() {
						goto l275
					}
				}
			l276:
				goto l274
			l275:
				position, thunkPosition = position275, thunkPosition275
			}
			if !p.rules[ruleHtmlBlockCloseFieldset]() {
				goto l273
			}
			return true
		l273:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 51 HtmlBlockOpenForm <- ('<' Spnl ('form' / 'FORM') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l279
			}
			if !p.rules[ruleSpnl]() {
				goto l279
			}
			{
				position280, thunkPosition280 := position, thunkPosition
				if !matchString("form") {
					goto l281
				}
				goto l280
			l281:
				position, thunkPosition = position280, thunkPosition280
				if !matchString("FORM") {
					goto l279
				}
			}
		l280:
			if !p.rules[ruleSpnl]() {
				goto l279
			}
		l282:
			{
				position283, thunkPosition283 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l283
				}
				goto l282
			l283:
				position, thunkPosition = position283, thunkPosition283
			}
			if !matchChar('>') {
				goto l279
			}
			return true
		l279:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 52 HtmlBlockCloseForm <- ('<' Spnl '/' ('form' / 'FORM') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l284
			}
			if !p.rules[ruleSpnl]() {
				goto l284
			}
			if !matchChar('/') {
				goto l284
			}
			{
				position285, thunkPosition285 := position, thunkPosition
				if !matchString("form") {
					goto l286
				}
				goto l285
			l286:
				position, thunkPosition = position285, thunkPosition285
				if !matchString("FORM") {
					goto l284
				}
			}
		l285:
			if !p.rules[ruleSpnl]() {
				goto l284
			}
			if !matchChar('>') {
				goto l284
			}
			return true
		l284:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 53 HtmlBlockForm <- (HtmlBlockOpenForm (HtmlBlockForm / (!HtmlBlockCloseForm .))* HtmlBlockCloseForm) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenForm]() {
				goto l287
			}
		l288:
			{
				position289, thunkPosition289 := position, thunkPosition
				{
					position290, thunkPosition290 := position, thunkPosition
					if !p.rules[ruleHtmlBlockForm]() {
						goto l291
					}
					goto l290
				l291:
					position, thunkPosition = position290, thunkPosition290
					{
						position292, thunkPosition292 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseForm]() {
							goto l292
						}
						goto l289
					l292:
						position, thunkPosition = position292, thunkPosition292
					}
					if !matchDot() {
						goto l289
					}
				}
			l290:
				goto l288
			l289:
				position, thunkPosition = position289, thunkPosition289
			}
			if !p.rules[ruleHtmlBlockCloseForm]() {
				goto l287
			}
			return true
		l287:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 54 HtmlBlockOpenH1 <- ('<' Spnl ('h1' / 'H1') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l293
			}
			if !p.rules[ruleSpnl]() {
				goto l293
			}
			{
				position294, thunkPosition294 := position, thunkPosition
				if !matchString("h1") {
					goto l295
				}
				goto l294
			l295:
				position, thunkPosition = position294, thunkPosition294
				if !matchString("H1") {
					goto l293
				}
			}
		l294:
			if !p.rules[ruleSpnl]() {
				goto l293
			}
		l296:
			{
				position297, thunkPosition297 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l297
				}
				goto l296
			l297:
				position, thunkPosition = position297, thunkPosition297
			}
			if !matchChar('>') {
				goto l293
			}
			return true
		l293:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 55 HtmlBlockCloseH1 <- ('<' Spnl '/' ('h1' / 'H1') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l298
			}
			if !p.rules[ruleSpnl]() {
				goto l298
			}
			if !matchChar('/') {
				goto l298
			}
			{
				position299, thunkPosition299 := position, thunkPosition
				if !matchString("h1") {
					goto l300
				}
				goto l299
			l300:
				position, thunkPosition = position299, thunkPosition299
				if !matchString("H1") {
					goto l298
				}
			}
		l299:
			if !p.rules[ruleSpnl]() {
				goto l298
			}
			if !matchChar('>') {
				goto l298
			}
			return true
		l298:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 56 HtmlBlockH1 <- (HtmlBlockOpenH1 (HtmlBlockH1 / (!HtmlBlockCloseH1 .))* HtmlBlockCloseH1) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH1]() {
				goto l301
			}
		l302:
			{
				position303, thunkPosition303 := position, thunkPosition
				{
					position304, thunkPosition304 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH1]() {
						goto l305
					}
					goto l304
				l305:
					position, thunkPosition = position304, thunkPosition304
					{
						position306, thunkPosition306 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH1]() {
							goto l306
						}
						goto l303
					l306:
						position, thunkPosition = position306, thunkPosition306
					}
					if !matchDot() {
						goto l303
					}
				}
			l304:
				goto l302
			l303:
				position, thunkPosition = position303, thunkPosition303
			}
			if !p.rules[ruleHtmlBlockCloseH1]() {
				goto l301
			}
			return true
		l301:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 57 HtmlBlockOpenH2 <- ('<' Spnl ('h2' / 'H2') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l307
			}
			if !p.rules[ruleSpnl]() {
				goto l307
			}
			{
				position308, thunkPosition308 := position, thunkPosition
				if !matchString("h2") {
					goto l309
				}
				goto l308
			l309:
				position, thunkPosition = position308, thunkPosition308
				if !matchString("H2") {
					goto l307
				}
			}
		l308:
			if !p.rules[ruleSpnl]() {
				goto l307
			}
		l310:
			{
				position311, thunkPosition311 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l311
				}
				goto l310
			l311:
				position, thunkPosition = position311, thunkPosition311
			}
			if !matchChar('>') {
				goto l307
			}
			return true
		l307:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 58 HtmlBlockCloseH2 <- ('<' Spnl '/' ('h2' / 'H2') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l312
			}
			if !p.rules[ruleSpnl]() {
				goto l312
			}
			if !matchChar('/') {
				goto l312
			}
			{
				position313, thunkPosition313 := position, thunkPosition
				if !matchString("h2") {
					goto l314
				}
				goto l313
			l314:
				position, thunkPosition = position313, thunkPosition313
				if !matchString("H2") {
					goto l312
				}
			}
		l313:
			if !p.rules[ruleSpnl]() {
				goto l312
			}
			if !matchChar('>') {
				goto l312
			}
			return true
		l312:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 59 HtmlBlockH2 <- (HtmlBlockOpenH2 (HtmlBlockH2 / (!HtmlBlockCloseH2 .))* HtmlBlockCloseH2) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH2]() {
				goto l315
			}
		l316:
			{
				position317, thunkPosition317 := position, thunkPosition
				{
					position318, thunkPosition318 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH2]() {
						goto l319
					}
					goto l318
				l319:
					position, thunkPosition = position318, thunkPosition318
					{
						position320, thunkPosition320 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH2]() {
							goto l320
						}
						goto l317
					l320:
						position, thunkPosition = position320, thunkPosition320
					}
					if !matchDot() {
						goto l317
					}
				}
			l318:
				goto l316
			l317:
				position, thunkPosition = position317, thunkPosition317
			}
			if !p.rules[ruleHtmlBlockCloseH2]() {
				goto l315
			}
			return true
		l315:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 60 HtmlBlockOpenH3 <- ('<' Spnl ('h3' / 'H3') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l321
			}
			if !p.rules[ruleSpnl]() {
				goto l321
			}
			{
				position322, thunkPosition322 := position, thunkPosition
				if !matchString("h3") {
					goto l323
				}
				goto l322
			l323:
				position, thunkPosition = position322, thunkPosition322
				if !matchString("H3") {
					goto l321
				}
			}
		l322:
			if !p.rules[ruleSpnl]() {
				goto l321
			}
		l324:
			{
				position325, thunkPosition325 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l325
				}
				goto l324
			l325:
				position, thunkPosition = position325, thunkPosition325
			}
			if !matchChar('>') {
				goto l321
			}
			return true
		l321:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 61 HtmlBlockCloseH3 <- ('<' Spnl '/' ('h3' / 'H3') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position327, thunkPosition327 := position, thunkPosition
				if !matchString("h3") {
					goto l328
				}
				goto l327
			l328:
				position, thunkPosition = position327, thunkPosition327
				if !matchString("H3") {
					goto l326
				}
			}
		l327:
			if !p.rules[ruleSpnl]() {
				goto l326
			}
			if !matchChar('>') {
				goto l326
			}
			return true
		l326:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 62 HtmlBlockH3 <- (HtmlBlockOpenH3 (HtmlBlockH3 / (!HtmlBlockCloseH3 .))* HtmlBlockCloseH3) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH3]() {
				goto l329
			}
		l330:
			{
				position331, thunkPosition331 := position, thunkPosition
				{
					position332, thunkPosition332 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH3]() {
						goto l333
					}
					goto l332
				l333:
					position, thunkPosition = position332, thunkPosition332
					{
						position334, thunkPosition334 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH3]() {
							goto l334
						}
						goto l331
					l334:
						position, thunkPosition = position334, thunkPosition334
					}
					if !matchDot() {
						goto l331
					}
				}
			l332:
				goto l330
			l331:
				position, thunkPosition = position331, thunkPosition331
			}
			if !p.rules[ruleHtmlBlockCloseH3]() {
				goto l329
			}
			return true
		l329:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 63 HtmlBlockOpenH4 <- ('<' Spnl ('h4' / 'H4') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l335
			}
			if !p.rules[ruleSpnl]() {
				goto l335
			}
			{
				position336, thunkPosition336 := position, thunkPosition
				if !matchString("h4") {
					goto l337
				}
				goto l336
			l337:
				position, thunkPosition = position336, thunkPosition336
				if !matchString("H4") {
					goto l335
				}
			}
		l336:
			if !p.rules[ruleSpnl]() {
				goto l335
			}
		l338:
			{
				position339, thunkPosition339 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l339
				}
				goto l338
			l339:
				position, thunkPosition = position339, thunkPosition339
			}
			if !matchChar('>') {
				goto l335
			}
			return true
		l335:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 64 HtmlBlockCloseH4 <- ('<' Spnl '/' ('h4' / 'H4') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l340
			}
			if !p.rules[ruleSpnl]() {
				goto l340
			}
			if !matchChar('/') {
				goto l340
			}
			{
				position341, thunkPosition341 := position, thunkPosition
				if !matchString("h4") {
					goto l342
				}
				goto l341
			l342:
				position, thunkPosition = position341, thunkPosition341
				if !matchString("H4") {
					goto l340
				}
			}
		l341:
			if !p.rules[ruleSpnl]() {
				goto l340
			}
			if !matchChar('>') {
				goto l340
			}
			return true
		l340:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 65 HtmlBlockH4 <- (HtmlBlockOpenH4 (HtmlBlockH4 / (!HtmlBlockCloseH4 .))* HtmlBlockCloseH4) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH4]() {
				goto l343
			}
		l344:
			{
				position345, thunkPosition345 := position, thunkPosition
				{
					position346, thunkPosition346 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH4]() {
						goto l347
					}
					goto l346
				l347:
					position, thunkPosition = position346, thunkPosition346
					{
						position348, thunkPosition348 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH4]() {
							goto l348
						}
						goto l345
					l348:
						position, thunkPosition = position348, thunkPosition348
					}
					if !matchDot() {
						goto l345
					}
				}
			l346:
				goto l344
			l345:
				position, thunkPosition = position345, thunkPosition345
			}
			if !p.rules[ruleHtmlBlockCloseH4]() {
				goto l343
			}
			return true
		l343:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 66 HtmlBlockOpenH5 <- ('<' Spnl ('h5' / 'H5') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l349
			}
			if !p.rules[ruleSpnl]() {
				goto l349
			}
			{
				position350, thunkPosition350 := position, thunkPosition
				if !matchString("h5") {
					goto l351
				}
				goto l350
			l351:
				position, thunkPosition = position350, thunkPosition350
				if !matchString("H5") {
					goto l349
				}
			}
		l350:
			if !p.rules[ruleSpnl]() {
				goto l349
			}
		l352:
			{
				position353, thunkPosition353 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l353
				}
				goto l352
			l353:
				position, thunkPosition = position353, thunkPosition353
			}
			if !matchChar('>') {
				goto l349
			}
			return true
		l349:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 67 HtmlBlockCloseH5 <- ('<' Spnl '/' ('h5' / 'H5') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position355, thunkPosition355 := position, thunkPosition
				if !matchString("h5") {
					goto l356
				}
				goto l355
			l356:
				position, thunkPosition = position355, thunkPosition355
				if !matchString("H5") {
					goto l354
				}
			}
		l355:
			if !p.rules[ruleSpnl]() {
				goto l354
			}
			if !matchChar('>') {
				goto l354
			}
			return true
		l354:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 68 HtmlBlockH5 <- (HtmlBlockOpenH5 (HtmlBlockH5 / (!HtmlBlockCloseH5 .))* HtmlBlockCloseH5) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH5]() {
				goto l357
			}
		l358:
			{
				position359, thunkPosition359 := position, thunkPosition
				{
					position360, thunkPosition360 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH5]() {
						goto l361
					}
					goto l360
				l361:
					position, thunkPosition = position360, thunkPosition360
					{
						position362, thunkPosition362 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH5]() {
							goto l362
						}
						goto l359
					l362:
						position, thunkPosition = position362, thunkPosition362
					}
					if !matchDot() {
						goto l359
					}
				}
			l360:
				goto l358
			l359:
				position, thunkPosition = position359, thunkPosition359
			}
			if !p.rules[ruleHtmlBlockCloseH5]() {
				goto l357
			}
			return true
		l357:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 69 HtmlBlockOpenH6 <- ('<' Spnl ('h6' / 'H6') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l363
			}
			if !p.rules[ruleSpnl]() {
				goto l363
			}
			{
				position364, thunkPosition364 := position, thunkPosition
				if !matchString("h6") {
					goto l365
				}
				goto l364
			l365:
				position, thunkPosition = position364, thunkPosition364
				if !matchString("H6") {
					goto l363
				}
			}
		l364:
			if !p.rules[ruleSpnl]() {
				goto l363
			}
		l366:
			{
				position367, thunkPosition367 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l367
				}
				goto l366
			l367:
				position, thunkPosition = position367, thunkPosition367
			}
			if !matchChar('>') {
				goto l363
			}
			return true
		l363:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 70 HtmlBlockCloseH6 <- ('<' Spnl '/' ('h6' / 'H6') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l368
			}
			if !p.rules[ruleSpnl]() {
				goto l368
			}
			if !matchChar('/') {
				goto l368
			}
			{
				position369, thunkPosition369 := position, thunkPosition
				if !matchString("h6") {
					goto l370
				}
				goto l369
			l370:
				position, thunkPosition = position369, thunkPosition369
				if !matchString("H6") {
					goto l368
				}
			}
		l369:
			if !p.rules[ruleSpnl]() {
				goto l368
			}
			if !matchChar('>') {
				goto l368
			}
			return true
		l368:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 71 HtmlBlockH6 <- (HtmlBlockOpenH6 (HtmlBlockH6 / (!HtmlBlockCloseH6 .))* HtmlBlockCloseH6) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenH6]() {
				goto l371
			}
		l372:
			{
				position373, thunkPosition373 := position, thunkPosition
				{
					position374, thunkPosition374 := position, thunkPosition
					if !p.rules[ruleHtmlBlockH6]() {
						goto l375
					}
					goto l374
				l375:
					position, thunkPosition = position374, thunkPosition374
					{
						position376, thunkPosition376 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseH6]() {
							goto l376
						}
						goto l373
					l376:
						position, thunkPosition = position376, thunkPosition376
					}
					if !matchDot() {
						goto l373
					}
				}
			l374:
				goto l372
			l373:
				position, thunkPosition = position373, thunkPosition373
			}
			if !p.rules[ruleHtmlBlockCloseH6]() {
				goto l371
			}
			return true
		l371:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 72 HtmlBlockOpenMenu <- ('<' Spnl ('menu' / 'MENU') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l377
			}
			if !p.rules[ruleSpnl]() {
				goto l377
			}
			{
				position378, thunkPosition378 := position, thunkPosition
				if !matchString("menu") {
					goto l379
				}
				goto l378
			l379:
				position, thunkPosition = position378, thunkPosition378
				if !matchString("MENU") {
					goto l377
				}
			}
		l378:
			if !p.rules[ruleSpnl]() {
				goto l377
			}
		l380:
			{
				position381, thunkPosition381 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l381
				}
				goto l380
			l381:
				position, thunkPosition = position381, thunkPosition381
			}
			if !matchChar('>') {
				goto l377
			}
			return true
		l377:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 73 HtmlBlockCloseMenu <- ('<' Spnl '/' ('menu' / 'MENU') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l382
			}
			if !p.rules[ruleSpnl]() {
				goto l382
			}
			if !matchChar('/') {
				goto l382
			}
			{
				position383, thunkPosition383 := position, thunkPosition
				if !matchString("menu") {
					goto l384
				}
				goto l383
			l384:
				position, thunkPosition = position383, thunkPosition383
				if !matchString("MENU") {
					goto l382
				}
			}
		l383:
			if !p.rules[ruleSpnl]() {
				goto l382
			}
			if !matchChar('>') {
				goto l382
			}
			return true
		l382:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 74 HtmlBlockMenu <- (HtmlBlockOpenMenu (HtmlBlockMenu / (!HtmlBlockCloseMenu .))* HtmlBlockCloseMenu) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenMenu]() {
				goto l385
			}
		l386:
			{
				position387, thunkPosition387 := position, thunkPosition
				{
					position388, thunkPosition388 := position, thunkPosition
					if !p.rules[ruleHtmlBlockMenu]() {
						goto l389
					}
					goto l388
				l389:
					position, thunkPosition = position388, thunkPosition388
					{
						position390, thunkPosition390 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseMenu]() {
							goto l390
						}
						goto l387
					l390:
						position, thunkPosition = position390, thunkPosition390
					}
					if !matchDot() {
						goto l387
					}
				}
			l388:
				goto l386
			l387:
				position, thunkPosition = position387, thunkPosition387
			}
			if !p.rules[ruleHtmlBlockCloseMenu]() {
				goto l385
			}
			return true
		l385:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 75 HtmlBlockOpenNoframes <- ('<' Spnl ('noframes' / 'NOFRAMES') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l391
			}
			if !p.rules[ruleSpnl]() {
				goto l391
			}
			{
				position392, thunkPosition392 := position, thunkPosition
				if !matchString("noframes") {
					goto l393
				}
				goto l392
			l393:
				position, thunkPosition = position392, thunkPosition392
				if !matchString("NOFRAMES") {
					goto l391
				}
			}
		l392:
			if !p.rules[ruleSpnl]() {
				goto l391
			}
		l394:
			{
				position395, thunkPosition395 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l395
				}
				goto l394
			l395:
				position, thunkPosition = position395, thunkPosition395
			}
			if !matchChar('>') {
				goto l391
			}
			return true
		l391:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 76 HtmlBlockCloseNoframes <- ('<' Spnl '/' ('noframes' / 'NOFRAMES') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l396
			}
			if !p.rules[ruleSpnl]() {
				goto l396
			}
			if !matchChar('/') {
				goto l396
			}
			{
				position397, thunkPosition397 := position, thunkPosition
				if !matchString("noframes") {
					goto l398
				}
				goto l397
			l398:
				position, thunkPosition = position397, thunkPosition397
				if !matchString("NOFRAMES") {
					goto l396
				}
			}
		l397:
			if !p.rules[ruleSpnl]() {
				goto l396
			}
			if !matchChar('>') {
				goto l396
			}
			return true
		l396:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 77 HtmlBlockNoframes <- (HtmlBlockOpenNoframes (HtmlBlockNoframes / (!HtmlBlockCloseNoframes .))* HtmlBlockCloseNoframes) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenNoframes]() {
				goto l399
			}
		l400:
			{
				position401, thunkPosition401 := position, thunkPosition
				{
					position402, thunkPosition402 := position, thunkPosition
					if !p.rules[ruleHtmlBlockNoframes]() {
						goto l403
					}
					goto l402
				l403:
					position, thunkPosition = position402, thunkPosition402
					{
						position404, thunkPosition404 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseNoframes]() {
							goto l404
						}
						goto l401
					l404:
						position, thunkPosition = position404, thunkPosition404
					}
					if !matchDot() {
						goto l401
					}
				}
			l402:
				goto l400
			l401:
				position, thunkPosition = position401, thunkPosition401
			}
			if !p.rules[ruleHtmlBlockCloseNoframes]() {
				goto l399
			}
			return true
		l399:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 78 HtmlBlockOpenNoscript <- ('<' Spnl ('noscript' / 'NOSCRIPT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l405
			}
			if !p.rules[ruleSpnl]() {
				goto l405
			}
			{
				position406, thunkPosition406 := position, thunkPosition
				if !matchString("noscript") {
					goto l407
				}
				goto l406
			l407:
				position, thunkPosition = position406, thunkPosition406
				if !matchString("NOSCRIPT") {
					goto l405
				}
			}
		l406:
			if !p.rules[ruleSpnl]() {
				goto l405
			}
		l408:
			{
				position409, thunkPosition409 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l409
				}
				goto l408
			l409:
				position, thunkPosition = position409, thunkPosition409
			}
			if !matchChar('>') {
				goto l405
			}
			return true
		l405:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 79 HtmlBlockCloseNoscript <- ('<' Spnl '/' ('noscript' / 'NOSCRIPT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position411, thunkPosition411 := position, thunkPosition
				if !matchString("noscript") {
					goto l412
				}
				goto l411
			l412:
				position, thunkPosition = position411, thunkPosition411
				if !matchString("NOSCRIPT") {
					goto l410
				}
			}
		l411:
			if !p.rules[ruleSpnl]() {
				goto l410
			}
			if !matchChar('>') {
				goto l410
			}
			return true
		l410:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 80 HtmlBlockNoscript <- (HtmlBlockOpenNoscript (HtmlBlockNoscript / (!HtmlBlockCloseNoscript .))* HtmlBlockCloseNoscript) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenNoscript]() {
				goto l413
			}
		l414:
			{
				position415, thunkPosition415 := position, thunkPosition
				{
					position416, thunkPosition416 := position, thunkPosition
					if !p.rules[ruleHtmlBlockNoscript]() {
						goto l417
					}
					goto l416
				l417:
					position, thunkPosition = position416, thunkPosition416
					{
						position418, thunkPosition418 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseNoscript]() {
							goto l418
						}
						goto l415
					l418:
						position, thunkPosition = position418, thunkPosition418
					}
					if !matchDot() {
						goto l415
					}
				}
			l416:
				goto l414
			l415:
				position, thunkPosition = position415, thunkPosition415
			}
			if !p.rules[ruleHtmlBlockCloseNoscript]() {
				goto l413
			}
			return true
		l413:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 81 HtmlBlockOpenOl <- ('<' Spnl ('ol' / 'OL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l419
			}
			if !p.rules[ruleSpnl]() {
				goto l419
			}
			{
				position420, thunkPosition420 := position, thunkPosition
				if !matchString("ol") {
					goto l421
				}
				goto l420
			l421:
				position, thunkPosition = position420, thunkPosition420
				if !matchString("OL") {
					goto l419
				}
			}
		l420:
			if !p.rules[ruleSpnl]() {
				goto l419
			}
		l422:
			{
				position423, thunkPosition423 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l423
				}
				goto l422
			l423:
				position, thunkPosition = position423, thunkPosition423
			}
			if !matchChar('>') {
				goto l419
			}
			return true
		l419:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 82 HtmlBlockCloseOl <- ('<' Spnl '/' ('ol' / 'OL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l424
			}
			if !p.rules[ruleSpnl]() {
				goto l424
			}
			if !matchChar('/') {
				goto l424
			}
			{
				position425, thunkPosition425 := position, thunkPosition
				if !matchString("ol") {
					goto l426
				}
				goto l425
			l426:
				position, thunkPosition = position425, thunkPosition425
				if !matchString("OL") {
					goto l424
				}
			}
		l425:
			if !p.rules[ruleSpnl]() {
				goto l424
			}
			if !matchChar('>') {
				goto l424
			}
			return true
		l424:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 83 HtmlBlockOl <- (HtmlBlockOpenOl (HtmlBlockOl / (!HtmlBlockCloseOl .))* HtmlBlockCloseOl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenOl]() {
				goto l427
			}
		l428:
			{
				position429, thunkPosition429 := position, thunkPosition
				{
					position430, thunkPosition430 := position, thunkPosition
					if !p.rules[ruleHtmlBlockOl]() {
						goto l431
					}
					goto l430
				l431:
					position, thunkPosition = position430, thunkPosition430
					{
						position432, thunkPosition432 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseOl]() {
							goto l432
						}
						goto l429
					l432:
						position, thunkPosition = position432, thunkPosition432
					}
					if !matchDot() {
						goto l429
					}
				}
			l430:
				goto l428
			l429:
				position, thunkPosition = position429, thunkPosition429
			}
			if !p.rules[ruleHtmlBlockCloseOl]() {
				goto l427
			}
			return true
		l427:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 84 HtmlBlockOpenP <- ('<' Spnl ('p' / 'P') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l433
			}
			if !p.rules[ruleSpnl]() {
				goto l433
			}
			{
				position434, thunkPosition434 := position, thunkPosition
				if !matchChar('p') {
					goto l435
				}
				goto l434
			l435:
				position, thunkPosition = position434, thunkPosition434
				if !matchChar('P') {
					goto l433
				}
			}
		l434:
			if !p.rules[ruleSpnl]() {
				goto l433
			}
		l436:
			{
				position437, thunkPosition437 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l437
				}
				goto l436
			l437:
				position, thunkPosition = position437, thunkPosition437
			}
			if !matchChar('>') {
				goto l433
			}
			return true
		l433:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 85 HtmlBlockCloseP <- ('<' Spnl '/' ('p' / 'P') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position439, thunkPosition439 := position, thunkPosition
				if !matchChar('p') {
					goto l440
				}
				goto l439
			l440:
				position, thunkPosition = position439, thunkPosition439
				if !matchChar('P') {
					goto l438
				}
			}
		l439:
			if !p.rules[ruleSpnl]() {
				goto l438
			}
			if !matchChar('>') {
				goto l438
			}
			return true
		l438:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 86 HtmlBlockP <- (HtmlBlockOpenP (HtmlBlockP / (!HtmlBlockCloseP .))* HtmlBlockCloseP) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenP]() {
				goto l441
			}
		l442:
			{
				position443, thunkPosition443 := position, thunkPosition
				{
					position444, thunkPosition444 := position, thunkPosition
					if !p.rules[ruleHtmlBlockP]() {
						goto l445
					}
					goto l444
				l445:
					position, thunkPosition = position444, thunkPosition444
					{
						position446, thunkPosition446 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseP]() {
							goto l446
						}
						goto l443
					l446:
						position, thunkPosition = position446, thunkPosition446
					}
					if !matchDot() {
						goto l443
					}
				}
			l444:
				goto l442
			l443:
				position, thunkPosition = position443, thunkPosition443
			}
			if !p.rules[ruleHtmlBlockCloseP]() {
				goto l441
			}
			return true
		l441:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 87 HtmlBlockOpenPre <- ('<' Spnl ('pre' / 'PRE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l447
			}
			if !p.rules[ruleSpnl]() {
				goto l447
			}
			{
				position448, thunkPosition448 := position, thunkPosition
				if !matchString("pre") {
					goto l449
				}
				goto l448
			l449:
				position, thunkPosition = position448, thunkPosition448
				if !matchString("PRE") {
					goto l447
				}
			}
		l448:
			if !p.rules[ruleSpnl]() {
				goto l447
			}
		l450:
			{
				position451, thunkPosition451 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l451
				}
				goto l450
			l451:
				position, thunkPosition = position451, thunkPosition451
			}
			if !matchChar('>') {
				goto l447
			}
			return true
		l447:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 88 HtmlBlockClosePre <- ('<' Spnl '/' ('pre' / 'PRE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l452
			}
			if !p.rules[ruleSpnl]() {
				goto l452
			}
			if !matchChar('/') {
				goto l452
			}
			{
				position453, thunkPosition453 := position, thunkPosition
				if !matchString("pre") {
					goto l454
				}
				goto l453
			l454:
				position, thunkPosition = position453, thunkPosition453
				if !matchString("PRE") {
					goto l452
				}
			}
		l453:
			if !p.rules[ruleSpnl]() {
				goto l452
			}
			if !matchChar('>') {
				goto l452
			}
			return true
		l452:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 89 HtmlBlockPre <- (HtmlBlockOpenPre (HtmlBlockPre / (!HtmlBlockClosePre .))* HtmlBlockClosePre) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenPre]() {
				goto l455
			}
		l456:
			{
				position457, thunkPosition457 := position, thunkPosition
				{
					position458, thunkPosition458 := position, thunkPosition
					if !p.rules[ruleHtmlBlockPre]() {
						goto l459
					}
					goto l458
				l459:
					position, thunkPosition = position458, thunkPosition458
					{
						position460, thunkPosition460 := position, thunkPosition
						if !p.rules[ruleHtmlBlockClosePre]() {
							goto l460
						}
						goto l457
					l460:
						position, thunkPosition = position460, thunkPosition460
					}
					if !matchDot() {
						goto l457
					}
				}
			l458:
				goto l456
			l457:
				position, thunkPosition = position457, thunkPosition457
			}
			if !p.rules[ruleHtmlBlockClosePre]() {
				goto l455
			}
			return true
		l455:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 90 HtmlBlockOpenTable <- ('<' Spnl ('table' / 'TABLE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l461
			}
			if !p.rules[ruleSpnl]() {
				goto l461
			}
			{
				position462, thunkPosition462 := position, thunkPosition
				if !matchString("table") {
					goto l463
				}
				goto l462
			l463:
				position, thunkPosition = position462, thunkPosition462
				if !matchString("TABLE") {
					goto l461
				}
			}
		l462:
			if !p.rules[ruleSpnl]() {
				goto l461
			}
		l464:
			{
				position465, thunkPosition465 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l465
				}
				goto l464
			l465:
				position, thunkPosition = position465, thunkPosition465
			}
			if !matchChar('>') {
				goto l461
			}
			return true
		l461:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 91 HtmlBlockCloseTable <- ('<' Spnl '/' ('table' / 'TABLE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l466
			}
			if !p.rules[ruleSpnl]() {
				goto l466
			}
			if !matchChar('/') {
				goto l466
			}
			{
				position467, thunkPosition467 := position, thunkPosition
				if !matchString("table") {
					goto l468
				}
				goto l467
			l468:
				position, thunkPosition = position467, thunkPosition467
				if !matchString("TABLE") {
					goto l466
				}
			}
		l467:
			if !p.rules[ruleSpnl]() {
				goto l466
			}
			if !matchChar('>') {
				goto l466
			}
			return true
		l466:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 92 HtmlBlockTable <- (HtmlBlockOpenTable (HtmlBlockTable / (!HtmlBlockCloseTable .))* HtmlBlockCloseTable) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTable]() {
				goto l469
			}
		l470:
			{
				position471, thunkPosition471 := position, thunkPosition
				{
					position472, thunkPosition472 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTable]() {
						goto l473
					}
					goto l472
				l473:
					position, thunkPosition = position472, thunkPosition472
					{
						position474, thunkPosition474 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTable]() {
							goto l474
						}
						goto l471
					l474:
						position, thunkPosition = position474, thunkPosition474
					}
					if !matchDot() {
						goto l471
					}
				}
			l472:
				goto l470
			l471:
				position, thunkPosition = position471, thunkPosition471
			}
			if !p.rules[ruleHtmlBlockCloseTable]() {
				goto l469
			}
			return true
		l469:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 93 HtmlBlockOpenUl <- ('<' Spnl ('ul' / 'UL') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l475
			}
			if !p.rules[ruleSpnl]() {
				goto l475
			}
			{
				position476, thunkPosition476 := position, thunkPosition
				if !matchString("ul") {
					goto l477
				}
				goto l476
			l477:
				position, thunkPosition = position476, thunkPosition476
				if !matchString("UL") {
					goto l475
				}
			}
		l476:
			if !p.rules[ruleSpnl]() {
				goto l475
			}
		l478:
			{
				position479, thunkPosition479 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l479
				}
				goto l478
			l479:
				position, thunkPosition = position479, thunkPosition479
			}
			if !matchChar('>') {
				goto l475
			}
			return true
		l475:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 94 HtmlBlockCloseUl <- ('<' Spnl '/' ('ul' / 'UL') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l480
			}
			if !p.rules[ruleSpnl]() {
				goto l480
			}
			if !matchChar('/') {
				goto l480
			}
			{
				position481, thunkPosition481 := position, thunkPosition
				if !matchString("ul") {
					goto l482
				}
				goto l481
			l482:
				position, thunkPosition = position481, thunkPosition481
				if !matchString("UL") {
					goto l480
				}
			}
		l481:
			if !p.rules[ruleSpnl]() {
				goto l480
			}
			if !matchChar('>') {
				goto l480
			}
			return true
		l480:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 95 HtmlBlockUl <- (HtmlBlockOpenUl (HtmlBlockUl / (!HtmlBlockCloseUl .))* HtmlBlockCloseUl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenUl]() {
				goto l483
			}
		l484:
			{
				position485, thunkPosition485 := position, thunkPosition
				{
					position486, thunkPosition486 := position, thunkPosition
					if !p.rules[ruleHtmlBlockUl]() {
						goto l487
					}
					goto l486
				l487:
					position, thunkPosition = position486, thunkPosition486
					{
						position488, thunkPosition488 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseUl]() {
							goto l488
						}
						goto l485
					l488:
						position, thunkPosition = position488, thunkPosition488
					}
					if !matchDot() {
						goto l485
					}
				}
			l486:
				goto l484
			l485:
				position, thunkPosition = position485, thunkPosition485
			}
			if !p.rules[ruleHtmlBlockCloseUl]() {
				goto l483
			}
			return true
		l483:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 96 HtmlBlockOpenDd <- ('<' Spnl ('dd' / 'DD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l489
			}
			if !p.rules[ruleSpnl]() {
				goto l489
			}
			{
				position490, thunkPosition490 := position, thunkPosition
				if !matchString("dd") {
					goto l491
				}
				goto l490
			l491:
				position, thunkPosition = position490, thunkPosition490
				if !matchString("DD") {
					goto l489
				}
			}
		l490:
			if !p.rules[ruleSpnl]() {
				goto l489
			}
		l492:
			{
				position493, thunkPosition493 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l493
				}
				goto l492
			l493:
				position, thunkPosition = position493, thunkPosition493
			}
			if !matchChar('>') {
				goto l489
			}
			return true
		l489:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 97 HtmlBlockCloseDd <- ('<' Spnl '/' ('dd' / 'DD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position495, thunkPosition495 := position, thunkPosition
				if !matchString("dd") {
					goto l496
				}
				goto l495
			l496:
				position, thunkPosition = position495, thunkPosition495
				if !matchString("DD") {
					goto l494
				}
			}
		l495:
			if !p.rules[ruleSpnl]() {
				goto l494
			}
			if !matchChar('>') {
				goto l494
			}
			return true
		l494:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 98 HtmlBlockDd <- (HtmlBlockOpenDd (HtmlBlockDd / (!HtmlBlockCloseDd .))* HtmlBlockCloseDd) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDd]() {
				goto l497
			}
		l498:
			{
				position499, thunkPosition499 := position, thunkPosition
				{
					position500, thunkPosition500 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDd]() {
						goto l501
					}
					goto l500
				l501:
					position, thunkPosition = position500, thunkPosition500
					{
						position502, thunkPosition502 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDd]() {
							goto l502
						}
						goto l499
					l502:
						position, thunkPosition = position502, thunkPosition502
					}
					if !matchDot() {
						goto l499
					}
				}
			l500:
				goto l498
			l499:
				position, thunkPosition = position499, thunkPosition499
			}
			if !p.rules[ruleHtmlBlockCloseDd]() {
				goto l497
			}
			return true
		l497:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 99 HtmlBlockOpenDt <- ('<' Spnl ('dt' / 'DT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l503
			}
			if !p.rules[ruleSpnl]() {
				goto l503
			}
			{
				position504, thunkPosition504 := position, thunkPosition
				if !matchString("dt") {
					goto l505
				}
				goto l504
			l505:
				position, thunkPosition = position504, thunkPosition504
				if !matchString("DT") {
					goto l503
				}
			}
		l504:
			if !p.rules[ruleSpnl]() {
				goto l503
			}
		l506:
			{
				position507, thunkPosition507 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l507
				}
				goto l506
			l507:
				position, thunkPosition = position507, thunkPosition507
			}
			if !matchChar('>') {
				goto l503
			}
			return true
		l503:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 100 HtmlBlockCloseDt <- ('<' Spnl '/' ('dt' / 'DT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l508
			}
			if !p.rules[ruleSpnl]() {
				goto l508
			}
			if !matchChar('/') {
				goto l508
			}
			{
				position509, thunkPosition509 := position, thunkPosition
				if !matchString("dt") {
					goto l510
				}
				goto l509
			l510:
				position, thunkPosition = position509, thunkPosition509
				if !matchString("DT") {
					goto l508
				}
			}
		l509:
			if !p.rules[ruleSpnl]() {
				goto l508
			}
			if !matchChar('>') {
				goto l508
			}
			return true
		l508:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 101 HtmlBlockDt <- (HtmlBlockOpenDt (HtmlBlockDt / (!HtmlBlockCloseDt .))* HtmlBlockCloseDt) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenDt]() {
				goto l511
			}
		l512:
			{
				position513, thunkPosition513 := position, thunkPosition
				{
					position514, thunkPosition514 := position, thunkPosition
					if !p.rules[ruleHtmlBlockDt]() {
						goto l515
					}
					goto l514
				l515:
					position, thunkPosition = position514, thunkPosition514
					{
						position516, thunkPosition516 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseDt]() {
							goto l516
						}
						goto l513
					l516:
						position, thunkPosition = position516, thunkPosition516
					}
					if !matchDot() {
						goto l513
					}
				}
			l514:
				goto l512
			l513:
				position, thunkPosition = position513, thunkPosition513
			}
			if !p.rules[ruleHtmlBlockCloseDt]() {
				goto l511
			}
			return true
		l511:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 102 HtmlBlockOpenFrameset <- ('<' Spnl ('frameset' / 'FRAMESET') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l517
			}
			if !p.rules[ruleSpnl]() {
				goto l517
			}
			{
				position518, thunkPosition518 := position, thunkPosition
				if !matchString("frameset") {
					goto l519
				}
				goto l518
			l519:
				position, thunkPosition = position518, thunkPosition518
				if !matchString("FRAMESET") {
					goto l517
				}
			}
		l518:
			if !p.rules[ruleSpnl]() {
				goto l517
			}
		l520:
			{
				position521, thunkPosition521 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l521
				}
				goto l520
			l521:
				position, thunkPosition = position521, thunkPosition521
			}
			if !matchChar('>') {
				goto l517
			}
			return true
		l517:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 103 HtmlBlockCloseFrameset <- ('<' Spnl '/' ('frameset' / 'FRAMESET') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
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
				position523, thunkPosition523 := position, thunkPosition
				if !matchString("frameset") {
					goto l524
				}
				goto l523
			l524:
				position, thunkPosition = position523, thunkPosition523
				if !matchString("FRAMESET") {
					goto l522
				}
			}
		l523:
			if !p.rules[ruleSpnl]() {
				goto l522
			}
			if !matchChar('>') {
				goto l522
			}
			return true
		l522:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 104 HtmlBlockFrameset <- (HtmlBlockOpenFrameset (HtmlBlockFrameset / (!HtmlBlockCloseFrameset .))* HtmlBlockCloseFrameset) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenFrameset]() {
				goto l525
			}
		l526:
			{
				position527, thunkPosition527 := position, thunkPosition
				{
					position528, thunkPosition528 := position, thunkPosition
					if !p.rules[ruleHtmlBlockFrameset]() {
						goto l529
					}
					goto l528
				l529:
					position, thunkPosition = position528, thunkPosition528
					{
						position530, thunkPosition530 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseFrameset]() {
							goto l530
						}
						goto l527
					l530:
						position, thunkPosition = position530, thunkPosition530
					}
					if !matchDot() {
						goto l527
					}
				}
			l528:
				goto l526
			l527:
				position, thunkPosition = position527, thunkPosition527
			}
			if !p.rules[ruleHtmlBlockCloseFrameset]() {
				goto l525
			}
			return true
		l525:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 105 HtmlBlockOpenLi <- ('<' Spnl ('li' / 'LI') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l531
			}
			if !p.rules[ruleSpnl]() {
				goto l531
			}
			{
				position532, thunkPosition532 := position, thunkPosition
				if !matchString("li") {
					goto l533
				}
				goto l532
			l533:
				position, thunkPosition = position532, thunkPosition532
				if !matchString("LI") {
					goto l531
				}
			}
		l532:
			if !p.rules[ruleSpnl]() {
				goto l531
			}
		l534:
			{
				position535, thunkPosition535 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l535
				}
				goto l534
			l535:
				position, thunkPosition = position535, thunkPosition535
			}
			if !matchChar('>') {
				goto l531
			}
			return true
		l531:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 106 HtmlBlockCloseLi <- ('<' Spnl '/' ('li' / 'LI') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l536
			}
			if !p.rules[ruleSpnl]() {
				goto l536
			}
			if !matchChar('/') {
				goto l536
			}
			{
				position537, thunkPosition537 := position, thunkPosition
				if !matchString("li") {
					goto l538
				}
				goto l537
			l538:
				position, thunkPosition = position537, thunkPosition537
				if !matchString("LI") {
					goto l536
				}
			}
		l537:
			if !p.rules[ruleSpnl]() {
				goto l536
			}
			if !matchChar('>') {
				goto l536
			}
			return true
		l536:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 107 HtmlBlockLi <- (HtmlBlockOpenLi (HtmlBlockLi / (!HtmlBlockCloseLi .))* HtmlBlockCloseLi) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenLi]() {
				goto l539
			}
		l540:
			{
				position541, thunkPosition541 := position, thunkPosition
				{
					position542, thunkPosition542 := position, thunkPosition
					if !p.rules[ruleHtmlBlockLi]() {
						goto l543
					}
					goto l542
				l543:
					position, thunkPosition = position542, thunkPosition542
					{
						position544, thunkPosition544 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseLi]() {
							goto l544
						}
						goto l541
					l544:
						position, thunkPosition = position544, thunkPosition544
					}
					if !matchDot() {
						goto l541
					}
				}
			l542:
				goto l540
			l541:
				position, thunkPosition = position541, thunkPosition541
			}
			if !p.rules[ruleHtmlBlockCloseLi]() {
				goto l539
			}
			return true
		l539:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 108 HtmlBlockOpenTbody <- ('<' Spnl ('tbody' / 'TBODY') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l545
			}
			if !p.rules[ruleSpnl]() {
				goto l545
			}
			{
				position546, thunkPosition546 := position, thunkPosition
				if !matchString("tbody") {
					goto l547
				}
				goto l546
			l547:
				position, thunkPosition = position546, thunkPosition546
				if !matchString("TBODY") {
					goto l545
				}
			}
		l546:
			if !p.rules[ruleSpnl]() {
				goto l545
			}
		l548:
			{
				position549, thunkPosition549 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l549
				}
				goto l548
			l549:
				position, thunkPosition = position549, thunkPosition549
			}
			if !matchChar('>') {
				goto l545
			}
			return true
		l545:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 109 HtmlBlockCloseTbody <- ('<' Spnl '/' ('tbody' / 'TBODY') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l550
			}
			if !p.rules[ruleSpnl]() {
				goto l550
			}
			if !matchChar('/') {
				goto l550
			}
			{
				position551, thunkPosition551 := position, thunkPosition
				if !matchString("tbody") {
					goto l552
				}
				goto l551
			l552:
				position, thunkPosition = position551, thunkPosition551
				if !matchString("TBODY") {
					goto l550
				}
			}
		l551:
			if !p.rules[ruleSpnl]() {
				goto l550
			}
			if !matchChar('>') {
				goto l550
			}
			return true
		l550:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 110 HtmlBlockTbody <- (HtmlBlockOpenTbody (HtmlBlockTbody / (!HtmlBlockCloseTbody .))* HtmlBlockCloseTbody) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTbody]() {
				goto l553
			}
		l554:
			{
				position555, thunkPosition555 := position, thunkPosition
				{
					position556, thunkPosition556 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTbody]() {
						goto l557
					}
					goto l556
				l557:
					position, thunkPosition = position556, thunkPosition556
					{
						position558, thunkPosition558 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTbody]() {
							goto l558
						}
						goto l555
					l558:
						position, thunkPosition = position558, thunkPosition558
					}
					if !matchDot() {
						goto l555
					}
				}
			l556:
				goto l554
			l555:
				position, thunkPosition = position555, thunkPosition555
			}
			if !p.rules[ruleHtmlBlockCloseTbody]() {
				goto l553
			}
			return true
		l553:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 111 HtmlBlockOpenTd <- ('<' Spnl ('td' / 'TD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l559
			}
			if !p.rules[ruleSpnl]() {
				goto l559
			}
			{
				position560, thunkPosition560 := position, thunkPosition
				if !matchString("td") {
					goto l561
				}
				goto l560
			l561:
				position, thunkPosition = position560, thunkPosition560
				if !matchString("TD") {
					goto l559
				}
			}
		l560:
			if !p.rules[ruleSpnl]() {
				goto l559
			}
		l562:
			{
				position563, thunkPosition563 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l563
				}
				goto l562
			l563:
				position, thunkPosition = position563, thunkPosition563
			}
			if !matchChar('>') {
				goto l559
			}
			return true
		l559:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 112 HtmlBlockCloseTd <- ('<' Spnl '/' ('td' / 'TD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l564
			}
			if !p.rules[ruleSpnl]() {
				goto l564
			}
			if !matchChar('/') {
				goto l564
			}
			{
				position565, thunkPosition565 := position, thunkPosition
				if !matchString("td") {
					goto l566
				}
				goto l565
			l566:
				position, thunkPosition = position565, thunkPosition565
				if !matchString("TD") {
					goto l564
				}
			}
		l565:
			if !p.rules[ruleSpnl]() {
				goto l564
			}
			if !matchChar('>') {
				goto l564
			}
			return true
		l564:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 113 HtmlBlockTd <- (HtmlBlockOpenTd (HtmlBlockTd / (!HtmlBlockCloseTd .))* HtmlBlockCloseTd) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTd]() {
				goto l567
			}
		l568:
			{
				position569, thunkPosition569 := position, thunkPosition
				{
					position570, thunkPosition570 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTd]() {
						goto l571
					}
					goto l570
				l571:
					position, thunkPosition = position570, thunkPosition570
					{
						position572, thunkPosition572 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTd]() {
							goto l572
						}
						goto l569
					l572:
						position, thunkPosition = position572, thunkPosition572
					}
					if !matchDot() {
						goto l569
					}
				}
			l570:
				goto l568
			l569:
				position, thunkPosition = position569, thunkPosition569
			}
			if !p.rules[ruleHtmlBlockCloseTd]() {
				goto l567
			}
			return true
		l567:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 114 HtmlBlockOpenTfoot <- ('<' Spnl ('tfoot' / 'TFOOT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l573
			}
			if !p.rules[ruleSpnl]() {
				goto l573
			}
			{
				position574, thunkPosition574 := position, thunkPosition
				if !matchString("tfoot") {
					goto l575
				}
				goto l574
			l575:
				position, thunkPosition = position574, thunkPosition574
				if !matchString("TFOOT") {
					goto l573
				}
			}
		l574:
			if !p.rules[ruleSpnl]() {
				goto l573
			}
		l576:
			{
				position577, thunkPosition577 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l577
				}
				goto l576
			l577:
				position, thunkPosition = position577, thunkPosition577
			}
			if !matchChar('>') {
				goto l573
			}
			return true
		l573:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 115 HtmlBlockCloseTfoot <- ('<' Spnl '/' ('tfoot' / 'TFOOT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l578
			}
			if !p.rules[ruleSpnl]() {
				goto l578
			}
			if !matchChar('/') {
				goto l578
			}
			{
				position579, thunkPosition579 := position, thunkPosition
				if !matchString("tfoot") {
					goto l580
				}
				goto l579
			l580:
				position, thunkPosition = position579, thunkPosition579
				if !matchString("TFOOT") {
					goto l578
				}
			}
		l579:
			if !p.rules[ruleSpnl]() {
				goto l578
			}
			if !matchChar('>') {
				goto l578
			}
			return true
		l578:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 116 HtmlBlockTfoot <- (HtmlBlockOpenTfoot (HtmlBlockTfoot / (!HtmlBlockCloseTfoot .))* HtmlBlockCloseTfoot) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTfoot]() {
				goto l581
			}
		l582:
			{
				position583, thunkPosition583 := position, thunkPosition
				{
					position584, thunkPosition584 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTfoot]() {
						goto l585
					}
					goto l584
				l585:
					position, thunkPosition = position584, thunkPosition584
					{
						position586, thunkPosition586 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTfoot]() {
							goto l586
						}
						goto l583
					l586:
						position, thunkPosition = position586, thunkPosition586
					}
					if !matchDot() {
						goto l583
					}
				}
			l584:
				goto l582
			l583:
				position, thunkPosition = position583, thunkPosition583
			}
			if !p.rules[ruleHtmlBlockCloseTfoot]() {
				goto l581
			}
			return true
		l581:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 117 HtmlBlockOpenTh <- ('<' Spnl ('th' / 'TH') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l587
			}
			if !p.rules[ruleSpnl]() {
				goto l587
			}
			{
				position588, thunkPosition588 := position, thunkPosition
				if !matchString("th") {
					goto l589
				}
				goto l588
			l589:
				position, thunkPosition = position588, thunkPosition588
				if !matchString("TH") {
					goto l587
				}
			}
		l588:
			if !p.rules[ruleSpnl]() {
				goto l587
			}
		l590:
			{
				position591, thunkPosition591 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l591
				}
				goto l590
			l591:
				position, thunkPosition = position591, thunkPosition591
			}
			if !matchChar('>') {
				goto l587
			}
			return true
		l587:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 118 HtmlBlockCloseTh <- ('<' Spnl '/' ('th' / 'TH') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l592
			}
			if !p.rules[ruleSpnl]() {
				goto l592
			}
			if !matchChar('/') {
				goto l592
			}
			{
				position593, thunkPosition593 := position, thunkPosition
				if !matchString("th") {
					goto l594
				}
				goto l593
			l594:
				position, thunkPosition = position593, thunkPosition593
				if !matchString("TH") {
					goto l592
				}
			}
		l593:
			if !p.rules[ruleSpnl]() {
				goto l592
			}
			if !matchChar('>') {
				goto l592
			}
			return true
		l592:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 119 HtmlBlockTh <- (HtmlBlockOpenTh (HtmlBlockTh / (!HtmlBlockCloseTh .))* HtmlBlockCloseTh) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTh]() {
				goto l595
			}
		l596:
			{
				position597, thunkPosition597 := position, thunkPosition
				{
					position598, thunkPosition598 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTh]() {
						goto l599
					}
					goto l598
				l599:
					position, thunkPosition = position598, thunkPosition598
					{
						position600, thunkPosition600 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTh]() {
							goto l600
						}
						goto l597
					l600:
						position, thunkPosition = position600, thunkPosition600
					}
					if !matchDot() {
						goto l597
					}
				}
			l598:
				goto l596
			l597:
				position, thunkPosition = position597, thunkPosition597
			}
			if !p.rules[ruleHtmlBlockCloseTh]() {
				goto l595
			}
			return true
		l595:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 120 HtmlBlockOpenThead <- ('<' Spnl ('thead' / 'THEAD') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l601
			}
			if !p.rules[ruleSpnl]() {
				goto l601
			}
			{
				position602, thunkPosition602 := position, thunkPosition
				if !matchString("thead") {
					goto l603
				}
				goto l602
			l603:
				position, thunkPosition = position602, thunkPosition602
				if !matchString("THEAD") {
					goto l601
				}
			}
		l602:
			if !p.rules[ruleSpnl]() {
				goto l601
			}
		l604:
			{
				position605, thunkPosition605 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l605
				}
				goto l604
			l605:
				position, thunkPosition = position605, thunkPosition605
			}
			if !matchChar('>') {
				goto l601
			}
			return true
		l601:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 121 HtmlBlockCloseThead <- ('<' Spnl '/' ('thead' / 'THEAD') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l606
			}
			if !p.rules[ruleSpnl]() {
				goto l606
			}
			if !matchChar('/') {
				goto l606
			}
			{
				position607, thunkPosition607 := position, thunkPosition
				if !matchString("thead") {
					goto l608
				}
				goto l607
			l608:
				position, thunkPosition = position607, thunkPosition607
				if !matchString("THEAD") {
					goto l606
				}
			}
		l607:
			if !p.rules[ruleSpnl]() {
				goto l606
			}
			if !matchChar('>') {
				goto l606
			}
			return true
		l606:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 122 HtmlBlockThead <- (HtmlBlockOpenThead (HtmlBlockThead / (!HtmlBlockCloseThead .))* HtmlBlockCloseThead) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenThead]() {
				goto l609
			}
		l610:
			{
				position611, thunkPosition611 := position, thunkPosition
				{
					position612, thunkPosition612 := position, thunkPosition
					if !p.rules[ruleHtmlBlockThead]() {
						goto l613
					}
					goto l612
				l613:
					position, thunkPosition = position612, thunkPosition612
					{
						position614, thunkPosition614 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseThead]() {
							goto l614
						}
						goto l611
					l614:
						position, thunkPosition = position614, thunkPosition614
					}
					if !matchDot() {
						goto l611
					}
				}
			l612:
				goto l610
			l611:
				position, thunkPosition = position611, thunkPosition611
			}
			if !p.rules[ruleHtmlBlockCloseThead]() {
				goto l609
			}
			return true
		l609:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 123 HtmlBlockOpenTr <- ('<' Spnl ('tr' / 'TR') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l615
			}
			if !p.rules[ruleSpnl]() {
				goto l615
			}
			{
				position616, thunkPosition616 := position, thunkPosition
				if !matchString("tr") {
					goto l617
				}
				goto l616
			l617:
				position, thunkPosition = position616, thunkPosition616
				if !matchString("TR") {
					goto l615
				}
			}
		l616:
			if !p.rules[ruleSpnl]() {
				goto l615
			}
		l618:
			{
				position619, thunkPosition619 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l619
				}
				goto l618
			l619:
				position, thunkPosition = position619, thunkPosition619
			}
			if !matchChar('>') {
				goto l615
			}
			return true
		l615:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 124 HtmlBlockCloseTr <- ('<' Spnl '/' ('tr' / 'TR') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l620
			}
			if !p.rules[ruleSpnl]() {
				goto l620
			}
			if !matchChar('/') {
				goto l620
			}
			{
				position621, thunkPosition621 := position, thunkPosition
				if !matchString("tr") {
					goto l622
				}
				goto l621
			l622:
				position, thunkPosition = position621, thunkPosition621
				if !matchString("TR") {
					goto l620
				}
			}
		l621:
			if !p.rules[ruleSpnl]() {
				goto l620
			}
			if !matchChar('>') {
				goto l620
			}
			return true
		l620:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 125 HtmlBlockTr <- (HtmlBlockOpenTr (HtmlBlockTr / (!HtmlBlockCloseTr .))* HtmlBlockCloseTr) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenTr]() {
				goto l623
			}
		l624:
			{
				position625, thunkPosition625 := position, thunkPosition
				{
					position626, thunkPosition626 := position, thunkPosition
					if !p.rules[ruleHtmlBlockTr]() {
						goto l627
					}
					goto l626
				l627:
					position, thunkPosition = position626, thunkPosition626
					{
						position628, thunkPosition628 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseTr]() {
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
			if !p.rules[ruleHtmlBlockCloseTr]() {
				goto l623
			}
			return true
		l623:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 126 HtmlBlockOpenScript <- ('<' Spnl ('script' / 'SCRIPT') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l629
			}
			if !p.rules[ruleSpnl]() {
				goto l629
			}
			{
				position630, thunkPosition630 := position, thunkPosition
				if !matchString("script") {
					goto l631
				}
				goto l630
			l631:
				position, thunkPosition = position630, thunkPosition630
				if !matchString("SCRIPT") {
					goto l629
				}
			}
		l630:
			if !p.rules[ruleSpnl]() {
				goto l629
			}
		l632:
			{
				position633, thunkPosition633 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l633
				}
				goto l632
			l633:
				position, thunkPosition = position633, thunkPosition633
			}
			if !matchChar('>') {
				goto l629
			}
			return true
		l629:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 127 HtmlBlockCloseScript <- ('<' Spnl '/' ('script' / 'SCRIPT') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l634
			}
			if !p.rules[ruleSpnl]() {
				goto l634
			}
			if !matchChar('/') {
				goto l634
			}
			{
				position635, thunkPosition635 := position, thunkPosition
				if !matchString("script") {
					goto l636
				}
				goto l635
			l636:
				position, thunkPosition = position635, thunkPosition635
				if !matchString("SCRIPT") {
					goto l634
				}
			}
		l635:
			if !p.rules[ruleSpnl]() {
				goto l634
			}
			if !matchChar('>') {
				goto l634
			}
			return true
		l634:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 128 HtmlBlockScript <- (HtmlBlockOpenScript (HtmlBlockScript / (!HtmlBlockCloseScript .))* HtmlBlockCloseScript) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHtmlBlockOpenScript]() {
				goto l637
			}
		l638:
			{
				position639, thunkPosition639 := position, thunkPosition
				{
					position640, thunkPosition640 := position, thunkPosition
					if !p.rules[ruleHtmlBlockScript]() {
						goto l641
					}
					goto l640
				l641:
					position, thunkPosition = position640, thunkPosition640
					{
						position642, thunkPosition642 := position, thunkPosition
						if !p.rules[ruleHtmlBlockCloseScript]() {
							goto l642
						}
						goto l639
					l642:
						position, thunkPosition = position642, thunkPosition642
					}
					if !matchDot() {
						goto l639
					}
				}
			l640:
				goto l638
			l639:
				position, thunkPosition = position639, thunkPosition639
			}
			if !p.rules[ruleHtmlBlockCloseScript]() {
				goto l637
			}
			return true
		l637:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 129 HtmlBlockInTags <- (HtmlBlockAddress / HtmlBlockBlockquote / HtmlBlockCenter / HtmlBlockDir / HtmlBlockDiv / HtmlBlockDl / HtmlBlockFieldset / HtmlBlockForm / HtmlBlockH1 / HtmlBlockH2 / HtmlBlockH3 / HtmlBlockH4 / HtmlBlockH5 / HtmlBlockH6 / HtmlBlockMenu / HtmlBlockNoframes / HtmlBlockNoscript / HtmlBlockOl / HtmlBlockP / HtmlBlockPre / HtmlBlockTable / HtmlBlockUl / HtmlBlockDd / HtmlBlockDt / HtmlBlockFrameset / HtmlBlockLi / HtmlBlockTbody / HtmlBlockTd / HtmlBlockTfoot / HtmlBlockTh / HtmlBlockThead / HtmlBlockTr / HtmlBlockScript) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position644, thunkPosition644 := position, thunkPosition
				if !p.rules[ruleHtmlBlockAddress]() {
					goto l645
				}
				goto l644
			l645:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockBlockquote]() {
					goto l646
				}
				goto l644
			l646:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockCenter]() {
					goto l647
				}
				goto l644
			l647:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockDir]() {
					goto l648
				}
				goto l644
			l648:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockDiv]() {
					goto l649
				}
				goto l644
			l649:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockDl]() {
					goto l650
				}
				goto l644
			l650:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockFieldset]() {
					goto l651
				}
				goto l644
			l651:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockForm]() {
					goto l652
				}
				goto l644
			l652:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockH1]() {
					goto l653
				}
				goto l644
			l653:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockH2]() {
					goto l654
				}
				goto l644
			l654:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockH3]() {
					goto l655
				}
				goto l644
			l655:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockH4]() {
					goto l656
				}
				goto l644
			l656:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockH5]() {
					goto l657
				}
				goto l644
			l657:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockH6]() {
					goto l658
				}
				goto l644
			l658:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockMenu]() {
					goto l659
				}
				goto l644
			l659:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockNoframes]() {
					goto l660
				}
				goto l644
			l660:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockNoscript]() {
					goto l661
				}
				goto l644
			l661:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockOl]() {
					goto l662
				}
				goto l644
			l662:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockP]() {
					goto l663
				}
				goto l644
			l663:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockPre]() {
					goto l664
				}
				goto l644
			l664:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockTable]() {
					goto l665
				}
				goto l644
			l665:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockUl]() {
					goto l666
				}
				goto l644
			l666:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockDd]() {
					goto l667
				}
				goto l644
			l667:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockDt]() {
					goto l668
				}
				goto l644
			l668:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockFrameset]() {
					goto l669
				}
				goto l644
			l669:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockLi]() {
					goto l670
				}
				goto l644
			l670:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockTbody]() {
					goto l671
				}
				goto l644
			l671:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockTd]() {
					goto l672
				}
				goto l644
			l672:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockTfoot]() {
					goto l673
				}
				goto l644
			l673:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockTh]() {
					goto l674
				}
				goto l644
			l674:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockThead]() {
					goto l675
				}
				goto l644
			l675:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockTr]() {
					goto l676
				}
				goto l644
			l676:
				position, thunkPosition = position644, thunkPosition644
				if !p.rules[ruleHtmlBlockScript]() {
					goto l643
				}
			}
		l644:
			return true
		l643:
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
				goto l677
			}
			begin = position
			{
				position678, thunkPosition678 := position, thunkPosition
				if !p.rules[ruleHtmlBlockInTags]() {
					goto l679
				}
				goto l678
			l679:
				position, thunkPosition = position678, thunkPosition678
				if !p.rules[ruleHtmlComment]() {
					goto l680
				}
				goto l678
			l680:
				position, thunkPosition = position678, thunkPosition678
				if !p.rules[ruleHtmlBlockSelfClosing]() {
					goto l677
				}
			}
		l678:
			end = position
			if !p.rules[ruleBlankLine]() {
				goto l677
			}
		l681:
			{
				position682, thunkPosition682 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l682
				}
				goto l681
			l682:
				position, thunkPosition = position682, thunkPosition682
			}
			do(40)
			return true
		l677:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 131 HtmlBlockSelfClosing <- ('<' Spnl HtmlBlockType Spnl HtmlAttribute* '/' Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l683
			}
			if !p.rules[ruleSpnl]() {
				goto l683
			}
			if !p.rules[ruleHtmlBlockType]() {
				goto l683
			}
			if !p.rules[ruleSpnl]() {
				goto l683
			}
		l684:
			{
				position685, thunkPosition685 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l685
				}
				goto l684
			l685:
				position, thunkPosition = position685, thunkPosition685
			}
			if !matchChar('/') {
				goto l683
			}
			if !p.rules[ruleSpnl]() {
				goto l683
			}
			if !matchChar('>') {
				goto l683
			}
			return true
		l683:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 132 HtmlBlockType <- ('address' / 'blockquote' / 'center' / 'dir' / 'div' / 'dl' / 'fieldset' / 'form' / 'h1' / 'h2' / 'h3' / 'h4' / 'h5' / 'h6' / 'hr' / 'isindex' / 'menu' / 'noframes' / 'noscript' / 'ol' / 'p' / 'pre' / 'table' / 'ul' / 'dd' / 'dt' / 'frameset' / 'li' / 'tbody' / 'td' / 'tfoot' / 'th' / 'thead' / 'tr' / 'script' / 'ADDRESS' / 'BLOCKQUOTE' / 'CENTER' / 'DIR' / 'DIV' / 'DL' / 'FIELDSET' / 'FORM' / 'H1' / 'H2' / 'H3' / 'H4' / 'H5' / 'H6' / 'HR' / 'ISINDEX' / 'MENU' / 'NOFRAMES' / 'NOSCRIPT' / 'OL' / 'P' / 'PRE' / 'TABLE' / 'UL' / 'DD' / 'DT' / 'FRAMESET' / 'LI' / 'TBODY' / 'TD' / 'TFOOT' / 'TH' / 'THEAD' / 'TR' / 'SCRIPT') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position687, thunkPosition687 := position, thunkPosition
				if !matchString("address") {
					goto l688
				}
				goto l687
			l688:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("blockquote") {
					goto l689
				}
				goto l687
			l689:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("center") {
					goto l690
				}
				goto l687
			l690:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("dir") {
					goto l691
				}
				goto l687
			l691:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("div") {
					goto l692
				}
				goto l687
			l692:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("dl") {
					goto l693
				}
				goto l687
			l693:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("fieldset") {
					goto l694
				}
				goto l687
			l694:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("form") {
					goto l695
				}
				goto l687
			l695:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("h1") {
					goto l696
				}
				goto l687
			l696:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("h2") {
					goto l697
				}
				goto l687
			l697:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("h3") {
					goto l698
				}
				goto l687
			l698:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("h4") {
					goto l699
				}
				goto l687
			l699:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("h5") {
					goto l700
				}
				goto l687
			l700:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("h6") {
					goto l701
				}
				goto l687
			l701:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("hr") {
					goto l702
				}
				goto l687
			l702:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("isindex") {
					goto l703
				}
				goto l687
			l703:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("menu") {
					goto l704
				}
				goto l687
			l704:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("noframes") {
					goto l705
				}
				goto l687
			l705:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("noscript") {
					goto l706
				}
				goto l687
			l706:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("ol") {
					goto l707
				}
				goto l687
			l707:
				position, thunkPosition = position687, thunkPosition687
				if !matchChar('p') {
					goto l708
				}
				goto l687
			l708:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("pre") {
					goto l709
				}
				goto l687
			l709:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("table") {
					goto l710
				}
				goto l687
			l710:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("ul") {
					goto l711
				}
				goto l687
			l711:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("dd") {
					goto l712
				}
				goto l687
			l712:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("dt") {
					goto l713
				}
				goto l687
			l713:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("frameset") {
					goto l714
				}
				goto l687
			l714:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("li") {
					goto l715
				}
				goto l687
			l715:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("tbody") {
					goto l716
				}
				goto l687
			l716:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("td") {
					goto l717
				}
				goto l687
			l717:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("tfoot") {
					goto l718
				}
				goto l687
			l718:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("th") {
					goto l719
				}
				goto l687
			l719:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("thead") {
					goto l720
				}
				goto l687
			l720:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("tr") {
					goto l721
				}
				goto l687
			l721:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("script") {
					goto l722
				}
				goto l687
			l722:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("ADDRESS") {
					goto l723
				}
				goto l687
			l723:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("BLOCKQUOTE") {
					goto l724
				}
				goto l687
			l724:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("CENTER") {
					goto l725
				}
				goto l687
			l725:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("DIR") {
					goto l726
				}
				goto l687
			l726:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("DIV") {
					goto l727
				}
				goto l687
			l727:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("DL") {
					goto l728
				}
				goto l687
			l728:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("FIELDSET") {
					goto l729
				}
				goto l687
			l729:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("FORM") {
					goto l730
				}
				goto l687
			l730:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("H1") {
					goto l731
				}
				goto l687
			l731:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("H2") {
					goto l732
				}
				goto l687
			l732:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("H3") {
					goto l733
				}
				goto l687
			l733:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("H4") {
					goto l734
				}
				goto l687
			l734:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("H5") {
					goto l735
				}
				goto l687
			l735:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("H6") {
					goto l736
				}
				goto l687
			l736:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("HR") {
					goto l737
				}
				goto l687
			l737:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("ISINDEX") {
					goto l738
				}
				goto l687
			l738:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("MENU") {
					goto l739
				}
				goto l687
			l739:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("NOFRAMES") {
					goto l740
				}
				goto l687
			l740:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("NOSCRIPT") {
					goto l741
				}
				goto l687
			l741:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("OL") {
					goto l742
				}
				goto l687
			l742:
				position, thunkPosition = position687, thunkPosition687
				if !matchChar('P') {
					goto l743
				}
				goto l687
			l743:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("PRE") {
					goto l744
				}
				goto l687
			l744:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("TABLE") {
					goto l745
				}
				goto l687
			l745:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("UL") {
					goto l746
				}
				goto l687
			l746:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("DD") {
					goto l747
				}
				goto l687
			l747:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("DT") {
					goto l748
				}
				goto l687
			l748:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("FRAMESET") {
					goto l749
				}
				goto l687
			l749:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("LI") {
					goto l750
				}
				goto l687
			l750:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("TBODY") {
					goto l751
				}
				goto l687
			l751:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("TD") {
					goto l752
				}
				goto l687
			l752:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("TFOOT") {
					goto l753
				}
				goto l687
			l753:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("TH") {
					goto l754
				}
				goto l687
			l754:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("THEAD") {
					goto l755
				}
				goto l687
			l755:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("TR") {
					goto l756
				}
				goto l687
			l756:
				position, thunkPosition = position687, thunkPosition687
				if !matchString("SCRIPT") {
					goto l686
				}
			}
		l687:
			return true
		l686:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 133 StyleOpen <- ('<' Spnl ('style' / 'STYLE') Spnl HtmlAttribute* '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l757
			}
			if !p.rules[ruleSpnl]() {
				goto l757
			}
			{
				position758, thunkPosition758 := position, thunkPosition
				if !matchString("style") {
					goto l759
				}
				goto l758
			l759:
				position, thunkPosition = position758, thunkPosition758
				if !matchString("STYLE") {
					goto l757
				}
			}
		l758:
			if !p.rules[ruleSpnl]() {
				goto l757
			}
		l760:
			{
				position761, thunkPosition761 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l761
				}
				goto l760
			l761:
				position, thunkPosition = position761, thunkPosition761
			}
			if !matchChar('>') {
				goto l757
			}
			return true
		l757:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 134 StyleClose <- ('<' Spnl '/' ('style' / 'STYLE') Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l762
			}
			if !p.rules[ruleSpnl]() {
				goto l762
			}
			if !matchChar('/') {
				goto l762
			}
			{
				position763, thunkPosition763 := position, thunkPosition
				if !matchString("style") {
					goto l764
				}
				goto l763
			l764:
				position, thunkPosition = position763, thunkPosition763
				if !matchString("STYLE") {
					goto l762
				}
			}
		l763:
			if !p.rules[ruleSpnl]() {
				goto l762
			}
			if !matchChar('>') {
				goto l762
			}
			return true
		l762:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 135 InStyleTags <- (StyleOpen (!StyleClose .)* StyleClose) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleStyleOpen]() {
				goto l765
			}
		l766:
			{
				position767, thunkPosition767 := position, thunkPosition
				{
					position768, thunkPosition768 := position, thunkPosition
					if !p.rules[ruleStyleClose]() {
						goto l768
					}
					goto l767
				l768:
					position, thunkPosition = position768, thunkPosition768
				}
				if !matchDot() {
					goto l767
				}
				goto l766
			l767:
				position, thunkPosition = position767, thunkPosition767
			}
			if !p.rules[ruleStyleClose]() {
				goto l765
			}
			return true
		l765:
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
				goto l769
			}
			end = position
		l770:
			{
				position771, thunkPosition771 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l771
				}
				goto l770
			l771:
				position, thunkPosition = position771, thunkPosition771
			}
			do(41)
			return true
		l769:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 137 Inlines <- (StartList ((!Endline Inline { a = cons(yy, a) }) / (Endline &Inline { a = cons(c, a) }))+ Endline? { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l772
			}
			doarg(yySet, -2)
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
					goto l772
				}
				doarg(yySet, -1)
				{
					position778, thunkPosition778 := position, thunkPosition
					if !p.rules[ruleInline]() {
						goto l772
					}
					position, thunkPosition = position778, thunkPosition778
				}
				do(43)
			}
		l775:
		l773:
			{
				position774, thunkPosition774 := position, thunkPosition
				{
					position779, thunkPosition779 := position, thunkPosition
					{
						position781, thunkPosition781 := position, thunkPosition
						if !p.rules[ruleEndline]() {
							goto l781
						}
						goto l780
					l781:
						position, thunkPosition = position781, thunkPosition781
					}
					if !p.rules[ruleInline]() {
						goto l780
					}
					do(42)
					goto l779
				l780:
					position, thunkPosition = position779, thunkPosition779
					if !p.rules[ruleEndline]() {
						goto l774
					}
					doarg(yySet, -1)
					{
						position782, thunkPosition782 := position, thunkPosition
						if !p.rules[ruleInline]() {
							goto l774
						}
						position, thunkPosition = position782, thunkPosition782
					}
					do(43)
				}
			l779:
				goto l773
			l774:
				position, thunkPosition = position774, thunkPosition774
			}
			{
				position783, thunkPosition783 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l783
				}
				goto l784
			l783:
				position, thunkPosition = position783, thunkPosition783
			}
		l784:
			do(44)
			doarg(yyPop, 2)
			return true
		l772:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 138 Inline <- (Str / Endline / UlOrStarLine / Space / Strong / Emph / Image / Link / NoteReference / InlineNote / Code / RawHtml / Entity / EscapedChar / Smart / Symbol) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position786, thunkPosition786 := position, thunkPosition
				if !p.rules[ruleStr]() {
					goto l787
				}
				goto l786
			l787:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleEndline]() {
					goto l788
				}
				goto l786
			l788:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleUlOrStarLine]() {
					goto l789
				}
				goto l786
			l789:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleSpace]() {
					goto l790
				}
				goto l786
			l790:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleStrong]() {
					goto l791
				}
				goto l786
			l791:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleEmph]() {
					goto l792
				}
				goto l786
			l792:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleImage]() {
					goto l793
				}
				goto l786
			l793:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleLink]() {
					goto l794
				}
				goto l786
			l794:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleNoteReference]() {
					goto l795
				}
				goto l786
			l795:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleInlineNote]() {
					goto l796
				}
				goto l786
			l796:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleCode]() {
					goto l797
				}
				goto l786
			l797:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleRawHtml]() {
					goto l798
				}
				goto l786
			l798:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleEntity]() {
					goto l799
				}
				goto l786
			l799:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleEscapedChar]() {
					goto l800
				}
				goto l786
			l800:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleSmart]() {
					goto l801
				}
				goto l786
			l801:
				position, thunkPosition = position786, thunkPosition786
				if !p.rules[ruleSymbol]() {
					goto l785
				}
			}
		l786:
			return true
		l785:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 139 Space <- (Spacechar+ { yy = mk_str(" ")
          yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSpacechar]() {
				goto l802
			}
		l803:
			{
				position804, thunkPosition804 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l804
				}
				goto l803
			l804:
				position, thunkPosition = position804, thunkPosition804
			}
			do(45)
			return true
		l802:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 140 Str <- (< NormalChar (NormalChar / ('_'+ &Alphanumeric))* > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleNormalChar]() {
				goto l805
			}
		l806:
			{
				position807, thunkPosition807 := position, thunkPosition
				{
					position808, thunkPosition808 := position, thunkPosition
					if !p.rules[ruleNormalChar]() {
						goto l809
					}
					goto l808
				l809:
					position, thunkPosition = position808, thunkPosition808
					if !matchChar('_') {
						goto l807
					}
				l810:
					{
						position811, thunkPosition811 := position, thunkPosition
						if !matchChar('_') {
							goto l811
						}
						goto l810
					l811:
						position, thunkPosition = position811, thunkPosition811
					}
					{
						position812, thunkPosition812 := position, thunkPosition
						if !p.rules[ruleAlphanumeric]() {
							goto l807
						}
						position, thunkPosition = position812, thunkPosition812
					}
				}
			l808:
				goto l806
			l807:
				position, thunkPosition = position807, thunkPosition807
			}
			end = position
			do(46)
			return true
		l805:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 141 EscapedChar <- ('\\' !Newline < [-\\`|*_{}[\]()#+.!><] > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\\') {
				goto l813
			}
			{
				position814, thunkPosition814 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l814
				}
				goto l813
			l814:
				position, thunkPosition = position814, thunkPosition814
			}
			begin = position
			if !matchClass(2) {
				goto l813
			}
			end = position
			do(47)
			return true
		l813:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 142 Entity <- ((HexEntity / DecEntity / CharEntity) { yy = mk_str(yytext); yy.key = HTML }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position816, thunkPosition816 := position, thunkPosition
				if !p.rules[ruleHexEntity]() {
					goto l817
				}
				goto l816
			l817:
				position, thunkPosition = position816, thunkPosition816
				if !p.rules[ruleDecEntity]() {
					goto l818
				}
				goto l816
			l818:
				position, thunkPosition = position816, thunkPosition816
				if !p.rules[ruleCharEntity]() {
					goto l815
				}
			}
		l816:
			do(48)
			return true
		l815:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 143 Endline <- (LineBreak / TerminalEndline / NormalEndline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position820, thunkPosition820 := position, thunkPosition
				if !p.rules[ruleLineBreak]() {
					goto l821
				}
				goto l820
			l821:
				position, thunkPosition = position820, thunkPosition820
				if !p.rules[ruleTerminalEndline]() {
					goto l822
				}
				goto l820
			l822:
				position, thunkPosition = position820, thunkPosition820
				if !p.rules[ruleNormalEndline]() {
					goto l819
				}
			}
		l820:
			return true
		l819:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 144 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line (('===' '='*) / ('---' '-'*)) Newline) { yy = mk_str("\n")
                    yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l823
			}
			if !p.rules[ruleNewline]() {
				goto l823
			}
			{
				position824, thunkPosition824 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l824
				}
				goto l823
			l824:
				position, thunkPosition = position824, thunkPosition824
			}
			if peekChar('>') {
				goto l823
			}
			{
				position825, thunkPosition825 := position, thunkPosition
				if !p.rules[ruleAtxStart]() {
					goto l825
				}
				goto l823
			l825:
				position, thunkPosition = position825, thunkPosition825
			}
			{
				position826, thunkPosition826 := position, thunkPosition
				if !p.rules[ruleLine]() {
					goto l826
				}
				{
					position827, thunkPosition827 := position, thunkPosition
					if !matchString("===") {
						goto l828
					}
				l829:
					{
						position830, thunkPosition830 := position, thunkPosition
						if !matchChar('=') {
							goto l830
						}
						goto l829
					l830:
						position, thunkPosition = position830, thunkPosition830
					}
					goto l827
				l828:
					position, thunkPosition = position827, thunkPosition827
					if !matchString("---") {
						goto l826
					}
				l831:
					{
						position832, thunkPosition832 := position, thunkPosition
						if !matchChar('-') {
							goto l832
						}
						goto l831
					l832:
						position, thunkPosition = position832, thunkPosition832
					}
				}
			l827:
				if !p.rules[ruleNewline]() {
					goto l826
				}
				goto l823
			l826:
				position, thunkPosition = position826, thunkPosition826
			}
			do(49)
			return true
		l823:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 145 TerminalEndline <- (Sp Newline Eof { yy = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l833
			}
			if !p.rules[ruleNewline]() {
				goto l833
			}
			if !p.rules[ruleEof]() {
				goto l833
			}
			do(50)
			return true
		l833:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 146 LineBreak <- ('  ' NormalEndline { yy = mk_element(LINEBREAK) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("  ") {
				goto l834
			}
			if !p.rules[ruleNormalEndline]() {
				goto l834
			}
			do(51)
			return true
		l834:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 147 Symbol <- (< SpecialChar > { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleSpecialChar]() {
				goto l835
			}
			end = position
			do(52)
			return true
		l835:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 148 UlOrStarLine <- ((UlLine / StarLine) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position837, thunkPosition837 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l838
				}
				goto l837
			l838:
				position, thunkPosition = position837, thunkPosition837
				if !p.rules[ruleStarLine]() {
					goto l836
				}
			}
		l837:
			do(53)
			return true
		l836:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 149 StarLine <- ((< '****' '*'* >) / (< Spacechar '*'+ &Spacechar >)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position840, thunkPosition840 := position, thunkPosition
				begin = position
				if !matchString("****") {
					goto l841
				}
			l842:
				{
					position843, thunkPosition843 := position, thunkPosition
					if !matchChar('*') {
						goto l843
					}
					goto l842
				l843:
					position, thunkPosition = position843, thunkPosition843
				}
				end = position
				goto l840
			l841:
				position, thunkPosition = position840, thunkPosition840
				begin = position
				if !p.rules[ruleSpacechar]() {
					goto l839
				}
				if !matchChar('*') {
					goto l839
				}
			l844:
				{
					position845, thunkPosition845 := position, thunkPosition
					if !matchChar('*') {
						goto l845
					}
					goto l844
				l845:
					position, thunkPosition = position845, thunkPosition845
				}
				{
					position846, thunkPosition846 := position, thunkPosition
					if !p.rules[ruleSpacechar]() {
						goto l839
					}
					position, thunkPosition = position846, thunkPosition846
				}
				end = position
			}
		l840:
			return true
		l839:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 150 UlLine <- ((< '____' '_'* >) / (< Spacechar '_'+ &Spacechar >)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position848, thunkPosition848 := position, thunkPosition
				begin = position
				if !matchString("____") {
					goto l849
				}
			l850:
				{
					position851, thunkPosition851 := position, thunkPosition
					if !matchChar('_') {
						goto l851
					}
					goto l850
				l851:
					position, thunkPosition = position851, thunkPosition851
				}
				end = position
				goto l848
			l849:
				position, thunkPosition = position848, thunkPosition848
				begin = position
				if !p.rules[ruleSpacechar]() {
					goto l847
				}
				if !matchChar('_') {
					goto l847
				}
			l852:
				{
					position853, thunkPosition853 := position, thunkPosition
					if !matchChar('_') {
						goto l853
					}
					goto l852
				l853:
					position, thunkPosition = position853, thunkPosition853
				}
				{
					position854, thunkPosition854 := position, thunkPosition
					if !p.rules[ruleSpacechar]() {
						goto l847
					}
					position, thunkPosition = position854, thunkPosition854
				}
				end = position
			}
		l848:
			return true
		l847:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 151 Emph <- (EmphStar / EmphUl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position856, thunkPosition856 := position, thunkPosition
				if !p.rules[ruleEmphStar]() {
					goto l857
				}
				goto l856
			l857:
				position, thunkPosition = position856, thunkPosition856
				if !p.rules[ruleEmphUl]() {
					goto l855
				}
			}
		l856:
			return true
		l855:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 152 OneStarOpen <- (!StarLine '*' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position859, thunkPosition859 := position, thunkPosition
				if !p.rules[ruleStarLine]() {
					goto l859
				}
				goto l858
			l859:
				position, thunkPosition = position859, thunkPosition859
			}
			if !matchChar('*') {
				goto l858
			}
			{
				position860, thunkPosition860 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l860
				}
				goto l858
			l860:
				position, thunkPosition = position860, thunkPosition860
			}
			{
				position861, thunkPosition861 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l861
				}
				goto l858
			l861:
				position, thunkPosition = position861, thunkPosition861
			}
			return true
		l858:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 153 OneStarClose <- (!Spacechar !Newline Inline !StrongStar '*' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position863, thunkPosition863 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l863
				}
				goto l862
			l863:
				position, thunkPosition = position863, thunkPosition863
			}
			{
				position864, thunkPosition864 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l864
				}
				goto l862
			l864:
				position, thunkPosition = position864, thunkPosition864
			}
			if !p.rules[ruleInline]() {
				goto l862
			}
			doarg(yySet, -1)
			{
				position865, thunkPosition865 := position, thunkPosition
				if !p.rules[ruleStrongStar]() {
					goto l865
				}
				goto l862
			l865:
				position, thunkPosition = position865, thunkPosition865
			}
			if !matchChar('*') {
				goto l862
			}
			do(54)
			doarg(yyPop, 1)
			return true
		l862:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 154 EmphStar <- (OneStarOpen StartList (!OneStarClose Inline { a = cons(yy, a) })* OneStarClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneStarOpen]() {
				goto l866
			}
			if !p.rules[ruleStartList]() {
				goto l866
			}
			doarg(yySet, -1)
		l867:
			{
				position868, thunkPosition868 := position, thunkPosition
				{
					position869, thunkPosition869 := position, thunkPosition
					if !p.rules[ruleOneStarClose]() {
						goto l869
					}
					goto l868
				l869:
					position, thunkPosition = position869, thunkPosition869
				}
				if !p.rules[ruleInline]() {
					goto l868
				}
				do(55)
				goto l867
			l868:
				position, thunkPosition = position868, thunkPosition868
			}
			if !p.rules[ruleOneStarClose]() {
				goto l866
			}
			do(56)
			do(57)
			doarg(yyPop, 1)
			return true
		l866:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 155 OneUlOpen <- (!UlLine '_' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position871, thunkPosition871 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l871
				}
				goto l870
			l871:
				position, thunkPosition = position871, thunkPosition871
			}
			if !matchChar('_') {
				goto l870
			}
			{
				position872, thunkPosition872 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l872
				}
				goto l870
			l872:
				position, thunkPosition = position872, thunkPosition872
			}
			{
				position873, thunkPosition873 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l873
				}
				goto l870
			l873:
				position, thunkPosition = position873, thunkPosition873
			}
			return true
		l870:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 156 OneUlClose <- (!Spacechar !Newline Inline !StrongUl '_' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position875, thunkPosition875 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l875
				}
				goto l874
			l875:
				position, thunkPosition = position875, thunkPosition875
			}
			{
				position876, thunkPosition876 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l876
				}
				goto l874
			l876:
				position, thunkPosition = position876, thunkPosition876
			}
			if !p.rules[ruleInline]() {
				goto l874
			}
			doarg(yySet, -1)
			{
				position877, thunkPosition877 := position, thunkPosition
				if !p.rules[ruleStrongUl]() {
					goto l877
				}
				goto l874
			l877:
				position, thunkPosition = position877, thunkPosition877
			}
			if !matchChar('_') {
				goto l874
			}
			{
				position878, thunkPosition878 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l878
				}
				goto l874
			l878:
				position, thunkPosition = position878, thunkPosition878
			}
			do(58)
			doarg(yyPop, 1)
			return true
		l874:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 157 EmphUl <- (OneUlOpen StartList (!OneUlClose Inline { a = cons(yy, a) })* OneUlClose { a = cons(yy, a) } { yy = mk_list(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleOneUlOpen]() {
				goto l879
			}
			if !p.rules[ruleStartList]() {
				goto l879
			}
			doarg(yySet, -1)
		l880:
			{
				position881, thunkPosition881 := position, thunkPosition
				{
					position882, thunkPosition882 := position, thunkPosition
					if !p.rules[ruleOneUlClose]() {
						goto l882
					}
					goto l881
				l882:
					position, thunkPosition = position882, thunkPosition882
				}
				if !p.rules[ruleInline]() {
					goto l881
				}
				do(59)
				goto l880
			l881:
				position, thunkPosition = position881, thunkPosition881
			}
			if !p.rules[ruleOneUlClose]() {
				goto l879
			}
			do(60)
			do(61)
			doarg(yyPop, 1)
			return true
		l879:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 158 Strong <- (StrongStar / StrongUl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position884, thunkPosition884 := position, thunkPosition
				if !p.rules[ruleStrongStar]() {
					goto l885
				}
				goto l884
			l885:
				position, thunkPosition = position884, thunkPosition884
				if !p.rules[ruleStrongUl]() {
					goto l883
				}
			}
		l884:
			return true
		l883:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 159 TwoStarOpen <- (!StarLine '**' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position887, thunkPosition887 := position, thunkPosition
				if !p.rules[ruleStarLine]() {
					goto l887
				}
				goto l886
			l887:
				position, thunkPosition = position887, thunkPosition887
			}
			if !matchString("**") {
				goto l886
			}
			{
				position888, thunkPosition888 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l888
				}
				goto l886
			l888:
				position, thunkPosition = position888, thunkPosition888
			}
			{
				position889, thunkPosition889 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l889
				}
				goto l886
			l889:
				position, thunkPosition = position889, thunkPosition889
			}
			return true
		l886:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 160 TwoStarClose <- (!Spacechar !Newline Inline '**' { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position891, thunkPosition891 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l891
				}
				goto l890
			l891:
				position, thunkPosition = position891, thunkPosition891
			}
			{
				position892, thunkPosition892 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l892
				}
				goto l890
			l892:
				position, thunkPosition = position892, thunkPosition892
			}
			if !p.rules[ruleInline]() {
				goto l890
			}
			doarg(yySet, -1)
			if !matchString("**") {
				goto l890
			}
			do(62)
			doarg(yyPop, 1)
			return true
		l890:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 161 StrongStar <- (TwoStarOpen StartList (!TwoStarClose Inline { a = cons(yy, a) })* TwoStarClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoStarOpen]() {
				goto l893
			}
			if !p.rules[ruleStartList]() {
				goto l893
			}
			doarg(yySet, -1)
		l894:
			{
				position895, thunkPosition895 := position, thunkPosition
				{
					position896, thunkPosition896 := position, thunkPosition
					if !p.rules[ruleTwoStarClose]() {
						goto l896
					}
					goto l895
				l896:
					position, thunkPosition = position896, thunkPosition896
				}
				if !p.rules[ruleInline]() {
					goto l895
				}
				do(63)
				goto l894
			l895:
				position, thunkPosition = position895, thunkPosition895
			}
			if !p.rules[ruleTwoStarClose]() {
				goto l893
			}
			do(64)
			do(65)
			doarg(yyPop, 1)
			return true
		l893:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 162 TwoUlOpen <- (!UlLine '__' !Spacechar !Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position898, thunkPosition898 := position, thunkPosition
				if !p.rules[ruleUlLine]() {
					goto l898
				}
				goto l897
			l898:
				position, thunkPosition = position898, thunkPosition898
			}
			if !matchString("__") {
				goto l897
			}
			{
				position899, thunkPosition899 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l899
				}
				goto l897
			l899:
				position, thunkPosition = position899, thunkPosition899
			}
			{
				position900, thunkPosition900 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l900
				}
				goto l897
			l900:
				position, thunkPosition = position900, thunkPosition900
			}
			return true
		l897:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 163 TwoUlClose <- (!Spacechar !Newline Inline '__' !Alphanumeric { yy = a }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position902, thunkPosition902 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l902
				}
				goto l901
			l902:
				position, thunkPosition = position902, thunkPosition902
			}
			{
				position903, thunkPosition903 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l903
				}
				goto l901
			l903:
				position, thunkPosition = position903, thunkPosition903
			}
			if !p.rules[ruleInline]() {
				goto l901
			}
			doarg(yySet, -1)
			if !matchString("__") {
				goto l901
			}
			{
				position904, thunkPosition904 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l904
				}
				goto l901
			l904:
				position, thunkPosition = position904, thunkPosition904
			}
			do(66)
			doarg(yyPop, 1)
			return true
		l901:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 164 StrongUl <- (TwoUlOpen StartList (!TwoUlClose Inline { a = cons(yy, a) })* TwoUlClose { a = cons(yy, a) } { yy = mk_list(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleTwoUlOpen]() {
				goto l905
			}
			if !p.rules[ruleStartList]() {
				goto l905
			}
			doarg(yySet, -1)
		l906:
			{
				position907, thunkPosition907 := position, thunkPosition
				{
					position908, thunkPosition908 := position, thunkPosition
					if !p.rules[ruleTwoUlClose]() {
						goto l908
					}
					goto l907
				l908:
					position, thunkPosition = position908, thunkPosition908
				}
				if !p.rules[ruleInline]() {
					goto l907
				}
				do(67)
				goto l906
			l907:
				position, thunkPosition = position907, thunkPosition907
			}
			if !p.rules[ruleTwoUlClose]() {
				goto l905
			}
			do(68)
			do(69)
			doarg(yyPop, 1)
			return true
		l905:
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
				goto l909
			}
			{
				position910, thunkPosition910 := position, thunkPosition
				if !p.rules[ruleExplicitLink]() {
					goto l911
				}
				goto l910
			l911:
				position, thunkPosition = position910, thunkPosition910
				if !p.rules[ruleReferenceLink]() {
					goto l909
				}
			}
		l910:
			do(70)
			return true
		l909:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 166 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position913, thunkPosition913 := position, thunkPosition
				if !p.rules[ruleExplicitLink]() {
					goto l914
				}
				goto l913
			l914:
				position, thunkPosition = position913, thunkPosition913
				if !p.rules[ruleReferenceLink]() {
					goto l915
				}
				goto l913
			l915:
				position, thunkPosition = position913, thunkPosition913
				if !p.rules[ruleAutoLink]() {
					goto l912
				}
			}
		l913:
			return true
		l912:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 167 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position917, thunkPosition917 := position, thunkPosition
				if !p.rules[ruleReferenceLinkDouble]() {
					goto l918
				}
				goto l917
			l918:
				position, thunkPosition = position917, thunkPosition917
				if !p.rules[ruleReferenceLinkSingle]() {
					goto l916
				}
			}
		l917:
			return true
		l916:
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
				goto l919
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto l919
			}
			end = position
			{
				position920, thunkPosition920 := position, thunkPosition
				if !matchString("[]") {
					goto l920
				}
				goto l919
			l920:
				position, thunkPosition = position920, thunkPosition920
			}
			if !p.rules[ruleLabel]() {
				goto l919
			}
			doarg(yySet, -2)
			do(71)
			doarg(yyPop, 2)
			return true
		l919:
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
				goto l921
			}
			doarg(yySet, -1)
			begin = position
			{
				position922, thunkPosition922 := position, thunkPosition
				if !p.rules[ruleSpnl]() {
					goto l922
				}
				if !matchString("[]") {
					goto l922
				}
				goto l923
			l922:
				position, thunkPosition = position922, thunkPosition922
			}
		l923:
			end = position
			do(72)
			doarg(yyPop, 1)
			return true
		l921:
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
				goto l924
			}
			doarg(yySet, -2)
			if !p.rules[ruleSpnl]() {
				goto l924
			}
			if !matchChar('(') {
				goto l924
			}
			if !p.rules[ruleSp]() {
				goto l924
			}
			if !p.rules[ruleSource]() {
				goto l924
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l924
			}
			if !p.rules[ruleTitle]() {
				goto l924
			}
			doarg(yySet, -3)
			if !p.rules[ruleSp]() {
				goto l924
			}
			if !matchChar(')') {
				goto l924
			}
			do(73)
			doarg(yyPop, 3)
			return true
		l924:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 171 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position926, thunkPosition926 := position, thunkPosition
				if !matchChar('<') {
					goto l927
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l927
				}
				end = position
				if !matchChar('>') {
					goto l927
				}
				goto l926
			l927:
				position, thunkPosition = position926, thunkPosition926
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l925
				}
				end = position
			}
		l926:
			do(74)
			return true
		l925:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 172 SourceContents <- (((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* / '') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position929, thunkPosition929 := position, thunkPosition
			l931:
				{
					position932, thunkPosition932 := position, thunkPosition
					{
						position933, thunkPosition933 := position, thunkPosition
						if peekChar('(') {
							goto l934
						}
						if peekChar(')') {
							goto l934
						}
						if peekChar('>') {
							goto l934
						}
						if !p.rules[ruleNonspacechar]() {
							goto l934
						}
					l935:
						{
							position936, thunkPosition936 := position, thunkPosition
							if peekChar('(') {
								goto l936
							}
							if peekChar(')') {
								goto l936
							}
							if peekChar('>') {
								goto l936
							}
							if !p.rules[ruleNonspacechar]() {
								goto l936
							}
							goto l935
						l936:
							position, thunkPosition = position936, thunkPosition936
						}
						goto l933
					l934:
						position, thunkPosition = position933, thunkPosition933
						if !matchChar('(') {
							goto l932
						}
						if !p.rules[ruleSourceContents]() {
							goto l932
						}
						if !matchChar(')') {
							goto l932
						}
					}
				l933:
					goto l931
				l932:
					position, thunkPosition = position932, thunkPosition932
				}
				goto l929
			l930:
				position, thunkPosition = position929, thunkPosition929
				if !matchString("") {
					goto l928
				}
			}
		l929:
			return true
		l928:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 173 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position938, thunkPosition938 := position, thunkPosition
				if !p.rules[ruleTitleSingle]() {
					goto l939
				}
				goto l938
			l939:
				position, thunkPosition = position938, thunkPosition938
				if !p.rules[ruleTitleDouble]() {
					goto l940
				}
				goto l938
			l940:
				position, thunkPosition = position938, thunkPosition938
				begin = position
				if !matchString("") {
					goto l937
				}
				end = position
			}
		l938:
			do(75)
			return true
		l937:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 174 TitleSingle <- ('\'' < (!('\'' Sp (')' / Newline)) .)* > '\'') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l941
			}
			begin = position
		l942:
			{
				position943, thunkPosition943 := position, thunkPosition
				{
					position944, thunkPosition944 := position, thunkPosition
					if !matchChar('\'') {
						goto l944
					}
					if !p.rules[ruleSp]() {
						goto l944
					}
					{
						position945, thunkPosition945 := position, thunkPosition
						if !matchChar(')') {
							goto l946
						}
						goto l945
					l946:
						position, thunkPosition = position945, thunkPosition945
						if !p.rules[ruleNewline]() {
							goto l944
						}
					}
				l945:
					goto l943
				l944:
					position, thunkPosition = position944, thunkPosition944
				}
				if !matchDot() {
					goto l943
				}
				goto l942
			l943:
				position, thunkPosition = position943, thunkPosition943
			}
			end = position
			if !matchChar('\'') {
				goto l941
			}
			return true
		l941:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 175 TitleDouble <- ('"' < (!('"' Sp (')' / Newline)) .)* > '"') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l947
			}
			begin = position
		l948:
			{
				position949, thunkPosition949 := position, thunkPosition
				{
					position950, thunkPosition950 := position, thunkPosition
					if !matchChar('"') {
						goto l950
					}
					if !p.rules[ruleSp]() {
						goto l950
					}
					{
						position951, thunkPosition951 := position, thunkPosition
						if !matchChar(')') {
							goto l952
						}
						goto l951
					l952:
						position, thunkPosition = position951, thunkPosition951
						if !p.rules[ruleNewline]() {
							goto l950
						}
					}
				l951:
					goto l949
				l950:
					position, thunkPosition = position950, thunkPosition950
				}
				if !matchDot() {
					goto l949
				}
				goto l948
			l949:
				position, thunkPosition = position949, thunkPosition949
			}
			end = position
			if !matchChar('"') {
				goto l947
			}
			return true
		l947:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 176 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position954, thunkPosition954 := position, thunkPosition
				if !p.rules[ruleAutoLinkUrl]() {
					goto l955
				}
				goto l954
			l955:
				position, thunkPosition = position954, thunkPosition954
				if !p.rules[ruleAutoLinkEmail]() {
					goto l953
				}
			}
		l954:
			return true
		l953:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 177 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = mk_link(mk_str(yytext), yytext, "") }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l956
			}
			begin = position
			if !matchClass(4) {
				goto l956
			}
		l957:
			{
				position958, thunkPosition958 := position, thunkPosition
				if !matchClass(4) {
					goto l958
				}
				goto l957
			l958:
				position, thunkPosition = position958, thunkPosition958
			}
			if !matchString("://") {
				goto l956
			}
			{
				position961, thunkPosition961 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l961
				}
				goto l956
			l961:
				position, thunkPosition = position961, thunkPosition961
			}
			if peekChar('>') {
				goto l956
			}
			if !matchDot() {
				goto l956
			}
		l959:
			{
				position960, thunkPosition960 := position, thunkPosition
				{
					position962, thunkPosition962 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l962
					}
					goto l960
				l962:
					position, thunkPosition = position962, thunkPosition962
				}
				if peekChar('>') {
					goto l960
				}
				if !matchDot() {
					goto l960
				}
				goto l959
			l960:
				position, thunkPosition = position960, thunkPosition960
			}
			end = position
			if !matchChar('>') {
				goto l956
			}
			do(76)
			return true
		l956:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 178 AutoLinkEmail <- ('<' < [-A-Za-z0-9+_]+ '@' (!Newline !'>' .)+ > '>' {
                    yy = mk_link(mk_str(yytext), "mailto:"+yytext, "")
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l963
			}
			begin = position
			if !matchClass(9) {
				goto l963
			}
		l964:
			{
				position965, thunkPosition965 := position, thunkPosition
				if !matchClass(9) {
					goto l965
				}
				goto l964
			l965:
				position, thunkPosition = position965, thunkPosition965
			}
			if !matchChar('@') {
				goto l963
			}
			{
				position968, thunkPosition968 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l968
				}
				goto l963
			l968:
				position, thunkPosition = position968, thunkPosition968
			}
			if peekChar('>') {
				goto l963
			}
			if !matchDot() {
				goto l963
			}
		l966:
			{
				position967, thunkPosition967 := position, thunkPosition
				{
					position969, thunkPosition969 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l969
					}
					goto l967
				l969:
					position, thunkPosition = position969, thunkPosition969
				}
				if peekChar('>') {
					goto l967
				}
				if !matchDot() {
					goto l967
				}
				goto l966
			l967:
				position, thunkPosition = position967, thunkPosition967
			}
			end = position
			if !matchChar('>') {
				goto l963
			}
			do(77)
			return true
		l963:
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
				goto l970
			}
			{
				position971, thunkPosition971 := position, thunkPosition
				if !matchString("[]") {
					goto l971
				}
				goto l970
			l971:
				position, thunkPosition = position971, thunkPosition971
			}
			if !p.rules[ruleLabel]() {
				goto l970
			}
			doarg(yySet, -2)
			if !matchChar(':') {
				goto l970
			}
			if !p.rules[ruleSpnl]() {
				goto l970
			}
			if !p.rules[ruleRefSrc]() {
				goto l970
			}
			doarg(yySet, -1)
			if !p.rules[ruleSpnl]() {
				goto l970
			}
			if !p.rules[ruleRefTitle]() {
				goto l970
			}
			doarg(yySet, -3)
		l972:
			{
				position973, thunkPosition973 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l973
				}
				goto l972
			l973:
				position, thunkPosition = position973, thunkPosition973
			}
			do(78)
			doarg(yyPop, 3)
			return true
		l970:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 180 Label <- ('[' ((!'^' &{ p.extension.Notes }) / (&. &{ !p.extension.Notes })) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = mk_list(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l974
			}
			{
				position975, thunkPosition975 := position, thunkPosition
				if peekChar('^') {
					goto l976
				}
				if !( p.extension.Notes ) {
					goto l976
				}
				goto l975
			l976:
				position, thunkPosition = position975, thunkPosition975
				if !peekDot() {
					goto l974
				}
				if !( !p.extension.Notes ) {
					goto l974
				}
			}
		l975:
			if !p.rules[ruleStartList]() {
				goto l974
			}
			doarg(yySet, -1)
		l977:
			{
				position978, thunkPosition978 := position, thunkPosition
				if peekChar(']') {
					goto l978
				}
				if !p.rules[ruleInline]() {
					goto l978
				}
				do(79)
				goto l977
			l978:
				position, thunkPosition = position978, thunkPosition978
			}
			if !matchChar(']') {
				goto l974
			}
			do(80)
			doarg(yyPop, 1)
			return true
		l974:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 181 RefSrc <- (< Nonspacechar+ > { yy = mk_str(yytext)
           yy.key = HTML }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto l979
			}
		l980:
			{
				position981, thunkPosition981 := position, thunkPosition
				if !p.rules[ruleNonspacechar]() {
					goto l981
				}
				goto l980
			l981:
				position, thunkPosition = position981, thunkPosition981
			}
			end = position
			do(81)
			return true
		l979:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 182 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position983, thunkPosition983 := position, thunkPosition
				if !p.rules[ruleRefTitleSingle]() {
					goto l984
				}
				goto l983
			l984:
				position, thunkPosition = position983, thunkPosition983
				if !p.rules[ruleRefTitleDouble]() {
					goto l985
				}
				goto l983
			l985:
				position, thunkPosition = position983, thunkPosition983
				if !p.rules[ruleRefTitleParens]() {
					goto l986
				}
				goto l983
			l986:
				position, thunkPosition = position983, thunkPosition983
				if !p.rules[ruleEmptyTitle]() {
					goto l982
				}
			}
		l983:
			do(82)
			return true
		l982:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 183 EmptyTitle <- (< '' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchString("") {
				goto l987
			}
			end = position
			return true
		l987:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 184 RefTitleSingle <- ('\'' < (!(('\'' Sp Newline) / Newline) .)* > '\'') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l988
			}
			begin = position
		l989:
			{
				position990, thunkPosition990 := position, thunkPosition
				{
					position991, thunkPosition991 := position, thunkPosition
					{
						position992, thunkPosition992 := position, thunkPosition
						if !matchChar('\'') {
							goto l993
						}
						if !p.rules[ruleSp]() {
							goto l993
						}
						if !p.rules[ruleNewline]() {
							goto l993
						}
						goto l992
					l993:
						position, thunkPosition = position992, thunkPosition992
						if !p.rules[ruleNewline]() {
							goto l991
						}
					}
				l992:
					goto l990
				l991:
					position, thunkPosition = position991, thunkPosition991
				}
				if !matchDot() {
					goto l990
				}
				goto l989
			l990:
				position, thunkPosition = position990, thunkPosition990
			}
			end = position
			if !matchChar('\'') {
				goto l988
			}
			return true
		l988:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 185 RefTitleDouble <- ('"' < (!(('"' Sp Newline) / Newline) .)* > '"') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l994
			}
			begin = position
		l995:
			{
				position996, thunkPosition996 := position, thunkPosition
				{
					position997, thunkPosition997 := position, thunkPosition
					{
						position998, thunkPosition998 := position, thunkPosition
						if !matchChar('"') {
							goto l999
						}
						if !p.rules[ruleSp]() {
							goto l999
						}
						if !p.rules[ruleNewline]() {
							goto l999
						}
						goto l998
					l999:
						position, thunkPosition = position998, thunkPosition998
						if !p.rules[ruleNewline]() {
							goto l997
						}
					}
				l998:
					goto l996
				l997:
					position, thunkPosition = position997, thunkPosition997
				}
				if !matchDot() {
					goto l996
				}
				goto l995
			l996:
				position, thunkPosition = position996, thunkPosition996
			}
			end = position
			if !matchChar('"') {
				goto l994
			}
			return true
		l994:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 186 RefTitleParens <- ('(' < (!((')' Sp Newline) / Newline) .)* > ')') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('(') {
				goto l1000
			}
			begin = position
		l1001:
			{
				position1002, thunkPosition1002 := position, thunkPosition
				{
					position1003, thunkPosition1003 := position, thunkPosition
					{
						position1004, thunkPosition1004 := position, thunkPosition
						if !matchChar(')') {
							goto l1005
						}
						if !p.rules[ruleSp]() {
							goto l1005
						}
						if !p.rules[ruleNewline]() {
							goto l1005
						}
						goto l1004
					l1005:
						position, thunkPosition = position1004, thunkPosition1004
						if !p.rules[ruleNewline]() {
							goto l1003
						}
					}
				l1004:
					goto l1002
				l1003:
					position, thunkPosition = position1003, thunkPosition1003
				}
				if !matchDot() {
					goto l1002
				}
				goto l1001
			l1002:
				position, thunkPosition = position1002, thunkPosition1002
			}
			end = position
			if !matchChar(')') {
				goto l1000
			}
			return true
		l1000:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 187 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1006
			}
			doarg(yySet, -1)
		l1007:
			{
				position1008, thunkPosition1008 := position, thunkPosition
				{
					position1009, thunkPosition1009 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto l1010
					}
					doarg(yySet, -2)
					do(83)
					goto l1009
				l1010:
					position, thunkPosition = position1009, thunkPosition1009
					if !p.rules[ruleSkipBlock]() {
						goto l1008
					}
				}
			l1009:
				goto l1007
			l1008:
				position, thunkPosition = position1008, thunkPosition1008
			}
			do(84)
			if !(commit(thunkPosition0)) {
				goto l1006
			}
			doarg(yyPop, 2)
			return true
		l1006:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 188 Ticks1 <- ('`' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('`') {
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
		/* 189 Ticks2 <- ('``' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("``") {
				goto l1012
			}
			if peekChar('`') {
				goto l1012
			}
			return true
		l1012:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 190 Ticks3 <- ('```' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("```") {
				goto l1013
			}
			if peekChar('`') {
				goto l1013
			}
			return true
		l1013:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 191 Ticks4 <- ('````' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("````") {
				goto l1014
			}
			if peekChar('`') {
				goto l1014
			}
			return true
		l1014:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 192 Ticks5 <- ('`````' !'`') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("`````") {
				goto l1015
			}
			if peekChar('`') {
				goto l1015
			}
			return true
		l1015:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 193 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / (!Ticks1 '`'+) / (!(Sp Ticks1) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / (!Ticks2 '`'+) / (!(Sp Ticks2) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / (!Ticks3 '`'+) / (!(Sp Ticks3) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / (!Ticks4 '`'+) / (!(Sp Ticks4) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / (!Ticks5 '`'+) / (!(Sp Ticks5) (Spacechar / (Newline !BlankLine))))+ > Sp Ticks5)) { yy = mk_str(yytext); yy.key = CODE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1017, thunkPosition1017 := position, thunkPosition
				if !p.rules[ruleTicks1]() {
					goto l1018
				}
				if !p.rules[ruleSp]() {
					goto l1018
				}
				begin = position
				{
					position1021, thunkPosition1021 := position, thunkPosition
					if peekChar('`') {
						goto l1022
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1022
					}
				l1023:
					{
						position1024, thunkPosition1024 := position, thunkPosition
						if peekChar('`') {
							goto l1024
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1024
						}
						goto l1023
					l1024:
						position, thunkPosition = position1024, thunkPosition1024
					}
					goto l1021
				l1022:
					position, thunkPosition = position1021, thunkPosition1021
					{
						position1026, thunkPosition1026 := position, thunkPosition
						if !p.rules[ruleTicks1]() {
							goto l1026
						}
						goto l1025
					l1026:
						position, thunkPosition = position1026, thunkPosition1026
					}
					if !matchChar('`') {
						goto l1025
					}
				l1027:
					{
						position1028, thunkPosition1028 := position, thunkPosition
						if !matchChar('`') {
							goto l1028
						}
						goto l1027
					l1028:
						position, thunkPosition = position1028, thunkPosition1028
					}
					goto l1021
				l1025:
					position, thunkPosition = position1021, thunkPosition1021
					{
						position1029, thunkPosition1029 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1029
						}
						if !p.rules[ruleTicks1]() {
							goto l1029
						}
						goto l1018
					l1029:
						position, thunkPosition = position1029, thunkPosition1029
					}
					{
						position1030, thunkPosition1030 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1031
						}
						goto l1030
					l1031:
						position, thunkPosition = position1030, thunkPosition1030
						if !p.rules[ruleNewline]() {
							goto l1018
						}
						{
							position1032, thunkPosition1032 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1032
							}
							goto l1018
						l1032:
							position, thunkPosition = position1032, thunkPosition1032
						}
					}
				l1030:
				}
			l1021:
			l1019:
				{
					position1020, thunkPosition1020 := position, thunkPosition
					{
						position1033, thunkPosition1033 := position, thunkPosition
						if peekChar('`') {
							goto l1034
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1034
						}
					l1035:
						{
							position1036, thunkPosition1036 := position, thunkPosition
							if peekChar('`') {
								goto l1036
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1036
							}
							goto l1035
						l1036:
							position, thunkPosition = position1036, thunkPosition1036
						}
						goto l1033
					l1034:
						position, thunkPosition = position1033, thunkPosition1033
						{
							position1038, thunkPosition1038 := position, thunkPosition
							if !p.rules[ruleTicks1]() {
								goto l1038
							}
							goto l1037
						l1038:
							position, thunkPosition = position1038, thunkPosition1038
						}
						if !matchChar('`') {
							goto l1037
						}
					l1039:
						{
							position1040, thunkPosition1040 := position, thunkPosition
							if !matchChar('`') {
								goto l1040
							}
							goto l1039
						l1040:
							position, thunkPosition = position1040, thunkPosition1040
						}
						goto l1033
					l1037:
						position, thunkPosition = position1033, thunkPosition1033
						{
							position1041, thunkPosition1041 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1041
							}
							if !p.rules[ruleTicks1]() {
								goto l1041
							}
							goto l1020
						l1041:
							position, thunkPosition = position1041, thunkPosition1041
						}
						{
							position1042, thunkPosition1042 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1043
							}
							goto l1042
						l1043:
							position, thunkPosition = position1042, thunkPosition1042
							if !p.rules[ruleNewline]() {
								goto l1020
							}
							{
								position1044, thunkPosition1044 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1044
								}
								goto l1020
							l1044:
								position, thunkPosition = position1044, thunkPosition1044
							}
						}
					l1042:
					}
				l1033:
					goto l1019
				l1020:
					position, thunkPosition = position1020, thunkPosition1020
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1018
				}
				if !p.rules[ruleTicks1]() {
					goto l1018
				}
				goto l1017
			l1018:
				position, thunkPosition = position1017, thunkPosition1017
				if !p.rules[ruleTicks2]() {
					goto l1045
				}
				if !p.rules[ruleSp]() {
					goto l1045
				}
				begin = position
				{
					position1048, thunkPosition1048 := position, thunkPosition
					if peekChar('`') {
						goto l1049
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1049
					}
				l1050:
					{
						position1051, thunkPosition1051 := position, thunkPosition
						if peekChar('`') {
							goto l1051
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1051
						}
						goto l1050
					l1051:
						position, thunkPosition = position1051, thunkPosition1051
					}
					goto l1048
				l1049:
					position, thunkPosition = position1048, thunkPosition1048
					{
						position1053, thunkPosition1053 := position, thunkPosition
						if !p.rules[ruleTicks2]() {
							goto l1053
						}
						goto l1052
					l1053:
						position, thunkPosition = position1053, thunkPosition1053
					}
					if !matchChar('`') {
						goto l1052
					}
				l1054:
					{
						position1055, thunkPosition1055 := position, thunkPosition
						if !matchChar('`') {
							goto l1055
						}
						goto l1054
					l1055:
						position, thunkPosition = position1055, thunkPosition1055
					}
					goto l1048
				l1052:
					position, thunkPosition = position1048, thunkPosition1048
					{
						position1056, thunkPosition1056 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1056
						}
						if !p.rules[ruleTicks2]() {
							goto l1056
						}
						goto l1045
					l1056:
						position, thunkPosition = position1056, thunkPosition1056
					}
					{
						position1057, thunkPosition1057 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1058
						}
						goto l1057
					l1058:
						position, thunkPosition = position1057, thunkPosition1057
						if !p.rules[ruleNewline]() {
							goto l1045
						}
						{
							position1059, thunkPosition1059 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1059
							}
							goto l1045
						l1059:
							position, thunkPosition = position1059, thunkPosition1059
						}
					}
				l1057:
				}
			l1048:
			l1046:
				{
					position1047, thunkPosition1047 := position, thunkPosition
					{
						position1060, thunkPosition1060 := position, thunkPosition
						if peekChar('`') {
							goto l1061
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1061
						}
					l1062:
						{
							position1063, thunkPosition1063 := position, thunkPosition
							if peekChar('`') {
								goto l1063
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1063
							}
							goto l1062
						l1063:
							position, thunkPosition = position1063, thunkPosition1063
						}
						goto l1060
					l1061:
						position, thunkPosition = position1060, thunkPosition1060
						{
							position1065, thunkPosition1065 := position, thunkPosition
							if !p.rules[ruleTicks2]() {
								goto l1065
							}
							goto l1064
						l1065:
							position, thunkPosition = position1065, thunkPosition1065
						}
						if !matchChar('`') {
							goto l1064
						}
					l1066:
						{
							position1067, thunkPosition1067 := position, thunkPosition
							if !matchChar('`') {
								goto l1067
							}
							goto l1066
						l1067:
							position, thunkPosition = position1067, thunkPosition1067
						}
						goto l1060
					l1064:
						position, thunkPosition = position1060, thunkPosition1060
						{
							position1068, thunkPosition1068 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1068
							}
							if !p.rules[ruleTicks2]() {
								goto l1068
							}
							goto l1047
						l1068:
							position, thunkPosition = position1068, thunkPosition1068
						}
						{
							position1069, thunkPosition1069 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1070
							}
							goto l1069
						l1070:
							position, thunkPosition = position1069, thunkPosition1069
							if !p.rules[ruleNewline]() {
								goto l1047
							}
							{
								position1071, thunkPosition1071 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1071
								}
								goto l1047
							l1071:
								position, thunkPosition = position1071, thunkPosition1071
							}
						}
					l1069:
					}
				l1060:
					goto l1046
				l1047:
					position, thunkPosition = position1047, thunkPosition1047
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1045
				}
				if !p.rules[ruleTicks2]() {
					goto l1045
				}
				goto l1017
			l1045:
				position, thunkPosition = position1017, thunkPosition1017
				if !p.rules[ruleTicks3]() {
					goto l1072
				}
				if !p.rules[ruleSp]() {
					goto l1072
				}
				begin = position
				{
					position1075, thunkPosition1075 := position, thunkPosition
					if peekChar('`') {
						goto l1076
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1076
					}
				l1077:
					{
						position1078, thunkPosition1078 := position, thunkPosition
						if peekChar('`') {
							goto l1078
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1078
						}
						goto l1077
					l1078:
						position, thunkPosition = position1078, thunkPosition1078
					}
					goto l1075
				l1076:
					position, thunkPosition = position1075, thunkPosition1075
					{
						position1080, thunkPosition1080 := position, thunkPosition
						if !p.rules[ruleTicks3]() {
							goto l1080
						}
						goto l1079
					l1080:
						position, thunkPosition = position1080, thunkPosition1080
					}
					if !matchChar('`') {
						goto l1079
					}
				l1081:
					{
						position1082, thunkPosition1082 := position, thunkPosition
						if !matchChar('`') {
							goto l1082
						}
						goto l1081
					l1082:
						position, thunkPosition = position1082, thunkPosition1082
					}
					goto l1075
				l1079:
					position, thunkPosition = position1075, thunkPosition1075
					{
						position1083, thunkPosition1083 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1083
						}
						if !p.rules[ruleTicks3]() {
							goto l1083
						}
						goto l1072
					l1083:
						position, thunkPosition = position1083, thunkPosition1083
					}
					{
						position1084, thunkPosition1084 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1085
						}
						goto l1084
					l1085:
						position, thunkPosition = position1084, thunkPosition1084
						if !p.rules[ruleNewline]() {
							goto l1072
						}
						{
							position1086, thunkPosition1086 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1086
							}
							goto l1072
						l1086:
							position, thunkPosition = position1086, thunkPosition1086
						}
					}
				l1084:
				}
			l1075:
			l1073:
				{
					position1074, thunkPosition1074 := position, thunkPosition
					{
						position1087, thunkPosition1087 := position, thunkPosition
						if peekChar('`') {
							goto l1088
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1088
						}
					l1089:
						{
							position1090, thunkPosition1090 := position, thunkPosition
							if peekChar('`') {
								goto l1090
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1090
							}
							goto l1089
						l1090:
							position, thunkPosition = position1090, thunkPosition1090
						}
						goto l1087
					l1088:
						position, thunkPosition = position1087, thunkPosition1087
						{
							position1092, thunkPosition1092 := position, thunkPosition
							if !p.rules[ruleTicks3]() {
								goto l1092
							}
							goto l1091
						l1092:
							position, thunkPosition = position1092, thunkPosition1092
						}
						if !matchChar('`') {
							goto l1091
						}
					l1093:
						{
							position1094, thunkPosition1094 := position, thunkPosition
							if !matchChar('`') {
								goto l1094
							}
							goto l1093
						l1094:
							position, thunkPosition = position1094, thunkPosition1094
						}
						goto l1087
					l1091:
						position, thunkPosition = position1087, thunkPosition1087
						{
							position1095, thunkPosition1095 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1095
							}
							if !p.rules[ruleTicks3]() {
								goto l1095
							}
							goto l1074
						l1095:
							position, thunkPosition = position1095, thunkPosition1095
						}
						{
							position1096, thunkPosition1096 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1097
							}
							goto l1096
						l1097:
							position, thunkPosition = position1096, thunkPosition1096
							if !p.rules[ruleNewline]() {
								goto l1074
							}
							{
								position1098, thunkPosition1098 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1098
								}
								goto l1074
							l1098:
								position, thunkPosition = position1098, thunkPosition1098
							}
						}
					l1096:
					}
				l1087:
					goto l1073
				l1074:
					position, thunkPosition = position1074, thunkPosition1074
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1072
				}
				if !p.rules[ruleTicks3]() {
					goto l1072
				}
				goto l1017
			l1072:
				position, thunkPosition = position1017, thunkPosition1017
				if !p.rules[ruleTicks4]() {
					goto l1099
				}
				if !p.rules[ruleSp]() {
					goto l1099
				}
				begin = position
				{
					position1102, thunkPosition1102 := position, thunkPosition
					if peekChar('`') {
						goto l1103
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1103
					}
				l1104:
					{
						position1105, thunkPosition1105 := position, thunkPosition
						if peekChar('`') {
							goto l1105
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1105
						}
						goto l1104
					l1105:
						position, thunkPosition = position1105, thunkPosition1105
					}
					goto l1102
				l1103:
					position, thunkPosition = position1102, thunkPosition1102
					{
						position1107, thunkPosition1107 := position, thunkPosition
						if !p.rules[ruleTicks4]() {
							goto l1107
						}
						goto l1106
					l1107:
						position, thunkPosition = position1107, thunkPosition1107
					}
					if !matchChar('`') {
						goto l1106
					}
				l1108:
					{
						position1109, thunkPosition1109 := position, thunkPosition
						if !matchChar('`') {
							goto l1109
						}
						goto l1108
					l1109:
						position, thunkPosition = position1109, thunkPosition1109
					}
					goto l1102
				l1106:
					position, thunkPosition = position1102, thunkPosition1102
					{
						position1110, thunkPosition1110 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1110
						}
						if !p.rules[ruleTicks4]() {
							goto l1110
						}
						goto l1099
					l1110:
						position, thunkPosition = position1110, thunkPosition1110
					}
					{
						position1111, thunkPosition1111 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1112
						}
						goto l1111
					l1112:
						position, thunkPosition = position1111, thunkPosition1111
						if !p.rules[ruleNewline]() {
							goto l1099
						}
						{
							position1113, thunkPosition1113 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1113
							}
							goto l1099
						l1113:
							position, thunkPosition = position1113, thunkPosition1113
						}
					}
				l1111:
				}
			l1102:
			l1100:
				{
					position1101, thunkPosition1101 := position, thunkPosition
					{
						position1114, thunkPosition1114 := position, thunkPosition
						if peekChar('`') {
							goto l1115
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1115
						}
					l1116:
						{
							position1117, thunkPosition1117 := position, thunkPosition
							if peekChar('`') {
								goto l1117
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1117
							}
							goto l1116
						l1117:
							position, thunkPosition = position1117, thunkPosition1117
						}
						goto l1114
					l1115:
						position, thunkPosition = position1114, thunkPosition1114
						{
							position1119, thunkPosition1119 := position, thunkPosition
							if !p.rules[ruleTicks4]() {
								goto l1119
							}
							goto l1118
						l1119:
							position, thunkPosition = position1119, thunkPosition1119
						}
						if !matchChar('`') {
							goto l1118
						}
					l1120:
						{
							position1121, thunkPosition1121 := position, thunkPosition
							if !matchChar('`') {
								goto l1121
							}
							goto l1120
						l1121:
							position, thunkPosition = position1121, thunkPosition1121
						}
						goto l1114
					l1118:
						position, thunkPosition = position1114, thunkPosition1114
						{
							position1122, thunkPosition1122 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1122
							}
							if !p.rules[ruleTicks4]() {
								goto l1122
							}
							goto l1101
						l1122:
							position, thunkPosition = position1122, thunkPosition1122
						}
						{
							position1123, thunkPosition1123 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1124
							}
							goto l1123
						l1124:
							position, thunkPosition = position1123, thunkPosition1123
							if !p.rules[ruleNewline]() {
								goto l1101
							}
							{
								position1125, thunkPosition1125 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1125
								}
								goto l1101
							l1125:
								position, thunkPosition = position1125, thunkPosition1125
							}
						}
					l1123:
					}
				l1114:
					goto l1100
				l1101:
					position, thunkPosition = position1101, thunkPosition1101
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1099
				}
				if !p.rules[ruleTicks4]() {
					goto l1099
				}
				goto l1017
			l1099:
				position, thunkPosition = position1017, thunkPosition1017
				if !p.rules[ruleTicks5]() {
					goto l1016
				}
				if !p.rules[ruleSp]() {
					goto l1016
				}
				begin = position
				{
					position1128, thunkPosition1128 := position, thunkPosition
					if peekChar('`') {
						goto l1129
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1129
					}
				l1130:
					{
						position1131, thunkPosition1131 := position, thunkPosition
						if peekChar('`') {
							goto l1131
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1131
						}
						goto l1130
					l1131:
						position, thunkPosition = position1131, thunkPosition1131
					}
					goto l1128
				l1129:
					position, thunkPosition = position1128, thunkPosition1128
					{
						position1133, thunkPosition1133 := position, thunkPosition
						if !p.rules[ruleTicks5]() {
							goto l1133
						}
						goto l1132
					l1133:
						position, thunkPosition = position1133, thunkPosition1133
					}
					if !matchChar('`') {
						goto l1132
					}
				l1134:
					{
						position1135, thunkPosition1135 := position, thunkPosition
						if !matchChar('`') {
							goto l1135
						}
						goto l1134
					l1135:
						position, thunkPosition = position1135, thunkPosition1135
					}
					goto l1128
				l1132:
					position, thunkPosition = position1128, thunkPosition1128
					{
						position1136, thunkPosition1136 := position, thunkPosition
						if !p.rules[ruleSp]() {
							goto l1136
						}
						if !p.rules[ruleTicks5]() {
							goto l1136
						}
						goto l1016
					l1136:
						position, thunkPosition = position1136, thunkPosition1136
					}
					{
						position1137, thunkPosition1137 := position, thunkPosition
						if !p.rules[ruleSpacechar]() {
							goto l1138
						}
						goto l1137
					l1138:
						position, thunkPosition = position1137, thunkPosition1137
						if !p.rules[ruleNewline]() {
							goto l1016
						}
						{
							position1139, thunkPosition1139 := position, thunkPosition
							if !p.rules[ruleBlankLine]() {
								goto l1139
							}
							goto l1016
						l1139:
							position, thunkPosition = position1139, thunkPosition1139
						}
					}
				l1137:
				}
			l1128:
			l1126:
				{
					position1127, thunkPosition1127 := position, thunkPosition
					{
						position1140, thunkPosition1140 := position, thunkPosition
						if peekChar('`') {
							goto l1141
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1141
						}
					l1142:
						{
							position1143, thunkPosition1143 := position, thunkPosition
							if peekChar('`') {
								goto l1143
							}
							if !p.rules[ruleNonspacechar]() {
								goto l1143
							}
							goto l1142
						l1143:
							position, thunkPosition = position1143, thunkPosition1143
						}
						goto l1140
					l1141:
						position, thunkPosition = position1140, thunkPosition1140
						{
							position1145, thunkPosition1145 := position, thunkPosition
							if !p.rules[ruleTicks5]() {
								goto l1145
							}
							goto l1144
						l1145:
							position, thunkPosition = position1145, thunkPosition1145
						}
						if !matchChar('`') {
							goto l1144
						}
					l1146:
						{
							position1147, thunkPosition1147 := position, thunkPosition
							if !matchChar('`') {
								goto l1147
							}
							goto l1146
						l1147:
							position, thunkPosition = position1147, thunkPosition1147
						}
						goto l1140
					l1144:
						position, thunkPosition = position1140, thunkPosition1140
						{
							position1148, thunkPosition1148 := position, thunkPosition
							if !p.rules[ruleSp]() {
								goto l1148
							}
							if !p.rules[ruleTicks5]() {
								goto l1148
							}
							goto l1127
						l1148:
							position, thunkPosition = position1148, thunkPosition1148
						}
						{
							position1149, thunkPosition1149 := position, thunkPosition
							if !p.rules[ruleSpacechar]() {
								goto l1150
							}
							goto l1149
						l1150:
							position, thunkPosition = position1149, thunkPosition1149
							if !p.rules[ruleNewline]() {
								goto l1127
							}
							{
								position1151, thunkPosition1151 := position, thunkPosition
								if !p.rules[ruleBlankLine]() {
									goto l1151
								}
								goto l1127
							l1151:
								position, thunkPosition = position1151, thunkPosition1151
							}
						}
					l1149:
					}
				l1140:
					goto l1126
				l1127:
					position, thunkPosition = position1127, thunkPosition1127
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l1016
				}
				if !p.rules[ruleTicks5]() {
					goto l1016
				}
			}
		l1017:
			do(85)
			return true
		l1016:
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
				position1153, thunkPosition1153 := position, thunkPosition
				if !p.rules[ruleHtmlComment]() {
					goto l1154
				}
				goto l1153
			l1154:
				position, thunkPosition = position1153, thunkPosition1153
				if !p.rules[ruleHtmlTag]() {
					goto l1152
				}
			}
		l1153:
			end = position
			do(86)
			return true
		l1152:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 195 BlankLine <- (Sp Newline) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l1155
			}
			if !p.rules[ruleNewline]() {
				goto l1155
			}
			return true
		l1155:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 196 Quoted <- (('"' (!'"' .)* '"') / ('\'' (!'\'' .)* '\'')) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1157, thunkPosition1157 := position, thunkPosition
				if !matchChar('"') {
					goto l1158
				}
			l1159:
				{
					position1160, thunkPosition1160 := position, thunkPosition
					if peekChar('"') {
						goto l1160
					}
					if !matchDot() {
						goto l1160
					}
					goto l1159
				l1160:
					position, thunkPosition = position1160, thunkPosition1160
				}
				if !matchChar('"') {
					goto l1158
				}
				goto l1157
			l1158:
				position, thunkPosition = position1157, thunkPosition1157
				if !matchChar('\'') {
					goto l1156
				}
			l1161:
				{
					position1162, thunkPosition1162 := position, thunkPosition
					if peekChar('\'') {
						goto l1162
					}
					if !matchDot() {
						goto l1162
					}
					goto l1161
				l1162:
					position, thunkPosition = position1162, thunkPosition1162
				}
				if !matchChar('\'') {
					goto l1156
				}
			}
		l1157:
			return true
		l1156:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 197 HtmlAttribute <- ((AlphanumericAscii / '-')+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1166, thunkPosition1166 := position, thunkPosition
				if !p.rules[ruleAlphanumericAscii]() {
					goto l1167
				}
				goto l1166
			l1167:
				position, thunkPosition = position1166, thunkPosition1166
				if !matchChar('-') {
					goto l1163
				}
			}
		l1166:
		l1164:
			{
				position1165, thunkPosition1165 := position, thunkPosition
				{
					position1168, thunkPosition1168 := position, thunkPosition
					if !p.rules[ruleAlphanumericAscii]() {
						goto l1169
					}
					goto l1168
				l1169:
					position, thunkPosition = position1168, thunkPosition1168
					if !matchChar('-') {
						goto l1165
					}
				}
			l1168:
				goto l1164
			l1165:
				position, thunkPosition = position1165, thunkPosition1165
			}
			if !p.rules[ruleSpnl]() {
				goto l1163
			}
			{
				position1170, thunkPosition1170 := position, thunkPosition
				if !matchChar('=') {
					goto l1170
				}
				if !p.rules[ruleSpnl]() {
					goto l1170
				}
				{
					position1172, thunkPosition1172 := position, thunkPosition
					if !p.rules[ruleQuoted]() {
						goto l1173
					}
					goto l1172
				l1173:
					position, thunkPosition = position1172, thunkPosition1172
					if peekChar('>') {
						goto l1170
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1170
					}
				l1174:
					{
						position1175, thunkPosition1175 := position, thunkPosition
						if peekChar('>') {
							goto l1175
						}
						if !p.rules[ruleNonspacechar]() {
							goto l1175
						}
						goto l1174
					l1175:
						position, thunkPosition = position1175, thunkPosition1175
					}
				}
			l1172:
				goto l1171
			l1170:
				position, thunkPosition = position1170, thunkPosition1170
			}
		l1171:
			if !p.rules[ruleSpnl]() {
				goto l1163
			}
			return true
		l1163:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 198 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("<!--") {
				goto l1176
			}
		l1177:
			{
				position1178, thunkPosition1178 := position, thunkPosition
				{
					position1179, thunkPosition1179 := position, thunkPosition
					if !matchString("-->") {
						goto l1179
					}
					goto l1178
				l1179:
					position, thunkPosition = position1179, thunkPosition1179
				}
				if !matchDot() {
					goto l1178
				}
				goto l1177
			l1178:
				position, thunkPosition = position1178, thunkPosition1178
			}
			if !matchString("-->") {
				goto l1176
			}
			return true
		l1176:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 199 HtmlTag <- ('<' Spnl '/'? AlphanumericAscii+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l1180
			}
			if !p.rules[ruleSpnl]() {
				goto l1180
			}
			{
				position1181, thunkPosition1181 := position, thunkPosition
				if !matchChar('/') {
					goto l1181
				}
				goto l1182
			l1181:
				position, thunkPosition = position1181, thunkPosition1181
			}
		l1182:
			if !p.rules[ruleAlphanumericAscii]() {
				goto l1180
			}
		l1183:
			{
				position1184, thunkPosition1184 := position, thunkPosition
				if !p.rules[ruleAlphanumericAscii]() {
					goto l1184
				}
				goto l1183
			l1184:
				position, thunkPosition = position1184, thunkPosition1184
			}
			if !p.rules[ruleSpnl]() {
				goto l1180
			}
		l1185:
			{
				position1186, thunkPosition1186 := position, thunkPosition
				if !p.rules[ruleHtmlAttribute]() {
					goto l1186
				}
				goto l1185
			l1186:
				position, thunkPosition = position1186, thunkPosition1186
			}
			{
				position1187, thunkPosition1187 := position, thunkPosition
				if !matchChar('/') {
					goto l1187
				}
				goto l1188
			l1187:
				position, thunkPosition = position1187, thunkPosition1187
			}
		l1188:
			if !p.rules[ruleSpnl]() {
				goto l1180
			}
			if !matchChar('>') {
				goto l1180
			}
			return true
		l1180:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 200 Eof <- !. */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if peekDot() {
				goto l1189
			}
			return true
		l1189:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 201 Spacechar <- (' ' / '\t') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1191, thunkPosition1191 := position, thunkPosition
				if !matchChar(' ') {
					goto l1192
				}
				goto l1191
			l1192:
				position, thunkPosition = position1191, thunkPosition1191
				if !matchChar('\t') {
					goto l1190
				}
			}
		l1191:
			return true
		l1190:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 202 Nonspacechar <- (!Spacechar !Newline .) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1194, thunkPosition1194 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1194
				}
				goto l1193
			l1194:
				position, thunkPosition = position1194, thunkPosition1194
			}
			{
				position1195, thunkPosition1195 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1195
				}
				goto l1193
			l1195:
				position, thunkPosition = position1195, thunkPosition1195
			}
			if !matchDot() {
				goto l1193
			}
			return true
		l1193:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 203 Newline <- ('\n' / ('\r' '\n'?)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1197, thunkPosition1197 := position, thunkPosition
				if !matchChar('\n') {
					goto l1198
				}
				goto l1197
			l1198:
				position, thunkPosition = position1197, thunkPosition1197
				if !matchChar('\r') {
					goto l1196
				}
				{
					position1199, thunkPosition1199 := position, thunkPosition
					if !matchChar('\n') {
						goto l1199
					}
					goto l1200
				l1199:
					position, thunkPosition = position1199, thunkPosition1199
				}
			l1200:
			}
		l1197:
			return true
		l1196:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 204 Sp <- Spacechar* */
		func() bool {
		l1202:
			{
				position1203, thunkPosition1203 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1203
				}
				goto l1202
			l1203:
				position, thunkPosition = position1203, thunkPosition1203
			}
			return true
		},
		/* 205 Spnl <- (Sp (Newline Sp)?) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l1204
			}
			{
				position1205, thunkPosition1205 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1205
				}
				if !p.rules[ruleSp]() {
					goto l1205
				}
				goto l1206
			l1205:
				position, thunkPosition = position1205, thunkPosition1205
			}
		l1206:
			return true
		l1204:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 206 SpecialChar <- ('*' / '_' / '`' / '&' / '[' / ']' / '<' / '!' / '#' / '\\' / ExtendedSpecialChar) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1208, thunkPosition1208 := position, thunkPosition
				if !matchChar('*') {
					goto l1209
				}
				goto l1208
			l1209:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar('_') {
					goto l1210
				}
				goto l1208
			l1210:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar('`') {
					goto l1211
				}
				goto l1208
			l1211:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar('&') {
					goto l1212
				}
				goto l1208
			l1212:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar('[') {
					goto l1213
				}
				goto l1208
			l1213:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar(']') {
					goto l1214
				}
				goto l1208
			l1214:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar('<') {
					goto l1215
				}
				goto l1208
			l1215:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar('!') {
					goto l1216
				}
				goto l1208
			l1216:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar('#') {
					goto l1217
				}
				goto l1208
			l1217:
				position, thunkPosition = position1208, thunkPosition1208
				if !matchChar('\\') {
					goto l1218
				}
				goto l1208
			l1218:
				position, thunkPosition = position1208, thunkPosition1208
				if !p.rules[ruleExtendedSpecialChar]() {
					goto l1207
				}
			}
		l1208:
			return true
		l1207:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 207 NormalChar <- (!(SpecialChar / Spacechar / Newline) .) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1220, thunkPosition1220 := position, thunkPosition
				{
					position1221, thunkPosition1221 := position, thunkPosition
					if !p.rules[ruleSpecialChar]() {
						goto l1222
					}
					goto l1221
				l1222:
					position, thunkPosition = position1221, thunkPosition1221
					if !p.rules[ruleSpacechar]() {
						goto l1223
					}
					goto l1221
				l1223:
					position, thunkPosition = position1221, thunkPosition1221
					if !p.rules[ruleNewline]() {
						goto l1220
					}
				}
			l1221:
				goto l1219
			l1220:
				position, thunkPosition = position1220, thunkPosition1220
			}
			if !matchDot() {
				goto l1219
			}
			return true
		l1219:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 208 NonAlphanumeric <- [\000-\057\072-\100\133-\140\173-\177] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(3) {
				goto l1224
			}
			return true
		l1224:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 209 Alphanumeric <- ([0-9A-Za-z] / '\200' / '\201' / '\202' / '\203' / '\204' / '\205' / '\206' / '\207' / '\210' / '\211' / '\212' / '\213' / '\214' / '\215' / '\216' / '\217' / '\220' / '\221' / '\222' / '\223' / '\224' / '\225' / '\226' / '\227' / '\230' / '\231' / '\232' / '\233' / '\234' / '\235' / '\236' / '\237' / '\240' / '\241' / '\242' / '\243' / '\244' / '\245' / '\246' / '\247' / '\250' / '\251' / '\252' / '\253' / '\254' / '\255' / '\256' / '\257' / '\260' / '\261' / '\262' / '\263' / '\264' / '\265' / '\266' / '\267' / '\270' / '\271' / '\272' / '\273' / '\274' / '\275' / '\276' / '\277' / '\300' / '\301' / '\302' / '\303' / '\304' / '\305' / '\306' / '\307' / '\310' / '\311' / '\312' / '\313' / '\314' / '\315' / '\316' / '\317' / '\320' / '\321' / '\322' / '\323' / '\324' / '\325' / '\326' / '\327' / '\330' / '\331' / '\332' / '\333' / '\334' / '\335' / '\336' / '\337' / '\340' / '\341' / '\342' / '\343' / '\344' / '\345' / '\346' / '\347' / '\350' / '\351' / '\352' / '\353' / '\354' / '\355' / '\356' / '\357' / '\360' / '\361' / '\362' / '\363' / '\364' / '\365' / '\366' / '\367' / '\370' / '\371' / '\372' / '\373' / '\374' / '\375' / '\376' / '\377') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1226, thunkPosition1226 := position, thunkPosition
				if !matchClass(1) {
					goto l1227
				}
				goto l1226
			l1227:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\200") {
					goto l1228
				}
				goto l1226
			l1228:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\201") {
					goto l1229
				}
				goto l1226
			l1229:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\202") {
					goto l1230
				}
				goto l1226
			l1230:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\203") {
					goto l1231
				}
				goto l1226
			l1231:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\204") {
					goto l1232
				}
				goto l1226
			l1232:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\205") {
					goto l1233
				}
				goto l1226
			l1233:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\206") {
					goto l1234
				}
				goto l1226
			l1234:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\207") {
					goto l1235
				}
				goto l1226
			l1235:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\210") {
					goto l1236
				}
				goto l1226
			l1236:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\211") {
					goto l1237
				}
				goto l1226
			l1237:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\212") {
					goto l1238
				}
				goto l1226
			l1238:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\213") {
					goto l1239
				}
				goto l1226
			l1239:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\214") {
					goto l1240
				}
				goto l1226
			l1240:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\215") {
					goto l1241
				}
				goto l1226
			l1241:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\216") {
					goto l1242
				}
				goto l1226
			l1242:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\217") {
					goto l1243
				}
				goto l1226
			l1243:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\220") {
					goto l1244
				}
				goto l1226
			l1244:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\221") {
					goto l1245
				}
				goto l1226
			l1245:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\222") {
					goto l1246
				}
				goto l1226
			l1246:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\223") {
					goto l1247
				}
				goto l1226
			l1247:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\224") {
					goto l1248
				}
				goto l1226
			l1248:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\225") {
					goto l1249
				}
				goto l1226
			l1249:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\226") {
					goto l1250
				}
				goto l1226
			l1250:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\227") {
					goto l1251
				}
				goto l1226
			l1251:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\230") {
					goto l1252
				}
				goto l1226
			l1252:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\231") {
					goto l1253
				}
				goto l1226
			l1253:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\232") {
					goto l1254
				}
				goto l1226
			l1254:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\233") {
					goto l1255
				}
				goto l1226
			l1255:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\234") {
					goto l1256
				}
				goto l1226
			l1256:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\235") {
					goto l1257
				}
				goto l1226
			l1257:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\236") {
					goto l1258
				}
				goto l1226
			l1258:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\237") {
					goto l1259
				}
				goto l1226
			l1259:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\240") {
					goto l1260
				}
				goto l1226
			l1260:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\241") {
					goto l1261
				}
				goto l1226
			l1261:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\242") {
					goto l1262
				}
				goto l1226
			l1262:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\243") {
					goto l1263
				}
				goto l1226
			l1263:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\244") {
					goto l1264
				}
				goto l1226
			l1264:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\245") {
					goto l1265
				}
				goto l1226
			l1265:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\246") {
					goto l1266
				}
				goto l1226
			l1266:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\247") {
					goto l1267
				}
				goto l1226
			l1267:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\250") {
					goto l1268
				}
				goto l1226
			l1268:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\251") {
					goto l1269
				}
				goto l1226
			l1269:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\252") {
					goto l1270
				}
				goto l1226
			l1270:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\253") {
					goto l1271
				}
				goto l1226
			l1271:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\254") {
					goto l1272
				}
				goto l1226
			l1272:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\255") {
					goto l1273
				}
				goto l1226
			l1273:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\256") {
					goto l1274
				}
				goto l1226
			l1274:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\257") {
					goto l1275
				}
				goto l1226
			l1275:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\260") {
					goto l1276
				}
				goto l1226
			l1276:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\261") {
					goto l1277
				}
				goto l1226
			l1277:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\262") {
					goto l1278
				}
				goto l1226
			l1278:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\263") {
					goto l1279
				}
				goto l1226
			l1279:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\264") {
					goto l1280
				}
				goto l1226
			l1280:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\265") {
					goto l1281
				}
				goto l1226
			l1281:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\266") {
					goto l1282
				}
				goto l1226
			l1282:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\267") {
					goto l1283
				}
				goto l1226
			l1283:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\270") {
					goto l1284
				}
				goto l1226
			l1284:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\271") {
					goto l1285
				}
				goto l1226
			l1285:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\272") {
					goto l1286
				}
				goto l1226
			l1286:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\273") {
					goto l1287
				}
				goto l1226
			l1287:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\274") {
					goto l1288
				}
				goto l1226
			l1288:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\275") {
					goto l1289
				}
				goto l1226
			l1289:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\276") {
					goto l1290
				}
				goto l1226
			l1290:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\277") {
					goto l1291
				}
				goto l1226
			l1291:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\300") {
					goto l1292
				}
				goto l1226
			l1292:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\301") {
					goto l1293
				}
				goto l1226
			l1293:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\302") {
					goto l1294
				}
				goto l1226
			l1294:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\303") {
					goto l1295
				}
				goto l1226
			l1295:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\304") {
					goto l1296
				}
				goto l1226
			l1296:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\305") {
					goto l1297
				}
				goto l1226
			l1297:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\306") {
					goto l1298
				}
				goto l1226
			l1298:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\307") {
					goto l1299
				}
				goto l1226
			l1299:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\310") {
					goto l1300
				}
				goto l1226
			l1300:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\311") {
					goto l1301
				}
				goto l1226
			l1301:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\312") {
					goto l1302
				}
				goto l1226
			l1302:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\313") {
					goto l1303
				}
				goto l1226
			l1303:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\314") {
					goto l1304
				}
				goto l1226
			l1304:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\315") {
					goto l1305
				}
				goto l1226
			l1305:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\316") {
					goto l1306
				}
				goto l1226
			l1306:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\317") {
					goto l1307
				}
				goto l1226
			l1307:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\320") {
					goto l1308
				}
				goto l1226
			l1308:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\321") {
					goto l1309
				}
				goto l1226
			l1309:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\322") {
					goto l1310
				}
				goto l1226
			l1310:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\323") {
					goto l1311
				}
				goto l1226
			l1311:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\324") {
					goto l1312
				}
				goto l1226
			l1312:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\325") {
					goto l1313
				}
				goto l1226
			l1313:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\326") {
					goto l1314
				}
				goto l1226
			l1314:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\327") {
					goto l1315
				}
				goto l1226
			l1315:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\330") {
					goto l1316
				}
				goto l1226
			l1316:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\331") {
					goto l1317
				}
				goto l1226
			l1317:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\332") {
					goto l1318
				}
				goto l1226
			l1318:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\333") {
					goto l1319
				}
				goto l1226
			l1319:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\334") {
					goto l1320
				}
				goto l1226
			l1320:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\335") {
					goto l1321
				}
				goto l1226
			l1321:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\336") {
					goto l1322
				}
				goto l1226
			l1322:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\337") {
					goto l1323
				}
				goto l1226
			l1323:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\340") {
					goto l1324
				}
				goto l1226
			l1324:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\341") {
					goto l1325
				}
				goto l1226
			l1325:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\342") {
					goto l1326
				}
				goto l1226
			l1326:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\343") {
					goto l1327
				}
				goto l1226
			l1327:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\344") {
					goto l1328
				}
				goto l1226
			l1328:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\345") {
					goto l1329
				}
				goto l1226
			l1329:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\346") {
					goto l1330
				}
				goto l1226
			l1330:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\347") {
					goto l1331
				}
				goto l1226
			l1331:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\350") {
					goto l1332
				}
				goto l1226
			l1332:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\351") {
					goto l1333
				}
				goto l1226
			l1333:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\352") {
					goto l1334
				}
				goto l1226
			l1334:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\353") {
					goto l1335
				}
				goto l1226
			l1335:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\354") {
					goto l1336
				}
				goto l1226
			l1336:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\355") {
					goto l1337
				}
				goto l1226
			l1337:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\356") {
					goto l1338
				}
				goto l1226
			l1338:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\357") {
					goto l1339
				}
				goto l1226
			l1339:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\360") {
					goto l1340
				}
				goto l1226
			l1340:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\361") {
					goto l1341
				}
				goto l1226
			l1341:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\362") {
					goto l1342
				}
				goto l1226
			l1342:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\363") {
					goto l1343
				}
				goto l1226
			l1343:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\364") {
					goto l1344
				}
				goto l1226
			l1344:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\365") {
					goto l1345
				}
				goto l1226
			l1345:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\366") {
					goto l1346
				}
				goto l1226
			l1346:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\367") {
					goto l1347
				}
				goto l1226
			l1347:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\370") {
					goto l1348
				}
				goto l1226
			l1348:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\371") {
					goto l1349
				}
				goto l1226
			l1349:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\372") {
					goto l1350
				}
				goto l1226
			l1350:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\373") {
					goto l1351
				}
				goto l1226
			l1351:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\374") {
					goto l1352
				}
				goto l1226
			l1352:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\375") {
					goto l1353
				}
				goto l1226
			l1353:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\376") {
					goto l1354
				}
				goto l1226
			l1354:
				position, thunkPosition = position1226, thunkPosition1226
				if !matchString("\377") {
					goto l1225
				}
			}
		l1226:
			return true
		l1225:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 210 AlphanumericAscii <- [A-Za-z0-9] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(8) {
				goto l1355
			}
			return true
		l1355:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 211 Digit <- [0-9] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(7) {
				goto l1356
			}
			return true
		l1356:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 212 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1357
			}
			if !matchChar('#') {
				goto l1357
			}
			if !matchClass(5) {
				goto l1357
			}
			if !matchClass(0) {
				goto l1357
			}
		l1358:
			{
				position1359, thunkPosition1359 := position, thunkPosition
				if !matchClass(0) {
					goto l1359
				}
				goto l1358
			l1359:
				position, thunkPosition = position1359, thunkPosition1359
			}
			if !matchChar(';') {
				goto l1357
			}
			end = position
			return true
		l1357:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 213 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1360
			}
			if !matchChar('#') {
				goto l1360
			}
			if !matchClass(7) {
				goto l1360
			}
		l1361:
			{
				position1362, thunkPosition1362 := position, thunkPosition
				if !matchClass(7) {
					goto l1362
				}
				goto l1361
			l1362:
				position, thunkPosition = position1362, thunkPosition1362
			}
			end = position
			if !matchChar(';') {
				goto l1360
			}
			end = position
			return true
		l1360:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 214 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchChar('&') {
				goto l1363
			}
			if !matchClass(8) {
				goto l1363
			}
		l1364:
			{
				position1365, thunkPosition1365 := position, thunkPosition
				if !matchClass(8) {
					goto l1365
				}
				goto l1364
			l1365:
				position, thunkPosition = position1365, thunkPosition1365
			}
			if !matchChar(';') {
				goto l1363
			}
			end = position
			return true
		l1363:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 215 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1367, thunkPosition1367 := position, thunkPosition
				if !matchString("   ") {
					goto l1368
				}
				goto l1367
			l1368:
				position, thunkPosition = position1367, thunkPosition1367
				if !matchString("  ") {
					goto l1369
				}
				goto l1367
			l1369:
				position, thunkPosition = position1367, thunkPosition1367
				if !matchChar(' ') {
					goto l1370
				}
				goto l1367
			l1370:
				position, thunkPosition = position1367, thunkPosition1367
				if !matchString("") {
					goto l1366
				}
			}
		l1367:
			return true
		l1366:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 216 Indent <- ('\t' / '    ') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1372, thunkPosition1372 := position, thunkPosition
				if !matchChar('\t') {
					goto l1373
				}
				goto l1372
			l1373:
				position, thunkPosition = position1372, thunkPosition1372
				if !matchString("    ") {
					goto l1371
				}
			}
		l1372:
			return true
		l1371:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 217 IndentedLine <- (Indent Line) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleIndent]() {
				goto l1374
			}
			if !p.rules[ruleLine]() {
				goto l1374
			}
			return true
		l1374:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 218 OptionallyIndentedLine <- (Indent? Line) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1376, thunkPosition1376 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l1376
				}
				goto l1377
			l1376:
				position, thunkPosition = position1376, thunkPosition1376
			}
		l1377:
			if !p.rules[ruleLine]() {
				goto l1375
			}
			return true
		l1375:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 219 StartList <- (&. { yy = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !peekDot() {
				goto l1378
			}
			do(87)
			return true
		l1378:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 220 Line <- (RawLine { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleRawLine]() {
				goto l1379
			}
			do(88)
			return true
		l1379:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 221 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > Eof)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1381, thunkPosition1381 := position, thunkPosition
				begin = position
			l1383:
				{
					position1384, thunkPosition1384 := position, thunkPosition
					if peekChar('\r') {
						goto l1384
					}
					if peekChar('\n') {
						goto l1384
					}
					if !matchDot() {
						goto l1384
					}
					goto l1383
				l1384:
					position, thunkPosition = position1384, thunkPosition1384
				}
				if !p.rules[ruleNewline]() {
					goto l1382
				}
				end = position
				goto l1381
			l1382:
				position, thunkPosition = position1381, thunkPosition1381
				begin = position
				if !matchDot() {
					goto l1380
				}
			l1385:
				{
					position1386, thunkPosition1386 := position, thunkPosition
					if !matchDot() {
						goto l1386
					}
					goto l1385
				l1386:
					position, thunkPosition = position1386, thunkPosition1386
				}
				end = position
				if !p.rules[ruleEof]() {
					goto l1380
				}
			}
		l1381:
			return true
		l1380:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 222 SkipBlock <- (((!BlankLine RawLine)+ BlankLine*) / BlankLine+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1388, thunkPosition1388 := position, thunkPosition
				{
					position1392, thunkPosition1392 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1392
					}
					goto l1389
				l1392:
					position, thunkPosition = position1392, thunkPosition1392
				}
				if !p.rules[ruleRawLine]() {
					goto l1389
				}
			l1390:
				{
					position1391, thunkPosition1391 := position, thunkPosition
					{
						position1393, thunkPosition1393 := position, thunkPosition
						if !p.rules[ruleBlankLine]() {
							goto l1393
						}
						goto l1391
					l1393:
						position, thunkPosition = position1393, thunkPosition1393
					}
					if !p.rules[ruleRawLine]() {
						goto l1391
					}
					goto l1390
				l1391:
					position, thunkPosition = position1391, thunkPosition1391
				}
			l1394:
				{
					position1395, thunkPosition1395 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1395
					}
					goto l1394
				l1395:
					position, thunkPosition = position1395, thunkPosition1395
				}
				goto l1388
			l1389:
				position, thunkPosition = position1388, thunkPosition1388
				if !p.rules[ruleBlankLine]() {
					goto l1387
				}
			l1396:
				{
					position1397, thunkPosition1397 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1397
					}
					goto l1396
				l1397:
					position, thunkPosition = position1397, thunkPosition1397
				}
			}
		l1388:
			return true
		l1387:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 223 ExtendedSpecialChar <- ((&{ p.extension.Smart } ('.' / '-' / '\'' / '"')) / (&{ p.extension.Notes } '^')) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1399, thunkPosition1399 := position, thunkPosition
				if !( p.extension.Smart ) {
					goto l1400
				}
				{
					position1401, thunkPosition1401 := position, thunkPosition
					if !matchChar('.') {
						goto l1402
					}
					goto l1401
				l1402:
					position, thunkPosition = position1401, thunkPosition1401
					if !matchChar('-') {
						goto l1403
					}
					goto l1401
				l1403:
					position, thunkPosition = position1401, thunkPosition1401
					if !matchChar('\'') {
						goto l1404
					}
					goto l1401
				l1404:
					position, thunkPosition = position1401, thunkPosition1401
					if !matchChar('"') {
						goto l1400
					}
				}
			l1401:
				goto l1399
			l1400:
				position, thunkPosition = position1399, thunkPosition1399
				if !( p.extension.Notes ) {
					goto l1398
				}
				if !matchChar('^') {
					goto l1398
				}
			}
		l1399:
			return true
		l1398:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 224 Smart <- (&{ p.extension.Smart } (Ellipsis / Dash / SingleQuoted / DoubleQuoted / Apostrophe)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !( p.extension.Smart ) {
				goto l1405
			}
			{
				position1406, thunkPosition1406 := position, thunkPosition
				if !p.rules[ruleEllipsis]() {
					goto l1407
				}
				goto l1406
			l1407:
				position, thunkPosition = position1406, thunkPosition1406
				if !p.rules[ruleDash]() {
					goto l1408
				}
				goto l1406
			l1408:
				position, thunkPosition = position1406, thunkPosition1406
				if !p.rules[ruleSingleQuoted]() {
					goto l1409
				}
				goto l1406
			l1409:
				position, thunkPosition = position1406, thunkPosition1406
				if !p.rules[ruleDoubleQuoted]() {
					goto l1410
				}
				goto l1406
			l1410:
				position, thunkPosition = position1406, thunkPosition1406
				if !p.rules[ruleApostrophe]() {
					goto l1405
				}
			}
		l1406:
			return true
		l1405:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 225 Apostrophe <- ('\'' { yy = mk_element(APOSTROPHE) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1411
			}
			do(89)
			return true
		l1411:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 226 Ellipsis <- (('...' / '. . .') { yy = mk_element(ELLIPSIS) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1413, thunkPosition1413 := position, thunkPosition
				if !matchString("...") {
					goto l1414
				}
				goto l1413
			l1414:
				position, thunkPosition = position1413, thunkPosition1413
				if !matchString(". . .") {
					goto l1412
				}
			}
		l1413:
			do(90)
			return true
		l1412:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 227 Dash <- (EmDash / EnDash) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1416, thunkPosition1416 := position, thunkPosition
				if !p.rules[ruleEmDash]() {
					goto l1417
				}
				goto l1416
			l1417:
				position, thunkPosition = position1416, thunkPosition1416
				if !p.rules[ruleEnDash]() {
					goto l1415
				}
			}
		l1416:
			return true
		l1415:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 228 EnDash <- ('-' &Digit { yy = mk_element(ENDASH) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('-') {
				goto l1418
			}
			{
				position1419, thunkPosition1419 := position, thunkPosition
				if !p.rules[ruleDigit]() {
					goto l1418
				}
				position, thunkPosition = position1419, thunkPosition1419
			}
			do(91)
			return true
		l1418:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 229 EmDash <- (('---' / '--') { yy = mk_element(EMDASH) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1421, thunkPosition1421 := position, thunkPosition
				if !matchString("---") {
					goto l1422
				}
				goto l1421
			l1422:
				position, thunkPosition = position1421, thunkPosition1421
				if !matchString("--") {
					goto l1420
				}
			}
		l1421:
			do(92)
			return true
		l1420:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 230 SingleQuoteStart <- ('\'' ![)!\],.;:-? \t\n] !(('s' / 't' / 'm' / 've' / 'll' / 're') !Alphanumeric)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1423
			}
			{
				position1424, thunkPosition1424 := position, thunkPosition
				if !matchClass(6) {
					goto l1424
				}
				goto l1423
			l1424:
				position, thunkPosition = position1424, thunkPosition1424
			}
			{
				position1425, thunkPosition1425 := position, thunkPosition
				{
					position1426, thunkPosition1426 := position, thunkPosition
					if !matchChar('s') {
						goto l1427
					}
					goto l1426
				l1427:
					position, thunkPosition = position1426, thunkPosition1426
					if !matchChar('t') {
						goto l1428
					}
					goto l1426
				l1428:
					position, thunkPosition = position1426, thunkPosition1426
					if !matchChar('m') {
						goto l1429
					}
					goto l1426
				l1429:
					position, thunkPosition = position1426, thunkPosition1426
					if !matchString("ve") {
						goto l1430
					}
					goto l1426
				l1430:
					position, thunkPosition = position1426, thunkPosition1426
					if !matchString("ll") {
						goto l1431
					}
					goto l1426
				l1431:
					position, thunkPosition = position1426, thunkPosition1426
					if !matchString("re") {
						goto l1425
					}
				}
			l1426:
				{
					position1432, thunkPosition1432 := position, thunkPosition
					if !p.rules[ruleAlphanumeric]() {
						goto l1432
					}
					goto l1425
				l1432:
					position, thunkPosition = position1432, thunkPosition1432
				}
				goto l1423
			l1425:
				position, thunkPosition = position1425, thunkPosition1425
			}
			return true
		l1423:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 231 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('\'') {
				goto l1433
			}
			{
				position1434, thunkPosition1434 := position, thunkPosition
				if !p.rules[ruleAlphanumeric]() {
					goto l1434
				}
				goto l1433
			l1434:
				position, thunkPosition = position1434, thunkPosition1434
			}
			return true
		l1433:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 232 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = mk_list(SINGLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto l1435
			}
			if !p.rules[ruleStartList]() {
				goto l1435
			}
			doarg(yySet, -1)
			{
				position1438, thunkPosition1438 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto l1438
				}
				goto l1435
			l1438:
				position, thunkPosition = position1438, thunkPosition1438
			}
			if !p.rules[ruleInline]() {
				goto l1435
			}
			doarg(yySet, -2)
			do(93)
		l1436:
			{
				position1437, thunkPosition1437 := position, thunkPosition
				{
					position1439, thunkPosition1439 := position, thunkPosition
					if !p.rules[ruleSingleQuoteEnd]() {
						goto l1439
					}
					goto l1437
				l1439:
					position, thunkPosition = position1439, thunkPosition1439
				}
				if !p.rules[ruleInline]() {
					goto l1437
				}
				doarg(yySet, -2)
				do(93)
				goto l1436
			l1437:
				position, thunkPosition = position1437, thunkPosition1437
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1435
			}
			do(94)
			doarg(yyPop, 2)
			return true
		l1435:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 233 DoubleQuoteStart <- '"' */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l1440
			}
			return true
		l1440:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 234 DoubleQuoteEnd <- '"' */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('"') {
				goto l1441
			}
			return true
		l1441:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 235 DoubleQuoted <- (DoubleQuoteStart StartList (!DoubleQuoteEnd Inline { a = cons(b, a) })+ DoubleQuoteEnd { yy = mk_list(DOUBLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleDoubleQuoteStart]() {
				goto l1442
			}
			if !p.rules[ruleStartList]() {
				goto l1442
			}
			doarg(yySet, -1)
			{
				position1445, thunkPosition1445 := position, thunkPosition
				if !p.rules[ruleDoubleQuoteEnd]() {
					goto l1445
				}
				goto l1442
			l1445:
				position, thunkPosition = position1445, thunkPosition1445
			}
			if !p.rules[ruleInline]() {
				goto l1442
			}
			doarg(yySet, -2)
			do(95)
		l1443:
			{
				position1444, thunkPosition1444 := position, thunkPosition
				{
					position1446, thunkPosition1446 := position, thunkPosition
					if !p.rules[ruleDoubleQuoteEnd]() {
						goto l1446
					}
					goto l1444
				l1446:
					position, thunkPosition = position1446, thunkPosition1446
				}
				if !p.rules[ruleInline]() {
					goto l1444
				}
				doarg(yySet, -2)
				do(95)
				goto l1443
			l1444:
				position, thunkPosition = position1444, thunkPosition1444
			}
			if !p.rules[ruleDoubleQuoteEnd]() {
				goto l1442
			}
			do(96)
			doarg(yyPop, 2)
			return true
		l1442:
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
				goto l1447
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1447
			}
			doarg(yySet, -1)
			do(97)
			doarg(yyPop, 1)
			return true
		l1447:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 237 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = mk_str(yytext) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("[^") {
				goto l1448
			}
			begin = position
			{
				position1451, thunkPosition1451 := position, thunkPosition
				if !p.rules[ruleNewline]() {
					goto l1451
				}
				goto l1448
			l1451:
				position, thunkPosition = position1451, thunkPosition1451
			}
			if peekChar(']') {
				goto l1448
			}
			if !matchDot() {
				goto l1448
			}
		l1449:
			{
				position1450, thunkPosition1450 := position, thunkPosition
				{
					position1452, thunkPosition1452 := position, thunkPosition
					if !p.rules[ruleNewline]() {
						goto l1452
					}
					goto l1450
				l1452:
					position, thunkPosition = position1452, thunkPosition1452
				}
				if peekChar(']') {
					goto l1450
				}
				if !matchDot() {
					goto l1450
				}
				goto l1449
			l1450:
				position, thunkPosition = position1450, thunkPosition1450
			}
			end = position
			if !matchChar(']') {
				goto l1448
			}
			do(98)
			return true
		l1448:
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
				goto l1453
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l1453
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1453
			}
			doarg(yySet, -1)
			if !matchChar(':') {
				goto l1453
			}
			if !p.rules[ruleSp]() {
				goto l1453
			}
			if !p.rules[ruleStartList]() {
				goto l1453
			}
			doarg(yySet, -2)
			if !p.rules[ruleRawNoteBlock]() {
				goto l1453
			}
			do(99)
		l1454:
			{
				position1455, thunkPosition1455 := position, thunkPosition
				{
					position1456, thunkPosition1456 := position, thunkPosition
					if !p.rules[ruleIndent]() {
						goto l1455
					}
					position, thunkPosition = position1456, thunkPosition1456
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto l1455
				}
				do(100)
				goto l1454
			l1455:
				position, thunkPosition = position1455, thunkPosition1455
			}
			do(101)
			doarg(yyPop, 2)
			return true
		l1453:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 239 InlineNote <- (&{ p.extension.Notes } '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = mk_list(NOTE, a)
                  yy.contents.str = "" }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !( p.extension.Notes ) {
				goto l1457
			}
			if !matchString("^[") {
				goto l1457
			}
			if !p.rules[ruleStartList]() {
				goto l1457
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto l1457
			}
			if !p.rules[ruleInline]() {
				goto l1457
			}
			do(102)
		l1458:
			{
				position1459, thunkPosition1459 := position, thunkPosition
				if peekChar(']') {
					goto l1459
				}
				if !p.rules[ruleInline]() {
					goto l1459
				}
				do(102)
				goto l1458
			l1459:
				position, thunkPosition = position1459, thunkPosition1459
			}
			if !matchChar(']') {
				goto l1457
			}
			do(103)
			doarg(yyPop, 1)
			return true
		l1457:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 240 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1460
			}
			doarg(yySet, -1)
		l1461:
			{
				position1462, thunkPosition1462 := position, thunkPosition
				{
					position1463, thunkPosition1463 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto l1464
					}
					doarg(yySet, -2)
					do(104)
					goto l1463
				l1464:
					position, thunkPosition = position1463, thunkPosition1463
					if !p.rules[ruleSkipBlock]() {
						goto l1462
					}
				}
			l1463:
				goto l1461
			l1462:
				position, thunkPosition = position1462, thunkPosition1462
			}
			do(105)
			if !(commit(thunkPosition0)) {
				goto l1460
			}
			doarg(yyPop, 2)
			return true
		l1460:
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
				goto l1465
			}
			doarg(yySet, -1)
			{
				position1468, thunkPosition1468 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l1468
				}
				goto l1465
			l1468:
				position, thunkPosition = position1468, thunkPosition1468
			}
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l1465
			}
			do(106)
		l1466:
			{
				position1467, thunkPosition1467 := position, thunkPosition
				{
					position1469, thunkPosition1469 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1469
					}
					goto l1467
				l1469:
					position, thunkPosition = position1469, thunkPosition1469
				}
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto l1467
				}
				do(106)
				goto l1466
			l1467:
				position, thunkPosition = position1467, thunkPosition1467
			}
			begin = position
		l1470:
			{
				position1471, thunkPosition1471 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l1471
				}
				goto l1470
			l1471:
				position, thunkPosition = position1471, thunkPosition1471
			}
			end = position
			do(107)
			do(108)
			doarg(yyPop, 1)
			return true
		l1465:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 242 DefinitionList <- (&{ p.extension.Dlists } StartList (Definition { a = cons(yy, a) })+ { yy = mk_list(DEFINITIONLIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !( p.extension.Dlists ) {
				goto l1472
			}
			if !p.rules[ruleStartList]() {
				goto l1472
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto l1472
			}
			do(109)
		l1473:
			{
				position1474, thunkPosition1474 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto l1474
				}
				do(109)
				goto l1473
			l1474:
				position, thunkPosition = position1474, thunkPosition1474
			}
			do(110)
			doarg(yyPop, 1)
			return true
		l1472:
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
				position1476, thunkPosition1476 := position, thunkPosition
				{
					position1479, thunkPosition1479 := position, thunkPosition
					if !p.rules[ruleDefmark]() {
						goto l1479
					}
					goto l1475
				l1479:
					position, thunkPosition = position1479, thunkPosition1479
				}
				if !p.rules[ruleRawLine]() {
					goto l1475
				}
			l1477:
				{
					position1478, thunkPosition1478 := position, thunkPosition
					{
						position1480, thunkPosition1480 := position, thunkPosition
						if !p.rules[ruleDefmark]() {
							goto l1480
						}
						goto l1478
					l1480:
						position, thunkPosition = position1480, thunkPosition1480
					}
					if !p.rules[ruleRawLine]() {
						goto l1478
					}
					goto l1477
				l1478:
					position, thunkPosition = position1478, thunkPosition1478
				}
				{
					position1481, thunkPosition1481 := position, thunkPosition
					if !p.rules[ruleBlankLine]() {
						goto l1481
					}
					goto l1482
				l1481:
					position, thunkPosition = position1481, thunkPosition1481
				}
			l1482:
				if !p.rules[ruleDefmark]() {
					goto l1475
				}
				position, thunkPosition = position1476, thunkPosition1476
			}
			if !p.rules[ruleStartList]() {
				goto l1475
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto l1475
			}
			do(111)
		l1483:
			{
				position1484, thunkPosition1484 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto l1484
				}
				do(111)
				goto l1483
			l1484:
				position, thunkPosition = position1484, thunkPosition1484
			}
			{
				position1485, thunkPosition1485 := position, thunkPosition
				if !p.rules[ruleDefTight]() {
					goto l1486
				}
				goto l1485
			l1486:
				position, thunkPosition = position1485, thunkPosition1485
				if !p.rules[ruleDefLoose]() {
					goto l1475
				}
			}
		l1485:
			do(112)
			do(113)
			doarg(yyPop, 1)
			return true
		l1475:
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
				goto l1487
			}
			{
				position1488, thunkPosition1488 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1488
				}
				goto l1487
			l1488:
				position, thunkPosition = position1488, thunkPosition1488
			}
			{
				position1489, thunkPosition1489 := position, thunkPosition
				if !p.rules[ruleNonspacechar]() {
					goto l1487
				}
				position, thunkPosition = position1489, thunkPosition1489
			}
			if !p.rules[ruleStartList]() {
				goto l1487
			}
			doarg(yySet, -1)
			{
				position1492, thunkPosition1492 := position, thunkPosition
				if !p.rules[ruleEndline]() {
					goto l1492
				}
				goto l1487
			l1492:
				position, thunkPosition = position1492, thunkPosition1492
			}
			if !p.rules[ruleInline]() {
				goto l1487
			}
			do(114)
		l1490:
			{
				position1491, thunkPosition1491 := position, thunkPosition
				{
					position1493, thunkPosition1493 := position, thunkPosition
					if !p.rules[ruleEndline]() {
						goto l1493
					}
					goto l1491
				l1493:
					position, thunkPosition = position1493, thunkPosition1493
				}
				if !p.rules[ruleInline]() {
					goto l1491
				}
				do(114)
				goto l1490
			l1491:
				position, thunkPosition = position1491, thunkPosition1491
			}
			if !p.rules[ruleSp]() {
				goto l1487
			}
			if !p.rules[ruleNewline]() {
				goto l1487
			}
			do(115)
			doarg(yyPop, 1)
			return true
		l1487:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 245 DefTight <- (&Defmark ListTight) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1495, thunkPosition1495 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1494
				}
				position, thunkPosition = position1495, thunkPosition1495
			}
			if !p.rules[ruleListTight]() {
				goto l1494
			}
			return true
		l1494:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 246 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlankLine]() {
				goto l1496
			}
			{
				position1497, thunkPosition1497 := position, thunkPosition
				if !p.rules[ruleDefmark]() {
					goto l1496
				}
				position, thunkPosition = position1497, thunkPosition1497
			}
			if !p.rules[ruleListLoose]() {
				goto l1496
			}
			return true
		l1496:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 247 Defmark <- (NonindentSpace (':' / '~') Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleNonindentSpace]() {
				goto l1498
			}
			{
				position1499, thunkPosition1499 := position, thunkPosition
				if !matchChar(':') {
					goto l1500
				}
				goto l1499
			l1500:
				position, thunkPosition = position1499, thunkPosition1499
				if !matchChar('~') {
					goto l1498
				}
			}
		l1499:
			if !p.rules[ruleSpacechar]() {
				goto l1498
			}
		l1501:
			{
				position1502, thunkPosition1502 := position, thunkPosition
				if !p.rules[ruleSpacechar]() {
					goto l1502
				}
				goto l1501
			l1502:
				position, thunkPosition = position1502, thunkPosition1502
			}
			return true
		l1498:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 248 DefMarker <- (&{ p.extension.Dlists } Defmark) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !( p.extension.Dlists ) {
				goto l1503
			}
			if !p.rules[ruleDefmark]() {
				goto l1503
			}
			return true
		l1503:
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
