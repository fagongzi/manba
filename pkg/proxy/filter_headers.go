package proxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/log"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

type GatewayClaims struct {
	HttpBody string `json:"httpBody"`
	jwt.StandardClaims
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// HeadersFilter HeadersFilter
type HeadersFilter struct {
	filter.BaseFilter
}

func newHeadersFilter() filter.Filter {
	return &HeadersFilter{}
}

// Name return name of this filter
func (f HeadersFilter) Name() string {
	return FilterHeader
}

// Pre execute before proxy
func (f HeadersFilter) Pre(c filter.Context) (statusCode int, err error) {

	//check===================================================start
	statusCode, err = validate(c)
	if statusCode != 200 {
		return statusCode, err
	}
	//check====================================================end

	for _, h := range hopHeaders {
		c.GetProxyOuterRequest().Header.Del(h)
	}

	return f.BaseFilter.Pre(c)
}

// Post execute after proxy
func (f HeadersFilter) Post(c filter.Context) (statusCode int, err error) {
	for _, h := range hopHeaders {
		c.GetProxyResponse().Header.Del(h)
	}

	// 需要合并处理的，不做header的复制，由proxy做合并
	if !c.NeedMerge() {
		c.GetOriginRequestCtx().Response.Header.Reset()
		c.GetProxyResponse().Header.CopyTo(&c.GetOriginRequestCtx().Response.Header)
	}

	return f.BaseFilter.Post(c)
}

// validity of the request data
func validate(c filter.Context) (statusCode int, err error) {

	cookie := c.GetProxyOuterRequest().Header.Cookie("Auth")

	log.Info(string(cookie))

	// gets the stream of request parameters；	request type is application/x-www-form-urlencoded
	c.GetProxyOuterRequest().Body()

	textBodyStream := c.GetProxyOuterRequest().Body()

	log.Infof("The requested httpBody Stream is: %v", textBodyStream)

	textBodyString := string(textBodyStream[:])

	log.Infof("The requested httpBody String is: %v", textBodyString)

	var password string

	token, err := jwt.ParseWithClaims(string(cookie), &GatewayClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Infof("Unexpected signing method %v", token.Header["alg"])
				return nil, nil
			}

			log.Infof("Enter ParseWithClaims method")

			password = finduserbyName(token.Claims.(*GatewayClaims).Issuer)

			log.Infof("return password by finduserbyName is: %v", password)

			return []byte(password), nil
		})

	log.Infof("The requested JWT's nonce is: %v", token.Claims.(*GatewayClaims).Id)

	// nonce is exist
	if checkNonce(token.Claims.(*GatewayClaims).Id) != 1 {
		log.Infof("This nonce has been used")
		return 201, errors.New("This nonce has been used")
	}

	log.Infof("The requested JWT'httpBody is: %v", token.Claims.(*GatewayClaims).HttpBody)
	log.Infof("The requested HttpBody Have ComputeHmac256 then String is: %v", ComputeHmac256(textBodyString, password))

	if token.Claims.(*GatewayClaims).HttpBody != ComputeHmac256(textBodyString, password) {
		log.Infof("User body has been modified")
		return 203, errors.New("User body has been modified")
	}

	if _, ok := token.Claims.(*GatewayClaims); ok && token.Valid {
		log.Infof("User Redirect Success")
		return 200, errors.New("success")

	} else {
		log.Infof("User password is wrong")
		return 202, errors.New("User password is wrong")
	}
}

// Check whether nonce exists；default 1（not exist nonce,execute insert);  2:exist 3: error
func checkNonce(gatewayNonce string) (nonceStatus int) {
	db, err := sql.Open("mysql", "root:pass123word01@tcp(172.16.192.91:3308)/gateway?charset=utf8")
	nonceStatus = 1 //default 1 is not exist
	if err != nil {
		log.Info(err)
		nonceStatus = 3
		return nonceStatus
	}

	defer db.Close()

	var rows *sql.Rows
	rows, err = db.Query("select * from gateway_nonce")
	if err != nil {
		log.Info(err)
		nonceStatus = 3
		return nonceStatus
	}

	for rows.Next() {
		var nonce string
		var id int
		rows.Scan(&id, &nonce)
		log.Infof("The Method checkNonce id is: %v", id)
		log.Infof("The Method checkNonce nonce is: %v", nonce)
		//if nonce hava been black dont insert table
		if strings.EqualFold(nonce, gatewayNonce) {
			nonceStatus = 2 // status 2 is exist
			break
		}
	}

	if nonceStatus == 1 {
		var result sql.Result
		result, err = db.Exec("insert into gateway_nonce(nonce) values(?)", gatewayNonce)
		if err != nil {
			log.Info(err)
			nonceStatus = 3
			return nonceStatus
		}
		lastId, _ := result.LastInsertId()
		log.Info("The Method checkNonce insert record's : %v", lastId)

	} else {
		log.Info("this nonce is exist!!!")
	}
	rows.Close()
	return nonceStatus
}

//find password by name
func finduserbyName(name string) (password string) {
	db, err := sql.Open("mysql", "root:pass123word01@tcp(172.16.192.91:3308)/gateway?charset=utf8")
	password = ""
	if err != nil {
		log.Info(err)
		return password
	}

	defer db.Close()

	var rows *sql.Rows
	log.Info("The Method finduserbyName Request Name is", name)

	rows, err = db.Query("select * from gateway_user where username = ?", name)
	if err != nil {
		log.Info(err)
		return password
	}
	for rows.Next() {
		var username string
		var id int
		var userpassword string
		var createtime string
		var updatetime string
		var status int
		rows.Scan(&id, &username, &userpassword, &createtime, &updatetime, &status)
		log.Info(id, "\t", username, "\t", userpassword)
		if !strings.EqualFold(userpassword, "") {
			password = userpassword
		}
	}
	return password
}

// HS256 Sign the string by key
func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	stringsVal := base64.StdEncoding.EncodeToString(h.Sum(nil))
	str := strings.Replace(stringsVal, "=", "", -1)
	str1 := strings.Replace(str, "+", "-", -1)
	str2 := strings.Replace(str1, "/", "_", -1)
	return str2
}
