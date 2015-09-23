package main

import (
	"crypto/x509/pkix"
	"fmt"
	"github.com/pki-io/core/crypto"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/fs"
	"github.com/pki-io/core/x509"
	"time"
)

type CertParams struct {
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
	certFile       *string
	keyFile        *string
}

func NewCertParams() *CertParams {
	return new(CertParams)
}

func (params *CertParams) ValidateName(required bool) error {
	if required && *params.name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

func (params *CertParams) ValidateStandalone(required bool) error    { return nil }
func (params *CertParams) ValidateTags(required bool) error          { return nil }
func (params *CertParams) ValidateExpiry(required bool) error        { return nil }
func (params *CertParams) ValidateKeyType(required bool) error       { return nil }
func (params *CertParams) ValidateDnLocality(required bool) error    { return nil }
func (params *CertParams) ValidateDnState(required bool) error       { return nil }
func (params *CertParams) ValidateDnOrg(required bool) error         { return nil }
func (params *CertParams) ValidateDnOrgUnit(required bool) error     { return nil }
func (params *CertParams) ValidateDnCountry(required bool) error     { return nil }
func (params *CertParams) ValidateDnStreet(required bool) error      { return nil }
func (params *CertParams) ValidateDnPostal(required bool) error      { return nil }
func (params *CertParams) ValidateConfirmDelete(required bool) error { return nil }
func (params *CertParams) ValidateExport(required bool) error        { return nil }
func (params *CertParams) ValidatePrivate(required bool) error       { return nil }
func (params *CertParams) ValidateCertFile(required bool) error      { return nil }
func (params *CertParams) ValidateKeyFile(required bool) error       { return nil }

type CertController struct {
	env *Environment
}

func NewCertController(env *Environment) (*CertController, error) {
	cont := new(CertController)
	cont.env = env
	return cont, nil
}

func (cont *CertController) GetCA(caId string) (*x509.CA, error) {
	caCont, err := NewCAController(cont.env)
	if err != nil {
		return nil, err
	}

	return caCont.GetCA(caId)
}

func (cont *CertController) ResetCertTags(certId, tags string) error {
	orgIndex, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	if err := orgIndex.ClearCertTags(certId); err != nil {
		return err
	}

	err = orgIndex.AddCertTags(certId, ParseTags(tags))
	if err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(orgIndex)
	if err != nil {
		return err
	}

	return nil
}

func (cont *CertController) GetCert(id string) (*x509.Certificate, error) {
	cont.env.logger.Debugf("Getting certificate '%s'", id)

	certContainerJson, err := cont.env.api.GetPrivate(cont.env.controllers.org.org.Data.Body.Id, id)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Creating new container")

	certContainer, err := document.NewContainer(certContainerJson)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Decrypting container")

	certJson, err := cont.env.controllers.org.org.VerifyThenDecrypt(certContainer)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Loading certificate json")

	cert, err := x509.NewCertificate(certJson)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (cont *CertController) SaveCert(cert *x509.Certificate) error {
	cont.env.logger.Debug("Encrypting cert for org")
	certContainer, err := cont.env.controllers.org.org.EncryptThenSignString(cert.Dump(), nil)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Saving encrypted cert")
	err = cont.env.api.SendPrivate(cont.env.controllers.org.org.Data.Body.Id, cert.Data.Body.Id, certContainer.Dump())
	if err != nil {
		return err
	}

	return nil
}

func (cont *CertController) AddCertToOrgIndex(cert *x509.Certificate, tags string) error {
	orgIndex, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	err = orgIndex.AddCert(cert.Data.Body.Name, cert.Data.Body.Id)
	if err != nil {
		return err
	}

	err = orgIndex.AddCertTags(cert.Data.Body.Id, ParseTags(tags))
	if err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(orgIndex)
	if err != nil {
		return err
	}

	return nil
}

func (cont *CertController) New(params *CertParams) (*x509.Certificate, error) {
	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	// TODO - This should really be in a certificate function
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

	cont.env.logger.Debug("Creating certificate struct")

	cert, err := x509.NewCertificate(nil)
	if err != nil {
		return nil, err
	}

	cert.Data.Body.Name = *params.name
	cert.Data.Body.Expiry = *params.expiry

	var files []ExportFile
	certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
	keyFile := fmt.Sprintf("%s-key.pem", cert.Data.Body.Name)

	if *params.certFile == "" && *params.keyFile == "" {
		cont.env.logger.Debug("Generating a new certificate")
		if *params.ca == "" {
			if err := cert.Generate(nil, &subject); err != nil {
				return nil, err
			}
		} else {
			cont.env.logger.Debug("Getting org index")

			index, err := cont.env.controllers.org.GetIndex()
			if err != nil {
				return nil, err
			}

			cont.env.logger.Debugf("Getting CA '%s' from index", *params.ca)
			caId, err := index.GetCA(*params.ca)
			if err != nil {
				return nil, err
			}

			ca, err := cont.GetCA(caId)
			if err != nil {
				return nil, err
			}

			if err := cert.Generate(ca, &subject); err != nil {
				return nil, err
			}

			caFile := fmt.Sprintf("%s-cacert.pem", cert.Data.Body.Name)
			files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(ca.Data.Body.Certificate)})
		}
	} else {
		if *params.certFile == "" {
			return nil, fmt.Errorf("certificate PEM file must be provided if importing")
		}

		cont.env.logger.Debug("Importing certificate")

		ok, err := fs.Exists(*params.certFile)
		if err != nil {
			return nil, err
		}

		if !ok {
			cont.env.logger.Warnf("Cert file '%s' does not exist", *params.certFile)
			return nil, nil
		}

		certPem, err := fs.ReadFile(*params.certFile)
		if err != nil {
			return nil, err
		}

		importCert, err := x509.PemDecodeX509Certificate([]byte(certPem))
		if err != nil {
			return nil, err
		}

		cert.Data.Body.Id = NewID()
		cert.Data.Body.Certificate = certPem
		certExpiry := int(importCert.NotAfter.Sub(importCert.NotBefore) / (time.Hour * 24))
		cert.Data.Body.Expiry = certExpiry

		if *params.keyFile != "" {
			ok, err := fs.Exists(*params.keyFile)
			if err != nil {
				return nil, err
			}

			if !ok {
				cont.env.logger.Warnf("Key file '%s' does not exist", *params.keyFile)
				return nil, nil
			}

			keyPem, err := fs.ReadFile(*params.keyFile)
			if err != nil {
				return nil, err
			}

			key, err := crypto.PemDecodePrivate([]byte(keyPem))
			if err != nil {
				return nil, err
			}

			keyType, err := crypto.GetKeyType(key)
			if err != nil {
				return nil, err
			}

			cert.Data.Body.KeyType = string(keyType)
			cert.Data.Body.PrivateKey = keyPem
		}
	}

	files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
	files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})

	if *params.standaloneFile == "" {
		cont.env.logger.Debug("Saving certificate")
		err = cont.SaveCert(cert)
		if err != nil {
			return nil, err
		}

		cont.env.logger.Debug("Adding cert to org")
		var tags string
		if *params.tags == "NAME" {
			tags = *params.name
		} else {
			tags = *params.tags
		}

		err = cont.AddCertToOrgIndex(cert, tags)
		if err != nil {
			return nil, err
		}

		return cert, nil

	} else {
		cont.env.logger.Infof("Exporting to '%s'", *params.standaloneFile)
		Export(files, *params.standaloneFile)
	}

	return nil, nil
}

