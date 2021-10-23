package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/twinj/uuid"
	"noz.zkip.cc/utils/model"

	_ "github.com/go-sql-driver/mysql"
)

func IsExpiredToken(err error) bool {
	return err != nil && (err.(*jwt.ValidationError).Errors&jwt.ValidationErrorExpired != 0)
}

func GenToken(userID string) string {
	atClaims := jwt.MapClaims{}
	atClaims["user_id"] = userID
	atClaims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, _ := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	return token
}

func CreateToken(userid uint64) (*model.TokenDetails, error) {
	td := &model.TokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * 24 * 30).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 30).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	return td, nil
}

func RecordAccessUuid(userid uint64, td *model.TokenDetails) error {
	redisClient := GetRedisClient()
	at := time.Unix(td.AtExpires, 0)
	now := time.Now()

	errAccess := redisClient.Set(ToString(userid), td.AccessUuid, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}

	return nil
}

func IsCurrentAccessUuid(userid uint64, accessUuid string) bool {
	redisClient := GetRedisClient()
	currentAccessUuid, err := redisClient.Get(ToString(userid)).Result()

	return err == nil && currentAccessUuid == accessUuid
}

func ExtractTokenCredential(r *http.Request) (*model.TokenCredential, error) {
	tokenString := ""
	tokenHeader := r.Header.Get("Authorization")
	inCookie := false
	inHeader := tokenHeader != ""

	if inHeader {
		tokenString = tokenHeader
	} else {
		tokenCookie, err := r.Cookie("access_token")
		inCookie = err == nil
		if inCookie {
			tokenString = tokenCookie.Value
		}
	}

	tokenString = strings.TrimSpace(tokenString)

	if !inCookie && !inHeader && tokenString == "" {
		return nil, &model.TokenCredentialEmptyErr{}
	}

	tokenCredential := &model.TokenCredential{}

	segs := strings.Split(tokenString, " ")
	if len(segs) == 1 {
		tokenCredential.Type = "Plain"
		tokenCredential.Value = segs[0]
	} else {
		tokenCredential.Type = segs[0]
		tokenCredential.Value = segs[1]
	}

	return tokenCredential, nil
}

func ParseToken(tokenCredential *model.TokenCredential) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenCredential.Value, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &model.TokenInvalidAlgErr{Alg: t.Header["alg"]}
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if err != nil {
		if IsExpiredToken(err) {
			return nil, &model.TokenExpiredErr{}
		}
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	useridRaw := claims["user_id"].(float64)
	accessUuid := claims["access_uuid"].(string)

	// Check the dirt token
	if !IsCurrentAccessUuid(uint64(useridRaw), accessUuid) {
		return nil, &model.TokenExpiredErr{}
	}

	if err != nil {
		return nil, &model.TokenParseErr{Msg: err.Error()}
	}

	return token, nil
}

// Valid before
func ExtractDataFromToken(token *jwt.Token) uint64 {
	claims := token.Claims.(jwt.MapClaims)
	useridRaw := claims["user_id"].(float64)

	return uint64(useridRaw)
}

func ExtractUserIDFromRequst(r *http.Request) uint64 {
	whichStr := r.Header.Get("X-Authenticate-Which")
	userID, err := strconv.Atoi(whichStr)
	if err != nil {
		return 0
	}
	return uint64(userID)
}

func ExtractUserPRIFromRequst(r *http.Request) string {
	userID := ExtractUserIDFromRequst(r)
	return GenPRI(userID, Resource_type_user)
}

func ExtractGID(groupPRI string) string {
	segs := strings.Split(groupPRI, "/")
	return segs[len(segs)-1]
}
func ExtractRID(resourcePRI string) string {
	segs := strings.Split(resourcePRI, "/")
	return segs[len(segs)-1]
}
func ExtractPRIID(PRI string) (uint64, error) {
	segs := strings.Split(PRI, "/")
	ID, err := strconv.Atoi(segs[1])
	if err != nil {
		return 0, err
	}
	return uint64(ID), nil
}

func GenPRI(id interface{}, rType uint8) string {
	idStr := ToString(id)
	typeStr := ToString(GetResoureceTypeIdent(rType))
	return strings.Join([]string{typeStr, idStr}, "/")
}

func GetResoureceTypeIdent(rType uint8) string {
	switch rType {
	case Resource_type_user:
		return "us"
	case Resource_type_group:
		return "gp"
	case Resource_type_paper:
		return "pr"
	case Resource_type_image:
		return "ig"
	default:
		return "un"
	}
}

const (
	Resource_type_user = iota
	Resource_type_group

	Resource_type_paper

	// binary
	Resource_type_image
)

var client *redis.Client
var redisReady = false

func GetRedisClient() *redis.Client {
	if !redisReady {
		initRedisClient()
	}
	return client
}

