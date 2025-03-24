package main

import (
	"strings"
	"testing"
)

func TestTopic(t *testing.T) {
	output := Topic()

	// Basic validation
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}

	// The output should indicate which year's data is being displayed
	if !strings.Contains(output, "F1 Data for ") {
		t.Error("Output should indicate the F1 data year")
	}

	// The output should contain race information or a status message
	if !strings.Contains(output, "Next Race:") && !strings.Contains(output, "Next race:") {
		t.Error("Output should contain race information or status")
	}

	// The output should contain driver standings or an error message
	if !strings.Contains(output, "Driver Standings") && !strings.Contains(output, "Driver standings error:") {
		t.Error("Output should contain driver standings or error message")
	}

	// The output should contain constructor standings or an error message
	if !strings.Contains(output, "Constructor Standings") && !strings.Contains(output, "Constructor standings error:") {
		t.Error("Output should contain constructor standings or error message")
	}

	// Log the output for manual verification
	t.Logf("Topic output:\n%s", output)
}

func TestSlackTopic(t *testing.T) {
	output := SlackTopic()

	// Basic validation
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}

	// The output should start with the F1 emoji and year
	if !strings.Contains(output, ":f1: ") {
		t.Error("Output should start with F1 emoji")
	}

	// The output should contain next race information
	if !strings.Contains(output, "Next: R") {
		t.Error("Output should contain next race information")
	}

	// The output should contain standings
	if !strings.Contains(output, "Standings:") {
		t.Error("Output should contain standings")
	}

	// The output should contain country flag emoji
	if !strings.Contains(output, ":flag-") {
		t.Error("Output should contain country flag emoji")
	}

	// The output should contain fantasy code
	if !strings.Contains(output, "Fantasy: `") {
		t.Error("Output should contain fantasy code")
	}

	// Check that output is within Slack character limit
	if len(output) > 250 {
		t.Error("Output exceeds Slack's 250 character limit")
	}

	// Log the output for manual verification
	t.Logf("SlackTopic output:\n%s", output)
	t.Logf("Character count: %d/250", len(output))
}
