.H 1 "Markdown: Syntax"
.BL
.LI
Overview (#overview)
.BL
.LI
Philosophy (#philosophy)
.LI
Inline HTML (#html)
.LI
Automatic Escaping for Special Characters (#autoescape)
.LE 1
.LI
Block Elements (#block)
.BL
.LI
Paragraphs and Line Breaks (#p)
.LI
Headers (#header)
.LI
Blockquotes (#blockquote)
.LI
Lists (#list)
.LI
Code Blocks (#precode)
.LI
Horizontal Rules (#hr)
.LE 1
.LI
Span Elements (#span)
.BL
.LI
Links (#link)
.LI
Emphasis (#em)
.LI
Code (#code)
.LI
Images (#img)
.LE 1
.LI
Miscellaneous (#misc)
.BL
.LI
Backslash Escapes (#backslash)
.LI
Automatic Links (#autolink)
.LE 1
.LE 1
.P
\fBNote:\fR This document is itself written using Markdown; you
can see the source for it by adding '.text' to the URL (/projects/markdown/syntax.text).
\l'\n(.lu*8u/10u'
.P
Markdown is intended to be as easy-to-read and easy-to-write as is feasible.
.P
Readability, however, is emphasized above all else. A Markdown-formatted
document should be publishable as-is, as plain text, without looking
like it's been marked up with tags or formatting instructions. While
Markdown's syntax has been influenced by several existing text-to-HTML
filters -- including Setext (http://docutils.sourceforge.net/mirror/setext.html), atx (http://www.aaronsw.com/2002/atx/), Textile (http://textism.com/tools/textile/), reStructuredText (http://docutils.sourceforge.net/rst.html),
Grutatext (http://www.triptico.com/software/grutatxt.html), and EtText (http://ettext.taint.org/doc/) -- the single biggest source of
inspiration for Markdown's syntax is the format of plain text email.
.P
To this end, Markdown's syntax is comprised entirely of punctuation
characters, which punctuation characters have been carefully chosen so
as to look like what they mean. E.g., asterisks around a word actually
look like *emphasis*. Markdown lists look like, well, lists. Even
blockquotes look like quoted passages of text, assuming you've ever
used email.
.P
Markdown's syntax is intended for one purpose: to be used as a
format for \fIwriting\fR for the web.
.P
Markdown is not a replacement for HTML, or even close to it. Its
syntax is very small, corresponding only to a very small subset of
HTML tags. The idea is \fInot\fR to create a syntax that makes it easier
to insert HTML tags. In my opinion, HTML tags are already easy to
insert. The idea for Markdown is to make it easy to read, write, and
edit prose. HTML is a \fIpublishing\fR format; Markdown is a \fIwriting\fR
format. Thus, Markdown's formatting syntax only addresses issues that
can be conveyed in plain text.
.P
For any markup that is not covered by Markdown's syntax, you simply
use HTML itself. There's no need to preface it or delimit it to
indicate that you're switching from Markdown to HTML; you just use
the tags.
.P
The only restrictions are that block-level HTML elements -- e.g. \fC<div>\fR,
\fC<table>\fR, \fC<pre>\fR, \fC<p>\fR, etc. -- must be separated from surrounding
content by blank lines, and the start and end tags of the block should
not be indented with tabs or spaces. Markdown is smart enough not
to add extra (unwanted) \fC<p>\fR tags around HTML block-level tags.
.P
For example, to add an HTML table to a Markdown article:
.VERBON 2
This is a regular paragraph.

<table>
    <tr>
        <td>Foo</td>
    </tr>
</table>

This is another regular paragraph.
.VERBOFF
.P
Note that Markdown formatting syntax is not processed within block-level
HTML tags. E.g., you can't use Markdown-style \fC*emphasis*\fR inside an
HTML block.
.P
Span-level HTML tags -- e.g. \fC<span>\fR, \fC<cite>\fR, or \fC<del>\fR -- can be
used anywhere in a Markdown paragraph, list item, or header. If you
want, you can even use HTML tags instead of Markdown formatting; e.g. if
you'd prefer to use HTML \fC<a>\fR or \fC<img>\fR tags instead of Markdown's
link or image syntax, go right ahead.
.P
Unlike block-level HTML tags, Markdown syntax \fIis\fR processed within
span-level tags.
.P
In HTML, there are two characters that demand special treatment: \fC<\fR
and \fC&\fR. Left angle brackets are used to start tags; ampersands are
used to denote HTML entities. If you want to use them as literal
characters, you must escape them as entities, e.g. \fC&lt;\fR, and
\fC&amp;\fR.
.P
Ampersands in particular are bedeviling for web writers. If you want to
write about 'AT&T', you need to write '\fCAT&amp;T\fR'. You even need to
escape ampersands within URLs. Thus, if you want to link to:
.VERBON 2
http://images.google.com/images?num=30&q=larry+bird
.VERBOFF
.P
you need to encode the URL as:
.VERBON 2
http://images.google.com/images?num=30&amp;q=larry+bird
.VERBOFF
.P
in your anchor tag \fChref\fR attribute. Needless to say, this is easy to
forget, and is probably the single most common source of HTML validation
errors in otherwise well-marked-up web sites.
.P
Markdown allows you to use these characters naturally, taking care of
all the necessary escaping for you. If you use an ampersand as part of
an HTML entity, it remains unchanged; otherwise it will be translated
into \fC&amp;\fR.
.P
So, if you want to include a copyright symbol in your article, you can write:
.VERBON 2
&copy;
.VERBOFF
.P
and Markdown will leave it alone. But if you write:
.VERBON 2
AT&T
.VERBOFF
.P
Markdown will translate it to:
.VERBON 2
AT&amp;T
.VERBOFF
.P
Similarly, because Markdown supports inline HTML (#html), if you use
angle brackets as delimiters for HTML tags, Markdown will treat them as
such. But if you write:
.VERBON 2
4 < 5
.VERBOFF
.P
Markdown will translate it to:
.VERBON 2
4 &lt; 5
.VERBOFF
.P
However, inside Markdown code spans and blocks, angle brackets and
ampersands are \fIalways\fR encoded automatically. This makes it easy to use
Markdown to write about HTML code. (As opposed to raw HTML, which is a
terrible format for writing about HTML syntax, because every single \fC<\fR
and \fC&\fR in your example code needs to be escaped.)
\l'\n(.lu*8u/10u'
.P
A paragraph is simply one or more consecutive lines of text, separated
by one or more blank lines. (A blank line is any line that looks like a
blank line -- a line containing nothing but spaces or tabs is considered
blank.) Normal paragraphs should not be intended with spaces or tabs.
.P
The implication of the "one or more consecutive lines of text" rule is
that Markdown supports "hard-wrapped" text paragraphs. This differs
significantly from most other text-to-HTML formatters (including Movable
Type's "Convert Line Breaks" option) which translate every line break
character in a paragraph into a \fC<br />\fR tag.
.P
When you \fIdo\fR want to insert a \fC<br />\fR break tag using Markdown, you
end a line with two or more spaces, then type return.
.P
Yes, this takes a tad more effort to create a \fC<br />\fR, but a simplistic
"every line break is a \fC<br />\fR" rule wouldn't work for Markdown.
Markdown's email-style blockquoting (#blockquote) and multi-paragraph list items (#list)
work best -- and look better -- when you format them with hard breaks.
.P
Markdown supports two styles of headers, Setext (http://docutils.sourceforge.net/mirror/setext.html) and atx (http://www.aaronsw.com/2002/atx/).
.P
Setext-style headers are "underlined" using equal signs (for first-level
headers) and dashes (for second-level headers). For example:
.VERBON 2
This is an H1
=============

This is an H2
-------------
.VERBOFF
.P
Any number of underlining \fC=\fR's or \fC-\fR's will work.
.P
Atx-style headers use 1-6 hash characters at the start of the line,
corresponding to header levels 1-6. For example:
.VERBON 2
# This is an H1

## This is an H2

###### This is an H6
.VERBOFF
.P
Optionally, you may "close" atx-style headers. This is purely
cosmetic -- you can use this if you think it looks better. The
closing hashes don't even need to match the number of hashes
used to open the header. (The number of opening hashes
determines the header level.) :
.VERBON 2
# This is an H1 #

## This is an H2 ##

### This is an H3 ######
.VERBOFF
.P
Markdown uses email-style \fC>\fR characters for blockquoting. If you're
familiar with quoting passages of text in an email message, then you
know how to create a blockquote in Markdown. It looks best if you hard
wrap the text and put a \fC>\fR before every line:
.VERBON 2
> This is a blockquote with two paragraphs. Lorem ipsum dolor sit amet,
> consectetuer adipiscing elit. Aliquam hendrerit mi posuere lectus.
> Vestibulum enim wisi, viverra nec, fringilla in, laoreet vitae, risus.
> 
> Donec sit amet nisl. Aliquam semper ipsum sit amet velit. Suspendisse
> id sem consectetuer libero luctus adipiscing.
.VERBOFF
.P
Markdown allows you to be lazy and only put the \fC>\fR before the first
line of a hard-wrapped paragraph:
.VERBON 2
> This is a blockquote with two paragraphs. Lorem ipsum dolor sit amet,
consectetuer adipiscing elit. Aliquam hendrerit mi posuere lectus.
Vestibulum enim wisi, viverra nec, fringilla in, laoreet vitae, risus.

> Donec sit amet nisl. Aliquam semper ipsum sit amet velit. Suspendisse
id sem consectetuer libero luctus adipiscing.
.VERBOFF
.P
Blockquotes can be nested (i.e. a blockquote-in-a-blockquote) by
adding additional levels of \fC>\fR:
.VERBON 2
> This is the first level of quoting.
>
> > This is nested blockquote.
>
> Back to the first level.
.VERBOFF
.P
Blockquotes can contain other Markdown elements, including headers, lists,
and code blocks:
.VERBON 2
> ## This is a header.
> 
> 1.   This is the first list item.
> 2.   This is the second list item.
> 
> Here's some example code:
> 
>     return shell_exec("echo $input | $markdown_script");
.VERBOFF
.P
Any decent text editor should make email-style quoting easy. For
example, with BBEdit, you can make a selection and choose Increase
Quote Level from the Text menu.
.P
Markdown supports ordered (numbered) and unordered (bulleted) lists.
.P
Unordered lists use asterisks, pluses, and hyphens -- interchangably
-- as list markers:
.VERBON 2
*   Red
*   Green
*   Blue
.VERBOFF
.P
is equivalent to:
.VERBON 2
+   Red
+   Green
+   Blue
.VERBOFF
.P
and:
.VERBON 2
-   Red
-   Green
-   Blue
.VERBOFF
.P
Ordered lists use numbers followed by periods:
.VERBON 2
1.  Bird
2.  McHale
3.  Parish
.VERBOFF
.P
It's important to note that the actual numbers you use to mark the
list have no effect on the HTML output Markdown produces. The HTML
Markdown produces from the above list is:
.VERBON 2
<ol>
<li>Bird</li>
<li>McHale</li>
<li>Parish</li>
</ol>
.VERBOFF
.P
If you instead wrote the list in Markdown like this:
.VERBON 2
1.  Bird
1.  McHale
1.  Parish
.VERBOFF
.P
or even:
.VERBON 2
3. Bird
1. McHale
8. Parish
.VERBOFF
.P
you'd get the exact same HTML output. The point is, if you want to,
you can use ordinal numbers in your ordered Markdown lists, so that
the numbers in your source match the numbers in your published HTML.
But if you want to be lazy, you don't have to.
.P
If you do use lazy list numbering, however, you should still start the
list with the number 1. At some point in the future, Markdown may support
starting ordered lists at an arbitrary number.
.P
List markers typically start at the left margin, but may be indented by
up to three spaces. List markers must be followed by one or more spaces
or a tab.
.P
To make lists look nice, you can wrap items with hanging indents:
.VERBON 2
*   Lorem ipsum dolor sit amet, consectetuer adipiscing elit.
    Aliquam hendrerit mi posuere lectus. Vestibulum enim wisi,
    viverra nec, fringilla in, laoreet vitae, risus.
*   Donec sit amet nisl. Aliquam semper ipsum sit amet velit.
    Suspendisse id sem consectetuer libero luctus adipiscing.
.VERBOFF
.P
But if you want to be lazy, you don't have to:
.VERBON 2
*   Lorem ipsum dolor sit amet, consectetuer adipiscing elit.
Aliquam hendrerit mi posuere lectus. Vestibulum enim wisi,
viverra nec, fringilla in, laoreet vitae, risus.
*   Donec sit amet nisl. Aliquam semper ipsum sit amet velit.
Suspendisse id sem consectetuer libero luctus adipiscing.
.VERBOFF
.P
If list items are separated by blank lines, Markdown will wrap the
items in \fC<p>\fR tags in the HTML output. For example, this input:
.VERBON 2
*   Bird
*   Magic
.VERBOFF
.P
will turn into:
.VERBON 2
<ul>
<li>Bird</li>
<li>Magic</li>
</ul>
.VERBOFF
.P
But this:
.VERBON 2
*   Bird

*   Magic
.VERBOFF
.P
will turn into:
.VERBON 2
<ul>
<li><p>Bird</p></li>
<li><p>Magic</p></li>
</ul>
.VERBOFF
.P
List items may consist of multiple paragraphs. Each subsequent
paragraph in a list item must be intended by either 4 spaces
or one tab:
.VERBON 2
1.  This is a list item with two paragraphs. Lorem ipsum dolor
    sit amet, consectetuer adipiscing elit. Aliquam hendrerit
    mi posuere lectus.

    Vestibulum enim wisi, viverra nec, fringilla in, laoreet
    vitae, risus. Donec sit amet nisl. Aliquam semper ipsum
    sit amet velit.

2.  Suspendisse id sem consectetuer libero luctus adipiscing.
.VERBOFF
.P
It looks nice if you indent every line of the subsequent
paragraphs, but here again, Markdown will allow you to be
lazy:
.VERBON 2
*   This is a list item with two paragraphs.

    This is the second paragraph in the list item. You're
only required to indent the first line. Lorem ipsum dolor
sit amet, consectetuer adipiscing elit.

*   Another item in the same list.
.VERBOFF
.P
To put a blockquote within a list item, the blockquote's \fC>\fR
delimiters need to be indented:
.VERBON 2
*   A list item with a blockquote:

    > This is a blockquote
    > inside a list item.
.VERBOFF
.P
To put a code block within a list item, the code block needs
to be indented \fItwice\fR -- 8 spaces or two tabs:
.VERBON 2
*   A list item with a code block:

        <code goes here>
.VERBOFF
.P
It's worth noting that it's possible to trigger an ordered list by
accident, by writing something like this:
.VERBON 2
1986. What a great season.
.VERBOFF
.P
In other words, a \fInumber-period-space\fR sequence at the beginning of a
line. To avoid this, you can backslash-escape the period:
.VERBON 2
1986\e. What a great season.
.VERBOFF
.P
Pre-formatted code blocks are used for writing about programming or
markup source code. Rather than forming normal paragraphs, the lines
of a code block are interpreted literally. Markdown wraps a code block
in both \fC<pre>\fR and \fC<code>\fR tags.
.P
To produce a code block in Markdown, simply indent every line of the
block by at least 4 spaces or 1 tab. For example, given this input:
.VERBON 2
This is a normal paragraph:

    This is a code block.
.VERBOFF
.P
Markdown will generate:
.VERBON 2
<p>This is a normal paragraph:</p>

<pre><code>This is a code block.
</code></pre>
.VERBOFF
.P
One level of indentation -- 4 spaces or 1 tab -- is removed from each
line of the code block. For example, this:
.VERBON 2
Here is an example of AppleScript:

    tell application "Foo"
        beep
    end tell
.VERBOFF
.P
will turn into:
.VERBON 2
<p>Here is an example of AppleScript:</p>

<pre><code>tell application "Foo"
    beep
end tell
</code></pre>
.VERBOFF
.P
A code block continues until it reaches a line that is not indented
(or the end of the article).
.P
Within a code block, ampersands (\fC&\fR) and angle brackets (\fC<\fR and \fC>\fR)
are automatically converted into HTML entities. This makes it very
easy to include example HTML source code using Markdown -- just paste
it and indent it, and Markdown will handle the hassle of encoding the
ampersands and angle brackets. For example, this:
.VERBON 2
    <div class="footer">
        &copy; 2004 Foo Corporation
    </div>
.VERBOFF
.P
will turn into:
.VERBON 2
<pre><code>&lt;div class="footer"&gt;
    &amp;copy; 2004 Foo Corporation
&lt;/div&gt;
</code></pre>
.VERBOFF
.P
Regular Markdown syntax is not processed within code blocks. E.g.,
asterisks are just literal asterisks within a code block. This means
it's also easy to use Markdown to write about Markdown's own syntax.
.P
You can produce a horizontal rule tag (\fC<hr />\fR) by placing three or
more hyphens, asterisks, or underscores on a line by themselves. If you
wish, you may use spaces between the hyphens or asterisks. Each of the
following lines will produce a horizontal rule:
.VERBON 2
* * *

***

*****

- - -

---------------------------------------

_ _ _
.VERBOFF
\l'\n(.lu*8u/10u'
.P
Markdown supports two style of links: \fIinline\fR and \fIreference\fR.
.P
In both styles, the link text is delimited by [square brackets].
.P
To create an inline link, use a set of regular parentheses immediately
after the link text's closing square bracket. Inside the parentheses,
put the URL where you want the link to point, along with an \fIoptional\fR
title for the link, surrounded in quotes. For example:
.VERBON 2
This is [an example](http://example.com/ "Title") inline link.

[This link](http://example.net/) has no title attribute.
.VERBOFF
.P
Will produce:
.VERBON 2
<p>This is <a href="http://example.com/" title="Title">
an example</a> inline link.</p>

<p><a href="http://example.net/">This link</a> has no
title attribute.</p>
.VERBOFF
.P
If you're referring to a local resource on the same server, you can
use relative paths:
.VERBON 2
See my [About](/about/) page for details.
.VERBOFF
.P
Reference-style links use a second set of square brackets, inside
which you place a label of your choosing to identify the link:
.VERBON 2
This is [an example][id] reference-style link.
.VERBOFF
.P
You can optionally use a space to separate the sets of brackets:
.VERBON 2
This is [an example] [id] reference-style link.
.VERBOFF
.P
Then, anywhere in the document, you define your link label like this,
on a line by itself:
.VERBON 2
[id]: http://example.com/  "Optional Title Here"
.VERBOFF
.P
That is:
.BL
.LI
Square brackets containing the link identifier (optionally
indented from the left margin using up to three spaces);
.LI
followed by a colon;
.LI
followed by one or more spaces (or tabs);
.LI
followed by the URL for the link;
.LI
optionally followed by a title attribute for the link, enclosed
in double or single quotes.
.LE 1
.P
The link URL may, optionally, be surrounded by angle brackets:
.VERBON 2
[id]: <http://example.com/>  "Optional Title Here"
.VERBOFF
.P
You can put the title attribute on the next line and use extra spaces
or tabs for padding, which tends to look better with longer URLs:
.VERBON 2
[id]: http://example.com/longish/path/to/resource/here
    "Optional Title Here"
.VERBOFF
.P
Link definitions are only used for creating links during Markdown
processing, and are stripped from your document in the HTML output.
.P
Link definition names may constist of letters, numbers, spaces, and punctuation -- but they are \fInot\fR case sensitive. E.g. these two links:
.VERBON 2
[link text][a]
[link text][A]
.VERBOFF
.P
are equivalent.
.P
The \fIimplicit link name\fR shortcut allows you to omit the name of the
link, in which case the link text itself is used as the name.
Just use an empty set of square brackets -- e.g., to link the word
"Google" to the google.com web site, you could simply write:
.VERBON 2
[Google][]
.VERBOFF
.P
And then define the link:
.VERBON 2
[Google]: http://google.com/
.VERBOFF
.P
Because link names may contain spaces, this shortcut even works for
multiple words in the link text:
.VERBON 2
Visit [Daring Fireball][] for more information.
.VERBOFF
.P
And then define the link:
.VERBON 2
[Daring Fireball]: http://daringfireball.net/
.VERBOFF
.P
Link definitions can be placed anywhere in your Markdown document. I
tend to put them immediately after each paragraph in which they're
used, but if you want, you can put them all at the end of your
document, sort of like footnotes.
.P
Here's an example of reference links in action:
.VERBON 2
I get 10 times more traffic from [Google] [1] than from
[Yahoo] [2] or [MSN] [3].

  [1]: http://google.com/        "Google"
  [2]: http://search.yahoo.com/  "Yahoo Search"
  [3]: http://search.msn.com/    "MSN Search"
.VERBOFF
.P
Using the implicit link name shortcut, you could instead write:
.VERBON 2
I get 10 times more traffic from [Google][] than from
[Yahoo][] or [MSN][].

  [google]: http://google.com/        "Google"
  [yahoo]:  http://search.yahoo.com/  "Yahoo Search"
  [msn]:    http://search.msn.com/    "MSN Search"
.VERBOFF
.P
Both of the above examples will produce the following HTML output:
.VERBON 2
<p>I get 10 times more traffic from <a href="http://google.com/"
title="Google">Google</a> than from
<a href="http://search.yahoo.com/" title="Yahoo Search">Yahoo</a>
or <a href="http://search.msn.com/" title="MSN Search">MSN</a>.</p>
.VERBOFF
.P
For comparison, here is the same paragraph written using
Markdown's inline link style:
.VERBON 2
I get 10 times more traffic from [Google](http://google.com/ "Google")
than from [Yahoo](http://search.yahoo.com/ "Yahoo Search") or
[MSN](http://search.msn.com/ "MSN Search").
.VERBOFF
.P
The point of reference-style links is not that they're easier to
write. The point is that with reference-style links, your document
source is vastly more readable. Compare the above examples: using
reference-style links, the paragraph itself is only 81 characters
long; with inline-style links, it's 176 characters; and as raw HTML,
it's 234 characters. In the raw HTML, there's more markup than there
is text.
.P
With Markdown's reference-style links, a source document much more
closely resembles the final output, as rendered in a browser. By
allowing you to move the markup-related metadata out of the paragraph,
you can add links without interrupting the narrative flow of your
prose.
.P
Markdown treats asterisks (\fC*\fR) and underscores (\fC_\fR) as indicators of
emphasis. Text wrapped with one \fC*\fR or \fC_\fR will be wrapped with an
HTML \fC<em>\fR tag; double \fC*\fR's or \fC_\fR's will be wrapped with an HTML
\fC<strong>\fR tag. E.g., this input:
.VERBON 2
*single asterisks*

_single underscores_

**double asterisks**

__double underscores__
.VERBOFF
.P
will produce:
.VERBON 2
<em>single asterisks</em>

<em>single underscores</em>

<strong>double asterisks</strong>

<strong>double underscores</strong>
.VERBOFF
.P
You can use whichever style you prefer; the lone restriction is that
the same character must be used to open and close an emphasis span.
.P
Emphasis can be used in the middle of a word:
.VERBON 2
un*fucking*believable
.VERBOFF
.P
But if you surround an \fC*\fR or \fC_\fR with spaces, it'll be treated as a
literal asterisk or underscore.
.P
To produce a literal asterisk or underscore at a position where it
would otherwise be used as an emphasis delimiter, you can backslash
escape it:
.VERBON 2
\e*this text is surrounded by literal asterisks\e*
.VERBOFF
.P
To indicate a span of code, wrap it with backtick quotes (\fC`\fR).
Unlike a pre-formatted code block, a code span indicates code within a
normal paragraph. For example:
.VERBON 2
Use the `printf()` function.
.VERBOFF
.P
will produce:
.VERBON 2
<p>Use the <code>printf()</code> function.</p>
.VERBOFF
.P
To include a literal backtick character within a code span, you can use
multiple backticks as the opening and closing delimiters:
.VERBON 2
``There is a literal backtick (`) here.``
.VERBOFF
.P
which will produce this:
.VERBON 2
<p><code>There is a literal backtick (`) here.</code></p>
.VERBOFF
.P
The backtick delimiters surrounding a code span may include spaces --
one after the opening, one before the closing. This allows you to place
literal backtick characters at the beginning or end of a code span:
.VERBON 2
A single backtick in a code span: `` ` ``

A backtick-delimited string in a code span: `` `foo` ``
.VERBOFF
.P
will produce:
.VERBON 2
<p>A single backtick in a code span: <code>`</code></p>

<p>A backtick-delimited string in a code span: <code>`foo`</code></p>
.VERBOFF
.P
With a code span, ampersands and angle brackets are encoded as HTML
entities automatically, which makes it easy to include example HTML
tags. Markdown will turn this:
.VERBON 2
Please don't use any `<blink>` tags.
.VERBOFF
.P
into:
.VERBON 2
<p>Please don't use any <code>&lt;blink&gt;</code> tags.</p>
.VERBOFF
.P
You can write this:
.VERBON 2
`&#8212;` is the decimal-encoded equivalent of `&mdash;`.
.VERBOFF
.P
to produce:
.VERBON 2
<p><code>&amp;#8212;</code> is the decimal-encoded
equivalent of <code>&amp;mdash;</code>.</p>
.VERBOFF
.P
Admittedly, it's fairly difficult to devise a "natural" syntax for
placing images into a plain text document format.
.P
Markdown uses an image syntax that is intended to resemble the syntax
for links, allowing for two styles: \fIinline\fR and \fIreference\fR.
.P
Inline image syntax looks like this:
.VERBON 2
![Alt text](/path/to/img.jpg)

![Alt text](/path/to/img.jpg "Optional title")
.VERBOFF
.P
That is:
.BL
.LI
An exclamation mark: \fC!\fR;
.LI
followed by a set of square brackets, containing the \fCalt\fR
attribute text for the image;
.LI
followed by a set of parentheses, containing the URL or path to
the image, and an optional \fCtitle\fR attribute enclosed in double
or single quotes.
.LE 1
.P
Reference-style image syntax looks like this:
.VERBON 2
![Alt text][id]
.VERBOFF
.P
Where "id" is the name of a defined image reference. Image references
are defined using syntax identical to link references:
.VERBON 2
[id]: url/to/image  "Optional title attribute"
.VERBOFF
.P
As of this writing, Markdown has no syntax for specifying the
dimensions of an image; if this is important to you, you can simply
use regular HTML \fC<img>\fR tags.
\l'\n(.lu*8u/10u'
.P
Markdown supports a shortcut style for creating "automatic" links for URLs and email addresses: simply surround the URL or email address with angle brackets. What this means is that if you want to show the actual text of a URL or email address, and also have it be a clickable link, you can do this:
.VERBON 2
<http://example.com/>
.VERBOFF
.P
Markdown will turn this into:
.VERBON 2
<a href="http://example.com/">http://example.com/</a>
.VERBOFF
.P
Automatic links for email addresses work similarly, except that
Markdown will also perform a bit of randomized decimal and hex
entity-encoding to help obscure your address from address-harvesting
spambots. For example, Markdown will turn this:
.VERBON 2
<address@example.com>
.VERBOFF
.P
into something like this:
.VERBON 2
<a href="&#x6D;&#x61;i&#x6C;&#x74;&#x6F;:&#x61;&#x64;&#x64;&#x72;&#x65;
&#115;&#115;&#64;&#101;&#120;&#x61;&#109;&#x70;&#x6C;e&#x2E;&#99;&#111;
&#109;">&#x61;&#x64;&#x64;&#x72;&#x65;&#115;&#115;&#64;&#101;&#120;&#x61;
&#109;&#x70;&#x6C;e&#x2E;&#99;&#111;&#109;</a>
.VERBOFF
.P
which will render in a browser as a clickable link to "address@example.com".
.P
(This sort of entity-encoding trick will indeed fool many, if not
most, address-harvesting bots, but it definitely won't fool all of
them. It's better than nothing, but an address published in this way
will probably eventually start receiving spam.)
.P
Markdown allows you to use backslash escapes to generate literal
characters which would otherwise have special meaning in Markdown's
formatting syntax. For example, if you wanted to surround a word with
literal asterisks (instead of an HTML \fC<em>\fR tag), you can backslashes
before the asterisks, like this:
.VERBON 2
\e*literal asterisks\e*
.VERBOFF
.P
Markdown provides backslash escapes for the following characters:
.VERBON 2
\e   backslash
`   backtick
*   asterisk
_   underscore
{}  curly braces
[]  square brackets
()  parentheses
#   hash mark
+   plus sign
-   minus sign (hyphen)
.   dot
!   exclamation mark
.VERBOFF
