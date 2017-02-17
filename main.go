package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var allowedFails = 2

func main() {
	for _, envVar := range []string{"PASSWORD", "USERNAME", "HOST"} {
		if os.Getenv(envVar) == "" {
			log.Panicf("Enviornment variable %s was not set", envVar)
		}
	}

	log.Println("Starting successful!")

	for {
		fails := 0
		for fails < allowedFails {
			if testInternet() {
				fails = 0
			} else {
				fails++
				log.Printf("Internet Check Failed %d times\n", fails)
			}
			time.Sleep(1 * time.Minute)
		}
		log.Println("Rebooting Router")
		err := rebootRouter()
		if err != nil {
			log.Printf("Failed to reboot router %s\n", err)
		} else {
			log.Println("Router Rebooted")
		}
		time.Sleep(5 * time.Minute)
	}

}

func testInternet() bool {
	_, err := http.Get("https://google.com/")
	return err == nil
}

func rebootRouter() error {
	cookieJar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Jar: cookieJar, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/adv_index.asp", os.Getenv("HOST")), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(os.Getenv("USERNAME"), os.Getenv("PASSWORD"))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to login. %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("./test", data, 0644)
	if err != nil {
		return err
	}

	re, err := regexp.Compile(`<SCRIPT language="JavaScript" type="text/javascript" > var webSecurityKey ="(.+)"</SCRIPT>`)
	if err != nil {
		return err
	}

	res := re.FindStringSubmatch(string(data))

	log.Println("Logged in for reboot.")

	form := url.Values{}
	form.Add("reset_delay", "1")
	form.Add("RebootEvent", "reboot")
	form.Add("webSecurityKey", res[1])
	form.Add("next_page", "/pls_wait_reboot.asp")
	form.Add("SessionID", "")

	req, err = http.NewRequest("POST", fmt.Sprintf("%s/goform/EventForm", os.Getenv("HOST")), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.SetBasicAuth(os.Getenv("USERNAME"), os.Getenv("PASSWORD"))
	resp, err = client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to reboot. %d", resp.StatusCode)
	}

	return nil
}
