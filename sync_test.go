package main

import (
	"testing"
)

func TestCleanMsgPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"[]", ""},
		{" [ ] ", ""},
		{"[  ]", ""},
		{"[WhatsApp Direct: ]", ""},
		{"[Signal Direct: ]", ""},
		{"[WhatsApp Direct: +123456]", "[WhatsApp Direct: +123456]"},
		{"[Group Name]", "[Group Name]"},
		{"  [Family]  ", "[Family]"},
	}

	for _, tt := range tests {
		got := cleanMsgPrefix(tt.input)
		if got != tt.expected {
			t.Errorf("cleanMsgPrefix(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatBlockquote(t *testing.T) {
	quote := formatBlockquote("Alice", "Hello world")
	expected := "> Alice: Hello world"
	if quote != expected {
		t.Errorf("formatBlockquote() = %q; want %q", quote, expected)
	}

	multilineQuote := formatBlockquote("Bob", "Line 1\nLine 2")
	expectedMultiline := "> Bob: Line 1\n> Line 2"
	if multilineQuote != expectedMultiline {
		t.Errorf("formatBlockquote() = %q; want %q", multilineQuote, expectedMultiline)
	}
}

func TestFormattedTextMessageAssembly(t *testing.T) {
	// Case 1: Reply with empty msgPrefix (should not have [] or leading \n before >)
	msgPrefix := cleanMsgPrefix("[]")
	quoteBlock := "> Alice: original message"
	cleanText := "Bob: reply message"

	var formattedText string
	if quoteBlock != "" {
		if msgPrefix != "" {
			formattedText = msgPrefix + "\n" + quoteBlock + "\n" + cleanText
		} else {
			formattedText = quoteBlock + "\n" + cleanText
		}
	}

	expectedReply := "> Alice: original message\nBob: reply message"
	if formattedText != expectedReply {
		t.Errorf("Formatted reply text = %q; want %q", formattedText, expectedReply)
	}

	// Case 2: Reply with non-empty msgPrefix
	msgPrefixGroup := cleanMsgPrefix("[Family]")
	if quoteBlock != "" {
		if msgPrefixGroup != "" {
			formattedText = msgPrefixGroup + "\n" + quoteBlock + "\n" + cleanText
		} else {
			formattedText = quoteBlock + "\n" + cleanText
		}
	}

	expectedGroupReply := "[Family]\n> Alice: original message\nBob: reply message"
	if formattedText != expectedGroupReply {
		t.Errorf("Formatted group reply text = %q; want %q", formattedText, expectedGroupReply)
	}

	// Case 3: Direct/General message with empty msgPrefix (should not have [] header)
	msgPrefixDirect := cleanMsgPrefix("[]")
	var formattedDirect string
	if msgPrefixDirect != "" {
		formattedDirect = msgPrefixDirect + " " + cleanText
	} else {
		formattedDirect = cleanText
	}

	expectedDirect := "Bob: reply message"
	if formattedDirect != expectedDirect {
		t.Errorf("Formatted direct text = %q; want %q", formattedDirect, expectedDirect)
	}
}
