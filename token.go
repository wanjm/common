package common

import "context"

type TokenUser struct {
	UserName     string `json:"userName"`
	UserType     int    `json:"userType"`
	Mobile       string `json:"mobile"`
	MCode        string `json:"mCode"`
	Platform     string `json:"platform"`
	Option       int    `json:"option"`
	Token        string `json:"token"`
	Location     string `json:"location"`
	LastLogin    string `json:"lastLogin"`
	UserId       int    `json:"userId"`
	Username     string `json:"username"`
	UserRealName string `json:"userRealName"`
	UserProfile  string `json:"userProfile"`
	UserOrgName  string `json:"userOrgName"`
	OrgId        int    `json:"orgId"`
	RootOrgId    int    `json:"rootOrgId"`
	OrgShortName string `json:"orgShortName"`
	PolyvCataId  string `json:"polyvCataId"`
	CreateTime   int64  `json:"createTime"`
	Language     string `json:"language"`
	Role         string `json:"role"`
	RoleType     int    `json:"roleType"`
	BindOrgIds   []int  `json:"bindOrgIds"`
	BindAllOrg   bool   `json:"bindAllOrg"`
}
type TokenKey struct{}

func GetToken(c context.Context) *TokenUser {
	return c.Value(TokenKey{}).(*TokenUser)
}
