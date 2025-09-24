# kubectl get migrations enhancements research

## Current State Analysis

### File Location
`/Users/joshs/Code/schemahero/schemahero/pkg/cli/schemaherokubectlcli/get_migrations.go`

### Current Implementation
- Lists all migrations across namespaces with database filtering
- Displays columns: ID, DATABASE, TABLE, PLANNED, EXECUTED, APPROVED, REJECTED
- Status flag is defined but commented out at line 125
- No sorting - migrations are displayed in the order returned by the API
- No status filtering - all migrations are shown regardless of status

### Data Structures
Migration status has 5 timestamp fields:
- `PlannedAt` - when the plan was generated
- `InvalidatedAt` - when plan was determined invalid/outdated
- `ApprovedAt` - when migration was approved
- `RejectedAt` - when migration was rejected
- `ExecutedAt` - when migration was executed

### Current Output Format
Uses tabwriter for formatted table output with age calculations (e.g., "449d19h" format)

## Requirements from examples.md

### Status Filtering
Mutually exclusive status categories:
1. **planned**: PlannedAt > 0 AND no other timestamps
2. **executed**: ExecutedAt > 0 AND NOT approved AND NOT rejected
3. **approved**: ApprovedAt > 0 AND NOT rejected
4. **rejected**: RejectedAt > 0 (final state, takes precedence)

### Special Case
Migration fe75c65 has all 4 timestamps and should ONLY appear in "rejected" status due to hierarchy.

### Sorting
All results should be sorted by age (newest first) based on most recent timestamp.

## Code Patterns

### Command Structure
- Uses cobra for command definition
- Uses viper for flag binding
- Pre-existing patterns for namespace and database filtering

### Testing
- No tests exist for the CLI package currently
- Standard Go test patterns used elsewhere in codebase
- Tests typically use table-driven approach

## Dependencies
- `github.com/spf13/cobra` - command framework
- `github.com/spf13/viper` - configuration
- `k8s.io/client-go` - Kubernetes client
- Standard library `sort` package available for sorting

## Implementation Considerations

### Status Filter Logic
Need clear hierarchy:
1. Check RejectedAt first (overrides all)
2. Then ApprovedAt (overrides executed)
3. Then ExecutedAt
4. Finally PlannedAt (only if nothing else set)

### Sorting Strategy
- Need to determine "most recent" timestamp for each migration
- Use that for sorting newest first
- Can use standard library sort.Slice()

### Backwards Compatibility
- Flag is already defined, just commented out
- Default behavior (no flag) should remain unchanged
- Adding sorting won't break existing scripts as output format stays same