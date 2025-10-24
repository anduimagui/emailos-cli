package helpers

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
	
	mailos "github.com/anduimagui/emailos-cli"
)

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	
	if !reflect.DeepEqual(expected, actual) {
		msg := "Values are not equal"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	
	if reflect.DeepEqual(expected, actual) {
		msg := "Values should not be equal"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: both values are %v", msg, expected)
	}
}

// AssertTrue checks if a value is true
func AssertTrue(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	
	if !value {
		msg := "Expected true but got false"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Error(msg)
	}
}

// AssertFalse checks if a value is false
func AssertFalse(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	
	if value {
		msg := "Expected false but got true"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Error(msg)
	}
}

// AssertNil checks if a value is nil
func AssertNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	
	if value != nil {
		msg := "Expected nil value"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: got %v", msg, value)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	
	if value == nil {
		msg := "Expected non-nil value"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Error(msg)
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	
	if !strings.Contains(str, substr) {
		msg := "String does not contain expected substring"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: '%s' does not contain '%s'", msg, str, substr)
	}
}

// AssertNotContains checks if a string does not contain a substring
func AssertNotContains(t *testing.T, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	
	if strings.Contains(str, substr) {
		msg := "String should not contain substring"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: '%s' contains '%s'", msg, str, substr)
	}
}

// AssertMatches checks if a string matches a regular expression
func AssertMatches(t *testing.T, str, pattern string, msgAndArgs ...interface{}) {
	t.Helper()
	
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		t.Fatalf("Invalid regex pattern: %s", pattern)
	}
	
	if !matched {
		msg := "String does not match pattern"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: '%s' does not match pattern '%s'", msg, str, pattern)
	}
}

// AssertLen checks if a slice/map/string has the expected length
func AssertLen(t *testing.T, obj interface{}, expectedLen int, msgAndArgs ...interface{}) {
	t.Helper()
	
	v := reflect.ValueOf(obj)
	actualLen := v.Len()
	
	if actualLen != expectedLen {
		msg := "Length mismatch"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: expected length %d, got %d", msg, expectedLen, actualLen)
	}
}

// AssertEmpty checks if a slice/map/string is empty
func AssertEmpty(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	
	v := reflect.ValueOf(obj)
	if v.Len() != 0 {
		msg := "Expected empty collection"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: collection has %d elements", msg, v.Len())
	}
}

// AssertNotEmpty checks if a slice/map/string is not empty
func AssertNotEmpty(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	
	v := reflect.ValueOf(obj)
	if v.Len() == 0 {
		msg := "Expected non-empty collection"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Error(msg)
	}
}

// AssertTimeAlmostEqual checks if two times are within a delta
func AssertTimeAlmostEqual(t *testing.T, expected, actual time.Time, delta time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	
	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	
	if diff > delta {
		msg := "Times are not within expected delta"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: expected %v, got %v (delta: %v)", msg, expected, actual, diff)
	}
}

// AssertValidEmail checks if a string is a valid email address
func AssertValidEmail(t *testing.T, email string, msgAndArgs ...interface{}) {
	t.Helper()
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		msg := "Invalid email format"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: '%s' is not a valid email", msg, email)
	}
}

// AssertEmailStructure validates the structure of an Email object
func AssertEmailStructure(t *testing.T, email *mailos.Email, msgAndArgs ...interface{}) {
	t.Helper()
	
	msg := "Email structure validation failed"
	if len(msgAndArgs) > 0 {
		msg = msgAndArgs[0].(string)
	}
	
	if email == nil {
		t.Errorf("%s: email is nil", msg)
		return
	}
	
	if email.ID <= 0 {
		t.Errorf("%s: email ID should be positive, got %d", msg, email.ID)
	}
	
	if email.From == "" {
		t.Errorf("%s: email From field is empty", msg)
	} else {
		AssertValidEmail(t, email.From, "Invalid From email")
	}
	
	if len(email.To) == 0 {
		t.Errorf("%s: email To field is empty", msg)
	} else {
		for i, to := range email.To {
			AssertValidEmail(t, to, "Invalid To[%d] email", i)
		}
	}
	
	if email.Subject == "" {
		t.Errorf("%s: email Subject field is empty", msg)
	}
	
	if email.Date.IsZero() {
		t.Errorf("%s: email Date field is zero", msg)
	}
}

// AssertAttachmentsValid validates email attachments
func AssertAttachmentsValid(t *testing.T, email *mailos.Email, msgAndArgs ...interface{}) {
	t.Helper()
	
	msg := "Attachment validation failed"
	if len(msgAndArgs) > 0 {
		msg = msgAndArgs[0].(string)
	}
	
	if email == nil {
		t.Errorf("%s: email is nil", msg)
		return
	}
	
	if len(email.Attachments) != len(email.AttachmentData) {
		t.Errorf("%s: attachment count mismatch - names: %d, data: %d", 
			msg, len(email.Attachments), len(email.AttachmentData))
	}
	
	for _, filename := range email.Attachments {
		if filename == "" {
			t.Errorf("%s: empty attachment filename", msg)
		}
		
		if _, exists := email.AttachmentData[filename]; !exists {
			t.Errorf("%s: missing attachment data for '%s'", msg, filename)
		}
	}
	
	for filename, data := range email.AttachmentData {
		if len(data) == 0 {
			t.Errorf("%s: empty attachment data for '%s'", msg, filename)
		}
	}
}

// AssertCommandOutputContains checks if command output contains expected text
func AssertCommandOutputContains(t *testing.T, output, expected string, msgAndArgs ...interface{}) {
	t.Helper()
	
	if !strings.Contains(output, expected) {
		msg := "Command output does not contain expected text"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: output '%s' does not contain '%s'", msg, output, expected)
	}
}

// AssertNoError checks that an error is nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	
	if err != nil {
		msg := "Unexpected error"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: %v", msg, err)
	}
}

// AssertError checks that an error is not nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	
	if err == nil {
		msg := "Expected error but got nil"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Error(msg)
	}
}

// AssertErrorContains checks that an error contains specific text
func AssertErrorContains(t *testing.T, err error, expectedText string, msgAndArgs ...interface{}) {
	t.Helper()
	
	if err == nil {
		msg := "Expected error but got nil"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Error(msg)
		return
	}
	
	if !strings.Contains(err.Error(), expectedText) {
		msg := "Error does not contain expected text"
		if len(msgAndArgs) > 0 {
			msg = msgAndArgs[0].(string)
		}
		t.Errorf("%s: error '%s' does not contain '%s'", msg, err.Error(), expectedText)
	}
}