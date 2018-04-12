package output

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintWarning(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	PrintWarning("boo", b)
	assert.Equal("WARNING: BOO \n", b.String())
}

func TestHeader(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	PrintHeader("HEADER", b)
	assert.Equal("\nHEADER\n\n", b.String())
}

func TestPrintBlock(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	PrintBlock(`HEADER
This is a line
This is another line`, b)

	assert.Equal("\nHEADER               \n\nThis is a line       \nThis is another line \n", b.String())
}

func TestPrintErr(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	PrintErr("Label", fmt.Errorf("Error Message"), b)
	assert.Equal("Label: Error Message\n", b.String())
}

func TestPrintErrNil(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	PrintErr("Label", nil, b)
	assert.Equal("Label\n", b.String())
}

func TestPrintErrNoLabel(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	PrintErr("", fmt.Errorf("Error Message"), b)
	assert.Equal("Error Message\n", b.String())
}
