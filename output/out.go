package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// NoColor disables outputing in color
func NoColor(noColor bool) {
	color.NoColor = noColor
}

// PrintWarning outputs a warning
func PrintWarning(txt string, writer io.Writer) {
	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintln(writer, color.New(color.FgYellow, color.Bold).Sprintf("WARNING: %s ", strings.ToUpper(txt)))
}

// PrintStoplight will print out red if stopped is true, green if not
func PrintStoplight(txt string, stopped bool, writer io.Writer) {
	c := color.FgGreen
	if stopped {
		c = color.FgRed
	}

	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintln(writer, color.New(c).Sprintf(txt))
}

// PrintHeader prints out the header
func PrintHeader(txt string, writer io.Writer) {
	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintln(writer, color.New(color.FgCyan, color.Bold).Sprintf("\n%s\n", txt))
}

// PrintBlock prints out a block of code with the same background
func PrintBlock(block string, writer io.Writer) {
	pieces := strings.Split(block, "\n")
	max := 0
	for _, v := range pieces {
		// do something
		l := len(v)
		if l > max {
			max = l
		}
	}

	var formatted []string
	for i, v := range pieces {
		l := len(v)
		padding := ""
		if l < max {
			padding = strings.Repeat(" ", max-l)
		}

		v = fmt.Sprintf("%s%s ", v, padding)
		if i == 0 {
			PrintHeader(v, writer)
			continue
		}

		formatted = append(formatted, color.New(color.FgBlack, color.BgCyan).Sprintf(v))
	}

	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintln(writer, strings.Join(formatted, "\n"))
}

// PrintErr prints out the text to Stderr
func PrintErr(txt string, err error, writer io.Writer) {
	if err != nil {
		if txt != "" {
			txt = fmt.Sprintf(txt+": %s", err)
		} else {
			txt = fmt.Sprintf("%s", err)
		}
	}

	if writer == nil {
		writer = os.Stderr
	}

	fmt.Fprintln(writer, color.New(color.FgWhite, color.BgRed).Sprintf(txt))
}

// ExitErr prints out an error and quits
func ExitErr(txt string, err error) {
	PrintErr(txt, err, nil)
	os.Exit(1)
}
