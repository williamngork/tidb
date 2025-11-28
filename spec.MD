Enhance DDL operations by embedding the ANALYZE operation for both `ALTER TABLE ... MODIFY COLUMN` and `ALTER TABLE ... CREATE INDEX` events, ensuring fresh statistics are available immediately after these operations complete.

### Scope
- Modify the DDL job execution flow to include statistics collection

### Core Functionality
1. **Configuration**
    - New session variable: `tidb_enable_ddl_analyze` (default: 0)
2. **DDL Operations with Embedded ANALYZE**
    - when tidb_enable_ddl_analyze is set to 1, Execute ANALYZE as part of the DDL job for both:
        - `ALTER TABLE ... CREATE INDEX` operations
        - `ALTER TABLE ... MODIFY COLUMN` operations

### Data Flow
1. User call `set @@tidb_enable_ddl_analyze = 1` to enable this feature in their session
2. User executes either:
    - `ALTER TABLE {table name} ADD INDEX...`
    - `ALTER TABLE {table name} MODIFY COLUMN...`
3. DDL job is created with new analyze operation
4. Index creation proceeds as usual
5. After DDL operation completes (index backfill or column modification), analyze operation begins
6. Statistics are collected and stored
7. Job completes

### API Changes

1. **New Session Variables**

```sql
-- Enable/disable embedded analyze
SET @@tidb_enable_ddl_analyze = {0|1};
```
