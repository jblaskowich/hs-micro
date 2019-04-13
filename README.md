# hs-micro-front

## Overview

This is a educational purpose app: a simple bloging like platform.

## Architecture

```
             ---------
             | EMAIL |
             ---------
                 ^
                 |
---------     --------     --------     -----------
| FRONT | --> | NATS | --> | BACK | --> | MariaDB |
---------     --------     --------     -----------
```

 - Front: a go frontend (gorilla, html/templatesn go-nats)
 - Back: a go backend (go-nats, database/sql)
 - Email: a python notification service

## Run

### Binary

```
$ go build -o app
$ export NATSURL="your_nats_url"               // default demo.nats.io
$ export NATSPORT="your_nats_port"             // default :4222
$ export NATSPOST="your_nats_post_channel"     // the channel used for posts, default zjnO12CgNkHD0IsuGd89zA
$ export NATSGET="your_nats_get_posts_channel" // the channel used get posts, default OWM7pKQNbXd7l75l21kOzA
$ ./app
```

### Docker

```
$ export NATSURL="your_nats_url"               // default demo.nats.io
$ export NATSPORT="your_nats_port"             // default :4222
$ export NATSPOST="your_nats_post_channel"     // the channel used for posts, default zjnO12CgNkHD0IsuGd89zA
$ export NATSGET="your_nats_get_posts_channel" // the channel used get posts, default OWM7pKQNbXd7l75l21kOzA
$ docker run -d -e NATSURL=${NATSURL} -e NATSPORT=${NATSPORT} -e NATSPOST=${NATSPOST} -e NATSGET=${NATSGET} -p 8080:8080 jblaskovich/hs-micro-front:$release
```