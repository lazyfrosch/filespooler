File Spooler
============

This is a file spool tool that transports files from a directory via a client-server connection to another server.

The idea is to be as minimalistic as possible without using tools like rsync or an SSH connection.

## Installation

    $ go get github.com/lazyfrosch/filespooler/cmd/filespooler
    
    $ git clone https://github.com/lazyfrosch/filespooler
    $ cd filespooler/
    $ go build ./cmd/filespooler
    $ ./filespooler

## Usage

Both client and server are daemons that work in the background.

    $ filespooler receiver -listen :5664 -target /var/spool/data
    
    $ filespooler sender -connect localhost:5664 -source /var/spool/tool/output

## Known Issues

* TLS encryption needs to be implemented
* Improve logging

## License

    Copyright (C) 2019 Markus Frosch <markus.frosch@netways.de>
                  2019 NETWAYS GmbH <info@netways.de>

    This program is free software; you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation; either version 2 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License along
    with this program; if not, write to the Free Software Foundation, Inc.,
    51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
