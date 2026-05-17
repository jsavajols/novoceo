# Application mobile Android

App Android native construite avec React Native et Expo 54. Elle reproduit le dashboard web (4 cartes IoT) et appelle directement l'API REST.

## Technologie

| Elément | Valeur |
|---------|--------|
| Framework | React Native 0.81 + Expo SDK 54 |
| Langage | TypeScript |
| Build cloud | EAS Build (Expo Application Services) |
| Package Android | `com.novoceo.mobile` |
| Répertoire | `mobile/` |

## Cartes (équivalent du front web)

| Carte | Couleur | Endpoint API | Refresh |
|-------|---------|--------------|---------|
| Bridge Health | Cyan | `GET /bridge/health` | 60s |
| Température | Amber | `GET /sensor/temperature` + `/history` | 60s |
| Prise | Emerald | `GET /device/Prise/state` | 5s |
| Porte | Violet | `GET /device/Contacteur%20porte/contact` | 5s |

Le bouton **Toggle** de la carte Prise appelle `POST /device/command` puis rafraichit l'état après 1,5s.

Un **pull-to-refresh** (tirer vers le bas) force le rechargement de toutes les cartes simultanément.

## Configuration

L'URL de l'API et le token Bearer sont des constantes en haut de `mobile/App.tsx` :

```typescript
const API_URL = 'https://novoceo.api.local.happyapi.fr';
const API_TOKEN = '...';
```

## Développement (Expo Go)

Prérequis : Node 20+, téléphone Android avec l'app **Expo Go** installée (Play Store).

```bash
cd mobile/
npm install        # première fois seulement

npx expo start     # démarre Metro Bundler
```

Scanne le QR code depuis Expo Go. Le téléphone doit être sur le même Wi-Fi que le PC.

> **Firewall NixOS** : ouvrir le port 8081 (Metro) si le téléphone ne parvient pas à se connecter.
> ```bash
> sudo iptables -I INPUT -p tcp --dport 8081 -j ACCEPT
> ```
> Ou de façon permanente dans `/etc/nixos/configuration.nix` :
> ```nix
> networking.firewall.allowedTCPPorts = [ 22 8000 8081 19000 ];
> ```

## Build APK (EAS Build)

EAS Build compile l'APK dans le cloud Expo — aucun Android SDK requis en local.

```bash
# Une seule fois : login et configuration
eas login
eas build:configure

# Build APK installable directement sur Android
eas build --platform android --profile preview

# Vérifier l'état des builds
eas build:list --platform android --limit 3

# Télécharger le dernier APK
eas build:download --platform android
```

Le build prend environ 5-10 minutes. A la fin, EAS fournit une URL de téléchargement directe.

### Profils EAS (`eas.json`)

| Profil | Format | Usage |
|--------|--------|-------|
| `preview` | APK | Installation directe sur le téléphone |
| `production` | AAB | Publication sur le Play Store |

## Installation sur Android

1. Télécharger l'APK (lien EAS ou `eas build:download`)
2. Transférer sur le téléphone (email, Nextcloud, Tailscale...)
3. Ouvrir le fichier APK sur le téléphone
4. Autoriser les **sources inconnues** si demandé
5. Installer

## Assets

| Fichier | Taille | Usage |
|---------|--------|-------|
| `assets/icon.png` | 1024x1024 | Icône de l'app |
| `assets/adaptive-icon.png` | 1024x1024 | Icône adaptive Android (fond #020617) |
| `assets/splash-icon.png` | 1080x1080 | Ecran de démarrage |
| `assets/favicon.png` | 48x48 | Favicon web |

Les assets source sont générés avec ImageMagick depuis des SVGs (design : fond `#020617`, lettre `n` en `#2563EB`, grille réseau IoT).
