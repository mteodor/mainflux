package groups

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mainflux/mainflux"
)

var (
	_ mainflux.Response = (*memberPageRes)(nil)
	_ mainflux.Response = (*groupRes)(nil)
	_ mainflux.Response = (*deleteRes)(nil)
	_ mainflux.Response = (*assignRes)(nil)
	_ mainflux.Response = (*unassignRes)(nil)
)

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

type viewGroupRes struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	ParentID    string                 `json:"parent_id,omitempty"`
	OwnerID     string                 `json:"owner_id"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	// Indicates a level in tree hierarchy from first group node - root.
	Level int `json:"level"`
	// Path is a path in a tree, consisted of group names
	// parentName.childrenName1.childrenName2 .
	Path      string          `json:"path"`
	Children  []*viewGroupRes `json:"children,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
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

type groupRes struct {
	id      string
	created bool
}

func (res groupRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res groupRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location": fmt.Sprintf("/groups/%s", res.id),
		}
	}

	return map[string]string{}
}

func (res groupRes) Empty() bool {
	return true
}

type groupPageRes struct {
	pageRes
	Groups []viewGroupRes `json:"groups"`
}

type pageRes struct {
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
	Total  uint64 `json:"total"`
	Level  uint64 `json:"level"`
	Name   string `json:"name"`
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

type deleteRes struct{}

func (res deleteRes) Code() int {
	return http.StatusNoContent
}

func (res deleteRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteRes) Empty() bool {
	return true
}

type assignRes struct{}

func (res assignRes) Code() int {
	return http.StatusNoContent
}

func (res assignRes) Headers() map[string]string {
	return map[string]string{}
}

func (res assignRes) Empty() bool {
	return true
}

type unassignRes struct{}

func (res unassignRes) Code() int {
	return http.StatusNoContent
}

func (res unassignRes) Headers() map[string]string {
	return map[string]string{}
}

func (res unassignRes) Empty() bool {
	return true
}
