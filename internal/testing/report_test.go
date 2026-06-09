package testing

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
)

func TestFormatText(t *testing.T) {
	report := &Report{
		Total:      3,
		Passed:     2,
		Failed:     1,
		DurationMs: 45,
		Results: []TestResult{
			{Name: "test 1", Passed: true, Expected: "allowed", Actual: "allowed", LatencyMs: 10},
			{Name: "test 2", Passed: true, Expected: "denied", Actual: "denied", LatencyMs: 15},
			{Name: "test 3", Passed: false, Expected: "allowed", Actual: "denied", LatencyMs: 20},
		},
	}

	var buf bytes.Buffer
	FormatText(&buf, report)
	output := buf.String()

	if !strings.Contains(output, "✓ test 1") {
		t.Error("expected pass marker for test 1")
	}
	if !strings.Contains(output, "✗ test 3") {
		t.Error("expected fail marker for test 3")
	}
	if !strings.Contains(output, "2 passed, 1 failed, 3 total") {
		t.Errorf("expected summary line, got: %s", output)
	}
	if !strings.Contains(output, "Failures:") {
		t.Error("expected Failures section")
	}
}

func TestFormatTextAllPassed(t *testing.T) {
	report := &Report{
		Total:      1,
		Passed:     1,
		Failed:     0,
		DurationMs: 5,
		Results: []TestResult{
			{Name: "test 1", Passed: true, Expected: "allowed", Actual: "allowed", LatencyMs: 5},
		},
	}

	var buf bytes.Buffer
	FormatText(&buf, report)
	output := buf.String()

	if !strings.Contains(output, "✓ All tests passed") {
		t.Errorf("expected all passed message, got: %s", output)
	}
}

func TestFormatJSON(t *testing.T) {
	report := &Report{
		Total:      1,
		Passed:     1,
		Failed:     0,
		DurationMs: 5,
		Results: []TestResult{
			{Name: "test 1", Passed: true, Expected: "allowed", Actual: "allowed", LatencyMs: 5},
		},
	}

	var buf bytes.Buffer
	if err := FormatJSON(&buf, report); err != nil {
		t.Fatal(err)
	}

	var decoded Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.Total != 1 {
		t.Errorf("expected total=1, got %d", decoded.Total)
	}
	if decoded.Results[0].Name != "test 1" {
		t.Errorf("expected name 'test 1', got %s", decoded.Results[0].Name)
	}
}

func TestFormatJUnit(t *testing.T) {
	report := &Report{
		Total:      2,
		Passed:     1,
		Failed:     1,
		DurationMs: 100,
		Results: []TestResult{
			{Name: "pass test", Passed: true, Expected: "allowed", Actual: "allowed", LatencyMs: 30},
			{Name: "fail test", Passed: false, Expected: "allowed", Actual: "denied", LatencyMs: 70},
		},
	}

	var buf bytes.Buffer
	if err := FormatJUnit(&buf, report, "permission-tests.yaml"); err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, "<?xml") {
		t.Error("expected XML header")
	}
	if !strings.Contains(output, `name="permission-tests.yaml"`) {
		t.Error("expected suite name in output")
	}

	// Verify valid XML
	var suites JUnitTestSuites
	if err := xml.Unmarshal(buf.Bytes(), &suites); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}
	if len(suites.Suites) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(suites.Suites))
	}
	suite := suites.Suites[0]
	if suite.Tests != 2 {
		t.Errorf("expected 2 tests, got %d", suite.Tests)
	}
	if suite.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", suite.Failures)
	}
	if suite.Cases[1].Failure == nil {
		t.Error("expected failure on second test case")
	}
}

func TestFormatRegressions(t *testing.T) {
	t.Run("no regressions", func(t *testing.T) {
		var buf bytes.Buffer
		FormatRegressions(&buf, nil)
		if !strings.Contains(buf.String(), "No regressions detected") {
			t.Error("expected no-regressions message")
		}
	})

	t.Run("with regressions", func(t *testing.T) {
		regressions := []Regression{
			{TestName: "test-escalation", Before: "denied", After: "allowed"},
			{TestName: "test-breakage", Before: "allowed", After: "denied"},
		}
		var buf bytes.Buffer
		FormatRegressions(&buf, regressions)
		output := buf.String()

		if !strings.Contains(output, "2 regression(s) detected") {
			t.Errorf("expected regression count, got: %s", output)
		}
		if !strings.Contains(output, "ESCALATION") {
			t.Error("expected ESCALATION label")
		}
		if !strings.Contains(output, "BREAKAGE") {
			t.Error("expected BREAKAGE label")
		}
		if !strings.Contains(output, "1 escalation(s)") {
			t.Error("expected escalation count in summary")
		}
		if !strings.Contains(output, "1 breakage(s)") {
			t.Error("expected breakage count in summary")
		}
	})
}
