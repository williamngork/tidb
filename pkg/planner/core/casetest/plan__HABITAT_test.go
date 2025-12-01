package casetest_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func TestOuterJoinElimination__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec(`create table t1 (a int, b int, c int)`)
	tk.MustExec(`create table t2 (a int, b int, c int)`)
	tk.MustExec(`create table t2_k (a int, b int, c int, key(a))`)
	tk.MustExec(`create table t2_uk (a int, b int, c int, unique key(a))`)
	tk.MustExec(`create table t2_nnuk (a int not null, b int, c int, unique key(a))`)
	tk.MustExec(`create table t2_pk (a int, b int, c int, primary key(a))`)

	// test distinct aggregation
	tk.MustNotHavePlan("select distinct t1.a from t1 left join t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct t1.a from t1 left join t2_k t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct t1.a from t1 left join t2_uk t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct t1.a from t1 left join t2_nnuk t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct t1.a from t1 left join t2_pk t2 on t1.a = t2.a", "Join")
	// test constant columns with distinct
	tk.MustNotHavePlan("select distinct 1 from t1 left join t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct 1 from t1 left join t2_k t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct 1 from t1 left join t2_uk t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct 1 from t1 left join t2_nnuk t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct 1 from t1 left join t2_pk t2 on t1.a = t2.a", "Join")
	// test constant columns with distinct
	tk.MustHavePlan("select 1 from t1 left join t2 on t1.a = t2.a", "Join")
	tk.MustHavePlan("select 1 from t1 left join t2_k t2 on t1.a = t2.a", "Join")
	tk.MustHavePlan("select 1 from t1 left join t2_uk t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select 1 from t1 left join t2_nnuk t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select 1 from t1 left join t2_pk t2 on t1.a = t2.a", "Join")
	// test subqueries
	tk.MustHavePlan("select 1 from (select distinct a from t1) t1 left join t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct 1 from (select distinct a from t1) t1 left join t2 on t1.a = t2.a", "Join")
	tk.MustHavePlan("select t1.a from (select distinct a from t1) t1 left join t2 on t1.a = t2.a", "Join")
	tk.MustNotHavePlan("select distinct t1.a from (select distinct a from t1) t1 left join t2 on t1.a = t2.a", "Join")
}
