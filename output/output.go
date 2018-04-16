package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Out prints the data to os.Stdout/os.StdErr
var Out *BuffOut

// BuffOut provides writers to handle output and err output
type BuffOut struct {
	Out io.Writer
	Err io.Writer
}

func init() {
	Out = &BuffOut{Out: os.Stdout, Err: os.Stderr}
}

// PrintWarning outputs a warning
func (bo *BuffOut) PrintWarning(txt string) {
	fmt.Fprintln(bo.Out, color.New(color.FgYellow, color.Bold).Sprintf("WARNING: %s ", strings.ToUpper(txt)))
}

// PrintStoplight will print out red if stopped is true, green if not
func (bo *BuffOut) PrintStoplight(txt string, stopped bool) {
	c := color.FgGreen
	if stopped {
		c = color.FgRed
	}

	fmt.Fprintln(bo.Out, color.New(c).Sprintf(txt))
}

// PrintHeader prints out the header
func (bo *BuffOut) PrintHeader(txt string) {
	fmt.Fprintln(bo.Out, color.New(color.FgCyan, color.Bold).Sprintf("\n%s\n", txt))
}

// PrintBlock prints out a block of code with the same background
func (bo *BuffOut) PrintBlock(block string) {
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
			bo.PrintHeader(v)
			continue
		}

		formatted = append(formatted, color.New(color.FgBlack, color.BgCyan).Sprintf(v))
	}

	fmt.Fprintln(bo.Out, strings.Join(formatted, "\n"))
}

// PrintErr prints out the text to Stderr
func (bo *BuffOut) PrintErr(txt string, err error) {
	if err != nil {
		if txt != "" {
			txt = fmt.Sprintf(txt+": %s", err)
		} else {
			txt = fmt.Sprintf("%s", err)
		}
	}

	fmt.Fprintln(bo.Err, color.New(color.FgWhite, color.BgRed).Sprintf(txt))
}

// PrintVersion will output the current version and colophon
func (bo *BuffOut) PrintVersion(version string) {
	bo.PrintBlock(fmt.Sprintf(`thoom.Gulp
version: %s
author: Z.d.Peacock <zdp@thoomtech.com>
link: https://github.com/thoom/gulp`, version))

	fmt.Fprintln(bo.Out, "")
}

// NoColor disables outputing in color
func NoColor(noColor bool) {
	color.NoColor = noColor
}

// ExitErr prints out an error and quits
func ExitErr(txt string, err error) {
	Out.PrintErr(txt, err)
	os.Exit(1)
}
