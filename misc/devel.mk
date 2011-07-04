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


bmprepare:
	rc ./misc/bmprepare.rc

benchmark: m ,,pmd ,,prevmd
	rc ./misc/benchmark.rc


#
# pprof
#
pprof: cmd m
	./cmd/markdown -cpuprofile /tmp/md.prof <m > /tmp/,,md.out
	@echo gopprof \'--nodefraction=0.1\' ./cmd/markdown /tmp/md.prof
	@echo gopprof ./cmd/markdown /tmp/md.prof

.PHONY:\
	diff\
	gofmt\
	pprof\
