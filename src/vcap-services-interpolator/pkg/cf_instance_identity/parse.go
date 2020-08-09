package cf_instance_identity

import (
	"crypto/x509"
	"fmt"
	"net"
	"strings"
)

// From https://docs.cloudfoundry.org/devguide/deploy-apps/instance-identity.html:
// * The Common Name property is set to the instance GUID for the given app instance.
// * The certificate contains an IP SAN set to the container IP address for the given
//   app instance.
// * The certificate contains a DNS SAN set to the instance GUID for the given app
//   instance. [This is also the Common Name.]
// * The Organizational Unit property in the certificate’s Subject Distinguished Name
//   contains the values organization:ORG-GUID, space:SPACE-GUID, and app:APP-GUID.
//   The ORG-GUID, SPACE-GUID, and APP-GUID are set to the GUIDs for the organization,
//   space, and app as assigned by Cloud Controller.
type CfInstanceIdentity struct {
	AppInstanceGuid string
	AppInstanceIP   net.IP
	AppGuid         string
	SpaceGuid       string
	OrgGuid         string
}

func ParseCfIID(cert *x509.Certificate) (*CfInstanceIdentity, error) {
	// FIXME: parse the guids strictly
	cfIID := CfInstanceIdentity{
		AppInstanceGuid: cert.Subject.CommonName,
	}

	if len(cert.IPAddresses) != 1 {
		return nil, fmt.Errorf("unexpected number of ip addresses in cert: %d", len(cert.IPAddresses))
	}
	cfIID.AppInstanceIP = cert.IPAddresses[0]

	err := parseCfIIDOrganizationalUnit(cert.Subject.OrganizationalUnit, &cfIID)
	if err != nil {
		return nil, err
	}

	return &cfIID, nil
}

func parseCfIIDOrganizationalUnit(organizationalUnit []string, cfIID *CfInstanceIdentity) error {
	if len(organizationalUnit) != 3 {
		return fmt.Errorf("unexpected number of entries in certificate organizational unit: %d", len(organizationalUnit))
	}

	units := [][]string{}
	for _, unit := range organizationalUnit {
		splitUnit := strings.Split(unit, ":")
		if len(splitUnit) != 2 {
			return fmt.Errorf("unexpected format of OU: '%s'", unit)
		}
		units = append(units, splitUnit)
	}

	if units[0][0] != "organization" {
		return fmt.Errorf("unexpected name of first OU: '%s'", units[0][0])
	}
	cfIID.OrgGuid = units[0][1]
	if units[1][0] != "space" {
		return fmt.Errorf("unexpected name of second OU: '%s'", units[1][0])
	}
	cfIID.SpaceGuid = units[1][1]
	if units[2][0] != "app" {
		return fmt.Errorf("unexpected name of third OU: '%s'", units[2][0])
	}
	cfIID.AppGuid = units[2][1]
	return nil
}

func (cfiid CfInstanceIdentity) String() string {
	return fmt.Sprintf(
		"cf-instance-identity(app instance %s at ip %v of app %s in space %s and org %s)",
		cfiid.AppInstanceGuid,
		cfiid.AppInstanceIP,
		cfiid.AppGuid,
		cfiid.SpaceGuid,
		cfiid.OrgGuid,
	)
}
