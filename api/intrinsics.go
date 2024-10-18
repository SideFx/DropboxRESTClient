// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// REST API & Dropbox data structures etc.
// ---------------------------------------------------------------------------------------------------------------------

package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	dropboxAuthURI    = "https://www.dropbox.com/oauth2/authorize"
	dropboxAPIURI     = "https://api.dropbox.com"
	dropboxContentURI = "https://content.dropbox.com"
)

const (
	endpointAuthToken          = "/oauth2/token"
	endpointGetCurrentUser     = "/2/users/get_current_account"
	endpointListFolder         = "/2/files/list_folder"
	endpointListFolderContinue = "/2/files/list_folder/continue"
	endPointFilesMove          = "/2/files/move_v2"
)

const (
	paraResponseType    = "response_type="
	paraClientId        = "client_id="
	paraTokenAccessType = "token_access_type="
)

const (
	paraAuthorization = "Authorization"
	paraContentType   = "Content-Type"
	paraGrantType     = "grant_type"
	paraCode          = "code"
	paraRefreshToken  = "refresh_token"
)

const (
	valResponseType       = "code"
	valTokenAccessType    = "offline"
	valContentTypeURLForm = "application/x-www-form-urlencoded"
	valContentTypeJson    = "application/json"
	valAuthorizationCode  = "authorization_code"
	valRefreshToken       = "refresh_token"
)

const (
	valAuthBasic  = "Basic "
	valAuthBearer = "Bearer "
)

const (
	DbxFile          = "file"
	DbxFolder        = "folder"
	DbxPathSeparator = "/"
)

const threshold = 10 // safety time span for requesting new access token

type AppAuthType struct {
	AppKey    string
	AppSecret string
}

type accessTokenType struct {
	token     string
	expiresIn int64
	fetchedAt int64
}

type KeyValueType struct {
	Key   string
	Value string
}

type RESTParaType struct {
	ParaURL    string
	ParaMethod string
	ParaHeader []KeyValueType
	ParaForm   url.Values
	ParaBody   string
}

//----------------------------------------------------------------------------------------------------------------------

type ListFoldersParaType struct {
	IncludeDeleted                  bool   `json:"include_deleted"`
	IncludeHasExplicitSharedMembers bool   `json:"include_has_explicit_shared_members"`
	IncludeMountedFolders           bool   `json:"include_mounted_folders"`
	IncludeNonDownloadableFiles     bool   `json:"include_non_downloadable_files"`
	Path                            string `json:"path"`
	Recursive                       bool   `json:"recursive"`
	Limit                           uint32 `json:"limit"`
}

type ListContinueType struct {
	Cursor string `json:"cursor"`
}

type FilesMoveParaType struct {
	AllowOwnershipTransfer bool   `json:"allow_ownership_transfer"`
	Autorename             bool   `json:"autorename"`
	FromPath               string `json:"from_path"`
	ToPath                 string `json:"to_path"`
}

//----------------------------------------------------------------------------------------------------------------------

