#!/bin/bash

. clear_db.sh
sqlite3 db.sqlite < db.sql
