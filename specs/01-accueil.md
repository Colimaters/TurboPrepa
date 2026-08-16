# 01 — Accueil

Dépend de : `00-architecture-generale-et-design.md`

## Objectif

Page d'atterrissage de l'application. Doit donner en un coup d'œil : de quoi tenir le moral,
et où en est l'utilisateur dans sa préparation.

## Contenu de la page

### 1. Carte "Pensée du jour"

- Affiche une **citation réelle et attribuée** (auteur, philosophe, figure historique,
  militaire ou politique) (ce sont des thèmes cohérents avec la
  préparation d'un concours : courage, discipline, droit, justice, persévérance, service
  public)
- Format d'affichage : citation entre guillemets, auteur en dessous (attribution visible).
- Si la paternité exacte d'une citation est débattue, préfixer par "Attribué à" plutôt que
  d'affirmer une attribution incertaine.
- Une citation différente chaque jour, non répétée avant d'avoir épuisé le stock disponible
  (rotation déterministe par date, pas aléatoire pure, pour éviter les répétitions
  rapprochées).

### 2. Résumé de progression

- Vue synthétique de l'avancement global des matières : nombre de chapitres par statut
  (à planifier / planifié / en cours / maîtrisé), toutes matières confondues.
- Barre ou indicateur visuel (ex. barre de progression, anneau) plutôt qu'un simple tableau
  de chiffres.
- Doit se mettre à jour automatiquement dès qu'un statut de chapitre change dans l'onglet
  Matières (lecture des mêmes données, pas de duplication de saisie).

### 3. Tâches du jour

- Liste des tâches planifiées pour la date du jour, issues de l'onglet Planning (lecture
  directe des données du planning, voir `03-planning.md`).
- Chaque tâche affiche : matière (avec sa couleur), intitulé, créneau horaire.
- Possibilité de cocher une tâche comme faite directement depuis l'accueil (répercuté dans
  le planning).
- Si aucune tâche n'est planifiée aujourd'hui : message encourageant + raccourci pour aller
  planifier (lien vers Planning).

## Interactions

- Clic sur une tâche du jour → navigue vers le jour correspondant dans Planning.
- Clic sur le résumé de progression → navigue vers Matières.

## Critères d'acceptation

- [ ] La citation affichée change chaque jour et n'est jamais générée à la volée par un
      modèle : elle vient d'une liste stockée avec attribution.
- [ ] Le résumé de progression reflète en temps réel les statuts réels des chapitres.
- [ ] Les tâches du jour correspondent exactement aux tâches planifiées à la date du jour
      dans l'onglet Planning, sans duplication de données.
- [ ] La page reste utilisable et lisible même sans aucune donnée saisie (état vide géré
      proprement, pas d'écran cassé).
