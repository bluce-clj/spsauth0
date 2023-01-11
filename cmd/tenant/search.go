package tenant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"spsauth0/common"
	"spsauth0/internal/config"
	"github.com/cenkalti/backoff"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/briandowns/spinner"
)

var (
	tenantSearchCmd = &cobra.Command{
		Use:     "search <clients>",
		Short:   "search tenant for clients, connects, users etc.",
		Aliases: []string{"sr"},
		Args:  cobra.ExactArgs(1),
		Run:     tenantSearchExecute,
	}
	clientListNames []string
	g *gocui.Gui
	clientList *[]Auth0Client
)

type requestHeaders map[string]string
type requestQueryParams map[string]string

const (
	authDomain 					  = "https://auth.test.spsapps.net"
	clientsUrlPattern 			  = "%s/api/v2/clients"
	authorizationHeaderKey        = "Authorization"
	pageParam                     = "page"
	perPageParam                  = "per_page"
	includeTotalsParam            = "include_totals"
	getClientPerPage              = 100
	includeTotals                 = "true"
	MaxAuth0Retries    = 4
	Auth0RetryInternal = 1 * time.Second
)

type Auth0Connector struct {
	Auth0Domain   string
	Client        *http.Client
	RetryInterval time.Duration
	MaxRetries    uint64
	ClientToUse 		  *config.Client
	Tenant			*config.Tenant
}

type Auth0PaginatedResponse struct {
	Total   int                  `json:"total"`
	Clients []Auth0Client `json:"clients"`
}

type Auth0ManagementApiErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
	Attributes *struct {
		Error string `json:"error"`
	} `json:"attributes"`
}

func tenantSearchExecute(cmd *cobra.Command, args []string) {
	// Validate arg is clients as it's the only one currently supported
	tenantConfig, err := config.LoadTenantConfigWithViper()
	if len(*tenantConfig.GetTenantProfileList()) == 0 {
		fmt.Printf("You must first configure a tenant before you are able to search a tenants configuration \n run `spsauth0 tenant add <tenant name>` ")
		os.Exit(1)
	}

	// also check to see if there are no configured tenant as error
	// should we still do this if there is only 1 tenant
	_, tenantName, err := common.PromptSelect("What tenant to search", tenantConfig.GetTenantListNames())
	if err != nil {
		fmt.Println(err.Error())
	}
	tenant := tenantConfig.GetTenantConfig(strings.ToLower(tenantName))
	client := getClientForSearch(tenant, tenantName)

	auth0 := Auth0Connector{
		Auth0Domain:   tenant.Tenant.Domain,
		Client:        &http.Client{Timeout: 10 * time.Second},
		RetryInterval: Auth0RetryInternal,
		MaxRetries:    uint64(MaxAuth0Retries),
		ClientToUse:   client,
	}

	// Handle Error
	clientList, _ = getAllClients(auth0)
	returnClientNameList(clientList)

	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.Mouse = false

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	//if err := g.SetKeybinding("finder", gocui.KeyArrowRight, gocui.ModNone, switchToMainView); err != nil {
	//	log.Panicln(err)
	//}
	//
	//if err := g.SetKeybinding("main", gocui.KeyArrowLeft, gocui.ModNone, switchToSideView); err != nil {
	//	log.Panicln(err)
	//}

	//if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
	//	log.Panicln(err)
	//}
	//if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
	//	log.Panicln(err)
	//}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func getClientForSearch(tenant *config.Tenant, tenantName string) *config.Client {

	if tenant.Tenant.DefaultClient != nil {
		_, useDefaultClient, _ := common.PromptSelect(fmt.Sprintf("Do you want to use the defaultClient %s set on the tenant?", tenant.Tenant.DefaultClient.ClientName), []string{"Yes", "No"})
		if useDefaultClient == "Yes"{
			return tenant.Tenant.DefaultClient
		}
	}

	// Get configured clients
	clientConfig, err := config.LoadClientConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load client config - %v\n", err)
		os.Exit(1)
	}

	if len(*clientConfig.GetClientList(tenantName)) == 0 {
		fmt.Printf("You have no configured clients for the %s tenant", tenantName)
		os.Exit(1)
	}

	// Wrap this is a user select y/n if they want to set a default client.- Give a short blurb on how this is used
	_, selectedClient, err := common.PromptSelect("Select a Default Client to use with this tenant ", config.GetClientListNames(*clientConfig.GetClientList(tenantName)))
	return clientConfig.GetClientConfig(strings.ToLower(selectedClient))
}

