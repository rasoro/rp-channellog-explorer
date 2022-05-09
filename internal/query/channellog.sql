
CREATE TABLE public.channels_channellog (
	id bigserial NOT NULL,
	description varchar(255) NOT NULL,
	is_error bool NOT NULL,
	url text NULL,
	"method" varchar(16) NULL,
	request text NULL,
	response text NULL,
	response_status int4 NULL,
	created_on timestamptz NOT NULL,
	request_time int4 NULL,
	channel_id int4 NOT NULL,
	connection_id int4 NULL,
	msg_id int8 NULL,
	CONSTRAINT channels_channellog_pkey PRIMARY KEY (id)
);



-- name: GetChannelLog :one
SELECT id, description, is_error, url, "method", request, response, response_status, created_on, request_time, channel_id, connection_id, msg_id
FROM public.channels_channellog
WHERE id=$1;

-- name: GetChannelLogFromChannelID :many
SELECT id, description, is_error, url, "method", request, response, response_status, created_on, request_time, channel_id, connection_id, msg_id
FROM public.channels_channellog
WHERE channel_id=$1;