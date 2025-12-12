package tables_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/pingcap/tidb/pkg/kv"
	"github.com/pingcap/tidb/pkg/meta/model"
	"github.com/pingcap/tidb/pkg/table/tables"
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/pingcap/tidb/pkg/types"
	"github.com/pingcap/tidb/pkg/util/mock"
	"github.com/stretchr/testify/require"
)

func TestForceLockNonUniqueIndexInDDLMergingTempIndex__HABITAT(t *testing.T) {
	tblInfo := buildTableInfo(t, "create table t (id int primary key, k int, key k(k))")

	var idxInfo *model.IndexInfo
	for _, info := range tblInfo.Indices {
		if info.Name.L == "k" {
			idxInfo = info
			break
		}
	}

	require.NotNil(t, idxInfo)
	cases := []struct {
		idxState      model.SchemaState
		backfillState model.BackfillState
		forceLock     bool
	}{
		{model.StateWriteReorganization, model.BackfillStateReadyToMerge, true},
		{model.StateWriteReorganization, model.BackfillStateMerging, true},
		{model.StatePublic, model.BackfillStateInapplicable, false},
	}

	mockCtx := mock.NewContext()
	store := testkit.CreateMockStore(t)
	h := kv.IntHandle(1)
	indexedValues := []types.Datum{types.NewIntDatum(100)}
	idx := tables.NewIndex(tblInfo.ID, tblInfo, idxInfo)
	indexKey, distinct, err := idx.GenIndexKey(mockCtx.ErrCtx(), time.UTC, indexedValues, h, nil)
	require.NoError(t, err)
	require.False(t, distinct)

	for _, c := range cases {
		idxInfo.State = c.idxState
		idxInfo.BackfillState = c.backfillState

		t.Run(fmt.Sprintf("DeleteIndex in %s-%s", c.idxState, c.backfillState), func(t *testing.T) {
			txn, err := store.Begin()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, txn.Rollback())
			}()
			txn.SetOption(kv.Pessimistic, true)

			err = idx.Delete(mockCtx.GetTableCtx(), txn, []types.Datum{types.NewIntDatum(100)}, kv.IntHandle(1))
			require.NoError(t, err)
			flags, err := txn.GetMemBuffer().GetFlags(indexKey)
			require.NoError(t, err)
			require.Equal(t, c.forceLock, flags.HasNeedLocked())
		})

		t.Run(fmt.Sprintf("CreateIndex in %s-%s", c.idxState, c.backfillState), func(t *testing.T) {
			txn, err := store.Begin()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, txn.Rollback())
			}()
			txn.SetOption(kv.Pessimistic, true)

			_, err = idx.Create(mockCtx.GetTableCtx(), txn, indexedValues, h, nil)
			require.NoError(t, err)
			flags, err := txn.GetMemBuffer().GetFlags(indexKey)
			require.NoError(t, err)
			require.Equal(t, c.forceLock, flags.HasNeedLocked())
		})
	}
}
