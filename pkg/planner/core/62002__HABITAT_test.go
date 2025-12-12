package core

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func Test62002__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")

	// Create the test table with the same structure as in the issue
	tk.MustExec(`CREATE TABLE table_test (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		apply_id bigint(20) NOT NULL,
		form_id int(11) NOT NULL,
		ws_id int(11) DEFAULT NULL,
		ordinal int(11) DEFAULT NULL,
		col1 json DEFAULT NULL,
		col2 json DEFAULT NULL,
		col3 json DEFAULT NULL,
		col4 json DEFAULT NULL,
		col5 json DEFAULT NULL,
		col6 json DEFAULT NULL,
		col7 json DEFAULT NULL,
		col8 json DEFAULT NULL,
		col9 json DEFAULT NULL,
		col10 json DEFAULT NULL,
		col11 json DEFAULT NULL,
		col12 json DEFAULT NULL,
		col13 json DEFAULT NULL,
		col14 json DEFAULT NULL,
		col15 json DEFAULT NULL,
		col16 json DEFAULT NULL,
		col17 json DEFAULT NULL,
		col18 json DEFAULT NULL,
		col19 json DEFAULT NULL,
		col20 json DEFAULT NULL,
		col21 json DEFAULT NULL,
		col22 json DEFAULT NULL,
		col23 json DEFAULT NULL,
		col24 json DEFAULT NULL,
		col25 json DEFAULT NULL,
		col26 json DEFAULT NULL,
		col27 json DEFAULT NULL,
		col28 json DEFAULT NULL,
		col29 json DEFAULT NULL,
		col30 json DEFAULT NULL,
		create_time datetime DEFAULT CURRENT_TIMESTAMP,
		last_update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		being_deleted tinyint(1) DEFAULT '0',
		apply_status int(11) NOT NULL DEFAULT '0',
		PRIMARY KEY (id,form_id),
		KEY idx_form_apply_deleted (form_id,apply_id,being_deleted),
		KEY idx_form_ordinal_deleted (form_id,ordinal,being_deleted)
	)`)

	// Test the query from the issue
	tk.MustQuery(`SELECT
		cast(if(s.column5 is null OR s.column5 = '', 0, s.column5) AS decimal(23,8)) AS column5,
		s.column1 AS column1,
		s.column9 AS column9,
		s.column16 AS column16,
		s.column17 AS column17,
		s.column18 AS column18,
		s.column10 AS column10,
		s.column3 AS column3,
		s.column4 AS column4,
		s.column6 AS column6,
		s.column7 AS column7
	FROM (
		SELECT
			apply_id,
			JSON_UNQUOTE(JSON_EXTRACT(col5, '$[0].value')) AS column5,
			replace(replace(replace(json_unquote(json_extract(col10, '$[*].value')), '[\"', ''), '\"]', ''), '\",\"', ',') AS column10,
			JSON_UNQUOTE(JSON_EXTRACT(col6, '$[0].value')) AS column6,
			JSON_UNQUOTE(JSON_EXTRACT(col7, '$[0].value')) AS column7,
			JSON_UNQUOTE(JSON_EXTRACT(col1, '$[0].value')) AS column1,
			JSON_UNQUOTE(JSON_EXTRACT(col9, '$[0].value')) AS column9,
			col16 -> '$[*].optUid' AS column16,
			replace(replace(replace(json_unquote(json_extract(col16, '$[*].value')), '[\"', ''), '\"]', ''), '\",\"', ',') AS order_column16,
			JSON_UNQUOTE(JSON_EXTRACT(col17, '$[0].value')) AS column17,
			JSON_UNQUOTE(JSON_EXTRACT(col3, '$[0].value')) AS column3,
			JSON_UNQUOTE(JSON_EXTRACT(col18, '$[0].value')) AS column18,
			JSON_UNQUOTE(JSON_EXTRACT(col4, '$[0].value')) AS column4
		FROM (
			SELECT
				table_test.apply_id,
				col5,
				col10,
				col6,
				col7,
				col1,
				col9,
				col16,
				col17,
				col3,
				col18,
				col4,
				table_test.ordinal,
				table_test.being_deleted
			FROM table_test
			WHERE table_test.form_id = 9093606
			AND table_test.being_deleted = 0
			AND table_test.apply_status != 1
		) ta24e
		WHERE ordinal = -1
		AND being_deleted = 0
	) AS s
	WHERE (s.column1 = '标准')
	ORDER BY
		column5 asc,
		CONVERT(column1 USING GBK) asc,
		CONVERT(column9 USING GBK) asc,
		CONVERT(column16 USING GBK) asc,
		column17 asc,
		column18 asc,
		CONVERT(column10 USING GBK) asc,
		column3 asc,
		CONVERT(column4 USING GBK) asc,
		CONVERT(column6 USING GBK) asc,
		CONVERT(column7 USING GBK) ASC
	LIMIT 0,20`)
}
