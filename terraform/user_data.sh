#!/bin/bash
set -e

# Update packages and install Docker
dnf update -y
dnf install -y docker
systemctl enable --now docker
usermod -aG docker ec2-user

echo "Docker installed successfully"
