all:
	@echo 'targets: test nuke parser clean'

#
# run MarkdownTests-1.0.3 that come with original C sources
#
test: package cmd orig-c-src
	cd orig-c-src/MarkdownTest_1.0.3; \
	./MarkdownTest.pl --script=../../cmd/markdown/markdown --tidy

cmd: package
	cd cmd/markdown && go build -v

package: parser.leg.go
	go install -v

clean:
	go clean . ./...
	rm -rf orig-c-src
	rm -rf ,,prevmd ,,pmd
	
parser:	parser.leg.go

nuke:
	rm -f parser.leg.go


# LEG parser rules
#
ifeq ($(MAKECMDGOALS),parser)
include $(shell go list -f '{{.Dir}}' github.com/knieriem/peg)/Make.inc
%.leg.go: %.leg $(LEG)
	$(LEG) -switch $<

endif


#
# get access to original C source files
#
orig-c-src:
	hg clone git://github.com/jgm/peg-markdown.git $@


include misc/devel.mk

.PHONY: \
	all\
	cmd\
	nuke\
	package\
	parser\
	test\
