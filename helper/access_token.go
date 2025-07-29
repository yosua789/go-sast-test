package helper

import (
	"assist-tix/config"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type JwtData struct {
	Exp           string
	TransactionID string
}

type JwtKeys struct {
	Exp           string
	TransactionID string
}

func GetJwtKeys() JwtKeys {
	return JwtKeys{
		Exp:           "exp",
		TransactionID: "transaction_id",
	}
}

func GenerateAccessToken(env *config.EnvironmentVariable, transactionID string) (token string, err error) {

	// Load env
	tokenSecretKey := env.AccessToken.SecretKey

	atClaims := jwt.MapClaims{}
	atClaims[GetJwtKeys().TransactionID] = transactionID
	atClaims[GetJwtKeys().Exp] = time.Now().Add(time.Hour * 24).Unix()

	log.Info().Interface("atClaims", atClaims).Msg("claims")

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err = at.SignedString([]byte(tokenSecretKey))

	if err != nil {
		return
	}

	return
}
func TokenExpired(token *jwt.Token) bool {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return true
	}

	exp := int64(claims[GetJwtKeys().Exp].(float64))
	expTime := time.Unix(exp, 0)
	now := time.Now()

	return now.After(expTime)
}

func AccessTokenValid(r *http.Request, env *config.EnvironmentVariable, cookies string) error {
	if cookies == "" {
		token, err := VerifyAccessToken(r, env)
		if err != nil {
			return errors.New("token is expired or invalid")
		}
		if token == nil || TokenExpired(token) {
			return errors.New("token is expired or invalid")
		}
	} else {
		token, err := VerifyCookieAccessToken(cookies, env)
		if err != nil {
			return errors.New("token is expired or invalid")
		}
		if token == nil || TokenExpired(token) {
			return errors.New("token is expired or invalid")
		}
	}
	return nil
}

func VerifyAccessToken(r *http.Request, env *config.EnvironmentVariable) (*jwt.Token, error) {
	tokenString := ExtractAccessToken(r)

	if tokenString == "" {
		err := errors.New("token is required")
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(env.AccessToken.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
func VerifyCookieAccessToken(cookies string, env *config.EnvironmentVariable) (*jwt.Token, error) {
	token, err := jwt.Parse(cookies, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(env.AccessToken.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func ExtractAccessToken(r *http.Request) string {
	tokenBearer := r.Header.Get("Authorization")

	strArr := strings.Split(tokenBearer, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func ExtractAccessTokenMetadata(r *http.Request, env *config.EnvironmentVariable) (*JwtData, error) {
	token, err := VerifyAccessToken(r, env)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		exp := fmt.Sprintf("%s", claims[GetJwtKeys().Exp])
		transactionID := fmt.Sprintf("%s", claims[GetJwtKeys().TransactionID])
		return &JwtData{
			Exp:           exp,
			TransactionID: transactionID,
		}, nil
	}

	return nil, err
}

func GetDataFromAccessToken(r *http.Request, env *config.EnvironmentVariable) (transactionID string, err error) {

	jwtData, err := ExtractAccessTokenMetadata(r, env)
	if err != nil {
		log.Error().Err(err).Msg("failed to extract data from token")
		return
	}

	transactionID = jwtData.TransactionID

	log.Info().Interface("jwtData", jwtData).Str("transactionID", transactionID).Msg("Jwt Data")

	if transactionID == "" {
		err = errors.New("transaction_id from token is empty")
		return
	}

	return
}
