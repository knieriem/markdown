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

The Go version is around 5x slower than the original C
version.  A marked speed improvement has been achieved by
converting function `preformat` from concatenating strings
to using bytes.Buffer. At other places, where this kind of
modification had been tried, performance did not improve. Also,
pre-allocating a large buffer for `element`s didn't show a
significant difference from allocating `element`s one at a time.

## Installation

Provided you have a recent copy of Go, and git is available,

	goinstall github.com/knieriem/markdown

should install the package into
`$GOROOT/src/pkg/github.com/knieriem/markdown`, and build
it. During the build, a copy of [knieriem/peg][] will be
downloaded from github and compiled (`make peg` if done
manually).

**NOTE:** At the moment, goinstall most likely will fail,
as it does not use the package's Makefile, but generates
its own, which is not sufficient as it does not know how
to build parser.leg.go from parser.leg.  As a workaround,
after the failed goinstall, please do the following steps to
finish the installation:

	cd $GOROOT/src/pkg/github.com/knieriem/markdown
	gomake install

See doc.go for an example how to use the package.

To update [knieriem/peg][] run `gomake update-peg`. This
will fetch available revisions from github, and remove the
old *leg* binary.

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
[knieriem/peg]: https://github.com/knieriem/peg


## Extensions

In addition to the extensions already present in peg-markdown,
this package also supports definition lists (option `-dlists`)
similar to the way they are described in the documentation of
[PHP Markdown Extra][].

Definitions (`<dd>...</dd>`) are implemented using [ListTight][]
and `ListLoose`, on which bullet lists and ordered lists are based
already. If there is an empty line between the definition title and
the first definition, a loose list is expected, a tight list otherwise.

As definition item markers both `:` and `~` can be used.

[PHP Markdown Extra]: http://michelf.com/projects/php-markdown/extra/#def-list
[ListTight]: https://github.com/knieriem/markdown/blob/master/parser.leg#L191


## Todo

*	Implement definition lists (work in progress), and perhaps tables

*	Rename element key identifiers, so that they are not public

*	Where appropriate, use more idiomatic Go code

## Subdirectory Index

*	peg – PEG parser generator (modified) from Andrew J Snodgrass

*	peg/leg – LEG parser generator, based on PEG

*	cmd	– command line program `markdown`