func (cont *CertController) List(params *CertParams) ([]*x509.Certificate, error) {
	cont.env.logger.Info("Listing certificates")

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return nil, err
	}

	certs := make([]*x509.Certificate, 0)
	for _, id := range index.GetCerts() {
		cert, err := cont.GetCert(id)
		if err != nil {
			return nil, err
		}

		certs = append(certs, cert)
	}

	return certs, nil
}

func (cont *CertController) Show(params *CertParams) (*x509.Certificate, error) {
	cont.env.logger.Info("Showing certificate")

	cont.env.logger.Debug("Validating parameters")
	if err := params.ValidateName(true); err != nil {
		return nil, err
	}

	if err := params.ValidateExport(false); err != nil {
		return nil, err
	}

	if err := params.ValidatePrivate(false); err != nil {
		return nil, err
	}

	cont.env.logger.Infof("Showing certificate '%s'", *params.name)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return nil, err
	}

	certId, err := index.GetCert(*params.name)
	if err != nil {
		return nil, err
	}

	cert, err := cont.GetCert(certId)
	if err != nil {
		return nil, err
	}

	if *params.export == "" {
		return cert, nil
	} else {
		var files []ExportFile
		certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
		keyFile := fmt.Sprintf("%s-key.pem", cert.Data.Body.Name)

		files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
		if cert.Data.Body.CACertificate != "" {
			caFile := fmt.Sprintf("%s-cacert.pem", cert.Data.Body.Name)
			files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(cert.Data.Body.CACertificate)})
		}

		if *params.private {
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})
		}

		cont.env.logger.Infof("Exporting to '%s'", *params.export)
		Export(files, *params.export)
	}

	return nil, nil
}

func (cont *CertController) Update(params *CertParams) error {
	cont.env.logger.Info("Updating certificate")

	cont.env.logger.Debug("Validating parameters")
	if err := params.ValidateName(true); err != nil {
		return err
	}

	cont.env.logger.Infof("Updating certificate '%s'", *params.name)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	certId, err := index.GetCert(*params.name)
	if err != nil {
		return err
	}

	cert, err := cont.GetCert(certId)
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
		cert.Data.Body.Certificate = certPem
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
		cert.Data.Body.KeyType = string(keyType)

		cont.env.logger.Debug("Setting key PEM")
		cert.Data.Body.PrivateKey = keyPem
	}

	if *params.tags != "" {
		cont.ResetCertTags(certId, *params.tags)
	}

	cont.env.logger.Debug("Saving certificate")

	err = cont.SaveCert(cert)
	if err != nil {
		return err
	}

	return nil
}

func (cont *CertController) Delete(params *CertParams) error {
	cont.env.logger.Info("Deleting certificate")

	cont.env.logger.Debug("Validating parameters")
	if err := params.ValidateName(true); err != nil {
		return err
	}

	if err := params.ValidateConfirmDelete(true); err != nil {
		return err
	}

	cont.env.logger.Infof("Deleting certificate '%s' with reason '%s'", *params.name, *params.confirmDelete)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	certId, err := index.GetCert(*params.name)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Removing certificate file")
	if err := cont.env.api.DeletePrivate(cont.env.controllers.org.org.Data.Body.Id, certId); err != nil {
		return err
	}

	cont.env.logger.Debug("Removing cert from org index")
	if err := index.RemoveCert(*params.name); err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(index)
	if err != nil {
		return err
	}

	return nil

}
