package core_test

import (
	"github.com/pingcap/tidb/pkg/testkit"
	"testing"
)

func TestSupportForIgnorePlanCacheHint__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)

	tk.MustExec(`use test`)
	tk.MustExec(`create table t (pk int, a int, primary key(pk))`)
	tk.MustExec(`set tidb_enable_non_prepared_plan_cache=1;`)

	tk.MustExec(`select /*+ ignore_plan_cache() */ * from t where pk >= 1`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("0 0"))
	tk.MustExec(`select /*+ ignore_plan_cache() */ * from t where pk >= 1`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("0 0"))

	tk.MustExec(`CREATE BINDING FOR select * from t where pk >= ? USING select * from t where pk >= ?`)
	tk.MustExec(`select  /*+ ignore_plan_cache() */ * from t where pk >= 1`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("1 0"))
	tk.MustExec(`select  /*+ ignore_plan_cache() */ * from t where pk >= 1`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("1 0"))

	tk.MustExec(`DROP BINDING FOR select * from t where pk >= ?`)
	tk.MustExec(`CREATE BINDING FOR select * from t where pk >= ? USING select /*+ ignore_plan_cache() */ * from t where pk >= ?`)
	tk.MustExec(`select * from t where pk >= 1`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("1 0"))
	tk.MustExec(`select * from t where pk >= 1`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("1 0"))

	tk.MustExec(`DROP BINDING FOR select * from t where pk >= ?`)

	tk.MustExec(`prepare st from 'select * from t where pk >= ?'`)
	tk.MustExec(`set @a=4`)
	tk.MustExec(`execute st using @a`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("0 0"))
	tk.MustExec(`execute st using @a`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("0 1"))

	tk.MustExec(`prepare st from 'select /*+ ignore_plan_cache() */ * from t where pk >= ?'`)
	tk.MustExec(`set @a=4`)
	tk.MustExec(`execute st using @a`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("0 0"))
	tk.MustExec(`execute st using @a`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("0 0"))

	tk.MustExec(`CREATE BINDING FOR select * from t where pk >= ? USING select /*+ ignore_plan_cache() */ * from t where pk >= ?`)
	tk.MustExec(`prepare st from 'select * from t where pk >= ?'`)
	tk.MustExec(`set @a=4`)
	tk.MustExec(`execute st using @a`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("1 0"))
	tk.MustExec(`execute st using @a`)
	tk.MustQuery(`select @@last_plan_from_binding, @@last_plan_from_cache`).Check(testkit.Rows("1 0"))
}
