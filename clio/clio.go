package clio

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/zeroSal/went-clio/ansi"
	"golang.org/x/term"
)

type Clio struct {
	out    io.Writer
	in     io.Reader
	reader *bufio.Reader
	mutex  sync.Mutex
	bannerTemplate string
}

func NewClio() *Clio {
	return &Clio{
		out:    os.Stdout,
		in:     os.Stdin,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *Clio) output(color, prefix, msg string, newline bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	format := "%s%s%s%s"
	if newline {
		format += "\n"
	}

	_, err := fmt.Fprintf(c.out, format, color, prefix, msg, ansi.Reset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to console: %v\n", err)
	}
}

func (c *Clio) SetBannerTemplate(template string) {
	c.bannerTemplate = template
}

func (c *Clio) Banner(values ...any) {
	if c.bannerTemplate == "" {
		return
	}

	c.output(ansi.White, "", fmt.Sprintf(c.bannerTemplate, values...), true)
}

func (c *Clio) Ask(question string, validate ValidatorFn) string {
	prompt := "[?] " + question + ": "
	for {
		c.output(ansi.White, "", prompt, false)
		line, err := c.reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if err != nil && err != io.EOF {
			c.Error(fmt.Sprintf("Read error: %v", err))
			continue
		}

		if validate == nil {
			return line
		}

		if err := validate(line); err != nil {
			c.Error(fmt.Sprintf("Invalid input: %s", err.Error()))
			continue
		}

		return line
	}
}

func (c *Clio) AskHidden(question string, validate ValidatorFn) (string, error) {
	prompt := "[?] " + question + ": "
	for {
		c.output(ansi.White, "", prompt, false)

		var value string
		f, ok := c.in.(*os.File)
		if !ok {
			line, err := c.reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return "", fmt.Errorf("error reading hidden input: %w", err)
			}
			value = strings.TrimRight(line, "\r\n")
		} else {
			raw, err := term.ReadPassword(int(f.Fd()))
			fmt.Fprintln(c.out)
			if err != nil {
				return "", fmt.Errorf("error reading hidden input: %w", err)
			}
			value = string(raw)
		}

		if validate == nil {
			return value, nil
		}

		if err := validate(value); err != nil {
			c.Error(fmt.Sprintf("Invalid input: %s", err.Error()))
			continue
		}

		return value, nil
	}
}

func (c *Clio) Confirm(question string, defaultVal bool) bool {
	hint := "y/N"
	if defaultVal {
		hint = "Y/n"
	}

	prompt := fmt.Sprintf("[?] %s [%s]: ", question, hint)
	for {
		c.output(ansi.White, "", prompt, false)
		line, err := c.reader.ReadString('\n')
		line = strings.ToLower(strings.TrimRight(line, "\r\n"))

		if err != nil && err != io.EOF {
			c.Error(fmt.Sprintf("Reading error: %s", err.Error()))
			continue
		}

		if line == "" {
			return defaultVal
		}

		if line == "y" || line == "n" {
			return line == "y"
		}

		c.Error("Invalid answer. Use y/n.")
	}
}

func (c *Clio) PickInRange(min, max int) int {
	prompt := fmt.Sprintf("[>] Pick a value (from %d to %d): ", min, max)
	for {
		c.output(ansi.White, "", prompt, false)
		line, err := c.reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if err != nil && err != io.EOF {
			c.Error(fmt.Sprintf("Reading error: %v", err))
			continue
		}

		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || n < min || n > max {
			c.Error(fmt.Sprintf("Invalid input: %q is not a valid value (%d-%d)", line, min, max))
			continue
		}

		return n
	}
}

func (c *Clio) MultipleChoice(question string, choices []Choice) string {
	if len(choices) == 0 {
		return ""
	}

	labels := make([]string, len(choices))
	for i, ch := range choices {
		labels[i] = fmt.Sprintf("%d) %s", i+1, ch.Label)
	}
	c.List(question, labels)

	n := c.PickInRange(1, len(choices))
	ch := choices[n-1]
	if ch.Value != "" {
		return ch.Value
	}
	return ch.Label
}

func (c *Clio) Debug(msg string) {
	c.output(ansi.White, "[·] ", msg, true)
}

func (c *Clio) Info(msg string) {
	c.output(ansi.Blue, "[i] ", msg, true)
}

func (c *Clio) Success(msg string) {
	c.output(ansi.Green, "[✓] ", msg, true)
}

func (c *Clio) Warn(msg string) {
	c.output(ansi.Yellow, "[!] ", msg, true)
}

func (c *Clio) Error(msg string) {
	c.output(ansi.Red, "[×] ", msg, true)
}

func (c *Clio) Fatal(msg string) {
	c.output(ansi.Red, "[FATAL] ", msg, true)
}

func (c *Clio) List(title string, elements []string) {
	c.output(ansi.White, "[*] ", title, true)
	for _, el := range elements {
		c.output(ansi.White, "", fmt.Sprintf("    • %s", el), true)
	}
}