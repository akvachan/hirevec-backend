create schema if not exists v1;

-- ULID generation
create or replace function v1.generate_ulid() returns text language plpgsql volatile as $$
  declare
    encoding constant text := '0123456789abcdefghjkmnpqrstvwxyz';
    ts bytea := repeat(E'\\000', 6);
    rand bytea := substring(uuid_send(gen_random_uuid()) from 1 for 10);
    ulid bytea;
    t bigint;
    b int[];
    out text[];
  begin
    t := (extract(epoch from clock_timestamp()) * 1000)::bigint;
    ts := set_byte(ts, 0, ((t >> 40) & 255)::int);
    ts := set_byte(ts, 1, ((t >> 32) & 255)::int);
    ts := set_byte(ts, 2, ((t >> 24) & 255)::int);
    ts := set_byte(ts, 3, ((t >> 16) & 255)::int);
    ts := set_byte(ts, 4, ((t >> 8) & 255)::int);
    ts := set_byte(ts, 5, (t & 255)::int);

    ulid := ts || rand;

    b := array[
      get_byte(ulid,0), get_byte(ulid,1), get_byte(ulid,2), get_byte(ulid,3),
      get_byte(ulid,4), get_byte(ulid,5), get_byte(ulid,6), get_byte(ulid,7),
      get_byte(ulid,8), get_byte(ulid,9), get_byte(ulid,10), get_byte(ulid,11),
      get_byte(ulid,12), get_byte(ulid,13), get_byte(ulid,14), get_byte(ulid,15)
    ];

    out := array[
      substr(encoding, ((b[1] & 224) >> 5) + 1, 1),
      substr(encoding, (b[1] & 31) + 1, 1),
      substr(encoding, ((b[2] & 248) >> 3) + 1, 1),
      substr(encoding, (((b[2] & 7) << 2) | ((b[3] & 192) >> 6)) + 1, 1),
      substr(encoding, ((b[3] & 62) >> 1) + 1, 1),
      substr(encoding, (((b[3] & 1) << 4) | ((b[4] & 240) >> 4)) + 1, 1),
      substr(encoding, (((b[4] & 15) << 1) | ((b[5] & 128) >> 7)) + 1, 1),
      substr(encoding, ((b[5] & 124) >> 2) + 1, 1),
      substr(encoding, (((b[5] & 3) << 3) | ((b[6] & 224) >> 5)) + 1, 1),
      substr(encoding, (b[6] & 31) + 1, 1),

      substr(encoding, ((b[7] & 248) >> 3) + 1, 1),
      substr(encoding, (((b[7] & 7) << 2) | ((b[8] & 192) >> 6)) + 1, 1),
      substr(encoding, ((b[8] & 62) >> 1) + 1, 1),
      substr(encoding, (((b[8] & 1) << 4) | ((b[9] & 240) >> 4)) + 1, 1),
      substr(encoding, (((b[9] & 15) << 1) | ((b[10] & 128) >> 7)) + 1, 1),
      substr(encoding, ((b[10] & 124) >> 2) + 1, 1),
      substr(encoding, (((b[10] & 3) << 3) | ((b[11] & 224) >> 5)) + 1, 1),
      substr(encoding, (b[11] & 31) + 1, 1),

      substr(encoding, ((b[12] & 248) >> 3) + 1, 1),
      substr(encoding, (((b[12] & 7) << 2) | ((b[13] & 192) >> 6)) + 1, 1),
      substr(encoding, ((b[13] & 62) >> 1) + 1, 1),
      substr(encoding, (((b[13] & 1) << 4) | ((b[14] & 240) >> 4)) + 1, 1),
      substr(encoding, (((b[14] & 15) << 1) | ((b[15] & 128) >> 7)) + 1, 1),
      substr(encoding, ((b[15] & 124) >> 2) + 1, 1),
      substr(encoding, (((b[15] & 3) << 3) | ((b[16] & 224) >> 5)) + 1, 1),
      substr(encoding, (b[16] & 31) + 1, 1)
    ];

    return array_to_string(out, '');
  end;
$$;

