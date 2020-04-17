package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	SDK "github.com/mainflux/mainflux/sdk/go"
)

func TestValidate(t *testing.T) {

	cases := map[string]struct {
		ExternalID  string
		ExternalKey string
		err         error
	}{
		"mac address for device": {
			ExternalID:  "11:22:33:44:55:66",
			ExternalKey: "key12345678",
			err:         nil,
		},
		"external id for device empty": {
			err: SDK.Err
		},
	}

	for desc, tc := range cases {
		req := addThingReq{
			ExternalID:  tc.ExternalID,
			ExternalKey: tc.ExternalKey,
		}

		err := req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, err, tc.err))
	}
}
