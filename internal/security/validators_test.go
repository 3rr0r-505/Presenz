// internal/security/validators_test.go

package security

import "testing"

func TestSanitizeText(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"trims whitespace", "  John Doe  ", "John Doe"},
		{"strips control chars", "John\x07Doe", "JohnDoe"},
		{"strips tab and trims", "\tJohn Doe\n", "John Doe"},
		{"empty string", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SanitizeText(c.input)
			if got != c.want {
				t.Errorf("SanitizeText(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	const maxLen = 100

	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid simple name", "John Doe", "John Doe", false},
		{"valid with dot", "John D. Doe", "John D. Doe", false},
		{"too short", "J", "", true},
		{"too long", stringOfLen(101), "", true},
		{"contains digits", "John123", "", true},
		{"contains special chars", "John@Doe", "", true},
		{"sanitized before length check", "  Jo  ", "Jo", false},
		{"empty string", "", "", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ValidateName(c.input, maxLen)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateName(%q) error = %v, wantErr %v", c.input, err, c.wantErr)
			}
			if !c.wantErr && got != c.want {
				t.Errorf("ValidateName(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestValidateRoll(t *testing.T) {
	const maxLen = 50

	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid alphanumeric", "abc123", "ABC123", false},
		{"valid with hyphen", "abc-123", "ABC-123", false},
		{"uppercased on success", "roll99", "ROLL99", false},
		{"single char valid", "a", "A", false},
		{"too long", stringOfLen(51), "", true},
		{"contains space", "abc 123", "", true},
		{"contains special char", "abc@123", "", true},
		{"empty string", "", "", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ValidateRoll(c.input, maxLen)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateRoll(%q) error = %v, wantErr %v", c.input, err, c.wantErr)
			}
			if !c.wantErr && got != c.want {
				t.Errorf("ValidateRoll(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestValidateSessionCode(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		expected string
		wantErr  bool
	}{
		{"exact match", "ABC123", "ABC123", false},
		{"mismatch", "XYZ999", "ABC123", true},
		{"sanitized before compare", "  ABC123  ", "ABC123", false},
		{"case sensitive mismatch", "abc123", "ABC123", true},
		{"empty vs non-empty", "", "ABC123", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateSessionCode(c.code, c.expected)
			if (err != nil) != c.wantErr {
				t.Errorf("ValidateSessionCode(%q, %q) error = %v, wantErr %v", c.code, c.expected, err, c.wantErr)
			}
		})
	}
}

func TestValidationErrorType(t *testing.T) {
	_, err := ValidateName("", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Errorf("expected *ValidationError, got %T", err)
	}
}

// stringOfLen builds a string of exactly n valid-alphabet characters,
// used to test max-length boundaries without hardcoding long literals.
func stringOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
