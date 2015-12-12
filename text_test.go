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
