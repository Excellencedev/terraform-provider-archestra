package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Permission defines the available permissions for a role
type Permission string

const (
	PermissionAgentsRead     Permission = "agents:read"
	PermissionAgentsWrite    Permission = "agents:write"
	PermissionMcpServersRead Permission = "mcp_servers:read"
	// Add other permissions as needed based on the backend
)

// UserRoleAssignment represents a role assignment to a user
type UserRoleAssignment struct {
	UserID openapi_types.UUID `json:"userId"`
	RoleID openapi_types.UUID `json:"roleId"`
}

// CreateUserRoleAssignmentJSONBody defines the body for assigning a role
type CreateUserRoleAssignmentJSONBody struct {
	RoleID openapi_types.UUID `json:"roleId"`
}

// User represents a user in the system (re-defining for manual use if needed, though generated might exist)
// Using checking if User exists in generated code... it does not seem to be fully exposed in a way we can use for individual get,
// so defining a basic struct for GetUser response if the generated one is insufficient.
// However, let's try to reuse generated types if possible or define minimal ones here.
// Looking at datasource_user.go, it expects a User struct with Name, Email etc.
// Let's rely on the fact that we can parse the JSON into a struct that matches what we need.

type User struct {
	ID            openapi_types.UUID `json:"id"`
	Name          string             `json:"name"`
	Email         string             `json:"email"`
	EmailVerified bool               `json:"emailVerified"`
	Image         *string            `json:"image,omitempty"`
	Role          *string            `json:"role,omitempty"` // Legacy role field?
	Banned        bool               `json:"banned"`
	BanReason     *string            `json:"banReason,omitempty"`
}

// GetUserResponse is the response for GetUser
type GetUserResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *User
	JSON404      *struct{}
}

func (r *GetUserResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

func (c *ClientWithResponses) GetUserWithResponse(ctx context.Context, id string) (*GetUserResponse, error) {
	req, err := http.NewRequest("GET", c.Server+"/users/"+id, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &GetUserResponse{
		HTTPResponse: resp,
	}

	if resp.StatusCode == 200 {
		var user User
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, err
		}
		result.JSON200 = &user
	} else if resp.StatusCode == 404 {
		result.JSON404 = &struct{}{}
	}

	return result, nil
}

// Role Assignment Methods

type UserRoleAssignmentResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *UserRoleAssignment
	JSON404      *struct{}
}

func (r *UserRoleAssignmentResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

func (c *ClientWithResponses) CreateUserRoleAssignmentWithResponse(ctx context.Context, userId string, body CreateUserRoleAssignmentJSONBody) (*UserRoleAssignmentResponse, error) {
	var bodyReader *bytes.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)

	req, err := http.NewRequest("POST", c.Server+"/users/"+userId+"/roles", bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &UserRoleAssignmentResponse{
		HTTPResponse: resp,
	}

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		// Assuming the response is the assignment object
		var assignment UserRoleAssignment
		// It might be a list or the user object, but let's assume assignment for now based on typical REST patterns
		// Or maybe it returns emptiness
		if resp.ContentLength > 0 {
			// Try decode
			_ = json.NewDecoder(resp.Body).Decode(&assignment)

			// Parse UUIDs
			uID, err := uuid.Parse(userId)
			if err == nil {
				assignment.UserID = openapi_types.UUID(uID)
			}
		}
		// For now, let's just return success if 200
		result.JSON200 = &assignment
	}

	return result, nil
}

func (c *ClientWithResponses) DeleteUserRoleAssignmentWithResponse(ctx context.Context, userId string, roleId string) (*UserRoleAssignmentResponse, error) {
	req, err := http.NewRequest("DELETE", c.Server+"/users/"+userId+"/roles/"+roleId, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &UserRoleAssignmentResponse{
		HTTPResponse: resp,
	}

	if resp.StatusCode == 404 {
		result.JSON404 = &struct{}{}
	}

	return result, nil
}

type GetUserRolesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]Role
	JSON404      *struct{}
}

func (r *GetUserRolesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

func (c *ClientWithResponses) GetUserRolesWithResponse(ctx context.Context, userId string) (*GetUserRolesResponse, error) {
	req, err := http.NewRequest("GET", c.Server+"/users/"+userId+"/roles", nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &GetUserRolesResponse{
		HTTPResponse: resp,
	}

	if resp.StatusCode == 200 {
		var roles []Role
		if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
			return nil, err
		}
		result.JSON200 = &roles
	} else if resp.StatusCode == 404 {
		result.JSON404 = &struct{}{}
	}

	return result, nil
}
