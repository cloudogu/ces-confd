# ces-confd testen

Der dogu nginx muss in der CES-Instanz laufen.

1. Bauen des Binaries mit `make`
2. Kopieren des Binaries in das (vagrant) ecosystem root Verzeichnis.
3. Im vagrant in den Pfad `/vagrant` navigieren 
4. den nginx-Container mit `docker cp ces-confd nginx:/usr/bin/`
5. Ggf. Anpassen der configuration.yml im nginx. Dafür mit `docker exec -it nginx sh`in den nginx Container gehen und die Konfiguration anpassen. Diese liegt im Container unter `/etc/ces-confd/config.yaml.tpl`. ***Achtung*** Die `config.yaml` in diesem Repository ist für Entwicklungszwecke und als Beispiel da. In der Produktion wird die Konfiguration durch den nginx dogu erstellt.
6. `docker restart nginx && tail -f /var/log/docker/nginx.log` => Sollte ebenfalls die Ausgaben von ces-confd zeigen
