package scim

import (
	"testing"
	"time"

	"github.com/turtacn/QuantaID/pkg/scim"
	"github.com/turtacn/QuantaID/pkg/types"
)

func TestToDomainUser(t *testing.T) {
	sUser := &scim.User{
		UserName:   "bjensen",
		ExternalID: "bjensen",
		Active:     true,
		Emails: []scim.Email{
			{Value: "bjensen@example.com", Type: "work", Primary: true},
		},
		Name: &scim.Name{
			GivenName:  "Barbara",
			FamilyName: "Jensen",
		},
	}

	dUser := ToDomainUser(sUser)

	if dUser.Username != "bjensen" {
		t.Errorf("Expected username bjensen, got %s", dUser.Username)
	}
	if dUser.Email != "bjensen@example.com" {
		t.Errorf("Expected email bjensen@example.com, got %s", dUser.Email)
	}
	if dUser.Status != types.UserStatusActive {
		t.Errorf("Expected status active, got %s", dUser.Status)
	}
	if dUser.Attributes["externalId"] != "bjensen" {
		t.Errorf("Expected externalId bjensen, got %v", dUser.Attributes["externalId"])
	}

	nameMap, ok := dUser.Attributes["name"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected name map in attributes")
	}
	if nameMap["givenName"] != "Barbara" {
		t.Errorf("Expected givenName Barbara, got %v", nameMap["givenName"])
	}
}

func TestToSCIMUser(t *testing.T) {
	now := time.Now()
	dUser := &types.User{
		ID:        "2819c223-7f76-453a-919d-413861904646",
		Username:  "bjensen",
		Email:     "bjensen@example.com",
		Status:    types.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
		Attributes: map[string]interface{}{
			"externalId": "bjensen",
			"name": map[string]interface{}{
				"givenName":  "Barbara",
				"familyName": "Jensen",
			},
		},
	}

	sUser := ToSCIMUser(dUser)

	if sUser.ID != dUser.ID {
		t.Errorf("Expected ID %s, got %s", dUser.ID, sUser.ID)
	}
	if sUser.UserName != "bjensen" {
		t.Errorf("Expected username bjensen, got %s", sUser.UserName)
	}
	if sUser.ExternalID != "bjensen" {
		t.Errorf("Expected externalId bjensen, got %s", sUser.ExternalID)
	}
	if !sUser.Active {
		t.Error("Expected active true")
	}
	if len(sUser.Emails) != 1 || sUser.Emails[0].Value != "bjensen@example.com" {
		t.Error("Expected email bjensen@example.com")
	}
	if sUser.Name.GivenName != "Barbara" {
		t.Errorf("Expected givenName Barbara, got %s", sUser.Name.GivenName)
	}
	if sUser.Meta.ResourceType != "User" {
		t.Errorf("Expected ResourceType User, got %s", sUser.Meta.ResourceType)
	}
}
