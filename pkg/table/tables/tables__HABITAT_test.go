package tables_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func TestCast__HABITAT(t *testing.T) {
	store, _ := testkit.CreateMockStoreAndDomain(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("DROP DATABASE IF EXISTS test_hidden;")
	tk.MustExec("CREATE DATABASE test_hidden;")
	tk.MustExec("USE test_hidden;")

	tk.MustExec("drop table if exists t;")
	tk.MustExec("create table t (col timestamp);")
	tk.MustExec("explain format='brief' select cast(col as char) from t group by cast(col as char);")
}
