package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/gabriel-vasile/mimetype"
	"noz.zkip.cc/utils"
)

func getQuota(targetPRI string) (*Quota, error) {
	db := utils.GetMySqlDB()
	var q = &Quota{}

	row := db.QueryRow("select capcity, used from tQuotas where targetPRI = ?", targetPRI)
	err := row.Scan(&q.Capcity, &q.Used)
	if err != nil {
		return nil, err
	}

	return q, nil
}

func getSupportedMimeType(mtype *mimetype.MIME) bool {
	ok, _ := regexp.MatchString(`^image`, mtype.String())
	return ok
}

func setupStoreEnv() {
	err := os.MkdirAll("data/images", os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("data/papers", os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func deleteResourceFile(rID string, rType uint8) error {
	storepath, ok := resource_store_path[rType]
	if !ok {
		return &NoStoreResourceTypeErr{rType: rType}
	}

	filepath := fmt.Sprintf("%s/%s", storepath, rID)

	err := os.Remove(filepath)
	if os.IsNotExist(err) {
		return &utils.NotFoundErr{Name: filepath}
	}
	if err != nil {
		return err
	}

	return nil
}

func removeResource(takerPRI, rID string, rType uint8) error {
	db := utils.GetMySqlDB()
	resourcePRI := utils.GenPRI(rID, rType)

	var err error

	tx, err := db.Begin()
	utils.PanicIfErr(err)

	rollback := func(err error) {
		if tx != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
		}
	}

	var quotaUsage uint64
	row := tx.QueryRow("select quotaUsage from tResources where rID = ?", rID)
	err = row.Scan(&quotaUsage)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	quota, err := getQuota(takerPRI)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	needntUpdate := quota.Used == 0

	err = quota.doDelta(-int64(quotaUsage))
	if utils.RunIfErr(err, rollback) {
		return err
	}

	if !needntUpdate {
		_, err := tx.Exec("update tQuotas set used = ? where targetPRI = ?", quota.Used, takerPRI)
		if utils.RunIfErr(err, rollback) {
			return err
		}
	}

	_, err = tx.Exec("delete from tResources where rID = ?", rID)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	_, err = tx.Exec("delete from tPermissions where resourcePRI = ?", resourcePRI)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	err = deleteResourceFile(rID, rType)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	return tx.Commit()
}

func saveImage(content []byte, name string) {
	file, err := os.Create("data/images/" + name)
	utils.PanicIfErr(err)

	defer file.Close()

	_, err = io.Copy(file, bytes.NewBuffer(content))
	utils.PanicIfErr(err)
}

func savePaper(content []byte, rID string, isNewOne bool) error {
	var file *os.File
	var err error
	var filename = fmt.Sprintf("data/papers/%s", rID)
	if isNewOne {
		file, err = os.Create(filename)
		if err != nil {
			return err
		}
	} else {
		file, err = os.OpenFile(filename, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
	}

	defer file.Close()

	_, err = io.Copy(file, bytes.NewBuffer(content))
	if err != nil {
		return err
	}
	return nil
}

func getResourceSum(sum string) (bool, string, error) {
	db := utils.GetMySqlDB()

	row := db.QueryRow("select rID from tResources where sum = ?", sum)

	var rID string
	err := row.Scan(&rID)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}

	return true, rID, nil
}
