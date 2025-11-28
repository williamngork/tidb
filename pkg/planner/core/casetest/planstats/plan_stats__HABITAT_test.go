package planstats_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pingcap/tidb/pkg/domain"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/pingcap/tidb/pkg/testkit/testdata"
	"github.com/pingcap/tidb/pkg/util/dbutil"
	"github.com/stretchr/testify/require"
)

func TestStatsAnalyzedInDDL__HABITAT(t *testing.T) {
	testkit.RunTestUnderCascadesWithDomain(t, func(t *testing.T, testKit *testkit.TestKit, dom *domain.Domain, cascades, caller string) {
		testKit.MustExec("use test")
		testKit.MustExec("set session tidb_enable_ddl_analyze = 1")
		// test normal table
		testKit.MustExec("create table t(a int, b int, c int, primary key(a), key idx(b))")

		// insert data
		for i := range 50 {
			testKit.MustExec("insert into t values (?,?,?)", i, i, i)
		}
		var (
			input  []string
			output []struct {
				Query  string
				Result []string
			}
		)
		testData := testDataMap["plan_stats_suite__HABITAT"]
		testData.LoadTestCases(t, &input, &output, cascades, caller)
		var (
			lastIsSelect     bool
			lastStatsVersion string
			curStatsVersion  string
		)
		getHistID := func(name string, isIndex bool) (int64, int64) {
			require.NoError(t, dom.Reload())
			is := dom.InfoSchema()
			tbl, err := is.TableByName(context.Background(), ast.NewCIStr("test"), ast.NewCIStr("t"))
			require.NoError(t, err)
			tableID := tbl.Meta().ID
			histID := int64(0)
			if isIndex {
				histID = tbl.Meta().FindIndexByName(name).ID
			} else {
				histID = dbutil.FindColumnByName(tbl.Meta().Columns, name).ID
			}
			return tableID, histID
		}
		for i, sql := range input {
			isSelect := strings.HasPrefix(sql, "select")
			testdata.OnRecord(func() {
				output[i].Query = input[i]
				if isSelect {
					// explain the query
					output[i].Result = testdata.ConvertRowsToStrings(testKit.MustQuery("explain format='brief' " + sql).Rows())
				} else {
					output[i].Result = nil
				}
			})
			if isSelect {
				// explain the query
				testKit.MustQuery("explain format='brief' " + sql).Check(testkit.Rows(output[i].Result...))
				// assert the version
				indexName := ""
				if strings.Contains(sql, "idx_c") {
					indexName = "idx_c"
				} else {
					indexName = "idx_bc"
				}
				tableID, histID := getHistID(indexName, true)
				res := testKit.MustQuery("select version from mysql.stats_histograms where table_id = ? and hist_id = ? and is_index=?;", tableID, histID, true)
				if len(res.Rows()) > 0 {
					curStatsVersion = res.Rows()[0][0].(string)
					// since the index is re-analyzed, so each usage of them use a new version.
					if lastIsSelect {
						require.Equal(t, lastStatsVersion, curStatsVersion)
					} else {
						// last is ddl
						require.NotEqual(t, lastStatsVersion, curStatsVersion)
					}
					lastStatsVersion = curStatsVersion
				}
				lastIsSelect = true
			} else {
				// run the ddl anyway
				testKit.MustExec(sql)
				lastIsSelect = false
			}
		}
	})
}
