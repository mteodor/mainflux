package provision_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mainflux/mainflux/pkg/errors"
	SDK "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/provision"
	"github.com/mainflux/mainflux/provision/mocks"

	logger "github.com/mainflux/mainflux/logger"
	"github.com/stretchr/testify/assert"
)

const (
	validEmail = "test@example.com"
	validPass  = "test"
	invalid    = "invalid"
	validToken = "valid_token"
)

var (
	cfg = provision.Config{
		Bootstrap: provision.Bootstrap{
			AutoWhiteList: true,
			Provision:     true,
			Content:       "",
			X509Provision: true,
		},
		Server: provision.ServiceConf{
			MfPass: "test",
			MfUser: "test@example.com",
		},
		Channels: []provision.Channel{
			provision.Channel{
				Name:     "control-channel",
				Metadata: map[string]interface{}{"type": "control"},
			},
			provision.Channel{
				Name:     "data-channel",
				Metadata: map[string]interface{}{"type": "data"},
			},
		},
		Things: []provision.Thing{
			provision.Thing{
				Name:     "thing",
				Metadata: map[string]interface{}{"external_id": "xxxxxx"},
			},
		},
	}
	log, _ = logger.New(os.Stdout, "info")
)

func TestProvision(t *testing.T) {
	// Create multiple services with different configurations.
	conf := SDK.Config{}
	sdk := mocks.NewSDK(conf)
	svc := provision.New(cfg, sdk, log)

	cases := []struct {
		desc        string
		externalID  string
		externalKey string
		svc         provision.Service
		err         error
	}{
		{
			desc:        "Provision successfully",
			externalID:  "id",
			externalKey: "key",
			svc:         svc,
			err:         nil,
		},
		{
			desc:        "Provision already existing config",
			externalID:  "id",
			externalKey: "key",
			svc:         svc,
			err:         provision.ErrFailedBootstrap,
		},
	}

	for _, tc := range cases {
		_, err := tc.svc.Provision("", "", tc.externalID, tc.externalKey)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected `%v` got `%v`", tc.desc, tc.err, err))
	}
}

func TestCert(t *testing.T) {
	conf := SDK.Config{}
	sdk := mocks.NewSDK(conf)
	svc := provision.New(cfg, sdk, log)

	nonExistingID, err := uuid.New().ID()
	assert.Nil(t, err, fmt.Sprintf("Failed to generate uuid: %s", err))

	existingID, err := sdk.CreateThing(SDK.Thing{Name: "test"}, validToken)

	cases := []struct {
		desc      string
		token     string
		thingId   string
		svc       provision.Service
		daysValid string
		rsaBits   int
		err       error
	}{
		{
			desc:      "Create certs successfully",
			token:     validToken,
			thingId:   existingID,
			svc:       svc,
			daysValid: "2400h",
			rsaBits:   4096,
			err:       nil,
		},
		{
			desc:      "Create certs for non existing id",
			token:     validToken,
			thingId:   nonExistingID,
			svc:       svc,
			daysValid: "2400h",
			rsaBits:   4096,
			err:       SDK.ErrUnauthorized,
		},
		{
			desc:      "Create certs for invalid token",
			token:     invalid,
			thingId:   existingID,
			svc:       svc,
			daysValid: "2400h",
			rsaBits:   4096,
			err:       SDK.ErrUnauthorized,
		},
	}

	for _, tc := range cases {
		_, _, err := tc.svc.Cert(tc.token, tc.thingId, tc.daysValid, tc.rsaBits)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected `%v` got `%v`", tc.desc, tc.err, err))
	}

}
