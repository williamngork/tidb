package core_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestPlanBuilder__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("create table t3 (id int, a int, b int, c int, primary key(id), index idx(b)) partition by hash (id) partitions 3;")
	tk.MustExec("insert into t3 values (1,1,1,1), (2,2,2,2);")
	rows := tk.MustQuery("select /*+ index_lookup_pushdown(t3, idx)*/ * from t3 use index(idx) where b>=0 and b<=10;").Rows()

	require.Equal(t, 2, len(rows))
}
