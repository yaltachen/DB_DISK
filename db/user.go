package db

import (
	"log"

	mydb "github.com/yaltachen/BD_Disk/db/mysql"
)

// TableUser : 用户表model
type TableUser struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// UserSignup : 通过用户名及密码完成user表的注册操作
func UserSignup(username string, passwd string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if err != nil {
		log.Println("Failed to insert, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, passwd)
	if err != nil {
		log.Println("Failed to insert, err:" + err.Error())
		return false
	}
	if rowsAffected, err := ret.RowsAffected(); nil == err && rowsAffected > 0 {
		return true
	}
	// 已注册
	return false
}

// UserSignin : 判断密码是否一致
func UserSignin(username string, encpwd string) bool {
	stmt, err := mydb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		log.Println(err.Error())
		return false
	} else if rows == nil {
		// 未注册
		log.Println("username not found: " + username)
		return false
	}

	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	// 用户名密码不匹配
	return false
}

// UpdateToken : 刷新用户登录的token
func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

// GetUserInfo : 查询用户信息
func GetUserInfo(username string) (*TableUser, error) {
	user := TableUser{}

	stmt, err := mydb.DBConn().Prepare(
		"select user_name,signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		log.Println(err.Error())
		// error不为nil, 返回时user应当置为nil
		//return user, err
		return nil, err
	}
	defer stmt.Close()

	// 执行查询的操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UserExist : 查询用户是否存在
func UserExist(username string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"select 1 from tbl_user where user_name=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)

	if err != nil {
		return false
	}

	return rows.Next()
}
