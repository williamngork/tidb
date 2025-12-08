package physicalplantest

import (
	"flag"
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/pingcap/tidb/pkg/testkit/testdata"
	"github.com/pingcap/tidb/pkg/testkit/testmain"
	"github.com/pingcap/tidb/pkg/testkit/testsetup"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRuleAggElimination4Join(t *testing.T) {
	testkit.RunTestUnderCascades(t, func(t *testing.T, tk *testkit.TestKit, cascades, caller string) {
		tk.MustExec("use test")
		tk.MustExec("drop table if exists t1, t2")
		tk.MustExec("CREATE TABLE t1 ( id1 int NOT NULL, id2 int NOT NULL, id3 int NOT NULL, id4 int NOT NULL, UNIQUE KEY `UK_id1_id2` (id1,id2));")
		tk.MustExec("CREATE TABLE t2 ( id1 int NOT NULL, id2 int NOT NULL, id3 int NOT NULL, id4 int NOT NULL, UNIQUE KEY `UK_id1_id2` (id1,id2));")
		tk.MustExec("CREATE TABLE t3 ( id1 int NOT NULL, id2 int NOT NULL, id3 int NOT NULL, id4 int NOT NULL, UNIQUE KEY `UK_id1_id2` (id1,id2));")
		tk.MustExec("CREATE TABLE t4 ( id1 int NOT NULL, id2 int NOT NULL, id3 int NOT NULL, id4 int NOT NULL, KEY `UK_id1_id2` (id1,id2));")

		var input []string
		var output []struct {
			SQL  string
			Plan []string
			Warn []string
		}

		cascadesData := getCascadesTemplateData()
		cascadesData.LoadTestCases(t, &input, &output, cascades, caller)

		for i, ts := range input {
			testdata.OnRecord(func() {
				output[i].SQL = ts
				output[i].Plan = testdata.ConvertRowsToStrings(tk.MustQuery("explain format = 'brief' " + ts).Rows())
				output[i].Warn = testdata.ConvertSQLWarnToStrings(tk.Session().GetSessionVars().StmtCtx.GetWarnings())
			})
			tk.MustQuery("explain format = 'brief' " + ts).Check(testkit.Rows(output[i].Plan...))
			require.Equal(t, output[i].Warn, testdata.ConvertSQLWarnToStrings(tk.Session().GetSessionVars().StmtCtx.GetWarnings()))
		}
	})
}

// Helper: load cascades_template test data
func getCascadesTemplateData() testdata.TestData {
	var testDataMap = make(testdata.BookKeeper)
	testDataMap.LoadTestSuiteData("testdata", "cascades_template", true)
	return testDataMap["cascades_template"]
}

var testDataMap = make(testdata.BookKeeper)

func TestCascades__HABITAT(m *testing.M) {
	testsetup.SetupForCommonTest()
	flag.Parse()
	testDataMap.LoadTestSuiteData("testdata", "plan_suite")
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("github.com/golang/glog.(*fileSink).flushDaemon"),
		goleak.IgnoreTopFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
		goleak.IgnoreTopFunction("github.com/lestrrat-go/httprc.runFetchWorker"),
		goleak.IgnoreTopFunction("go.etcd.io/etcd/client/pkg/v3/logutil.(*MergeLogger).outputLoop"),
		goleak.IgnoreTopFunction("gopkg.in/natefinch/lumberjack%2ev2.(*Logger).millRun"),
		goleak.IgnoreTopFunction("github.com/tikv/client-go/v2/txnkv/transaction.keepAlive"),
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	}

	callback := func(i int) int {
		testDataMap.GenerateOutputIfNeeded()
		return i
	}

	goleak.VerifyTestMain(testmain.WrapTestingM(m, callback), opts...)
}

func GetPlanSuiteData() testdata.TestData {
	return testDataMap["plan_suite"]
}
