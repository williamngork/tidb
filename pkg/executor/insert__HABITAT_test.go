package executor_test

import (
	"fmt"
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func TestInsertDuplicateToGeneratedColumns__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec(`
		CREATE TABLE tmv (
  			J1 json,
			J2 json GENERATED ALWAYS AS (j1) VIRTUAL,
  			UNIQUE KEY i1 ((cast(j1 as signed array))),
  			KEY i2 ((cast(j2 as signed array))))`)
	tk.MustExec("insert into tmv set j1 = '[1]'")
	tk.MustExec("insert ignore into tmv set j1 = '[1]' on duplicate key update j1 = '[2]'")
	for _, enabled := range []bool{false, true} {
		tk.MustExec(fmt.Sprintf("set @@tidb_enable_fast_table_check = %v", enabled))
		tk.MustExec("admin check table tmv;")
	}

	tk.MustExec(`
		CREATE TABLE ttime (
			id int primary key,
			t1 datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			t2 datetime GENERATED ALWAYS AS (date_add(t1, interval 1 day)) VIRTUAL,
			t3 datetime GENERATED ALWAYS AS (date_add(t2, interval 1 day)) VIRTUAL,
			t4 datetime GENERATED ALWAYS AS (date_sub(t2, interval 10 MINUTE)) VIRTUAL,
			UNIQUE KEY i1 (t1),
			UNIQUE KEY i2 (t2),
			UNIQUE KEY i3 (id, t3),
			KEY i4 (t4))`)
	tk.MustExec(`insert into ttime set id = 1, t1 = "2011-12-20 17:15:50"`)
	tk.MustExec("insert into ttime set id = 1 on duplicate key update id = 2")
	for _, enabled := range []bool{false, true} {
		tk.MustExec(fmt.Sprintf("set @@tidb_enable_fast_table_check = %v", enabled))
		tk.MustExec("admin check table ttime;")
	}
}
