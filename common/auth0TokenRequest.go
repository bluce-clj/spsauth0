package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"spsauth0/internal/config"
	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"github.com/skratchdot/open-golang/open"
)

type auth0TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Audience     string `json:"audience"`
}

const (
	OAuthTokenPattern = "/oauth/token"
)

type auth0TokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"errorDescription"`
}

type auth0TokenSuccessResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

var accessToken = ""

func GetTokenHandler(client *config.Client) string {

	switch client.ClientType {
	case "Machine-to-Machine Application":
		return getClientToken(client)
	default:
		return getUserTokenPKCE(client)
	}
}

// GetClientToken used for client_credential flow
func getClientToken(client *config.Client)  string{

	tenantConfig, err := config.LoadTenantConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	tenant := tenantConfig.GetTenantConfig(strings.ToLower(client.TenantName))
	if tenant == nil {
		fmt.Println("Something went wrong in getting the tenant information for the client.")
		os.Exit(1)
	}

	audience := getAudienceFromTenant(tenant, tenantConfig)

	jsonBody, _ := json.Marshal(auth0TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     client.ClientId,
		ClientSecret: client.ClientSecret,
		Audience:     audience,
	})

	url := "https://" + tenant.Tenant.Domain + OAuthTokenPattern
	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Error executing http request: %v", err)
		os.Exit(1)
	}

	if res.StatusCode != http.StatusOK {
		var body auth0TokenErrorResponse
		json.NewDecoder(res.Body).Decode(&body)
		defer res.Body.Close()
		fmt.Printf("Call to obtain auth0 token returned non-OK status %d: %v\n", res.StatusCode, body)
		os.Exit(1)
	}

	var body auth0TokenSuccessResponse
	err = json.NewDecoder(res.Body).Decode(&body)
	defer res.Body.Close()

	return body.AccessToken
}

func getAudienceFromTenant(tenant *config.Tenant, tenantConfig *config.TenantConfig) string {
	if len(tenant.Tenant.APIs) == 0 {
		// add Use command to output
		fmt.Printf("To request a token for this client you need to add configures API to the %s tenant. Use ``")
		os.Exit(1)
	}

	i, _, err := PromptSelect("Select the audience that this token is for", tenantConfig.GetTenantAPINames(tenant.Tenant.APIs))
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	return tenant.Tenant.APIs[i].Audience
}

// AuthorizeUser implements the PKCE OAuth2 flow.
func getUserTokenPKCE(client *config.Client)  string{

	tenantConfig, err := config.LoadTenantConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	tenant := tenantConfig.GetTenantConfig(strings.ToLower(client.TenantName))
	if tenant == nil {
		fmt.Println("Something went wrong in getting the tenant information for the client.")
		os.Exit(1)
	}
	
	additionalQueryParams := ""
	codeVerifier := ""
	switch client.ClientType {
	case "Native Application":
			additionalQueryParams,  codeVerifier = pkceAuthorizationQueryParams()
	case "Web Service Application":
			additionalQueryParams = webServiceAppAuthorizationQueryParams()
	case "Single-Page Application (SPA)":
			additionalQueryParams = spaAuthorizationQueryParams()
	}

	// construct the authorization URL (with Auth0 as the authorization provider)
	redirectURL :=  "http://localhost:1000"
	authorizationURL := fmt.Sprintf(
		"https://" + tenant.Tenant.Domain + "/authorize" +
			"?audience=api://api.spscommerce.com/"+
			"&client_id=%s"+
			"&redirect_uri=%s"+
			additionalQueryParams,
		client.ClientId, redirectURL)

	// start a web server to listen on a callback URL
	server := &http.Server{Addr: redirectURL}

	// define a handler that will get the authorization code, call the token endpoint, and close the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if client.ClientType != "Single-Page Application (SPA)" {
			// get the authorization code
			code := r.URL.Query().Get("code")
			if code == "" {
				fmt.Println("Url Param 'code' is missing")
				io.WriteString(w, "Error: could not find 'code' URL parameter\n")

				// close the HTTP server and return
				cleanup(server)
				return
			}

			accessToken, err = getAccessToken(client, codeVerifier, code, redirectURL, tenant.Tenant.Domain)

			if err != nil {
				fmt.Println("could not get access token")
				io.WriteString(w, "Error: could not retrieve access token\n")

				// close the HTTP server and return
				cleanup(server)
				return
			}
		}

		// return an indication of success to the caller
		io.WriteString(w, `
		<html>
			<body>
				<pre>
 ________   __  __   _________  ___   ___   _________  ______   ______   __          
