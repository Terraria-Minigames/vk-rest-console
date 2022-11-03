package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/broothie/qst"
	"github.com/spf13/pflag"
)

const ApiVersion = "5.131"

var version = "" // Set during build via -ldflags

func main() {
	configPath := pflag.StringP("config-path", "c", "./config.yaml", "path to config.yaml, file name can not be omitted")
	shouldPrintVersion := pflag.BoolP("version", "v", false, "prints version")
	pflag.Parse()

	if *shouldPrintVersion {
		fmt.Printf("%s\n", version)
		return
	}

	config, err := LoadConfig(*configPath, true)
	if err != nil {
		fmt.Printf("Failed to load config from (%q): %v\n", *configPath, err)
		os.Exit(1)
	}

	if err := ValidateConfigValues(config); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	tshockConfig, err := LoadTShockConfig(config.TShockConfigPath)
	if err != nil {
		fmt.Printf("Could not load read TShock config from (%q): %v\n", config.TShockConfigPath, err)
		os.Exit(1)
	}

	if tshockConfig.RestApiEnabled == false {
		fmt.Printf("Rest API is not enabled in TShock config (%q)\n", config.TShockConfigPath)
		os.Exit(1)
	}

	if config.CommandPrefix == "" {
		config.CommandPrefix = tshockConfig.CommandSpecifier
	}
	if config.RestAddr == "" {
		config.RestAddr = fmt.Sprintf("http://127.0.0.1:%d", tshockConfig.RestApiPort)
	}

	http.HandleFunc("/", Handler(config, tshockConfig))

	fmt.Printf("Starting callback server on port %d\n", config.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil); err != nil {
		fmt.Printf("Failed to start callback server on port %d: %v\n", config.Port, err)
		os.Exit(1)
	}
}

func Handler(c Config, tc TShockConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Deny all requests coming to an invalid path
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		req := struct {
			EventType string `json:"type"`
			Version   string `json:"v"`
			Secret    string `json:"secret"`
			Object    struct {
				Message struct {
					FromID    int    `json:"from_id"`
					PeerID    int    `json:"peer_id"`
					MessageID int    `json:"id"`
					Text      string `json:"text"`
				} `json:"message"`
			} `json:"object"`
		}{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Version != ApiVersion {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "unsupported callback api version; set it to %q", ApiVersion)
			return
		}

		if req.Secret != c.VK.Secret {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "incorrect secret")
			return
		}

		if req.EventType == "confirmation" {
			fmt.Fprintf(w, c.VK.ConfirmationToken)
			return
		}

		fmt.Fprintf(w, "ok")

		if req.EventType != "message_new" {
			return
		}

		if !strings.HasPrefix(req.Object.Message.Text, c.CommandPrefix) {
			return
		}

		for token, d := range tc.ApplicationRestTokens {
			if d.VKId != req.Object.Message.FromID {
				continue
			}

			output, err := ExecRESTCommand(c.RestAddr, token, req.Object.Message.Text)
			if err != nil {
				fmt.Printf("Failed to execute %q: %v\n", req.Object.Message.Text, err)
				if err := SendVKMessage(
					c.VK.Token,
					req.Object.Message.PeerID,
					req.Object.Message.MessageID,
					c.Messages.RestRequestFailed,
					c.VK.Keyboard,
				); err != nil {
					fmt.Printf("Failed to send VK reply: %v", err)
				}

				return
			}

			if output == "" {
				output = c.Messages.NoCommandOutput
			}

			if c.RemoveChatTags {
				output = regexp.MustCompile(`(?:\[c\/.+?:(.+?)\])|(?:\[i:.*?\])`).ReplaceAllString(output, "$1")
			}

			fmt.Printf("%s (%s) executed %s\n", d.Username, d.UserGroupName, req.Object.Message.Text)

			if err := SendVKMessage(
				c.VK.Token,
				req.Object.Message.PeerID,
				req.Object.Message.MessageID,
				output,
				c.VK.Keyboard,
			); err != nil {
				fmt.Printf("Failed to send VK reply: %v", err)
			}

			return
		}
	}
}

func SendVKMessage(token string, peerId, replyTo int, message string, keyboard any) error {
	kb := ""
	if keyboard != nil {
		keyboardBytes, err := json.Marshal(keyboard)
		if err != nil {
			return err
		}
		kb = string(keyboardBytes)
	}

	r, err := qst.Get("https://api.vk.com/method/messages.send",
		qst.QueryValue("access_token", token),
		qst.QueryValue("message", message),
		qst.QueryValue("peer_id", strconv.Itoa(peerId)),
		qst.QueryValue("reply_to", strconv.Itoa(replyTo)),
		qst.QueryValue("keyboard", kb),
		qst.QueryValue("random_id", strconv.FormatInt(time.Now().UnixNano(), 10)),
		qst.QueryValue("v", ApiVersion),
	)
	if err != nil {
		return err
	}

	resp := struct {
		Error map[any]any `json:"error"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return err
	}

	if resp.Error != nil {
		b, _ := json.Marshal(resp.Error)
		return errors.New(string(b))
	}

	return nil
}

func ExecRESTCommand(restAddr, token, command string) (string, error) {
	data := struct {
		OutputLines []string `json:"response"`
	}{}

	resp, err := qst.Get(strings.TrimSuffix(restAddr, "/")+"/v3/server/rawcmd",
		qst.QueryValue("token", token),
		qst.QueryValue("cmd", command),
	)
	if err != nil {
		return "", err
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return strings.Join(data.OutputLines, "\n"), nil
}