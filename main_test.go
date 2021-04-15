package main

import (
	"testing"
)

func TestPrintTable(t *testing.T) {
	outLine("admin balance", "user1 balance")
	// Output:
	// |admin balance       |user1 balance       |
}

func TestTokenTransfer(t *testing.T) {
	enrollDemo()
	transferDemo()
	// Output:
	// 200 Admin@org1.example.com enroll successfully
	// 200 User1@org1.example.com enroll successfully
	// |admin balance       |user1 balance       |
	// |100                 |100                 |
	// |95                  |105                 |
	// |103                 |97                  |
}
