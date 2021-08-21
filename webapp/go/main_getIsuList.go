package main

import (
	"net/http"

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

	// isuList := []Isu{}
	// err = tx.Select(
	// 	&isuList,
	// 	"SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC",
	// 	jiaUserID)
	// if err != nil {
	// 	c.Logger().Errorf("db error: %v", err)
	// 	return c.NoContent(http.StatusInternalServerError)
	// }

	// responseList := []GetIsuListResponse{}
	// for _, isu := range isuList {
	// 	var lastCondition IsuCondition
	// 	foundLastCondition := true
	// 	err = tx.Get(&lastCondition, "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` DESC LIMIT 1",
	// 		isu.JIAIsuUUID)
	// 	if err != nil {
	// 		if errors.Is(err, sql.ErrNoRows) {
	// 			foundLastCondition = false
	// 		} else {
	// 			c.Logger().Errorf("db error: %v", err)
	// 			return c.NoContent(http.StatusInternalServerError)
	// 		}
	// 	}

	// 	var formattedCondition *GetIsuConditionResponse
	// 	if foundLastCondition {
	// 		conditionLevel, err := calculateConditionLevel(lastCondition.Condition)
	// 		if err != nil {
	// 			c.Logger().Error(err)
	// 			return c.NoContent(http.StatusInternalServerError)
	// 		}

	// 		formattedCondition = &GetIsuConditionResponse{
	// 			JIAIsuUUID:     lastCondition.JIAIsuUUID,
	// 			IsuName:        isu.Name,
	// 			Timestamp:      lastCondition.Timestamp.Unix(),
	// 			IsSitting:      lastCondition.IsSitting,
	// 			Condition:      lastCondition.Condition,
	// 			ConditionLevel: conditionLevel,
	// 			Message:        lastCondition.Message,
	// 		}
	// 	}

	// 	res := GetIsuListResponse{
	// 		ID:                 isu.ID,
	// 		JIAIsuUUID:         isu.JIAIsuUUID,
	// 		Name:               isu.Name,
	// 		Character:          isu.Character,
	// 		LatestIsuCondition: formattedCondition}
	// 	responseList = append(responseList, res)
	// }

	responseList := []GetIsuListResponse{}
	isuLastConditionList := []GetIsuLastConditionResponse{}
	err = tx.Select(
		&isuLastConditionList,
		`
SELECT
t1.id,
t1.jia_isu_uuid,
t1.name,
t1.character,
t4.timestamp,
t4.is_sitting,
t4.condition,
t4.message
FROM
isu t1
INNER JOIN
(
SELECT
t2.jia_isu_uuid,
t2.timestamp,
t2.is_sitting,
t2.condition,
t2.message
FROM isu_condition t2
INNER JOIN
(
SELECT jia_isu_uuid, MAX(timestamp) as timestamp
FROM isu_condition
GROUP BY jia_isu_uuid
) t3
ON (t2.timestamp = t3.timestamp AND t2.jia_isu_uuid = t3.jia_isu_uuid)
) t4
ON t1.jia_isu_uuid = t4.jia_isu_uuid AND t1.jia_user_id = ?
ORDER BY t4.timestamp DESC;
		`,
		jiaUserID)
	if err != nil {
		c.Logger().Errorf("db error: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	for _, lastCondition := range isuLastConditionList {
		conditionLevel, err := calculateConditionLevel(lastCondition.Condition)
		if err != nil {
			c.Logger().Error(err)
			return c.NoContent(http.StatusInternalServerError)
		}
		var formattedCondition = &GetIsuConditionResponse{
			JIAIsuUUID:     lastCondition.JIAIsuUUID,
			IsuName:        lastCondition.Name,
			Timestamp:      lastCondition.Timestamp.Unix(),
			IsSitting:      lastCondition.IsSitting,
			Condition:      lastCondition.Condition,
			ConditionLevel: conditionLevel,
			Message:        lastCondition.Message,
		}
		res := GetIsuListResponse{
			ID:                 lastCondition.ID,
			JIAIsuUUID:         lastCondition.JIAIsuUUID,
			Name:               lastCondition.Name,
			Character:          lastCondition.Character,
			LatestIsuCondition: formattedCondition,
		}
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
