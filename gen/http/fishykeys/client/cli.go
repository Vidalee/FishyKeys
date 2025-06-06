// Code generated by goa v3.21.1, DO NOT EDIT.
//
// fishykeys HTTP client CLI support package
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/backend/design

package client

import (
	"encoding/json"
	"fmt"

	fishykeys "github.com/Vidalee/FishyKeys/backend/gen/fishykeys"
)

// BuildCreateMasterKeyPayload builds the payload for the fishykeys
// create_master_key endpoint from CLI flags.
func BuildCreateMasterKeyPayload(fishykeysCreateMasterKeyBody string) (*fishykeys.CreateMasterKeyPayload, error) {
	var err error
	var body CreateMasterKeyRequestBody
	{
		err = json.Unmarshal([]byte(fishykeysCreateMasterKeyBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"min_shares\": 3,\n      \"total_shares\": 5\n   }'")
		}
	}
	v := &fishykeys.CreateMasterKeyPayload{
		TotalShares: body.TotalShares,
		MinShares:   body.MinShares,
	}

	return v, nil
}

// BuildAddSharePayload builds the payload for the fishykeys add_share endpoint
// from CLI flags.
func BuildAddSharePayload(fishykeysAddShareBody string) (*fishykeys.AddSharePayload, error) {
	var err error
	var body AddShareRequestBody
	{
		err = json.Unmarshal([]byte(fishykeysAddShareBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"share\": \"EXAMPLEA5ZKwDn8Zotr3B+d+F+UzrcJ1Yhl2rU0\"\n   }'")
		}
	}
	v := &fishykeys.AddSharePayload{
		Share: body.Share,
	}

	return v, nil
}

// BuildDeleteSharePayload builds the payload for the fishykeys delete_share
// endpoint from CLI flags.
func BuildDeleteSharePayload(fishykeysDeleteShareBody string) (*fishykeys.DeleteSharePayload, error) {
	var err error
	var body DeleteShareRequestBody
	{
		err = json.Unmarshal([]byte(fishykeysDeleteShareBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"index\": 1\n   }'")
		}
	}
	v := &fishykeys.DeleteSharePayload{
		Index: body.Index,
	}

	return v, nil
}
