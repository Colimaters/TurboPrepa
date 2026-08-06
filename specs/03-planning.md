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
  - Matière associée (liste déroulante depuis l'onglet Matières) → détermine la couleur par
    défaut.
  - Intitulé / chapitre concerné (optionnel, peut être lié à un chapitre précis de la
    matière).
  - Date, heure de début, heure de fin.
  - **Couleur** : panel de **15 couleurs pastel proposées**, sélectionnable indépendamment
    de la couleur par défaut de la matière si l'utilisateur veut la personnaliser.
  - Notes libres (optionnel).
- La tâche créée apparaît immédiatement dans le calendrier.

### 2. Importer un emploi du temps
- Permet d'importer un fichier (PDF, image, Excel/.xlsx, Word) contenant un emploi du temps
  existant, comme pièce jointe consultable (voir `10-import-fichiers.md` pour les limites :
  pas de lecture automatique du contenu en V1).
- Après import, l'interface guide l'utilisateur pour **recopier manuellement** les créneaux
  dans le calendrier (l'app n'extrait pas automatiquement les horaires du fichier importé).
- Le fichier importé reste accessible en pièce jointe de référence.

### 3. Planifier automatiquement ma semaine
Fonctionnalité de génération automatique de planning, **synchronisée avec l'onglet
Matières** :
- L'utilisateur sélectionne une ou plusieurs matières/chapitres (ceux au statut "à
  planifier" en priorité, mais tous doivent rester sélectionnables) dans l'onglet Matières
  et indique le temps à y consacrer.
- Un questionnaire s'affiche pour chaque sélection, avec les champs suivants :
  - Jour(s) en particulier souhaité(s) (ou "peu importe").
  - Temps total à consacrer à ce chapitre/cette matière.
  - Pauses souhaitées dans la journée (fréquence/durée).
  - Horaires de début et de fin de journée de travail.
  - Alternance ou non avec d'autres matières dans la même journée (oui/non, et lesquelles).
- Bouton **"Générer automatiquement l'emploi du temps"** : produit un jeu de blocs sur le
  calendrier respectant les contraintes saisies (pas de chevauchement, respect des pauses et
  des horaires de journée définis).
- Les blocs générés sont des tâches normales, éditables ensuite comme n'importe quelle
  tâche manuelle (voir sous-onglet 1).
- Le rythme par défaut suggéré si l'utilisateur ne précise rien doit rester raisonnable et
  reproductible (éviter de proposer des journées surchargées) — cadre indicatif à proposer
  en pré-remplissage du questionnaire :
  - Matin : 8h–12h
  - Après-midi : 14h–19h
  - Soirée : 20h–22h
  - Temps personnel (sport, repos, vie sociale) à laisser explicitement libre dans le
    planning généré, pas comblé automatiquement.

## Calendrier / agenda (commun aux 3 sous-onglets, affiché en bas ou à côté du sous-onglet actif)
- Trois vues : **mois**, **semaine**, **jour** (détail), bascule facile entre les trois.
- Démarre à **août 2026** (première date affichable/navigable).
- Chaque tâche affichée avec la couleur de sa matière (panel des 15 couleurs pastel, partagé
  avec Matières).
- **Glisser-déposer** : déplacer une tâche vers un autre jour/horaire sans toucher au reste
  du planning (pas de recalcul en cascade des autres tâches).
- Clic sur une tâche → édition (matière, horaires, couleur, notes) ou suppression.
- Vue jour : granularité fine (créneaux, ex. par tranche de 15–30 min).
- Vue semaine/mois : vue condensée, lisible, avec indication visuelle de charge (ex. jours
  très chargés vs. légers).

## Panel de couleurs (référence partagée)
- 15 couleurs pastel prédéfinies, utilisées à la fois :
  - comme couleur par défaut d'une matière (`02-matieres.md`),
  - comme couleur assignable à une tâche individuelle.
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
- [ ] Le panel de 15 couleurs est identique entre Matières et Planning.
