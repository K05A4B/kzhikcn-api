package data

import (
	"time"
)

var (
	schemaVersion = 1
)

type SchemaState struct {
	Version       int `gorm:"primaryKey;autoIncrement" json:"version"`
	LastMigration time.Time
}

func GetSchemaState() (*SchemaState, error) {
	state := &SchemaState{}
	err := db.Model(SchemaState{}).First(state).Error
	if err != nil {
		return nil, err
	}
	return state, nil
}

func ExistSchemaState() bool {
	return db.Migrator().HasTable(&SchemaState{})
}
