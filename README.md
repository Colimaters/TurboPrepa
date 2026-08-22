# TurboPrepa

Application de bureau macOS, construite avec Wails et Go.

## Structure du projet

```text
internal/application/         Cycle de vie Wails, SQLite, migrations et domaines Go actuels
internal/application/assets/  Données initiales embarquées (citations)
frontend/src/app/             Point d'entrée et orchestration de l'interface
frontend/src/domains/         Vues et modules par domaine métier
frontend/src/shared/          Composants et comportements réutilisables
frontend/src/styles/          Styles globaux et composants communs
specs/                        Spécifications et critères d'acceptation
```

Le code déjà livré couvre Accueil et Matières. Les prochains domaines suivent la même
frontière dans `frontend/src/domains/` et dans `internal/application/` : Planning,
Annuaire, Jurisprudence, Veille juridique, Quiz du jour, Concours et Textes de référence.
Les éléments transverses, notamment les migrations SQLite, la palette et l'import de
fichiers, restent uniques afin d'éviter toute duplication entre domaines.

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
