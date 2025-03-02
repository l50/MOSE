# MOSE (Master Of SErvers)

[![Dc27Badge](https://img.shields.io/badge/DEF%20CON-27-green)](https://defcon.org/html/defcon-27/dc-27-speakers.html#Grace)
[![Go Report Card](https://goreportcard.com/badge/github.com/master-of-servers/mose)](https://goreportcard.com/report/github.com/master-of-servers/mose)
[![License](https://img.shields.io/github/license/master-of-servers/mose?label=License&style=flat&color=blue&logo=github)](https://github.com/master-of-servers/mose/blob/master/LICENSE)
[![Build Status](https://dev.azure.com/jaysonegrace/MOSE/_apis/build/status/master-of-servers.MOSE?branchName=master)](https://dev.azure.com/jaysonegrace/MOSE/_build/latest?definitionId=5&branchName=master)

> Copyright (c) 2022 National Technology & Engineering Solutions of Sandia, LLC (NTESS).
> Under the terms of Contract DE-NA0003525 with NTESS,
> the U.S. Government retains certain rights in this software

MOSE is a post exploitation tool that enables security professionals with little or no experience with configuration management (CM) technologies to leverage them to compromise environments. CM tools, such as [Puppet](https://puppet.com/), [Chef](https://www.chef.io/), [Salt](https://www.saltstack.com/), and [Ansible](https://www.ansible.com/) are used to provision systems in a uniform manner based on their function in a network.

Upon successfully compromising a CM server, an attacker can use these tools to run commands on any and all systems that are in the CM server’s inventory. However, if the attacker does not have experience with these types of tools, there can be a very time-consuming learning curve. MOSE allows an operator to specify what they want to run without having to get bogged down in the details of how to write code specific to a proprietary CM tool. It also automatically incorporates the desired commands into existing code on the system, removing that burden from the user.

MOSE allows the operator to choose which assets they want to target within the scope of the server’s inventory, whether this is a subset of clients or all clients. This is useful for targeting specific assets such as web servers or choosing to take over all of the systems in the CM server’s inventory.

## MOSE + Puppet

![MOSE+Puppet](docs/images/mose_and_puppet.gif)

## MOSE + Chef

![MOSE+Chef](docs/images/mose_and_chef.gif)

## Dependencies

You must download and install the following for MOSE to work:

- [Golang](https://golang.org/)
- [Docker](https://docs.docker.com/install/)

## Getting started

Grab the code without having to clone the repo:

```bash
go get -u -v github.com/master-of-servers/mose
```

Install all go-specific dependencies and build the binary (be sure to `cd` into the repo before running this):

```bash
make build
```

### Usage

```bash
Usage:
  github.com/master-of-servers/mose [command]

Available Commands:
  ansible     Create MOSE payload for ansible
  chef        Create MOSE payload for chef
  help        Help about any command
  puppet      Create MOSE payload for puppet
  salt        Create MOSE payload for salt

Flags:
      --basedir string            Location of payloads output by mose (default "/Users/l/programs/go/src/github.com/master-of-servers/mose")
  -c, --cmd string                Command to run on the targets
      --config string             config file (default is $PWD/.settings.yaml)
      --debug                     Display debug output
      --exfilport int             Port used to exfil data from chef server (default 9090, 443 with SSL) (default 9090)
  -f, --filepath string           Output binary locally at <filepath>
  -u, --fileupload string         File upload option
  -h, --help                      help for github.com/master-of-servers/mose
  -l, --localip string            Local IP Address
      --nocolor                   Disable colors for mose
  -a, --osarch string             Architecture that the target CM tool is running on
  -o, --ostarget string           Operating system that the target CM server is on (default "linux")
  -m, --payloadname string        Name for backdoor payload (default "my_cmd")
      --payloads string           Location of payloads output by mose (default "/Users/l/programs/go/src/github.com/master-of-servers/mose/payloads")
      --remoteuploadpath string   Remote file path to upload a script to (used in conjunction with -fu) (default "/root/.definitelynotevil")
  -r, --rhost string              Set the remote host for /etc/hosts in the chef workstation container (format is hostname:ip)
      --ssl                       Serve payload over TLS
      --tts int                   Number of seconds to serve the payload (default 60)
      --websrvport int            Port used to serve payloads (default 8090, 443 with SSL) (default 8090)

Use "github.com/master-of-servers/mose [command] --help" for more information about a command.
```

### TLS Certificates

**You should generate and use a TLS certificate
signed by a trusted Certificate Authority**

A self-signed certificate and key are provided for you, although you really shouldn't use them. This key and certificate are widely distributed, so you can not expect privacy if you do choose to use them. They can be found in the `data` directory.

### Examples

You can find some examples of how to run MOSE in [EXAMPLES.md](EXAMPLES.md).

### Test Labs

Test labs that can be run with MOSE are at these locations:

- [Puppet Test Lab](https://github.com/master-of-servers/puppet-test-lab)
- [Chef Test Lab](https://github.com/master-of-servers/chef-test-lab)
- [Ansible Test Lab](https://github.com/master-of-servers/ansible-test-lab)
- [Salt Test Lab](https://github.com/master-of-servers/salt-test-lab)

### Credits

The following resources were used to help motivate the creation of this project:

- <https://n0tty.github.io/2017/06/11/Enterprise-Offense-IT-Operations-Part-1/>
- <http://www.ryanwendel.com/2017/10/03/cooking-up-shells-with-a-compromised-chef-server/>
