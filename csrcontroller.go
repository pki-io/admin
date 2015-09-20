package main

import (
	"crypto/x509/pkix"
	"fmt"
	"github.com/pki-io/core/crypto"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/fs"
	"github.com/pki-io/core/x509"
)

type CSRParams struct {
	name           *string
	tags           *string
	standaloneFile *string
	expiry         *int
	ca             *string
	keyType        *string
	dnLocality     *string
	dnState        *string
	dnOrg          *string
	dnOrgUnit      *string
	dnCountry      *string
	dnStreet       *string
	dnPostal       *string
	confirmDelete  *string
	export         *string
	private        *bool
	csrFile        *string
	keyFile        *string
}

func NewCSRParams() *CSRParams {
	return new(CSRParams)
}

func (params *CSRParams) ValidateName(required bool) error {
	if required && *params.name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

func (params *CSRParams) ValidateStandalone(required bool) error    { return nil }
func (params *CSRParams) ValidateTags(required bool) error          { return nil }
func (params *CSRParams) ValidateExpiry(required bool) error        { return nil }
func (params *CSRParams) ValidateKeyType(required bool) error       { return nil }
func (params *CSRParams) ValidateDnLocality(required bool) error    { return nil }
func (params *CSRParams) ValidateDnState(required bool) error       { return nil }
func (params *CSRParams) ValidateDnOrg(required bool) error         { return nil }
func (params *CSRParams) ValidateDnOrgUnit(required bool) error     { return nil }
func (params *CSRParams) ValidateDnCountry(required bool) error     { return nil }
func (params *CSRParams) ValidateDnStreet(required bool) error      { return nil }
func (params *CSRParams) ValidateDnPostal(required bool) error      { return nil }
func (params *CSRParams) ValidateConfirmDelete(required bool) error { return nil }
func (params *CSRParams) ValidateExport(required bool) error        { return nil }
func (params *CSRParams) ValidatePrivate(required bool) error       { return nil }
func (params *CSRParams) ValidateCSRFile(required bool) error       { return nil }
func (params *CSRParams) ValidateKeyFile(required bool) error       { return nil }

type CSRController struct {
	env *Environment
}

func NewCSRController(env *Environment) (*CSRController, error) {
	cont := new(CSRController)
	cont.env = env
	return cont, nil
}

func (cont *CSRController) GetCA(caId string) (*x509.CA, error) {
	caCont, err := NewCAController(cont.env)
	if err != nil {
		return nil, err
	}

	return caCont.GetCA(caId)
}

func (cont *CSRController) ResetCSRTags(csrId, tags string) error {
	orgIndex, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	if err := orgIndex.ClearCSRTags(csrId); err != nil {
		return err
	}

	err = orgIndex.AddCSRTags(csrId, ParseTags(tags))
	if err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(orgIndex)
	if err != nil {
		return err
	}

	return nil
}

func (cont *CSRController) GetCSR(id string) (*x509.CSR, error) {
	cont.env.logger.Debugf("Getting CSR '%s'", id)

	csrContainerJson, err := cont.env.api.GetPrivate(cont.env.controllers.org.org.Data.Body.Id, id)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Creating new container")

	csrContainer, err := document.NewContainer(csrContainerJson)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Decrypting container")

	csrJson, err := cont.env.controllers.org.org.VerifyThenDecrypt(csrContainer)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Loading CSR json")

	csr, err := x509.NewCSR(csrJson)
	if err != nil {
		return nil, err
	}

	return csr, nil
}

func (cont *CSRController) SaveCSR(csr *x509.CSR) error {
	cont.env.logger.Debug("Encrypting csr for org")
	csrContainer, err := cont.env.controllers.org.org.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Saving encrypted csr")
	err = cont.env.api.SendPrivate(cont.env.controllers.org.org.Data.Body.Id, csr.Data.Body.Id, csrContainer.Dump())
	if err != nil {
		return err
	}

	return nil
}

func (cont *CSRController) AddCSRToOrgIndex(csr *x509.CSR, tags string) error {
	orgIndex, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	err = orgIndex.AddCSR(csr.Data.Body.Name, csr.Data.Body.Id)
	if err != nil {
		return err
	}

	err = orgIndex.AddCSRTags(csr.Data.Body.Id, ParseTags(tags))
	if err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(orgIndex)
	if err != nil {
		return err
	}

	return nil
}

func (cont *CSRController) New(params *CSRParams) error {
	cont.env.logger.Info("Creating new CSR")

	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	// TODO - This should really be in a CSR function
	subject := pkix.Name{CommonName: *params.name}

	if *params.dnLocality != "" {
		subject.Locality = []string{*params.dnLocality}
	}
	if *params.dnState != "" {
		subject.Province = []string{*params.dnState}
	}
	if *params.dnOrg != "" {
		subject.Organization = []string{*params.dnOrg}
	}
	if *params.dnOrgUnit != "" {
		subject.OrganizationalUnit = []string{*params.dnOrgUnit}
	}
	if *params.dnCountry != "" {
		subject.Country = []string{*params.dnCountry}
	}
	if *params.dnStreet != "" {
		subject.StreetAddress = []string{*params.dnStreet}
	}
	if *params.dnPostal != "" {
		subject.PostalCode = []string{*params.dnPostal}
	}

	cont.env.logger.Debug("Creating CSR struct")

	csr, err := x509.NewCSR(nil)
	if err != nil {
		return err
	}

	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = *params.name

	var files []ExportFile
	csrFile := fmt.Sprintf("%s-csr.pem", csr.Data.Body.Name)
	keyFile := fmt.Sprintf("%s-key.pem", csr.Data.Body.Name)

	if *params.csrFile == "" && *params.keyFile == "" {
		cont.env.logger.Debug("Generating a new CSR")
		csr.Generate(&subject)
	} else {
		if *params.csrFile == "" {
			return fmt.Errorf("CSR PEM file must be provided if importing")
		}

		cont.env.logger.Debug("Importing CSR")

		ok, err := fs.Exists(*params.csrFile)
		if err != nil {
			return err
		}

		if !ok {
			cont.env.logger.Warnf("CSR file '%s' does not exist", *params.csrFile)
			return nil
		}

		csrPem, err := fs.ReadFile(*params.csrFile)
		if err != nil {
			return err
		}

		_, err = x509.PemDecodeX509CSR([]byte(csrPem))
		if err != nil {
			return err
		}

		csr.Data.Body.CSR = csrPem

		if *params.keyFile != "" {
			ok, err := fs.Exists(*params.keyFile)
			if err != nil {
				return err
			}

			if !ok {
				cont.env.logger.Warnf("Key file '%s' does not exist", *params.keyFile)
				return nil
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

			csr.Data.Body.KeyType = string(keyType)
			csr.Data.Body.PrivateKey = keyPem
		}
	}

	files = append(files, ExportFile{Name: csrFile, Mode: 0644, Content: []byte(csr.Data.Body.CSR)})
	files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(csr.Data.Body.PrivateKey)})

	if *params.standaloneFile == "" {
		cont.env.logger.Debug("Saving CSR")
		err = cont.SaveCSR(csr)
		if err != nil {
			return err
		}

		cont.env.logger.Debug("Adding csr to org")
		var tags string
		if *params.tags == "NAME" {
			tags = *params.name
		} else {
			tags = *params.tags
		}

		err = cont.AddCSRToOrgIndex(csr, tags)
		if err != nil {
			return err
		}
	} else {
		cont.env.logger.Infof("Exporting to '%s'", *params.standaloneFile)
		Export(files, *params.standaloneFile)
	}

	return nil
}

