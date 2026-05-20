# Architecture et topologie MQTT

## Diagramme d'architecture

```mermaid
graph TB
    subgraph rpi["RPi novoceo-os"]
        direction TB
        DONGLE["🔌 Dongle Zigbee USB"]
        Z2M["zigbee2mqtt\nbase_topic: zigbee2mqtt/novoceo-os"]
        WDG["watchdog-net\npublish rpi/watchdog-net"]
        DONGLE --> Z2M

        subgraph devices["Devices Zigbee"]
            BOUTON["🔘 Bouton\n(interrupteur)"]
            PRISE["⚡ Prise\n(connectée)"]
            TEMP["🌡 Température\n(capteur)"]
        end

        BOUTON -- "Zigbee RF" --> Z2M
        TEMP   -- "Zigbee RF" --> Z2M
        Z2M    -- "Zigbee RF" --> PRISE
    end

    subgraph k3s["k3s cluster — 192.168.1.128  |  namespace novoceo"]
        direction TB

        MOSQ["🦟 mosquitto\nNodePort :32500"]

        subgraph consumers["Consommateurs MQTT"]
            SNUG["snug\nBouton → TOGGLE Prise"]
            REC["recorder\nzigbee2mqtt/# + rpi/# → PostgreSQL"]
        end

        subgraph web["Couche web"]
            API["api\n:5000  REST + Bearer auth"]
            FRONT["front\n:3000  HTMX dashboard"]
        end

        MOSQ --> SNUG
        MOSQ --> REC
        API  --> MOSQ
        REC  --> DB
        API  --> DB
        FRONT --> API
    end

    subgraph ext["Externe"]
        DB[("PostgreSQL\nScaleway\n51.159.109.144:30015")]
        BROWSER["🌐 Navigateur"]
        INGRESS["Ingress nginx\nTLS Let's Encrypt"]
    end

    Z2M  -- "MQTT TCP :32500" --> MOSQ
    WDG  -- "MQTT TCP :32500" --> MOSQ
    SNUG -- "MQTT TOGGLE /set" --> MOSQ

    BROWSER --> INGRESS
    INGRESS -- "novoceo.front.local.happyapi.fr" --> FRONT
    INGRESS -- "novoceo.api.local.happyapi.fr"   --> API

    style rpi      fill:#1e293b,stroke:#475569,color:#cbd5e1
    style k3s      fill:#0f172a,stroke:#334155,color:#94a3b8
    style ext      fill:#0f172a,stroke:#334155,color:#94a3b8
    style devices  fill:#1e293b,stroke:#334155,color:#94a3b8
    style consumers fill:#1e293b,stroke:#334155,color:#94a3b8
    style web      fill:#1e293b,stroke:#334155,color:#94a3b8
    style MOSQ     fill:#0e7490,stroke:#06b6d4,color:#fff
    style Z2M      fill:#7c3aed,stroke:#8b5cf6,color:#fff
    style SNUG     fill:#065f46,stroke:#10b981,color:#fff
    style REC      fill:#92400e,stroke:#f59e0b,color:#fff
    style API      fill:#1e40af,stroke:#3b82f6,color:#fff
    style FRONT    fill:#1e40af,stroke:#3b82f6,color:#fff
    style DB       fill:#581c87,stroke:#a855f7,color:#fff
    style DONGLE   fill:#374151,stroke:#6b7280,color:#d1d5db
    style BOUTON   fill:#374151,stroke:#6b7280,color:#d1d5db
    style PRISE    fill:#374151,stroke:#6b7280,color:#d1d5db
    style TEMP     fill:#374151,stroke:#6b7280,color:#d1d5db
    style BROWSER  fill:#374151,stroke:#6b7280,color:#d1d5db
    style INGRESS  fill:#374151,stroke:#6b7280,color:#d1d5db
```

## Topologie réseau (ASCII)

## Topics MQTT

Zigbee2MQTT du RPi publie sur le broker central avec son hostname comme préfixe :

```
zigbee2mqtt/<rpi>/<device>        # état d'un device (ex: zigbee2mqtt/novoceo-os/Bouton)
zigbee2mqtt/<rpi>/<device>/set    # commande vers un device
zigbee2mqtt/<rpi>/bridge/devices  # liste des devices Zigbee du RPi
zigbee2mqtt/<rpi>/bridge/health   # état de santé du bridge
zigbee2mqtt/bridge/health         # santé du bridge local (si zigbee2mqtt local)
```

Le watchdog réseau publie sur un topic distinct :

```
rpi/watchdog-net                  # heartbeat toutes les minutes + event reboot
```

Payload heartbeat : `{"event":"heartbeat","loss":N}` (N = % paquets perdus)
Payload reboot : `{"event":"reboot","loss":N}`

### Appuyer sur le bouton — flux complet

```mermaid
sequenceDiagram
    participant B  as 🔘 Bouton (Zigbee)
    participant Z  as zigbee2mqtt (RPi)
    participant M  as mosquitto (broker)
    participant S  as snug
    participant R  as recorder
    participant P  as ⚡ Prise (Zigbee)
    participant DB as PostgreSQL

    B  ->> Z  : Appui (RF Zigbee)
    Z  ->> M  : PUBLISH zigbee2mqtt/novoceo-os/Bouton<br/>{"action":"single"}
    M  -->> S  : (abonné) message reçu
    M  -->> R  : (abonné zigbee2mqtt/#) message reçu
    R  ->> DB : INSERT INTO states (topic, device_state)

    S  ->> M  : PUBLISH zigbee2mqtt/novoceo-os/Prise/set<br/>{"state":"TOGGLE"}
    M  ->> Z  : livraison commande
    Z  ->> P  : Commande Zigbee TOGGLE
    P  ->> Z  : Ack état (ON/OFF)
    Z  ->> M  : PUBLISH zigbee2mqtt/novoceo-os/Prise<br/>{"state":"ON"}
    M  -->> R  : message reçu
    R  ->> DB : INSERT INTO states (topic, device_state)
```

## Configuration zigbee2mqtt sur le RPi

Le RPi `novoceo-os` doit avoir dans `configuration.yaml` de zigbee2mqtt :

```yaml
mqtt:
  server: mqtt://192.168.1.128:32500   # broker central sur k3s
  base_topic: zigbee2mqtt/novoceo-os   # préfixe = hostname du RPi
```

Zigbee2MQTT tourne en tant que service systemd sur le RPi (`rpi/zigbee2mqtt.service`).

```bash
# Copier le service sur le RPi
sudo cp rpi/zigbee2mqtt.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now zigbee2mqtt

# Opérations courantes
sudo systemctl status zigbee2mqtt
sudo systemctl restart zigbee2mqtt
journalctl -u zigbee2mqtt -f
```

> Adapter `User=jerome` dans le fichier service si l'utilisateur système est différent.

## Noms des devices

Les friendly names sont définis dans zigbee2mqtt et utilisés partout :

| Friendly name | Type | Usage |
|---------------|------|-------|
| `Bouton` | Interrupteur Zigbee | Déclencheur snug |
| `Prise` | Prise connectée Zigbee | Cible du TOGGLE |
| `Température` | Capteur temp/humidity | Affiché dans front/api |

## Règles de routage snug

snug souscrit à deux topics pour gérer les deux formats :

```
zigbee2mqtt/Bouton          # format local (zigbee2mqtt sans préfixe RPi)
zigbee2mqtt/+/Bouton        # format multi-RPi (wildcard sur le nom du RPi)
```

Flag `-action single` : filtre sur le champ `action` du payload pour éviter
le double-toggle sur les appuis longs ou double-clic.
