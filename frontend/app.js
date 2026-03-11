/* ============================================================
   OpenClaw Installer — Frontend Logic
   ============================================================ */

// --------------- State ---------------
const state = {
  currentStep: 1,
  status: null,       // from /api/status
  projectDir: '',     // default install path
  devServerRunning: false,
  cloneDone: false,
  completedSteps: new Set(),
};

// --------------- DOM Helpers ---------------
const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

// --------------- Initialization ---------------
document.addEventListener('DOMContentLoaded', () => {
  renderStep(1);
});

// ============================================================
// Step Navigation
// ============================================================
function goToStep(n) {
  if (n < 1 || n > 4) return;
  state.currentStep = n;
  renderStep(n);

  // Trigger step-specific actions
  if (n === 2) checkEnvironment();
}

function nextStep() {
  if (!validateCurrentStep()) return;
  state.completedSteps.add(state.currentStep);
  goToStep(state.currentStep + 1);
}

function prevStep() {
  goToStep(state.currentStep - 1);
}

function renderStep(n) {
  // Hide all steps, show current
  $$('.step').forEach((el) => (el.style.display = 'none'));
  const target = $(`#step-${n}`);
  if (target) target.style.display = 'block';

  // Update step dots
  $$('.step-dot').forEach((dot, i) => {
    const stepNum = i + 1;
    dot.classList.remove('active', 'done');
    if (stepNum === n) dot.classList.add('active');
    else if (state.completedSteps.has(stepNum)) dot.classList.add('done');
  });

  // Update lines between dots
  $$('.step-line').forEach((line, i) => {
    line.classList.toggle('done', state.completedSteps.has(i + 1));
  });

  // Navigation buttons
  const btnPrev = $('#btn-prev');
  const btnNext = $('#btn-next');
  btnPrev.style.display = n > 1 ? 'inline-flex' : 'none';
  btnNext.style.display = n < 4 ? 'inline-flex' : 'none';
}

function validateCurrentStep() {
  switch (state.currentStep) {
    case 2:
      if (!state.status) {
        toast('请先完成环境检测', 'error');
        return false;
      }
      if (!state.status.git.installed || !state.status.node.installed) {
        toast('请先安装所有缺失的组件', 'error');
        return false;
      }
      return true;
    case 3:
      if (!state.cloneDone) {
        toast('请先下载并安装项目', 'error');
        return false;
      }
      return true;
    default:
      return true;
  }
}

// ============================================================
// Step 2: Environment Detection
// ============================================================
async function checkEnvironment() {
  // Reset UI
  setStatusLoading('git-status');
  setStatusLoading('node-status');
  $('#btn-install-git').style.display = 'none';
  $('#btn-install-node').style.display = 'none';
  $('#install-all-section').style.display = 'none';
  $('#btn-recheck').style.display = 'none';
  $('#card-git').classList.remove('ok', 'fail');
  $('#card-node').classList.remove('ok', 'fail');

  try {
    const res = await api('/api/status');
    state.status = res;
    state.projectDir = res.projectDir || '';
    state.cloneDone = res.cloned || false;
    $('#project-dir').value = state.projectDir;

    // Git status
    if (res.git.installed) {
      setStatusOk('git-status', `已安装 ${res.git.version || ''}`);
      $('#card-git').classList.add('ok');
    } else {
      setStatusFail('git-status');
      $('#card-git').classList.add('fail');
      $('#btn-install-git').style.display = 'inline-flex';
    }

    // Node status
    if (res.node.installed) {
      setStatusOk('node-status', `已安装 ${res.node.version || ''}`);
      $('#card-node').classList.add('ok');
    } else {
      setStatusFail('node-status');
      $('#card-node').classList.add('fail');
      $('#btn-install-node').style.display = 'inline-flex';
    }

    // Show install-all button if anything missing
    if (!res.git.installed || !res.node.installed) {
      $('#install-all-section').style.display = 'block';
    }

    $('#btn-recheck').style.display = 'inline-flex';
  } catch (err) {
    setStatusFail('git-status');
    setStatusFail('node-status');
    $('#btn-recheck').style.display = 'inline-flex';
    toast('环境检测失败: ' + err.message, 'error');
  }
}

function setStatusLoading(id) {
  $(`#${id}`).innerHTML = '<span class="status-loading"><span class="spinner-sm"></span> 检测中...</span>';
}
function setStatusOk(id, text) {
  $(`#${id}`).innerHTML = `<span style="color:var(--success)">&#10003; ${text}</span>`;
}
function setStatusFail(id) {
  $(`#${id}`).innerHTML = '<span style="color:var(--error)">&#10007; 未检测到</span>';
}

async function installTool(tool) {
  // Use the unified prereqs install endpoint for any tool
  return installAllMissing();
}

