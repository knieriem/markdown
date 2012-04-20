/*
A translation of peg-markdown [1] into Go.

Usage example:

	package main

	import (
		md "github.com/knieriem/markdown"
		"os"
		"bufio"
	)

	func main() {
		doc := md.Parse(os.Stdin, md.Extensions{Smart: true})

		w := bufio.NewWriter(os.Stdout)
		doc.WriteHtml(w)	
		w.Flush()
	}

[1]: https://github.com/jgm/peg-markdown/
*/
package markdown
