package core_test

import (
	"testing"
	"time"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestSetVarHintsWorks__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec(`use test`)

	// test that the timestamp continues to update (default behavior)
	tk.MustExec(`set timestamp=default;`)
	require.Equal(t, "42", tk.MustQuery(`select /*+ set_var(timestamp=1) */ @@timestamp + 41;`).Rows()[0][0].(string))
	firstts := tk.MustQuery(`select @@timestamp;`).Rows()[0][0].(string)
	require.Eventually(t, func() bool {
		return firstts < tk.MustQuery(`select @@timestamp;`).Rows()[0][0].(string)
	}, time.Second, time.Microsecond*10)
}
