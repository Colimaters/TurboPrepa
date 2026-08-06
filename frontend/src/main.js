import './style.css';
import './app.css';

import { GetDashboard, ToggleTodayTask } from '../wailsjs/go/main/App';

const statuses = [
  { key: 'toPlan', label: 'À planifier', color: 'var(--slate)' },
  { key: 'planned', label: 'Planifié', color: 'var(--lavender)' },
  { key: 'inProgress', label: 'En cours', color: 'var(--peach)' },
  { key: 'mastered', label: 'Maîtrisé', color: 'var(--sage)' },
];

document.querySelector('#app').innerHTML = `
  <div class="shell">
    <header class="topbar">
      <a class="brand" href="#" aria-label="TurboPrepa, accueil">
        <span class="brand-mark">T</span><span>Turbo<span>Prepa</span></span>
      </a>
      <nav aria-label="Navigation principale">
        <button class="nav-item active" type="button" aria-current="page">Accueil</button>
        <button class="nav-item" type="button" disabled>Matières</button>
        <button class="nav-item" type="button" disabled>Planning</button>
        <button class="nav-item" type="button" disabled>Annuaire</button>
        <button class="nav-item" type="button" disabled>Jurisprudence</button>
        <button class="nav-item" type="button" disabled>Veille juridique</button>
        <button class="nav-item" type="button" disabled>Quiz du jour</button>
        <button class="nav-item" type="button" disabled>Concours</button>
        <button class="nav-item" type="button" disabled>Textes</button>
      </nav>
    </header>
    <main class="dashboard">
      <section class="page-heading">
        <div>
          <p class="eyebrow">TABLEAU DE BORD</p>
          <h1>Bonjour, prête à avancer ?</h1>
          <p id="date-label" class="date-label"></p>
        </div>
        <div class="daily-badge"><span aria-hidden="true">✦</span> Un pas à la fois</div>
      </section>
      <p id="load-error" class="load-error" role="alert" hidden></p>
      <section class="dashboard-grid" aria-label="Résumé de préparation">
        <article class="card quote-card">
          <div class="card-heading"><span class="card-icon quote-icon" aria-hidden="true">“</span><p>PENSÉE DU JOUR</p></div>
          <blockquote id="quote-text"></blockquote>
          <p id="quote-author" class="quote-author"></p>
          <p id="quote-source" class="quote-source"></p>
        </article>
        <article class="card progress-card">
          <button id="progress-link" class="card-link" type="button" aria-label="Voir les matières">
            <div class="card-heading"><span class="card-icon progress-icon" aria-hidden="true">◔</span><p>MA PROGRESSION</p><span class="arrow" aria-hidden="true">→</span></div>
            <div class="progress-main">
              <div id="progress-ring" class="progress-ring"><strong id="mastered-percent">0%</strong><span>maîtrisé</span></div>
              <div id="status-list" class="status-list"></div>
            </div>
            <p id="progress-note" class="progress-note"></p>
          </button>
        </article>
      </section>
      <section class="card tasks-card">
        <div class="tasks-header">
          <div class="card-heading"><span class="card-icon task-icon" aria-hidden="true">✓</span><div><p>AUJOURD'HUI</p><h2>Mes tâches du jour</h2></div></div>
          <button id="planning-link" class="text-action" type="button">Voir le planning <span aria-hidden="true">→</span></button>
        </div>
        <div id="tasks" class="tasks-list"></div>
      </section>
    </main>
  </div>
`;

const dateFormatter = new Intl.DateTimeFormat('fr-FR', {
  weekday: 'long', day: 'numeric', month: 'long',
});

function showUnavailable(section) {
  const error = document.querySelector('#load-error');
  error.hidden = false;
  error.textContent = `${section} sera disponible dans une prochaine étape.`;
}

document.querySelector('#progress-link').addEventListener('click', () => showUnavailable('Matières'));
document.querySelector('#planning-link').addEventListener('click', () => showUnavailable('Planning'));

function renderProgress(progress) {
  const total = statuses.reduce((sum, status) => sum + progress[status.key], 0);
  const percent = total ? Math.round((progress.mastered / total) * 100) : 0;
  document.querySelector('#mastered-percent').textContent = `${percent}%`;
  document.querySelector('#progress-ring').style.setProperty('--progress', `${percent * 3.6}deg`);
  document.querySelector('#status-list').innerHTML = statuses.map((status) => `
    <div class="status-row">
      <span class="status-dot" style="background:${status.color}"></span>
      <span>${status.label}</span><strong>${progress[status.key]}</strong>
    </div>
  `).join('');
  document.querySelector('#progress-note').textContent = total
    ? `${progress.mastered} chapitre${progress.mastered > 1 ? 's' : ''} maîtrisé${progress.mastered > 1 ? 's' : ''} sur ${total}`
    : 'Ajoutez vos chapitres dans Matières pour suivre votre progression.';
}

function renderTasks(tasks) {
  const taskList = document.querySelector('#tasks');
  if (!tasks.length) {
    taskList.innerHTML = `
      <div class="empty-state">
        <span class="empty-icon" aria-hidden="true">☀</span>
        <div><h3>Aucune tâche prévue aujourd'hui</h3><p>Profitez-en pour planifier une séance à votre rythme.</p></div>
        <button id="empty-planning-link" class="primary-action" type="button">Planifier ma journée</button>
      </div>`;
    document.querySelector('#empty-planning-link').addEventListener('click', () => showUnavailable('Planning'));
    return;
  }
  taskList.innerHTML = tasks.map((task) => `
    <button class="task-row ${task.completed ? 'done' : ''}" type="button" data-task-id="${task.id}" aria-pressed="${task.completed}">
      <span class="checkmark" aria-hidden="true">${task.completed ? '✓' : ''}</span>
      <span class="task-color" style="background:${task.color}"></span>
      <span class="task-copy"><strong>${task.title}</strong><span>${task.subject}</span></span>
      <time>${task.startTime} – ${task.endTime}</time>
    </button>
  `).join('');
  taskList.querySelectorAll('.task-row').forEach((taskButton) => {
    taskButton.addEventListener('click', async () => {
      const completed = taskButton.getAttribute('aria-pressed') !== 'true';
      taskButton.disabled = true;
      try {
        await ToggleTodayTask(Number(taskButton.dataset.taskId), completed);
        await loadDashboard();
      } catch (error) {
        document.querySelector('#load-error').hidden = false;
        document.querySelector('#load-error').textContent = `Impossible de mettre à jour la tâche : ${error}`;
      } finally {
        taskButton.disabled = false;
      }
    });
  });
}

async function loadDashboard() {
  try {
    const dashboard = await GetDashboard();
    const date = new Date(`${dashboard.today}T12:00:00`);
    document.querySelector('#date-label').textContent = dateFormatter.format(date);
    document.querySelector('#quote-text').textContent = `« ${dashboard.quote.text} »`;
    document.querySelector('#quote-author').textContent = dashboard.quote.uncertainAttribution
      ? `Attribué à ${dashboard.quote.author}` : dashboard.quote.author;
    document.querySelector('#quote-source').textContent = dashboard.quote.source || '';
    renderProgress(dashboard.progress);
    renderTasks(dashboard.tasks);
  } catch (error) {
    const errorElement = document.querySelector('#load-error');
    errorElement.hidden = false;
    errorElement.textContent = `Impossible de charger l'accueil : ${error}`;
  }
}

loadDashboard();
