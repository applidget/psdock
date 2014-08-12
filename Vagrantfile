# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "hashicorp/precise64"
  
  #supposed clean workspace: eg ~/code/go
  config.vm.synced_folder "~/code/go", "/home/vagrant/code/go/"
  
  config.vm.provision "shell", path: "setup.sh"

  config.vm.network :public_network

end
