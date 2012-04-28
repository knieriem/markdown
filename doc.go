/*
A translation of peg-markdown [1] into Go.

Usage example:

	package main

	import (
		"github.com/knieriem/markdown"
		"os"
		"bufio"
	)

	func main() {
		p := markdown.NewParser(&markdown.Options{Smart: true})

		w := bufio.NewWriter(os.Stdout)
		p.Markdown(os.Stdin, markdown.ToHTML(w))
		w.Flush()
	}

[1]: https://github.com/jgm/peg-markdown/
*/
package markdown
