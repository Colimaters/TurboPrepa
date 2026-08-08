# TurboPrepa

Application de bureau macOS, construite avec Wails et Go.

## Lancer l'application

Prérequis : Go, Node.js et le CLI Wails (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).

```bash
wails dev
```

Une fenêtre TurboPrepa s'ouvre avec une page de confirmation. Le bouton « Vérifier le lancement » confirme la communication entre l'interface et Go.

## Construire l'application macOS

À exécuter depuis un Mac : Wails ne prend pas en charge la compilation croisée vers macOS.

```bash
wails build -clean -o TurboPrepa
```

Le bundle de test est généré dans `build/bin/TurboPrepa.app`.
