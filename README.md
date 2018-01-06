# varnish-broadcast

[![Build Status](https://travis-ci.org/emgag/varnish-broadcast.svg?branch=master)](https://travis-ci.org/emgag/varnish-broadcast)
[![Go Report Card](https://goreportcard.com/badge/github.com/emgag/varnish-broadcast)](https://goreportcard.com/report/github.com/emgag/varnish-broadcast)

**WORK IN PROGRESS**: more or less feature complete, but not used in production yet.

**varnish-broadcast** is designed to distribute cache invalidation requests to a fleet of
[varnish](http://varnish-cache.org/) instances running in an dynamic environment (e.g. AWS Auto Scaling, Azure
Autoscale). The agent daemon is listening for PURGE and BAN requests on a [Redis
Pub/Sub](https://redis.io/topics/pubsub) channel and forwards incoming cache invalidation requests to its local varnish
instance. It's the successor of [varnish-cache-reaper](https://github.com/emgag/varnish-cache-reaper), which is also
used to fan out invalidation requests to multiple varnish instances, though its host list is static while with
varnish-broadcast, each varnish instance registers itself automatically.

It supports PURGE and BAN requests as well as surrogate keys (cache tags) using the 
[xkey module](https://github.com/varnish/varnish-modules/blob/master/docs/vmod_xkey.rst), formerly known as Hashtwo.

## Requirements

* [Redis](https://redis.io) service, accessible from both the applications
    issuing invalidation requests as well as the varnish instances
    running the agent.
* [Varnish](http://varnish-cache.org/), obviously. Although as it doesn't use any specific varnish APIs and uses plain 
    HTTP, it can probably be configured for other proxies as well. 
* VCL has to be modified to support purging, banning and distinguishing the two different xkey purging methods 
    supported by varnish-broadcast. See VCL example below.

## Agent

### Configuration

The agent configuration is done using a YAML file (see [varnish-broadcast.yml.dist]([varnish-broadcast.yml.dist])), default location is
*/etc/varnish-broadcast.yml*.

**redis** section:

* uri: redis host to connect to. Use redis:// for unencrypted, rediss:// for an encrypted connection.
* password: provide password if the connection needs to be authenticated.
* subscribe: list of pubsub channels the agent will subscribe to. 

**endpoint** section:

* uri: the HTTP endpoint of the varnish instance. 
* xkeyheader: The header used to supply list of keys to purge using *xkey.purge()* 
* softxkeyheader: The header used to supply list of keys to purge using *xkey.softpurge()*
* banheader: The header used to supply the expression for *ban()*
* banurlheader: The header used to supply the pattern for an URL ban 
 
Example:

```YAML
redis:
  uri: redis://127.0.0.1:6379
  password: thepasswordifneeeded
  subscribe:
    - varnish.purge
endpoint: 
  uri: http://127.0.0.1:8080/
  xkeyheader: x-xkey
  softxkeyheader: x-xkey-soft
  banheader: x-ban-expression
  banurlheader: x-ban-url  
```

### Usage

```
Distribute cache invalidation requests to a fleet of varnish instances.

Usage:
  varnish-broadcast [command]

Available Commands:
  ban         Issue ban request to all registered instances
  help        Help about any command
  listen      Listen for incoming invalidation requests
  purge       Issue purge request to all registered instances
  version     Print the version number of varnish-broadcast
  xkey        Invalidate selected surrogate keys on all registered instances

Flags:
  -c, --config string   config file (default is /etc/varnish-broadcast.yml)
  -h, --help            help for varnish-broadcast

Use "varnish-broadcast [command] --help" for more information about a command.
```

Example:
```
$ varnish-broadcast -c varnish-broadcast.yml listen 
2017/12/14 01:09:14 Connecting to redis...
2017/12/14 01:09:14 Connected to redis://127.0.0.1:6379
2017/12/14 01:09:14 subscribe: varnish.purge (1)
[...]
```

## Invalidation request API

Invalidation requests can be sent by publishing to a [Redis Pub/Sub](https://redis.io/topics/pubsub) channel.

### Message format

The publish message payload consists of a JSON object with following properties:

* **command**: string. Required. Either _ban_, _ban.url_, _purge_, _xkey_ or _xkey.soft_.
* **host**: string. Required. The _Host_ header used in the PURGE/BAN request to varnish.
* **value**: string[]. Required. Meaning depends on the command.
_ban_: List of [ban() expressions](https://varnish-cache.org/docs/5.2/reference/vcl.html#vcl-7-ban), 
_ban.url_: List of regular expressions matching the path portion of the URL to be banned,
_purge_: The path portion of the URL to be purged,
_xkey_ and _xkey.soft_: List of keys to (soft-)purge.

Example:

```JSON
{
   "command" : "xkey",
   "host" : "www.example.org",
   "value" : ["still", "flying"]
}
```

Using _varnish-broadcast_:

```
$ varnish-broadcast -c config.yml xkey --host www.example.org still flying
```

Using _redis-cli_:

```
$ redis-cli
127.0.0.1:6379> publish varnish.purge '{"command": "xkey", "host": "www.example.org", "value": ["still", "flying"]}'
```

Using PHP & [Predis](https://github.com/nrk/predis):

```PHP
$client = new Predis\Client([
    'scheme' => 'tcp',
    'host'   => '127.0.0.1',
    'port'   => '6379'
]);

$message = json_encode([
    'command' => 'xkey',
    'host'    => 'www.example.org',
    'value'   => ['still', 'flying']
]);

$client->publish('varnish.purge', $message);
```

Using PHP, [Predis](https://github.com/nrk/predis) & [varnish-broadcast-php](https://github.com/emgag/varnish-broadcast-php):

```PHP
$client = new Predis\Client([
    'scheme' => 'tcp',
    'host'   => '127.0.0.1',
    'port'   => '6379'
]);

$vb = new VarnishBroadcast($client);
$vb->xkey('example.org', ['still', 'flying']);
```

## VCL example

### Varnish 4.1 / 5.x

Varnish documentation on [purging and banning in varnish 4](https://www.varnish-cache.org/docs/4.1/users-guide/purging.html), or in [varnish 5.2](https://www.varnish-cache.org/docs/5.2/users-guide/purging.html) 

```VCL
[...]
# use xkey module
import xkey;

# purgers acl
# - who is allowed to issue PURGE and BAN requests
# 
acl purgers {
    "localhost";
}

sub vcl_recv {
    [...]
    if (req.method == "PURGE") {
        if (!client.ip ~ purgers) {
            return(synth(405,"Method not allowed"));
        }
        
        if(req.http.x-xkey) {
            set req.http.n-gone = xkey.purge(req.http.x-xkey);
            return (synth(200, "Got " + req.http.x-xkey + ", invalidated " + req.http.n-gone + " objects"));
        }
        
        if(req.http.x-xkey-soft) {
            set req.http.n-gone = xkey.softpurge(req.http.x-xkey-soft);
            return (synth(200, "Got " + req.http.x-xkey-soft + ", invalidated " + req.http.n-gone + " objects"));
        }
        
        return (purge);
    }

    if (req.method == "BAN") {
        if (!client.ip ~ purgers) {
            return(synth(405,"Method not allowed"));
        }
        
        if (req.http.x-ban-expression) {
            ban(req.http.x-ban-expression)
            return(synth(200, "Banned expression"));
        
        } else if (req.http.x-ban-url) {
            ban(
                "obj.http.host == " + req.http.host + " && " +
                "obj.http.url ~ " + req.http.x-ban-url
            )
            
            return(synth(200, "Banned URL"));
        }
        
        return(synth(400, "No bans"));        
    }

    [...]

    return(hash);
}

sub vcl_backend_response {
    [...]

    # be friendly to ban lurker
    set beresp.http.url = bereq.url;
    
    [...]
}

sub vcl_deliver {
    [...]
    
    # remove some variables we used before
    unset resp.http.url;
    unset resp.http.xkey;
    
    [...]
}
```

## Build

On Linux:

```
$ mkdir varnish-broadcast && cd varnish-broadcast
$ export GOPATH=$PWD
$ go get -d github.com/emgag/varnish-broadcast
$ cd src/github.com/emgag/varnish-broadcast
$ make install
```

will download the source and builds binary called _varnish-broadcast_ in $GOPATH/bin.

## License

varnish-broadcast is licensed under the [MIT License](http://opensource.org/licenses/MIT).
