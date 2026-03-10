create schema if not exists v1;

-- ulid base generator
create or replace function generate_ulid() returns text as $$
declare
  encoding   bytea = '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
  timestamp  bytea = e'\\000\\000\\000\\000\\000\\000';
  output     text  = '';
  unix_time  bigint;
  ulid       bytea;
begin
  unix_time = (extract(epoch from clock_timestamp()) * 1000)::bigint;

  timestamp = set_byte(timestamp, 0, ((unix_time >> 40) & 255)::int);
  timestamp = set_byte(timestamp, 1, ((unix_time >> 32) & 255)::int);
  timestamp = set_byte(timestamp, 2, ((unix_time >> 24) & 255)::int);
  timestamp = set_byte(timestamp, 3, ((unix_time >> 16) & 255)::int);
  timestamp = set_byte(timestamp, 4, ((unix_time >> 8)  & 255)::int);
  timestamp = set_byte(timestamp, 5, (unix_time         & 255)::int);


  -- uuid_send converts uuid to raw 16 bytes; take 10 for the random component
  ulid = timestamp || substring(uuid_send(gen_random_uuid()) from 1 for 10);

  output = output || chr(get_byte(encoding, (get_byte(ulid, 0) & 224) >> 5));
  output = output || chr(get_byte(encoding,  get_byte(ulid, 0) & 31));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 1) & 248) >> 3));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 1) & 7) << 2) | ((get_byte(ulid, 2) & 192) >> 6)));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 2) & 62) >> 1));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 2) & 1) << 4) | ((get_byte(ulid, 3) & 240) >> 4)));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 3) & 15) << 1) | ((get_byte(ulid, 4) & 128) >> 7)));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 4) & 124) >> 2));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 4) & 3) << 3) | ((get_byte(ulid, 5) & 224) >> 5)));
  output = output || chr(get_byte(encoding,  get_byte(ulid, 5) & 31));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 6) & 248) >> 3));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 6) & 7) << 2) | ((get_byte(ulid, 7) & 192) >> 6)));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 7) & 62) >> 1));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 7) & 1) << 4) | ((get_byte(ulid, 8) & 240) >> 4)));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 8) & 15) << 1) | ((get_byte(ulid, 9) & 128) >> 7)));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 9) & 124) >> 2));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 9) & 3) << 3) | ((get_byte(ulid, 10) & 224) >> 5)));
  output = output || chr(get_byte(encoding,  get_byte(ulid, 10) & 31));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 11) & 248) >> 3));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 11) & 7) << 2) | ((get_byte(ulid, 12) & 192) >> 6)));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 12) & 62) >> 1));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 12) & 1) << 4) | ((get_byte(ulid, 13) & 240) >> 4)));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 13) & 15) << 1) | ((get_byte(ulid, 14) & 128) >> 7)));
  output = output || chr(get_byte(encoding, (get_byte(ulid, 14) & 124) >> 2));
  output = output || chr(get_byte(encoding, ((get_byte(ulid, 14) & 3) << 3) | ((get_byte(ulid, 15) & 224) >> 5)));
  output = output || chr(get_byte(encoding,  get_byte(ulid, 15) & 31));

  return output;
end
$$ language plpgsql volatile;

-- prefixed ulid helper
create or replace function generate_ulid(prefix text) returns text as $$
begin
  return prefix || '_' || generate_ulid();
end
$$ language plpgsql volatile;

-- per-table ulid generators
create or replace function generate_ulid_usr() returns text as $$ begin return generate_ulid('usr'); end $$ language plpgsql volatile;
create or replace function generate_ulid_rtk() returns text as $$ begin return generate_ulid('rtk'); end $$ language plpgsql volatile;
create or replace function generate_ulid_can() returns text as $$ begin return generate_ulid('can'); end $$ language plpgsql volatile;
create or replace function generate_ulid_rec() returns text as $$ begin return generate_ulid('rec'); end $$ language plpgsql volatile;
create or replace function generate_ulid_pos() returns text as $$ begin return generate_ulid('pos'); end $$ language plpgsql volatile;

-- providers
do $$ begin
    create type v1.provider_type as enum ('google', 'apple');
exception
    when duplicate_object then null;
end $$;

-- reactions
do $$ begin
    create type v1.reaction_type as enum ('positive', 'negative', 'neutral');
exception
    when duplicate_object then null;
end $$;

