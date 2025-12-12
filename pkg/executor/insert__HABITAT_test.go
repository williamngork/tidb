package executor_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func TestInsertDuplicateToGeneratedColumns__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec(`
		CREATE TABLE t3 (
		id int,
		time1 datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		time2 datetime GENERATED ALWAYS AS (date_add(time1, interval 1 day)) VIRTUAL,
		KEY idx (id, time2)
		);
		`)
	tk.MustExec("insert into t3 set id = 1;")
	tk.MustExec("update t3 set id = 2;")
	tk.MustExec("admin check table t3;")
}
