package gltext

import (
	"testing"
)

func TestHasRune(t *testing.T) {
	f := &Font{}
	f.Config = &FontConfig{}
	f.Config.Glyphs = make(Charset, 100)
	f.Config.RuneRanges = make(RuneRanges, 0)

	r := RuneRange{Low: 30, High: 40}
	f.Config.RuneRanges = append(f.Config.RuneRanges, r)
	r = RuneRange{Low: 100, High: 400}
	f.Config.RuneRanges = append(f.Config.RuneRanges, r)

	if !f.Config.RuneRanges.Validate() {
		t.Error("Not validating properly.")
	}
	text := &Text{}
	text.font = f
	if !text.HasRune(40) {
		t.Error("Missing rune 40.")
	}
	if text.HasRune(41) {
		t.Error("Should not have 41.")
	}
}

// TestClickedCharacter tests a hypothetical string of length 3 with variable width chars
func TestClickedCharacter(t *testing.T) {
	text := &Text{}
	text.font = &Font{}
	text.font.WindowWidth = 100
	text.X1.X = -20
	text.String = "ABC"

	// click was just around the middle of the screen
	xPos := float64(51)

	// -20 to -10 is A
	// -10 to +10 is B
	// +10 to +20 is C
	text.CharSpacing = make([]float32, 0)
	text.CharSpacing = append(text.CharSpacing, 10)
	text.CharSpacing = append(text.CharSpacing, 20)
	text.CharSpacing = append(text.CharSpacing, 10)

	index, side := text.ClickedCharacter(xPos)
	if index != 1 {
		t.Error("Expecting index 1")
	}
	if side != CSRight {
		t.Error("Expecting right side click")
	}
}
