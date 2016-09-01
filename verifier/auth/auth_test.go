package auth

import (
	"testing"

	"fmt"

	"os"

	. "github.com/bborbe/assert"
	auth_model "github.com/bborbe/auth/model"
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/golang/glog"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestVerifyFailed(t *testing.T) {
	username := model.UserName("user123")
	password := model.Password("pass123")

	authVerifier := New(func(authToken auth_model.AuthToken, requiredGroups []auth_model.GroupName) (*auth_model.UserName, error) {
		return nil, fmt.Errorf("not found")
	}, model.GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NotNilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(false)); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyNotFound(t *testing.T) {
	username := model.UserName("user123")
	password := model.Password("pass123")

	authVerifier := New(func(authToken auth_model.AuthToken, requiredGroups []auth_model.GroupName) (*auth_model.UserName, error) {
		u := auth_model.UserName("")
		return &u, nil
	}, model.GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(false)); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyFound(t *testing.T) {
	username := model.UserName("user123")
	password := model.Password("pass123")

	authVerifier := New(func(authToken auth_model.AuthToken, requiredGroups []auth_model.GroupName) (*auth_model.UserName, error) {
		u := auth_model.UserName(username)
		return &u, nil
	}, model.GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(true)); err != nil {
		t.Fatal(err)
	}
}