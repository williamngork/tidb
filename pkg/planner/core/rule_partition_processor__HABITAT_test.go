package core_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func TestRulePartition__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec(` CREATE TABLE tla842d94a (
       col_1 varchar(188) CHARACTER SET gbk COLLATE gbk_bin NOT NULL,
       col_2 double NOT NULL,
       PRIMARY KEY (col_1,col_2) /*T![clustered_index] NONCLUSTERED */,
       UNIQUE KEY idx_2 (col_1,col_2),
       UNIQUE KEY idx_3 (col_1,col_2),
       KEY idx_4 (col_1,col_2) /*T![global_index] GLOBAL */
     ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
     PARTITION BY RANGE COLUMNS(col_1)
     (PARTITION p0 VALUES LESS THAN ('E恘l57'),
      PARTITION p1 VALUES LESS THAN ('MboOU0'),
      PARTITION p2 VALUES LESS THAN ('Q&h髑UDZ娻躸(襲!籂35'),
      PARTITION p3 VALUES LESS THAN ('f獟@'),
      PARTITION p4 VALUES LESS THAN ('~W噽纓'));`)
	tk.MustQuery(`SELECT /*+ set_var(tidb_partition_prune_mode="static") */
    1,
    char(tla842d94a.col_2, tla842d94a.col_2 using utf8mb4) AS col_383,
    tla842d94a.col_2 AS col_384
FROM tla842d94a
WHERE tla842d94a.col_1 IN ('與P)凥i5', 'AI禡=Ymm滕籔湾$IUKiF3撔')
AND char(tla842d94a.col_2, tla842d94a.col_2 using utf8mb4) IN ('9eQ)6nzji', 'bF!pOc~')
AND NOT (tla842d94a.col_2 <> 3496.9237290113774)
ORDER BY char(tla842d94a.col_2, tla842d94a.col_2 using utf8mb4), tla842d94a.col_2;
`)
}
