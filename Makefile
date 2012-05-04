all:
	@echo 'targets: nuke parser clean'

cmd: package
	cd cmd/markdown && go build -v

package: parser.leg.go
	go install -v

clean:
	go clean . ./...
	rm -rf ,,prevmd ,,pmd
	
parser:	parser.leg.go

nuke:
	rm -f parser.leg.go


# LEG parser rules
#
ifeq ($(MAKECMDGOALS),parser)
include $(shell go list -f '{{.Dir}}' github.com/knieriem/peg)/Make.inc
%.leg.go: %.leg $(LEG)
	$(LEG) -verbose -switch -O all $< > $@

endif


include misc/devel.mk

.PHONY: \
	all\
	cmd\
	nuke\
	package\
	parser\
