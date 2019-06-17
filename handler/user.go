package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	dblayer "github.com/yaltachen/emulate/BD_Disk/FILESTORE-SERVER/db"
	"github.com/yaltachen/emulate/BD_Disk/FILESTORE-SERVER/util"
)

const (
	pwd_salt = ".#890"
)

// SignupHandler: 用户注册
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// response signup html
		SignupHandlerGET(w, r)
	} else {
		SignUpHandlerPOST(w, r)
	}
}
func SignupHandlerGET(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		data []byte
	)
	if data, err = ioutil.ReadFile("./static/view/signup.html"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
	return
}
func SignUpHandlerPOST(w http.ResponseWriter, r *http.Request) {
	var (
		username string
		password string
		encpwd   string
		suc      bool
	)
	r.ParseForm()

	username = r.Form.Get("username")
	password = r.Form.Get("password")

	if len(username) < 3 || len(password) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}

	encpwd = util.Sha1([]byte(password + pwd_salt))
	if suc = dblayer.UserSignup(username, encpwd); suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

// SigninHandler: 用户登录
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 校验用户

	var (
		username string
		password string
		encpwd   string
		same     bool
		token    string
	)

	r.ParseForm()

	username = r.PostForm["username"][0]
	password = r.PostForm["password"][0]

	encpwd = util.Sha1([]byte(password + pwd_salt))

	if same = dblayer.UserSignin(username, encpwd); !same {
		w.Write([]byte("FAILED"))
		return
	}
	// 2. 生成访问凭证（token）

	token = GenToken(username)
	if !dblayer.UpdateToken(username, token) {
		w.Write([]byte("FAILED"))
		return
	}
	// 3. 登录后定向到首页

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username string
		// token    string
		user *dblayer.TableUser
		err  error
	)
	// 1. 解析请求参数
	r.ParseForm()
	username = r.Form.Get("username")
	// token = r.Form.Get("token")

	// 2. 验证token,由拦截器处理
	// if !isTokenValid(token) {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	return
	// }

	// 3. 查询用户
	if user, err = dblayer.GetUserInfo(username); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 4. 组装并响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

func GenToken(username string) string {
	// 40位字符
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// isTokenValid：验证token
func isTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token时效性
	// TODO: 从tbl_user_token查询username对应的token信息
	// TODO: token与数据库表中username对应的token一致
	return true
}
