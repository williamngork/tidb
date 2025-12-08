package parser_test

import "testing"

func TestExplain__HABITAT(t *testing.T) {
	table := []testCase{
		{"EXPLAIN 'sqldigest'", true, "EXPLAIN FORMAT = 'row' 'sqldigest'"},
		{"EXPLAIN ANALYZE 'sqldigest'", true, "EXPLAIN ANALYZE 'sqldigest'"},
		{"EXPLAIN format='json' 'sqldigest'", true, "EXPLAIN FORMAT = 'json' 'sqldigest'"},
		{"EXPLAIN ANALYZE format='json' 'sqldigest'", true, "EXPLAIN ANALYZE FORMAT = 'json' 'sqldigest'"},
	}
	RunTest(t, table, false)
}
