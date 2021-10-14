package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/twinj/uuid"
	"noz.zkip.cc/utils"
	"noz.zkip.cc/utils/model"
)

type R = model.Result

type setterImageAliasOption struct {
	ID   string
	Name string
}

type setterPaperOption struct {
	ID      string
	Name    string
	Content string
}

type PaperMetaData struct {
	Name string
	ID   string
}

type ImageMetaData struct {
	Name string
	ID   string
}

type PaperData struct {
	Name    string
	ID      string
	Content []byte
}

type PaperDataOption struct {
	ID      string
	Content string
}

type PaperMetaListResult struct {
	Data []*PaperMetaData
}

type ImageMetaListResult struct {
	Data []*ImageMetaData
}

type PaperListResult struct {
	Data []PaperData
}

type AccountResult struct {
	ID       string
	Nickname string
	Email    string
}

type AccountPatchOption struct {
	ID       string
	Nickname string
	Email    string
}

type Quota struct {
	Capcity uint64
	Used    uint64
}

func (q *Quota) doDelta(size int64) error {
	used := int64(q.Used) + size

	if used+size < 0 {
		used = 0
	}

	if used+size > int64(q.Capcity) {
		return &QuotaLackErr{q}
	}

	q.Used = uint64(used)

	return nil
}

type QuotaResult struct {
	Capcity uint64
	Used    uint64
}

type NoStoreResourceTypeErr struct {
	rType uint8
}

func (ne *NoStoreResourceTypeErr) Error() string {
	return fmt.Sprintf("%s is no store resource type.", utils.GetResoureceTypeIdent(ne.rType))
}

type QuotaLackErr struct {
	quota *Quota
}

func (qe *QuotaLackErr) Error() string {
	return fmt.Sprintf("Quota is lacked (%d/%d)).", qe.quota.Capcity, qe.quota.Used)
}

var maxUploadSize int64 = 1024 * 1024 * 1024 * 1024

const (
	dynamic_route_pattern_image = `^/image/([^\/]+)$`
	dynamic_route_pattern_paper = `^/paper/([^\/]+)$`
)

var resource_store_path = map[uint8]string{
	utils.Resource_type_image: "data/images",
	utils.Resource_type_paper: "data/papers",
}

