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

The Go version is around 3.5x slower than the original C
version.  A marked speed improvement has been achieved by
converting function `preformat` from concatenating strings
to using bytes.Buffer. At other places, where this kind of
modification had been tried, performance did not improve.

## Installation

Provided you have a recent copy of Go, and git is available,

	goinstall github.com/knieriem/markdown

should install the package into
`$GOROOT/src/pkg/github.com/knieriem/markdown`, and build
it.

See doc.go for an example how to use the package.

To create the command line program *markdown,* run

	cd $GOROOT/src/pkg/github.com/knieriem/markdown
	gomake cmd

the binary should then be available in subdirectory *cmd.*

To run the Markdown 1.0.3 test suite, type

	make mdtest

This will download peg-markdown, in case you have `git`
available, build cmd/markdown, and run the test suite.

The test suite will fail on one test, for the same reason which
applies to peg-markdown, because the grammar is the same.
See the [original README][] for details.

[original README]: https://github.com/jgm/peg-markdown/blob/master/README.markdown

## Development

[`Goinstall`][Goinstall] is creating its own Makefiles to build
packages, based on the `.go` files found in the directory.
It would not know about `parser.leg.go`, which had to be built
by `leg` from the `parser.leg` grammar source file first.
Because of this, to make *markdown* installable using
`goinstall`, `parser.leg.go` has been added to the VCS.

`Make` will update `parser.leg.go` using `leg`, which is part of
[knieriem/peg][] at github, if parser.leg has been changed. If
a copy of this package has not yet been downloaded -- i.e. no
directory `./peg` is present --, `make` will perform the
neccessary steps automatically (run `make peg` to manually
download [knieriem/peg][]).

To update [knieriem/peg][] run `gomake update-peg`. This will
fetch available revisions from github, and remove the old
*leg* binary.

[goinstall]: http://golang.org/cmd/goinstall/
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

*	Rename element key identifiers, so that they are not public

*	Where appropriate, use more idiomatic Go code

## Subdirectory Index

*	peg – PEG parser generator (modified) from Andrew J Snodgrass

*	peg/leg – LEG parser generator, based on PEG

*	cmd	– command line program `markdown`

