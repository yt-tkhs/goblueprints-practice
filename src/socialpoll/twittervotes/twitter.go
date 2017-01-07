package main

import (
	"net"
	"time"
	"fmt"
	"io"
	"github.com/garyburd/go-oauth/oauth"
	"sync"
	"net/http"
	"net/url"
	"strconv"
	"log"
	"strings"
	"encoding/json"
	"bufio"
)

var (
	conn net.Conn
	reader io.ReadCloser

	authClient *oauth.Client
	creds *oauth.Credentials

	authSetupOnce sync.Once
	httpClient *http.Client
)

type tweet struct {
	Text string
}

func dial(netw, addr string) (net.Conn, error) {
	fmt.Printf("netw: %s, addr: %s\n", netw, addr)

	if conn != nil {
		conn.Close()
		conn = nil
	}

	netc, err := net.DialTimeout(netw, addr, 5 * time.Second)
	if err != nil {
		return nil, err
	}

	conn = netc
	return netc, nil
}

func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

func setupTwitterAuth() {
	creds = &oauth.Credentials{
		Token:  SP_TWITTER_ACCESSTOKEN,
		Secret: SP_TWITTER_ACCESSSECRET,
	}

	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  SP_TWITTER_KEY,
			Secret: SP_TWITTER_SECRET,
		},
	}
}

func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	authSetupOnce.Do(func() {
		setupTwitterAuth()
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: dial,
			},
		}
	})

	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", req.URL, params))

	fmt.Println("REQUEST:", req.Header, "\n", req.Body)

	return httpClient.Do(req)
}

func readFromTwitter(votes chan<- string) {
	options, err := loadOptions()
	if err != nil {
		log.Println("Failed to load options:", err)
		return
	}
	fmt.Println(options)

	u, err := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
	if err != nil {
		log.Println("Failed to parse URL:", err)
		return
	}

	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	if err != nil {
		log.Println("Failed to create search request:", err)
		return
	}

	resp, err := makeRequest(req, query)
	if err != nil {
		log.Println("Failed to request of search:", err)
		return
	}

	s := bufio.NewScanner(resp.Body)
	s.Scan()
	log.Println(s.Text())
	log.Println("StatusCode=", resp.StatusCode)

	reader = resp.Body
	decoder := json.NewDecoder(reader)

	var jsonStr []byte
	reader.Read(jsonStr)
	fmt.Println("decoder: ", string(jsonStr))

	for {
		var tw tweet

		if err := decoder.Decode(&tw); err != nil {
			fmt.Println("Failed to decode:", err)
			break
		}

		fmt.Println(tw)

		for _, option := range options {
			if strings.Contains(strings.ToLower(tw.Text), strings.ToLower(option)) {
				log.Println("投票: " + option)
				votes <- option
			}
		}
	}
}

func startTwitterStream(stopchan <-chan struct{}, votes chan<- string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)

	go func() {
		defer func() {
			stoppedchan <- struct {}{}
		}()

		for {
			select {
			case <-stopchan:
				log.Println("Finishing request to Twitter...")
				return
			default:
				log.Println("dialing to Twitter...")
				readFromTwitter(votes)
				log.Println("(WAITING)")
				time.Sleep(10 * time.Second)
			}
		}
	}()

	return stoppedchan
}