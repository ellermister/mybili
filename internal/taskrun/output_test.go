package taskrun

import "testing"

func TestTaskOutputKeys(t *testing.T) {
	if taskLogKey(12) != "mybili:tasklog:12" {
		t.Fatalf("log key = %s", taskLogKey(12))
	}
	if taskProgKey(12) != "mybili:taskprog:12" {
		t.Fatalf("progress key = %s", taskProgKey(12))
	}
	if outputTTL.Hours() != 12 {
		t.Fatalf("ttl = %s", outputTTL)
	}
}
