.P
Foo bar (/url/).
.P
Foo bar (/url/).
.P
Foo bar (/url/).
.P
With embedded [brackets] (/url/).
.P
Indented once (/url).
.P
Indented twice (/url).
.P
Indented thrice (/url).
.P
Indented [four][] times.
.VERBON 2
[four]: /url
.VERBOFF
\l'\n(.lu*8u/10u'
.P
this (foo) should work
.P
So should this (foo).
.P
And this (foo).
.P
And this (foo).
.P
And this (foo).
.P
But not [that] [].
.P
Nor [that][].
.P
Nor [that].
.P
[Something in brackets like this (foo) should work]
.P
[Same with this (foo).]
.P
In this case, this (/somethingelse/) points to something else.
.P
Backslashing should suppress [this] and [this].
\l'\n(.lu*8u/10u'
.P
Here's one where the link
breaks (/url/) across lines.
.P
Here's another where the link
breaks (/url/) across lines, but with a line-ending space.
