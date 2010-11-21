#
# development utilities
#
gofmt:
	rc ./misc/gofmt.rc

diff: ,,c
	tkdiff $< parser.leg

,,c:	orig-c-src/markdown_parser.leg
	sed -f misc/c2go.sed < $< > $@

orig-c-src/markdown_parser.leg: orig-c-src


.PHONY: diff gofmt
