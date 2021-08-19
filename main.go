package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	sjson "github.com/bitly/go-simplejson"
	requests "github.com/hiroakis/go-requests"
)

var config = LoadConfig()

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Deny all requests coming to an invalid path
		if r.URL.Path != "/" {
			return
		}

		json, err := sjson.NewFromReader(r.Body)
		if err != nil {
			fmt.Fprintf(w, "500")
		}

		secret, _ := json.Get("secret").String()
		eventtype, _ := json.Get("type").String()

		if secret != config.VKSecret {
			fmt.Fprint(w, "Nope.")
			return
		}

		if eventtype == "confirmation" {
			fmt.Fprint(w, config.VKConfirmationToken)
			return
		}

		if eventtype == "message_new" {
			go HandleNewMessage(json)
		}

		fmt.Fprint(w, "ok")
	})

	fmt.Println("Listening on port 80")
	http.ListenAndServe(":80", nil)
}

func HandleNewMessage(json *sjson.Json) {
	text, _ := json.Get("object").Get("message").Get("text").String()
	from_id, _ := json.Get("object").Get("message").Get("from_id").Int()

	// If isn't a valid command
	if !strings.HasPrefix(text, "/") {
		return
	}

	if resttoken, ok := config.VKUserTokens[from_id]; ok {
		qs := &url.Values{}
		qs.Add("token", resttoken)
		qs.Add("cmd", text)

		resp, err := requests.Get(config.RestUrl+"/v3/server/rawcmd", qs, nil)
		if err != nil {
			SendVKMessage("[Server request failed]\n"+err.Error(), from_id)
			return
		}

		json, err := sjson.NewJson(resp.Raw().Bytes())
		if err != nil {
			SendVKMessage("[Server didn't return anything]\n"+err.Error(), from_id)
			return
		}

		if response, err := json.Get("response").StringArray(); err == nil {
			if len(response) > 0 {
				result := strings.Join(response, "\n")
				SendVKMessage(result, from_id)
			} else {
				SendVKMessage("[No responce]", from_id)
			}
			fmt.Println("id" + strconv.Itoa(from_id) + " executed " + text)
		}
	}
}

func SendVKMessage(text string, user_id int) {
	qs := &url.Values{}
	qs.Add("message", text)
	qs.Add("access_token", config.VKToken)
	qs.Add("keyboard", config.VKKeyboard)
	qs.Add("user_id", strconv.Itoa(user_id))
	qs.Add("random_id", strconv.Itoa(int(time.Now().UnixNano())))
	qs.Add("v", "5.131")
	requests.Get("https://api.vk.com/method/messages.send", qs, nil)
}
