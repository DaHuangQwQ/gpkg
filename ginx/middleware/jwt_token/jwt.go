package jwt_token

import (
	ijwt "github.com/DaHuangQwQ/gpkg/ginx/jwt"
	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Builder struct {
	publicPaths set.Set[string]

	ijwt.Handler
}

func NewBuilder(handler ijwt.Handler) *Builder {
	return &Builder{
		publicPaths: set.NewMapSet[string](16),
		Handler:     handler,
	}
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if b.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}

		tokenStr := b.ExtractToken(ctx)
		uc := ijwt.UserClaims{}
		err := b.ParseWithClaims(tokenStr, &uc)
		if err != nil {
			// 不正确的 token
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expireTime, err := uc.GetExpirationTime()
		if err != nil {
			// 拿不到过期时间
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if expireTime.Before(time.Now()) {
			// 已经过期
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 浏览器指纹
		//if ctx.GetHeader("User-Agent") != uc.UserAgent {
		//	// 换了一个 User-Agent，可能是攻击者
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		err = b.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// 已经推出登入
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 链路传递
		ctx.Set("claims", uc)
	}
}

func (b *Builder) IgnorePaths(path ...string) *Builder {
	for _, p := range path {
		b.publicPaths.Add(p)
	}
	return b
}
