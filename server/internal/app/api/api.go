package api

import (
	"github.com/gin-gonic/gin"
)

type Service interface {
	AuthRequired(*gin.Context)
	ShamirRequired(*gin.Context)

	SignUp(*gin.Context)
	SignIn(*gin.Context)

	Create(*gin.Context)
	Get(*gin.Context)
	Delete(*gin.Context)
	Update(*gin.Context)

	ListSecrets(*gin.Context)
	ListSecretsRecursively(*gin.Context)

	Unseal(*gin.Context)
	UnsealComplete(*gin.Context)
	Master(*gin.Context)
	MasterComplete(*gin.Context)

	IsReady(*gin.Context)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func New(service Service) *gin.Engine {
	r := gin.Default()

	r.Use()

	r.Use(CORSMiddleware())

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	apiGroup := r.Group("/api")
	{
		apiGroup.Any("/ready", service.IsReady) // like healhcheck for tests

		apiGroup.POST("/unseal", service.Unseal)
		apiGroup.POST("/unseal/complete", service.UnsealComplete)
		apiGroup.GET("/master", service.Master)
		apiGroup.GET("/master/complete", service.MasterComplete)

		apiGroup.Use(service.ShamirRequired)

		apiGroup.POST("/signup", service.SignUp)
		apiGroup.POST("/signin", service.SignIn)

		authorized := apiGroup.Group("/", service.AuthRequired)
		{
			sercretManage := authorized.Group("/secrets") // query param /api/secrets?path=lvl1/lvl2/key_of_secret
			{
				sercretManage.POST("/", service.Create)
				sercretManage.GET("/:key", service.Get)
				sercretManage.DELETE("/:key", service.Delete)
				sercretManage.PATCH("/:key", service.Update)
			}

			authorized.GET("/list", service.ListSecrets)
			authorized.GET("/reclist", service.ListSecretsRecursively)
		}
	}

	return r
}
