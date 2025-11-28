package ddl_test

import (
	"context"
	"testing"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestAddIndexWithAnalyze__HABITAT(t *testing.T) {
	store, dom := testkit.CreateMockStoreAndDomain(t)
	tk := testkit.NewTestKit(t, store)
	// add index
	tk.MustExec("use test")
	tk.MustExec("set @@tidb_enable_ddl_analyze = 1")
	tk.MustExec("create table t(a int NOT NULL DEFAULT 10, b int, index idx_b(b))")
	for i := range 50 {
		tk.MustExec("insert into t values (?, ?)", i, i)
	}
	tk.MustExec("ALTER TABLE t ADD index idx(a)")
	tk.MustQuery("select * from t use index(idx) where a >1")
	tk.MustQuery("select * from t use index(idx_b) where b >1")
	// get meta elements
	tbl, err := dom.InfoSchema().TableByName(context.Background(), ast.NewCIStr("test"), ast.NewCIStr("t"))
	require.NoError(t, err)
	aInfo := tbl.Meta().FindPublicColumnByName("a")
	require.NotNil(t, aInfo)
	bInfo := tbl.Meta().FindPublicColumnByName("b")
	require.NotNil(t, bInfo)
	idxInfo := tbl.Meta().FindIndexByName("idx")
	require.NotNil(t, idxInfo)
	idxBInfo := tbl.Meta().FindIndexByName("idx_b")
	require.NotNil(t, idxBInfo)
	// check the stats handle
	statsTable, ok := dom.StatsHandle().StatsCache.Get(tbl.Meta().ID)
	require.True(t, ok)
	// check the stats element is analyzed
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(aInfo.ID, false))
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(idxInfo.ID, true))
	require.NotNil(t, statsTable.HistColl.GetCol(aInfo.ID))
	require.NotNil(t, statsTable.HistColl.GetIdx(idxInfo.ID))
	colAStatsVer := statsTable.HistColl.GetCol(aInfo.ID).Histogram.LastUpdateVersion
	indexAStatsVer := statsTable.HistColl.GetIdx(idxInfo.ID).Histogram.LastUpdateVersion
	require.Equal(t, colAStatsVer, indexAStatsVer)
	// for other columns and indexes, they are also analyzed.
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(bInfo.ID, false))
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(idxBInfo.ID, true))
	require.NotNil(t, statsTable.HistColl.GetCol(bInfo.ID))
	require.NotNil(t, statsTable.HistColl.GetIdx(idxBInfo.ID))
	colBStatsVer := statsTable.HistColl.GetCol(bInfo.ID).Histogram.LastUpdateVersion
	indexBStatsVer := statsTable.HistColl.GetIdx(idxBInfo.ID).Histogram.LastUpdateVersion
	require.Equal(t, colBStatsVer, indexBStatsVer)
	require.Equal(t, indexAStatsVer, indexBStatsVer)

	// test alter column
	tk.MustExec("ALTER TABLE t modify column a varchar(10)")
	tk.MustQuery("select * from t use index(idx) where a >1")
	tk.MustQuery("select * from t use index(idx_b) where b >1")
	// reload the schema info
	tbl, err = dom.InfoSchema().TableByName(context.Background(), ast.NewCIStr("test"), ast.NewCIStr("t"))
	require.NoError(t, err)
	aInfo = tbl.Meta().FindPublicColumnByName("a")
	require.NotNil(t, aInfo)
	idxInfo = tbl.Meta().FindIndexByName("idx")
	require.NotNil(t, idxInfo)
	// check the stats handle
	statsTable, ok = dom.StatsHandle().StatsCache.Get(tbl.Meta().ID)
	require.True(t, ok)
	// check the stats element is analyzed
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(aInfo.ID, false))
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(idxInfo.ID, true))
	require.NotNil(t, statsTable.HistColl.GetCol(aInfo.ID))
	require.NotNil(t, statsTable.HistColl.GetIdx(idxInfo.ID))
	colAStatsVer2 := statsTable.HistColl.GetCol(aInfo.ID).Histogram.LastUpdateVersion
	indexAStatsVer2 := statsTable.HistColl.GetIdx(idxInfo.ID).Histogram.LastUpdateVersion
	require.Equal(t, colAStatsVer2, indexAStatsVer2)
	// colsAStatsVer2 is not same as colAStatsVer
	require.True(t, colAStatsVer != colAStatsVer2)
	// for other columns and indexes, they are also analyzed.
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(bInfo.ID, false))
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(idxBInfo.ID, true))
	require.NotNil(t, statsTable.HistColl.GetCol(bInfo.ID))
	require.NotNil(t, statsTable.HistColl.GetIdx(idxBInfo.ID))
	colBStatsVer2 := statsTable.HistColl.GetCol(bInfo.ID).Histogram.LastUpdateVersion
	indexBStatsVer2 := statsTable.HistColl.GetIdx(idxBInfo.ID).Histogram.LastUpdateVersion
	require.Equal(t, colAStatsVer2, colBStatsVer2)
	require.Equal(t, indexAStatsVer2, indexBStatsVer2)
	// colsAStatsVer2 is not same as colAStatsVer
	require.True(t, colBStatsVer != colBStatsVer2)

	// for partition table, add index with analyze should be banned.
	tk.MustExec("CREATE TABLE pt(id INT NOT NULL, stu_id INT NOT NULL) " +
		"PARTITION BY RANGE (stu_id) (PARTITION p0 VALUES LESS THAN (25),PARTITION p1 VALUES LESS THAN (51))")
	for i := range 50 {
		tk.MustExec("insert into pt values (?,?)", i, i)
	}
	tk.MustExec("analyze table pt all columns")
	// reload the schema info
	tbl, err = dom.InfoSchema().TableByName(context.Background(), ast.NewCIStr("test"), ast.NewCIStr("pt"))
	require.NoError(t, err)
	idInfo := tbl.Meta().FindPublicColumnByName("id")
	require.NotNil(t, idInfo)
	// check the stats handle
	statsTable, ok = dom.StatsHandle().StatsCache.Get(tbl.Meta().ID)
	require.True(t, ok)
	// column is analyzed from last time.
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(idInfo.ID, false))
	require.NotNil(t, statsTable.HistColl.GetCol(idInfo.ID))
	idColStatsVer := statsTable.HistColl.GetCol(idInfo.ID).Histogram.LastUpdateVersion

	tk.MustExec("ALTER TABLE pt ADD index idx(id)")
	tk.MustQuery("select * from pt use index(idx) where id >1")
	// reload the schema info
	tbl, err = dom.InfoSchema().TableByName(context.Background(), ast.NewCIStr("test"), ast.NewCIStr("pt"))
	require.NoError(t, err)
	idInfo = tbl.Meta().FindPublicColumnByName("id")
	require.NotNil(t, idInfo)
	idxInfo = tbl.Meta().FindIndexByName("idx")
	require.NotNil(t, idxInfo)

	// check the stats handle
	statsTable, ok = dom.StatsHandle().StatsCache.Get(tbl.Meta().ID)
	require.True(t, ok)
	// column is analyzed from last time.
	require.True(t, statsTable.ColAndIdxExistenceMap.Has(idInfo.ID, false))
	// index is not analyzed.
	require.False(t, statsTable.ColAndIdxExistenceMap.Has(idInfo.ID, true))
	// column analyze time is same as before.
	require.NotNil(t, statsTable.HistColl.GetCol(idInfo.ID))
	idColStatsVer2 := statsTable.HistColl.GetCol(idInfo.ID).Histogram.LastUpdateVersion
	// assert that column id is analyzed from
	require.Equal(t, idColStatsVer, idColStatsVer2)

	// modify column for partitioned table.
	tk.MustGetErrMsg("ALTER TABLE pt modify column id varchar(10)", "[ddl:8200]Unsupported modify column: table is partition table")

	tk.MustExec("drop table if exists t1;")
	tk.MustExec("create table t1( id int, a int, b int, index idx(id, a));")
	tk.MustExec("insert into t1 values (1, 1, 1), (2, 2, 2), (3, 3, 3), (4, 4, 4), (5, 5, 5);")
	tk.MustExec("analyze table t1 all columns with 1 topn, 10 buckets;")
	tk.MustExec(" ALTER TABLE t1 ADD INDEX idx_a(a), ADD INDEX idx_b(b);")
	tk.MustQuery("explain select * from t1 where a = 1;").CheckContain("TableFullScan")
}
