package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(EveryLogicalType{}) }

// EveryLogicalType is the fixture that spells all eighteen logical types once each.
type EveryLogicalType struct{}

// Up creates a table carrying one column of every type the IR defines.
func (m EveryLogicalType) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.CreateTable("specimens").
		Column("c_string", migrasyon.String(255)).
		Column("c_text", migrasyon.Text()).
		Column("c_bool", migrasyon.Bool()).
		Column("c_int16", migrasyon.Int16()).
		Column("c_int32", migrasyon.Int32()).
		Column("c_int64", migrasyon.Int64()).
		Column("c_float32", migrasyon.Float32()).
		Column("c_float64", migrasyon.Float64()).
		Column("c_decimal", migrasyon.Decimal(10, 2)).
		Column("c_uuid", migrasyon.UUID()).
		Column("c_json", migrasyon.JSON()).
		Column("c_timestamp", migrasyon.Timestamp()).
		Column("c_timestamptz", migrasyon.TimestampTZ()).
		Column("c_date", migrasyon.Date()).
		Column("c_time", migrasyon.Time()).
		Column("c_binary", migrasyon.Binary()).
		Column("c_enum", migrasyon.Enum("draft", "live")).
		Column("c_array", migrasyon.Array(migrasyon.String(64))).
		PrimaryKey("c_uuid").
		Apply(ctx)
}

// Down drops the table, which needs no record of what it held to be a valid operation.
func (m EveryLogicalType) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("specimens").Drop().Apply(ctx)
}
