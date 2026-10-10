-- Test fixture for columns referenced by the application.
-- The application never creates or alters production tables.
CREATE TABLE invite_codes (id serial PRIMARY KEY, code text UNIQUE NOT NULL, used integer NOT NULL DEFAULT 0);
CREATE TABLE admin_users (id serial PRIMARY KEY, username text UNIQUE NOT NULL, password text NOT NULL,
    level integer NOT NULL, invite_code_id integer REFERENCES invite_codes(id));
CREATE TABLE virtual_domains (id serial PRIMARY KEY, name text UNIQUE NOT NULL, admin_user_id integer REFERENCES admin_users(id));
CREATE TABLE virtual_users (id serial PRIMARY KEY, domain_id integer REFERENCES virtual_domains(id), password text NOT NULL, email text UNIQUE NOT NULL);
CREATE TABLE virtual_aliases (id serial PRIMARY KEY, domain_id integer NOT NULL, source text NOT NULL, destination text NOT NULL);
CREATE TABLE recipient_bcc (id serial PRIMARY KEY, domain_id integer NOT NULL, source text NOT NULL, destination text NOT NULL, region text NOT NULL);
CREATE TABLE transport_domains (id serial PRIMARY KEY, domain_id integer NOT NULL, source text NOT NULL, destination text NOT NULL, region text NOT NULL);
-- Production has only my_networks; do not mask misspelled queries with aliases.
CREATE TABLE my_networks (id serial PRIMARY KEY, domain_name text, server_mark text, region_mark text, default_mark text);
CREATE TABLE opendkim_keys (id serial PRIMARY KEY, domain_name text UNIQUE NOT NULL, selector text NOT NULL, private_key text NOT NULL);
CREATE TABLE opendkim_signings (id serial PRIMARY KEY, author text UNIQUE NOT NULL, dkim_id integer REFERENCES opendkim_keys(id));
