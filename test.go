package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"

	_ "github.com/go-sql-driver/mysql"
)

func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))

}

func CreateCaptcha() string {
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}

func main() {
	v := url.Values{}
	_now := strconv.FormatInt(time.Now().Unix(), 10)
	//fmt.Printf(_now)
	_account := "C39781260"                         //查看用户名 登录用户中心->验证码通知短信>产品总览->API接口信息->APIid
	_password := "acf3be18589488307fc6995bc2b28bf0" //查看密码 登录用户中心->验证码通知短信>产品总览->API接口信息->APIKEY
	_mobile := "13087213080"
	code := CreateCaptcha()
	_content := fmt.Sprintf("您的验证码是：%s。请不要把验证码泄露给其他人。", code)
	v.Set("account", _account)
	v.Set("password", GetMd5String(_account+_password+_mobile+_content+_now))
	v.Set("mobile", _mobile)
	v.Set("content", _content)
	v.Set("time", _now)
	body := strings.NewReader(v.Encode()) //把form数据编下码
	client := &http.Client{}

	req, _ := http.NewRequest("POST", "https://106.ihuyi.com/webservice/sms.php?method=Submit&format=json", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	fmt.Printf("%+v\n", req)    //看下发送的结构
	resp, err := client.Do(req) //发送
	if err != nil {
		fmt.Printf("send fail")
		return
	}
	defer resp.Body.Close() //一定要关闭resp.Body

	db, err := sql.Open("mysql", "root:238238238@tcp(127.0.0.1:3306)/mysql?charset=utf8")
	if err != nil {
		fmt.Printf("conn db err")
		return
	}
	defer db.Close()
	//哈哈huohuo

	stmt, err := db.Prepare("select f_id,f_name from md_test.t_test where 1=1")
	if err != nil {
		fmt.Printf("Prepare table err")
		return
	}

	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		fmt.Printf("Query db err")
		return
	}

	defer rows.Close()

	for rows.Next() {
		var f_id string
		var f_name string
		err := rows.Scan(&f_id, &f_name)
		if err != nil {
			fmt.Printf("parse data err")
			return
		}
		fmt.Println(f_id, f_name)
	}

	c, err1 := redis.Dial("tcp", "localhost:6379") // 指定端口，连接方式
	if err1 != nil {
		fmt.Println("conn redis failed,", err1)

		return
	}
	defer func() {
		err := c.Close()
		if err != nil {
			return
		}
	}()
	fmt.Println("redis connect success")
	_, err = c.Do("Set", "13087213080", code)
	if err != nil {
		fmt.Println(err)
		return
	}

	r, err := redis.Int(c.Do("Get", "13087213080"))
	if err != nil {
		fmt.Println("get code failed,", err)
		return
	}
	fmt.Println(r)

	_, err = c.Do("expire", "13087213080", 60)
	if err != nil {
		fmt.Println(err)
		return
	}
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data), err)

}
