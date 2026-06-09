package testing

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// FormatText writes a human-readable test report.
func FormatText(w io.Writer, report *Report) {
	for _, r := range report.Results {
		if r.Passed {
			fmt.Fprintf(w, "  ✓ %s (%0.fms)\n", r.Name, r.LatencyMs)
		} else if r.Error != "" {
			fmt.Fprintf(w, "  ✗ %s — error: %s\n", r.Name, r.Error)
		} else {
			fmt.Fprintf(w, "  ✗ %s — expected %s, got %s\n", r.Name, r.Expected, r.Actual)
		}
	}

	fmt.Fprintf(w, "\n%d passed, %d failed, %d total (%0.fms)\n", report.Passed, report.Failed, report.Total, report.DurationMs)

	if report.Failed > 0 {
		fmt.Fprintf(w, "\nFailures:\n")
		for _, r := range report.Results {
			if !r.Passed {
				if r.Error != "" {
					fmt.Fprintf(w, "  ✗ %s — error: %s\n", r.Name, r.Error)
				} else {
					fmt.Fprintf(w, "  ✗ %s — expected %s, got %s\n", r.Name, r.Expected, r.Actual)
				}
			}
		}
	} else {
		fmt.Fprintf(w, "\n✓ All tests passed\n")
	}
}

// FormatJSON writes a JSON test report.
func FormatJSON(w io.Writer, report *Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// JUnitTestSuites represents a JUnit XML test report.
type JUnitTestSuites struct {
	XMLName xml.Name        `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite represents a single test suite in JUnit format.
type JUnitTestSuite struct {
	XMLName  xml.Name       `xml:"testsuite"`
	Name     string         `xml:"name,attr"`
	Tests    int            `xml:"tests,attr"`
	Failures int            `xml:"failures,attr"`
	Errors   int            `xml:"errors,attr"`
	Time     float64        `xml:"time,attr"`
	Cases    []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase represents a single test case in JUnit format.
type JUnitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
}

// JUnitFailure represents a test failure in JUnit format.
type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// FormatJUnit writes a JUnit XML test report (for CI/CD integration).
func FormatJUnit(w io.Writer, report *Report, suiteName string) error {
	suite := JUnitTestSuite{
		Name:     suiteName,
		Tests:    report.Total,
		Failures: report.Failed,
		Time:     report.DurationMs / 1000.0,
		Cases:    make([]JUnitTestCase, 0, len(report.Results)),
	}

	for _, r := range report.Results {
		tc := JUnitTestCase{
			Name:      r.Name,
			ClassName: "authgraph.permissions",
			Time:      r.LatencyMs / 1000.0,
		}
		if !r.Passed {
			msg := fmt.Sprintf("expected %s, got %s", r.Expected, r.Actual)
			if r.Error != "" {
				msg = r.Error
			}
			tc.Failure = &JUnitFailure{
				Message: msg,
				Type:    "AssertionError",
				Content: msg,
			}
		}
		suite.Cases = append(suite.Cases, tc)
	}

	suites := JUnitTestSuites{Suites: []JUnitTestSuite{suite}}

	fmt.Fprint(w, xml.Header)
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(suites)
}

// FormatRegressions writes regression detection results.
func FormatRegressions(w io.Writer, regressions []Regression) {
	if len(regressions) == 0 {
		fmt.Fprintf(w, "\n✓ No regressions detected — schema change is safe\n")
		return
	}

	fmt.Fprintf(w, "\n⚠ %d regression(s) detected:\n", len(regressions))
	for _, r := range regressions {
		arrow := "→"
		impact := ""
		if r.Before == "denied" && r.After == "allowed" {
			impact = " (ESCALATION — previously denied access now allowed)"
		} else if r.Before == "allowed" && r.After == "denied" {
			impact = " (BREAKAGE — previously allowed access now denied)"
		}
		fmt.Fprintf(w, "  %s %s: %s %s %s%s\n", arrow, r.TestName, r.Before, arrow, r.After, impact)
	}

	// Summary
	var escalations, breakages int
	for _, r := range regressions {
		if r.Before == "denied" && r.After == "allowed" {
			escalations++
		} else {
			breakages++
		}
	}

	parts := make([]string, 0, 2)
	if escalations > 0 {
		parts = append(parts, fmt.Sprintf("%d escalation(s)", escalations))
	}
	if breakages > 0 {
		parts = append(parts, fmt.Sprintf("%d breakage(s)", breakages))
	}
	fmt.Fprintf(w, "\n  Summary: %s\n", strings.Join(parts, ", "))
}
