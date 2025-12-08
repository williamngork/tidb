package ttlworker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pingcap/tidb/pkg/meta/model"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/pingcap/tidb/pkg/ttl/cache"
	"github.com/pingcap/tidb/pkg/ttl/ttlworker"
	"github.com/stretchr/testify/require"
)

func TestTTLSummaryForTimeoutJob(t *testing.T) {
	store, dom := testkit.CreateMockStoreAndDomain(t)
	waitAndStopTTLManager(t, dom)
	sessionFactory := sessionFactory(t, dom)

	tk := testkit.NewTestKit(t, store)
	m := ttlworker.NewJobManager("test-job-manager", dom.AdvancedSysSessionPool(), store, nil, func() bool { return true })

	se, closeSe := sessionFactory()
	defer closeSe()

	testTable := &cache.PhysicalTable{ID: 1, TableInfo: &model.TableInfo{ID: 1, TTLInfo: &model.TTLInfo{IntervalExprStr: "1", IntervalTimeUnit: int(ast.TimeUnitDay), JobInterval: "1h"}}}
	m.InfoSchemaCache().Tables[testTable.ID] = testTable
	jobID := uuid.NewString()
	_, err := m.LockJob(context.Background(), se, testTable, se.Now(), jobID, false)
	require.NoError(t, err)
	tk.MustQuery("SELECT current_job_id, current_job_owner_id FROM mysql.tidb_ttl_table_status WHERE table_id = ?", 1).Check(testkit.Rows(fmt.Sprintf("%s %s", jobID, m.ID())))

	// insert some finished task
	tk.MustExec("INSERT INTO mysql.tidb_ttl_task (scan_id, job_id, table_id, status, state, expire_time, created_time)"+
		" VALUES (1, ?, 1, 'finished', ?, NOW(), NOW())", jobID, `{"total_rows": 100 ,"success_rows": 100}`)

	// report timeout
	err = m.UpdateHeartBeatForJob(context.Background(), se, se.Now().Add(8*time.Hour), m.RunningJobs()[0])
	require.NoError(t, err)

	// the job should contain summary
	rows := tk.MustQuery("SELECT last_job_summary FROM mysql.tidb_ttl_table_status WHERE table_id = ?", 1).Rows()
	summary := &ttlworker.TTLSummary{}
	require.NoError(t, json.Unmarshal([]byte(rows[0][0].(string)), summary))
	require.Equal(t, uint64(100), summary.TotalRows)
	require.Equal(t, uint64(100), summary.SuccessRows)
	require.Equal(t, uint64(0), summary.ErrorRows)
	require.Equal(t, "job is timeout", summary.ScanTaskErr)
}
