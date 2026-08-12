package vt

import "testing"

func TestTextCannotInjectTerminalControls(t *testing.T) {
	got := string(Encode([]TerminalOp{WriteText("a\x1B[31m")}, ModernCapabilities()))
	want := "a\uFFFD[31m"
	if got != want {
		t.Fatalf("Encode() = %q, want %q", got, want)
	}
}

func TestMinimumRelativeDeltaDoesNotOverflow(t *testing.T) {
	got := string(Encode([]TerminalOp{MoveRelative(-1<<31, -1<<31)}, BaselineCapabilities()))
	want := "\x1B[2147483648A\x1B[2147483648D"
	if got != want {
		t.Fatalf("Encode() = %q, want %q", got, want)
	}
}

func TestAppendEncodedPreservesDestinationPrefix(t *testing.T) {
	got := AppendEncoded([]byte("prefix:"), []TerminalOp{MoveTo(2, 3), WriteText("ok")}, BaselineCapabilities())
	want := "prefix:\x1B[4;3Hok"
	if string(got) != want {
		t.Fatalf("AppendEncoded() = %q, want %q", got, want)
	}
}

func TestSetClipboardNormalizesInvalidUTF8AndEncodesControls(t *testing.T) {
	got := string(Encode([]TerminalOp{SetClipboard("a\xFF\x1B日")}, BaselineCapabilities()))
	want := "\x1B]52;c;Ye+/vRvml6U=\x1B\\"
	if got != want {
		t.Fatalf("Encode(SetClipboard) = %q, want %q", got, want)
	}
}
