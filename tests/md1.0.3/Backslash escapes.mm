.P
These should all get escaped:
.P
Backslash: \e
.P
Backtick: `
.P
Asterisk: *
.P
Underscore: _
.P
Left brace: {
.P
Right brace: }
.P
Left bracket: [
.P
Right bracket: ]
.P
Left paren: (
.P
Right paren: )
.P
Greater-than: >
.P
Hash: #
.P
Period: .
.P
Bang: !
.P
Plus: +
.P
Minus: -
.P
These should not, because they occur within a code block:
.VERBON 2
Backslash: \e\e

Backtick: \e`

Asterisk: \e*

Underscore: \e_

Left brace: \e{

Right brace: \e}

Left bracket: \e[

Right bracket: \e]

Left paren: \e(

Right paren: \e)

Greater-than: \e>

Hash: \e#

Period: \e.

Bang: \e!

Plus: \e+

Minus: \e-
.VERBOFF
.P
Nor should these, which occur in code spans:
.P
Backslash: \fC\e\e\fR
.P
Backtick: \fC\e`\fR
.P
Asterisk: \fC\e*\fR
.P
Underscore: \fC\e_\fR
.P
Left brace: \fC\e{\fR
.P
Right brace: \fC\e}\fR
.P
Left bracket: \fC\e[\fR
.P
Right bracket: \fC\e]\fR
.P
Left paren: \fC\e(\fR
.P
Right paren: \fC\e)\fR
.P
Greater-than: \fC\e>\fR
.P
Hash: \fC\e#\fR
.P
Period: \fC\e.\fR
.P
Bang: \fC\e!\fR
.P
Plus: \fC\e+\fR
.P
Minus: \fC\e-\fR
.P
These should get escaped, even though they're matching pairs for
other Markdown constructs:
.P
*asterisks*
.P
_underscores_
.P
`backticks`
.P
This is a code span with a literal backslash-backtick sequence: \fC\e`\fR
.P
This is a tag with unescaped backticks bar.
.P
This is a tag with backslashes bar.
