package casetest_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestOuterJoinEliminationInExecution__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec(`create table t1 (a int, b int, c int)`)
	tk.MustExec("drop table if exists t1,t2,t3")
	tk.MustExec("create table t1 (id bigint primary key, parent_id bigint NOT NULL, payload varchar(255));")
	tk.MustExec("create table t2 (id bigint primary key, payload varchar(255));")

	tk.MustExec("create table t3 (parent_id bigint primary key, search_field varchar(255), key (search_field));")

	tk.MustExec("insert into t3 values (1,1),(2,2),(3,3),(4,4),(5,5),(6,6),(7,7),(8,8);")
	tk.MustExec("insert ignore into t3 select (select max(parent_id) from t3) + t3.parent_id * (t3_2.parent_id + 1), concat(t3.parent_id, t3_2.parent_id) from t3, t3 t3_2;")
	tk.MustExec("insert ignore into t3 select (select max(parent_id) from t3) + t3.parent_id * (t3_2.parent_id + 1), concat(t3.parent_id, t3_2.parent_id) from t3, t3 t3_2;")
	tk.MustExec("insert ignore into t3 select (select max(parent_id) from t3) + t3.parent_id * (t3_2.parent_id + 1), concat(t3.parent_id, t3_2.parent_id) from t3, t3 t3_2;")
	tk.MustExec("insert into t2 select * from t3;")
	tk.MustExec("insert into t1 select parent_id, parent_id, search_field from t3;")
	tk.MustExec("update t3 set search_field = 'Find me' order by rand() limit 4;")
	tk.MustExec("analyze table t1,t2,t3;")
	// Run the query and get EXPLAIN ANALYZE output
	explainRows := tk.MustQuery(`EXPLAIN ANALYZE
		SELECT DISTINCT 'Found' as w1
		FROM (SELECT t3.parent_id AS found_id, t3.search_field as search_term
		      FROM t3
		      JOIN t2 ON t3.parent_id = t2.id
		      WHERE t3.search_field = 'Find me'
		      LIMIT 202) sub_query
		LEFT OUTER JOIN t3 ON t3.parent_id = sub_query.found_id
		LEFT OUTER JOIN t1 ON t1.parent_id = sub_query.found_id
		LIMIT 101;`).Rows()

	for _, row := range explainRows {
		// should not contain outer join
		for _, item := range row {
			require.NotContains(t, item, "outer join")
		}
	}
}

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
