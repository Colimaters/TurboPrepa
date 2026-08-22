export function showNotice(message, type = 'success') {
  const notice = document.querySelector('#notice');
  if (!notice) return;
  notice.textContent = message;
  notice.className = `notice ${type}`;
  notice.hidden = false;
  window.setTimeout(() => { notice.hidden = true; }, 3500);
}
