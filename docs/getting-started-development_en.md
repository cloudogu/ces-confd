# test ces-confd

The dogu nginx must be running in the CES instance.

1. build the binary with `make`.
2. copy the binary to the (vagrant) ecosystem root directory.
3. navigate in the vagrant to the path `/vagrant
4. create the nginx container with `docker cp ces-confd nginx:/usr/bin/`.
5. if necessary adjust the configuration.yml in nginx. To do this, go into the nginx container with `docker exec -it nginx sh` and adjust the configuration. This is in the container under `/etc/ces-confd/config.yaml.tpl`. 6.

**NOTE**
The `config.yaml` in this repository is there for development purposes and to have an example. In production the configuration is build by the nginx dogu. 

6. `docker restart nginx && tail -f /var/log/docker/nginx.log` => Should show the output of ces-confd as well.