var db *sql.DB
var mysqlReady = false

func GetMySqlDB() *sql.DB {
	if !mysqlReady {
		initMySqlDB()
	}
	return db
}

func initMySqlDB() {
	var err error
	mysqlHost := ResolveMySqlServiceHost().Host
	dbname := "jjj"
	password := "123456"
	dsn := fmt.Sprintf("root:%s@(%s)/%s", password, mysqlHost, dbname)
	db, err = sql.Open("mysql", dsn)

	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	mysqlReady = true
}

func cleanSQL() {
	if mysqlReady {
		db.Close()
	}
}

func initRedisClient() {
	redisHost := ResolveRedisServiceHost()
	dsn := redisHost.Host

	client = redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	redisReady = true
}

func ParseJsonFromResponse(rw *http.Response, data interface{}) error {
	bytes, err := ioutil.ReadAll(rw.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &data)
}

type ResponseJsonProvider struct {
	Rw http.ResponseWriter
}

type NotFoundErr struct {
	Name string
}

func (nf *NotFoundErr) Error() string {
	return fmt.Sprintf("%s is not found.", nf.Name)
}

func (rjp *ResponseJsonProvider) Send(code int, msg string, data ...model.JsonResponser) {
	var tData interface{}

	if len(data) > 0 {
		tData = data[0]
	}

	result := map[string]interface{}{
		"Code": code,
		"Msg":  msg,
		"Data": tData,
	}

	bytes, _ := json.Marshal(result)
	rjp.Rw.Write(bytes)
}

// ---
var NoopErr = errors.New("NOOP")

func ExtractDynamicRouteID(pattern string, r *http.Request) string {
	rxp := regexp.MustCompile(pattern)
	c := rxp.FindSubmatch([]byte(r.URL.Path))
	id := string(c[1])
	return strings.TrimSpace(id)
}

func RunIfErr(err error, fns ...ErrHandler) bool {
	if err != nil {
		for _, fn := range fns {
			defer fn(err)
		}
		return true
	}
	return false
}

func RunIfOK(ok bool, fns ...ErrHandler) bool {
	if ok {
		for _, fn := range fns {
			defer fn(NoopErr)
		}
	}
	return ok
}

func RunIfPickErr(err, target error) func(...ErrHandler) bool {
	ok := reflect.TypeOf(err) == reflect.TypeOf(target)
	return func(fns ...ErrHandler) bool {
		if ok {
			for _, fn := range fns {
				fn(err)
			}
			return true
		}
		return false
	}
}

func PanicIfErr(err error, fns ...ErrHandler) {
	if err != nil {
		for _, fn := range fns {
			if fn != nil {
				defer fn(err)
			}
		}
		panic(err)
	}
}

func PickErr(vs ...interface{}) bool {
	if len(vs) > 1 {
		return vs[1].(bool)
	}
	return false
}

type ErrHandler func(error)

type HandlerUtils struct {
	wr http.ResponseWriter
}

func GenHandlerUtils(wr http.ResponseWriter) *HandlerUtils {
	return &HandlerUtils{wr}
}

func (hu *HandlerUtils) SendJSON(code int, msg string, data ...model.JsonResponser) error {
	var tData interface{}

	if len(data) > 0 {
		tData = data[0]
	}

	result := map[string]interface{}{
		"Code": code,
		"Msg":  msg,
		"Data": tData,
	}

	hu.wr.Header().Set("Content-Type", "application/json")

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	_, err = hu.wr.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func (hu *HandlerUtils) SendInternalServerErr(err error) {
	hu.wr.WriteHeader(http.StatusInternalServerError)
}
func (hu *HandlerUtils) SendOptionErr(err error) {
	hu.wr.WriteHeader(http.StatusBadRequest)
}
func (hu *HandlerUtils) SendNotFoundErr(err error) {
	hu.wr.WriteHeader(http.StatusNotFound)
}
func (hu *HandlerUtils) SendPermDeniedErr(err error) {
	hu.wr.WriteHeader(http.StatusForbidden)
}
func (hu *HandlerUtils) SendIllegalJsonOptionErr(err error) {
	hu.wr.WriteHeader(http.StatusBadRequest)
	hu.SendJSON(17, "")
}
func (hu *HandlerUtils) ByErrJSON(code int, msg ...string) ErrHandler {
	_msg := ""
	if len(msg) > 0 {
		_msg = msg[0]
	}

	return func(e error) {
		hu.SendJSON(code, _msg)
	}
}
func (hu *HandlerUtils) ByErrJSONEM(code int) ErrHandler {
	return func(e error) {
		hu.SendJSON(code, e.Error())
	}
}

func ExtractJsonBodyFromRequest(r *http.Request, data interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &data)
}
