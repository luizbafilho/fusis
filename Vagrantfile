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
    "/home/vagrant/go/src/github.com/luizbafilho/fusis",
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

  config.vm.post_up_message = <<-MSG
    Fusis VM ready!
    To login run `vagrant ssh`
    The user is 'vagrant' with password 'vagrant'
  MSG

  ##########################################
  # The folowing tasks are running as root #
  ##########################################
  config.vm.provision "shell",
    privileged: true,
    keep_color: true,
    name: 'Install dependencies',
    env: { DEBIAN_FRONTEND: 'noninteractive' },
    inline: <<-SHELL

    # Add docker repo
    echo '\033[0;32m''Add docker apt repo'
    apt-key adv --keyserver hkp://ha.pool.sks-keyservers.net:80 \
      --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
    echo "deb http://apt.dockerproject.org/repo ubuntu-xenial main" > \
      /etc/apt/sources.list.d/docker.list

    # Add repo responsible for golang
    echo '\033[0;32m''Add lxd apt repo'
    add-apt-repository ppa:ubuntu-lxc/lxd-stable

    # Configure apt just because we can
    echo '\033[0;32m''Add some custom acquire conf to apt.conf.d/99acquire''\e[0m'
    cat << "EOF" > /etc/apt/apt.conf.d/99acquire
Acquire::Queue-Mode "host";
Acquire::Retries "10";
Acquire::http::Timeout "10";
Acquire::https::Timeout "10";
Acquire::ftp::Timeout "10";
EOF

    # Doing this instead of disabling ubuntu auto update
    echo '\033[0;32m''Wait for apt lock'
    while fuser /var/lib/dpkg/lock >/dev/null 2>&1; do
      sleep 1
    done

    # Install packages
    echo '\033[0;32m''Update apt and install packages'
    apt-get -y update &&
    apt-get install -y --allow-unauthenticated \
      docker-engine libnl-3-dev libnl-genl-3-dev build-essential git ipvsadm golang unzip

    # Check if docker is started
    echo '\033[0;32m''Ensure docker service is running'
    systemctl start docker

    # Unfortunately there is no upto date working consul package
    echo '\033[0;32m''Manually installing consul and creating a service'

    ## create consul user and group
    addgroup --system consul
    adduser --system --no-create-home --ingroup consul consul

    ## download and install binary
    curl -s -o consul.zip https://releases.hashicorp.com/consul/0.7.1/consul_0.7.1_linux_amd64.zip
    unzip -o consul.zip -d /usr/bin
    rm consul.zip

    ## create default environment
    echo 'CONSUL_FLAGS="-dev"' > /etc/default/consul

    ## create folder to persist data, in case -dev flag is disabled
    mkdir /var/lib/consul
    chown -R consul: /var/lib/consul

    ## create configuration folder and configure data_dir and syslog
    mkdir /etc/consul.d
    cat << "EOF" > /etc/consul.d/20-agent.json
{
  "data_dir": "/var/lib/consul",
  "enable_syslog": true
}
EOF
    chown -R consul: /etc/consul.d

    ## create systemd service
    cat << "EOF" > /lib/systemd/system/consul.service
[Unit]
Description=Consul agent
After=network.target
Documentation=man:consul(1)

[Service]
Type=simple
Environment=GOMAXPROCS=2
EnvironmentFile=/etc/default/consul
ExecStart=/usr/bin/consul agent -config-dir=/etc/consul.d $CONSUL_FLAGS
ExecReload=/bin/kill -HUP $MAINPID
User=consul
Group=consul
Restart=on-failure
RestartSec=10
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
EOF
    ## enable and start systemd consul.service
    systemctl enable consul
    systemctl start consul

    # configure a welcome message for motd
    echo '\033[0;32m''Add a welcome message on login'
    cat << "EOF" > /etc/update-motd.d/20-fusis
#!/bin/sh
printf "\n"
printf '\033[0;32m'
cat <<"MSG"
  ___            _
 / __)          (_)
| |__ _   _  ___ _  ___
|  __) | | |/___) |/___)
| |  | |_| |___ | |___ |
|_|   \\____(___/|_(___/

MSG
printf '\e[0m'
echo 'Your $GOPATH is /home/vagrant/go'
echo 'Fusis code is in /home/vagrant/go/src/github.com/luizbafilho/fusis'
echo "For your convinience it's linked in /home/vagrant/fusis"
echo 'Build fusis with `make` and start `make run`'
printf "\n"
EOF
    chmod +x /etc/update-motd.d/20-fusis

    # fusis folder was mounted as root, we neet the parent folders as vagrant
    echo '\033[0;32m''Ensure project folder tree has the right ownership'
    f='/home/vagrant/go/src/github.com/luizbafilho'
    while [[ $f != '/home/vagrant' ]]; do chown vagrant: $f; f=$(dirname $f); done;
  SHELL

  #############################################
  # The folowing tasks are running as vagrant #
  #############################################
  config.vm.provision "shell",
    privileged: false,
    keep_color: true,
    name: 'Configure development environment',
    env: { HOME: '/home/vagrant', GOPATH: '/home/vagrant/go' },
    inline: <<-SHELL

    # configure GOPATH and and go binaries to vagrant user
    echo '\033[0;32m''Add go envs to .profile'
    cat << "EOF" >> $HOME/.profile
# Golang
export GOPATH="$HOME/go"
PATH="$GOPATH/bin:$PATH"
EOF

    # is too tedious navigating gopath, this link is quality of life
    echo '\033[0;32m''Link fusis in /home/vagrant for convinience'
    ln -s $GOPATH/src/github.com/luizbafilho/fusis $HOME/fusis

    # fusis binary checks ~/.fusis for configuration file as fallback
    echo '\033[0;32m''Create a sample config at /home/vagrant/.fusis/fusis.toml'
    mkdir $HOME/.fusis
    cat << EOF > $HOME/.fusis/fusis.toml
store-address = "consul://127.0.0.1:8500"

[interfaces]
inbound = "$(ip r | grep '^192' | cut -f 3 -d ' ')"
outbound = "$(ip r | grep '^default' | cut -f 5 -d ' ')"

[ipam]
ranges = ["192.168.0.0/24"]
EOF

    # better download the project dependencies
    echo '\033[0;32m''go get'
    PATH="$GOPATH/bin:$PATH"
    cd $HOME/fusis
    go get -v .
  SHELL
end