func getTokenHeaderVal(auth0 Auth0Connector) (string, error) {
	token := common.GetTokenHandler(auth0.ClientToUse)

	return "Bearer " + token, nil
}

func getAllClients(auth0 Auth0Connector) (*[]Auth0Client, error) {
	tokenHeaderVal, err := getTokenHeaderVal(auth0)

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)  // Build our new spinner
	s.Start()
	page := 0
	results, err := getPaginatedResults(tokenHeaderVal, page, auth0)
	if err != nil {
		return nil, err
	}

	perPage := getClientPerPage
	remainingPages := math.Ceil(float64(results.Total)/float64(perPage) - 1)

	clientList := make([]Auth0Client, 0)
	for _, client := range results.Clients {
		if client.Callbacks == nil {
			client.Callbacks = make([]string, 0)
		}
		clientList = append(clientList, client)
	}

	for getPage := 1; getPage <= int(remainingPages); getPage++ {
		results, err := getPaginatedResults(tokenHeaderVal, getPage, auth0)
		if err != nil {
			return nil, err
		}
		for _, client := range results.Clients {
			if client.Callbacks == nil {
				client.Callbacks = make([]string, 0)
			}
			clientList = append(clientList, client)
		}
	}
	s.Stop()
	return &clientList, nil
}

func getPaginatedResults(tokenHeaderVal string, page int, auth0 Auth0Connector) (*Auth0PaginatedResponse, error) {
	res, err := doRequest(http.MethodGet, fmt.Sprintf(clientsUrlPattern, authDomain), nil,
		requestHeaders{
			authorizationHeaderKey: tokenHeaderVal,
		},
		requestQueryParams{
			pageParam:          strconv.Itoa(page),
			perPageParam:       strconv.Itoa(getClientPerPage),
			includeTotalsParam: includeTotals,
		},
		http.StatusOK, auth0)
	if err != nil {
		return nil, err
	}


	defer res.Body.Close()

	var body Auth0PaginatedResponse
	decodeErr := json.NewDecoder(res.Body).Decode(&body)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return &body, nil
}

func doRequest(method string, url string, body []byte, headers requestHeaders, queryParams requestQueryParams, expectedStatus int, auth0 Auth0Connector) (*http.Response, error) {
	var res *http.Response
	retryErr := backoff.Retry(func() error {
		req, getReqErr := getRequest(method, url, body, headers, queryParams)
		if getReqErr != nil {
			return getReqErr
		}
		var doReqErr error
		res, doReqErr = auth0.Client.Do(req)
		if doReqErr != nil {
			return doReqErr
		}
		if res.StatusCode != expectedStatus {
			defer res.Body.Close()
			var body Auth0ManagementApiErrorResponse
			decodeErr := json.NewDecoder(res.Body).Decode(&body)
			if decodeErr != nil {
				return decodeErr
			}
			return decodeErr
		}
		return nil
	}, getRetryBackOff(auth0))

	if retryErr != nil {
		return nil, retryErr
	}
	return res, nil
}

func getRequest(method string, url string, body []byte, headers requestHeaders, queryParams requestQueryParams) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	if queryParams != nil {
		q := req.URL.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	return req, nil
}

func getRetryBackOff(auth0 Auth0Connector) backoff.BackOff {
	return backoff.WithMaxRetries(backoff.NewConstantBackOff(auth0.RetryInterval), auth0.MaxRetries)
}

