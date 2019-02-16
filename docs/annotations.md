| Attack | Annotation                                    | Value       |
|--------|-----------------------------------------------|------------ |
| CPU    | gremlin.gremlin.kubedex.com/attack            | [cpu]       |
|        | gremlin.gremlin.kubedex.com/length            | int         |
|        | gremlin.gremlin.kubedex.com/cores             | int         |
|        |                                               |             |
| Disk   | gremlin.gremlin.kubedex.com/attack            | [disk]      |
|        | gremlin.gremlin.kubedex.com/length            | int         |
|        | gremlin.gremlin.kubedex.com/directory         | string      |
|        | gremlin.gremlin.kubedex.com/volume_percentage | int         |
|        |                                               |             |
| I/O    | gremlin.gremlin.kubedex.com/attack            | [io]        |
|        | gremlin.gremlin.kubedex.com/length            | int         |
|        | gremlin.gremlin.kubedex.com/directory         | string      |
|        | gremlin.gremlin.kubedex.com/mode              | string [rwx]|
|        |                                               |             |
| Memory | gremlin.gremlin.kubedex.com/attack            | [memory]    |
|        | gremlin.gremlin.kubedex.com/allocation_type   | [gb|mb|%]   |
|        |                                               |             |


`metadata.annotations.gremlin.gremlin.kubedex.com/team_id`

`metadata.annotations.gremlin.gremlin.kubedex.com/container`