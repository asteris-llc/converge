# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
	config.vm.box = "centos/7"
	config.vm.synced_folder ".", "/vagrant", disabled: true
	config.vm.synced_folder "./samples", "/converge_samples"
	config.vm.provision :converge do |cvg|
		cvg.bikeshed = [
			"/converge_samples/basic.hcl"
		]
		cvg.install = true
	end
end
