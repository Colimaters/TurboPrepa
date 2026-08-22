import { GetDashboard, ToggleTodayTask } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const app = document.querySelector('#app');
const api = (name, ...args) => {
  const method = window.go?.main?.App?.[name];
  if (typeof method !== 'function') {
    return Promise.reject(new Error(`La fonctionnalité « ${name} » n’est pas encore disponible.`));
  }
  return method(...args);
};
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
const ui = {
  page: 'Accueil', subjectId: null, tab: 'programme',
  planningTab: 'add', calendarView: 'month', calendarDate: '2026-08-01',
  planningSelection: new Map(), editingPlanningTask: null,
  planningImportStatus: null,
};

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

function setPlanningImportStatus(message, state = 'info', busy = false) {
  ui.planningImportStatus = message ? { message, state, busy } : null;
  const feedback = document.querySelector('#planning-import-feedback');
  if (feedback) {
    feedback.textContent = message || '';
    feedback.className = `planning-import-feedback ${state}`;
    feedback.hidden = !message;
  }
  const button = document.querySelector('#import-template');
  if (button) {
    button.disabled = busy;
    button.textContent = busy ? 'Import en cours…' : 'Importer un fichier Excel';
    button.setAttribute('aria-busy', String(busy));
  }
}

EventsOn('planning-import-status', (status) => {
  if (ui.page !== 'Planning' || ui.planningTab !== 'import') return;
  setPlanningImportStatus(status.message, status.state, true);
});

