(() => {
  const state = {
    ruangID: '',
    pekerjaanID: '',
    agen: [],
    assignments: new Map(),
  };

  const $ = (selector) => document.querySelector(selector);

  async function request(path, options = {}) {
    const response = await fetch(path, { ...options, headers: { Accept: 'application/json', ...(options.headers || {}) } });
    if (!response.ok) throw new Error(`request ${response.status}`);
    return response.status === 204 ? null : response.json();
  }

  function escapeHTML(value) {
    return String(value ?? '').replace(/[&<>"']/g, (char) => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;',
    })[char]);
  }

  function id() {
    if (globalThis.crypto?.randomUUID) return crypto.randomUUID();
    return `klip-${Date.now()}-${Math.random().toString(16).slice(2)}`;
  }

  async function loadAgents(ruangID) {
    state.ruangID = ruangID;
    state.agen = await request(`/api/v1/ruang/${encodeURIComponent(ruangID)}/agen`);
    state.assignments.clear();
  }

  async function loadAssignment(tugasID) {
    if (state.assignments.has(tugasID)) return state.assignments.get(tugasID);
    const assignment = await request(`/api/v1/tugas/${encodeURIComponent(tugasID)}/agen`);
    state.assignments.set(tugasID, assignment);
    return assignment;
  }

  function agentName(agentID) {
    return state.agen.find((agent) => agent.id === agentID)?.name || 'Agen tidak ditemukan';
  }

  async function prepareTaskForm() {
    const select = $('#task-agent');
    if (!select || !state.ruangID) return;
    select.disabled = true;
    select.innerHTML = '<option value="">Tanpa agen pelaksana</option>';
    for (const agent of state.agen) {
      const option = document.createElement('option');
      option.value = agent.id;
      option.textContent = agent.name || 'Tanpa nama';
      select.appendChild(option);
    }
    select.disabled = false;
  }

  async function refreshTaskList() {
    const target = $('#task-list');
    if (!target) return;
    const buttons = Array.from(target.querySelectorAll('[data-task]'));
    await Promise.all(buttons.map(async (button) => {
      try {
        const assignment = await loadAssignment(button.dataset.task);
        let meta = button.querySelector('[data-task-agent]');
        if (!meta) {
          meta = document.createElement('span');
          meta.dataset.taskAgent = 'true';
          button.appendChild(meta);
        }
        meta.textContent = assignment ? `Agen · ${agentName(assignment.agen_id)}` : 'Agen · Belum ditugaskan';
      } catch (_) {
        // Daftar tugas tetap dapat digunakan saat assignment belum dapat dibaca.
      }
    }));
  }

  async function renderTaskAssignment(tugasID) {
    const card = $('#task-detail-card');
    if (!card) return;
    try {
      const assignment = await loadAssignment(tugasID);
      const row = document.createElement('div');
      row.innerHTML = `<dt>Agen pelaksana</dt><dd>${escapeHTML(assignment ? agentName(assignment.agen_id) : 'Belum ditugaskan')}</dd>`;
      card.querySelector('dl')?.appendChild(row);
    } catch (_) {
      // Detail tugas tetap dapat digunakan tanpa data assignment.
    }
  }

  async function loadAgentTasks(agentID) {
    if (!state.ruangID) return;
    const jobs = await request(`/api/v1/ruang/${encodeURIComponent(state.ruangID)}/pekerjaan`);
    const tasks = [];
    await Promise.all(jobs.map(async (job) => {
      const jobTasks = await request(`/api/v1/pekerjaan/${encodeURIComponent(job.id)}/tugas`);
      await Promise.all(jobTasks.map(async (task) => {
        const assignment = await loadAssignment(task.id);
        if (assignment?.agen_id === agentID) tasks.push({ ...task, pekerjaan: job.title });
      }));
    }));

    const sections = Array.from($('#agent-detail-card')?.querySelectorAll('.item') || []);
    const section = sections.find((item) => item.querySelector('strong')?.textContent?.trim() === 'Pekerjaan');
    if (!section) return;
    if (!tasks.length) {
      section.innerHTML = '<strong>Pekerjaan</strong><p>Belum ada tugas yang ditugaskan ke agen ini.</p>';
      return;
    }
    tasks.sort((a, b) => String(a.title).localeCompare(String(b.title)));
    section.innerHTML = `<strong>Pekerjaan</strong><div class="stack">${tasks.map((task) => `<div class="item"><strong>${escapeHTML(task.title)}</strong><span>Pekerjaan · ${escapeHTML(task.pekerjaan || 'Tanpa pekerjaan')}</span><span>Status · ${escapeHTML(task.status || 'Belum diatur')}</span></div>`).join('')}</div>`;
  }

  document.addEventListener('click', (event) => {
    const space = event.target.closest('[data-space]');
    if (space) {
      loadAgents(space.dataset.space).catch(() => {});
      return;
    }

    const job = event.target.closest('[data-job]');
    if (job) {
      state.pekerjaanID = job.dataset.job;
      setTimeout(() => refreshTaskList().catch(() => {}), 0);
      return;
    }

    const taskToggle = event.target.closest('[data-toggle="task-form"]');
    if (taskToggle) {
      prepareTaskForm().catch(() => {});
      return;
    }

    const task = event.target.closest('[data-task]');
    if (task) {
      setTimeout(() => renderTaskAssignment(task.dataset.task).catch(() => {}), 0);
      return;
    }

    const agent = event.target.closest('[data-agent]');
    if (agent) {
      setTimeout(() => loadAgentTasks(agent.dataset.agent).catch(() => {}), 0);
    }
  }, true);

  document.addEventListener('submit', async (event) => {
    if (event.target.id !== 'task-form') return;
    event.preventDefault();
    event.stopImmediatePropagation();
    if (!state.pekerjaanID) {
      alert('Pekerjaan belum dipilih.');
      return;
    }

    const formElement = event.target;
    const form = new FormData(formElement);
    const submit = formElement.querySelector('button[type="submit"]');
    const agentID = String(form.get('agen_id') || '').trim();
    submit.disabled = true;
    try {
      const task = await request(`/api/v1/pekerjaan/${encodeURIComponent(state.pekerjaanID)}/tugas`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: id(), title: String(form.get('title') || '').trim(), position: 0 }),
      });
      if (agentID) {
        await request(`/api/v1/tugas/${encodeURIComponent(task.id)}/agen`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ agen_id: agentID }),
        });
      }
      state.assignments.delete(task.id);
      formElement.reset();
      formElement.hidden = true;
      const button = document.querySelector(`[data-job="${CSS.escape(state.pekerjaanID)}"]`);
      button?.click();
    } catch (_) {
      alert('Tugas belum dapat disimpan.');
    } finally {
      submit.disabled = false;
    }
  }, true);
})();
