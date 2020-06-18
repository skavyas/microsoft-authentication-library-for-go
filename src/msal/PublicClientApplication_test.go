// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package msalgo

import (
	"errors"
	"reflect"
	"testing"

	"github.com/AzureAD/microsoft-authentication-library-for-go/src/internal/msalbase"
	"github.com/AzureAD/microsoft-authentication-library-for-go/src/internal/requests"
	"github.com/AzureAD/microsoft-authentication-library-for-go/src/internal/tokencache"
)

var tokenCommonParams = &acquireTokenCommonParameters{
	scopes: []string{"openid"},
}
var testAuthorityEndpoints = msalbase.CreateAuthorityEndpoints("https://login.microsoftonline.com/v2.0/authorize",
	"https://login.microsoftonline.com/v2.0/token",
	"https://login.microsoftonline.com/v2.0",
	"login.microsoftonline.com")
var testAuthorityInfo, err = msalbase.CreateAuthorityInfoFromAuthorityUri("https://login.microsoftonline.com/v2.0/", true)
var testAuthParams = msalbase.CreateAuthParametersInternal("clientID", testAuthorityInfo)
var appCommonParams = &applicationCommonParameters{
	clientID:      "clientID",
	authorityInfo: testAuthorityInfo,
}
var pcaParams = &PublicClientApplicationParameters{
	commonParameters: appCommonParams,
}
var tdr = &requests.TenantDiscoveryResponse{
	AuthorizationEndpoint: "https://login.microsoftonline.com/v2.0/authorize",
	TokenEndpoint:         "https://login.microsoftonline.com/v2.0/token",
	Issuer:                "https://login.microsoftonline.com/v2.0",
}
var wrm = new(requests.MockWebRequestManager)
var cacheManager = new(tokencache.MockCacheManager)
var testPCA = &PublicClientApplication{
	pcaParameters:     pcaParams,
	webRequestManager: wrm,
	cacheManager:      cacheManager,
}

func TestAcquireAuthCodeURL(t *testing.T) {
	authCodeParams := &AcquireTokenAuthCodeParameters{
		commonParameters:    tokenCommonParams,
		codeChallenge:       "codeChallenge",
		codeChallengeMethod: "plain",
		redirectURI:         "redirect",
	}

	wrm.On("GetTenantDiscoveryResponse",
		"https://login.microsoftonline.com/v2.0/v2.0/.well-known/openid-configuration").Return(tdr, nil)
	url, err := testPCA.AcquireAuthCodeURL(authCodeParams)
	if err != nil {
		t.Errorf("Error should be nil, instead it is %v", err)
	}
	actualURL := "https://login.microsoftonline.com/v2.0/authorize?client_id=clientID&code_challenge=codeChallenge" +
		"&code_challenge_method=plain&redirect_uri=redirect&response_type=code&scope=openid"
	if !reflect.DeepEqual(actualURL, url) {
		t.Errorf("URL should be %v, instead it is %v", actualURL, url)
	}
}

func TestAcquireTokenByAuthCode(t *testing.T) {
	testAuthParams.SetAuthorityEndpoints(testAuthorityEndpoints)
	testAuthParams.SetAuthorizationType(msalbase.AuthorizationTypeAuthCode)
	testAuthParams.SetScopes(tokenCommonParams.scopes)
	authCodeParams := &AcquireTokenAuthCodeParameters{
		commonParameters: tokenCommonParams,
	}
	wrm.On("GetTenantDiscoveryResponse",
		"https://login.microsoftonline.com/v2.0/v2.0/.well-known/openid-configuration").Return(tdr, nil)
	actualTokenResp := &msalbase.TokenResponse{}
	wrm.On("GetAccessTokenFromAuthCode", testAuthParams, "", "").Return(actualTokenResp, nil)
	_, err := testPCA.AcquireTokenByAuthCode(authCodeParams)
	if err != nil {
		t.Errorf("Error should be nil, instead it is %v", err)
	}
}

func TestExecuteTokenRequestWithoutCacheWrite(t *testing.T) {
	req := new(requests.MockTokenRequest)
	actualTokenResp := &msalbase.TokenResponse{}
	req.On("Execute").Return(actualTokenResp, nil)
	_, err := testPCA.executeTokenRequestWithoutCacheWrite(req, testAuthParams)
	if err != nil {
		t.Errorf("Error should be nil, instead it is %v", err)
	}
	mockError := errors.New("This is a mock error")
	errorReq := new(requests.MockTokenRequest)
	errorReq.On("Execute").Return(nil, mockError)
	_, err = testPCA.executeTokenRequestWithoutCacheWrite(errorReq, testAuthParams)
	if !reflect.DeepEqual(err, mockError) {
		t.Errorf("Actual error is %v, expected error is %v", err, mockError)
	}
}
