// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// REST API
// ---------------------------------------------------------------------------------------------------------------------

package api

import (
	"Dropbox_REST_Client/assets"
	"encoding/base64"
	"errors"
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
			{paraContentType, string(valContentTypeURLForm)},
			{paraAuthorization, string(valAuthBasic) + authString},
		},
		ParaForm: url.Values{
			paraCode:      {code},
			paraGrantType: {valAuthorizationCode},
		},
		ParaBody: nil,
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
			{paraAuthorization, string(valAuthBearer) + accessToken.token},
		},
		ParaForm: url.Values{},
		ParaBody: nil,
	}
	r, err = restCall[UserInfoType](para)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CurrentUserGetPicture -fetch user account picture
func CurrentUserGetPicture(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
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
			{paraAuthorization, string(valAuthBearer) + accessToken.token},
			{paraContentType, string(valContentTypeJson)},
		},
		ParaForm: url.Values{},
		ParaBody: []byte(jdbxpara),
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
				{paraAuthorization, string(valAuthBearer) + accessToken.token},
				{paraContentType, string(valContentTypeJson)},
			},
			ParaForm: url.Values{},
			ParaBody: []byte(jdbxcont),
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

// MoveFiles -move files to destination folder
func MoveFiles(from, to string) (*FileItemMetadataType, error) {
	var metadata *FileItemMetadataType
	var err error
	err = requestAccessToken()
	if err != nil {
		return nil, err
	}
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
			{paraAuthorization, string(valAuthBearer) + accessToken.token},
			{paraContentType, string(valContentTypeJson)},
		},
		ParaForm: url.Values{},
		ParaBody: []byte(jdbxpara),
	}
	metadata, err = restCall[*FileItemMetadataType](para)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// DeleteFile -delete single file
func DeleteFile(path string) (*FileItemMetadataType, error) {
	var err error
	var metadata *FileItemMetadataType
	var dbxpara FilePathParaType
	err = requestAccessToken()
	if err != nil {
		return nil, err
	}
	dbxpara = FilePathParaType{path}
	jdbxpara, err := anyToJson[FilePathParaType](dbxpara)
	if err != nil {
		return nil, err
	}
	var para = RESTParaType{
		ParaURL:    dropboxAPIURI + endPointFilesDelete,
		ParaMethod: http.MethodPost,
		ParaHeader: []KeyValueType{
			{paraAuthorization, string(valAuthBearer) + accessToken.token},
			{paraContentType, string(valContentTypeJson)},
		},
		ParaForm: url.Values{},
		ParaBody: []byte(jdbxpara),
	}
	metadata, err = restCall[*FileItemMetadataType](para)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// BatchDeleteFiles -delete a bunch of files
func BatchDeleteFiles(path []string) (*FileItemBatchDeletedType, error) {
	var err error
	var metadata *FileItemBatchDeletedType
	var dbxpara DeleteBatchParaType
	var _path FilePathParaType
	var para RESTParaType
	err = requestAccessToken()
	if err != nil {
		return nil, err
	}
	for _, p := range path {
		_path = FilePathParaType{p}
		dbxpara.Entries = append(dbxpara.Entries, _path)
	}
	jdbxpara, err := anyToJson[DeleteBatchParaType](dbxpara)
	if err != nil {
		return nil, err
	}
	para = RESTParaType{
		ParaURL:    dropboxAPIURI + endPointFilesDeleteBatch,
		ParaMethod: http.MethodPost,
		ParaHeader: []KeyValueType{
			{paraAuthorization, string(valAuthBearer) + accessToken.token},
			{paraContentType, string(valContentTypeJson)},
		},
		ParaForm: url.Values{},
		ParaBody: []byte(jdbxpara),
	}
	metadata, err = restCall[*FileItemBatchDeletedType](para)
	if err != nil {
		return nil, err
	}
	tag := metadata.Tag
	id := metadata.AsyncJobId
	var loop = 0
	for tag == DbxAsyncJobId && id != "" {
		err = requestAccessToken()
		if err != nil {
			return nil, err
		}
		// poll async job
		metadata = nil
		batchcheck := BatchCheckParaType{"", id}
		jbatchcheck, err := anyToJson[BatchCheckParaType](batchcheck)
		if err != nil {
			return nil, err
		}
		para = RESTParaType{
			ParaURL:    dropboxAPIURI + endPointFilesDeleteBatchCheck,
			ParaMethod: http.MethodPost,
			ParaHeader: []KeyValueType{
				{paraAuthorization, string(valAuthBearer) + accessToken.token},
				{paraContentType, string(valContentTypeJson)},
			},
			ParaForm: url.Values{},
			ParaBody: []byte(jbatchcheck),
		}
		metadata, err = restCall[*FileItemBatchDeletedType](para)
		if err != nil {
			return nil, err
		}
		switch metadata.Tag {
		case DbxInProgress:
			time.Sleep(pollSleepTime * time.Second)
			loop++
			if loop > maxJobPolls { // deploy parachute
				return nil, errors.New(assets.ErrorAsyncJobTimeOut)
			}
			break
		case DbxComplete:
			id, tag = "", ""
			break
		case DbxFailed:
			return nil, errors.New(assets.ErrorAsyncJobFailed)
		default:
			return nil, errors.New(assets.ErrorAsyncJobUnknownStatus)
		}
	}
	return metadata, nil
}

// UploadFile -upload a single file to Dropbox (max. file size 150MB)
func UploadFile(path string, payload []byte) (*FileItemType, error) {
	var err error
	var para RESTParaType
	var metadata *FileItemType
	opts := UploadFileParaType{
		AutoRename:     false,
		Path:           path,
		Mode:           OverWrite,
		Mute:           false,
		StrictConflict: false,
	}
	jopts, err := anyToJson(opts)
	if err != nil {
		return nil, err
	}
	para = RESTParaType{
		ParaURL:    dropboxContentURI + endPointFilesUpload,
		ParaMethod: http.MethodPost,
		ParaHeader: []KeyValueType{
			{paraAuthorization, string(valAuthBearer) + accessToken.token},
			{paraContentType, string(valContentTypeJson)},
			{paraDbxAPIArg, jopts},
		},
		ParaForm: url.Values{},
		ParaBody: payload,
	}
	metadata, err = restCall[*FileItemType](para)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// CreateFolder -create new folder in Dropbox
func CreateFolder(path string) (*FileItemType, error) {
	var err error
	var metadata *FileItemMetadataType
	var dbxpara = CreateFolderParaType{true, path}
	jdbxpara, err := anyToJson[CreateFolderParaType](dbxpara)
	if err != nil {
		return nil, err
	}
	var para = RESTParaType{
		ParaURL:    dropboxAPIURI + endPointCreateFolder,
		ParaMethod: http.MethodPost,
		ParaHeader: []KeyValueType{
			{paraAuthorization, string(valAuthBearer) + accessToken.token},
			{paraContentType, string(valContentTypeJson)},
		},
		ParaForm: url.Values{},
		ParaBody: []byte(jdbxpara),
	}
	metadata, err = restCall[*FileItemMetadataType](para)
	if err != nil {
		return nil, err
	}
	metadata.Metadata.Tag = DbxFolder
	return &metadata.Metadata, nil
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
				{paraContentType, string(valContentTypeURLForm)},
				{paraAuthorization, string(valAuthBasic) + authString},
			},
			ParaForm: url.Values{
				paraGrantType:    {valRefreshToken},
				paraRefreshToken: {refreshToken},
			},
			ParaBody: nil,
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
