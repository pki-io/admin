// ThreatSpec package main
package main

import (
	"fmt"
	"strings"
)

type PairingKeyParams struct {
	id            *string
	tags          *string
	confirmDelete *string
	private       *bool
}

func NewPairingKeyParams() *PairingKeyParams {
	return new(PairingKeyParams)
}

func (params *PairingKeyParams) ValidateID(required bool) error {
	if required && *params.id == "" {
		return fmt.Errorf("id cannot be empty")
	}
	return nil
}

func (params *PairingKeyParams) ValidateTags(required bool) error          { return nil }
func (params *PairingKeyParams) ValidatePrivate(required bool) error       { return nil }
func (params *PairingKeyParams) ValidateConfirmDelete(required bool) error { return nil }

type PairingKeyController struct {
	env *Environment
}

func NewPairingKeyController(env *Environment) (*PairingKeyController, error) {
	cont := new(PairingKeyController)
	cont.env = env
	return cont, nil
}

func (cont *PairingKeyController) GeneratePairingKey() (string, string) {
	logger.Debug("generating pairing key")
	id := NewID()
	key := NewID()

	logger.Trace("returning pairing key")
	return id, key
}

func (cont *PairingKeyController) AddPairingKeyToOrgIndex(id, key, tags string) error {
	logger.Debug("adding pairing key to org index")
	logger.Tracef("received id '%s', key [NOT LOGGED], and tags '%s'", id, tags)

	orgIndex, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	if err := orgIndex.AddPairingKey(id, key, ParseTags(tags)); err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(orgIndex)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *PairingKeyController) New(params *PairingKeyParams) (string, string, error) {
	logger.Debug("creating new pairing key")
	logger.Tracef("received params: %s", params)

	if err := params.ValidateTags(true); err != nil {
		return "", "", err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return "", "", err
	}

	id, key := cont.GeneratePairingKey()

	err := cont.AddPairingKeyToOrgIndex(id, key, *params.tags)
	if err != nil {
		return "", "", err
	}

	logger.Trace("returning pairing key")
	return id, key, nil
}

func (cont *PairingKeyController) List(params *PairingKeyParams) ([][]string, error) {
	logger.Debug("listing pairing keys")
	logger.Tracef("received params: %s", params)

	keys := [][]string{}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return keys, err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return keys, err
	}

	logger.Flush()
	for id, pk := range index.GetPairingKeys() {
		keys = append(keys, []string{id, strings.Join(pk.Tags[:], ",")})
	}

	logger.Trace("returning keys")
	return keys, nil
}

func (cont *PairingKeyController) Show(params *PairingKeyParams) (string, string, string, error) {
	logger.Debug("showing pairing key")
	logger.Tracef("received params: %s", params)

	if err := params.ValidateID(true); err != nil {
		return "", "", "", err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return "", "", "", err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return "", "", "", err
	}

	pk, err := index.GetPairingKey(*params.id)
	if err != nil {
		return "", "", "", err
	}

	logger.Trace("returning pairing key")
	return *params.id, pk.Key, strings.Join(pk.Tags[:], ","), nil
}

func (cont *PairingKeyController) Delete(params *PairingKeyParams) error {
	logger.Debug("deleting pairing key")
	logger.Tracef("received params: %s", params)

	if err := params.ValidateID(true); err != nil {
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

	if err := index.RemovePairingKey(*params.id); err != nil {
		return err
	}

	err = cont.env.controllers.org.SaveIndex(index)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}
