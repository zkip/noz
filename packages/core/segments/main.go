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

	"github.com/gabriel-vasile/mimetype"
	"github.com/twinj/uuid"
	"noz.zkip.cc/utils"
	"noz.zkip.cc/utils/model"
)

type R = model.Result

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
	R
	Data []*PaperMetaData
}

type ImageMetaListResult struct {
	R
	Data []*ImageMetaData
}

type PaperListResult struct {
	R
	Data []PaperData
}

type NoStoreResourceTypeErr struct {
	rType uint8
}

func (ne *NoStoreResourceTypeErr) Error() string {
	return fmt.Sprintf("%s is no store resource type.", utils.GetResoureceTypeIdent(ne.rType))
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

	tx, err := db.Begin()
	utils.PanicIfErr(err)

	rollback := func(err error) {
		if tx != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
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

	mimeType := mimetype.Detect(fileBytes)

	rID := uuid.NewV4().String()

	userID := utils.ExtractUserIDFromRequst(r)
	ownerPRI := utils.GenPRI(userID, utils.Resource_type_user)

	resourcePRI := utils.GenPRI(rID, utils.Resource_type_image)

	fmt.Println("Size: ", len(fileBytes), mimeType, "by: ", ownerPRI)

	db := utils.GetMySqlDB()
	if getSupportedMimeType(mimeType) {
		saveImage(fileBytes, rID)

		_, err := db.Exec("insert into tResources(rid, ownerPRI, mimeType, sum, kind) values( ?, ?, ?, ?, ? )", rID, ownerPRI, mimeType.String(), sum, utils.Resource_type_image)

		if err != nil {
			panic(err)
		}

		_, err = db.Exec("insert into tPermissions(which, resourcePRI) values( ?, ? )", visibility_private, resourcePRI)

		if err != nil {
			panic(err)
		}

		rjp.Send(0, rID)
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

func setupServer() {
	httpServer := utils.NewNOZHTTPServer()

	httpServer.HandleFunc(newImageProvider, "^/image", utils.Http_method_new)
	httpServer.HandleFunc(imageProvider, "^/image", utils.Http_method_get)
	httpServer.HandleFunc(imageRemoverProvider, "^/image", utils.Http_method_delete)

	httpServer.HandleFunc(newPaperProvider, "^/paper", utils.Http_method_new)
	httpServer.HandleFunc(paperProvider, "^/paper", utils.Http_method_get)
	httpServer.HandleFunc(paperRemoverProvider, "^/paper", utils.Http_method_delete)

	httpServer.HandleFunc(paperAliasSetterProvider, "^/paper/name", utils.Http_method_set)
	httpServer.HandleFunc(paperContentSetterProvider, "^/paper/content", utils.Http_method_set)

	httpServer.HandleFunc(paperListProvider, "^/paper/list", utils.Http_method_action)
	httpServer.HandleFunc(imageListProvider, "^/image/list", utils.Http_method_action)

	utils.PanicIfErr(http.ListenAndServe(utils.GetServeHost(7703), httpServer))
}

func main() {
	defer utils.Setup()()

	setupStoreEnv()
	setupServer()
}
