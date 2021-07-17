package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/manifoldco/promptui"
)

var isGlobal = flag.Bool("global", false, "Set user as global")
var emailRexp = regexp.MustCompile("^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}$")

const (
	add = "Add a new git user"
	sel = "Select a git user"
	del = "Delete an existing git user"
)

func usage() {
	format := `Usage:
  usegit [flags]

Flags:
  --global              Set user as global.

Author:
  Ahmed Waleed <ahmedwaleed11599@gmail.com>
`
	fmt.Fprintln(os.Stderr, format)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	os.Exit(run())
}

func run() int {
	action := promptui.Select{
		Label: "Select action",
		Items: []string{sel, add, del},
	}

	_, result, err := action.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not perform action: %v\n", err)
	}

	switch result {
	case add:
		if err := addUser(); err != nil {
			fmt.Fprintf(os.Stderr, "could not add new user %v\n", err)
			return 1
		}
	case del:
		if err := deleteUser(); err != nil {
			fmt.Fprintf(os.Stderr, "could not delete new user %v\n", err)
			return 1
		}
	case sel:
		if err := selectUser(); err != nil {
			fmt.Fprintf(os.Stderr, "could not select user %v\n", err)
			return 1
		}
	default:
		fmt.Fprintf(os.Stderr, "invalid operation")
		return 1
	}

	return 0
}

func addUser() error {
	inputName := promptui.Prompt{
		Label: "Enter git user name",
	}

	name, err := inputName.Run()
	if err != nil {
		return err
	}

	inputEmail := promptui.Prompt{
		Label: "Enter git user email",
		Validate: func(email string) error {
			if !emailRexp.MatchString(email) {
				return errors.New("Invalid email address")
			}

			return nil
		},
	}

	email, err := inputEmail.Run()
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	inputSshKeyFile := promptui.Prompt{
		Label: fmt.Sprintf("Enter ssh key file path, default: (%s/.ssh/id_rsa)", homeDir),
	}

	sshKeyPath, err := inputSshKeyFile.Run()
	if err != nil {
		return err
	}

	if sshKeyPath == "" {
		sshKeyPath = fmt.Sprintf("%s/.ssh/id_rsa", homeDir)
	}

	conf, err := NewConfig()
	if err != nil {
		return err
	}

	usr := User{
		Name:           name,
		Email:          email,
		SshKeyFilePath: sshKeyPath,
	}

	if err := conf.Add(usr); err != nil {
		return err
	}

	return nil
}

func deleteUser() error {
	conf, err := NewConfig()
	if err != nil {
		return err
	}

	if len(conf.Users) == 0 {
		fmt.Println("no users found")
	}

	action := promptui.Select{
		Label: "Select a git user",
		Items: stringifyUsers(conf.Users),
	}

	index, _, err := action.Run()
	if err != nil {
		return err
	}

	if err := conf.Remove(index); err != nil {
		return err
	}

	return nil
}

func selectUser() error {
	conf, err := NewConfig()
	if err != nil {
		return err
	}

	if len(conf.Users) == 0 {
		fmt.Println("no users found")
	}

	action := promptui.Select{
		Label: "Select a git user",
		Items: stringifyUsers(conf.Users),
	}

	index, _, err := action.Run()
	if err != nil {
		return err
	}

	option := "--local"
	if *isGlobal {
		option = "--global"
	}

	if err := shellExec("ssh-add", "-D"); err != nil {
		return err
	}
	if err := shellExec("ssh-add", "-D"); err != nil {
		return err
	}
	if err := shellExec("ssh-add", conf.Users[index].SshKeyFilePath); err != nil {
		return err
	}
	if err := shellExec("git", "config", option, "user.name", conf.Users[index].Name); err != nil {
		return err
	}
	if err := shellExec("git", "config", option, "user.email", conf.Users[index].Email); err != nil {
		return err
	}

	return nil
}

func stringifyUsers(users []User) []string {
	items := make([]string, len(users))
	for i, usr := range users {
		items[i] = usr.String()
	}

	return items
}

func shellExec(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
