package main

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func getIsuList(c echo.Context) error {
	jiaUserID, errStatusCode, err := getUserIDFromSession(c)
	if err != nil {
		if errStatusCode == http.StatusUnauthorized {
			return c.String(http.StatusUnauthorized, "you are not signed in")
		}

		c.Logger().Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	tx, err := db.Beginx()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	isuList := []Isu{}
	err = tx.Select(
		&isuList,
		"SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC",
		jiaUserID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	responseList := []GetIsuListResponse{}
	jia_isu_uuids := []string{}
	for _, isu := range isuList {
		jia_isu_uuids = append(jia_isu_uuids, isu.JIAIsuUUID)
	}
	inQuery, inArgs, err := sqlx.In(
		`
		SELECT isu_condition.*
		FROM
		isu_condition
		INNER JOIN
		(
			SELECT jia_isu_uuid, MAX(timestamp) as timestamp
			FROM isu_condition
			GROUP BY jia_isu_uuid
		) t
		ON isu_condition.jia_isu_uuid = t.jia_isu_uuid AND isu_condition.timestamp = t.timestamp
		WHERE isu_condition.jia_isu_uuid IN (?)
		`, jia_isu_uuids)
	if err != nil {
		log.Print(err)
	}
	lastConditionList := []IsuCondition{}
	if err = tx.Select(&lastConditionList, inQuery, inArgs...); err != nil {
		// c.Logger().Errorf("db error: %v", err)
		// return c.NoContent(http.StatusInternalServerError)
	}
	lastConditionMap := map[string]IsuCondition{}
	for _, lastCondition := range lastConditionList {
		lastConditionMap[lastCondition.JIAIsuUUID] = lastCondition
	}
	for _, isu := range isuList {
		var lastCondition IsuCondition
		if val, ok := lastConditionMap[isu.JIAIsuUUID]; ok {
			lastCondition = val
		} else {
			continue
		}
		var formattedCondition *GetIsuConditionResponse
		conditionLevel, err := calculateConditionLevel(lastCondition.Condition)
		if err != nil {
			c.Logger().Error(err)
			return c.NoContent(http.StatusInternalServerError)
		}

		formattedCondition = &GetIsuConditionResponse{
			JIAIsuUUID:     lastCondition.JIAIsuUUID,
			IsuName:        isu.Name,
			Timestamp:      lastCondition.Timestamp.Unix(),
			IsSitting:      lastCondition.IsSitting,
			Condition:      lastCondition.Condition,
			ConditionLevel: conditionLevel,
			Message:        lastCondition.Message,
		}

		res := GetIsuListResponse{
			ID:                 isu.ID,
			JIAIsuUUID:         isu.JIAIsuUUID,
			Name:               isu.Name,
			Character:          isu.Character,
			LatestIsuCondition: formattedCondition}
		responseList = append(responseList, res)
	}

	err = tx.Commit()
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, responseList)
}

// POST /api/isu
// ISUを登録
