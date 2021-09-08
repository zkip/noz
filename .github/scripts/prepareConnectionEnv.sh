#!/bin/bash

chmod 600 id_rsa
mkdir -p ~/.ssh
echo "StrictHostKeyChecking=no" >> ~/.ssh/config