create or replace function v1.generate_ulid(prefix text) returns text language plpgsql volatile as $$
  begin
    return prefix || '_' || v1.generate_ulid();
  end;
$$;

create or replace function v1.generate_ulid_usr() returns text as $$ begin return v1.generate_ulid('usr'); end $$ language plpgsql volatile;
create or replace function v1.generate_ulid_rtk() returns text as $$ begin return v1.generate_ulid('rtk'); end $$ language plpgsql volatile;
create or replace function v1.generate_ulid_can() returns text as $$ begin return v1.generate_ulid('can'); end $$ language plpgsql volatile;
create or replace function v1.generate_ulid_rec() returns text as $$ begin return v1.generate_ulid('rec'); end $$ language plpgsql volatile;
create or replace function v1.generate_ulid_pos() returns text as $$ begin return v1.generate_ulid('pos'); end $$ language plpgsql volatile;
create or replace function v1.generate_ulid_rcm() returns text as $$ begin return v1.generate_ulid('rcm'); end $$ language plpgsql volatile;

do $$ begin create type v1.provider_type as enum ('google','apple'); exception when duplicate_object then null; end $$;
do $$ begin create type v1.reaction_type as enum ('positive','negative','neutral'); exception when duplicate_object then null; end $$;

create table if not exists v1.users (
    id text primary key default v1.generate_ulid_usr(),
    provider v1.provider_type not null,
    provider_user_id varchar(255) not null,
    email varchar(255),
    full_name varchar(255),
    user_name varchar(100) unique,
    updated_at timestamp default now(),
    unique(provider, provider_user_id)
);

create table if not exists v1.refresh_tokens (
    jti text primary key default v1.generate_ulid_rtk(),
    user_id text not null references v1.users(id) on delete cascade,
    expires_at timestamp not null,
    revoked boolean default false,
    unique(jti)
);
create index if not exists idx_refresh_tokens_user_id on v1.refresh_tokens(user_id);

create table if not exists v1.candidates (
    id text primary key default v1.generate_ulid_can(),
    user_id text not null references v1.users(id) on delete cascade,
    about text not null,
    unique(user_id)
);

create table if not exists v1.recruiters (
    id text primary key default v1.generate_ulid_rec(),
    user_id text not null references v1.users(id) on delete cascade
);

create table if not exists v1.positions (
    id text primary key default v1.generate_ulid_pos(),
    recruiter_id text not null references v1.recruiters(id) on delete cascade,
    title text not null,
    description text not null,
    company text,
    unique(title, description, company)
);

create table if not exists v1.recommendations (
    id text primary key default v1.generate_ulid_rcm(),
    position_id text not null references v1.positions(id) on delete cascade,
    candidate_id text not null references v1.candidates(id) on delete cascade,
    unique(position_id, candidate_id)
);
create index if not exists idx_recommendations_position on v1.recommendations(position_id);
create index if not exists idx_recommendations_candidate on v1.recommendations(candidate_id);
create index if not exists idx_recommendations_candidate_id on v1.recommendations(candidate_id, id asc);

create table if not exists v1.reactions (
    recommendation_id text not null references v1.recommendations(id) on delete cascade,
    reactor_type text not null check (reactor_type in ('candidate','recruiter')),
    reactor_id text not null,
    reaction_type v1.reaction_type not null,
    created_at timestamp not null default now(),
    primary key (recommendation_id, reactor_type, reactor_id)
);
create index if not exists idx_reactions_recommendation on v1.reactions(recommendation_id);

create table if not exists v1.matches (
    candidate_id text not null references v1.candidates(id) on delete cascade,
    position_id text not null references v1.positions(id) on delete cascade,
    created_at timestamp not null default now(),
    primary key (candidate_id, position_id)
);

-- users
insert into v1.users (provider, provider_user_id, email, full_name, user_name)
values
    ('google', 'google-001', 'alice@example.com', 'Alice Doe', 'alice_doe'),
    ('google', 'google-002', 'bob@example.com', 'Bob Smith', 'bob_smith'),
    ('google', 'google-003', 'carol@example.com', 'Carol Jones', 'carol_jones')
on conflict (provider, provider_user_id) do nothing;

