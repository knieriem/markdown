package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/knieriem/markdown"
	"log"
	"os"
)

func main() {
	var opt markdown.Options
	flag.BoolVar(&opt.Notes, "notes", false, "turn on footnote syntax")
	flag.BoolVar(&opt.Smart, "smart", false, "turn on smart quotes, dashes, and ellipses")
	flag.BoolVar(&opt.Dlists, "dlists", false, "support definitions lists")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [FILE]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	r := os.Stdin
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r = f
	}

	startPProf()
	defer stopPProf()

	doc := markdown.Parse(r, opt)
	w := bufio.NewWriter(os.Stdout)
	doc.WriteHtml(w)
	w.Flush()
}
