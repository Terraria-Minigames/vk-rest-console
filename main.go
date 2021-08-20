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
			return
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

	fmt.Println("Listening on port " + strconv.Itoa(config.Port))
	http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
}

func HandleNewMessage(json *sjson.Json) {
	text, _ := json.Get("object").Get("message").Get("text").String()
	fromId, _ := json.Get("object").Get("message").Get("from_id").Int()

	// If isn't a valid command
	if !strings.HasPrefix(text, "/") {
		return
	}

	resttoken, ok := config.VKUserTokens[fromId]

	if !ok {
		fmt.Println("id" + strconv.Itoa(fromId) + " tried to execute " + text)
		return
	}

	qs := &url.Values{}
	qs.Add("token", resttoken)
	qs.Add("cmd", text)

	resp, reqerr := requests.Get(config.RestUrl+"/v3/server/rawcmd", qs, nil)
	if reqerr != nil {
		SendVKMessage("Server request failed.", fromId) // Sending back err.Error() expsoses token
		return
	}

	json, jsonerr := sjson.NewJson(resp.Raw().Bytes())
	if jsonerr != nil {
		SendVKMessage("Failed parsing server response.\n"+jsonerr.Error(), fromId)
		return
	}

	if response, err := json.Get("response").StringArray(); err == nil {
		if len(response) > 0 {
			result := strings.Join(response, "\n")
			SendVKMessage(result, fromId)
		} else {
			SendVKMessage("Command didn't return output.", fromId)
		}

		fmt.Println("id" + strconv.Itoa(fromId) + " executed " + text)
	}

}

func SendVKMessage(text string, userId int) {
	qs := &url.Values{}
	qs.Add("message", text)
	qs.Add("access_token", config.VKToken)
	qs.Add("keyboard", config.VKKeyboard)
	qs.Add("user_id", strconv.Itoa(userId))
	qs.Add("random_id", strconv.Itoa(int(time.Now().UnixNano())))
	qs.Add("v", "5.131")
	resp, err := requests.Get("https://api.vk.com/method/messages.send", qs, nil)
	if err != nil {
		fmt.Println("Error occured when sending VK message to id " + strconv.Itoa(userId) + ": " + text)
		fmt.Println(err.Error())
		fmt.Println(resp.Text())
	}
}
