# varnish-broadcast

**WORK IN PROGRESS**: not ready yet.

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
*/etc/varnish/broadcast-agent.yml*.

**redis** section:

* uri: redis host to connect to. Use redis:// for unencrypted, rediss:// for an encrypted connection.
* password: provide password if the connection needs to be authenticated.
* subscribe: list of pubsub channels the agent will subscribe to. 

**endpoint** section:

* uri: the HTTP endpoint of the varnish instance. 
* xkeyheader: The header used to supply list of tags to purge using *xkey.purge()* 
* softxkeyheader: The header used to supply list of tags to purge using *xkey.softpurge()*
 
Example:

```YAML
redis:
  uri: redis://127.0.0.1:6379
  password: thepasswordifneeeded
  subscribe:
    - varnish.purge
endpoint: 
  uri: http://127.0.0.1:8080/
  xkeyheader: xkey
  softxkeyheader: xkey-soft
```

### Usage

The agent has just one command, *listen*:

```
$ varnish-broadcast-agent listen -h  
Usage of listen:
  -config string
        Config file to use. (default "/etc/varnish/broadcast-agent.yml")
```

Example:
```
$ varnish-broadcast-agent listen -config config.yml
2017/12/14 01:09:14 Connecting to redis...
2017/12/14 01:09:14 Connected to redis://127.0.0.1:6379
2017/12/14 01:09:14 subscribe: varnish.purge (1)
[...]
```


## Invalidation request API

Invalidation requests can be sent by publishing to a [Redis Pub/Sub](https://redis.io/topics/pubsub) channel.

### Message format

The publish message payload consists of JSON object with following properties:

* **command**: string. Required. Either _ban_, _purge_, _xkey_ or _softxkey_.
* **host**: string. Required. The _Host_ header used in the PURGE/BAN request to varnish.
* **path**: string. Required for _purge_ command. The path portion of the URL to be purged.
* **pattern**: string. Required for _ban_ command. The pattern used for banning.
* **tags**: string[]. Required for _xkey_ and _softxkey_ commands. A list of tags to purge. 
...

Example:

```JSON
{
   "command" : "xkey",
   "host" : "www.example.org",
   "tags" : ["still", "flying"]
}
```

Using _redis-cli_:

```
$ redis-cli
127.0.0.1:6379> publish varnish.purge '{"command": "xkey", "host": "www.example.org", "tags": ["still", "flying"]}'
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
    'tags'    => ['still', 'flying']
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
        
        if(req.http.xkey) {
            set req.http.n-gone = xkey.purge(req.http.xkey);
            return (synth(200, "Got " + req.http.xkey + ", invalidated " + req.http.n-gone + " objects"));
        }
        
        if(req.http.xkey-soft) {
            set req.http.n-gone = xkey.softpurge(req.http.xkey-soft);
            return (synth(200, "Got " + req.http.xkey-soft + ", invalidated " + req.http.n-gone + " objects"));
        }
        
        return (purge);
    }

    if (req.method == "BAN") {
        if (!client.ip ~ purgers) {
            return(synth(405,"Method not allowed"));
        }
    
        # remove leading / to not confuse regular expression
        ban(
            "obj.http.x-host == " + req.http.host + " && " +
            "obj.http.x-url ~ " + regsub(req.url, "^/", "")
        );

        return(synth(200, "Banned"));
    }

    [...]

    return(hash);
}

sub vcl_backend_response {
    [...]

    # be friendly to ban lurker
    set beresp.http.x-url = bereq.url;
    set beresp.http.x-host = bereq.http.host;
    
    [...]
}

sub vcl_deliver {
    [...]
    
    # remove some variables we used before
    unset resp.http.x-url;
    unset resp.http.x-host;
    unset resp.http.xkey;
    
    [...]
}
```

## Build

TBD

## License

varnish-broadcast is licensed under the [MIT License](http://opensource.org/licenses/MIT).
