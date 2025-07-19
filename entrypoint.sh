#!/bin/sh

if [ ! -f .env ]; then
    env > .env
fi 

$@