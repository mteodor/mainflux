// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"net/http"
	"os/user"

	"github.com/mainflux/mainflux"
)

var (
	_ mainflux.Response = (*tokenRes)(nil)
	_ mainflux.Response = (*viewUserRes)(nil)
	_ mainflux.Response = (*passwChangeRes)(nil)
	_ mainflux.Response = (*updateGroupRes)(nil)
	_ mainflux.Response = (*viewGroupRes)(nil)
	_ mainflux.Response = (*groupRes)(nil)
	_ mainflux.Response = (*userCreatedRes)(nil)
)

// MailSent message response when link is sent
const MailSent = "Email with reset link is sent"

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

type userCreatedRes struct {
	ID      string
	created bool
}

func (res userCreatedRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res userCreatedRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location": fmt.Sprintf("/users/%s", res.ID),
		}
	}

	return map[string]string{}
}

func (res userCreatedRes) Empty() bool {
	return true
}

type tokenRes struct {
	Token string `json:"token,omitempty"`
}

func (res tokenRes) Code() int {
	return http.StatusCreated
}

func (res tokenRes) Headers() map[string]string {
	return map[string]string{}
}

func (res tokenRes) Empty() bool {
	return res.Token == ""
}

type updateUserRes struct{}

func (res updateUserRes) Code() int {
	return http.StatusOK
}

func (res updateUserRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updateUserRes) Empty() bool {
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

type viewUserRes struct {
	ID       string                 `json:"id"`
	Email    string                 `json:"email"`
	Groups   []user.Group           `json:"groups"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (res viewUserRes) Code() int {
	return http.StatusOK
}

func (res viewUserRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewUserRes) Empty() bool {
	return false
}

type userPageRes struct {
	pageRes
	Users []viewUserRes
}

func (res userPageRes) Code() int {
	return http.StatusOK
}

func (res userPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res userPageRes) Empty() bool {
	return false
}

type viewGroupRes struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	ParentID    string                 `json:"parent_id"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
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

type errorRes struct {
	Err string `json:"error"`
}

type passwChangeRes struct {
	Msg string `json:"msg"`
}

func (res passwChangeRes) Code() int {
	return http.StatusCreated
}

func (res passwChangeRes) Headers() map[string]string {
	return map[string]string{}
}

func (res passwChangeRes) Empty() bool {
	return false
}

type groupRes struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	ParentID    string                 `json:"parent_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	created     bool
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
			"Location": fmt.Sprintf("/users/groups/%s", res.ID),
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

func (res groupPageRes) Code() int {
	return http.StatusOK
}

func (res groupPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res groupPageRes) Empty() bool {
	return false
}
