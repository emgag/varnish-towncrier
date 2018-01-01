# varnish-broadcast

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

The agent configuration is done using a YAML file (see [config.yml.dist]([config.yml.dist])), default location is
*/etc/varnish-broadcast.yml*.

**redis** section:

* uri: redis host to connect to. Use redis:// for unencrypted, rediss:// for an encrypted connection.
* password: provide password if the connection needs to be authenticated.
* subscribe: list of pubsub channels the agent will subscribe to. 

**endpoint** section:

* uri: the HTTP endpoint of the varnish instance. 
* xkeyheader: The header used to supply list of keys to purge using *xkey.purge()* 
* softxkeyheader: The header used to supply list of keys to purge using *xkey.softpurge()*
 
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
```

### Usage

```
NAME:
   varnish-broadcast - Distribute cache invalidation requests to a fleet of varnish instances.

USAGE:
   varnish-broadcast [global options] command [command options] [arguments...]

VERSION:
   0.1

COMMANDS:
     help, h  Shows a list of commands or help for one command
   Agent:
     listen  Listen for incoming invalidation requests
   Client:
     ban    Issue ban request to all registered instances
     purge  Issue purge request to all registered instances
     xkey   Invalidate selected surrogate keys on all registered instances

GLOBAL OPTIONS:
   --config FILE, -c FILE  Load configuration from FILE (default: "/etc/varnish-broadcast.yml")
   --help, -h              show help
   --version, -v           print the version
```

Example:
```
$ varnish-broadcast -c config.yml listen 
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
* **expression**: string. Required for _ban_. A [ban() expression](https://varnish-cache.org/docs/5.2/reference/vcl.html#vcl-7-ban).
* **host**: string. Required. The _Host_ header used in the PURGE/BAN request to varnish.
* **path**: string. Required for _purge_ command. The path portion of the URL to be purged.
* **pattern**: string. Required for _ban.url_. Regular expression matching the path portion of the URL to be banned.
* **keys**: string[]. Required for _xkey_ and _xkey.soft_ commands. A list of keys to purge. 

Example:

```JSON
{
   "command" : "xkey",
   "host" : "www.example.org",
   "keys" : ["still", "flying"]
}
```

Using _varnish-broadcast_:

```
$ varnish-broadcast -c config.yml xkey --host www.example.org still flying
```

Using _redis-cli_:

```
$ redis-cli
127.0.0.1:6379> publish varnish.purge '{"command": "xkey", "host": "www.example.org", "keys": ["still", "flying"]}'
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
    'keys'    => ['still', 'flying']
]);

$client->publish('varnish.purge', $message);
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
        } else {
            # remove leading /
            ban(
                "obj.http.host == " + req.http.host + " && " +
                "obj.http.url ~ " + regsub(req.url, "^/", "")
            )        
        }

        return(synth(200, "Banned"));
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

TBD

## License

varnish-broadcast is licensed under the [MIT License](http://opensource.org/licenses/MIT).
