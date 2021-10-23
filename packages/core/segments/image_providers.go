package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/twinj/uuid"
	"noz.zkip.cc/utils"
)

func imageFactoryProvider(rw http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(rw)
	rjp := utils.ResponseJsonProvider{Rw: rw}

	r.Body = http.MaxBytesReader(rw, r.Body, maxUploadSize)

	err := r.ParseMultipartForm(int64(maxUploadSize))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		rjp.Send(71, err.Error(), nil)
		return
	}

	fileBytes, _ := ioutil.ReadAll(file)

	sum := fmt.Sprintf("%x", md5.Sum(fileBytes))

	existed, iRID, err := getResourceSum(sum)
	if err != nil {
		rjp.Send(17, err.Error())
		return
	}

	if existed {
		rjp.Send(42, iRID)
		return
	}

	rID := uuid.NewV4().String()
	size := int64(len(fileBytes))
	mimeType := mimetype.Detect(fileBytes)
	userID := utils.ExtractUserIDFromRequst(r)
	ownerPRI := utils.GenPRI(userID, utils.Resource_type_user)
	resourcePRI := utils.GenPRI(rID, utils.Resource_type_image)

	quota, err := getQuota(ownerPRI)
	utils.PanicIfErr(err)

	err = quota.doDelta(size)
	if utils.RunIfPickErr(err, &QuotaLackErr{})(hu.SendOptionErr, hu.ByErrJSON(39)) {
		return
	}
	utils.PanicIfErr(err)

	db := utils.GetMySqlDB()
	if getSupportedMimeType(mimeType) {
		saveImage(fileBytes, rID)

		_, err := db.Exec("insert into tResources(rid, ownerPRI, quotaUsage, mimeType, sum, kind) values( ?, ?, ?, ?, ?, ? )", rID, ownerPRI, size, mimeType.String(), sum, utils.Resource_type_image)
		utils.PanicIfErr(err)

		_, err = db.Exec("insert into tPermissions(which, resourcePRI) values( ?, ? )", visibility_private, resourcePRI)
		utils.PanicIfErr(err)

		_, err = db.Exec("update tQuotas set used = ? where targetPRI = ?", quota.Used, ownerPRI)
		utils.PanicIfErr(err)

		hu.SendJSON(0, rID)
	}
}

func imageProvider(rw http.ResponseWriter, r *http.Request) {
	id := utils.ExtractDynamicRouteID(dynamic_route_pattern_image, r)

	userID := utils.ExtractUserIDFromRequst(r)
	userPRI := utils.GenPRI(userID, utils.Resource_type_user)

	resourcePRI := utils.GenPRI(id, utils.Resource_type_image)

	perm, err := getPermission(resourcePRI, userPRI)
	utils.PanicIfErr(err)

	if !perm.canRead() {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	filepath := fmt.Sprintf("data/images/%s", id)

	fileBytes, err := ioutil.ReadFile(filepath)
	utils.PanicIfErr(err)

	mimeType := mimetype.Detect(fileBytes)

	rw.Header().Add("Content-Type", mimeType.String())

	rw.Write(fileBytes)
}

func imageRemoverProvider(rw http.ResponseWriter, r *http.Request) {
	es := utils.GenHandlerUtils(rw)
	id := utils.ExtractDynamicRouteID(dynamic_route_pattern_image, r)

	userID := utils.ExtractUserIDFromRequst(r)
	takerPRI := utils.GenPRI(userID, utils.Resource_type_user)
	resourcePRI := utils.GenPRI(id, utils.Resource_type_image)

	perm, err := getPermission(resourcePRI, takerPRI)
	if utils.RunIfPickErr(err, ResourceNonExistErr_E)(es.SendNotFoundErr) {
		return
	}

	if utils.RunIfOK(!perm.canWrite(), es.SendPermDeniedErr) {
		return
	}

	err = removeResource(takerPRI, id, utils.Resource_type_image)
	if utils.RunIfPickErr(err, &utils.NotFoundErr{})() {
		return
	}
	utils.PanicIfErr(err)
}

func imageAliasSetterProvider(rw http.ResponseWriter, r *http.Request) {
	hc := utils.GenHandlerUtils(rw)
	db := utils.GetMySqlDB()

	var option setterImageAliasOption

	if utils.RunIfErr(utils.ExtractJsonBodyFromRequest(r, &option), hc.SendOptionErr, hc.ByErrJSON(17)) {
		// JSON解析失败
		return
	}

	// check args
	if utils.RunIfOK(option.ID == "", hc.SendOptionErr, hc.ByErrJSON(18)) {
		return
	}

	takerPRI := utils.ExtractUserPRIFromRequst(r)
	resourcePRI := utils.GenPRI(option.ID, utils.Resource_type_image)

	perm, err := getPermission(resourcePRI, takerPRI)
	if utils.RunIfPickErr(err, ResourceNonExistErr_E)(hc.SendNotFoundErr) {
		return
	}

	if utils.RunIfOK(!perm.canWrite(), hc.SendPermDeniedErr) {
		return
	}

	_, err = db.Exec("update tResources set alias = ? where rID = ?", option.Name, option.ID)
	utils.PanicIfErr(err)
}

func imageListProvider(rw http.ResponseWriter, r *http.Request) {
	es := utils.GenHandlerUtils(rw)
	db := utils.GetMySqlDB()

	userID := utils.ExtractUserIDFromRequst(r)
	ownerPRI := utils.GenPRI(userID, utils.Resource_type_user)

	var imageMetaList = make([]*ImageMetaData, 0)

	rows, err := db.Query("select rID, alias from tResources where ownerPRI = ? and kind = ?", ownerPRI, utils.Resource_type_image)
	if utils.RunIfErr(err, es.SendInternalServerErr) {
		return
	}

	for rows.Next() {
		var imageMeta ImageMetaData
		var alias sql.NullString

		err := rows.Scan(&imageMeta.ID, &alias)
		if utils.RunIfErr(err, es.SendInternalServerErr) {
			return
		}

		imageMeta.Name = alias.String
		imageMetaList = append(imageMetaList, &imageMeta)
	}

	utils.RunIfErr(es.SendJSON(0, "", imageMetaList), es.SendInternalServerErr)
}
