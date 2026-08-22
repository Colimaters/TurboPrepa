# 03 — Planning

Dépend de : `00-architecture-generale-et-design.md`, `02-matieres.md`,
`10-import-fichiers.md`

## Objectif
Organiser le temps de révision : poser un cadre régulier et soutenable, visualiser le
travail sur le calendrier, et pouvoir générer automatiquement un planning à partir du
programme des matières.

## Structure : 3 sous-onglets

### 1. Ajouter une tâche
- Formulaire de création manuelle d'une tâche :
  - Matière associée obligatoire (liste déroulante depuis l'onglet Matières) → détermine la
    couleur de la tâche.
  - Intitulé / chapitre concerné (optionnel, peut être lié à un chapitre précis de la
    matière).
  - Date, heure de début, heure de fin.
  - Notes libres (optionnel).
- La tâche créée apparaît immédiatement dans le calendrier.

### 2. Importer un emploi du temps
- Permet d'importer un fichier (Excel/.xlsx) contenant un emploi du temps existant.
- Possibilité de télécharger d'abord un template afin de faciliter l'import.
- Après import, l'interface les créneaux qui ont été importés dans le calendrier.

### 3. Planifier automatiquement ma semaine
Fonctionnalité de génération automatique de planning, **synchronisée avec l'onglet
Matières** :
- L'utilisateur sélectionne une ou plusieurs matières/chapitres (ceux au statut "à
  planifier" en priorité, mais tous doivent rester sélectionnables) dans l'onglet Matières
  et indique le temps à y consacrer.
- Un questionnaire s'affiche pour chaque sélection, avec les champs suivants :
  - Jour de début en souhaité(s).
  - Nombre de révisions souhaité pour ce chapitre.
  - Durée de la révision pour ce chapitre
  - Espacement des révisions pour ce chapitre (en nombre de jour)s
- Bouton **"Générer automatiquement l'emploi du temps"** : produit un jeu de blocs sur le
  calendrier respectant les contraintes saisies (pas de chevauchement et respect
  des horaires de journée définis).
- Les blocs générés sont des tâches normales, éditables ensuite comme n'importe quelle
  tâche manuelle (voir sous-onglet 1).
- Pouvoir éditer sa journée de travail type qui servira pour placer les activités généré automatiquement 
  - Matin : 8h–12h
  - Après-midi : 14h–19h
  - Soirée : 20h–22h


## Calendrier / agenda (commun aux 3 sous-onglets, affiché en bas ou à côté du sous-onglet actif)
- Trois vues : **mois**, **semaine**, **jour** (détail), bascule facile entre les trois.
- Démarre à **août 2026** (première date affichable/navigable).
- Chaque tâche affichée avec la couleur de sa matière (panel des 15 couleurs pastel, partagé
  avec Matières), sans personnalisation individuelle.
- **Glisser-déposer** : déplacer une tâche vers un autre jour/horaire sans toucher au reste
  du planning (pas de recalcul en cascade des autres tâches).
- Clic sur une tâche → édition (matière, horaires, notes) ou suppression.
- Vue jour : granularité fine (créneaux, ex. par tranche de 15–30 min).
- Vue semaine/mois : vue condensée, lisible, avec indication visuelle de charge (ex. jours
  très chargés vs. légers).

## Panel de couleurs (référence partagée)
- 15 couleurs pastel prédéfinies, utilisées à la fois :
  - comme couleur par défaut d'une matière (`02-matieres.md`),
  - comme couleur d'une tâche rattachée à cette matière.
- Stocker ce panel comme une liste centrale réutilisable (pas dupliquée entre Planning et
  Matières) pour garantir la cohérence visuelle.

## Critères d'acceptation
- [ ] Les 3 sous-onglets (Ajouter une tâche / Importer un emploi du temps / Planifier
      automatiquement) sont bien distincts et accessibles.
- [ ] Le calendrier propose les 3 vues (mois/semaine/jour) et démarre en août 2026.
- [ ] Une tâche créée manuellement apparaît immédiatement dans le calendrier avec la bonne
      couleur.
- [ ] Le glisser-déposer déplace une tâche sans modifier les autres tâches du planning.
- [ ] La planification automatique lit bien les chapitres/matières existants (pas de
      re-saisie) et applique les contraintes du questionnaire (jours, durée, pauses,
      horaires, alternance).
- [ ] Le planning généré automatiquement ne remplit pas les plages explicitement laissées
      comme temps personnel.
- [ ] Les tâches utilisent exclusivement la couleur de leur matière.
