package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"noz.zkip.cc/utils"
)

func quotaProvider(wr http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(wr)
	db := utils.GetMySqlDB()

	userID := utils.ExtractUserIDFromRequst(r)
	userPRI := utils.GenPRI(userID, utils.Resource_type_user)
	var qr QuotaResult

	row := db.QueryRow("select capcity, used from tQuotas where targetPRI = ?", userPRI)

	err := row.Scan(&qr.Capcity, &qr.Used)
	if utils.RunIfPickErr(err, sql.ErrNoRows)(hu.SendNotFoundErr) {
		return
	}
	utils.PanicIfErr(err)

	hu.SendJSON(0, "", qr)
}

// accounts
func accountProvider(wr http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(wr)
	db := utils.GetMySqlDB()

	data := &AccountResult{}

	userID := utils.ExtractUserIDFromRequst(r)

	row := db.QueryRow("select id, nickname, email from tAccounts where id = ?", userID)
	utils.PanicIfErr(row.Scan(&data.ID, &data.Nickname, &data.Email))

	utils.PanicIfErr(hu.SendJSON(0, "", data))
}
func accountPatcherProvider(wr http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(wr)
	db := utils.GetMySqlDB()

	var option *AccountPatchOption

	userID := utils.ExtractUserIDFromRequst(r)
	utils.ExtractJsonBodyFromRequest(r, &option)

	// check args
	if utils.RunIfOK(option.Email != "" && !utils.IsEmail(option.Email), hu.SendOptionErr, hu.ByErrJSON(27)) {
		// 参数不合法
		return
	}

	holders := make([]string, 0)
	values := make([]interface{}, 0)

	if option.Nickname != "" {
		holders = append(holders, "nickname = ?")
		values = append(values, option.Nickname)
	}
	if option.Email != "" {
		holders = append(holders, "email = ?")
		values = append(values, option.Email)
	}

	holderQtr := strings.Join(holders, ", ")
	values = append(values, userID)

	qtr := fmt.Sprintf("update tAccounts set %s where id = ?", holderQtr)
	_, err := db.Exec(qtr, values...)
	utils.PanicIfErr(err)
}