//func cursorDown(g *gocui.Gui, v *gocui.View) error {
//	if v != nil {
//		cx, cy := v.Cursor()
//		if err := v.SetCursor(cx, cy+1); err != nil {
//			ox, oy := v.Origin()
//			if err := v.SetOrigin(ox, oy+1); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func cursorUp(g *gocui.Gui, v *gocui.View) error {
//	if v != nil {
//		ox, oy := v.Origin()
//		cx, cy := v.Cursor()
//		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
//			if err := v.SetOrigin(ox, oy-1); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func switchToSideView(g *gocui.Gui, view *gocui.View) error {
//	if _, err := g.SetCurrentView("finder"); err != nil {
//		return err
//	}
//	return nil
//}

//func switchToMainView(g *gocui.Gui, view *gocui.View) error {
//	if _, err := g.SetCurrentView("main"); err != nil {
//		return err
//	}
//	return nil
//}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("finder", -1, 0, 80, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Editable = true
		v.Frame = true
		v.Title = "Type pattern here. Press -> or <- to switch between panes"
		if _, err := g.SetCurrentView("finder"); err != nil {
			return err
		}
		v.Editor = gocui.EditorFunc(finder)
	}
	if v, err := g.SetView("Top Match", 79, 0, maxX, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.Title = "Top Match"
	}
	if v, err := g.SetView("results", -1, 3, 79, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = false
		v.Wrap = true
		v.Frame = true
		v.Title = "Search Results"
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func finder(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
		g.Update(func(gui *gocui.Gui) error {
			results, err := g.View("results")
			object, err := g.View("Top Match")
			if err != nil {
				// handle error
			}
			object.Clear()
			results.Clear()
			t := time.Now()
			matches := fuzzy.Find(strings.TrimSpace(v.ViewBuffer()), clientListNames)
			elapsed := time.Since(t)
			fmt.Fprintf(results, "found %v matches in %v\n", len(matches), elapsed)
			//topMatch, err := json.MarshalIndent(getClientbByClientName(matches[0].Str), "", "  ")
			//fmt.Fprintf(object, "%s\n", topMatch)
			for k, match := range matches {
				if k == 0 {
					topMatch, _ := json.MarshalIndent(getClientbByClientName(matches[0].Str), "", "  ")
					fmt.Fprintf(object, "%s\n", topMatch)
				}
				for i := 0; i < len(match.Str); i++ {
					if contains(i, match.MatchedIndexes) {
						fmt.Fprintf(results, fmt.Sprintf("\033[1;31m%s\033[0m", string(match.Str[i])))
					} else {
						fmt.Fprintf(results, string(match.Str[i]))
					}

				}
				fmt.Fprintln(results, "")
			}
			return nil
		})
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
		g.Update(func(gui *gocui.Gui) error {
			results, err := g.View("results")
			object, err := g.View("Top Match")
			if err != nil {
				// handle error
			}
			object.Clear()
			results.Clear()
			t := time.Now()
			matches := fuzzy.Find(strings.TrimSpace(v.ViewBuffer()), clientListNames)
			elapsed := time.Since(t)
			fmt.Fprintf(results, "found %v matches in %v\n", len(matches), elapsed)
			//topMatch, err := json.MarshalIndent(getClientbByClientName(matches[0].Str), "", "  ")
			//fmt.Fprintf(object, "%s\n", topMatch)
			for k, match := range matches {
				if k == 0 {
					topMatch, _ := json.MarshalIndent(getClientbByClientName(matches[0].Str), "", "  ")
					fmt.Fprintf(object, "%s\n", topMatch)
				}
				for i := 0; i < len(match.Str); i++ {
					if contains(i, match.MatchedIndexes) {
						fmt.Fprintf(results, fmt.Sprintf("\033[1;31m%s\033[0m", string(match.Str[i])))
					} else {
						fmt.Fprintf(results, string(match.Str[i]))
					}

				}
				fmt.Fprintln(results, "")
			}
			return nil
		})
	case key == gocui.KeyDelete:
		v.EditDelete(false)
		g.Update(func(gui *gocui.Gui) error {
			results, err := g.View("results")
			object, err := g.View("Top Match")
			if err != nil {
				// handle error
			}
			object.Clear()
			results.Clear()
			t := time.Now()
			matches := fuzzy.Find(strings.TrimSpace(v.ViewBuffer()), clientListNames)
			elapsed := time.Since(t)
			fmt.Fprintf(results, "found %v matches in %v\n", len(matches), elapsed)

			for k, match := range matches {
				if k == 0 {
					topMatch, _ := json.MarshalIndent(getClientbByClientName(matches[0].Str), "", "  ")
					fmt.Fprintf(object, "%s\n", topMatch)
				}
				for i := 0; i < len(match.Str); i++ {
					if contains(i, match.MatchedIndexes) {
						fmt.Fprintf(results, fmt.Sprintf("\033[1;31m%s\033[0m", string(match.Str[i])))
					} else {
						fmt.Fprintf(results, string(match.Str[i]))
					}

				}
				fmt.Fprintln(results, "")
			}
			return nil
		})
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	}
}

