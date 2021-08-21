package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type IsuConditionBulk struct {
	JIAIsuUUID string    `db:"jia_isu_uuid"`
	Timestamp  time.Time `db:"timestamp"`
	IsSitting  bool      `db:"is_sitting"`
	Condition  string    `db:"condition"`
	Message    string    `db:"message"`
}

func postIsuCondition(c echo.Context) error {
	// TODO: 一定割合リクエストを落としてしのぐようにしたが、本来は全量さばけるようにすべき
	dropProbability := 0.9
	if rand.Float64() <= dropProbability {
		c.Logger().Warnf("drop post isu condition request")
		return c.NoContent(http.StatusAccepted)
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")
	if jiaIsuUUID == "" {
		return c.String(http.StatusBadRequest, "missing: jia_isu_uuid")
	}

	req := []PostIsuConditionRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request body")
	} else if len(req) == 0 {
		return c.String(http.StatusBadRequest, "bad request body")
	}

	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?", jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if count == 0 {
		return c.String(http.StatusNotFound, "not found: isu")
	}

	conditions := []IsuConditionBulk{}

	for _, cond := range req {
		timestamp := time.Unix(cond.Timestamp, 0)

		if !isValidConditionFormat(cond.Condition) {
			return c.String(http.StatusBadRequest, "bad request body")
		}

		conditions = append(conditions, IsuConditionBulk{
			JIAIsuUUID: jiaIsuUUID,
			Timestamp:  timestamp,
			IsSitting:  cond.IsSitting,
			Condition:  cond.Condition,
			Message:    cond.Message,
		})
	}
	_, err_ := tx.NamedExec("INSERT INTO `isu_condition`"+
		"	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)"+
		"	VALUES (:jia_isu_uuid, :timestamp, :is_sitting, :condition, :message)", conditions)
	if err_ != nil {
		c.Logger().Errorf("db error: %v", err_)
		return c.NoContent(http.StatusInternalServerError)
	}

	// for _, cond := range req {
	// 	timestamp := time.Unix(cond.Timestamp, 0)

	// 	if !isValidConditionFormat(cond.Condition) {
	// 		return c.String(http.StatusBadRequest, "bad request body")
	// 	}

	// 	_, err = tx.Exec(
	// 		"INSERT INTO `isu_condition`"+
	// 			"	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)"+
	// 			"	VALUES (?, ?, ?, ?, ?)",
	// 		jiaIsuUUID, timestamp, cond.IsSitting, cond.Condition, cond.Message)
	// 	if err != nil {
	// 		c.Logger().Errorf("db error: %v", err)
	// 		return c.NoContent(http.StatusInternalServerError)
	// 	}

	// }

	err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusAccepted)
}

// ISUのコンディションの文字列がcsv形式になっているか検証
