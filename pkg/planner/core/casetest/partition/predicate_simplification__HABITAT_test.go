package partition

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func TestIssue62923__HABITAT(t *testing.T) {
	testkit.RunTestUnderCascades(t, func(t *testing.T, tk *testkit.TestKit, cascades, caller string) {
		tk.MustExec("use test")

		// Create test tables with the same structure as in the issue
		tk.MustExec(`CREATE TABLE tlfdfece63 (
		  col_41 timestamp NULL DEFAULT NULL,
		  col_42 json NOT NULL,
		  col_43 varchar(330) COLLATE utf8_general_ci DEFAULT NULL,
		  col_44 char(192) COLLATE utf8_general_ci NOT NULL DEFAULT '^_',
		  col_45 text COLLATE utf8_general_ci DEFAULT NULL,
		  col_46 double DEFAULT '8900.485367052326',
		  col_47 decimal(59,2) DEFAULT NULL,
		  col_48 varchar(493) COLLATE utf8_general_ci DEFAULT NULL,
		  PRIMARY KEY (col_44) /*T![clustered_index] NONCLUSTERED */,
		  KEY idx_20 (col_41,col_47),
		  UNIQUE KEY idx_21 ((cast(col_42 as char(64) array)),col_45(4),col_43(3)),
		  KEY idx_22 (col_48(4),col_46)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci`)

		tk.MustExec(`CREATE TABLE tad03b424 (
	  col_41 timestamp NULL DEFAULT NULL,
	  col_42 json NOT NULL,
	  col_43 varchar(330) COLLATE utf8_general_ci DEFAULT NULL,
	  col_44 char(192) COLLATE utf8_general_ci NOT NULL DEFAULT '^_',
	  col_45 text COLLATE utf8_general_ci DEFAULT NULL,
	  col_46 double DEFAULT '8900.485367052326',
	  col_47 decimal(59,2) DEFAULT NULL,
	  col_48 varchar(493) COLLATE utf8_general_ci DEFAULT NULL,
	  PRIMARY KEY (col_44) /*T![clustered_index] NONCLUSTERED */,
	  KEY idx_20 (col_41,col_47),
	  UNIQUE KEY idx_21 ((cast(col_42 as char(64) array)),col_45(4),col_43(3)),
	  KEY idx_22 (col_48(4),col_46)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci`)

		// The query that triggers the assertion failure
		sql := `
with cte_263 ( col_1350,col_1351,col_1352 ) AS (
select /*+ read_from_storage(tiflash[ tad03b424,tlfdfece63 ]) */ /*+ use_index_merge( tlfdfece63,tad03b424 ) */ /*+ merge_join( tlfdfece63 , tad03b424 */ rpad( tad03b424.col_44 , 6 , tad03b424.col_48 ) as r0 ,
insert( tlfdfece63.col_44 , 0 , 10 , tlfdfece63.col_43 ) as r1 ,
tad03b424.col_41 as r2 from tlfdfece63 ,
tad03b424 where not( tlfdfece63.col_42 = '[\"Sl9DRlDnSdIOxbfequ02VeikDWiphuDO6suBf0F7esJeCWrRJWQbd3BK3vT58Coz\",\"MmC5saHdTUqosY50IrxprAR52oD08XgGhqJCcYeoaDJKrYxBdbi0QuVDDArCghyL\"]' )
order by r0,r1,r2 limit 348170821
) (
select 1,col_1350,col_1351,col_1352
from cte_263
where not( cte_263.col_1352 != '2020-06-30' )
and cte_263.col_1352 in ( null ,'1983-08-09' )
order by 1,2,3,4  )`

		// This should not cause an assertion failure
		tk.MustExec(sql)
	})
}
