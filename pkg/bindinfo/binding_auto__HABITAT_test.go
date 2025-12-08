package bindinfo_test

import (
	"strings"
	"testing"

	"github.com/pingcap/tidb/pkg/parser/auth"
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestExplainExploreVerifyAndBind__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	require.NoError(t, tk.Session().Auth(&auth.UserIdentity{Username: "root", Hostname: "%"}, nil, nil, nil))
	tk.MustExec("use test")
	tk.MustExec(`create table t (a int, b int, key(a))`)
	tk.MustExec(`insert into t values (1, 2), (2, 3), (3, 4), (4, 5)`)

	tk.MustQuery(`select * from t`)
	tk.MustQuery(`select @@last_plan_from_binding`).Check(testkit.Rows("0"))
	require.True(t, len(tk.MustQuery(`show global bindings`).Rows()) == 0) // no binding

	rs := tk.MustQuery(`explain explore select * from t`).Rows()
	runStmt := rs[0][12].(string)     // explain analyze <plan_digest>
	bindingStmt := rs[0][13].(string) // create global binding from history using plan digest <plan_digest>

	require.True(t, strings.HasPrefix(runStmt, "EXPLAIN ANALYZE"))
	require.True(t, strings.HasPrefix(bindingStmt, "CREATE GLOBAL BINDING"))

	rs = tk.MustQuery(runStmt).Rows()
	require.True(t, strings.Contains(rs[0][0].(string), "TableReader")) // table scan and no error

	tk.MustExec(bindingStmt)
	tk.MustQuery(`select * from t`)
	tk.MustQuery(`select @@last_plan_from_binding`).Check(testkit.Rows("1"))
	require.True(t, len(tk.MustQuery(`show global bindings`).Rows()) == 1)
}
