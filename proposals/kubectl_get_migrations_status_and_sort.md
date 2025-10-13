# kubectl get migrations status filtering and sorting

## TL;DR (solution in one paragraph)

Implement the `--status` flag for the `kubectl schemahero get migrations` command to filter migrations by their lifecycle state (planned, executed, approved, rejected), and add automatic sorting by age (newest first) to all output. This enhances usability by allowing users to quickly identify migrations in specific states and see the most recent activity first, improving operational workflows for managing schema migrations.

## The problem

Users cannot filter migrations by status, making it difficult to identify which migrations need attention (planned but not executed), which have been completed (approved/executed), or which were rejected. Additionally, migrations are displayed in arbitrary order, requiring users to manually scan timestamps to find recent changes. The examples.md file shows clear user expectations for these features, with the `--status` flag already defined but commented out in the code.

## Prototype / design

The solution adds minimal code changes to implement status filtering and sorting:

```
Input: kubectl schemahero get migrations --status=planned
Processing:
1. Fetch all migrations (existing logic)
2. Filter by status using matchesMigrationStatus()
3. Sort by most recent timestamp (newest first)
4. Display in table format (existing logic)

Status hierarchy (mutually exclusive):
- rejected: RejectedAt > 0 (highest precedence)
- approved: ApprovedAt > 0 && RejectedAt == 0
- executed: ExecutedAt > 0 && ApprovedAt == 0 && RejectedAt == 0
- planned: PlannedAt > 0 && ExecutedAt == 0 && ApprovedAt == 0 && RejectedAt == 0
```

## New Subagents / Commands

No new subagents or commands will be created. This is an enhancement to the existing `get migrations` command.

## Database

**No database changes required.** This is a read-only CLI enhancement that filters and sorts existing data.

## Implementation plan

### Files to modify
1. `/Users/joshs/Code/schemahero/schemahero/pkg/cli/schemaherokubectlcli/get_migrations.go`

### Changes to get_migrations.go

**1. Uncomment status flag (line 125)**
```go
cmd.Flags().StringP("status", "s", "", "status to filter to results to (planned, executed, approved, rejected)")
```

**2. Add status filter variable after database filter (line 32)**
```go
databaseNameFilter := v.GetString("database")
statusFilter := v.GetString("status")
```

**3. Add status validation after statusFilter declaration**
```go
if statusFilter != "" {
    validStatuses := map[string]bool{"planned": true, "executed": true, "approved": true, "rejected": true}
    if !validStatuses[statusFilter] {
        return fmt.Errorf("invalid status: %s. Valid values are: planned, executed, approved, rejected", statusFilter)
    }
}
```

**4. Add matchesMigrationStatus function before GetMigrationsCmd**
```go
func matchesMigrationStatus(m schemasv1alpha4.Migration, status string) bool {
    s := m.Status

    switch status {
    case "rejected":
        return s.RejectedAt > 0
    case "approved":
        return s.ApprovedAt > 0 && s.RejectedAt == 0
    case "executed":
        return s.ExecutedAt > 0 && s.ApprovedAt == 0 && s.RejectedAt == 0
    case "planned":
        return s.PlannedAt > 0 && s.ExecutedAt == 0 && s.ApprovedAt == 0 && s.RejectedAt == 0
    default:
        return false
    }
}
```

**5. Add getMostRecentTimestamp function**
```go
func getMostRecentTimestamp(m schemasv1alpha4.Migration) int64 {
    timestamps := []int64{
        m.Status.PlannedAt,
        m.Status.ExecutedAt,
        m.Status.ApprovedAt,
        m.Status.RejectedAt,
    }

    var mostRecent int64
    for _, ts := range timestamps {
        if ts > mostRecent {
            mostRecent = ts
        }
    }
    return mostRecent
}
```

**6. Filter by status in the migration collection loop (after line 81)**
```go
if databaseNameFilter == "" && statusFilter == "" {
    matchingMigrations = append(matchingMigrations, m)
    continue
}

if databaseNameFilter != "" && m.Spec.DatabaseName != databaseNameFilter {
    continue
}

if statusFilter != "" && !matchesMigrationStatus(m, statusFilter) {
    continue
}

matchingMigrations = append(matchingMigrations, m)
```

**7. Add sorting after collection (before line 87)**
```go
// Sort by most recent timestamp, newest first
sort.Slice(matchingMigrations, func(i, j int) bool {
    return getMostRecentTimestamp(matchingMigrations[i]) > getMostRecentTimestamp(matchingMigrations[j])
})
```

**8. Add import for sort package**
```go
import (
    // existing imports...
    "sort"
)
```

