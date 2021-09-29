package main

import (
	"fmt"
	"net/http"
	"os"

	"noz.zkip.cc/utils"
	"noz.zkip.cc/utils/model"
)

var UserDB = struct {
	secret map[uint64]string
	name   map[uint64]string
	email  map[string]uint64
}{
	map[uint64]string{
		1000: utils.EncodingStringMd5("pbo980"),
	},
	map[uint64]string{
		1000: "zkip",
	},
	map[string]uint64{
		"zkiplan@qq.com": 1000,
	},
}

type WR = model.WhichResult
type TR = model.TokenResult
type R = model.Result

func isValid(credential, secret string) bool {
	userID, ok := getUserIDByEmail(credential)
	return ok && UserDB.secret[userID] == secret
}

func getUserIDByEmail(email string) (uint64, bool) {
	userID, ok := UserDB.email[email]

	return userID, ok
}

func whichProvider(rw http.ResponseWriter, r *http.Request) {
	rjp := utils.ResponseJsonProvider{Rw: rw}
	tokenCredential, err := utils.ExtractTokenCredential(r)
	if _, ok := err.(*model.TokenCredentialEmptyErr); ok {
		rjp.Send(&WR{Result: R{Code: 50, Msg: ""}, Which: 0}) // 没有指定token
		return
	}

	token, err := utils.GetTokenValidation(tokenCredential)
	if _, ok := err.(*model.TokenExpiredErr); ok {
		rjp.Send(&WR{Result: R{Code: 51, Msg: ""}, Which: 0}) // token过期
		return
	}
	if err, ok := err.(*model.TokenExpiredErr); ok {
		rjp.Send(&WR{Result: R{Code: 52, Msg: err.Error()}, Which: 0}) // token解析错误
		return
	}

	fmt.Println(err)

	which := utils.ExtractDataFromToken(token)
	rjp.Send(&WR{Result: R{Code: 0, Msg: ""}, Which: which}) // token解析错误
}

func authProvider(rw http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	credential := query.Get("credential")
	secret := query.Get("secret")

	hasCredential := credential != ""
	hasSecret := secret != ""

	rjp := utils.ResponseJsonProvider{Rw: rw}

	if !hasCredential || !hasSecret {
		rjp.Send(&R{Code: 30, Msg: ""}) // 没有提供完整的认证信息
		return
	}

	userid, ok := getUserIDByEmail(credential)

	tokenDetails, err := utils.CreateToken(userid)
	if err != nil {
		rjp.Send(&TR{Result: R{Code: 33}}) // token生成错误
	}

	if !ok {
		rjp.Send(&TR{Result: R{Code: 31}}) // 该凭据没有被记录

	} else if !isValid(credential, secret) {
		rjp.Send(&TR{Result: R{Code: 32}}) // 凭据与密钥不符

	} else {
		rjp.Send(&TR{
			Result:       R{Code: 0, Msg: ""},
			AccessToken:  tokenDetails.AccessToken,
			RefreshToken: tokenDetails.RefreshToken,
		})
	}
}

// TODO: Refresh Token
func accessProvider(rw http.ResponseWriter, r *http.Request) {

}

func logoffProvider(rw http.ResponseWriter, r *http.Request) {

}

func main() {

	// bec84dd977e71cb7642e0785b7a7d972
	// fmt.Println(encodingStringMd5("pbo980"))
	os.Setenv("ACCESS_SECRET", "fallingtosky")
	os.Setenv("REFRESH_SECRET", "poppullpump")

	http.HandleFunc("/", whichProvider)
	http.HandleFunc("/auth", authProvider)
	http.ListenAndServe(utils.GetServeHost(7000), nil)
}
