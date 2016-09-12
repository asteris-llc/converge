#!/bin/bash
cd $(dirname "$0")

vagrant up --provision
vagrant destroy --force
