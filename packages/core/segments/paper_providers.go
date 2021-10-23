package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/twinj/uuid"
	"noz.zkip.cc/utils"
)

func paperFactoryProvider(rw http.ResponseWriter, r *http.Request) {
	es := utils.GenHandlerUtils(rw)
	db := utils.GetMySqlDB()

	var spo *setterPaperOption
	var err error

	rID := uuid.NewV4().String()
	ownerPRI := utils.ExtractUserPRIFromRequst(r)
	resourcePRI := utils.GenPRI(rID, utils.Resource_type_paper)

	if utils.RunIfErr(utils.ExtractJsonBodyFromRequest(r, &spo), es.SendOptionErr) {
		return
	}

	initContent := []byte(spo.Content)

	tx, err := db.Begin()
	if utils.RunIfErr(err, es.SendInternalServerErr) {
		return
	}

	rollback := func(err error) {
		if tx != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
		}
	}

	sum := utils.GenSum([]byte(``))

	_, err = tx.Exec("insert into tResources(rid, ownerPRI, alias, mimeType, sum, kind) values( ?, ?, ?, 'text/noz-paper', ?, ? )", rID, ownerPRI, spo.Name, sum, utils.Resource_type_paper)
	if utils.RunIfErr(err, es.SendInternalServerErr, rollback) {
		return
	}

	_, err = tx.Exec("insert into tPermissions(which, resourcePRI) values( ?, ? )", visibility_private, resourcePRI)
	if utils.RunIfErr(err, es.SendInternalServerErr, rollback) {
		return
	}

	if utils.RunIfErr(savePaper(initContent, rID, true), es.SendInternalServerErr, rollback) {
		return
	}

	if utils.RunIfErr(tx.Commit(), es.SendInternalServerErr) {
		return
	}

	es.SendJSON(0, rID)
}

func paperProvider(rw http.ResponseWriter, r *http.Request) {
	es := utils.GenHandlerUtils(rw)
	db := utils.GetMySqlDB()

	id := utils.ExtractDynamicRouteID(dynamic_route_pattern_paper, r)

	userPRI := utils.ExtractUserPRIFromRequst(r)
	resourcePRI := utils.GenPRI(id, utils.Resource_type_paper)

	perm, err := getPermission(resourcePRI, userPRI)
	if utils.RunIfPickErr(err, ResourceNonExistErr_E)(es.SendNotFoundErr) {
		return
	}
	if utils.RunIfErr(err, es.SendOptionErr) {
		return
	}

	if utils.RunIfOK(!perm.canRead(), es.SendPermDeniedErr) {
		return
	}

	filepath := fmt.Sprintf("data/papers/%s", id)

	fileBytes, err := ioutil.ReadFile(filepath)
	if utils.RunIfErr(err, es.SendInternalServerErr) {
		return
	}

	paperData := &PaperData{"", id, fileBytes}

	row := db.QueryRow("select alias from tResource where rID = ?", id)
	err = row.Scan(&paperData.Name)
	if utils.RunIfErr(err, es.SendInternalServerErr) {
		return
	}

	es.SendJSON(0, "", paperData)
}

func paperRemoverProvider(rw http.ResponseWriter, r *http.Request) {
	es := utils.GenHandlerUtils(rw)
	id := utils.ExtractDynamicRouteID(dynamic_route_pattern_paper, r)

	userID := utils.ExtractUserIDFromRequst(r)
	takerPRI := utils.GenPRI(userID, utils.Resource_type_user)
	resourcePRI := utils.GenPRI(id, utils.Resource_type_paper)

	perm, err := getPermission(resourcePRI, takerPRI)
	if utils.RunIfPickErr(err, ResourceNonExistErr_E)(es.SendNotFoundErr) {
		return
	}
	if utils.RunIfOK(!perm.canWrite(), es.SendPermDeniedErr) {
		return
	}

	err = removeResource(takerPRI, id, utils.Resource_type_paper)
	if utils.RunIfPickErr(err, &utils.NotFoundErr{})() {
		return
	}
	utils.PanicIfErr(err)
}

func paperContentSetterProvider(rw http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(rw)

	var paperData *PaperDataOption

	if utils.RunIfErr(utils.ExtractJsonBodyFromRequest(r, &paperData), hu.SendOptionErr) {
		return
	}

	// check option
	if utils.RunIfOK(paperData.ID == "", hu.SendOptionErr) {
		return
	}

	rID := paperData.ID
	userPRI := utils.ExtractUserPRIFromRequst(r)
	resourcePRI := utils.GenPRI(rID, utils.Resource_type_paper)

	perm, err := getPermission(resourcePRI, userPRI)
	if utils.RunIfPickErr(err, ResourceNonExistErr_E)(hu.SendNotFoundErr) {
		return
	}
	if utils.RunIfOK(!perm.canWrite(), hu.SendPermDeniedErr) {
		return
	}
	err = savePaper([]byte(paperData.Content), rID, false)
	if utils.RunIfErr(err, hu.SendInternalServerErr) {
		return
	}
}

func paperAliasSetterProvider(rw http.ResponseWriter, r *http.Request) {
	hc := utils.GenHandlerUtils(rw)
	db := utils.GetMySqlDB()

	var data setterPaperOption

	if utils.RunIfErr(utils.ExtractJsonBodyFromRequest(r, &data), hc.SendOptionErr, hc.ByErrJSON(17)) {
		// JSON解析失败
		return
	}

	// check args
	if utils.RunIfOK(data.ID == "", hc.SendOptionErr, hc.ByErrJSON(18)) {
		return
	}

	takerPRI := utils.ExtractUserPRIFromRequst(r)
	resourcePRI := utils.GenPRI(data.ID, utils.Resource_type_paper)

	perm, err := getPermission(resourcePRI, takerPRI)
	if utils.RunIfPickErr(err, ResourceNonExistErr_E)(hc.SendNotFoundErr) {
		return
	}

	if utils.RunIfOK(!perm.canWrite(), hc.SendPermDeniedErr) {
		return
	}

	_, err = db.Exec("update tResources set alias = ? where rID = ?", data.Name, data.ID)
	utils.PanicIfErr(err)
}

func paperListProvider(rw http.ResponseWriter, r *http.Request) {
	es := utils.GenHandlerUtils(rw)
	db := utils.GetMySqlDB()

	var data = make([]PaperMetaData, 0)

	takerPRI := utils.ExtractUserPRIFromRequst(r)

	rows, err := db.Query("select rID, alias from tResources where kind = ? and ownerPRI = ?", utils.Resource_type_paper, takerPRI)
	utils.PanicIfErr(err)

	for rows.Next() {
		var meta PaperMetaData

		utils.PanicIfErr(rows.Scan(&meta.ID, &meta.Name))

		data = append(data, meta)
	}

	utils.PanicIfErr(es.SendJSON(0, "", data))
}
