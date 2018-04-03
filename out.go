package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// PrintWarning outputs a warning
func PrintWarning(txt string) {
	txt = fmt.Sprintf("WARNING: %s ", strings.ToUpper(txt))
	if config.UseColor() {
		txt = color.New(color.FgHiYellow, color.Bold).Sprintf(txt)
	}

	fmt.Println(txt)
}

// PrintStoplight will print out red if stopped is true, green if not
func PrintStoplight(txt string, stopped bool) {
	if config.UseColor() {
		c := color.FgHiGreen
		if stopped {
			c = color.FgHiRed
		}

		txt = color.New(c).Sprintf(txt)
	}

	fmt.Println(txt)
}

// PrintHeader prints out the header
func PrintHeader(txt string) {
	if config.UseColor() {
		txt = color.New(color.FgHiCyan, color.Bold).Sprintf(fmt.Sprintf("\n%s\n", txt))
	}
	fmt.Println(txt)
}

// PrintBlock prints out a block of code with the same background
func PrintBlock(block string) {
	txt := ""
	if config.UseColor() {
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
				PrintHeader(v)
				// fmt.Println(color.New(color.FgHiCyan, color.Bold).Sprintf(v))
				continue
			}

			formatted = append(formatted, v)
		}
		txt = color.New(color.FgBlack, color.BgHiCyan).Sprintf(strings.Join(formatted, "\n"))
	} else {
		txt = block
	}

	fmt.Println(txt)
}

// ExitErr prints out an error and quits
func ExitErr(txt string, err error) {
	if err != nil {
		txt = fmt.Sprintf(txt+": ", err)
	}

	if config.UseColor() {
		c := color.New(color.FgHiWhite, color.BgHiRed)
		txt = c.Sprintf(txt)
	}

	fmt.Fprintln(os.Stderr, txt)
	os.Exit(1)
}
