package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/46bit/vcap-services-interpolator/src/vcap-services-interpolator/pkg/cf_instance_identity"

	"code.cloudfoundry.org/lager"
	"github.com/gin-gonic/gin"
)

type VcapServices = map[string]interface{}
type VcapService = []interface{}
type VcapServiceInstance = map[string]interface{}
type VcapServiceInstanceCredentials = map[string]interface{}

type Interpolator interface {
	Interpolate(_ cf_instance_identity.CfInstanceIdentity, _ DetailsToInterpolate) error
}

type DetailsToInterpolate struct {
	ServiceName     string
	ServiceInstance VcapServiceInstance
	Credentials     VcapServiceInstanceCredentials
	CredhubRef      string
}

func InterpolationEndpoint(interpolator Interpolator, logger lager.Logger) gin.HandlerFunc {
	logger = logger.Session("interpolation-endpoint")

	return func(c *gin.Context) {
		cfIID := c.MustGet("cf-instance-identity").(*cf_instance_identity.CfInstanceIdentity)

		// FIXME: This can actually be a Gin-style bind
		vcapServices, err := parseRequestBodyAsVcapServices(c.Request)
		if err != nil {
			logger.Error("err-could-not-get-vcap-services-from-request-body", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "error decoding request body",
			})
			return
		}

		err = applyInterpolationToVcapServices(vcapServices, interpolator, cfIID)
		if err != nil {
			logger.Error("err-could-apply-interpolation", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "error applying interpolation",
			})
			return
		}

		c.JSON(http.StatusOK, vcapServices)
	}
}

func parseRequestBodyAsVcapServices(request *http.Request) (VcapServices, error) {
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	vcapServices := VcapServices{}
	err = json.Unmarshal(requestBody, &vcapServices)
	if err != nil {
		return nil, fmt.Errorf("error decoding request body as VCAP_SERVICES: %v", err)
	}
	return vcapServices, nil
}

func applyInterpolationToVcapServices(
	vcapServices VcapServices,
	interpolator Interpolator,
	cfIID *cf_instance_identity.CfInstanceIdentity,
) error {
	// This length may seem insane, but it's important to not erase anything
	// unexpected in VCAP_SERVICES. Hopefully this can be relaxed with more study.
	for serviceName, value := range vcapServices {
		service, ok := value.(VcapService)
		if !ok {
			continue
		}

		for _, serviceInstanceValue := range service {
			serviceInstance, ok := serviceInstanceValue.(VcapServiceInstance)
			if !ok {
				continue
			}

			if _, ok = serviceInstance["credentials"]; !ok {
				continue
			}
			credentials, ok := serviceInstance["credentials"].(VcapServiceInstanceCredentials)
			if !ok {
				continue
			}

			credhubRef, ok := credentials["credhub-ref"]
			if !ok {
				credhubRef, ok = credentials["credhub_ref"]
				if !ok {
					continue
				}
			}

			credhubRefString, ok := credhubRef.(string)
			if !ok {
				return fmt.Errorf("credhub ref was not a string in service: '%s'", serviceName)
			}

			detailsToInterpolate := DetailsToInterpolate{
				ServiceName:     serviceName,
				ServiceInstance: serviceInstance,
				Credentials:     credentials,
				CredhubRef:      credhubRefString,
			}

			err := interpolator.Interpolate(*cfIID, detailsToInterpolate)
			if err != nil {
				return err
			}
		}
		vcapServices[serviceName] = service
	}
	return nil
}
