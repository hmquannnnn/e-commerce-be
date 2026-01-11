package utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNoFieldsToUpdate = errors.New("no fields to update")
	ErrNoRowsAffected   = errors.New("no rows affected")
)

// PatchUpdate represents a partial update operation
type PatchUpdate struct {
	Table     string                 // Table name
	Updates   map[string]interface{} // Field name -> value to update
	Where     string                 // WHERE clause (e.g., "id = $1 AND deleted_at IS NULL")
	WhereArgs []interface{}          // Arguments for WHERE clause
}

// ExecPatch executes a dynamic UPDATE query based on provided fields
// It automatically adds updated_at field with current timestamp
// Returns number of rows affected and error
func ExecPatch(ctx context.Context, db *sql.DB, patch PatchUpdate) (int64, error) {
	if len(patch.Updates) == 0 {
		return 0, ErrNoFieldsToUpdate
	}

	if patch.Table == "" {
		return 0, errors.New("table name is required")
	}

	if patch.Where == "" {
		return 0, errors.New("where clause is required")
	}

	// Build SET clause
	setClauses := make([]string, 0, len(patch.Updates)+1)
	args := make([]interface{}, 0, len(patch.Updates)+1+len(patch.WhereArgs))
	paramIndex := 1

	// Add user-provided fields
	for field, value := range patch.Updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, paramIndex))
		args = append(args, value)
		paramIndex++
	}

	// Always add updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", paramIndex))
	args = append(args, time.Now())
	paramIndex++

	// Build WHERE clause with adjusted parameter indices
	whereClause := patch.Where
	for i := range patch.WhereArgs {
		placeholder := fmt.Sprintf("$%d", i+1)
		newPlaceholder := fmt.Sprintf("$%d", paramIndex)
		whereClause = strings.Replace(whereClause, placeholder, newPlaceholder, 1)
		paramIndex++
	}
	args = append(args, patch.WhereArgs...)

	// Build final query
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		patch.Table,
		strings.Join(setClauses, ", "),
		whereClause,
	)

	// Execute query
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute patch update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// BuildPatchUpdates is a helper function to build Updates map from struct fields
// It only includes non-nil pointer fields
// Example:
//
//	updates := BuildPatchUpdates(map[string]*interface{}{
//	    "name": ptrTo(user.Name),
//	    "phone": ptrTo(user.Phone),
//	})
func BuildPatchUpdates(fields map[string]interface{}) map[string]interface{} {
	updates := make(map[string]interface{})
	for field, value := range fields {
		if value != nil {
			updates[field] = value
		}
	}
	return updates
}
