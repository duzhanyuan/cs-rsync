package account

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type AccountInfoer struct {
	Username string
	Password string
}

var accountInfo map[string]AccountInfoer

func init() {
	accountInfo = make(map[string]AccountInfoer)
	InitAccountInfo()
}

func InitAccountInfo() {
	path := "../data/account_info.json"
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(file, &accountInfo)
	if err != nil {
		fmt.Println(err)
	}
}

func GetAccountInfo(username string) AccountInfoer {

	if _, ok := accountInfo[username]; ok {
		return accountInfo[username]
	} else {
		return AccountInfoer{}
	}
}

func CheckPassword(username, password string) bool {

	if val, ok := accountInfo[username]; ok {
		if val.Password == "*" {
			return true
		}
		md5h := md5.New()
		io.WriteString(md5h, password)
		password = fmt.Sprintf("%X", md5h.Sum(nil))
		if password == val.Password {
			return true
		}
	}
	return false
}

// todo
func writeAccountInfo() error {
	path := "../data/account_info.json"
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
	}
	if err != nil {
		fmt.Println(err)
		file, err = os.Create(path)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	defer func() {
		err = file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	jsonByte, err := json.MarshalIndent(&accountInfo, "", "	")
	if err != nil {
		fmt.Println("writeAccountInfo-json.MarshalIndent:", err)
		return err
	}

	jsonStr := string(jsonByte)

	file.Truncate(int64(len(jsonStr)))
	_, err = io.WriteString(file, jsonStr)

	if err != nil {
		fmt.Println("writeAccountInfo-io.WriteString:", err)
		return nil
	}

	return nil
}

func UpdatePassword(username, password, newPassword string) bool {
	if CheckPassword(username, password) {
		md5h := md5.New()
		io.WriteString(md5h, newPassword)
		newPassword = fmt.Sprintf("%X", md5h.Sum(nil))
		delete(accountInfo, username)
		accountInfo[username] = AccountInfoer{Username:username, Password:newPassword}
		writeAccountInfo()
		return true
	}
	return false
}