// Copyright 2022 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/planner/core"
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/pingcap/tidb/pkg/testkit/testdata"
	"github.com/stretchr/testify/require"
)

func TestPlanCacheForIntersectionIndexMerge(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("drop table if exists t")
	tk.MustExec("create table t(a int, b int, c int, d int, e int, index ia(a), index ib(b), index ic(c), index id(d), index ie(e))")
	tk.MustExec("prepare stmt from 'select /*+ use_index_merge(t, ia, ib, ic, id, ie) */ * from t where a = 10 and b = ? and c > ? and d is null and e in (0, 100)'")
	tk.MustExec("set @a=1, @b=3")
	tk.MustQuery("execute stmt using @a,@b").Check(testkit.Rows())
	tk.MustQuery("select @@last_plan_from_cache").Check(testkit.Rows("0"))
	tk.MustQuery("execute stmt using @a,@b").Check(testkit.Rows())
	tk.MustQuery("select @@last_plan_from_cache").Check(testkit.Rows("1"))
	tk.MustExec("set @a=100, @b=500")
	tk.MustQuery("execute stmt using @a,@b").Check(testkit.Rows())
	tk.MustQuery("select @@last_plan_from_cache").Check(testkit.Rows("1"))
	tk.MustQuery("execute stmt using @a,@b").Check(testkit.Rows())
	require.True(t, tk.HasPlanForLastExecution("IndexMerge"))
}

func TestIndexMergeWithOrderProperty(t *testing.T) {
	testkit.RunTestUnderCascades(t, func(t *testing.T, testKit *testkit.TestKit, cascades, caller string) {
		testKit.MustExec("use test")
		testKit.MustExec("drop table if exists t")
		testKit.MustExec("create table t (a int, b int, c int, d int, e int, key a(a), key b(b), key c(c), key ac(a, c), key bc(b, c), key ae(a, e), key be(b, e)," +
			" key abd(a, b, d), key cd(c, d))")
		testKit.MustExec("create table t2 (a int, b int, c int, key a(a), key b(b), key ac(a, c))")

		var (
			input  []string
			output []struct {
				SQL  string
				Plan []string
			}
		)
		planSuiteData := core.GetIndexMergeSuiteData()
		planSuiteData.LoadTestCases(t, &input, &output, cascades, caller)
		for i, ts := range input {
			testdata.OnRecord(func() {
				output[i].SQL = ts
				output[i].Plan = testdata.ConvertRowsToStrings(testKit.MustQuery("explain format = 'brief' " + ts).Rows())
			})
			testKit.MustQuery("explain format = 'brief' " + ts).Check(testkit.Rows(output[i].Plan...))
			// Expect no warnings.
			testKit.MustQuery("show warnings").Check(testkit.Rows())
		}
	})
}
