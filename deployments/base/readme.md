##### How deployment works?

`conf`, `npc`... contains configuration files that we are actually want to run on our server and they are mapped into configMaps. To apply our files we do:

1. copy pure `conf`, `npc` from rathena to new volumes
2. copy our configsMaps on top
3. replace `$PASSWORD` varaibles in `conf/import` directory with secret values
4. mount result back to rathena container in place of source files and run it


yes it means that any config/script change will require server redeploy/restart.