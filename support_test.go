package main

import (
	"testing"
)

func expectInt(t *testing.T, expected, actual int) (ok bool) {
	if actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
		return
	}

	return true
}

func expectBool(t *testing.T, expected, actual bool) (ok bool) {
	if actual != expected {
		t.Errorf("expected: %t, got: %t", expected, actual)
		return
	}

	return true
}

func expectString(t *testing.T, expected, actual string) (ok bool) {
	if actual != expected {
		t.Errorf("expected: %s, got: %s", expected, actual)
		return
	}

	return true
}

func expectKeyValue(t *testing.T, m map[string]string, k, expected string) (ok bool) {
	v, ok := m[k]
	if !ok {
		t.Errorf("expected key %s not found", k)
		return
	}

	if !expectString(t, expected, v) {
		return
	}

	return true
}
