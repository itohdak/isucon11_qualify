package main

import (
	"time"
)

type GetIsuLastConditionResponse struct {
	ID         int    `db:"id" json:"id"`
	JIAIsuUUID string `db:"jia_isu_uuid" json:"jia_isu_uuid"`
	Name       string `db:"name" json:"name"`
	Character  string `db:"character" json:"character"`
	// JIAIsuUUID string    `db:"jia_isu_uuid"`
	Timestamp time.Time `db:"timestamp"`
	IsSitting bool      `db:"is_sitting"`
	Condition string    `db:"condition"`
	Message   string    `db:"message"`
}
