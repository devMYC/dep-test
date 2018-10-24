package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	humanize "github.com/dustin/go-humanize"
)

const Github = "github"
const Typicode = "typicode"

func handleRequest(url string, data []byte, headers *http.Header) func() ([]byte, error) {
	c := make(chan interface{})
	go func() {
		var method string
		if data == nil {
			method = http.MethodGet
		} else {
			method = http.MethodPost
		}
		req, err := http.NewRequest(method, url, bytes.NewReader(data))
		if err != nil {
			c <- err
			return
		}

		for k := range *headers {
			req.Header.Set(k, headers.Get(k))
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c <- err
			return
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c <- err
			return
		}

		fmt.Printf("\u279c  Response size: %s\n", humanize.Bytes(uint64(len(b))))
		c <- b
	}()

	return func() ([]byte, error) {
		v := <-c
		switch v.(type) {
		case error:
			return nil, v.(error)
		}
		return v.([]byte), nil
	}
}

// FetchUserByID fetches user info from typicode.com
func FetchUserByID(from string, idOrLogin string) (User, error) {
	var api string
	var user User

	if len(idOrLogin) == 0 {
		return user, &ArgError{"FetchUserByID", "idOrLogin", "Empty string."}
	}

	switch from {
	case Github:
		token := os.Getenv("GITHUB_AUTH_TOKEN")
		if len(token) == 0 {
			return user, &ArgError{"FetchUserByID", "GITHUB_AUTH_TOKEN", "Could not get auth key from env variable"}
		}
		api = "https://api.github.com/graphql"
		headers := http.Header{}
		headers.Add("Authorization", "bearer "+token)
		headers.Add("Content-Type", "application/json")

		type params struct {
			Login string `json:"login"`
		}
		type payload struct {
			Query  string `json:"query"`
			params `json:"variables"`
		}

		data := payload{
			`query ($login: String!) { user (login: $login) { id, name, location, url } }`,
			params{Login: idOrLogin},
		}

		b, err := json.Marshal(&data)
		if err != nil {
			return user, err
		}

		res, err := handleRequest(api, b, &headers)()
		if err != nil {
			return user, err
		}

		type gqlRes struct {
			Data struct {
				User struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Location string `json:"location"`
					URL      string `json:"url"`
				} `json:"user"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}

		var githubUser gqlRes
		err = json.Unmarshal(res, &githubUser)
		if err != nil {
			return user, err
		}

		// fmt.Printf("Result: %v\n", githubUser)
		if len(githubUser.Errors) > 0 {
			return user, &ArgError{"FetchUserByID", "idOrLogin", githubUser.Errors[0].Message}
		}

		user.ID = githubUser.Data.User.ID
		user.Name = githubUser.Data.User.Name
		user.Location = githubUser.Data.User.Location
		user.WebsiteURL = githubUser.Data.User.URL
		return user, nil

	case Typicode:
		api = "https://jsonplaceholder.typicode.com/users/" + idOrLogin
		res, err := handleRequest(api, nil, &http.Header{})()
		if err != nil {
			return user, err
		}
		var u typicodeUser
		err = json.Unmarshal(res, &u)
		if err != nil {
			return user, err
		}

		user.ID = strconv.Itoa(int(u.ID))
		user.Name = u.Name
		user.Location = u.Address.City + ", " + u.Address.Street
		user.WebsiteURL = u.Website
		return user, nil
	}
	return user, &ArgError{"FetchUserByID", "from", "Unknown query target."}
}

// User structure of typicode /users/:id
type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Location   string `json:"location"`
	WebsiteURL string `json:"website_url"`
}

type typicodeUser struct {
	ID       uint32 `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Address  struct {
		Street  string `json:"street"`
		Suite   string `json:"suite"`
		City    string `json:"city"`
		Zipcode string `json:"zipcode"`
		Geo     struct {
			Lat string `json:"lat"`
			Lng string `json:"lng"`
		} `json:"geo"`
	} `json:"address"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	Company struct {
		Name        string `json:"name"`
		CatchPhrase string `json:"catchPhrase"`
		Bs          string `json:"bs"`
	} `json:"company"`
}
