#!/bin/bash

for tb in $(sqlite3 db.sqlite .tables)
do
    $(sqlite3 db.sqlite "drop table $tb")
done
