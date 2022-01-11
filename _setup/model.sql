create table users (
	user_id serial not null primary key,
	email character varying(384) not null,
	password character varying(60) not null,
	activated_at timestamp with time zone default null,
	created_at timestamp with time zone default current_timestamp
);


create table activation (
	activation_id serial not null primary key,
	id character varying(36) not null,
	created_at timestamp with time zone default current_timestamp,
	user_id int not null references users (user_id)
);


create table courses (
	course_id serial not null primary key,
	name character varying(64) not null,
	price decimal(16, 2)
);

create or replace function ver () returns trigger language plpgsql as
$$
	begin

		insert into activation (id, user_id) values (
			uuid_generate_v4()::varchar,
			NEW.user_id
		);

		return new;

	end;
$$
;

create trigger ver_tg after insert on users
for each row execute procedure ver()
;


create or replace function done_ver () returns trigger language plpgsql as
$$
	begin

		if NEW.activated_at is not null then

			delete from activation where user_id = OLD.user_id;

		end if;

		return null;

	end;
$$
;

create trigger done_ver_tg after update on users
for each row execute procedure done_ver()
;

insert into courses (name, price) values ('Golang Basics', 800000);