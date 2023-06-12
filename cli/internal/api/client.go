package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/nstehr/pitwall/cli/internal/config"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/enroll"
)

type Identity struct {
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
	RoleAttributes          []string    `json:"roleAttributes"`
	SdkInfo                 struct {
	} `json:"sdkInfo"`
	ServiceHostingCosts struct {
	} `json:"serviceHostingCosts"`
	ServiceHostingPrecedences struct {
	} `json:"serviceHostingPrecedences"`
	TypeID string `json:"typeId"`
}

type IdentityType int64
type ServicePolicyType int64
type ServicePolicySemantic int64

type VirtualMachine struct {
	ID             int       `json:"id"`
	Image          string    `json:"image"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	OrchestratorID int       `json:"orchestrator_id"`
	Status         string    `json:"status"`
	PublicKey      string    `json:"public_key"`
	UserID         int       `json:"user_id"`
	Name           string    `json:"name"`
	Services       []Service `json:"services"`
}

type Service struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Port             int       `json:"port"`
	Private          bool      `json:"private"`
	Protocol         string    `json:"protocol"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	VirtualMachineID int       `json:"virtual_machine_id"`
}

type Resp[T any] struct {
	data []byte
}

func (r Resp[T]) PrettyString() (string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, r.data, "", "\t")

	return prettyJSON.String(), err
}

func (r Resp[T]) Parse() (*T, error) {
	out := new(T)
	if err := json.Unmarshal(r.data, out); err != nil {
		return nil, err
	}
	return out, nil
}

func GetVMs(ctx context.Context, cfg *config.Config) (*Resp[[]VirtualMachine], error) {
	client := getAuthenticatedClient(ctx, cfg)
	resp, err := client.Get(fmt.Sprintf("%s/virtual_machines", cfg.ApiEndpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := Resp[[]VirtualMachine]{data: b}
	return &r, err
}

func GetVMByName(ctx context.Context, cfg *config.Config, name string) (*Resp[VirtualMachine], error) {
	client := getAuthenticatedClient(ctx, cfg)
	resp, err := client.Get(fmt.Sprintf("%s/virtual_machines?name=%s", cfg.ApiEndpoint, name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := Resp[VirtualMachine]{data: b}
	return &r, err
}

func generateZitiIdentity(ctx context.Context, cfg *config.Config) (*Identity, error) {
	client := getAuthenticatedClient(ctx, cfg)
	url := fmt.Sprintf("%s/profile/ztIdentity", cfg.ApiEndpoint)
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var identity Identity
	json.Unmarshal(b, &identity)
	return &identity, err
}

func EnrollIdentity(ctx context.Context, cfg *config.Config) (*ziti.Config, error) {
	identity, err := generateZitiIdentity(ctx, cfg)
	if err != nil {
		return nil, err
	}
	tkn, _, err := enroll.ParseToken(string(identity.Enrollment.Ott.Jwt))
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