async function installAllMissing() {
  const btn = $('#btn-install-all');
  const btnGit = $('#btn-install-git');
  const btnNode = $('#btn-install-node');
  setButtonLoading(btn, true);
  if (btnGit) setButtonLoading(btnGit, true);
  if (btnNode) setButtonLoading(btnNode, true);
  showInstallProgress();

  try {
    await api('/api/prereqs/install', { method: 'POST' });
    appendLog('install-log', '正在安装所有缺失组件...');

    const es = new EventSource('/api/prereqs/progress');
    attachSSE(es, 'install-log', () => {
      es.close();
      setButtonLoading(btn, false);
      if (btnGit) setButtonLoading(btnGit, false);
      if (btnNode) setButtonLoading(btnNode, false);
      checkEnvironment();
      toast('所有组件安装完成', 'success');
    });
  } catch (err) {
    setButtonLoading(btn, false);
    if (btnGit) setButtonLoading(btnGit, false);
    if (btnNode) setButtonLoading(btnNode, false);
    toast('安装失败: ' + err.message, 'error');
  }
}

function showInstallProgress() {
  $('#install-progress').style.display = 'block';
  clearLog('install-log');
}

// ============================================================
// Step 3: Project Clone
// ============================================================
function toggleDirEdit() {
  const input = $('#project-dir');
  input.focus();
  input.select();
}

async function cloneProject() {
  const dir = $('#project-dir').value.trim();
  if (!dir) {
    toast('请填写安装路径', 'error');
    return;
  }
  state.projectDir = dir;

  const btn = $('#btn-clone');
  setButtonLoading(btn, true);
  $('#clone-progress').style.display = 'block';
  $('#btn-clone-retry').style.display = 'none';
  clearLog('clone-log');

  // Reset phases
  setPhase('phase-clone', 'active');
  setPhase('phase-npm', 'pending');
  setProgressBar('clone-progress-fill', 0);

  try {
    await api('/api/project/clone', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ targetDir: dir }),
    });

    // Listen to SSE for clone + dependency install progress
    // Backend sends plain strings: "Starting dependency install..." marks phase transition
    const es = new EventSource('/api/project/progress');
    let currentPhase = 'clone';

    es.addEventListener('message', (e) => {
      const msg = e.data;

      // Completion
      if (msg === 'DONE') {
        es.close();
        setPhase('phase-npm', 'done');
        setProgressBar('clone-progress-fill', 100);
        setButtonLoading(btn, false);
        state.cloneDone = true;
        toast('项目下载并安装完成', 'success');
        return;
      }

      // Error
      if (msg.startsWith('ERROR:')) {
        es.close();
        appendLog('clone-log', msg, 'error');
        setButtonLoading(btn, false);
        $('#btn-clone-retry').style.display = 'inline-flex';
        toast('项目下载失败', 'error');
        return;
      }

      // Phase transition: detect dependency install start
      if (msg.includes('Starting dependency install') && currentPhase === 'clone') {
        currentPhase = 'npm';
        setPhase('phase-clone', 'done');
        setPhase('phase-npm', 'active');
        setProgressBar('clone-progress-fill', 50);
      }

      appendLog('clone-log', msg);
    });

    es.addEventListener('error', () => {
      es.close();
      setButtonLoading(btn, false);
      $('#btn-clone-retry').style.display = 'inline-flex';
    });
  } catch (err) {
    setButtonLoading(btn, false);
    $('#btn-clone-retry').style.display = 'inline-flex';
    toast('项目下载失败: ' + err.message, 'error');
  }
}

function setPhase(id, status) {
  const el = $(`#${id}`);
  el.classList.remove('active', 'done');
  if (status === 'active') {
    el.style.opacity = '1';
    el.classList.add('active');
    // Show spinner
    el.querySelector('.phase-icon').innerHTML = '<span class="spinner-sm"></span>';
  } else if (status === 'done') {
    el.style.opacity = '1';
    el.classList.add('done');
    el.querySelector('.phase-icon').innerHTML = '<span style="color:var(--success);font-size:16px">&#10003;</span>';
  } else {
    el.style.opacity = '0.4';
  }
}

// ============================================================
// Step 4: Dev Server
// ============================================================
let serverEventSource = null;

async function startServer() {
  const btn = $('#btn-start-server');
  setButtonLoading(btn, true);
  $('#btn-server-retry').style.display = 'none';
  $('#server-log').style.display = 'block';
  clearLog('server-log');
  updateServerStatus('starting');

  try {
    await api('/api/dev/start', { method: 'POST' });
    state.devServerRunning = true;

    btn.style.display = 'none';
    $('#btn-stop-server').style.display = 'inline-flex';
    setButtonLoading(btn, false);

    // SSE for dev server logs (plain string messages)
    serverEventSource = new EventSource('/api/dev/logs');
    serverEventSource.addEventListener('message', (e) => {
      const msg = e.data;

      // DONE means server stopped
      if (msg === 'DONE') {
        serverEventSource.close();
        serverEventSource = null;
        state.devServerRunning = false;
        updateServerStatus(false);
        $('#btn-start-server').style.display = 'inline-flex';
        $('#btn-stop-server').style.display = 'none';
        $('#btn-open-browser').style.display = 'none';
        return;
      }

      appendLog('server-log', msg);

      if (msg === '[dev server exited]') {
        state.devServerRunning = false;
        updateServerStatus(false);
        $('#btn-start-server').style.display = 'inline-flex';
        $('#btn-stop-server').style.display = 'none';
        $('#btn-server-retry').style.display = 'inline-flex';
        $('#btn-open-browser').style.display = 'none';
        return;
      }

      // Detect runtime ready URLs
      if (msg.includes('Dashboard URL:') || (msg.includes('Local:') && msg.includes('http'))) {
        const match = msg.match(/(https?:\/\/[^\s]+)/);
        if (match) {
          updateServerStatus(true);
          showOpenBrowserButton(match[1]);
        }
      }
    });

    serverEventSource.addEventListener('error', () => {
      if (serverEventSource) {
        serverEventSource.close();
        serverEventSource = null;
      }
      // Don't update status immediately — may be a transient error
    });
  } catch (err) {
    setButtonLoading(btn, false);
    $('#btn-server-retry').style.display = 'inline-flex';
    toast('启动服务器失败: ' + err.message, 'error');
  }
}