-- users
create table if not exists v1.users (
    id               text primary key default generate_ulid_usr(),
    provider         v1.provider_type not null,
    provider_user_id varchar(255) not null,
    email            varchar(255),
    full_name        varchar(255),
    user_name        varchar(100) unique,
    updated_at       timestamp default now(),
    unique(provider, provider_user_id)
);

-- refresh tokens
create table if not exists v1.refresh_tokens (
    jti        text primary key default generate_ulid_rtk(),
    user_id    text      not null references v1.users(id) on delete cascade,
    expires_at timestamp not null,
    revoked    boolean   default false,
    unique(jti)
);

create index if not exists idx_refresh_tokens_user_id
    on v1.refresh_tokens(user_id);

-- candidates
create table if not exists v1.candidates (
    id      text primary key default generate_ulid_can(),
    user_id text not null references v1.users(id) on delete cascade,
    about   text not null,
    unique(user_id)
);

-- recruiters
create table if not exists v1.recruiters (
    id      text primary key default generate_ulid_rec(),
    user_id text not null references v1.users(id) on delete cascade
);

-- positions
create table if not exists v1.positions (
    id          text primary key default generate_ulid_pos(),
    title       text not null,
    description text not null,
    company     text,
    unique(title, description, company)
);

create table if not exists v1.candidates_reactions (
    candidate_id  text not null references v1.candidates(id) on delete cascade,
    position_id   text not null references v1.positions(id)  on delete cascade,
    reaction_type v1.reaction_type not null,
    created_at    timestamp not null default now(),
    primary key (candidate_id, position_id)
);

create table if not exists v1.recruiters_reactions (
    recruiter_id  text not null references v1.recruiters(id)  on delete cascade,
    position_id   text not null references v1.positions(id)   on delete cascade,
    candidate_id  text not null references v1.candidates(id)  on delete cascade,
    reaction_type v1.reaction_type not null,
    created_at    timestamp not null default now(),
    primary key (recruiter_id, position_id, candidate_id)
);

-- matches
create table if not exists v1.matches (
    candidate_id text not null references v1.candidates(id) on delete cascade,
    position_id  text not null references v1.positions(id)  on delete cascade,
    created_at   timestamp not null default now(),
    primary key (candidate_id, position_id)
);

-- admins
create table if not exists v1.admins (
    id      text primary key default generate_ulid_rec(),
    user_id text not null references v1.users(id) on delete cascade
);

-- test data
insert into v1.users (provider, provider_user_id, email, full_name, user_name)
values
    ('google', 'google-123', 'candidate@test.com', 'jane doe',   'jane_doe'),
    ('google', 'google-456', 'recruiter@test.com', 'john smith', 'john_smith'),
    ('google', 'google-789', 'combined@test.com', 'alice parker', 'alice_parker')
on conflict (provider, provider_user_id) do nothing;

insert into v1.candidates (user_id, about)
select id, 'Backend developer with 5 years of experience'
from v1.users
where email = 'candidate@test.com'
on conflict do nothing;

insert into v1.candidates (user_id, about)
select id, 'Frontend developer'
from v1.users
where email = 'combined@test.com'
on conflict do nothing;

insert into v1.recruiters (user_id)
select id
from v1.users
where email = 'recruiter@test.com'
on conflict do nothing;

insert into v1.recruiters (user_id)
select id
from v1.users
where email = 'combined@test.com'
on conflict do nothing;

insert into v1.positions (title, description, company)
values
    ('backend engineer',    'work on apis and databases', 'acme inc'),
    ('fullstack developer', 'react + node.js role',       'tech corp')
on conflict do nothing;

insert into v1.candidates_reactions (candidate_id, position_id, reaction_type)
select c.id, p.id, 'positive'
from v1.candidates c, v1.positions p
where c.id = (select id from v1.candidates limit 1)
  and p.id = (select id from v1.positions  limit 1)
on conflict do nothing;

insert into v1.recruiters_reactions (recruiter_id, position_id, candidate_id, reaction_type)
select r.id, p.id, c.id, 'positive'
from v1.recruiters r, v1.positions p, v1.candidates c
where r.id = (select id from v1.recruiters limit 1)
  and p.id = (select id from v1.positions  limit 1)
  and c.id = (select id from v1.candidates limit 1)
on conflict do nothing;

insert into v1.matches (candidate_id, position_id)
select c.id, p.id
from v1.candidates c, v1.positions p
where c.id = (select id from v1.candidates limit 1)
  and p.id = (select id from v1.positions  limit 1)
on conflict do nothing;

insert into v1.refresh_tokens (user_id, expires_at)
select id, now() + interval '30 days'
from v1.users
where email = 'candidate@test.com';
