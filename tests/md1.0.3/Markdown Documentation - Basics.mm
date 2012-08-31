.H 1 "Markdown: Basics"
.H 2 "Getting the Gist of Markdown's Formatting Syntax"
.P
This page offers a brief overview of what it's like to use Markdown.
The syntax page (/projects/markdown/syntax) provides complete, detailed documentation for
every feature, but Markdown should be very easy to pick up simply by
looking at a few examples of it in action. The examples on this page
are written in a before/after style, showing example syntax and the
HTML output produced by Markdown.
.P
It's also helpful to simply try Markdown out; the Dingus (/projects/markdown/dingus) is a
web application that allows you type your own Markdown-formatted text
and translate it to XHTML.
.P
\fBNote:\fR This document is itself written using Markdown; you
can see the source for it by adding '.text' to the URL (/projects/markdown/basics.text).
.H 2 "Paragraphs, Headers, Blockquotes"
.P
A paragraph is simply one or more consecutive lines of text, separated
by one or more blank lines. (A blank line is any line that looks like a
blank line -- a line containing nothing spaces or tabs is considered
blank.) Normal paragraphs should not be intended with spaces or tabs.
.P
Markdown offers two styles of headers: \fISetext\fR and \fIatx\fR.
Setext-style headers for \fC<h1>\fR and \fC<h2>\fR are created by
"underlining" with equal signs (\fC=\fR) and hyphens (\fC-\fR), respectively.
To create an atx-style header, you put 1-6 hash marks (\fC#\fR) at the
beginning of the line -- the number of hashes equals the resulting
HTML header level.
.P
Blockquotes are indicated using email-style '\fC>\fR' angle brackets.
.P
Markdown:
.VERBON 2
A First Level Header
====================

A Second Level Header
---------------------

Now is the time for all good men to come to
the aid of their country. This is just a
regular paragraph.

The quick brown fox jumped over the lazy
dog's back.

### Header 3

> This is a blockquote.
> 
> This is the second paragraph in the blockquote.
>
> ## This is an H2 in a blockquote
.VERBOFF
.P
Output:
.VERBON 2
<h1>A First Level Header</h1>

<h2>A Second Level Header</h2>

<p>Now is the time for all good men to come to
the aid of their country. This is just a
regular paragraph.</p>

<p>The quick brown fox jumped over the lazy
dog's back.</p>

<h3>Header 3</h3>

<blockquote>
    <p>This is a blockquote.</p>

    <p>This is the second paragraph in the blockquote.</p>

    <h2>This is an H2 in a blockquote</h2>
</blockquote>
.VERBOFF
.H 3 "Phrase Emphasis"
.P
Markdown uses asterisks and underscores to indicate spans of emphasis.
.P
Markdown:
.VERBON 2
Some of these words *are emphasized*.
Some of these words _are emphasized also_.

Use two asterisks for **strong emphasis**.
Or, if you prefer, __use two underscores instead__.
.VERBOFF
.P
Output:
.VERBON 2
<p>Some of these words <em>are emphasized</em>.
Some of these words <em>are emphasized also</em>.</p>

<p>Use two asterisks for <strong>strong emphasis</strong>.
Or, if you prefer, <strong>use two underscores instead</strong>.</p>
.VERBOFF
.H 2 "Lists"
.P
Unordered (bulleted) lists use asterisks, pluses, and hyphens (\fC*\fR,
\fC+\fR, and \fC-\fR) as list markers. These three markers are
interchangable; this:
.VERBON 2
*   Candy.
*   Gum.
*   Booze.
.VERBOFF
.P
this:
.VERBON 2
+   Candy.
+   Gum.
+   Booze.
.VERBOFF
.P
and this:
.VERBON 2
-   Candy.
-   Gum.
-   Booze.
.VERBOFF
.P
all produce the same output:
.VERBON 2
<ul>
<li>Candy.</li>
<li>Gum.</li>
<li>Booze.</li>
</ul>
.VERBOFF
.P
Ordered (numbered) lists use regular numbers, followed by periods, as
list markers:
.VERBON 2
1.  Red
2.  Green
3.  Blue
.VERBOFF
.P
Output:
.VERBON 2
<ol>
<li>Red</li>
<li>Green</li>
<li>Blue</li>
</ol>
.VERBOFF
.P
If you put blank lines between items, you'll get \fC<p>\fR tags for the
list item text. You can create multi-paragraph list items by indenting
the paragraphs by 4 spaces or 1 tab:
.VERBON 2
*   A list item.

    With multiple paragraphs.

*   Another item in the list.
.VERBOFF
.P
Output:
.VERBON 2
<ul>
<li><p>A list item.</p>
<p>With multiple paragraphs.</p></li>
<li><p>Another item in the list.</p></li>
</ul>
.VERBOFF
.H 3 "Links"
.P
Markdown supports two styles for creating links: \fIinline\fR and
\fIreference\fR. With both styles, you use square brackets to delimit the
text you want to turn into a link.
.P
Inline-style links use parentheses immediately after the link text.
For example:
.VERBON 2
This is an [example link](http://example.com/).
.VERBOFF
.P
Output:
.VERBON 2
<p>This is an <a href="http://example.com/">
example link</a>.</p>
.VERBOFF
.P
Optionally, you may include a title attribute in the parentheses:
.VERBON 2
This is an [example link](http://example.com/ "With a Title").
.VERBOFF
.P
Output:
.VERBON 2
<p>This is an <a href="http://example.com/" title="With a Title">
example link</a>.</p>
.VERBOFF
.P
Reference-style links allow you to refer to your links by names, which
you define elsewhere in your document:
.VERBON 2
I get 10 times more traffic from [Google][1] than from
[Yahoo][2] or [MSN][3].

[1]: http://google.com/        "Google"
[2]: http://search.yahoo.com/  "Yahoo Search"
[3]: http://search.msn.com/    "MSN Search"
.VERBOFF
.P
Output:
.VERBON 2
<p>I get 10 times more traffic from <a href="http://google.com/"
title="Google">Google</a> than from <a href="http://search.yahoo.com/"
title="Yahoo Search">Yahoo</a> or <a href="http://search.msn.com/"
title="MSN Search">MSN</a>.</p>
.VERBOFF
.P
The title attribute is optional. Link names may contain letters,
numbers and spaces, but are \fInot\fR case sensitive:
.VERBON 2
I start my morning with a cup of coffee and
[The New York Times][NY Times].

[ny times]: http://www.nytimes.com/
.VERBOFF
.P
Output:
.VERBON 2
<p>I start my morning with a cup of coffee and
<a href="http://www.nytimes.com/">The New York Times</a>.</p>
.VERBOFF
.H 3 "Images"
.P
Image syntax is very much like link syntax.
.P
Inline (titles are optional):
.VERBON 2
![alt text](/path/to/img.jpg "Title")
.VERBOFF
.P
Reference-style:
.VERBON 2
![alt text][id]

[id]: /path/to/img.jpg "Title"
.VERBOFF
.P
Both of the above examples produce the same output:
.VERBON 2
<img src="/path/to/img.jpg" alt="alt text" title="Title" />
.VERBOFF
.H 3 "Code"
.P
In a regular paragraph, you can create code span by wrapping text in
backtick quotes. Any ampersands (\fC&\fR) and angle brackets (\fC<\fR or
\fC>\fR) will automatically be translated into HTML entities. This makes
it easy to use Markdown to write about HTML example code:
.VERBON 2
I strongly recommend against using any `<blink>` tags.

I wish SmartyPants used named entities like `&mdash;`
instead of decimal-encoded entites like `&#8212;`.
.VERBOFF
.P
Output:
.VERBON 2
<p>I strongly recommend against using any
<code>&lt;blink&gt;</code> tags.</p>

<p>I wish SmartyPants used named entities like
<code>&amp;mdash;</code> instead of decimal-encoded
entites like <code>&amp;#8212;</code>.</p>
.VERBOFF
.P
To specify an entire block of pre-formatted code, indent every line of
the block by 4 spaces or 1 tab. Just like with code spans, \fC&\fR, \fC<\fR,
and \fC>\fR characters will be escaped automatically.
.P
Markdown:
.VERBON 2
If you want your page to validate under XHTML 1.0 Strict,
you've got to put paragraph tags in your blockquotes:

    <blockquote>
        <p>For example.</p>
    </blockquote>
.VERBOFF
.P
Output:
.VERBON 2
<p>If you want your page to validate under XHTML 1.0 Strict,
you've got to put paragraph tags in your blockquotes:</p>

<pre><code>&lt;blockquote&gt;
    &lt;p&gt;For example.&lt;/p&gt;
&lt;/blockquote&gt;
</code></pre>
.VERBOFF
