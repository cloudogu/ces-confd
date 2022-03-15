# Einbauen ins nginx zum testen

1. Bauen des Binaries mit `make`
2. Kopieren des BInaries in den nginx-Container mit `docker cp ces-confd nginx:/usr/bin/`
3. Ggf. Anpassen der configuration.yml im nginx. Diese liegt im container unter `/etc/ces-confd/config.yaml.tpl`
4. `docker restart nginx && tail -f /var/log/docker/nginx.log` => Sollte ebenfalls die ausgaben von ces-confd zeigen
