package groups

import (
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux"
)

var (
	_ mainflux.Response = (*memberPageRes)(nil)
	_ mainflux.Response = (*createGroupRes)(nil)
	_ mainflux.Response = (*updateGroupRes)(nil)
	_ mainflux.Response = (*viewGroupRes)(nil)
	_ mainflux.Response = (*groupDeleteRes)(nil)
	_ mainflux.Response = (*assignMemberToGroupRes)(nil)
	_ mainflux.Response = (*removeMemberFromGroupRes)(nil)
)

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}
type memberPageRes struct {
	pageRes
	Members []interface{}
}

func (res memberPageRes) Code() int {
	return http.StatusOK
}

func (res memberPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res memberPageRes) Empty() bool {
	return false
}

type createGroupRes struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	ParentID    string                 `json:"parent_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	created     bool
}

func (res createGroupRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res createGroupRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location": fmt.Sprintf("/groups/%s", res.ID),
		}
	}
	return map[string]string{}
}

func (res createGroupRes) Empty() bool {
	return true
}

type updateGroupRes struct{}

func (res updateGroupRes) Code() int {
	return http.StatusOK
}

func (res updateGroupRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updateGroupRes) Empty() bool {
	return true
}

type viewGroupRes struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	ParentID    string                 `json:"parent_id"`
	OwnerID     string                 `json:"owner_id"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Level       int                    `json:"level"`
	Path        string                 `json:"path"`
}

func (res viewGroupRes) Code() int {
	return http.StatusOK
}

func (res viewGroupRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewGroupRes) Empty() bool {
	return false
}

type groupPageRes struct {
	pageRes
	Groups []viewGroupRes `json:"groups"`
}

func (res groupPageRes) Code() int {
	return http.StatusOK
}

func (res groupPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res groupPageRes) Empty() bool {
	return false
}

type groupDeleteRes struct{}

func (res groupDeleteRes) Code() int {
	return http.StatusNoContent
}

func (res groupDeleteRes) Headers() map[string]string {
	return map[string]string{}
}

func (res groupDeleteRes) Empty() bool {
	return true
}

type assignMemberToGroupRes struct{}

func (res assignMemberToGroupRes) Code() int {
	return http.StatusNoContent
}

func (res assignMemberToGroupRes) Headers() map[string]string {
	return map[string]string{}
}

func (res assignMemberToGroupRes) Empty() bool {
	return true
}

type removeMemberFromGroupRes struct{}

func (res removeMemberFromGroupRes) Code() int {
	return http.StatusNoContent
}

func (res removeMemberFromGroupRes) Headers() map[string]string {
	return map[string]string{}
}

func (res removeMemberFromGroupRes) Empty() bool {
	return true
}
