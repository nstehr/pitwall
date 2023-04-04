package ziti

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/enroll"
)

type Client struct {
	username string
	password string
	ctrlUrl  string
	token    string
}

type Identity struct {
	Data struct {
		Links struct {
			AuthPolicies struct {
				Href string `json:"href"`
			} `json:"auth-policies"`
			Authenticators struct {
				Href string `json:"href"`
			} `json:"authenticators"`
			EdgeRouterPolicies struct {
				Href string `json:"href"`
			} `json:"edge-router-policies"`
			EdgeRouters struct {
				Href string `json:"href"`
			} `json:"edge-routers"`
			Enrollments struct {
				Href string `json:"href"`
			} `json:"enrollments"`
			FailedServiceRequests struct {
				Href string `json:"href"`
			} `json:"failed-service-requests"`
			PostureData struct {
				Href string `json:"href"`
			} `json:"posture-data"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
			ServiceConfigs struct {
				Href string `json:"href"`
			} `json:"service-configs"`
			ServicePolicies struct {
				Href string `json:"href"`
			} `json:"service-policies"`
			Services struct {
				Href string `json:"href"`
			} `json:"services"`
		} `json:"_links"`
		CreatedAt time.Time `json:"createdAt"`
		ID        string    `json:"id"`
		Tags      struct {
		} `json:"tags"`
		UpdatedAt time.Time `json:"updatedAt"`
		AppData   struct {
		} `json:"appData"`
		AuthPolicyID   string `json:"authPolicyId"`
		Authenticators struct {
		} `json:"authenticators"`
		DefaultHostingCost       int    `json:"defaultHostingCost"`
		DefaultHostingPrecedence string `json:"defaultHostingPrecedence"`
		Disabled                 bool   `json:"disabled"`
		Enrollment               struct {
			Ott struct {
				ExpiresAt time.Time `json:"expiresAt"`
				ID        string    `json:"id"`
				Jwt       string    `json:"jwt"`
				Token     string    `json:"token"`
			} `json:"ott"`
		} `json:"enrollment"`
		EnvInfo struct {
		} `json:"envInfo"`
		ExternalID              interface{} `json:"externalId"`
		HasAPISession           bool        `json:"hasApiSession"`
		HasEdgeRouterConnection bool        `json:"hasEdgeRouterConnection"`
		IsAdmin                 bool        `json:"isAdmin"`
		IsDefaultAdmin          bool        `json:"isDefaultAdmin"`
		IsMfaEnabled            bool        `json:"isMfaEnabled"`
		Name                    string      `json:"name"`
		RoleAttributes          interface{} `json:"roleAttributes"`
		SdkInfo                 struct {
		} `json:"sdkInfo"`
		ServiceHostingCosts struct {
		} `json:"serviceHostingCosts"`
		ServiceHostingPrecedences struct {
		} `json:"serviceHostingPrecedences"`
		Type struct {
			Links struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
			Entity string `json:"entity"`
			ID     string `json:"id"`
			Name   string `json:"name"`
		} `json:"type"`
		TypeID string `json:"typeId"`
	} `json:"data"`
	Meta struct {
	} `json:"meta"`
}

type IdentityType int64
type ServicePolicyType int64
type ServicePolicySemantic int64

const (
	ztSession = "zt-session"
)

const (
	User IdentityType = iota
	Device
	Service
	Router
)

const (
	Dial ServicePolicyType = iota
	Bind
)

const (
	AnyOf = iota
	AllOf
)

func (it IdentityType) String() string {
	return []string{"User", "Device", "Service", "Router"}[it]
}

func (spt ServicePolicyType) String() string {
	return []string{"Dial", "Bind"}[spt]
}

func (sc ServicePolicySemantic) String() string {
	return []string{"AnyOf", "AllOf"}[sc]
}

func NewClient(ctrlUrl, username, password string) (*Client, error) {
	// writing my own client for management to be able to create services and identities, programmatically
	// initially looked at embedding the sdk RestClient, but found at this point I wasn't using much from it
	// TODO: we since this all about secure, we should fix this
	return &Client{ctrlUrl: ctrlUrl, username: username, password: password}, nil
}

func (c *Client) Login() error {
	// for now, just supporting password auth here, but will look into cert based auth
	// and when we do cert based, we can use the official rest client
	authUrl := fmt.Sprintf("%s/edge/management/v1/authenticate?method=password", c.ctrlUrl)
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	tr := &http.Transport{
		TLSClientConfig: config,
	}

	payload := map[string]interface{}{"username": c.username, "password": c.password}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	httpClient := &http.Client{Transport: tr}
	req, err := http.NewRequest(http.MethodPost, authUrl, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	token := resp.Header.Get(ztSession)
	if token == "" {
		return errors.New("no token found in header")
	}
	c.token = token
	return nil
}

func (c *Client) CreateIdentity(it IdentityType, name string, isAdmin bool, attributes []string) (string, error) {
	// TODO: for now only supporting ott method of enrollment
	enrollment := map[string]interface{}{"ott": true}
	payload := map[string]interface{}{"enrollment": enrollment, "isAdmin": isAdmin, "name": name, "roleAttributes": attributes, "type": it.String()}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.makeRequest(http.MethodPost, "edge/management/v1/identities", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	val, ok := result["error"]
	if ok {
		cause := val.(map[string]interface{})["cause"]
		reason := cause.(map[string]interface{})["reason"]
		return "", errors.New(reason.(string))
	}
	data := result["data"].(map[string]interface{})
	return data["id"].(string), nil
}

func (c *Client) EnrollIdentity(identityId string) (*config.Config, error) {
	identity, err := c.GetIdentity(identityId)
	if err != nil {
		return nil, err
	}
	tkn, _, err := enroll.ParseToken(string(identity.Data.Enrollment.Ott.Jwt))

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %s", err.Error())
	}

	enFlags := enroll.EnrollmentFlags{
		Token:  tkn,
		KeyAlg: "RSA",
	}
	config, err := enroll.Enroll(enFlags)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Client) GetIdentity(identityId string) (*Identity, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("edge/management/v1/identities/%s", identityId), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var identity Identity
	err = json.NewDecoder(resp.Body).Decode(&identity)
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (c *Client) CreateService(name string, encryptionRequired bool, attributes []string) (string, error) {
	payload := map[string]interface{}{"encryptionRequired": encryptionRequired, "name": name, "roleAttributes": attributes}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.makeRequest(http.MethodPost, "edge/management/v1/services", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return extractId(resp.Body)
}

func (c *Client) DeleteService(id string) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("edge/management/v1/services/%s", id), nil)
	log.Println(resp)
	return err
}

func (c *Client) CreateServicePolicy(policyType ServicePolicyType, name string, identityRoles []string, serviceRoles []string, semantic ServicePolicySemantic) (string, error) {
	payload := map[string]interface{}{"type": policyType.String(), "name": name, "identityRoles": identityRoles, "serviceRoles": serviceRoles, "semantic": semantic.String()}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	resp, err := c.makeRequest(http.MethodPost, "edge/management/v1/service-policies", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return extractId(resp.Body)
}

func (c *Client) makeRequest(method string, url string, body []byte) (*http.Response, error) {
	if c.token == "" {
		return nil, fmt.Errorf("no session found, please login")
	}
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{
		TLSClientConfig: config,
	}

	httpClient := &http.Client{Transport: tr}
	fullUrl := fmt.Sprintf("%s/%s", c.ctrlUrl, url)
	req, err := http.NewRequest(method, fullUrl, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(ztSession, c.token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func extractId(body io.ReadCloser) (string, error) {
	var result map[string]interface{}
	err := json.NewDecoder(body).Decode(&result)
	if err != nil {
		return "", err
	}
	val, ok := result["error"]
	if ok {
		cause := val.(map[string]interface{})["cause"]
		reason := cause.(map[string]interface{})["reason"]
		return "", errors.New(reason.(string))
	}
	data := result["data"].(map[string]interface{})
	return data["id"].(string), nil
}
