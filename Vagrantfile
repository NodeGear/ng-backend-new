# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
	config.vm.define "dev_backend"
	config.vm.hostname = "dev-backend"

	config.vm.box = "ubuntu/trusty64"
	config.vm.network "private_network", ip: "10.0.3.6"

	config.vm.synced_folder ".", "/var/lib/backend", type: "rsync", rsync__exclude: [".git/"]

	config.vm.provider "virtualbox" do |v|
		v.memory = 2048
		v.cpus = 2
	end

	config.vm.provision :ansible do |ansible|
		ansible.groups = {
			"proxy" => ["dev_backend"],
			"backend" => ["dev_backend"],
			"development:children" => ["backend"]
		}

		ansible.playbook = "~/p/ng-infrastructure/infrastructure.yml"

		ansible.limit = 'all'
	end

#	config.vm.provision :ansible do |ansible|
#		ansible.groups = {
#			"proxy" => ["dev_backend"],
#			"backend" => ["dev_backend"],
#			"development:children" => ["backend"]
#		}
#
#		ansible.playbook = "/Users/matejkramny/Projects/ng-infrastructure/backend.yml"
#	end

	config.vm.provision :ansible do |ansible|
		ansible.groups = {
			"proxy" => ["dev_backend"],
			"backend" => ["dev_backend"],
			"development:children" => ["backend", "proxy"]
		}

		ansible.playbook = "~/p/ng-infrastructure/proxy.yml"
	end
end
