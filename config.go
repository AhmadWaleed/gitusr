package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type User struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	SshKeyFilePath string `json:"ssh_key_file_path"`
}

func (usr *User) String() string {
	return fmt.Sprintf("%s\t<%s> (%s)", usr.Name, usr.Email, usr.SshKeyFilePath)
}

type Config struct {
	Users    []User `json:"Users"`
	Filename string
}

func NewConfig() (Config, error) {
	c := Config{Users: []User{}, Filename: "config.json"}

	dir, err := c.Directory()
	if err != nil {
		return c, err
	}

	if !c.Exist() {
		if err := os.Mkdir(dir, 0744); err != nil {
			return c, err
		}
	}

	c.Read()

	return c, nil
}

func (c *Config) Directory() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "use-git"), nil
}

func (c *Config) Path() (string, error) {
	dir, err := c.Directory()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, c.Filename), nil
}

func (c *Config) Exist() bool {
	dir, err := c.Directory()
	if err != nil {
		return false
	}

	_, err = os.Stat(dir)

	return err == nil
}

func (c *Config) Add(user User) error {
	for _, usr := range c.Users {
		if user.Email == usr.Email {
			return errors.New("user already exist.")
		}
	}

	c.Users = append(c.Users, user)
	if err := c.Save(); err != nil {
		return err
	}

	return nil
}

func (c *Config) Remove(idx int) error {
	newUsers := make([]User, len(c.Users)-1)
	if idx+1 == len(c.Users) {
		newUsers = c.Users[:idx]
	} else {
		newUsers = append(c.Users[:idx], c.Users[idx+1:]...)
	}

	c.Users = newUsers

	if err := c.Save(); err != nil {
		return err
	}

	return nil
}

func (c *Config) Save() error {
	path, err := c.Path()
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	if _, err := file.Write(content); err != nil {
		return err
	}

	return nil
}

func (c *Config) Read() error {
	path, err := c.Path()
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, c); err != nil {
		return err
	}

	return nil
}
