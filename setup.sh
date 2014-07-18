##!/bin/ash

set -e
 
sudo apt-get update -qq

echo "Installing base stack"

packagelist=(
  curl
  build-essential
  bison
  openssl
  libreadline6
  libreadline-dev
  git-core
  zlib1g
  zlib1g-dev
  libssl-dev
  libyaml-dev
  libxml2-dev
  libxslt-dev
  autoconf
  ssl-cert
  libcurl4-openssl-dev
  lxc
  python-software-properties
  golang
  zsh
)
 
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y ${packagelist[@]}

echo "GOPATH=~/code/go" >> ~/.bashrc

echo "Installing Ruby 2.0"

wget http://ftp.ruby-lang.org/pub/ruby/2.0/ruby-2.0.0-p247.tar.gz &> /dev/null 
tar xf ruby-2.0.0-p247.tar.gz &> /dev/null 
cd ruby-2.0.0-p247/
./configure &> /dev/null 
make &> /dev/null 
sudo make install
cd ..
rm -rf ruby-2.0.0-p247*

sudo gem install bundler

echo "Installing nodejs"
sudo add-apt-repository -y ppa:chris-lea/node.js
sudo apt-get update -qq
sudo apt-get install -y nodejs