function shell(content) {
  app.innerHTML = `
    <div class="shell">
      <header class="topbar">
        <button class="brand" type="button" data-page="Accueil" aria-label="TurboPrepa, accueil">
          <span class="brand-mark">T</span><span>Turbo<span>Prepa</span></span>
        </button>
        <nav aria-label="Navigation principale">
          ${['Accueil', 'Matières', 'Planning', 'Annuaire', 'Jurisprudence', 'Veille juridique', 'Quiz du jour', 'Concours', 'Textes']
    .map((page) => `<button class="nav-item ${ui.page === page ? 'active' : ''}" type="button" data-page="${page}" ${page !== 'Accueil' && page !== 'Matières' && page !== 'Planning' ? 'disabled' : ''}>${page}</button>`).join('')}
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
  if (ui.page === 'Planning') {
    await renderPlanning();
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

function isoDate(date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

function planningRange() {
  const cursor = new Date(`${ui.calendarDate}T12:00:00`);
  if (ui.calendarView === 'day') return [ui.calendarDate, ui.calendarDate];
  if (ui.calendarView === 'month') {
    const start = new Date(cursor.getFullYear(), cursor.getMonth(), 1, 12);
    start.setDate(start.getDate() - ((start.getDay() + 6) % 7));
    const end = new Date(start);
    end.setDate(start.getDate() + 41);
    return [isoDate(start), isoDate(end)];
  }
  const start = new Date(cursor);
  start.setDate(cursor.getDate() - ((cursor.getDay() + 6) % 7));
  const end = new Date(start);
  end.setDate(start.getDate() + 6);
  return [isoDate(start), isoDate(end)];
}

function planningTaskInput(form) {
  const value = (name) => form.elements[name].value.trim();
  const identifier = (name) => {
    const parsed = Number(value(name));
    return parsed || null;
  };
  return {
    matiereId: identifier('matiereId'), chapitreId: identifier('chapitreId'), title: value('title'),
    date: value('date'), startTime: value('startTime'), endTime: value('endTime'),
    notes: value('notes'),
  };
}

function subjectOptions(subjects, selectedID = null) {
  return `<option value="">Sans matière</option>${subjects.map((subject) => `<option value="${subject.id}" ${Number(subject.id) === Number(selectedID) ? 'selected' : ''}>${escapeHTML(subject.name)}</option>`).join('')}`;
}

function chapterOptions(subjects, subjectID = null, selectedID = null) {
  const subject = subjects.find((item) => Number(item.id) === Number(subjectID));
  return `<option value="">Aucun chapitre</option>${(subject?.chapters || []).map((chapter) => `<option value="${chapter.id}" ${Number(chapter.id) === Number(selectedID) ? 'selected' : ''}>${escapeHTML(chapter.name)}${chapter.status === 'a_planifier' ? ' · À planifier' : ''}</option>`).join('')}`;
}

function taskForm(subjects, task = null, formID = 'planning-task-form') {
  const value = (field, fallback = '') => task ? task[field] ?? fallback : fallback;
  const subjectID = value('matiereId');
  return `<form class="planning-form" id="${formID}">
    <label class="planning-full">Intitulé<input name="title" required maxlength="180" value="${escapeHTML(value('title'))}" placeholder="Ex. Revoir le chapitre 1"></label>
    <label>Matière<select data-task-subject name="matiereId" required>${subjectOptions(subjects, subjectID)}</select></label>
    <label>Chapitre <span class="optional">facultatif</span><select data-task-chapter name="chapitreId">${chapterOptions(subjects, subjectID, value('chapitreId'))}</select></label>
    <label>Date<input name="date" type="date" required value="${escapeHTML(value('date', ui.calendarDate))}"></label>
    <label>Début<input name="startTime" type="time" required value="${escapeHTML(value('startTime', '08:00'))}"></label>
    <label>Fin<input name="endTime" type="time" required value="${escapeHTML(value('endTime', '09:00'))}"></label>
    <p class="form-help planning-full">La couleur de la matière sélectionnée est appliquée automatiquement.</p>
    <label class="planning-full">Notes<textarea name="notes" rows="3">${escapeHTML(value('notes'))}</textarea></label>
    <div class="planning-full form-actions"><button class="primary-action" type="submit">${task ? 'Enregistrer les modifications' : 'Ajouter la tâche'}</button></div>
  </form>`;
}

async function renderPlanning() {
  shell('<main class="planning-page"><p class="loading">Chargement du planning…</p></main>');
  try {
    const [data, preferences] = await Promise.all([api('ListPlanningData'), api('GetWorkdayPreferences')]);
    const [start, end] = planningRange();
    const tasks = await api('ListPlanningTasks', start, end);
    shell(`<main class="planning-page">
      <section class="page-heading"><div><p class="eyebrow">ORGANISATION</p><h1>Mon planning</h1><p class="date-label">Planifiez vos révisions à votre rythme.</p></div></section>
      <div class="planning-tabs" role="tablist">${[['add', 'Ajouter une tâche'], ['import', 'Importer un emploi du temps'], ['generate', 'Planifier automatiquement ma semaine']].map(([id, label]) => `<button class="${ui.planningTab === id ? 'active' : ''}" type="button" data-planning-tab="${id}">${label}</button>`).join('')}</div>
      <section class="planning-panel card">${renderPlanningPanel(data, preferences)}</section>
      ${renderCalendar(tasks)}${renderTaskModal(data)}</main>`);
    bindPlanningActions(data, preferences, tasks);
  } catch (error) {
    shell(`<main class="planning-page"><p class="load-error">Impossible de charger le planning : ${escapeHTML(errorMessage(error))}</p></main>`);
  }
}

function renderPlanningPanel(data, preferences) {
  if (ui.planningTab === 'add') return `<div class="panel-heading"><div><p class="eyebrow">CRÉATION MANUELLE</p><h2>Ajouter une tâche</h2><p>Une tâche créée apparaît immédiatement dans le calendrier.</p></div></div>${taskForm(data.subjects)}`;
  if (ui.planningTab === 'import') {
    const importStatus = ui.planningImportStatus;
    return `<div class="panel-heading"><div><p class="eyebrow">IMPORT EXCEL</p><h2>Importer un emploi du temps</h2><p>Téléchargez le modèle, complétez vos créneaux, puis importez-le dans le calendrier.</p></div></div>
    <div class="planning-import-card"><span class="import-symbol" aria-hidden="true">⇩</span><div><h3>Modèle TurboPrepa</h3><p>Chaque ligne valide du fichier Excel devient une tâche. Les créneaux importés sont immédiatement visibles dans le calendrier.</p></div><div class="planning-actions"><button class="secondary-action" id="download-template" type="button">Télécharger le modèle</button><button class="primary-action" id="import-template" type="button" ${importStatus?.busy ? 'disabled aria-busy="true"' : ''}>${importStatus?.busy ? 'Import en cours…' : 'Importer un fichier Excel'}</button></div></div>
    <p id="planning-import-feedback" class="planning-import-feedback ${importStatus?.state || 'info'}" role="status" ${importStatus?.message ? '' : 'hidden'}>${escapeHTML(importStatus?.message || '')}</p>
    <p class="attachment-notice">L’import accepte uniquement le fichier Excel conforme au modèle. Aucun contenu de pièce jointe n’est lu automatiquement.</p>`;
  }
  return renderGeneratorPanel(data, preferences);
}

function selectionKey(subjectID, chapterID) {
  return `${subjectID}:${chapterID}`;
}

function generatorChapterList(subjects) {
  if (!subjects.length) return '<p class="empty-list">Ajoutez des matières et des chapitres dans Matières pour les planifier ici.</p>';
  return subjects.map((subject) => `<div class="generator-subject"><strong><span style="background:${escapeHTML(subject.color)}"></span>${escapeHTML(subject.name)}</strong>${subject.chapters.length ? `<div>${subject.chapters.map((chapter) => {
    const key = selectionKey(subject.id, chapter.id);
    return `<label><input data-planning-chapter="${key}" data-chapter-id="${chapter.id}" type="checkbox" ${ui.planningSelection.has(key) ? 'checked' : ''}> ${escapeHTML(chapter.name)}${chapter.status === 'a_planifier' ? '<em>À planifier</em>' : ''}</label>`;
  }).join('')}</div>` : '<p class="form-help">Aucun chapitre de programme.</p>'}</div>`).join('');
}

function selectedGeneratorChapters(subjects) {
  return subjects.flatMap((subject) => subject.chapters.map((chapter) => ({ subject, chapter, key: selectionKey(subject.id, chapter.id) }))).filter((item) => ui.planningSelection.has(item.key));
}

function generatorQuestionnaire(subjects) {
  const selected = selectedGeneratorChapters(subjects);
  if (!selected.length) return '<p class="empty-list">Sélectionnez au moins un chapitre dans la liste.</p>';
  return selected.map(({ subject, chapter, key }) => {
    const settings = ui.planningSelection.get(key);
    return `<article class="generator-question"><h3><span style="background:${escapeHTML(subject.color)}"></span>${escapeHTML(subject.name)} · ${escapeHTML(chapter.name)}</h3><div>
      <fieldset><legend>Jour(s) de début</legend><div class="weekday-choice">${['Dim.', 'Lun.', 'Mar.', 'Mer.', 'Jeu.', 'Ven.', 'Sam.'].map((label, day) => `<label><input data-start-day="${key}" type="checkbox" value="${day}" ${settings.startDays.includes(day) ? 'checked' : ''}>${label}</label>`).join('')}</div></fieldset>
      <label>Nombre de révisions<input data-generator-field="${key}" data-field="revisionCount" type="number" min="1" max="20" value="${settings.revisionCount}" required></label>
      <label>Durée (minutes)<input data-generator-field="${key}" data-field="durationMinutes" type="number" min="15" max="720" step="15" value="${settings.durationMinutes}" required></label>
      <label>Espacement (jours)<input data-generator-field="${key}" data-field="spacingDays" type="number" min="0" max="90" value="${settings.spacingDays}" required></label>
    </div></article>`;
  }).join('');
}

function renderGeneratorPanel(data, preferences) {
  return `<div class="panel-heading"><div><p class="eyebrow">GÉNÉRATION GUIDÉE</p><h2>Planifier automatiquement ma semaine</h2><p>Vos chapitres existants sont proposés ici, avec « À planifier » en priorité.</p></div></div>
    <div class="generator-layout"><section><h3>Chapitres à planifier</h3>${generatorChapterList(data.subjects)}</section><section><h3>Contraintes par chapitre</h3><form id="generate-form"><label>Date de départ<input name="startDate" type="date" value="${ui.calendarDate}"></label>${generatorQuestionnaire(data.subjects)}<button class="primary-action" type="submit" ${ui.planningSelection.size ? '' : 'disabled'}>Générer automatiquement l’emploi du temps</button></form></section></div>
    <p class="generator-limit">Dans ce lot, les plages de temps personnel, les pauses et l’alternance entre matières ne sont pas configurables. Seuls les horaires de travail, les jours de début, la durée et l’espacement sont transmis au générateur.</p>
    <form id="workday-form" class="workday-form"><div><h3>Ma journée de travail type</h3><p>Ces créneaux servent à placer les tâches générées.</p></div>${preferences.slots.map((slot) => `<label><input type="checkbox" name="${slot.period}-enabled" ${slot.enabled ? 'checked' : ''}> ${slot.period === 'morning' ? 'Matin' : slot.period === 'afternoon' ? 'Après-midi' : 'Soirée'} <input type="time" name="${slot.period}-start" value="${escapeHTML(slot.start)}"> à <input type="time" name="${slot.period}-end" value="${escapeHTML(slot.end)}"></label>`).join('')}<button class="secondary-action" type="submit">Enregistrer les horaires</button></form>`;
}

function renderTaskModal(data) {
  const task = ui.editingPlanningTask;
  if (!task) return '';
  return `<div class="modal-backdrop" data-close-planning-modal><section class="planning-modal card" role="dialog" aria-modal="true" aria-labelledby="planning-modal-title"><button class="modal-close" data-close-planning-modal type="button" aria-label="Fermer">×</button><p class="eyebrow">MODIFIER UNE TÂCHE</p><h2 id="planning-modal-title">Modifier le créneau</h2>${taskForm(data.subjects, task, 'planning-edit-form')}<div class="modal-actions"><button class="danger-action" id="delete-planning-task" type="button">Supprimer</button></div></section></div>`;
}

function renderCalendar(tasks) {
  const [start, end] = planningRange();
  const title = new Intl.DateTimeFormat('fr-FR', { month: 'long', year: 'numeric', day: ui.calendarView === 'day' ? 'numeric' : undefined, weekday: ui.calendarView === 'day' ? 'long' : undefined }).format(new Date(`${ui.calendarDate}T12:00:00`));
  const dayCount = Math.round((new Date(`${end}T12:00:00`) - new Date(`${start}T12:00:00`)) / 86400000) + 1;
  const days = Array.from({ length: dayCount }, (_, index) => { const date = new Date(`${start}T12:00:00`); date.setDate(date.getDate() + index); return isoDate(date); });
  const content = ui.calendarView === 'day' ? renderDayTimeline(tasks) : `<div class="${ui.calendarView === 'month' ? 'calendar-weekdays' : ''}">${ui.calendarView === 'month' ? ['Lun.', 'Mar.', 'Mer.', 'Jeu.', 'Ven.', 'Sam.', 'Dim.'].map((name) => `<span>${name}</span>`).join('') : ''}</div><div class="calendar-grid ${ui.calendarView}">${days.map((date) => {
    const dayTasks = tasks.filter((task) => task.date === date);
    return `<article class="calendar-day ${date === INITIAL_PLANNING_DATE ? 'planning-initial-date' : ''}" data-drop-date="${date}"><h3>${new Intl.DateTimeFormat('fr-FR', { weekday: ui.calendarView === 'week' ? 'short' : undefined, day: 'numeric' }).format(new Date(`${date}T12:00:00`))}<small>${dayTasks.length ? `${dayTasks.length} tâche${dayTasks.length > 1 ? 's' : ''}` : ''}</small></h3>${dayTasks.map((task) => calendarTaskButton(task)).join('') || '<p class="empty-calendar">Libre</p>'}</article>`;
  }).join('')}</div>`;
  return `<section class="calendar card"><header class="calendar-header"><div><p class="eyebrow">AGENDA</p><h2>Calendrier</h2><p>${escapeHTML(title)}</p></div><div><button data-calendar-nav="-1" type="button" aria-label="Période précédente">←</button>${['month', 'week', 'day'].map((view) => `<button data-calendar-view="${view}" class="${ui.calendarView === view ? 'active' : ''}" type="button">${view === 'month' ? 'Mois' : view === 'week' ? 'Semaine' : 'Jour'}</button>`).join('')}<button data-calendar-nav="1" type="button" aria-label="Période suivante">→</button></div></header>${content}</section>`;
}

const INITIAL_PLANNING_DATE = '2026-08-01';

function calendarTaskButton(task) {
  return `<button draggable="true" data-calendar-task="${task.id}" class="calendar-task ${task.completed ? 'completed' : ''}" style="--task-color:${escapeHTML(task.color || '#D4D4D4')}" type="button"><span>${escapeHTML(task.startTime)} · ${escapeHTML(task.title)}</span>${task.subjectName ? `<small>${escapeHTML(task.subjectName)}${task.chapterName ? ` · ${escapeHTML(task.chapterName)}` : ''}</small>` : ''}</button>`;
}

function renderDayTimeline(tasks) {
  const rows = Array.from({ length: 60 }, (_, index) => {
    const minutes = 7 * 60 + index * 15;
    const time = `${String(Math.floor(minutes / 60)).padStart(2, '0')}:${String(minutes % 60).padStart(2, '0')}`;
    return `<div class="planning-time-slot" data-drop-date="${ui.calendarDate}" data-drop-time="${time}"><time>${time}</time><div>${tasks.filter((task) => task.startTime === time).map((task) => calendarTaskButton(task)).join('')}</div></div>`;
  }).join('');
  return `<div class="planning-day-timeline">${rows}</div>`;
}

function bindTaskFormDependencies(form, subjects) {
  const subjectInput = form.querySelector('[data-task-subject]');
  const chapterInput = form.querySelector('[data-task-chapter]');
  if (!subjectInput || !chapterInput) return;
  subjectInput.addEventListener('change', () => {
    chapterInput.innerHTML = chapterOptions(subjects, subjectInput.value);
  });
}

function generatorInput(subjects, form) {
  return selectedGeneratorChapters(subjects).map(({ chapter, key }) => ({
    chapterId: Number(chapter.id),
    startDays: [...form.querySelectorAll(`[data-start-day="${key}"]:checked`)].map((input) => Number(input.value)),
    revisionCount: Math.max(1, Number(form.querySelector(`[data-generator-field="${key}"][data-field="revisionCount"]`)?.value || 1)),
    durationMinutes: Math.max(15, Number(form.querySelector(`[data-generator-field="${key}"][data-field="durationMinutes"]`)?.value || 60)),
    spacingDays: Math.max(0, Number(form.querySelector(`[data-generator-field="${key}"][data-field="spacingDays"]`)?.value || 0)),
  }));
}

function timeToMinutes(value) {
  const [hours, minutes] = String(value).split(':').map(Number);
  return hours * 60 + minutes;
}

function minutesToTime(value) {
  const limited = Math.min(Math.max(value, 0), (23 * 60) + 59);
  return `${String(Math.floor(limited / 60)).padStart(2, '0')}:${String(limited % 60).padStart(2, '0')}`;
}

function bindPlanningActions(data, preferences, tasks) {
  document.querySelectorAll('[data-planning-tab]').forEach((button) => button.addEventListener('click', () => {
    ui.planningTab = button.dataset.planningTab;
    ui.editingPlanningTask = null;
    renderPlanning();
  }));
  ['planning-task-form', 'planning-edit-form'].forEach((formID) => {
    const form = document.querySelector(`#${formID}`);
    if (form) bindTaskFormDependencies(form, data.subjects);
  });

  const taskFormElement = document.querySelector('#planning-task-form');
  if (taskFormElement) taskFormElement.addEventListener('submit', async (event) => {
    event.preventDefault();
    if (!taskFormElement.reportValidity()) return;
    try {
      await api('CreatePlanningTask', planningTaskInput(taskFormElement));
      await renderPlanning();
      showNotice('Tâche ajoutée au calendrier.');
    } catch (error) { showNotice(`Impossible d’ajouter la tâche : ${errorMessage(error)}`, 'error'); }
  });

  const download = document.querySelector('#download-template');
  if (download) download.addEventListener('click', async () => {
    try {
      await api('DownloadPlanningTemplate');
      showNotice('Le modèle Excel a été téléchargé.');
    } catch (error) { showNotice(`Impossible de télécharger le modèle : ${errorMessage(error)}`, 'error'); }
  });
  const importTemplate = document.querySelector('#import-template');
  if (importTemplate) importTemplate.addEventListener('click', async () => {
    setPlanningImportStatus('Ouverture du sélecteur de fichiers…', 'info', true);
    try {
      const result = await api('ImportPlanningTemplate');
      const imported = result?.imported?.length || 0;
      const skipped = result?.skipped || [];
      if (!imported && !skipped.length) {
        setPlanningImportStatus('Aucun fichier n’a été sélectionné. L’import a été annulé.', 'info');
        return;
      }
      const message = `${imported} créneau${imported > 1 ? 'x' : ''} importé${imported > 1 ? 's' : ''}${skipped.length ? ` ; ${skipped.join(' · ')}` : ''}.`;
      setPlanningImportStatus(message, skipped.length ? 'warning' : 'success');
      await renderPlanning();
      showNotice(message, skipped.length ? 'warning' : 'success');
    } catch (error) {
      setPlanningImportStatus(`Impossible d’importer le fichier : ${errorMessage(error)}`, 'error');
    }
  });

  document.querySelectorAll('[data-planning-chapter]').forEach((input) => input.addEventListener('change', () => {
    const key = input.dataset.planningChapter;
    if (input.checked) ui.planningSelection.set(key, { startDays: [1], revisionCount: 2, durationMinutes: 60, spacingDays: 7 });
    else ui.planningSelection.delete(key);
    renderPlanning();
  }));
  document.querySelectorAll('[data-generator-field]').forEach((input) => input.addEventListener('change', () => {
    const setting = ui.planningSelection.get(input.dataset.generatorField);
    if (setting) setting[input.dataset.field] = Number(input.value);
  }));
  document.querySelectorAll('[data-start-day]').forEach((input) => input.addEventListener('change', () => {
    const setting = ui.planningSelection.get(input.dataset.startDay);
    if (!setting) return;
    setting.startDays = [...document.querySelectorAll(`[data-start-day="${input.dataset.startDay}"]:checked`)].map((item) => Number(item.value));
  }));
  const generate = document.querySelector('#generate-form');
  if (generate) generate.addEventListener('submit', async (event) => {
    event.preventDefault();
    const selections = generatorInput(data.subjects, generate);
    if (!selections.length || selections.some((selection) => !selection.startDays.length)) {
      showNotice('Choisissez au moins un jour de début pour chaque chapitre.', 'warning');
      return;
    }
    try {
      const generated = await api('GeneratePlanning', { selections, startDate: new FormData(generate).get('startDate') });
      ui.planningSelection.clear();
      await renderPlanning();
      showNotice(`${generated.length} révision${generated.length > 1 ? 's' : ''} ajoutée${generated.length > 1 ? 's' : ''} au calendrier.`);
    } catch (error) { showNotice(`Impossible de générer le planning : ${errorMessage(error)}`, 'error'); }
  });

  const workday = document.querySelector('#workday-form');
  if (workday) workday.addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = new FormData(workday);
    const slots = preferences.slots.map((slot) => ({
      period: slot.period,
      enabled: form.get(`${slot.period}-enabled`) === 'on',
      start: form.get(`${slot.period}-start`),
      end: form.get(`${slot.period}-end`),
    }));
    try {
      await api('SaveWorkdayPreferences', { slots });
      await renderPlanning();
      showNotice('Journée de travail enregistrée.');
    } catch (error) { showNotice(`Impossible d’enregistrer les horaires : ${errorMessage(error)}`, 'error'); }
  });

  document.querySelectorAll('[data-calendar-view]').forEach((button) => button.addEventListener('click', () => {
    ui.calendarView = button.dataset.calendarView;
    renderPlanning();
  }));
  document.querySelectorAll('[data-calendar-nav]').forEach((button) => button.addEventListener('click', () => {
    const date = new Date(`${ui.calendarDate}T12:00:00`);
    const direction = Number(button.dataset.calendarNav);
    if (ui.calendarView === 'month') date.setMonth(date.getMonth() + direction);
    else date.setDate(date.getDate() + (ui.calendarView === 'week' ? direction * 7 : direction));
    ui.calendarDate = isoDate(date);
    renderPlanning();
  }));
  document.querySelectorAll('[data-calendar-task]').forEach((button) => button.addEventListener('click', () => {
    ui.editingPlanningTask = tasks.find((task) => Number(task.id) === Number(button.dataset.calendarTask)) || null;
    renderPlanning();
  }));
  document.querySelectorAll('[data-close-planning-modal]').forEach((element) => element.addEventListener('click', (event) => {
    if (event.target !== element && !element.classList.contains('modal-close')) return;
    ui.editingPlanningTask = null;
    renderPlanning();
  }));

  const editForm = document.querySelector('#planning-edit-form');
  if (editForm) {
    editForm.addEventListener('submit', async (event) => {
      event.preventDefault();
      if (!editForm.reportValidity()) return;
      try {
        await api('UpdatePlanningTask', Number(ui.editingPlanningTask.id), planningTaskInput(editForm));
        ui.editingPlanningTask = null;
        await renderPlanning();
        showNotice('Tâche modifiée.');
      } catch (error) { showNotice(`Impossible de modifier la tâche : ${errorMessage(error)}`, 'error'); }
    });
    document.querySelector('#delete-planning-task').addEventListener('click', async () => {
      if (!window.confirm('Supprimer cette tâche du planning ?')) return;
      try {
        await api('DeletePlanningTask', Number(ui.editingPlanningTask.id));
        ui.editingPlanningTask = null;
        await renderPlanning();
        showNotice('Tâche supprimée.');
      } catch (error) { showNotice(`Impossible de supprimer la tâche : ${errorMessage(error)}`, 'error'); }
    });
  }

  document.querySelectorAll('[data-calendar-task]').forEach((button) => button.addEventListener('dragstart', (event) => {
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', button.dataset.calendarTask);
  }));
  document.querySelectorAll('[data-drop-date]').forEach((target) => {
    target.addEventListener('dragover', (event) => {
      event.preventDefault();
      target.classList.add('drop-target');
    });
    target.addEventListener('dragleave', () => target.classList.remove('drop-target'));
    target.addEventListener('drop', async (event) => {
      event.preventDefault();
      target.classList.remove('drop-target');
      const task = tasks.find((item) => Number(item.id) === Number(event.dataTransfer.getData('text/plain')));
      if (!task) return;
      const startTime = target.dataset.dropTime || task.startTime;
      const duration = timeToMinutes(task.endTime) - timeToMinutes(task.startTime);
      const endTime = target.dataset.dropTime ? minutesToTime(timeToMinutes(startTime) + duration) : task.endTime;
      try {
        await api('MovePlanningTask', task.id, target.dataset.dropDate, startTime, endTime);
        await renderPlanning();
        showNotice('Tâche déplacée dans le calendrier.');
      } catch (error) { showNotice(`Impossible de déplacer la tâche : ${errorMessage(error)}`, 'error'); }
    });
  });
}

render();