type RefreshTokenType struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	AccountId    string `json:"account_id"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
	Uid          string `json:"uid"`
}

type UserMessageType struct {
	Text   string `json:"text"`
	Locale string `json:"locale"`
}

type ErrorType struct {
	ErrorSummary     string          `json:"error_summary"`
	Error            string          `json:"error"`
	ErrorDescription string          `json:"error_description"`
	UserMessage      UserMessageType `json:"user_message"`
}

type NameType struct {
	AbbreviatedName string `json:"abbreviated_name"`
	DisplayName     string `json:"display_name"`
	FamiliarName    string `json:"familiar_name"`
	GivenName       string `json:"given_name"`
	Surname         string `json:"surname"`
}

type RootInfoType struct {
	Tag             string `json:".tag"`
	HomeNamespaceId string `json:"home_namespace_id"`
	RootNamespaceId string `json:"root_namespace_id"`
}

type AccountTypeType struct {
	Tag string `json:".tag"`
}

type UserInfoType struct {
	AccountId       string          `json:"account_id"`
	AccountType     AccountTypeType `json:"account_type"`
	Country         string          `json:"country"`
	Disabled        bool            `json:"disabled"`
	Email           string          `json:"email"`
	EmailVerified   bool            `json:"email_verified"`
	IsPaired        bool            `json:"is_paired"`
	Locale          string          `json:"locale"`
	Name            NameType        `json:"name"`
	ProfilePhotoUrl string          `json:"profile_photo_url"`
	ReferralLink    string          `json:"referral_link"`
	RootInfo        RootInfoType    `json:"root_info"`
}

type FileLockInfoType struct {
	Created        string `json:"created"`
	IsLockholder   bool   `json:"is_lockholder"`
	LockholderName string `json:"lockholder_name"`
}

type FieldType struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PropertyGroupType struct {
	Fields     []FieldType `json:"fields"`
	TemplateId string      `json:"template_id"`
}

type SharingInfoType struct {
	ModifiedBy           string `json:"modified_by"`
	ParentSharedFolderId string `json:"parent_shared_folder_id"`
	ReadOnly             bool   `json:"read_only"`
	NoAccess             bool   `json:"no_access"`
	TraverseOnly         bool   `json:"traverse_only"`
}

type FileItemType struct {
	Tag                      string              `json:".tag"`
	ClientModified           string              `json:"client_modified"`
	ContentHash              string              `json:"content_hash"`
	FileLockInfo             FileLockInfoType    `json:"file_lock_info"`
	HasExplicitSharedMembers bool                `json:"has_explicit_shared_members"`
	Id                       string              `json:"id"`
	IsDownloadable           bool                `json:"is_downloadable"`
	Name                     string              `json:"name"`
	PathDisplay              string              `json:"path_display"`
	PathLower                string              `json:"path_lower"`
	PropertyGroups           []PropertyGroupType `json:"property_groups"`
	Rev                      string              `json:"rev"`
	ServerModified           string              `json:"server_modified"`
	SharingInfo              SharingInfoType     `json:"sharing_info"`
	Size                     int64               `json:"size"`
}

type FileItemMetadataType struct {
	Metadata FileItemType `json:"metadata"`
}

type ItemInfoType struct {
	Cursor  string         `json:"cursor"`
	Entries []FileItemType `json:"entries"`
	HasMore bool           `json:"has_more"`
}

//----------------------------------------------------------------------------------------------------------------------

var authkey AppAuthType
var accessToken accessTokenType
var refreshToken string
var currentUserId string

//----------------------------------------------------------------------------------------------------------------------

// SetConnectionData -receive connection data from ui
func SetConnectionData(key AppAuthType, token string) {
	authkey = key
	refreshToken = token
}

// restCall -generic REST call
func restCall[T any](para RESTParaType) (T, error) {
	var result T
	var dbxerror ErrorType
	var err error
	var errorString string
	var requestbody io.Reader = nil
	if len(para.ParaForm) > 0 {
		requestbody = strings.NewReader(para.ParaForm.Encode()) // form fields
	} else {
		if para.ParaBody != "" {
			requestbody = strings.NewReader(para.ParaBody) // json
		}
	}
	req, err := http.NewRequest(para.ParaMethod, para.ParaURL, requestbody)
	if err != nil {
		return result, err
	}
	for _, h := range para.ParaHeader {
		req.Header.Add(h.Key, h.Value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &result)
		return result, err
	} else {
		_ = json.Unmarshal(body, &dbxerror)
		if dbxerror.ErrorSummary != "" {
			errorString = dbxerror.ErrorSummary
		} else {
			if dbxerror.ErrorDescription != "" {
				errorString = dbxerror.Error + " " + dbxerror.ErrorDescription
			} else {
				errorString = string(body)
			}
		}
		return result, errors.New(errorString)
	}
}

// anyToJson -generic JSON transformation
func anyToJson[T any](v T) (string, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(j), nil
}
