(() => {
  const $ = (selector) => document.querySelector(selector);

  async function loadSettings() {
    const target = $('#settings-list');
    if (!target) return;
    target.innerHTML = '<div class="empty"><strong>Memuat pengaturan...</strong></div>';
    try {
      const response = await fetch('/api/v1/settings', { headers: { Accept: 'application/json' } });
      if (!response.ok) throw new Error(`request ${response.status}`);
      const settings = await response.json();
      const rows = [
        ['Data lokal', settings.data_dir],
        ['Alamat layanan', settings.listen],
        ['Bahasa', settings.locale],
        ['Penyedia AI', settings.provider],
        ['Alamat penyedia', settings.base_url],
        ['Model', settings.model || 'Belum diatur'],
        ['Kunci API', settings.api_key_set ? 'Sudah diatur' : 'Tidak diatur']
      ];
      target.innerHTML = `<dl class="settings-grid">${rows.map(([label, value]) => `<div><dt>${escapeHTML(label)}</dt><dd>${escapeHTML(value)}</dd></div>`).join('')}</dl>`;
    } catch (_) {
      target.innerHTML = '<div class="empty"><strong>Pengaturan belum dapat dimuat.</strong></div>';
    }
  }

  function escapeHTML(value) {
    return String(value ?? '').replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  }

  document.addEventListener('DOMContentLoaded', () => {
    loadSettings();
    $('#refresh-settings')?.addEventListener('click', loadSettings);
  });
})();