func (cont *CSRController) List(params *CSRParams) error {
	cont.env.logger.Info("Listing CSRs")

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
	for name, id := range index.GetCSRs() {
		fmt.Printf("* %s %s\n", name, id)
	}

	return nil
}

func (cont *CSRController) Show(params *CSRParams) error {
	cont.env.logger.Info("Showing CSR")

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

	cont.env.logger.Infof("Showing CSR '%s'", *params.name)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	csrId, err := index.GetCSR(*params.name)
	if err != nil {
		return err
	}

	csr, err := cont.GetCSR(csrId)
	if err != nil {
		return err
	}

	if *params.export == "" {
		cont.env.logger.Flush()
		fmt.Printf("Name: %s\n", csr.Data.Body.Name)
		fmt.Printf("ID: %s\n", csr.Data.Body.Id)
		fmt.Printf("Key type: %s\n", csr.Data.Body.KeyType)
		fmt.Printf("CSR:\n%s\n", csr.Data.Body.CSR)

		if *params.private {
			fmt.Printf("Private key:\n%s\n", csr.Data.Body.PrivateKey)
		}
	} else {
		var files []ExportFile
		csrFile := fmt.Sprintf("%s-csr.pem", csr.Data.Body.Name)
		keyFile := fmt.Sprintf("%s-key.pem", csr.Data.Body.Name)

		files = append(files, ExportFile{Name: csrFile, Mode: 0644, Content: []byte(csr.Data.Body.CSR)})

		if *params.private {
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(csr.Data.Body.PrivateKey)})
		}

		cont.env.logger.Infof("Exporting to '%s'", *params.export)
		Export(files, *params.export)
	}

	return nil
}

