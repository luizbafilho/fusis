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

  config.vm.synced_folder File.dirname(__FILE__), "/home/vagrant/go/src/github.com/luizbafilho/fusis", type: "nfs"

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

  config.vm.provision "shell",
    privileged: true,
    keep_color: true,
    name: 'Install dependencies',
    env: { DEBIAN_FRONTEND: 'noninteractive' },
    inline: <<-SHELL

    echo '\033[0;32m''Add docker apt repo'
    apt-key adv --keyserver hkp://ha.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
    echo "deb https://apt.dockerproject.org/repo ubuntu-xenial main" > /etc/apt/sources.list.d/docker.list

    echo '\033[0;32m''Add lxd apt repo'
    add-apt-repository ppa:ubuntu-lxc/lxd-stable

    echo '\033[0;32m''Add consul apt repo'
    add-apt-repository ppa:bcandrea/consul

    echo '\033[0;32m''Wait for apt lock' # doing this instead of disabling ubuntu auto update
    while fuser /var/lib/dpkg/lock >/dev/null 2>&1; do
      sleep 1
    done

    echo '\033[0;32m''Update apt and install packages'
    apt-get -y update &&
    apt-get install -y docker-engine libnl-3-dev libnl-genl-3-dev build-essential git ipvsadm golang consul

    echo '\033[0;32m''Start docker service'
    systemctl start docker

    echo '\033[0;32m''Start consul service in dev mode'
    echo 'CONSUL_FLAGS="-dev"' >> /etc/default/consul
    systemctl enable consul
    systemctl start consul

    echo '\033[0;32m''Ensure project folder tree has the right ownership'
    f='/home/vagrant/go/src/github.com/luizbafilho'
    while [[ $f != '/home/vagrant' ]]; do chown vagrant: $f; f=$(dirname $f); done;
  SHELL

  config.vm.provision "shell",
    privileged: false,
    keep_color: true,
    name: 'Configure development environment',
    env: { HOME: '/home/vagrant', GOPATH: '/home/vagrant/go' },
    inline: <<-SHELL

    echo '\033[0;32m''Add go envs to .profile'
    cat << EOF >> $HOME/.profile
export GOPATH="$HOME/go"
PATH="$GOPATH/bin:$PATH"
EOF

    echo '\033[0;32m''Link fusis in /home/vagrant for convinience'
    ln -s $GOPATH/src/github.com/luizbafilho/fusis $HOME/fusis

    echo '\033[0;32m''go get'
    PATH="$GOPATH/bin:$PATH"
    cd $GOPATH/src/github.com/luizbafilho/fusis
    go get -v .
  SHELL

  config.vm.post_up_message = <<-MSG
    Fusis VM ready!
    your user is 'vagrant' with password 'vagrant'
    your $GOPATH is /home/vagrant/go
    Fusis code is in /home/vagrant/go/src/github.com/luizbafilho/fusis
    for your convinience it's linked in /home/vagrant/fusis
  MSG
end
