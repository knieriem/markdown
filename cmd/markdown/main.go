package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/knieriem/markdown"
	"log"
	"os"
)

var format = flag.String("t", "html", "output format")

func main() {
	var opt markdown.Extensions
	flag.BoolVar(&opt.Notes, "notes", false, "turn on footnote syntax")
	flag.BoolVar(&opt.Smart, "smart", false, "turn on smart quotes, dashes, and ellipses")
	flag.BoolVar(&opt.Strike, "strike", false, "turn on strike-through syntax")
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

	p := markdown.NewParser(&opt)

	startPProf()
	defer stopPProf()

	w := bufio.NewWriter(os.Stdout)

	switch *format {
	case "groff-mm":
		p.Markdown(r, markdown.ToGroffMM(w))
	default:
		p.Markdown(r, markdown.ToHTML(w))
	}
	w.Flush()
}
