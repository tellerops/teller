package lastpass

import (
	"golang.org/x/crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
)

type blob struct {
	bytes             []byte
	keyIterationCount int
}

type session struct {
	id                string
	keyIterationCount int
	cookieJar         http.CookieJar
}

func login(username, password string) (*session, error) {
	iterationCount, err := requestIterationCount(username)
	if err != nil {
		return nil, err
	}
	return make_session(username, password, iterationCount)
}

func make_session(username, password string, iterationCount int) (*session, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar: cookieJar,
	}
	res, err := client.PostForm(
		"https://lastpass.com/login.php",
		url.Values{
			"method":     []string{"mobile"},
			"web":        []string{"1"},
			"xml":        []string{"1"},
			"username":   []string{username},
			"hash":       []string{string(makeHash(username, password, iterationCount))},
			"iterations": []string{fmt.Sprint(iterationCount)},
		})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response struct {
		SessionId string `xml:"sessionid,attr"`
	}
	err = xml.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return &session{response.SessionId, iterationCount, cookieJar}, nil
}

func fetch(s *session) (*blob, error) {
	u, err := url.Parse("https://lastpass.com/getaccts.php")
	if err != nil {
		return nil, err
	}
	u.RawQuery = (&url.Values{
		"mobile":    []string{"1"},
		"b64":       []string{"1"},
		"hash":      []string{"0.0"},
		"PHPSESSID": []string{s.id},
	}).Encode()
	client := &http.Client{
		Jar: s.cookieJar,
	}
	res, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}
	b, err = base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return nil, err
	}
	return &blob{b, s.keyIterationCount}, nil
}

func requestIterationCount(username string) (int, error) {
	res, err := http.DefaultClient.PostForm(
		"https://lastpass.com/iterations.php",
		url.Values{
			"email": []string{username},
		})
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func makeKey(username, password string, iterationCount int) []byte {
	if iterationCount == 1 {
		b := sha256.Sum256([]byte(username + password))
		return b[:]
	}
	return pbkdf2.Key([]byte(password), []byte(username), iterationCount, 32, sha256.New)
}

func makeHash(username, password string, iterationCount int) []byte {
	key := makeKey(username, password, iterationCount)
	if iterationCount == 1 {
		b := sha256.Sum256([]byte(string(encodeHex(key)) + password))
		return encodeHex(b[:])
	}
	return encodeHex(pbkdf2.Key([]byte(key), []byte(password), 1, 32, sha256.New))
}
