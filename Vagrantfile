# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure(2) do |config|
  config.ssh.forward_x11 = true
  config.ssh.forward_agent = true

  

  config.vm.synced_folder File.dirname(__FILE__),
    "/home/vagrant/gocode/src/github.com/luizbafilho/fusis"

  config.vm.provider "vmware_fusion" do |provider, override|
    override.vm.box = "bento/ubuntu-16.04"
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "virtualbox" do |provider, override|
    override.vm.box = "bento/ubuntu-16.04"
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "parallels" do |provider, override|
    override.vm.box = "bento/ubuntu-16.04"
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "libvirt" do |provider, override|
    override.vm.box = "yk0/ubuntu-xenial"
    provider.cpus = 4
    provider.memory = "2048"
    provider.driver = "kvm"
  end

  config.vm.define "fusis-1" do |fusis|
    fusis.vm.hostname = "fusis"
    fusis.vm.network "private_network", ip: "192.168.33.10"
    fusis.vm.provision "shell",
      privileged: true,
      name: 'Install dependencies',
      env: { DEBIAN_FRONTEND: 'noninteractive' },
      inline: <<-SHELL
      # Add docker repo
      echo 'Add docker apt repo'
      apt install -y linux-image-extra-$(uname -r) linux-image-extra-virtual apt-transport-https ca-certificates curl software-properties-common
      curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
      add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
      apt -y update
      apt install -y docker-ce
      groupadd docker
      usermod -aG docker vagrant

      #Instaling golang
      wget --show-progress https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz -O /tmp/go1.8.3.linux-amd64.tar.gz
      tar -C /usr/local -xzf /tmp/go1.8.3.linux-amd64.tar.gz
    SHELL
    
    fusis.vm.provision "shell",
      privileged: false,
      name: 'Configure development environment',
      inline: <<-SHELL

      # configure GOPATH and and go binaries to vagrant user
      echo 'Add go envs to $HOME/.bashrc'
      echo 'export GOPATH=$HOME/gocode' >> $HOME/.bashrc
      echo 'export PATH=$GOPATH/bin:/usr/local/go/bin:$PATH' >> $HOME/.bashrc
    SHELL
  end

  config.vm.define "client" do |client|
    client.vm.hostname = "client"
    client.vm.network "private_network", ip: "192.168.33.11"
  end

  config.vm.define "node" do |node|
    node.vm.hostname = "node"
    node.vm.network "private_network", ip: "192.168.33.12"
  end
end
