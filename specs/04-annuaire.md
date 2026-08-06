# 04 — Annuaire (jurisprudence par référence de code)

Dépend de : `00-architecture-generale-et-design.md`

## Objectif
Retrouver instantanément, en épreuve, où se trouve un arrêt donné dans son code (édition
Dalloz), sans perdre de temps à feuilleter au hasard.

## Distinction avec l'onglet "Jurisprudence" (`05-jurisprudence.md`)
- **Annuaire** = un **index de localisation** : pour un arrêt donné, où le trouver
  précisément dans l'édition Dalloz du code concerné (numéro d'article, section, page si
  pertinent).
- **Jurisprudence** = des **fiches de synthèse** du contenu et de la portée des grandes
  décisions.
Les deux onglets peuvent se référencer mutuellement (lien croisé depuis une entrée de
l'Annuaire vers la fiche de synthèse correspondante si elle existe), mais restent deux
structures de données distinctes.

## Fonctionnalités

### Saisie d'une entrée
Formulaire pour ajouter une entrée d'annuaire :
- Nom de l'arrêt (ex. "Benjamin", "Dame Lamotte"…).
- Référence complète de la décision (juridiction, date, numéro).
- Code concerné (pénal, administratif, constitutionnel, etc. — aligné sur la liste des
  matières de droit).
- **Localisation dans le code Dalloz** : article ou section sous laquelle l'arrêt est
  répertorié dans l'édition Dalloz (champ texte libre, car la codification peut varier
  d'une édition à l'autre).
- Notion(s) juridique(s) associée(s) (mots-clés, pour la recherche).
- Notes libres (optionnel).

### Recherche
Barre de recherche avec filtres combinables :
- Par **référence** (numéro de décision, juridiction, date).
- Par **nom d'arrêt**.
- Par **notion** (mot-clé juridique).
Recherche instantanée (résultats filtrés au fur et à mesure de la frappe), tolérante aux
fautes de frappe mineures si possible (recherche approximative), sinon au minimum
insensible à la casse et aux accents.

### Affichage des résultats
- Liste de résultats montrant : nom de l'arrêt, référence, localisation dans le code Dalloz
  en évidence (c'est l'information la plus recherchée en situation d'épreuve, doit être
  visible sans clic supplémentaire).
- Clic sur un résultat → fiche détaillée complète.

## Critères d'acceptation
- [ ] On peut ajouter, modifier, supprimer une entrée d'annuaire.
- [ ] La recherche fonctionne par référence, par nom d'arrêt et par notion, séparément ou
      combinés.
- [ ] La localisation dans le code Dalloz est immédiatement visible dans les résultats de
      recherche, sans navigation supplémentaire.
- [ ] L'onglet reste utilisable et rapide même avec plusieurs centaines d'entrées (recherche
      réactive).
