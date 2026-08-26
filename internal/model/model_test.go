package model

import "testing"

func TestFigureAndEncodingStateRules(t *testing.T) {
	if !IsValidFigureStatus(FigureStatusFrozen) || IsValidFigureStatus("unknown") {
		t.Fatal("figure status validation is incorrect")
	}
	if (Figure{Status: FigureStatusFrozen}).CanEdit() {
		t.Fatal("frozen figure is editable")
	}
	if !(Figure{Status: FigureStatusPublishable}).CanEdit() {
		t.Fatal("publishable figure should be editable")
	}
	if !IsValidEncodingStatus(EncodingStatusMissingLgnd) {
		t.Fatal("missing legend encoding status is not recognized")
	}
}

func TestExceptionAndChannelValidation(t *testing.T) {
	if !IsValidExceptionKind(ExceptionReuse) || IsValidExceptionKind("other") {
		t.Fatal("exception kind validation is incorrect")
	}
	for _, channel := range []string{ChannelColor, ChannelShape, ChannelSize, ChannelLinetype} {
		if !IsValidChannel(channel) {
			t.Fatalf("expected valid channel %q", channel)
		}
	}
	if IsValidChannel("opacity") {
		t.Fatal("unknown channel accepted")
	}
}
