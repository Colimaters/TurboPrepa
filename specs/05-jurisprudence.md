# 05 — Jurisprudence (fiches de synthèse)

Dépend de : `00-architecture-generale-et-design.md`, `04-annuaire.md`

## Objectif
Résumer les grandes décisions de jurisprudence à connaître pour les concours, et les
répertorier en indiquant leur emplacement dans le code pénal (ou le code concerné).

## Distinction avec l'onglet "Annuaire" (`04-annuaire.md`)
Voir section correspondante dans `04-annuaire.md`. Ici, l'accent est mis sur le **contenu**
(faits, solution, portée) plutôt que sur la seule localisation.

## Fonctionnalités

### Création d'une fiche de jurisprudence
Formulaire de fiche avec :
- Nom de l'arrêt et référence (juridiction, date, numéro).
- Code concerné et **localisation dans le code** (article/section — peut être repris ou lié
  depuis une entrée existante de l'Annuaire pour éviter la double-saisie).
- **Faits** (résumé court).
- **Solution / principe posé par la décision**.
- **Portée / évolutions ultérieures** (revirements, confirmations) si pertinent.
- Notion(s) juridique(s) associée(s) (mots-clés).
- Matière de rattachement (liste des matières de droit, cohérente avec `02-matieres.md`).

### Organisation et navigation
- Liste des fiches, filtrable par matière et par notion.
- Recherche texte libre sur l'ensemble des champs (nom, faits, solution, notions).
- Vue détaillée d'une fiche = tous les champs affichés de façon structurée et lisible
  (typographie hiérarchisée : nom de l'arrêt en titre, faits/solution/portée en sections
  distinctes).

### Lien avec Annuaire
- Depuis une fiche de Jurisprudence, lien optionnel vers l'entrée correspondante de
  l'Annuaire (et réciproquement) si l'utilisateur les a rattachées.

## Critères d'acceptation
- [ ] On peut créer, modifier, supprimer une fiche de jurisprudence.
- [ ] Chaque fiche indique clairement sa localisation dans le code concerné.
- [ ] Le filtrage par matière et la recherche texte libre fonctionnent correctement.
- [ ] Une fiche peut être liée à une entrée d'Annuaire sans dupliquer la saisie de la
      référence.
