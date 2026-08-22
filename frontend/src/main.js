import './style.css';
import './app.css';

import { GetDashboard, ToggleTodayTask } from '../wailsjs/go/main/App';

const app = document.querySelector('#app');
const api = (name, ...args) => window.go.main.App[name](...args);
const tabs = [
  ['programme', 'Programme'],
  ['fiches', 'Fiches'],
  ['cours_entier', 'Cours entier'],
  ['annales', 'Annales'],
];
const statusLabels = {
  a_planifier: 'À planifier',
  planifie: 'Planifié',
  en_cours: 'En cours',
  maitrise: 'Maîtrisé',
};
const ui = { page: 'Accueil', subjectId: null, tab: 'programme' };

function escapeHTML(value) {
  return String(value).replace(/[&<>"']/g, (character) => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
  })[character]);
}

function errorMessage(error) {
  return error instanceof Error ? error.message : String(error);
}

function showNotice(message, type = 'success') {
  const notice = document.querySelector('#notice');
  if (!notice) return;
  notice.textContent = message;
  notice.className = `notice ${type}`;
  notice.hidden = false;
  window.setTimeout(() => { notice.hidden = true; }, 3500);
}

function shell(content) {
  app.innerHTML = `
    <div class="shell">
      <header class="topbar">
        <button class="brand" type="button" data-page="Accueil" aria-label="TurboPrepa, accueil">
          <span class="brand-mark">T</span><span>Turbo<span>Prepa</span></span>
        </button>
        <nav aria-label="Navigation principale">
          ${['Accueil', 'Matières', 'Planning', 'Annuaire', 'Jurisprudence', 'Veille juridique', 'Quiz du jour', 'Concours', 'Textes']
    .map((page) => `<button class="nav-item ${ui.page === page ? 'active' : ''}" type="button" data-page="${page}" ${page !== 'Accueil' && page !== 'Matières' ? 'disabled' : ''}>${page}</button>`).join('')}
        </nav>
      </header>
      <p id="notice" class="notice" role="status" hidden></p>
      ${content}
    </div>`;
  app.querySelectorAll('[data-page]').forEach((button) => button.addEventListener('click', () => {
    ui.page = button.dataset.page;
    ui.subjectId = null;
    render();
  }));
}

async function render() {
  if (ui.page === 'Matières') {
    await renderMatieres();
    return;
  }
  await renderDashboard();
}

async function renderDashboard() {
  shell('<main class="dashboard"><p class="loading">Chargement de l’accueil…</p></main>');
  try {
    const dashboard = await GetDashboard();
    const total = dashboard.progress.toPlan + dashboard.progress.planned + dashboard.progress.inProgress + dashboard.progress.mastered;
    const percent = total ? Math.round((dashboard.progress.mastered / total) * 100) : 0;
    shell(`
      <main class="dashboard">
        <section class="page-heading"><div><p class="eyebrow">TABLEAU DE BORD</p><h1>Bonjour, prête à avancer ?</h1>
          <p class="date-label">${escapeHTML(new Intl.DateTimeFormat('fr-FR', { weekday: 'long', day: 'numeric', month: 'long' }).format(new Date(`${dashboard.today}T12:00:00`)))}</p></div>
          <div class="daily-badge">✦ Un pas à la fois</div></section>
        <section class="dashboard-grid">
          <article class="card quote-card"><div class="card-heading"><span class="card-icon quote-icon">“</span><p>PENSÉE DU JOUR</p></div>
            <blockquote>« ${escapeHTML(dashboard.quote.text)} »</blockquote><p class="quote-author">${escapeHTML(dashboard.quote.author)}</p><p class="quote-source">${escapeHTML(dashboard.quote.source || '')}</p></article>
          <article class="card progress-card"><button id="go-subjects" class="card-link" type="button"><div class="card-heading"><span class="card-icon progress-icon">◔</span><p>MA PROGRESSION</p><span class="arrow">→</span></div>
            <div class="progress-main"><div class="progress-ring" style="--progress:${percent * 3.6}deg"><strong>${percent}%</strong><span>maîtrisé</span></div>
            <div class="status-list">${Object.entries(statusLabels).map(([key, label]) => `<div class="status-row"><span>${label}</span><strong>${dashboard.progress[{ a_planifier: 'toPlan', planifie: 'planned', en_cours: 'inProgress', maitrise: 'mastered' }[key]]}</strong></div>`).join('')}</div></div>
            <p class="progress-note">${total ? `${dashboard.progress.mastered} chapitre${dashboard.progress.mastered > 1 ? 's' : ''} maîtrisé${dashboard.progress.mastered > 1 ? 's' : ''} sur ${total}` : 'Ajoutez vos chapitres dans Matières pour suivre votre progression.'}</p></button></article>
        </section>
        <section class="card tasks-card"><div class="tasks-header"><div class="card-heading"><span class="card-icon task-icon">✓</span><div><p>AUJOURD'HUI</p><h2>Mes tâches du jour</h2></div></div></div>
          <div class="tasks-list">${dashboard.tasks.length ? dashboard.tasks.map((task) => `<div class="task-row ${task.completed ? 'done' : ''}"><button class="task-toggle checkmark" data-task="${task.id}" data-completed="${!task.completed}" type="button">${task.completed ? '✓' : ''}</button><span class="task-color" style="background:${escapeHTML(task.color)}"></span><span class="task-copy"><strong>${escapeHTML(task.title)}</strong><span>${escapeHTML(task.subject)} · ${escapeHTML(task.startTime)}–${escapeHTML(task.endTime)}</span></span></div>`).join('') : '<div class="empty-state"><div><h3>Aucune tâche prévue aujourd’hui</h3><p>Profitez-en pour planifier une séance à votre rythme.</p></div></div>'}</div>
        </section>
      </main>`);
    document.querySelector('#go-subjects').addEventListener('click', () => { ui.page = 'Matières'; render(); });
    document.querySelectorAll('[data-task]').forEach((button) => button.addEventListener('click', async () => {
      try { await ToggleTodayTask(Number(button.dataset.task), button.dataset.completed === 'true'); await renderDashboard(); } catch (error) { showNotice(`Impossible de mettre à jour la tâche : ${errorMessage(error)}`, 'error'); }
    }));
  } catch (error) {
    shell(`<main class="dashboard"><p class="load-error">Impossible de charger l’accueil : ${escapeHTML(errorMessage(error))}</p></main>`);
  }
}

