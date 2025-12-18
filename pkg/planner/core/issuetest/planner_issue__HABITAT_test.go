package issuetest

import (
	"github.com/pingcap/tidb/pkg/testkit"
	"testing"
)

func TestOnlyFullGroupCantFeelUnaryConstant__HABITAT(t *testing.T) {
	testkit.RunTestUnderCascades(t, func(t *testing.T, testKit *testkit.TestKit, cascades, caller string) {
		testKit.MustExec("use test")
		testKit.MustExec("drop table if exists t")
		testKit.MustExec("create table t(a int);")
		testKit.MustQuery("select a,min(a) from t where a=-1;").Check(testkit.Rows("<nil> <nil>"))
		testKit.MustQuery("select a,min(a) from t where -1=a;").Check(testkit.Rows("<nil> <nil>"))
	})
}
