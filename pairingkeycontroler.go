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
	id := NewID()
	key := NewID()

	return id, key
}

func (cont *PairingKeyController) AddPairingKeyToOrgIndex(id, key, tags string) error {
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

	return nil

}

func (cont *PairingKeyController) New(params *PairingKeyParams) error {

	cont.env.logger.Info("Creating new pairing key")

	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateTags(true); err != nil {
		return err
	}

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	id, key := cont.GeneratePairingKey()

	cont.env.logger.Debug("Adding pairing key to org")

	err := cont.AddPairingKeyToOrgIndex(id, key, *params.tags)
	if err != nil {
		return err
	}

	cont.env.logger.Flush()
	fmt.Printf("Pairing ID: %s\n", id)
	fmt.Printf("Pairing key: %s\n", key)

	return nil
}

func (cont *PairingKeyController) List(params *PairingKeyParams) error {
	cont.env.logger.Info("Listing pairing keys")

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
	for id, pk := range index.GetPairingKeys() {
		fmt.Printf("* %s %s\n", id, strings.Join(pk.Tags[:], ","))
	}

	return nil
}

func (cont *PairingKeyController) Show(params *PairingKeyParams) error {

	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateID(true); err != nil {
		return err
	}

	cont.env.logger.Info("Showing pairing key")

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
	pk, err := index.GetPairingKey(*params.id)
	if err != nil {
		return err
	}

	fmt.Printf("ID: %s\n", *params.id)
	if *params.private {
		fmt.Printf("Key: %s\n", pk.Key)
	}
	fmt.Printf("Tags: %s\n", strings.Join(pk.Tags[:], ","))

	return nil
}

func (cont *PairingKeyController) Delete(params *PairingKeyParams) error {

	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateID(true); err != nil {
		return err
	}

	if err := params.ValidateConfirmDelete(true); err != nil {
		return err
	}

	cont.env.logger.Infof("Deleting pairing key '%s'", *params.id)

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.env.logger.Debug("Getting org index")

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

	return nil
}
