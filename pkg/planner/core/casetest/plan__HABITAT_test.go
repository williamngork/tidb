package casetest

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func TestOuterJoinElimination__HABITAT(t *testing.T) {
	testkit.RunTestUnderCascades(t, func(t *testing.T, testKit *testkit.TestKit, cascades, caller string) {
		testKit.MustExec("use test")
		testKit.MustExec(`create table t1 (a int, b int, c int)`)
		testKit.MustExec(`create table t2 (a int, b int, c int)`)
		testKit.MustExec(`create table t2_k (a int, b int, c int, key(a))`)
		testKit.MustExec(`create table t2_uk (a int, b int, c int, unique key(a))`)
		testKit.MustExec(`create table t2_nnuk (a int not null, b int, c int, unique key(a))`)
		testKit.MustExec(`create table t2_pk (a int, b int, c int, primary key(a))`)

		// test subqueries in the select list with no_decorrelate_in_select=OFF
		testKit.MustExec("set @@tidb_opt_enable_no_decorrelate_in_select=OFF")
		testKit.MustHavePlan("select t1a.a, if(exists(select 1 from t2_uk t2b where t2b.a = t1a.a), 1, 0) as founda from t1 t1a left join t2_pk t2 on t1a.a = t2.a", "Join")
		// test subqueries in the select list with no_decorrelate_in_select=ON
		testKit.MustExec("set @@tidb_opt_enable_no_decorrelate_in_select=ON")
		testKit.MustNotHavePlan("select t1a.a, if(exists(select 1 from t2_uk t2b where t2b.a = t1a.a), 1, 0) as founda from t1 t1a left join t2_pk t2 on t1a.a = t2.a", "Join")
		// next query correlates on t2, so outer join elimination can't be applied
		testKit.MustHavePlan("select t1a.a, if(exists(select 1 from t2_uk t2b where t2b.a = t2.a), 1, 0) as founda from t1 t1a left join t2_pk t2 on t1a.a = t2.a", "Join")
	})
}
