package ldap

import (
	"github.com/bborbe/auth-http-proxy/model"
	"github.com/golang/glog"
	"github.com/jtblin/go-ldap-client"
)

const ldapConnectionSize = 5

type Authenticator interface {
	Authenticate(model.UserName, model.Password) (bool, map[string]string, error)
	GetGroupsOfUser(model.UserName) ([]string, error)
}

type ldapAuth struct {
	ldapBaseDn       model.LdapBaseDn
	ldapHost         model.LdapHost
	ldapServerName   model.LdapServerName
	ldapPort         model.LdapPort
	ldapUseSSL       model.LdapUseSSL
	ldapSkipTls      model.LdapSkipTls
	ldapBindDN       model.LdapBindDN
	ldapBindPassword model.LdapBindPassword
	ldapUserFilter   model.LdapUserFilter
	ldapGroupFilter  model.LdapGroupFilter
	ldapUserDn       model.LdapUserDn
	ldapGroupDn      model.LdapGroupDn
	ldapUserField    model.LdapUserField
	ldapGroupField   model.LdapGroupField
	ldapClients      chan *ldap.LDAPClient
}

func NewAuthenticator(
	ldapBaseDn model.LdapBaseDn,
	ldapHost model.LdapHost,
	ldapServerName model.LdapServerName,
	ldapPort model.LdapPort,
	ldapUseSSL model.LdapUseSSL,
	ldapSkipTls model.LdapSkipTls,
	ldapBindDN model.LdapBindDN,
	ldapBindPassword model.LdapBindPassword,
	ldapUserDn model.LdapUserDn,
	ldapUserFilter model.LdapUserFilter,
	ldapUserField model.LdapUserField,
	ldapGroupDn model.LdapGroupDn,
	ldapGroupFilter model.LdapGroupFilter,
	ldapGroupField model.LdapGroupField,
) Authenticator {
	a := new(ldapAuth)
	a.ldapBaseDn = ldapBaseDn
	a.ldapHost = ldapHost
	a.ldapServerName = ldapServerName
	a.ldapPort = ldapPort
	a.ldapUseSSL = ldapUseSSL
	a.ldapSkipTls = ldapSkipTls
	a.ldapBindDN = ldapBindDN
	a.ldapBindPassword = ldapBindPassword
	a.ldapUserFilter = ldapUserFilter
	a.ldapGroupFilter = ldapGroupFilter
	a.ldapUserField = ldapUserField
	a.ldapGroupField = ldapGroupField
	a.ldapUserDn = ldapUserDn
	a.ldapGroupDn = ldapGroupDn
	a.ldapClients = make(chan *ldap.LDAPClient, ldapConnectionSize)
	return a
}

func (a *ldapAuth) getClient() *ldap.LDAPClient {
	select {
	case client := <-a.ldapClients:
		glog.V(2).Infof("got client from pool")
		return client
	default:
		glog.V(2).Infof("created new client")
		return a.createClient()
	}
}

func (a *ldapAuth) releaseClient(client *ldap.LDAPClient) {
	glog.V(2).Infof("release client")
	select {
	case a.ldapClients <- client:
		glog.V(2).Infof("returned client to pool")
	default:
		a.closeClient(client)
	}
}
func (a *ldapAuth) closeClient(client *ldap.LDAPClient) {
	glog.V(2).Infof("closed client")
	client.Close()
}

func (a *ldapAuth) Close() {
	glog.V(2).Infof("close all ldap connections")
	for client := range a.ldapClients {
		client.Close()
	}
}

func (a *ldapAuth) Authenticate(username model.UserName, password model.Password) (ok bool, data map[string]string, err error) {
	glog.V(2).Infof("Authenticate user %s", username)
	ldapClient := a.getClient()
	ok, data, err = ldapClient.Authenticate(username.String(), password.String())
	if err != nil {
		glog.V(1).Infof("Authenticate failed, retry with new connection: %v", err)
		a.closeClient(ldapClient)
		ldapClient = a.createClient()
		ok, data, err = ldapClient.Authenticate(username.String(), password.String())
	}
	a.releaseClient(ldapClient)
	return
}

func (a *ldapAuth) GetGroupsOfUser(username model.UserName) (groups []string, err error) {
	glog.V(2).Infof("GetGroupsOfUser for user %s", username)
	ldapClient := a.getClient()
	groups, err = ldapClient.GetGroupsOfUser(username.String())
	if err != nil {
		glog.V(1).Infof("GetGroupsOfUser failed, retry with new connection: %v", err)
		a.closeClient(ldapClient)
		ldapClient = a.createClient()
		groups, err = ldapClient.GetGroupsOfUser(username.String())
	}
	a.releaseClient(ldapClient)
	return
}

func (a *ldapAuth) createClient() *ldap.LDAPClient {
	serverName := a.ldapServerName.String()
	if len(serverName) == 0 {
		serverName = a.ldapHost.String()
	}
	glog.V(2).Infof("create new ldap client for %s:%d with servername %s", a.ldapHost, a.ldapPort, serverName)
	client := &ldap.LDAPClient{
		Base:         a.ldapBaseDn.String(),
		BindDN:       a.ldapBindDN.String(),
		BindPassword: a.ldapBindPassword.String(),
		GroupDN:      a.ldapGroupDn.String(),
		GroupField:   a.ldapGroupField.String(),
		GroupFilter:  a.ldapGroupFilter.String(),
		Host:         a.ldapHost.String(),
		Port:         a.ldapPort.Int(),
		ServerName:   serverName,
		SkipTLS:      a.ldapSkipTls.Bool(),
		UseSSL:       a.ldapUseSSL.Bool(),
		UserDN:       a.ldapUserDn.String(),
		UserField:    a.ldapUserField.String(),
		UserFilter:   a.ldapUserFilter.String(),
	}
	if glog.V(4) {
		glog.Infof("client %+v", client)
	}
	return client
}