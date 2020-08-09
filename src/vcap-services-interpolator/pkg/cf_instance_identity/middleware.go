package cf_instance_identity

import (
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/gin-gonic/gin"
)

func Middleware(logger lager.Logger) gin.HandlerFunc {
	logger = logger.Session("cf-instance-identity-middleware")
	return func(c *gin.Context) {
		if len(c.Request.TLS.VerifiedChains) != 1 {
			logger.Error("err-wrong-number-of-verified-chains", nil, lager.Data{
				"number-of-verified-chains": len(c.Request.TLS.VerifiedChains),
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "error with certificate chains",
			})
			return
		}

		instanceIdentityCert := c.Request.TLS.VerifiedChains[0][0]
		cfIID, err := ParseCfIID(instanceIdentityCert)
		if err != nil {
			logger.Error("err-parsing-instance-identity-from-cert", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "error with certificate chains",
			})
			return
		}

		logger.Info("successfully authenticated instance identity", lager.Data{
			"cf-instance-identity": cfIID,
		})
		c.Set("cf-instance-identity", cfIID)

		c.Next()
	}
}
