CREATE TABLE public.channels_channel (
	id serial4 NOT NULL,
	is_active bool NOT NULL,
	created_on timestamptz NOT NULL,
	modified_on timestamptz NOT NULL,
	uuid varchar(36) NOT NULL,
	channel_type varchar(3) NOT NULL,
	"name" varchar(64) NULL,
	address varchar(255) NULL,
	country varchar(2) NULL,
	claim_code varchar(16) NULL,
	secret varchar(64) NULL,
	last_seen timestamptz NOT NULL,
	device varchar(255) NULL,
	os varchar(255) NULL,
	alert_email varchar(254) NULL,
	config text NULL,
	schemes _varchar NOT NULL,
	"role" varchar(4) NOT NULL,
	bod text NULL,
	tps int4 NULL,
	created_by_id int4 NOT NULL,
	modified_by_id int4 NOT NULL,
	org_id int4 NULL,
	parent_id int4 NULL,
	CONSTRAINT channels_channel_claim_code_key UNIQUE (claim_code),
	CONSTRAINT channels_channel_pkey PRIMARY KEY (id),
	CONSTRAINT channels_channel_secret_key UNIQUE (secret),
	CONSTRAINT channels_channel_uuid_key UNIQUE (uuid)
);

-- name: GetChannel :one
SELECT id, is_active, created_on, modified_on, uuid, channel_type, "name", address, country, claim_code, secret, last_seen, device, os, alert_email, config, schemes, "role", bod, tps, created_by_id, modified_by_id, org_id, parent_id
FROM public.channels_channel
WHERE uuid=$1;
