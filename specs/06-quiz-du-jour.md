# 06 — Quiz du jour

Dépend de : `00-architecture-generale-et-design.md`

## Objectif
Rendez-vous quotidien rapide avec le droit : muscler la culture juridique fondamentale sans
y consacrer trop de temps.

## Fonctionnement

### Questions
- **5 nouvelles questions chaque jour**, calibrées sur le niveau attendu au concours
  (référence donnée par l'utilisateur : niveau commissaire de police — à considérer comme le
  niveau d'exigence le plus élevé, donc pertinent aussi pour les autres concours visés).
- Questions à choix multiple ou vrai/faux (format permettant une correction immédiate et
  automatique).
- Une nouvelle série chaque jour, non répétée tant que le stock de questions n'est pas
  épuisé (rotation déterministe par date, comme pour les citations de l'Accueil).
- Le stock de questions doit être une donnée éditable/extensible (liste de
  `{ question, choix[], bonneRéponse, matière/thème, explication? }`) pour permettre d'en
  ajouter facilement dans le temps.

### Déroulé
- **Minuteur de 2 minutes top chrono** pour répondre aux 5 questions.
- **Correction immédiate** après validation (bonne/mauvaise réponse affichée pour chaque
  question, avec explication courte si disponible).
- Score du jour affiché en fin de série.

### Suivi personnel (remplace le classement multi-utilisateurs, hors périmètre — voir
`00-architecture-generale-et-design.md`)
- **Série de jours consécutifs** (streak) où le quiz du jour a été complété.
- **Score personnel** cumulé et/ou historique des scores par jour, consultable (ex. petit
  graphique ou liste des 30 derniers jours).
- Pas de comparaison avec d'autres utilisateurs réels dans cette version.

## Critères d'acceptation
- [ ] 5 questions différentes sont proposées chaque jour, avec rotation sans répétition
      prématurée.
- [ ] Le chronomètre de 2 minutes est visible et fonctionnel, avec verrouillage des réponses
      à expiration.
- [ ] La correction s'affiche immédiatement après la fin du quiz (ou à expiration du temps).
- [ ] Le streak de jours consécutifs se met à jour correctement (y compris remise à zéro en
      cas de jour manqué).
- [ ] L'historique des scores est consultable.
