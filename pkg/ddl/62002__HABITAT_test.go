package ddl_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func Test62002__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")

	// Create a range-partitioned table similar to the one in the issue
	tk.MustExec(`CREATE TABLE t_part_range (
		col_95 char(181) COLLATE gbk_bin NOT NULL DEFAULT 'SaMKHTyg+nlID-X3Y',
		PRIMARY KEY (col_95) CLUSTERED
	) ENGINE=InnoDB DEFAULT CHARSET=gbk COLLATE=gbk_bin
	PARTITION BY RANGE COLUMNS(col_95) (
		PARTITION p0 VALUES LESS THAN ('6)nvX^uj0UGxqX'),
		PARTITION p1 VALUES LESS THAN ('BHSluf6'),
		PARTITION p2 VALUES LESS THAN (MAXVALUE)
	)`)

	// Test case 1: The original query from the issue
	tk.MustQuery(`SELECT col_95 FROM t_part_range WHERE
		col_95 BETWEEN 'Dyw=*7nigCMh' AND 'Im0*7sZ'
		OR col_95 IN ('58y-j)84-&Y*', 'WNe(rS5uwmvIvFnHw', 'j9FsMawX5uBro%$p', 'C(#EQm@J')
	GROUP BY col_95
	HAVING col_95 BETWEEN '%^2' AND '38ABfC-'
	   OR col_95 BETWEEN 'eKCAE$d2x_hxscj' AND 'zcw35^ATEEp1md=L'`)

	// Test case 2: Simplified version to verify basic functionality
	tk.MustExec(`INSERT INTO t_part_range (col_95) VALUES
		('Dyw=*7nigCMh'),  -- Should be in p2
		('BHSluf6'),       -- Boundary of p1/p2
		('5)nvX^uj0UGxqX') -- Should be in p0
	`)

	// Verify data is in the correct partitions
	result := tk.MustQuery(`SELECT col_95 FROM t_part_range PARTITION(p0)`).Rows()
	require.Equal(t, 1, len(result), "Expected 1 row in p0")
	require.Equal(t, "5)nvX^uj0UGxqX", result[0][0])

	result = tk.MustQuery(`SELECT col_95 FROM t_part_range PARTITION(p2)`).Sort().Rows()
	require.Equal(t, 2, len(result), "Expected 2 rows in p2")
	require.Equal(t, "BHSluf6", result[0][0])
	require.Equal(t, "Dyw=*7nigCMh", result[1][0])

	// Clean up
	tk.MustExec("DROP TABLE IF EXISTS t_part_range")
}
