package proxy

import "github.com/dgrijalva/jwt-go"
import "strings"
// import "encoding/json"
import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

import (
	"errors"
	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/log"
)


type MyCustomClaims struct {
	jwt.StandardClaims
}

type GatewayResponse struct {
	GatewayCode string `json:"gatewayCode"`
    GatewayMsg  string `json:"gatewayMsg"`
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
	log.Infof("555555")
	log.Infof(string(c.GetProxyOuterRequest().Header.Cookie("Auth")))
	
	//check==================================yinzf==================start
	statusCode,err = validate(c)
	if(statusCode !=200){
		return statusCode,err
	}
	
	// gatewayResponse := GatewayResponse{
	// 	GatewayCode: string(statusCode), 
	// 	GatewayMsg: "xxxxx",
	// }
	// 	jsons1, errs := json.Marshal(gatewayResponse)

	// 	if errs != nil {
	// 		fmt.Println(errs.Error())
	// 	}
	// 	fmt.Println(string(jsons1))

	// 	// return jsons1,err
	// 	// c.resp([]byte(jsons))
	// 	//		res.Write([]byte(jsons))
	// }
	//check==================================yinzf==================end

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


func validate(c filter.Context) (statusCode int, err error) {

		log.Infof("aaaaaaaaaaa")
		cookie := c.GetProxyOuterRequest().Header.Cookie("Auth")
		log.Infof("bbbbbbbbbbb" )
		// splitCookie := strings.Split(string(cookie), "Auth=")
		
		// log.Infof(splitCookie[1])
		log.Infof("XXXXXXX")
		log.Infof(string(cookie))
		token, err := jwt.ParseWithClaims(string(cookie), &MyCustomClaims{}, 
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method %v", token.Header["alg"])
			}
			log.Infof("cccccccccccc")
			var password string
			password = finduserbyName(token.Claims.(*MyCustomClaims).Issuer)
			fmt.Println(password)
			return []byte(password), nil
		})
		log.Infof("ddddddddddddd")
		log.Infof(token.Claims.(*MyCustomClaims).Id)
		// nonce is exist
		if checkNonce(token.Claims.(*MyCustomClaims).Id) != 1 {
			log.Infof("eeeeeeeeee")
			// gatewayResponse := GatewayResponse{
			// 	GatewayCode: "0001",
			// 	GatewayMsg: "此nonce值已经使用过了",
			// }
			// jsons, errs := json.Marshal(gatewayResponse) 
            // if errs != nil {
            //   fmt.Println(errs.Error())
            // }
            // fmt.Println(gatewayResponse)
            // fmt.Println(string(b))
			// res.Write([]byte(jsons))
			// http.NotFound(res, req)
			return 201,errors.New("此nonce值已经使用过了")
		}
		log.Infof("fffffffffff")

		if _, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
			log.Infof("ggggggggg")
			// context.Set(req, "Claims", claims)
			// http.NotFound(res, req)
			// gatewayResponse := GatewayResponse{
			// 	GatewayCode: "0000",
			// 	GatewayMsg:  "跳转成功",
            // }
            // jsons, errs := json.Marshal(gatewayResponse) 
            // if errs != nil {
            //   fmt.Println(errs.Error())
            // }
			// fmt.Println(gatewayResponse)
			// res.Write([]byte(jsons))
			return 200,errors.New("跳转成功")

		} else {
			log.Infof("hhhhhhhhhhhhhh")
			// http.NotFound(res, req)
			// gatewayResponse := GatewayResponse{
			// 	GatewayCode: "0002",
			// 	GatewayMsg:  "用户提供密码有误",
			// }
			// jsons, errs := json.Marshal(gatewayResponse) 
            // if errs != nil {
            //   fmt.Println(errs.Error())
            // }
			// fmt.Println(gatewayResponse)
			// res.Write([]byte(jsons))

			return 202,errors.New("用户提供密码有误")
		}
}

//check nonce does it exist ；default 1（not exist nonce,execute insert);  2:exist 3: error
func checkNonce(gatewayNonce string) (nonceStatus int) {
	db, err := sql.Open("mysql", "root:pass123word01@tcp(172.16.192.91:3308)/gateway?charset=utf8")
	nonceStatus = 1 //default 1 is not exist
	if err != nil {
		fmt.Println(err)
		nonceStatus = 3
		return nonceStatus
	}

	defer db.Close()

	var rows *sql.Rows
	rows, err = db.Query("select * from gateway_nonce")
	if err != nil {
		fmt.Println(err)
		nonceStatus = 3
		return nonceStatus
	}

	for rows.Next() {
		var nonce string
		var id int
		rows.Scan(&id, &nonce)
		fmt.Println(id, "\t", nonce)
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
			fmt.Println(err)
			nonceStatus = 3
			return nonceStatus
		}
		lastId, _ := result.LastInsertId()
		fmt.Println("insert record's id:", lastId)

	} else {
		fmt.Println("this nonce is exist!!!")
	}
	rows.Close()
	return nonceStatus
}

//find user information
func finduserbyName(name string) (password string) {
	db, err := sql.Open("mysql", "root:pass123word01@tcp(172.16.192.91:3308)/gateway?charset=utf8")
	password = ""
	if err != nil {
		fmt.Println(err)
		return password
	}

	defer db.Close()

	var rows *sql.Rows
	fmt.Println(name)
	rows, err = db.Query("select * from gateway_user where username = ?", name)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(id, "\t", username, "\t", userpassword)
		if !strings.EqualFold(userpassword, "") {
			password = userpassword
		}
	}
	return password
}