### External contracts
- No API changes
- No new events emitted
- Command-line interface enhanced with optional `--status` flag

### Toggle strategy
No feature flags needed - the enhancement is backwards compatible:
- Without `--status` flag: existing behavior (all migrations, now sorted)
- With `--status` flag: filtered results (also sorted)

## Testing

### Unit tests
Create `/Users/joshs/Code/schemahero/schemahero/pkg/cli/schemaherokubectlcli/get_migrations_test.go`:

```go
func TestMatchesMigrationStatus(t *testing.T) {
    tests := []struct {
        name     string
        status   MigrationStatus
        filter   string
        expected bool
    }{
        // Test rejected takes precedence
        {
            name: "rejected with all timestamps",
            status: MigrationStatus{
                PlannedAt: 100, ExecutedAt: 200, ApprovedAt: 300, RejectedAt: 400,
            },
            filter: "rejected",
            expected: true,
        },
        {
            name: "rejected excludes from approved",
            status: MigrationStatus{
                PlannedAt: 100, ExecutedAt: 200, ApprovedAt: 300, RejectedAt: 400,
            },
            filter: "approved",
            expected: false,
        },
        // Test approved
        {
            name: "approved without rejected",
            status: MigrationStatus{
                PlannedAt: 100, ExecutedAt: 200, ApprovedAt: 300,
            },
            filter: "approved",
            expected: true,
        },
        // Test executed
        {
            name: "executed only",
            status: MigrationStatus{
                PlannedAt: 100, ExecutedAt: 200,
            },
            filter: "executed",
            expected: true,
        },
        // Test planned
        {
            name: "planned only",
            status: MigrationStatus{
                PlannedAt: 100,
            },
            filter: "planned",
            expected: true,
        },
    }
}

func TestGetMostRecentTimestamp(t *testing.T) {
    // Test finding most recent among multiple timestamps
}

func TestSortingByAge(t *testing.T) {
    // Test that migrations are sorted newest first
}
```

### Manual testing checklist
- [ ] Run without --status flag, verify all migrations shown and sorted
- [ ] Run with --status=planned, verify only planned migrations shown
- [ ] Run with --status=executed, verify only executed migrations shown
- [ ] Run with --status=approved, verify only approved migrations shown
- [ ] Run with --status=rejected, verify only rejected migrations shown
- [ ] Test migration fe75c65 appears only in rejected status
- [ ] Run with --status=invalid, verify error message
- [ ] Verify sorting (newest first) in all cases

## Backward compatibility

- Command remains fully backwards compatible
- Existing scripts using `kubectl schemahero get migrations` continue to work
- Output format unchanged (same columns, same formatting)
- Only behavior change: results are now sorted (improvement, non-breaking)

## Migrations

No special deployment handling required. This is a client-side CLI change that works with existing API versions.

## Trade-offs

Optimizing for simplicity and clarity:
- **Chosen**: Simple mutually exclusive status hierarchy (rejected > approved > executed > planned)
- **Benefit**: Clear, predictable behavior matching user mental model
- **Cost**: Cannot query for "all executed regardless of approval/rejection" in single command

Optimizing for performance:
- **Chosen**: Client-side filtering after fetching all migrations
- **Benefit**: Simple implementation, works with existing API
- **Cost**: Fetches all migrations even when filtering (acceptable for typical cluster sizes)

## Alternative solutions considered

1. **Server-side filtering**: Add status filtering to the Kubernetes API List operation
   - Rejected: Requires API changes, more complex deployment

2. **Multiple status flags**: Allow `--status=executed,approved` for OR queries
   - Rejected: Adds complexity, current requirements show single status is sufficient

3. **Separate sort flag**: Add `--sort=age` flag instead of always sorting
   - Rejected: Sorting by age is always beneficial, no need for configurability

4. **Complex status logic**: Allow executed+approved as valid combination
   - Rejected: Requirements explicitly state statuses are mutually exclusive

## Research

Reference: `/Users/joshs/Code/schemahero/schemahero/proposals/kubectl_get_migrations_enhancements_research.md`

### Prior art in codebase
- Similar filtering pattern in database name filtering (lines 76-83)
- Age formatting function already exists (timestampToAge, lines 130-147)
- Standard cobra/viper patterns for flag handling

### External references
- Kubernetes kubectl uses similar status filtering for pods, deployments
- Standard Go sort.Slice pattern for custom sorting

## Checkpoints (PR plan)

Single PR containing:
1. Uncommented status flag
2. Status filtering logic with matchesMigrationStatus function
3. Sorting implementation with getMostRecentTimestamp function
4. Unit tests for new functions
5. Updated command help text if needed

This is a focused change that can be reviewed and merged as a single unit.