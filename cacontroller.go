// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/pki-io/core/crypto"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/fs"
	"github.com/pki-io/core/x509"
	"time"
)

// First-class types only
type CAParams struct {
	name          *string
	tags          *string
	caExpiry      *int
	certExpiry    *int
	keyType       *string
	dnLocality    *string
	dnState       *string
	dnOrg         *string
	dnOrgUnit     *string
	dnCountry     *string
	dnStreet      *string
	dnPostal      *string
	confirmDelete *string
	export        *string
	private       *bool
	certFile      *string
	keyFile       *string
}

func NewCAParams() *CAParams {
	return new(CAParams)
}

// ThreatSpec TMv0.1 for CAParams.ValidateName
// Does name parameter validation for App:CAController

func (params *CAParams) ValidateName(required bool) error {
	logger.Tracef("validating name '%s'", *params.name)
	if required && *params.name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

func (params *CAParams) ValidateTags(required bool) error          { return nil }
func (params *CAParams) ValidateCAExpiry(required bool) error      { return nil }
func (params *CAParams) ValidateCertExpiry(required bool) error    { return nil }
func (params *CAParams) ValidateKeyType(required bool) error       { return nil }
func (params *CAParams) ValidateDnLocality(required bool) error    { return nil }
func (params *CAParams) ValidateDnState(required bool) error       { return nil }
func (params *CAParams) ValidateDnOrg(required bool) error         { return nil }
func (params *CAParams) ValidateDnOrgUnit(required bool) error     { return nil }
func (params *CAParams) ValidateDnCountry(required bool) error     { return nil }
func (params *CAParams) ValidateDnStreet(required bool) error      { return nil }
func (params *CAParams) ValidateDnPostal(required bool) error      { return nil }
func (params *CAParams) ValidateConfirmDelete(required bool) error { return nil }
func (params *CAParams) ValidateExport(required bool) error        { return nil }
func (params *CAParams) ValidatePrivate(required bool) error       { return nil }
func (params *CAParams) ValidateCertFile(required bool) error      { return nil }
func (params *CAParams) ValidateKeyFile(required bool) error       { return nil }

type CAController struct {
	env *Environment
}

// ThreatSpec TMv0.1 for NewCAController
// Creates new CA controller for App:CAController

func NewCAController(env *Environment) (*CAController, error) {
	cont := new(CAController)
	cont.env = env
	return cont, nil
}

func (cont *CAController) GetCA(id string) (*x509.CA, error) {
	logger.Debugf("getting CA")
	logger.Tracef("received id '%s", id)

	logger.Debugf("getting CA json '%s' for org '%s'", id, cont.env.controllers.org.OrgId())
	caContainerJson, err := cont.env.api.GetPrivate(cont.env.controllers.org.OrgId(), id)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating new container from json")
	caContainer, err := document.NewContainer(caContainerJson)
	if err != nil {
		return nil, err
	}

	logger.Debug("decrypting container")
	caJson, err := cont.env.controllers.org.org.VerifyThenDecrypt(caContainer)
	if err != nil {
		return nil, err
	}

	logger.Debug("loading CA json to struct")
	ca, err := x509.NewCA(caJson)
	if err != nil {
		return nil, err
	}

	logger.Trace("returning CA")
	return ca, nil
}

func (cont *CAController) SaveCA(ca *x509.CA) error {
	logger.Debug("saving CA")
	logger.Trace("received CA [NOT LOGGED]")

	logger.Debug("encrypting CA for org")
	caContainer, err := cont.env.controllers.org.org.EncryptThenSignString(ca.Dump(), nil)
	if err != nil {
		return err
	}

	logger.Debug("saving encrypted CA")
	err = cont.env.api.SendPrivate(cont.env.controllers.org.org.Data.Body.Id, ca.Data.Body.Id, caContainer.Dump())
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *CAController) ResetCATags(caId, tags string) error {
	logger.Debug("resetting CA tags")
	logger.Tracef("received caId '%s' and tags '%s", caId, tags)

	orgIndex, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	if err := orgIndex.ClearCATags(caId); err != nil {
		return err
	}

	err = orgIndex.AddCATags(caId, ParseTags(tags))
	if err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(orgIndex)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *CAController) AddCAToOrgIndex(ca *x509.CA, tags string) error {
	logger.Debug("Adding CA to org index")
	logger.Tracef("received ca [NOT LOGGED] with tags '%s'", tags)

	orgIndex, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	err = orgIndex.AddCA(ca.Data.Body.Name, ca.Data.Body.Id)
	if err != nil {
		return err
	}

	err = orgIndex.AddCATags(ca.Data.Body.Id, ParseTags(tags))
	if err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(orgIndex)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *CAController) RemoveCAFromOrgIndex(name string) error {
	logger.Debug("removing CA from org index")
	logger.Tracef("received name '%s'", name)

	orgIndex, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	if err := orgIndex.RemoveCA(name); err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(orgIndex)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

// ThreatSpec TMv0.1 for CAController.New
// Creates new CA for App:CAController

func (cont *CAController) New(params *CAParams) (*x509.CA, error) {
	logger.Debug("creating new CA")
	logger.Trace("received params [NOT LOGGED]")

	if err := params.ValidateName(true); err != nil {
		return nil, err
	}

	if err := params.ValidateCAExpiry(true); err != nil {
		return nil, err
	}

	if err := params.ValidateCertExpiry(true); err != nil {
		return nil, err
	}

	if err := params.ValidateKeyType(true); err != nil {
		return nil, err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	logger.Debug("creating CA struct")
	ca, err := x509.NewCA(nil)
	if err != nil {
		return nil, err
	}

	ca.Data.Body.Name = *params.name
	ca.Data.Body.CAExpiry = *params.caExpiry
	ca.Data.Body.CertExpiry = *params.certExpiry
	ca.Data.Body.KeyType = *params.keyType
	ca.Data.Body.DNScope.Locality = *params.dnLocality
	ca.Data.Body.DNScope.Province = *params.dnState
	ca.Data.Body.DNScope.Organization = *params.dnOrg
	ca.Data.Body.DNScope.OrganizationalUnit = *params.dnOrgUnit
	ca.Data.Body.DNScope.Country = *params.dnCountry
	ca.Data.Body.DNScope.StreetAddress = *params.dnStreet
	ca.Data.Body.DNScope.PostalCode = *params.dnPostal

	if *params.certFile == "" && *params.keyFile == "" {
		logger.Debug("generating keys")
		ca.GenerateRoot()
	} else {
		if *params.certFile == "" {
			return nil, fmt.Errorf("certificate PEM file must be provided if importing")
		}

		ok, err := fs.Exists(*params.certFile)
		if err != nil {
			return nil, err
		}

		if !ok {
			logger.Warnf("certificate file '%s' does not exist", *params.certFile)
			return nil, nil
		}

		logger.Debugf("reading certificate PEM file '%s", *params.certFile)
		certPem, err := fs.ReadFile(*params.certFile)
		if err != nil {
			return nil, err
		}

		logger.Debug("decoding certificate PEM")
		cert, err := x509.PemDecodeX509Certificate([]byte(certPem))
		if err != nil {
			return nil, err
		}

		ca.Data.Body.Id = NewID()
		ca.Data.Body.Certificate = certPem
		ca.Data.Body.CertExpiry = *params.certExpiry
		caExpiry := int(cert.NotAfter.Sub(cert.NotBefore) / (time.Hour * 24))
		ca.Data.Body.CAExpiry = caExpiry

		if *params.keyFile != "" {
			ok, err = fs.Exists(*params.keyFile)
			if err != nil {
				return nil, err
			}

			if !ok {
				logger.Warnf("key file '%s' does not exist", *params.keyFile)
				return nil, nil
			}

			logger.Debugf("reading private key PEM file '%s'", *params.keyFile)
			keyPem, err := fs.ReadFile(*params.keyFile)
			if err != nil {
				return nil, err
			}

			logger.Debug("decoding private key")
			key, err := crypto.PemDecodePrivate([]byte(keyPem))
			if err != nil {
				return nil, err
			}

			logger.Debug("getting key type")
			keyType, err := crypto.GetKeyType(key)
			if err != nil {
				return nil, err
			}

			ca.Data.Body.KeyType = string(keyType)
			ca.Data.Body.PrivateKey = keyPem
		}
	}

	err = cont.SaveCA(ca)
	if err != nil {
		return nil, err
	}

	err = cont.AddCAToOrgIndex(ca, *params.tags)
	if err != nil {
		return nil, err
	}

	logger.Trace("returning CA")
	return ca, nil
}

func (cont *CAController) List(params *CAParams) ([]*x509.CA, error) {
	logger.Debug("listing CAs")
	logger.Trace("received params [NOT LOGGED]")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return nil, err
	}

	cas := make([]*x509.CA, 0)
	for _, id := range index.GetCAs() {
		ca, err := cont.GetCA(id)
		if err != nil {
			return nil, err
		}
		cas = append(cas, ca)
	}

	logger.Trace("returning CA list")
	return cas, nil
}

func (cont *CAController) Show(params *CAParams) (*x509.CA, error) {
	logger.Debug("showing CA")
	logger.Trace("received params [NOT LOGGED]")

	if err := params.ValidateName(true); err != nil {
		return nil, err
	}

	if err := params.ValidateExport(false); err != nil {
		return nil, err
	}

	if err := params.ValidatePrivate(false); err != nil {
		return nil, err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return nil, err
	}

	caId, err := index.GetCA(*params.name)
	if err != nil {
		return nil, err
	}

	ca, err := cont.GetCA(caId)
	if err != nil {
		return nil, err
	}

	if *params.export == "" {
		logger.Trace("returning CA")
		return ca, nil
	} else {
		var files []ExportFile
		certFile := fmt.Sprintf("%s-cert.pem", ca.Data.Body.Name)
		keyFile := fmt.Sprintf("%s-key.pem", ca.Data.Body.Name)

		files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(ca.Data.Body.Certificate)})

		if *params.private {
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(ca.Data.Body.PrivateKey)})
		}

		Export(files, *params.export)
	}

	return nil, nil
}

