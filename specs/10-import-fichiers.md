# 10 — Import de fichiers (fonctionnalité transverse)

Dépend de : `00-architecture-generale-et-design.md`
Utilisé par : `02-matieres.md`, `03-planning.md`

## Objectif
Fournir un composant d'import de fichiers unique et réutilisable, utilisé partout dans
l'application où l'utilisateur doit rattacher un document à un élément (chapitre, dossier,
tâche de planning).

## Formats supportés
| Format | Extensions | Notes |
|---|---|---|
| Word | .doc, .docx | |
| Pages | .pages | format propriétaire Apple, traité comme pièce jointe opaque |
| PDF | .pdf | |
| Images | .jpg, .jpeg, .png | |
| Excel | .xlsx, .xls | |
| Numbers | .numbers | format propriétaire Apple, traité comme pièce jointe opaque |
| Archive | .zip | uniquement comme **conteneur** d'import groupé, voir ci-dessous |

## Modes d'import

### Import unitaire
- Sélection d'un seul fichier via un input file classique ou glisser-déposer.
- Rattachement immédiat à l'élément cible (dossier de matière, tâche de planning…).

### Import groupé
- Sélection multiple de fichiers en une seule opération (multi-sélection dans le sélecteur
  système), **ou**
- Dépôt d'une **archive .zip** : l'application doit la décompresser côté client (ou côté
  traitement si backend) et créer une pièce jointe par fichier contenu, rattachées toutes
  au même dossier/élément cible. Les fichiers non pris en charge à l'intérieur du .zip
  (formats hors de la liste ci-dessus) doivent être ignorés avec un message explicite listant
  ce qui n'a pas été importé, sans faire échouer tout l'import.

### Ajout à un élément existant
- L'import n'est pas limité à la création d'un dossier/tâche : il doit être possible
  d'ajouter un ou plusieurs fichiers à un dossier ou une tâche **déjà existants**, à tout
  moment, via un bouton "Ajouter des fichiers" visible sur l'élément concerné.

## Limites et comportement
- Taille maximale par fichier : à définir techniquement selon le mécanisme de stockage
  retenu (indicative : jusqu'à quelques Mo par fichier en stockage navigateur ; à revoir à
  la hausse si un vrai backend/serveur de fichiers est mis en place par Claude Code).
- Les fichiers importés sont des **pièces jointes consultables et téléchargeables** :
  affichage d'un nom, d'une icône selon le type, d'une taille, d'une date d'ajout, et d'un
  bouton d'ouverture/téléchargement.
- **Pas de lecture automatique du contenu** en V1 (pas d'OCR, pas d'extraction de texte, pas
  d'extraction automatique d'un emploi du temps depuis un fichier importé). Ce point doit
  être communiqué clairement dans l'interface (ex. info-bulle "fichier joint, à consulter
  manuellement") pour éviter toute confusion côté utilisateur — cf. retour explicite lors des
  échanges précédents sur cette limite.
- Suppression d'un fichier : possible à tout moment, avec confirmation.
- Renommage du nom affiché d'un fichier (indépendamment du nom du fichier d'origine) :
  possible.

## Critères d'acceptation
- [ ] Tous les formats listés sont acceptés par le sélecteur de fichiers.
- [ ] L'import multiple (plusieurs fichiers ou un .zip) fonctionne en une seule opération et
      répartit correctement les fichiers dans l'élément cible.
- [ ] Un fichier peut être ajouté à un dossier/tâche existant sans recréer l'élément.
- [ ] Un fichier importé peut être renommé (affichage) et supprimé.
- [ ] L'absence de lecture automatique du contenu est explicite dans l'interface, pas juste
      dans la documentation.
