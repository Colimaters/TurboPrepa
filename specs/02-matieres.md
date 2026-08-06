# 02 — Matières

Dépend de : `00-architecture-generale-et-design.md`, `10-import-fichiers.md`

## Objectif
Espace central où l'utilisateur organise tout son contenu de révision, matière par matière :
cours, fiches, annales, et suivi de statut par chapitre.

## Liste des matières (données initiales à charger)
1. Droit pénal général
2. Droit pénal spécial
3. Procédure pénale
4. Droit européen
5. Droit constitutionnel
6. Droit administratif
7. Droit de la fonction publique
8. Libertés publiques
9. Anglais
10. Dictée
11. Culture générale
12. Connaissance du monde contemporain
13. Connaissance de l'institution policière
14. Note de synthèse
15. Cas pratique
16. Annales (transverses — voir note ci-dessous)
17. Sport
18. Lecture

> Note : "Annales" apparaît dans la demande initiale comme matière à part entière ET,
> suite à une demande ultérieure de l'utilisateur, comme **sous-onglet à ajouter dans
> chaque matière**. Implémenter les deux : garder "Annales" comme matière transverse dans
> la liste (pour des annales générales/toutes matières confondues), ET ajouter un
> sous-onglet "Annales" dans chaque matière individuelle (annales spécifiques à cette
> matière). Ne pas fusionner les deux usages.

La liste des matières doit être **modifiable par l'utilisateur** : ajout, suppression,
renommage d'une matière, sans limite au nombre de matières (l'app ne doit pas être codée en
dur avec exactement 18 entrées non modifiables).

## Vue liste des matières
- Vue d'ensemble type grille ou liste des matières, chacune avec :
  - Nom de la matière.
  - Couleur associée (cohérente avec le panel de 15 couleurs pastel utilisé en Planning,
    voir `03-planning.md`).
  - Indicateur de progression synthétique (ex. proportion de chapitres "maîtrisé").
- Clic sur une matière → ouvre la vue détaillée de la matière.

## Vue détaillée d'une matière
Sous-onglets, **au format liste** (et non plus grille/cartes, suite au retour utilisateur
explicite "que le visuel des onglets au sein des matières soit au format de liste") :

1. **Programme**
   - Liste des chapitres du programme de cette matière.
   - Chaque chapitre a un **statut** parmi : à planifier / planifié / en cours / maîtrisé,
     modifiable directement dans la liste (ex. menu déroulant ou badge cliquable par ligne).
   - Chaque chapitre peut avoir des fichiers rattachés (voir Dossiers/fichiers ci-dessous).
   - Section "Travaux à rendre et révisions à faire" centralisée pour la matière : liste
     d'items avec date d'échéance, cochable.
2. **Fiches**
   - Liste des fiches de révision de la matière (fichiers importés et/ou texte libre créé
     dans l'app).
3. **Cours entier**
   - Liste des supports de cours complets importés pour la matière.
4. **Annales** (nouveau sous-onglet, spécifique à cette matière)
   - Liste des annales/sujets d'examens propres à cette matière, avec leurs fichiers
     rattachés (sujet, corrigé le cas échéant).

## Gestion des dossiers et fichiers au sein d'une matière
- Chaque sous-onglet (Programme/Fiches/Cours entier/Annales) fonctionne comme un
  **gestionnaire de dossiers** :
  - Créer un dossier/chapitre.
  - **Renommer** un dossier existant.
  - **Supprimer** un dossier (avec confirmation, car cela supprime son contenu).
  - **Réorganiser** (déplacer un fichier d'un dossier à un autre).
- Import de fichiers :
  - Import **unitaire** : un fichier à la fois.
  - Import **groupé** : plusieurs fichiers sélectionnés en une fois, ou une archive .zip
    contenant plusieurs documents, décompressée automatiquement et répartie comme pièces
    jointes du dossier cible.
  - **Ajout** de fichiers à un dossier déjà existant (pas seulement à la création) — point
    explicitement demandé par l'utilisateur : l'import ne doit pas être limité au moment de
    la création du dossier.
  - Formats supportés : Word (.doc/.docx), Pages (.pages), PDF, images (JPG/PNG), **Excel
    (.xlsx/.xls)**, **Numbers (.numbers)**. Détail technique dans `10-import-fichiers.md`.

## Interactions avec le Planning
- Les chapitres au statut "à planifier" doivent être sélectionnables depuis l'onglet
  Planning pour la planification automatique (voir `03-planning.md`). Pas de duplication de
  données : le Planning lit la liste des chapitres/statuts de Matières.
- Marquer un chapitre "maîtrisé" depuis Matières doit être reflété dans le résumé de
  progression de l'Accueil (voir `01-accueil.md`).

## Critères d'acceptation
- [ ] Les 18 matières listées ci-dessus sont préchargées à l'initialisation.
- [ ] Ajouter/renommer/supprimer une matière fonctionne sans casser les données des autres
      matières.
- [ ] Chaque matière expose bien 4 sous-onglets : Programme, Fiches, Cours entier, Annales,
      affichés en **format liste**.
- [ ] On peut importer plusieurs fichiers d'un coup (ou un .zip) dans un dossier.
- [ ] On peut ajouter un fichier à un dossier déjà existant, à tout moment.
- [ ] On peut renommer et supprimer un dossier existant.
- [ ] Le statut de chaque chapitre est visible et modifiable directement dans la liste.
- [ ] Les formats Excel et Numbers sont acceptés à l'import au même titre que Word/PDF/Pages/
      images.
