package main

import (
	"testing"
)

func TestPrintTable(t *testing.T) {
	outLine("admin balance", "user1 balance")
	// Output:
	// |admin balance       |user1 balance       |
}
