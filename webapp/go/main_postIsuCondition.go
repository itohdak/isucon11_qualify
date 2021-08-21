package main

import (
	// "math/rand"
	"net/http"
	"sync"
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

type counter struct {
	mu         sync.Mutex // 追加
	conditions []IsuConditionBulk
}

func (m *counter) Set(jiaIsuUUID string, req []PostIsuConditionRequest, c echo.Context) int {
	m.mu.Lock() // 追加
	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return 2
	}
	defer tx.Rollback()

	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?", jiaIsuUUID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return 2
	}
	if count == 0 {
		return 3
	}
	for _, cond := range req {
		timestamp := time.Unix(cond.Timestamp, 0)

		if !isValidConditionFormat(cond.Condition) {
			return 3
		}

		m.conditions = append(m.conditions, IsuConditionBulk{
			JIAIsuUUID: jiaIsuUUID,
			Timestamp:  timestamp,
			IsSitting:  cond.IsSitting,
			Condition:  cond.Condition,
			Message:    cond.Message,
		})
	}

	if len(req) >= 100 {
		_, err_ := tx.NamedExec("INSERT INTO `isu_condition`"+
			"	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)"+
			"	VALUES (:jia_isu_uuid, :timestamp, :is_sitting, :condition, :message)", m.conditions)
		if err_ != nil {
			c.Logger().Errorf("db error: %v", err_)
			return 2
		}
		m.conditions = []IsuConditionBulk{}
	}
	err = tx.Commit()

	m.mu.Unlock()
	return 1
}

func postIsuCondition(c echo.Context) error {
	// TODO: 一定割合リクエストを落としてしのぐようにしたが、本来は全量さばけるようにすべき

	req := []PostIsuConditionRequest{}
	err := c.Bind(&req)

	// if len(req) < 15 {
	// 	c.Logger().Warnf("under 5 data")
	// 	return c.NoContent(http.StatusAccepted)
	// }

	// dropProbability := 0.5
	// if rand.Float64() <= dropProbability {
	// 	c.Logger().Warnf("drop post isu condition request")
	// 	return c.NoContent(http.StatusAccepted)
	// }

	jiaIsuUUID := c.Param("jia_isu_uuid")
	if jiaIsuUUID == "" {
		return c.String(http.StatusBadRequest, "missing: jia_isu_uuid")
	}

	if err != nil {
		return c.String(http.StatusBadRequest, "bad request body")
	} else if len(req) == 0 {
		return c.String(http.StatusBadRequest, "bad request body")
	}

	// tx, err := db.Beginx()
	// if err != nil {
	// 	c.Logger().Errorf("db error: %v", err)
	// 	return c.NoContent(http.StatusInternalServerError)
	// }
	// defer tx.Rollback()

	// var count int
	// err = tx.Get(&count, "SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?", jiaIsuUUID)
	// if err != nil {
	// 	c.Logger().Errorf("db error: %v", err)
	// 	return c.NoContent(http.StatusInternalServerError)
	// }
	// if count == 0 {
	// 	return c.String(http.StatusNotFound, "not found: isu")
	// }

	// conditions := []IsuConditionBulk{}

	isSuccess := m.Set(jiaIsuUUID, req, c)
	if isSuccess == 2 {
		return c.NoContent(http.StatusInternalServerError)
	} else if isSuccess == 3 {
		return c.String(http.StatusNotFound, "not found: isu")
	}

	// for _, cond := range req {
	// 	timestamp := time.Unix(cond.Timestamp, 0)

	// 	if !isValidConditionFormat(cond.Condition) {
	// 		return c.String(http.StatusBadRequest, "bad request body")
	// 	}

	// 	conditions = append(conditions, IsuConditionBulk{
	// 		JIAIsuUUID: jiaIsuUUID,
	// 		Timestamp:  timestamp,
	// 		IsSitting:  cond.IsSitting,
	// 		Condition:  cond.Condition,
	// 		Message:    cond.Message,
	// 	})
	// }
	// _, err_ := tx.NamedExec("INSERT INTO `isu_condition`"+
	// 	"	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)"+
	// 	"	VALUES (:jia_isu_uuid, :timestamp, :is_sitting, :condition, :message)", conditions)
	// if err_ != nil {
	// 	c.Logger().Errorf("db error: %v", err_)
	// 	return c.NoContent(http.StatusInternalServerError)
	// }

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

	// err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusAccepted)
}

// ISUのコンディションの文字列がcsv形式になっているか検証
