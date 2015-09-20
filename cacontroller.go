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

func (params *CAParams) ValidateName(required bool) error {
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

func NewCAController(env *Environment) (*CAController, error) {
	cont := new(CAController)
	cont.env = env
	return cont, nil
}

func (cont *CAController) GetCA(id string) (*x509.CA, error) {

	cont.env.logger.Debugf("Getting CA '%s'", id)

	caContainerJson, err := cont.env.api.GetPrivate(cont.env.controllers.org.org.Data.Body.Id, id)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Creating new container")

	caContainer, err := document.NewContainer(caContainerJson)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Decrypting container")

	caJson, err := cont.env.controllers.org.org.VerifyThenDecrypt(caContainer)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Loading CA json")

	ca, err := x509.NewCA(caJson)
	if err != nil {
		return nil, err
	}

	return ca, nil
}

func (cont *CAController) SaveCA(ca *x509.CA) error {

	cont.env.logger.Debug("Encrypting CA for org")
	caContainer, err := cont.env.controllers.org.org.EncryptThenSignString(ca.Dump(), nil)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Saving encrypted CA")
	err = cont.env.api.SendPrivate(cont.env.controllers.org.org.Data.Body.Id, ca.Data.Body.Id, caContainer.Dump())
	if err != nil {
		return err
	}

	return nil
}

func (cont *CAController) ResetCATags(caId, tags string) error {

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

	return nil
}

func (cont *CAController) AddCAToOrgIndex(ca *x509.CA, tags string) error {

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

	return nil
}

func (cont *CAController) RemoveCAFromOrgIndex(name string) error {

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

	return nil
}

func (cont *CAController) New(params *CAParams) error {

	cont.env.logger.Info("Creating new ca")

	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	if err := params.ValidateCAExpiry(true); err != nil {
		return err
	}

	if err := params.ValidateCertExpiry(true); err != nil {
		return err
	}

	if err := params.ValidateKeyType(true); err != nil {
		return err
	}

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Creating CA struct")
	ca, err := x509.NewCA(nil)
	if err != nil {
		return err
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
		cont.env.logger.Debug("Generating keys")
		ca.GenerateRoot()
	} else {
		if *params.certFile == "" {
			return fmt.Errorf("certificate PEM file must be provided if importing")
		}

		ok, err := fs.Exists(*params.certFile)
		if err != nil {
			return err
		}

		if !ok {
			cont.env.logger.Warnf("Certificate file '%s' does not exist", *params.certFile)
			return nil
		}

		certPem, err := fs.ReadFile(*params.certFile)
		if err != nil {
			return err
		}

		cert, err := x509.PemDecodeX509Certificate([]byte(certPem))
		if err != nil {
			return err
		}

		ca.Data.Body.Id = NewID()
		ca.Data.Body.Certificate = certPem
		ca.Data.Body.CertExpiry = *params.certExpiry
		caExpiry := int(cert.NotAfter.Sub(cert.NotBefore) / (time.Hour * 24))
		ca.Data.Body.CAExpiry = caExpiry

		if *params.keyFile != "" {
			ok, err = fs.Exists(*params.keyFile)
			if err != nil {
				return err
			}

			if !ok {
				cont.env.logger.Warnf("Key file '%s' does not exist", *params.keyFile)
			}

			keyPem, err := fs.ReadFile(*params.keyFile)
			if err != nil {
				return err
			}

			key, err := crypto.PemDecodePrivate([]byte(keyPem))
			if err != nil {
				return err
			}

			keyType, err := crypto.GetKeyType(key)
			if err != nil {
				return err
			}

			ca.Data.Body.KeyType = string(keyType)
			ca.Data.Body.PrivateKey = keyPem
		}
	}

	cont.env.logger.Debug("Saving CA")
	err = cont.SaveCA(ca)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Adding CA to org")
	err = cont.AddCAToOrgIndex(ca, *params.tags)
	if err != nil {
		return err
	}

	return nil
}

func (cont *CAController) List(params *CAParams) error {
	cont.env.logger.Info("Listing CAs")

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	cont.env.logger.Flush()
	for name, id := range index.GetCAs() {
		fmt.Printf("* %s %s\n", name, id)
	}

	return nil
}

func (cont *CAController) Show(params *CAParams) error {
	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	if err := params.ValidateExport(false); err != nil {
		return err
	}

	if err := params.ValidatePrivate(false); err != nil {
		return err
	}

	cont.env.logger.Infof("Showing CA '%s'", *params.name)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

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

	if *params.export == "" {
		cont.env.logger.Flush()
		fmt.Printf("Name: %s\n", ca.Data.Body.Name)
		fmt.Printf("ID: %s\n", ca.Data.Body.Id)
		fmt.Printf("CA expiry period: %d\n", ca.Data.Body.CAExpiry)
		fmt.Printf("Cert expiry period: %d\n", ca.Data.Body.CertExpiry)
		fmt.Printf("Key type: %s\n", ca.Data.Body.KeyType)
		fmt.Printf("DN country: %s\n", ca.Data.Body.DNScope.Country)
		fmt.Printf("DN organization: %s\n", ca.Data.Body.DNScope.Organization)
		fmt.Printf("DN organizational unit: %s\n", ca.Data.Body.DNScope.OrganizationalUnit)
		fmt.Printf("DN locality: %s\n", ca.Data.Body.DNScope.Locality)
		fmt.Printf("DN province: %s\n", ca.Data.Body.DNScope.Province)
		fmt.Printf("DN street address: %s\n", ca.Data.Body.DNScope.StreetAddress)
		fmt.Printf("DN postal code: %s\n", ca.Data.Body.DNScope.PostalCode)
		fmt.Printf("Certficate:\n%s\n", ca.Data.Body.Certificate)

		if *params.private {
			fmt.Printf("Private key:\n%s\n", ca.Data.Body.PrivateKey)
		}
	} else {
		var files []ExportFile
		certFile := fmt.Sprintf("%s-cert.pem", ca.Data.Body.Name)
		keyFile := fmt.Sprintf("%s-key.pem", ca.Data.Body.Name)

		files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(ca.Data.Body.Certificate)})

		if *params.private {
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(ca.Data.Body.PrivateKey)})
		}

		cont.env.logger.Infof("Exporting to '%s'", *params.export)
		Export(files, *params.export)
	}

	return nil
}

