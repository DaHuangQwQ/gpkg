package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"strings"
	"time"
)

type LocalJWTOption func(*LocalJWTHandler)

type LocalJWTHandler struct {
	token map[string]UserClaims

	// 哈希算法
	signingMethod jwt.SigningMethod
	//
	rcExpiration time.Duration

	JWTKey   []byte
	RCJWTKey []byte
}

func NewLocalJWTHandler(JWTKey []byte, RCJWTKey []byte, rcExpiration time.Duration, opts ...LocalJWTOption) *LocalJWTHandler {
	res := &LocalJWTHandler{
		token:         make(map[string]UserClaims),
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  rcExpiration,
		JWTKey:        JWTKey,
		RCJWTKey:      RCJWTKey,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func (l *LocalJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("claims").(UserClaims)

	delete(l.token, l.key(uc.Ssid))
	return nil
}

func (l *LocalJWTHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return authCode
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

func (l *LocalJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := l.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return l.SetJWTToken(ctx, uid, ssid)
}

func (l *LocalJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(l.rcExpiration)),
		},
	}
	// 加密
	token := jwt.NewWithClaims(l.signingMethod, uc)
	tokenStr, err := token.SignedString(l.JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (l *LocalJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	_, ok := l.token[ssid]
	if ok {
		return errors.New("session 无效")
	}
	return nil
}

func (l *LocalJWTHandler) ParseWithClaims(tokenStr string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return l.JWTKey, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token is invalid")
	}
	return nil
}

func (l *LocalJWTHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}
	token := jwt.NewWithClaims(l.signingMethod, rc)
	tokenStr, err := token.SignedString(l.RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (l *LocalJWTHandler) key(key string) string {
	return fmt.Sprintf("users:ssid:%s", key)
}

func LocalJWTWithSigningMethod(method jwt.SigningMethod) LocalJWTOption {
	return func(handler *LocalJWTHandler) {
		handler.signingMethod = method
	}
}
