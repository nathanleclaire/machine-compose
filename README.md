# moby

Recently I've been working on `libmachine` and I wanted to show off some of what
it can do.

`moby` is a little toy program that takes the same trick `docker-compose` does
with containers, and starts to apply it towards machines that run docker.

## building

The dependencies are vendored so you just need to:

```
$ make
```

No fancy container stuff unfortunately, so you need to have Go installed.

It will spit out `moby` binary.

I have used a custom branch of `libmachine` so sorry if it is a little hard to
follow the dependencies.

## usage

Define a `moby.yml` file, such as the one included in this repository, it should
look like this:

```yaml
hostoptions:
  engineoptions:
    storagedriver: aufs
    labels:
      - foo=bar
      - baz=quux
      - far=go
    arbitraryflags:
      - dns=8.8.8.8
driveroptions:
  accesstoken: <token>
  region: sfo1
  size: 2gb
```

The keys current map very directly to internal `libmachine` stuff, and loosely
to the command line flags that you may be familiar with through Docker Machine.

Currently `digitalocean` is the only provider which is supported so it's implied
that `driveroptions` refers to that.

Eventually all drivers, and hopefully a variety of different resource types,
will be supported, and the YAML keys will be cleaned up.  I'd like to integrate
it into Docker Machine proper or a related project.

When you have your settings done, just:

```
$ ./moby up
```

And you will see the pretty output as the machine gets created, Docker gets
installed and configured, etc.  The machine store will appear as a local
directory.  You can interact with the created machine (`mobydick`) with the
normal `docker-machine` client by specifying that as store path, like so:

```console
$ docker-machine -s store ls
```

Now, here is the fun bit:  What if we decided after the fact that we want to
change a Docker daemon options, like set `--log-driver=syslog` instead of
`json-file`?  `moby` can do that without you having to re-create your whole
instance.

You can change the YAML:

```
hostoptions:
  engineoptions:
    storagedriver: aufs
    labels:
      - foo=bar
      - baz=quux
      - far=go
    arbitraryflags:
      - dns=8.8.8.8
      - log-driver=syslog
driveroptions:
  accesstoken: <token>
  region: sfo1
  size: 2gb
```

Then run, `./moby apply`.  It will go out and change the Docker daemon settings
for you!

Oh yeah, I forgot to mention, but `swarmoptions` is a supported key too :D

## caveat emptor

This is very much just a toy and not meant for any real kinds of usage, only
hacking fun.  But, someday the foundations presented here as POC I hope will
make it upstream and become a very, very useful tool for folks doing Real Stuff.
For instance, I think the Virtualbox configuration / usage experience could be
way better with Machine than it is today.  I want to use such a YAML to define
all of your Compose files that you want to run automatically, different types of
shares so the VirtualBox shared folders can be replaced and/or manipulated more
effectively, and so on.  Then, you can pass around a little `moby.yml` (or
equivalent) to your team and the workflow is much easier, like Vagrant's.

## have fun and let me know what you think!
