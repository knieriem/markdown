package markdown

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// for each pair of .text/.html files in the given subdirectory
// of `./tests' compare the expected html output with
// the output of Parser.Markdown.
func runDirTests(dir string, t *testing.T) {

	dirPath := filepath.Join("tests", dir)
	f, err := os.Open(dirPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	names, err := f.Readdirnames(-1)
	if err != nil {
		t.Fatal(err)
	}

	p := NewParser(nil)
	for _, name := range names {
		if filepath.Ext(name) != ".text" {
			continue
		}
		if err = compareOutput(filepath.Join(dirPath, name), p); err != nil {
			t.Error(err)
		}
	}
}

// Compare the output of the C-based peg-markdown, which
// is, for each test, available in a .html file accompanying the
// .text file, with the output of this package's Markdown processor.
func compareOutput(textPath string, p *Parser) (err error) {
	var bOrig, bThis bytes.Buffer

	r, err := os.Open(textPath)
	if err != nil {
		return
	}
	defer r.Close()

	bThis.Reset()
	out := ToHTML(&bThis)
	p.Markdown(r, out)

	// replace .text extension by .html
	base := textPath[:len(textPath)-len(".text")]
	htmlPath := base + ".html"

	r, err = os.Open(htmlPath)
	if err != nil {
		return
	}
	defer r.Close()
	bOrig.ReadFrom(r)
	if bytes.Compare(bOrig.Bytes(), bThis.Bytes()) != 0 {
		err = fmt.Errorf("test %q failed", base)
	}
	return
}

func TestMarkdown103(t *testing.T) {
	runDirTests("md1.0.3", t)
}
