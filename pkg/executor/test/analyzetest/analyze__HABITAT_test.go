package analyzetest

import (
	"github.com/pingcap/tidb/pkg/testkit"
	"testing"
)

func TestSkipStatsForGeneratedColumnsOnSkippedColumns__HABITAT(t *testing.T) {
	store, _ := testkit.CreateMockStoreAndDomain(t)
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
}
