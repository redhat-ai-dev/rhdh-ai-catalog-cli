package common

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func AssertLineCompare(t *testing.T, str, expect string, minLine int) {
	t.Helper()
	scanner1 := bufio.NewScanner(strings.NewReader(str))
	scanner2 := bufio.NewScanner(strings.NewReader(expect))
	lineNum := 1
	for scanner1.Scan() && scanner2.Scan() {
		line1 := scanner1.Text()
		line2 := scanner2.Text()

		if line1 != line2 {
			if lineNum < minLine {
				return
			}

			t.Errorf("line diff at line %d between %s and %s", lineNum, line1, line2)
		}
		lineNum++
	}
}

func AssertError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func Equal(expected, got interface{}) bool {
	return reflect.DeepEqual(expected, got)
}

func AssertContains(t *testing.T, s string, subs []string) {
	t.Helper()
	missing := []string{}
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			missing = append(missing, sub)
		}
	}
	if len(missing) > 0 {
		t.Errorf("Expected '%s' to be in '%s'", strings.Join(missing, "\n"), s)
	}
	return
}

func AssertEqual(t *testing.T, e, g interface{}) (r bool) {
	t.Helper()
	if !Equal(e, g) {
		t.Errorf("Expected [%v], got [%v]", e, g)
	}

	return
}

func AssertNotNil(t *testing.T, g interface{}) {
	t.Helper()
	if g == nil {
		t.Errorf("got unexpected nil")
	}
}