-- candidates
insert into v1.candidates (user_id, about)
select id, 'Backend developer with 5 years experience'
from v1.users where email = 'alice@example.com'
on conflict do nothing;

insert into v1.candidates (user_id, about)
select id, 'Frontend developer expert in React'
from v1.users where email = 'carol@example.com'
on conflict do nothing;

-- recruiters
insert into v1.recruiters (user_id)
select id from v1.users where email = 'bob@example.com'
on conflict do nothing;

-- positions
insert into v1.positions (recruiter_id, title, description, company)
select r.id, 'Backend Engineer', 'Develop APIs and databases', 'TechCorp'
from v1.recruiters r
where r.user_id = (select id from v1.users where email = 'bob@example.com')
on conflict do nothing;

insert into v1.positions (recruiter_id, title, description, company)
select r.id, 'Frontend Engineer', 'React & UI focused role', 'DesignHub'
from v1.recruiters r
where r.user_id = (select id from v1.users where email = 'bob@example.com')
on conflict do nothing;

insert into v1.positions (recruiter_id, title, description, company)
select r.id, 'Fullstack Developer', 'Backend + Frontend role', 'TechFusion'
from v1.recruiters r
where r.user_id = (select id from v1.users where email = 'dave@example.com')
on conflict do nothing;

insert into v1.positions (recruiter_id, title, description, company)
select r.id, 'UI/UX Designer', 'Design and frontend focus', 'DesignHub'
from v1.recruiters r
where r.user_id = (select id from v1.users where email = 'dave@example.com')
on conflict do nothing;

-- recommendations
insert into v1.recommendations (position_id, candidate_id)
select p.id, c.id
from v1.positions p, v1.candidates c
where p.title = 'Backend Engineer'
  and c.user_id = (select id from v1.users where email = 'alice@example.com')
on conflict do nothing;

insert into v1.recommendations (position_id, candidate_id)
select p.id, c.id
from v1.positions p, v1.candidates c
where p.title = 'Frontend Engineer'
  and c.user_id = (select id from v1.users where email = 'carol@example.com')
on conflict do nothing;

-- candidate reacts to position 
insert into v1.reactions (recommendation_id, reactor_type, reactor_id, reaction_type)
select r.id, 'candidate', c.id, 'positive'
from v1.recommendations r
join v1.candidates c on c.id = r.candidate_id
where c.user_id = (select id from v1.users where email = 'alice@example.com')
on conflict do nothing;

-- recruiter reacts to candidate
insert into v1.reactions (recommendation_id, reactor_type, reactor_id, reaction_type)
select r.id, 'recruiter', rec.id, 'positive'
from v1.recommendations r
join v1.recruiters rec on rec.user_id = (select id from v1.users where email = 'bob@example.com')
join v1.candidates c on c.id = r.candidate_id
where c.user_id = (select id from v1.users where email = 'alice@example.com')
on conflict do nothing;

insert into v1.reactions (recommendation_id, reactor_type, reactor_id, reaction_type)
select r.id, 'candidate', c.id, 'neutral'
from v1.recommendations r
join v1.candidates c on c.id = r.candidate_id
where c.user_id = (select id from v1.users where email = 'carol@example.com')
on conflict do nothing;

insert into v1.reactions (recommendation_id, reactor_type, reactor_id, reaction_type)
select r.id, 'recruiter', rec.id, 'negative'
from v1.recommendations r
join v1.recruiters rec on rec.user_id = (select id from v1.users where email = 'bob@example.com')
join v1.candidates c on c.id = r.candidate_id
where c.user_id = (select id from v1.users where email = 'carol@example.com')
on conflict do nothing;

-- combined user
insert into v1.users (provider, provider_user_id, email, full_name, user_name)
values
    ('google', 'google-004', 'dave@example.com', 'Dave Miller', 'dave_miller')
on conflict (provider, provider_user_id) do nothing;

-- as candidate
insert into v1.candidates (user_id, about)
select id, 'Fullstack developer and part-time recruiter'
from v1.users
where email = 'dave@example.com'
on conflict do nothing;

