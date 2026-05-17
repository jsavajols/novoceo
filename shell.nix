{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "zigbee-server";

  packages = with pkgs; [
    mosquitto
    go
  ];

  shellHook = ''
    export SERVER_DIR="$(pwd)"
    export DATA_DIR="$SERVER_DIR/run"

    mkdir -p "$DATA_DIR/mosquitto"

    # Génère mosquitto.conf si absent
    if [ ! -f "$DATA_DIR/mosquitto/mosquitto.conf" ]; then
      cat > "$DATA_DIR/mosquitto/mosquitto.conf" <<CONF
# Ecoute sur toutes les interfaces (réseau local)
listener 1883
allow_anonymous true
persistence true
persistence_location $DATA_DIR/mosquitto/
log_dest stdout
CONF
      echo "[server] mosquitto.conf généré"
    fi

    echo ""
    echo "  Zigbee Server - shell prêt"
    echo "  ─────────────────────────────────────────"
    echo "  Démarrer Mosquitto :  ./start.sh"
    echo "  Stopper :             ./stop.sh"
    echo "  Lancer le monitor :   go run main.go tcp://localhost:1883"
    echo "  Debug MQTT :          mosquitto_sub -h localhost -t 'zigbee2mqtt/#' -v"
    echo "  ─────────────────────────────────────────"
    echo ""
  '';
}
