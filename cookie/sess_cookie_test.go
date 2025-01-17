// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cookie

import (
	"crypto/aes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/session"
)

func TestCookie(t *testing.T) {
	config := `{"cookieName":"gosessionid","enableSetCookie":false,"gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"beegocookiehashkey\"}"}`
	conf := new(session.ManagerConfig)
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		t.Fatal("json decode error", err)
	}
	globalSessions, err := session.NewManager("cookie", conf)
	if err != nil {
		t.Fatal("init cookie session err", err)
	}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		t.Fatal("set error,", err)
	}
	err = sess.Set(nil, "username", "astaxie")
	if err != nil {
		t.Fatal("set error,", err)
	}
	if username := sess.Get(nil, "username"); username != "astaxie" {
		t.Fatal("get username error")
	}
	sess.SessionRelease(nil, w)
	if cookiestr := w.Header().Get("Set-Cookie"); cookiestr == "" {
		t.Fatal("setcookie error")
	} else {
		parts := strings.Split(strings.TrimSpace(cookiestr), ";")
		for k, v := range parts {
			nameval := strings.Split(v, "=")
			if k == 0 && nameval[0] != "gosessionid" {
				t.Fatal("error")
			}
		}
	}
}

func TestDestroySessionCookie(t *testing.T) {
	config := `{"cookieName":"gosessionid","enableSetCookie":true,"gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"beegocookiehashkey\"}"}`
	conf := new(session.ManagerConfig)
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		t.Fatal("json decode error", err)
	}
	globalSessions, err := session.NewManager("cookie", conf)
	if err != nil {
		t.Fatal("init cookie session err", err)
	}

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	session, err := globalSessions.SessionStart(w, r)
	if err != nil {
		t.Fatal("session start err,", err)
	}

	// request again ,will get same sesssion id .
	r1, _ := http.NewRequest("GET", "/", nil)
	r1.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
	w = httptest.NewRecorder()
	newSession, err := globalSessions.SessionStart(w, r1)
	if err != nil {
		t.Fatal("session start err,", err)
	}
	if newSession.SessionID(nil) != session.SessionID(nil) {
		t.Fatal("get cookie session id is not the same again.")
	}

	// After destroy session , will get a new session id .
	globalSessions.SessionDestroy(w, r1)
	r2, _ := http.NewRequest("GET", "/", nil)
	r2.Header.Set("Cookie", w.Header().Get("Set-Cookie"))

	w = httptest.NewRecorder()
	newSession, err = globalSessions.SessionStart(w, r2)
	if err != nil {
		t.Fatal("session start error")
	}
	if newSession.SessionID(nil) == session.SessionID(nil) {
		t.Fatal("after destroy session and reqeust again ,get cookie session id is same.")
	}
}

func TestGenerate(t *testing.T) {
	str := generateRandomKey(20)
	if len(str) != 20 {
		t.Fatal("generate length is not equal to 20")
	}
}

func TestCookieEncodeDecode(t *testing.T) {
	hashKey := "testhashKey"
	blockkey := generateRandomKey(16)
	block, err := aes.NewCipher(blockkey)
	if err != nil {
		t.Fatal("NewCipher:", err)
	}
	securityName := string(generateRandomKey(20))
	val := make(map[interface{}]interface{})
	val["name"] = "astaxie"
	val["gender"] = "male"
	str, err := encodeCookie(block, hashKey, securityName, val)
	if err != nil {
		t.Fatal("encodeCookie:", err)
	}
	dst, err := decodeCookie(block, hashKey, securityName, str, 3600)
	if err != nil {
		t.Fatal("decodeCookie", err)
	}
	if dst["name"] != "astaxie" {
		t.Fatal("dst get map error")
	}
	if dst["gender"] != "male" {
		t.Fatal("dst get map error")
	}
}

func TestParseConfig(t *testing.T) {
	s := `{"cookieName":"gosessionid","gclifetime":3600}`
	cf := new(session.ManagerConfig)
	cf.EnableSetCookie = true
	err := json.Unmarshal([]byte(s), cf)
	if err != nil {
		t.Fatal("parse json error,", err)
	}
	if cf.CookieName != "gosessionid" {
		t.Fatal("parseconfig get cookiename error")
	}
	if cf.Gclifetime != 3600 {
		t.Fatal("parseconfig get gclifetime error")
	}

	cc := `{"cookieName":"gosessionid","enableSetCookie":false,"gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"beegocookiehashkey\"}"}`
	cf2 := new(session.ManagerConfig)
	cf2.EnableSetCookie = true
	err = json.Unmarshal([]byte(cc), cf2)
	if err != nil {
		t.Fatal("parse json error,", err)
	}
	if cf2.CookieName != "gosessionid" {
		t.Fatal("parseconfig get cookiename error")
	}
	if cf2.Gclifetime != 3600 {
		t.Fatal("parseconfig get gclifetime error")
	}
	if cf2.EnableSetCookie {
		t.Fatal("parseconfig get enableSetCookie error")
	}
	cconfig := new(cookieConfig)
	err = json.Unmarshal([]byte(cf2.ProviderConfig), cconfig)
	if err != nil {
		t.Fatal("parse ProviderConfig err,", err)
	}
	if cconfig.CookieName != "gosessionid" {
		t.Fatal("ProviderConfig get cookieName error")
	}
	if cconfig.SecurityKey != "beegocookiehashkey" {
		t.Fatal("ProviderConfig get securityKey error")
	}
}
