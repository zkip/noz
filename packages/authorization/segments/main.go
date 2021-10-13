package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"noz.zkip.cc/utils"
	"noz.zkip.cc/utils/model"
)

const (
	credential_kind_unknow = iota
	credential_kind_email
)

type Credential struct {
	kind   uint8
	email  string
	passwd string
}

func genCredentialByMasterMode(major, minor string) *Credential {
	credential := &Credential{}
	kind := resolveMajorKind(major)
	switch kind {
	case credential_kind_email:
		credential.email = major
		credential.passwd = minor
	default:
		//
	}

	credential.kind = kind
	return credential
}

func resolveMajorKind(major string) uint8 {
	if ok, _ := regexp.MatchString(`^([^@]+)@[^\.]+\.([^\.]+)$`, major); ok {
		return credential_kind_email
	}

	return credential_kind_unknow
}

type WR = model.WhichResult
type TR = model.TokenResult
type R = model.Result

type SecretWrongErr struct{}

func (sw *SecretWrongErr) Error() string {
	return "Secret wrong."
}

type UnsupportedCredentialKindErr struct{}

func (sw *UnsupportedCredentialKindErr) Error() string {
	return "Unsupported credential kind."
}

type CredentialIncompleteErr struct{}

func (ci *CredentialIncompleteErr) Error() string {
	return "Credential incomplete."
}

func resolveUserID(credential *Credential) (uint64, error) {
	var userID uint64
	var passwd string

	db := utils.GetMySqlDB()

	if credential.kind != credential_kind_email {
		return 0, &UnsupportedCredentialKindErr{}
	}

	row := db.QueryRow("select id, passwd from tAccounts where email = ?", credential.email)
	err := row.Scan(&userID, &passwd)

	if err == sql.ErrNoRows {
		return 0, &utils.NotFoundErr{Name: "email"}
	} else if err != nil {
		return 0, err
	}

	if passwd != credential.passwd {
		return 0, &SecretWrongErr{}
	}

	return userID, nil
}

var (
	TokenExpiredErr_E         = &model.TokenExpiredErr{}
	TokenCredentialEmptyErr_E = &model.TokenCredentialEmptyErr{}

	NotFoundErr_E = &utils.NotFoundErr{}

	CredentialIncompleteErr_E      = &CredentialIncompleteErr{}
	UnsupportedCredentialKindErr_E = &UnsupportedCredentialKindErr{}
	SecretWrongErr_E               = &SecretWrongErr{}
)

func whichProvider(wr http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(wr)

	tokenCredential, err := utils.ExtractTokenCredential(r)
	if utils.RunIfPickErr(err, TokenCredentialEmptyErr_E)(hu.SendOptionErr, hu.ByErrJSON(50)) {
		// 没有指定token
		return
	}
	if utils.RunIfErr(err, hu.SendOptionErr) {
		// unkown reason
		return
	}

	token, err := utils.ParseToken(tokenCredential)
	if utils.RunIfPickErr(err, TokenExpiredErr_E)(hu.ByErrJSON(51)) {
		// token过期
		return
	}
	if utils.RunIfErr(err, hu.SendOptionErr, hu.ByErrJSON(52)) {
		// token解析错误
		return
	}

	which := utils.ExtractDataFromToken(token)

	wr.Header().Add("X-Authenticate-Which", utils.ToString(which))
	wr.WriteHeader(901)
}

func extractCredential(r *http.Request) (*Credential, error) {
	query := r.URL.Query()

	credential := query.Get("credential")
	secret := query.Get("secret")

	hasCredential := credential != ""
	hasSecret := secret != ""

	if !hasCredential || !hasSecret {
		return nil, &CredentialIncompleteErr{}
	}

	result := genCredentialByMasterMode(credential, secret)

	return result, nil
}

var RunIfPickErr = utils.RunIfPickErr
var RunIfErr = utils.RunIfErr
var RunIfOK = utils.RunIfOK

func authProvider(rw http.ResponseWriter, r *http.Request) {
	es := utils.GenHandlerUtils(rw)

	credential, err := extractCredential(r)
	if RunIfPickErr(err, CredentialIncompleteErr_E)(es.SendOptionErr, es.ByErrJSON(35)) {
		// 没有提供完整的认证信息
		return
	}

	userID, err := resolveUserID(credential)
	if RunIfPickErr(err, UnsupportedCredentialKindErr_E)(es.SendOptionErr, es.ByErrJSON(34, utils.ToString(credential.kind))) {
		// 不支持的凭据类型
		return
	}
	if RunIfPickErr(err, NotFoundErr_E)(es.ByErrJSON(31)) {
		// 该凭据没有被记录
		return
	}
	if RunIfPickErr(err, SecretWrongErr_E)(es.SendOptionErr, es.ByErrJSON(32)) {
		// 凭据与密钥不符
		return
	}

	tokenDetails, err := utils.CreateToken(userID)
	if RunIfErr(err, es.SendOptionErr, es.ByErrJSON(33)) {
		// token生成错误
		return
	}

	if RunIfErr(utils.RecordAccessUuid(userID, tokenDetails), es.SendInternalServerErr) {
		return
	}

	es.SendJSON(0, "", &TR{
		AccessToken:  tokenDetails.AccessToken,
		RefreshToken: tokenDetails.RefreshToken,
	})
}

func logoffProvider(rw http.ResponseWriter, r *http.Request) {

	tokenCredential, err := utils.ExtractTokenCredential(r)
	if err != nil {
		panic(err)
	}

	token, err := utils.ParseToken(tokenCredential)
	if err != nil {
		panic(err)
	}

	userID := utils.ExtractDataFromToken(token)

	redis := utils.GetRedisClient()
	result, err := redis.Del(utils.ToString(userID)).Result()

	fmt.Println(result, err)
}

func main() {
	defer utils.Setup()()

	os.Setenv("ACCESS_SECRET", "fallingtosky")
	os.Setenv("REFRESH_SECRET", "poppullpump")

	httpServer := utils.NewNOZHTTPServer()

	httpServer.HandleFunc(whichProvider, "/")
	httpServer.HandleFunc(authProvider, "/auth")

	http.ListenAndServe(utils.GetServeHost(7000), httpServer)
}
