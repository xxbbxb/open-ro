package main

import (
	"fmt"
	"regexp"
	"strings"
)

func handleUpdateAsCommand(t *TelegramBot, u Update) {
	norm := regexp.MustCompile(`\s+`)
	u.Msg.Text = norm.ReplaceAllString(u.Msg.Text, " ")
	command := strings.Split(u.Msg.Text, " ")
	var c Command
	switch strings.ToLower(command[0]) {
	case "account":
		for len(command) < 4 {
			command = append(command, "")
		}
		c = AccountCmd{
			Name:   command[1],
			Passwd: command[2],
			Sex:    command[3],
			Author: fmt.Sprintf("%d@t.me", u.Msg.User.Id),
		}

	case "accounts":
		c = AccountsCmd{
			Author: fmt.Sprintf("%d@t.me", u.Msg.User.Id),
		}
	default:
		c = &UnknownCmd{
			name: command[0],
		}
	}
	ok, message := c.Process()
	if ok {
		message = fmt.Sprintf("✅ %s", message)
	} else {
		message = fmt.Sprintf("❕ %s", message)
	}
	t.SendMessage(MessageToSend{
		ChatId:                u.Msg.Chat.Id,
		Text:                  message,
		DisableWebPagePreview: true,
	})
}

const defaultResponse = "database error, please try later"

type Command interface {
	Process() (bool, string)
}

type UnknownCmd struct {
	name string
}

func (c *UnknownCmd) Process() (bool, string) {
	return false, fmt.Sprintf("Sorry, bot does not know command: %s. Available commands: account, accounts", c.name)
}

type AccountCmd struct {
	Name   string `db:"userid"`
	Passwd string `db:"user_pass"`
	Sex    string `db:"sex"`
	Author string `db:"email"`
}

func (c AccountCmd) isValid() (bool, string) {
	var alerts []string
	if !(c.Sex == "M" || c.Sex == "F" || c.Sex == "") {
		alerts = append(alerts, "sex must be one of M, F or empty")
	}
	if matched, _ := regexp.MatchString("[A-Za-z0-9_]+", c.Name); len(c.Name) < 6 || len(c.Name) > 23 || !matched {
		alerts = append(alerts, "login must be 6 to 23 characters A-Za-z0-9_")
	}
	if len(c.Passwd) < 6 {
		alerts = append(alerts, "password must contain at least 6 characters")
	}
	if matched, _ := regexp.MatchString("[A-Za-z0-9_]+", c.Name); len(c.Author) == 0 || !matched {
		log.Info("author id can not be empty")
	}
	if len(alerts) > 0 {
		alerts = append(alerts, "---")
		alerts = append(alerts, "create account: account login password sex")
		alerts = append(alerts, "reset password: account login new-password")
		return false, strings.Join(alerts, "\n")
	}
	return true, ""
}

func (c AccountCmd) Process() (bool, string) {
	if valid, message := c.isValid(); !valid {
		return valid, message
	}
	l := log.WithField("email", c.Author).WithField("account_name", c.Name).WithField("component", "DB")
	var (
		err error
	)
	con, err := newDB()
	if err != nil {
		l.WithError(err).Error("unable to connect database")
		return false, defaultResponse
	}
	var numAccounts int
	if err := con.QueryRowx(`SELECT COUNT(account_id) FROM login WHERE email = ?`, c.Author).Scan(&numAccounts); err != nil {
		l.WithError(err).Error("unable to get number of accounts")
		return false, defaultResponse
	}
	if numAccounts > 30 {
		return false, "error: 30 accounts per user limit reached"
	}
	rows, err := con.NamedQuery(`SELECT userid, user_pass, sex, email FROM login WHERE userid=:userid LIMIT 1`, c)
	if err != nil {
		l.WithError(err).Error("unable to check account presence")
		return false, defaultResponse
	}
	defer rows.Close()
	if rows.Next() {
		var row AccountCmd
		// TODO: need to add locks, but race chance is very low
		err = rows.StructScan(&row)
		if err != nil {
			l.WithError(err).Error("unable to read account record from database")
			return false, fmt.Sprint(err)
		}
		if row.Author != c.Author {
			return false, "account name is already taken"
		}
		_, err = con.NamedExec(`UPDATE login SET user_pass=:user_pass where userid=:userid`, c)
		if err != nil {
			l.WithError(err).Error("unable to update password")
			return false, fmt.Sprint(err)
		}
		return true, fmt.Sprintf("account %s password updated", c.Name)
	} else {
		if !(c.Sex == "M" || c.Sex == "F") {
			return false, "sex must be one of M, F"
		}
		_, err := con.NamedExec(`INSERT INTO login (userid, user_pass, sex, email) VALUES (:userid, :user_pass, :sex, :email)`, c)
		if err != nil {
			l.WithError(err).Error("unable to create new acccount")
			return false, fmt.Sprint(err)
		}
		return true, fmt.Sprintf("account %s created", c.Name)
	}
}

type AccountsCmd struct {
	Author string `db:"email"`
}

func (c AccountsCmd) Process() (bool, string) {
	l := log.WithField("email", c.Author).WithField("component", "DB")
	var (
		err error
	)
	con, err := newDB()
	if err != nil {
		l.WithError(err).Error("unable to connect database")
		return false, defaultResponse
	}
	rows, err := con.NamedQuery(`SELECT userid, user_pass, sex, email FROM login WHERE email=:email`, c)
	if err != nil {
		l.WithError(err).Error("unable to check account presence")
		return false, defaultResponse
	}
	defer rows.Close()
	a := []string{"Your accounts:"}
	for rows.Next() {
		var row AccountCmd
		err = rows.StructScan(&row)
		if err != nil {
			l.WithError(err).Error("unable to read account row")
			return false, defaultResponse
		}
		a = append(a, row.Name)
	}
	if len(a) == 0 {
		return false, "You don't have game accounts yet."
	}
	return true, strings.Join(a, "\n")
}
