package cardinality_test

import (
	"strconv"
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestOrderingIdxSelectivityRatioForJoin__HABITAT(t *testing.T) {
	store, dom := testkit.CreateMockStoreAndDomain(t)
	testKit := testkit.NewTestKit(t, store)

	testKit.MustExec("use test")
	testKit.MustExec("drop table if exists t")
	testKit.MustExec("create table t(a int, b int, c int, index ibc(b, c))")
	testKit.MustExec("insert into t values (1,1,1), (2,2,2), (3,3,3), (4,4,4), (5,5,5), (6,6,6), (7,7,7), (8,8,8), (9,9,9), (10,10,10)")
	testKit.MustExec("insert into t select a,b,c from t")
	testKit.MustExec("analyze table t")
	h := dom.StatsHandle()
	require.Nil(t, h.DumpStatsDeltaToKV(true))

	// Discourage merge join and hash join to encourage index join, and discourage topn to encourage limit.
	testKit.MustExec("set @@session.tidb_opt_merge_join_cost_factor = 1000")
	testKit.MustExec("set @@session.tidb_opt_hash_join_cost_factor = 1000")
	testKit.MustExec("set @@session.tidb_opt_topn_cost_factor = 1000")
	// Disable idx_selectivity_ratio using -1 and 0 - both should have no effect.
	testKit.MustExec("set @@session.tidb_opt_ordering_index_selectivity_ratio = -1")
	rs := testKit.MustQuery("explain format=verbose select t1.* from t t1 use index (ibc) join t t2 on t1.b=t2.b where t2.c=5 order by t1.b limit 2").Rows()
	planCost1, err1 := strconv.ParseFloat(rs[0][2].(string), 64)
	require.Nil(t, err1)
	testKit.MustExec("set @@session.tidb_opt_ordering_index_selectivity_ratio = 0")
	rs = testKit.MustQuery("explain format=verbose select t1.* from t t1 use index (ibc) join t t2 on t1.b=t2.b where t2.c=5 order by t1.b limit 2").Rows()
	planCost2, err2 := strconv.ParseFloat(rs[0][2].(string), 64)
	require.Nil(t, err2)
	require.Equal(t, planCost1, planCost2)
	// Increasing the ratio should increase the cost of index join, so the plan cost should increase.
	testKit.MustExec("set @@session.tidb_opt_ordering_index_selectivity_ratio = 0.5")
	rs = testKit.MustQuery("explain format=verbose select t1.* from t t1 use index (ibc) join t t2 on t1.b=t2.b where t2.c=5 order by t1.b limit 2").Rows()
	planCost3, err3 := strconv.ParseFloat(rs[0][2].(string), 64)
	require.Nil(t, err3)
	require.Less(t, planCost2, planCost3)
	testKit.MustExec("set @@session.tidb_opt_ordering_index_selectivity_ratio = 1")
	rs = testKit.MustQuery("explain format=verbose select t1.* from t t1 use index (ibc) join t t2 on t1.b=t2.b where t2.c=5 order by t1.b limit 2").Rows()
	planCost4, err4 := strconv.ParseFloat(rs[0][2].(string), 64)
	require.Nil(t, err4)
	require.Less(t, planCost3, planCost4)
}
