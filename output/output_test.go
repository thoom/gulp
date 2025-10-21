package output

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestNoColor(t *testing.T) {
	assert := assert.New(t)

	NoColor(false)
	assert.False(color.NoColor)

	NoColor(true)
	assert.True(color.NoColor)
}

func TestPrintWarning(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintWarning("boo")
	assert.Equal("WARNING: BOO\n", b.String())
}

func TestSpotlightStop(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintStoplight("stop", true)
	assert.Equal("stop\n", b.String())
}
func TestSpotlightGo(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintStoplight("go", false)
	assert.Equal("go\n", b.String())
}

func TestHeader(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintHeader("HEADER")
	assert.Equal("\nHEADER\n\n", b.String())
}

func TestPrintBlock(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintBlock(`HEADER
This is a line
This is another line`)

	lines := []string{"HEADER", "This is a line", "This is another line"}
	for i, line := range lines {
		t.Logf("Line %d: %q (len=%d)", i, line, len(line))
	}

	// Debug: print the actual output
	t.Logf("Actual output: %q", b.String())

	// The first line (HEADER) is printed as a header without padding
	// The remaining lines are part of the colored block with padding
	// Since "This is another line" is 20 chars, and "This is a line" is 14 chars,
	// "This is a line" should be padded with 6 spaces
	assert.Equal("\nHEADER\n\nThis is a line      \nThis is another line\n", b.String())
}

func TestPrintErr(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintErr("Label", fmt.Errorf("Error Message"))
	assert.Equal("Label: Error Message\n", b.String())
}

func TestPrintErrNil(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintErr("Label", nil)
	assert.Equal("Label\n", b.String())
}

func TestPrintErrNoLabel(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintErr("", fmt.Errorf("Error Message"))
	assert.Equal("Error Message\n", b.String())
}

func TestPrintVersion(t *testing.T) {
	assert := assert.New(t)
	tst := &BuffOut{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

	expected := "version: abc123def"
	tst.PrintVersion("abc123def")

	assert.Contains(tst.Out.(*bytes.Buffer).String(), expected)
}

func TestPrintVersionWithUpdates(t *testing.T) {
	assert := assert.New(t)

	t.Run("with update available", func(t *testing.T) {
		tst := &BuffOut{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		tst.PrintVersionWithUpdates("1.0.0", true, "1.1.0", "https://github.com/thoom/gulp/releases/tag/v1.1.0")
		output := tst.Out.(*bytes.Buffer).String()

		assert.Contains(output, "1.0.0")
		assert.Contains(output, "Update available")
		assert.Contains(output, "1.1.0")
		assert.Contains(output, "https://github.com/thoom/gulp/releases/tag/v1.1.0")
	})

	t.Run("no update available", func(t *testing.T) {
		tst := &BuffOut{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		tst.PrintVersionWithUpdates("1.0.0", false, "1.0.0", "")
		output := tst.Out.(*bytes.Buffer).String()

		assert.Contains(output, "1.0.0")
		assert.Contains(output, "latest version")
	})
}
