// Miminal package for terminal colors/ANSI escape code.
// Check out the source here https://github.com/acmacalister/skittles.
// Also see the example directory for another example on how to use skittles.
//
//  package main
//
//  import (
//    "fmt"
//    "github.com/acmacalister/skittles"
//  )
//
//  func main() {
//    fmt.Println(skittles.Red("Red's my favorite color"))
//  }
package skittles

import (
	"fmt"
)

const (
	black = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
	regular   = ""
	bold      = "1;"
	blink     = "5;"
	underline = "4;"
	inverse   = "7;"
)

func makeString(attr string, text interface{}, color int) string {
	return fmt.Sprintf("\033[%s%dm%v\033[0m", attr, color, text)
}

// make the terminal text black - \033[30m
func Black(text interface{}) string {
	return makeString(regular, text, black)
}

// make the terminal text red - \033[30m
func Red(text interface{}) string {
	return makeString(regular, text, red)
}

// make the terminal text green - \033[30m
func Green(text interface{}) string {
	return makeString(regular, text, green)
}

// make the terminal text yellow - \033[30m
func Yellow(text interface{}) string {
	return makeString(regular, text, yellow)
}

//  make the terminal text blue - \033[30m
func Blue(text interface{}) string {
	return makeString(regular, text, blue)
}

//  make the terminal text magenta - \033[30m
func Magenta(text interface{}) string {
	return makeString(regular, text, magenta)
}

// make the terminal text cyan - \033[30m
func Cyan(text interface{}) string {
	return makeString(regular, text, cyan)
}

// make the terminal text white - \033[30m
func White(text interface{}) string {
	return makeString(regular, text, white)
}

// make the terminal text bold black - \033[1;30m
func BoldBlack(text interface{}) string {
	return makeString(bold, text, black)
}

// make the terminal text bold red - \1;033[30m
func BoldRed(text interface{}) string {
	return makeString(bold, text, red)
}

// make the terminal text bold green - \033[1;30m
func BoldGreen(text interface{}) string {
	return makeString(bold, text, green)
}

// make the terminal text bold yellow - \033[1;30m
func BoldYellow(text interface{}) string {
	return makeString(bold, text, yellow)
}

// make the terminal text bold blue - \033[1;30m
func BoldBlue(text interface{}) string {
	return makeString(bold, text, blue)
}

// make the terminal text bold magenta - \033[1;30m
func BoldMagenta(text interface{}) string {
	return makeString(bold, text, magenta)
}

// make the terminal text bold cyan - \033[1;30m
func BoldCyan(text interface{}) string {
	return makeString(bold, text, cyan)
}

// make the terminal text bold white - \033[1;30m
func BoldWhite(text interface{}) string {
	return makeString(bold, text, white)
}

// make the terminal text blink black - \033[5;30m
func BlinkBlack(text interface{}) string {
	return makeString(blink, text, black)
}

// make the terminal text blink red - \033[5;30m
func BlinkRed(text interface{}) string {
	return makeString(blink, text, red)
}

// make the terminal text blink green - \033[5;30m
func BlinkGreen(text interface{}) string {
	return makeString(blink, text, green)
}

// make the terminal text blink yellow - \033[5;30m
func BlinkYellow(text interface{}) string {
	return makeString(blink, text, yellow)
}

// make the terminal text blink blue - \033[5;30m
func BlinkBlue(text interface{}) string {
	return makeString(blink, text, blue)
}

// make the terminal text blink magenta - \033[5;30m
func BlinkMagenta(text interface{}) string {
	return makeString(blink, text, magenta)
}

// make the terminal text blink cyan - \033[5;30m
func BlinkCyan(text interface{}) string {
	return makeString(blink, text, cyan)
}

// make the terminal text blink white - \033[5;30m
func BlinkWhite(text interface{}) string {
	return makeString(blink, text, white)
}

// make the terminal text underline black - \033[4;30m
func UnderlineBlack(text interface{}) string {
	return makeString(underline, text, black)
}

// make the terminal text underline red - \033[4;30m
func UnderlineRed(text interface{}) string {
	return makeString(underline, text, red)
}

// make the terminal text underline green - \033[4;30m
func UnderlineGreen(text interface{}) string {
	return makeString(underline, text, green)
}

// make the terminal text underline yellow - \033[4;30m
func UnderlineYellow(text interface{}) string {
	return makeString(underline, text, yellow)
}

// make the terminal text underline blue - \033[4;30m
func UnderlineBlue(text interface{}) string {
	return makeString(underline, text, blue)
}

// make the terminal text underline magenta - \033[4;30m
func UnderlineMagenta(text interface{}) string {
	return makeString(underline, text, magenta)
}

// make the terminal text underline cyan - \033[4;30m
func UnderlineCyan(text interface{}) string {
	return makeString(underline, text, cyan)
}

// make the terminal text underline white - \033[4;30m
func UnderlineWhite(text interface{}) string {
	return makeString(underline, text, white)
}

// make the terminal text inverse black - \033[7;30m
func InverseBlack(text interface{}) string {
	return makeString(inverse, text, black)
}

// make the terminal text inverse red - \033[7;30m
func InverseRed(text interface{}) string {
	return makeString(inverse, text, red)
}

// make the terminal text inverse green - \033[7;30m
func InverseGreen(text interface{}) string {
	return makeString(inverse, text, green)
}

// make the terminal text inverse yellow - \033[7;30m
func InverseYellow(text interface{}) string {
	return makeString(inverse, text, yellow)
}

// make the terminal text inverse blue - \033[7;30m
func InverseBlue(text interface{}) string {
	return makeString(inverse, text, blue)
}

// make the terminal text inverse magenta - \033[7;30m
func InverseMagenta(text interface{}) string {
	return makeString(inverse, text, magenta)
}

// make the terminal text inverse cyan - \033[7;30m
func InverseCyan(text interface{}) string {
	return makeString(inverse, text, cyan)
}

// make the terminal text inverse white - \033[7;30m
func InverseWhite(text interface{}) string {
	return makeString(inverse, text, white)
}
