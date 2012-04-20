This is an implementation of John Gruber's [markdown][] in
[Go][].  It is a translation of [peg-markdown][], written by
John MacFarlane in C, into Go.  It is using a modified version
of Andrew J Snodgrass' PEG parser [peg][] -- now supporting
LEG grammars --, which is itself based on the parser used
by peg-markdown.

[markdown]: http://daringfireball.net/projects/markdown/
[peg-markdown]: https://github.com/jgm/peg-markdown
[peg]: https://github.com/pointlander/peg
[Go]: http://golang.org/

Support for HTML output is implemented, but Groff and LaTeX
output have not been ported. The output should be identical
to that of peg-markdown.

A simple benchmark has been done by comparing the
execution time of the Go binary (cmd/main.go) and the
original C implementation's binary needed for processing
a Markdown document, which had been created by
concatenating ten [Markdown syntax descriptions][syntax].

  [syntax]: http://daringfireball.net/projects/markdown/syntax.text

The C version is still around 1.9x faster than the Go version.


## Installation

Provided you have a copy of Go 1, and git is available,

	go get github.com/knieriem/markdown

should download and install the package according to
your GOPATH settings.

See doc.go for an example how to use the package.

To create the command line program *markdown,* run

	go build github.com/knieriem/markdown/cmd/markdown

the binary should then be available in the current directory.

To run the Markdown 1.0.3 test suite, type

	make test

This will download peg-markdown, in case you have Mercurial
and the hg-git extension available, build cmd/markdown, and
run the test suite.

The test suite should fail on exactly one test –
*Ordered and unordered lists* –, for the same reason which
applies to peg-markdown, because the grammar is the same.
See the [original README][] for details.

[original README]: https://github.com/jgm/peg-markdown/blob/master/README.markdown


## Development

There is not yet a way to create a Go source file like
`parser.leg.go` automatically from another file, `parser.leg`,
when building packages and commands using the Go tool.  To make
*markdown* installable using `go get`, `parser.leg.go` has
been added to the VCS.

`Make parser` will update `parser.leg.go` using `leg`, which
is part of [knieriem/peg][] at github, if parser.leg has
been changed, or if the Go file is missing. If a copy of *peg*
is not yet present on your system, run

	go get github.com/knieriem/peg

Then `make parser` should succeed.

[knieriem/peg]: https://github.com/knieriem/peg


## Extensions

In addition to the extensions already present in peg-markdown,
this package also supports definition lists (option `-dlists`)
similar to the way they are described in the documentation of
[PHP Markdown Extra][].

Definitions (`<dd>...</dd>`) are implemented using [`ListTight`][ListTight]
and `ListLoose`, on which bullet lists and ordered lists are based
already. If there is an empty line between the definition title and
the first definition, a loose list is expected, a tight list otherwise.

As definition item markers both `:` and `~` can be used.

[PHP Markdown Extra]: http://michelf.com/projects/php-markdown/extra/#def-list
[ListTight]: https://github.com/knieriem/markdown/blob/master/parser.leg#L191


## Todo

*	Implement tables

*	Where appropriate, use more idiomatic Go code

## Subdirectory Index

*	peg – PEG parser generator (modified) from Andrew J Snodgrass

*	peg/leg – LEG parser generator, based on PEG

*	cmd	– command line program `markdown`

