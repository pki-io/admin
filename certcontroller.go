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
	logger.Debug("getting CA")
	logger.Tracef("received CA id '%s'", caId)

	logger.Debug("creating new CA controller")
	caCont, err := NewCAController(cont.env)
	if err != nil {
		return nil, err
	}

	logger.Debug("returning CA")
	return caCont.GetCA(caId)
}

func (cont *CertController) ResetCertTags(certId, tags string) error {
	logger.Debug("resetting certificate tags")
	logger.Tracef("received certificate id '%s' and tags '%s'", certId, tags)

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

	logger.Trace("returning nil error")
	return nil
}

func (cont *CertController) GetCert(id string) (*x509.Certificate, error) {
	logger.Debug("getting certificate")
	logger.Tracef("received certificate id '%s'", id)

	logger.Debugf("getting private file '%s' from org", id)
	certContainerJson, err := cont.env.api.GetPrivate(cont.env.controllers.org.org.Data.Body.Id, id)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating new container")
	certContainer, err := document.NewContainer(certContainerJson)
	if err != nil {
		return nil, err
	}

	logger.Debug("decrypting container")
	certJson, err := cont.env.controllers.org.org.VerifyThenDecrypt(certContainer)
	if err != nil {
		return nil, err
	}

	logger.Debug("loading certificate json")
	cert, err := x509.NewCertificate(certJson)
	if err != nil {
		return nil, err
	}

	logger.Trace("returning nil error")
	return cert, nil
}

func (cont *CertController) SaveCert(cert *x509.Certificate) error {
	logger.Debug("saving certificate")
	logger.Tracef("received certificate with id '%s'", cert.Id())

	logger.Debug("encrypting cert for org")
	certContainer, err := cont.env.controllers.org.org.EncryptThenSignString(cert.Dump(), nil)
	if err != nil {
		return err
	}

	logger.Debug("saving encrypted cert")
	err = cont.env.api.SendPrivate(cont.env.controllers.org.org.Data.Body.Id, cert.Data.Body.Id, certContainer.Dump())
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *CertController) AddCertToOrgIndex(cert *x509.Certificate, tags string) error {
	logger.Debug("adding certificate to org index")
	logger.Tracef("received certificate with id '%s' and tags '%s'", cert.Id(), tags)

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

	logger.Trace("returning nil error")
	return nil
}

func (cont *CertController) New(params *CertParams) (*x509.Certificate, error) {
	logger.Debug("creating new certificate")
	logger.Tracef("received params: %s", params)

	if err := params.ValidateName(true); err != nil {
		return nil, err
	}

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

	logger.Debug("creating certificate struct")
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
		logger.Debug("generating certificate and key")
		if *params.ca == "" {
			if err := cert.Generate(nil, &subject); err != nil {
				return nil, err
			}
		} else {
			index, err := cont.env.controllers.org.GetIndex()
			if err != nil {
				return nil, err
			}

			caId, err := index.GetCA(*params.ca)
			if err != nil {
				return nil, err
			}

			ca, err := cont.GetCA(caId)
			if err != nil {
				return nil, err
			}

			logger.Debugf("generating certificate and signing with CA '%s'", caId)
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

		logger.Debugf("importing certificate from '%s'", *params.certFile)
		ok, err := fs.Exists(*params.certFile)
		if err != nil {
			return nil, err
		}

		if !ok {
			logger.Warnf("certificate file '%s' does not exist", *params.certFile)
			return nil, nil
		}

		logger.Debug("reading certificate from file")
		certPem, err := fs.ReadFile(*params.certFile)
		if err != nil {
			return nil, err
		}

		logger.Debug("decoding certificate PEM")
		importCert, err := x509.PemDecodeX509Certificate([]byte(certPem))
		if err != nil {
			return nil, err
		}

		cert.Data.Body.Id = NewID()
		cert.Data.Body.Certificate = certPem
		certExpiry := int(importCert.NotAfter.Sub(importCert.NotBefore) / (time.Hour * 24))
		cert.Data.Body.Expiry = certExpiry

		if *params.keyFile != "" {
			logger.Debugf("importing certificate privte key from '%s'", *params.keyFile)
			ok, err := fs.Exists(*params.keyFile)
			if err != nil {
				return nil, err
			}

			if !ok {
				logger.Warnf("key file '%s' does not exist", *params.keyFile)
				return nil, nil
			}

			logger.Debug("reading private key file")
			keyPem, err := fs.ReadFile(*params.keyFile)
			if err != nil {
				return nil, err
			}

			logger.Debug("decoding private key PEM")
			key, err := crypto.PemDecodePrivate([]byte(keyPem))
			if err != nil {
				return nil, err
			}

			logger.Debug("getting key type")
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
		err = cont.SaveCert(cert)
		if err != nil {
			return nil, err
		}

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

		logger.Trace("returning certificate")
		return cert, nil

	} else {
		logger.Debugf("Exporting to '%s'", *params.standaloneFile)
		Export(files, *params.standaloneFile)
	}

	logger.Trace("returning nil error")
	return nil, nil
}

func (cont *CertController) List(params *CertParams) ([]*x509.Certificate, error) {
	logger.Debug("listing certificates")
	logger.Tracef("received params: %s", params)

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

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

	logger.Trace("returning certificates")
	return certs, nil
}

func (cont *CertController) Show(params *CertParams) (*x509.Certificate, error) {
	logger.Debug("showing certificate")
	logger.Tracef("received params: %s", params)

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

	certId, err := index.GetCert(*params.name)
	if err != nil {
		return nil, err
	}

	cert, err := cont.GetCert(certId)
	if err != nil {
		return nil, err
	}

	if *params.export == "" {
		logger.Debug("returning certificate")
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

		logger.Debugf("Exporting to '%s'", *params.export)
		Export(files, *params.export)
	}

	logger.Trace("returning nil error")
	return nil, nil
}

func (cont *CertController) Update(params *CertParams) error {
	logger.Debug("updating certificate")
	logger.Tracef("received params: %s", params)

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

		cert.Data.Body.Certificate = certPem
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

		logger.Debugf("reading key file '%s'", *params.keyFile)

		keyPem, err := fs.ReadFile(*params.keyFile)
		if err != nil {
			return err
		}

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

		cert.Data.Body.KeyType = string(keyType)
		cert.Data.Body.PrivateKey = keyPem
	}

	if *params.tags != "" {
		cont.ResetCertTags(certId, *params.tags)
	}

	err = cont.SaveCert(cert)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *CertController) Delete(params *CertParams) error {
	logger.Debug("deleting certificate")
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

	certId, err := index.GetCert(*params.name)
	if err != nil {
		return err
	}

	logger.Debugf("removing certificate file '%s'", certId)
	if err := cont.env.api.DeletePrivate(cont.env.controllers.org.OrgId(), certId); err != nil {
		return err
	}

	if err := index.RemoveCert(*params.name); err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(index)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}
