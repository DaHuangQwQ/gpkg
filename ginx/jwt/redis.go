package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var _ Handler = &RedisJWTHandler{}

type RedisJWTOption func(*RedisJWTHandler)

type RedisJWTHandler struct {
	client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration

	JWTKey   []byte
	RCJWTKey []byte
}

func NewRedisJWTHandler(JWTKey, RCJWTKey []byte, rcExpiration time.Duration, client redis.Cmdable, opts ...RedisJWTOption) Handler {
	res := &RedisJWTHandler{
		client:        client,
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  rcExpiration,

		JWTKey:   JWTKey,
		RCJWTKey: RCJWTKey,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch err {
	case redis.Nil:
		return nil
	case nil:
		if cnt == 0 {
			return nil
		}
		return errors.New("session 无效")
	default:
		return err
	}
}

// ExtractToken 根据约定，token 在 Authorization 头部
// Bearer XXXX
func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
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

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return h.SetJWTToken(ctx, uid, ssid)
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("claims").(UserClaims)

	return h.client.Set(ctx,
		fmt.Sprintf("users:ssid:%s", uc.Ssid),
		"", h.rcExpiration).Err()
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			// 1 分钟过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30 * 1000)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, err := token.SignedString(h.JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) ParseWithClaims(tokenStr string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return h.JWTKey, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token is invalid")
	}
	return nil
}

func (h *RedisJWTHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := token.SignedString(h.RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func RedisJWTWithSigningMethod(signingMethod jwt.SigningMethod) RedisJWTOption {
	return func(ctx *RedisJWTHandler) {
		ctx.signingMethod = signingMethod
	}
}
