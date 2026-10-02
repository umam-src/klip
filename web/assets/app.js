(() => {
  const list = document.querySelector('#space-list');
  const refresh = document.querySelector('#refresh');

  async function loadSpaces() {
    refresh.disabled = true;
    try {
      const response = await fetch('/api/v1/ruang', { headers: { Accept: 'application/json' } });
      if (!response.ok) throw new Error('request gagal');
      const spaces = await response.json();
      if (!Array.isArray(spaces) || spaces.length === 0) {
        list.innerHTML = '<div class="empty"><strong>Belum ada ruang.</strong><span>Buat ruang pertama melalui API atau nanti dari antarmuka Klip.</span></div>';
        return;
      }
      list.innerHTML = spaces.map((space) => `<article class="card"><h3>${escapeHTML(space.name || 'Tanpa nama')}</h3><p>${escapeHTML(space.id || '')}</p></article>`).join('');
    } catch (_) {
      list.innerHTML = '<div class="empty"><strong>Ruang belum dapat dimuat.</strong><span>Pastikan layanan Klip sedang berjalan.</span></div>';
    } finally {
      refresh.disabled = false;
    }
  }

  function escapeHTML(value) {
    return String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  }

  refresh.addEventListener('click', loadSpaces);
  loadSpaces();
})();