func contains(needle int, haystack []int) bool {
	for _, i := range haystack {
		if needle == i {
			return true
		}
	}
	return false
}

func returnClientNameList(clients *[]Auth0Client) {
	for _, v := range *clients {
		clientListNames = append(clientListNames, v.Name)
	}
}

func getClientbByClientName(clientName string) *Auth0Client{

	for _,v := range *clientList {
		if strings.ToLower(clientName) == strings.ToLower(v.Name) {
			return &v
		}
	}
	return nil
}

type Auth0Client struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	ClientID     string `json:"client_id,omitempty"`     // resp-only
	ClientSecret string `json:"client_secret,omitempty"` // resp-only
	LogoUri      string `json:"logo_uri,omitempty"`
	IsFirstParty *bool  `json:"is_first_party,omitempty"` // resp-only
	// for callbacks we don't want to omit the field if the value is empty as this will not allow users to clear out all
	// of their clients callback redirect_urls.
	Callbacks               []string `json:"callbacks"`
	AllowedOrigins          []string `json:"allowed_origins,omitempty"`
	WebOrigins              []string `json:"web_origins,omitempty"`
	ClientAliases           []string `json:"client_aliases,omitempty"`
	AllowedClients          []string `json:"allowed_clients,omitempty"`
	AllowedLogoutUrls       []string `json:"allowed_logout_urls,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	AppType                 string   `json:"app_type,omitempty"`
	OidcConformant          bool     `json:"oidc_conformant,omitempty"`
	JwtConfiguration        *struct {
		LifetimeInSeconds float64                `json:"lifetime_in_seconds,omitempty"`
		SecretEncoded     bool                   `json:"secret_encoded,omitempty"` // resp-only
		Scopes            map[string]interface{} `json:"scopes,omitempty"`
		Alg               string                 `json:"alg,omitempty"`
	} `json:"jwt_configuration,omitempty"`
	SigningKeys   *[]map[string]interface{} `json:"signing_keys,omitempty"` // resp-only
	EncryptionKey *struct {
		Pub     string `json:"pub,omitempty"`
		Cert    string `json:"cert,omitempty"`
		Subject string `json:"subject,omitempty"`
	} `json:"encryption_key,omitempty"`
	Sso                    bool                    `json:"sso,omitempty"`
	CrossOriginAuth        bool                    `json:"cross_origin_auth,omitempty"`
	CrossOriginLoc         string                  `json:"cross_origin_loc,omitempty"`
	SsoDisabled            bool                    `json:"sso_disabled,omitempty"`
	CustomLoginPageOn      bool                    `json:"custom_login_page_on,omitempty"`
	CustomLoginPage        string                  `json:"custom_login_page,omitempty"`
	CustomLoginPagePreview string                  `json:"custom_login_page_preview,omitempty"`
	FormTemplate           string                  `json:"form_template,omitempty"`
	IsHerokuApp            bool                    `json:"is_heroku_app,omitempty"`
	Addons                 *map[string]interface{} `json:"addons,omitempty"`
	ClientMetadata         *map[string]interface{} `json:"client_metadata,omitempty"`
	Mobile                 *struct {
		Android map[string]interface{}
		IOS     map[string]interface{}
	} `json:"mobile,omitempty"`
	IsTestOnly bool     `json:"is_test_only,omitempty"`
	Scopes     []string `json:"scopes,omitempty"`
}