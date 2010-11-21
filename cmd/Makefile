include $(GOROOT)/src/Make.inc 

TARG=markdown
GOFILES=\
	main.go\

R = ..
PREREQ += $(R)/_obj/markdown.a

include $(GOROOT)/src/Make.cmd 


test: $(TARG)
	cd $(R)/orig-c-src/MarkdownTest_1.0.3; \
	./MarkdownTest.pl --script=../../cmd/$< --tidy
