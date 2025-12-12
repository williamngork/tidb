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
CREATE TABLE tmv (
  J1 json, J2 json GENERATED ALWAYS AS (j1) VIRTUAL,
  UNIQUE KEY i1 ((cast(j1 as signed array))),
  KEY i2 ((cast(j2 as signed array)))
);		`)
	tk.MustExec("insert into tmv set j1 = '[1]';")
	tk.MustExec("insert ignore into tmv set j1 = '[1]' on duplicate key update j1 = '[2]';")
	tk.MustExec("admin check table tmv;")
}