/_______/\ /_/\/_/\ /________/\/__/\ /__/\ /________/\/_____/\ /_____/\ /_/\         
\::: _  \ \\:\ \:\ \\__.::.__\/\::\ \\  \ \\__.::.__\/\:::_ \ \\:::_ \ \\:\ \        
 \::(_)  \ \\:\ \:\ \  \::\ \   \::\/_\ .\ \  \::\ \   \:\ \ \ \\:\ \ \ \\:\ \       
  \:: __  \ \\:\ \:\ \  \::\ \   \:: ___::\ \  \::\ \   \:\ \ \ \\:\ \ \ \\:\ \____  
   \:.\ \  \ \\:\_\:\ \  \::\ \   \: \ \\::\ \  \::\ \   \:\_\ \ \\:\_\ \ \\:\/___/\ 
    \__\/\__\/ \_____\/   \__\/    \__\/ \::\/   \__\/    \_____\/ \_____\/ \_____\/
				</pre>

				<h2>You can close this window and return to authtool.</h2>
				<p style="overflow-wrap: anywhere" id="authToken"></p>

				<div> 
				<script>
    					document.getElementById("authToken").innerHTML = 
						window.location.hash;
				</script>
				</div>
			</body>
		</html>`)

		// close the HTTP server
		cleanup(server)
	})

	// parse the redirect URL for the port number
	u, err := url.Parse(redirectURL)
	if err != nil {
		fmt.Printf("bad redirect URL: %s\n", err)
		os.Exit(1)
	}

	// set up a listener on the redirect port
	port := fmt.Sprintf(":%s", u.Port())
	l, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("can't listen to port %s: %s\n", port, err)
		os.Exit(1)
	}

	// open a browser window to the authorizationURL
	err = open.Start(authorizationURL)
	if err != nil {
		fmt.Printf("can't open browser to URL %s: %s\n", authorizationURL, err)
		os.Exit(1)
	}

	// start the blocking web server loop
	// this will exit when the handler gets fired and calls server.Close()
	server.Serve(l)
	return accessToken
}

// getAccessToken trades the authorization code retrieved from the first OAuth2 leg for an access token
func getAccessToken(client *config.Client, codeVerifier string, authorizationCode string, callbackURL string, domain string) (string, error) {
	// set the url and form-encoded data for the POST to the access token endpoint
	url := "https://" + domain + "/oauth/token"
	
	additionalQueryParams := ""
	switch  client.ClientType{
	case "Native Application":
		additionalQueryParams = pkceAccessTokenQueryParams(codeVerifier)
	case "Web Service Application":
		additionalQueryParams = webServiceAppTokenQueryParams(client.ClientSecret)
	}

	data := fmt.Sprintf(
		"grant_type=authorization_code&client_id=%s"+
			"&code=%s"+
			"&redirect_uri=%s"+
			additionalQueryParams,
		client.ClientId, authorizationCode, callbackURL)
	payload := strings.NewReader(data)

	// create the request and execute it
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("snap: HTTP error: %s", err)
		return "", err
	}

	// process the response
	defer res.Body.Close()
	var responseData map[string]interface{}
	body, _ := ioutil.ReadAll(res.Body)

	// unmarshal the json into a string map
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Printf("JSON error: %s", err)
		return "", err
	}

	// retrieve the access token out of the map, and return to caller
	accessToken := responseData["access_token"].(string)
	return accessToken, nil
}

func pkceAuthorizationQueryParams() (string, string) {
	// initialize the code verifier
	var CodeVerifier, _ = cv.CreateCodeVerifier()

	// Create code_challenge with S256 method
	codeChallenge := CodeVerifier.CodeChallengeS256()
	
	return fmt.Sprintf("&code_challenge=%s&code_challenge_method=S256&scope=offline_access&response_type=code",
		codeChallenge), CodeVerifier.String()
}

func pkceAccessTokenQueryParams(codeVerifier string) string{
	return fmt.Sprintf("&code_verifier=%s",codeVerifier)
}

func webServiceAppAuthorizationQueryParams() string {
	return fmt.Sprintf("&scope=offline_access&response_type=code")
}

func webServiceAppTokenQueryParams(clientSecret string)  string{
	return fmt.Sprintf("&client_secret=%s", clientSecret)
}

func spaAuthorizationQueryParams() string{
	return "&response_type=token"
}

// cleanup closes the HTTP server
func cleanup(server *http.Server) {
	// we run this as a goroutine so that this function falls through and
	// the socket to the browser gets flushed/closed before the server goes away
	go server.Close()
}