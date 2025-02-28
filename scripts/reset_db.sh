#!/bin/bash

. scripts/clear_db.sh
sqlite3 ../db.sqlite < db.sql
