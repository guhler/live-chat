#!/bin/bash

for i in {0..100}; do
    sqlite3 ../db.sqlite "insert into messages (user_id, room_id, time, content) values (1, 1, datetime('now'), 'asdf')"
done
