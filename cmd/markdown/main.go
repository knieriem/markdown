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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [FILE]\n", os.Args[0])
		flag.PrintDefaults()
	}
	optNotes := flag.Bool("notes", false, "turn on footnote syntax")
	optSmart := flag.Bool("smart", false, "turn on smart quotes, dashes, and ellipses")
	optDlists := flag.Bool("dlists", false, "support definitions lists")
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

	e := markdown.Extensions{
		Notes:  *optNotes,
		Smart:  *optSmart,
		Dlists: *optDlists,
	}

	startPProf()
	defer stopPProf()

	doc := markdown.Parse(r, e)
	w := bufio.NewWriter(os.Stdout)
	doc.WriteHtml(w)
	w.Flush()
}
