
create table users (
    id integer primary key autoincrement,
    name text unique not null check (length(name) > 2 and length(name) <= 32),
    password_hash text not null,
    logout_time datetime default current_timestamp
);

create table rooms (
    id integer primary key autoincrement,
    name text unique not null check (length(name) > 0)
);

create table room_user (
    user_id int not null,
    room_id int not null,

    primary key (user_id, room_id),
    foreign key (user_id) references users(id),
    foreign key (room_id) references rooms(id)
);

create table messages (
    id integer primary key autoincrement,
    user_id int not null,
    room_id int not null,
    time datetime not null,
    content text not null check (length(content) > 0),

    foreign key (user_id) references users(id),
    foreign key (room_id) references rooms(id)
);