func (cont *CSRController) Sign(params *CSRParams) error {

	cont.env.logger.Debug("Validating parameters")
	if err := params.ValidateName(true); err != nil {
		return err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	csrId, err := index.GetCSR(*params.name)
	if err != nil {
		return err
	}

	csr, err := cont.GetCSR(csrId)
	if err != nil {
		return err
	}

	caId, err := index.GetCA(*params.ca)
	if err != nil {
		return err
	}

	caCont, err := NewCAController(cont.env)
	if err != nil {
		return err
	}

	ca, err := caCont.GetCA(caId)
	if err != nil {
		return err
	}

	cert, err := ca.Sign(csr)

	org := cont.env.controllers.org.org
	certContainer, err := org.EncryptThenSignString(cert.Dump(), nil)
	if err != nil {
		return err
	}

	if err := cont.env.api.SendPrivate(org.Data.Body.Id, cert.Data.Body.Id, certContainer.Dump()); err != nil {
		return err
	}

	index.AddCert(cert.Data.Body.Name, cert.Data.Body.Id)
	index.AddCertTags(cert.Data.Body.Id, ParseTags(*params.tags))

	if err := cont.env.controllers.org.SaveIndex(index); err != nil {
		return err
	}

	return nil
}
func (cont *CSRController) Update(params *CSRParams) error {
	cont.env.logger.Info("Updating CSR")

	cont.env.logger.Debug("Validating parameters")
	if err := params.ValidateName(true); err != nil {
		return err
	}

	cont.env.logger.Infof("Updating CSR '%s'", *params.name)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	csrId, err := index.GetCSR(*params.name)
	if err != nil {
		return err
	}

	csr, err := cont.GetCSR(csrId)
	if err != nil {
		return err
	}

	if *params.csrFile != "" {
		ok, err := fs.Exists(*params.csrFile)
		if err != nil {
			return err
		}
		if !ok {
			cont.env.logger.Warnf("CSR file '%s' does not exist", *params.csrFile)
			return nil
		}

		cont.env.logger.Debugf("Reading CSR file '%s'", *params.csrFile)

		csrPem, err := fs.ReadFile(*params.csrFile)
		if err != nil {
			return err
		}

		// TODO - better validation of pem
		cont.env.logger.Debug("Decoding CSR file PEM")
		_, err = x509.PemDecodeX509CSR([]byte(csrPem))
		if err != nil {
			return err
		}

		cont.env.logger.Debug("Setting CSR PEM")
		csr.Data.Body.CSR = csrPem
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
		csr.Data.Body.KeyType = string(keyType)

		cont.env.logger.Debug("Setting key PEM")
		csr.Data.Body.PrivateKey = keyPem
	}

	if *params.tags != "" {
		cont.ResetCSRTags(csrId, *params.tags)
	}

	cont.env.logger.Debug("Saving CSR")

	err = cont.SaveCSR(csr)
	if err != nil {
		return err
	}

	return nil
}

func (cont *CSRController) Delete(params *CSRParams) error {
	cont.env.logger.Info("Deleting CSR")

	cont.env.logger.Debug("Validating parameters")
	if err := params.ValidateName(true); err != nil {
		return err
	}

	if err := params.ValidateConfirmDelete(true); err != nil {
		return err
	}

	cont.env.logger.Infof("Deleting CSR '%s' with reason '%s'", *params.name, *params.confirmDelete)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	csrId, err := index.GetCSR(*params.name)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Removing CSR file")
	if err := cont.env.api.DeletePrivate(cont.env.controllers.org.org.Data.Body.Id, csrId); err != nil {
		return err
	}

	cont.env.logger.Debug("Removing csr from org index")
	if err := index.RemoveCSR(*params.name); err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(index)
	if err != nil {
		return err
	}

	return nil

}
