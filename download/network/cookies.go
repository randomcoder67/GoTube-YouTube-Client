package network

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// This file contains the retreiving and parsing of the Firefox youtube.com cookies, as well as retrieving the SAPISID to use for API post requests

const BASE_FIREFOX_PATH = "/.mozilla/firefox/"
const TEMP_COOKIES_LOCATION = "/.cache/gotube/firefoxCookies.sqlite"

func getCookiesFile() string {
	homeDir, _ := os.UserHomeDir()
	files, err := os.ReadDir(homeDir + BASE_FIREFOX_PATH)
	if err != nil {
		panic(err)
	}
	for _, x := range files {
		if x.IsDir() && strings.Contains(x.Name(), "default-release") {
			cpCmd := exec.Command("cp", homeDir + BASE_FIREFOX_PATH + x.Name() + "/cookies.sqlite", homeDir + TEMP_COOKIES_LOCATION)
			err = cpCmd.Run()
			if err != nil {
				panic(err)
			}
			return homeDir + TEMP_COOKIES_LOCATION
		}
	}
	return ""
}

func GetCookies() *cookiejar.Jar {
	db, err := sql.Open("sqlite3", getCookiesFile())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	row, err := db.Query("SELECT name, value, host, path, expiry, isHttpOnly FROM moz_cookies where host= \".youtube.com\"")
	//row, err := db.Query("SELECT name, value, host_key, path, expires_utc, is_httponly FROM Cookies where host_key= \".youtube.com\"")
	if err != nil {
		panic(err)
	}
	defer row.Close()

	var cookiesList []*http.Cookie

	for row.Next() {
		var name, value, domain, path, httpOnly, expiresAt string
		err := row.Scan(&name, &value, &domain, &path, &expiresAt, &httpOnly)

		var httpOnlyBool = false
		if httpOnly == "TRUE" {
			httpOnlyBool = true
		}

		expiresAt64, err := strconv.ParseInt(expiresAt, 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(expiresAt64, 0)

		if name == "__Secure-1PAPISID" || name == "__Secure-1PSID" || name == "__Secure-1PSIDCC" || name == "__Secure-1PSIDTS" || name == "__Secure-3PAPISID" || name == "__Secure-3PSID" || name == "__Secure-3PSIDCC" || name == "__Secure-3PSIDTS" || name == "APISID" || name == "CONSENT" || name == "CONSISTENCY" || name == "HSID" || name == "LOGIN_INFO" || name == "PREF" || name == "SAPISID" || name == "SID" || name == "SIDCC" || name == "SOCS" || name == "SSID" || name == "VISITOR_INFO1_LIVE" || name == "VISITOR_PRIVACY_METADATA" || name == "YSC" {
			cookie := &http.Cookie{
				Name:     name,
				Value:    value,
				Path:     path,
				Domain:   domain,
				Expires:  tm,
				HttpOnly: httpOnlyBool,
			}
			cookiesList = append(cookiesList, cookie)
			err = cookie.Valid()
			if err != nil {
				panic(err)
			}
		}
	}
	//os.Exit(0)

	jar, err := cookiejar.New(nil)
	u, _ := url.Parse("https://www.youtube.com/")
	jar.SetCookies(u, cookiesList)
	row.Close()
	db.Close()
	return jar
}

// Extract the SAPIDID from YouTube cookies
func getSapis(jar *cookiejar.Jar) string {
	u, _ := url.Parse("https://www.youtube.com")

	for _, cookie := range jar.Cookies(u) {
		if cookie.Name == "SAPISID" {
			return cookie.Value
		}
	}
	return ""
}

// Debug function
func Print(str string) {
	cmd := exec.Command("notify-send", str)
	cmd.Run()
}