func (cont *CAController) Update(params *CAParams) error {

	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	cont.env.logger.Infof("Updating CA '%s'", *params.name)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

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
			cont.env.logger.Warnf("Certificate file '%s' does not exist", *params.certFile)
			return nil
		}

		cont.env.logger.Debugf("Reading certificate file '%s'", *params.certFile)

		certPem, err := fs.ReadFile(*params.certFile)
		if err != nil {
			return err
		}

		// TODO - better validation of pem
		cont.env.logger.Debug("Decoding certificate file PEM")
		_, err = x509.PemDecodeX509Certificate([]byte(certPem))
		if err != nil {
			return err
		}

		cont.env.logger.Debug("Setting certificate PEM")
		ca.Data.Body.Certificate = certPem
	}

	if *params.keyFile != "" {
		ok, err := fs.Exists(*params.keyFile)
		if err != nil {
			return err
		}
		if !ok {
			cont.env.logger.Warnf("Key file '%s' does not exist", *params.keyFile)
			return nil
		}

		cont.env.logger.Debugf("Reading key file '%s'", *params.keyFile)

		keyPem, err := fs.ReadFile(*params.keyFile)
		if err != nil {
			return err
		}

		// TODO - better validation of pem
		cont.env.logger.Debug("Decoding key file PEM")
		key, err := crypto.PemDecodePrivate([]byte(keyPem))
		if err != nil {
			return err
		}

		cont.env.logger.Debug("Getting key type")
		keyType, err := crypto.GetKeyType(key)
		if err != nil {
			return err
		}

		cont.env.logger.Debug("Setting key type")
		ca.Data.Body.KeyType = string(keyType)

		cont.env.logger.Debug("Setting key PEM")
		ca.Data.Body.PrivateKey = keyPem
	}

	if *params.tags != "" {
		cont.ResetCATags(caId, *params.tags)
	}

	if *params.caExpiry != 0 {
		ca.Data.Body.CAExpiry = *params.caExpiry
	}

	if *params.certExpiry != 0 {
		ca.Data.Body.CertExpiry = *params.certExpiry
	}

	if *params.dnLocality != "" {
		ca.Data.Body.DNScope.Locality = *params.dnLocality
	}

	if *params.dnState != "" {
		ca.Data.Body.DNScope.Province = *params.dnState
	}

	if *params.dnOrg != "" {
		ca.Data.Body.DNScope.Organization = *params.dnOrg
	}

	if *params.dnOrgUnit != "" {
		ca.Data.Body.DNScope.OrganizationalUnit = *params.dnOrgUnit
	}

	if *params.dnCountry != "" {
		ca.Data.Body.DNScope.Country = *params.dnCountry
	}

	if *params.dnStreet != "" {
		ca.Data.Body.DNScope.StreetAddress = *params.dnStreet
	}

	if *params.dnPostal != "" {
		ca.Data.Body.DNScope.PostalCode = *params.dnPostal
	}
	cont.env.logger.Debug("Saving CA")

	err = cont.SaveCA(ca)
	if err != nil {
		return err
	}
	return nil
}

func (cont *CAController) Delete(params *CAParams) error {
	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	if err := params.ValidateConfirmDelete(true); err != nil {
		return err
	}

	cont.env.logger.Infof("Deleting CA '%s' with reason '%s'", *params.name, *params.confirmDelete)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	caId, err := index.GetCA(*params.name)
	if err != nil {
		return err
	}

	if err := cont.env.api.DeletePrivate(cont.env.controllers.org.org.Data.Body.Id, caId); err != nil {
		return err
	}

	if err := index.RemoveCA(*params.name); err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(index)
	if err != nil {
		return err
	}

	return nil
}