async function renderMatieres() {
  shell('<main class="subjects-page"><p class="loading">Chargement des matières…</p></main>');
  try {
    if (ui.subjectId) {
      const detail = await api('GetMatiereDetail', ui.subjectId);
      renderSubjectDetail(detail);
      return;
    }
    const [subjects, colors] = await Promise.all([api('ListMatieres'), api('ListPastelColors')]);
    renderSubjectList(subjects, colors);
  } catch (error) {
    shell(`<main class="subjects-page"><p class="load-error">Impossible de charger les matières : ${escapeHTML(errorMessage(error))}</p></main>`);
  }
}

function renderSubjectList(subjects, colors) {
  shell(`
    <main class="subjects-page">
      <section class="page-heading"><div><p class="eyebrow">ORGANISATION</p><h1>Mes matières</h1><p class="date-label">Organisez vos cours, fiches, annales et révisions.</p></div>
      <button class="primary-action" id="add-subject" type="button">Ajouter une matière</button></section>
      <section class="subject-grid">${subjects.map((subject) => {
    const percent = subject.chapters ? Math.round((subject.mastered / subject.chapters) * 100) : 0;
    return `<article class="subject-card" style="--subject-color:${escapeHTML(subject.color)}"><button class="subject-open" data-subject="${subject.id}" type="button"><span class="subject-swatch"></span><span><strong>${escapeHTML(subject.name)}</strong><small>${subject.mastered}/${subject.chapters} chapitres maîtrisés</small></span><span class="subject-percent">${percent}%</span></button><div class="subject-actions"><button data-rename-subject="${subject.id}" data-name="${escapeHTML(subject.name)}" type="button">Renommer</button><button data-delete-subject="${subject.id}" type="button">Supprimer</button></div></article>`;
  }).join('')}</section>
    </main>`);
  document.querySelector('#add-subject').addEventListener('click', async () => {
    const name = window.prompt('Nom de la matière :');
    if (!name) return;
    const color = colors[subjects.length % colors.length];
    try { await api('CreateMatiere', name, color); showNotice('Matière ajoutée.'); await renderMatieres(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  });
  document.querySelectorAll('[data-subject]').forEach((button) => button.addEventListener('click', () => { ui.subjectId = Number(button.dataset.subject); render(); }));
  document.querySelectorAll('[data-rename-subject]').forEach((button) => button.addEventListener('click', async () => {
    const name = window.prompt('Nouveau nom de la matière :', button.dataset.name);
    if (!name) return;
    try { await api('RenameMatiere', Number(button.dataset.renameSubject), name); showNotice('Matière renommée.'); await renderMatieres(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
  document.querySelectorAll('[data-delete-subject]').forEach((button) => button.addEventListener('click', async () => {
    if (!window.confirm('Supprimer cette matière et tout son contenu ? Cette action est irréversible.')) return;
    try { await api('DeleteMatiere', Number(button.dataset.deleteSubject)); showNotice('Matière supprimée.'); await renderMatieres(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
}

function renderSubjectDetail(detail) {
  const chapters = detail.chapters.filter((chapter) => chapter.tab === ui.tab);
  shell(`
    <main class="subjects-page">
      <button class="back-action" id="back-subjects" type="button">← Toutes les matières</button>
      <section class="subject-detail-heading"><span class="subject-swatch large" style="background:${escapeHTML(detail.subject.color)}"></span><div><p class="eyebrow">MATIÈRE</p><h1>${escapeHTML(detail.subject.name)}</h1><p class="date-label">${detail.subject.mastered}/${detail.subject.chapters} chapitres du programme maîtrisés</p></div></section>
      <div class="subject-tabs" role="tablist">${tabs.map(([id, label]) => `<button class="${ui.tab === id ? 'active' : ''}" data-tab="${id}" type="button" role="tab" aria-selected="${ui.tab === id}">${label}</button>`).join('')}</div>
      <section class="subject-panel card">
        <div class="panel-heading"><div><h2>${tabs.find(([id]) => id === ui.tab)[1]}</h2><p>${ui.tab === 'programme' ? 'Suivez les chapitres et leur progression.' : 'Dossiers et pièces jointes à consulter manuellement.'}</p></div><button class="primary-action" id="add-chapter" type="button">Ajouter ${ui.tab === 'programme' ? 'un chapitre' : 'un dossier'}</button></div>
        <p class="attachment-notice">Les fichiers joints sont consultables manuellement : TurboPrepa ne lit ni n’extrait leur contenu.</p>
        <div class="chapter-list">${chapters.length ? chapters.map((chapter) => chapterRow(chapter, detail.chapters)).join('') : '<p class="empty-list">Aucun élément pour le moment.</p>'}</div>
      </section>
      ${ui.tab === 'programme' ? renderWorks(detail.works) : ''}
    </main>`);
  document.querySelector('#back-subjects').addEventListener('click', () => { ui.subjectId = null; render(); });
  document.querySelectorAll('[data-tab]').forEach((button) => button.addEventListener('click', () => { ui.tab = button.dataset.tab; render(); }));
  document.querySelector('#add-chapter').addEventListener('click', async () => {
    const name = window.prompt(ui.tab === 'programme' ? 'Nom du chapitre :' : 'Nom du dossier :');
    if (!name) return;
    const content = ui.tab === 'fiches' ? (window.prompt('Texte libre de la fiche (facultatif) :') || '') : '';
    try { await api('CreateChapter', detail.subject.id, ui.tab, name, content); showNotice('Élément ajouté.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  });
  bindChapterActions(detail);
  bindWorkActions(detail);
}

function chapterRow(chapter, allChapters) {
  return `<article class="chapter-row">
    <div class="chapter-main"><strong>${escapeHTML(chapter.name)}</strong>${chapter.content ? `<p>${escapeHTML(chapter.content)}</p>` : ''}
      ${chapter.tab === 'programme' ? `<label>Statut <select data-status="${chapter.id}">${Object.entries(statusLabels).map(([value, label]) => `<option value="${value}" ${chapter.status === value ? 'selected' : ''}>${label}</option>`).join('')}</select></label>` : ''}</div>
    <div class="chapter-actions"><button data-edit-chapter="${chapter.id}" data-name="${escapeHTML(chapter.name)}" data-content="${escapeHTML(chapter.content)}" type="button">Modifier</button><button data-import="${chapter.id}" type="button">Ajouter des fichiers</button><button data-delete-chapter="${chapter.id}" type="button">Supprimer</button></div>
    <ul class="attachment-list">${chapter.files.map((file) => `<li><span>${escapeHTML(file.displayName)} <small>${formatBytes(file.size)}</small></span><label>Déplacer vers <select data-move-file="${file.id}"><option value="${chapter.id}">Ce dossier</option>${allChapters.filter((target) => target.id !== chapter.id).map((target) => `<option value="${target.id}">${escapeHTML(target.name)}</option>`).join('')}</select></label><button data-open-file="${file.id}" type="button">Ouvrir</button><button data-rename-file="${file.id}" data-name="${escapeHTML(file.displayName)}" type="button">Renommer</button><button data-delete-file="${file.id}" type="button">Supprimer</button></li>`).join('')}</ul>
  </article>`;
}

function renderWorks(works) {
  return `<section class="subject-panel card works-panel"><div class="panel-heading"><div><h2>Travaux à rendre et révisions à faire</h2><p>Une liste centralisée pour cette matière.</p></div><button class="secondary-action" id="add-work" type="button">Ajouter</button></div>
    <div class="work-list">${works.length ? works.map((work) => `<div class="work-row ${work.completed ? 'done' : ''}"><label><input data-work-toggle="${work.id}" type="checkbox" ${work.completed ? 'checked' : ''}> <span>${escapeHTML(work.title)}</span></label><span>${work.dueDate ? `Échéance : ${escapeHTML(work.dueDate)}` : 'Sans échéance'}</span><button data-delete-work="${work.id}" type="button">Supprimer</button></div>`).join('') : '<p class="empty-list">Aucun travail ou révision à suivre.</p>'}</div>
  </section>`;
}

function bindChapterActions(detail) {
  document.querySelectorAll('[data-status]').forEach((input) => input.addEventListener('change', async () => {
    try { await api('SetChapterStatus', Number(input.dataset.status), input.value); showNotice('Statut enregistré.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
  document.querySelectorAll('[data-edit-chapter]').forEach((button) => button.addEventListener('click', async () => {
    const name = window.prompt('Nom :', button.dataset.name); if (!name) return;
    const content = ui.tab === 'fiches' ? (window.prompt('Texte libre :', button.dataset.content) || '') : button.dataset.content;
    try { await api('UpdateChapter', Number(button.dataset.editChapter), name, content); showNotice('Élément modifié.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
  document.querySelectorAll('[data-delete-chapter]').forEach((button) => button.addEventListener('click', async () => {
    if (!window.confirm('Supprimer cet élément et ses fichiers ?')) return;
    try { await api('DeleteChapter', Number(button.dataset.deleteChapter)); showNotice('Élément supprimé.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
  document.querySelectorAll('[data-import]').forEach((button) => button.addEventListener('click', async () => {
    try {
      const result = await api('SelectAndImportChapterFiles', Number(button.dataset.import));
      showNotice(`${result.imported.length} fichier(s) importé(s)${result.skipped.length ? ` ; ignorés : ${result.skipped.join(', ')}` : ''}.`, result.skipped.length ? 'warning' : 'success');
      await render();
    } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
  document.querySelectorAll('[data-open-file]').forEach((button) => button.addEventListener('click', () => api('OpenAttachment', Number(button.dataset.openFile)).catch((error) => showNotice(errorMessage(error), 'error'))));
  document.querySelectorAll('[data-rename-file]').forEach((button) => button.addEventListener('click', async () => {
    const name = window.prompt('Nom affiché :', button.dataset.name); if (!name) return;
    try { await api('RenameAttachment', Number(button.dataset.renameFile), name); showNotice('Fichier renommé.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
  document.querySelectorAll('[data-delete-file]').forEach((button) => button.addEventListener('click', async () => {
    if (!window.confirm('Supprimer ce fichier ?')) return;
    try { await api('DeleteAttachment', Number(button.dataset.deleteFile)); showNotice('Fichier supprimé.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
  document.querySelectorAll('[data-move-file]').forEach((select) => select.addEventListener('change', async () => {
    try { await api('MoveAttachment', Number(select.dataset.moveFile), Number(select.value)); showNotice('Fichier déplacé.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
}

function bindWorkActions(detail) {
  const add = document.querySelector('#add-work');
  if (add) add.addEventListener('click', async () => {
    const title = window.prompt('Travail ou révision :'); if (!title) return;
    const dueDate = window.prompt('Échéance au format AAAA-MM-JJ (facultatif) :') || '';
    try { await api('CreateSubjectWork', detail.subject.id, title, dueDate); showNotice('Travail ajouté.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  });
  document.querySelectorAll('[data-work-toggle]').forEach((input) => input.addEventListener('change', async () => {
    try { await api('SetSubjectWorkCompleted', Number(input.dataset.workToggle), input.checked); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
  document.querySelectorAll('[data-delete-work]').forEach((button) => button.addEventListener('click', async () => {
    if (!window.confirm('Supprimer ce travail ?')) return;
    try { await api('DeleteSubjectWork', Number(button.dataset.deleteWork)); showNotice('Travail supprimé.'); await render(); } catch (error) { showNotice(errorMessage(error), 'error'); }
  }));
}

function formatBytes(size) {
  if (size < 1024) return `${size} o`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} Ko`;
  return `${(size / (1024 * 1024)).toFixed(1)} Mo`;
}

render();
