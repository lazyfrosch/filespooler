package util

import "testing"

func TestGetNameFromTCPAddr(t *testing.T) {
	var result string
	tests := map[string]string{
		"localhost":              "localhost",
		"localhost:12345":        "localhost",
		"fqdn.example.com:12345": "fqdn.example.com",
		"[127.0.0.1]:12345":      "127.0.0.1",
		"[::1]:12345":            "::1",
	}

	for name, expected := range tests {
		result = GetNameFromTCPAddr(name)
		if result != expected {
			t.Fatalf("result %s for %s is not expected %s", result, name, expected)
		}
	}
}
