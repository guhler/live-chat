#!/bin/bash

for t in $(sqlite3 db.sqlite .tables); do
    sqlite3 db.sqlite "drop table $t"
done
