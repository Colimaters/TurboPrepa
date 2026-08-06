# 00 — Architecture générale & Design System

## Contexte
Application d'organisation pour la préparation de plusieurs concours de la fonction
publique (Gardien de la paix, Officier de police, Commissaire de police, Officier de
gendarmerie). Une première version a été prototypée sous forme de fichier HTML unique
(artifact Claude) avec stockage clé-valeur lié au compte. **Ce prototype sert de référence
fonctionnelle** (contenu des onglets, interactions, design), mais l'implémentation cible
est une **application de bureau native pour Mac**, à usage strictement local et
mono-utilisateur (pas de compétence technique attendue à l'installation).

Ce fichier décrit ce qui est **commun à tous les onglets** : stack technique, packaging,
design system, et modèle de données. Les fichiers `01-…` à `09-…` décrivent chaque
fonctionnalité en détail et référencent ce document plutôt que de répéter ces règles.

## Stack technique retenue

### Framework applicatif : **Wails**
- Backend en **Go**, interface en **HTML/CSS/JS** (réutilisation directe du design pastel
  et des interactions déjà pensées dans le prototype), rendu via le moteur **WebKit natif
  de macOS** — pas de Chromium embarqué, donc application légère (de l'ordre de 10 à 20 Mo)
  et rapide au démarrage.
- Le frontend peut être en JS/HTML/CSS "vanilla" ou avec un framework léger (ex. Vue, Svelte)
  au choix de l'implémentation ; ce n'est pas prescriptif ici, seul le rendu final (design
  system ci-dessous) compte.
- Toute la logique métier sensible (accès aux fichiers, accès à la base de données,
  génération de la planification automatique, etc.) doit vivre côté **Go**, exposée au
  frontend via les bindings Wails, plutôt que dans le JS du frontend.

### Stockage : **SQLite**
- Base de données **SQLite embarquée**, un seul fichier `.db` stocké localement dans le
  répertoire applicatif standard macOS : `~/Library/Application Support/<NomApp>/data.db`.
- Driver Go recommandé : **`modernc.org/sqlite`** (implémentation SQLite pure Go, sans
  CGO). Ce choix est volontaire : il évite toute dépendance à un compilateur C ou à des
  bibliothèques système externes lors du build, ce qui simplifie la compilation ET la
  distribution du binaire final — cohérent avec l'objectif "installation ultra simple".
  Ne pas utiliser `mattn/go-sqlite3` (nécessite CGO) sauf raison technique impérieuse.
- Les fichiers importés par l'utilisateur (Word, PDF, images, Excel, Numbers…) ne sont
  **pas** stockés en BLOB dans SQLite : ils sont copiés dans un sous-dossier du répertoire
  applicatif (ex. `~/Library/Application Support/<NomApp>/fichiers/{uuid}-{nom-original}`),
  et seule la référence (chemin, métadonnées) est stockée en base. Cela évite d'alourdir le
  fichier `.db` et facilite l'accès direct aux fichiers si besoin.
- Toutes les migrations de schéma (voir "Modèle de données" ci-dessous) doivent être
  gérées par un mécanisme de migration versionné exécuté au démarrage de l'application
  (ex. table `schema_migrations`, ou bibliothèque de migration Go dédiée), pour ne jamais
  perdre les données existantes lors d'une mise à jour de l'application.

### Packaging et distribution
- Build via `wails build` générant un bundle `.app` macOS natif.
- Distribution à l'utilisateur sous forme de **`.dmg`** contenant l'icône de l'application
  et un raccourci vers le dossier Applications, pour reproduire le geste d'installation
  standard Mac ("glisser l'icône dans Applications").
- **Aucune licence Apple Developer** : l'application n'est pas signée/notarisée par un
  compte payant. Conséquence à documenter pour l'utilisateur : au tout premier lancement,
  macOS (Gatekeeper) affichera un avertissement "développeur non identifié". La procédure à
  suivre (à fournir sous forme d'une notice courte, avec capture d'écran, livrée avec le
  `.dmg`) : **clic droit sur l'icône de l'application → Ouvrir → confirmer**. Cette étape
  n'est nécessaire qu'une seule fois ; les lancements suivants se font par simple double-clic
  ou depuis le Dock/Launchpad comme n'importe quelle application.
- Une fois lancée, l'icône de l'application apparaît normalement dans le Dock (et peut y
  être épinglée par l'utilisateur) — comportement natif de toute `.app` macOS, aucune
  configuration spécifique à prévoir côté code.
- L'application fonctionne **entièrement hors connexion**, à l'exception des clics
  volontaires de l'utilisateur vers les liens externes de l'onglet Veille juridique
  (`07-veille-juridique.md`), qui s'ouvrent dans le navigateur par défaut du système.

## Navigation générale
Barre d'onglets cliquables, toujours visible (en haut ou sur le côté selon la maquette
retenue — voir "Layout"), avec 9 sections :

1. Accueil
2. Matières
3. Planning
4. Annuaire (jurisprudence par code)
5. Jurisprudence (fiches de synthèse)
6. Veille juridique
7. Quiz du jour
8. Concours (structure des épreuves)
9. Textes de référence

Chaque section est un module indépendant au niveau du code (composant/route séparé) mais
partage le design system et la couche de données décrits ici.

## Layout
- Bandeau de navigation **en haut** ou **barre latérale** (au choix, mais un seul des deux,
  pas les deux en même temps) — cliquable, avec l'onglet actif visuellement distinct.
- Sur mobile / écran étroit : la barre latérale se transforme en menu hamburger ou en barre
  du bas ; la barre du haut reste sticky.
- Contenu principal en dessous/à côté, scrollable indépendamment de la nav.
- Pas de rechargement de page entre onglets (SPA-like), transition douce (fade/slide léger,
  200–300ms).

## Design System (à respecter dans toutes les fonctionnalités)
- **Ambiance** : moderne, dynamique, mais reposante — l'utilisateur va passer des mois
  dessus en période de révisions intenses, éviter le criard.
- **Palette** : tons pastels. Proposer une palette de base (ex. bleu poudré, lavande, vert
  sauge, pêche, gris perle) utilisée pour les fonds, cartes, accents. Le panel de 15
  couleurs utilisé pour catégoriser les matières (voir fichier `03-planning.md`) doit rester
  cohérent avec cette palette pastel (pas de couleurs saturées agressives).
- **Typographie** : police douce, lisible, arrondie (ex. type Nunito, Quicksand, Poppins ou
  équivalent système). Taille de base confortable (16px mini pour le corps de texte),
  hiérarchie claire des titres. Doit rester ajustable (voir retour utilisateur "la police
  est trop petite dans Matières" → prévoir un réglage de taille de police global, pas juste
  un correctif ponctuel).
- **Composants** : cartes à coins arrondis, ombres légères, boutons avec état hover/actif
  visibles, feedback visuel systématique sur les actions (sauvegarde, ajout, suppression).
- **Accessibilité minimale** : contraste suffisant malgré les pastels (texte toujours
  lisible), tailles de cible tactile correctes (boutons ≥ 40px sur mobile).

## Modèle de données — principes transverses
Toutes les données saisies par l'utilisateur (matières, chapitres, fichiers, planning,
annuaire, jurisprudence, scores de quiz, réponses Q/R d'oral, textes de référence complétés
par l'utilisateur) sont stockées dans la base **SQLite** locale décrite ci-dessus, selon les
principes suivants :

1. **Persister** entre les sessions et entre les mises à jour de l'application. Une mise à
   jour du code (nouvelle fonctionnalité, correctif visuel) ne doit **jamais** effacer ou
   altérer les données déjà saisies : le fichier `data.db` n'est jamais recréé depuis zéro
   lors d'une mise à jour, uniquement migré si le schéma évolue.
2. **Schéma versionné** : une table `schema_migrations` (ou équivalent) enregistre la
   version de schéma appliquée, pour exécuter au démarrage uniquement les migrations
   manquantes, de façon idempotente.
3. **Tables principales** (schéma indicatif, à affiner en implémentation — les noms de
   colonnes ci-dessous font référence pour rester cohérents entre toutes les specs) :
   - `matieres (id, nom, couleur, ordre)`
   - `chapitres (id, matiere_id, sous_onglet, nom, statut, ordre)`
     — `sous_onglet` ∈ {programme, fiches, cours_entier, annales}, voir `02-matieres.md`
   - `fichiers (id, chapitre_id?, tache_id?, nom_affiche, nom_original, chemin_disque,
     type_mime, taille, date_ajout)`
     — rattaché soit à un chapitre (Matières), soit à une tâche (Planning), selon le
     contexte d'import, cf. `10-import-fichiers.md`
   - `taches_planning (id, matiere_id?, chapitre_id?, titre, date, heure_debut, heure_fin,
     couleur, notes)`
   - `annuaire_entrees (id, nom_arret, reference, code, localisation_dalloz, notions, notes)`
   - `jurisprudence_fiches (id, nom_arret, reference, code, localisation_dalloz, faits,
     solution, portee, notions, matiere_id?)`
   - `quiz_questions (id, question, choix_json, bonne_reponse, theme, explication?)`
   - `quiz_historique (id, date, score, questions_json)`
   - `citations_motivation (id, texte, auteur, source?, attribution_incertaine)`
   - `concours (id, nom)` avec sous-tables `concours_epreuves (id, concours_id, categorie,
     contenu, source, statut_verification)` où `categorie` ∈ {ecrit, oral, sport, admission}
   - `oral_questions_suggerees (id, concours_id, theme, question)`
   - `oral_questions_perso (id, concours_id, question, reponse)`
   - `textes_reference (id, concours_id, nature, intitule, date, lien, statut_verification,
     note)`
   - `veille_liens (id, nom, url, description)`
4. Toutes les tables avec hiérarchie (dossiers dans une matière, sous-éléments d'un
   concours…) utilisent des **clés étrangères** classiques plutôt que des clés composées
   textuelles, SQLite gérant nativement les contraintes d'intégrité référentielle
   (`PRAGMA foreign_keys = ON`).
5. Prévoir un mécanisme d'**export/sauvegarde manuelle** accessible depuis l'application
   (ex. bouton "Exporter mes données") qui copie le fichier `data.db` et le dossier
   `fichiers/` vers un emplacement choisi par l'utilisateur (Finder natif via Wails) — utile
   en cas de changement de machine ou de sauvegarde de précaution, l'utilisateur n'étant pas
   à l'aise pour aller chercher lui-même le fichier dans `~/Library/Application Support`.

## Import de fichiers — règles communes
Le détail fonctionnel de l'import est dans `09-import-fichiers.md`, mais les règles de
base s'appliquent partout où un import est proposé (Matières, Planning) :

- Formats acceptés : Word (.doc/.docx), Pages (.pages), PDF (.pdf), images (.jpg/.jpeg/.png),
  Excel (.xlsx/.xls), Numbers (.numbers).
- Import à l'unité **et** import groupé (plusieurs fichiers ou une archive .zip en un seul
  glisser-déposer).
- Les fichiers importés sont des **pièces jointes consultables/téléchargeables** rattachées
  à un élément (chapitre, dossier, tâche). L'application n'a pas d'obligation de lire/
  extraire automatiquement le contenu texte du fichier (OCR, parsing) sauf mention contraire
  dans une spec dédiée — le champ concerné doit alors être rempli manuellement.
- Chaque fichier importé garde ses métadonnées : nom d'origine, type, taille, date d'ajout.

## Hors périmètre explicite (pour ce lot de specs)
- Authentification multi-utilisateurs / comptes.
- Classement compétitif entre candidats réels (remplacé par des statistiques personnelles,
  voir `06-quiz-du-jour.md`).
- Lecture/OCR automatique du contenu des fichiers importés.

## Découpage des fichiers de spec
| Fichier | Fonctionnalité |
|---|---|
| `01-accueil.md` | Page d'accueil, citation du jour, résumé de progression |
| `02-matieres.md` | Onglet Matières (17+ matières, sous-onglets, statuts, dossiers) |
| `03-planning.md` | Onglet Planning (calendrier, tâches, planification automatique) |
| `04-annuaire.md` | Onglet Annuaire (jurisprudence classée par code) |
| `05-jurisprudence.md` | Onglet Jurisprudence (fiches de synthèse) |
| `06-quiz-du-jour.md` | Onglet Quiz du jour |
| `07-veille-juridique.md` | Onglet Veille juridique |
| `08-concours.md` | Onglet Concours (structure des épreuves, oral) |
| `09-textes-reference.md` | Onglet Textes de référence |
| `10-import-fichiers.md` | Système d'import de fichiers (transverse) |
