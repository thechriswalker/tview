package tview

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

// Text alignment within a box.
const (
	AlignLeft = iota
	AlignCenter
	AlignRight
)

// Semigraphical runes.
const (
	GraphicsHoriBar             = '\u2500'
	GraphicsVertBar             = '\u2502'
	GraphicsTopLeftCorner       = '\u250c'
	GraphicsTopRightCorner      = '\u2510'
	GraphicsBottomRightCorner   = '\u2518'
	GraphicsBottomLeftCorner    = '\u2514'
	GraphicsDbVertBar           = '\u2550'
	GraphicsDbHorBar            = '\u2551'
	GraphicsDbTopLeftCorner     = '\u2554'
	GraphicsDbTopRightCorner    = '\u2557'
	GraphicsDbBottomRightCorner = '\u255d'
	GraphicsDbBottomLeftCorner  = '\u255a'
	GraphicsRightT              = '\u2524'
	GraphicsLeftT               = '\u251c'
	GraphicsTopT                = '\u252c'
	GraphicsBottomT             = '\u2534'
	GraphicsCross               = '\u253c'
	GraphicsEllipsis            = '\u2026'
)

var (
	// InputFieldInteger accepts integers.
	InputFieldInteger func(text string, ch rune) bool

	// InputFieldFloat accepts floating-point numbers.
	InputFieldFloat func(text string, ch rune) bool

	// InputFieldMaxLength returns an input field accept handler which accepts
	// input strings up to a given length. Use it like this:
	//
	//   inputField.SetAcceptanceFunc(InputFieldMaxLength(10)) // Accept up to 10 characters.
	InputFieldMaxLength func(maxLength int) func(text string, ch rune) bool
)

// Package initialization.
func init() {
	// Initialize the predefined input field handlers.
	InputFieldInteger = func(text string, ch rune) bool {
		if text == "-" {
			return true
		}
		_, err := strconv.Atoi(text)
		return err == nil
	}
	InputFieldFloat = func(text string, ch rune) bool {
		if text == "-" || text == "." {
			return true
		}
		_, err := strconv.ParseFloat(text, 64)
		return err == nil
	}
	InputFieldMaxLength = func(maxLength int) func(text string, ch rune) bool {
		return func(text string, ch rune) bool {
			return len([]rune(text)) <= maxLength
		}
	}

	// Regular expressions.
	var colors string
	for color := range textColors {
		if len(colors) > 0 {
			colors += "|"
		}
		colors += color
	}
	colorPattern = regexp.MustCompile(`\[(` + colors + `)\]`)
}

// Print prints text onto the screen into the given box at (x,y,maxWidth,1),
// not exceeding that box. "align" is one of AlignLeft, AlignCenter, or
// AlignRight. The screen's background color will not be changed.
//
// Returns the number of actual runes printed.
func Print(screen tcell.Screen, text string, x, y, maxWidth, align int, color tcell.Color) int {
	// We deal with runes, not with bytes.
	runes := []rune(text)
	if maxWidth < 0 {
		return 0
	}

	// AlignCenter is split into two parts.
	if align == AlignCenter {
		half := len(runes) / 2
		halfWidth := maxWidth / 2
		return Print(screen, string(runes[:half]), x, y, halfWidth, AlignRight, color) +
			Print(screen, string(runes[half:]), x+halfWidth, y, maxWidth-halfWidth, AlignLeft, color)
	}

	// Draw text.
	drawn := 0
	for pos, ch := range runes {
		if pos >= maxWidth {
			break
		}
		finalX := x + pos
		if align == AlignRight {
			ch = runes[len(runes)-1-pos]
			finalX = x + maxWidth - 1 - pos
		}
		_, _, style, _ := screen.GetContent(finalX, y)
		style = style.Foreground(color)
		screen.SetContent(finalX, y, ch, nil, style)
		drawn++
	}

	return drawn
}

// PrintSimple prints white text to the screen at the given position.
func PrintSimple(screen tcell.Screen, text string, x, y int) {
	Print(screen, text, x, y, math.MaxInt64, AlignLeft, Styles.PrimaryTextColor)
}

// WordWrap splits a text such that each resulting line does not exceed the
// given width. Possible split points are after commas, dots, dashes, and any
// whitespace. Whitespace at split points will be dropped.
//
// Text is always split at newline characters ('\n').
func WordWrap(text string, width int) (lines []string) {
	x := 0
	start := 0
	candidate := -1 // -1 = no candidate yet.
	startAfterCandidate := 0
	countAfterCandidate := 0
	var evaluatingCandidate bool
	text = strings.TrimSpace(text)

	for pos, ch := range text {
		if !evaluatingCandidate && x >= width {
			// We've exceeded the width, we must split.
			if candidate >= 0 {
				lines = append(lines, text[start:candidate])
				start = startAfterCandidate
				x = countAfterCandidate
			} else {
				lines = append(lines, text[start:pos])
				start = pos
				x = 0
			}
			candidate = -1
			evaluatingCandidate = false
		}

		switch ch {
		// We have a candidate.
		case ',', '.', '-':
			if x > 0 {
				candidate = pos + 1
				evaluatingCandidate = true
			}
		// If we've had a candidate, skip whitespace. If not, we have a candidate.
		case ' ', '\t':
			if x > 0 && !evaluatingCandidate {
				candidate = pos
				evaluatingCandidate = true
			}
		// Split in any case.
		case '\n':
			lines = append(lines, text[start:pos])
			start = pos + 1
			evaluatingCandidate = false
			countAfterCandidate = 0
			x = 0
			continue
		// If we've had a candidate, we have a new start.
		default:
			if evaluatingCandidate {
				startAfterCandidate = pos
				evaluatingCandidate = false
				countAfterCandidate = 0
			}
		}
		x++
		countAfterCandidate++
	}

	// Process remaining text.
	text = strings.TrimSpace(text[start:])
	if len(text) > 0 {
		lines = append(lines, text)
	}

	return
}
