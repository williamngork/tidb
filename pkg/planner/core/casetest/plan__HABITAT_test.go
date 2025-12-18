package casetest

import (
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPreferRangeScanForDNF__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("set @@session.tidb_opt_prefer_range_scan=1")
	tk.MustExec("set @@global.tidb_enable_auto_analyze='OFF'")
	tk.MustExec("use test")
	tk.MustExec("drop table if exists t")
	// Create table without inserting data to test pseudo stats behavior
	tk.MustExec("create table t (a int, b int, c int, index idx_a_b(a, b))")

	// DNF with only equal predicates - should prefer IndexLookUp
	//result := tk.MustQuery("explain format='brief' select * from t where (a = 1 and b = 1) or (a = 2 and b = 2)")
	//require.Contains(t, result.Rows()[0][0], "IndexLookUp")
	// Test 1: Simple DNF with two conditions
	dnfQuery1 := "explain format='brief' select * from t where (a = 1 and b = 1) or (a = 2 and b = 2)"
	inQuery1 := "explain format='brief' select * from t where (a, b) in ((1, 1), (2, 2))"
	resultDNF := tk.MustQuery(dnfQuery1)
	resultIN := tk.MustQuery(inQuery1)
	require.Contains(t, resultDNF.Rows()[0][0], "IndexLookUp")
	require.Equal(t, resultDNF.Rows()[0][0], resultIN.Rows()[0][0], "DNF and IN should produce the same plan")

	// Longer DNF with only equal predicates - should prefer IndexLookUp
	//result := tk.MustQuery("explain format='brief' select * from t where a = 1 or a = 3 or a = 5 or a = 7 or a = 9 or a = 11 or a = 13 or a = 15 or a = 17 or a = 19 or a = 21 or a = 23 or a = 25 or a = 27 or a = 29 or a = 31 or a = 33 or a = 35 or a = 37 or a = 39 or a = 41 or a = 43 or a = 45 or a = 47 or a = 49 or a = 51 or a = 53 or a = 55 or a = 57 or a = 59")
	//require.Contains(t, result.Rows()[0][0], "IndexLookUp")
	longDNF := "a = 1 or a = 3 or a = 5 or a = 7 or a = 9 or a = 11 or a = 13 or a = 15 or a = 17 or a = 19 or " +
		"a = 21 or a = 23 or a = 25 or a = 27 or a = 29 or a = 31 or a = 33 or a = 35 or a = 37 or a = 39 or " +
		"a = 41 or a = 43 or a = 45 or a = 47 or a = 49 or a = 51 or a = 53 or a = 55 or a = 57 or a = 59"
	longIN := "a in (1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 35, 37, 39, " +
		"41, 43, 45, 47, 49, 51, 53, 55, 57, 59)"

	dnfQuery2 := "explain format='brief' select * from t where " + longDNF
	inQuery2 := "explain format='brief' select * from t where " + longIN
	resultDNF = tk.MustQuery(dnfQuery2)
	resultIN = tk.MustQuery(inQuery2)
	require.Contains(t, resultDNF.Rows()[0][0], "IndexLookUp")
	require.Equal(t, resultDNF.Rows()[0][0], resultIN.Rows()[0][0], "Long DNF and IN should produce the same plan")

	// Restore settings
	tk.MustExec("set @@global.tidb_enable_auto_analyze='ON'")
	tk.MustExec("set @@session.tidb_opt_prefer_range_scan=1")
}
