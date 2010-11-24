/*
A translation of peg-markdown [1] into Go.

Usage example:

	import (
		md "markdown"
		"os"
		"io/ioutil"
		"bufio"
	)

	func main() {
		b, _ := ioutil.ReadAll(os.Stdin)

		doc := md.Parse(string(b), md.Extensions{Smart: true})

		w := bufio.NewWriter(os.Stdout)
		doc.WriteHtml(w)	
		w.Flush()
	}

[1]: https://github.com/jgm/peg-markdown/
*/
package markdown
