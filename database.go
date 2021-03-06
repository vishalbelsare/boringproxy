package boringproxy

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"sync"
)

type Database struct {
	Tokens  map[string]TokenData `json:"tokens"`
	Tunnels map[string]Tunnel    `json:"tunnels"`
	Users   map[string]User      `json:"users"`
	SshKeys map[string]SshKey    `json:"ssh_keys"`
	mutex   *sync.Mutex
}

type TokenData struct {
	Owner string `json:"owner"`
}

type User struct {
	IsAdmin bool                `json:"is_admin"`
	Clients map[string]DbClient `json:"clients"`
}

type SshKey struct {
	Owner string `json:"owner"`
	Key   string `json:"key"`
}

type DbClient struct {
}

type Tunnel struct {
	Owner            string `json:"owner"`
	Domain           string `json:"domain"`
	SshKey           string `json:"ssh_key"`
	ServerAddress    string `json:"server_address"`
	ServerPort       int    `json:"server_port"`
	ServerPublicKey  string `json:"server_public_key"`
	Username         string `json:"username"`
	TunnelPort       int    `json:"tunnel_port"`
	TunnelPrivateKey string `json:"tunnel_private_key"`
	ClientName       string `json:"client_name"`
	ClientAddress    string `json:"client_address"`
	ClientPort       int    `json:"client_port"`
	AllowExternalTcp bool   `json:"allow_external_tcp"`
	AuthUsername     string `json:"auth_username"`
	AuthPassword     string `json:"auth_password"`
	CssId            string `json:"css_id"`
	TlsTermination   string `json:"tls_termination"`
}

func NewDatabase() (*Database, error) {

	dbJson, err := ioutil.ReadFile("boringproxy_db.json")
	if err != nil {
		log.Println("failed reading boringproxy_db.json")
		dbJson = []byte("{}")
	}

	var db *Database

	err = json.Unmarshal(dbJson, &db)
	if err != nil {
		log.Println(err)
		db = &Database{}
	}

	if db.Tokens == nil {
		db.Tokens = make(map[string]TokenData)
	}

	if db.Tunnels == nil {
		db.Tunnels = make(map[string]Tunnel)
	}

	if db.Users == nil {
		db.Users = make(map[string]User)
	}

	if db.SshKeys == nil {
		db.SshKeys = make(map[string]SshKey)
	}

	db.mutex = &sync.Mutex{}

	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.persist()

	return db, nil
}

func (d *Database) AddToken(owner string) (string, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, exists := d.Users[owner]
	if !exists {
		return "", errors.New("Owner doesn't exist")
	}

	token, err := genRandomCode(32)
	if err != nil {
		return "", errors.New("Could not generat token")
	}

	d.Tokens[token] = TokenData{owner}

	d.persist()

	return token, nil
}

func (d *Database) GetTokens() map[string]TokenData {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	tokens := make(map[string]TokenData)

	for k, v := range d.Tokens {
		tokens[k] = v
	}

	return tokens
}

func (d *Database) GetTokenData(token string) (TokenData, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	tokenData, exists := d.Tokens[token]

	if !exists {
		return TokenData{}, false
	}

	return tokenData, true
}

func (d *Database) SetTokenData(token string, tokenData TokenData) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Tokens[token] = tokenData
	d.persist()
}

func (d *Database) DeleteTokenData(token string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.Tokens, token)

	d.persist()
}

func (d *Database) GetTunnels() map[string]Tunnel {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	tunnels := make(map[string]Tunnel)

	for k, v := range d.Tunnels {
		tunnels[k] = v
	}

	return tunnels
}

func (d *Database) GetTunnel(domain string) (Tunnel, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	tun, exists := d.Tunnels[domain]

	if !exists {
		return Tunnel{}, false
	}

	return tun, true
}

func (d *Database) SetTunnel(domain string, tun Tunnel) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Tunnels[domain] = tun
	d.persist()
}

func (d *Database) DeleteTunnel(domain string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.Tunnels, domain)

	d.persist()
}

func (d *Database) GetUsers() map[string]User {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	users := make(map[string]User)

	for k, v := range d.Users {
		users[k] = v
	}

	return users
}

func (d *Database) GetUser(username string) (User, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	user, exists := d.Users[username]

	if !exists {
		return User{}, false
	}

	return user, true
}

func (d *Database) SetUser(username string, user User) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Users[username] = user
	d.persist()

	return nil
}

func (d *Database) AddUser(username string, isAdmin bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, exists := d.Users[username]

	if exists {
		return errors.New("User exists")
	}

	d.Users[username] = User{
		IsAdmin: isAdmin,
		Clients: make(map[string]DbClient),
	}

	d.persist()

	return nil
}

func (d *Database) DeleteUser(username string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.Users, username)

	d.persist()
}

func (d *Database) GetSshKey(id string) (SshKey, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	key, exists := d.SshKeys[id]

	if !exists {
		return SshKey{}, false
	}

	return key, true
}

func (d *Database) GetSshKeys() map[string]SshKey {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	keys := make(map[string]SshKey)

	for k, v := range d.SshKeys {
		keys[k] = v
	}

	return keys
}

func (d *Database) AddSshKey(id string, key SshKey) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, exists := d.SshKeys[id]

	if exists {
		return errors.New("SSH key id exists")
	}

	d.SshKeys[id] = key

	d.persist()

	return nil
}

func (d *Database) DeleteSshKey(id string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.SshKeys, id)

	d.persist()
}

func (d *Database) persist() {
	saveJson(d, "boringproxy_db.json")
}
