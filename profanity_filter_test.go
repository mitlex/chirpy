package main

import (
	"testing"
)

func TestProfanityFilter(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "  hello  world  ",
			expected: "  hello  world  ",
		},
		{
			input:    "HELLO WORLD",
			expected: "HELLO WORLD",
		},
		{
			input:    "what a kerfuffle",
			expected: "what a ****",
		},
		{
			input:    "sharbert on it!",
			expected: "**** on it!",
		},
		{
			input:    "fornax this!",
			expected: "**** this!",
		},
		{
			input:    "FORNAX this!",
			expected: "**** this!",
		},
		{
			input:    "foRNaX this!",
			expected: "**** this!",
		},
		{
			input:    "  hello  world  ",
			expected: "  hello  world  ",
		},
		{
			input:    "HELLO WORLD",
			expected: "HELLO WORLD",
		},
		{
			input:    "what a kerfuffle",
			expected: "what a ****",
		},
		{
			input:    "sharbert on it!",
			expected: "**** on it!",
		},
		{
			input:    "fornax this!",
			expected: "**** this!",
		},
		{
			input:    "FORNAX this!",
			expected: "**** this!",
		},
		{
			input:    "what a kerfuffle!",
			expected: "what a kerfuffle!",
		},
		{
			input:    "kerfuffle sharbert fornax",
			expected: "**** **** ****",
		},
		{
			input:    "Kerfuffle!",
			expected: "Kerfuffle!",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "kerfufflewhatever",
			expected: "kerfufflewhatever",
		},
		{
			input:    "KERFUFFLEWHATEVER",
			expected: "KERFUFFLEWHATEVER",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			actual := profanityFilter(c.input)

			if actual != c.expected {
				t.Errorf("actual chirp: %v, expected chirp: %v", actual, c.expected)
			}
		})
	}
}
