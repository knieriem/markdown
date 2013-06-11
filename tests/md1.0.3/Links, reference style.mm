.P
Foo bar (/url/)\[char46]
.P
Foo bar (/url/)\[char46]
.P
Foo bar (/url/)\[char46]
.P
With embedded [brackets] (/url/)\[char46]
.P
Indented once (/url)\[char46]
.P
Indented twice (/url)\[char46]
.P
Indented thrice (/url)\[char46]
.P
Indented [four][] times.
.VERBON 2
[four]: /url
.VERBOFF
\l'\n(.lu*8u/10u'
.P
this (foo) should work
.P
So should this (foo)\[char46]
.P
And this (foo)\[char46]
.P
And this (foo)\[char46]
.P
And this (foo)\[char46]
.P
But not [that] []\[char46]
.P
Nor [that][]\[char46]
.P
Nor [that]\[char46]
.P
[Something in brackets like this (foo) should work]
.P
[Same with this (foo)\[char46]]
.P
In this case, this (/somethingelse/) points to something else.
.P
Backslashing should suppress [this] and [this]\[char46]
\l'\n(.lu*8u/10u'
.P
Here's one where the link
breaks (/url/) across lines.
.P
Here's another where the link
breaks (/url/) across lines, but with a line-ending space.
