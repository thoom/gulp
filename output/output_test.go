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

	assert.Equal("\nHEADER               \n\nThis is a line       \nThis is another line \n", b.String())
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

	b := &bytes.Buffer{}
	tst := &BuffOut{Out: b, Err: b}
	tst.PrintVersion("abc123def")
	assert.Equal("\nthoom.Gulp                              \n\nversion: abc123def                      \nauthor: Z.d.Peacock <zdp@thoomtech.com> \nlink: https://github.com/thoom/gulp     \n\n", b.String())
}
