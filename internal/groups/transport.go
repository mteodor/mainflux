package groups

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux/pkg/errors"
)

func DecodeListGroupsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}
	o, err := readUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := readUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	n, err := readStringQuery(r, nameKey)
	if err != nil {
		return nil, err
	}

	m, err := readMetadataQuery(r, metadataKey)
	if err != nil {
		return nil, err
	}

	req := listGroupsReq{
		token:    r.Header.Get("Authorization"),
		offset:   o,
		limit:    l,
		name:     n,
		metadata: m,
		groupID:  bone.GetValue(r, "groupID"),
	}
	return req, nil
}

func DecodeListMemberGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}
	o, err := readUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := readUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	n, err := readStringQuery(r, nameKey)
	if err != nil {
		return nil, err
	}

	m, err := readMetadataQuery(r, metadataKey)
	if err != nil {
		return nil, err
	}

	req := listMemberGroupReq{
		token:    r.Header.Get("Authorization"),
		groupID:  bone.GetValue(r, "groupID"),
		memberID: bone.GetValue(r, "memberID"),
		offset:   o,
		limit:    l,
		name:     n,
		metadata: m,
	}
	return req, nil
}

func DecodeGroupCreate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}
	var req createGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(ErrFailedDecode, err)
	}
	req.token = r.Header.Get("Authorization")
	return req, nil
}

func DecodeGroupUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}
	var req updateGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(ErrFailedDecode, err)
	}
	req.ID = bone.GetValue(r, "groupID")
	req.token = r.Header.Get("Authorization")
	return req, nil
}

func DecodeGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := groupReq{
		token:   r.Header.Get("Authorization"),
		groupID: bone.GetValue(r, "groupID"),
		name:    bone.GetValue(r, "name"),
	}

	return req, nil
}

func DecodeMemberGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := memberGroupReq{
		token:    r.Header.Get("Authorization"),
		groupID:  bone.GetValue(r, "groupID"),
		memberID: bone.GetValue(r, "memberID"),
	}
	return req, nil
}

func readUintQuery(r *http.Request, key string, def uint64) (uint64, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return 0, errInvalidQueryParams
	}

	if len(vals) == 0 {
		return def, nil
	}

	strval := vals[0]
	val, err := strconv.ParseUint(strval, 10, 64)
	if err != nil {
		return 0, errInvalidQueryParams
	}

	return val, nil
}

func readStringQuery(r *http.Request, key string) (string, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return "", errInvalidQueryParams
	}

	if len(vals) == 0 {
		return "", nil
	}

	return vals[0], nil
}

func readMetadataQuery(r *http.Request, key string) (map[string]interface{}, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return nil, errInvalidQueryParams
	}

	if len(vals) == 0 {
		return nil, nil
	}

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(vals[0]), &m)
	if err != nil {
		return nil, errors.Wrap(errInvalidQueryParams, err)
	}

	return m, nil
}
