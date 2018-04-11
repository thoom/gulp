package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	// UseColor will either output in color or not
	UseColor = true
)

// PrintWarning outputs a warning
func PrintWarning(txt string, writer io.Writer) {
	txt = fmt.Sprintf("WARNING: %s ", strings.ToUpper(txt))
	if UseColor {
		txt = color.New(color.FgHiYellow, color.Bold).Sprintf(txt)
	}

	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintln(writer, txt)
}

// PrintStoplight will print out red if stopped is true, green if not
func PrintStoplight(txt string, stopped bool, writer io.Writer) {
	if UseColor {
		c := color.FgHiGreen
		if stopped {
			c = color.FgHiRed
		}

		txt = color.New(c).Sprintf(txt)
	}

	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintln(writer, txt)
}

// PrintHeader prints out the header
func PrintHeader(txt string, writer io.Writer) {
	if UseColor {
		txt = color.New(color.FgHiCyan, color.Bold).Sprintf(fmt.Sprintf("\n%s\n", txt))
	}

	if writer == nil {
		writer = os.Stdout
	}
	fmt.Fprintln(writer, txt)
}

// PrintBlock prints out a block of code with the same background
func PrintBlock(block string, writer io.Writer) {
	txt := ""
	if UseColor {
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
				// fmt.Println(color.New(color.FgHiCyan, color.Bold).Sprintf(v))
				continue
			}

			formatted = append(formatted, v)
		}
		txt = color.New(color.FgBlack, color.BgHiCyan).Sprintf(strings.Join(formatted, "\n"))
	} else {
		txt = block
	}

	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintln(writer, txt)
}

// ExitErr prints out an error and quits
func ExitErr(txt string, err error) {
	if err != nil {
		txt = fmt.Sprintf(txt+": ", err)
	}

	if UseColor {
		c := color.New(color.FgHiWhite, color.BgHiRed)
		txt = c.Sprintf(txt)
	}

	fmt.Fprintln(os.Stderr, txt)
	os.Exit(1)
}
