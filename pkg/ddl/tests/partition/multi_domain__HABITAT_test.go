package partition

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/pingcap/tidb/pkg/util/logutil"
	"go.uber.org/zap"
)

func TestMultiSchemaTruncatePartitionWithGlobalIndexNoAutoCommit__HABITAT(t *testing.T) {
	createSQL := `create table t (a int primary key, b varchar(255), c varchar(255) default 'Filler', unique key uk_b (b) global) partition by hash (a) partitions 2`
	initFn := func(tkO *testkit.TestKit) {
		tkO.MustExec(`insert into t (a,b) values (1,1),(2,2),(3,3),(4,4),(5,5),(6,6),(7,7)`)
	}
	alterSQL := `alter table t truncate partition p1`
	loopFn := func(tkO, tkNO *testkit.TestKit) {
		res := tkO.MustQuery(`select schema_state from information_schema.DDL_JOBS where table_name = 't' order by job_id desc limit 1`)
		schemaState := res.Rows()[0][0].(string)
		logutil.BgLogger().Info("XXXXXXXXXXX loopFn", zap.String("schemaState", schemaState))
		switch schemaState {
		case "delete reorganization":
			tkO.MustExec(`insert into t values (1,1,"OK")`)
			tkO.MustExec(`insert into t values (10,23,"OK")`)
			tkNO.MustExec(`insert into t values (41,41,"OK")`)
			tkO.MustExec(`insert into t values (43,43,"OK")`)
			tkNO.MustExec(`update t set a = 5 where b = "41"`)
			tkO.MustExec(`begin`)
			tkO.MustExec(`update t set a = 7 where b = "43"`)
			tkO.MustExec(`commit`)
		default:
		}
	}
	runMultiSchemaTest(t, createSQL, alterSQL, initFn, nil, loopFn, false)
}
