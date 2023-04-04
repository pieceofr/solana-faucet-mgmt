package main

import (
	"testing"
	"time"
)

func TestIsExpired(t *testing.T) {
	// Test case 1: Time is after duration
	t1 := time.Now().UTC()
	d1 := time.Hour * 24
	t1 = t1.Add(d1)
	result1 := IsExpired(t1)
	expected1 := false
	if result1 != expected1 {
		t.Errorf("Test case 1 failed: Expected %v but got %v", expected1, result1)
	}

	// Test case 2: Time is before duration
	t2 := time.Now().UTC()
	time.Sleep(3 * time.Second)
	result2 := IsExpired(t2)
	expected2 := true
	if result2 != expected2 {
		t.Errorf("Test case 2 failed: Expected %v but got %v", expected2, result2)
	}
}

func TestTimePlusSeconds(t *testing.T) {
	// Test case 1: Add 60 seconds to the current time
	now := time.Now()
	future := TimePlusSeconds(now, 60)
	expected := now.Add(60 * time.Second)
	if future != expected {
		t.Errorf("Expected %v but got %v", expected, future)
	}

	// Test case 2: Add 120 seconds to a specific time
	t1 := time.Date(2023, time.April, 1, 0, 0, 0, 0, time.UTC)
	future = TimePlusSeconds(t1, 120)
	expected = time.Date(2023, time.April, 1, 0, 2, 0, 0, time.UTC)
	if future != expected {
		t.Errorf("Expected %v but got %v", expected, future)
	}
}
