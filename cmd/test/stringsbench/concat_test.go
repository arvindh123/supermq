package main

import (
	"strings"
	"testing"
)

func concatUsingPlus(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "a"
	}
	return s
}

func concatUsingBuilder(n int) string {
	var builder strings.Builder
	for i := 0; i < n; i++ {
		builder.WriteString("a")
	}
	return builder.String()
}

func BenchmarkConcatUsingPlus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		concatUsingPlus(1000)
	}
}

func BenchmarkConcatUsingBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		concatUsingBuilder(1000)
	}
}
