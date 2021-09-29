package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/twinj/uuid"
	"noz.zkip.cc/utils/model"
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
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
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
	at := time.Unix(td.AtExpires, 0)
	now := time.Now()

	errAccess := client.Set(ToString(userid), td.AccessUuid, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}

	return nil
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

func GetTokenValidation(tokenCredential *model.TokenCredential) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenCredential.Value, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &model.TokenInvalidAlgErr{Alg: t.Header["alg"]}
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if IsExpiredToken(err) {
		return nil, &model.TokenExpiredErr{}
	}

	// Check the dirt token

	if err != nil {
		return nil, &model.TokenParseErr{Msg: err.Error()}
	}

	return token, nil
}

// func IshDirtToken(access_uuid string) {
// 	redisClient := GetRedisClient()
// 	// uid, err := redisClient.GeoHash("sdf",)
// }

// Valid before
func ExtractDataFromToken(token *jwt.Token) uint64 {
	claims := token.Claims.(jwt.MapClaims)
	useridRaw := claims["user_id"].(float64)

	return uint64(useridRaw)
}

var client *redis.Client
var redisInited = false

func GetRedisClient() *redis.Client {
	if !redisInited {
		initRedisClient()
	}
	return client
}

func initRedisClient() {
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	client = redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	redisInited = true
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

func (rjp *ResponseJsonProvider) Send(result model.JsonResponser) {
	bytes, _ := json.Marshal(result)
	rjp.Rw.Write(bytes)
}