async function stopServer() {
  const btn = $('#btn-stop-server');
  setButtonLoading(btn, true);

  try {
    await api('/api/dev/stop', { method: 'POST' });
    if (serverEventSource) {
      serverEventSource.close();
      serverEventSource = null;
    }
    state.devServerRunning = false;
    updateServerStatus(false);
    btn.style.display = 'none';
    $('#btn-start-server').style.display = 'inline-flex';
    $('#btn-open-browser').style.display = 'none';
    appendLog('server-log', '服务器已停止', 'success');
  } catch (err) {
    toast('停止服务器失败: ' + err.message, 'error');
  } finally {
    setButtonLoading(btn, false);
  }
}

function updateServerStatus(running) {
  const dot = $('.status-dot');
  const text = $('.status-text');
  dot.classList.remove('status-running', 'status-stopped');
  if (running === 'starting') {
    dot.classList.add('status-running');
    text.textContent = '启动中';
  } else if (running) {
    dot.classList.add('status-running');
    text.textContent = '运行中';
  } else {
    dot.classList.add('status-stopped');
    text.textContent = '已停止';
  }
}

function showOpenBrowserButton(url) {
  const btn = $('#btn-open-browser');
  btn.href = url;
  btn.style.display = 'inline-flex';
}

function openInBrowser(e) {
  // Just let the link open naturally in a new tab
  // No need to prevent default
}

// ============================================================
// Shared Utilities
// ============================================================

// --- API wrapper ---
async function api(url, options = {}) {
  let res;
  try {
    res = await fetch(url, options);
  } catch (e) {
    throw new Error('网络连接失败，请检查安装器是否在运行');
  }
  if (!res.ok) {
    let body;
    try {
      body = await res.json();
    } catch {
      throw new Error(`请求失败 (${res.status})`);
    }
    throw new Error(body.error || body.message || `请求失败 (${res.status})`);
  }
  // Some endpoints return empty body
  const text = await res.text();
  if (!text) return {};
  try {
    return JSON.parse(text);
  } catch {
    return { message: text };
  }
}

// --- SSE helper ---
// Backend sends plain strings via SSE. "DONE" signals completion.
// Lines starting with "ERROR:" signal errors.
function attachSSE(es, logId, onDone) {
  es.addEventListener('message', (e) => {
    const msg = e.data;

    // Completion sentinel
    if (msg === 'DONE') {
      onDone();
      return;
    }

    // Error messages
    if (msg.startsWith('ERROR:')) {
      appendLog(logId, msg, 'error');
      es.close();
      return;
    }

    // Normal log line
    appendLog(logId, msg);
  });

  es.addEventListener('error', () => {
    es.close();
  });
}

// --- Log viewer ---
function appendLog(logId, text, type) {
  const viewer = $(`#${logId}`);
  const entry = document.createElement('div');
  entry.className = 'log-entry' + (type ? ` ${type}` : '');
  entry.textContent = text;
  viewer.appendChild(entry);

  // Auto-scroll unless user scrolled up
  const isNearBottom = viewer.scrollHeight - viewer.scrollTop - viewer.clientHeight < 60;
  if (isNearBottom) {
    viewer.scrollTop = viewer.scrollHeight;
  }
}

function clearLog(logId) {
  $(`#${logId}`).innerHTML = '';
}

// --- Progress bar ---
function setProgressBar(id, pct) {
  $(`#${id}`).style.width = Math.min(100, Math.max(0, pct)) + '%';
}

// --- Button loading ---
function setButtonLoading(btn, loading) {
  if (loading) {
    btn.classList.add('loading');
    btn.disabled = true;
  } else {
    btn.classList.remove('loading');
    btn.disabled = false;
  }
}

// --- Toast notifications ---
function toast(message, type = 'info') {
  const container = $('#toast-container');
  const el = document.createElement('div');
  el.className = `toast toast-${type}`;
  el.textContent = message;
  container.appendChild(el);

  setTimeout(() => {
    if (el.parentNode) el.remove();
  }, 3000);
}
