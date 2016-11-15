# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure(2) do |config|
  config.ssh.forward_x11 = true
  config.ssh.forward_agent = true

  config.vm.network "forwarded_port", guest: 8000, host: 8000

  config.vm.network "private_network", ip: "192.168.33.10"

  config.vm.synced_folder "#{ENV['GOPATH']}/src", "/home/vagrant/gocode/src", type: "nfs"

  config.vm.provider "vmware_fusion" do |provider, override|
    override.vm.box = "geerlingguy/ubuntu1404"
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "virtualbox" do |provider, override|
    override.vm.box = "ubuntu/trusty64"
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "parallels" do |provider, override|
    override.vm.box = "parallels/ubuntu-14.04"
    provider.cpus = 4
    provider.memory = "2048"
  end

  config.vm.provider "libvirt" do |provider, override|
    override.vm.box = "sputnik13/trusty64"
    provider.cpus = 4
    provider.memory = "2048"
    provider.driver = "kvm"
  end

  config.vm.provision "shell", privileged: false, inline: <<-SHELL
    sudo apt-get update -y
    sudo apt-get install curl wget python3-pip

    HOME=/home/vagrant

    echo "====> Installing docker"
    sudo apt-get -y install apt-transport-https ca-certificates
    sudo apt-key -y adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
    echo "deb https://apt.dockerproject.org/repo ubuntu-trusty main" | sudo tee /etc/apt/sources.list.d/docker.list
    sudo apt-get update -y
    sudo apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual
    sudo apt-get install -y docker-engine
    sudo service docker start
    sudo usermod -aG docker vagrant

    echo "====> Installing vim-gnome"
    sudo locale-gen UTF-8
    sudo apt-get install -y vim-gnome

    echo "====> Installing dependencies"
    sudo apt-get install -y zsh silversearcher-ag software-properties-common libnl-3-dev libnl-genl-3-dev build-essential vim git cmake python-dev ipvsadm exuberant-ctags autojump xauth
    sudo add-apt-repository ppa:neovim-ppa/unstable
    sudo apt-get update -y
    sudo apt-get install -y neovim

    echo "====> Installing Oh my ZSH"
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"

    echo "====> Installing Go"
    curl -O https://storage.googleapis.com/golang/go1.6.3.linux-amd64.tar.gz
    tar -xvf go1.6.3.linux-amd64.tar.gz
    sudo mv go /usr/local
    echo export PATH=$PATH:/usr/local/go/bin >> $HOME/.zshrc
    mkdir $HOME/gocode
    echo export GOPATH=$HOME/gocode >> $HOME/.zshrc

    echo "====> Installing tmux 2.1"
    sudo apt-get build-dep -y tmux
    git clone https://github.com/tmux/tmux.git
    cd tmux
    git checkout 2.1
    sh autogen.sh
    ./configure && make
    sudo make install
    wget https://gist.githubusercontent.com/luizbafilho/99c6ec91b0c3415df75b4c4cf7d0265a/raw/bb10b105f4809c3549e20777e1afdde9b50bc915/.tmux.conf -O  $HOME/.tmux.conf

    echo "====> Downloading vimfiles"
    mkdir $HOME/.config
    git clone https://github.com/luizbafilho/vimfiles.git $HOME/.config/nvim
    nvim +PlugInstall +qa! && echo "Done! :)"
  SHELL
end
