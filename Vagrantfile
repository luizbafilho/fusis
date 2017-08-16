# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure(2) do |config|
  config.ssh.forward_x11 = true
  config.ssh.forward_agent = true

  config.vm.hostname = "fusis"
  config.vm.network "forwarded_port", guest: 8000, host: 8000
  config.vm.network "private_network", ip: "192.168.33.10"

  config.vm.synced_folder File.dirname(__FILE__),
    "/home/vagrant/gocode/src/github.com/luizbafilho/fusis",
    type: "nfs"

  config.vm.provider "vmware_fusion" do |provider, override|
    override.vm.box = "bento/ubuntu-16.04"
    provider.name = 'fusis'
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "virtualbox" do |provider, override|
    override.vm.box = "bento/ubuntu-16.04"
    provider.name = 'fusis'
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "parallels" do |provider, override|
    override.vm.box = "bento/ubuntu-16.04"
    provider.name = 'fusis'
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "libvirt" do |provider, override|
    override.vm.box = "yk0/ubuntu-xenial"
    provider.name = 'fusis'
    provider.cpus = 4
    provider.memory = "2048"
    provider.driver = "kvm"
  end

  ##########################################
  # The folowing tasks are running as root #
  ##########################################
  config.vm.provision "shell",
    privileged: true,
    name: 'Install dependencies',
    env: { DEBIAN_FRONTEND: 'noninteractive' },
    inline: <<-SHELL
    # Add docker repo
    echo 'Add docker apt repo'
    apt-key adv --keyserver hkp://ha.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
    echo "deb http://apt.dockerproject.org/repo ubuntu-xenial main" >  /etc/apt/sources.list.d/docker.list

    #Instaling golang
    wget --show-progress https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz -O /tmp/go1.8.3.linux-amd64.tar.gz
    tar -C /usr/local -xzf /tmp/go1.8.3.linux-amd64.tar.gz
  SHELL

  #############################################
  # The folowing tasks are running as vagrant #
  #############################################
  config.vm.provision "shell",
    privileged: false,
    name: 'Configure development environment',
    inline: <<-SHELL

    # configure GOPATH and and go binaries to vagrant user
    echo 'Add go envs to $HOME/.bashrc'
    echo 'export GOPATH=$HOME/gocode' >> $HOME/.bashrc
    echo 'export PATH=$GOPATH/bin:/usr/local/go/bin:$PATH' >> $HOME/.bashrc
  SHELL
end
