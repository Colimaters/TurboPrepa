# 09 — Textes de référence

Dépend de : `00-architecture-generale-et-design.md`, `08-concours.md`

## Objectif
Recenser, pour chacun des 4 concours (Gardien de la paix, Officier de police, Commissaire de
police, Officier de gendarmerie), les textes réglementaires qui organisent leurs épreuves
(décrets, arrêtés organisant le concours et son contenu).

## Contenu
Pour chaque concours :
- Liste des textes de référence identifiés : nature du texte (décret, arrêté…), intitulé,
  date, et si possible un lien vers le texte sur Légifrance.
- Champ **statut de vérification** : "vérifié" / "à confirmer" — parce que ces textes
  évoluent avec les réformes, l'application doit afficher clairement ce qui a pu être
  confirmé par une recherche fiable au moment du développement, et laisser à l'utilisateur
  la responsabilité de confirmer/compléter via Légifrance, comme demandé.
- **Champ libre éditable par l'utilisateur** pour ajouter ou corriger un texte une fois
  vérifié de son côté (formulaire simple : nature, intitulé, date, lien, note).

## Interactions
- Lien croisé avec l'onglet Concours (`08-concours.md`) : possibilité de naviguer d'un
  concours vers ses textes de référence et inversement.

## Critères d'acceptation
- [ ] Les 4 concours ont chacun leur liste de textes de référence, même partielle au départ.
- [ ] Le statut "vérifié / à confirmer" est visible pour chaque texte.
- [ ] L'utilisateur peut ajouter, modifier, supprimer une entrée de texte de référence.
- [ ] Un lien permet de naviguer entre l'onglet Concours et l'onglet Textes de référence pour
      un même concours.
