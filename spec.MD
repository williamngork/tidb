Enhance DDL operations by embedding the ANALYZE operation for both `ALTER TABLE ... MODIFY COLUMN` and `ALTER TABLE ... CREATE INDEX` events, ensuring fresh statistics are available immediately after these operations complete.

### Scope
- Modify the DDL job execution flow to include statistics collection

### Core Functionality
1. **Configuration**
    - New session variable: `tidb_enable_ddl_analyze`
2. **DDL Operations with Embedded ANALYZE**
    - when tidb_enable_ddl_analyze is set to 1, Execute ANALYZE as part of the DDL job for both:
        - `ALTER TABLE ... CREATE INDEX` operations
        - `ALTER TABLE ... MODIFY COLUMN` operations
    - partition tables or multi-index ADD INDEX DDL should not get automatic ANALYZE
### Data Flow
1. User call `set @@tidb_enable_ddl_analyze = 1` to enable this feature in their session
2. User executes either:
    - `ALTER TABLE {table name} ADD INDEX...`
    - `ALTER TABLE {table name} MODIFY COLUMN...`
3. Index creation / column modification proceeds as usual, along with analyze operation
4. Statistics are collected and stored

### API Changes

1. **New Session Variables**

```sql
-- Enable/disable embedded analyze
SET @@tidb_enable_ddl_analyze = {0|1};
```
