package provider

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// NullInt64ToIntPtr converts sql.NullInt64 to *int
func NullInt64ToIntPtr(ni sql.NullInt64) *int {
	if ni.Valid {
		val := int(ni.Int64)
		return &val
	}
	return nil
}

// NullTimeToTimePtr converts sql.NullTime to *time.Time
func NullTimeToTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// NullUUIDToUUID converts uuid.NullUUID to uuid.UUID
func NullUUIDToUUID(nu uuid.NullUUID) uuid.UUID {
	if nu.Valid {
		return nu.UUID
	}
	return uuid.Nil
}

// NullFloat64ToFloat64 converts sql.NullFloat64 to float64 (or 0.0 if null)
func NullFloat64ToFloat64(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0 // Or return a default value like 0.0, or use a pointer for nullability
}

// NullFloat64ToFloat64Ptr converts sql.NullFloat64 to *float64
func NullFloat64ToFloat64Ptr(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

// NullInt64ToInt converts sql.NullInt64 to int (or 0 if null)
func NullInt64ToInt(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0 // Or return a default value like 0, or use a pointer for nullability
}

// IntPtrToInt converts an *int to int, returning a defaultValue if the pointer is nil.
func IntPtrToInt(p *int, defaultValue int) int {
	if p != nil {
		return *p
	}
	return defaultValue
}

// Float64PtrToFloat64 converts an *float64 to float64, returning a defaultValue if the pointer is nil.
func Float64PtrToFloat64(p *float64, defaultValue float64) float64 {
	if p != nil {
		return *p
	}
	return defaultValue
}

