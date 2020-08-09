package main

import (
	//"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	//"os/signal"
	//"sync"
	//"syscall"

	"github.com/46bit/vcap-services-interpolator/src/vcap-services-interpolator/pkg/cf_instance_identity"

	"code.cloudfoundry.org/lager"
	"github.com/gin-gonic/gin"
)

func main() {
	// ctx, shutdown := context.WithCancel(context.Background())
	// go func() {
	// 	sigChan := make(chan os.Signal, 1)
	// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// 	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	// 	<-sigChan
	// 	shutdown()
	// }()
	// var wg sync.WaitGroup

	cfg := NewConfigFromEnv()

	mikiWasHereInterpolator := &MikiWasHereInterpolator{
		logger: cfg.Logger.Session("miki-was-here-interpolator"),
	}

	router := gin.Default()
	router.Use(cf_instance_identity.Middleware(cfg.Logger))
	// FIXME: Health endpoint will need to be separate to mutual TLS
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "online",
		})
	})
	router.POST("/api/v1/interpolate", InterpolationEndpoint(mikiWasHereInterpolator, cfg.Logger))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ListenPort),
		Handler: router,
	}

	instanceIdentityCA, err := ioutil.ReadFile(cfg.InstanceIdentityCAPath)
	if err != nil {
		cfg.Logger.Fatal("err-instance-identity-ca-unreadable", err)
	}
	cfIIDTlsServer := cf_instance_identity.NewServer(cfg.ServerCertPath, cfg.ServerKeyPath, instanceIdentityCA)

	//wg.Add(1)
	//go func() {
	err = cfIIDTlsServer.ServeWithMutualTLS(httpServer)
	if err != nil {
		cfg.Logger.Error("err-fatal-server", err)
	}
	//shutdown()
	os.Exit(1)
	//}()

	//wg.Wait()
}

type MikiWasHereInterpolator struct {
	logger lager.Logger
}

func (i *MikiWasHereInterpolator) Interpolate(
	cfIID cf_instance_identity.CfInstanceIdentity,
	detailsToInterpolate DetailsToInterpolate,
) error {
	if len(detailsToInterpolate.CredhubRef) == 0 {
		return fmt.Errorf("credhub ref was empty")
	}
	i.logger.Info("interpolated-service", lager.Data{
		"cf-instance-identity": cfIID.String(),
		"service-name":         detailsToInterpolate.ServiceName,
		"credhub-ref":          detailsToInterpolate.CredhubRef,
	})
	detailsToInterpolate.Credentials["miki"] = fmt.Sprintf(
		"hi from miki, app %s at %v",
		cfIID.AppGuid,
		cfIID.AppInstanceIP,
	)
	detailsToInterpolate.Credentials["username"] = "username-would-go-here"
	detailsToInterpolate.Credentials["password"] = "password-would-go-here"
	detailsToInterpolate.Credentials["name-of-another-secret"] = "another-secret-would-go-here"
	return nil
}
