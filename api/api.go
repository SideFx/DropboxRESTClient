// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// REST API
// ---------------------------------------------------------------------------------------------------------------------

package api

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"
)

// AuthorizeApp -calling system browser and load the Dropbox app authorization page
func AuthorizeApp(auth AppAuthType) error {
	var err error = nil
	targetUrl := dropboxAuthURI + "?" + paraClientId + auth.AppKey + "&" + paraTokenAccessType + valTokenAccessType +
		"&" + paraResponseType + valResponseType
	err = openURL(targetUrl)
	return err
}

// RequestRefreshToken -fetch refresh token after app has been authorized
func RequestRefreshToken(auth AppAuthType, code string) (string, error) {
	var r *RefreshTokenType
	var err error
	// create base64 encoded auth. key (app key + app secret, separated by ":")
	authString := base64.StdEncoding.EncodeToString([]byte(auth.AppKey + ":" + auth.AppSecret))
	var para = RESTParaType{
		ParaURL:    dropboxAPIURI + endpointAuthToken,
		ParaMethod: http.MethodPost,
		ParaHeader: []KeyValueType{
			{paraContentType, valContentTypeURLForm},
			{paraAuthorization, valAuthBasic + authString},
		},
		ParaForm: url.Values{
			paraCode:      {code},
			paraGrantType: {valAuthorizationCode},
		},
		ParaBody: "",
	}
	r, err = restCall[*RefreshTokenType](para)
	if err != nil {
		return "", err
	}
	return r.RefreshToken, nil
}

// GetCurrentUser -get Dropbox user id, needed for user authorization (making api calls)
func GetCurrentUser() (*UserInfoType, error) {
	var err error
	var r UserInfoType
	err = requestAccessToken()
	if err != nil {
		return nil, err
	}
	var para = RESTParaType{
		ParaURL:    dropboxAPIURI + endpointGetCurrentUser,
		ParaMethod: http.MethodPost,
		ParaHeader: []KeyValueType{
			{paraAuthorization, valAuthBearer + accessToken.token},
		},
		ParaForm: url.Values{},
		ParaBody: "",
	}
	r, err = restCall[UserInfoType](para)
	if err != nil {
		return nil, err
	}
	currentUserId = r.AccountId
	return &r, nil
}

// CurrentUserGetPicture -fetch user account picture
func CurrentUserGetPicture(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// ListFolders -list folders && list folders continue
func ListFolders(path string, recursive bool, limit uint32) ([]*FileItemType, error) {
	var err error
	var hasmore = false
	var cursor string
	var entries []*FileItemType
	err = requestAccessToken()
	if err != nil {
		return nil, err
	}
	var r, c ItemInfoType
	var dbxpara = ListFoldersParaType{
		false,
		false,
		true,
		true,
		path,
		recursive,
		limit,
	}
	jdbxpara, err := anyToJson[ListFoldersParaType](dbxpara)
	if err != nil {
		return nil, err
	}
	var paraStart = RESTParaType{
		ParaURL:    dropboxAPIURI + endpointListFolder,
		ParaMethod: http.MethodPost,
		ParaHeader: []KeyValueType{
			{paraAuthorization, valAuthBearer + accessToken.token},
			{paraContentType, valContentTypeJson},
		},
		ParaForm: url.Values{},
		ParaBody: jdbxpara,
	}
	r, err = restCall[ItemInfoType](paraStart)
	if err != nil {
		return nil, err
	}
	for _, e := range r.Entries {
		entries = append(entries, &e)
	}
	hasmore = r.HasMore
	cursor = r.Cursor
	for hasmore {
		err = requestAccessToken()
		if err != nil {
			return nil, err
		}
		var dbxcont = ListContinueType{
			cursor,
		}
		jdbxcont, err := anyToJson[ListContinueType](dbxcont)
		if err != nil {
			return nil, err
		}
		var paraCont = RESTParaType{
			ParaURL:    dropboxAPIURI + endpointListFolderContinue,
			ParaMethod: http.MethodPost,
			ParaHeader: []KeyValueType{
				{paraAuthorization, valAuthBearer + accessToken.token},
				{paraContentType, valContentTypeJson},
			},
			ParaForm: url.Values{},
			ParaBody: jdbxcont,
		}
		c, err = restCall[ItemInfoType](paraCont)
		if err != nil {
			return nil, err
		}
		for _, e := range c.Entries {
			entries = append(entries, &e)
		}
		hasmore = c.HasMore
		cursor = c.Cursor
	}
	return entries, nil
}

func MoveFiles(from, to string) (*FileItemMetadataType, error) {
	var metadata *FileItemMetadataType
	var err error
	var dbxpara = FilesMoveParaType{
		false,
		true,
		from,
		to,
	}
	jdbxpara, err := anyToJson[FilesMoveParaType](dbxpara)
	if err != nil {
		return nil, err
	}
	var para = RESTParaType{
		ParaURL:    dropboxAPIURI + endPointFilesMove,
		ParaMethod: http.MethodPost,
		ParaHeader: []KeyValueType{
			{paraAuthorization, valAuthBearer + accessToken.token},
			{paraContentType, valContentTypeJson},
		},
		ParaForm: url.Values{},
		ParaBody: jdbxpara,
	}
	metadata, err = restCall[*FileItemMetadataType](para)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// requestAccessToken -checks if the current access token has expired and fetches a new one, if needed,
// should be called before making any other dropbox api call
func requestAccessToken() error {
	var r RefreshTokenType
	var err error
	if accessToken.token == "" || (accessToken.fetchedAt+accessToken.expiresIn+threshold) < time.Now().Unix() {
		// create base64 encoded auth. key (app key + app secret, separated by ":")
		authString := base64.StdEncoding.EncodeToString([]byte(authkey.AppKey + ":" + authkey.AppSecret))
		var para = RESTParaType{
			ParaURL:    dropboxAPIURI + endpointAuthToken,
			ParaMethod: http.MethodPost,
			ParaHeader: []KeyValueType{
				{paraContentType, valContentTypeURLForm},
				{paraAuthorization, valAuthBasic + authString},
			},
			ParaForm: url.Values{
				paraGrantType:    {valRefreshToken},
				paraRefreshToken: {refreshToken},
			},
			ParaBody: "",
		}
		r, err = restCall[RefreshTokenType](para)
		if err != nil {
			return err
		}
		accessToken.token = r.AccessToken
		accessToken.expiresIn = r.ExpiresIn
		accessToken.fetchedAt = time.Now().Unix()
	}
	return nil
}

// https://gist.github.com/sevkin/9798d67b2cb9d07cb05f89f14ba682f8
func openURL(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}
	if len(args) > 1 {
		args = append(args[:1], append([]string{""}, args[1:]...)...)
	}
	return exec.Command(cmd, args...).Start()
}
