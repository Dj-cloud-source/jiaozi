package orderapp

import "testing"

func TestMaskIDCardMasksMiddleDigits(t *testing.T) {
	masked := maskIDCard("320101200001011234")
	if masked != "3201********1234" {
		t.Fatalf("expected masked id card, got %s", masked)
	}
}

func TestMaskIDCardMasksShortValue(t *testing.T) {
	masked := maskIDCard("12345678")
	if masked != "****" {
		t.Fatalf("expected ****, got %s", masked)
	}
}
