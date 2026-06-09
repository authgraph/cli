package testing

// Suite represents a complete permission test file with setup, tests, and teardown.
type Suite struct {
	Setup    *Setup     `yaml:"setup,omitempty"`
	Tests    []TestCase `yaml:"tests"`
	Teardown *Teardown  `yaml:"teardown,omitempty"`
}

// Setup defines pre-test schema and tuple provisioning.
type Setup struct {
	Schema *SchemaSetup `yaml:"schema,omitempty"`
	Tuples []Tuple      `yaml:"tuples,omitempty"`
}

// SchemaSetup specifies a schema to push before running tests.
type SchemaSetup struct {
	File   string `yaml:"file,omitempty"`
	Inline string `yaml:"inline,omitempty"`
}

// Teardown defines cleanup after tests.
type Teardown struct {
	Tuples       []Tuple `yaml:"tuples,omitempty"`
	DeleteSchema bool    `yaml:"delete_schema,omitempty"`
	CleanupSetup bool    `yaml:"cleanup_setup,omitempty"` // auto-remove setup tuples
}

// Tuple represents a relationship tuple for setup/teardown.
type Tuple struct {
	Subject  string `yaml:"subject"`
	Relation string `yaml:"relation"`
	Resource string `yaml:"resource"`
	// Conditions
	ExpiresIn      string `yaml:"expires_in,omitempty"`
	NonEscalatable bool   `yaml:"non_escalatable,omitempty"`
	Budget         int    `yaml:"budget,omitempty"`
}

// TestCase represents a single permission assertion.
type TestCase struct {
	Name  string `yaml:"name"`
	Check *Check `yaml:"check,omitempty"`
	// Support legacy flat format
	Subject    string `yaml:"subject,omitempty"`
	Permission string `yaml:"permission,omitempty"`
	Resource   string `yaml:"resource,omitempty"`
	Expected   string `yaml:"expected,omitempty"`
	// Nested check format uses "expect" field
	Expect string `yaml:"expect,omitempty"`
}

// Check represents a permission check assertion (nested format).
type Check struct {
	Subject  string `yaml:"subject"`
	Action   string `yaml:"action"`
	Resource string `yaml:"resource"`
}

// GetSubject returns the subject from either flat or nested format.
func (tc *TestCase) GetSubject() string {
	if tc.Check != nil {
		return tc.Check.Subject
	}
	return tc.Subject
}

// GetAction returns the action/permission from either format.
func (tc *TestCase) GetAction() string {
	if tc.Check != nil {
		return tc.Check.Action
	}
	return tc.Permission
}

// GetResource returns the resource from either format.
func (tc *TestCase) GetResource() string {
	if tc.Check != nil {
		return tc.Check.Resource
	}
	return tc.Resource
}

// GetExpected returns the expected result ("allowed" or "denied").
func (tc *TestCase) GetExpected() string {
	if tc.Check != nil {
		return tc.Expect
	}
	return tc.Expected
}

// IsAllowed returns true if the test expects the check to be allowed.
func (tc *TestCase) IsAllowed() bool {
	e := tc.GetExpected()
	return e == "allowed" || e == "allow" || e == "true"
}

// TestResult holds the outcome of a single test case.
type TestResult struct {
	Name      string  `json:"name"`
	Passed    bool    `json:"passed"`
	Expected  string  `json:"expected"`
	Actual    string  `json:"actual"`
	LatencyMs float64 `json:"latency_ms"`
	Error     string  `json:"error,omitempty"`
	Path      string  `json:"path,omitempty"`
}

// Report holds the full test run results.
type Report struct {
	Total      int          `json:"total"`
	Passed     int          `json:"passed"`
	Failed     int          `json:"failed"`
	Skipped    int          `json:"skipped"`
	DurationMs float64      `json:"duration_ms"`
	Results    []TestResult `json:"results"`
}
