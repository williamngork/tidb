package ddl_test

import (
	"context"
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestMultiSchemaChangeWithoutMDL__HABITAT(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	defer func() {
		tk.MustExec("set global tidb_enable_metadata_lock = default;")
	}()

	testcases := []struct {
		name string
		sql  string
	}{
		{"drop column", "alter table t drop column col2, drop column col3"},
		{"drop index", "alter table t drop index idx2, drop index idx3"},
		{"modify column", "alter table t modify column col2 bigint, modify column col3 bigint"},
		{"modify column with reorg", "alter table t modify column col2 varchar(4), modify column col3 varchar(4)"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tk.MustExec("set global tidb_enable_metadata_lock = on;")
			tk.MustExec("drop table if exists t;")
			tk.MustExec("create table t(col1 int, col2 int, col3 int, index idx2(col2), index idx3(col2));")
			tk.MustExec("set global tidb_enable_metadata_lock = off;")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Use a channel to capture the result
			done := make(chan error, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						if err, ok := r.(error); ok {
							done <- err
						} else {
							done <- errors.Errorf("%v", r)
						}
					}
				}()
				tk.MustExecWithContext(ctx, tc.sql)
				done <- nil
			}()

			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					t.Fatalf("Test timed out: %s", tc.name)
				}
			case err := <-done:
				if err != nil {
					t.Fatalf("Test failed: %v", err)
				}
			}
		})
	}
}
