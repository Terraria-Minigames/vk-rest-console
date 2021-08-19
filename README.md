## VK Rest Console
Bring tshock console to your server staff without giving them access to host machine.

## Setup guide.
Run the binary, close it. It will generate `config.json`, which you will use to set up your everything.

### 0. Download
Go to [releases tab](https://github.com/btvoidx/vk-rest-console/releases), download binary for your os (windows or linux, because who the hell hosts professional tshock servers from mac), place it somewhere on the same machine as your tshock server.

### 1. Obtain VK Group token
Go to your VK group, click *Manage*, head to *Settings* > *API usage*.

In **Access tokens** tab press *Create token* button, check the checkmark near *Allow access to community messages* line, press *Create*. Copy created token and place it into `config.json` as a value for `"VKToken"`.

### 2. Set up callback api
Go to your VK group, click *Manage*, head to *Settings* > *API usage*.
In **Callback API** tab press *Add server* button.
- Set **API version** to **5.131**.
- Set **URL** to url to your server.
- Set **Secret key** to any hard-to-guess string.

Place **Secret key** into `config.json` as a value for `"VKSecret"`. Do the same to **String to be returned** value as a value for `"VKConfirmationToken"`.

### 3. Set up TShock config
Set `"TShockConfig"` in `config.json` (*not* tshock one yet, it may get confusing) to **absoulte path** to your *tshock* `config.json`. Examples below.

In your *tshock* `config.json` modify `"ApplicationRestTokens"` accordingly to [tshock official guide](https://tshock.readme.io/reference/rest-api-endpoints#setting-it-all-up).

After you've done that, add `"VKId": {VK_PROFILE_ID}` to each token you want to be executable from VK.

### 4. Use
If everything is done correctly, anyone whos ids you assigned to tshock application rest tokens is now able to simply DM commands to you group, and they'll get executed. Try it: `/help`.

Be aware, that commands which do not work from normal TShock console will not work here either.

### Examples
tshock's `config.json`
```json
...
"ApplicationRestTokens": {
  "my_very_own_and_very_secret_rest_token": {
    "Username": "btvoidx",
    "UserGroupName": "superadmin",
    "VKId": 187569882
  },
  "not_mine_and_not_so_secret_rest_token": {
    "Username": "Nikita Matrosoff",
    "UserGroupName": "supersuperadmin",
    "VKId": 206352149
  }
}
```

our `config.json`
```
{
	"Port": 80,
	"RestUrl": "http://localhost:7878",
	"TShockConfig": "C:/Users/Admin/server/tshock/config.json",
	"VKConfirmationToken": "ba7bf260",
	"VKSecret": "3vUA8FBVJLpjjQd37ZxqZVFLRi93rRuC",
	"VKKeyboard": "",
	"VKToken": "a9eabd3ef76be720606e010107a339e203fe2cid81b67715ebaf823e8e52380f634516850cf0ab8344bb1"
}
```

## Configuration
Restart binary after editing config for changes to apply.

Possible values:
- `"Port": 80` - int - Changes port on which http server is running. Useful when running with proxy, like Nginx.
- `"RestUrl": "http://localhost:7878"` - string - Base URI for your TShock REST api.
- `"TShockConfig": ""` - string - Absoulte path to your tshock `config.json`.
- `"VKConfirmationToken": ""` - string - "String to be returned" value from VK callback api server setup screen.
- `"VKSecret": ""` - string - "Secret key" which is sent with every request by VK.
- `"VKKeyboard": ""` - string - [Keyboard object](https://vk.com/dev/bots_docs_3) to be sent with all messages.
- `"VKToken": ""` - string - Group Access Token with **community messages** access

## Contribute
Open a new issue or DM me wherever you want to suggest changes or report issues.
