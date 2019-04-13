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