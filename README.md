Mentor
======

A simple file sharing app over HTTP/HTTPS

Simply provide a list of files and directories to share and they will be
shared without the need for any webserver or any extra configuration.

Installation
------------

 * Linux

```sh
$ curl -sL http://mentor.sh/linux -o /usr/local/bin/mentor
$ sha256sum !$
# Verify SHA
$ chmod +x !$
```

 * Mac

```sh
$ curl -sL http://mentor.sh/darwin -o /usr/local/bin/mentor
$ sha256sum !$
# Verify SHA
$ chmod +x !$
```

 * FreeBSD

```sh
$ curl -sL http://mentor.sh/freebsd -o /usr/local/bin/mentor
$ sha256sum !$
# Verify SHA
$ chmod +x !$
```

Usage
-----

* Supported options

```sh
$ mentor -help
Usage of ./bin/linux/mentor:
  -password string
        Password for accessing the service (username is mentor)
  -port int
        Port number (default 61234)
  -ssl
        Enable HTTPS (SSL)
  -upload
        Allow uploads through the service
  -upload_dir string
        Upload directory (default ".")
  -upload_limit int
        Upload size limit (in MB) (default 2)
```

 * Share specific files

```sh
$ mentor file1 file2
```
