package gameforge

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/alaingilbert/ogame/pkg/device"
	"github.com/alaingilbert/ogame/pkg/utils"
)

func AddAccountNew(dev *device.Device, ctx context.Context, lobby, accountGroup, sessionToken string) (*AddAccountRes, error) {
	var payload struct {
		AccountGroup string `json:"accountGroup"`
		Locale       string `json:"locale"`
		Kid          string `json:"kid"`
		Blackbox     string `json:"blackbox"`
	}
	payload.Blackbox, _ = dev.GetBlackbox()

	payload.AccountGroup = accountGroup // en_181
	payload.Locale = "en_GB"
	jsonPayloadBytes, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPut, getGameforgeLobbyBaseURL(lobby)+"/api/users/me/accounts", strings.NewReader(string(jsonPayloadBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Set(authorizationHeaderKey, buildBearerHeaderValue(sessionToken))
	req.Header.Set(contentTypeHeaderKey, applicationJson)
	req.Header.Set(acceptEncodingHeaderKey, gzipEncoding)
	req.WithContext(ctx)
	resp, err := dev.GetClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	by, err := utils.ReadBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, errors.New("invalid request, account already in lobby ?")
	}
	var newAccount AddAccountRes
	if err := json.Unmarshal(by, &newAccount); err != nil {
		return nil, errors.New(err.Error() + " : " + string(by))
	}
	if newAccount.Error != "" {
		return nil, errors.New(newAccount.Error)
	}
	newAccount.BearerToken = sessionToken
	return &newAccount, nil
}
