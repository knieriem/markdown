package markdown

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// for each pair of .text/.html files in the given subdirectory
// of `./tests' compare the expected html output with
// the output of Parser.Markdown.
func runDirTests(dir string, opt *Extensions, t *testing.T) {

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

	var buf bytes.Buffer
	fHTML := ToHTML(&buf)
	fGroff := ToGroffMM(&buf)
	p := NewParser(opt)
	for _, name := range names {
		if filepath.Ext(name) != ".text" {
			continue
		}
		if err = compareOutput(&buf, fHTML, ".html", filepath.Join(dirPath, name), p); err != nil {
			t.Error(err)
		}
		if err = compareOutput(&buf, fGroff, ".mm", filepath.Join(dirPath, name), p); err != nil {
			t.Error(err)
		}
	}
}

// Compare the output of the C-based peg-markdown, which
// is, for each test, available in either a .html or a .mm file accompanying
// the .text file, with the output of this package's Markdown processor.
func compareOutput(w *bytes.Buffer, f Formatter, ext string, textPath string, p *Parser) (err error) {
	var bOrig bytes.Buffer

	r, err := os.Open(textPath)
	if err != nil {
		return
	}
	defer r.Close()

	w.Reset()
	p.Markdown(r, f)

	// replace .text extension by `ext'
	base := textPath[:len(textPath)-len(".text")]
	refPath := base + ext

	r, err = os.Open(refPath)
	if err != nil {
		return
	}
	defer r.Close()
	bOrig.ReadFrom(r)
	if bytes.Compare(bOrig.Bytes(), w.Bytes()) != 0 {
		err = fmt.Errorf("test %q failed", refPath)
	}
	return
}

func TestMarkdown103(t *testing.T) {
	runDirTests("md1.0.3", nil, t)
}

func TestMarkdownIssues(t *testing.T) {
	runDirTests("issues", &Extensions{Notes: true}, t)
}

// This test will make the test run fail with a
// message like "Buffer not empty" under the
// following condition:
//
// There exists an unprocessed, remaining portion of the
// input buffer after the previous parser call, which
// consists only of whitespace.
// This whitespace should have been ignored, but, due to
// a bug, hasn't.
func TestTrailingWhitespaceBug(t *testing.T) {
	const input = `* foo

    # bar

* baz
`
	var buf bytes.Buffer
	p := NewParser(nil)
	p.Markdown(strings.NewReader(input), ToHTML(&buf))
}
