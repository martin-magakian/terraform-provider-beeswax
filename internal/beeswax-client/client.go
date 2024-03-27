package beeswax

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
)

type User struct {
	ID               int64  `json:"id"`
	SuperUser        bool   `json:"super_user"`
	Email            string `json:"email"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	RoleID           int64  `json:"role_id"`
	AccountID        int64  `json:"account_id"`
	Active           bool   `json:"active"`
	AllAccountAccess bool   `json:"all_account_access"`
	AccountGroupIDs  []int  `json:"account_group_ids"`
}

type Role struct {
	ID                   int64        `json:"id"`
	Name                 string       `json:"name"`
	ParentRoleID         int64        `json:"parent_role_id"`
	Archived             bool         `json:"archived"`
	Notes                string       `json:"notes"`
	SharedAcrossAccounts bool         `json:"shared_across_accounts"`
	Permissions          []Permission `json:"permissions"`
	ReportIDs            []int        `json:"report_ids"`
}

type Permission struct {
	ObjectType string `json:"object_type"`
	Permission int64  `json:"permission"`
}

type Client struct {
	APIURL   string
	email    string
	password string
	client   *http.Client
}

func NewClient(apiURL, email, password string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		APIURL:   apiURL,
		email:    email,
		password: password,
		client: &http.Client{
			Jar: jar, // will keep the cookie to stay logged in
		},
	}
}

func (bx *Client) Login() error {
	loginPayload, err := json.Marshal(map[string]string{"email": bx.email, "password": bx.password})
	if err != nil {
		return err
	}
	resp, err := bx.client.Post(bx.APIURL+"/rest/v2/authenticate", "application/json", bytes.NewBuffer(loginPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("login failed")
	}
	// Authentication cookie is stored in the client
	return nil
}

func (bx *Client) request(method, path string, data interface{}) ([]byte, error) {
	dataPayload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshall: %w", err)
	}

	req, err := http.NewRequest(method, bx.APIURL+path, bytes.NewBuffer([]byte(dataPayload)))
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := bx.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	bodyStr, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read body response: %w", err)
	}

	// Manage error responses
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return bodyStr, errors.New("response " + resp.Status + " instead. API response: " + string(bodyStr))
	}

	return bodyStr, nil
}

func (bx *Client) GetUser(userID int64) (User, error) {
	response, err := bx.request("GET", fmt.Sprintf("/rest/v2/users/%d", userID), "")
	if err != nil {
		return User{}, err
	}
	user := User{}
	err = json.Unmarshal(response, &user)
	return user, err
}

func (bx *Client) CreateUser(user User) (int64, error) {
	response, err := bx.request("POST", "/rest/v2/users", user)
	if err != nil {
		return 0, err
	}
	createdUser := User{}
	err = json.Unmarshal(response, &createdUser)
	return createdUser.ID, err
}

func (bx *Client) UpdateUser(user User) error {
	_, err := bx.request("PUT", fmt.Sprintf("/rest/v2/users/%d", user.ID), user)
	return err
}

func (bx *Client) DeleteUser(userID int64) error {
	_, err := bx.request("DELETE", fmt.Sprintf("/rest/v2/users/%d", userID), "")
	return err
}

func (bx *Client) GetRole(roleID int64) (Role, error) {
	response, err := bx.request("GET", fmt.Sprintf("/rest/v2/roles/%d", roleID), "")
	if err != nil {
		return Role{}, err
	}
	role := Role{}
	err = json.Unmarshal(response, &role)
	return role, err
}

func (bx *Client) CreateRole(role Role) (int64, error) {
	response, err := bx.request("POST", "/rest/v2/roles", role)
	if err != nil {
		return 0, err
	}
	createdRole := Role{}
	err = json.Unmarshal(response, &createdRole)
	return createdRole.ID, err
}

func (bx *Client) UpdateRole(role Role) error {
	_, err := bx.request("PUT", fmt.Sprintf("/rest/v2/roles/%d", role.ID), role)
	return err
}

func (bx *Client) DeleteRole(roleID int64) error {
	_, err := bx.request("DELETE", fmt.Sprintf("/rest/v2/roles/%d", roleID), "")
	return err
}