func (cont *CAController) Update(params *CAParams) error {
	logger.Debug("updating CA")
	logger.Trace("received params [NOT LOGGED]")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	caId, err := index.GetCA(*params.name)
	if err != nil {
		return err
	}

	ca, err := cont.GetCA(caId)
	if err != nil {
		return err
	}

	if *params.certFile != "" {
		ok, err := fs.Exists(*params.certFile)
		if err != nil {
			return err
		}
		if !ok {
			logger.Warnf("certificate file '%s' does not exist", *params.certFile)
			return nil
		}

		logger.Debugf("reading certificate file '%s'", *params.certFile)
		certPem, err := fs.ReadFile(*params.certFile)
		if err != nil {
			return err
		}

		// TODO - better validation of pem
		logger.Debug("decoding certificate file PEM")
		_, err = x509.PemDecodeX509Certificate([]byte(certPem))
		if err != nil {
			return err
		}

		logger.Trace("setting certificate")
		ca.Data.Body.Certificate = certPem
	}

	if *params.keyFile != "" {
		ok, err := fs.Exists(*params.keyFile)
		if err != nil {
			return err
		}
		if !ok {
			logger.Warnf("key file '%s' does not exist", *params.keyFile)
			return nil
		}

		logger.Debugf("seading key file '%s'", *params.keyFile)
		keyPem, err := fs.ReadFile(*params.keyFile)
		if err != nil {
			return err
		}

		// TODO - better validation of pem
		logger.Debug("decoding key file PEM")
		key, err := crypto.PemDecodePrivate([]byte(keyPem))
		if err != nil {
			return err
		}

		logger.Debug("getting key type")
		keyType, err := crypto.GetKeyType(key)
		if err != nil {
			return err
		}

		ca.Data.Body.KeyType = string(keyType)
		ca.Data.Body.PrivateKey = keyPem
	}

	if *params.tags != "" {
		cont.ResetCATags(caId, *params.tags)
	}

	if *params.caExpiry != 0 {
		logger.Tracef("setting CA expiry to %d", *params.caExpiry)
		ca.Data.Body.CAExpiry = *params.caExpiry
	}

	if *params.certExpiry != 0 {
		logger.Tracef("setting certificate expiry to %d", *params.certExpiry)
		ca.Data.Body.CertExpiry = *params.certExpiry
	}

	if *params.dnLocality != "" {
		logger.Tracef("setting DN locality to %s", *params.dnLocality)
		ca.Data.Body.DNScope.Locality = *params.dnLocality
	}

	if *params.dnState != "" {
		logger.Tracef("setting DN state to %s", *params.dnState)
		ca.Data.Body.DNScope.Province = *params.dnState
	}

	if *params.dnOrg != "" {
		logger.Tracef("setting DN organisation to %s", *params.dnOrg)
		ca.Data.Body.DNScope.Organization = *params.dnOrg
	}

	if *params.dnOrgUnit != "" {
		logger.Tracef("setting DN organisational unit to %s", *params.dnOrgUnit)
		ca.Data.Body.DNScope.OrganizationalUnit = *params.dnOrgUnit
	}

	if *params.dnCountry != "" {
		logger.Tracef("setting DN country to %s", *params.dnCountry)
		ca.Data.Body.DNScope.Country = *params.dnCountry
	}

	if *params.dnStreet != "" {
		logger.Tracef("setting DN street address to %s", *params.dnStreet)
		ca.Data.Body.DNScope.StreetAddress = *params.dnStreet
	}

	if *params.dnPostal != "" {
		logger.Tracef("setting DN postal code to %s", *params.dnPostal)
		ca.Data.Body.DNScope.PostalCode = *params.dnPostal
	}

	err = cont.SaveCA(ca)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *CAController) Delete(params *CAParams) error {
	logger.Debug("deleting CA")
	logger.Tracef("received params: %s", params)

	if err := params.ValidateName(true); err != nil {
		return err
	}

	if err := params.ValidateConfirmDelete(true); err != nil {
		return err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	caId, err := index.GetCA(*params.name)
	if err != nil {
		return err
	}

	logger.Debugf("deleting private file for CA '%s' in org '%s'", caId, cont.env.controllers.org.OrgId())
	if err := cont.env.api.DeletePrivate(cont.env.controllers.org.OrgId(), caId); err != nil {
		return err
	}

	if err := index.RemoveCA(*params.name); err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(index)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}
