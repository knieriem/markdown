package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/knieriem/markdown"
	"io/ioutil"
	"os"
)

func main() {
	var b []byte

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [FILE]\n", os.Args[0])
		flag.PrintDefaults()
	}
	optNotes := flag.Bool("notes", false, "turn on footnote syntax")
	optSmart := flag.Bool("smart", false, "turn on smart quotes, dashes, and ellipses")
	optDlists := flag.Bool("dlists", false, "support definitions lists")
	flag.Parse()

	if flag.NArg() > 0 {
		b, _ = ioutil.ReadFile(flag.Arg(0))
	} else {
		b, _ = ioutil.ReadAll(os.Stdin)
	}

	e := markdown.Extensions{
		Notes:  *optNotes,
		Smart:  *optSmart,
		Dlists: *optDlists,
	}

	startPProf()
	defer stopPProf()

	doc := markdown.Parse(string(b), e)
	w := bufio.NewWriter(os.Stdout)
	doc.WriteHtml(w)
	w.Flush()
}
