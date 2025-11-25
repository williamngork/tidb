# Technical Specification: Embedded ANALYZE in DDL Operations

## 1. Overview

### 1.1 Purpose
Enhance DDL operations by embedding the ANALYZE operation for both `ALTER TABLE MODIFY COLUMN` and `CREATE INDEX` events, ensuring fresh statistics are available immediately after these operations complete.

### 1.2 Scope
- Modify the DDL job execution flow to include statistics collection
- Add new session variables and execution options
- Ensure backward compatibility

## 2. Technical Requirements

### 2.1 Functional Requirements

#### 2.1.1 Core Functionality
1. **DDL Operations with Embedded ANALYZE**
    - Execute ANALYZE as part of the DDL job for both:
      - `CREATE INDEX` operations
      - `ALTER TABLE MODIFY COLUMN` operations
    - Support for analyzing tables and indexes in "reorg" state
    - Handle both single and multi-column modifications

2. **State Management**
    - Add new job states for analyze phase
    - Track analyze progress within the DDL job
    - Handle job retries and failures gracefully

3. **Configuration**
    - New session variable: `tidb_enable_ddl_analyze` (default: true)
    - New execution option: `AnalyzeReorgIndexes`
    - Support for `ANALYZE` options (e.g., sample rate)

### 2.2 Non-Functional Requirements

1. **Performance**
    - Must not significantly impact existing index creation performance
    - Reduce total execution time compared to separate CREATE INDEX and ANALYZE

2. **Reliability**
    - Handle large tables without OOM
    - Proper cleanup on failure
    - Transaction safety

3. **Observability**
    - Detailed logging
    - Metrics for analyze phase
    - Progress reporting

## 3. System Design

### 3.1 Architecture

```
+------------------+     +------------------+     +------------------+
|   DDL Worker     | --> |  Index Creation  | --> |  Analyze Phase   |
+------------------+     +------------------+     +------------------+
         |                      |                          |
         v                      v                          v
+------------------+     +------------------+     +------------------+
|  Job Scheduler   |     |  Storage Layer   |     |  Statistics API  |
+------------------+     +------------------+     +------------------+
```

### 3.2 Data Flow
1. User call `set @@tidb_enable_ddl_analyze = 1` to enable this feature in their session
2. User executes either:
   - `ALTER TABLE {table name} ADD INDEX...`
   - `ALTER TABLE {table name} MODIFY COLUMN...`
3. DDL job is created with new analyze state
4. Index creation proceeds as usual
5. After DDL operation completes (index backfill or column modification), analyze phase begins
6. Statistics are collected and stored
7. Job completes

## 4. Implementation Details

### 4.1 New Components

1. **DDL Event Tracker**
    - Track schema changes from MODIFY COLUMN operations
    - Identify tables requiring statistics updates
    - Maintain mapping between DDL operations and affected statistics

1. **Analyze State Manager**
    - Track analyze progress
    - Handle state transitions
    - Manage retries

2. **Statistics Collector**
    - Optimized for both table and index analysis
    - Support for partial analysis
    - Progress reporting
    - Special handling for modified columns to ensure accurate statistics

### 4.2 Modified Components

1. **DDL Worker**
    - Handle new job states
    - Coordinate DDL operations (index creation, column modification) and analysis
    - Error handling
    - Track modified columns for accurate statistics collection

2. **Session Manager**
    - New session variables
    - Execution context

## 5. API Changes

### 5.1 New Session Variables

```sql
-- Enable/disable embedded analyze
SET tidb_enable_ddl_analyze = {0|1};

-- Control analyze behavior
SET tidb_analyze_version = {1|2};
```
