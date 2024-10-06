package logging

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	Trace   *log.Logger
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

const (
	fg_black = iota + 30
	fg_red
	fg_green
	fg_yellow
	fg_blue
	fg_magenta
	fg_cyan
	fg_white
)

const (
	bg_black = iota + 40
	bg_red
	bg_green
	bg_yellow
	bg_blue
	bg_magenta
	bg_cyan
	bg_white
)

const (
	fg_bold = iota + 1
	fg_dim
	fg_italic
	fg_underline
	fg_blinking
	fg_inverse
	fg_hidden
	fg_striketrough
)

const reset_styles = 0

func InitializeLoggers(out io.Writer) {
	useColor := out == os.Stdout || out == os.Stderr

	prefixes := map[string][]any{
		"TRACE": {fg_magenta, ""},
		"DEBUG": {fg_cyan, ""},
		"INFO":  {fg_green, ""},
		"WARN":  {fg_yellow, ""},
		"ERROR": {fg_red, ""},
	}

	for level, values := range prefixes {
		if useColor {
			prefixes[level][1] = withColor(level, values[0].(int)) + " "
		} else {
			prefixes[level][1] = level + " "
		}
	}

	baseFlags := log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC

	Trace = log.New(out, prefixes["TRACE"][1].(string), baseFlags|log.Llongfile)
	Debug = log.New(out, prefixes["DEBUG"][1].(string), baseFlags|log.Lshortfile)
	Info = log.New(out, prefixes["INFO"][1].(string), baseFlags)
	Warning = log.New(out, prefixes["WARN"][1].(string), baseFlags)
	Error = log.New(out, prefixes["ERROR"][1].(string), baseFlags)
}

func withColor(s string, color int) string {
	if color >= fg_black && color <= fg_white {
		s = fmt.Sprintf("\033[%dm%s\033[%dm", color, s, reset_styles)
	}

	return s
}
