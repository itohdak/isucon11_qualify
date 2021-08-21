package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func getIsuIcon(c echo.Context) error {
	jiaUserID, errStatusCode, err := getUserIDFromSession(c)
	if err != nil {
		if errStatusCode == http.StatusUnauthorized {
			return c.String(http.StatusUnauthorized, "you are not signed in")
		}

		c.Logger().Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	jiaIsuUUID := c.Param("jia_isu_uuid")
	var file *os.File
	if file, err = os.Open("/home/isucon/webapp/images/" + jiaUserID + "_" + jiaIsuUUID); err != nil {
		return c.String(http.StatusNotFound, "not found: isu")
	}
	defer file.Close()

	var image []byte
	if image, err = ioutil.ReadAll(file); err != nil {
		return c.String(http.StatusNotFound, "not found: isu")
	}

	// var image []byte
	// err = db.Get(&image, "SELECT `image` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
	// 	jiaUserID, jiaIsuUUID)
	// if err != nil {
	// 	if errors.Is(err, sql.ErrNoRows) {
	// 		return c.String(http.StatusNotFound, "not found: isu")
	// 	}

	// 	c.Logger().Errorf("db error: %v", err)
	// 	return c.NoContent(http.StatusInternalServerError)
	// }

	return c.Blob(http.StatusOK, "", image)
}

// GET /api/isu/:jia_isu_uuid/graph
// ISUのコンディショングラフ描画のための情報を取得

type IsuImageWithUUID struct {
	JIAIsuUUID string `db:"jia_isu_uuid"`
	Image      []byte `db:"image" json:"-"`
	JIAUserID  string `db:"jia_user_uuid"`
}

func saveIconsToLocal(c echo.Context) error {
	images := []IsuImageWithUUID{}
	err := db.Select(&images, "SELECT `image`, `jia_isu_uuid`, 'jia_user_uuid` FROM `isu`")
	if err != nil {
		fmt.Println(err)
		return c.Blob(http.StatusOK, "", nil)
	}
	var cnt = 0
	for _, image := range images {
		cnt += 1
		var file *os.File
		if file, err = os.OpenFile("/home/isucon/webapp/images/"+image.JIAUserID+"_"+image.JIAIsuUUID, os.O_RDWR|os.O_CREATE, 0755); err != nil {
			fmt.Println(err)
			continue
		}
		if _, err := file.Write(image.Image); err != nil {
			fmt.Println(err)
			continue
		}
		file.Close()
	}
	fmt.Println(cnt)
	return nil
}
