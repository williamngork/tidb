package analyzetest

import (
	"context"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGeneratedColumns__HABITAT(t *testing.T) {
	store, dom := testkit.CreateMockStoreAndDomain(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")

	// Create table with JSON column and generated columns
	tk.MustExec(`CREATE TABLE test_gen_cols (
		id INT PRIMARY KEY,
		data JSON,
		virtual_col VARCHAR(50) AS (JSON_UNQUOTE(JSON_EXTRACT(data, '$.name'))) VIRTUAL,
		stored_col VARCHAR(50) AS (JSON_UNQUOTE(JSON_EXTRACT(data, '$.status'))) STORED,
		json_but_not_used_by_generated_column JSON,
		INDEX idx_virtual (virtual_col),
		INDEX idx_stored (stored_col)
	)`)

	// Insert test data with simple JSON
	tk.MustExec(`INSERT INTO test_gen_cols (id, data, json_but_not_used_by_generated_column) VALUES
		(1, '{"name": "user1", "status": "active"}', '{"category": "admin", "level": 1}'),
		(2, '{"name": "user2", "status": "inactive"}', '{"category": "user", "level": 2}'),
		(3, '{"name": "user3", "status": "active"}', '{"category": "user", "level": 1}')`)

	// Analyze the table
	tk.MustExec("ANALYZE TABLE test_gen_cols")

	h := dom.StatsHandle()
	tbl, err := dom.InfoSchema().TableByName(context.Background(), ast.NewCIStr("test"), ast.NewCIStr("test_gen_cols"))
	require.NoError(t, err)

	// Get the table statistics
	tblStats := h.GetPhysicalTableStats(tbl.Meta().ID, tbl.Meta())
	require.NotNil(t, tblStats)
	require.True(t, tblStats.IsAnalyzed())

	// For the base column used by generated columns, we should collect statistics even it is not used by any indexes.
	require.True(t, tblStats.GetCol(tbl.Meta().Columns[1].ID).IsAnalyzed())

	// For virtual generated columns, we don't collect statistics because we cannot evaluate the expression on the TiKV side
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[2].ID).IsAnalyzed())

	// For stored generated columns, we collect statistics because the values are stored in TiKV
	require.True(t, tblStats.GetCol(tbl.Meta().Columns[3].ID).IsAnalyzed())

	// For JSON columns that are not used by generated columns, we don't collect statistics because we exclude it by tidb_analyze_skip_column_types.
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[4].ID).IsAnalyzed())

	// For indexes on generated columns, we collect statistics because index entries are stored in TiKV regardless of whether the column is virtual or stored
	require.True(t, tblStats.GetIdx(tbl.Meta().Indices[0].ID).IsAnalyzed())
	require.True(t, tblStats.GetIdx(tbl.Meta().Indices[1].ID).IsAnalyzed())
}

func TestSkipStatsForGeneratedColumnsOnSkippedColumns__HABITAT(t *testing.T) {
	store, dom := testkit.CreateMockStoreAndDomain(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	// Create table with JSON column and generated columns
	tk.MustExec(`CREATE TABLE test_gen_cols (
		id INT PRIMARY KEY,
		data JSON,
		virtual_col VARCHAR(50) AS (JSON_UNQUOTE(JSON_EXTRACT(data, '$.name'))) VIRTUAL,
		stored_col VARCHAR(50) AS (JSON_UNQUOTE(JSON_EXTRACT(data, '$.status'))) STORED
	)`)
	// Insert test data with simple JSON
	tk.MustExec(`INSERT INTO test_gen_cols (id, data) VALUES
		(1, '{"name": "user1", "status": "active"}'),
		(2, '{"name": "user2", "status": "inactive"}'),
		(3, '{"name": "user3", "status": "active"}')`)

	// Explicitly set tidb_analyze_skip_column_types to skip JSON columns
	tk.MustExec("set @@tidb_analyze_skip_column_types = 'json, text, blob'")
	// Check tidb_analyze_skip_column_types setting
	tk.MustQuery("select @@tidb_analyze_skip_column_types").Check(testkit.Rows("json,text,blob"))
	// Analyze the table with all columns
	tk.MustExec("ANALYZE TABLE test_gen_cols all columns")
	h := dom.StatsHandle()
	tbl, err := dom.InfoSchema().TableByName(context.Background(), ast.NewCIStr("test"), ast.NewCIStr("test_gen_cols"))
	require.NoError(t, err)
	// Get the table statistics
	tblStats := h.GetPhysicalTableStats(tbl.Meta().ID, tbl.Meta())
	require.NotNil(t, tblStats)
	require.True(t, tblStats.IsAnalyzed())
	// For JSON column, it should not collect statistics because we skip it
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[1].ID).IsAnalyzed())
	// For virtual generated columns, because it depends on the skipped JSON column, we also skip it
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[2].ID).IsAnalyzed())
	// For stored columns, because it depends on the skipped JSON column, we also skip it
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[3].ID).IsAnalyzed())

	// Test the predicate columns.
	tk.MustExec("select * from test_gen_cols where virtual_col = 'a' and stored_col = 'b'")
	require.NoError(t, h.DumpColStatsUsageToKV())
	// Check the predicate columns collection.
	rows := tk.MustQuery("show column_stats_usage where table_name = 'test_gen_cols'").Rows()
	require.Len(t, rows, 3)
	require.Equal(t, "id", rows[0][3])
	require.Equal(t, "virtual_col", rows[1][3])
	require.Equal(t, "stored_col", rows[2][3])

	tk.MustExec("ANALYZE TABLE test_gen_cols")
	tblStats = h.GetPhysicalTableStats(tbl.Meta().ID, tbl.Meta())
	require.NotNil(t, tblStats)
	require.True(t, tblStats.IsAnalyzed())
	// For JSON column, it should not collect statistics because we skip it
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[1].ID).IsAnalyzed())
	// For virtual generated columns, because it depends on the skipped JSON column, we also skip it
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[2].ID).IsAnalyzed())
	// For stored columns, because it depends on the skipped JSON column, we also skip it
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[3].ID).IsAnalyzed())

	// Remove the skip setting and re-analyze
	tk.MustExec("set @@tidb_analyze_skip_column_types = 'text, blob'")
	tk.MustQuery("select @@tidb_analyze_skip_column_types").Check(testkit.Rows("text,blob"))
	tk.MustExec("ANALYZE TABLE test_gen_cols")
	tblStats = h.GetPhysicalTableStats(tbl.Meta().ID, tbl.Meta())
	require.NotNil(t, tblStats)
	require.True(t, tblStats.IsAnalyzed())
	// For JSON column, it should be analyzed now
	require.True(t, tblStats.GetCol(tbl.Meta().Columns[1].ID).IsAnalyzed())
	// For virtual generated columns, we still could not collect statistics because we couldn't evaluate the expression on the TiKV side
	require.False(t, tblStats.GetCol(tbl.Meta().Columns[2].ID).IsAnalyzed())
	// For stored columns, we can collect statistics because the values are stored in TiKV
	require.True(t, tblStats.GetCol(tbl.Meta().Columns[3].ID).IsAnalyzed())
}