-- as recruiter
insert into v1.recruiters (user_id)
select id
from v1.users
where email = 'dave@example.com'
on conflict do nothing;

-- recommendations for combined user as candidate
insert into v1.recommendations (position_id, candidate_id)
select p.id, c.id
from v1.positions p
join v1.candidates c on c.user_id = (select id from v1.users where email = 'dave@example.com')
where p.title = 'Backend Engineer'
on conflict do nothing;

insert into v1.recommendations (position_id, candidate_id)
select p.id, c.id
from v1.positions p
join v1.candidates c on c.user_id = (select id from v1.users where email = 'dave@example.com')
where p.title = 'Frontend Engineer'
on conflict do nothing;

-- reactions by combined user as candidate
insert into v1.reactions (recommendation_id, reactor_type, reactor_id, reaction_type)
select r.id, 'candidate', c.id, 'positive'
from v1.recommendations r
join v1.candidates c on c.id = r.candidate_id
where c.user_id = (select id from v1.users where email = 'dave@example.com')
on conflict do nothing;

-- reactions by combined user as recruiter
insert into v1.reactions (recommendation_id, reactor_type, reactor_id, reaction_type)
select r.id, 'recruiter', rec.id, 'neutral'
from v1.recommendations r
join v1.recruiters rec on rec.user_id = (select id from v1.users where email = 'dave@example.com')
join v1.candidates c on c.id = r.candidate_id
where c.user_id <> rec.user_id  -- avoid reacting to self
on conflict do nothing;

-- Add test user
insert into v1.users (provider, provider_user_id, email, full_name, user_name)
values ('google', 'google-test-001', 'test@example.com', 'Test User', 'test_user')
on conflict (provider, provider_user_id) do nothing;

-- Test candidate account
insert into v1.candidates (user_id, about)
select id, 'Test candidate with full-stack experience'
from v1.users where email = 'test@example.com'
on conflict do nothing;

-- Test recruiter account
insert into v1.recruiters (user_id)
select id from v1.users where email = 'test@example.com'
on conflict do nothing;

-- Positions posted by test recruiter
insert into v1.positions (recruiter_id, title, description, company)
select r.id, 'Test Engineer', 'QA and testing focused role', 'TestCorp'
from v1.recruiters r
where r.user_id = (select id from v1.users where email = 'test@example.com')
on conflict do nothing;

-- Recommendations: test candidate recommended for existing positions
insert into v1.recommendations (position_id, candidate_id)
select p.id, c.id
from v1.positions p
join v1.candidates c on c.user_id = (select id from v1.users where email = 'test@example.com')
where p.title in ('Backend Engineer', 'Frontend Engineer', 'Test Engineer')
on conflict do nothing;

-- Recommendations: existing candidates recommended for test recruiter's position
insert into v1.recommendations (position_id, candidate_id)
select p.id, c.id
from v1.positions p
join v1.candidates c on c.user_id != (select id from v1.users where email = 'test@example.com')
where p.recruiter_id = (
    select id from v1.recruiters
    where user_id = (select id from v1.users where email = 'test@example.com')
)
on conflict do nothing;

-- Sample reaction: test candidate reacts positively to Backend Engineer
insert into v1.reactions (recommendation_id, reactor_type, reactor_id, reaction_type)
select r.id, 'candidate', c.id, 'positive'
from v1.recommendations r
join v1.candidates c on c.id = r.candidate_id
join v1.positions p on p.id = r.position_id
where c.user_id = (select id from v1.users where email = 'test@example.com')
  and p.title = 'Backend Engineer'
on conflict do nothing;

-- Match: test candidate + Backend Engineer
insert into v1.matches (candidate_id, position_id)
select c.id, p.id
from v1.candidates c
join v1.positions p on p.title = 'Backend Engineer'
where c.user_id = (select id from v1.users where email = 'test@example.com')
on conflict do nothing;

-- Match: alice + Test Engineer 
insert into v1.matches (candidate_id, position_id)
select c.id, p.id
from v1.candidates c
join v1.users u on u.id = c.user_id
join v1.positions p on p.title = 'Test Engineer'
where u.email = 'alice@example.com'
on conflict do nothing;
