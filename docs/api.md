# API REST

Service Go/Fiber exposant les données Zigbee2MQTT et le contrôle des devices.

URL en production : `https://novoceo.api.local.happyapi.fr`

## Authentification

Toutes les routes (sauf `/healthz`) requièrent un Bearer token :

```
Authorization: Bearer <API_TOKEN>
```

## Endpoints

### GET /healthz

Liveness probe — pas d'authentification requise.

```
200 OK
ok
```

### POST /device/command

Envoie une commande MQTT à un device.

**Body JSON :**

```json
{
  "device": "Prise",          // requis
  "rpi": "novoceo-os",        // optionnel — si vide, topic sans préfixe RPi
  "state": "TOGGLE",          // ON | OFF | TOGGLE (défaut: TOGGLE)
  "payload": "{\"brightness\":128}" // optionnel — JSON brut, prioritaire sur state
}
```

**Réponse :**

```json
{
  "ok": true,
  "topic": "zigbee2mqtt/novoceo-os/Prise/set",
  "payload": "{\"state\":\"TOGGLE\"}"
}
```

**Erreurs :**
- `400` : device manquant ou payload JSON invalide
- `503` : broker MQTT non disponible
- `504` : timeout broker (> 3s)

### GET /device/:device/contact

Dernier état d'un contacteur (capteur d'ouverture de porte).

```
GET /device/Contacteur%20porte/contact
```

```json
{
  "contact": true,
  "battery": 85,
  "created_at": "2026-05-16T10:23:45Z"
}
```

`contact: true` = porte fermée, `contact: false` = porte ouverte.

Retourne `404` si aucune donnée n'existe encore pour ce device.

Le friendly name du device dans Zigbee2MQTT doit correspondre au paramètre `:device` (ici `Contacteur porte`, encodé `%20` dans l'URL).

### GET /device/:device/state

Dernier état connu d'un device (lu en base).

```
GET /device/Prise/state
```

```json
{
  "state": "ON",
  "created_at": "2026-05-16T10:23:45Z"
}
```

### GET /sensor/temperature

Dernière mesure du capteur Température.

```json
{
  "battery": 85,
  "temperature": 21.5,
  "humidity": 58.0,
  "created_at": "2026-05-16T10:20:00Z"
}
```

### GET /sensor/temperature/history

Historique des températures sur les 24 dernières heures, agrégées par tranches de 15 minutes.

```json
[
  {"temperature": 19.8, "created_at": "2026-05-15T10:00:00Z"},
  {"temperature": 20.1, "created_at": "2026-05-15T10:15:00Z"},
  ...
]
```

Retourne un tableau vide `[]` si aucune donnée dans les dernières 24h.
Maximum 96 points (4 par heure × 24h).

### GET /bridge/health

Dernier état de santé du bridge Zigbee2MQTT (charge CPU, mémoire).

```json
{
  "load_average": [0.12, 0.08, 0.05],
  "memory_percent": 42.3,
  "memory_free": 57.7,
  "created_at": "2026-05-16T10:22:00Z"
}
```

## Variables d'environnement

| Variable | Défaut | Description |
|----------|--------|-------------|
| `MQTT_BROKER` | `tcp://192.168.1.128:1883` | URL du broker MQTT |
| `API_PORT` | `5000` | Port d'écoute |
| `API_TOKEN` | — | Bearer token (requis, sinon arrêt) |
| `DB_HOST` | `localhost` | Hôte PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_USER` | `postgres` | Utilisateur PostgreSQL |
| `DB_PASSWORD` | — | Mot de passe PostgreSQL |
| `DB_NAME` | `postgres` | Base de données |

En k3s : variables injectées via ConfigMap `api-config` + Secrets `recorder-secret` et `api-secret`.

## Lancement local

```bash
go run ./api/
```

Ou avec le fichier `.env` :

```bash
# .env chargé automatiquement via godotenv
go run ./api/
```
