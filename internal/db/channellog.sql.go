// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: channellog.sql

package db

import (
	"context"
	"time"
)

const getChannelLog = `-- name: GetChannelLog :one
SELECT id, description, is_error, url, "method", request, response, response_status, created_on, request_time, channel_id, connection_id, msg_id
FROM public.channels_channellog
WHERE id=$1
`

func (q *Queries) GetChannelLog(ctx context.Context, id int64) (ChannelsChannellog, error) {
	row := q.db.QueryRowContext(ctx, getChannelLog, id)
	var i ChannelsChannellog
	err := row.Scan(
		&i.ID,
		&i.Description,
		&i.IsError,
		&i.Url,
		&i.Method,
		&i.Request,
		&i.Response,
		&i.ResponseStatus,
		&i.CreatedOn,
		&i.RequestTime,
		&i.ChannelID,
		&i.ConnectionID,
		&i.MsgID,
	)
	return i, err
}

const getChannelLogFromChannelID = `-- name: GetChannelLogFromChannelID :many
SELECT id, description, is_error, url, "method", request, response, response_status, created_on, request_time, channel_id, connection_id, msg_id
FROM public.channels_channellog
WHERE channel_id=$1
`

func (q *Queries) GetChannelLogFromChannelID(ctx context.Context, channelID int32) ([]ChannelsChannellog, error) {
	rows, err := q.db.QueryContext(ctx, getChannelLogFromChannelID, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChannelsChannellog
	for rows.Next() {
		var i ChannelsChannellog
		if err := rows.Scan(
			&i.ID,
			&i.Description,
			&i.IsError,
			&i.Url,
			&i.Method,
			&i.Request,
			&i.Response,
			&i.ResponseStatus,
			&i.CreatedOn,
			&i.RequestTime,
			&i.ChannelID,
			&i.ConnectionID,
			&i.MsgID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getChannelLogWithParams = `-- name: GetChannelLogWithParams :many
SELECT id, description, is_error, url, "method", request, response, response_status, created_on, request_time, channel_id, connection_id, msg_id
FROM public.channels_channellog
WHERE channel_id=$1 AND created_on >= $2 AND created_on <= $3
ORDER BY created_on
`

type GetChannelLogWithParamsParams struct {
	ChannelID int32
	After     time.Time
	Before    time.Time
}

func (q *Queries) GetChannelLogWithParams(ctx context.Context, arg GetChannelLogWithParamsParams) ([]ChannelsChannellog, error) {
	rows, err := q.db.QueryContext(ctx, getChannelLogWithParams, arg.ChannelID, arg.After, arg.Before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChannelsChannellog
	for rows.Next() {
		var i ChannelsChannellog
		if err := rows.Scan(
			&i.ID,
			&i.Description,
			&i.IsError,
			&i.Url,
			&i.Method,
			&i.Request,
			&i.Response,
			&i.ResponseStatus,
			&i.CreatedOn,
			&i.RequestTime,
			&i.ChannelID,
			&i.ConnectionID,
			&i.MsgID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