var (
	ResourceNonExistErr_E = &ResourceNonExistErr{}
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

func newImageProvider(rw http.ResponseWriter, r *http.Request) {
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

func newPaperProvider(rw http.ResponseWriter, r *http.Request) {
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

func setupServer() {
	httpServer := utils.NewNOZHTTPServer()

	httpServer.HandleFunc(newImageProvider, "^/image", utils.Http_method_new)
	httpServer.HandleFunc(imageProvider, "^/image", utils.Http_method_get)
	httpServer.HandleFunc(imageRemoverProvider, "^/image", utils.Http_method_delete)
	httpServer.HandleFunc(imageAliasSetterProvider, "^/image/alias", utils.Http_method_set)

	httpServer.HandleFunc(newPaperProvider, "^/paper", utils.Http_method_new)
	httpServer.HandleFunc(paperProvider, "^/paper", utils.Http_method_get)
	httpServer.HandleFunc(paperRemoverProvider, "^/paper", utils.Http_method_delete)

	httpServer.HandleFunc(paperAliasSetterProvider, "^/paper/name", utils.Http_method_set)
	httpServer.HandleFunc(paperContentSetterProvider, "^/paper/content", utils.Http_method_set)

	httpServer.HandleFunc(paperListProvider, "^/paper/list", utils.Http_method_action)
	httpServer.HandleFunc(imageListProvider, "^/image/list", utils.Http_method_action)

	httpServer.HandleFunc(quotaProvider, "^/quota", utils.Http_method_get)
	httpServer.HandleFunc(accountProvider, "^/account", utils.Http_method_get)
	httpServer.HandleFunc(accountPatcherProvider, "^/account", utils.Http_method_patch)

	utils.PanicIfErr(http.ListenAndServe(utils.GetServeHost(7703), httpServer))
}

func main() {
	defer utils.Setup()()

	// setupStoreEnv()
	// setupServer()

	test()

}

func newHierarchyRecord(parentID string, targetPRI string, name string) (string, error) {
	db := utils.GetMySqlDB()
	hierarchyID := genHierarchyID(targetPRI)

	var err error

	tx, err := db.Begin()
	if err != nil {
		return "", err
	}

	rollback := func(err error) {
		if tx != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
		}
	}

	_, err = tx.Exec("insert into tHierarchy( ancestor, descendant, distance, targetPRI ) ( select ancestor, ?, distance + 1, ? from tHierarchy where descendant = ? )", hierarchyID, targetPRI, parentID)
	if utils.RunIfErr(err, rollback) {
		return "", err
	}

	_, err = tx.Exec("insert into tHierarchy( ancestor, descendant, distance, targetPRI ) values( ?, ?, 0, ? )", hierarchyID, hierarchyID, targetPRI)
	if utils.RunIfErr(err, rollback) {
		return "", err
	}

	_, err = tx.Exec("insert into tHierarchyData( hierarchyID, name ) values( ?, ? )", hierarchyID, name)
	if utils.RunIfErr(err, rollback) {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", nil
	}

	return hierarchyID, nil
}

func findChildren(hierarchyID string) ([]string, error) {
	db := utils.GetMySqlDB()
	var descendants = make([]string, 0)

	rows, err := db.Query("select name from tHierarchyData where hierarchyID in (select descendant from tHierarchy where ancestor = ? and distance = 1)", hierarchyID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var descendant string

		err := rows.Scan(&descendant)
		if err != nil {
			return nil, err
		}

		descendants = append(descendants, descendant)
	}

	return descendants, nil
}

func findPath(hierarchyID string) ([]string, error) {
	db := utils.GetMySqlDB()
	var ancestors = make([]string, 0)

	rows, err := db.Query("select name from tHierarchyData where hierarchyID in (select ancestor from tHierarchy where descendant = ? order by distance desc)", hierarchyID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var descendant string

		err := rows.Scan(&descendant)
		if err != nil {
			return nil, err
		}

		ancestors = append(ancestors, descendant)
	}

	return ancestors, err
}

func delete(hierarchyID string) error {
	db := utils.GetMySqlDB()

	var err error

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	rollback := func(err error) {
		if tx != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
		}
	}

	_, err = db.Exec("delete from tHierarchy where descendant in (select * from (select descendant from tHierarchy where ancestor = ?) as _)", hierarchyID)
	if utils.RunIfErr(err, rollback) {
		return err
	}
	_, err = db.Exec("delete from tHierarchyData where hierarchyID = ?", hierarchyID)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	return tx.Commit()
}

func renameHierarchyRecord(hierarchyID string, name string) error {
	db := utils.GetMySqlDB()
	_, err := db.Exec("update tHierarchyData set name = ? where hierarchyID = ?", name, hierarchyID)
	if err != nil {
		return err
	}
	return nil
}

func genHierarchyID(userPRI string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(userPRI+time.Now().String())))
}

func test() {
	// var err error
	// ownerPRI := "us/2"

	// idRoot, _ := newHierarchyRecord("", ownerPRI, "crack")
	// idPlanA, _ := newHierarchyRecord(idRoot, ownerPRI, "plan A")
	// idCrashStock, _ := newHierarchyRecord(idPlanA, ownerPRI, "crash stock")
	// newHierarchyRecord(idPlanA, ownerPRI, "field jump")
	// idPlanB, _ := newHierarchyRecord(idRoot, ownerPRI, "plan B")
	// newHierarchyRecord(idPlanB, ownerPRI, "run")

	// fmt.Println(findPath("20046eb7d5ca154469c4d07cd0de5b61"))

	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(id, "@@@@")

	// err := remove("cff12c3f3a04205e657cdebb35a63dc0")
	// utils.PanicIfErr(err)

	// findDepth(2)
	// fmt.Println(findChildren("bda78e4a601297d7eb9aa6d608e801c2"))
	// findPath(3)

	renameHierarchyRecord("bda78e4a601297d7eb9aa6d608e801c2", "Plan X")
}
